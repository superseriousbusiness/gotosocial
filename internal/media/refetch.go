/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type DereferenceMedia func(ctx context.Context, iri *url.URL) (io.ReadCloser, int64, error)

func (m *manager) RefetchEmojis(ctx context.Context, domain string, dereferenceMedia DereferenceMedia) (int, error) {
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
		emojis, err := m.state.DB.GetEmojis(ctx, domain, false, true, "", maxShortcodeDomain, "", 20)
		if err != nil {
			if !errors.Is(err, db.ErrNoEntries) {
				// an actual error has occurred
				log.Errorf(ctx, "error fetching emojis from database: %s", err)
			}
			break
		}

		for _, emoji := range emojis {
			if emoji.Domain == "" {
				// never try to refetch local emojis
				continue
			}

			if refetch, err := m.emojiRequiresRefetch(ctx, emoji); err != nil {
				// an error here indicates something is wrong with storage, so we should stop
				return 0, fmt.Errorf("error checking refetch requirement for emoji %s: %w", util.ShortcodeDomain(emoji), err)
			} else if !refetch {
				continue
			}

			refetchIDs = append(refetchIDs, emoji.ID)
		}

		// Update next maxShortcodeDomain from last emoji
		maxShortcodeDomain = util.ShortcodeDomain(emojis[len(emojis)-1])
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
		shortcodeDomain := util.ShortcodeDomain(emoji)

		if emoji.ImageRemoteURL == "" {
			log.Errorf(ctx, "remote emoji %s could not be refreshed because it has no ImageRemoteURL set", shortcodeDomain)
			continue
		}

		emojiImageIRI, err := url.Parse(emoji.ImageRemoteURL)
		if err != nil {
			log.Errorf(ctx, "remote emoji %s could not be refreshed because its ImageRemoteURL (%s) is not a valid uri: %s", shortcodeDomain, emoji.ImageRemoteURL, err)
			continue
		}

		dataFunc := func(ctx context.Context) (reader io.ReadCloser, fileSize int64, err error) {
			return dereferenceMedia(ctx, emojiImageIRI)
		}

		processingEmoji, err := m.PreProcessEmoji(ctx, dataFunc, nil, emoji.Shortcode, emoji.ID, emoji.URI, &AdditionalEmojiInfo{
			Domain:               &emoji.Domain,
			ImageRemoteURL:       &emoji.ImageRemoteURL,
			ImageStaticRemoteURL: &emoji.ImageStaticRemoteURL,
			Disabled:             emoji.Disabled,
			VisibleInPicker:      emoji.VisibleInPicker,
		}, true)
		if err != nil {
			log.Errorf(ctx, "emoji %s could not be refreshed because of an error during processing: %s", shortcodeDomain, err)
			continue
		}

		if _, err := processingEmoji.LoadEmoji(ctx); err != nil {
			log.Errorf(ctx, "emoji %s could not be refreshed because of an error during loading: %s", shortcodeDomain, err)
			continue
		}

		log.Tracef(ctx, "refetched emoji %s successfully from remote", shortcodeDomain)
		totalRefetched++
	}

	return totalRefetched, nil
}

func (m *manager) emojiRequiresRefetch(ctx context.Context, emoji *gtsmodel.Emoji) (bool, error) {
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
