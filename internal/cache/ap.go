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

type APCaches interface {
	// Init will initialize all the ActivityPub caches in this collection.
	// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
	Init()

	// Start will attempt to start all of the ActivityPub caches, or panic.
	Start()

	// Stop will attempt to stop all of the ActivityPub caches, or panic.
	Stop()
}

// NewAP returns a new default implementation of APCaches.
func NewAP() APCaches {
	return &apCaches{}
}

type apCaches struct{}

func (c *apCaches) Init() {}

func (c *apCaches) Start() {}

func (c *apCaches) Stop() {}
