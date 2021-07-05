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
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) AccountCreate(authed *oauth.Auth, form *apimodel.AccountCreateRequest) (*apimodel.Token, error) {
	return p.accountProcessor.Create(authed.Token, authed.Application, form)
}

func (p *processor) AccountGet(authed *oauth.Auth, targetAccountID string) (*apimodel.Account, error) {
	return p.accountProcessor.Get(authed.Account, targetAccountID)
}

func (p *processor) AccountUpdate(authed *oauth.Auth, form *apimodel.UpdateCredentialsRequest) (*apimodel.Account, error) {
	return p.accountProcessor.Update(authed.Account, form)
}

func (p *processor) AccountStatusesGet(authed *oauth.Auth, targetAccountID string, limit int, excludeReplies bool, maxID string, pinnedOnly bool, mediaOnly bool) ([]apimodel.Status, gtserror.WithCode) {
	return p.accountProcessor.StatusesGet(authed.Account, targetAccountID, limit, excludeReplies, maxID, pinnedOnly, mediaOnly)
}

func (p *processor) AccountFollowersGet(authed *oauth.Auth, targetAccountID string) ([]apimodel.Account, gtserror.WithCode) {
	return p.accountProcessor.FollowersGet(authed.Account, targetAccountID)
}

func (p *processor) AccountFollowingGet(authed *oauth.Auth, targetAccountID string) ([]apimodel.Account, gtserror.WithCode) {
	return p.accountProcessor.FollowingGet(authed.Account, targetAccountID)
}

func (p *processor) AccountRelationshipGet(authed *oauth.Auth, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	return p.accountProcessor.RelationshipGet(authed.Account, targetAccountID)
}

func (p *processor) AccountFollowCreate(authed *oauth.Auth, form *apimodel.AccountFollowRequest) (*apimodel.Relationship, gtserror.WithCode) {
	return p.accountProcessor.FollowCreate(authed.Account, form)
}

func (p *processor) AccountFollowRemove(authed *oauth.Auth, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	return p.accountProcessor.FollowRemove(authed.Account, targetAccountID)
}
