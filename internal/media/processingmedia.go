// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package media

import (
	"bytes"
	"context"
	"image/jpeg"
	"io"
	"time"

	errorsv2 "codeberg.org/gruf/go-errors/v2"
	"codeberg.org/gruf/go-runners"
	terminator "codeberg.org/superseriousbusiness/exif-terminator"
	"github.com/disintegration/imaging"
	"github.com/h2non/filetype"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// ProcessingMedia represents a piece of media
// currently being processed. It exposes functions
// for retrieving data from the process.
type ProcessingMedia struct {
	media  *gtsmodel.MediaAttachment // processing media attachment details
	dataFn DataFunc                  // load-data function, returns media stream
	done   bool                      // done is set when process finishes with non ctx canceled type error
	proc   runners.Processor         // proc helps synchronize only a singular running processing instance
	err    error                     // error stores permanent error value when done
	mgr    *Manager                  // mgr instance (access to db / storage)
}

// AttachmentID returns the ID of the underlying
// media attachment without blocking processing.
func (p *ProcessingMedia) AttachmentID() string {
	return p.media.ID // immutable, safe outside mutex.
}

// LoadAttachment blocks until the thumbnail and
// fullsize content has been processed, and then
// returns the attachment.
//
// If processing could not be completed fully
// then an error will be returned. The attachment
// will still be returned in that case, but it will
// only be partially complete and should be treated
// as a placeholder.
func (p *ProcessingMedia) LoadAttachment(ctx context.Context) (*gtsmodel.MediaAttachment, error) {
	// Attempt to load synchronously.
	media, done, err := p.load(ctx)
	if err == nil {
		// No issue, return media.
		return media, nil
	}

	if !done {
		// Provided context was cancelled,
		// e.g. request aborted early before
		// its context could be used to finish
		// loading the attachment. Enqueue for
		// asynchronous processing, which will
		// use a background context.
		log.Warnf(ctx, "reprocessing media %s after canceled ctx", p.media.ID)
		p.mgr.state.Workers.Media.Queue.Push(p.Process)
	}

	// Media could not be retrieved FULLY,
	// but partial attachment should be present.
	return media, err
}

// Process allows the receiving object to fit the
// runners.WorkerFunc signature. It performs a
// (blocking) load and logs on error.
func (p *ProcessingMedia) Process(ctx context.Context) {
	if _, _, err := p.load(ctx); err != nil {
		log.Errorf(ctx, "error(s) processing media: %v", err)
	}
}

// load performs a concurrency-safe load of ProcessingMedia, only
// marking itself as complete when returned error != context.Canceled.
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
			done = err == nil || !errorsv2.IsV2(err,
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

		var errs gtserror.MultiError

		// Attempt to store media and calculate
		// full-size media attachment details.
		//
		// This will update p.media as it goes.
		storeErr := p.store(ctx)
		if storeErr != nil {
			errs.Append(storeErr)
		}

		// Finish processing by reloading media into
		// memory to get dimension and generate a thumb.
		//
		// This will update p.media as it goes.
		if finishErr := p.finish(ctx); finishErr != nil {
			errs.Append(finishErr)
		}

		// If this isn't a file we were able to process,
		// we may have partially stored it (eg., it's a
		// jpeg, which is fine, but streaming it to storage
		// was interrupted halfway through and so it was
		// never decoded). Try to clean up in this case.
		if p.media.Type == gtsmodel.FileTypeUnknown {
			deleteErr := p.mgr.state.Storage.Delete(ctx, p.media.File.Path)
			if deleteErr != nil && !storage.IsNotFound(deleteErr) {
				errs.Append(deleteErr)
			}
		}

		if len(errs) == 0 {
			// Update attachment with latest details now its cached.
			dbErr := p.mgr.state.DB.UpdateAttachment(ctx, p.media)
			if dbErr != nil {
				errs.Append(dbErr)
			}
		}

		// Combine all errors.
		err = errs.Combine()
		return err
	})

	return p.media, done, err
}

