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

package hostmeta

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// HostMetaGETHandler swagger:operation GET /.well-known/host-meta hostMetaGet
//
// Returns a compliant hostmeta response to web host metadata queries.
//
// See: https://www.rfc-editor.org/rfc/rfc6415.html
//
//	---
//	tags:
//	- .well-known
//
//	produces:
//	- application/xrd+xml"
//
//	responses:
//		'200':
//			schema:
//				"$ref": "#/definitions/hostmeta"
func (m *Module) HostMetaGETHandler(c *gin.Context) {
	if _, err := apiutil.NegotiateAccept(c, apiutil.HostMetaHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	hostMeta := m.processor.Fedi().HostMetaGet()

	// this setup with a separate buffer we encode into is used because
	// xml.Marshal does not emit xml.Header by itself
	var b bytes.Buffer
	data := bufio.NewWriter(&b)
	if _, err := data.Write([]byte(xml.Header)); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGetV1)
		return
	}

	enc := xml.NewEncoder(data)
	if err := enc.Encode(hostMeta); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGetV1)
		return
	}

	c.Data(http.StatusOK, HostMetaContentType, b.Bytes())
}
