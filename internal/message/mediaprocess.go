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

package message

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) MediaCreate(authed *oauth.Auth, form *apimodel.AttachmentRequest) (*apimodel.Attachment, error) {
	// First check this user/account is permitted to create media
	// There's no point continuing otherwise.
	//
	// TODO: move this check to the oauth.Authed function and do it for all accounts
	if authed.User.Disabled || !authed.User.Approved || !authed.Account.SuspendedAt.IsZero() {
		return nil, errors.New("not authorized to post new media")
	}

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

	// allow the mediaHandler to work its magic of processing the attachment bytes, and putting them in whatever storage backend we're using
	attachment, err := p.mediaHandler.ProcessAttachment(buf.Bytes(), authed.Account.ID, "")
	if err != nil {
		return nil, fmt.Errorf("error reading attachment: %s", err)
	}

	// now we need to add extra fields that the attachment processor doesn't know (from the form)
	// TODO: handle this inside mediaHandler.ProcessAttachment (just pass more params to it)

	// first description
	attachment.Description = form.Description

	// now parse the focus parameter
	focusx, focusy, err := parseFocus(form.Focus)
	if err != nil {
		return nil, err
	}
	attachment.FileMeta.Focus.X = focusx
	attachment.FileMeta.Focus.Y = focusy

	// prepare the frontend representation now -- if there are any errors here at least we can bail without
	// having already put something in the database and then having to clean it up again (eugh)
	mastoAttachment, err := p.tc.AttachmentToMasto(attachment)
	if err != nil {
		return nil, fmt.Errorf("error parsing media attachment to frontend type: %s", err)
	}

	// now we can confidently put the attachment in the database
	if err := p.db.Put(attachment); err != nil {
		return nil, fmt.Errorf("error storing media attachment in db: %s", err)
	}

	return &mastoAttachment, nil
}

func (p *processor) MediaGet(authed *oauth.Auth, mediaAttachmentID string) (*apimodel.Attachment, ErrorWithCode) {
	attachment := &gtsmodel.MediaAttachment{}
	if err := p.db.GetByID(mediaAttachmentID, attachment); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			// attachment doesn't exist
			return nil, NewErrorNotFound(errors.New("attachment doesn't exist in the db"))
		}
		return nil, NewErrorNotFound(fmt.Errorf("db error getting attachment: %s", err))
	}

	if attachment.AccountID != authed.Account.ID {
		return nil, NewErrorNotFound(errors.New("attachment not owned by requesting account"))
	}

	a, err := p.tc.AttachmentToMasto(attachment)
	if err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("error converting attachment: %s", err))
	}

	return &a, nil
}

func (p *processor) MediaUpdate(authed *oauth.Auth, mediaAttachmentID string, form *apimodel.AttachmentUpdateRequest) (*apimodel.Attachment, ErrorWithCode) {
	attachment := &gtsmodel.MediaAttachment{}
	if err := p.db.GetByID(mediaAttachmentID, attachment); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			// attachment doesn't exist
			return nil, NewErrorNotFound(errors.New("attachment doesn't exist in the db"))
		}
		return nil, NewErrorNotFound(fmt.Errorf("db error getting attachment: %s", err))
	}

	if attachment.AccountID != authed.Account.ID {
		return nil, NewErrorNotFound(errors.New("attachment not owned by requesting account"))
	}

	if form.Description != nil {
		attachment.Description = *form.Description
		if err := p.db.UpdateByID(mediaAttachmentID, attachment); err != nil {
			return nil, NewErrorInternalError(fmt.Errorf("database error updating description: %s", err))
		}
	}

	if form.Focus != nil {
		focusx, focusy, err := parseFocus(*form.Focus)
		if err != nil {
			return nil, NewErrorBadRequest(err)
		}
		attachment.FileMeta.Focus.X = focusx
		attachment.FileMeta.Focus.Y = focusy
		if err := p.db.UpdateByID(mediaAttachmentID, attachment); err != nil {
			return nil, NewErrorInternalError(fmt.Errorf("database error updating focus: %s", err))
		}
	}

	a, err := p.tc.AttachmentToMasto(attachment)
	if err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("error converting attachment: %s", err))
	}

	return &a, nil
}

