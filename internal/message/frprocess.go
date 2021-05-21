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

package message

import (
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) FollowRequestsGet(auth *oauth.Auth) ([]apimodel.Account, ErrorWithCode) {
	frs := []gtsmodel.FollowRequest{}
	if err := p.db.GetFollowRequestsForAccountID(auth.Account.ID, &frs); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, NewErrorInternalError(err)
		}
	}

	accts := []apimodel.Account{}
	for _, fr := range frs {
		acct := &gtsmodel.Account{}
		if err := p.db.GetByID(fr.AccountID, acct); err != nil {
			return nil, NewErrorInternalError(err)
		}
		mastoAcct, err := p.tc.AccountToMastoPublic(acct)
		if err != nil {
			return nil, NewErrorInternalError(err)
		}
		accts = append(accts, *mastoAcct)
	}
	return accts, nil
}

func (p *processor) FollowRequestAccept(auth *oauth.Auth, accountID string) (*apimodel.Relationship, ErrorWithCode) {
	follow, err := p.db.AcceptFollowRequest(accountID, auth.Account.ID)
	if err != nil {
		return nil, NewErrorNotFound(err)
	}

	p.fromClientAPI <- gtsmodel.FromClientAPI{
		APActivityType: gtsmodel.ActivityStreamsAccept,
		GTSModel:       follow,
	}

	gtsR, err := p.db.GetRelationship(auth.Account.ID, accountID)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	r, err := p.tc.RelationshipToMasto(gtsR)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	return r, nil
}

func (p *processor) FollowRequestDeny(auth *oauth.Auth) ErrorWithCode {
	return nil
}
