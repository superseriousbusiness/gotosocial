/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"codeberg.org/gruf/go-store/kv"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// ProcessingEmoji represents an emoji currently processing. It exposes
// various functions for retrieving data from the process.
type ProcessingEmoji struct {
	mu sync.Mutex

	/*
		below fields should be set on newly created media;
		emoji will be updated incrementally as media goes through processing
	*/

	emoji *gtsmodel.Emoji
	data  DataFunc

	rawData []byte // will be set once the fetchRawData function has been called

	/*
		below fields represent the processing state of the static version of the emoji
	*/

	staticState processState
	static      *ImageMeta

	/*
		below fields represent the processing state of the emoji image
	*/

	fullSizeState processState
	fullSize      *ImageMeta

	/*
		below pointers to database and storage are maintained so that
		the media can store and update itself during processing steps
	*/

	database db.DB
	storage  *kv.KVStore

	err error // error created during processing, if any
}

// EmojiID returns the ID of the underlying emoji without blocking processing.
func (p *ProcessingEmoji) EmojiID() string {
	return p.emoji.ID
}

// LoadEmoji blocks until the static and fullsize image
// has been processed, and then returns the completed emoji.
func (p *ProcessingEmoji) LoadEmoji(ctx context.Context) (*gtsmodel.Emoji, error) {
	if err := p.fetchRawData(ctx); err != nil {
		return nil, err
	}

	if _, err := p.loadStatic(ctx); err != nil {
		return nil, err
	}

	if _, err := p.loadFullSize(ctx); err != nil {
		return nil, err
	}

	return p.emoji, nil
}

// Finished returns true if processing has finished for both the thumbnail
// and full fized version of this piece of media.
func (p *ProcessingEmoji) Finished() bool {
	return p.staticState == complete && p.fullSizeState == complete
}

func (p *ProcessingEmoji) loadStatic(ctx context.Context) (*ImageMeta, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.staticState {
	case received:
		// we haven't processed a static version of this emoji yet so do it now
		static, err := deriveStaticEmoji(p.rawData, p.emoji.ImageContentType)
		if err != nil {
			p.err = fmt.Errorf("error deriving static: %s", err)
			p.staticState = errored
			return nil, p.err
		}

		// put the static in storage
		if err := p.storage.Put(p.emoji.ImageStaticPath, static.image); err != nil {
			p.err = fmt.Errorf("error storing static: %s", err)
			p.staticState = errored
			return nil, p.err
		}

		// set appropriate fields on the emoji based on the static version we derived
		p.attachment.FileMeta.Small = gtsmodel.Small{
			Width:  static.width,
			Height: static.height,
			Size:   static.size,
			Aspect: static.aspect,
		}
		p.attachment.Thumbnail.FileSize = static.size

		if err := putOrUpdateAttachment(ctx, p.database, p.attachment); err != nil {
			p.err = err
			p.thumbstate = errored
			return nil, err
		}

		// set the thumbnail of this media
		p.thumb = static

		// we're done processing the thumbnail!
		p.thumbstate = complete
		fallthrough
	case complete:
		return p.thumb, nil
	case errored:
		return nil, p.err
	}

	return nil, fmt.Errorf("thumbnail processing status %d unknown", p.thumbstate)
}

func (p *ProcessingEmoji) loadFullSize(ctx context.Context) (*ImageMeta, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.fullSizeState {
	case received:
		var clean []byte
		var err error
		var decoded *ImageMeta

		ct := p.attachment.File.ContentType
		switch ct {
		case mimeImageJpeg, mimeImagePng:
			// first 'clean' image by purging exif data from it
			var exifErr error
			if clean, exifErr = purgeExif(p.rawData); exifErr != nil {
				err = exifErr
				break
			}
			decoded, err = decodeImage(clean, ct)
		case mimeImageGif:
			// gifs are already clean - no exif data to remove
			clean = p.rawData
			decoded, err = decodeGif(clean)
		default:
			err = fmt.Errorf("content type %s not a processible image type", ct)
		}

		if err != nil {
			p.err = err
			p.fullSizeState = errored
			return nil, err
		}

		// put the full size in storage
		if err := p.storage.Put(p.attachment.File.Path, decoded.image); err != nil {
			p.err = fmt.Errorf("error storing full size image: %s", err)
			p.fullSizeState = errored
			return nil, p.err
		}

		// set appropriate fields on the attachment based on the image we derived
		p.attachment.FileMeta.Original = gtsmodel.Original{
			Width:  decoded.width,
			Height: decoded.height,
			Size:   decoded.size,
			Aspect: decoded.aspect,
		}
		p.attachment.File.FileSize = decoded.size
		p.attachment.File.UpdatedAt = time.Now()
		p.attachment.Processing = gtsmodel.ProcessingStatusProcessed

		if err := putOrUpdateAttachment(ctx, p.database, p.attachment); err != nil {
			p.err = err
			p.fullSizeState = errored
			return nil, err
		}

		// set the fullsize of this media
		p.fullSize = decoded

		// we're done processing the full-size image
		p.fullSizeState = complete
		fallthrough
	case complete:
		return p.fullSize, nil
	case errored:
		return nil, p.err
	}

	return nil, fmt.Errorf("full size processing status %d unknown", p.fullSizeState)
}

