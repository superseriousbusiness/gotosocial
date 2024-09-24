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
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/filter/interaction"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/processing/common"
	"github.com/superseriousbusiness/gotosocial/internal/processing/interactionrequests"
	"github.com/superseriousbusiness/gotosocial/internal/processing/polls"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

type Processor struct {
	// embedded common logic
	c *common.Processor

	state        *state.State
	federator    *federation.Federator
	converter    *typeutils.Converter
	visFilter    *visibility.Filter
	intFilter    *interaction.Filter
	formatter    *text.Formatter
	parseMention gtsmodel.ParseMentionFunc

	// other processors
	polls   *polls.Processor
	intReqs *interactionrequests.Processor
}

// New returns a new status processor.
func New(
	state *state.State,
	common *common.Processor,
	polls *polls.Processor,
	intReqs *interactionrequests.Processor,
	federator *federation.Federator,
	converter *typeutils.Converter,
	visFilter *visibility.Filter,
	intFilter *interaction.Filter,
	parseMention gtsmodel.ParseMentionFunc,
) Processor {
	return Processor{
		c:            common,
		state:        state,
		federator:    federator,
		converter:    converter,
		visFilter:    visFilter,
		intFilter:    intFilter,
		formatter:    text.NewFormatter(state.DB),
		parseMention: parseMention,
		polls:        polls,
		intReqs:      intReqs,
	}
}
