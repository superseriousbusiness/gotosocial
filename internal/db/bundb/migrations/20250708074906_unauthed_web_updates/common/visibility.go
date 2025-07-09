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

package common

// Visibility represents the
// visibility granularity of a status.
type Visibility int16

const (
	// VisibilityNone means nobody can see this.
	// It's only used for web status visibility.
	VisibilityNone Visibility = 1

	// VisibilityPublic means this status will
	// be visible to everyone on all timelines.
	VisibilityPublic Visibility = 2

	// VisibilityUnlocked means this status will be visible to everyone,
	// but will only show on home timeline to followers, and in lists.
	VisibilityUnlocked Visibility = 3

	// VisibilityFollowersOnly means this status is viewable to followers only.
	VisibilityFollowersOnly Visibility = 4

	// VisibilityMutualsOnly means this status
	// is visible to mutual followers only.
	VisibilityMutualsOnly Visibility = 5

	// VisibilityDirect means this status is
	// visible only to mentioned recipients.
	VisibilityDirect Visibility = 6

	// VisibilityDefault is used when no other setting can be found.
	VisibilityDefault Visibility = VisibilityUnlocked
)
