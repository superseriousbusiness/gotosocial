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

package media

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
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
	if _, errWithCode := apiutil.ParseAPIVersion(
		c.Param(apiutil.APIVersionKey),
		[]string{apiutil.APIv1, apiutil.APIv2}...,
	); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		// This takes write even
		// though it's a read.
		apiutil.ScopeWriteMedia,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	attachmentID := c.Param(IDKey)
	if attachmentID == "" {
		err := errors.New("no attachment id specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	attachment, errWithCode := m.processor.Media().Get(c.Request.Context(), authed.Account, attachmentID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, attachment)
}
