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
	terminator "github.com/superseriousbusiness/exif-terminator"
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
	postData   PostDataCallbackFunc
	read       bool // bool indicating that data function has been triggered already

	thumbState    int32 // the processing state of the media thumbnail
	fullSizeState int32 // the processing state of the full-sized media

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
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.store(ctx); err != nil {
		return nil, err
	}

	if err := p.loadThumb(ctx); err != nil {
		return nil, err
	}

	if err := p.loadFullSize(ctx); err != nil {
		return nil, err
	}

	// store the result in the database before returning it
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
	return atomic.LoadInt32(&p.thumbState) == int32(complete) && atomic.LoadInt32(&p.fullSizeState) == int32(complete)
}

func (p *ProcessingMedia) loadThumb(ctx context.Context) error {
	thumbState := atomic.LoadInt32(&p.thumbState)
	switch processState(thumbState) {
	case received:
		// we haven't processed a thumbnail for this media yet so do it now

		// check if we need to create a blurhash or if there's already one set
		var createBlurhash bool
		if p.attachment.Blurhash == "" {
			// no blurhash created yet
			createBlurhash = true
		}

		// stream the original file out of storage...
		stored, err := p.storage.GetStream(p.attachment.File.Path)
		if err != nil {
			p.err = fmt.Errorf("loadThumb: error fetching file from storage: %s", err)
			atomic.StoreInt32(&p.thumbState, int32(errored))
			return p.err
		}

		// ... and into the derive thumbnail function
		thumb, err := deriveThumbnail(stored, p.attachment.File.ContentType, createBlurhash)
		if err != nil {
			p.err = fmt.Errorf("loadThumb: error deriving thumbnail: %s", err)
			atomic.StoreInt32(&p.thumbState, int32(errored))
			return p.err
		}

		if err := stored.Close(); err != nil {
			p.err = fmt.Errorf("loadThumb: error closing stored full size: %s", err)
			atomic.StoreInt32(&p.thumbState, int32(errored))
			return p.err
		}

		// put the thumbnail in storage
		if err := p.storage.Put(p.attachment.Thumbnail.Path, thumb.small); err != nil {
			p.err = fmt.Errorf("loadThumb: error storing thumbnail: %s", err)
			atomic.StoreInt32(&p.thumbState, int32(errored))
			return p.err
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
		p.attachment.Thumbnail.FileSize = len(thumb.small)

		// we're done processing the thumbnail!
		atomic.StoreInt32(&p.thumbState, int32(complete))
		fallthrough
	case complete:
		return nil
	case errored:
		return p.err
	}

	return fmt.Errorf("loadThumb: thumbnail processing status %d unknown", p.thumbState)
}

func (p *ProcessingMedia) loadFullSize(ctx context.Context) error {
	fullSizeState := atomic.LoadInt32(&p.fullSizeState)
	switch processState(fullSizeState) {
	case received:
		var err error
		var decoded *imageMeta

		// stream the original file out of storage...
		stored, err := p.storage.GetStream(p.attachment.File.Path)
		if err != nil {
			p.err = fmt.Errorf("loadFullSize: error fetching file from storage: %s", err)
			atomic.StoreInt32(&p.fullSizeState, int32(errored))
			return p.err
		}

		// decode the image
		ct := p.attachment.File.ContentType
		switch ct {
		case mimeImageJpeg, mimeImagePng:
			decoded, err = decodeImage(stored, ct)
		case mimeImageGif:
			decoded, err = decodeGif(stored)
		default:
			err = fmt.Errorf("loadFullSize: content type %s not a processible image type", ct)
		}

		if err != nil {
			p.err = err
			atomic.StoreInt32(&p.fullSizeState, int32(errored))
			return p.err
		}

		if err := stored.Close(); err != nil {
			p.err = fmt.Errorf("loadFullSize: error closing stored full size: %s", err)
			atomic.StoreInt32(&p.fullSizeState, int32(errored))
			return p.err
		}

		// set appropriate fields on the attachment based on the image we derived
		p.attachment.FileMeta.Original = gtsmodel.Original{
			Width:  decoded.width,
			Height: decoded.height,
			Size:   decoded.size,
			Aspect: decoded.aspect,
		}
		p.attachment.File.UpdatedAt = time.Now()
		p.attachment.Processing = gtsmodel.ProcessingStatusProcessed

		// we're done processing the full-size image
		atomic.StoreInt32(&p.fullSizeState, int32(complete))
		fallthrough
	case complete:
		return nil
	case errored:
		return p.err
	}

	return fmt.Errorf("loadFullSize: full size processing status %d unknown", p.fullSizeState)
}

// store calls the data function attached to p if it hasn't been called yet,
// and updates the underlying attachment fields as necessary. It will then stream
// bytes from p's reader directly into storage so that it can be retrieved later.
func (p *ProcessingMedia) store(ctx context.Context) error {
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
	if !supportedImage(contentType) {
		return fmt.Errorf("store: media type %s not (yet) supported", contentType)
	}

	// extract the file extension
	split := strings.Split(contentType, "/")
	if len(split) != 2 {
		return fmt.Errorf("store: content type %s was not valid", contentType)
	}
	extension := split[1] // something like 'jpeg'

	// concatenate the cleaned up first bytes with the existing bytes still in the reader (thanks Mara)
	multiReader := io.MultiReader(bytes.NewBuffer(firstBytes), reader)

	// we'll need to clean exif data from the first bytes; while we're
	// here, we can also use the extension to derive the attachment type
	var clean io.Reader
	switch extension {
	case mimeGif:
		p.attachment.Type = gtsmodel.FileTypeGif
		clean = multiReader // nothing to clean from a gif
	case mimeJpeg, mimePng:
		p.attachment.Type = gtsmodel.FileTypeImage
		purged, err := terminator.Terminate(multiReader, fileSize, extension)
		if err != nil {
			return fmt.Errorf("store: exif error: %s", err)
		}
		clean = purged
	default:
		return fmt.Errorf("store: couldn't process %s", extension)
	}

	// now set some additional fields on the attachment since
	// we know more about what the underlying media actually is
	p.attachment.URL = uris.GenerateURIForAttachment(p.attachment.AccountID, string(TypeAttachment), string(SizeOriginal), p.attachment.ID, extension)
	p.attachment.File.Path = fmt.Sprintf("%s/%s/%s/%s.%s", p.attachment.AccountID, TypeAttachment, SizeOriginal, p.attachment.ID, extension)
	p.attachment.File.ContentType = contentType
	p.attachment.File.FileSize = fileSize

	// store this for now -- other processes can pull it out of storage as they please
	if err := p.storage.PutStream(p.attachment.File.Path, clean); err != nil {
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

func (m *manager) preProcessMedia(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, accountID string, ai *AdditionalMediaInfo) (*ProcessingMedia, error) {
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
		ContentType: mimeImageJpeg,
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
		postData:      postData,
		thumbState:    int32(received),
		fullSizeState: int32(received),
		database:      m.db,
		storage:       m.storage,
	}

	return processingMedia, nil
}
