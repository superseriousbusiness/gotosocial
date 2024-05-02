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

// FilterV2 represents a user-defined filter for determining which statuses should not be shown to the user.
// v2 filters have names and can include multiple phrases and status IDs to filter.
//
// swagger:model filterV2
//
// ---
// tags:
// - filters
type FilterV2 struct {
	// The ID of the filter in the database.
	ID string `json:"id"`
	// The name of the filter.
	//
	// Example: Linux Words
	Title string `json:"title"`
	// The contexts in which the filter should be applied.
	//
	// Minimum items: 1
	// Unique: true
	// Enum:
	//	- home
	//	- notifications
	//	- public
	//	- thread
	//	- account
	// Example: ["home", "public"]
	Context []FilterContext `json:"context"`
	// When the filter should no longer be applied. Null if the filter does not expire.
	//
	// Example: 2024-02-01T02:57:49Z
	ExpiresAt *string `json:"expires_at"`
	// The action to be taken when a status matches this filter.
	// Enum:
	//	- warn
	//	- hide
	FilterAction FilterAction `json:"filter_action"`
	// The keywords grouped under this filter.
	Keywords []FilterKeyword `json:"keywords"`
	// The statuses grouped under this filter.
	Statuses []FilterStatus `json:"statuses"`
}

// FilterAction is the action to apply to statuses matching a filter.
type FilterAction string

const (
	// FilterActionNone filters should not exist, except internally, for partially constructed or invalid filters.
	FilterActionNone FilterAction = ""
	// FilterActionWarn filters will include this status in API results with a warning.
	FilterActionWarn FilterAction = "warn"
	// FilterActionHide filters will remove this status from API results.
	FilterActionHide FilterAction = "hide"
)

// FilterKeyword represents text to filter within a v2 filter.
//
// swagger:model filterKeyword
//
// ---
// tags:
// - filters
type FilterKeyword struct {
	// The ID of the filter keyword entry in the database.
	ID string `json:"id"`
	// The text to be filtered.
	//
	// Example: fnord
	Keyword string `json:"keyword"`
	// Should the filter consider word boundaries?
	//
	// Example: true
	WholeWord bool `json:"whole_word"`
}

// FilterStatus represents a single status to filter within a v2 filter.
//
// swagger:model filterStatus
//
// ---
// tags:
// - filters
type FilterStatus struct {
	// The ID of the filter status entry in the database.
	ID string `json:"id"`
	// The status ID to be filtered.
	StatusID string `json:"phrase"`
}
