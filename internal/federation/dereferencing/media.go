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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

// GetMedia ...
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

	// Force attachment loading.
	return processing.Load(ctx)
}

// RefreshMedia ...
func (d *Dereferencer) RefreshMedia(
	ctx context.Context,
	requestUser string,
	media *gtsmodel.MediaAttachment,
	force bool,
) (
	*gtsmodel.MediaAttachment,
	error,
) {
	// Can't refresh local.
	if media.IsLocal() {
		return media, nil
	}

	if !force {
		// Already cached.
		if *media.Cached {
			return media, nil
		}

		// TODO: in time update this
		// to perhaps follow a similar
		// freshness window to statuses
		// / accounts? But that's a big
		// maybe, media don't change in
		// the same way so this is largely
		// just to slow down fail retries.
		const maxfreq = 6 * time.Hour

		// Check whether media is uncached but repeatedly failing,
		// specifically limit the frequency at which we allow this.
		if !media.UpdatedAt.Equal(media.CreatedAt) && // i.e. not new
			media.UpdatedAt.Add(maxfreq).Before(time.Now()) {
			return media, nil
		}
	}

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

	// Force attachment loading.
	return processing.Load(ctx)
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
	var force bool

	if attach != nil {
		// Check if attachment description has changed.
		if existing.Description != attach.Description {
			existing.Description = attach.Description
			force = true
		}

		// Check if attachment blurhash has changed.
		if existing.Blurhash != attach.Blurhash {
			existing.Blurhash = attach.Blurhash
			force = true
		}
	}

	// Ensure media is cached.
	return d.RefreshMedia(ctx,
		requestUser,
		existing,
		force,
	)
}
