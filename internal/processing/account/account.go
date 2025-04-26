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

package account

import (
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/processing/common"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/text"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// Processor wraps functionality for updating, creating, and deleting accounts in response to API requests.
//
// It also contains logic for actions towards accounts such as following, blocking, seeing follows, etc.
type Processor struct {
	// common processor logic
	c *common.Processor

	state        *state.State
	converter    *typeutils.Converter
	mediaManager *media.Manager
	visFilter    *visibility.Filter
	formatter    *text.Formatter
	federator    *federation.Federator
	parseMention gtsmodel.ParseMentionFunc
	themes       *Themes
}

// New returns a new account processor.
func New(
	common *common.Processor,
	state *state.State,
	converter *typeutils.Converter,
	mediaManager *media.Manager,
	federator *federation.Federator,
	visFilter *visibility.Filter,
	parseMention gtsmodel.ParseMentionFunc,
) Processor {
	return Processor{
		c:            common,
		state:        state,
		converter:    converter,
		mediaManager: mediaManager,
		visFilter:    visFilter,
		formatter:    text.NewFormatter(state.DB),
		federator:    federator,
		parseMention: parseMention,
		themes:       PopulateThemes(),
	}
}
