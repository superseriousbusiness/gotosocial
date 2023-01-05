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

package web

import (
	"html"
	"strconv"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

const maxOGDescriptionLength = 300

// ogMeta represents supported OpenGraph Meta tags
//
// see eg https://developer.yoast.com/features/opengraph/functional-specification/
type ogMeta struct {
	// vanilla og tags

	Locale       string // og:locale
	ResourceType string // og:type
	Title        string // og:title
	URL          string // og:url
	SiteName     string // og:site_name
	Description  string // og:description
	Image        string // og:image
	ImageWidth   string // og:image:width
	ImageHeight  string // og:image:height

	// article tags

	ArticlePublisher     string // article:publisher
	ArticleAuthor        string // article:author
	ArticleModifiedTime  string // article:modified_time
	ArticlePublishedTime string // article:published_time
}

// ogBase returns an *ogMeta suitable for serving at
// the base root of an instance. It also serves as a
// foundation for building account / status ogMeta on
// top of.
func ogBase(instance *apimodel.Instance) *ogMeta {
	var locale string
	if len(instance.Languages) > 0 {
		locale = instance.Languages[0]
	}

	og := &ogMeta{
		Locale:       locale,
		ResourceType: "website",
		Title:        text.SanitizePlaintext(instance.Title) + " - GoToSocial",
		URL:          instance.URI,
		SiteName:     instance.AccountDomain,
		Description:  parseDescription(instance.ShortDescription),
		Image:        instance.Thumbnail,
	}

	return og
}

// withAccount uses the given account to build an ogMeta
// struct specific to that account. It's suitable for serving
// at account profile pages.
func (og *ogMeta) withAccount(account *apimodel.Account) *ogMeta {
	og.ResourceType = "profile"
	og.Title = parseTitle(account, og.SiteName)
	og.URL = account.URL
	og.Description = parseDescription(account.Note)
	og.Image = account.Avatar
	return og
}

// withStatus uses the given status to build an ogMeta
// struct specific to that status. It's suitable for serving
// at status pages.
func (og *ogMeta) withStatus(status *apimodel.Status) *ogMeta {
	if !status.Sensitive && len(status.MediaAttachments) > 0 {
		a := status.MediaAttachments[0]
		og.Image = a.PreviewURL
		og.ImageWidth = strconv.Itoa(a.Meta.Small.Width)
		og.ImageHeight = strconv.Itoa(a.Meta.Small.Height)
	} else {
		og.Image = status.Account.Avatar
	}

	if status.SpoilerText != "" {
		og.Description = parseDescription("CW: " + status.SpoilerText)
	} else {
		og.Description = parseDescription(status.Text)
	}

	if status.Language != nil {
		og.Locale = *status.Language
	}
	og.ResourceType = "article"
	og.Title = "Post by " + parseTitle(status.Account, og.SiteName)
	og.URL = status.URL
	og.ArticlePublisher = status.Account.URL
	og.ArticleAuthor = status.Account.URL
	og.ArticlePublishedTime = status.CreatedAt
	og.ArticleModifiedTime = status.CreatedAt
	return og
}

// parseTitle parses a page title from account and accountDomain
func parseTitle(account *apimodel.Account, accountDomain string) string {
	user := "@" + account.Acct + "@" + accountDomain

	if len(account.DisplayName) == 0 {
		return user
	}

	return account.DisplayName + " (" + user + ")"
}

// parseDescription returns a string description which is
// safe to use as a template.HTMLAttr inside templates.
func parseDescription(in string) string {
	i := html.UnescapeString(in)
	i = text.SanitizePlaintext(i)
	i = strings.ReplaceAll(i, "\"", "'")
	i = strings.ReplaceAll(i, `\`, "")
	i = strings.ReplaceAll(i, "\n", " ")
	i = trim(i, maxOGDescriptionLength)
	return `content="` + i + `"`
}

// trim strings trim s to specified length
func trim(s string, length int) string {
	if len(s) < length {
		return s
	}

	return s[:length]
}
