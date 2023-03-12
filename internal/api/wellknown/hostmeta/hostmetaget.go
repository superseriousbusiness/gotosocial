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

package hostmeta

import (
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
	var buf bytes.Buffer

	// Preallocate buffer of reasonable length.
	buf.Grow(len(xml.Header) + 64)

	// No need to check for error on write to buffer.
	_, _ = buf.WriteString(xml.Header)

	// Encode host-meta as XML to in-memory buffer.
	if err := xml.NewEncoder(&buf).Encode(hostMeta); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGetV1)
		return
	}

	c.Data(http.StatusOK, HostMetaContentType, buf.Bytes())
}
