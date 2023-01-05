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
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"

	"github.com/gin-gonic/gin"
)

// InstanceInformationGETHandler swagger:operation GET /api/v1/instance instanceGet
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
//				"$ref": "#/definitions/instance"
//		'406':
//			description: not acceptable
//		'500':
//			description: internal error
func (m *Module) InstanceInformationGETHandler(c *gin.Context) {
	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	instance, errWithCode := m.processor.InstanceGet(c.Request.Context(), config.GetHost())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, instance)
}
