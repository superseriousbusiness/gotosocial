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
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/api/health"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"github.com/gin-gonic/gin"
)

type Health struct {
	health *health.Module
}

func (mt *Health) Route(r *router.Router, m ...gin.HandlerFunc) {
	// Create new group on top level prefix.
	healthGroup := r.AttachGroup("")
	healthGroup.Use(m...)
	healthGroup.Use(
		middleware.CacheControl(middleware.CacheControlConfig{
			// Never cache health responses.
			Directives: []string{"no-store"},
		}),
	)

	mt.health.Route(healthGroup.Handle)
}

func NewHealth(readyF func(context.Context) error) *Health {
	return &Health{
		health: health.New(readyF),
	}
}
