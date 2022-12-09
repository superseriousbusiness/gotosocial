/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package notifications

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
)

const (
	// IDKey is for notification UUIDs
	IDKey = "id"
	// BasePath is the base path for serving the notification API, minus the 'api' prefix.
	BasePath = "/v1/notifications"
	// BasePathWithID is just the base path with the ID key in it.
	// Use this anywhere you need to know the ID of the notification being queried.
	BasePathWithID    = BasePath + "/:" + IDKey
	BasePathWithClear = BasePath + "/clear"

	// ExcludeTypes is an array specifying notification types to exclude
	ExcludeTypesKey = "exclude_types[]"
	// MaxIDKey is the url query for setting a max notification ID to return
	MaxIDKey = "max_id"
	// LimitKey is for specifying maximum number of notifications to return.
	LimitKey = "limit"
	// SinceIDKey is for specifying the minimum notification ID to return.
	SinceIDKey = "since_id"
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
	attachHandler(http.MethodGet, BasePath, m.NotificationsGETHandler)
	attachHandler(http.MethodPost, BasePathWithClear, m.NotificationsClearPOSTHandler)
}