// store calls the data function attached to p if it hasn't been called yet,
// and updates the underlying attachment fields as necessary. It will then stream
// bytes from p's reader directly into storage so that it can be retrieved later.
func (p *ProcessingMedia) store(ctx context.Context) error {
	// Load media from provided data fun
	rc, sz, err := p.dataFn(ctx)
	if err != nil {
		return gtserror.Newf("error executing data function: %w", err)
	}

	defer func() {
		// Ensure data reader gets closed on return.
		if err := rc.Close(); err != nil {
			log.Errorf(ctx, "error closing data reader: %v", err)
		}
	}()

	// Assume we're given correct file
	// size, we can overwrite this later
	// once we know THE TRUTH.
	fileSize := int(sz)
	p.media.File.FileSize = fileSize

	// Prepare to read bytes from
	// file header or magic number.
	hdrBuf := newHdrBuf(fileSize)

	// Read into buffer as much as possible.
	//
	// UnexpectedEOF means we couldn't read up to the
	// given size, but we may still have read something.
	//
	// EOF means we couldn't read anything at all.
	//
	// Any other error likely means the connection messed up.
	//
	// In other words, rather counterintuitively, we
	// can only proceed on no error or unexpected error!
	n, err := io.ReadFull(rc, hdrBuf)
	if err != nil {
		if err != io.ErrUnexpectedEOF {
			return gtserror.Newf("error reading first bytes of incoming media: %w", err)
		}

		// Initial file size was misreported, so we didn't read
		// fully into hdrBuf. Reslice it to the size we did read.
		log.Warnf(ctx,
			"recovered from misreported file size; reported %d; read %d",
			fileSize, n,
		)
		hdrBuf = hdrBuf[:n]
		fileSize = n
		p.media.File.FileSize = fileSize
	}

	// Parse file type info from header buffer.
	// This should only ever error if the buffer
	// is empty (ie., the attachment is 0 bytes).
	info, err := filetype.Match(hdrBuf)
	if err != nil {
		return gtserror.Newf("error parsing file type: %w", err)
	}

	// Recombine header bytes with remaining stream
	r := io.MultiReader(bytes.NewReader(hdrBuf), rc)

	// Assume we'll put
	// this file in storage.
	store := true

	switch info.Extension {
	case "mp4":
		// No problem.

	case "gif":
		// No problem

	case "jpg", "jpeg", "png", "webp":
		if fileSize > 0 {
			// A file size was provided so we can clean
			// exif data from image as we're streaming it.
			r, err = terminator.Terminate(r, fileSize, info.Extension)
			if err != nil {
				return gtserror.Newf("error cleaning exif data: %w", err)
			}
		}

	default:
		// The file is not a supported format that
		// we can process, so we can't do much with it.
		log.Warnf(ctx,
			"media extension '%s' not officially supported, will be processed as "+
				"type '%s' with minimal metadata, and will not be cached locally",
			info.Extension, gtsmodel.FileTypeUnknown,
		)

		// Don't bother storing this.
		store = false
	}

	// Fill in correct attachment
	// data now we've parsed it.
	p.media.URL = uris.URIForAttachment(
		p.media.AccountID,
		string(TypeAttachment),
		string(SizeOriginal),
		p.media.ID,
		info.Extension,
	)

	// Prefer discovered mime type, fall back to
	// generic "this contains some bytes" type.
	mime := info.MIME.Value
	if mime == "" {
		mime = "application/octet-stream"
	}
	p.media.File.ContentType = mime

	// Calculate final media attachment file path.
	p.media.File.Path = uris.StoragePathForAttachment(
		p.media.AccountID,
		string(TypeAttachment),
		string(SizeOriginal),
		p.media.ID,
		info.Extension,
	)

	// We should only try to store the file if it's
	// a format we can keep processing, otherwise be
	// a bit cheeky: don't store it and let users
	// click through to the remote server instead.
	if !store {
		return nil
	}

	// File shouldn't already exist in storage at this point,
	// but we do a check as it's worth logging / cleaning up.
	if have, _ := p.mgr.state.Storage.Has(ctx, p.media.File.Path); have {
		log.Warnf(ctx, "media already exists at storage path: %s", p.media.File.Path)

		// Attempt to remove existing media at storage path (might be broken / out-of-date)
		if err := p.mgr.state.Storage.Delete(ctx, p.media.File.Path); err != nil {
			return gtserror.Newf("error removing media from storage: %v", err)
		}
	}

	// Write the final reader stream to our storage.
	wroteSize, err := p.mgr.state.Storage.PutStream(ctx, p.media.File.Path, r)
	if err != nil {
		return gtserror.Newf("error writing media to storage: %w", err)
	}

	// Set actual written size
	// as authoritative file size.
	p.media.File.FileSize = int(wroteSize)

	// We can now consider this cached.
	p.media.Cached = util.Ptr(true)

	return nil
}

