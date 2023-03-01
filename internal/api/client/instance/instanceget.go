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

package instance

import (
	"net/http"

	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"

	"github.com/gin-gonic/gin"
)

// InstanceInformationV1GETHandlerV1 swagger:operation GET /api/v1/instance instanceGetV1
//
// View instance information.
//
//	---
//	tags:
//	- instance
//
//	produces:
//	- application/json
//
//	responses:
//		'200':
//			description: "Instance information."
//			schema:
//				"$ref": "#/definitions/instanceV1"
//		'406':
//			description: not acceptable
//		'500':
//			description: internal error
func (m *Module) InstanceInformationGETHandlerV1(c *gin.Context) {
	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	instance, errWithCode := m.processor.InstanceGetV1(c.Request.Context())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	c.JSON(http.StatusOK, instance)
}

// InstanceInformationGETHandlerV2 swagger:operation GET /api/v2/instance instanceGetV2
//
// View instance information.
//
//	---
//	tags:
//	- instance
//
//	produces:
//	- application/json
//
//	responses:
//		'200':
//			description: "Instance information."
//			schema:
//				"$ref": "#/definitions/instanceV2"
//		'406':
//			description: not acceptable
//		'500':
//			description: internal error
func (m *Module) InstanceInformationGETHandlerV2(c *gin.Context) {
	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	instance, errWithCode := m.processor.InstanceGetV2(c.Request.Context())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	c.JSON(http.StatusOK, instance)
}
