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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
	"github.com/temoto/robotstxt"
)

func (t *transport) DereferenceInstance(ctx context.Context, iri *url.URL) (*gtsmodel.Instance, error) {
	// Try to fetch robots.txt to check
	// if we're allowed to try endpoints:
	//
	//   - /api/v1/instance
	//   - /.well-known/nodeinfo
	//   - /nodeinfo/2.0|2.1 endpoints
	robotsTxt, err := t.DereferenceRobots(ctx, iri.Scheme, iri.Host)
	if err != nil {
		log.Debugf(ctx, "couldn't fetch robots.txt from %s: %v", iri.Host, err)
	}

	var i *gtsmodel.Instance

	// First try to dereference using /api/v1/instance.
	// This will provide the most complete picture of an instance, and avoid unnecessary api calls.
	//
	// This will only work with Mastodon-api compatible instances: Mastodon, some Pleroma instances, GoToSocial.
	log.Debugf(ctx, "trying to dereference instance %s by /api/v1/instance", iri.Host)
	i, err = t.dereferenceByAPIV1Instance(ctx, iri, robotsTxt)
	if err == nil {
		log.Debugf(ctx, "successfully dereferenced instance using /api/v1/instance")
		return i, nil
	}
	log.Debugf(ctx, "couldn't dereference instance using /api/v1/instance: %s", err)

	// If that doesn't work, try to dereference using /.well-known/nodeinfo.
	// This will involve two API calls and return less info overall, but should be more widely compatible.
	log.Debugf(ctx, "trying to dereference instance %s by /.well-known/nodeinfo", iri.Host)
	i, err = t.dereferenceByNodeInfo(ctx, iri, robotsTxt)
	if err == nil {
		log.Debugf(ctx, "successfully dereferenced instance using /.well-known/nodeinfo")
		return i, nil
	}
	log.Debugf(ctx, "couldn't dereference instance using /.well-known/nodeinfo: %s", err)

	// we couldn't dereference the instance using any of the known methods, so just return a minimal representation
	log.Debugf(ctx, "returning minimal representation of instance %s", iri.Host)
	id, err := id.NewRandomULID()
	if err != nil {
		return nil, fmt.Errorf("error creating new id for instance %s: %s", iri.Host, err)
	}

	return &gtsmodel.Instance{
		ID:     id,
		Domain: iri.Host,
		URI:    iri.String(),
	}, nil
}

