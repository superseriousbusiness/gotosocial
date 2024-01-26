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

package v1

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
)

// Create a new filter and filter keyword for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) Create(ctx context.Context, account *gtsmodel.Account, form *apimodel.FilterCreateUpdateRequestV1) (*apimodel.FilterV1, gtserror.WithCode) {
	filter := &gtsmodel.Filter{
		ID:        id.NewULID(),
		AccountID: account.ID,
		Title:     form.Phrase,
		Action:    gtsmodel.FilterActionWarn,
	}
	if *form.Irreversible {
		filter.Action = gtsmodel.FilterActionHide
	}
	if form.ExpiresIn != nil {
		filter.ExpiresAt = time.Now().Add(time.Second * time.Duration(*form.ExpiresIn))
	}
	//goland:noinspection GoImportUsedAsName
	for _, context := range form.Context {
		switch context {
		case apimodel.FilterContextHome:
			filter.ContextHome = true
		case apimodel.FilterContextNotifications:
			filter.ContextNotifications = true
		case apimodel.FilterContextPublic:
			filter.ContextPublic = true
		case apimodel.FilterContextThread:
			filter.ContextThread = true
		case apimodel.FilterContextAccount:
			filter.ContextAccount = true
		default:
			return nil, gtserror.NewErrorUnprocessableEntity(
				fmt.Errorf("unsupported filter context '%s'", context),
			)
		}
	}

	filterKeyword := &gtsmodel.FilterKeyword{
		FilterEntry: gtsmodel.FilterEntry{
			ID:        id.NewULID(),
			AccountID: account.ID,
			FilterID:  filter.ID,
			Filter:    filter,
		},
		Keyword:   form.Phrase,
		WholeWord: *form.WholeWord,
	}
	filter.Keywords = []*gtsmodel.FilterKeyword{filterKeyword}

	if err := p.state.DB.PutFilter(ctx, filter); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			err = errors.New("you already have a filter with this title")
			return nil, gtserror.NewErrorConflict(err, err.Error())
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiFilter(ctx, filterKeyword)
}
