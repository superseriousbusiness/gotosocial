/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package processing

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) FollowRequestsGet(ctx context.Context, auth *oauth.Auth) ([]apimodel.Account, gtserror.WithCode) {
	frs, err := p.db.GetAccountFollowRequests(ctx, auth.Account.ID)
	if err != nil {
		if err != db.ErrNoEntries {
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	accts := []apimodel.Account{}
	for _, fr := range frs {
		if fr.Account == nil {
			frAcct, err := p.db.GetAccountByID(ctx, fr.AccountID)
			if err != nil {
				return nil, gtserror.NewErrorInternalError(err)
			}
			fr.Account = frAcct
		}

		apiAcct, err := p.tc.AccountToAPIAccountPublic(ctx, fr.Account)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		accts = append(accts, *apiAcct)
	}
	return accts, nil
}

func (p *processor) FollowRequestAccept(ctx context.Context, auth *oauth.Auth, accountID string) (*apimodel.Relationship, gtserror.WithCode) {
	follow, err := p.db.AcceptFollowRequest(ctx, accountID, auth.Account.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	if follow.Account == nil {
		followAccount, err := p.db.GetAccountByID(ctx, follow.AccountID)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		follow.Account = followAccount
	}

	if follow.TargetAccount == nil {
		followTargetAccount, err := p.db.GetAccountByID(ctx, follow.TargetAccountID)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		follow.TargetAccount = followTargetAccount
	}

	p.clientWorker.Queue(messages.FromClientAPI{
		APObjectType:   ap.ActivityFollow,
		APActivityType: ap.ActivityAccept,
		GTSModel:       follow,
		OriginAccount:  follow.Account,
		TargetAccount:  follow.TargetAccount,
	})

	gtsR, err := p.db.GetRelationship(ctx, auth.Account.ID, accountID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	r, err := p.tc.RelationshipToAPIRelationship(ctx, gtsR)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return r, nil
}

func (p *processor) FollowRequestReject(ctx context.Context, auth *oauth.Auth, accountID string) (*apimodel.Relationship, gtserror.WithCode) {
	followRequest, err := p.db.RejectFollowRequest(ctx, accountID, auth.Account.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	if followRequest.Account == nil {
		a, err := p.db.GetAccountByID(ctx, followRequest.AccountID)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		followRequest.Account = a
	}

	if followRequest.TargetAccount == nil {
		a, err := p.db.GetAccountByID(ctx, followRequest.TargetAccountID)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		followRequest.TargetAccount = a
	}

	p.clientWorker.Queue(messages.FromClientAPI{
		APObjectType:   ap.ActivityFollow,
		APActivityType: ap.ActivityReject,
		GTSModel:       followRequest,
		OriginAccount:  followRequest.Account,
		TargetAccount:  followRequest.TargetAccount,
	})

	gtsR, err := p.db.GetRelationship(ctx, auth.Account.ID, accountID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	r, err := p.tc.RelationshipToAPIRelationship(ctx, gtsR)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return r, nil
}
