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

package conversations

import (
	"context"
	"errors"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/filter/mutes"
	"code.superseriousbusiness.org/gotosocial/internal/filter/status"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

type Processor struct {
	state        *state.State
	converter    *typeutils.Converter
	visFilter    *visibility.Filter
	muteFilter   *mutes.Filter
	statusFilter *status.Filter
}

func New(
	state *state.State,
	converter *typeutils.Converter,
	visFilter *visibility.Filter,
	muteFilter *mutes.Filter,
	statusFilter *status.Filter,
) Processor {
	return Processor{
		state:        state,
		converter:    converter,
		visFilter:    visFilter,
		muteFilter:   muteFilter,
		statusFilter: statusFilter,
	}
}

const conversationNotFoundHelpText = "conversation not found"

// getConversationOwnedBy gets a conversation by ID and checks that it is owned by the given account.
func (p *Processor) getConversationOwnedBy(
	ctx context.Context,
	id string,
	requestingAccount *gtsmodel.Account,
) (*gtsmodel.Conversation, gtserror.WithCode) {
	// Get the conversation so that we can check its owning account ID.
	conversation, err := p.state.DB.GetConversationByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(
			gtserror.Newf(
				"DB error getting conversation %s for account %s: %w",
				id,
				requestingAccount.ID,
				err,
			),
		)
	}
	if conversation == nil {
		return nil, gtserror.NewErrorNotFound(
			gtserror.Newf(
				"conversation %s not found: %w",
				id,
				err,
			),
			conversationNotFoundHelpText,
		)
	}
	if conversation.AccountID != requestingAccount.ID {
		return nil, gtserror.NewErrorNotFound(
			gtserror.Newf(
				"conversation %s not owned by account %s: %w",
				id,
				requestingAccount.ID,
				err,
			),
			conversationNotFoundHelpText,
		)
	}

	return conversation, nil
}
