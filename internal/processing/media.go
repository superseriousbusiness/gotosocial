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

func (p *processor) MediaCreate(ctx context.Context, authed *oauth.Auth, form *apimodel.AttachmentRequest) (*apimodel.Attachment, gtserror.WithCode) {
	return p.mediaProcessor.Create(ctx, authed.Account, form)
}

func (p *processor) MediaGet(ctx context.Context, authed *oauth.Auth, mediaAttachmentID string) (*apimodel.Attachment, gtserror.WithCode) {
	return p.mediaProcessor.GetMedia(ctx, authed.Account, mediaAttachmentID)
}

func (p *processor) MediaUpdate(ctx context.Context, authed *oauth.Auth, mediaAttachmentID string, form *apimodel.AttachmentUpdateRequest) (*apimodel.Attachment, gtserror.WithCode) {
	return p.mediaProcessor.Update(ctx, authed.Account, mediaAttachmentID, form)
}

func (p *processor) FileGet(ctx context.Context, authed *oauth.Auth, form *apimodel.GetContentRequestForm) (*apimodel.Content, gtserror.WithCode) {
	return p.mediaProcessor.GetFile(ctx, authed.Account, form)
}

func (p *processor) CustomEmojisGet(ctx context.Context) ([]*apimodel.Emoji, gtserror.WithCode) {
	return p.mediaProcessor.GetCustomEmojis(ctx)
}
