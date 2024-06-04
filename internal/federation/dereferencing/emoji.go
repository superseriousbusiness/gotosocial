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
	"fmt"
	"io"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (d *Dereferencer) GetRemoteEmoji(ctx context.Context, requestUser string, remoteURL string, shortcode string, domain string, id string, emojiURI string, ai *media.AdditionalEmojiInfo, refresh bool) (*media.ProcessingEmoji, error) {
	var shortcodeDomain = shortcode + "@" + domain

	// Ensure we have been passed a valid URL.
	derefURI, err := url.Parse(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteEmoji: error parsing url for emoji %s: %s", shortcodeDomain, err)
	}

	// Acquire derefs lock.
	d.derefEmojisMu.Lock()

	// Ensure unlock only done once.
	unlock := d.derefEmojisMu.Unlock
	unlock = util.DoOnce(unlock)
	defer unlock()

	// Look for an existing dereference in progress.
	processing, ok := d.derefEmojis[shortcodeDomain]

	if !ok {
		// Fetch a transport for current request user in order to perform request.
		tsport, err := d.transportController.NewTransportForUsername(ctx, requestUser)
		if err != nil {
			return nil, gtserror.Newf("couldn't create transport: %w", err)
		}

		// Set the media data function to dereference emoji from URI.
		data := func(ctx context.Context) (io.ReadCloser, int64, error) {
			return tsport.DereferenceMedia(ctx, derefURI)
		}

		// Create new emoji processing request from the media manager.
		processing, err = d.mediaManager.PreProcessEmoji(ctx, data,
			shortcode,
			id,
			emojiURI,
			ai,
			refresh,
		)
		if err != nil {
			return nil, gtserror.Newf("error preprocessing emoji %s: %s", shortcodeDomain, err)
		}

		// Store media in map to mark as processing.
		d.derefEmojis[shortcodeDomain] = processing

		defer func() {
			// On exit safely remove emoji from map.
			d.derefEmojisMu.Lock()
			delete(d.derefEmojis, shortcodeDomain)
			d.derefEmojisMu.Unlock()
		}()
	}

	// Unlock map.
	unlock()

	// Start emoji attachment loading (blocking call).
	if _, err := processing.LoadEmoji(ctx); err != nil {
		return nil, err
	}

	return processing, nil
}

func (d *Dereferencer) populateEmojis(ctx context.Context, rawEmojis []*gtsmodel.Emoji, requestingUsername string) ([]*gtsmodel.Emoji, error) {
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
		} else if gotEmoji, err = d.state.DB.GetEmojiByShortcodeDomain(ctx, e.Shortcode, e.Domain); err != nil && err != db.ErrNoEntries {
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
