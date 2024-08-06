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
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

func (p *Processor) InteractionRequestsGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	statusID string,
	likes bool,
	replies bool,
	boosts bool,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	reqs, err := p.state.DB.GetPendingInteractionsForAcct(
		ctx,
		requester.ID,
		statusID,
		likes,
		replies,
		boosts,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting pending interactions: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(reqs)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	

	return nil, nil
}
