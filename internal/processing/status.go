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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) StatusCreate(ctx context.Context, authed *oauth.Auth, form *apimodel.AdvancedStatusCreateForm) (*apimodel.Status, gtserror.WithCode) {
	return p.statusProcessor.Create(ctx, authed.Account, authed.Application, form)
}

func (p *processor) StatusDelete(ctx context.Context, authed *oauth.Auth, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	return p.statusProcessor.Delete(ctx, authed.Account, targetStatusID)
}

func (p *processor) StatusFave(ctx context.Context, authed *oauth.Auth, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	return p.statusProcessor.Fave(ctx, authed.Account, targetStatusID)
}

func (p *processor) StatusBoost(ctx context.Context, authed *oauth.Auth, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	return p.statusProcessor.Boost(ctx, authed.Account, authed.Application, targetStatusID)
}

func (p *processor) StatusUnboost(ctx context.Context, authed *oauth.Auth, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	return p.statusProcessor.Unboost(ctx, authed.Account, authed.Application, targetStatusID)
}

func (p *processor) StatusBoostedBy(ctx context.Context, authed *oauth.Auth, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode) {
	return p.statusProcessor.BoostedBy(ctx, authed.Account, targetStatusID)
}

func (p *processor) StatusFavedBy(ctx context.Context, authed *oauth.Auth, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode) {
	return p.statusProcessor.FavedBy(ctx, authed.Account, targetStatusID)
}

func (p *processor) StatusGet(ctx context.Context, authed *oauth.Auth, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	return p.statusProcessor.Get(ctx, authed.Account, targetStatusID)
}

func (p *processor) StatusUnfave(ctx context.Context, authed *oauth.Auth, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	return p.statusProcessor.Unfave(ctx, authed.Account, targetStatusID)
}

func (p *processor) StatusGetContext(ctx context.Context, authed *oauth.Auth, targetStatusID string) (*apimodel.Context, gtserror.WithCode) {
	return p.statusProcessor.Context(ctx, authed.Account, targetStatusID)
}

func (p *processor) StatusBookmark(ctx context.Context, authed *oauth.Auth, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	return p.statusProcessor.Bookmark(ctx, authed.Account, targetStatusID)
}

func (p *processor) StatusUnbookmark(ctx context.Context, authed *oauth.Auth, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	return p.statusProcessor.Unbookmark(ctx, authed.Account, targetStatusID)
}
