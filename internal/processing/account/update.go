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
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"codeberg.org/gruf/go-bytesize"
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
	"github.com/superseriousbusiness/gotosocial/internal/util"
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
	// Ensure account populated; we'll need settings.
	if err := p.state.DB.PopulateAccount(ctx, account); err != nil {
		log.Errorf(ctx, "error(s) populating account, will continue: %s", err)
	}

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
		f := p.selectNoteFormatter(account.Settings.StatusContentType)
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
		avatarInfo, errWithCode := p.UpdateAvatar(ctx,
			account,
			form.Avatar,
			nil,
		)
		if errWithCode != nil {
			return nil, errWithCode
		}
		account.AvatarMediaAttachmentID = avatarInfo.ID
		account.AvatarMediaAttachment = avatarInfo
		log.Tracef(ctx, "new avatar info for account %s is %+v", account.ID, avatarInfo)
	}

	if form.Header != nil && form.Header.Size != 0 {
		headerInfo, errWithCode := p.UpdateHeader(ctx,
			account,
			form.Header,
			nil,
		)
		if errWithCode != nil {
			return nil, errWithCode
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
			account.Settings.Language = language
		}

		if form.Source.Sensitive != nil {
			account.Settings.Sensitive = form.Source.Sensitive
		}

		if form.Source.Privacy != nil {
			if err := validate.Privacy(*form.Source.Privacy); err != nil {
				return nil, gtserror.NewErrorBadRequest(err)
			}
			privacy := typeutils.APIVisToVis(apimodel.Visibility(*form.Source.Privacy))
			account.Settings.Privacy = privacy
		}

		if form.Source.StatusContentType != nil {
			if err := validate.StatusContentType(*form.Source.StatusContentType); err != nil {
				return nil, gtserror.NewErrorBadRequest(err, err.Error())
			}

			account.Settings.StatusContentType = *form.Source.StatusContentType
		}
	}

	if form.Theme != nil {
		theme := *form.Theme
		if theme == "" {
			// Empty is easy, just clear this.
			account.Settings.Theme = ""
		} else {
			// Theme was provided, check
			// against known available themes.
			if _, ok := p.themes.ByFileName[theme]; !ok {
				err := fmt.Errorf("theme %s not available on this instance, see /api/v1/accounts/themes for available themes", theme)
				return nil, gtserror.NewErrorBadRequest(err, err.Error())
			}
			account.Settings.Theme = theme
		}
	}

	if form.CustomCSS != nil {
		customCSS := *form.CustomCSS
		if err := validate.CustomCSS(customCSS); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		account.Settings.CustomCSS = text.SanitizeToPlaintext(customCSS)
	}

	if form.EnableRSS != nil {
		account.Settings.EnableRSS = form.EnableRSS
	}

	if form.HideCollections != nil {
		account.Settings.HideCollections = form.HideCollections
	}

	if err := p.state.DB.UpdateAccount(ctx, account); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("could not update account %s: %s", account.ID, err))
	}

	if err := p.state.DB.UpdateAccountSettings(ctx, account.Settings); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("could not update account settings %s: %s", account.ID, err))
	}

	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActorPerson,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       account,
		Origin:         account,
	})

	acctSensitive, err := p.converter.AccountToAPIAccountSensitive(ctx, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("could not convert account into apisensitive account: %s", err))
	}
	return acctSensitive, nil
}

// UpdateAvatar does the dirty work of checking the avatar
// part of an account update form, parsing and checking the
// media, and doing the necessary updates in the database
// for this to become the account's new avatar.
func (p *Processor) UpdateAvatar(
	ctx context.Context,
	account *gtsmodel.Account,
	avatar *multipart.FileHeader,
	description *string,
) (
	*gtsmodel.MediaAttachment,
	gtserror.WithCode,
) {
	max := config.GetMediaImageMaxSize()
	if sz := bytesize.Size(avatar.Size); sz > max {
		text := fmt.Sprintf("size %s exceeds max media size %s", sz, max)
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		f, err := avatar.Open()
		return f, avatar.Size, err
	}

	// Write to instance storage.
	return p.c.StoreLocalMedia(ctx,
		account.ID,
		data,
		media.AdditionalMediaInfo{
			Avatar:      util.Ptr(true),
			Description: description,
		},
	)
}

// UpdateHeader does the dirty work of checking the header
// part of an account update form, parsing and checking the
// media, and doing the necessary updates in the database
// for this to become the account's new header.
func (p *Processor) UpdateHeader(
	ctx context.Context,
	account *gtsmodel.Account,
	header *multipart.FileHeader,
	description *string,
) (
	*gtsmodel.MediaAttachment,
	gtserror.WithCode,
) {
	max := config.GetMediaImageMaxSize()
	if sz := bytesize.Size(header.Size); sz > max {
		text := fmt.Sprintf("size %s exceeds max media size %s", sz, max)
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		f, err := header.Open()
		return f, header.Size, err
	}

	// Write to instance storage.
	return p.c.StoreLocalMedia(ctx,
		account.ID,
		data,
		media.AdditionalMediaInfo{
			Header:      util.Ptr(true),
			Description: description,
		},
	)
}
