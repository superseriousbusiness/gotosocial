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
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

// InstanceDomainBlocksGETHandler swagger:operation GET /api/v1/instance/domain_blocks instanceDomainBlocksGet
//
// List blocked domains.
//
// OAuth token may need to be provided depending on setting `instance-expose-blocklist`.
//
//	---
//	tags:
//	- instance
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer: []
//
//	responses:
//		'200':
//			description: List of blocked domains.
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
func (m *Module) InstanceDomainBlocksGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		false, false, false, false,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if (authed.Account == nil || authed.User == nil) && !config.GetInstanceExposeBlocklist() {
		const errText = "domain blocks endpoint requires an authenticated account/user"
		errWithCode := gtserror.NewErrorUnauthorized(errors.New(errText), errText)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	data, errWithCode := m.processor.InstancePeersGet(
		c.Request.Context(),
		true,  // Include blocked.
		false, // Don't include allowed.
		false, // Don't include open.
		false, // Don't flatten.
		true,  // Include severity.
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, data)
}

// InstanceDomainAllowsGETHandler swagger:operation GET /api/v1/instance/domain_allows instanceDomainAllowsGet
//
// List explicitly allowed domains.
//
// OAuth token may need to be provided depending on setting `instance-expose-allowlist`.
//
//	---
//	tags:
//	- instance
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer: []
//
//	responses:
//		'200':
//			description: List of explicitly allowed domains.
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
func (m *Module) InstanceDomainAllowsGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		false, false, false, false,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if (authed.Account == nil || authed.User == nil) && !config.GetInstanceExposeAllowlist() {
		const errText = "domain allows endpoint requires an authenticated account/user"
		errWithCode := gtserror.NewErrorUnauthorized(errors.New(errText), errText)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	data, errWithCode := m.processor.InstancePeersGet(
		c.Request.Context(),
		false, // Don't include blocked.
		true,  // Include allowed.
		false, // Don't include open.
		false, // Don't flatten.
		false, // Don't include severity.
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, data)
}
