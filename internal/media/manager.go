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
	"io"
	"time"

	"codeberg.org/gruf/go-iotools"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
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
	mimeImageWebp,
}

type Manager struct {
	state *state.State
}

// NewManager returns a media manager with given state.
func NewManager(state *state.State) *Manager {
	return &Manager{state: state}
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

// CreateMedia ...
func (m *Manager) CreateMedia(
	ctx context.Context,
	data DataFunc,
	accountID string,
	info AdditionalMediaInfo,
) (
	*ProcessingMedia,
	error,
) {
	now := time.Now()

	// Generate new ID.
	id := id.NewULID()

	// Calculate URI for attachment.
	uri := uris.URIForAttachment(
		accountID,
		string(TypeAttachment),
		string(SizeOriginal),
		id,
		"unknown",
	)

	// Calculate storage path for attachment.
	path := uris.StoragePathForAttachment(
		accountID,
		string(TypeAttachment),
		string(SizeOriginal),
		id,
		"unknown",
	)

	// Populate initial fields on the new media,
	// leaving out fields with values we don't know
	// yet. These will be overwritten as we go.
	attachment := &gtsmodel.MediaAttachment{
		ID:         id,
		CreatedAt:  now,
		UpdatedAt:  now,
		URL:        uri,
		Type:       gtsmodel.FileTypeUnknown,
		AccountID:  accountID,
		Processing: gtsmodel.ProcessingStatusReceived,
		File: gtsmodel.File{
			UpdatedAt:   now,
			ContentType: "application/octet-stream",
			Path:        path,
		},
		Thumbnail: gtsmodel.Thumbnail{UpdatedAt: now},
		Avatar:    util.Ptr(false),
		Header:    util.Ptr(false),
		Cached:    util.Ptr(false),
	}

	// Check if we were provided additional info
	// to add to the attachment, and overwrite
	// some of the attachment fields if so.
	if info.CreatedAt != nil {
		attachment.CreatedAt = *info.CreatedAt
	}
	if info.StatusID != nil {
		attachment.StatusID = *info.StatusID
	}
	if info.RemoteURL != nil {
		attachment.RemoteURL = *info.RemoteURL
	}
	if info.Description != nil {
		attachment.Description = *info.Description
	}
	if info.ScheduledStatusID != nil {
		attachment.ScheduledStatusID = *info.ScheduledStatusID
	}
	if info.Blurhash != nil {
		attachment.Blurhash = *info.Blurhash
	}
	if info.Avatar != nil {
		attachment.Avatar = info.Avatar
	}
	if info.Header != nil {
		attachment.Header = info.Header
	}
	if info.FocusX != nil {
		attachment.FileMeta.Focus.X = *info.FocusX
	}
	if info.FocusY != nil {
		attachment.FileMeta.Focus.Y = *info.FocusY
	}

	// Store attachment in database in initial form.
	err := m.state.DB.PutAttachment(ctx, attachment)
	if err != nil {
		return nil, err
	}

	// Return wrapped media for later processing.
	processingMedia := &ProcessingMedia{
		media:  attachment,
		dataFn: data,
		mgr:    m,
	}

	return processingMedia, nil
}

// PreProcessMediaRecache refetches, reprocesses,
// and recaches an existing attachment that has
// been uncached via cleaner pruning.
//
// Note: unlike ProcessMedia, this will NOT queue
// the media to be asychronously processed.

// RecacheMedia ...
func (m *Manager) RecacheMedia(
	ctx context.Context,
	data DataFunc,
	mediaID string,
) (
	*ProcessingMedia,
	error,
) {
	// Get the existing media attachment from database.
	attachment, err := m.state.DB.GetAttachmentByID(ctx, mediaID)
	if err != nil {
		return nil, err
	}

	// Return wrapped media for later processing.
	processingMedia := &ProcessingMedia{
		media:  attachment,
		dataFn: data,
		mgr:    m,
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

// CreateEmoji ...
func (m *Manager) CreateEmoji(
	ctx context.Context,
	data DataFunc,
	shortcode string,
	emojiID string,
	uri string,
	info AdditionalEmojiInfo,
	refresh bool,
) (
	*ProcessingEmoji,
	error,
) {
	now := time.Now()

	// Fetch the local instance account for emoji path generation.
	instanceAcc, err := m.state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, gtserror.Newf("error fetching instance account: %w", err)
	}

	// Generate new static URL for attachment.
	imageStaticURL := uris.URIForAttachment(
		instanceAcc.ID,
		string(TypeEmoji),
		string(SizeStatic),
		emojiID,

		// All static emojis
		// are encoded as png.
		mimePng,
	)

	// Generate new static image storage path for attachment.
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
	emoji := &gtsmodel.Emoji{
		ID:                     emojiID,
		Shortcode:              shortcode,
		ImageStaticURL:         imageStaticURL,
		ImageStaticPath:        imageStaticPath,
		ImageStaticContentType: mimeImagePng,
		ImageUpdatedAt:         now,
		Disabled:               util.Ptr(false),
		URI:                    uri,
		VisibleInPicker:        util.Ptr(true),
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	// Finally, create new emoji.
	return m.createEmoji(ctx,
		data,
		emoji,
		info,
	)
}

// RefreshEmoji ...
func (m *Manager) RefreshEmoji(
	ctx context.Context,
	data DataFunc,
	shortcode string,
	emojiID string,
	uri string,
	info AdditionalEmojiInfo,
) (
	*ProcessingEmoji,
	error,
) {
	// Fetch the local instance account for emoji path generation.
	instanceAcc, err := m.state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, gtserror.Newf("error fetching instance account: %w", err)
	}

	// Fetch existing emoji with ID from database.
	emoji, err := m.state.DB.GetEmojiByID(ctx, emojiID)
	if err != nil {
		err = gtserror.Newf("error fetching emoji to refresh from the db: %w", err)
		return nil, err
	}

	// Since this is a refresh, we will end up with
	// new images stored for this emoji, so we should
	// use an io.Closer callback to perform clean up
	// of the original images from storage.
	originalImageStaticPath := emoji.ImageStaticPath
	originalImagePath := emoji.ImagePath
	originalData := data

	data = func(ctx context.Context) (io.ReadCloser, int64, error) {
		// Call original data func.
		rc, sz, err := originalData(ctx)
		if err != nil {
			return nil, 0, err
		}

		// Wrap closer to cleanup old data.
		c := iotools.CloserCallback(rc, func() {
			if err := m.state.Storage.Delete(ctx, originalImagePath); err != nil && !storage.IsNotFound(err) {
				log.Errorf(ctx, "error removing old emoji %s@%s from storage: %v", emoji.Shortcode, emoji.Domain, err)
			}

			if err := m.state.Storage.Delete(ctx, originalImageStaticPath); err != nil && !storage.IsNotFound(err) {
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
	newPathID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.Newf("error generating newPathID for emoji refresh: %s", err)
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

	// Finally, create new emoji in database.
	processingEmoji, err := m.createEmoji(ctx,
		data,
		emoji,
		info,
	)
	if err != nil {
		return nil, err
	}

	// Indicate this was existing
	// emoji requiring refresh.
	processingEmoji.existing = true
	processingEmoji.newPathID = newPathID

	return processingEmoji, nil
}

func (m *Manager) createEmoji(
	ctx context.Context,
	data DataFunc,
	emoji *gtsmodel.Emoji,
	info AdditionalEmojiInfo,
) (
	*ProcessingEmoji,
	error,
) {
	// Check if we have additional info to add to the emoji,
	// and overwrite some of the emoji fields if so.
	if info.CreatedAt != nil {
		emoji.CreatedAt = *info.CreatedAt
	}
	if info.Domain != nil {
		emoji.Domain = *info.Domain
	}
	if info.ImageRemoteURL != nil {
		emoji.ImageRemoteURL = *info.ImageRemoteURL
	}
	if info.ImageStaticRemoteURL != nil {
		emoji.ImageStaticRemoteURL = *info.ImageStaticRemoteURL
	}
	if info.Disabled != nil {
		emoji.Disabled = info.Disabled
	}
	if info.VisibleInPicker != nil {
		emoji.VisibleInPicker = info.VisibleInPicker
	}
	if info.CategoryID != nil {
		emoji.CategoryID = *info.CategoryID
	}

	// Return wrapped emoji for later processing.
	processingEmoji := &ProcessingEmoji{
		emoji:  emoji,
		dataFn: data,
		mgr:    m,
	}

	return processingEmoji, nil
}

// PreProcessEmojiRecache refetches, reprocesses, and recaches
// an existing emoji that has been uncached via cleaner pruning.
//
// Note: unlike ProcessEmoji, this will NOT queue the emoji to
// be asychronously processed.

// RecacheEmoji ...
func (m *Manager) RecacheEmoji(
	ctx context.Context,
	data DataFunc,
	emojiID string,
) (
	*ProcessingEmoji,
	error,
) {
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
