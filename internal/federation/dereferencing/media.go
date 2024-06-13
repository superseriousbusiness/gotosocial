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

package dereferencing

import (
	"context"
	"io"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

// GetMedia fetches the media at given remote URL by
// dereferencing it. The passed accountID is used to
// store it as being owned by that account. Additional
// information to set on the media attachment may also
// be provided.
//
// Please note that even if an error is returned,
// a media model may still be returned if the error
// was only encountered during actual dereferencing.
// In this case, it will act as a placeholder.
//
// Also note that since account / status dereferencing is
// already protected by per-uri locks, and that fediverse
// media is generally not shared between accounts (etc),
// there aren't any concurrency protections against multiple
// insertion / dereferencing of media at remoteURL. Worst
// case scenario, an extra media entry will be inserted
// and the scheduled cleaner.Cleaner{} will catch it!
func (d *Dereferencer) GetMedia(
	ctx context.Context,
	requestUser string,
	accountID string, // media account owner
	remoteURL string,
	info media.AdditionalMediaInfo,
) (
	*gtsmodel.MediaAttachment,
	error,
) {
	// Parse str as valid URL object.
	url, err := url.Parse(remoteURL)
	if err != nil {
		return nil, gtserror.Newf("invalid remote media url %q: %v", remoteURL, err)
	}

	// Fetch transport for the provided request user from controller.
	tsport, err := d.transportController.NewTransportForUsername(ctx,
		requestUser,
	)
	if err != nil {
		return nil, gtserror.Newf("failed getting transport for %s: %w", requestUser, err)
	}

	// Start processing remote attachment at URL.
	processing, err := d.mediaManager.CreateMedia(
		ctx,
		accountID,
		func(ctx context.Context) (io.ReadCloser, int64, error) {
			return tsport.DereferenceMedia(ctx, url)
		},
		info,
	)
	if err != nil {
		return nil, err
	}

	// Perform media load operation.
	media, err := processing.Load(ctx)
	if err != nil {
		err = gtserror.Newf("error loading media %s: %w", media.RemoteURL, err)

		// TODO: in time we should return checkable flags by gtserror.Is___()
		// which can determine if loading error should allow remaining placeholder.
	}

	return media, err
}

// RefreshMedia ensures that given media is up-to-date,
// both in terms of being cached in local instance,
// storage and compared to extra info in information
// in given gtsmodel.AdditionMediaInfo{}. This handles
// the case of local emoji by returning early.
//
// Please note that even if an error is returned,
// a media model may still be returned if the error
// was only encountered during actual dereferencing.
// In this case, it will act as a placeholder.
//
// Also note that since account / status dereferencing is
// already protected by per-uri locks, and that fediverse
// media is generally not shared between accounts (etc),
// there aren't any concurrency protections against multiple
// insertion / dereferencing of media at remoteURL. Worst
// case scenario, an extra media entry will be inserted
// and the scheduled cleaner.Cleaner{} will catch it!
func (d *Dereferencer) RefreshMedia(
	ctx context.Context,
	requestUser string,
	media *gtsmodel.MediaAttachment,
	info media.AdditionalMediaInfo,
	force bool,
) (
	*gtsmodel.MediaAttachment,
	error,
) {
	// Can't refresh local.
	if media.IsLocal() {
		return media, nil
	}

	// Check emoji is up-to-date
	// with provided extra info.
	switch {
	case info.Blurhash != nil &&
		*info.Blurhash != media.Blurhash:
		force = true
	case info.Description != nil &&
		*info.Description != media.Description:
		force = true
	case info.RemoteURL != nil &&
		*info.RemoteURL != media.RemoteURL:
		force = true
	}

	// Check if needs updating.
	if !force && *media.Cached {
		return media, nil
	}

	// TODO: more finegrained freshness checks.

	// Ensure we have a valid remote URL.
	url, err := url.Parse(media.RemoteURL)
	if err != nil {
		err := gtserror.Newf("invalid media remote url %s: %w", media.RemoteURL, err)
		return nil, err
	}

	// Fetch transport for the provided request user from controller.
	tsport, err := d.transportController.NewTransportForUsername(ctx,
		requestUser,
	)
	if err != nil {
		return nil, gtserror.Newf("failed getting transport for %s: %w", requestUser, err)
	}

	// Start processing remote attachment recache.
	processing := d.mediaManager.RecacheMedia(
		media,
		func(ctx context.Context) (io.ReadCloser, int64, error) {
			return tsport.DereferenceMedia(ctx, url)
		},
	)

	// Perform media load operation.
	media, err = processing.Load(ctx)
	if err != nil {
		err = gtserror.Newf("error loading media %s: %w", media.RemoteURL, err)

		// TODO: in time we should return checkable flags by gtserror.Is___()
		// which can determine if loading error should allow remaining placeholder.
	}

	return media, err
}

// updateAttachment handles the case of an existing media attachment
// that *may* have changes or need recaching. it checks for changed
// fields, updating in the database if so, and recaches uncached media.
func (d *Dereferencer) updateAttachment(
	ctx context.Context,
	requestUser string,
	existing *gtsmodel.MediaAttachment, // existing attachment
	attach *gtsmodel.MediaAttachment, // (optional) changed media
) (
	*gtsmodel.MediaAttachment, // always set
	error,
) {
	var info media.AdditionalMediaInfo

	if attach != nil {
		// Set optional extra information,
		// (will later check for changes).
		info.Description = &attach.Description
		info.Blurhash = &attach.Blurhash
		info.RemoteURL = &attach.RemoteURL
	}

	// Ensure media is cached.
	return d.RefreshMedia(ctx,
		requestUser,
		existing,
		info,
		false,
	)
}
