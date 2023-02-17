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

package dereferencing

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

func (d *deref) GetRemoteEmoji(ctx context.Context, requestingUsername string, remoteURL string, shortcode string, domain string, id string, emojiURI string, ai *media.AdditionalEmojiInfo, refresh bool) (*media.ProcessingEmoji, error) {
	var (
		shortcodeDomain = shortcode + "@" + domain
		processingEmoji *media.ProcessingEmoji
	)

	// Acquire lock for derefs map.
	unlock := d.derefEmojisMu.Lock()
	defer unlock()

	// first check if we're already processing this emoji
	if alreadyProcessing, ok := d.derefEmojis[shortcodeDomain]; ok {
		// we're already on it, no worries
		processingEmoji = alreadyProcessing
	} else {
		// not processing it yet, let's start
		t, err := d.transportController.NewTransportForUsername(ctx, requestingUsername)
		if err != nil {
			return nil, fmt.Errorf("GetRemoteEmoji: error creating transport to fetch emoji %s: %s", shortcodeDomain, err)
		}

		derefURI, err := url.Parse(remoteURL)
		if err != nil {
			return nil, fmt.Errorf("GetRemoteEmoji: error parsing url for emoji %s: %s", shortcodeDomain, err)
		}

		dataFunc := func(innerCtx context.Context) (io.ReadCloser, int64, error) {
			return t.DereferenceMedia(innerCtx, derefURI)
		}

		newProcessing, err := d.mediaManager.PreProcessEmoji(ctx, dataFunc, nil, shortcode, id, emojiURI, ai, refresh)
		if err != nil {
			return nil, fmt.Errorf("GetRemoteEmoji: error processing emoji %s: %s", shortcodeDomain, err)
		}

		// store it in our map to indicate it's in process
		d.derefEmojis[shortcodeDomain] = newProcessing
		processingEmoji = newProcessing
	}

	// Unlock map.
	unlock()

	defer func() {
		// On exit safely remove emoji from map.
		unlock := d.derefEmojisMu.Lock()
		delete(d.derefEmojis, shortcodeDomain)
		unlock()
	}()

	// Start emoji attachment loading (blocking call).
	if _, err := processingEmoji.LoadEmoji(ctx); err != nil {
		return nil, err
	}

	return processingEmoji, nil
}

func (d *deref) populateEmojis(ctx context.Context, rawEmojis []*gtsmodel.Emoji, requestingUsername string) ([]*gtsmodel.Emoji, error) {
	// At this point we should know:
	// * the AP uri of the emoji
	// * the domain of the emoji
	// * the shortcode of the emoji
	// * the remote URL of the image
	// This should be enough to dereference the emoji

	gotEmojis := make([]*gtsmodel.Emoji, 0, len(rawEmojis))

	for _, e := range rawEmojis {
		var gotEmoji *gtsmodel.Emoji
		var err error
		shortcodeDomain := e.Shortcode + "@" + e.Domain

		// check if we already know this emoji
		if e.ID != "" {
			// we had an ID for this emoji already, which means
			// it should be fleshed out already and we won't
			// have to get it from the database again
			gotEmoji = e
		} else if gotEmoji, err = d.db.GetEmojiByShortcodeDomain(ctx, e.Shortcode, e.Domain); err != nil && err != db.ErrNoEntries {
			log.Errorf(ctx, "error checking database for emoji %s: %s", shortcodeDomain, err)
			continue
		}

		var refresh bool

		if gotEmoji != nil {
			// we had the emoji already, but refresh it if necessary
			if e.UpdatedAt.Unix() > gotEmoji.ImageUpdatedAt.Unix() {
				log.Tracef(ctx, "emoji %s was updated since we last saw it, will refresh", shortcodeDomain)
				refresh = true
			}

			if !refresh && (e.URI != gotEmoji.URI) {
				log.Tracef(ctx, "emoji %s changed URI since we last saw it, will refresh", shortcodeDomain)
				refresh = true
			}

			if !refresh && (e.ImageRemoteURL != gotEmoji.ImageRemoteURL) {
				log.Tracef(ctx, "emoji %s changed image URL since we last saw it, will refresh", shortcodeDomain)
				refresh = true
			}

			if !refresh {
				log.Tracef(ctx, "emoji %s is up to date, will not refresh", shortcodeDomain)
			} else {
				log.Tracef(ctx, "refreshing emoji %s", shortcodeDomain)
				emojiID := gotEmoji.ID // use existing ID
				processingEmoji, err := d.GetRemoteEmoji(ctx, requestingUsername, e.ImageRemoteURL, e.Shortcode, e.Domain, emojiID, e.URI, &media.AdditionalEmojiInfo{
					Domain:               &e.Domain,
					ImageRemoteURL:       &e.ImageRemoteURL,
					ImageStaticRemoteURL: &e.ImageStaticRemoteURL,
					Disabled:             gotEmoji.Disabled,
					VisibleInPicker:      gotEmoji.VisibleInPicker,
				}, refresh)
				if err != nil {
					log.Errorf(ctx, "couldn't refresh remote emoji %s: %s", shortcodeDomain, err)
					continue
				}

				if gotEmoji, err = processingEmoji.LoadEmoji(ctx); err != nil {
					log.Errorf(ctx, "couldn't load refreshed remote emoji %s: %s", shortcodeDomain, err)
					continue
				}
			}
		} else {
			// it's new! go get it!
			newEmojiID, err := id.NewRandomULID()
			if err != nil {
				log.Errorf(ctx, "error generating id for remote emoji %s: %s", shortcodeDomain, err)
				continue
			}

			processingEmoji, err := d.GetRemoteEmoji(ctx, requestingUsername, e.ImageRemoteURL, e.Shortcode, e.Domain, newEmojiID, e.URI, &media.AdditionalEmojiInfo{
				Domain:               &e.Domain,
				ImageRemoteURL:       &e.ImageRemoteURL,
				ImageStaticRemoteURL: &e.ImageStaticRemoteURL,
				Disabled:             e.Disabled,
				VisibleInPicker:      e.VisibleInPicker,
			}, refresh)
			if err != nil {
				log.Errorf(ctx, "couldn't get remote emoji %s: %s", shortcodeDomain, err)
				continue
			}

			if gotEmoji, err = processingEmoji.LoadEmoji(ctx); err != nil {
				log.Errorf(ctx, "couldn't load remote emoji %s: %s", shortcodeDomain, err)
				continue
			}
		}

		// if we get here, we either had the emoji already or we successfully fetched it
		gotEmojis = append(gotEmojis, gotEmoji)
	}

	return gotEmojis, nil
}
