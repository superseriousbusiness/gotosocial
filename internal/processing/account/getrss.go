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

package account

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/text"
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

		var feedTitle string
		if account.DisplayName != "" {
			feedTitle = account.DisplayName
		} else {
			feedTitle = account.Username
		}

		author := "@" + account.Username + "@" + config.GetAccountDomain()

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
			Title:       feedTitle,
			Description: "Public posts from " + author,
			Link:        &feeds.Link{Href: account.URL},
			Image:       image,
		}

		feed.Items = []*feeds.Item{}
		for i, s := range statuses {
			// take the date of the first (ie., latest) status as feed updated value
			if i == 0 {
				feed.Updated = s.UpdatedAt
			}

			apiStatus, err := p.tc.StatusToAPIStatus(ctx, s, nil)
			if err != nil {
				return "", gtserror.NewErrorInternalError(fmt.Errorf("GetRSSFeedForUsername: error converting status to api model: %s", err))
			}

			// build description field
			descriptionBuilder := strings.Builder{}
			attachments := len(apiStatus.MediaAttachments)
			switch {
			case attachments > 1:
				descriptionBuilder.WriteString(fmt.Sprintf("Posted [%d] attachments", attachments))
			case attachments == 1:
				descriptionBuilder.WriteString("Posted 1 attachment")
			default:
				descriptionBuilder.WriteString("Made a new post")
			}
			description := trimTo(descriptionBuilder.String(), 256)

			// build title field
			var title string
			if apiStatus.SpoilerText != "" {
				title = trimTo(apiStatus.SpoilerText, 64)
			} else {
				title = trimTo(s.Text, 64)
			}

			feed.Add(&feeds.Item{
				Id:          apiStatus.URL,
				Title:       title,
				Link:        &feeds.Link{Href: apiStatus.URL},
				Description: description,
				Author:      feed.Author,
				Created:     s.CreatedAt,
				Content:     text.Emojify(apiStatus.Emojis, apiStatus.Content),
			})
		}

		rss, err := feed.ToRss()
		if err != nil {
			return "", gtserror.NewErrorInternalError(fmt.Errorf("GetRSSFeedForUsername: error converting feed to rss string: %s", err))
		}

		return rss, nil
	}, lastModified, nil
}

func trimTo(in string, to int) string {
	if len(in) <= to {
		return in
	}

	return in[:to]
}
