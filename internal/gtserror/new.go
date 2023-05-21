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

package gtserror

import (
	"errors"
	"net/http"

	"codeberg.org/gruf/go-byteutil"
)

// NewResponseError crafts an error from provided HTTP response
// including the method, status and body (if any provided). This
// will also wrap the returned error using WithStatusCode().
func NewResponseError(rsp *http.Response) error {
	var buf byteutil.Buffer

	// Get URL string ahead of time.
	urlStr := rsp.Request.URL.String()

	// Alloc guesstimate of required buf size.
	buf.Guarantee(0 +
		len(rsp.Request.Method) +
		12 + //  request to
		len(urlStr) +
		17 + //  failed: status="
		len(rsp.Status) +
		8 + // " body="
		256 + // max body size
		1, // "
	)

	// Build error message string without
	// using "fmt", as chances are this will
	// be used in a hot code path and we
	// know all the incoming types involved.
	buf.WriteString(rsp.Request.Method)
	buf.WriteString(" request to ")
	buf.WriteString(urlStr)
	buf.WriteString(" failed: status=\"")
	buf.WriteString(rsp.Status)
	buf.WriteString("\" body=\"")
	buf.WriteString(drainBody(rsp.Body, 256))
	buf.WriteString("\"")

	// Create new error from msg.
	err := errors.New(buf.String())

	// Wrap error to provide status code.
	return WithStatusCode(err, rsp.StatusCode)
}
