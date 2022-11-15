/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

func (p *processor) AdminAccountAction(ctx context.Context, authed *oauth.Auth, form *apimodel.AdminAccountActionRequest) gtserror.WithCode {
	return p.adminProcessor.AccountAction(ctx, authed.Account, form)
}

func (p *processor) AdminEmojiCreate(ctx context.Context, authed *oauth.Auth, form *apimodel.EmojiCreateRequest) (*apimodel.Emoji, gtserror.WithCode) {
	return p.adminProcessor.EmojiCreate(ctx, authed.Account, authed.User, form)
}

func (p *processor) AdminEmojisGet(ctx context.Context, authed *oauth.Auth, domain string, includeDisabled bool, includeEnabled bool, shortcode string, maxShortcodeDomain string, minShortcodeDomain string, limit int) (*apimodel.PageableResponse, gtserror.WithCode) {
	return p.adminProcessor.EmojisGet(ctx, authed.Account, authed.User, domain, includeDisabled, includeEnabled, shortcode, maxShortcodeDomain, minShortcodeDomain, limit)
}

func (p *processor) AdminEmojiGet(ctx context.Context, authed *oauth.Auth, id string) (*apimodel.AdminEmoji, gtserror.WithCode) {
	return p.adminProcessor.EmojiGet(ctx, authed.Account, authed.User, id)
}

func (p *processor) AdminEmojiDelete(ctx context.Context, authed *oauth.Auth, id string) (*apimodel.AdminEmoji, gtserror.WithCode) {
	return p.adminProcessor.EmojiDelete(ctx, id)
}

func (p *processor) AdminEmojiCategoriesGet(ctx context.Context) ([]*apimodel.EmojiCategory, gtserror.WithCode) {
	return p.adminProcessor.EmojiCategoriesGet(ctx)
}

func (p *processor) AdminDomainBlockCreate(ctx context.Context, authed *oauth.Auth, form *apimodel.DomainBlockCreateRequest) (*apimodel.DomainBlock, gtserror.WithCode) {
	return p.adminProcessor.DomainBlockCreate(ctx, authed.Account, form.Domain, form.Obfuscate, form.PublicComment, form.PrivateComment, "")
}

func (p *processor) AdminDomainBlocksImport(ctx context.Context, authed *oauth.Auth, form *apimodel.DomainBlockCreateRequest) ([]*apimodel.DomainBlock, gtserror.WithCode) {
	return p.adminProcessor.DomainBlocksImport(ctx, authed.Account, form.Domains)
}

func (p *processor) AdminDomainBlocksGet(ctx context.Context, authed *oauth.Auth, export bool) ([]*apimodel.DomainBlock, gtserror.WithCode) {
	return p.adminProcessor.DomainBlocksGet(ctx, authed.Account, export)
}

func (p *processor) AdminDomainBlockGet(ctx context.Context, authed *oauth.Auth, id string, export bool) (*apimodel.DomainBlock, gtserror.WithCode) {
	return p.adminProcessor.DomainBlockGet(ctx, authed.Account, id, export)
}

func (p *processor) AdminDomainBlockDelete(ctx context.Context, authed *oauth.Auth, id string) (*apimodel.DomainBlock, gtserror.WithCode) {
	return p.adminProcessor.DomainBlockDelete(ctx, authed.Account, id)
}

func (p *processor) AdminMediaPrune(ctx context.Context, mediaRemoteCacheDays int) gtserror.WithCode {
	return p.adminProcessor.MediaPrune(ctx, mediaRemoteCacheDays)
}
