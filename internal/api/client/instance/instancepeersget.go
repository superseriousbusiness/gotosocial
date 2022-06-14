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

package instance

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"

	"github.com/gin-gonic/gin"
)

// InstancePeersGETHandler swagger:operation GET /api/v1/instance/peers instancePeersGet
//
// ---
// tags:
// - instance
//
// produces:
// - application/json
//
// parameters:
// - name: filter
//   type: string
//   description: |-
//     Comma-separated list of filters to apply to results. Recognized values are:
//     'open' -- include peers that are not suspended or silenced
//     'suspended' -- include peers that have been suspended.
//     If filter is 'open', only instances that haven't been suspended or silenced will be returned.
//     If filter is 'suspended', only suspended instances will be shown.
//     If filter is 'open,suspended', then all known instances will be returned.
//     If filter is an empty string or not set, then 'open' will be assumed as the default.
//   in: query
//   required: false
//
// responses:
//   '200':
//     description: |-
//       An array of objects with at least a `domain` key will be returned.
//       Domains that are silenced or suspended will also have a key
//       'suspended_at' or 'silenced_at' that contains an iso8601 date string.
//       If one of these keys is not present on the domain object, it is open.
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/domain"
//   '400':
//      description: bad request
//   '401':
//      description: unauthorized
//   '403':
//      description: forbidden
//   '404':
//      description: not found
//   '406':
//      description: not acceptable
//   '500':
//      description: internal server error
func (m *Module) InstancePeersGETHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	var includeSuspended bool
	var includeOpen bool
	if filterParam := c.Query(PeersFilterKey); filterParam != "" {
		filters := strings.Split(filterParam, ",")
		for _, f := range filters {
			trimmed := strings.TrimSpace(f)
			switch {
			case strings.EqualFold(trimmed, "suspended"):
				includeSuspended = true
			case strings.EqualFold(trimmed, "open"):
				includeOpen = true
			default:
				err := fmt.Errorf("filter %s not recognized; accepted values are 'open', 'suspended'", trimmed)
				api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
				return
			}
		}
	} else {
		// default is to only include open domains
		includeOpen = true
	}

	data, errWithCode := m.processor.InstancePeersGet(c.Request.Context(), authed, includeSuspended, includeOpen)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, data)
}
