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
	accountID string,
	data DataFunc,
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

	// Pass prepared media as ready to be cached.
	return m.RecacheMedia(attachment, data), nil
}

// PreProcessMediaRecache refetches, reprocesses,
// and recaches an existing attachment that has
// been uncached via cleaner pruning.
//
// Note: unlike ProcessMedia, this will NOT queue
// the media to be asychronously processed.

// RecacheMedia ...
func (m *Manager) RecacheMedia(
	media *gtsmodel.MediaAttachment,
	data DataFunc,
) *ProcessingMedia {
	return &ProcessingMedia{
		media:  media,
		dataFn: data,
		mgr:    m,
	}
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
	shortcode string,
	domain string,
	data DataFunc,
	info AdditionalEmojiInfo,
) (
	*ProcessingEmoji,
	error,
) {
	now := time.Now()

	// Generate new ID.
	id := id.NewULID()

	// Fetch the local instance account for emoji path generation.
	instanceAcc, err := m.state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, gtserror.Newf("error fetching instance account: %w", err)
	}

	// Create new ActivityPub URI.
	uri := uris.URIForEmoji(id)

	// Generate static URL for attachment.
	staticURL := uris.URIForAttachment(
		instanceAcc.ID,
		string(TypeEmoji),
		string(SizeStatic),
		id,

		// All static emojis
		// are encoded as png.
		mimePng,
	)

	// Generate static image path for attachment.
	staticPath := uris.StoragePathForAttachment(
		instanceAcc.ID,
		string(TypeEmoji),
		string(SizeStatic),
		id,

		// All static emojis
		// are encoded as png.
		mimePng,
	)

	// Populate initial fields on the new emoji,
	// leaving out fields with values we don't know
	// yet. These will be overwritten as we go.
	emoji := &gtsmodel.Emoji{
		ID:                     id,
		Shortcode:              shortcode,
		Domain:                 domain,
		ImageStaticURL:         staticURL,
		ImageStaticPath:        staticPath,
		ImageStaticContentType: mimeImagePng,
		ImageUpdatedAt:         now,
		Disabled:               util.Ptr(false),
		VisibleInPicker:        util.Ptr(true),
		URI:                    uri,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	// Finally, create new emoji.
	return m.createEmoji(ctx,
		m.state.DB.PutEmoji,
		data,
		emoji,
		info,
	)
}

// RefreshEmoji ...
func (m *Manager) RefreshEmoji(
	ctx context.Context,
	emoji *gtsmodel.Emoji,
	data DataFunc,
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

	// Create references to old emoji image
	// paths before they get updated with new
	// path ID. These are required for later
	// deleting the old image files on refresh.
	shortcodeDomain := util.ShortcodeDomain(emoji)
	oldStaticPath := emoji.ImageStaticPath
	oldPath := emoji.ImagePath

	// Since this is a refresh we will end up storing new images at new
	// paths, so we should wrap closer to delete old paths at completion.
	wrapped := func(ctx context.Context) (io.ReadCloser, int64, error) {

		// Call original data func.
		rc, sz, err := data(ctx)
		if err != nil {
			return nil, 0, err
		}

		// Wrap closer to cleanup old data.
		c := iotools.CloserFunc(func() error {

			// First try close original.
			if rc.Close(); err != nil {
				return err
			}

			// Remove any *old* emoji image file path now stream is closed.
			if err := m.state.Storage.Delete(ctx, oldPath); err != nil &&
				!storage.IsNotFound(err) {
				log.Errorf(ctx, "error deleting old emoji %s from storage: %v", shortcodeDomain, err)
			}

			// Remove any *old* emoji static image file path now stream is closed.
			if err := m.state.Storage.Delete(ctx, oldStaticPath); err != nil &&
				!storage.IsNotFound(err) {
				log.Errorf(ctx, "error deleting old static emoji %s from storage: %v", shortcodeDomain, err)
			}

			return nil
		})

		// Return newly wrapped readcloser and size.
		return iotools.ReadCloser(rc, c), sz, nil
	}

	// Use a new ID to create a new path
	// for the new images, to get around
	// needing to do cache invalidation.
	newPathID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.Newf("error generating newPathID for emoji refresh: %s", err)
	}

	// Generate new static URL for emoji.
	emoji.ImageStaticURL = uris.URIForAttachment(
		instanceAcc.ID,
		string(TypeEmoji),
		string(SizeStatic),
		newPathID,

		// All static emojis
		// are encoded as png.
		mimePng,
	)

	// Generate new static image storage path for emoji.
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
		func(ctx context.Context, emoji *gtsmodel.Emoji) error {
			return m.state.DB.UpdateEmoji(ctx, emoji)
		},
		wrapped,
		emoji,
		info,
	)
	if err != nil {
		return nil, err
	}

	// Set the refreshed path ID used.
	processingEmoji.newPathID = newPathID

	return processingEmoji, nil
}

// createEmoji ...
func (m *Manager) createEmoji(
	ctx context.Context,
	putDB func(context.Context, *gtsmodel.Emoji) error,
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

	// Store emoji in database in initial form.
	if err := putDB(ctx, emoji); err != nil {
		return nil, err
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
	emoji *gtsmodel.Emoji,
	data DataFunc,
) *ProcessingEmoji {
	return &ProcessingEmoji{
		emoji:  emoji,
		dataFn: data,
		mgr:    m,
	}
}