func (t *transport) dereferenceByAPIV1Instance(
	ctx context.Context,
	iri *url.URL,
	robotsTxt *robotstxt.RobotsData,
) (*gtsmodel.Instance, error) {
	const path = "api/v1/instance"

	// Bail if we're not allowed to fetch this endpoint.
	if robotsTxt != nil && !robotsTxt.TestAgent("/"+path, t.controller.userAgent) {
		err := gtserror.Newf("can't fetch %s: robots.txt disallows it", path)
		return nil, gtserror.SetNotPermitted(err)
	}

	cleanIRI := &url.URL{
		Scheme: iri.Scheme,
		Host:   iri.Host,
		Path:   path,
	}

	// Build IRI just once
	iriStr := cleanIRI.String()

	req, err := http.NewRequestWithContext(ctx, "GET", iriStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", string(apiutil.AppJSON))

	resp, err := t.GET(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Ensure a non-error status response.
	if resp.StatusCode != http.StatusOK {
		return nil, gtserror.NewFromResponse(resp)
	}

	// Ensure that we can use data returned from this endpoint.
	robots := resp.Header.Values("X-Robots-Tag")
	if slices.ContainsFunc(
		robots,
		func(key string) bool {
			return strings.Contains(key, "noindex")
		},
	) {
		err := gtserror.Newf("can't use fetched %s: robots tags disallows it", path)
		return nil, gtserror.SetNotPermitted(err)
	}

	// Ensure that the incoming request content-type is expected.
	if ct := resp.Header.Get("Content-Type"); !apiutil.JSONContentType(ct) {
		err := gtserror.Newf("non json response type: %s", ct)
		return nil, gtserror.SetMalformed(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if len(b) == 0 {
		return nil, errors.New("response bytes was len 0")
	}

	// Try to parse the returned bytes
	// directly into an Instance model.
	apiResp := &apimodel.InstanceV1{}
	if err := json.Unmarshal(b, apiResp); err != nil {
		return nil, err
	}

	var contactUsername string
	if apiResp.ContactAccount != nil {
		contactUsername = apiResp.ContactAccount.Username
	}

	ulid, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	i := &gtsmodel.Instance{
		ID:                     ulid,
		Domain:                 iri.Host,
		Title:                  apiResp.Title,
		URI:                    iri.Scheme + "://" + iri.Host,
		ShortDescription:       apiResp.ShortDescription,
		Description:            apiResp.Description,
		ContactEmail:           apiResp.Email,
		ContactAccountUsername: contactUsername,
		Version:                apiResp.Version,
	}

	return i, nil
}

func (t *transport) dereferenceByNodeInfo(
	ctx context.Context,
	iri *url.URL,
	robotsTxt *robotstxt.RobotsData,
) (*gtsmodel.Instance, error) {
	// Retrieve the nodeinfo IRI from .well-known/nodeinfo.
	niIRI, err := t.callNodeInfoWellKnown(ctx, iri, robotsTxt)
	if err != nil {
		return nil, gtserror.Newf("error during initial call to .well-known: %w", err)
	}

	// Use the returned nodeinfo IRI to make a followup call.
	ni, err := t.callNodeInfo(ctx, niIRI, robotsTxt)
	if err != nil {
		return nil, gtserror.Newf("error during call to %s: %w", niIRI.String(), err)
	}

	// We got a response of some kind!
	//
	// Start building out the bare minimum
	// instance model, we'll add to it if we can.
	id, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.Newf("error creating new id for instance %s: %w", iri.Host, err)
	}

	i := &gtsmodel.Instance{
		ID:     id,
		Domain: iri.Host,
		URI:    iri.String(),
	}

	var title string
	if i, present := ni.Metadata["nodeName"]; present {
		// it's present, check it's a string
		if v, ok := i.(string); ok {
			// it is a string!
			title = v
		}
	}
	i.Title = title

	var shortDescription string
	if i, present := ni.Metadata["nodeDescription"]; present {
		// it's present, check it's a string
		if v, ok := i.(string); ok {
			// it is a string!
			shortDescription = v
		}
	}
	i.ShortDescription = shortDescription

	var contactEmail string
	var contactAccountUsername string
	if i, present := ni.Metadata["maintainer"]; present {
		// it's present, check it's a map
		if v, ok := i.(map[string]string); ok {
			// see if there's an email in the map
			if email, present := v["email"]; present {
				if err := validate.Email(email); err == nil {
					// valid email address
					contactEmail = email
				}
			}
			// see if there's a 'name' in the map
			if name, present := v["name"]; present {
				// name could be just a username, or could be a mention string eg @whatever@aaaa.com
				username, _, err := util.ExtractNamestringParts(name)
				if err == nil {
					// it was a mention string
					contactAccountUsername = username
				} else {
					// not a mention string
					contactAccountUsername = name
				}
			}
		}
	}
	i.ContactEmail = contactEmail
	i.ContactAccountUsername = contactAccountUsername

	var software string
	if ni.Software.Name != "" {
		software = ni.Software.Name
	}
	if ni.Software.Version != "" {
		software = software + " " + ni.Software.Version
	}
	i.Version = software

	return i, nil
}

func (t *transport) callNodeInfoWellKnown(
	ctx context.Context,
	iri *url.URL,
	robotsTxt *robotstxt.RobotsData,
) (*url.URL, error) {
	const path = ".well-known/nodeinfo"

	// Bail if we're not allowed to fetch this endpoint.
	if robotsTxt != nil && !robotsTxt.TestAgent("/"+path, t.controller.userAgent) {
		err := gtserror.Newf("can't fetch %s: robots.txt disallows it", path)
		return nil, gtserror.SetNotPermitted(err)
	}

	cleanIRI := &url.URL{
		Scheme: iri.Scheme,
		Host:   iri.Host,
		Path:   path,
	}

	// Build IRI just once
	iriStr := cleanIRI.String()

	req, err := http.NewRequestWithContext(ctx, "GET", iriStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", string(apiutil.AppJSON))

	resp, err := t.GET(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Ensure a non-error status response.
	if resp.StatusCode != http.StatusOK {
		return nil, gtserror.NewFromResponse(resp)
	}

	// Ensure that we can use data returned from this endpoint.
	robots := resp.Header.Values("X-Robots-Tag")
	if slices.ContainsFunc(
		robots,
		func(key string) bool {
			return strings.Contains(key, "noindex")
		},
	) {
		err := gtserror.Newf("can't use fetched %s: robots tags disallows it", path)
		return nil, gtserror.SetNotPermitted(err)
	}

	// Ensure that the returned content-type is expected.
	if ct := resp.Header.Get("Content-Type"); !apiutil.JSONContentType(ct) {
		err := gtserror.Newf("non json response type: %s", ct)
		return nil, gtserror.SetMalformed(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if len(b) == 0 {
		return nil, gtserror.New("response bytes was len 0")
	}

	wellKnownResp := &apimodel.WellKnownResponse{}
	if err := json.Unmarshal(b, wellKnownResp); err != nil {
		return nil, gtserror.Newf("could not unmarshal server response as WellKnownResponse: %w", err)
	}

	// Look through the links for the first one that
	// matches nodeinfo schema, this is what we need.
	var nodeinfoHref *url.URL
	for _, l := range wellKnownResp.Links {
		if l.Href == "" || !strings.HasPrefix(l.Rel, "http://nodeinfo.diaspora.software/ns/schema/2") {
			continue
		}
		nodeinfoHref, err = url.Parse(l.Href)
		if err != nil {
			return nil, gtserror.Newf("couldn't parse url %s: %w", l.Href, err)
		}
	}
	if nodeinfoHref == nil {
		return nil, gtserror.New("could not find nodeinfo rel in well known response")
	}

	return nodeinfoHref, nil
}

func (t *transport) callNodeInfo(
	ctx context.Context,
	iri *url.URL,
	robotsTxt *robotstxt.RobotsData,
) (*apimodel.Nodeinfo, error) {
	// Normalize robots.txt test path.
	testPath := iri.Path
	if !strings.HasPrefix(testPath, "/") {
		testPath = "/" + testPath
	}

	// Bail if we're not allowed to fetch this endpoint.
	if robotsTxt != nil && !robotsTxt.TestAgent(testPath, t.controller.userAgent) {
		err := gtserror.Newf("can't fetch %s: robots.txt disallows it", testPath)
		return nil, gtserror.SetNotPermitted(err)
	}

	// Build IRI just once
	iriStr := iri.String()

	req, err := http.NewRequestWithContext(ctx, "GET", iriStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", string(apiutil.AppJSON))

	resp, err := t.GET(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Ensure a non-error status response.
	if resp.StatusCode != http.StatusOK {
		return nil, gtserror.NewFromResponse(resp)
	}

	// Ensure that the incoming request content-type is expected.
	if ct := resp.Header.Get("Content-Type"); !apiutil.NodeInfo2ContentType(ct) {
		err := gtserror.Newf("non nodeinfo schema 2.0 response: %s", ct)
		return nil, gtserror.SetMalformed(err)
	}

	// Ensure that we can use data returned from this endpoint.
	robots := resp.Header.Values("X-Robots-Tag")
	if slices.ContainsFunc(
		robots,
		func(key string) bool {
			return strings.Contains(key, "noindex")
		},
	) {
		err := gtserror.Newf("can't use fetched %s: robots tags disallows it", iri.Path)
		return nil, gtserror.SetNotPermitted(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if len(b) == 0 {
		return nil, gtserror.New("response bytes was len 0")
	}

	niResp := &apimodel.Nodeinfo{}
	if err := json.Unmarshal(b, niResp); err != nil {
		return nil, gtserror.Newf("could not unmarshal server response as Nodeinfo: %w", err)
	}

	return niResp, nil
}
