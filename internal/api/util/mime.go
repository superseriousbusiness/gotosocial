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

import "strings"

const (
	// Possible GoToSocial mimetypes.
	AppJSON           = `application/json`
	AppXML            = `application/xml`
	appXMLText        = `text/xml` // AppXML is only *recommended* in RFC7303
	AppXMLXRD         = `application/xrd+xml`
	AppRSSXML         = `application/rss+xml`
	AppActivityJSON   = `application/activity+json`
	appActivityLDJSON = `application/ld+json` // without profile
	AppActivityLDJSON = appActivityLDJSON + `; profile="https://www.w3.org/ns/activitystreams"`
	AppJRDJSON        = `application/jrd+json` // https://www.rfc-editor.org/rfc/rfc7033#section-10.2
	AppForm           = `application/x-www-form-urlencoded`
	MultipartForm     = `multipart/form-data`
	TextXML           = `text/xml`
	TextHTML          = `text/html`
	TextCSS           = `text/css`
	TextCSV           = `text/csv`
)

// JSONContentType returns whether is application/json(;charset=utf-8)? content-type.
func JSONContentType(ct string) bool {
	p := splitContentType(ct)
	p, ok := isUTF8ContentType(p)
	return ok && len(p) == 1 &&
		p[0] == AppJSON
}

// JSONJRDContentType returns whether is application/(jrd+)?json(;charset=utf-8)? content-type.
func JSONJRDContentType(ct string) bool {
	p := splitContentType(ct)
	p, ok := isUTF8ContentType(p)
	return ok && len(p) == 1 &&
		p[0] == AppJSON ||
		p[0] == AppJRDJSON
}

// XMLContentType returns whether is application/xml(;charset=utf-8)? content-type.
func XMLContentType(ct string) bool {
	p := splitContentType(ct)
	p, ok := isUTF8ContentType(p)
	return ok && len(p) == 1 &&
		p[0] == AppXML ||
		p[0] == appXMLText
}

// XMLXRDContentType returns whether is application/(xrd+)?xml(;charset=utf-8)? content-type.
func XMLXRDContentType(ct string) bool {
	p := splitContentType(ct)
	p, ok := isUTF8ContentType(p)
	return ok && len(p) == 1 &&
		p[0] == AppXML ||
		p[0] == appXMLText ||
		p[0] == AppXMLXRD
}

// ASContentType returns whether is valid ActivityStreams content-types:
// - application/activity+json
// - application/ld+json;profile=https://w3.org/ns/activitystreams
func ASContentType(ct string) bool {
	p := splitContentType(ct)
	p, ok := isUTF8ContentType(p)
	if !ok {
		return false
	}
	switch len(p) {
	case 1:
		return p[0] == AppActivityJSON
	case 2:
		return p[0] == appActivityLDJSON &&
			p[1] == "profile=https://www.w3.org/ns/activitystreams" ||
			p[1] == "profile=\"https://www.w3.org/ns/activitystreams\""
	default:
		return false
	}
}

// NodeInfo2ContentType returns whether is nodeinfo schema 2.0 content-type.
func NodeInfo2ContentType(ct string) bool {
	p := splitContentType(ct)
	p, ok := isUTF8ContentType(p)
	if !ok {
		return false
	}
	switch len(p) {
	case 1:
		return p[0] == AppJSON
	case 2:
		return p[0] == AppJSON &&
			p[1] == "profile=\"http://nodeinfo.diaspora.software/ns/schema/2.0#\"" ||
			p[1] == "profile=http://nodeinfo.diaspora.software/ns/schema/2.0#"
	default:
		return false
	}
}

// isUTF8ContentType checks for a provided charset in given
// type parts list, removes it and returns whether is utf-8.
func isUTF8ContentType(p []string) ([]string, bool) {
	const charset = "charset="
	const charsetUTF8 = charset + "utf-8"
	for i, part := range p {

		// Only handle charset slice parts.
		if strings.HasPrefix(part, charset) {

			// Check if is UTF-8 charset.
			ok := (part == charsetUTF8)

			// Drop this slice part.
			_ = copy(p[i:], p[i+1:])
			p = p[:len(p)-1]

			return p, ok
		}
	}
	return p, true
}

// splitContentType splits content-type into semi-colon
// separated parts. useful when a charset is provided.
// note this also maps all chars to their lowercase form.
func splitContentType(ct string) []string {
	s := strings.Split(ct, ";")
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
		s[i] = strings.ToLower(s[i])
	}
	return s
}
