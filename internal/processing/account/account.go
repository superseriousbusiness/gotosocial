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
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

// Processor wraps functionality for updating, creating, and deleting accounts in response to API requests.
//
// It also contains logic for actions towards accounts such as following, blocking, seeing follows, etc.
type Processor struct {
	state        *state.State
	tc           typeutils.TypeConverter
	mediaManager media.Manager
	oauthServer  oauth.Server
	filter       *visibility.Filter
	formatter    text.Formatter
	federator    federation.Federator
	parseMention gtsmodel.ParseMentionFunc
}

// New returns a new account processor.
func New(
	state *state.State,
	tc typeutils.TypeConverter,
	mediaManager media.Manager,
	oauthServer oauth.Server,
	federator federation.Federator,
	filter *visibility.Filter,
	parseMention gtsmodel.ParseMentionFunc,
) Processor {
	return Processor{
		state:        state,
		tc:           tc,
		mediaManager: mediaManager,
		oauthServer:  oauthServer,
		filter:       filter,
		formatter:    text.NewFormatter(state.DB),
		federator:    federator,
		parseMention: parseMention,
	}
}
