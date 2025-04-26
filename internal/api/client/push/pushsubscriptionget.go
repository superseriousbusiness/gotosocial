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

package push

import (
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"github.com/gin-gonic/gin"
)

// PushSubscriptionGETHandler swagger:operation GET /api/v1/push/subscription pushSubscriptionGet
//
// Get the push subscription for the current access token.
//
//	---
//	tags:
//	- push
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- push
//
//	responses:
//		'200':
//			description: Web Push subscription for current access token.
//			schema:
//				"$ref": "#/definitions/webPushSubscription"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'404':
//			description: This access token doesn't have an associated subscription.
//		'500':
//			description: internal server error
func (m *Module) PushSubscriptionGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopePush,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiSubscription, errWithCode := m.processor.Push().Get(c, authed.Token.GetAccess())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiSubscription)
}
