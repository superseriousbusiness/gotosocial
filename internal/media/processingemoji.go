/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"codeberg.org/gruf/go-bytesize"
	gostore "codeberg.org/gruf/go-store/v2/storage"
	"github.com/h2non/filetype"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// ProcessingEmoji represents an emoji currently processing. It exposes
// various functions for retrieving data from the process.
type ProcessingEmoji struct {
	instAccID string               // instance account ID
	emoji     *gtsmodel.Emoji      // processing emoji details
	refresh   bool                 // whether this is an existing emoji being refreshed
	newPathID string               // new emoji path ID to use if refreshed
	dataFn    DataFunc             // load-data function, returns media stream
	postFn    PostDataCallbackFunc // post data callback function
	err       error                // error encountered during processing
	manager   *manager             // manager instance (access to db / storage)
	once      sync.Once            // once ensures processing only occurs once
}

// EmojiID returns the ID of the underlying emoji without blocking processing.
func (p *ProcessingEmoji) EmojiID() string {
	return p.emoji.ID // immutable, safe outside mutex.
}

// LoadEmoji blocks until the static and fullsize image
// has been processed, and then returns the completed emoji.
func (p *ProcessingEmoji) LoadEmoji(ctx context.Context) (*gtsmodel.Emoji, error) {
	// only process once.
	p.once.Do(func() {
		var err error

		defer func() {
			if r := recover(); r != nil {
				if err != nil {
					rOld := r // wrap the panic so we don't lose existing returned error
					r = fmt.Errorf("panic occured after error %q: %v", err.Error(), rOld)
				}

				// Catch any panics and wrap as error.
				err = fmt.Errorf("caught panic: %v", r)
			}

			if err != nil {
				// Store error.
				p.err = err
			}
		}()

		// Attempt to store media and calculate
		// full-size media attachment details.
		if err = p.store(ctx); err != nil {
			return
		}

		// Finish processing by reloading media into
		// memory to get dimension and generate a thumb.
		if err = p.finish(ctx); err != nil {
			return
		}

		if p.refresh {
			columns := []string{
				"updated_at",
				"image_remote_url",
				"image_static_remote_url",
				"image_url",
				"image_static_url",
				"image_path",
				"image_static_path",
				"image_content_type",
				"image_file_size",
				"image_static_file_size",
				"image_updated_at",
				"shortcode",
				"uri",
			}

			// Existing emoji we're refreshing, so only need to update.
			_, err = p.manager.db.UpdateEmoji(ctx, p.emoji, columns...)
			return
		}

		// New emoji media, first time caching.
		err = p.manager.db.PutEmoji(ctx, p.emoji)
		return //nolint shutup linter i like this here
	})

	if p.err != nil {
		return nil, p.err
	}

	return p.emoji, nil
}

