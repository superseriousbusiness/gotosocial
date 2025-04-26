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
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/processing/account"
	"code.superseriousbusiness.org/gotosocial/internal/processing/common"
	"code.superseriousbusiness.org/gotosocial/internal/processing/conversations"
	"code.superseriousbusiness.org/gotosocial/internal/processing/media"
	"code.superseriousbusiness.org/gotosocial/internal/processing/stream"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/webpush"
	"code.superseriousbusiness.org/gotosocial/internal/workers"
)

type Processor struct {
	clientAPI clientAPI
	fediAPI   fediAPI
	workers   *workers.Workers
}

func New(
	state *state.State,
	common *common.Processor,
	federator *federation.Federator,
	converter *typeutils.Converter,
	visFilter *visibility.Filter,
	emailSender email.Sender,
	webPushSender webpush.Sender,
	account *account.Processor,
	media *media.Processor,
	stream *stream.Processor,
	conversations *conversations.Processor,
) Processor {
	// Init federate logic
	// wrapper struct.
	federate := &federate{
		Federator: federator,
		state:     state,
		converter: converter,
	}

	// Init surface logic
	// wrapper struct.
	surface := &Surface{
		State:         state,
		Converter:     converter,
		Stream:        stream,
		VisFilter:     visFilter,
		EmailSender:   emailSender,
		WebPushSender: webPushSender,
		Conversations: conversations,
	}

	// Init shared util funcs.
	utils := &utils{
		state:     state,
		media:     media,
		account:   account,
		surface:   surface,
		converter: converter,
	}

	return Processor{
		workers: &state.Workers,
		clientAPI: clientAPI{
			state:     state,
			converter: converter,
			surface:   surface,
			federate:  federate,
			account:   account,
			common:    common,
			utils:     utils,
		},
		fediAPI: fediAPI{
			state:    state,
			surface:  surface,
			federate: federate,
			account:  account,
			common:   common,
			utils:    utils,
		},
	}
}
