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
	"strings"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Update an existing filter and filter keyword for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) Update(
	ctx context.Context,
	account *gtsmodel.Account,
	filterKeywordID string,
	form *apimodel.FilterCreateUpdateRequestV1,
) (*apimodel.FilterV1, gtserror.WithCode) {
	// Get enough of the filter keyword that we can look up its filter ID.
	filterKeyword, err := p.state.DB.GetFilterKeywordByID(gtscontext.SetBarebones(ctx), filterKeywordID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}
	if filterKeyword.AccountID != account.ID {
		return nil, gtserror.NewErrorNotFound(nil)
	}

	// Get the filter for this keyword.
	filter, err := p.state.DB.GetFilterByID(ctx, filterKeyword.FilterID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	title := form.Phrase
	action := gtsmodel.FilterActionWarn
	if *form.Irreversible {
		action = gtsmodel.FilterActionHide
	}
	expiresAt := time.Time{}
	if form.ExpiresIn != nil && *form.ExpiresIn != 0 {
		expiresAt = time.Now().Add(time.Second * time.Duration(*form.ExpiresIn))
	}
	contextHome := false
	contextNotifications := false
	contextPublic := false
	contextThread := false
	contextAccount := false
	for _, context := range form.Context {
		switch context {
		case apimodel.FilterContextHome:
			contextHome = true
		case apimodel.FilterContextNotifications:
			contextNotifications = true
		case apimodel.FilterContextPublic:
			contextPublic = true
		case apimodel.FilterContextThread:
			contextThread = true
		case apimodel.FilterContextAccount:
			contextAccount = true
		default:
			return nil, gtserror.NewErrorUnprocessableEntity(
				fmt.Errorf("unsupported filter context '%s'", context),
			)
		}
	}

	// v1 filter APIs can't change certain fields for a filter with multiple keywords or any statuses,
	// since it would be an unexpected side effect on filters that, to the v1 API, appear separate.
	// See https://docs.joinmastodon.org/methods/filters/#update-v1
	if len(filter.Keywords) > 1 || len(filter.Statuses) > 0 {
		forbiddenFields := make([]string, 0, 4)
		if title != filter.Title {
			forbiddenFields = append(forbiddenFields, "phrase")
		}
		if action != filter.Action {
			forbiddenFields = append(forbiddenFields, "irreversible")
		}
		if expiresAt != filter.ExpiresAt {
			forbiddenFields = append(forbiddenFields, "expires_in")
		}
		if contextHome != util.PtrOrValue(filter.ContextHome, false) ||
			contextNotifications != util.PtrOrValue(filter.ContextNotifications, false) ||
			contextPublic != util.PtrOrValue(filter.ContextPublic, false) ||
			contextThread != util.PtrOrValue(filter.ContextThread, false) ||
			contextAccount != util.PtrOrValue(filter.ContextAccount, false) {
			forbiddenFields = append(forbiddenFields, "context")
		}
		if len(forbiddenFields) > 0 {
			return nil, gtserror.NewErrorUnprocessableEntity(
				fmt.Errorf("v1 filter backwards compatibility: can't change these fields for a filter with multiple keywords or any statuses: %s", strings.Join(forbiddenFields, ", ")),
			)
		}
	}

	// Now that we've checked that the changes are legal, apply them to the filter and keyword.
	filter.Title = title
	filter.Action = action
	filter.ExpiresAt = expiresAt
	filter.ContextHome = &contextHome
	filter.ContextNotifications = &contextNotifications
	filter.ContextPublic = &contextPublic
	filter.ContextThread = &contextThread
	filter.ContextAccount = &contextAccount
	filterKeyword.Keyword = form.Phrase
	filterKeyword.WholeWord = util.Ptr(util.PtrOrValue(form.WholeWord, false))

	// We only want to update the relevant filter keyword.
	filter.Keywords = []*gtsmodel.FilterKeyword{filterKeyword}
	filter.Statuses = nil
	filterKeyword.Filter = filter

	filterColumns := []string{
		"title",
		"action",
		"expires_at",
		"context_home",
		"context_notifications",
		"context_public",
		"context_thread",
		"context_account",
	}
	filterKeywordColumns := [][]string{
		{
			"keyword",
			"whole_word",
		},
	}
	if err := p.state.DB.UpdateFilter(ctx, filter, filterColumns, filterKeywordColumns, nil, nil); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			err = errors.New("you already have a filter with this title")
			return nil, gtserror.NewErrorConflict(err, err.Error())
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiFilter, errWithCode := p.apiFilter(ctx, filterKeyword)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Send a filters changed event.
	p.stream.FiltersChanged(ctx, account)

	return apiFilter, nil
}
