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

import (
	"fmt"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/regexes"
)

// ExtractNamestringParts extracts the username test_user and
// the domain example.org from a string like @test_user@example.org.
//
// If nothing is matched, it will return an error.
func ExtractNamestringParts(mention string) (username, host string, err error) {
	matches := regexes.MentionName.FindStringSubmatch(mention)
	switch len(matches) {
	case 2:
		return matches[1], "", nil
	case 3:
		return matches[1], matches[2], nil
	default:
		return "", "", fmt.Errorf("couldn't match mention %s", mention)
	}
}

// ExtractWebfingerParts returns username test_user and
// domain example.org from a string like acct:test_user@example.org,
// or acct:@test_user@example.org.
//
// If nothing is extracted, it will return an error.
func ExtractWebfingerParts(webfinger string) (username, host string, err error) {
	// remove the acct: prefix if it's present
	webfinger = strings.TrimPrefix(webfinger, "acct:")

	// prepend an @ if necessary
	if webfinger[0] != '@' {
		webfinger = "@" + webfinger
	}

	return ExtractNamestringParts(webfinger)
}
