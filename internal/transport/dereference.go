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

package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// ErrGone is returned from Dereference when the remote resource returns 410 GONE.
// This is useful in cases where we're processing a delete of a resource that's already
// been removed from the remote server, so we know we don't need to keep trying to
// dereference it.
var ErrGone = errors.New("remote resource returned HTTP code 410 GONE")

func (t *transport) Dereference(ctx context.Context, iri *url.URL) ([]byte, error) {
	// if the request is to us, we can shortcut for certain URIs rather than going through
	// the normal request flow, thereby saving time and energy
	if iri.Host == config.GetHost() {
		if uris.IsFollowersPath(iri) {
			// the request is for followers of one of our accounts, which we can shortcut
			return t.controller.dereferenceLocalFollowers(ctx, iri)
		}

		if uris.IsUserPath(iri) {
			// the request is for one of our accounts, which we can shortcut
			return t.controller.dereferenceLocalUser(ctx, iri)
		}
	}

	// Build IRI just once
	iriStr := iri.String()

	// Prepare new HTTP request to endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", iriStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", string(api.AppActivityLDJSON)+","+string(api.AppActivityJSON))
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Add("User-Agent", t.controller.userAgent)
	req.Header.Set("Host", iri.Host)

	// Perform the HTTP request
	rsp, err := t.GET(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case http.StatusOK:
		return io.ReadAll(rsp.Body)
	case http.StatusGone:
		return nil, ErrGone
	default:
		return nil, fmt.Errorf("GET request to %s failed (%d): %s", iriStr, rsp.StatusCode, rsp.Status)
	}
}
