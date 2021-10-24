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
	"context"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) AccountCreate(ctx context.Context, authed *oauth.Auth, form *apimodel.AccountCreateRequest) (*apimodel.Token, error) {
	return p.accountProcessor.Create(ctx, authed.Token, authed.Application, form)
}

func (p *processor) AccountGet(ctx context.Context, authed *oauth.Auth, targetAccountID string) (*apimodel.Account, error) {
	return p.accountProcessor.Get(ctx, authed.Account, targetAccountID)
}

func (p *processor) AccountUpdate(ctx context.Context, authed *oauth.Auth, form *apimodel.UpdateCredentialsRequest) (*apimodel.Account, error) {
	return p.accountProcessor.Update(ctx, authed.Account, form)
}

func (p *processor) AccountStatusesGet(ctx context.Context, authed *oauth.Auth, targetAccountID string, limit int, excludeReplies bool, maxID string, minID string, pinnedOnly bool, mediaOnly bool, publicOnly bool) ([]apimodel.Status, gtserror.WithCode) {
	return p.accountProcessor.StatusesGet(ctx, authed.Account, targetAccountID, limit, excludeReplies, maxID, minID, pinnedOnly, mediaOnly, publicOnly)
}

func (p *processor) AccountFollowersGet(ctx context.Context, authed *oauth.Auth, targetAccountID string) ([]apimodel.Account, gtserror.WithCode) {
	return p.accountProcessor.FollowersGet(ctx, authed.Account, targetAccountID)
}

func (p *processor) AccountFollowingGet(ctx context.Context, authed *oauth.Auth, targetAccountID string) ([]apimodel.Account, gtserror.WithCode) {
	return p.accountProcessor.FollowingGet(ctx, authed.Account, targetAccountID)
}

func (p *processor) AccountRelationshipGet(ctx context.Context, authed *oauth.Auth, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	return p.accountProcessor.RelationshipGet(ctx, authed.Account, targetAccountID)
}

func (p *processor) AccountFollowCreate(ctx context.Context, authed *oauth.Auth, form *apimodel.AccountFollowRequest) (*apimodel.Relationship, gtserror.WithCode) {
	return p.accountProcessor.FollowCreate(ctx, authed.Account, form)
}

func (p *processor) AccountFollowRemove(ctx context.Context, authed *oauth.Auth, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	return p.accountProcessor.FollowRemove(ctx, authed.Account, targetAccountID)
}

func (p *processor) AccountBlockCreate(ctx context.Context, authed *oauth.Auth, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	return p.accountProcessor.BlockCreate(ctx, authed.Account, targetAccountID)
}

func (p *processor) AccountBlockRemove(ctx context.Context, authed *oauth.Auth, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	return p.accountProcessor.BlockRemove(ctx, authed.Account, targetAccountID)
}
