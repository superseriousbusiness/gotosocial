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
	"errors"
	"fmt"

	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (f *federatingDB) Accept(ctx context.Context, accept vocab.ActivityStreamsAccept) error {
	if log.Level() >= level.DEBUG {
		i, err := marshalItem(accept)
		if err != nil {
			return err
		}
		l := log.WithContext(ctx).
			WithField("accept", i)
		l.Debug("entering Accept")
	}

	receivingAccount, _, internal := extractFromCtx(ctx)
	if internal {
		return nil // Already processed.
	}

	// Iterate all provided objects in the activity.
	for _, object := range ap.ExtractObjects(accept) {

		// Check and handle any vocab.Type objects.
		if objType := object.GetType(); objType != nil {
			switch objType.GetTypeName() { //nolint:gocritic

			case ap.ActivityFollow:
				// Cast the vocab.Type object to known AS type.
				asFollow := objType.(vocab.ActivityStreamsFollow)

				// convert the follow to something we can understand
				gtsFollow, err := f.converter.ASFollowToFollow(ctx, asFollow)
				if err != nil {
					return fmt.Errorf("ACCEPT: error converting asfollow to gtsfollow: %s", err)
				}

				// make sure the addressee of the original follow is the same as whatever inbox this landed in
				if gtsFollow.AccountID != receivingAccount.ID {
					return errors.New("ACCEPT: follow object account and inbox account were not the same")
				}

				follow, err := f.state.DB.AcceptFollowRequest(ctx, gtsFollow.AccountID, gtsFollow.TargetAccountID)
				if err != nil {
					return err
				}

				f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
					APObjectType:     ap.ActivityFollow,
					APActivityType:   ap.ActivityAccept,
					GTSModel:         follow,
					ReceivingAccount: receivingAccount,
				})
			}

			continue
		}

		// Check and handle any
		// IRI type objects.
		if object.IsIRI() {

			// Extract IRI from object.
			iri := object.GetIRI()
			if !uris.IsFollowPath(iri) {
				continue
			}

			// Serialize IRI.
			iriStr := iri.String()

			// ACCEPT FOLLOW
			followReq, err := f.state.DB.GetFollowRequestByURI(ctx, iriStr)
			if err != nil {
				return fmt.Errorf("ACCEPT: couldn't get follow request with id %s from the database: %s", iriStr, err)
			}

			// make sure the addressee of the original follow is the same as whatever inbox this landed in
			if followReq.AccountID != receivingAccount.ID {
				return errors.New("ACCEPT: follow object account and inbox account were not the same")
			}

			follow, err := f.state.DB.AcceptFollowRequest(ctx, followReq.AccountID, followReq.TargetAccountID)
			if err != nil {
				return err
			}

			f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
				APObjectType:     ap.ActivityFollow,
				APActivityType:   ap.ActivityAccept,
				GTSModel:         follow,
				ReceivingAccount: receivingAccount,
			})

			continue
		}

	}

	return nil
}
