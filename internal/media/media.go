package media

import (
	"context"
	"fmt"
	"sync"

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

type Media struct {
	mu sync.Mutex

	/*
		below fields should be set on newly created media;
		attachment will be updated incrementally as media goes through processing
	*/

	attachment *gtsmodel.MediaAttachment
	rawData    []byte

	/*
		below fields represent the processing state of the media thumbnail
	*/

	thumbing processState
	thumb    *ImageMeta

	/*
		below fields represent the processing state of the full-sized media
	*/

	processing processState
	processed  *ImageMeta

	/*
		below pointers to database and storage are maintained so that
		the media can store and update itself during processing steps
	*/

	database db.DB
	storage  *kv.KVStore

	err error // error created during processing, if any
}

func (m *Media) Thumb(ctx context.Context) (*ImageMeta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch m.thumbing {
	case received:
		// we haven't processed a thumbnail for this media yet so do it now
		thumb, err := deriveThumbnail(m.rawData, m.attachment.File.ContentType)
		if err != nil {
			m.err = fmt.Errorf("error deriving thumbnail: %s", err)
			m.thumbing = errored
			return nil, m.err
		}

		// put the thumbnail in storage
		if err := m.storage.Put(m.attachment.Thumbnail.Path, thumb.image); err != nil {
			m.err = fmt.Errorf("error storing thumbnail: %s", err)
			m.thumbing = errored
			return nil, m.err
		}

		// set appropriate fields on the attachment based on the thumbnail we derived
		m.attachment.Blurhash = thumb.blurhash
		m.attachment.FileMeta.Small = gtsmodel.Small{
			Width:  thumb.width,
			Height: thumb.height,
			Size:   thumb.size,
			Aspect: thumb.aspect,
		}
		m.attachment.Thumbnail.FileSize = thumb.size

		// put or update the attachment in the database
		if err := m.database.Put(ctx, m.attachment); err != nil {
			if err != db.ErrAlreadyExists {
				m.err = fmt.Errorf("error putting attachment: %s", err)
				m.thumbing = errored
				return nil, m.err
			}
			if err := m.database.UpdateByPrimaryKey(ctx, m.attachment); err != nil {
				m.err = fmt.Errorf("error updating attachment: %s", err)
				m.thumbing = errored
				return nil, m.err
			}
		}

		// set the thumbnail of this media
		m.thumb = thumb

		// we're done processing the thumbnail!
		m.thumbing = complete
		fallthrough
	case complete:
		return m.thumb, nil
	case errored:
		return nil, m.err
	}

	return nil, fmt.Errorf("thumbnail processing status %d unknown", m.thumbing)
}

func (m *Media) Full(ctx context.Context) (*ImageMeta, error) {
	var clean []byte
	var err error
	var original *ImageMeta

	ct := m.attachment.File.ContentType
aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
	switch ct {
	case mimeImageJpeg, mimeImagePng:
		// first 'clean' image by purging exif data from it
		var exifErr error
		if clean, exifErr = purgeExif(m.rawData); exifErr != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", exifErr)
		}
		original, err = decodeImage(clean, ct)
	case mimeImageGif:
		// gifs are already clean - no exif data to remove
		clean = m.rawData
		original, err = decodeGif(clean)
	default:
		err = fmt.Errorf("content type %s not a processible image type", ct)
	}

	if err != nil {
		return nil, err
	}

	return original, nil
}

func (m *Media) PreLoad(ctx context.Context) {
	go m.Thumb(ctx)
	m.mu.Lock()
	defer m.mu.Unlock()
}

func (m *Media) Load() {
	m.mu.Lock()
	defer m.mu.Unlock()
}
