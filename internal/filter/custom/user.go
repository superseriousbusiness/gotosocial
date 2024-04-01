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

// Package custom represents custom filters managed by the user through the API.
package custom

import (
	"errors"
)

// ErrHideStatus indicates that a status has been filtered and should not be returned at all.
var ErrHideStatus = errors.New("hide status")

// FilterContext determines the filters that apply to a given status or list of statuses.
type FilterContext string

const (
	// FilterContextNone means no filters should be applied.
	// There are no filters with this context; it's for internal use only.
	FilterContextNone FilterContext = ""
	// FilterContextHome means this status is being filtered as part of a home or list timeline.
	FilterContextHome FilterContext = "home"
	// FilterContextNotifications means this status is being filtered as part of the notifications timeline.
	FilterContextNotifications FilterContext = "notifications"
	// FilterContextPublic means this status is being filtered as part of a public or tag timeline.
	FilterContextPublic FilterContext = "public"
	// FilterContextThread means this status is being filtered as part of a thread's context.
	FilterContextThread FilterContext = "thread"
	// FilterContextAccount means this status is being filtered as part of an account's statuses.
	FilterContextAccount FilterContext = "account"
)
