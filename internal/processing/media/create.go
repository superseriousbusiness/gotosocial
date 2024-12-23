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
	"io"

	"codeberg.org/gruf/go-iotools"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

// Create creates a new media attachment belonging to the given account, using the request form.
func (p *Processor) Create(ctx context.Context, account *gtsmodel.Account, form *apimodel.AttachmentRequest) (*apimodel.Attachment, gtserror.WithCode) {

	// Get maximum supported local media size.
	maxsz := config.GetMediaLocalMaxSize()
	maxszInt64 := int64(maxsz) // #nosec G115 -- Already validated.

	// Ensure media within size bounds.
	if form.File.Size > maxszInt64 {
		text := fmt.Sprintf("media exceeds configured max size: %s", maxsz)
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Parse focus details from API form input.
	focusX, focusY, errWithCode := apiutil.ParseFocus(form.Focus)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// If description provided,
	// process and validate it.
	//
	// This may not yet be set as it
	// is often set on status post.
	if form.Description != "" {
		form.Description, errWithCode = processDescription(form.Description)
		if errWithCode != nil {
			return nil, errWithCode
		}
	}

	// Open multipart file reader.
	mpfile, err := form.File.Open()
	if err != nil {
		err := gtserror.Newf("error opening multipart file: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Wrap multipart file reader to ensure is limited to max size.
	rc, _, _ := iotools.UpdateReadCloserLimit(mpfile, maxszInt64)

	// Create local media and write to instance storage.
	attachment, errWithCode := p.c.StoreLocalMedia(ctx,
		account.ID,
		func(ctx context.Context) (reader io.ReadCloser, err error) {
			return rc, nil
		},
		media.AdditionalMediaInfo{
			Description: &form.Description,
			FocusX:      &focusX,
			FocusY:      &focusY,
		},
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	apiAttachment, err := p.converter.AttachmentToAPIAttachment(ctx, attachment)
	if err != nil {
		err := fmt.Errorf("error parsing media attachment to frontend type: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return &apiAttachment, nil
}