// store calls the data function attached to p if it hasn't been called yet,
// and updates the underlying attachment fields as necessary. It will then stream
// bytes from p's reader directly into storage so that it can be retrieved later.
func (p *ProcessingEmoji) store(ctx context.Context) error {
	defer func() {
		if p.postFn == nil {
			return
		}

		// Ensure post callback gets called.
		if err := p.postFn(ctx); err != nil {
			log.Errorf("error executing postdata function: %v", err)
		}
	}()

	// Load media from provided data fn.
	rc, sz, err := p.dataFn(ctx)
	if err != nil {
		return fmt.Errorf("error executing data function: %w", err)
	}

	defer func() {
		// Ensure data reader gets closed on return.
		if err := rc.Close(); err != nil {
			log.Errorf("error closing data reader: %v", err)
		}
	}()

	// Byte buffer to read file header into.
	// See: https://en.wikipedia.org/wiki/File_format#File_header
	// and https://github.com/h2non/filetype
	hdrBuf := make([]byte, 261)

	// Read the first 261 header bytes into buffer.
	if _, err := io.ReadFull(rc, hdrBuf); err != nil {
		return fmt.Errorf("error reading incoming media: %w", err)
	}

	// Parse file type info from header buffer.
	info, err := filetype.Match(hdrBuf)
	if err != nil {
		return fmt.Errorf("error parsing file type: %w", err)
	}

	switch info.Extension {
	// only supported emoji types
	case "gif", "png":

	// unhandled
	default:
		return fmt.Errorf("unsupported emoji filetype: %s", info.Extension)
	}

	// Recombine header bytes with remaining stream
	r := io.MultiReader(bytes.NewReader(hdrBuf), rc)

	var maxSize bytesize.Size

	if p.emoji.Domain == "" {
		// this is a local emoji upload
		maxSize = config.GetMediaEmojiLocalMaxSize()
	} else {
		// this is a remote incoming emoji
		maxSize = config.GetMediaEmojiRemoteMaxSize()
	}

	// Check that provided size isn't beyond max. We check beforehand
	// so that we don't attempt to stream the emoji into storage if not needed.
	if size := bytesize.Size(sz); sz > 0 && size > maxSize {
		return fmt.Errorf("given emoji size %s greater than max allowed %s", size, maxSize)
	}

	var pathID string

	if p.refresh {
		// This is a refreshed emoji with a new
		// path ID that this will be stored under.
		pathID = p.newPathID
	} else {
		// This is a new emoji, simply use provided ID.
		pathID = p.emoji.ID
	}

	// Calculate emoji file path.
	p.emoji.ImagePath = fmt.Sprintf(
		"%s/%s/%s/%s.%s",
		p.instAccID,
		TypeEmoji,
		SizeOriginal,
		pathID,
		info.Extension,
	)

	// This shouldn't already exist, but we do a check as it's worth logging.
	if have, _ := p.manager.storage.Has(ctx, p.emoji.ImagePath); have {
		log.Warnf("emoji already exists at storage path: %s", p.emoji.ImagePath)

		// Attempt to remove existing emoji at storage path (might be broken / out-of-date)
		if err := p.manager.storage.Delete(ctx, p.emoji.ImagePath); err != nil {
			return fmt.Errorf("error removing emoji from storage: %v", err)
		}
	}

	// Write the final image reader stream to our storage.
	sz, err = p.manager.storage.PutStream(ctx, p.emoji.ImagePath, r)
	if err != nil {
		return fmt.Errorf("error writing emoji to storage: %w", err)
	}

	// Once again check size in case none was provided previously.
	if size := bytesize.Size(sz); size > maxSize {
		if err := p.manager.storage.Delete(ctx, p.emoji.ImagePath); err != nil {
			log.Errorf("error removing too-large-emoji from storage: %v", err)
		}
		return fmt.Errorf("calculated emoji size %s greater than max allowed %s", size, maxSize)
	}

	// Fill in remaining attachment data now it's stored.
	p.emoji.ImageURL = uris.GenerateURIForAttachment(
		p.instAccID,
		string(TypeEmoji),
		string(SizeOriginal),
		pathID,
		info.Extension,
	)
	p.emoji.ImageContentType = info.MIME.Value
	p.emoji.ImageFileSize = int(sz)

	return nil
}

func (p *ProcessingEmoji) finish(ctx context.Context) error {
	// Fetch a stream to the original file in storage.
	rc, err := p.manager.storage.GetStream(ctx, p.emoji.ImagePath)
	if err != nil {
		return fmt.Errorf("error loading file from storage: %w", err)
	}
	defer rc.Close()

	// Decode the image from storage.
	staticImg, err := decodeImage(rc)
	if err != nil {
		return fmt.Errorf("error decoding image: %w", err)
	}

	// The image should be in-memory by now.
	if err := rc.Close(); err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}

	// This shouldn't already exist, but we do a check as it's worth logging.
	if have, _ := p.manager.storage.Has(ctx, p.emoji.ImageStaticPath); have {
		log.Warnf("static emoji already exists at storage path: %s", p.emoji.ImagePath)

		// Attempt to remove static existing emoji at storage path (might be broken / out-of-date)
		if err := p.manager.storage.Delete(ctx, p.emoji.ImageStaticPath); err != nil {
			return fmt.Errorf("error removing static emoji from storage: %v", err)
		}
	}

	// Create an emoji PNG encoder stream.
	enc := staticImg.ToPNG()

	// Stream-encode the PNG static image into storage.
	sz, err := p.manager.storage.PutStream(ctx, p.emoji.ImageStaticPath, enc)
	if err != nil {
		return fmt.Errorf("error stream-encoding static emoji to storage: %w", err)
	}

	// Set written image size.
	p.emoji.ImageStaticFileSize = int(sz)

	return nil
}

