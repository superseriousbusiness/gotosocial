/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package text

import (
	"context"
	"fmt"
	"net/url"

	"mvdan.cc/xurls/v2"
)

// schemes is the regex for schemes we accept when looking for links.
// Basically, we accept https or http.
var schemes = `(((http|https))://)`

// FindLinks parses the given string looking for recognizable URLs (including scheme).
// It returns a list of those URLs, without changing the string, or an error if something goes wrong.
// If no URLs are found within the given string, an empty slice and nil will be returned.
func FindLinks(in string) ([]*url.URL, error) {
	rxStrict, err := xurls.StrictMatchingScheme(schemes)
	if err != nil {
		return nil, err
	}

	urls := []*url.URL{}

	// bail already if we don't find anything
	found := rxStrict.FindAllString(in, -1)
	if len(found) == 0 {
		return urls, nil
	}

	// for each string we find, we want to parse it into a URL if we can
	// if we fail to parse it, just ignore this match and continue
	for _, f := range found {
		u, err := url.Parse(f)
		if err != nil {
			continue
		}
		urls = append(urls, u)
	}

	// deduplicate the URLs
	urlsDeduped := []*url.URL{}

	for _, u := range urls {
		if !contains(urlsDeduped, u) {
			urlsDeduped = append(urlsDeduped, u)
		}
	}

	return urlsDeduped, nil
}

// contains checks if the given url is already within a slice of URLs
func contains(urls []*url.URL, url *url.URL) bool {
	for _, u := range urls {
		if u.String() == url.String() {
			return true
		}
	}
	return false
}

// ReplaceLinks replaces all detected links in a piece of text with their HTML (href) equivalents.
// Note: because Go doesn't allow negative lookbehinds in regex, it's possible that an already-formatted
// href will end up double-formatted, if the text you pass here contains one or more hrefs already.
// To avoid this, you should sanitize any HTML out of text before you pass it into this function.
func (f *formatter) ReplaceLinks(ctx context.Context, in string) string {
	rxStrict, err := xurls.StrictMatchingScheme(schemes)
	if err != nil {
		panic(err)
	}

	replaced := rxStrict.ReplaceAllStringFunc(in, func(urlString string) string {
		thisURL, err := url.Parse(urlString)
		if err != nil {
			return urlString // we can't parse it as a URL so don't replace it
		}

		shortString := thisURL.Hostname()

		if thisURL.Path != "" {
			shortString = shortString + thisURL.Path
		}

		if thisURL.Fragment != "" {
			shortString = shortString + "#" + thisURL.Fragment
		}

		if thisURL.RawQuery != "" {
			shortString = shortString + "?" + thisURL.RawQuery
		}

		replacement := fmt.Sprintf(`<a href="%s" rel="noopener">%s</a>`, urlString, shortString)
		return replacement
	})
	return replaced
}
