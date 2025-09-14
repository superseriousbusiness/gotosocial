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

package fedi

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// AuthorizationGet handles the getting of a fedi/activitypub
// representation of a local interaction authorization.
//
// It performs appropriate authentication before
// returning a JSON serializable interface.
func (p *Processor) AuthorizationGet(
	ctx context.Context,
	requestedUser string,
	intReqID string,
) (any, gtserror.WithCode) {
	// Ensure valid request, intReq exists, etc.
	intReq, errWithCode := p.validateIntReqRequest(ctx, requestedUser, intReqID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Convert + serialize the Authorization.
	authorization, err := p.converter.InteractionReqToASAuthorization(ctx, intReq)
	if err != nil {
		err := gtserror.Newf("error converting to authorization: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.Serialize(authorization)
	if err != nil {
		err := gtserror.Newf("error serializing accept: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
