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
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Create a new filter for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) Create(ctx context.Context, account *gtsmodel.Account, form *apimodel.FilterCreateRequestV2) (*apimodel.FilterV2, gtserror.WithCode) {
	filter := &gtsmodel.Filter{
		ID:        id.NewULID(),
		AccountID: account.ID,
		Title:     form.Title,
		Action:    typeutils.APIFilterActionToFilterAction(*form.FilterAction),
	}
	if form.ExpiresIn != nil && *form.ExpiresIn != 0 {
		filter.ExpiresAt = time.Now().Add(time.Second * time.Duration(*form.ExpiresIn))
	}
	for _, context := range form.Context {
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

	for _, formKeyword := range form.Keywords {
		filterKeyword := &gtsmodel.FilterKeyword{
			ID:        id.NewULID(),
			AccountID: account.ID,
			FilterID:  filter.ID,
			Filter:    filter,
			Keyword:   formKeyword.Keyword,
			WholeWord: formKeyword.WholeWord,
		}
		filter.Keywords = append(filter.Keywords, filterKeyword)
	}

	for _, formStatus := range form.Statuses {
		filterStatus := &gtsmodel.FilterStatus{
			ID:        id.NewULID(),
			AccountID: account.ID,
			FilterID:  filter.ID,
			Filter:    filter,
			StatusID:  formStatus.StatusID,
		}
		filter.Statuses = append(filter.Statuses, filterStatus)
	}

	if err := p.state.DB.PutFilter(ctx, filter); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			err = errors.New("duplicate title, keyword, or status")
			return nil, gtserror.NewErrorConflict(err, err.Error())
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiFilter, errWithCode := p.apiFilter(ctx, filter)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Send a filters changed event.
	p.stream.FiltersChanged(ctx, account)

	return apiFilter, nil
}
