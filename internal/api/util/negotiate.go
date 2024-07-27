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
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// JSONAcceptHeaders is a slice of offers that just contains application/json types.
var JSONAcceptHeaders = []string{
	AppJSON,
}

// WebfingerJSONAcceptHeaders is a slice of offers that prefers the
// jrd+json content type, but will be chill and fall back to app/json.
// This is to be used specifically for webfinger responses.
// See https://www.rfc-editor.org/rfc/rfc7033#section-10.2
var WebfingerJSONAcceptHeaders = []string{
	AppJRDJSON,
	AppJSON,
}

// JSONOrHTMLAcceptHeaders is a slice of offers that prefers AppJSON and will
// fall back to HTML if necessary. This is useful for error handling, since it can
// be used to serve a nice HTML page if the caller accepts that, or just JSON if not.
var JSONOrHTMLAcceptHeaders = []string{
	AppJSON,
	TextHTML,
}

// HTMLAcceptHeaders is a slice of offers that just contains text/html types.
var HTMLAcceptHeaders = []string{
	TextHTML,
}

// HTMLOrActivityPubHeaders matches text/html first, then activitypub types.
// This is useful for user URLs that a user might go to in their browser,
// but which should also be able to serve ActivityPub as a fallback.
//
// https://www.w3.org/TR/activitypub/#retrieving-objects
var HTMLOrActivityPubHeaders = []string{
	TextHTML,
	AppActivityLDJSON,
	AppActivityJSON,
}

// ActivityPubOrHTMLHeaders matches activitypub types first, then text/html.
// This is useful for URLs that should serve ActivityPub by default, but
// which a user might also go to in their browser sometimes.
//
// https://www.w3.org/TR/activitypub/#retrieving-objects
var ActivityPubOrHTMLHeaders = []string{
	AppActivityLDJSON,
	AppActivityJSON,
	TextHTML,
}

// ActivityPubHeaders matches only activitypub Accept headers.
// This is useful for URLs should only serve ActivityPub.
//
// https://www.w3.org/TR/activitypub/#retrieving-objects
var ActivityPubHeaders = []string{
	AppActivityLDJSON,
	AppActivityJSON,
}

var HostMetaHeaders = []string{
	AppXMLXRD,
	AppXML,
}

// CSVHeaders just contains the text/csv
// MIME type, used for import/export.
var CSVHeaders = []string{
	TextCSV,
}

// NegotiateAccept takes the *gin.Context from an incoming request, and a
// slice of Offers, and performs content negotiation for the given request
// with the given content-type offers. It will return a string representation
// of the first suitable content-type, or an error if something goes wrong or
// a suitable content-type cannot be matched.
//
// For example, if the request in the *gin.Context has Accept headers of value
// [application/json, text/html], and the provided offers are of value
// [application/json, application/xml], then the returned string will be
// 'application/json', which indicates the content-type that should be returned.
//
// If the length of offers is 0, then an error will be returned, so this function
// should only be called in places where format negotiation is actually needed.
//
// If there are no Accept headers in the request, then the first offer will be returned,
// under the assumption that it's better to serve *something* than error out completely.
//
// Callers can use the offer slices exported in this package as shortcuts for
// often-used Accept types.
//
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Content_negotiation#server-driven_content_negotiation
func NegotiateAccept(c *gin.Context, offers ...string) (string, error) {
	if len(offers) == 0 {
		return "", errors.New("no format offered")
	}

	strings := []string{}
	for _, o := range offers {
		strings = append(strings, string(o))
	}

	accepts := c.Request.Header.Values("Accept")
	if len(accepts) == 0 {
		// there's no accept header set, just return the first offer
		return strings[0], nil
	}

	format := NegotiateFormat(c, strings...)
	if format == "" {
		return "", fmt.Errorf("no format can be offered for requested Accept header(s) %s; this endpoint offers %s", accepts, offers)
	}

	return format, nil
}

// This is the exact same thing as gin.Context.NegotiateFormat except it contains
// tsmethurst's fix to make it work properly with multiple accept headers.
//
// https://github.com/gin-gonic/gin/pull/3156
func NegotiateFormat(c *gin.Context, offered ...string) string {
	if len(offered) == 0 {
		panic("you must provide at least one offer")
	}

	if c.Accepted == nil {
		for _, a := range c.Request.Header.Values("Accept") {
			c.Accepted = append(c.Accepted, parseAccept(a)...)
		}
	}
	if len(c.Accepted) == 0 {
		return offered[0]
	}
	for _, accepted := range c.Accepted {
		for _, offer := range offered {
			// According to RFC 2616 and RFC 2396, non-ASCII characters are not allowed in headers,
			// therefore we can just iterate over the string without casting it into []rune
			i := 0
			for ; i < len(accepted); i++ {
				if accepted[i] == '*' || offer[i] == '*' {
					return offer
				}
				if accepted[i] != offer[i] {
					break
				}
			}
			if i == len(accepted) {
				return offer
			}
		}
	}
	return ""
}

// https://github.com/gin-gonic/gin/blob/4787b8203b79012877ac98d7806422da3a678ba2/utils.go#L103
func parseAccept(acceptHeader string) []string {
	parts := strings.Split(acceptHeader, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if i := strings.IndexByte(part, ';'); i > 0 {
			part = part[:i]
		}
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
