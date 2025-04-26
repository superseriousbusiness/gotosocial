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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/headerfilter"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// GetAllowHeaderFilter fetches allow HTTP header filter with provided ID from the database.
func (p *Processor) GetAllowHeaderFilter(ctx context.Context, id string) (*apimodel.HeaderFilter, gtserror.WithCode) {
	return p.getHeaderFilter(ctx, id, p.state.DB.GetAllowHeaderFilter)
}

// GetBlockHeaderFilter fetches block HTTP header filter with provided ID from the database.
func (p *Processor) GetBlockHeaderFilter(ctx context.Context, id string) (*apimodel.HeaderFilter, gtserror.WithCode) {
	return p.getHeaderFilter(ctx, id, p.state.DB.GetBlockHeaderFilter)
}

// GetAllowHeaderFilters fetches all allow HTTP header filters stored in the database.
func (p *Processor) GetAllowHeaderFilters(ctx context.Context) ([]*apimodel.HeaderFilter, gtserror.WithCode) {
	return p.getHeaderFilters(ctx, p.state.DB.GetAllowHeaderFilters)
}

// GetBlockHeaderFilters fetches all block HTTP header filters stored in the database.
func (p *Processor) GetBlockHeaderFilters(ctx context.Context) ([]*apimodel.HeaderFilter, gtserror.WithCode) {
	return p.getHeaderFilters(ctx, p.state.DB.GetBlockHeaderFilters)
}

// CreateAllowHeaderFilter inserts the incoming allow HTTP header filter into the database, marking as authored by provided admin account.
func (p *Processor) CreateAllowHeaderFilter(ctx context.Context, admin *gtsmodel.Account, request *apimodel.HeaderFilterRequest) (*apimodel.HeaderFilter, gtserror.WithCode) {
	return p.createHeaderFilter(ctx, admin, request, p.state.DB.PutAllowHeaderFilter)
}

// CreateBlockHeaderFilter inserts the incoming block HTTP header filter into the database, marking as authored by provided admin account.
func (p *Processor) CreateBlockHeaderFilter(ctx context.Context, admin *gtsmodel.Account, request *apimodel.HeaderFilterRequest) (*apimodel.HeaderFilter, gtserror.WithCode) {
	return p.createHeaderFilter(ctx, admin, request, p.state.DB.PutBlockHeaderFilter)
}

// DeleteAllowHeaderFilter deletes the allowing HTTP header filter with provided ID from the database.
func (p *Processor) DeleteAllowHeaderFilter(ctx context.Context, id string) gtserror.WithCode {
	return p.deleteHeaderFilter(ctx, id, p.state.DB.DeleteAllowHeaderFilter)
}

// DeleteBlockHeaderFilter deletes the blocking HTTP header filter with provided ID from the database.
func (p *Processor) DeleteBlockHeaderFilter(ctx context.Context, id string) gtserror.WithCode {
	return p.deleteHeaderFilter(ctx, id, p.state.DB.DeleteBlockHeaderFilter)
}

// getHeaderFilter fetches an HTTP header filter with
// provided ID, using given get function, converting the
// resulting filter to returnable frontend API model.
func (p *Processor) getHeaderFilter(
	ctx context.Context,
	id string,
	get func(context.Context, string) (*gtsmodel.HeaderFilter, error),
) (
	*apimodel.HeaderFilter,
	gtserror.WithCode,
) {
	// Select filter by ID from db.
	filter, err := get(ctx, id)

	switch {
	// Successfully found.
	case err == nil:
		return toAPIHeaderFilter(filter), nil

	// Filter does not exist with ID.
	case errors.Is(err, db.ErrNoEntries):
		const text = "filter not found"
		return nil, gtserror.NewErrorNotFound(errors.New(text), text)

	// Any other error type.
	default:
		err := gtserror.Newf("error selecting from database: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
}

// getHeaderFilters fetches all HTTP header filters
// using given get function, converting the resulting
// filters to returnable frontend API models.
func (p *Processor) getHeaderFilters(
	ctx context.Context,
	get func(context.Context) ([]*gtsmodel.HeaderFilter, error),
) (
	[]*apimodel.HeaderFilter,
	gtserror.WithCode,
) {
	// Select all filters from DB.
	filters, err := get(ctx)

	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Only handle errors other than not-found types.
		err := gtserror.Newf("error selecting from database: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Convert passed header filters to apimodel filters.
	apiFilters := make([]*apimodel.HeaderFilter, len(filters))
	for i := range filters {
		apiFilters[i] = toAPIHeaderFilter(filters[i])
	}

	return apiFilters, nil
}

// createHeaderFilter inserts the given HTTP header
// filter into database, marking as authored by the
// provided admin, using the given insert function.
func (p *Processor) createHeaderFilter(
	ctx context.Context,
	admin *gtsmodel.Account,
	request *apimodel.HeaderFilterRequest,
	insert func(context.Context, *gtsmodel.HeaderFilter) error,
) (
	*apimodel.HeaderFilter,
	gtserror.WithCode,
) {
	// Convert header key to canonical mime header format.
	request.Header = textproto.CanonicalMIMEHeaderKey(request.Header)

	// Validate incoming header filter.
	if errWithCode := validateHeaderFilter(
		request.Header,
		request.Regex,
	); errWithCode != nil {
		return nil, errWithCode
	}

	// Create new database model with ID.
	var filter gtsmodel.HeaderFilter
	filter.ID = id.NewULID()
	filter.Header = request.Header
	filter.Regex = request.Regex
	filter.AuthorID = admin.ID
	filter.Author = admin

	// Insert new header filter into the database.
	if err := insert(ctx, &filter); err != nil {
		err := gtserror.Newf("error inserting into database: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Finally return API model response.
	return toAPIHeaderFilter(&filter), nil
}

// deleteHeaderFilter deletes the HTTP header filter
// with provided ID, using the given delete function.
func (p *Processor) deleteHeaderFilter(
	ctx context.Context,
	id string,
	delete func(context.Context, string) error,
) gtserror.WithCode {
	if err := delete(ctx, id); err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error deleting from database: %w", err)
		return gtserror.NewErrorInternalError(err)
	}
	return nil
}

// toAPIFilter performs a simple conversion of database model HeaderFilter to API model.
func toAPIHeaderFilter(filter *gtsmodel.HeaderFilter) *apimodel.HeaderFilter {
	return &apimodel.HeaderFilter{
		ID:        filter.ID,
		Header:    filter.Header,
		Regex:     filter.Regex,
		CreatedBy: filter.AuthorID,
		CreatedAt: util.FormatISO8601(filter.CreatedAt),
	}
}

// validateHeaderFilter validates incoming filter's header key, and regular expression.
func validateHeaderFilter(header, regex string) gtserror.WithCode {
	// Check header validity (within our own bound checks).
	if header == "" || len(header) > headerfilter.MaxHeaderValue {
		const text = "invalid request header key (empty or too long)"
		return gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Ensure this is compilable regex.
	_, err := regexp.Compile(regex)
	if err != nil {
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	return nil
}
