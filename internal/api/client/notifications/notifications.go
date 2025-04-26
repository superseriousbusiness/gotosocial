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

package notifications

import (
	"net/http"

	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"github.com/gin-gonic/gin"
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

	// TypesKey names an array param specifying notification types to include.
	TypesKey = "types[]"
	// ExcludeTypesKey names an array param specifying notification types to exclude.
	ExcludeTypesKey = "exclude_types[]"
	MaxIDKey        = "max_id"
	LimitKey        = "limit"
	SinceIDKey      = "since_id"
	MinIDKey        = "min_id"
)

type Module struct {
	processor *processing.Processor
}

func New(processor *processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, BasePath, m.NotificationsGETHandler)
	attachHandler(http.MethodGet, BasePathWithID, m.NotificationGETHandler)
	attachHandler(http.MethodPost, BasePathWithClear, m.NotificationsClearPOSTHandler)
}
