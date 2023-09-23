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

package account

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

func (p *Processor) selectNoteFormatter(contentType string) text.FormatFunc {
	if contentType == "text/markdown" {
		return p.formatter.FromMarkdown
	}

	return p.formatter.FromPlain
}

// Update processes the update of an account with the given form.
func (p *Processor) Update(ctx context.Context, account *gtsmodel.Account, form *apimodel.UpdateCredentialsRequest) (*apimodel.Account, gtserror.WithCode) {
	if form.Discoverable != nil {
		account.Discoverable = form.Discoverable
	}

	if form.Bot != nil {
		account.Bot = form.Bot
	}

	// Via the process of updating the account,
	// it is possible that the emojis used by
	// that account in note/display name/fields
	// may change; we need to keep track of this.
	var emojisChanged bool

	if form.DisplayName != nil {
		displayName := *form.DisplayName
		if err := validate.DisplayName(displayName); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		// Parse new display name (always from plaintext).
		account.DisplayName = text.SanitizeToPlaintext(displayName)

		// If display name has changed, account emojis may have also changed.
		emojisChanged = true
	}

	if form.Note != nil {
		note := *form.Note
		if err := validate.Note(note); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		// Store raw version of the note for now,
		// we'll process the proper version later.
		account.NoteRaw = note

		// If note has changed, account emojis may have also changed.
		emojisChanged = true
	}

	if form.FieldsAttributes != nil {
		var (
			fieldsAttributes = *form.FieldsAttributes
			fieldsLen        = len(fieldsAttributes)
			fieldsRaw        = make([]*gtsmodel.Field, 0, fieldsLen)
		)

		for _, updateField := range fieldsAttributes {
			if updateField.Name == nil || updateField.Value == nil {
				continue
			}

			var (
				name  string = *updateField.Name
				value string = *updateField.Value
			)

			if name == "" || value == "" {
				continue
			}

			// Sanitize raw field values.
			fieldRaw := &gtsmodel.Field{
				Name:  text.SanitizeToPlaintext(name),
				Value: text.SanitizeToPlaintext(value),
			}
			fieldsRaw = append(fieldsRaw, fieldRaw)
		}

		// Check length of parsed raw fields.
		if err := validate.ProfileFields(fieldsRaw); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		// OK, new raw fields are valid.
		account.FieldsRaw = fieldsRaw
		account.Fields = make([]*gtsmodel.Field, 0, fieldsLen) // process these in a sec

		// If fields have changed, account emojis may also have changed.
		emojisChanged = true
	}

	if emojisChanged {
		// Use map to deduplicate emojis by their ID.
		emojis := make(map[string]*gtsmodel.Emoji)

		// Retrieve display name emojis.
		for _, emoji := range p.formatter.FromPlainEmojiOnly(
			ctx,
			p.parseMention,
			account.ID,
			"",
			account.DisplayName,
		).Emojis {
			emojis[emoji.ID] = emoji
		}

		// Format + set note according to user prefs.
		f := p.selectNoteFormatter(account.StatusContentType)
		formatNoteResult := f(ctx, p.parseMention, account.ID, "", account.NoteRaw)
		account.Note = formatNoteResult.HTML

		// Retrieve note emojis.
		for _, emoji := range formatNoteResult.Emojis {
			emojis[emoji.ID] = emoji
		}

		// Process the raw fields we stored earlier.
		account.Fields = make([]*gtsmodel.Field, 0, len(account.FieldsRaw))
		for _, fieldRaw := range account.FieldsRaw {
			field := &gtsmodel.Field{}

			// Name stays plain, but we still need to
			// see if there are any emojis set in it.
			field.Name = fieldRaw.Name
			for _, emoji := range p.formatter.FromPlainEmojiOnly(
				ctx,
				p.parseMention,
				account.ID,
				"",
				fieldRaw.Name,
			).Emojis {
				emojis[emoji.ID] = emoji
			}

			// Value can be HTML, but we don't want
			// to wrap the result in <p> tags.
			fieldFormatValueResult := p.formatter.FromPlainNoParagraph(ctx, p.parseMention, account.ID, "", fieldRaw.Value)
			field.Value = fieldFormatValueResult.HTML

			// Retrieve field emojis.
			for _, emoji := range fieldFormatValueResult.Emojis {
				emojis[emoji.ID] = emoji
			}

			// We're done, append the shiny new field.
			account.Fields = append(account.Fields, field)
		}

		emojisCount := len(emojis)
		account.Emojis = make([]*gtsmodel.Emoji, 0, emojisCount)
		account.EmojiIDs = make([]string, 0, emojisCount)

		for id, emoji := range emojis {
			account.Emojis = append(account.Emojis, emoji)
			account.EmojiIDs = append(account.EmojiIDs, id)
		}
	}

	if form.Avatar != nil && form.Avatar.Size != 0 {
		avatarInfo, err := p.UpdateAvatar(ctx, form.Avatar, nil, account.ID)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}
		account.AvatarMediaAttachmentID = avatarInfo.ID
		account.AvatarMediaAttachment = avatarInfo
		log.Tracef(ctx, "new avatar info for account %s is %+v", account.ID, avatarInfo)
	}

	if form.Header != nil && form.Header.Size != 0 {
		headerInfo, err := p.UpdateHeader(ctx, form.Header, nil, account.ID)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}
		account.HeaderMediaAttachmentID = headerInfo.ID
		account.HeaderMediaAttachment = headerInfo
		log.Tracef(ctx, "new header info for account %s is %+v", account.ID, headerInfo)
	}

	if form.Locked != nil {
		account.Locked = form.Locked
	}

	if form.Source != nil {
		if form.Source.Language != nil {
			language, err := validate.Language(*form.Source.Language)
			if err != nil {
				return nil, gtserror.NewErrorBadRequest(err)
			}
			account.Language = language
		}

		if form.Source.Sensitive != nil {
			account.Sensitive = form.Source.Sensitive
		}

		if form.Source.Privacy != nil {
			if err := validate.Privacy(*form.Source.Privacy); err != nil {
				return nil, gtserror.NewErrorBadRequest(err)
			}
			privacy := typeutils.APIVisToVis(apimodel.Visibility(*form.Source.Privacy))
			account.Privacy = privacy
		}

		if form.Source.StatusContentType != nil {
			if err := validate.StatusContentType(*form.Source.StatusContentType); err != nil {
				return nil, gtserror.NewErrorBadRequest(err, err.Error())
			}

			account.StatusContentType = *form.Source.StatusContentType
		}
	}

	if form.CustomCSS != nil {
		customCSS := *form.CustomCSS
		if err := validate.CustomCSS(customCSS); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		account.CustomCSS = text.SanitizeToPlaintext(customCSS)
	}

	if form.EnableRSS != nil {
		account.EnableRSS = form.EnableRSS
	}

	err := p.state.DB.UpdateAccount(ctx, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("could not update account %s: %s", account.ID, err))
	}

	p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ObjectProfile,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       account,
		OriginAccount:  account,
	})

	acctSensitive, err := p.converter.AccountToAPIAccountSensitive(ctx, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("could not convert account into apisensitive account: %s", err))
	}
	return acctSensitive, nil
}

