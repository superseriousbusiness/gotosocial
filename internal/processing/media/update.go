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
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

// Update updates a media attachment with the given id, using the provided form parameters.
func (p *Processor) Update(ctx context.Context, account *gtsmodel.Account, mediaAttachmentID string, form *apimodel.AttachmentUpdateRequest) (*apimodel.Attachment, gtserror.WithCode) {
	attachment, err := p.state.DB.GetAttachmentByID(ctx, mediaAttachmentID)
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

	var updatingColumns []string

	if form.Description != nil {
		// Sanitize and validate incoming description.
		description, errWithCode := processDescription(
			*form.Description,
		)
		if errWithCode != nil {
			return nil, errWithCode
		}

		attachment.Description = description
		updatingColumns = append(updatingColumns, "description")
	}

	if form.Focus != nil {
		// Parse focus details from API form input.
		focusX, focusY, errWithCode := apiutil.ParseFocus(*form.Focus)
		if errWithCode != nil {
			return nil, errWithCode
		}

		attachment.FileMeta.Focus.X = focusX
		attachment.FileMeta.Focus.Y = focusY
		updatingColumns = append(updatingColumns, "focus_x", "focus_y")
	}

	if err := p.state.DB.UpdateAttachment(ctx, attachment, updatingColumns...); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("database error updating media: %s", err))
	}

	a, err := p.converter.AttachmentToAPIAttachment(ctx, attachment)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error converting attachment: %s", err))
	}

	return &a, nil
}

// processDescription will sanitize and valid description against server configuration.
func processDescription(description string) (string, gtserror.WithCode) {
	description = text.StripHTMLFromText(description)
	chars := len([]rune(description))

	if min := config.GetMediaDescriptionMinChars(); chars < min {
		text := fmt.Sprintf("media description less than min chars (%d)", min)
		return "", gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	if max := config.GetMediaDescriptionMaxChars(); chars > max {
		text := fmt.Sprintf("media description exceeds max chars (%d)", max)
		return "", gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	return description, nil
}
