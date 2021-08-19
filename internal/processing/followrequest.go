/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) FollowRequestsGet(auth *oauth.Auth) ([]apimodel.Account, gtserror.WithCode) {
	frs, err := p.db.GetAccountFollowRequests(auth.Account.ID)
	if err != nil {
		if err != db.ErrNoEntries {
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	accts := []apimodel.Account{}
	for _, fr := range frs {
		acct := &gtsmodel.Account{}
		if err := p.db.GetByID(fr.AccountID, acct); err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		mastoAcct, err := p.tc.AccountToMastoPublic(acct)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		accts = append(accts, *mastoAcct)
	}
	return accts, nil
}

func (p *processor) FollowRequestAccept(auth *oauth.Auth, accountID string) (*apimodel.Relationship, gtserror.WithCode) {
	follow, err := p.db.AcceptFollowRequest(accountID, auth.Account.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	originAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(follow.AccountID, originAccount); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(follow.TargetAccountID, targetAccount); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	p.fromClientAPI <- gtsmodel.FromClientAPI{
		APObjectType:   gtsmodel.ActivityStreamsFollow,
		APActivityType: gtsmodel.ActivityStreamsAccept,
		GTSModel:       follow,
		OriginAccount:  originAccount,
		TargetAccount:  targetAccount,
	}

	gtsR, err := p.db.GetRelationship(auth.Account.ID, accountID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	r, err := p.tc.RelationshipToMasto(gtsR)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return r, nil
}

func (p *processor) FollowRequestDeny(auth *oauth.Auth) gtserror.WithCode {
	return nil
}
