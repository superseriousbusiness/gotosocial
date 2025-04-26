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
	"io"
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-iotools"
)

func (t *transport) DereferenceMedia(ctx context.Context, iri *url.URL, maxsz int64) (io.ReadCloser, error) {
	// Build IRI just once
	iriStr := iri.String()

	// Prepare HTTP request to this media's IRI
	req, err := http.NewRequestWithContext(ctx, "GET", iriStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "*/*") // we don't know what kind of media we're going to get here

	// Perform the HTTP request
	rsp, err := t.GET(req)
	if err != nil {
		return nil, err
	}

	// Check for an expected status code
	if rsp.StatusCode != http.StatusOK {
		return nil, gtserror.NewFromResponse(rsp)
	}

	// Check media within size limit.
	if rsp.ContentLength > maxsz {
		_ = rsp.Body.Close()       // close early.
		sz := bytesize.Size(maxsz) //nolint:gosec
		return nil, gtserror.Newf("media body exceeds max size %s", sz)
	}

	// Update response body with maximum supported media size.
	rsp.Body, _, _ = iotools.UpdateReadCloserLimit(rsp.Body, maxsz)

	return rsp.Body, nil
}
