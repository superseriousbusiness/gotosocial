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

package model

// Marker represents the last read position within a user's timelines.
type Marker struct {
	// Information about the user's position in the home timeline.
	Home *TimelineMarker `json:"home"`
	// Information about the user's position in their notifications.
	Notifications *TimelineMarker `json:"notifications"`
}

// TimelineMarker contains information about a user's progress through a specific timeline.
type TimelineMarker struct {
	// The ID of the most recently viewed entity.
	LastReadID string `json:"last_read_id"`
	// The timestamp of when the marker was set (ISO 8601 Datetime)
	UpdatedAt string `json:"updated_at"`
	// Used for locking to prevent write conflicts.
	Version string `json:"version"`
}
