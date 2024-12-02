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

	"codeberg.org/gruf/go-iotools"
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

	var (
		// Indicates that the account's
		// note, display name, and/or fields
		// have changed, and so emojis should
		// be re-parsed and updated as well.
		textChanged bool

		// DB columns on the account
		// that need to be updated.
		acctColumns []string

		// DB columns on the settings
		// that need to be updated.
		settingsColumns []string
	)

	// Account flags.

	if form.Discoverable != nil {
		account.Discoverable = form.Discoverable
		acctColumns = append(acctColumns, "discoverable")
	}

	if form.Bot != nil {
		account.Bot = form.Bot
		acctColumns = append(acctColumns, "bot")
	}

	if form.Locked != nil {
		account.Locked = form.Locked
		acctColumns = append(acctColumns, "locked")
	}

	if form.DisplayName != nil {
		// Display name text
		// is changing.
		textChanged = true

		displayName := *form.DisplayName
		if err := validate.DisplayName(displayName); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		// Parse new display name (always from plaintext).
		account.DisplayName = text.SanitizeToPlaintext(displayName)
		acctColumns = append(acctColumns, "display_name")
	}

	if form.Note != nil {
		// Note text is changing.
		textChanged = true

		note := *form.Note
		if err := validate.Note(note); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		// Store raw version of note
		// for now, we'll process
		// the proper version later.
		account.NoteRaw = note
		acctColumns = append(acctColumns, []string{
			"note",
			"note_raw",
		}...)
	}

	if form.FieldsAttributes != nil {
		// Field text is changing.
		textChanged = true

		if err := p.updateFields(
			account,
			*form.FieldsAttributes,
		); err != nil {
			return nil, err
		}
		acctColumns = append(acctColumns, []string{
			"fields",
			"fields_raw",
		}...)
	}

	if textChanged {
		// Process display name, note, fields,
		// and any concomitant emoji changes.
		p.processAccountText(ctx, account)
		acctColumns = append(acctColumns, "emojis")
	}

	if form.AvatarDescription != nil {
		desc := text.SanitizeToPlaintext(*form.AvatarDescription)
		form.AvatarDescription = &desc
	}

	if form.Avatar != nil && form.Avatar.Size != 0 {
		avatarInfo, errWithCode := p.UpdateAvatar(ctx,
			account,
			form.Avatar,
			form.AvatarDescription,
		)
		if errWithCode != nil {
			return nil, errWithCode
		}
		account.AvatarMediaAttachmentID = avatarInfo.ID
		account.AvatarMediaAttachment = avatarInfo
		acctColumns = append(acctColumns, "avatar_media_attachment_id")
	} else if form.AvatarDescription != nil && account.AvatarMediaAttachment != nil {
		// Update just existing description if possible.
		account.AvatarMediaAttachment.Description = *form.AvatarDescription
		if err := p.state.DB.UpdateAttachment(
			ctx,
			account.AvatarMediaAttachment,
			"description",
		); err != nil {
			err := gtserror.Newf("db error updating account avatar description: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	if form.HeaderDescription != nil {
		desc := text.SanitizeToPlaintext(*form.HeaderDescription)
		form.HeaderDescription = util.Ptr(desc)
	}

	if form.Header != nil && form.Header.Size != 0 {
		headerInfo, errWithCode := p.UpdateHeader(ctx,
			account,
			form.Header,
			form.HeaderDescription,
		)
		if errWithCode != nil {
			return nil, errWithCode
		}
		account.HeaderMediaAttachmentID = headerInfo.ID
		account.HeaderMediaAttachment = headerInfo
		acctColumns = append(acctColumns, "header_media_attachment_id")
	} else if form.HeaderDescription != nil && account.HeaderMediaAttachment != nil {
		// Update just existing description if possible.
		account.HeaderMediaAttachment.Description = *form.HeaderDescription
		if err := p.state.DB.UpdateAttachment(
			ctx,
			account.HeaderMediaAttachment,
			"description",
		); err != nil {
			err := gtserror.Newf("db error updating account avatar description: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	// Account settings flags.

	if form.Source != nil {
		if form.Source.Language != nil {
			language, err := validate.Language(*form.Source.Language)
			if err != nil {
				return nil, gtserror.NewErrorBadRequest(err, err.Error())
			}

			account.Settings.Language = language
			settingsColumns = append(settingsColumns, "language")
		}

		if form.Source.Sensitive != nil {
			account.Settings.Sensitive = form.Source.Sensitive
			settingsColumns = append(settingsColumns, "sensitive")
		}

		if form.Source.Privacy != nil {
			if err := validate.Privacy(*form.Source.Privacy); err != nil {
				return nil, gtserror.NewErrorBadRequest(err, err.Error())
			}

			priv := apimodel.Visibility(*form.Source.Privacy)
			account.Settings.Privacy = typeutils.APIVisToVis(priv)
			settingsColumns = append(settingsColumns, "privacy")
		}

		if form.Source.StatusContentType != nil {
			if err := validate.StatusContentType(*form.Source.StatusContentType); err != nil {
				return nil, gtserror.NewErrorBadRequest(err, err.Error())
			}

			account.Settings.StatusContentType = *form.Source.StatusContentType
			settingsColumns = append(settingsColumns, "status_content_type")
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
		settingsColumns = append(settingsColumns, "theme")
	}

	if form.CustomCSS != nil {
		customCSS := *form.CustomCSS
		if err := validate.CustomCSS(customCSS); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		account.Settings.CustomCSS = text.SanitizeToPlaintext(customCSS)
		settingsColumns = append(settingsColumns, "custom_css")
	}

	if form.EnableRSS != nil {
		account.Settings.EnableRSS = form.EnableRSS
		settingsColumns = append(settingsColumns, "enable_rss")
	}

	if form.HideCollections != nil {
		account.Settings.HideCollections = form.HideCollections
		settingsColumns = append(settingsColumns, "hide_collections")
	}

	if form.WebVisibility != nil {
		apiVis := apimodel.Visibility(*form.WebVisibility)
		webVisibility := typeutils.APIVisToVis(apiVis)
		if webVisibility != gtsmodel.VisibilityPublic &&
			webVisibility != gtsmodel.VisibilityUnlocked &&
			webVisibility != gtsmodel.VisibilityNone {
			const text = "web_visibility must be one of public, unlocked, or none"
			err := errors.New(text)
			return nil, gtserror.NewErrorBadRequest(err, text)
		}

		account.Settings.WebVisibility = webVisibility
		settingsColumns = append(settingsColumns, "web_visibility")
	}

	// We've parsed + set everything, do
	// necessary database updates now.

	if len(acctColumns) > 0 {
		if err := p.state.DB.UpdateAccount(ctx, account, acctColumns...); err != nil {
			err := gtserror.Newf("db error updating account %s: %w", account.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	if len(settingsColumns) > 0 {
		if err := p.state.DB.UpdateAccountSettings(ctx, account.Settings, settingsColumns...); err != nil {
			err := gtserror.Newf("db error updating account settings %s: %w", account.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	// Send out Update message over the s2s (fedi) API.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActorPerson,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       account,
		Origin:         account,
	})

	acctSensitive, err := p.converter.AccountToAPIAccountSensitive(ctx, account)
	if err != nil {
		err := gtserror.Newf("error converting account: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return acctSensitive, nil
}

// updateFields sets FieldsRaw on the given
// account, and resets account.Fields to an
// empty slice, ready for further processing.
func (p *Processor) updateFields(
	account *gtsmodel.Account,
	fieldsAttributes []apimodel.UpdateField,
) gtserror.WithCode {
	var (
		fieldsLen = len(fieldsAttributes)
		fieldsRaw = make([]*gtsmodel.Field, 0, fieldsLen)
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
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	// OK, new raw fields are valid.
	account.FieldsRaw = fieldsRaw
	account.Fields = make([]*gtsmodel.Field, 0, fieldsLen)
	return nil
}

// processAccountText processes the raw versions of the given
// account's display name, note, and fields, and sets those
// processed versions on the account, while also updating the
// account's emojis entry based on the results of the processing.
func (p *Processor) processAccountText(
	ctx context.Context,
	account *gtsmodel.Account,
) {
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

	// Process raw fields.
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

	// Update the account's emojis.
	emojisCount := len(emojis)
	account.Emojis = make([]*gtsmodel.Emoji, 0, emojisCount)
	account.EmojiIDs = make([]string, 0, emojisCount)

	for id, emoji := range emojis {
		account.Emojis = append(account.Emojis, emoji)
		account.EmojiIDs = append(account.EmojiIDs, id)
	}
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
	// Get maximum supported local media size.
	maxsz := config.GetMediaLocalMaxSize()
	maxszInt64 := int64(maxsz) // #nosec G115 -- Already validated.

	// Ensure media within size bounds.
	if avatar.Size > maxszInt64 {
		text := fmt.Sprintf("media exceeds configured max size: %s", maxsz)
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Open multipart file reader.
	mpfile, err := avatar.Open()
	if err != nil {
		err := gtserror.Newf("error opening multipart file: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Wrap the multipart file reader to ensure is limited to max.
	rc, _, _ := iotools.UpdateReadCloserLimit(mpfile, maxszInt64)

	// Write to instance storage.
	return p.c.StoreLocalMedia(ctx,
		account.ID,
		func(ctx context.Context) (reader io.ReadCloser, err error) {
			return rc, nil
		},
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
	// Get maximum supported local media size.
	maxsz := config.GetMediaLocalMaxSize()
	maxszInt64 := int64(maxsz) // #nosec G115 -- Already validated.

	// Ensure media within size bounds.
	if header.Size > maxszInt64 {
		text := fmt.Sprintf("media exceeds configured max size: %s", maxsz)
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Open multipart file reader.
	mpfile, err := header.Open()
	if err != nil {
		err := gtserror.Newf("error opening multipart file: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Wrap the multipart file reader to ensure is limited to max.
	rc, _, _ := iotools.UpdateReadCloserLimit(mpfile, maxszInt64)

	// Write to instance storage.
	return p.c.StoreLocalMedia(ctx,
		account.ID,
		func(ctx context.Context) (reader io.ReadCloser, err error) {
			return rc, nil
		},
		media.AdditionalMediaInfo{
			Header:      util.Ptr(true),
			Description: description,
		},
	)
}
