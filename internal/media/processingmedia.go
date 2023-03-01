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
	"image/jpeg"
	"io"
	"time"

	"codeberg.org/gruf/go-errors/v2"
	"codeberg.org/gruf/go-runners"
	"github.com/disintegration/imaging"
	"github.com/h2non/filetype"
	terminator "github.com/superseriousbusiness/exif-terminator"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// ProcessingMedia represents a piece of media that is currently being processed. It exposes
// various functions for retrieving data from the process.
type ProcessingMedia struct {
	media   *gtsmodel.MediaAttachment // processing media attachment details
	dataFn  DataFunc                  // load-data function, returns media stream
	postFn  PostDataCallbackFunc      // post data callback function
	recache bool                      // recaching existing (uncached) media
	done    bool                      // done is set when process finishes with non ctx canceled type error
	proc    runners.Processor         // proc helps synchronize only a singular running processing instance
	err     error                     // error stores permanent error value when done
	mgr     *manager                  // mgr instance (access to db / storage)
}

// AttachmentID returns the ID of the underlying media attachment without blocking processing.
func (p *ProcessingMedia) AttachmentID() string {
	return p.media.ID // immutable, safe outside mutex.
}

// LoadAttachment blocks until the thumbnail and fullsize content has been processed, and then returns the completed attachment.
func (p *ProcessingMedia) LoadAttachment(ctx context.Context) (*gtsmodel.MediaAttachment, error) {
	// Attempt to load synchronously.
	media, done, err := p.load(ctx)

	if err == nil {
		// No issue, return media.
		return media, nil
	}

	if !done {
		// Provided context was cancelled, e.g. request cancelled
		// early. Queue this item for asynchronous processing.
		log.Warnf(ctx, "reprocessing media %s after canceled ctx", p.media.ID)
		go p.mgr.state.Workers.Media.Enqueue(p.Process)
	}

	return nil, err
}

// Process allows the receiving object to fit the runners.WorkerFunc signature. It performs a (blocking) load and logs on error.
func (p *ProcessingMedia) Process(ctx context.Context) {
	if _, _, err := p.load(ctx); err != nil {
		log.Errorf(ctx, "error processing media: %v", err)
	}
}

// load performs a concurrency-safe load of ProcessingMedia, only marking itself as complete when returned error is NOT a context cancel.
func (p *ProcessingMedia) load(ctx context.Context) (*gtsmodel.MediaAttachment, bool, error) {
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

		if p.recache {
			// Existing attachment we're recaching, so only need to update.
			err = p.mgr.state.DB.UpdateByID(ctx, p.media, p.media.ID)
			return err
		}

		// New attachment, first time caching.
		err = p.mgr.state.DB.Put(ctx, p.media)
		return err
	})

	if err != nil {
		return nil, done, err
	}

	return p.media, done, nil
}

// store calls the data function attached to p if it hasn't been called yet,
// and updates the underlying attachment fields as necessary. It will then stream
// bytes from p's reader directly into storage so that it can be retrieved later.
func (p *ProcessingMedia) store(ctx context.Context) error {
	defer func() {
		if p.postFn == nil {
			return
		}

		// ensure post callback gets called.
		if err := p.postFn(ctx); err != nil {
			log.Errorf(ctx, "error executing postdata function: %v", err)
		}
	}()

	// Load media from provided data fun
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

	// Recombine header bytes with remaining stream
	r := io.MultiReader(bytes.NewReader(hdrBuf), rc)

	switch info.Extension {
	case "mp4":
		p.media.Type = gtsmodel.FileTypeVideo

	case "gif":
		p.media.Type = gtsmodel.FileTypeImage

	case "jpg", "jpeg", "png", "webp":
		p.media.Type = gtsmodel.FileTypeImage
		if sz > 0 {
			// A file size was provided so we can clean exif data from image.
			r, err = terminator.Terminate(r, int(sz), info.Extension)
			if err != nil {
				return fmt.Errorf("error cleaning exif data: %w", err)
			}
		}

	default:
		return fmt.Errorf("unsupported file type: %s", info.Extension)
	}

	// Calculate attachment file path.
	p.media.File.Path = fmt.Sprintf(
		"%s/%s/%s/%s.%s",
		p.media.AccountID,
		TypeAttachment,
		SizeOriginal,
		p.media.ID,
		info.Extension,
	)

	// This shouldn't already exist, but we do a check as it's worth logging.
	if have, _ := p.mgr.state.Storage.Has(ctx, p.media.File.Path); have {
		log.Warnf(ctx, "media already exists at storage path: %s", p.media.File.Path)

		// Attempt to remove existing media at storage path (might be broken / out-of-date)
		if err := p.mgr.state.Storage.Delete(ctx, p.media.File.Path); err != nil {
			return fmt.Errorf("error removing media from storage: %v", err)
		}
	}

	// Write the final image reader stream to our storage.
	sz, err = p.mgr.state.Storage.PutStream(ctx, p.media.File.Path, r)
	if err != nil {
		return fmt.Errorf("error writing media to storage: %w", err)
	}

	// Set written image size.
	p.media.File.FileSize = int(sz)

	// Fill in remaining attachment data now it's stored.
	p.media.URL = uris.GenerateURIForAttachment(
		p.media.AccountID,
		string(TypeAttachment),
		string(SizeOriginal),
		p.media.ID,
		info.Extension,
	)
	p.media.File.ContentType = info.MIME.Value
	cached := true
	p.media.Cached = &cached

	return nil
}

