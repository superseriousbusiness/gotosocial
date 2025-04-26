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
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Update updates one list for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) Update(
	ctx context.Context,
	account *gtsmodel.Account,
	id string,
	title *string,
	repliesPolicy *gtsmodel.RepliesPolicy,
	exclusive *bool,
) (*apimodel.List, gtserror.WithCode) {
	list, errWithCode := p.getList(
		// Use barebones ctx; no embedded
		// structs necessary for this call.
		gtscontext.SetBarebones(ctx),
		account.ID,
		id,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Only update columns we're told to update.
	columns := make([]string, 0, 3)

	if title != nil {
		list.Title = *title
		columns = append(columns, "title")
	}

	if repliesPolicy != nil {
		list.RepliesPolicy = *repliesPolicy
		columns = append(columns, "replies_policy")
	}

	if exclusive != nil {
		list.Exclusive = exclusive
		columns = append(columns, "exclusive")
	}

	if err := p.state.DB.UpdateList(ctx, list, columns...); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			err = errors.New("you already have a list with this title")
			return nil, gtserror.NewErrorConflict(err, err.Error())
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiList(ctx, list)
}
