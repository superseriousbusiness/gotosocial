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
	"runtime"

	"codeberg.org/gruf/go-runners"
	"codeberg.org/gruf/go-store/kv"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

// Manager provides an interface for managing media: parsing, storing, and retrieving media objects like photos, videos, and gifs.
type Manager interface {
	// ProcessMedia begins the process of decoding and storing the given data as an attachment.
	// It will return a pointer to a ProcessingMedia struct upon which further actions can be performed, such as getting
	// the finished media, thumbnail, attachment, etc.
	//
	// data should be a function that the media manager can call to return a reader containing the media data.
	//
	// postData will be called after data has been called; it can be used to clean up any remaining resources.
	// The provided function can be nil, in which case it will not be executed.
	//
	// accountID should be the account that the media belongs to.
	//
	// ai is optional and can be nil. Any additional information about the attachment provided will be put in the database.
	ProcessMedia(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, accountID string, ai *AdditionalMediaInfo) (*ProcessingMedia, error)
	// ProcessEmoji begins the process of decoding and storing the given data as an emoji.
	// It will return a pointer to a ProcessingEmoji struct upon which further actions can be performed, such as getting
	// the finished media, thumbnail, attachment, etc.
	//
	// data should be a function that the media manager can call to return a reader containing the emoji data.
	//
	// postData will be called after data has been called; it can be used to clean up any remaining resources.
	// The provided function can be nil, in which case it will not be executed.
	//
	// shortcode should be the emoji shortcode without the ':'s around it.
	//
	// id is the database ID that should be used to store the emoji.
	//
	// uri is the ActivityPub URI/ID of the emoji.
	//
	// ai is optional and can be nil. Any additional information about the emoji provided will be put in the database.
	ProcessEmoji(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, shortcode string, id string, uri string, ai *AdditionalEmojiInfo) (*ProcessingEmoji, error)
	// NumWorkers returns the total number of workers available to this manager.
	NumWorkers() int
	// QueueSize returns the total capacity of the queue.
	QueueSize() int
	// JobsQueued returns the number of jobs currently in the task queue.
	JobsQueued() int
	// ActiveWorkers returns the number of workers currently performing jobs.
	ActiveWorkers() int
	// Stop stops the underlying worker pool of the manager. It should be called
	// when closing GoToSocial in order to cleanly finish any in-progress jobs.
	// It will block until workers are finished processing.
	Stop() error
}

type manager struct {
	db         db.DB
	storage    *kv.KVStore
	pool       runners.WorkerPool
	numWorkers int
	queueSize  int
}

// NewManager returns a media manager with the given db and underlying storage.
//
// A worker pool will also be initialized for the manager, to ensure that only
// a limited number of media will be processed in parallel.
//
// The number of workers will be the number of CPUs available to the Go runtime,
// divided by 2 (rounding down, but always at least 1).
//
// The length of the queue will be the number of workers multiplied by 10.
//
// So for an 8 core machine, the media manager will get 4 workers, and a queue of length 40.
// For a 4 core machine, this will be 2 workers, and a queue length of 20.
// For a single or 2-core machine, the media manager will get 1 worker, and a queue of length 10.
func NewManager(database db.DB, storage *kv.KVStore) (Manager, error) {
	numWorkers := runtime.NumCPU() / 2
	// make sure we always have at least 1 worker even on single-core machines
	if numWorkers == 0 {
		numWorkers = 1
	}
	queueSize := numWorkers * 10

	m := &manager{
		db:         database,
		storage:    storage,
		pool:       runners.NewWorkerPool(numWorkers, queueSize),
		numWorkers: numWorkers,
		queueSize:  queueSize,
	}

	if start := m.pool.Start(); !start {
		return nil, errors.New("could not start worker pool")
	}
	logrus.Debugf("started media manager worker pool with %d workers and queue capacity of %d", numWorkers, queueSize)

	return m, nil
}

func (m *manager) ProcessMedia(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, accountID string, ai *AdditionalMediaInfo) (*ProcessingMedia, error) {
	processingMedia, err := m.preProcessMedia(ctx, data, postData, accountID, ai)
	if err != nil {
		return nil, err
	}

	logrus.Tracef("ProcessMedia: about to enqueue media with attachmentID %s, queue length is %d", processingMedia.AttachmentID(), m.pool.Queue())
	m.pool.Enqueue(func(innerCtx context.Context) {
		select {
		case <-innerCtx.Done():
			// if the inner context is done that means the worker pool is closing, so we should just return
			return
		default:
			// start loading the media already for the caller's convenience
			if _, err := processingMedia.LoadAttachment(innerCtx); err != nil {
				logrus.Errorf("ProcessMedia: error processing media with attachmentID %s: %s", processingMedia.AttachmentID(), err)
			}
		}
	})
	logrus.Tracef("ProcessMedia: succesfully queued media with attachmentID %s, queue length is %d", processingMedia.AttachmentID(), m.pool.Queue())

	return processingMedia, nil
}

func (m *manager) ProcessEmoji(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, shortcode string, id string, uri string, ai *AdditionalEmojiInfo) (*ProcessingEmoji, error) {
	processingEmoji, err := m.preProcessEmoji(ctx, data, postData, shortcode, id, uri, ai)
	if err != nil {
		return nil, err
	}

	logrus.Tracef("ProcessEmoji: about to enqueue emoji with id %s, queue length is %d", processingEmoji.EmojiID(), m.pool.Queue())
	m.pool.Enqueue(func(innerCtx context.Context) {
		select {
		case <-innerCtx.Done():
			// if the inner context is done that means the worker pool is closing, so we should just return
			return
		default:
			// start loading the emoji already for the caller's convenience
			if _, err := processingEmoji.LoadEmoji(innerCtx); err != nil {
				logrus.Errorf("ProcessEmoji: error processing emoji with id %s: %s", processingEmoji.EmojiID(), err)
			}
		}
	})
	logrus.Tracef("ProcessEmoji: succesfully queued emoji with id %s, queue length is %d", processingEmoji.EmojiID(), m.pool.Queue())

	return processingEmoji, nil
}

func (m *manager) NumWorkers() int {
	return m.numWorkers
}

func (m *manager) QueueSize() int {
	return m.queueSize
}

func (m *manager) JobsQueued() int {
	return m.pool.Queue()
}

func (m *manager) ActiveWorkers() int {
	return m.pool.Workers()
}

func (m *manager) Stop() error {
	logrus.Info("stopping media manager worker pool")

	stopped := m.pool.Stop()
	if !stopped {
		return errors.New("could not stop media manager worker pool")
	}
	return nil
}