func (p *ProcessingMedia) finish(ctx context.Context) error {
	// Fetch a stream to the original file in storage.
	rc, err := p.mgr.state.Storage.GetStream(ctx, p.media.File.Path)
	if err != nil {
		return fmt.Errorf("error loading file from storage: %w", err)
	}
	defer rc.Close()

	var fullImg *gtsImage

	switch p.media.File.ContentType {
	// .jpeg, .gif, .webp image type
	case mimeImageJpeg, mimeImageGif, mimeImageWebp:
		fullImg, err = decodeImage(rc, imaging.AutoOrientation(true))
		if err != nil {
			return fmt.Errorf("error decoding image: %w", err)
		}

	// .png image (requires ancillary chunk stripping)
	case mimeImagePng:
		fullImg, err = decodeImage(&pngAncillaryChunkStripper{
			Reader: rc,
		}, imaging.AutoOrientation(true))
		if err != nil {
			return fmt.Errorf("error decoding image: %w", err)
		}

	// .mp4 video type
	case mimeVideoMp4:
		video, err := decodeVideoFrame(rc)
		if err != nil {
			return fmt.Errorf("error decoding video: %w", err)
		}

		// Set video frame as image.
		fullImg = video.frame

		// Set video metadata in attachment info.
		p.media.FileMeta.Original.Duration = &video.duration
		p.media.FileMeta.Original.Framerate = &video.framerate
		p.media.FileMeta.Original.Bitrate = &video.bitrate
	}

	// The image should be in-memory by now.
	if err := rc.Close(); err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}

	// Set full-size dimensions in attachment info.
	p.media.FileMeta.Original.Width = int(fullImg.Width())
	p.media.FileMeta.Original.Height = int(fullImg.Height())
	p.media.FileMeta.Original.Size = int(fullImg.Size())
	p.media.FileMeta.Original.Aspect = fullImg.AspectRatio()

	// Calculate attachment thumbnail file path
	p.media.Thumbnail.Path = fmt.Sprintf(
		"%s/%s/%s/%s.jpg",
		p.media.AccountID,
		TypeAttachment,
		SizeSmall,
		p.media.ID,
	)

	// Get smaller thumbnail image
	thumbImg := fullImg.Thumbnail()

	// Garbage collector, you may
	// now take our large son.
	fullImg = nil

	// Blurhash needs generating from thumb.
	hash, err := thumbImg.Blurhash()
	if err != nil {
		return fmt.Errorf("error generating blurhash: %w", err)
	}

	// Set the attachment blurhash.
	p.media.Blurhash = hash

	// This shouldn't already exist, but we do a check as it's worth logging.
	if have, _ := p.mgr.state.Storage.Has(ctx, p.media.Thumbnail.Path); have {
		log.Warnf(ctx, "thumbnail already exists at storage path: %s", p.media.Thumbnail.Path)

		// Attempt to remove existing thumbnail at storage path (might be broken / out-of-date)
		if err := p.mgr.state.Storage.Delete(ctx, p.media.Thumbnail.Path); err != nil {
			return fmt.Errorf("error removing thumbnail from storage: %v", err)
		}
	}

	// Create a thumbnail JPEG encoder stream.
	enc := thumbImg.ToJPEG(&jpeg.Options{
		Quality: 70, // enough for a thumbnail.
	})

	// Stream-encode the JPEG thumbnail image into storage.
	sz, err := p.mgr.state.Storage.PutStream(ctx, p.media.Thumbnail.Path, enc)
	if err != nil {
		return fmt.Errorf("error stream-encoding thumbnail to storage: %w", err)
	}

	// Fill in remaining thumbnail now it's stored
	p.media.Thumbnail.ContentType = mimeImageJpeg
	p.media.Thumbnail.URL = uris.GenerateURIForAttachment(
		p.media.AccountID,
		string(TypeAttachment),
		string(SizeSmall),
		p.media.ID,
		"jpg", // always jpeg
	)

	// Set thumbnail dimensions in attachment info.
	p.media.FileMeta.Small = gtsmodel.Small{
		Width:  int(thumbImg.Width()),
		Height: int(thumbImg.Height()),
		Size:   int(thumbImg.Size()),
		Aspect: thumbImg.AspectRatio(),
	}

	// Set written image size.
	p.media.Thumbnail.FileSize = int(sz)

	// Finally set the attachment as processed and update time.
	p.media.Processing = gtsmodel.ProcessingStatusProcessed
	p.media.File.UpdatedAt = time.Now()

	return nil
}
