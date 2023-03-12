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

package nodeinfo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// NodeInfoWellKnownGETHandler swagger:operation GET /.well-known/nodeinfo nodeInfoWellKnownGet
//
// Returns a well-known response which redirects callers to `/nodeinfo/2.0`.
//
// eg. `{"links":[{"rel":"http://nodeinfo.diaspora.software/ns/schema/2.0","href":"http://example.org/nodeinfo/2.0"}]}`
// See: https://nodeinfo.diaspora.software/protocol.html
//
//	---
//	tags:
//	- .well-known
//
//	produces:
//	- application/json
//
//	responses:
//		'200':
//			schema:
//				"$ref": "#/definitions/wellKnownResponse"
func (m *Module) NodeInfoWellKnownGETHandler(c *gin.Context) {
	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Fedi().NodeInfoRelGet(c.Request.Context())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	c.JSON(http.StatusOK, resp)
}
