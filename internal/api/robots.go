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

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/robots"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

type Robots struct {
	robots *robots.Module
}

func (rb *Robots) Route(r *router.Router, m ...gin.HandlerFunc) {
	// Create a group so we can attach middlewares.
	robotsGroup := r.AttachGroup("robots.txt")

	// Use passed-in middlewares.
	robotsGroup.Use(m...)

	// Allow caching for 24 hrs.
	// https://www.rfc-editor.org/rfc/rfc9309.html#section-2.4
	robotsGroup.Use(
		middleware.CacheControl(middleware.CacheControlConfig{
			Directives: []string{"public", "no-cache"},
			Vary:       []string{"Accept-Encoding"},
		}),
	)

	rb.robots.Route(robotsGroup.Handle)
}

func NewRobots() *Robots {
	return &Robots{}
}
