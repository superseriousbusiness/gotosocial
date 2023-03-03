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

package typeutils

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gorilla/feeds"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

const (
	rssMaxTitleChars       = 128
	rssDescriptionMaxChars = 256
)

func (c *converter) StatusToRSSItem(ctx context.Context, s *gtsmodel.Status) (*feeds.Item, error) {
	// see https://cyber.harvard.edu/rss/rss.html

	// Title -- The title of the item.
	// example: Venice Film Festival Tries to Quit Sinking
	var title string
	if s.ContentWarning != "" {
		title = trimTo(s.ContentWarning, rssMaxTitleChars)
	} else {
		title = trimTo(s.Text, rssMaxTitleChars)
	}

	// Link -- The URL of the item.
	// example: http://nytimes.com/2004/12/07FEST.html
	link := &feeds.Link{
		Href: s.URL,
	}

	// Author -- Email address of the author of the item.
	// example: oprah\@oxygen.net
	if s.Account == nil {
		a, err := c.db.GetAccountByID(ctx, s.AccountID)
		if err != nil {
			return nil, fmt.Errorf("error getting status author: %s", err)
		}
		s.Account = a
	}
	authorName := "@" + s.Account.Username + "@" + config.GetAccountDomain()
	author := &feeds.Author{
		Name: authorName,
	}

	// Source -- The RSS channel that the item came from.
	source := &feeds.Link{
		Href: s.Account.URL + "/feed.rss",
	}

	// Description -- The item synopsis.
	// example: Some of the most heated chatter at the Venice Film Festival this week was about the way that the arrival of the stars at the Palazzo del Cinema was being staged.
	descriptionBuilder := strings.Builder{}
	descriptionBuilder.WriteString(authorName + " ")

	attachmentCount := len(s.Attachments)
	if len(s.AttachmentIDs) > attachmentCount {
		attachmentCount = len(s.AttachmentIDs)
	}
	switch {
	case attachmentCount > 1:
		descriptionBuilder.WriteString(fmt.Sprintf("posted [%d] attachments", attachmentCount))
	case attachmentCount == 1:
		descriptionBuilder.WriteString("posted 1 attachment")
	default:
		descriptionBuilder.WriteString("made a new post")
	}

	if s.Text != "" {
		descriptionBuilder.WriteString(": \"")
		descriptionBuilder.WriteString(s.Text)
		descriptionBuilder.WriteString("\"")
	}

	description := trimTo(descriptionBuilder.String(), rssDescriptionMaxChars)

	// ID -- A string that uniquely identifies the item.
	// example: http://inessential.com/2002/09/01.php#a2
	id := s.URL

	// Enclosure -- Describes a media object that is attached to the item.
	enclosure := &feeds.Enclosure{}
	// get first attachment if present
	var attachment *gtsmodel.MediaAttachment
	if len(s.Attachments) > 0 {
		attachment = s.Attachments[0]
	} else if len(s.AttachmentIDs) > 0 {
		a, err := c.db.GetAttachmentByID(ctx, s.AttachmentIDs[0])
		if err == nil {
			attachment = a
		}
	}
	if attachment != nil {
		enclosure.Type = attachment.File.ContentType
		enclosure.Length = strconv.Itoa(attachment.File.FileSize)
		enclosure.Url = attachment.URL
	}

	// Content
	apiEmojis := []apimodel.Emoji{}
	// the status might already have some gts emojis on it if it's not been pulled directly from the database
	// if so, we can directly convert the gts emojis into api ones
	if s.Emojis != nil {
		for _, gtsEmoji := range s.Emojis {
			apiEmoji, err := c.EmojiToAPIEmoji(ctx, gtsEmoji)
			if err != nil {
				log.Errorf(ctx, "error converting emoji with id %s: %s", gtsEmoji.ID, err)
				continue
			}
			apiEmojis = append(apiEmojis, apiEmoji)
		}
		// the status doesn't have gts emojis on it, but it does have emoji IDs
		// in this case, we need to pull the gts emojis from the db to convert them into api ones
	} else {
		for _, e := range s.EmojiIDs {
			gtsEmoji := &gtsmodel.Emoji{}
			if err := c.db.GetByID(ctx, e, gtsEmoji); err != nil {
				log.Errorf(ctx, "error getting emoji with id %s: %s", e, err)
				continue
			}
			apiEmoji, err := c.EmojiToAPIEmoji(ctx, gtsEmoji)
			if err != nil {
				log.Errorf(ctx, "error converting emoji with id %s: %s", gtsEmoji.ID, err)
				continue
			}
			apiEmojis = append(apiEmojis, apiEmoji)
		}
	}
	content := text.Emojify(apiEmojis, s.Content)

	return &feeds.Item{
		Title:       title,
		Link:        link,
		Author:      author,
		Source:      source,
		Description: description,
		Id:          id,
		Updated:     s.UpdatedAt,
		Created:     s.CreatedAt,
		Enclosure:   enclosure,
		Content:     content,
	}, nil
}

func trimTo(in string, to int) string {
	if len(in) <= to {
		return in
	}

	return in[:to-3] + "..."
}
