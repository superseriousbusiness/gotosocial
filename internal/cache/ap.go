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

type APCaches struct{}

// Init will initialize all the ActivityPub caches in this collection.
// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
func (c *APCaches) Init() {}

// Start will attempt to start all of the ActivityPub caches, or panic.
func (c *APCaches) Start() {}

// Stop will attempt to stop all of the ActivityPub caches, or panic.
func (c *APCaches) Stop() {}
