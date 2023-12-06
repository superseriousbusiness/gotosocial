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

package admin

import (
	"context"
	"errors"
	"net/textproto"
	"regexp"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func (p *Processor) CreateAllowHeaderFilter(ctx context.Context, admin *gtsmodel.Account, headerKey string, valueExpr string) (*gtsmodel.HeaderFilterAllow, gtserror.WithCode) {
	// Validate incoming header filter.
	if errWithCode := validateFilter(
		headerKey,
		valueExpr,
	); errWithCode != nil {
		return nil, errWithCode
	}

	// Create new database model with ID.
	var filter gtsmodel.HeaderFilterAllow
	filter.ID = id.NewULID()
	filter.Key = headerKey
	filter.Regex = valueExpr
	filter.AuthorID = admin.ID
	filter.Author = admin

	// Insert new header filter into the database, all errs here are fatal.
	if err := p.state.DB.PutAllowHeaderFilter(ctx, &filter); err != nil {
		err := gtserror.Newf("error inserting into database: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return &filter, nil
}

func (p *Processor) CreateBlockHeaderFilter(ctx context.Context, admin *gtsmodel.Account, headerKey string, valueExpr string) (*gtsmodel.HeaderFilterBlock, gtserror.WithCode) {
	// Validate incoming header filter.
	if errWithCode := validateFilter(
		headerKey,
		valueExpr,
	); errWithCode != nil {
		return nil, errWithCode
	}

	// Create new database model with ID.
	var filter gtsmodel.HeaderFilterBlock
	filter.ID = id.NewULID()
	filter.Key = headerKey
	filter.Regex = valueExpr
	filter.AuthorID = admin.ID
	filter.Author = admin

	// Insert new header filter into the database, all errs here are fatal.
	if err := p.state.DB.PutBlockHeaderFilter(ctx, &filter); err != nil {
		err := gtserror.Newf("error inserting into database: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return &filter, nil
}

func (p *Processor) GetAllowHeaderFilters(ctx context.Context) ([]*gtsmodel.HeaderFilterAllow, gtserror.WithCode) {

}

func (p *Processor) GetBlockHeaderFilters(ctx context.Context) ([]*gtsmodel.HeaderFilterAllow, gtserror.WithCode) {

}

func (p *Processor) DeleteAllowHeaderFilter(ctx context.Context, id string) gtserror.WithCode {

}

func (p *Processor) DeleteBlockHeaderFilter(ctx context.Context, id string) gtserror.WithCode {

}

func validateFilter(headerKey, valueExpr string) gtserror.WithCode {
	// Canonicalize the mime header key and check validity.
	headerKey = textproto.CanonicalMIMEHeaderKey(headerKey)
	if headerKey == "" || len(headerKey) > 1024 {
		const text = "invalid request header key (empty or too long)"
		return gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Ensure value regexp validity.
	_, err := regexp.Compile(valueExpr)
	if err != nil {
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	return nil
}
