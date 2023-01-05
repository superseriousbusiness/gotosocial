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

package followrequests

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
)

const (
	// IDKey is for account IDs
	IDKey = "id"
	// BasePath is the base path for serving the follow request API, minus the 'api' prefix
	BasePath = "/v1/follow_requests"
	// BasePathWithID is just the base path with the ID key in it.
	// Use this anywhere you need to know the ID of the account that owns the follow request being queried.
	BasePathWithID = BasePath + "/:" + IDKey
	// AuthorizePath is used for authorizing follow requests
	AuthorizePath = BasePathWithID + "/authorize"
	// RejectPath is used for rejecting follow requests
	RejectPath = BasePathWithID + "/reject"
)

type Module struct {
	processor processing.Processor
}

func New(processor processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, BasePath, m.FollowRequestGETHandler)
	attachHandler(http.MethodPost, AuthorizePath, m.FollowRequestAuthorizePOSTHandler)
	attachHandler(http.MethodPost, RejectPath, m.FollowRequestRejectPOSTHandler)
}
