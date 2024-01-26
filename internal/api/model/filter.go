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

// FilterContext represents the context in which to apply a filter.
// v1 and v2 filter APIs use the same set of contexts.
//
// swagger:model filterContext
type FilterContext string

const (
	// FilterContextHome means this filter should be applied to the home timeline and lists.
	FilterContextHome FilterContext = "home"
	// FilterContextNotifications means this filter should be applied to the notifications timeline.
	FilterContextNotifications FilterContext = "notifications"
	// FilterContextPublic means this filter should be applied to public timelines.
	FilterContextPublic FilterContext = "public"
	// FilterContextThread means this filter should be applied to the expanded thread of a detailed status.
	FilterContextThread FilterContext = "thread"
	// FilterContextAccount means this filter should be applied when viewing a profile.
	FilterContextAccount FilterContext = "account"

	FilterContextNumValues = 5
)
