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
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"codeberg.org/gruf/go-runners"
	"codeberg.org/gruf/go-store/kv"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// Manager provides an interface for managing media: parsing, storing, and retrieving media objects like photos, videos, and gifs.
type Manager interface {
	// ProcessMedia begins the process of decoding and storing the given data as a piece of media (aka an attachment).
	// It will return a pointer to a Media struct upon which further actions can be performed, such as getting
	// the finished media, thumbnail, decoded bytes, attachment, and setting additional fields.
	//
	// accountID should be the account that the media belongs to.
	//
	// RemoteURL is optional, and can be an empty string. Setting this to a non-empty string indicates that
	// the piece of media originated on a remote instance and has been dereferenced to be cached locally.
	ProcessMedia(ctx context.Context, data []byte, accountID string, ai *AdditionalInfo) (*Media, error)

	ProcessEmoji(ctx context.Context, data []byte, accountID string) (*Media, error)
}

type manager struct {
	db      db.DB
	storage *kv.KVStore
	pool    runners.WorkerPool
}

// New returns a media manager with the given db and underlying storage.
func New(database db.DB, storage *kv.KVStore) (Manager, error) {
	workers := runtime.NumCPU() / 2
	queue := workers * 10
	pool := runners.NewWorkerPool(workers, queue)

	if start := pool.Start(); !start {
		return nil, errors.New("could not start worker pool")
	}
	logrus.Debugf("started media manager worker pool with %d workers and queue capacity of %d", workers, queue)

	m := &manager{
		db:      database,
		storage: storage,
		pool:    pool,
	}

	return m, nil
}

/*
	INTERFACE FUNCTIONS
*/

func (m *manager) ProcessMedia(ctx context.Context, data []byte, accountID string, ai *AdditionalInfo) (*Media, error) {
	contentType, err := parseContentType(data)
	if err != nil {
		return nil, err
	}

	split := strings.Split(contentType, "/")
	if len(split) != 2 {
		return nil, fmt.Errorf("content type %s malformed", contentType)
	}

	mainType := split[0]

	switch mainType {
	case mimeImage:
		media, err := m.preProcessImage(ctx, data, contentType, accountID, ai)
		if err != nil {
			return nil, err
		}

		m.pool.Enqueue(func(innerCtx context.Context) {
			select {
			case <-innerCtx.Done():
				// if the inner context is done that means the worker pool is closing, so we should just return
				return
			default:
				// start preloading the media for the caller's convenience
				media.preLoad(innerCtx)
			}
		})

		return media, nil
	default:
		return nil, fmt.Errorf("content type %s not (yet) supported", contentType)
	}
}

func (m *manager) ProcessEmoji(ctx context.Context, data []byte, accountID string) (*Media, error) {
	return nil, nil
}

// preProcessImage initializes processing
func (m *manager) preProcessImage(ctx context.Context, data []byte, contentType string, accountID string, ai *AdditionalInfo) (*Media, error) {
	if !supportedImage(contentType) {
		return nil, fmt.Errorf("image type %s not supported", contentType)
	}

	if len(data) == 0 {
		return nil, errors.New("image was of size 0")
	}

	id, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	extension := strings.Split(contentType, "/")[1]

	attachment := &gtsmodel.MediaAttachment{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		StatusID:  "",
		URL:       uris.GenerateURIForAttachment(accountID, string(TypeAttachment), string(SizeOriginal), id, extension),
		RemoteURL: "",
		Type:      gtsmodel.FileTypeImage,
		FileMeta: gtsmodel.FileMeta{
			Focus: gtsmodel.Focus{
				X: 0,
				Y: 0,
			},
		},
		AccountID:         accountID,
		Description:       "",
		ScheduledStatusID: "",
		Blurhash:          "",
		Processing:        0,
		File: gtsmodel.File{
			Path:        fmt.Sprintf("%s/%s/%s/%s.%s", accountID, TypeAttachment, SizeOriginal, id, extension),
			ContentType: contentType,
			UpdatedAt:   time.Now(),
		},
		Thumbnail: gtsmodel.Thumbnail{
			URL:         uris.GenerateURIForAttachment(accountID, string(TypeAttachment), string(SizeSmall), id, mimeJpeg), // all thumbnails are encoded as jpeg,
			Path:        fmt.Sprintf("%s/%s/%s/%s.%s", accountID, TypeAttachment, SizeSmall, id, mimeJpeg),                 // all thumbnails are encoded as jpeg,
			ContentType: mimeJpeg,
			UpdatedAt:   time.Now(),
		},
		Avatar: false,
		Header: false,
	}

	// check if we have additional info to add to the attachment
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

	media := &Media{
		attachment:    attachment,
		rawData:       data,
		thumbstate:    received,
		fullSizeState: received,
		database:      m.db,
		storage:       m.storage,
	}

	return media, nil
}
