/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package notification

import (
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	// IDKey is for notification UUIDs
	IDKey = "id"
	// BasePath is the base path for serving the notification API
	BasePath = "/api/v1/notifications"
	// BasePathWithID is just the base path with the ID key in it.
	// Use this anywhere you need to know the ID of the notification being queried.
	BasePathWithID = BasePath + "/:" + IDKey

	// MaxIDKey is the url query for setting a max notification ID to return
	MaxIDKey = "max_id"
	// LimitKey is for specifying maximum number of notifications to return.
	LimitKey = "limit"
	// SinceIDKey is for specifying the minimum notification ID to return.
	SinceIDKey = "since_id"
)

// Module implements the ClientAPIModule interface for every related to posting/deleting/interacting with notifications
type Module struct {
	config    *config.Config
	processor processing.Processor
}

// New returns a new notification module
func New(config *config.Config, processor processing.Processor) api.ClientModule {
	return &Module{
		config:    config,
		processor: processor,
	}
}

// Route attaches all routes from this module to the given router
func (m *Module) Route(r router.Router) error {
	r.AttachHandler(http.MethodGet, BasePath, m.NotificationsGETHandler)
	return nil
}
