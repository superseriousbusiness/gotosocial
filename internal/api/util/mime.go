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

// MIME represents a mime-type.
type MIME string

// MIME type
const (
	AppJSON           MIME = `application/json`
	AppXML            MIME = `application/xml`
	AppXMLXRD         MIME = `application/xrd+xml`
	AppRSSXML         MIME = `application/rss+xml`
	AppActivityJSON   MIME = `application/activity+json`
	AppActivityLDJSON MIME = `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`
	AppForm           MIME = `application/x-www-form-urlencoded`
	MultipartForm     MIME = `multipart/form-data`
	TextXML           MIME = `text/xml`
	TextHTML          MIME = `text/html`
	TextCSS           MIME = `text/css`
)
