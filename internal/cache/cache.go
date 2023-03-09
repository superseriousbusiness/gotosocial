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

package cache

type Caches struct {
	// GTS provides access to the collection of gtsmodel object caches.
	// (used by the database).
	GTS GTSCaches

	// AP provides access to the collection of ActivityPub object caches.
	// (planned to be used by the typeconverter).
	AP APCaches

	// Visibility provides access to the item visibility cache.
	// (used by the visibility filter).
	Visibility VisibilityCache

	// prevent pass-by-value.
	_ nocopy
}

// Init will (re)initialize both the GTS and AP cache collections.
// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
func (c *Caches) Init() {
	c.GTS.Init()
	c.AP.Init()
	c.Visibility.Init()
}

// Start will start both the GTS and AP cache collections.
func (c *Caches) Start() {
	c.GTS.Start()
	c.AP.Start()
	c.Visibility.Start()
}

// Stop will stop both the GTS and AP cache collections.
func (c *Caches) Stop() {
	c.GTS.Stop()
	c.AP.Stop()
	c.Visibility.Stop()
}
