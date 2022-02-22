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
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
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
	return atomic.LoadInt32(&p.staticState) == int32(complete)
}

func (p *ProcessingEmoji) loadStatic(ctx context.Context) error {
	staticState := atomic.LoadInt32(&p.staticState)
	switch processState(staticState) {
	case received:
		// stream the original file out of storage...
		stored, err := p.storage.GetStream(p.emoji.ImagePath)
		if err != nil {
			p.err = fmt.Errorf("loadStatic: error fetching file from storage: %s", err)
			atomic.StoreInt32(&p.staticState, int32(errored))
			return p.err
		}

		// we haven't processed a static version of this emoji yet so do it now
		static, err := deriveStaticEmoji(stored, p.emoji.ImageContentType)
		if err != nil {
			p.err = fmt.Errorf("loadStatic: error deriving static: %s", err)
			atomic.StoreInt32(&p.staticState, int32(errored))
			return p.err
		}

		if err := stored.Close(); err != nil {
			p.err = fmt.Errorf("loadStatic: error closing stored full size: %s", err)
			atomic.StoreInt32(&p.staticState, int32(errored))
			return p.err
		}

		// put the static in storage
		if err := p.storage.Put(p.emoji.ImageStaticPath, static.small); err != nil {
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

	// execute the data function to get the reader out of it
	reader, fileSize, err := p.data(ctx)
	if err != nil {
		return fmt.Errorf("store: error executing data function: %s", err)
	}

	// extract no more than 261 bytes from the beginning of the file -- this is the header
	firstBytes := make([]byte, maxFileHeaderBytes)
	if _, err := reader.Read(firstBytes); err != nil {
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
	p.emoji.ImageURL = uris.GenerateURIForAttachment(p.instanceAccountID, string(TypeEmoji), string(SizeOriginal), p.emoji.ID, extension)
	p.emoji.ImagePath = fmt.Sprintf("%s/%s/%s/%s.%s", p.instanceAccountID, TypeEmoji, SizeOriginal, p.emoji.ID, extension)
	p.emoji.ImageContentType = contentType
	p.emoji.ImageFileSize = fileSize

	// concatenate the first bytes with the existing bytes still in the reader (thanks Mara)
	multiReader := io.MultiReader(bytes.NewBuffer(firstBytes), reader)

	// store this for now -- other processes can pull it out of storage as they please
	if err := p.storage.PutStream(p.emoji.ImagePath, multiReader); err != nil {
		return fmt.Errorf("store: error storing stream: %s", err)
	}

	// if the original reader is a readcloser, close it since we're done with it now
	if rc, ok := reader.(io.ReadCloser); ok {
		if err := rc.Close(); err != nil {
			return fmt.Errorf("store: error closing readcloser: %s", err)
		}
	}

	p.read = true

	if p.postData != nil {
		return p.postData(ctx)
	}

	return nil
}

func (m *manager) preProcessEmoji(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, shortcode string, id string, uri string, ai *AdditionalEmojiInfo) (*ProcessingEmoji, error) {
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
		postData:          postData,
		staticState:       int32(received),
		database:          m.db,
		storage:           m.storage,
	}

	return processingEmoji, nil
}
