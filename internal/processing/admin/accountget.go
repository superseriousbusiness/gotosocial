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

package admin

import (
	"context"
	"errors"
	"fmt"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

func (p *Processor) AccountGet(ctx context.Context, accountID string) (*apimodel.AdminAccountInfo, gtserror.WithCode) {
	account, err := p.state.DB.GetAccountByID(ctx, accountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting account %s: %w", accountID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if account == nil {
		err := fmt.Errorf("account %s not found", accountID)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	apiAccount, err := p.converter.AccountToAdminAPIAccount(ctx, account)
	if err != nil {
		err := gtserror.Newf("error converting account %s to admin api model: %w", accountID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiAccount, nil
}
