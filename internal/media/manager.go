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
	"github.com/superseriousbusiness/gotosocial/internal/util"
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

// NewManager returns a media manager with given state.
func NewManager(state *state.State) *Manager {
	m := &Manager{state: state}
	return m
}

// PreProcessMedia begins the process of decoding
// and storing the given data as an attachment.
// It will return a pointer to a ProcessingMedia
// struct upon which further actions can be performed,
// such as getting the finished media, thumbnail,
// attachment, etc.
//
//   - data: a function that the media manager can call
//     to return a reader containing the media data.
//   - accountID: the account that the media belongs to.
//   - ai: optional and can be nil. Any additional information
//     about the attachment provided will be put in the database.
//
// Note: unlike ProcessMedia, this will NOT
// queue the media to be asynchronously processed.
func (m *Manager) PreProcessMedia(
	data DataFunc,
	accountID string,
	ai *AdditionalMediaInfo,
) *ProcessingMedia {
	// Populate initial fields on the new media,
	// leaving out fields with values we don't know
	// yet. These will be overwritten as we go.
	now := time.Now()
	attachment := &gtsmodel.MediaAttachment{
		ID:         id.NewULID(),
		CreatedAt:  now,
		UpdatedAt:  now,
		Type:       gtsmodel.FileTypeUnknown,
		FileMeta:   gtsmodel.FileMeta{},
		AccountID:  accountID,
		Processing: gtsmodel.ProcessingStatusReceived,
		File: gtsmodel.File{
			UpdatedAt:   now,
			ContentType: "application/octet-stream",
		},
		Thumbnail: gtsmodel.Thumbnail{UpdatedAt: now},
		Avatar:    util.Ptr(false),
		Header:    util.Ptr(false),
		Cached:    util.Ptr(false),
	}

	attachment.URL = uris.URIForAttachment(
		accountID,
		string(TypeAttachment),
		string(SizeOriginal),
		attachment.ID,
		"unknown",
	)

	attachment.File.Path = uris.StoragePathForAttachment(
		accountID,
		string(TypeAttachment),
		string(SizeOriginal),
		attachment.ID,
		"unknown",
	)

	// Check if we were provided additional info
	// to add to the attachment, and overwrite
	// some of the attachment fields if so.
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

	return processingMedia
}

// PreProcessMediaRecache refetches, reprocesses,
// and recaches an existing attachment that has
// been uncached via cleaner pruning.
//
// Note: unlike ProcessMedia, this will NOT queue
// the media to be asychronously processed.
func (m *Manager) PreProcessMediaRecache(
	ctx context.Context,
	data DataFunc,
	attachmentID string,
) (*ProcessingMedia, error) {
	// Get the existing attachment from database.
	attachment, err := m.state.DB.GetAttachmentByID(ctx, attachmentID)
	if err != nil {
		return nil, err
	}

	processingMedia := &ProcessingMedia{
		media:   attachment,
		dataFn:  data,
		recache: true, // Indicate it's a recache.
		mgr:     m,
	}

	return processingMedia, nil
}

