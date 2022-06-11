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

package account

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/sirupsen/logrus"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

func (p *processor) Update(ctx context.Context, account *gtsmodel.Account, form *apimodel.UpdateCredentialsRequest) (*apimodel.Account, gtserror.WithCode) {
	l := logrus.WithField("func", "AccountUpdate")

	if form.Discoverable != nil {
		account.Discoverable = *form.Discoverable
	}

	if form.Bot != nil {
		account.Bot = *form.Bot
	}

	if form.DisplayName != nil {
		if err := validate.DisplayName(*form.DisplayName); err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}
		account.DisplayName = text.SanitizePlaintext(*form.DisplayName)
	}

	if form.Note != nil {
		if err := validate.Note(*form.Note); err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}

		// Set the raw note before processing
		account.NoteRaw = *form.Note

		// Process note to generate a valid HTML representation
		note, err := p.processNote(ctx, *form.Note, account.ID)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}

		// Set updated HTML-ified note
		account.Note = note
	}

	if form.Avatar != nil && form.Avatar.Size != 0 {
		avatarInfo, err := p.UpdateAvatar(ctx, form.Avatar, account.ID)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}
		account.AvatarMediaAttachmentID = avatarInfo.ID
		account.AvatarMediaAttachment = avatarInfo
		l.Tracef("new avatar info for account %s is %+v", account.ID, avatarInfo)
	}

	if form.Header != nil && form.Header.Size != 0 {
		headerInfo, err := p.UpdateHeader(ctx, form.Header, account.ID)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}
		account.HeaderMediaAttachmentID = headerInfo.ID
		account.HeaderMediaAttachment = headerInfo
		l.Tracef("new header info for account %s is %+v", account.ID, headerInfo)
	}

	if form.Locked != nil {
		account.Locked = *form.Locked
	}

	if form.Source != nil {
		if form.Source.Language != nil {
			if err := validate.Language(*form.Source.Language); err != nil {
				return nil, gtserror.NewErrorBadRequest(err)
			}
			account.Language = *form.Source.Language
		}

		if form.Source.Sensitive != nil {
			account.Sensitive = *form.Source.Sensitive
		}

		if form.Source.Privacy != nil {
			if err := validate.Privacy(*form.Source.Privacy); err != nil {
				return nil, gtserror.NewErrorBadRequest(err)
			}
			privacy := p.tc.APIVisToVis(apimodel.Visibility(*form.Source.Privacy))
			account.Privacy = privacy
		}
	}

	updatedAccount, err := p.db.UpdateAccount(ctx, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("could not update account %s: %s", account.ID, err))
	}

	p.clientWorker.Queue(messages.FromClientAPI{
		APObjectType:   ap.ObjectProfile,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       updatedAccount,
		OriginAccount:  updatedAccount,
	})

	acctSensitive, err := p.tc.AccountToAPIAccountSensitive(ctx, updatedAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("could not convert account into apisensitive account: %s", err))
	}
	return acctSensitive, nil
}

// UpdateAvatar does the dirty work of checking the avatar part of an account update form,
// parsing and checking the image, and doing the necessary updates in the database for this to become
// the account's new avatar image.
func (p *processor) UpdateAvatar(ctx context.Context, avatar *multipart.FileHeader, accountID string) (*gtsmodel.MediaAttachment, error) {
	maxImageSize := config.GetMediaImageMaxSize()
	if int(avatar.Size) > maxImageSize {
		return nil, fmt.Errorf("UpdateAvatar: avatar with size %d exceeded max image size of %d bytes", avatar.Size, maxImageSize)
	}

	dataFunc := func(innerCtx context.Context) (io.Reader, int, error) {
		f, err := avatar.Open()
		return f, int(avatar.Size), err
	}

	isAvatar := true
	ai := &media.AdditionalMediaInfo{
		Avatar: &isAvatar,
	}

	processingMedia, err := p.mediaManager.ProcessMedia(ctx, dataFunc, nil, accountID, ai)
	if err != nil {
		return nil, fmt.Errorf("UpdateAvatar: error processing avatar: %s", err)
	}

	return processingMedia.LoadAttachment(ctx)
}

// UpdateHeader does the dirty work of checking the header part of an account update form,
// parsing and checking the image, and doing the necessary updates in the database for this to become
// the account's new header image.
func (p *processor) UpdateHeader(ctx context.Context, header *multipart.FileHeader, accountID string) (*gtsmodel.MediaAttachment, error) {
	maxImageSize := config.GetMediaImageMaxSize()
	if int(header.Size) > maxImageSize {
		return nil, fmt.Errorf("UpdateHeader: header with size %d exceeded max image size of %d bytes", header.Size, maxImageSize)
	}

	dataFunc := func(innerCtx context.Context) (io.Reader, int, error) {
		f, err := header.Open()
		return f, int(header.Size), err
	}

	isHeader := true
	ai := &media.AdditionalMediaInfo{
		Header: &isHeader,
	}

	processingMedia, err := p.mediaManager.ProcessMedia(ctx, dataFunc, nil, accountID, ai)
	if err != nil {
		return nil, fmt.Errorf("UpdateHeader: error processing header: %s", err)
	}
	if err != nil {
		return nil, fmt.Errorf("UpdateHeader: error processing header: %s", err)
	}

	return processingMedia.LoadAttachment(ctx)
}

func (p *processor) processNote(ctx context.Context, note string, accountID string) (string, error) {
	if note == "" {
		return "", nil
	}

	tagStrings := util.DeriveHashtagsFromText(note)
	tags, err := p.db.TagStringsToTags(ctx, tagStrings, accountID)
	if err != nil {
		return "", err
	}

	mentionStrings := util.DeriveMentionNamesFromText(note)
	mentions := []*gtsmodel.Mention{}
	for _, mentionString := range mentionStrings {
		mention, err := p.parseMention(ctx, mentionString, accountID, "")
		if err != nil {
			continue
		}
		mentions = append(mentions, mention)
	}

	// TODO: support emojis in account notes
	// emojiStrings := util.DeriveEmojisFromText(note)
	// emojis, err := p.db.EmojiStringsToEmojis(ctx, emojiStrings)

	return p.formatter.FromPlain(ctx, note, mentions, tags), nil
}
