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

	errorsv2 "codeberg.org/gruf/go-errors/v2"
	"codeberg.org/gruf/go-runners"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// ProcessingEmoji represents an emoji currently processing. It exposes
// various functions for retrieving data from the process.
type ProcessingEmoji struct {
	emoji     *gtsmodel.Emoji   // processing emoji details
	newPathID string            // new emoji path ID to use when being refreshed
	dataFn    DataFunc          // load-data function, returns media stream
	done      bool              // done is set when process finishes with non ctx canceled type error
	proc      runners.Processor // proc helps synchronize only a singular running processing instance
	err       error             // error stores permanent error value when done
	mgr       *Manager          // mgr instance (access to db / storage)
}

// ID returns the ID of the underlying emoji.
func (p *ProcessingEmoji) ID() string {
	return p.emoji.ID // immutable, safe outside mutex.
}

// LoadEmoji blocks until the static and fullsize image has been processed, and then returns the completed emoji.
func (p *ProcessingEmoji) Load(ctx context.Context) (*gtsmodel.Emoji, error) {
	emoji, done, err := p.load(ctx)
	if !done {
		// On a context-canceled error (marked as !done), requeue for loading.
		p.mgr.state.Workers.Dereference.Queue.Push(func(ctx context.Context) {
			if _, _, err := p.load(ctx); err != nil {
				log.Errorf(ctx, "error loading emoji: %v", err)
			}
		})
	}
	return emoji, err
}

// load is the package private form of load() that is wrapped to catch context canceled.
func (p *ProcessingEmoji) load(ctx context.Context) (
	emoji *gtsmodel.Emoji,
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
			done = (err == nil || !errorsv2.IsV2(err,
				context.Canceled,
				context.DeadlineExceeded,
			))

			if !done {
				return
			}

			// Anything from here, we
			// need to ensure happens
			// (i.e. no ctx canceled).
			ctx = gtscontext.WithValues(
				context.Background(),
				ctx, // values
			)

			// On error, clean
			// downloaded files.
			if err != nil {
				p.cleanup(ctx)
			}

			if !done {
				return
			}

			// Update with latest details, whatever happened.
			e := p.mgr.state.DB.UpdateEmoji(ctx, p.emoji)
			if e != nil {
				log.Errorf(ctx, "error updating emoji in db: %v", e)
			}

			// Store final values.
			p.done = true
			p.err = err
		}()

		// Attempt to store media and calculate
		// full-size media attachment details.
		//
		// This will update p.emoji as it goes.
		err = p.store(ctx)
		return err
	})
	emoji = p.emoji
	return
}

// store calls the data function attached to p if it hasn't been called yet,
// and updates the underlying attachment fields as necessary. It will then stream
// bytes from p's reader directly into storage so that it can be retrieved later.
func (p *ProcessingEmoji) store(ctx context.Context) error {
	// Load media from data func.
	rc, err := p.dataFn(ctx)
	if err != nil {
		return gtserror.Newf("error executing data function: %w", err)
	}

	var (
		// predfine temporary media
		// file path variables so we
		// can remove them on error.
		temppath   string
		staticpath string
	)

	defer func() {
		if err := remove(temppath, staticpath); err != nil {
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
	result, err := ffprobe(ctx, temppath)
	if err != nil {
		return gtserror.Newf("error ffprobing data: %w", err)
	}

	switch {
	// No errors parsing data.
	case result.Error == nil:

	// Data type unhandleable by ffprobe.
	case result.Error.Code == -1094995529:
		log.Warn(ctx, "unsupported data type")
		return nil

	default:
		return gtserror.Newf("ffprobe error: %w", err)
	}

	var ext string

	// Set media type from ffprobe format data.
	fileType, ext := result.Format.GetFileType()
	if fileType != gtsmodel.FileTypeImage {
		return gtserror.Newf("unsupported emoji filetype: %s (%s)", fileType, ext)
	}

	// Generate a static image from input emoji path.
	staticpath, err = ffmpegGenerateStatic(ctx, temppath)
	if err != nil {
		return gtserror.Newf("error generating emoji static: %w", err)
	}

	var pathID string
	if p.newPathID != "" {
		// This is a refreshed emoji with a new
		// path ID that this will be stored under.
		pathID = p.newPathID
	} else {
		// This is a new emoji, simply use provided ID.
		pathID = p.emoji.ID
	}

	// Determine instance account ID from generated image static path.
	instanceAccID, ok := getInstanceAccountID(p.emoji.ImageStaticPath)
	if !ok {
		return gtserror.Newf("invalid emoji static path; no instance account id: %s", p.emoji.ImageStaticPath)
	}

	// Calculate final media attachment file path.
	p.emoji.ImagePath = uris.StoragePathForAttachment(
		instanceAccID,
		string(TypeEmoji),
		string(SizeOriginal),
		pathID,
		ext,
	)

	// Copy temporary file into storage at path.
	filesz, err := p.mgr.state.Storage.PutFile(ctx,
		p.emoji.ImagePath,
		temppath,
	)
	if err != nil {
		return gtserror.Newf("error writing emoji to storage: %w", err)
	}

	// Copy static emoji file into storage at path.
	staticsz, err := p.mgr.state.Storage.PutFile(ctx,
		p.emoji.ImageStaticPath,
		staticpath,
	)
	if err != nil {
		return gtserror.Newf("error writing static to storage: %w", err)
	}

	// Set final determined file sizes.
	p.emoji.ImageFileSize = int(filesz)
	p.emoji.ImageStaticFileSize = int(staticsz)

	// Fill in remaining emoji data now it's stored.
	p.emoji.ImageURL = uris.URIForAttachment(
		instanceAccID,
		string(TypeEmoji),
		string(SizeOriginal),
		pathID,
		ext,
	)

	// Get mimetype for the file container
	// type, falling back to generic data.
	p.emoji.ImageContentType = getMimeType(ext)

	// We can now consider this cached.
	p.emoji.Cached = util.Ptr(true)

	return nil
}

// cleanup will remove any traces of processing emoji from storage,
// and perform any other necessary cleanup steps after failure.
func (p *ProcessingEmoji) cleanup(ctx context.Context) {
	var err error

	if p.emoji.ImagePath != "" {
		// Ensure emoji file at path is deleted from storage.
		err = p.mgr.state.Storage.Delete(ctx, p.emoji.ImagePath)
		if err != nil && !storage.IsNotFound(err) {
			log.Errorf(ctx, "error deleting %s: %v", p.emoji.ImagePath, err)
		}
	}

	if p.emoji.ImageStaticPath != "" {
		// Ensure emoji static file at path is deleted from storage.
		err = p.mgr.state.Storage.Delete(ctx, p.emoji.ImageStaticPath)
		if err != nil && !storage.IsNotFound(err) {
			log.Errorf(ctx, "error deleting %s: %v", p.emoji.ImageStaticPath, err)
		}
	}

	// Ensure marked as not cached.
	p.emoji.Cached = util.Ptr(false)
}

// getInstanceAccountID determines the instance account ID from
// emoji static image storage path. returns false on failure.
func getInstanceAccountID(staticPath string) (string, bool) {
	matches := regexes.FilePath.FindStringSubmatch(staticPath)
	if len(matches) < 2 {
		return "", false
	}
	return matches[1], true
}
