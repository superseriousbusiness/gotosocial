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

	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (t *transport) Dereference(ctx context.Context, iri *url.URL) (*http.Response, error) {
	// If the request is to us, we can shortcut for
	// certain URIs rather than going through the normal
	// request flow, thereby saving time and energy.
	if iri.Host == config.GetHost() {
		switch {

		case uris.IsFollowersPath(iri):
			// The request is for followers of one of
			// our accounts, which we can shortcut.
			return t.controller.dereferenceLocalFollowers(ctx, iri)

		case uris.IsUserPath(iri):
			// The request is for one of our
			// accounts, which we can shortcut.
			return t.controller.dereferenceLocalUser(ctx, iri)

		case uris.IsAcceptsPath(iri):
			// The request is for an Accept on
			// our instance, which we can shortcut.
			return t.controller.dereferenceLocalAccept(ctx, iri)
		}
	}

	// Build IRI just once
	iriStr := iri.String()

	// Prepare new HTTP request to endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", iriStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", string(apiutil.AppActivityLDJSON)+","+string(apiutil.AppActivityJSON))
	req.Header.Add("Accept-Charset", "utf-8")

	// Perform the HTTP request
	rsp, err := t.GET(req)
	if err != nil {
		return nil, err
	}

	// Ensure a non-error status response.
	if rsp.StatusCode != http.StatusOK {
		err := gtserror.NewFromResponse(rsp)
		_ = rsp.Body.Close() // done with body
		return nil, err
	}

	// Ensure that the incoming request content-type is expected.
	if ct := rsp.Header.Get("Content-Type"); !apiutil.ASContentType(ct) {
		err := gtserror.Newf("non activity streams response: %s", ct)
		_ = rsp.Body.Close() // done with body
		return nil, gtserror.SetMalformed(err)
	}

	return rsp, nil
}
