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

package cache

import (
	"sync"
	"time"
)

// Cache defines an in-memory cache that is safe to be wiped when the application is restarted
type Cache interface {
	Store(k string, v interface{}) error
	Fetch(k string) (interface{}, error)
}

type cache struct {
	stored *sync.Map
}

// New returns a new in-memory cache.
func New() Cache {
	cache := &cache{
		stored: &sync.Map{},
	}
	go cache.sweep()
	return cache
}

type cacheEntry struct {
	updated time.Time
	value   interface{}
}
