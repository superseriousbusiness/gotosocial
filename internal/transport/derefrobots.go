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

package transport

import (
	"context"
	"net/http"
	"net/url"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-iotools"
	"github.com/temoto/robotstxt"
)

func (t *transport) DereferenceRobots(ctx context.Context, protocol string, host string) (*robotstxt.RobotsData, error) {
	robotsIRI := &url.URL{
		Scheme: protocol,
		Host:   host,
		Path:   "robots.txt",
	}

	// Build IRI just once
	iriStr := robotsIRI.String()

	// Prepare new HTTP request to endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", iriStr, nil)
	if err != nil {
		return nil, err
	}

	// We want text/plain utf-8 encoding.
	//
	// https://www.rfc-editor.org/rfc/rfc9309.html#name-access-method
	req.Header.Add("Accept", apiutil.TextPlain)
	req.Header.Add("Accept-Charset", apiutil.UTF8)

	// Perform the HTTP request
	rsp, err := t.GET(req)
	if err != nil {
		return nil, err
	}

	// Ensure a non-error status response.
	if rsp.StatusCode != http.StatusOK {
		err := gtserror.NewFromResponse(rsp)
		_ = rsp.Body.Close() // close early.
		return nil, err
	}

	// Ensure that the incoming request content-type is expected.
	if ct := rsp.Header.Get("Content-Type"); !apiutil.TextPlainContentType(ct) {
		err := gtserror.Newf("non text/plain response: %s", ct)
		_ = rsp.Body.Close() // close early.
		return nil, gtserror.SetMalformed(err)
	}

	// Limit the robots.txt size to 500KiB
	//
	// https://www.rfc-editor.org/rfc/rfc9309.html#name-limits
	const maxsz = int64(500 * bytesize.KiB)

	// Check body claims to be within size limit.
	if rsp.ContentLength > maxsz {
		_ = rsp.Body.Close()       // close early.
		sz := bytesize.Size(maxsz) //nolint:gosec
		return nil, gtserror.Newf("robots.txt body exceeds max size %s", sz)
	}

	// Update response body with maximum size.
	rsp.Body, _, _ = iotools.UpdateReadCloserLimit(rsp.Body, maxsz)
	defer rsp.Body.Close()

	return robotstxt.FromResponse(rsp)
}
