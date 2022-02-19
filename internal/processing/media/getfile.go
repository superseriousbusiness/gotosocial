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

package media

import (
	"context"
	"fmt"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
)

func (p *processor) GetFile(ctx context.Context, account *gtsmodel.Account, form *apimodel.GetContentRequestForm) (*apimodel.Content, error) {
	// parse the form fields
	mediaSize, err := media.ParseMediaSize(form.MediaSize)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("media size %s not valid", form.MediaSize))
	}

	mediaType, err := media.ParseMediaType(form.MediaType)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("media type %s not valid", form.MediaType))
	}

	spl := strings.Split(form.FileName, ".")
	if len(spl) != 2 || spl[0] == "" || spl[1] == "" {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("file name %s not parseable", form.FileName))
	}
	wantedMediaID := spl[0]

	// get the account that owns the media and make sure it's not suspended
	acct, err := p.db.GetAccountByID(ctx, form.AccountID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("account with id %s could not be selected from the db: %s", form.AccountID, err))
	}
	if !acct.SuspendedAt.IsZero() {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("account with id %s is suspended", form.AccountID))
	}

	// make sure the requesting account and the media account don't block each other
	if account != nil {
		blocked, err := p.db.IsBlocked(ctx, account.ID, form.AccountID, true)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("block status could not be established between accounts %s and %s: %s", form.AccountID, account.ID, err))
		}
		if blocked {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("block exists between accounts %s and %s", form.AccountID, account.ID))
		}
	}

	// the way we store emojis is a little different from the way we store other attachments,
	// so we need to take different steps depending on the media type being requested
	content := &apimodel.Content{}
	var storagePath string
	switch mediaType {
	case media.TypeEmoji:
		e := &gtsmodel.Emoji{}
		if err := p.db.GetByID(ctx, wantedMediaID, e); err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("emoji %s could not be taken from the db: %s", wantedMediaID, err))
		}
		if e.Disabled {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("emoji %s has been disabled", wantedMediaID))
		}
		switch mediaSize {
		case media.SizeOriginal:
			content.ContentType = e.ImageContentType
			content.ContentLength = int64(e.ImageFileSize)
			storagePath = e.ImagePath
		case media.SizeStatic:
			content.ContentType = e.ImageStaticContentType
			content.ContentLength = int64(e.ImageStaticFileSize)
			storagePath = e.ImageStaticPath
		default:
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("media size %s not recognized for emoji", mediaSize))
		}
	case media.TypeAttachment, media.TypeHeader, media.TypeAvatar:
		a, err := p.db.GetAttachmentByID(ctx, wantedMediaID)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("attachment %s could not be taken from the db: %s", wantedMediaID, err))
		}
		if a.AccountID != form.AccountID {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("attachment %s is not owned by %s", wantedMediaID, form.AccountID))
		}
		switch mediaSize {
		case media.SizeOriginal:
			content.ContentType = a.File.ContentType
			content.ContentLength = int64(a.File.FileSize)
			storagePath = a.File.Path
		case media.SizeSmall:
			content.ContentType = a.Thumbnail.ContentType
			content.ContentLength = int64(a.Thumbnail.FileSize)
			storagePath = a.Thumbnail.Path
		default:
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("media size %s not recognized for attachment", mediaSize))
		}
	}

	reader, err := p.storage.GetStream(storagePath)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error retrieving from storage: %s", err))
	}

	content.Content = reader
	return content, nil
}
