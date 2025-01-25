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

package admin

import (
	"github.com/superseriousbusiness/gotosocial/internal/cleaner"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/processing/common"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/subscriptions"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

type Processor struct {
	// common processor logic
	c *common.Processor

	state         *state.State
	cleaner       *cleaner.Cleaner
	subscriptions *subscriptions.Subscriptions
	converter     *typeutils.Converter
	federator     *federation.Federator
	media         *media.Manager
	transport     transport.Controller
	email         email.Sender
}

// New returns a new admin processor.
func New(
	common *common.Processor,
	state *state.State,
	cleaner *cleaner.Cleaner,
	subscriptions *subscriptions.Subscriptions,
	federator *federation.Federator,
	converter *typeutils.Converter,
	mediaManager *media.Manager,
	transportController transport.Controller,
	emailSender email.Sender,
) Processor {
	return Processor{
		c:             common,
		state:         state,
		cleaner:       cleaner,
		subscriptions: subscriptions,
		converter:     converter,
		federator:     federator,
		media:         mediaManager,
		transport:     transportController,
		email:         emailSender,
	}
}
