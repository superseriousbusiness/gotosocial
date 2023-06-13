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

package util

import "net/url"

// UniqueStrings returns a deduplicated version of the given
// slice of strings, without changing the order of the entries.
func UniqueStrings(strings []string) []string {
	var (
		l      = len(strings)
		keys   = make(map[string]any, l) // Use map to dedupe items.
		unique = make([]string, 0, l)    // Return slice.
	)

	for _, str := range strings {
		// Check if already set as a key in the map;
		// if not, add to return slice + mark key as set.
		if _, set := keys[str]; !set {
			keys[str] = nil // Value doesn't matter.
			unique = append(unique, str)
		}
	}

	return unique
}

// UniqueURIs returns a deduplicated version of the given
// slice of URIs, without changing the order of the entries.
func UniqueURIs(uris []*url.URL) []*url.URL {
	var (
		l      = len(uris)
		keys   = make(map[string]any, l) // Use map to dedupe items.
		unique = make([]*url.URL, 0, l)  // Return slice.
	)

	for _, uri := range uris {
		uriStr := uri.String()

		// Check if already set as a key in the map;
		// if not, add to return slice + mark key as set.
		if _, set := keys[uriStr]; !set {
			keys[uriStr] = nil // Value doesn't matter.
			unique = append(unique, uri)
		}
	}

	return unique
}
