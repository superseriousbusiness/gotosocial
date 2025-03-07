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

package status

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// Get gets the given status, taking account of privacy settings and blocks etc.
func (p *Processor) Get(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	target, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requestingAccount,
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}
	return p.c.GetAPIStatus(ctx, requestingAccount, target)
}

// SourceGet returns the *apimodel.StatusSource version of the targetStatusID.
// Status must belong to the requester, and must not be a boost.
func (p *Processor) SourceGet(ctx context.Context, requester *gtsmodel.Account, statusID string) (*apimodel.StatusSource, gtserror.WithCode) {
	status, errWithCode := p.c.GetOwnStatus(ctx, requester, statusID)
	if errWithCode != nil {
		return nil, errWithCode
	}
	if status.BoostOfID != "" {
		return nil, gtserror.NewErrorNotFound(
			errors.New("status is a boost wrapper"),
			"target status not found",
		)
	}

	// Try to use unparsed content
	// warning text if available,
	// fall back to parsed cw html.
	var spoilerText string
	if status.ContentWarningText != "" {
		spoilerText = status.ContentWarningText
	} else {
		spoilerText = status.ContentWarning
	}

	return &apimodel.StatusSource{
		ID:          status.ID,
		Text:        status.Text,
		SpoilerText: spoilerText,
		ContentType: typeutils.ContentTypeToAPIContentType(status.ContentType),
	}, nil
}
