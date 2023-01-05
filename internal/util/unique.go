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

package util

import "net/url"

// UniqueStrings returns a deduplicated version of a given string slice.
func UniqueStrings(s []string) []string {
	keys := make(map[string]bool, len(s))
	list := []string{}
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// UniqueURIs returns a deduplicated version of a given *url.URL slice.
func UniqueURIs(s []*url.URL) []*url.URL {
	keys := make(map[string]bool, len(s))
	list := []*url.URL{}
	for _, entry := range s {
		if _, value := keys[entry.String()]; !value {
			keys[entry.String()] = true
			list = append(list, entry)
		}
	}
	return list
}
