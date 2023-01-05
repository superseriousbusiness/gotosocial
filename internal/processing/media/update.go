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

package media

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

func (p *processor) Update(ctx context.Context, account *gtsmodel.Account, mediaAttachmentID string, form *apimodel.AttachmentUpdateRequest) (*apimodel.Attachment, gtserror.WithCode) {
	attachment, err := p.db.GetAttachmentByID(ctx, mediaAttachmentID)
	if err != nil {
		if err == db.ErrNoEntries {
			// attachment doesn't exist
			return nil, gtserror.NewErrorNotFound(errors.New("attachment doesn't exist in the db"))
		}
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("db error getting attachment: %s", err))
	}

	if attachment.AccountID != account.ID {
		return nil, gtserror.NewErrorNotFound(errors.New("attachment not owned by requesting account"))
	}

	updatingColumns := []string{}

	if form.Description != nil {
		attachment.Description = text.SanitizePlaintext(*form.Description)
		updatingColumns = append(updatingColumns, "description")
	}

	if form.Focus != nil {
		focusx, focusy, err := parseFocus(*form.Focus)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}
		attachment.FileMeta.Focus.X = focusx
		attachment.FileMeta.Focus.Y = focusy
		updatingColumns = append(updatingColumns, "focus_x", "focus_y")
	}

	if err := p.db.UpdateByID(ctx, attachment, attachment.ID, updatingColumns...); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("database error updating media: %s", err))
	}

	a, err := p.tc.AttachmentToAPIAttachment(ctx, attachment)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error converting attachment: %s", err))
	}

	return &a, nil
}
