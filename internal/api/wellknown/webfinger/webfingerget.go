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

package webfinger

import (
	"errors"
	"fmt"
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/gin-gonic/gin"
)

// WebfingerGETRequest swagger:operation GET /.well-known/webfinger webfingerGet
//
// Handles webfinger account lookup requests.
//
// For example, a GET to `https://goblin.technology/.well-known/webfinger?resource=acct:tobi@goblin.technology` would return:
//
// ```
//
//	{"subject":"acct:tobi@goblin.technology","aliases":["https://goblin.technology/users/tobi","https://goblin.technology/@tobi"],"links":[{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"https://goblin.technology/@tobi"},{"rel":"self","type":"application/activity+json","href":"https://goblin.technology/users/tobi"}]}
//
// ```
//
// See: https://webfinger.net/
//
//	---
//	tags:
//	- .well-known
//
//	produces:
//	- application/jrd+json
//
//	responses:
//		'200':
//			schema:
//				"$ref": "#/definitions/wellKnownResponse"
func (m *Module) WebfingerGETRequest(c *gin.Context) {
	if _, err := apiutil.NegotiateAccept(c, apiutil.WebfingerJSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	resourceQuery, set := c.GetQuery("resource")
	if !set || resourceQuery == "" {
		err := errors.New("no 'resource' in request query")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	requestedUsername, requestedHost, err := util.ExtractWebfingerParts(resourceQuery)
	if err != nil {
		err := fmt.Errorf("bad webfinger request with resource query %s: %w", resourceQuery, err)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if requestedHost != config.GetHost() && requestedHost != config.GetAccountDomain() {
		err := fmt.Errorf("requested host %s does not belong to this instance", requestedHost)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Fedi().WebfingerGet(c.Request.Context(), requestedUsername)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Encode JSON HTTP response.
	apiutil.EncodeJSONResponse(
		c.Writer,
		c.Request,
		http.StatusOK,
		apiutil.AppJRDJSON,
		resp,
	)
}
