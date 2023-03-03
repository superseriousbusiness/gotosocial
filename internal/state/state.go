/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package state

import (
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/workers"
)

// State provides a means of dependency injection and sharing of resources
// across different subpackages of the GoToSocial codebase. DO NOT assume
// that any particular field will be initialized if you are accessing this
// during initialization. A pointer to a State{} is often passed during
// subpackage initialization, while the returned subpackage type will later
// then be set and stored within the State{} itself.
type State struct {
	// Caches provides access to this state's collection of caches.
	Caches cache.Caches

	// DB provides access to the database.
	DB db.DB

	// Storage provides access to the storage driver.
	Storage *storage.Driver

	// Workers provides access to this state's collection of worker pools.
	Workers workers.Workers

	// prevent pass-by-value.
	_ nocopy
}

// nocopy when embedded will signal linter to
// error on pass-by-value of parent struct.
type nocopy struct{}

func (*nocopy) Lock() {}

func (*nocopy) Unlock() {}
