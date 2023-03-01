/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"time"

	"codeberg.org/gruf/go-runners"
	"codeberg.org/gruf/go-sched"
	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

var SupportedMIMETypes = []string{
	mimeImageJpeg,
	mimeImageGif,
	mimeImagePng,
	mimeImageWebp,
	mimeVideoMp4,
}

var SupportedEmojiMIMETypes = []string{
	mimeImageGif,
	mimeImagePng,
}

// Manager provides an interface for managing media: parsing, storing, and retrieving media objects like photos, videos, and gifs.
type Manager interface {
	/*
		PROCESSING FUNCTIONS
	*/

	// PreProcessMedia begins the process of decoding and storing the given data as an attachment.
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
	//
	// Note: unlike ProcessMedia, this will NOT queue the media to be asynchronously processed.
	PreProcessMedia(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, accountID string, ai *AdditionalMediaInfo) (*ProcessingMedia, error)

	// PreProcessMediaRecache refetches, reprocesses, and recaches an existing attachment that has been uncached via pruneRemote.
	//
	// Note: unlike ProcessMedia, this will NOT queue the media to be asychronously processed.
	PreProcessMediaRecache(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, attachmentID string) (*ProcessingMedia, error)

	// ProcessMedia will call PreProcessMedia, followed by queuing the media to be processing in the media worker queue.
	ProcessMedia(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, accountID string, ai *AdditionalMediaInfo) (*ProcessingMedia, error)

	// PreProcessEmoji begins the process of decoding and storing the given data as an emoji.
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
	// Note: unlike ProcessEmoji, this will NOT queue the emoji to be asynchronously processed.
	PreProcessEmoji(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, shortcode string, id string, uri string, ai *AdditionalEmojiInfo, refresh bool) (*ProcessingEmoji, error)

	// ProcessEmoji will call PreProcessEmoji, followed by queuing the emoji to be processing in the emoji worker queue.
	ProcessEmoji(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, shortcode string, id string, uri string, ai *AdditionalEmojiInfo, refresh bool) (*ProcessingEmoji, error)

	/*
		PRUNING/UNCACHING FUNCTIONS
	*/

	// PruneAll runs all of the below pruning/uncacheing functions, and then cleans up any resulting
	// empty directories from the storage driver. It can be called as a shortcut for calling the below
	// pruning functions one by one.
	//
	// If blocking is true, then any errors encountered during the prune will be combined + returned to
	// the caller. If blocking is false, the prune is run in the background and errors are just logged
	// instead.
	PruneAll(ctx context.Context, mediaCacheRemoteDays int, blocking bool) error
	// UncacheRemote uncaches all remote media attachments older than the given amount of days.
	//
	// In this context, uncacheing means deleting media files from storage and marking the attachment
	// as cached=false in the database.
	//
	// If 'dry' is true, then only a dry run will be performed: nothing will actually be changed.
	//
	// The returned int is the amount of media that was/would be uncached by this function.
	UncacheRemote(ctx context.Context, olderThanDays int, dry bool) (int, error)
	// PruneUnusedRemote prunes unused/out of date headers and avatars cached on this instance.
	//
	// The returned int is the amount of media that was pruned by this function.
	PruneUnusedRemote(ctx context.Context, dry bool) (int, error)
	// PruneUnusedLocal prunes unused media attachments that were uploaded by
	// a user on this instance, but never actually attached to a status, or attached but
	// later detached.
	//
	// The returned int is the amount of media that was pruned by this function.
	PruneUnusedLocal(ctx context.Context, dry bool) (int, error)
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
	state *state.State
}

// NewManager returns a media manager with the given db and underlying storage.
//
// A worker pool will also be initialized for the manager, to ensure that only
// a limited number of media will be processed in parallel. The numbers of workers
// is determined from the $GOMAXPROCS environment variable (usually no. CPU cores).
// See internal/concurrency.NewWorkerPool() documentation for further information.
func NewManager(state *state.State) Manager {
	m := &manager{state: state}
	scheduleCleanupJobs(m)
	return m
}

