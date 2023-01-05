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

package text

import (
	"bytes"
	"context"
	"net/url"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/regexes"
)

// FindLinks parses the given string looking for recognizable URLs (including scheme).
// It returns a list of those URLs, without changing the string, or an error if something goes wrong.
// If no URLs are found within the given string, an empty slice and nil will be returned.
func FindLinks(in string) []*url.URL {
	var urls []*url.URL

	// bail already if we don't find anything
	found := regexes.LinkScheme.FindAllString(in, -1)
	if len(found) == 0 {
		return nil
	}

	urlmap := map[string]struct{}{}

	// for each string we find, we want to parse it into a URL if we can
	// if we fail to parse it, just ignore this match and continue
	for _, f := range found {
		u, err := url.Parse(f)
		if err != nil {
			continue
		}

		// Calculate string
		ustr := u.String()

		if _, ok := urlmap[ustr]; !ok {
			// Has not been encountered yet
			urls = append(urls, u)
			urlmap[ustr] = struct{}{}
		}
	}

	return urls
}

// ReplaceLinks replaces all detected links in a piece of text with their HTML (href) equivalents.
// Note: because Go doesn't allow negative lookbehinds in regex, it's possible that an already-formatted
// href will end up double-formatted, if the text you pass here contains one or more hrefs already.
// To avoid this, you should sanitize any HTML out of text before you pass it into this function.
func (f *formatter) ReplaceLinks(ctx context.Context, in string) string {
	return regexes.ReplaceAllStringFunc(regexes.LinkScheme, in, func(urlString string, buf *bytes.Buffer) string {
		thisURL, err := url.Parse(urlString)
		if err != nil {
			return urlString // we can't parse it as a URL so don't replace it
		}
		// <a href="thisURL.String()" rel="noopener">urlString</a>
		urlString = thisURL.String()
		buf.WriteString(`<a href="`)
		buf.WriteString(thisURL.String())
		buf.WriteString(`" rel="noopener">`)
		urlString = strings.TrimPrefix(urlString, thisURL.Scheme)
		urlString = strings.TrimPrefix(urlString, "://")
		buf.WriteString(urlString)
		buf.WriteString(`</a>`)
		return buf.String()
	})
}
