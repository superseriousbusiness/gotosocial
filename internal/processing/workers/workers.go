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

package workers

import (
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/processing/account"
	"github.com/superseriousbusiness/gotosocial/internal/processing/media"
	"github.com/superseriousbusiness/gotosocial/internal/processing/stream"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/workers"
)

type Processor struct {
	workers   *workers.Workers
	clientAPI *clientAPI
	fediAPI   *fediAPI
}

func New(
	state *state.State,
	federator federation.Federator,
	converter *typeutils.Converter,
	filter *visibility.Filter,
	emailSender email.Sender,
	account *account.Processor,
	media *media.Processor,
	stream *stream.Processor,
) Processor {
	// Init surface logic
	// wrapper struct.
	surface := &surface{
		state:       state,
		converter:   converter,
		stream:      stream,
		filter:      filter,
		emailSender: emailSender,
	}

	// Init federate logic
	// wrapper struct.
	federate := &federate{
		Federator: federator,
		state:     state,
		converter: converter,
	}

	// Init shared logic wipe
	// status util func.
	wipeStatus := wipeStatusF(
		state,
		media,
		surface,
	)

	return Processor{
		workers: &state.Workers,
		clientAPI: &clientAPI{
			state:      state,
			converter:  converter,
			surface:    surface,
			federate:   federate,
			wipeStatus: wipeStatus,
			account:    account,
		},
		fediAPI: &fediAPI{
			state:      state,
			surface:    surface,
			federate:   federate,
			wipeStatus: wipeStatus,
			account:    account,
		},
	}
}
