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
	"strings"
	"sync"
	"sync/atomic"
	"time"

	gostore "codeberg.org/gruf/go-store/v2/storage"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// ProcessingEmoji represents an emoji currently processing. It exposes
// various functions for retrieving data from the process.
type ProcessingEmoji struct {
	mu sync.Mutex

	// id of this instance's account -- pinned for convenience here so we only need to fetch it once
	instanceAccountID string

	/*
		below fields should be set on newly created media;
		emoji will be updated incrementally as media goes through processing
	*/

	emoji    *gtsmodel.Emoji
	data     DataFunc
	postData PostDataCallbackFunc
	read     bool // bool indicating that data function has been triggered already

	/*
		below fields represent the processing state of the static of the emoji
	*/
	staticState int32

	/*
		below pointers to database and storage are maintained so that
		the media can store and update itself during processing steps
	*/

	database db.DB
	storage  *storage.Driver

	err error // error created during processing, if any

	// track whether this emoji has already been put in the databse
	insertedInDB bool

	// is this a refresh of an existing emoji?
	refresh bool
	// if it is a refresh, which alternate ID should we use in the storage and URL paths?
	newPathID string
}

// EmojiID returns the ID of the underlying emoji without blocking processing.
func (p *ProcessingEmoji) EmojiID() string {
	return p.emoji.ID
}

// LoadEmoji blocks until the static and fullsize image
// has been processed, and then returns the completed emoji.
func (p *ProcessingEmoji) LoadEmoji(ctx context.Context) (*gtsmodel.Emoji, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.store(ctx); err != nil {
		return nil, err
	}

	if err := p.loadStatic(ctx); err != nil {
		return nil, err
	}

	// store the result in the database before returning it
	if !p.insertedInDB {
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
			if _, err := p.database.UpdateEmoji(ctx, p.emoji, columns...); err != nil {
				return nil, err
			}
		} else {
			if err := p.database.PutEmoji(ctx, p.emoji); err != nil {
				return nil, err
			}
		}
		p.insertedInDB = true
	}

	return p.emoji, nil
}

// Finished returns true if processing has finished for both the thumbnail
// and full fized version of this piece of media.
func (p *ProcessingEmoji) Finished() bool {
	return atomic.LoadInt32(&p.staticState) == int32(complete)
}

func (p *ProcessingEmoji) loadStatic(ctx context.Context) error {
	staticState := atomic.LoadInt32(&p.staticState)
	switch processState(staticState) {
	case received:
		// stream the original file out of storage...
		stored, err := p.storage.GetStream(ctx, p.emoji.ImagePath)
		if err != nil {
			p.err = fmt.Errorf("loadStatic: error fetching file from storage: %s", err)
			atomic.StoreInt32(&p.staticState, int32(errored))
			return p.err
		}
		defer stored.Close()

		// we haven't processed a static version of this emoji yet so do it now
		static, err := deriveStaticEmoji(stored, p.emoji.ImageContentType)
		if err != nil {
			p.err = fmt.Errorf("loadStatic: error deriving static: %s", err)
			atomic.StoreInt32(&p.staticState, int32(errored))
			return p.err
		}

		// Close stored emoji now we're done
		if err := stored.Close(); err != nil {
			log.Errorf("loadStatic: error closing stored full size: %s", err)
		}

		// put the static image in storage
		if err := p.storage.Put(ctx, p.emoji.ImageStaticPath, static.small); err != nil && err != storage.ErrAlreadyExists {
			p.err = fmt.Errorf("loadStatic: error storing static: %s", err)
			atomic.StoreInt32(&p.staticState, int32(errored))
			return p.err
		}

		p.emoji.ImageStaticFileSize = len(static.small)

		// we're done processing the static version of the emoji!
		atomic.StoreInt32(&p.staticState, int32(complete))
		fallthrough
	case complete:
		return nil
	case errored:
		return p.err
	}

	return fmt.Errorf("static processing status %d unknown", p.staticState)
}

