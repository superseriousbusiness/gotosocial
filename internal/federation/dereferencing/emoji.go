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
	"errors"
	"io"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// GetEmoji ...
func (d *Dereferencer) GetEmoji(
	ctx context.Context,
	shortcode string,
	domain string,
	remoteURL string,
	info media.AdditionalEmojiInfo,
	refresh bool,
) (
	*gtsmodel.Emoji,
	error,
) {
	// Look for an existing emoji with shortcode domain.
	emoji, err := d.state.DB.GetEmojiByShortcodeDomain(ctx,
		shortcode,
		domain,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.Newf("error fetching emoji from db: %w", err)
	}

	if emoji != nil {
		if emoji.ImageRemoteURL != remoteURL {
			// Remote URL has oddly changed...
			// Force an emoji refresh.
			refresh = true
		}

		// This was an existing emoji, pass to refresh func.
		return d.RefreshEmoji(ctx, emoji, info, refresh)
	}

	if domain == "" {
		// failed local lookup, will be db.ErrNoEntries.
		return nil, gtserror.SetUnretrievable(err)
	}

	// Generate shortcode domain for locks + logging.
	shortcodeDomain := emoji.Shortcode + "@" + emoji.Domain

	// Ensure we have a valid remote URL.
	url, err := url.Parse(remoteURL)
	if err != nil {
		err := gtserror.Newf("invalid image remote url %s for emoji %s: %w", remoteURL, shortcodeDomain, err)
		return nil, err
	}

	// Acquire new instance account transport for emoji dereferencing.
	tsport, err := d.transportController.NewTransportForUsername(ctx, "")
	if err != nil {
		err := gtserror.Newf("error getting instance transport: %w", err)
		return nil, err
	}

	// Prepare data function to dereference remote emoji media.
	data := func(context.Context) (io.ReadCloser, int64, error) {
		return tsport.DereferenceMedia(ctx, url)
	}

	// Pass along for safe processing.
	return d.processEmojiSafely(ctx,
		shortcodeDomain,
		func() (*media.ProcessingEmoji, error) {
			return d.mediaManager.CreateEmoji(ctx,
				shortcode,
				domain,
				data,
				info,
			)
		},
	)
}

// RefreshEmoji ...
func (d *Dereferencer) RefreshEmoji(
	ctx context.Context,
	emoji *gtsmodel.Emoji,
	info media.AdditionalEmojiInfo,
	force bool,
) (
	*gtsmodel.Emoji,
	error,
) {
	// Can't refresh local.
	if emoji.IsLocal() {
		return emoji, nil
	}

	// Check if refresh needed.
	if *emoji.Cached && !force {
		return emoji, nil
	}

	// Generate shortcode domain for locks + logging.
	shortcodeDomain := emoji.Shortcode + "@" + emoji.Domain

	// Ensure we have a valid image remote URL.
	url, err := url.Parse(emoji.ImageRemoteURL)
	if err != nil {
		err := gtserror.Newf("invalid image remote url %s for emoji %s: %w", emoji.ImageRemoteURL, shortcodeDomain, err)
		return nil, err
	}

	// Acquire new instance account transport for emoji dereferencing.
	tsport, err := d.transportController.NewTransportForUsername(ctx, "")
	if err != nil {
		err := gtserror.Newf("error getting instance transport: %w", err)
		return nil, err
	}

	// Prepare data function to dereference remote emoji media.
	data := func(context.Context) (io.ReadCloser, int64, error) {
		return tsport.DereferenceMedia(ctx, url)
	}

	// Pass along for safe processing.
	return d.processEmojiSafely(ctx,
		shortcodeDomain,
		func() (*media.ProcessingEmoji, error) {
			return d.mediaManager.RefreshEmoji(ctx,
				emoji,
				data,
				info,
			)
		},
	)
}

// processingEmojiSafely ...
func (d *Dereferencer) processEmojiSafely(
	ctx context.Context,
	shortcodeDomain string,
	process func() (*media.ProcessingEmoji, error),
) (
	*gtsmodel.Emoji,
	error,
) {
	// Ensure unlock only done once.
	unlock := d.derefEmojisMu.Unlock
	unlock = util.DoOnce(unlock)
	defer unlock()

	// Look for an existing dereference in progress.
	processing, ok := d.derefEmojis[shortcodeDomain]

	if !ok {
		var err error

		// Start new processing emoji.
		processing, err = process()
		if err != nil {
			return nil, err
		}
	}

	// Unlock map.
	unlock()

	// Perform (blocking) load operation.
	emoji, err := processing.Load(ctx)
	if err != nil {
		err := gtserror.Newf("error loading emoji %s: %w", shortcodeDomain, err)
		return nil, err
	}

	// Return a COPY of emoji.
	emoji2 := new(gtsmodel.Emoji)
	*emoji2 = *emoji
	return emoji2, nil
}

func (d *Dereferencer) fetchEmojis(
	ctx context.Context,
	tsport transport.Transport,
	existing []*gtsmodel.Emoji,
	emojis []*gtsmodel.Emoji, // newly dereferenced
) (
	[]*gtsmodel.Emoji,
	bool, // any changes?
	error,
) {
	// Track any changes.
	changed := false

	for i, emoji := range emojis {
		// Look for an existing emoji with shortcode + domain.
		existing, ok := getEmojiByShortcodeDomain(existing,
			emoji.Shortcode,
			emoji.Domain,
		)
		if ok && existing.ID != "" {

			// Ensure that the existing emoji model is up-to-date and cached.
			existing, err := d.RefreshEmoji(ctx, existing, media.AdditionalEmojiInfo{

				// Set the newly fetched image remote URLs.
				ImageRemoteURL:       &emoji.ImageRemoteURL,
				ImageStaticRemoteURL: &emoji.ImageStaticRemoteURL,

				// Refresh is only forced if image URL changed.
			}, (existing.ImageRemoteURL != emoji.ImageRemoteURL))
			if err != nil {
				log.Errorf(ctx, "error refreshing emoji: %v", err)

				// specifically do NOT continue here,
				// we already have a model, we don't
				// want to drop it from the slice, just
				// log that an update for it failed.
			}

			// Set existing emoji.
			emojis[i] = existing
			continue
		}

		// Emojis changed!
		changed = true

		// Fetch this newly added emoji,
		// this function handles the case
		// of existing cached emojis and
		// new ones requiring dereference.
		emoji, err := d.GetEmoji(ctx,
			emoji.Shortcode,
			emoji.Domain,
			emoji.ImageRemoteURL,
			media.AdditionalEmojiInfo{
				ImageRemoteURL:       &emoji.ImageRemoteURL,
				ImageStaticRemoteURL: &emoji.ImageStaticRemoteURL,
			},
			false,
		)
		if err != nil {
			if emoji == nil {
				log.Errorf(ctx, "error loading emoji %s: %v", emoji.ImageRemoteURL, err)
				continue
			}

			// non-fatal error occurred during loading, still use it.
			log.Warnf(ctx, "partially loaded emoji: %v", err)
		}
	}

	for i := 0; i < len(emojis); {
		if emojis[i].ID == "" {
			// Remove failed emoji populations.
			copy(emojis[i:], emojis[i+1:])
			emojis = emojis[:len(emojis)-1]
			continue
		}
		i++
	}

	return emojis, changed, nil
}
