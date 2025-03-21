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

package status

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/PuerkitoBio/goquery"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

var urlRegex = regexp.MustCompile(`https?://[a-zA-Z0-9./?=_-]+`)

func extractLastURL(text string) string {
	matches := urlRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1]
}

// FetchPreview retrieves OpenGraph metadata from a URL.
func FetchPreview(text string) (*gtsmodel.Card, gtserror.WithCode) {
	link := extractLastURL(text)
	if link == "" {
		return nil, nil
	}

	parsed, err := url.ParseRequestURI(link)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err, "invalid URL")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("unsupported scheme: %s", parsed.Scheme))
	}

	resp, err := http.Get(link)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err, "request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("unexpected status: %s", resp.Status))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("failed to parse HTML: %w", err))
	}

	card := &gtsmodel.Card{
		URL: link,
	}

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		property, _ := s.Attr("property")
		content, _ := s.Attr("content")

		switch property {
		case "og:title":
			card.Title = content
		case "og:description":
			card.Description = content
		case "og:type":
			card.Type = content
		case "og:image":
			card.Image = content
		case "og:url":
			if content != "" {
				card.URL = content
			}
		case "og:site_name":
			card.ProviderName = content
		}
	})

	if card.Title == "" {
		card.Title = doc.Find("title").Text()
	}

	if card.Description == "" {
		desc, exists := doc.Find("meta[name='description']").Attr("content")
		if exists {
			card.Description = desc
		}
	}

	return card, nil
}