func (m *manager) PreProcessMedia(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, accountID string, ai *AdditionalMediaInfo) (*ProcessingMedia, error) {
	id, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	avatar := false
	header := false
	cached := false
	now := time.Now()

	// populate initial fields on the media attachment -- some of these will be overwritten as we proceed
	attachment := &gtsmodel.MediaAttachment{
		ID:                id,
		CreatedAt:         now,
		UpdatedAt:         now,
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
		File:              gtsmodel.File{UpdatedAt: now},
		Thumbnail:         gtsmodel.Thumbnail{UpdatedAt: now},
		Avatar:            &avatar,
		Header:            &header,
		Cached:            &cached,
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
			attachment.Avatar = ai.Avatar
		}

		if ai.Header != nil {
			attachment.Header = ai.Header
		}

		if ai.FocusX != nil {
			attachment.FileMeta.Focus.X = *ai.FocusX
		}

		if ai.FocusY != nil {
			attachment.FileMeta.Focus.Y = *ai.FocusY
		}
	}

	processingMedia := &ProcessingMedia{
		media:  attachment,
		dataFn: data,
		postFn: postData,
		mgr:    m,
	}

	return processingMedia, nil
}

func (m *manager) PreProcessMediaRecache(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, attachmentID string) (*ProcessingMedia, error) {
	// get the existing attachment from database.
	attachment, err := m.state.DB.GetAttachmentByID(ctx, attachmentID)
	if err != nil {
		return nil, err
	}

	processingMedia := &ProcessingMedia{
		media:   attachment,
		dataFn:  data,
		postFn:  postData,
		recache: true, // indicate it's a recache
		mgr:     m,
	}

	return processingMedia, nil
}

func (m *manager) ProcessMedia(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, accountID string, ai *AdditionalMediaInfo) (*ProcessingMedia, error) {
	// Create a new processing media object for this media request.
	media, err := m.PreProcessMedia(ctx, data, postData, accountID, ai)
	if err != nil {
		return nil, err
	}

	// Attempt to add this media processing item to the worker queue.
	_ = m.state.Workers.Media.MustEnqueueCtx(ctx, media.Process)

	return media, nil
}

