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

package dereferencing

import (
	"context"
	"encoding/json"
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// fingerRemoteAccount performs a webfinger call for the
// given username and host, using the provided transport.
//
// The webfinger response will be parsed, and the subject
// domain and AP URI will be extracted and returned.
//
// In case the response cannot be parsed, or the response
// does not contain a valid subject string or AP URI, an
// error will be returned instead.
func (d *Dereferencer) fingerRemoteAccount(
	ctx context.Context,
	transport transport.Transport,
	username string,
	host string,
) (
	string, // discovered username
	string, // discovered account domain
	*url.URL, // discovered account URI
	error,
) {
	// Assemble target namestring for logging.
	var target = "@" + username + "@" + host

	b, err := transport.Finger(ctx, username, host)
	if err != nil {
		err = gtserror.Newf("error webfingering %s: %w", target, err)
		return "", "", nil, err
	}

	var resp apimodel.WellKnownResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		err = gtserror.Newf("error parsing response as JSON for %s: %w", target, err)
		return "", "", nil, err
	}

	if len(resp.Links) == 0 {
		err = gtserror.Newf("no links found in response for %s", target)
		return "", "", nil, err
	}

	if resp.Subject == "" {
		err = gtserror.Newf("no subject found in response for %s", target)
		return "", "", nil, err
	}

	accUsername, accDomain, err := util.ExtractWebfingerParts(resp.Subject)
	if err != nil {
		return "", "", nil, gtserror.Newf("error extracting subject parts for %s: %w", target, err)
	} else if accUsername != username {
		return "", "", nil, gtserror.Newf("response username does not match input for %s: %w", target, err)
	}

	// Look through links for the first
	// one that matches what we need:
	//
	//   - Must be self link.
	//   - Must be AP type.
	//   - Valid https/http URI.
	for _, link := range resp.Links {
		if link.Rel != "self" {
			// Not self link, ignore.
			continue
		}

		if !apiutil.ASContentType(link.Type) {
			// Not an AP type, ignore.
			continue
		}

		uri, err := url.Parse(link.Href)
		if err != nil {
			log.Warnf(ctx,
				"invalid href for ActivityPub self link %s for %s: %v",
				link.Href, target, err,
			)

			// Funky URL, ignore.
			continue
		}

		if uri.Scheme != "http" && uri.Scheme != "https" {
			log.Warnf(ctx,
				"invalid href for ActivityPub self link %s for %s: schema must be http or https",
				link.Href, target,
			)

			// Can't handle this
			// schema, ignore.
			continue
		}

		// All looks good, return happily!
		return accUsername, accDomain, uri, nil
	}

	return "", "", nil, gtserror.Newf("no suitable self, AP-type link found in webfinger response for %s", target)
}
