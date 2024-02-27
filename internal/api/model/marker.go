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

package model

// Marker represents the last read position within a user's timelines.
//
// swagger:model markers
type Marker struct {
	// Information about the user's position in the home timeline.
	Home *TimelineMarker `json:"home,omitempty"`
	// Information about the user's position in their notifications.
	Notifications *TimelineMarker `json:"notifications,omitempty"`
}

// TimelineMarker contains information about a user's progress through a specific timeline.
type TimelineMarker struct {
	// The ID of the most recently viewed entity.
	LastReadID string `json:"last_read_id"`
	// The timestamp of when the marker was set (ISO 8601 Datetime)
	UpdatedAt string `json:"updated_at"`
	// Used for locking to prevent write conflicts.
	Version int `json:"version"`
}

// MarkerName is the name of one of the timelines we can store markers for.
type MarkerName string

const (
	MarkerNameHome          MarkerName = "home"
	MarkerNameNotifications MarkerName = "notifications"
	MarkerNameNumValues                = 2
)

// MarkerPostRequest models a request to update one or more markers.
// This has two sets of fields to support a goofy nested map structure in both form data and JSON bodies.
//
// swagger:ignore
type MarkerPostRequest struct {
	Home                        *MarkerPostRequestMarker `json:"home"`
	FormHomeLastReadID          string                   `form:"home[last_read_id]"`
	Notifications               *MarkerPostRequestMarker `json:"notifications"`
	FormNotificationsLastReadID string                   `form:"notifications[last_read_id]"`
}

type MarkerPostRequestMarker struct {
	// The ID of the most recently viewed entity.
	LastReadID string `json:"last_read_id"`
}

// HomeLastReadID should be used instead of Home or FormHomeLastReadID.
func (r *MarkerPostRequest) HomeLastReadID() string {
	if r.Home != nil {
		return r.Home.LastReadID
	}
	return r.FormHomeLastReadID
}

// NotificationsLastReadID should be used instead of Notifications or FormNotificationsLastReadID.
func (r *MarkerPostRequest) NotificationsLastReadID() string {
	if r.Notifications != nil {
		return r.Notifications.LastReadID
	}
	return r.FormNotificationsLastReadID
}
