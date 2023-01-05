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

package media

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// MediaGETHandler swagger:operation GET /api/v1/media/{id} mediaGet
//
// Get a media attachment that you own.
//
//	---
//	tags:
//	- media
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		description: id of the attachment
//		type: string
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- read:media
//
//	responses:
//		'200':
//			description: The requested media attachment.
//			schema:
//				"$ref": "#/definitions/attachment"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//		   description: internal server error
func (m *Module) MediaGETHandler(c *gin.Context) {
	if apiVersion := c.Param(APIVersionKey); apiVersion != APIv1 {
		err := errors.New("api version must be one v1 for this path")
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err, err.Error()), m.processor.InstanceGet)
		return
	}

	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	attachmentID := c.Param(IDKey)
	if attachmentID == "" {
		err := errors.New("no attachment id specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	attachment, errWithCode := m.processor.MediaGet(c.Request.Context(), authed, attachmentID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, attachment)
}
