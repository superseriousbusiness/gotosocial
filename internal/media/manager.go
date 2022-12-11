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
	"time"

	"github.com/robfig/cron/v3"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

// selectPruneLimit is the amount of media entries to select at a time from the db when pruning
const selectPruneLimit = 20

// UnusedLocalAttachmentCacheDays is the amount of days to keep local media in storage if it
// is not attached to a status, or was never attached to a status.
const UnusedLocalAttachmentCacheDays = 3

// Manager provides an interface for managing media: parsing, storing, and retrieving media objects like photos, videos, and gifs.
type Manager interface {
	// Stop stops the underlying worker pool of the manager. It should be called
	// when closing GoToSocial in order to cleanly finish any in-progress jobs.
	// It will block until workers are finished processing.
	Stop() error

	/*
		PROCESSING FUNCTIONS
	*/

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
	//
	// If refresh is true, this indicates that the emoji image has changed and should be updated.
	ProcessEmoji(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, shortcode string, id string, uri string, ai *AdditionalEmojiInfo, refresh bool) (*ProcessingEmoji, error)
	// RecacheMedia refetches, reprocesses, and recaches an existing attachment that has been uncached via pruneRemote.
	RecacheMedia(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, attachmentID string) (*ProcessingMedia, error)

	/*
		PRUNING FUNCTIONS
	*/

	// PruneAllRemote prunes all remote media attachments cached on this instance which are older than the given amount of days.
	// 'Pruning' in this context means removing the locally stored data of the attachment (both thumbnail and full size),
	// and setting 'cached' to false on the associated attachment.
	//
	// The returned int is the amount of media that was pruned by this function.
	PruneAllRemote(ctx context.Context, olderThanDays int) (int, error)
	// PruneAllMeta prunes unused/out of date headers and avatars cached on this instance.
	//
	// The returned int is the amount of media that was pruned by this function.
	PruneAllMeta(ctx context.Context) (int, error)
	// PruneUnusedLocalAttachments prunes unused media attachments that were uploaded by
	// a user on this instance, but never actually attached to a status, or attached but
	// later detached.
	//
	// The returned int is the amount of media that was pruned by this function.
	PruneUnusedLocalAttachments(ctx context.Context) (int, error)
	// PruneOrphaned prunes files that exist in storage but which do not have a corresponding
	// entry in the database.
	//
	// If dry is true, then nothing will be changed, only the amount that *would* be removed
	// is returned to the caller.
	PruneOrphaned(ctx context.Context, dry bool) (int, error)

	/*
		REFETCHING FUNCTIONS
		Useful when data loss has occurred.
	*/

	// RefetchEmojis iterates through remote emojis (for the given domain, or all if domain is empty string).
	//
	// For each emoji, the manager will check whether both the full size and static images are present in storage.
	// If not, the manager will refetch and reprocess full size and static images for the emoji.
	//
	// The provided DereferenceMedia function will be used when it's necessary to refetch something this way.
	RefetchEmojis(ctx context.Context, domain string, dereferenceMedia DereferenceMedia) (int, error)
}

type manager struct {
	db           db.DB
	storage      *storage.Driver
	emojiWorker  *concurrency.WorkerPool[*ProcessingEmoji]
	mediaWorker  *concurrency.WorkerPool[*ProcessingMedia]
	stopCronJobs func() error
}

// NewManager returns a media manager with the given db and underlying storage.
//
// A worker pool will also be initialized for the manager, to ensure that only
// a limited number of media will be processed in parallel. The numbers of workers
// is determined from the $GOMAXPROCS environment variable (usually no. CPU cores).
// See internal/concurrency.NewWorkerPool() documentation for further information.
func NewManager(database db.DB, storage *storage.Driver) (Manager, error) {
	m := &manager{
		db:      database,
		storage: storage,
	}

	// Prepare the media worker pool
	m.mediaWorker = concurrency.NewWorkerPool[*ProcessingMedia](-1, 10)
	m.mediaWorker.SetProcessor(func(ctx context.Context, media *ProcessingMedia) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := media.LoadAttachment(ctx); err != nil {
			return fmt.Errorf("error loading media %s: %v", media.AttachmentID(), err)
		}
		return nil
	})

	// Prepare the emoji worker pool
	m.emojiWorker = concurrency.NewWorkerPool[*ProcessingEmoji](-1, 10)
	m.emojiWorker.SetProcessor(func(ctx context.Context, emoji *ProcessingEmoji) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := emoji.LoadEmoji(ctx); err != nil {
			return fmt.Errorf("error loading emoji %s: %v", emoji.EmojiID(), err)
		}
		return nil
	})

	// Start the worker pools
	if err := m.mediaWorker.Start(); err != nil {
		return nil, err
	}
	if err := m.emojiWorker.Start(); err != nil {
		return nil, err
	}

	if err := scheduleCleanupJobs(m); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *manager) ProcessMedia(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, accountID string, ai *AdditionalMediaInfo) (*ProcessingMedia, error) {
	processingMedia, err := m.preProcessMedia(ctx, data, postData, accountID, ai)
	if err != nil {
		return nil, err
	}
	m.mediaWorker.Queue(processingMedia)
	return processingMedia, nil
}

