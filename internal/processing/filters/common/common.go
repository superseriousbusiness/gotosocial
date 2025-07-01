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

package common

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/processing/stream"
	"code.superseriousbusiness.org/gotosocial/internal/state"
)

type Processor struct {
	state  *state.State
	stream *stream.Processor
}

func New(state *state.State, stream *stream.Processor) *Processor {
	return &Processor{state, stream}
}

// CheckFilterExists calls .GetFilter() with a barebones context to not
// fetch any sub-models, and not returning the result. this functionally
// just uses .GetFilter() for the ownership and existence checks.
func (p *Processor) CheckFilterExists(
	ctx context.Context,
	requester *gtsmodel.Account,
	id string,
) gtserror.WithCode {
	_, errWithCode := p.GetFilter(gtscontext.SetBarebones(ctx), requester, id)
	return errWithCode
}

// GetFilter fetches the filter with given ID, also checking
// the given requesting account is the owner of the filter.
func (p *Processor) GetFilter(
	ctx context.Context,
	requester *gtsmodel.Account,
	id string,
) (
	*gtsmodel.Filter,
	gtserror.WithCode,
) {
	// Get the filter from the database with given ID.
	filter, err := p.state.DB.GetFilterByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error getting filter: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check it exists.
	if filter == nil {
		const text = "filter not found"
		return nil, gtserror.NewWithCode(http.StatusNotFound, text)
	}

	// Check that the requester owns it.
	if filter.AccountID != requester.ID {
		const text = "filter not found"
		err := gtserror.New("filter does not belong to account")
		return nil, gtserror.NewErrorNotFound(err, text)
	}

	return filter, nil
}

// GetFilterStatus fetches the filter status with given ID, also
// checking the given requesting account is the owner of it.
func (p *Processor) GetFilterStatus(
	ctx context.Context,
	requester *gtsmodel.Account,
	id string,
) (
	*gtsmodel.FilterStatus,
	*gtsmodel.Filter,
	gtserror.WithCode,
) {

	// Get the filter status from the database with given ID.
	filterStatus, err := p.state.DB.GetFilterStatusByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error getting filter status: %w", err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	// Check it even exists.
	if filterStatus == nil {
		const text = "filter status not found"
		return nil, nil, gtserror.NewWithCode(http.StatusNotFound, text)
	}

	// Get the filter this filter status is
	// associated with, without sub-models.
	// (this also checks filter ownership).
	filter, errWithCode := p.GetFilter(
		gtscontext.SetBarebones(ctx),
		requester,
		filterStatus.FilterID,
	)
	if errWithCode != nil {
		return nil, nil, errWithCode
	}

	return filterStatus, filter, nil
}

// GetFilterKeyword fetches the filter keyword with given ID,
// also checking the given requesting account is the owner of it.
func (p *Processor) GetFilterKeyword(
	ctx context.Context,
	requester *gtsmodel.Account,
	id string,
) (
	*gtsmodel.FilterKeyword,
	*gtsmodel.Filter,
	gtserror.WithCode,
) {

	// Get the filter keyword from the database with given ID.
	keyword, err := p.state.DB.GetFilterKeywordByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error getting filter keyword: %w", err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	// Check it exists.
	if keyword == nil {
		const text = "filter keyword not found"
		return nil, nil, gtserror.NewWithCode(http.StatusNotFound, text)
	}

	// Get the filter this filter keyword is
	// associated with, without sub-models.
	// (this also checks filter ownership).
	filter, errWithCode := p.GetFilter(
		gtscontext.SetBarebones(ctx),
		requester,
		keyword.FilterID,
	)
	if errWithCode != nil {
		return nil, nil, errWithCode
	}

	return keyword, filter, nil
}

// OnFilterChanged ...
func (p *Processor) OnFilterChanged(ctx context.Context, requester *gtsmodel.Account) {

	// Get list of list IDs created by this requesting account.
	listIDs, err := p.state.DB.GetListIDsByAccountID(ctx, requester.ID)
	if err != nil {
		log.Errorf(ctx, "error getting account '%s' lists: %v", requester.Username, err)
	}

	// Unprepare this requester's home timeline.
	p.state.Caches.Timelines.Home.Unprepare(requester.ID)

	// Unprepare list timelines.
	for _, id := range listIDs {
		p.state.Caches.Timelines.List.Unprepare(id)
	}

	// Send filter changed event for account.
	p.stream.FiltersChanged(ctx, requester)
}

// FromAPIContexts converts a slice of frontend API model FilterContext types to our internal FilterContexts bit field.
func FromAPIContexts(apiContexts []apimodel.FilterContext) (gtsmodel.FilterContexts, gtserror.WithCode) {
	var contexts gtsmodel.FilterContexts
	for _, context := range apiContexts {
		switch context {
		case apimodel.FilterContextHome:
			contexts.SetHome()
		case apimodel.FilterContextNotifications:
			contexts.SetNotifications()
		case apimodel.FilterContextPublic:
			contexts.SetPublic()
		case apimodel.FilterContextThread:
			contexts.SetThread()
		case apimodel.FilterContextAccount:
			contexts.SetAccount()
		default:
			text := fmt.Sprintf("unsupported filter context: %s", context)
			return 0, gtserror.NewWithCode(http.StatusBadRequest, text)
		}
	}
	return contexts, nil
}
