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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

func (p *Processor) implicitlyAccept(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (bool, gtserror.WithCode) {
	if status.InReplyToAccountID != requester.ID {
		// Status doesn't reply to us,
		// we can't accept on behalf
		// of someone else.
		return false, nil
	}

	targetPendingApproval := util.PtrOrValue(status.PendingApproval, false)
	if !targetPendingApproval {
		// Status isn't pending approval,
		// nothing to implicitly accept.
		return false, nil
	}

	// Status is pending approval,
	// check for an interaction request.
	intReq, err := p.state.DB.GetInteractionRequestByInteractionURI(ctx, status.URI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Something's gone wrong.
		err := gtserror.Newf("db error getting interaction request for %s: %w", status.URI, err)
		return false, gtserror.NewErrorInternalError(err)
	}

	// No interaction request present
	// for this status. Race condition?
	if intReq == nil {
		return false, nil
	}

	// Accept the interaction.
	if _, errWithCode := p.intReqs.Accept(ctx,
		requester, intReq.ID,
	); errWithCode != nil {
		return false, errWithCode
	}

	return true, nil
}
