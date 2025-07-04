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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/util"
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

	// Check whether status if filtered by local participant in context.
	filtered, _, err := p.statusFilter.StatusFilterResultsInContext(ctx,
		requestingAccount,
		conversation.LastStatus,
		gtsmodel.FilterContextNotifications,
	)
	if err != nil {
		log.Errorf(ctx, "error filtering status: %v", err)
	}

	apiConversation, err := p.converter.ConversationToAPIConversation(ctx,
		conversation,
		requestingAccount,
	)
	if err != nil {
		err = gtserror.Newf("error converting conversation %s to API representation: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Set filter results on attached status model.
	apiConversation.LastStatus.Filtered = filtered

	return apiConversation, nil
}
