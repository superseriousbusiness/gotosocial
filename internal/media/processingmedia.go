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
	"context"
	"os"

	errorsv2 "codeberg.org/gruf/go-errors/v2"
	"codeberg.org/gruf/go-runners"

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

// ID returns the ID of the underlying media.
func (p *ProcessingMedia) ID() string {
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
func (p *ProcessingMedia) Load(ctx context.Context) (*gtsmodel.MediaAttachment, error) {
	media, done, err := p.load(ctx)
	if !done {
		// On a context-canceled error (marked as !done), requeue for loading.
		log.Warnf(ctx, "reprocessing media %s after canceled ctx", p.media.ID)
		p.mgr.state.Workers.Dereference.Queue.Push(func(ctx context.Context) {
			if _, _, err := p.load(ctx); err != nil {
				log.Errorf(ctx, "error loading media: %v", err)
			}
		})
	}
	return media, err
}

// load is the package private form of load() that is wrapped to catch context canceled.
func (p *ProcessingMedia) load(ctx context.Context) (
	media *gtsmodel.MediaAttachment,
	done bool,
	err error,
) {
	err = p.proc.Process(func() error {
		if done = p.done; done {
			// Already proc'd.
			return p.err
		}

		defer func() {
			// This is only done when ctx NOT cancelled.
			if done = (err == nil || !errorsv2.IsV2(err,
				context.Canceled,
				context.DeadlineExceeded,
			)); done {
				// Processing finished,
				// whether error or not!

				// Anything from here, we
				// need to ensure happens
				// (i.e. no ctx canceled).
				ctx = context.WithoutCancel(ctx)

				// On error or unknown media types, perform error cleanup.
				if err != nil || p.media.Type == gtsmodel.FileTypeUnknown {
					p.cleanup(ctx)
				}

				// Update with latest details, whatever happened.
				e := p.mgr.state.DB.UpdateAttachment(ctx, p.media)
				if e != nil {
					log.Errorf(ctx, "error updating media in db: %v", e)
				}

				// Store values.
				p.done = true
				p.err = err
			}
		}()

		// Attempt to store media and calculate
		// full-size media attachment details.
		//
		// This will update p.media as it goes.
		err = p.store(ctx)
		return err
	})

	// Return a copy of media attachment.
	media = new(gtsmodel.MediaAttachment)
	*media = *p.media
	return
}

// store calls the data function attached to p if it hasn't been called yet,
// and updates the underlying attachment fields as necessary. It will then stream
// bytes from p's reader directly into storage so that it can be retrieved later.
func (p *ProcessingMedia) store(ctx context.Context) error {
	// Load media from data func.
	rc, err := p.dataFn(ctx)
	if err != nil {
		return gtserror.Newf("error executing data function: %w", err)
	}

	var (
		// predfine temporary media
		// file path variables so we
		// can remove them on error.
		temppath  string
		thumbpath string
	)

	defer func() {
		if err := remove(temppath, thumbpath); err != nil {
			log.Errorf(ctx, "error(s) cleaning up files: %v", err)
		}
	}()

	// Drain reader to tmp file
	// (this reader handles close).
	temppath, err = drainToTmp(rc)
	if err != nil {
		return gtserror.Newf("error draining data to tmp: %w", err)
	}

	// Pass input file through ffprobe to
	// parse further metadata information.
	result, err := probe(ctx, temppath)
	if err != nil && !isUnsupportedTypeErr(err) {
		return gtserror.Newf("ffprobe error: %w", err)
	} else if result == nil {
		log.Warn(ctx, "unsupported data type")
		return nil
	}

	var ext string

	// Extract any video stream metadata from media.
	// This will always be used regardless of type,
	// as even audio files may contain embedded album art.
	width, height, framerate := result.ImageMeta()
	aspect := util.Div(float32(width), float32(height))
	p.media.FileMeta.Original.Width = width
	p.media.FileMeta.Original.Height = height
	p.media.FileMeta.Original.Size = (width * height)
	p.media.FileMeta.Original.Aspect = aspect
	p.media.FileMeta.Original.Framerate = util.PtrIf(framerate)
	p.media.FileMeta.Original.Duration = util.PtrIf(float32(result.duration))
	p.media.FileMeta.Original.Bitrate = util.PtrIf(result.bitrate)

	// Set media type from ffprobe format data.
	p.media.Type, ext = result.GetFileType()

	// Add file extension to path.
	newpath := temppath + "." + ext

	// Before ffmpeg processing, rename to set file ext.
	if err := os.Rename(temppath, newpath); err != nil {
		return gtserror.Newf("error renaming to %s - >%s: %w", temppath, newpath, err)
	}

	// Update path var
	// AFTER successful.
	temppath = newpath

	switch p.media.Type {
	case gtsmodel.FileTypeImage,
		gtsmodel.FileTypeVideo,
		gtsmodel.FileTypeGifv:
		// Attempt to clean as metadata from file as possible.
		if err := clearMetadata(ctx, temppath); err != nil {
			return gtserror.Newf("error cleaning metadata: %w", err)
		}

	case gtsmodel.FileTypeAudio:
		// NOTE: we do not clean audio file
		// metadata, in order to keep tags.

	default:
		log.Warn(ctx, "unsupported data type: %s", result.format)
		return nil
	}

	if width > 0 && height > 0 {
		// Determine thumbnail dimens to use.
		thumbWidth, thumbHeight := thumbSize(
			width,
			height,
			aspect,
		)
		p.media.FileMeta.Small.Width = thumbWidth
		p.media.FileMeta.Small.Height = thumbHeight
		p.media.FileMeta.Small.Size = (thumbWidth * thumbHeight)
		p.media.FileMeta.Small.Aspect = aspect

		// Determine if blurhash needs generating.
		needBlurhash := (p.media.Blurhash == "")
		var newBlurhash string

		// Generate thumbnail, and new blurhash if need from media.
		thumbpath, newBlurhash, err = generateThumb(ctx, temppath,
			thumbWidth,
			thumbHeight,
			result.orientation,
			result.PixFmt(),
			needBlurhash,
		)
		if err != nil {
			return gtserror.Newf("error generating image thumb: %w", err)
		}

		if needBlurhash {
			// Set newly determined blurhash.
			p.media.Blurhash = newBlurhash
		}
	}

	// Calculate final media attachment file path.
	p.media.File.Path = uris.StoragePathForAttachment(
		p.media.AccountID,
		string(TypeAttachment),
		string(SizeOriginal),
		p.media.ID,
		ext,
	)

	// Get mimetype for the file container
	// type, falling back to generic data.
	p.media.File.ContentType = getMimeType(ext)

	// Copy temporary file into storage at path.
	filesz, err := p.mgr.state.Storage.PutFile(ctx,
		p.media.File.Path,
		temppath,
		p.media.File.ContentType,
	)
	if err != nil {
		return gtserror.Newf("error writing media to storage: %w", err)
	}

	// Set final determined file size.
	p.media.File.FileSize = int(filesz)

	if thumbpath != "" {
		// Determine final thumbnail ext.
		thumbExt := getExtension(thumbpath)

		// Calculate final media attachment thumbnail path.
		p.media.Thumbnail.Path = uris.StoragePathForAttachment(
			p.media.AccountID,
			string(TypeAttachment),
			string(SizeSmall),
			p.media.ID,
			thumbExt,
		)

		// Determine thumbnail content-type from thumb ext.
		p.media.Thumbnail.ContentType = getMimeType(thumbExt)

		// Copy thumbnail file into storage at path.
		thumbsz, err := p.mgr.state.Storage.PutFile(ctx,
			p.media.Thumbnail.Path,
			thumbpath,
			p.media.Thumbnail.ContentType,
		)
		if err != nil {
			return gtserror.Newf("error writing thumb to storage: %w", err)
		}

		// Set final determined thumbnail size.
		p.media.Thumbnail.FileSize = int(thumbsz)

		// Generate a media attachment thumbnail URL.
		p.media.Thumbnail.URL = uris.URIForAttachment(
			p.media.AccountID,
			string(TypeAttachment),
			string(SizeSmall),
			p.media.ID,
			thumbExt,
		)
	}

	// Generate a media attachment URL.
	p.media.URL = uris.URIForAttachment(
		p.media.AccountID,
		string(TypeAttachment),
		string(SizeOriginal),
		p.media.ID,
		ext,
	)

	// We can now consider this cached.
	p.media.Cached = util.Ptr(true)

	// Finally set the attachment as finished processing.
	p.media.Processing = gtsmodel.ProcessingStatusProcessed

	return nil
}

// cleanup will remove any traces of processing media from storage.
// and perform any other necessary cleanup steps after failure.
func (p *ProcessingMedia) cleanup(ctx context.Context) {
	if p.media.File.Path != "" {
		// Ensure media file at path is deleted from storage.
		err := p.mgr.state.Storage.Delete(ctx, p.media.File.Path)
		if err != nil && !storage.IsNotFound(err) {
			log.Errorf(ctx, "error deleting %s: %v", p.media.File.Path, err)
		}
	}

	if p.media.Thumbnail.Path != "" {
		// Ensure media thumbnail at path is deleted from storage.
		err := p.mgr.state.Storage.Delete(ctx, p.media.Thumbnail.Path)
		if err != nil && !storage.IsNotFound(err) {
			log.Errorf(ctx, "error deleting %s: %v", p.media.Thumbnail.Path, err)
		}
	}

	// Unset all processor-calculated media fields.
	p.media.FileMeta.Original = gtsmodel.Original{}
	p.media.FileMeta.Small = gtsmodel.Small{}
	p.media.File.ContentType = ""
	p.media.File.FileSize = 0
	p.media.File.Path = ""
	p.media.Thumbnail.FileSize = 0
	p.media.Thumbnail.ContentType = ""
	p.media.Thumbnail.Path = ""
	p.media.Thumbnail.URL = ""
	p.media.URL = ""

	// Also ensure marked as unknown and finished
	// processing so gets inserted as placeholder URL.
	p.media.Processing = gtsmodel.ProcessingStatusProcessed
	p.media.Type = gtsmodel.FileTypeUnknown
	p.media.Cached = util.Ptr(false)
}
