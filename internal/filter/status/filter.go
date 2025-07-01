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

package status

import (
	"code.superseriousbusiness.org/gotosocial/internal/state"
)

// noauth is a placeholder ID used in cache lookups
// when there is no authorized account ID to use.
const noauth = "noauth"

// Filter packages up logic for checking whether
// given status is muted by a given requester (user).
type Filter struct{ state *state.State }

// New returns a new Filter interface that will use the provided state.
func NewFilter(state *state.State) *Filter { return &Filter{state} }
