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

package timelines

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
)

const (
	// BasePath is the base URI path for serving timelines, minus the 'api' prefix.
	BasePath = "/v1/timelines"
	// HomeTimeline is the path for the home timeline
	HomeTimeline = BasePath + "/home"
	// PublicTimeline is the path for the public (and public local) timeline
	PublicTimeline = BasePath + "/public"
	// MaxIDKey is the url query for setting a max status ID to return
	MaxIDKey = "max_id"
	// SinceIDKey is the url query for returning results newer than the given ID
	SinceIDKey = "since_id"
	// MinIDKey is the url query for returning results immediately newer than the given ID
	MinIDKey = "min_id"
	// LimitKey is for specifying maximum number of results to return.
	LimitKey = "limit"
	// LocalKey is for specifying whether only local statuses should be returned
	LocalKey = "local"
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
	attachHandler(http.MethodGet, HomeTimeline, m.HomeTimelineGETHandler)
	attachHandler(http.MethodGet, PublicTimeline, m.PublicTimelineGETHandler)
}
