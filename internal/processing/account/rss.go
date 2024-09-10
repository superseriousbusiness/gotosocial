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

package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/feeds"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

const (
	rssFeedLength = 20
)

type GetRSSFeed func() (string, gtserror.WithCode)

// GetRSSFeedForUsername returns a function to return the RSS feed of a local account
// with the given username, and the last-modified time (time that the account last
// posted a status eligible to be included in the rss feed).
//
// To save db calls, callers to this function should only call the returned GetRSSFeed
// func if the last-modified time is newer than the last-modified time they have cached.
//
// If the account has not yet posted an RSS-eligible status, the returned last-modified
// time will be zero, and the GetRSSFeed func will return a valid RSS xml with no items.
func (p *Processor) GetRSSFeedForUsername(ctx context.Context, username string) (GetRSSFeed, time.Time, gtserror.WithCode) {
	var (
		never = time.Time{}
	)

	account, err := p.state.DB.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// Simply no account with this username.
			err = gtserror.New("account not found")
			return nil, never, gtserror.NewErrorNotFound(err)
		}

		// Real db error.
		err = gtserror.Newf("db error getting account %s: %w", username, err)
		return nil, never, gtserror.NewErrorInternalError(err)
	}

	// Ensure account has rss feed enabled.
	if !*account.Settings.EnableRSS {
		err = gtserror.New("account RSS feed not enabled")
		return nil, never, gtserror.NewErrorNotFound(err)
	}

	// Ensure account stats populated.
	if err := p.state.DB.PopulateAccountStats(ctx, account); err != nil {
		err = gtserror.Newf("db error getting account stats %s: %w", username, err)
		return nil, never, gtserror.NewErrorInternalError(err)
	}

	// LastModified time is needed by callers to check freshness for cacheing.
	// This might be a zero time.Time if account has never posted a status that's
	// eligible to appear in the RSS feed; that's fine.
	lastPostAt := account.Stats.LastStatusAt

	return func() (string, gtserror.WithCode) {
		// Assemble author namestring once only.
		author := "@" + account.Username + "@" + config.GetAccountDomain()

		// Derive image/thumbnail for this account (may be nil).
		image, errWithCode := p.rssImageForAccount(ctx, account, author)
		if errWithCode != nil {
			return "", errWithCode
		}

		feed := &feeds.Feed{
			Title:       "Posts from " + author,
			Description: "Posts from " + author,
			Link:        &feeds.Link{Href: account.URL},
			Image:       image,
		}

		// If the account has never posted anything, just use
		// account creation time as Updated value for the feed;
		// we could use time.Now() here but this would likely
		// mess up cacheing; we want something determinate.
		//
		// We can also return early rather than wasting a db call,
		// since we already know there's no eligible statuses.
		if lastPostAt.IsZero() {
			feed.Updated = account.CreatedAt
			return stringifyFeed(feed)
		}

		// Account has posted at least one status that's
		// eligible to appear in the RSS feed.
		//
		// Reuse the lastPostAt value for feed.Updated.
		feed.Updated = lastPostAt

		// Retrieve latest statuses as they'd be shown on the web view of the account profile.
		statuses, err := p.state.DB.GetAccountWebStatuses(ctx, account, rssFeedLength, "")
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("db error getting account web statuses: %w", err)
			return "", gtserror.NewErrorInternalError(err)
		}

		// Add each status to the rss feed.
		for _, status := range statuses {
			item, err := p.converter.StatusToRSSItem(ctx, status)
			if err != nil {
				err = gtserror.Newf("error converting status to feed item: %w", err)
				return "", gtserror.NewErrorInternalError(err)
			}

			feed.Add(item)
		}

		return stringifyFeed(feed)
	}, lastPostAt, nil
}

func (p *Processor) rssImageForAccount(ctx context.Context, account *gtsmodel.Account, author string) (*feeds.Image, gtserror.WithCode) {
	if account.AvatarMediaAttachmentID == "" {
		// No image, no problem!
		return nil, nil
	}

	// Ensure account avatar attachment populated.
	if account.AvatarMediaAttachment == nil {
		var err error
		account.AvatarMediaAttachment, err = p.state.DB.GetAttachmentByID(ctx, account.AvatarMediaAttachmentID)
		if err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				// No attachment found with this ID (race condition?).
				return nil, nil
			}

			// Real db error.
			err = gtserror.Newf("db error fetching avatar media attachment: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	return &feeds.Image{
		Url:   account.AvatarMediaAttachment.Thumbnail.URL,
		Title: "Avatar for " + author,
		Link:  account.URL,
	}, nil
}

func stringifyFeed(feed *feeds.Feed) (string, gtserror.WithCode) {
	// Stringify the feed. Even with no statuses,
	// this will still produce valid rss xml.
	rss, err := feed.ToRss()
	if err != nil {
		err := gtserror.Newf("error converting feed to rss string: %w", err)
		return "", gtserror.NewErrorInternalError(err)
	}

	return rss, nil
}
