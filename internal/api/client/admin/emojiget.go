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

package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// EmojisGETHandler swagger:operation GET /api/v1/admin/custom_emojis emojisGet
//
// View local and remote emojis available to / known by this instance.
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
//
//	responses:
//		'200':
//			description:
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
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		api.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	var domain string
	var includeDisabled bool
	var includeEnabled bool
	var shortcode string
	if filterParam := c.Query(FilterQueryKey); filterParam != "" {
		filters := strings.Split(filterParam, ",")
		for _, f := range filters {
			trimmed := strings.TrimSpace(f)
			trimmedLower := strings.ToLower(trimmed)
			switch {
			case strings.HasPrefix(trimmedLower, "domain:"):
				domain = strings.TrimPrefix(trimmedLower, "domain:")
				// if the part after `domain:` is an empty
				// string, assume the caller wants all
				if domain == "" {
					domain = db.EmojiAllDomains
				}
			case trimmedLower == "disabled":
				includeDisabled = true
			case trimmedLower == "enabled":
				includeEnabled = true
			case strings.HasPrefix(trimmedLower, "shortcode:"):
				shortcode = strings.Trim(trimmed[10:], ":") // remove any errant ":"
			default:
				err := fmt.Errorf("filter %s not recognized; accepted values are 'domain:[domain]', 'disabled', 'enabled', 'shortcode:[shortcode]'", trimmed)
				api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
				return
			}
		}
	} else {
		// default is to show all domains
		domain = db.EmojiAllDomains
	}

	// normalize filters
	if !includeDisabled && !includeEnabled {
		// include both if neither specified
		includeDisabled = true
		includeEnabled = true
	}
	if domain == "local" || domain == config.GetHost() || domain == config.GetAccountDomain() {
		// pass empty string for local domain
		domain = ""
	}

	apiEmojis, errWithCode := m.processor.AdminEmojisGet(c.Request.Context(), authed, domain, includeDisabled, includeEnabled, shortcode)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, apiEmojis)
}
