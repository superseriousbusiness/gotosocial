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

package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

func (p *processor) Create(ctx context.Context, account *gtsmodel.Account, form *apimodel.AttachmentRequest) (*apimodel.Attachment, error) {
	// open the attachment and extract the bytes from it
	f, err := form.File.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening attachment: %s", err)
	}
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		return nil, fmt.Errorf("error reading attachment: %s", err)
	}
	if size == 0 {
		return nil, errors.New("could not read provided attachment: size 0 bytes")
	}

	// now parse the focus parameter
	focusx, focusy, err := parseFocus(form.Focus)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse attachment focus: %s", err)
	}

	minAttachment := &gtsmodel.MediaAttachment{
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		AccountID:   account.ID,
		Description: text.RemoveHTML(form.Description),
		FileMeta: gtsmodel.FileMeta{
			Focus: gtsmodel.Focus{
				X: focusx,
				Y: focusy,
			},
		},
	}

	// allow the mediaHandler to work its magic of processing the attachment bytes, and putting them in whatever storage backend we're using
	attachment, err := p.mediaHandler.ProcessAttachment(ctx, buf.Bytes(), minAttachment)
	if err != nil {
		return nil, fmt.Errorf("error reading attachment: %s", err)
	}

	// prepare the frontend representation now -- if there are any errors here at least we can bail without
	// having already put something in the database and then having to clean it up again (eugh)
	apiAttachment, err := p.tc.AttachmentToAPIAttachment(ctx, attachment)
	if err != nil {
		return nil, fmt.Errorf("error parsing media attachment to frontend type: %s", err)
	}

	// now we can confidently put the attachment in the database
	if err := p.db.Put(ctx, attachment); err != nil {
		return nil, fmt.Errorf("error storing media attachment in db: %s", err)
	}

	return &apiAttachment, nil
}
