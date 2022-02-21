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

func (p *processor) GetFile(ctx context.Context, account *gtsmodel.Account, form *apimodel.GetContentRequestForm) (*apimodel.Content, gtserror.WithCode) {
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
	expectedAccountID := form.AccountID

	// get the account that owns the media and make sure it's not suspended
	acct, err := p.db.GetAccountByID(ctx, expectedAccountID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("account with id %s could not be selected from the db: %s", expectedAccountID, err))
	}
	if !acct.SuspendedAt.IsZero() {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("account with id %s is suspended", expectedAccountID))
	}

	// make sure the requesting account and the media account don't block each other
	if account != nil {
		blocked, err := p.db.IsBlocked(ctx, account.ID, expectedAccountID, true)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("block status could not be established between accounts %s and %s: %s", expectedAccountID, account.ID, err))
		}
		if blocked {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("block exists between accounts %s and %s", expectedAccountID, account.ID))
		}
	}

	// the way we store emojis is a little different from the way we store other attachments,
	// so we need to take different steps depending on the media type being requested
	switch mediaType {
	case media.TypeEmoji:
		return p.getEmojiContent(ctx, wantedMediaID, mediaSize)
	case media.TypeAttachment, media.TypeHeader, media.TypeAvatar:
		return p.getAttachmentContent(ctx, wantedMediaID, expectedAccountID, mediaSize)
	default:
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("media type %s not recognized", mediaType))
	}
}

func (p *processor) getAttachmentContent(ctx context.Context, wantedMediaID string, expectedAccountID string, mediaSize media.Size) (*apimodel.Content, gtserror.WithCode) {
	attachmentContent := &apimodel.Content{}
	var storagePath string

	a, err := p.db.GetAttachmentByID(ctx, wantedMediaID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("attachment %s could not be taken from the db: %s", wantedMediaID, err))
	}

	if a.AccountID != expectedAccountID {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("attachment %s is not owned by %s", wantedMediaID, expectedAccountID))
	}

	switch mediaSize {
	case media.SizeOriginal:
		attachmentContent.ContentType = a.File.ContentType
		attachmentContent.ContentLength = int64(a.File.FileSize)
		storagePath = a.File.Path
	case media.SizeSmall:
		attachmentContent.ContentType = a.Thumbnail.ContentType
		attachmentContent.ContentLength = int64(a.Thumbnail.FileSize)
		storagePath = a.Thumbnail.Path
	default:
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("media size %s not recognized for attachment", mediaSize))
	}

	reader, err := p.storage.GetStream(storagePath)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error retrieving from storage: %s", err))
	}

	attachmentContent.Content = reader
	return attachmentContent, nil
}

func (p *processor) getEmojiContent(ctx context.Context, wantedEmojiID string, emojiSize media.Size) (*apimodel.Content, gtserror.WithCode) {
	emojiContent := &apimodel.Content{}
	var storagePath string

	e := &gtsmodel.Emoji{}
	if err := p.db.GetByID(ctx, wantedEmojiID, e); err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("emoji %s could not be taken from the db: %s", wantedEmojiID, err))
	}

	if e.Disabled {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("emoji %s has been disabled", wantedEmojiID))
	}

	switch emojiSize {
	case media.SizeOriginal:
		emojiContent.ContentType = e.ImageContentType
		emojiContent.ContentLength = int64(e.ImageFileSize)
		storagePath = e.ImagePath
	case media.SizeStatic:
		emojiContent.ContentType = e.ImageStaticContentType
		emojiContent.ContentLength = int64(e.ImageStaticFileSize)
		storagePath = e.ImageStaticPath
	default:
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("media size %s not recognized for emoji", emojiSize))
	}

	reader, err := p.storage.GetStream(storagePath)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error retrieving from storage: %s", err))
	}

	emojiContent.Content = reader
	return emojiContent, nil
}
