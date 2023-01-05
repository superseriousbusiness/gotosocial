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

package cache

type Caches struct {
	// GTS provides access to the collection of gtsmodel object caches.
	GTS GTSCaches

	// AP provides access to the collection of ActivityPub object caches.
	AP APCaches

	// prevent pass-by-value.
	_ nocopy
}

// Init will (re)initialize both the GTS and AP cache collections.
// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
func (c *Caches) Init() {
	if c.GTS == nil {
		// use default impl
		c.GTS = NewGTS()
	}

	if c.AP == nil {
		// use default impl
		c.AP = NewAP()
	}

	// initialize caches
	c.GTS.Init()
	c.AP.Init()
}

// Start will start both the GTS and AP cache collections.
func (c *Caches) Start() {
	c.GTS.Start()
	c.AP.Start()
}

// Stop will stop both the GTS and AP cache collections.
func (c *Caches) Stop() {
	c.GTS.Stop()
	c.AP.Stop()
}
