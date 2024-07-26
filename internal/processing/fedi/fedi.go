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

package fedi

import (
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/processing/common"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

type Processor struct {
	// embed common logic
	c *common.Processor

	state     *state.State
	federator *federation.Federator
	converter *typeutils.Converter
	visFilter *visibility.Filter
}

// New returns a
// new fedi processor.
func New(
	state *state.State,
	common *common.Processor,
	converter *typeutils.Converter,
	federator *federation.Federator,
	visFilter *visibility.Filter,
) Processor {
	return Processor{
		c:         common,
		state:     state,
		federator: federator,
		converter: converter,
		visFilter: visFilter,
	}
}
