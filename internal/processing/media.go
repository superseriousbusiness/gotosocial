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

func (p *processor) MediaCreate(authed *oauth.Auth, form *apimodel.AttachmentRequest) (*apimodel.Attachment, error) {
	return p.mediaProcessor.Create(authed.Account, form)
}

func (p *processor) MediaGet(authed *oauth.Auth, mediaAttachmentID string) (*apimodel.Attachment, gtserror.WithCode) {
	return p.mediaProcessor.GetMedia(authed.Account, mediaAttachmentID)
}

func (p *processor) MediaUpdate(authed *oauth.Auth, mediaAttachmentID string, form *apimodel.AttachmentUpdateRequest) (*apimodel.Attachment, gtserror.WithCode) {
	return p.mediaProcessor.Update(authed.Account, mediaAttachmentID, form)
}

func (p *processor) FileGet(authed *oauth.Auth, form *apimodel.GetContentRequestForm) (*apimodel.Content, error) {
	return p.mediaProcessor.GetFile(authed.Account, form)
}
