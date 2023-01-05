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

package transport

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (t *transport) DereferenceMedia(ctx context.Context, iri *url.URL) (io.ReadCloser, int64, error) {
	// Build IRI just once
	iriStr := iri.String()

	// Prepare HTTP request to this media's IRI
	req, err := http.NewRequestWithContext(ctx, "GET", iriStr, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Add("Accept", "*/*") // we don't know what kind of media we're going to get here
	req.Header.Set("Host", iri.Host)

	// Perform the HTTP request
	rsp, err := t.GET(req)
	if err != nil {
		return nil, 0, err
	}

	// Check for an expected status code
	if rsp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("GET request to %s failed (%d): %s", iriStr, rsp.StatusCode, rsp.Status)
	}

	return rsp.Body, rsp.ContentLength, nil
}
