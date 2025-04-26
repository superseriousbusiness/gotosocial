// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package media

import (
	"context"
	"fmt"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// DeleteAvatar deletes the account's avatar, if one exists, and returns the updated account.
// If no avatar exists, it returns anyway with no error.
func (p *Processor) DeleteAvatar(
	ctx context.Context,
	account *gtsmodel.Account,
) (*apimodel.Account, gtserror.WithCode) {
	attachmentID := account.AvatarMediaAttachmentID
	account.AvatarMediaAttachmentID = ""
	return p.deleteProfileAttachment(ctx, account, "avatar_media_attachment_id", attachmentID)
}

// DeleteHeader deletes the account's header, if one exists, and returns the updated account.
// If no header exists, it returns anyway with no error.
func (p *Processor) DeleteHeader(
	ctx context.Context,
	account *gtsmodel.Account,
) (*apimodel.Account, gtserror.WithCode) {
	attachmentID := account.HeaderMediaAttachmentID
	account.HeaderMediaAttachmentID = ""
	return p.deleteProfileAttachment(ctx, account, "header_media_attachment_id", attachmentID)
}

// deleteProfileAttachment updates an attachment ID column and then deletes the attachment.
// Precondition: the relevant attachment ID field of the account model has already been set to the empty string.
func (p *Processor) deleteProfileAttachment(
	ctx context.Context,
	account *gtsmodel.Account,
	attachmentIDColumn string,
	attachmentID string,
) (*apimodel.Account, gtserror.WithCode) {
	if attachmentID != "" {
		// Remove attachment from account.
		if err := p.state.DB.UpdateAccount(ctx, account, attachmentIDColumn); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("could not update account: %s", err))
		}

		// Delete attachment media.
		if err := p.Delete(ctx, attachmentID); err != nil {
			return nil, err
		}
	}

	acctSensitive, err := p.converter.AccountToAPIAccountSensitive(ctx, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("could not convert account into apisensitive account: %s", err))
	}

	return acctSensitive, nil
}
