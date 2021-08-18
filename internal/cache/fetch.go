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

func (c *cache) Fetch(k string) (interface{}, error) {
	ceI, stored := c.stored.Load(k)
	if !stored {
		return nil, ErrNotFound
	}

	ce, ok := ceI.(*cacheEntry)
	if !ok {
		panic("cache entry was not a *cacheEntry -- this should never happen")
	}

	return ce.value, nil
}
