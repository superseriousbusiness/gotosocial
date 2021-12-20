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

package router

import "github.com/gin-gonic/gin"

// AttachHandler attaches the given gin.HandlerFunc to the router with the specified method and path.
// If the path is set to ANY, then the handlerfunc will be used for ALL methods at its given path.
func (r *router) AttachHandler(method string, path string, handler gin.HandlerFunc) {
	if method == "ANY" {
		r.engine.Any(path, handler)
	} else {
		r.engine.Handle(method, path, handler)
	}
}

// AttachMiddleware attaches a gin middleware to the router that will be used globally
func (r *router) AttachMiddleware(middleware gin.HandlerFunc) {
	r.engine.Use(middleware)
}

// AttachNoRouteHandler attaches a gin.HandlerFunc to NoRoute to handle 404's
func (r *router) AttachNoRouteHandler(handler gin.HandlerFunc) {
	r.engine.NoRoute(handler)
}
