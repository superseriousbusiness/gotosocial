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

package ap

import (
	"net/url"
	"sync"

	"codeberg.org/superseriousbusiness/activity/streams/vocab"
)

const mapmax = 256

// mapPool is a memory pool
// of maps for JSON decoding.
var mapPool sync.Pool

// getMap acquires a map from memory pool.
func getMap() map[string]any {
	v := mapPool.Get()
	if v == nil {
		// preallocate map of max-size.
		m := make(map[string]any, mapmax)
		v = m
	}
	return v.(map[string]any) //nolint
}

// putMap clears and places map back in pool.
func putMap(m map[string]any) {
	if len(m) > mapmax {
		// drop maps beyond
		// our maximum size.
		return
	}
	clear(m)
	mapPool.Put(m)
}

// _TypeOrIRI wraps a vocab.Type to implement TypeOrIRI.
type _TypeOrIRI struct {
	vocab.Type
}

func (t *_TypeOrIRI) GetType() vocab.Type {
	return t.Type
}

func (t *_TypeOrIRI) GetIRI() *url.URL {
	return nil
}

func (t *_TypeOrIRI) IsIRI() bool {
	return false
}

func (t *_TypeOrIRI) SetIRI(*url.URL) {}
