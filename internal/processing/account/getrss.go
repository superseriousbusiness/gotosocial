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

	"github.com/gorilla/feeds"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

const rssFeedLength = 20

func (p *processor) GetRSSFeedForUsername(ctx context.Context, requestingAccount *gtsmodel.Account, username string) (string, gtserror.WithCode) {
	account, err := p.db.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		if err == db.ErrNoEntries {
			return "", gtserror.NewErrorNotFound(errors.New("GetRSSFeedForUsername: account not found"))
		}
		return "", gtserror.NewErrorInternalError(fmt.Errorf("GetRSSFeedForUsername: db error: %s", err))
	}

	if !*account.EnableRSS {
		return "", gtserror.NewErrorNotFound(errors.New("GetRSSFeedForUsername: account RSS feed not enabled"))
	}

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
			Url:   account.AvatarMediaAttachment.URL,
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

		// build description field
		descriptionBuilder := strings.Builder{}
		attachments := len(s.Attachments)
		switch {
		case attachments > 1:
			descriptionBuilder.WriteString(fmt.Sprintf("Posted [%d] attachments", attachments))
		case attachments == 1:
			descriptionBuilder.WriteString("Posted 1 attachment")
			if s.Content != "" {
				descriptionBuilder.WriteString("; " + s.Content)
			}
		default:
			descriptionBuilder.WriteString(s.Content)
		}
		description := trimTo(descriptionBuilder.String(), 256)

		// build title field
		var title string
		if s.ContentWarning != "" {
			title = trimTo(s.ContentWarning, 64)
		} else {
			title = trimTo(description, 64)
		}

		feed.Add(&feeds.Item{
			Id:          s.URL,
			Title:       title,
			Link:        &feeds.Link{Href: s.URL},
			Description: description,
			Author:      feed.Author,
			Created:     s.CreatedAt,
			Content:     s.Content,
		})
	}

	rss, err := feed.ToRss()
	if err != nil {
		return "", gtserror.NewErrorInternalError(fmt.Errorf("GetRSSFeedForUsername: error converting feed to rss string: %s", err))
	}

	return rss, nil
}

func trimTo(in string, to int) string {
	if len(in) <= to {
		return in
	}

	return in[:to]
}
