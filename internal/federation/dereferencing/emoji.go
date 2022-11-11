/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

func (d *deref) GetRemoteEmoji(ctx context.Context, requestingUsername string, remoteURL string, shortcode string, id string, emojiURI string, ai *media.AdditionalEmojiInfo, refresh bool) (*media.ProcessingEmoji, error) {
	t, err := d.transportController.NewTransportForUsername(ctx, requestingUsername)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteEmoji: error creating transport: %s", err)
	}

	derefURI, err := url.Parse(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteEmoji: error parsing url: %s", err)
	}

	dataFunc := func(innerCtx context.Context) (io.ReadCloser, int64, error) {
		return t.DereferenceMedia(innerCtx, derefURI)
	}

	processingMedia, err := d.mediaManager.ProcessEmoji(ctx, dataFunc, nil, shortcode, id, emojiURI, ai, refresh)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteEmoji: error processing emoji: %s", err)
	}

	return processingMedia, nil
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

		// check if we've already got this emoji in the db
		if gotEmoji, err = d.db.GetEmojiByShortcodeDomain(ctx, e.Shortcode, e.Domain); err != nil && err != db.ErrNoEntries {
			log.Errorf("populateEmojis: error checking database for emoji %s: %s", e.URI, err)
			continue
		}

		if gotEmoji != nil {
			// we had the emoji in our database already; make sure the one we have is up to date
			if (e.UpdatedAt.After(gotEmoji.ImageUpdatedAt)) || (e.URI != gotEmoji.URI) || (e.ImageRemoteURL != gotEmoji.ImageRemoteURL) {
				emojiID := gotEmoji.ID // use existing ID
				processingEmoji, err := d.GetRemoteEmoji(ctx, requestingUsername, e.ImageRemoteURL, e.Shortcode, emojiID, e.URI, &media.AdditionalEmojiInfo{
					Domain:               &e.Domain,
					ImageRemoteURL:       &e.ImageRemoteURL,
					ImageStaticRemoteURL: &e.ImageStaticRemoteURL,
					Disabled:             gotEmoji.Disabled,
					VisibleInPicker:      gotEmoji.VisibleInPicker,
				}, true)

				if err != nil {
					log.Errorf("populateEmojis: couldn't refresh remote emoji %s: %s", e.URI, err)
					continue
				}

				if gotEmoji, err = processingEmoji.LoadEmoji(ctx); err != nil {
					log.Errorf("populateEmojis: couldn't load refreshed remote emoji %s: %s", e.URI, err)
					continue
				}
			}
		} else {
			// it's new! go get it!
			newEmojiID, err := id.NewRandomULID()
			if err != nil {
				log.Errorf("populateEmojis: error generating id for remote emoji %s: %s", e.URI, err)
				continue
			}

			processingEmoji, err := d.GetRemoteEmoji(ctx, requestingUsername, e.ImageRemoteURL, e.Shortcode, newEmojiID, e.URI, &media.AdditionalEmojiInfo{
				Domain:               &e.Domain,
				ImageRemoteURL:       &e.ImageRemoteURL,
				ImageStaticRemoteURL: &e.ImageStaticRemoteURL,
				Disabled:             e.Disabled,
				VisibleInPicker:      e.VisibleInPicker,
			}, false)

			if err != nil {
				log.Errorf("populateEmojis: couldn't get remote emoji %s: %s", e.URI, err)
				continue
			}

			if gotEmoji, err = processingEmoji.LoadEmoji(ctx); err != nil {
				log.Errorf("populateEmojis: couldn't load remote emoji %s: %s", e.URI, err)
				continue
			}
		}

		// if we get here, we either had the emoji already or we successfully fetched it
		gotEmojis = append(gotEmojis, gotEmoji)
	}

	return gotEmojis, nil
}
