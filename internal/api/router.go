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

package api

import "github.com/gin-gonic/gin"

// Router provides the http routes used by the API
type Router interface {
	Route()
}

// NewRouter returns a new router
func NewRouter() Router {
	return &router{}
}

// router implements the router interface
type router struct {

}

func (r *router) Route() {
	ginRouter := gin.Default()
	ginRouter.LoadHTMLGlob("web/template/*")

	apiGroup := ginRouter.Group("/api")
	{
		v1 := apiGroup.Group("/v1")
		{
			statusesGroup := v1.Group("/statuses")
			{
				statusesGroup.GET(":id", statusGet)
			}

		}
	}
	ginRouter.Run()
}
