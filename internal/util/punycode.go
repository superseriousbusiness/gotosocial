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

import (
	"strings"

	"golang.org/x/net/idna"
)

// Punify converts the given domain to lowercase
// then to punycode (for international domain names).
//
// Returns the resulting domain or an error if the
// punycode conversion fails.
func Punify(domain string) (string, error) {
	domain = strings.ToLower(domain)
	return idna.ToASCII(domain)
}

// DePunify converts the given punycode string
// to its original unicode representation (lowercased).
// Noop if the domain is (already) not puny.
//
// Returns an error if conversion fails.
func DePunify(domain string) (string, error) {
	out, err := idna.ToUnicode(domain)
	return strings.ToLower(out), err
}