// PreProcessEmoji begins the process of decoding and storing
// the given data as an emoji. It will return a pointer to a
// ProcessingEmoji struct upon which further actions can be
// performed, such as getting the finished media, thumbnail,
// attachment, etc.
//
//   - data: function that the media manager can call
//     to return a reader containing the emoji data.
//   - shortcode: the emoji shortcode without the ':'s around it.
//   - emojiID: database ID that should be used to store the emoji.
//   - uri: ActivityPub URI/ID of the emoji.
//   - ai: optional and can be nil. Any additional information
//     about the emoji provided will be put in the database.
//   - refresh: refetch/refresh the emoji.
//
// Note: unlike ProcessEmoji, this will NOT queue
// the emoji to be asynchronously processed.
func (m *Manager) PreProcessEmoji(
	ctx context.Context,
	data DataFunc,
	shortcode string,
	emojiID string,
	uri string,
	ai *AdditionalEmojiInfo,
	refresh bool,
) (*ProcessingEmoji, error) {
	var (
		newPathID string
		emoji     *gtsmodel.Emoji
		now       = time.Now()
	)

	// Fetch the local instance account for emoji path generation.
	instanceAcc, err := m.state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, gtserror.Newf("error fetching instance account: %w", err)
	}

	if refresh {
		// Existing emoji!

		emoji, err = m.state.DB.GetEmojiByID(ctx, emojiID)
		if err != nil {
			err = gtserror.Newf("error fetching emoji to refresh from the db: %w", err)
			return nil, err
		}

		// Since this is a refresh, we will end up with
		// new images stored for this emoji, so we should
		// use an io.Closer callback to perform clean up
		// of the original images from storage.
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

		// Reuse existing shortcode and URI -
		// these don't change when we refresh.
		emoji.Shortcode = shortcode
		emoji.URI = uri

		// Use a new ID to create a new path
		// for the new images, to get around
		// needing to do cache invalidation.
		newPathID, err = id.NewRandomULID()
		if err != nil {
			return nil, gtserror.Newf("error generating alternateID for emoji refresh: %s", err)
		}

		emoji.ImageStaticURL = uris.URIForAttachment(
			instanceAcc.ID,
			string(TypeEmoji),
			string(SizeStatic),
			newPathID,
			// All static emojis
			// are encoded as png.
			mimePng,
		)

		emoji.ImageStaticPath = uris.StoragePathForAttachment(
			instanceAcc.ID,
			string(TypeEmoji),
			string(SizeStatic),
			newPathID,
			// All static emojis
			// are encoded as png.
			mimePng,
		)
	} else {
		// New emoji!

		imageStaticURL := uris.URIForAttachment(
			instanceAcc.ID,
			string(TypeEmoji),
			string(SizeStatic),
			emojiID,
			// All static emojis
			// are encoded as png.
			mimePng,
		)

		imageStaticPath := uris.StoragePathForAttachment(
			instanceAcc.ID,
			string(TypeEmoji),
			string(SizeStatic),
			emojiID,
			// All static emojis
			// are encoded as png.
			mimePng,
		)

		// Populate initial fields on the new emoji,
		// leaving out fields with values we don't know
		// yet. These will be overwritten as we go.
		emoji = &gtsmodel.Emoji{
			ID:                     emojiID,
			CreatedAt:              now,
			UpdatedAt:              now,
			Shortcode:              shortcode,
			ImageStaticURL:         imageStaticURL,
			ImageStaticPath:        imageStaticPath,
			ImageStaticContentType: mimeImagePng,
			ImageUpdatedAt:         now,
			Disabled:               util.Ptr(false),
			URI:                    uri,
			VisibleInPicker:        util.Ptr(true),
		}
	}

	// Check if we have additional info to add to the emoji,
	// and overwrite some of the emoji fields if so.
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
		emoji:     emoji,
		existing:  refresh,
		newPathID: newPathID,
		dataFn:    data,
		mgr:       m,
	}

	return processingEmoji, nil
}

// PreProcessEmojiRecache refetches, reprocesses, and recaches
// an existing emoji that has been uncached via cleaner pruning.
//
// Note: unlike ProcessEmoji, this will NOT queue the emoji to
// be asychronously processed.
func (m *Manager) PreProcessEmojiRecache(
	ctx context.Context,
	data DataFunc,
	emojiID string,
) (*ProcessingEmoji, error) {
	// Get the existing emoji from the database.
	emoji, err := m.state.DB.GetEmojiByID(ctx, emojiID)
	if err != nil {
		return nil, err
	}

	processingEmoji := &ProcessingEmoji{
		emoji:    emoji,
		dataFn:   data,
		existing: true, // Indicate recache.
		mgr:      m,
	}

	return processingEmoji, nil
}

// ProcessEmoji will call PreProcessEmoji, followed
// by queuing the emoji in the emoji worker queue.
func (m *Manager) ProcessEmoji(
	ctx context.Context,
	data DataFunc,
	shortcode string,
	id string,
	uri string,
	ai *AdditionalEmojiInfo,
	refresh bool,
) (*ProcessingEmoji, error) {
	// Create a new processing emoji object for this emoji request.
	emoji, err := m.PreProcessEmoji(ctx, data, shortcode, id, uri, ai, refresh)
	if err != nil {
		return nil, err
	}

	// Attempt to add this emoji processing item to the worker queue.
	_ = m.state.Workers.Media.MustEnqueueCtx(ctx, emoji.Process)

	return emoji, nil
}
