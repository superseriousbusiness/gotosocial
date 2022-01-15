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

// ProcessingMedia represents a piece of media that is currently being processed. It exposes
// various functions for retrieving data from the process.
type ProcessingMedia struct {
	mu sync.Mutex

	/*
		below fields should be set on newly created media;
		attachment will be updated incrementally as media goes through processing
	*/

	attachment *gtsmodel.MediaAttachment
	data       DataFunc

	rawData []byte // will be set once the fetchRawData function has been called

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

	// track whether this media has already been put in the databse
	insertedInDB bool
}

// AttachmentID returns the ID of the underlying media attachment without blocking processing.
func (p *ProcessingMedia) AttachmentID() string {
	return p.attachment.ID
}

// LoadAttachment blocks until the thumbnail and fullsize content
// has been processed, and then returns the completed attachment.
func (p *ProcessingMedia) LoadAttachment(ctx context.Context) (*gtsmodel.MediaAttachment, error) {
	if err := p.fetchRawData(ctx); err != nil {
		return nil, err
	}

	if _, err := p.loadThumb(ctx); err != nil {
		return nil, err
	}

	if _, err := p.loadFullSize(ctx); err != nil {
		return nil, err
	}

	// store the result in the database before returning it
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.insertedInDB {
		if err := p.database.Put(ctx, p.attachment); err != nil {
			return nil, err
		}
		p.insertedInDB = true
	}

	return p.attachment, nil
}

// Finished returns true if processing has finished for both the thumbnail
// and full fized version of this piece of media.
func (p *ProcessingMedia) Finished() bool {
	return p.thumbstate == complete && p.fullSizeState == complete
}

func (p *ProcessingMedia) loadThumb(ctx context.Context) (*ImageMeta, error) {
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
		p.attachment.Thumbnail.FileSize = len(thumb.image)

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

func (p *ProcessingMedia) loadFullSize(ctx context.Context) (*ImageMeta, error) {
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
		p.attachment.File.FileSize = len(decoded.image)
		p.attachment.File.UpdatedAt = time.Now()
		p.attachment.Processing = gtsmodel.ProcessingStatusProcessed

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
func (p *ProcessingMedia) fetchRawData(ctx context.Context) error {
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

	if !supportedImage(contentType) {
		return fmt.Errorf("fetchRawData: media type %s not (yet) supported", contentType)
	}

	split := strings.Split(contentType, "/")
	if len(split) != 2 {
		return fmt.Errorf("fetchRawData: content type %s was not valid", contentType)
	}

	extension := split[1] // something like 'jpeg'

	// set some additional fields on the attachment now that
	// we know more about what the underlying media actually is
	if extension == mimeGif {
		p.attachment.Type = gtsmodel.FileTypeGif
	} else {
		p.attachment.Type = gtsmodel.FileTypeImage
	}
	p.attachment.URL = uris.GenerateURIForAttachment(p.attachment.AccountID, string(TypeAttachment), string(SizeOriginal), p.attachment.ID, extension)
	p.attachment.File.Path = fmt.Sprintf("%s/%s/%s/%s.%s", p.attachment.AccountID, TypeAttachment, SizeOriginal, p.attachment.ID, extension)
	p.attachment.File.ContentType = contentType

	return nil
}

func (m *manager) preProcessMedia(ctx context.Context, data DataFunc, accountID string, ai *AdditionalMediaInfo) (*ProcessingMedia, error) {
	id, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	file := gtsmodel.File{
		Path:        "", // we don't know yet because it depends on the uncalled DataFunc
		ContentType: "", // we don't know yet because it depends on the uncalled DataFunc
		UpdatedAt:   time.Now(),
	}

	thumbnail := gtsmodel.Thumbnail{
		URL:         uris.GenerateURIForAttachment(accountID, string(TypeAttachment), string(SizeSmall), id, mimeJpeg), // all thumbnails are encoded as jpeg,
		Path:        fmt.Sprintf("%s/%s/%s/%s.%s", accountID, TypeAttachment, SizeSmall, id, mimeJpeg),                 // all thumbnails are encoded as jpeg,
		ContentType: mimeJpeg,
		UpdatedAt:   time.Now(),
	}

	// populate initial fields on the media attachment -- some of these will be overwritten as we proceed
	attachment := &gtsmodel.MediaAttachment{
		ID:                id,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		StatusID:          "",
		URL:               "", // we don't know yet because it depends on the uncalled DataFunc
		RemoteURL:         "",
		Type:              gtsmodel.FileTypeUnknown, // we don't know yet because it depends on the uncalled DataFunc
		FileMeta:          gtsmodel.FileMeta{},
		AccountID:         accountID,
		Description:       "",
		ScheduledStatusID: "",
		Blurhash:          "",
		Processing:        gtsmodel.ProcessingStatusReceived,
		File:              file,
		Thumbnail:         thumbnail,
		Avatar:            false,
		Header:            false,
	}

	// check if we have additional info to add to the attachment,
	// and overwrite some of the attachment fields if so
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

	processingMedia := &ProcessingMedia{
		attachment:    attachment,
		data:          data,
		thumbstate:    received,
		fullSizeState: received,
		database:      m.db,
		storage:       m.storage,
	}

	return processingMedia, nil
}
