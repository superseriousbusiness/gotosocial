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

package federation

import (
	"time"

	"codeberg.org/superseriousbusiness/activity/pub"
)

/*
	GOFED CLOCK INTERFACE
	Determines the time.
*/

// Clock implements the Clock interface of go-fed
type Clock struct{}

// Now just returns the time now
func (c *Clock) Now() time.Time {
	return time.Now()
}

// NewClock returns a simple pub.Clock for use in federation interfaces.
func NewClock() pub.Clock {
	return &Clock{}
}
