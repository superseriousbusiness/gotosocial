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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

func (p *Processor) TokensGet(
	ctx context.Context,
	userID string,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	tokens, err := p.state.DB.GetAccessTokens(ctx, userID, page)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting tokens: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(tokens)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	var (
		// Get the lowest and highest
		// ID values, used for paging.
		lo = tokens[count-1].ID
		hi = tokens[0].ID

		// Best-guess items length.
		items = make([]interface{}, 0, count)
	)

	for _, token := range tokens {
		tokenInfo, err := p.converter.TokenToAPITokenInfo(ctx, token)
		if err != nil {
			log.Errorf(ctx, "error converting token to api token info: %v", err)
			continue
		}

		// Append req to return items.
		items = append(items, tokenInfo)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/tokens",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}

func (p *Processor) TokenGet(
	ctx context.Context,
	userID string,
	tokenID string,
) (*apimodel.TokenInfo, gtserror.WithCode) {
	token, err := p.state.DB.GetTokenByID(ctx, tokenID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting token %s: %w", tokenID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if token == nil {
		err := gtserror.Newf("token %s not found in the db", tokenID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if token.UserID != userID {
		err := gtserror.Newf("token %s does not belong to user %s", tokenID, userID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	tokenInfo, err := p.converter.TokenToAPITokenInfo(ctx, token)
	if err != nil {
		err := gtserror.Newf("error converting token to api token info: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return tokenInfo, nil
}

func (p *Processor) TokenInvalidate(
	ctx context.Context,
	userID string,
	tokenID string,
) (*apimodel.TokenInfo, gtserror.WithCode) {
	tokenInfo, errWithCode := p.TokenGet(ctx, userID, tokenID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if err := p.state.DB.DeleteTokenByID(ctx, tokenID); err != nil {
		err := gtserror.Newf("db error deleting token %s: %w", tokenID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return tokenInfo, nil
}
