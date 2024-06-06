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

// UserMuteCreateUpdateRequest captures params for creating or updating a user mute.
//
// swagger:ignore
type UserMuteCreateUpdateRequest struct {
	// Should the mute apply to notifications from that user?
	//
	// Example: true
	Notifications *bool `form:"notifications" json:"notifications" xml:"notifications"`
	// Number of seconds from now that the mute should expire. If omitted or 0, mute never expires.
	Duration *int `json:"-" form:"duration" xml:"duration"`
	// Number of seconds from now that the mute should expire. If omitted or 0, mute never expires.
	//
	// Example: 86400
	DurationI interface{} `json:"duration"`
}
