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
	"errors"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// AcceptGet handles the getting of a fedi/activitypub
// representation of a local interaction acceptance.
//
// It performs appropriate authentication before
// returning a JSON serializable interface.
func (p *Processor) AcceptGet(
	ctx context.Context,
	requestedUser string,
	reqID string,
) (interface{}, gtserror.WithCode) {
	// Authenticate incoming request, getting related accounts.
	auth, errWithCode := p.authenticate(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if auth.handshakingURI != nil {
		// We're currently handshaking, which means
		// we don't know this account yet. This should
		// be a very rare race condition.
		err := gtserror.Newf("network race handshaking %s", auth.handshakingURI)
		return nil, gtserror.NewErrorInternalError(err)
	}

	receivingAcct := auth.receivingAcct

	req, err := p.state.DB.GetInteractionRequestByID(ctx, reqID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction request %s: %w", reqID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if req == nil || !req.IsAccepted() {
		// Request doesn't exist or hasn't been accepted.
		err := gtserror.Newf("interaction request %s not found", reqID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if req.TargetAccountID != receivingAcct.ID {
		const text = "interaction request does not belong to receiving account"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	accept, err := p.converter.InteractionReqToASAccept(ctx, req)
	if err != nil {
		err := gtserror.Newf("error converting accept: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.Serialize(accept)
	if err != nil {
		err := gtserror.Newf("error serializing accept: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
