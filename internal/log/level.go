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

package log

// LEVEL defines a level of logging.
type LEVEL uint8

// Default levels of logging.
const (
	UNSET LEVEL = 0
	PANIC LEVEL = 1
	ERROR LEVEL = 100
	WARN  LEVEL = 150
	INFO  LEVEL = 200
	DEBUG LEVEL = 250
	TRACE LEVEL = 254
	ALL   LEVEL = ^LEVEL(0)
)

// CanLog returns whether an incoming log of 'lvl' can be logged against receiving level.
func (loglvl LEVEL) CanLog(lvl LEVEL) bool {
	return loglvl > lvl
}
