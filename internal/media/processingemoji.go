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

	// track whether this emoji has already been put in the databse
	insertedInDB bool
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

	// store the result in the database before returning it
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.insertedInDB {
		if err := p.database.Put(ctx, p.emoji); err != nil {
			return nil, err
		}
		p.insertedInDB = true
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
		p.emoji.ImageStaticFileSize = len(static.image)

		// set the static on the processing emoji
		p.static = static

		// we're done processing the static version of the emoji!
		p.staticState = complete
		fallthrough
	case complete:
		return p.static, nil
	case errored:
		return nil, p.err
	}

	return nil, fmt.Errorf("static processing status %d unknown", p.staticState)
}

func (p *ProcessingEmoji) loadFullSize(ctx context.Context) (*ImageMeta, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.fullSizeState {
	case received:
		var err error
		var decoded *ImageMeta

		ct := p.emoji.ImageContentType
		switch ct {
		case mimeImagePng:
			decoded, err = decodeImage(p.rawData, ct)
		case mimeImageGif:
			decoded, err = decodeGif(p.rawData)
		default:
			err = fmt.Errorf("content type %s not a processible emoji type", ct)
		}

		if err != nil {
			p.err = err
			p.fullSizeState = errored
			return nil, err
		}

		// put the full size emoji in storage
		if err := p.storage.Put(p.emoji.ImagePath, decoded.image); err != nil {
			p.err = fmt.Errorf("error storing full size emoji: %s", err)
			p.fullSizeState = errored
			return nil, p.err
		}

		// set the fullsize of this media
		p.fullSize = decoded

		// we're done processing the full-size emoji
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
// and updates the underlying emoji fields as necessary.
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
	extension := split[1] // something like 'gif'

	// set some additional fields on the emoji now that
	// we know more about what the underlying image actually is
	p.emoji.ImageURL = uris.GenerateURIForAttachment(p.instanceAccountID, string(TypeEmoji), string(SizeOriginal), p.emoji.ID, extension)
	p.emoji.ImagePath = fmt.Sprintf("%s/%s/%s/%s.%s", p.instanceAccountID, TypeEmoji, SizeOriginal, p.emoji.ID, extension)
	p.emoji.ImageContentType = contentType
	p.emoji.ImageFileSize = len(p.rawData)

	return nil
}

func (m *manager) preProcessEmoji(ctx context.Context, data DataFunc, shortcode string, id string, uri string, ai *AdditionalEmojiInfo) (*ProcessingEmoji, error) {
	instanceAccount, err := m.db.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("preProcessEmoji: error fetching this instance account from the db: %s", err)
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
		URI:                    uri,
		VisibleInPicker:        true,
		CategoryID:             "",
	}

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
			emoji.Disabled = *ai.Disabled
		}

		if ai.VisibleInPicker != nil {
			emoji.VisibleInPicker = *ai.VisibleInPicker
		}

		if ai.CategoryID != nil {
			emoji.CategoryID = *ai.CategoryID
		}
	}

	processingEmoji := &ProcessingEmoji{
		instanceAccountID: instanceAccount.ID,
		emoji:             emoji,
		data:              data,
		staticState:       received,
		fullSizeState:     received,
		database:          m.db,
		storage:           m.storage,
	}

	return processingEmoji, nil
}
