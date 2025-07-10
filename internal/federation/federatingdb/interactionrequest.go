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

package federatingdb

import (
	"context"
	"net/http"

	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

func (f *DB) LikeRequest(ctx context.Context, likeReq vocab.GoToSocialLikeRequest) error {
	log.DebugKV(ctx, "like", serialize{likeReq})

	// Mark activity as handled.
	f.storeActivityID(likeReq)

	// Extract relevant values from passed ctx.
	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requesting := activityContext.requestingAcct
	receiving := activityContext.receivingAcct

	if requesting.IsMoving() {
		// A Moving account
		// can't do this.
		return nil
	}

	if receiving.IsMoving() {
		// Moving accounts can't
		// do anything with interaction
		// requests, so ignore it.
		return nil
	}

	// Make sure we have a single
	// object of the interaction request.
	objectIRIs := ap.GetObjectIRIs(likeReq)
	if l := len(objectIRIs); l != 1 {
		err := gtserror.Newf("invalid object len %d, wanted 1", l)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}
	statusIRI := objectIRIs[0]
	statusIRIStr := statusIRI.String()

	// Object should be a status.
	status, err := f.state.DB.GetStatusByURI(ctx, statusIRIStr)
	if err != nil {

	}

	

	return nil
}

func (f *DB) ReplyRequest(ctx context.Context, replyReq vocab.GoToSocialReplyRequest) error {
	return nil
}

func (f *DB) AnnounceRequest(ctx context.Context, announceReq vocab.GoToSocialAnnounceRequest) error {
	return nil
}
