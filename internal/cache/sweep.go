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

import "time"

// sweep removes all entries more than 5 minutes old, on a loop.
func (c *cache) sweep() {
	t := time.NewTicker(5 * time.Minute)
	for range t.C {
		toRemove := []interface{}{}
		c.stored.Range(func(key interface{}, value interface{}) bool {
			ce, ok := value.(*cacheEntry)
			if !ok {
				panic("cache entry was not a *cacheEntry -- this should never happen")
			}
			if ce.updated.Add(5 * time.Minute).After(time.Now()) {
				toRemove = append(toRemove, key)
			}
			return true
		})
		for _, r := range toRemove {
			c.stored.Delete(r)
		}
	}
}