func (m *manager) ProcessEmoji(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, shortcode string, id string, uri string, ai *AdditionalEmojiInfo, refresh bool) (*ProcessingEmoji, error) {
	processingEmoji, err := m.preProcessEmoji(ctx, data, postData, shortcode, id, uri, ai, refresh)
	if err != nil {
		return nil, err
	}
	m.emojiWorker.Queue(processingEmoji)
	return processingEmoji, nil
}

func (m *manager) RecacheMedia(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, attachmentID string) (*ProcessingMedia, error) {
	processingRecache, err := m.preProcessRecache(ctx, data, postData, attachmentID)
	if err != nil {
		return nil, err
	}
	m.mediaWorker.Queue(processingRecache)
	return processingRecache, nil
}

func (m *manager) Stop() error {
	// Stop media and emoji worker pools
	mediaErr := m.mediaWorker.Stop()
	emojiErr := m.emojiWorker.Stop()

	var cronErr error
	if m.stopCronJobs != nil {
		cronErr = m.stopCronJobs()
	}

	if mediaErr != nil {
		return mediaErr
	} else if emojiErr != nil {
		return emojiErr
	}

	return cronErr
}

func scheduleCleanupJobs(m *manager) error {
	// create a new cron instance for scheduling cleanup jobs
	c := cron.New(cron.WithLogger(&logrusWrapper{}))
	pruneCtx, pruneCancel := context.WithCancel(context.Background())

	if _, err := c.AddFunc("@midnight", func() {
		begin := time.Now()
		pruned, err := m.PruneAllMeta(pruneCtx)
		if err != nil {
			log.Errorf("media manager: error pruning meta: %s", err)
			return
		}
		log.Infof("media manager: pruned %d meta entries in %s", pruned, time.Since(begin))
	}); err != nil {
		pruneCancel()
		return fmt.Errorf("error starting media manager meta cleanup job: %s", err)
	}

	if _, err := c.AddFunc("@midnight", func() {
		begin := time.Now()
		pruned, err := m.PruneUnusedLocalAttachments(pruneCtx)
		if err != nil {
			log.Errorf("media manager: error pruning unused local attachments: %s", err)
			return
		}
		log.Infof("media manager: pruned %d unused local attachments in %s", pruned, time.Since(begin))
	}); err != nil {
		pruneCancel()
		return fmt.Errorf("error starting media manager unused local attachments cleanup job: %s", err)
	}

	// start remote cache cleanup cronjob if configured
	if mediaRemoteCacheDays := config.GetMediaRemoteCacheDays(); mediaRemoteCacheDays > 0 {
		if _, err := c.AddFunc("@midnight", func() {
			begin := time.Now()
			pruned, err := m.PruneAllRemote(pruneCtx, mediaRemoteCacheDays)
			if err != nil {
				log.Errorf("media manager: error pruning remote cache: %s", err)
				return
			}
			log.Infof("media manager: pruned %d remote cache entries in %s", pruned, time.Since(begin))
		}); err != nil {
			pruneCancel()
			return fmt.Errorf("error starting media manager remote cache cleanup job: %s", err)
		}
	}

	// try to stop any jobs gracefully by waiting til they're finished
	m.stopCronJobs = func() error {
		cronCtx := c.Stop()

		select {
		case <-cronCtx.Done():
			log.Infof("media manager: cron finished jobs and stopped gracefully")
		case <-time.After(1 * time.Minute):
			log.Infof("media manager: cron didn't stop after 60 seconds, will force close jobs")
			break
		}

		pruneCancel()
		return nil
	}

	c.Start()
	return nil
}
