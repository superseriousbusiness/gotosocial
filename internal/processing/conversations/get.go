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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// GetAll returns conversations owned by the given account.
// The additional parameters can be used for paging.
func (p *Processor) GetAll(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	conversations, err := p.state.DB.GetConversationsByOwnerAccountID(
		ctx,
		requestingAccount.ID,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(
			gtserror.Newf(
				"DB error getting conversations for account %s: %w",
				requestingAccount.ID,
				err,
			),
		)
	}

	// Check for empty response.
	count := len(conversations)
	if len(conversations) == 0 {
		return util.EmptyPageableResponse(), nil
	}

	// Get the lowest and highest last status ID values, used for paging.
	lo := conversations[count-1].LastStatusID
	hi := conversations[0].LastStatusID

	items := make([]interface{}, 0, count)

	filters, mutes, errWithCode := p.getFiltersAndMutes(ctx, requestingAccount)
	if errWithCode != nil {
		return nil, errWithCode
	}

	for _, conversation := range conversations {
		// Convert conversation to frontend API model.
		apiConversation, err := p.converter.ConversationToAPIConversation(
			ctx,
			conversation,
			requestingAccount,
			filters,
			mutes,
		)
		if err != nil {
			log.Errorf(
				ctx,
				"error converting conversation %s to API representation: %v",
				conversation.ID,
				err,
			)
			continue
		}

		// Append conversation to return items.
		items = append(items, apiConversation)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/conversations",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}
