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
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// webfingerURLFor returns the URL to try a webfinger request against, as
// well as if the URL was retrieved from cache. When the URL is retrieved
// from cache we don't have to try and do host-meta discovery
func (t *transport) webfingerURLFor(targetDomain string) (string, bool) {
	url := "https://" + targetDomain + "/.well-known/webfinger"

	wc := t.controller.state.Caches.Webfinger

	// We're doing the manual locking/unlocking here to be able to
	// safely call Cache.Get instead of Get, as the latter updates the
	// item expiry which we don't want to do here
	wc.Lock()
	item, ok := wc.Cache.Get(targetDomain)
	wc.Unlock()

	if ok {
		url = item.Value
	}

	return url, ok
}

func prepWebfingerReq(ctx context.Context, loc, domain, username string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loc, nil)
	if err != nil {
		return nil, err
	}

	value := url.QueryEscape("acct:" + username + "@" + domain)
	req.URL.RawQuery = "resource=" + value

	// Prefer application/jrd+json, fall back to application/json.
	// See https://www.rfc-editor.org/rfc/rfc7033#section-10.2.
	//
	// Some implementations don't handle multiple accept headers properly,
	// including Gin itself. So concat the accept header with a comma
	// instead which seems to work reliably
	req.Header.Add("Accept", string(apiutil.AppJRDJSON)+","+string(apiutil.AppJSON))

	return req, nil
}

func (t *transport) Finger(ctx context.Context, targetUsername string, targetDomain string) ([]byte, error) {
	// Remotes seem to prefer having their punycode
	// domain used in webfinger requests, so let's oblige.
	punyDomain, err := util.Punify(targetDomain)
	if err != nil {
		return nil, gtserror.Newf("error punifying %s: %w", targetDomain, err)
	}

	// Generate new GET request
	url, cached := t.webfingerURLFor(punyDomain)
	req, err := prepWebfingerReq(ctx, url, punyDomain, targetUsername)
	if err != nil {
		return nil, err
	}

	// Perform the HTTP request
	rsp, err := t.GET(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	// Check if the request succeeded so we can bail out early or if we explicitly
	// got a "this resource is gone" response which will happen when a user has
	// deleted the account
	if rsp.StatusCode == http.StatusOK || rsp.StatusCode == http.StatusGone {
		if cached {
			// If we got a response we consider successful on a cached URL, i.e one set
			// by us later on when a host-meta based webfinger request succeeded, set it
			// again here to renew the TTL
			t.controller.state.Caches.Webfinger.Set(punyDomain, url)
		}

		if rsp.StatusCode == http.StatusGone {
			return nil, fmt.Errorf("account has been deleted/is gone")
		}

		// Ensure that the incoming request content-type is expected.
		if ct := rsp.Header.Get("Content-Type"); !apiutil.JSONJRDContentType(ct) {
			err := gtserror.Newf("non webfinger type response: %s", ct)
			return nil, gtserror.SetMalformed(err)
		}

		return io.ReadAll(rsp.Body)
	}

	// From here on out, we're handling different failure scenarios and
	// deciding whether we should do a host-meta based fallback or not

	// Response status codes >= 500 are returned as errors by the wrapped HTTP client.
	//
	// if (rsp.StatusCode >= 500 && rsp.StatusCode < 600) || cached {
	// In case we got a 5xx, bail out irrespective of if the value
	// was cached or not. The target may be broken or be signalling
	// us to back-off.
	//
	// If it's any error but the URL was cached, bail out too
	// return nil, gtserror.NewResponseError(rsp)
	// }

	// So far we've failed to get a successful response from the expected
	// webfinger endpoint. Lets try and discover the webfinger endpoint
	// through /.well-known/host-meta
	host, err := t.webfingerFromHostMeta(ctx, punyDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to discover webfinger URL fallback for: %s through host-meta: %w", targetDomain, err)
	}

	// Check if the original and host-meta URL are the same. If they
	// are there's no sense in us trying the request again as it just
	// failed
	if host == url {
		return nil, fmt.Errorf("webfinger discovery on %s returned endpoint we already tried: %s", targetDomain, host)
	}

	// Now that we have a different URL for the webfinger
	// endpoint, try the request against that endpoint instead
	req, err = prepWebfingerReq(ctx, host, punyDomain, targetUsername)
	if err != nil {
		return nil, err
	}

	// Perform the HTTP request
	rsp, err = t.GET(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		// A HTTP 410 indicates we got a response to our webfinger query, but the resource
		// we asked for is gone. This means the endpoint itself is valid and we should
		// cache it for future queries to the same domain
		if rsp.StatusCode == http.StatusGone {
			t.controller.state.Caches.Webfinger.Set(targetDomain, host)
			return nil, fmt.Errorf("account has been deleted/is gone")
		}
		// We've reached the end of the line here, both the original request
		// and our attempt to resolve it through the fallback have failed
		return nil, gtserror.NewFromResponse(rsp)
	}

	// Set the URL in cache here, since host-meta told us this should be the
	// valid one, it's different from the default and our request to it did
	// not fail in any manner
	t.controller.state.Caches.Webfinger.Set(targetDomain, host)

	return io.ReadAll(rsp.Body)
}

func (t *transport) webfingerFromHostMeta(ctx context.Context, targetDomain string) (string, error) {
	// Build the request for the host-meta endpoint
	hmurl := "https://" + targetDomain + "/.well-known/host-meta"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hmurl, nil)
	if err != nil {
		return "", err
	}

	// We're doing XML
	req.Header.Add("Accept", string(apiutil.AppXML))
	req.Header.Add("Accept", "application/xrd+xml")

	// Perform the HTTP request
	rsp, err := t.GET(req)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	// Doesn't look like host-meta is working for this instance
	if rsp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET request for %s failed: %s", req.URL.String(), rsp.Status)
	}

	// Ensure that the incoming request content-type is expected.
	if ct := rsp.Header.Get("Content-Type"); !apiutil.XMLXRDContentType(ct) {
		err := gtserror.Newf("non host-meta type response: %s", ct)
		return "", gtserror.SetMalformed(err)
	}

	e := xml.NewDecoder(rsp.Body)
	var hm apimodel.HostMeta
	if err := e.Decode(&hm); err != nil {
		// We got something, but it's not a host-meta document we understand
		return "", fmt.Errorf("failed to decode host-meta response for %s at %s: %w", targetDomain, req.URL.String(), err)
	}

	for _, link := range hm.Link {
		// Based on what we currently understand, there should not be more than one
		// of these with Rel="lrdd" in a host-meta document
		if link.Rel == "lrdd" {
			u, err := url.Parse(link.Template)
			if err != nil {
				return "", fmt.Errorf("lrdd link is not a valid url: %w", err)
			}
			// Get rid of the query template, we only want the scheme://host/path part
			u.RawQuery = ""
			urlStr := u.String()
			return urlStr, nil
		}
	}
	return "", fmt.Errorf("no webfinger URL found")
}
