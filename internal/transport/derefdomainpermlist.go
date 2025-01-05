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

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type DereferenceDomainPermissionsResp struct {
	// Set only if response was 200 OK.
	// It's up to the caller to close
	// this when they're done with it.
	Body io.ReadCloser

	// True if response
	// was 304 Not Modified.
	Unmodified bool

	// May be set
	// if 200 or 304.
	ETag string
}

func (t *transport) DereferenceDomainPermissions(
	ctx context.Context,
	permSub *gtsmodel.DomainPermissionSubscription,
	force bool,
) (*DereferenceDomainPermissionsResp, error) {
	// Prepare new HTTP request to endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", permSub.URI, nil)
	if err != nil {
		return nil, err
	}

	// Set basic auth header if necessary.
	if permSub.FetchUsername != "" || permSub.FetchPassword != "" {
		req.SetBasicAuth(permSub.FetchUsername, permSub.FetchPassword)
	}

	// Set relevant Accept headers.
	// Allow fallback in case target doesn't
	// negotiate content type correctly.
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Add("Accept", permSub.ContentType.String()+","+"*/*")

	// If force is true, we want to skip setting Cache
	// headers so that we definitely don't get a 304 back.
	if !force {
		// If we've successfully fetched this list
		// before, set If-Modified-Since to last
		// success to make the request conditional.
		//
		// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
		if !permSub.SuccessfullyFetchedAt.IsZero() {
			timeStr := permSub.SuccessfullyFetchedAt.Format(http.TimeFormat)
			req.Header.Add("If-Modified-Since", timeStr)
		}

		// If we've got an ETag stored for this list, set
		// If-None-Match to make the request conditional.
		// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag#caching_of_unchanged_resources.
		if len(permSub.ETag) != 0 {
			req.Header.Add("If-None-Match", permSub.ETag)
		}
	}

	// Perform the HTTP request
	rsp, err := t.GET(req)
	if err != nil {
		return nil, err
	}

	// If we have an unexpected / error response,
	// wrap + return as error. This will also drain
	// and close the response body for us.
	if rsp.StatusCode != http.StatusOK &&
		rsp.StatusCode != http.StatusNotModified {
		err := gtserror.NewFromResponse(rsp)
		return nil, err
	}

	// Check already if we were given an ETag
	// we can use, as ETag is often returned
	// even on 304 Not Modified responses.
	permsResp := &DereferenceDomainPermissionsResp{
		ETag: rsp.Header.Get("Etag"),
	}

	if rsp.StatusCode == http.StatusNotModified {
		// Nothing has changed on the remote side
		// since we last fetched, so there's nothing
		// to do and we don't need to read the body.
		rsp.Body.Close()
		permsResp.Unmodified = true
	} else {
		// Return the live body to the caller.
		permsResp.Body = rsp.Body
	}

	return permsResp, nil
}
