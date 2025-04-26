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

package interactionrequests

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// GetPage returns a page of interaction requests targeting
// the requester and (optionally) the given status ID.
func (p *Processor) GetPage(
	ctx context.Context,
	requester *gtsmodel.Account,
	statusID string,
	likes bool,
	replies bool,
	boosts bool,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	reqs, err := p.state.DB.GetInteractionsRequestsForAcct(
		ctx,
		requester.ID,
		statusID,
		likes,
		replies,
		boosts,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction requests: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(reqs)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	var (
		// Get the lowest and highest
		// ID values, used for paging.
		lo = reqs[count-1].ID
		hi = reqs[0].ID

		// Best-guess items length.
		items = make([]interface{}, 0, count)
	)

	for _, req := range reqs {
		apiReq, err := p.converter.InteractionReqToAPIInteractionReq(
			ctx, req, requester,
		)
		if err != nil {
			log.Errorf(ctx, "error converting interaction req to api req: %v", err)
			continue
		}

		// Append req to return items.
		items = append(items, apiReq)
	}

	// Build extra query params to return in Link header.
	extraParams := make(url.Values, 4)
	extraParams.Set(apiutil.InteractionFavouritesKey, strconv.FormatBool(likes))
	extraParams.Set(apiutil.InteractionRepliesKey, strconv.FormatBool(replies))
	extraParams.Set(apiutil.InteractionReblogsKey, strconv.FormatBool(boosts))
	if statusID != "" {
		extraParams.Set(apiutil.InteractionStatusIDKey, statusID)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/interaction_requests",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: extraParams,
	}), nil
}

// GetOne returns one interaction
// request with the given ID.
func (p *Processor) GetOne(
	ctx context.Context,
	requester *gtsmodel.Account,
	id string,
) (*apimodel.InteractionRequest, gtserror.WithCode) {
	req, err := p.state.DB.GetInteractionRequestByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction request: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if req == nil {
		err := gtserror.New("interaction request not found")
		return nil, gtserror.NewErrorNotFound(err)
	}

	if req.TargetAccountID != requester.ID {
		err := gtserror.Newf(
			"interaction request %s does not target account %s",
			req.ID, requester.ID,
		)
		return nil, gtserror.NewErrorNotFound(err)
	}

	apiReq, err := p.converter.InteractionReqToAPIInteractionReq(
		ctx, req, requester,
	)
	if err != nil {
		err := gtserror.Newf("error converting interaction req to api req: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiReq, nil
}
