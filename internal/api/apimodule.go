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

import (
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

// ClientModule represents a chunk of code (usually contained in a single package) that adds a set
// of functionalities and/or side effects to a router, by mapping routes and/or middlewares onto it--in other words, a REST API ;)
// A ClientAPIMpdule with routes corresponds roughly to one main path of the gotosocial REST api, for example /api/v1/accounts/ or /oauth/
type ClientModule interface {
	Route(s router.Router) error
}

// FederationModule represents a chunk of code (usually contained in a single package) that adds a set
// of functionalities and/or side effects to a router, by mapping routes and/or middlewares onto it--in other words, a REST API ;)
// Unlike ClientAPIModule, federation API module is not intended to be interacted with by clients directly -- it is primarily a server-to-server interface.
type FederationModule interface {
	Route(s router.Router) error
}