func (p *processor) FileGet(authed *oauth.Auth, form *apimodel.GetContentRequestForm) (*apimodel.Content, error) {
	// parse the form fields
	mediaSize, err := media.ParseMediaSize(form.MediaSize)
	if err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("media size %s not valid", form.MediaSize))
	}

	mediaType, err := media.ParseMediaType(form.MediaType)
	if err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("media type %s not valid", form.MediaType))
	}

	spl := strings.Split(form.FileName, ".")
	if len(spl) != 2 || spl[0] == "" || spl[1] == "" {
		return nil, NewErrorNotFound(fmt.Errorf("file name %s not parseable", form.FileName))
	}
	wantedMediaID := spl[0]

	// get the account that owns the media and make sure it's not suspended
	acct := &gtsmodel.Account{}
	if err := p.db.GetByID(form.AccountID, acct); err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("account with id %s could not be selected from the db: %s", form.AccountID, err))
	}
	if !acct.SuspendedAt.IsZero() {
		return nil, NewErrorNotFound(fmt.Errorf("account with id %s is suspended", form.AccountID))
	}

	// make sure the requesting account and the media account don't block each other
	if authed.Account != nil {
		blocked, err := p.db.Blocked(authed.Account.ID, form.AccountID)
		if err != nil {
			return nil, NewErrorNotFound(fmt.Errorf("block status could not be established between accounts %s and %s: %s", form.AccountID, authed.Account.ID, err))
		}
		if blocked {
			return nil, NewErrorNotFound(fmt.Errorf("block exists between accounts %s and %s", form.AccountID, authed.Account.ID))
		}
	}

	// the way we store emojis is a little different from the way we store other attachments,
	// so we need to take different steps depending on the media type being requested
	content := &apimodel.Content{}
	var storagePath string
	switch mediaType {
	case media.Emoji:
		e := &gtsmodel.Emoji{}
		if err := p.db.GetByID(wantedMediaID, e); err != nil {
			return nil, NewErrorNotFound(fmt.Errorf("emoji %s could not be taken from the db: %s", wantedMediaID, err))
		}
		if e.Disabled {
			return nil, NewErrorNotFound(fmt.Errorf("emoji %s has been disabled", wantedMediaID))
		}
		switch mediaSize {
		case media.Original:
			content.ContentType = e.ImageContentType
			storagePath = e.ImagePath
		case media.Static:
			content.ContentType = e.ImageStaticContentType
			storagePath = e.ImageStaticPath
		default:
			return nil, NewErrorNotFound(fmt.Errorf("media size %s not recognized for emoji", mediaSize))
		}
	case media.Attachment, media.Header, media.Avatar:
		a := &gtsmodel.MediaAttachment{}
		if err := p.db.GetByID(wantedMediaID, a); err != nil {
			return nil, NewErrorNotFound(fmt.Errorf("attachment %s could not be taken from the db: %s", wantedMediaID, err))
		}
		if a.AccountID != form.AccountID {
			return nil, NewErrorNotFound(fmt.Errorf("attachment %s is not owned by %s", wantedMediaID, form.AccountID))
		}
		switch mediaSize {
		case media.Original:
			content.ContentType = a.File.ContentType
			storagePath = a.File.Path
		case media.Small:
			content.ContentType = a.Thumbnail.ContentType
			storagePath = a.Thumbnail.Path
		default:
			return nil, NewErrorNotFound(fmt.Errorf("media size %s not recognized for attachment", mediaSize))
		}
	}

	bytes, err := p.storage.RetrieveFileFrom(storagePath)
	if err != nil {
		return nil, NewErrorNotFound(fmt.Errorf("error retrieving from storage: %s", err))
	}

	content.ContentLength = int64(len(bytes))
	content.Content = bytes
	return content, nil
}

func parseFocus(focus string) (focusx, focusy float32, err error) {
	if focus == "" {
		return
	}
	spl := strings.Split(focus, ",")
	if len(spl) != 2 {
		err = fmt.Errorf("improperly formatted focus %s", focus)
		return
	}
	xStr := spl[0]
	yStr := spl[1]
	if xStr == "" || yStr == "" {
		err = fmt.Errorf("improperly formatted focus %s", focus)
		return
	}
	fx, err := strconv.ParseFloat(xStr, 32)
	if err != nil {
		err = fmt.Errorf("improperly formatted focus %s: %s", focus, err)
		return
	}
	if fx > 1 || fx < -1 {
		err = fmt.Errorf("improperly formatted focus %s", focus)
		return
	}
	focusx = float32(fx)
	fy, err := strconv.ParseFloat(yStr, 32)
	if err != nil {
		err = fmt.Errorf("improperly formatted focus %s: %s", focus, err)
		return
	}
	if fy > 1 || fy < -1 {
		err = fmt.Errorf("improperly formatted focus %s", focus)
		return
	}
	focusy = float32(fy)
	return
}
