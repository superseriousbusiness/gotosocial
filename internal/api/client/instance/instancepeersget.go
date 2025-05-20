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

package instance

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"

	"github.com/gin-gonic/gin"
)

// InstancePeersGETHandler swagger:operation GET /api/v1/instance/peers instancePeersGet
//
// List peer domains.
//
//	---
//	tags:
//	- instance
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: filter
//		type: string
//		description: |-
//			Comma-separated list of filters to apply to results. Recognized filters are:
//				- `open` -- include known domains that are not in the domain blocklist
//				- `allowed` -- include domains that are in the domain allowlist
//				- `blocked` -- include domains that are in the domain blocklist
//				- `suspended` -- DEPRECATED! Use `blocked` instead. Same as `blocked`: include domains that are in the domain blocklist;
//
//			If filter is `open`, only domains that aren't in the blocklist will be shown.
//
//			If filter is `blocked`, only domains that *are* in the blocklist will be shown.
//
//			If filter is `allowed`, only domains that are in the allowlist will be shown.
//
//			If filter is `open,blocked`, then blocked domains and known domains not on the blocklist will be shown.
//
//			If filter is `open,allowed`, then allowed domains and known domains not on the blocklist will be shown.
//
//			If filter is an empty string or not set, then `open` will be assumed as the default.
//		in: query
//		required: false
//		default: flat
//	-
//		name: flat
//		type: boolean
//		description: If true, a "flat" array of strings will be returned corresponding to just domain names.
//		in: query
//		required: false
//		default: false
//
//	security:
//	- OAuth2 Bearer: []
//
//	responses:
//		'200':
//			description: >-
//				If no filter parameter is provided, or filter is empty, then a legacy,
//				Mastodon-API compatible response will be returned. This will consist of
//				just a 'flat' array of strings like `["example.com", "example.org"]`,
//				which corresponds to setting a filter of `open` and flat=true.
//
//				If a filter parameter is provided and flat is not true, then an array
//				of objects with at least a `domain` key set on each object will be returned.
//
//				Domains that are silenced or suspended will also have a key
//				`suspended_at` or `silenced_at` that contains an iso8601 date string.
//				If one of these keys is not present on the domain object, it is open.
//				Suspended instances may in some cases be obfuscated, which means they
//				will have some letters replaced by `*` to make it more difficult for
//				bad actors to target instances with harassment.
//
//				Whether a flat response or a more detailed response is returned, domains
//				will be sorted alphabetically by hostname.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/domain"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) InstancePeersGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		false, false, false, false,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	isUnauthenticated := authed.Account == nil || authed.User == nil

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	var (
		includeBlocked bool
		includeAllowed bool
		includeOpen    bool
		flatten        bool
	)

	if filterParam := c.Query(PeersFilterKey); filterParam != "" {
		filters := strings.Split(filterParam, ",")
		for _, f := range filters {
			trimmed := strings.TrimSpace(f)
			switch {
			case strings.EqualFold(trimmed, "blocked") || strings.EqualFold(trimmed, "suspended"):
				includeBlocked = true
			case strings.EqualFold(trimmed, "allowed"):
				includeAllowed = true
			case strings.EqualFold(trimmed, "open"):
				includeOpen = true
			default:
				err := fmt.Errorf("filter %s not recognized; accepted values are 'open', 'blocked', 'allowed', and 'suspended' (deprecated)", trimmed)
				apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
				return
			}
		}
	} else {
		// Default is to only include open domains, and present
		// them in a 'flat' manner (just an array of strings),
		// to maintain compatibility with the Mastodon API.
		includeOpen = true
		flatten = true
	}

	if includeBlocked && isUnauthenticated && !config.GetInstanceExposeBlocklist() {
		const errText = "peers blocked query requires an authenticated account/user"
		errWithCode := gtserror.NewErrorUnauthorized(errors.New(errText), errText)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if includeAllowed && isUnauthenticated && !config.GetInstanceExposeAllowlist() {
		const errText = "peers allowed query requires an authenticated account/user"
		errWithCode := gtserror.NewErrorUnauthorized(errors.New(errText), errText)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if includeOpen && isUnauthenticated && !config.GetInstanceExposePeers() {
		const errText = "peers open query requires an authenticated account/user"
		errWithCode := gtserror.NewErrorUnauthorized(errors.New(errText), errText)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if includeBlocked && includeAllowed {
		const errText = "cannot include blocked + allowed filters at the same time"
		errWithCode := gtserror.NewErrorBadRequest(errors.New(errText), errText)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if flatStr := c.Query(PeersFlatKey); flatStr != "" {
		var err error
		flatten, err = strconv.ParseBool(flatStr)
		if err != nil {
			err := fmt.Errorf("error parsing 'flat' key as boolean: %w", err)
			errWithCode := gtserror.NewErrorBadRequest(err, err.Error())
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
			return
		}
	}

	data, errWithCode := m.processor.InstancePeersGet(
		c.Request.Context(),
		includeBlocked,
		includeAllowed,
		includeOpen,
		flatten,
		false, // Don't include severity.
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, data)
}
