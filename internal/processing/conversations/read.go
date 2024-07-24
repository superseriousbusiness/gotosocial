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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
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
	conversation, errWithCode := p.getConversationOwnedBy(ctx, id, requestingAccount)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Mark the conversation as read.
	conversation.Read = util.Ptr(true)
	if err := p.state.DB.UpsertConversation(ctx, conversation, "read"); err != nil {
		err = gtserror.Newf("DB error updating conversation %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	filters, mutes, errWithCode := p.getFiltersAndMutes(ctx, requestingAccount)
	if errWithCode != nil {
		return nil, errWithCode
	}

	apiConversation, err := p.converter.ConversationToAPIConversation(
		ctx,
		conversation,
		requestingAccount,
		filters,
		mutes,
	)
	if err != nil {
		err = gtserror.Newf("error converting conversation %s to API representation: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiConversation, nil
}