// fetchRawData calls the data function attached to p if it hasn't been called yet,
// and updates the underlying attachment fields as necessary.
// It should only be called from within a function that already has a lock on p!
func (p *ProcessingEmoji) fetchRawData(ctx context.Context) error {
	// check if we've already done this and bail early if we have
	if p.rawData != nil {
		return nil
	}

	// execute the data function and pin the raw bytes for further processing
	b, err := p.data(ctx)
	if err != nil {
		return fmt.Errorf("fetchRawData: error executing data function: %s", err)
	}
	p.rawData = b

	// now we have the data we can work out the content type
	contentType, err := parseContentType(p.rawData)
	if err != nil {
		return fmt.Errorf("fetchRawData: error parsing content type: %s", err)
	}

	if !supportedEmoji(contentType) {
		return fmt.Errorf("fetchRawData: content type %s was not valid for an emoji", contentType)
	}

	split := strings.Split(contentType, "/")
	mainType := split[0]  // something like 'image'
	extension := split[1] // something like 'gif'

	// set some additional fields on the emoji now that
	// we know more about what the underlying image actually is
	p.emoji.ImageURL = uris.GenerateURIForAttachment(p.attachment.AccountID, string(TypeAttachment), string(SizeOriginal), p.attachment.ID, extension)
	p.attachment.File.Path = fmt.Sprintf("%s/%s/%s/%s.%s", p.attachment.AccountID, TypeAttachment, SizeOriginal, p.attachment.ID, extension)
	p.attachment.File.ContentType = contentType

	switch mainType {
	case mimeImage:
		if extension == mimeGif {
			p.attachment.Type = gtsmodel.FileTypeGif
		} else {
			p.attachment.Type = gtsmodel.FileTypeImage
		}
	default:
		return fmt.Errorf("fetchRawData: cannot process mime type %s (yet)", mainType)
	}

	return nil
}

// putOrUpdateEmoji is just a convenience function for first trying to PUT the emoji in the database,
// and then if that doesn't work because the emoji already exists, updating it instead.
func putOrUpdateEmoji(ctx context.Context, database db.DB, emoji *gtsmodel.Emoji) error {
	if err := database.Put(ctx, emoji); err != nil {
		if err != db.ErrAlreadyExists {
			return fmt.Errorf("putOrUpdateEmoji: proper error while putting emoji: %s", err)
		}
		if err := database.UpdateByPrimaryKey(ctx, emoji); err != nil {
			return fmt.Errorf("putOrUpdateEmoji: error while updating emoji: %s", err)
		}
	}

	return nil
}

func (m *manager) preProcessEmoji(ctx context.Context, data DataFunc, shortcode string, ai *AdditionalEmojiInfo) (*ProcessingEmoji, error) {
	instanceAccount, err := m.db.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("preProcessEmoji: error fetching this instance account from the db: %s", err)
	}

	id, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	// populate initial fields on the emoji -- some of these will be overwritten as we proceed
	emoji := &gtsmodel.Emoji{
		ID:                     id,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		Shortcode:              shortcode,
		Domain:                 "", // assume our own domain unless told otherwise
		ImageRemoteURL:         "",
		ImageStaticRemoteURL:   "",
		ImageURL:               "",                                                                                                    // we don't know yet
		ImageStaticURL:         uris.GenerateURIForAttachment(instanceAccount.ID, string(TypeEmoji), string(SizeStatic), id, mimePng), // all static emojis are encoded as png
		ImagePath:              "",                                                                                                    // we don't know yet
		ImageStaticPath:        fmt.Sprintf("%s/%s/%s/%s.%s", instanceAccount.ID, TypeEmoji, SizeStatic, id, mimePng),                 // all static emojis are encoded as png
		ImageContentType:       "",                                                                                                    // we don't know yet
		ImageStaticContentType: mimeImagePng,                                                                                          // all static emojis are encoded as png
		ImageFileSize:          0,
		ImageStaticFileSize:    0,
		ImageUpdatedAt:         time.Now(),
		Disabled:               false,
		URI:                    "", // we don't know yet
		VisibleInPicker:        true,
		CategoryID:             "",
	}

	// check if we have additional info to add to the emoji,
	// and overwrite some of the emoji fields if so
	if ai != nil {
		if ai.CreatedAt != nil {
			attachment.CreatedAt = *ai.CreatedAt
		}

		if ai.StatusID != nil {
			attachment.StatusID = *ai.StatusID
		}

		if ai.RemoteURL != nil {
			attachment.RemoteURL = *ai.RemoteURL
		}

		if ai.Description != nil {
			attachment.Description = *ai.Description
		}

		if ai.ScheduledStatusID != nil {
			attachment.ScheduledStatusID = *ai.ScheduledStatusID
		}

		if ai.Blurhash != nil {
			attachment.Blurhash = *ai.Blurhash
		}

		if ai.Avatar != nil {
			attachment.Avatar = *ai.Avatar
		}

		if ai.Header != nil {
			attachment.Header = *ai.Header
		}

		if ai.FocusX != nil {
			attachment.FileMeta.Focus.X = *ai.FocusX
		}

		if ai.FocusY != nil {
			attachment.FileMeta.Focus.Y = *ai.FocusY
		}
	}

	processingEmoji := &ProcessingEmoji{
		emoji:         emoji,
		data:          data,
		staticState:   received,
		fullSizeState: received,
		database:      m.db,
		storage:       m.storage,
	}

	return processingEmoji, nil
}
