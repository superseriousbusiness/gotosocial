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

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/util"
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
	// Ensure we have a valid remote URL.
	url, err := url.Parse(remoteURL)
	if err != nil {
		err := gtserror.Newf("invalid media remote url %s: %w", remoteURL, err)
		return nil, err
	}

	return d.processMediaSafeley(ctx,
		remoteURL,
		func() (*media.ProcessingMedia, error) {

			// Fetch transport for the provided request user from controller.
			tsport, err := d.transportController.NewTransportForUsername(ctx,
				requestUser,
			)
			if err != nil {
				return nil, gtserror.Newf("failed getting transport for %s: %w", requestUser, err)
			}

			// Get maximum supported remote media size.
			maxsz := int64(config.GetMediaRemoteMaxSize()) // #nosec G115 -- Already validated.

			// Create media with prepared info.
			return d.mediaManager.CreateMedia(
				ctx,
				accountID,
				func(ctx context.Context) (io.ReadCloser, error) {
					return tsport.DereferenceMedia(ctx, url, maxsz)
				},
				info,
			)
		},
	)
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
	attach *gtsmodel.MediaAttachment,
	info media.AdditionalMediaInfo,
	force bool,
) (
	*gtsmodel.MediaAttachment,
	error,
) {
	// Can't refresh local.
	if attach.IsLocal() {
		return attach, nil
	}

	// Check blurhash up-to-date.
	if info.Blurhash != nil &&
		*info.Blurhash != attach.Blurhash {
		attach.Blurhash = *info.Blurhash
		force = true
	}

	// Check description up-to-date.
	if info.Description != nil &&
		*info.Description != attach.Description {
		attach.Description = *info.Description
		force = true
	}

	// Check remote URL up-to-date.
	if info.RemoteURL != nil &&
		*info.RemoteURL != attach.RemoteURL {
		attach.RemoteURL = *info.RemoteURL
		force = true
	}

	// Check if needs updating.
	if *attach.Cached && !force {
		return attach, nil
	}

	// Ensure we have a valid remote URL.
	url, err := url.Parse(attach.RemoteURL)
	if err != nil {
		err := gtserror.Newf("invalid media remote url %s: %w", attach.RemoteURL, err)
		return nil, err
	}

	// Pass along for safe processing.
	return d.processMediaSafeley(ctx,
		attach.RemoteURL,
		func() (*media.ProcessingMedia, error) {

			// Fetch transport for the provided request user from controller.
			tsport, err := d.transportController.NewTransportForUsername(ctx,
				requestUser,
			)
			if err != nil {
				return nil, gtserror.Newf("failed getting transport for %s: %w", requestUser, err)
			}

			// Get maximum supported remote media size.
			maxsz := int64(config.GetMediaRemoteMaxSize()) // #nosec G115 -- Already validated.

			// Recache media with prepared info,
			// this will also update media in db.
			return d.mediaManager.CacheMedia(
				attach,
				func(ctx context.Context) (io.ReadCloser, error) {
					return tsport.DereferenceMedia(ctx, url, maxsz)
				},
			), nil
		},
	)
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

// processingMediaSafely provides concurrency-safe processing of
// a media with given remote URL string. if a copy of the media is
// not already being processed, the given 'process' callback will
// be used to generate new *media.ProcessingMedia{} instance.
func (d *Dereferencer) processMediaSafeley(
	ctx context.Context,
	remoteURL string,
	process func() (*media.ProcessingMedia, error),
) (
	media *gtsmodel.MediaAttachment,
	err error,
) {

	// Acquire map lock.
	d.derefMediaMu.Lock()

	// Ensure unlock only done once.
	unlock := d.derefMediaMu.Unlock
	unlock = util.DoOnce(unlock)
	defer unlock()

	// Look for an existing deref in progress.
	processing, ok := d.derefMedia[remoteURL]

	if !ok {
		// Start new processing emoji.
		processing, err = process()
		if err != nil {
			return nil, err
		}

		// Add processing media to hash map.
		d.derefMedia[remoteURL] = processing

		defer func() {
			// Remove on finish.
			d.derefMediaMu.Lock()
			delete(d.derefMedia, remoteURL)
			d.derefMediaMu.Unlock()
		}()
	}

	// Unlock map.
	unlock()

	// Perform media load operation.
	media, err = processing.Load(ctx)
	if err != nil {
		err = gtserror.Newf("error loading media %s: %w", remoteURL, err)

		// TODO: in time we should return checkable flags by gtserror.Is___()
		// which can determine if loading error should allow remaining placeholder.
	}

	return
}
