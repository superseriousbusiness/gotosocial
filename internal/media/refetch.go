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
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type DereferenceMedia func(ctx context.Context, iri *url.URL, maxsz int64) (io.ReadCloser, error)

// RefetchEmojis iterates through remote emojis (for the given domain, or all if domain is empty string).
//
// For each emoji, the manager will check whether both the full size and static images are present in storage.
// If not, the manager will refetch and reprocess full size and static images for the emoji.
//
// The provided DereferenceMedia function will be used when it's necessary to refetch something this way.
func (m *Manager) RefetchEmojis(ctx context.Context, domain string, dereferenceMedia DereferenceMedia) (int, error) {
	// normalize domain
	if domain == "" {
		domain = db.EmojiAllDomains
	}

	var (
		maxShortcodeDomain string
		refetchIDs         []string
	)

	// page through emojis 20 at a time, looking for those with missing images
	for {
		// Fetch next block of emojis from database
		emojis, err := m.state.DB.GetEmojisBy(ctx, domain, false, true, "", maxShortcodeDomain, "", 20)
		if err != nil {
			if !errors.Is(err, db.ErrNoEntries) {
				log.Errorf(ctx, "error fetching emojis from database: %s", err)
			}
			break
		}

		for _, emoji := range emojis {
			if emoji.IsLocal() {
				// never try to refetch local emojis
				continue
			}

			if refetch, err := m.emojiRequiresRefetch(ctx, emoji); err != nil {
				// an error here indicates something is wrong with storage, so we should stop
				return 0, fmt.Errorf("error checking refetch requirement for emoji %s: %w", emoji.ShortcodeDomain(), err)
			} else if !refetch {
				continue
			}

			refetchIDs = append(refetchIDs, emoji.ID)
		}

		// Update next maxShortcodeDomain from last emoji
		maxShortcodeDomain = emojis[len(emojis)-1].ShortcodeDomain()
	}

	// bail early if we've got nothing to do
	toRefetchCount := len(refetchIDs)
	if toRefetchCount == 0 {
		log.Debug(ctx, "no remote emojis require a refetch")
		return 0, nil
	}
	log.Debugf(ctx, "%d remote emoji(s) require a refetch, doing that now...", toRefetchCount)

	var totalRefetched int
	for _, emojiID := range refetchIDs {
		emoji, err := m.state.DB.GetEmojiByID(ctx, emojiID)
		if err != nil {
			// this shouldn't happen--since we know we have the emoji--so return if it does
			return 0, fmt.Errorf("error getting emoji %s: %w", emojiID, err)
		}
		shortcodeDomain := emoji.ShortcodeDomain()

		if emoji.ImageRemoteURL == "" {
			log.Errorf(ctx, "remote emoji %s could not be refreshed because it has no ImageRemoteURL set", shortcodeDomain)
			continue
		}

		emojiImageIRI, err := url.Parse(emoji.ImageRemoteURL)
		if err != nil {
			log.Errorf(ctx, "remote emoji %s could not be refreshed because its ImageRemoteURL (%s) is not a valid uri: %s", shortcodeDomain, emoji.ImageRemoteURL, err)
			continue
		}

		// Get max supported remote emoji media size.
		maxsz := int64(config.GetMediaEmojiRemoteMaxSize()) // #nosec G115 -- Already validated.
		dataFunc := func(ctx context.Context) (reader io.ReadCloser, err error) {
			return dereferenceMedia(ctx, emojiImageIRI, maxsz)
		}

		processingEmoji, err := m.UpdateEmoji(ctx, emoji, dataFunc, AdditionalEmojiInfo{
			Domain:               &emoji.Domain,
			ImageRemoteURL:       &emoji.ImageRemoteURL,
			ImageStaticRemoteURL: &emoji.ImageStaticRemoteURL,
			Disabled:             emoji.Disabled,
			VisibleInPicker:      emoji.VisibleInPicker,
		})
		if err != nil {
			log.Errorf(ctx, "emoji %s could not be updated because of an error during processing: %s", shortcodeDomain, err)
			continue
		}

		if _, err := processingEmoji.Load(ctx); err != nil {
			log.Errorf(ctx, "emoji %s could not be updated because of an error during loading: %s", shortcodeDomain, err)
			continue
		}

		log.Tracef(ctx, "refetched + updated emoji %s successfully from remote", shortcodeDomain)
		totalRefetched++
	}

	return totalRefetched, nil
}

func (m *Manager) emojiRequiresRefetch(ctx context.Context, emoji *gtsmodel.Emoji) (bool, error) {
	if has, err := m.state.Storage.Has(ctx, emoji.ImagePath); err != nil {
		return false, err
	} else if !has {
		return true, nil
	}

	if has, err := m.state.Storage.Has(ctx, emoji.ImageStaticPath); err != nil {
		return false, err
	} else if !has {
		return true, nil
	}

	return false, nil
}
