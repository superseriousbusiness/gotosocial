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

package processing

import (
	"context"
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *Processor) FollowRequestsGet(ctx context.Context, auth *oauth.Auth) ([]apimodel.Account, gtserror.WithCode) {
	followRequests, err := p.state.DB.GetAccountFollowRequests(ctx, auth.Account.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	accts := make([]apimodel.Account, 0, len(followRequests))
	for _, followRequest := range followRequests {
		if followRequest.Account == nil {
			// The creator of the follow doesn't exist,
			// just skip this one.
			log.WithContext(ctx).WithField("followRequest", followRequest).Warn("follow request had no associated account")
			continue
		}

		apiAcct, err := p.tc.AccountToAPIAccountPublic(ctx, followRequest.Account)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		accts = append(accts, *apiAcct)
	}

	return accts, nil
}

func (p *Processor) FollowRequestAccept(ctx context.Context, auth *oauth.Auth, accountID string) (*apimodel.Relationship, gtserror.WithCode) {
	follow, err := p.state.DB.AcceptFollowRequest(ctx, accountID, auth.Account.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	if follow.Account == nil {
		// The creator of the follow doesn't exist,
		// so we can't do further processing.
		log.WithContext(ctx).WithField("follow", follow).Warn("follow had no associated account")
		return p.relationship(ctx, auth.Account.ID, accountID)
	}

	p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ActivityFollow,
		APActivityType: ap.ActivityAccept,
		GTSModel:       follow,
		OriginAccount:  follow.Account,
		TargetAccount:  follow.TargetAccount,
	})

	return p.relationship(ctx, auth.Account.ID, accountID)
}

func (p *Processor) FollowRequestReject(ctx context.Context, auth *oauth.Auth, accountID string) (*apimodel.Relationship, gtserror.WithCode) {
	followRequest, err := p.state.DB.GetFollowRequest(ctx, accountID, auth.Account.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	err = p.state.DB.RejectFollowRequest(ctx, accountID, auth.Account.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	if followRequest.Account == nil {
		// The creator of the request doesn't exist,
		// so we can't do further processing.
		return p.relationship(ctx, auth.Account.ID, accountID)
	}

	p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ActivityFollow,
		APActivityType: ap.ActivityReject,
		GTSModel:       followRequest,
		OriginAccount:  followRequest.Account,
		TargetAccount:  followRequest.TargetAccount,
	})

	return p.relationship(ctx, auth.Account.ID, accountID)
}

func (p *Processor) relationship(ctx context.Context, accountID string, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	relationship, err := p.state.DB.GetRelationship(ctx, accountID, targetAccountID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiRelationship, err := p.tc.RelationshipToAPIRelationship(ctx, relationship)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiRelationship, nil
}
