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

import (
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// Processor provides a processor with logic
// common to multiple logical domains of the
// processing subsection of the codebase.
type Processor struct {
	state     *state.State
	media     *media.Manager
	converter *typeutils.Converter
	federator *federation.Federator
	visFilter *visibility.Filter
}

// New returns a new Processor instance.
func New(
	state *state.State,
	media *media.Manager,
	converter *typeutils.Converter,
	federator *federation.Federator,
	visFilter *visibility.Filter,
) Processor {
	return Processor{
		state:     state,
		media:     media,
		converter: converter,
		federator: federator,
		visFilter: visFilter,
	}
}
