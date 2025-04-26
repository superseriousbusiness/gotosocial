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

package list

import (
	"context"
	"errors"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
)

// Create creates one a new list for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) Create(
	ctx context.Context,
	account *gtsmodel.Account,
	title string,
	repliesPolicy gtsmodel.RepliesPolicy,
	exclusive bool,
) (*apimodel.List, gtserror.WithCode) {
	list := &gtsmodel.List{
		ID:            id.NewULID(),
		Title:         title,
		AccountID:     account.ID,
		RepliesPolicy: repliesPolicy,
		Exclusive:     &exclusive,
	}

	if err := p.state.DB.PutList(ctx, list); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			err = errors.New("you already have a list with this title")
			return nil, gtserror.NewErrorConflict(err, err.Error())
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiList(ctx, list)
}
