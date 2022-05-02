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

package text

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/regexes"
)

// FindLinks parses the given string looking for recognizable URLs (including scheme).
// It returns a list of those URLs, without changing the string, or an error if something goes wrong.
// If no URLs are found within the given string, an empty slice and nil will be returned.
func FindLinks(in string) ([]*url.URL, error) {
	urls := []*url.URL{}

	// bail already if we don't find anything
	found := regexes.LinkScheme.FindAllString(in, -1)
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
	return regexes.ReplaceAllStringFunc(regexes.LinkScheme, in, func(urlString string, buf *bytes.Buffer) string {
		// Check we have received parseable URL
		thisURL, err := url.Parse(urlString)
		if err != nil {
			return urlString
		}

		// Write HTML href with actual URL
		fmt.Fprintf(buf, `<a href="%s" rel="noopener">`, urlString)

		// Write hostname to buf
		buf.WriteString(thisURL.Hostname())

		// Write any path to buf
		if thisURL.Path != "" {
			buf.WriteString(thisURL.Path)
		}

		// Write any query to buf
		if thisURL.RawQuery != "" {
			buf.WriteByte('?')
			buf.WriteString(thisURL.RawQuery)
		}

		// Write any fragment to buf
		if thisURL.Fragment != "" {
			buf.WriteByte('#')
			buf.WriteString(thisURL.RawFragment)
		}

		// Write remainder of href
		buf.WriteString(`</a>`)

		return buf.String()
	})
}
