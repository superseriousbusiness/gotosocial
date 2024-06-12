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
	"io"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

// Create creates a new media attachment belonging to the given account, using the request form.
func (p *Processor) Create(ctx context.Context, account *gtsmodel.Account, form *apimodel.AttachmentRequest) (*apimodel.Attachment, gtserror.WithCode) {
	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		f, err := form.File.Open()
		return f, form.File.Size, err
	}

	focusX, focusY, err := parseFocus(form.Focus)
	if err != nil {
		err := fmt.Errorf("could not parse focus value %s: %s", form.Focus, err)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Create local media and write to instance storage.
	attachment, errWithCode := p.c.StoreLocalMedia(ctx,
		account.ID,
		data,
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
