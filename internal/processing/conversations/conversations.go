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

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

type Processor struct {
	state     *state.State
	converter *typeutils.Converter
	filter    *visibility.Filter
}

func New(
	state *state.State,
	converter *typeutils.Converter,
	filter *visibility.Filter,
) Processor {
	return Processor{
		state:     state,
		converter: converter,
		filter:    filter,
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

// getFiltersAndMutes gets the given account's filters and compiled mute list.
func (p *Processor) getFiltersAndMutes(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
) ([]*gtsmodel.Filter, *usermute.CompiledUserMuteList, gtserror.WithCode) {
	filters, err := p.state.DB.GetFiltersForAccountID(ctx, requestingAccount.ID)
	if err != nil {
		return nil, nil, gtserror.NewErrorInternalError(
			gtserror.Newf(
				"DB error getting filters for account %s: %w",
				requestingAccount.ID,
				err,
			),
		)
	}

	mutes, err := p.state.DB.GetAccountMutes(gtscontext.SetBarebones(ctx), requestingAccount.ID, nil)
	if err != nil {
		return nil, nil, gtserror.NewErrorInternalError(
			gtserror.Newf(
				"DB error getting mutes for account %s: %w",
				requestingAccount.ID,
				err,
			),
		)
	}
	compiledMutes := usermute.NewCompiledUserMuteList(mutes)

	return filters, compiledMutes, nil
}