// store calls the data function attached to p if it hasn't been called yet,
// and updates the underlying attachment fields as necessary. It will then stream
// bytes from p's reader directly into storage so that it can be retrieved later.
func (p *ProcessingEmoji) store(ctx context.Context) error {
	// check if we've already done this and bail early if we have
	if p.read {
		return nil
	}

	// execute the data function to get the readcloser out of it
	rc, fileSize, err := p.data(ctx)
	if err != nil {
		return fmt.Errorf("store: error executing data function: %s", err)
	}

	// defer closing the reader when we're done with it
	defer func() {
		if err := rc.Close(); err != nil {
			log.Errorf("store: error closing readcloser: %s", err)
		}
	}()

	// execute the postData function no matter what happens
	defer func() {
		if p.postData != nil {
			if err := p.postData(ctx); err != nil {
				log.Errorf("store: error executing postData: %s", err)
			}
		}
	}()

	// extract no more than 261 bytes from the beginning of the file -- this is the header
	firstBytes := make([]byte, maxFileHeaderBytes)
	if _, err := rc.Read(firstBytes); err != nil {
		return fmt.Errorf("store: error reading initial %d bytes: %s", maxFileHeaderBytes, err)
	}

	// now we have the file header we can work out the content type from it
	contentType, err := parseContentType(firstBytes)
	if err != nil {
		return fmt.Errorf("store: error parsing content type: %s", err)
	}

	// bail if this is a type we can't process
	if !supportedEmoji(contentType) {
		return fmt.Errorf("store: content type %s was not valid for an emoji", contentType)
	}

	// extract the file extension
	split := strings.Split(contentType, "/")
	extension := split[1] // something like 'gif'

	// set some additional fields on the emoji now that
	// we know more about what the underlying image actually is
	var pathID string
	if p.refresh {
		pathID = p.newPathID
	} else {
		pathID = p.emoji.ID
	}
	p.emoji.ImageURL = uris.GenerateURIForAttachment(p.instanceAccountID, string(TypeEmoji), string(SizeOriginal), pathID, extension)
	p.emoji.ImagePath = fmt.Sprintf("%s/%s/%s/%s.%s", p.instanceAccountID, TypeEmoji, SizeOriginal, pathID, extension)
	p.emoji.ImageContentType = contentType

	// concatenate the first bytes with the existing bytes still in the reader (thanks Mara)
	readerToStore := io.MultiReader(bytes.NewBuffer(firstBytes), rc)

	var maxEmojiSize int64
	if p.emoji.Domain == "" {
		maxEmojiSize = int64(config.GetMediaEmojiLocalMaxSize())
	} else {
		maxEmojiSize = int64(config.GetMediaEmojiRemoteMaxSize())
	}

	// if we know the fileSize already, make sure it's not bigger than our limit
	var checkedSize bool
	if fileSize > 0 {
		checkedSize = true
		if fileSize > maxEmojiSize {
			return fmt.Errorf("store: given emoji fileSize (%db) is larger than allowed size (%db)", fileSize, maxEmojiSize)
		}
	}

	// store this for now -- other processes can pull it out of storage as they please
	if fileSize, err = putStream(ctx, p.storage, p.emoji.ImagePath, readerToStore, fileSize); err != nil {
		if !errors.Is(err, storage.ErrAlreadyExists) {
			return fmt.Errorf("store: error storing stream: %s", err)
		}
		log.Warnf("emoji %s already exists at storage path: %s", p.emoji.ID, p.emoji.ImagePath)
	}

	// if we didn't know the fileSize yet, we do now, so check if we need to
	if !checkedSize && fileSize > maxEmojiSize {
		err = fmt.Errorf("store: discovered emoji fileSize (%db) is larger than allowed emojiRemoteMaxSize (%db), will delete from the store now", fileSize, maxEmojiSize)
		log.Warn(err)
		if deleteErr := p.storage.Delete(ctx, p.emoji.ImagePath); deleteErr != nil {
			log.Errorf("store: error removing too-large emoji from the store: %s", deleteErr)
		}
		return err
	}

	p.emoji.ImageFileSize = int(fileSize)
	p.read = true

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
		instanceAccountID: instanceAccount.ID,
		emoji:             emoji,
		data:              data,
		postData:          postData,
		staticState:       int32(received),
		database:          m.db,
		storage:           m.storage,
		refresh:           refresh,
		newPathID:         newPathID,
	}

	return processingEmoji, nil
}
