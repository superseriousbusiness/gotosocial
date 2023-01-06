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

package streaming

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
)

const (
	// BasePath is the path for the streaming api, minus the 'api' prefix
	BasePath = "/v1/streaming"

	// StreamQueryKey is the query key for the type of stream being requested
	StreamQueryKey = "stream"

	// AccessTokenQueryKey is the query key for an oauth access token that should be passed in streaming requests.
	AccessTokenQueryKey = "access_token"
	// AccessTokenHeader is the header for an oauth access token that can be passed in streaming requests instead of AccessTokenQueryKey
	//nolint:gosec
	AccessTokenHeader = "Sec-Websocket-Protocol"
)

type Module struct {
	processor processing.Processor
	dTicker   time.Duration
	wsUpgrade websocket.Upgrader
}

func New(processor processing.Processor, dTicker time.Duration, wsBuf int) *Module {
	return &Module{
		processor: processor,
		dTicker:   dTicker,
		wsUpgrade: websocket.Upgrader{
			ReadBufferSize:  wsBuf, // we don't expect reads
			WriteBufferSize: wsBuf,

			// we expect cors requests (via eg., pinafore.social) so be lenient
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, BasePath, m.StreamGETHandler)
}
