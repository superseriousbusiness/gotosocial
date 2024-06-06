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
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// loadAttachment handles the case of a new media attachment
// that requires loading. it stores and caches from given data.
func (d *Dereferencer) loadAttachment(
	ctx context.Context,
	tsport transport.Transport,
	accountID string, // media account owner
	remoteURL string,
	info *media.AdditionalMediaInfo,
) (
	*gtsmodel.MediaAttachment,
	error,
) {
	// Parse str as valid URL object.
	url, err := url.Parse(remoteURL)
	if err != nil {
		return nil, gtserror.Newf("invalid remote media url %q: %v", remoteURL, err)
	}

	// Start pre-processing remote media at remote URL.
	processing := d.mediaManager.PreProcessMedia(
		func(ctx context.Context) (io.ReadCloser, int64, error) {
			return tsport.DereferenceMedia(ctx, url)
		},
		accountID,
		info,
	)

	// Force attachment loading *right now*.
	return processing.LoadAttachment(ctx)
}

// updateAttachment handles the case of an existing media attachment
// that *may* have changes or need recaching. it checks for changed
// fields, updating in the database if so, and recaches uncached media.
func (d *Dereferencer) updateAttachment(
	ctx context.Context,
	tsport transport.Transport,
	existing *gtsmodel.MediaAttachment,
	media *gtsmodel.MediaAttachment,
) (
	*gtsmodel.MediaAttachment, // always set
	error,
) {

	// Possible changed media columns.
	changed := make([]string, 0, 3)

	// Check if attachment description has changed.
	if existing.Description != media.Description {
		changed = append(changed, "description")
		existing.Description = media.Description
	}

	// Check if attachment blurhash has changed (i.e. content change).
	if existing.Blurhash != media.Blurhash && media.Blurhash != "" {
		changed = append(changed, "blurhash", "cached")
		existing.Blurhash = media.Blurhash
		existing.Cached = util.Ptr(false)
	}

	if len(changed) > 0 {
		// Update the existing attachment model in the database.
		err := d.state.DB.UpdateAttachment(ctx, existing, changed...)
		if err != nil {
			return media, gtserror.Newf("error updating media: %w", err)
		}
	}

	// Check if cached.
	if *media.Cached {
		return media, nil
	}

	// Parse str as valid URL object.
	url, err := url.Parse(media.RemoteURL)
	if err != nil {
		return nil, gtserror.Newf("invalid remote media url %q: %v", media.RemoteURL, err)
	}

	// Start pre-processing remote media recaching from remote.
	processing, err := d.mediaManager.PreProcessMediaRecache(
		ctx,
		func(ctx context.Context) (io.ReadCloser, int64, error) {
			return tsport.DereferenceMedia(ctx, url)
		},
		media.ID,
	)
	if err != nil {
		return nil, gtserror.Newf("error processing recache: %w", err)
	}

	// Force load attachment recache *right now*.
	recached, err := processing.LoadAttachment(ctx)

	if recached != nil {
		// Only set recached media
		// file if it was at least
		// partially downloaded.
		media = recached
	}

	return media, err
}

// pollChanged returns whether a poll has changed in way that
// indicates that this should be an entirely new poll. i.e. if
// the available options have changed, or the expiry has increased.
func pollChanged(existing, latest *gtsmodel.Poll) bool {
	return !slices.Equal(existing.Options, latest.Options) ||
		!existing.ExpiresAt.Equal(latest.ExpiresAt)
}

// pollUpdated returns whether a poll has updated, i.e. if the
// vote counts have changed, or if it has expired / been closed.
func pollUpdated(existing, latest *gtsmodel.Poll) bool {
	return *existing.Voters != *latest.Voters ||
		!slices.Equal(existing.Votes, latest.Votes) ||
		!existing.ClosedAt.Equal(latest.ClosedAt)
}

// pollJustClosed returns whether a poll has *just* closed.
func pollJustClosed(existing, latest *gtsmodel.Poll) bool {
	return existing.ClosedAt.IsZero() && latest.Closed()
}