func (m *manager) preProcessEmoji(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, shortcode string, emojiID string, uri string, ai *AdditionalEmojiInfo, refresh bool) (*ProcessingEmoji, error) {
	instanceAccount, err := m.db.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("preProcessEmoji: error fetching this instance account from the db: %s", err)
	}

	var newPathID string
	var emoji *gtsmodel.Emoji
	if refresh {
		emoji, err = m.db.GetEmojiByID(ctx, emojiID)
		if err != nil {
			return nil, fmt.Errorf("preProcessEmoji: error fetching emoji to refresh from the db: %s", err)
		}

		// if this is a refresh, we will end up with new images
		// stored for this emoji, so we can use the postData function
		// to perform clean up of the old images from storage
		originalPostData := postData
		originalImagePath := emoji.ImagePath
		originalImageStaticPath := emoji.ImageStaticPath
		postData = func(innerCtx context.Context) error {
			// trigger the original postData function if it was provided
			if originalPostData != nil {
				if err := originalPostData(innerCtx); err != nil {
					return err
				}
			}

			l := log.WithField("shortcode@domain", emoji.Shortcode+"@"+emoji.Domain)
			l.Debug("postData: cleaning up old emoji files for refreshed emoji")
			if err := m.storage.Delete(innerCtx, originalImagePath); err != nil && !errors.Is(err, gostore.ErrNotFound) {
				l.Errorf("postData: error cleaning up old emoji image at %s for refreshed emoji: %s", originalImagePath, err)
			}
			if err := m.storage.Delete(innerCtx, originalImageStaticPath); err != nil && !errors.Is(err, gostore.ErrNotFound) {
				l.Errorf("postData: error cleaning up old emoji static image at %s for refreshed emoji: %s", originalImageStaticPath, err)
			}

			return nil
		}

		newPathID, err = id.NewRandomULID()
		if err != nil {
			return nil, fmt.Errorf("preProcessEmoji: error generating alternateID for emoji refresh: %s", err)
		}

		// store + serve static image at new path ID
		emoji.ImageStaticURL = uris.GenerateURIForAttachment(instanceAccount.ID, string(TypeEmoji), string(SizeStatic), newPathID, mimePng)
		emoji.ImageStaticPath = fmt.Sprintf("%s/%s/%s/%s.%s", instanceAccount.ID, TypeEmoji, SizeStatic, newPathID, mimePng)

		emoji.Shortcode = shortcode
		emoji.URI = uri
	} else {
		disabled := false
		visibleInPicker := true

		// populate initial fields on the emoji -- some of these will be overwritten as we proceed
		emoji = &gtsmodel.Emoji{
			ID:                     emojiID,
			CreatedAt:              time.Now(),
			Shortcode:              shortcode,
			Domain:                 "", // assume our own domain unless told otherwise
			ImageRemoteURL:         "",
			ImageStaticRemoteURL:   "",
			ImageURL:               "",                                                                                                         // we don't know yet
			ImageStaticURL:         uris.GenerateURIForAttachment(instanceAccount.ID, string(TypeEmoji), string(SizeStatic), emojiID, mimePng), // all static emojis are encoded as png
			ImagePath:              "",                                                                                                         // we don't know yet
			ImageStaticPath:        fmt.Sprintf("%s/%s/%s/%s.%s", instanceAccount.ID, TypeEmoji, SizeStatic, emojiID, mimePng),                 // all static emojis are encoded as png
			ImageContentType:       "",                                                                                                         // we don't know yet
			ImageStaticContentType: mimeImagePng,                                                                                               // all static emojis are encoded as png
			ImageFileSize:          0,
			ImageStaticFileSize:    0,
			Disabled:               &disabled,
			URI:                    uri,
			VisibleInPicker:        &visibleInPicker,
			CategoryID:             "",
		}
	}

	emoji.ImageUpdatedAt = time.Now()
	emoji.UpdatedAt = time.Now()

	// check if we have additional info to add to the emoji,
	// and overwrite some of the emoji fields if so
	if ai != nil {
		if ai.CreatedAt != nil {
			emoji.CreatedAt = *ai.CreatedAt
		}

		if ai.Domain != nil {
			emoji.Domain = *ai.Domain
		}

		if ai.ImageRemoteURL != nil {
			emoji.ImageRemoteURL = *ai.ImageRemoteURL
		}

		if ai.ImageStaticRemoteURL != nil {
			emoji.ImageStaticRemoteURL = *ai.ImageStaticRemoteURL
		}

		if ai.Disabled != nil {
			emoji.Disabled = ai.Disabled
		}

		if ai.VisibleInPicker != nil {
			emoji.VisibleInPicker = ai.VisibleInPicker
		}

		if ai.CategoryID != nil {
			emoji.CategoryID = *ai.CategoryID
		}
	}

	processingEmoji := &ProcessingEmoji{
		instAccID: instanceAccount.ID,
		emoji:     emoji,
		refresh:   refresh,
		newPathID: newPathID,
		dataFn:    data,
		postFn:    postData,
		manager:   m,
	}

	return processingEmoji, nil
}
