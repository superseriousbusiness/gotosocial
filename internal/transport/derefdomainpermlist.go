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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
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

	// May be set
	// if 200 or 304.
	LastModified time.Time
}

func (t *transport) DereferenceDomainPermissions(
	ctx context.Context,
	permSub *gtsmodel.DomainPermissionSubscription,
	skipCache bool,
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
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("Accept", permSub.ContentType.String()+","+"*/*")

	// If skipCache is true, we want to skip setting Cache
	// headers so that we definitely don't get a 304 back.
	if !skipCache {
		// If we've got a Last-Modified stored for this list,
		// set If-Modified-Since to make the request conditional.
		//
		// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
		if !permSub.LastModified.IsZero() {
			// http.Time wants UTC.
			lmUTC := permSub.LastModified.UTC()
			req.Header.Set("If-Modified-Since", lmUTC.Format(http.TimeFormat))
		}

		// If we've got an ETag stored for this list, set
		// If-None-Match to make the request conditional.
		// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag#caching_of_unchanged_resources.
		if permSub.ETag != "" {
			req.Header.Set("If-None-Match", permSub.ETag)
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

	// Check already if we were given a valid ETag or
	// Last-Modified we can use, as these cache headers
	// are often returned even on Not Modified responses.
	permsResp := &DereferenceDomainPermissionsResp{
		ETag:         rsp.Header.Get("ETag"),
		LastModified: validateLastModified(ctx, rsp.Header.Get("Last-Modified")),
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

// Validate Last-Modified to ensure it's not
// garbagio, and not more than a minute in the
// future (to allow for clock issues + rounding).
func validateLastModified(
	ctx context.Context,
	lastModified string,
) time.Time {
	if lastModified == "" {
		// Not set,
		// no problem.
		return time.Time{}
	}

	// Try to parse and see what we get.
	switch lm, err := http.ParseTime(lastModified); {
	case err != nil:
		// No good,
		// chuck it.
		log.Debugf(ctx,
			"discarding invalid Last-Modified header %s: %+v",
			lastModified, err,
		)
		return time.Time{}

	case lm.Unix() > time.Now().Add(1*time.Minute).Unix():
		// In the future,
		// chuck it.
		log.Debugf(ctx,
			"discarding in-the-future Last-Modified header %s",
			lastModified,
		)
		return time.Time{}

	default:
		// It's fine,
		// keep it.
		return lm
	}
}
