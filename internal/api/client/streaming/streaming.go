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

package streaming

import (
	"net/http"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	BasePath            = "/v1/streaming"          // path for the streaming api, minus the 'api' prefix
	StreamQueryKey      = "stream"                 // type of stream being requested
	StreamListKey       = "list"                   // id of list being requested
	StreamTagKey        = "tag"                    // name of tag being requested
	AccessTokenQueryKey = "access_token"           // oauth access token
	AccessTokenHeader   = "Sec-Websocket-Protocol" //nolint:gosec
)

type Module struct {
	processor *processing.Processor
	dTicker   time.Duration
	wsUpgrade websocket.Upgrader
}

func New(processor *processing.Processor, dTicker time.Duration, wsBuf int) *Module {
	// We expect CORS requests for websockets,
	// (via eg., semaphore.social) so be lenient.
	// TODO: make this customizable?
	checkOrigin := func(r *http.Request) bool { return true }

	return &Module{
		processor: processor,
		dTicker:   dTicker,
		wsUpgrade: websocket.Upgrader{
			ReadBufferSize:  wsBuf,
			WriteBufferSize: wsBuf,
			CheckOrigin:     checkOrigin,
		},
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, BasePath, m.StreamGETHandler)
}
