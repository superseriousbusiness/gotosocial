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
	"fmt"
	"io"

	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-errors/v2"
	"codeberg.org/gruf/go-runners"
	"github.com/h2non/filetype"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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
	done      bool                 // done is set when process finishes with non ctx canceled type error
	proc      runners.Processor    // proc helps synchronize only a singular running processing instance
	err       error                // error stores permanent error value when done
	mgr       *manager             // mgr instance (access to db / storage)
}

// EmojiID returns the ID of the underlying emoji without blocking processing.
func (p *ProcessingEmoji) EmojiID() string {
	return p.emoji.ID // immutable, safe outside mutex.
}

// LoadEmoji blocks until the static and fullsize image has been processed, and then returns the completed emoji.
func (p *ProcessingEmoji) LoadEmoji(ctx context.Context) (*gtsmodel.Emoji, error) {
	// Attempt to load synchronously.
	emoji, done, err := p.load(ctx)

	if err == nil {
		// No issue, return media.
		return emoji, nil
	}

	if !done {
		// Provided context was cancelled, e.g. request cancelled
		// early. Queue this item for asynchronous processing.
		log.Warnf(ctx, "reprocessing emoji %s after canceled ctx", p.emoji.ID)
		go p.mgr.state.Workers.Media.Enqueue(p.Process)
	}

	return nil, err
}

// Process allows the receiving object to fit the runners.WorkerFunc signature. It performs a (blocking) load and logs on error.
func (p *ProcessingEmoji) Process(ctx context.Context) {
	if _, _, err := p.load(ctx); err != nil {
		log.Errorf(ctx, "error processing emoji: %v", err)
	}
}

// load performs a concurrency-safe load of ProcessingEmoji, only marking itself as complete when returned error is NOT a context cancel.
func (p *ProcessingEmoji) load(ctx context.Context) (*gtsmodel.Emoji, bool, error) {
	var (
		done bool
		err  error
	)

	err = p.proc.Process(func() error {
		if p.done {
			// Already proc'd.
			return p.err
		}

		defer func() {
			// This is only done when ctx NOT cancelled.
			done = err == nil || !errors.Is(err,
				context.Canceled,
				context.DeadlineExceeded,
			)

			if !done {
				return
			}

			// Store final values.
			p.done = true
			p.err = err
		}()

		// Attempt to store media and calculate
		// full-size media attachment details.
		if err = p.store(ctx); err != nil {
			return err
		}

		// Finish processing by reloading media into
		// memory to get dimension and generate a thumb.
		if err = p.finish(ctx); err != nil {
			return err
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
			_, err = p.mgr.state.DB.UpdateEmoji(ctx, p.emoji, columns...)
			return err
		}

		// New emoji media, first time caching.
		err = p.mgr.state.DB.PutEmoji(ctx, p.emoji)
		return err
	})

	if err != nil {
		return nil, done, err
	}

	return p.emoji, done, nil
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
			log.Errorf(ctx, "error executing postdata function: %v", err)
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
			log.Errorf(ctx, "error closing data reader: %v", err)
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
	if have, _ := p.mgr.state.Storage.Has(ctx, p.emoji.ImagePath); have {
		log.Warnf(ctx, "emoji already exists at storage path: %s", p.emoji.ImagePath)

		// Attempt to remove existing emoji at storage path (might be broken / out-of-date)
		if err := p.mgr.state.Storage.Delete(ctx, p.emoji.ImagePath); err != nil {
			return fmt.Errorf("error removing emoji from storage: %v", err)
		}
	}

	// Write the final image reader stream to our storage.
	sz, err = p.mgr.state.Storage.PutStream(ctx, p.emoji.ImagePath, r)
	if err != nil {
		return fmt.Errorf("error writing emoji to storage: %w", err)
	}

	// Once again check size in case none was provided previously.
	if size := bytesize.Size(sz); size > maxSize {

		if err := p.mgr.state.Storage.Delete(ctx, p.emoji.ImagePath); err != nil {
			log.Errorf(ctx, "error removing too-large-emoji from storage: %v", err)
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
	rc, err := p.mgr.state.Storage.GetStream(ctx, p.emoji.ImagePath)
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
	if have, _ := p.mgr.state.Storage.Has(ctx, p.emoji.ImageStaticPath); have {
		log.Warnf(ctx, "static emoji already exists at storage path: %s", p.emoji.ImagePath)
		// Attempt to remove static existing emoji at storage path (might be broken / out-of-date)
		if err := p.mgr.state.Storage.Delete(ctx, p.emoji.ImageStaticPath); err != nil {
			return fmt.Errorf("error removing static emoji from storage: %v", err)
		}
	}

	// Create an emoji PNG encoder stream.
	enc := staticImg.ToPNG()

	// Stream-encode the PNG static image into storage.
	sz, err := p.mgr.state.Storage.PutStream(ctx, p.emoji.ImageStaticPath, enc)
	if err != nil {
		return fmt.Errorf("error stream-encoding static emoji to storage: %w", err)
	}

	// Set written image size.
	p.emoji.ImageStaticFileSize = int(sz)

	return nil
}
