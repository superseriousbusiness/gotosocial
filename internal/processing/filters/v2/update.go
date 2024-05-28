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

package v2

import (
	"context"
	"errors"
	"fmt"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Update an existing filter for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) Update(
	ctx context.Context,
	account *gtsmodel.Account,
	filterID string,
	form *apimodel.FilterUpdateRequestV2,
) (*apimodel.FilterV2, gtserror.WithCode) {
	// Get the filter by ID, with existing keywords and statuses.
	filter, err := p.state.DB.GetFilterByID(ctx, filterID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}
	if filter.AccountID != account.ID {
		return nil, gtserror.NewErrorNotFound(nil)
	}

	// Filter columns that we're going to update.
	filterColumns := []string{}

	// Apply filter changes.
	if form.Title != nil {
		filterColumns = append(filterColumns, "title")
		filter.Title = *form.Title
	}
	if form.FilterAction != nil {
		filterColumns = append(filterColumns, "action")
		filter.Action = typeutils.APIFilterActionToFilterAction(*form.FilterAction)
	}
	// TODO: (Vyr) is it possible to unset a filter expiration with this API?
	if form.ExpiresIn != nil {
		filterColumns = append(filterColumns, "expires_at")
		filter.ExpiresAt = time.Now().Add(time.Second * time.Duration(*form.ExpiresIn))
	}
	if form.Context != nil {
		filterColumns = append(filterColumns,
			"context_home",
			"context_notifications",
			"context_public",
			"context_thread",
			"context_account",
		)
		filter.ContextHome = util.Ptr(false)
		filter.ContextNotifications = util.Ptr(false)
		filter.ContextPublic = util.Ptr(false)
		filter.ContextThread = util.Ptr(false)
		filter.ContextAccount = util.Ptr(false)
		for _, context := range *form.Context {
			switch context {
			case apimodel.FilterContextHome:
				filter.ContextHome = util.Ptr(true)
			case apimodel.FilterContextNotifications:
				filter.ContextNotifications = util.Ptr(true)
			case apimodel.FilterContextPublic:
				filter.ContextPublic = util.Ptr(true)
			case apimodel.FilterContextThread:
				filter.ContextThread = util.Ptr(true)
			case apimodel.FilterContextAccount:
				filter.ContextAccount = util.Ptr(true)
			default:
				return nil, gtserror.NewErrorUnprocessableEntity(
					fmt.Errorf("unsupported filter context '%s'", context),
				)
			}
		}
	}

	// Temporarily detach keywords and statuses from filter, since we're not updating them below.
	filterKeywords := filter.Keywords
	filterStatuses := filter.Statuses
	filter.Keywords = nil
	filter.Statuses = nil

	if err := p.state.DB.UpdateFilter(ctx, filter, filterColumns, nil, nil, nil); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			err = errors.New("you already have a filter with this title")
			return nil, gtserror.NewErrorConflict(err, err.Error())
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Re-attach keywords and statuses before returning.
	filter.Keywords = filterKeywords
	filter.Statuses = filterStatuses

	return p.apiFilter(ctx, filter)
}
