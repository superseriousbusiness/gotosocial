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
)

const rssFeedLength = 20

func (p *processor) GetRSSFeedForUsername(ctx context.Context, username string) (func() (string, gtserror.WithCode), time.Time, gtserror.WithCode) {
	account, err := p.db.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, time.Time{}, gtserror.NewErrorNotFound(errors.New("GetRSSFeedForUsername: account not found"))
		}
		return nil, time.Time{}, gtserror.NewErrorInternalError(fmt.Errorf("GetRSSFeedForUsername: db error: %s", err))
	}

	if !*account.EnableRSS {
		return nil, time.Time{}, gtserror.NewErrorNotFound(errors.New("GetRSSFeedForUsername: account RSS feed not enabled"))
	}

	lastModified, err := p.db.GetAccountLastPosted(ctx, account.ID, true)
	if err != nil {
		return nil, time.Time{}, gtserror.NewErrorInternalError(fmt.Errorf("GetRSSFeedForUsername: db error: %s", err))
	}

	return func() (string, gtserror.WithCode) {
		statuses, err := p.db.GetAccountWebStatuses(ctx, account.ID, rssFeedLength, "")
		if err != nil && err != db.ErrNoEntries {
			return "", gtserror.NewErrorInternalError(fmt.Errorf("GetRSSFeedForUsername: db error: %s", err))
		}

		author := "@" + account.Username + "@" + config.GetAccountDomain()
		title := "Posts from " + author
		description := "Posts from " + author
		link := &feeds.Link{Href: account.URL}

		var image *feeds.Image
		if account.AvatarMediaAttachmentID != "" {
			if account.AvatarMediaAttachment == nil {
				avatar, err := p.db.GetAttachmentByID(ctx, account.AvatarMediaAttachmentID)
				if err != nil {
					return "", gtserror.NewErrorInternalError(fmt.Errorf("GetRSSFeedForUsername: db error fetching avatar attachment: %s", err))
				}
				account.AvatarMediaAttachment = avatar
			}
			image = &feeds.Image{
				Url:   account.AvatarMediaAttachment.Thumbnail.URL,
				Title: "Avatar for " + author,
				Link:  account.URL,
			}
		}

		feed := &feeds.Feed{
			Title:       title,
			Description: description,
			Link:        link,
			Image:       image,
		}

		for i, s := range statuses {
			// take the date of the first (ie., latest) status as feed updated value
			if i == 0 {
				feed.Updated = s.UpdatedAt
			}

			item, err := p.tc.StatusToRSSItem(ctx, s)
			if err != nil {
				return "", gtserror.NewErrorInternalError(fmt.Errorf("GetRSSFeedForUsername: error converting status to feed item: %s", err))
			}

			feed.Add(item)
		}

		rss, err := feed.ToRss()
		if err != nil {
			return "", gtserror.NewErrorInternalError(fmt.Errorf("GetRSSFeedForUsername: error converting feed to rss string: %s", err))
		}

		return rss, nil
	}, lastModified, nil
}
