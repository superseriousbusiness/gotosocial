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

package visibility

import (
	"code.superseriousbusiness.org/gotosocial/internal/state"
)

// NoAuth is a placeholder ID used in cache lookups
// when there is no authorized account ID to use.
const NoAuth = "noauth"

// Filter packages up a bunch of logic for checking whether
// given statuses or accounts are visible to a requester.
type Filter struct {
	state *state.State
}

// NewFilter returns a new Filter interface that will use the provided database.
func NewFilter(state *state.State) *Filter {
	return &Filter{state: state}
}
