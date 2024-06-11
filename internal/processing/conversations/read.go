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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *Processor) Read(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	id string,
) (*apimodel.Conversation, gtserror.WithCode) {
	// Get the conversation, including participating accounts and last status.
	conversation, err := p.state.DB.GetConversationByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		err = gtserror.Newf("db error getting conversation %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	if conversation.AccountID != requestingAccount.ID {
		return nil, gtserror.NewErrorNotFound(nil)
	}

	// Mark the conversation as read.
	conversation.Read = util.Ptr(true)
	if err := p.state.DB.PutConversation(ctx, conversation, "read"); err != nil {
		err = gtserror.Newf("db error updating conversation %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	filters, err := p.state.DB.GetFiltersForAccountID(ctx, requestingAccount.ID)
	if err != nil {
		err = gtserror.Newf("couldn't retrieve filters for account %s: %w", requestingAccount.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	mutes, err := p.state.DB.GetAccountMutes(gtscontext.SetBarebones(ctx), requestingAccount.ID, nil)
	if err != nil {
		err = gtserror.Newf("couldn't retrieve mutes for account %s: %w", requestingAccount.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	compiledMutes := usermute.NewCompiledUserMuteList(mutes)

	apiConversation, err := p.converter.ConversationToAPIConversation(
		ctx,
		conversation,
		requestingAccount,
		filters,
		compiledMutes,
	)
	if err != nil {
		err = gtserror.Newf("db error converting conversation %s to API representation: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiConversation, nil
}
