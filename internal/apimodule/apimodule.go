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

// Package apimodule is basically a wrapper for a lot of modules (in subdirectories) that satisfy the ClientAPIModule interface.
package apimodule

import "github.com/superseriousbusiness/gotosocial/internal/router"

// ClientAPIModule represents a chunk of code (usually contained in a single package) that adds a set
// of functionalities and side effects to a router, by mapping routes and handlers onto it--in other words, a REST API ;)
// A ClientAPIMpdule corresponds roughly to one main path of the gotosocial REST api, for example /api/v1/accounts/ or /oauth/
type ClientAPIModule interface {
	Route(s router.Router) error
}
