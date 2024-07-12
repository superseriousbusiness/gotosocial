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

package status

import (
	"context"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// HistoryGet gets edit history for the target status, taking account of privacy settings and blocks etc.
// TODO: currently this just returns the latest version of the status.
func (p *Processor) HistoryGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) ([]*apimodel.StatusEdit, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requestingAccount,
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	apiStatus, errWithCode := p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return []*apimodel.StatusEdit{
		{
			Content:          apiStatus.Content,
			SpoilerText:      apiStatus.SpoilerText,
			Sensitive:        apiStatus.Sensitive,
			CreatedAt:        util.FormatISO8601(targetStatus.UpdatedAt),
			Account:          apiStatus.Account,
			Poll:             apiStatus.Poll,
			MediaAttachments: apiStatus.MediaAttachments,
			Emojis:           apiStatus.Emojis,
		},
	}, nil
}

// Get gets the given status, taking account of privacy settings and blocks etc.
func (p *Processor) Get(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requestingAccount,
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
}

// SourceGet returns the *apimodel.StatusSource version of the targetStatusID.
// Status must belong to the requester, and must not be a boost.
func (p *Processor) SourceGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.StatusSource, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requestingAccount,
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Redirect to wrapped status if boost.
	targetStatus, errWithCode = p.c.UnwrapIfBoost(
		ctx,
		requestingAccount,
		targetStatus,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if targetStatus.AccountID != requestingAccount.ID {
		err := gtserror.Newf(
			"status %s does not belong to account %s",
			targetStatusID, requestingAccount.ID,
		)
		return nil, gtserror.NewErrorNotFound(err)
	}

	statusSource, err := p.converter.StatusToAPIStatusSource(ctx, targetStatus)
	if err != nil {
		err = gtserror.Newf("error converting status: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return statusSource, nil
}
