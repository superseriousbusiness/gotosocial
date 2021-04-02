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

package util

import (
	"fmt"
	"regexp"
)

// To play around with these regexes, see: https://regex101.com/r/2km2EK/1
var (
	hostnameRegexString = `(?:(?:[a-zA-Z]{1})|(?:[a-zA-Z]{1}[a-zA-Z]{1})|(?:[a-zA-Z]{1}[0-9]{1})|(?:[0-9]{1}[a-zA-Z]{1})|(?:[a-zA-Z0-9][a-zA-Z0-9-_]{1,61}[a-zA-Z0-9]))\.(?:[a-zA-Z]{2,6}|[a-zA-Z0-9-]{2,30}\.[a-zA-Z]{2,5}))`
	mentionRegexString  = fmt.Sprintf(`(?: |^|\W)(@[a-zA-Z0-9_]+@%s(?: |\n)`, hostnameRegexString)
	mentionRegex        = regexp.MustCompile(mentionRegexString)
)

// DeriveMentions takes a plaintext (ie., not html-formatted) status,
// and applies a regex to it to return a deduplicated list of accounts
// mentioned in that status.
//
// It will look for fully-qualified account names in the form "@user@example.org".
// Mentions that are just in the form "@username" will not be detected.
func DeriveMentions(status string) []string {
	menchies := []string{}
	for _, match := range mentionRegex.FindAllStringSubmatch(status, -1) {
		menchies = append(menchies, match[1])
	}
	return Unique(menchies)
}

// Unique returns a deduplicated version of a given string slice.
func Unique(s []string) []string {
    keys := make(map[string]bool)
    list := []string{}
    for _, entry := range s {
        if _, value := keys[entry]; !value {
            keys[entry] = true
            list = append(list, entry)
        }
    }
    return list
}

// HTMLFormat takes a plaintext formatted status string, and converts it into
// a nice HTML-formatted string.
//
// This includes:
//
// - Replacing line-breaks with <p>
//
// - Replacing URLs with hrefs.
//
// - Replacing mentions with links to that account's URL as stored in the database.
func HTMLFormat(status string) string {
	// TODO: write proper HTML formatting logic for a status
	return status
}
