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

package account

import (
	"context"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
)

// PutNote updates the requesting account's private note on the target account.
func (p *Processor) PutNote(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string, comment string) (*apimodel.Relationship, gtserror.WithCode) {
	targetAccount, errWithCode := p.Get(ctx, requestingAccount, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	note := &gtsmodel.AccountNote{
		ID:              id.NewULID(),
		AccountID:       requestingAccount.ID,
		TargetAccountID: targetAccount.ID,
		Comment:         comment,
	}
	err := p.state.DB.PutNote(ctx, note)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.RelationshipGet(ctx, requestingAccount, targetAccount.ID)
}
