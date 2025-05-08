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

package util

import (
	"html"
	"slices"
	"strconv"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/text"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// OGMeta represents supported OpenGraph Meta tags
//
// see eg https://ogp.me/
type OGMeta struct {
	/* Vanilla og tags */

	Title       string // og:title
	Type        string // og:type
	Locale      string // og:locale
	URL         string // og:url
	SiteName    string // og:site_name
	Description string // og:description

	// Zero or more media entries of type image,
	// video, or audio (https://ogp.me/#array).
	Media []OGMedia

	/* Article tags. */

	ArticlePublisher     string // article:publisher
	ArticleAuthor        string // article:author
	ArticleModifiedTime  string // article:modified_time
	ArticlePublishedTime string // article:published_time

	/* Profile tags. */

	ProfileUsername string // profile:username

	/*
		Twitter card stuff
		https://developer.twitter.com/en/docs/twitter-for-websites/cards/overview/abouts-cards
	*/

	// Set to media URL for media posts.
	TwitterSummaryLargeImage string
	TwitterImageAlt          string
}

func (o *OGMeta) prependMedia(i ...OGMedia) {
	if len(o.Media) == 0 {
		// Set as
		// only entries.
		o.Media = i
	} else {
		// Prepend as higher
		// priority entries.
		o.Media = slices.Insert(o.Media, 0, i...)
	}
}

// OGMedia represents one OpenGraph media
// entry of type image, video, or audio.
type OGMedia struct {
	OGType   string // image/video/audio
	URL      string // og:${type}
	MIMEType string // og:${type}:type
	Width    string // og:${type}:width
	Height   string // og:${type}:height
	Alt      string // og:${type}:alt
}

// OGBase returns an *ogMeta suitable for serving at
// the base root of an instance. It also serves as a
// foundation for building account / status ogMeta.
func OGBase(instance *apimodel.InstanceV1) *OGMeta {
	var locale string
	if len(instance.Languages) > 0 {
		locale = instance.Languages[0]
	}

	og := &OGMeta{
		Title:       text.StripHTMLFromText(instance.Title) + " - GoToSocial",
		Type:        "website",
		Locale:      locale,
		URL:         instance.URI,
		SiteName:    instance.AccountDomain,
		Description: ParseDescription(instance.ShortDescription),
		Media: []OGMedia{
			{
				OGType:   "image",
				URL:      instance.Thumbnail,
				Alt:      instance.ThumbnailDescription,
				MIMEType: instance.ThumbnailType,
			},
		},
	}

	return og
}

// WithAccount uses the given account to build an ogMeta
// struct specific to that account. It's suitable for serving
// at account profile pages.
func (o *OGMeta) WithAccount(acct *apimodel.WebAccount) *OGMeta {
	o.Title = AccountTitle(acct, o.SiteName)
	o.ProfileUsername = acct.Username
	o.Type = "profile"
	o.URL = acct.URL
	if acct.Note != "" {
		o.Description = ParseDescription(acct.Note)
	} else {
		const desc = "This GoToSocial user hasn't written a bio yet!"
		o.Description = desc
	}

	// Add avatar image.
	o.prependMedia(ogImgForAcct(acct))

	return o
}

// util funct to return OGImage using account.
func ogImgForAcct(account *apimodel.WebAccount) OGMedia {
	ogMedia := OGMedia{
		OGType: "image",
		URL:    account.Avatar,
		Alt:    "Avatar for " + account.Username,
	}

	if desc := account.AvatarDescription; desc != "" {
		ogMedia.Alt += ": " + desc
	}

	// Add extra info if not default avi.
	if a := account.AvatarAttachment; a != nil {
		ogMedia.MIMEType = a.MIMEType
		ogMedia.Width = strconv.Itoa(a.Meta.Original.Width)
		ogMedia.Height = strconv.Itoa(a.Meta.Original.Height)
	}

	return ogMedia
}

// WithStatus uses the given status to build an ogMeta
// struct specific to that status. It's suitable for serving
// at status pages.
func (o *OGMeta) WithStatus(status *apimodel.WebStatus) *OGMeta {
	o.Title = "Post by " + AccountTitle(status.Account, o.SiteName)
	o.Type = "article"
	if status.Language != nil {
		o.Locale = *status.Language
	}
	o.URL = status.URL
	switch {
	case status.SpoilerText != "":
		o.Description = ParseDescription("CW: " + status.SpoilerText)
	case status.Text != "":
		o.Description = ParseDescription(status.Text)
	default:
		o.Description = o.Title
	}

	// Prepend account image.
	o.prependMedia(ogImgForAcct(status.Account))

	if l := len(status.MediaAttachments); l != 0 && !status.Sensitive {

		// Take first not "unknown"
		// attachment as the "main" one.
		for _, a := range status.MediaAttachments {
			if a.Type == "unknown" {
				// Skip unknown.
				continue
			}

			// Start with
			// common media tags.
			desc := util.PtrOrZero(a.Description)
			ogMedia := OGMedia{
				URL:      *a.URL,
				MIMEType: a.MIMEType,
				Alt:      desc,
			}

			// Gather ogMedias for
			// this attachment.
			ogMedias := []OGMedia{}

			// Add further tags
			// depending on type.
			switch a.Type {

			case "image":
				ogMedia.OGType = "image"
				ogMedia.Width = strconv.Itoa(a.Meta.Original.Width)
				ogMedia.Height = strconv.Itoa(a.Meta.Original.Height)

				// If this image is the only piece of media,
				// set TwitterSummaryLargeImage to indicate
				// that a large image summary is preferred.
				if l == 1 {
					o.TwitterSummaryLargeImage = *a.URL
					o.TwitterImageAlt = desc
				}

			case "audio":
				ogMedia.OGType = "audio"

			case "video", "gifv":
				ogMedia.OGType = "video"
				ogMedia.Width = strconv.Itoa(a.Meta.Original.Width)
				ogMedia.Height = strconv.Itoa(a.Meta.Original.Height)
			}

			// Add this to our gathered entries.
			ogMedias = append(ogMedias, ogMedia)

			if a.Type != "image" {
				// Add static/thumbnail
				// for non-images.
				ogMedias = append(
					ogMedias,
					OGMedia{
						OGType:   "image",
						URL:      *a.PreviewURL,
						MIMEType: a.PreviewMIMEType,
						Width:    strconv.Itoa(a.Meta.Small.Width),
						Height:   strconv.Itoa(a.Meta.Small.Height),
						Alt:      util.PtrOrZero(a.Description),
					},
				)
			}

			// Prepend gathered entries.
			//
			// This will cause the full-size
			// entry to appear before its
			// thumbnail entry (if set).
			o.prependMedia(ogMedias...)

			// Done!
			break
		}
	}

	o.ArticlePublisher = status.Account.URL
	o.ArticleAuthor = status.Account.URL
	o.ArticlePublishedTime = status.CreatedAt
	o.ArticleModifiedTime = util.PtrOrValue(status.EditedAt, status.CreatedAt)

	return o
}

// AccountTitle parses a page title from account and accountDomain
func AccountTitle(account *apimodel.WebAccount, accountDomain string) string {
	user := "@" + account.Acct + "@" + accountDomain

	if len(account.DisplayName) == 0 {
		return user
	}

	return account.DisplayName + ", " + user
}

// ParseDescription returns a string description which is
// safe to use as the content of a `content="..."` attribute.
func ParseDescription(in string) string {
	i := text.StripHTMLFromText(in)
	i = strings.ReplaceAll(i, "\n", " ")
	i = strings.Join(strings.Fields(i), " ")
	i = html.EscapeString(i)
	i = strings.ReplaceAll(i, `\`, "&bsol;")
	return truncate(i)
}

// truncate trims string
// to maximum 160 runes.
func truncate(s string) string {
	const truncateLen = 160

	r := []rune(s)
	if len(r) < truncateLen {
		// No need
		// to trim.
		return s
	}

	return string(r[:truncateLen-3]) + "…"
}
