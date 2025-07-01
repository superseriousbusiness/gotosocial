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
	"net/http"
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/processing/filters/common"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// Create a new filter and filter keyword for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) Create(ctx context.Context, requester *gtsmodel.Account, form *apimodel.FilterCreateUpdateRequestV1) (*apimodel.FilterV1, gtserror.WithCode) {
	var errWithCode gtserror.WithCode

	// Create new wrapping filter.
	filter := &gtsmodel.Filter{
		ID:        id.NewULID(),
		AccountID: requester.ID,
		Title:     form.Phrase,
	}

	if *form.Irreversible {
		// Irreversible = action hide.
		filter.Action = gtsmodel.FilterActionHide
	} else {
		// Default action = action warn.
		filter.Action = gtsmodel.FilterActionWarn
	}

	// Check form for valid expiry and set on filter.
	if form.ExpiresIn != nil && *form.ExpiresIn > 0 {
		expiresIn := time.Duration(*form.ExpiresIn) * time.Second
		filter.ExpiresAt = time.Now().Add(expiresIn)
	}

	// Parse contexts filter applies in from incoming request form data.
	filter.Contexts, errWithCode = common.FromAPIContexts(form.Context)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Create new keyword attached to filter.
	filterKeyword := &gtsmodel.FilterKeyword{
		ID:        id.NewULID(),
		FilterID:  filter.ID,
		Keyword:   form.Phrase,
		WholeWord: util.Ptr(util.PtrOrValue(form.WholeWord, false)),
	}

	// Attach the new keyword to filter before insert.
	filter.Keywords = append(filter.Keywords, filterKeyword)
	filter.KeywordIDs = append(filter.KeywordIDs, filterKeyword.ID)

	// Insert newly created filter into the database.
	switch err := p.state.DB.PutFilter(ctx, filter); {
	case err == nil:
		// no issue

	case errors.Is(err, db.ErrAlreadyExists):
		const text = "duplicate title"
		return nil, gtserror.NewWithCode(http.StatusConflict, text)

	default:
		err := gtserror.Newf("error inserting filter: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Handle filter change side-effects.
	p.c.OnFilterChanged(ctx, requester)

	// Return as converted frontend filter keyword model.
	return typeutils.FilterKeywordToAPIFilterV1(filter, filterKeyword), nil
}
