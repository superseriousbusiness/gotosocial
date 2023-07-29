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

package gtsmodel

import "time"

// Marker stores a local account's read position on a given timeline.
type Marker struct {
	AccountID  string     `validate:"required,ulid" bun:"type:CHAR(26),pk,unique:markers_account_id_timeline_uniq,notnull,nullzero"` // ID of the local account that owns the marker
	Name       MarkerName `validate:"oneof=home notifications" bun:",nullzero,notnull,pk,unique:markers_account_id_timeline_uniq"`   // Name of the marked timeline
	UpdatedAt  time.Time  `validate:"required" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                    // When marker was last updated
	Version    int        `validate:"required,min=0" bun:",nullzero,notnull,default:0"`                                              // For optimistic concurrency control
	LastReadID string     `validate:"required,ulid" bun:"type:CHAR(26),notnull,nullzero"`                                            // Last ID read on this timeline (status ID for home, notification ID for notifications)
}

// MarkerName is the name of one of the timelines we can store markers for.
type MarkerName string

const (
	MarkerNameHome          MarkerName = "home"
	MarkerNameNotifications MarkerName = "notifications"
)
