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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

func (t *transport) DereferenceInstance(ctx context.Context, iri *url.URL) (*gtsmodel.Instance, error) {
	var i *gtsmodel.Instance
	var err error

	// First try to dereference using /api/v1/instance.
	// This will provide the most complete picture of an instance, and avoid unnecessary api calls.
	//
	// This will only work with Mastodon-api compatible instances: Mastodon, some Pleroma instances, GoToSocial.
	log.Debugf("trying to dereference instance %s by /api/v1/instance", iri.Host)
	i, err = dereferenceByAPIV1Instance(ctx, t, iri)
	if err == nil {
		log.Debugf("successfully dereferenced instance using /api/v1/instance")
		return i, nil
	}
	log.Debugf("couldn't dereference instance using /api/v1/instance: %s", err)

	// If that doesn't work, try to dereference using /.well-known/nodeinfo.
	// This will involve two API calls and return less info overall, but should be more widely compatible.
	log.Debugf("trying to dereference instance %s by /.well-known/nodeinfo", iri.Host)
	i, err = dereferenceByNodeInfo(ctx, t, iri)
	if err == nil {
		log.Debugf("successfully dereferenced instance using /.well-known/nodeinfo")
		return i, nil
	}
	log.Debugf("couldn't dereference instance using /.well-known/nodeinfo: %s", err)

	// we couldn't dereference the instance using any of the known methods, so just return a minimal representation
	log.Debugf("returning minimal representation of instance %s", iri.Host)
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

func dereferenceByAPIV1Instance(ctx context.Context, t *transport, iri *url.URL) (*gtsmodel.Instance, error) {
	cleanIRI := &url.URL{
		Scheme: iri.Scheme,
		Host:   iri.Host,
		Path:   "api/v1/instance",
	}

	// Build IRI just once
	iriStr := cleanIRI.String()

	req, err := http.NewRequestWithContext(ctx, "GET", iriStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", string(apiutil.AppJSON))
	req.Header.Set("Host", cleanIRI.Host)

	resp, err := t.GET(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET request to %s failed (%d): %s", iriStr, resp.StatusCode, resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if len(b) == 0 {
		return nil, errors.New("response bytes was len 0")
	}

	// try to parse the returned bytes directly into an Instance model
	apiResp := &apimodel.Instance{}
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
		URI:                    fmt.Sprintf("%s://%s", iri.Scheme, iri.Host),
		ShortDescription:       apiResp.ShortDescription,
		Description:            apiResp.Description,
		ContactEmail:           apiResp.Email,
		ContactAccountUsername: contactUsername,
		Version:                apiResp.Version,
	}

	return i, nil
}

func dereferenceByNodeInfo(c context.Context, t *transport, iri *url.URL) (*gtsmodel.Instance, error) {
	niIRI, err := callNodeInfoWellKnown(c, t, iri)
	if err != nil {
		return nil, fmt.Errorf("dereferenceByNodeInfo: error during initial call to well-known nodeinfo: %s", err)
	}

	ni, err := callNodeInfo(c, t, niIRI)
	if err != nil {
		return nil, fmt.Errorf("dereferenceByNodeInfo: error doing second call to nodeinfo uri %s: %s", niIRI.String(), err)
	}

	// we got a response of some kind! take what we can from it...
	id, err := id.NewRandomULID()
	if err != nil {
		return nil, fmt.Errorf("dereferenceByNodeInfo: error creating new id for instance %s: %s", iri.Host, err)
	}

	// this is the bare minimum instance we'll return, and we'll add more stuff to it if we can
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

func callNodeInfoWellKnown(ctx context.Context, t *transport, iri *url.URL) (*url.URL, error) {
	cleanIRI := &url.URL{
		Scheme: iri.Scheme,
		Host:   iri.Host,
		Path:   ".well-known/nodeinfo",
	}

	// Build IRI just once
	iriStr := cleanIRI.String()

	req, err := http.NewRequestWithContext(ctx, "GET", iriStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", string(apiutil.AppJSON))
	req.Header.Set("Host", cleanIRI.Host)

	resp, err := t.GET(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("callNodeInfoWellKnown: GET request to %s failed (%d): %s", iriStr, resp.StatusCode, resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if len(b) == 0 {
		return nil, errors.New("callNodeInfoWellKnown: response bytes was len 0")
	}

	wellKnownResp := &apimodel.WellKnownResponse{}
	if err := json.Unmarshal(b, wellKnownResp); err != nil {
		return nil, fmt.Errorf("callNodeInfoWellKnown: could not unmarshal server response as WellKnownResponse: %s", err)
	}

	// look through the links for the first one that matches the nodeinfo schema, this is what we need
	var nodeinfoHref *url.URL
	for _, l := range wellKnownResp.Links {
		if l.Href == "" || !strings.HasPrefix(l.Rel, "http://nodeinfo.diaspora.software/ns/schema/2") {
			continue
		}
		nodeinfoHref, err = url.Parse(l.Href)
		if err != nil {
			return nil, fmt.Errorf("callNodeInfoWellKnown: couldn't parse url %s: %s", l.Href, err)
		}
	}
	if nodeinfoHref == nil {
		return nil, errors.New("callNodeInfoWellKnown: could not find nodeinfo rel in well known response")
	}

	return nodeinfoHref, nil
}

func callNodeInfo(ctx context.Context, t *transport, iri *url.URL) (*apimodel.Nodeinfo, error) {
	// Build IRI just once
	iriStr := iri.String()

	req, err := http.NewRequestWithContext(ctx, "GET", iriStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", string(apiutil.AppJSON))
	req.Header.Set("Host", iri.Host)

	resp, err := t.GET(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("callNodeInfo: GET request to %s failed (%d): %s", iriStr, resp.StatusCode, resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if len(b) == 0 {
		return nil, errors.New("callNodeInfo: response bytes was len 0")
	}

	niResp := &apimodel.Nodeinfo{}
	if err := json.Unmarshal(b, niResp); err != nil {
		return nil, fmt.Errorf("callNodeInfo: could not unmarshal server response as Nodeinfo: %s", err)
	}

	return niResp, nil
}
