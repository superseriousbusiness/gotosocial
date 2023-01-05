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
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) AccountCreate(ctx context.Context, authed *oauth.Auth, form *apimodel.AccountCreateRequest) (*apimodel.Token, gtserror.WithCode) {
	return p.accountProcessor.Create(ctx, authed.Token, authed.Application, form)
}

func (p *processor) AccountDeleteLocal(ctx context.Context, authed *oauth.Auth, form *apimodel.AccountDeleteRequest) gtserror.WithCode {
	return p.accountProcessor.DeleteLocal(ctx, authed.Account, form)
}

func (p *processor) AccountGet(ctx context.Context, authed *oauth.Auth, targetAccountID string) (*apimodel.Account, gtserror.WithCode) {
	return p.accountProcessor.Get(ctx, authed.Account, targetAccountID)
}

func (p *processor) AccountGetLocalByUsername(ctx context.Context, authed *oauth.Auth, username string) (*apimodel.Account, gtserror.WithCode) {
	return p.accountProcessor.GetLocalByUsername(ctx, authed.Account, username)
}

func (p *processor) AccountGetCustomCSSForUsername(ctx context.Context, username string) (string, gtserror.WithCode) {
	return p.accountProcessor.GetCustomCSSForUsername(ctx, username)
}

func (p *processor) AccountGetRSSFeedForUsername(ctx context.Context, username string) (func() (string, gtserror.WithCode), time.Time, gtserror.WithCode) {
	return p.accountProcessor.GetRSSFeedForUsername(ctx, username)
}

func (p *processor) AccountUpdate(ctx context.Context, authed *oauth.Auth, form *apimodel.UpdateCredentialsRequest) (*apimodel.Account, gtserror.WithCode) {
	return p.accountProcessor.Update(ctx, authed.Account, form)
}

func (p *processor) AccountStatusesGet(ctx context.Context, authed *oauth.Auth, targetAccountID string, limit int, excludeReplies bool, excludeReblogs bool, maxID string, minID string, pinnedOnly bool, mediaOnly bool, publicOnly bool) (*apimodel.PageableResponse, gtserror.WithCode) {
	return p.accountProcessor.StatusesGet(ctx, authed.Account, targetAccountID, limit, excludeReplies, excludeReblogs, maxID, minID, pinnedOnly, mediaOnly, publicOnly)
}

func (p *processor) AccountWebStatusesGet(ctx context.Context, targetAccountID string, maxID string) (*apimodel.PageableResponse, gtserror.WithCode) {
	return p.accountProcessor.WebStatusesGet(ctx, targetAccountID, maxID)
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