// UpdateAvatar does the dirty work of checking the avatar part of an account update form,
// parsing and checking the image, and doing the necessary updates in the database for this to become
// the account's new avatar image.
func (p *Processor) UpdateAvatar(ctx context.Context, avatar *multipart.FileHeader, description *string, accountID string) (*gtsmodel.MediaAttachment, error) {
	maxImageSize := config.GetMediaImageMaxSize()
	if avatar.Size > int64(maxImageSize) {
		return nil, fmt.Errorf("UpdateAvatar: avatar with size %d exceeded max image size of %d bytes", avatar.Size, maxImageSize)
	}

	dataFunc := func(innerCtx context.Context) (io.ReadCloser, int64, error) {
		f, err := avatar.Open()
		return f, avatar.Size, err
	}

	isAvatar := true
	ai := &media.AdditionalMediaInfo{
		Avatar:      &isAvatar,
		Description: description,
	}

	processingMedia, err := p.mediaManager.PreProcessMedia(ctx, dataFunc, accountID, ai)
	if err != nil {
		return nil, fmt.Errorf("UpdateAvatar: error processing avatar: %s", err)
	}

	return processingMedia.LoadAttachment(ctx)
}

// UpdateHeader does the dirty work of checking the header part of an account update form,
// parsing and checking the image, and doing the necessary updates in the database for this to become
// the account's new header image.
func (p *Processor) UpdateHeader(ctx context.Context, header *multipart.FileHeader, description *string, accountID string) (*gtsmodel.MediaAttachment, error) {
	maxImageSize := config.GetMediaImageMaxSize()
	if header.Size > int64(maxImageSize) {
		return nil, fmt.Errorf("UpdateHeader: header with size %d exceeded max image size of %d bytes", header.Size, maxImageSize)
	}

	dataFunc := func(innerCtx context.Context) (io.ReadCloser, int64, error) {
		f, err := header.Open()
		return f, header.Size, err
	}

	isHeader := true
	ai := &media.AdditionalMediaInfo{
		Header: &isHeader,
	}

	processingMedia, err := p.mediaManager.PreProcessMedia(ctx, dataFunc, accountID, ai)
	if err != nil {
		return nil, fmt.Errorf("UpdateHeader: error processing header: %s", err)
	}

	return processingMedia.LoadAttachment(ctx)
}