func (m *manager) PreProcessEmoji(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, shortcode string, emojiID string, uri string, ai *AdditionalEmojiInfo, refresh bool) (*ProcessingEmoji, error) {
	instanceAccount, err := m.state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("preProcessEmoji: error fetching this instance account from the db: %s", err)
	}

	var (
		newPathID string
		emoji     *gtsmodel.Emoji
		now       = time.Now()
	)

	if refresh {
		emoji, err = m.state.DB.GetEmojiByID(ctx, emojiID)
		if err != nil {
			return nil, fmt.Errorf("preProcessEmoji: error fetching emoji to refresh from the db: %s", err)
		}

		// if this is a refresh, we will end up with new images
		// stored for this emoji, so we can use the postData function
		// to perform clean up of the old images from storage
		originalPostData := postData
		originalImagePath := emoji.ImagePath
		originalImageStaticPath := emoji.ImageStaticPath
		postData = func(innerCtx context.Context) error {
			// trigger the original postData function if it was provided
			if originalPostData != nil {
				if err := originalPostData(innerCtx); err != nil {
					return err
				}
			}

			l := log.WithContext(ctx).
				WithField("shortcode@domain", emoji.Shortcode+"@"+emoji.Domain)
			l.Debug("postData: cleaning up old emoji files for refreshed emoji")
			if err := m.state.Storage.Delete(innerCtx, originalImagePath); err != nil && !errors.Is(err, storage.ErrNotFound) {
				l.Errorf("postData: error cleaning up old emoji image at %s for refreshed emoji: %s", originalImagePath, err)
			}
			if err := m.state.Storage.Delete(innerCtx, originalImageStaticPath); err != nil && !errors.Is(err, storage.ErrNotFound) {
				l.Errorf("postData: error cleaning up old emoji static image at %s for refreshed emoji: %s", originalImageStaticPath, err)
			}

			return nil
		}

		newPathID, err = id.NewRandomULID()
		if err != nil {
			return nil, fmt.Errorf("preProcessEmoji: error generating alternateID for emoji refresh: %s", err)
		}

		// store + serve static image at new path ID
		emoji.ImageStaticURL = uris.GenerateURIForAttachment(instanceAccount.ID, string(TypeEmoji), string(SizeStatic), newPathID, mimePng)
		emoji.ImageStaticPath = fmt.Sprintf("%s/%s/%s/%s.%s", instanceAccount.ID, TypeEmoji, SizeStatic, newPathID, mimePng)

		emoji.Shortcode = shortcode
		emoji.URI = uri
	} else {
		disabled := false
		visibleInPicker := true

		// populate initial fields on the emoji -- some of these will be overwritten as we proceed
		emoji = &gtsmodel.Emoji{
			ID:                     emojiID,
			CreatedAt:              now,
			Shortcode:              shortcode,
			Domain:                 "", // assume our own domain unless told otherwise
			ImageRemoteURL:         "",
			ImageStaticRemoteURL:   "",
			ImageURL:               "",                                                                                                         // we don't know yet
			ImageStaticURL:         uris.GenerateURIForAttachment(instanceAccount.ID, string(TypeEmoji), string(SizeStatic), emojiID, mimePng), // all static emojis are encoded as png
			ImagePath:              "",                                                                                                         // we don't know yet
			ImageStaticPath:        fmt.Sprintf("%s/%s/%s/%s.%s", instanceAccount.ID, TypeEmoji, SizeStatic, emojiID, mimePng),                 // all static emojis are encoded as png
			ImageContentType:       "",                                                                                                         // we don't know yet
			ImageStaticContentType: mimeImagePng,                                                                                               // all static emojis are encoded as png
			ImageFileSize:          0,
			ImageStaticFileSize:    0,
			Disabled:               &disabled,
			URI:                    uri,
			VisibleInPicker:        &visibleInPicker,
			CategoryID:             "",
		}
	}

	emoji.ImageUpdatedAt = now
	emoji.UpdatedAt = now

	// check if we have additional info to add to the emoji,
	// and overwrite some of the emoji fields if so
	if ai != nil {
		if ai.CreatedAt != nil {
			emoji.CreatedAt = *ai.CreatedAt
		}

		if ai.Domain != nil {
			emoji.Domain = *ai.Domain
		}

		if ai.ImageRemoteURL != nil {
			emoji.ImageRemoteURL = *ai.ImageRemoteURL
		}

		if ai.ImageStaticRemoteURL != nil {
			emoji.ImageStaticRemoteURL = *ai.ImageStaticRemoteURL
		}

		if ai.Disabled != nil {
			emoji.Disabled = ai.Disabled
		}

		if ai.VisibleInPicker != nil {
			emoji.VisibleInPicker = ai.VisibleInPicker
		}

		if ai.CategoryID != nil {
			emoji.CategoryID = *ai.CategoryID
		}
	}

	processingEmoji := &ProcessingEmoji{
		instAccID: instanceAccount.ID,
		emoji:     emoji,
		refresh:   refresh,
		newPathID: newPathID,
		dataFn:    data,
		postFn:    postData,
		mgr:       m,
	}

	return processingEmoji, nil
}

func (m *manager) ProcessEmoji(ctx context.Context, data DataFunc, postData PostDataCallbackFunc, shortcode string, id string, uri string, ai *AdditionalEmojiInfo, refresh bool) (*ProcessingEmoji, error) {
	// Create a new processing emoji object for this emoji request.
	emoji, err := m.PreProcessEmoji(ctx, data, postData, shortcode, id, uri, ai, refresh)
	if err != nil {
		return nil, err
	}

	// Attempt to add this emoji processing item to the worker queue.
	_ = m.state.Workers.Media.MustEnqueueCtx(ctx, emoji.Process)

	return emoji, nil
}

func scheduleCleanupJobs(m *manager) {
	const day = time.Hour * 24

	// Calculate closest midnight.
	now := time.Now()
	midnight := now.Round(day)

	if midnight.Before(now) {
		// since <= 11:59am rounds down.
		midnight = midnight.Add(day)
	}

	// Get ctx associated with scheduler run state.
	done := m.state.Workers.Scheduler.Done()
	doneCtx := runners.CancelCtx(done)

	// TODO: we'll need to do some thinking to make these
	// jobs restartable if we want to implement reloads in
	// the future that make call to Workers.Stop() -> Workers.Start().

	// Schedule the PruneAll task to execute every day at midnight.
	m.state.Workers.Scheduler.Schedule(sched.NewJob(func(now time.Time) {
		err := m.PruneAll(doneCtx, config.GetMediaRemoteCacheDays(), true)
		if err != nil {
			log.Errorf(nil, "error during prune: %v", err)
		}
		log.Infof(nil, "finished pruning all in %s", time.Since(now))
	}).EveryAt(midnight, day))
}
