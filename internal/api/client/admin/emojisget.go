/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// EmojisGETHandler swagger:operation GET /api/v1/admin/custom_emojis emojisGet
//
// View local and remote emojis available to / known by this instance.
//
// The next and previous queries can be parsed from the returned Link header.
// Example:
//
// `<http://localhost:8080/api/v1/admin/custom_emojis?limit=30&max_shortcode_domain=yell@fossbros-anonymous.io&filter=domain:all>; rel="next", <http://localhost:8080/api/v1/admin/custom_emojis?limit=30&min_shortcode_domain=rainbow@&filter=domain:all>; rel="prev"`
//
//	---
//	tags:
//	- admin
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
//
//			`domain:[domain]` -- show emojis from the given domain, eg `?filter=domain:example.org` will show emojis from `example.org` only.
//			Instead of giving a specific domain, you can also give either one of the key words `local` or `all` to show either local emojis only (`domain:local`) or show all emojis from all domains (`domain:all`).
//			Note: `domain:*` is equivalent to `domain:all` (including local).
//			If no domain filter is provided, `domain:all` will be assumed.
//
//			`disabled` -- include emojis that have been disabled.
//
//			`enabled` -- include emojis that are enabled.
//
//			`shortcode:[shortcode]` -- show only emojis with the given shortcode, eg `?filter=shortcode:blob_cat_uwu` will show only emojis with the shortcode `blob_cat_uwu` (case sensitive).
//
//			If neither `disabled` or `enabled` are provided, both disabled and enabled emojis will be shown.
//
//			If no filter query string is provided, the default `domain:all` will be used, which will show all emojis from all domains.
//		in: query
//		required: false
//		default: "domain:all"
//	-
//		name: limit
//		type: integer
//		description: Number of emojis to return. Less than 1, or not set, means unlimited (all emojis).
//		default: 50
//		in: query
//	-
//		name: max_shortcode_domain
//		type: string
//		description: >-
//			Return only emojis with `[shortcode]@[domain]` *LOWER* (alphabetically) than given `[shortcode]@[domain]`.
//			For example, if `max_shortcode_domain=beep@example.org`, then returned values might include emojis with
//			`[shortcode]@[domain]`s like `car@example.org`, `debian@aaa.com`, `test@` (local emoji), etc.
//
//			Emoji with the given `[shortcode]@[domain]` will not be included in the result set.
//		in: query
//	-
//		name: min_shortcode_domain
//		type: string
//		description: >-
//			Return only emojis with `[shortcode]@[domain]` *HIGHER* (alphabetically) than given `[shortcode]@[domain]`.
//			For example, if `max_shortcode_domain=beep@example.org`, then returned values might include emojis with
//			`[shortcode]@[domain]`s like `arse@test.com`, `0101_binary@hackers.net`, `bee@` (local emoji), etc.
//
//			Emoji with the given `[shortcode]@[domain]` will not be included in the result set.
//		in: query
//
//	responses:
//		'200':
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
//			description: An array of emojis, arranged alphabetically by shortcode and domain.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/adminEmoji"
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
func (m *Module) EmojisGETHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	maxShortcodeDomain := c.Query(MaxShortcodeDomainKey)
	minShortcodeDomain := c.Query(MinShortcodeDomainKey)

	limit := 50
	limitString := c.Query(LimitKey)
	if limitString != "" {
		i, err := strconv.ParseInt(limitString, 10, 32)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", LimitKey, err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
		limit = int(i)
	}
	if limit < 0 {
		limit = 0
	}

	var domain string
	var includeDisabled bool
	var includeEnabled bool
	var shortcode string
	if filterParam := c.Query(FilterQueryKey); filterParam != "" {
		filters := strings.Split(filterParam, ",")
		for _, filter := range filters {
			lower := strings.ToLower(filter)
			switch {
			case strings.HasPrefix(lower, "domain:"):
				domain = strings.TrimPrefix(lower, "domain:")
			case lower == "disabled":
				includeDisabled = true
			case lower == "enabled":
				includeEnabled = true
			case strings.HasPrefix(lower, "shortcode:"):
				shortcode = strings.Trim(filter[10:], ":") // remove any errant ":"
			default:
				err := fmt.Errorf("filter %s not recognized; accepted values are 'domain:[domain]', 'disabled', 'enabled', 'shortcode:[shortcode]'", filter)
				apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
				return
			}
		}
	}

	if domain == "" {
		// default is to show all domains
		domain = db.EmojiAllDomains
	} else if domain == "local" || domain == config.GetHost() || domain == config.GetAccountDomain() {
		// pass empty string for local domain
		domain = ""
	}

	// normalize filters
	if !includeDisabled && !includeEnabled {
		// include both if neither specified
		includeDisabled = true
		includeEnabled = true
	}

	resp, errWithCode := m.processor.AdminEmojisGet(c.Request.Context(), authed, domain, includeDisabled, includeEnabled, shortcode, maxShortcodeDomain, minShortcodeDomain, limit)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}
	c.JSON(http.StatusOK, resp.Items)
}
