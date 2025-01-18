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

import "time"

// ISO8601 is a formatter for serializing times that forces ISO8601 behavior.
const ISO8601 = "2006-01-02T15:04:05.000Z"
const ISO8601Date = "2006-01-02"

// FormatISO8601 converts the given time to UTC and then formats it
// using the ISO8601 const, which the Mastodon API is able to understand.
func FormatISO8601(t time.Time) string {
	return t.UTC().Format(ISO8601)
}

// Mastodon returns UTC dates (without time) for last_status_at/LastStatusAt as
// a special case, but most of the time you want to use FormatISO8601 instead.
func FormatISO8601Date(t time.Time) string {
	return t.UTC().Format(ISO8601Date)
}

// ParseISO8601 parses the given time string according to the ISO8601 const.
func ParseISO8601(in string) (time.Time, error) {
	return time.Parse(ISO8601, in)
}