func (p *ProcessingMedia) finish(ctx context.Context) error {
	// Make a jolly assumption about thumbnail type.
	p.media.Thumbnail.ContentType = mimeImageJpeg

	// Calculate attachment thumbnail file path
	p.media.Thumbnail.Path = uris.StoragePathForAttachment(
		p.media.AccountID,
		string(TypeAttachment),
		string(SizeSmall),
		p.media.ID,

		// Always encode attachment
		// thumbnails as jpg.
		"jpg",
	)

	// Calculate attachment thumbnail serve path.
	p.media.Thumbnail.URL = uris.URIForAttachment(
		p.media.AccountID,
		string(TypeAttachment),
		string(SizeSmall),
		p.media.ID,

		// Always encode attachment
		// thumbnails as jpg.
		"jpg",
	)

	// If original file hasn't been stored, there's
	// likely something wrong with the data, or we
	// don't want to store it. Skip everything else.
	if !*p.media.Cached {
		p.media.Processing = gtsmodel.ProcessingStatusProcessed
		return nil
	}

	// Get a stream to the original file for further processing.
	rc, err := p.mgr.state.Storage.GetStream(ctx, p.media.File.Path)
	if err != nil {
		return gtserror.Newf("error loading file from storage: %w", err)
	}
	defer rc.Close()

	// fullImg is the processed version of
	// the original (stripped + reoriented).
	var fullImg *gtsImage

	// Depending on the content type, we
	// can do various types of decoding.
	switch p.media.File.ContentType {

	// .jpeg, .gif, .webp image type
	case mimeImageJpeg, mimeImageGif, mimeImageWebp:
		fullImg, err = decodeImage(rc,
			imaging.AutoOrientation(true),
		)
		if err != nil {
			return gtserror.Newf("error decoding image: %w", err)
		}

		// Mark as no longer unknown type now
		// we know for sure we can decode it.
		p.media.Type = gtsmodel.FileTypeImage

	// .png image (requires ancillary chunk stripping)
	case mimeImagePng:
		fullImg, err = decodeImage(
			&pngAncillaryChunkStripper{Reader: rc},
			imaging.AutoOrientation(true),
		)
		if err != nil {
			return gtserror.Newf("error decoding image: %w", err)
		}

		// Mark as no longer unknown type now
		// we know for sure we can decode it.
		p.media.Type = gtsmodel.FileTypeImage

	// .mp4 video type
	case mimeVideoMp4:
		video, err := decodeVideoFrame(rc)
		if err != nil {
			return gtserror.Newf("error decoding video: %w", err)
		}

		// Set video frame as image.
		fullImg = video.frame

		// Set video metadata in attachment info.
		p.media.FileMeta.Original.Duration = &video.duration
		p.media.FileMeta.Original.Framerate = &video.framerate
		p.media.FileMeta.Original.Bitrate = &video.bitrate

		// Mark as no longer unknown type now
		// we know for sure we can decode it.
		p.media.Type = gtsmodel.FileTypeVideo
	}

	// fullImg should be in-memory by
	// now so we're done with storage.
	if err := rc.Close(); err != nil {
		return gtserror.Newf("error closing file: %w", err)
	}

	// Set full-size dimensions in attachment info.
	p.media.FileMeta.Original.Width = fullImg.Width()
	p.media.FileMeta.Original.Height = fullImg.Height()
	p.media.FileMeta.Original.Size = fullImg.Size()
	p.media.FileMeta.Original.Aspect = fullImg.AspectRatio()

	// Get smaller thumbnail image
	thumbImg := fullImg.Thumbnail()

	// Garbage collector, you may
	// now take our large son.
	fullImg = nil

	// Only generate blurhash
	// from thumb if necessary.
	if p.media.Blurhash == "" {
		hash, err := thumbImg.Blurhash()
		if err != nil {
			return gtserror.Newf("error generating blurhash: %w", err)
		}

		// Set the attachment blurhash.
		p.media.Blurhash = hash
	}

	// Thumbnail shouldn't already exist in storage at this point,
	// but we do a check as it's worth logging / cleaning up.
	if have, _ := p.mgr.state.Storage.Has(ctx, p.media.Thumbnail.Path); have {
		log.Warnf(ctx, "thumbnail already exists at storage path: %s", p.media.Thumbnail.Path)

		// Attempt to remove existing thumbnail at storage path (might be broken / out-of-date)
		if err := p.mgr.state.Storage.Delete(ctx, p.media.Thumbnail.Path); err != nil {
			return gtserror.Newf("error removing thumbnail from storage: %v", err)
		}
	}

	// Create a thumbnail JPEG encoder stream.
	enc := thumbImg.ToJPEG(&jpeg.Options{
		// Good enough for
		// a thumbnail.
		Quality: 70,
	})

	// Stream-encode the JPEG thumbnail image into storage.
	sz, err := p.mgr.state.Storage.PutStream(ctx, p.media.Thumbnail.Path, enc)
	if err != nil {
		return gtserror.Newf("error stream-encoding thumbnail to storage: %w", err)
	}

	// Set thumbnail dimensions in attachment info.
	p.media.FileMeta.Small = gtsmodel.Small{
		Width:  thumbImg.Width(),
		Height: thumbImg.Height(),
		Size:   thumbImg.Size(),
		Aspect: thumbImg.AspectRatio(),
	}

	// Set written image size.
	p.media.Thumbnail.FileSize = int(sz)

	// Finally set the attachment as processed and update time.
	p.media.Processing = gtsmodel.ProcessingStatusProcessed
	p.media.File.UpdatedAt = time.Now()

	return nil
}
