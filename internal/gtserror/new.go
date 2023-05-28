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
	"net/http"
)

// New returns a new error, prepended with caller function name if gtserror.Caller is enabled.
func New(msg string) error {
	return newAt(3, msg)
}

// Newf returns a new formatted error, prepended with caller function name if gtserror.Caller is enabled.
func Newf(msgf string, args ...any) error {
	return newfAt(3, msgf, args...)
}

// NewResponseError crafts an error from provided HTTP response
// including the method, status and body (if any provided). This
// will also wrap the returned error using WithStatusCode() and
// will include the caller function name as a prefix.
func NewFromResponse(rsp *http.Response) error {
	// Build error with message without
	// using "fmt", as chances are this will
	// be used in a hot code path and we
	// know all the incoming types involved.
	err := newAt(3, ""+
		rsp.Request.Method+
		" request to "+
		rsp.Request.URL.String()+
		" failed: status=\""+
		rsp.Status+
		"\" body=\""+
		drainBody(rsp.Body, 256)+
		"\"",
	)

	// Wrap error to provide status code.
	return WithStatusCode(err, rsp.StatusCode)
}
