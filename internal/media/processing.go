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
	"sync"
	"time"

	"codeberg.org/gruf/go-store/kv"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type processState int

const (
	received processState = iota // processing order has been received but not done yet
	complete                     // processing order has been completed successfully
	errored                      // processing order has been completed with an error
)

// Processing represents a piece of media that is currently being processed. It exposes
// various functions for retrieving data from the process.
type Processing struct {
	mu sync.Mutex

	/*
		below fields should be set on newly created media;
		attachment will be updated incrementally as media goes through processing
	*/

	attachment *gtsmodel.MediaAttachment // will only be set if the media is an attachment
	emoji      *gtsmodel.Emoji           // will only be set if the media is an emoji

	rawData []byte

	/*
		below fields represent the processing state of the media thumbnail
	*/

	thumbstate processState
	thumb      *ImageMeta

	/*
		below fields represent the processing state of the full-sized media
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

func (p *Processing) Thumb(ctx context.Context) (*ImageMeta, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.thumbstate {
	case received:
		// we haven't processed a thumbnail for this media yet so do it now

		// check if we need to create a blurhash or if there's already one set
		var createBlurhash bool
		if p.attachment.Blurhash == "" {
			// no blurhash created yet
			createBlurhash = true
		}

		thumb, err := deriveThumbnail(p.rawData, p.attachment.File.ContentType, createBlurhash)
		if err != nil {
			p.err = fmt.Errorf("error deriving thumbnail: %s", err)
			p.thumbstate = errored
			return nil, p.err
		}

		// put the thumbnail in storage
		if err := p.storage.Put(p.attachment.Thumbnail.Path, thumb.image); err != nil {
			p.err = fmt.Errorf("error storing thumbnail: %s", err)
			p.thumbstate = errored
			return nil, p.err
		}

		// set appropriate fields on the attachment based on the thumbnail we derived
		if createBlurhash {
			p.attachment.Blurhash = thumb.blurhash
		}

		p.attachment.FileMeta.Small = gtsmodel.Small{
			Width:  thumb.width,
			Height: thumb.height,
			Size:   thumb.size,
			Aspect: thumb.aspect,
		}
		p.attachment.Thumbnail.FileSize = thumb.size

		if err := putOrUpdateAttachment(ctx, p.database, p.attachment); err != nil {
			p.err = err
			p.thumbstate = errored
			return nil, err
		}

		// set the thumbnail of this media
		p.thumb = thumb

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

func (p *Processing) FullSize(ctx context.Context) (*ImageMeta, error) {
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

// AttachmentID returns the ID of the underlying media attachment without blocking processing.
func (p *Processing) AttachmentID() string {
	return p.attachment.ID
}

// Load blocks until the thumbnail and fullsize content has been processed, and then
// returns the completed attachment.
func (p *Processing) Load(ctx context.Context) (*gtsmodel.MediaAttachment, error) {
	if _, err := p.Thumb(ctx); err != nil {
		return nil, err
	}

	if _, err := p.FullSize(ctx); err != nil {
		return nil, err
	}

	return p.attachment, nil
}

func (p *Processing) LoadEmoji(ctx context.Context) (*gtsmodel.Emoji, error) {
	return nil, nil
}

func (p *Processing) Finished() bool {
	return p.thumbstate == complete && p.fullSizeState == complete
}

// putOrUpdateAttachment is just a convenience function for first trying to PUT the attachment in the database,
// and then if that doesn't work because the attachment already exists, updating it instead.
func putOrUpdateAttachment(ctx context.Context, database db.DB, attachment *gtsmodel.MediaAttachment) error {
	if err := database.Put(ctx, attachment); err != nil {
		if err != db.ErrAlreadyExists {
			return fmt.Errorf("putOrUpdateAttachment: proper error while putting attachment: %s", err)
		}
		if err := database.UpdateByPrimaryKey(ctx, attachment); err != nil {
			return fmt.Errorf("putOrUpdateAttachment: error while updating attachment: %s", err)
		}
	}

	return nil
}
