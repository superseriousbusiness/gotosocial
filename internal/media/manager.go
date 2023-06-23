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
	"errors"
	"fmt"
	"io"
	"time"

	"codeberg.org/gruf/go-iotools"
	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
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

type Manager struct {
	state *state.State
}

// NewManager returns a media manager with the given db and underlying storage.
//
// A worker pool will also be initialized for the manager, to ensure that only
// a limited number of media will be processed in parallel. The numbers of workers
// is determined from the $GOMAXPROCS environment variable (usually no. CPU cores).
// See internal/concurrency.NewWorkerPool() documentation for further information.
func NewManager(state *state.State) *Manager {
	m := &Manager{state: state}
	return m
}

// PreProcessMedia begins the process of decoding and storing the given data as an attachment.
// It will return a pointer to a ProcessingMedia struct upon which further actions can be performed, such as getting
// the finished media, thumbnail, attachment, etc.
//
// data should be a function that the media manager can call to return a reader containing the media data.
//
// accountID should be the account that the media belongs to.
//
// ai is optional and can be nil. Any additional information about the attachment provided will be put in the database.
//
// Note: unlike ProcessMedia, this will NOT queue the media to be asynchronously processed.
func (m *Manager) PreProcessMedia(ctx context.Context, data DataFunc, accountID string, ai *AdditionalMediaInfo) (*ProcessingMedia, error) {
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
		mgr:    m,
	}

	return processingMedia, nil
}

// PreProcessMediaRecache refetches, reprocesses, and recaches an existing attachment that has been uncached via pruneRemote.
//
// Note: unlike ProcessMedia, this will NOT queue the media to be asychronously processed.
func (m *Manager) PreProcessMediaRecache(ctx context.Context, data DataFunc, attachmentID string) (*ProcessingMedia, error) {
	// get the existing attachment from database.
	attachment, err := m.state.DB.GetAttachmentByID(ctx, attachmentID)
	if err != nil {
		return nil, err
	}

	processingMedia := &ProcessingMedia{
		media:   attachment,
		dataFn:  data,
		recache: true, // indicate it's a recache
		mgr:     m,
	}

	return processingMedia, nil
}

// ProcessMedia will call PreProcessMedia, followed by queuing the media to be processing in the media worker queue.
func (m *Manager) ProcessMedia(ctx context.Context, data DataFunc, accountID string, ai *AdditionalMediaInfo) (*ProcessingMedia, error) {
	// Create a new processing media object for this media request.
	media, err := m.PreProcessMedia(ctx, data, accountID, ai)
	if err != nil {
		return nil, err
	}

	// Attempt to add this media processing item to the worker queue.
	_ = m.state.Workers.Media.MustEnqueueCtx(ctx, media.Process)

	return media, nil
}

// PreProcessEmoji begins the process of decoding and storing the given data as an emoji.
// It will return a pointer to a ProcessingEmoji struct upon which further actions can be performed, such as getting
// the finished media, thumbnail, attachment, etc.
//
// data should be a function that the media manager can call to return a reader containing the emoji data.
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
func (m *Manager) PreProcessEmoji(ctx context.Context, data DataFunc, shortcode string, emojiID string, uri string, ai *AdditionalEmojiInfo, refresh bool) (*ProcessingEmoji, error) {
	instanceAccount, err := m.state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, gtserror.Newf("error fetching this instance account from the db: %s", err)
	}

	var (
		newPathID string
		emoji     *gtsmodel.Emoji
		now       = time.Now()
	)

	if refresh {
		// Look for existing emoji by given ID.
		emoji, err = m.state.DB.GetEmojiByID(ctx, emojiID)
		if err != nil {
			return nil, gtserror.Newf("error fetching emoji to refresh from the db: %s", err)
		}

		// if this is a refresh, we will end up with new images
		// stored for this emoji, so we can use an io.Closer callback
		// to perform clean up of the old images from storage
		originalData := data
		originalImagePath := emoji.ImagePath
		originalImageStaticPath := emoji.ImageStaticPath
		data = func(ctx context.Context) (io.ReadCloser, int64, error) {
			// Call original data func.
			rc, sz, err := originalData(ctx)
			if err != nil {
				return nil, 0, err
			}

			// Wrap closer to cleanup old data.
			c := iotools.CloserCallback(rc, func() {
				if err := m.state.Storage.Delete(ctx, originalImagePath); err != nil && !errors.Is(err, storage.ErrNotFound) {
					log.Errorf(ctx, "error removing old emoji %s@%s from storage: %v", emoji.Shortcode, emoji.Domain, err)
				}

				if err := m.state.Storage.Delete(ctx, originalImageStaticPath); err != nil && !errors.Is(err, storage.ErrNotFound) {
					log.Errorf(ctx, "error removing old static emoji %s@%s from storage: %v", emoji.Shortcode, emoji.Domain, err)
				}
			})

			// Return newly wrapped readcloser and size.
			return iotools.ReadCloser(rc, c), sz, nil
		}

		newPathID, err = id.NewRandomULID()
		if err != nil {
			return nil, gtserror.Newf("error generating alternateID for emoji refresh: %s", err)
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
		mgr:       m,
	}

	return processingEmoji, nil
}

// ProcessEmoji will call PreProcessEmoji, followed by queuing the emoji to be processing in the emoji worker queue.
func (m *Manager) ProcessEmoji(ctx context.Context, data DataFunc, shortcode string, id string, uri string, ai *AdditionalEmojiInfo, refresh bool) (*ProcessingEmoji, error) {
	// Create a new processing emoji object for this emoji request.
	emoji, err := m.PreProcessEmoji(ctx, data, shortcode, id, uri, ai, refresh)
	if err != nil {
		return nil, err
	}

	// Attempt to add this emoji processing item to the worker queue.
	_ = m.state.Workers.Media.MustEnqueueCtx(ctx, emoji.Process)

	return emoji, nil
}
