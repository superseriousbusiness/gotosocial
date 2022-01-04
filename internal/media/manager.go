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

	"codeberg.org/gruf/go-runners"
	"codeberg.org/gruf/go-store/kv"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

// Manager provides an interface for managing media: parsing, storing, and retrieving media objects like photos, videos, and gifs.
type Manager interface {
	ProcessMedia(ctx context.Context, data []byte, accountID string) (*Media, error)
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

func (m *manager) ProcessMedia(ctx context.Context, data []byte, accountID string) (*Media, error) {
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
		media, err := m.preProcessImage(ctx, data, contentType, accountID)
		if err != nil {
			return nil, err
		}

		m.pool.Enqueue(func(innerCtx context.Context) {
			select {
			case <-innerCtx.Done():
				// if the inner context is done that means the worker pool is closing, so we should just return
				return
			default:
				media.PreLoad(innerCtx)
			}
		})

		return media, nil
	default:
		return nil, fmt.Errorf("content type %s not (yet) supported", contentType)
	}
}
