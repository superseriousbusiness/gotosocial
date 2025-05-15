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
	"errors"
	"net/http"
	"net/url"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/federation/federatingdb"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

func (t *transport) Dereference(ctx context.Context, iri *url.URL) (*http.Response, error) {
	// If the request is to us, we can try to shortcut
	// rather than going through the normal request flow.
	//
	// Only bail on a real error, otherwise continue
	// to just make a normal http request to ourself.
	if iri.Host == config.GetHost() {
		rsp, err := t.controller.dereferenceLocal(ctx, iri)
		if err != nil && !errors.Is(err, federatingdb.ErrNotImplemented) {
			return nil, gtserror.Newf("error dereferencing local: %w", err)
		}

		if rsp != nil {
			// Got something!
			//
			// No need for
			// further business.
			return rsp, nil
		}

		// Blast out a cheeky warning so we can keep track of this.
		log.Warnf(ctx, "about to perform request to self: GET %s", iri)
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
