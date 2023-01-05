/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

func (p *processor) Update(ctx context.Context, account *gtsmodel.Account, form *apimodel.UpdateCredentialsRequest) (*apimodel.Account, gtserror.WithCode) {
	if form.Discoverable != nil {
		account.Discoverable = form.Discoverable
	}

	if form.Bot != nil {
		account.Bot = form.Bot
	}

	var updateEmojis bool

	if form.DisplayName != nil {
		if err := validate.DisplayName(*form.DisplayName); err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}
		account.DisplayName = text.SanitizePlaintext(*form.DisplayName)
		updateEmojis = true
	}

	if form.Note != nil {
		if err := validate.Note(*form.Note); err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}

		// Set the raw note before processing
		account.NoteRaw = *form.Note

		// Process note to generate a valid HTML representation
		note, err := p.processNote(ctx, *form.Note, account)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}

		// Set updated HTML-ified note
		account.Note = note
		updateEmojis = true
	}

	if updateEmojis {
		// account emojis -- treat the sanitized display name and raw
		// note like one long text for the purposes of deriving emojis
		accountEmojiShortcodes := util.DeriveEmojisFromText(account.DisplayName + "\n\n" + account.NoteRaw)
		account.Emojis = make([]*gtsmodel.Emoji, 0, len(accountEmojiShortcodes))
		account.EmojiIDs = make([]string, 0, len(accountEmojiShortcodes))

		for _, shortcode := range accountEmojiShortcodes {
			emoji, err := p.db.GetEmojiByShortcodeDomain(ctx, shortcode, "")
			if err != nil {
				if err != db.ErrNoEntries {
					log.Errorf("error getting local emoji with shortcode %s: %s", shortcode, err)
				}
				continue
			}

			if *emoji.VisibleInPicker && !*emoji.Disabled {
				account.Emojis = append(account.Emojis, emoji)
				account.EmojiIDs = append(account.EmojiIDs, emoji.ID)
			}
		}
	}

	if form.Avatar != nil && form.Avatar.Size != 0 {
		avatarInfo, err := p.UpdateAvatar(ctx, form.Avatar, nil, account.ID)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}
		account.AvatarMediaAttachmentID = avatarInfo.ID
		account.AvatarMediaAttachment = avatarInfo
		log.Tracef("new avatar info for account %s is %+v", account.ID, avatarInfo)
	}

	if form.Header != nil && form.Header.Size != 0 {
		headerInfo, err := p.UpdateHeader(ctx, form.Header, nil, account.ID)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err)
		}
		account.HeaderMediaAttachmentID = headerInfo.ID
		account.HeaderMediaAttachment = headerInfo
		log.Tracef("new header info for account %s is %+v", account.ID, headerInfo)
	}

	if form.Locked != nil {
		account.Locked = form.Locked
	}

	if form.Source != nil {
		if form.Source.Language != nil {
			if err := validate.Language(*form.Source.Language); err != nil {
				return nil, gtserror.NewErrorBadRequest(err)
			}
			account.Language = *form.Source.Language
		}

		if form.Source.Sensitive != nil {
			account.Sensitive = form.Source.Sensitive
		}

		if form.Source.Privacy != nil {
			if err := validate.Privacy(*form.Source.Privacy); err != nil {
				return nil, gtserror.NewErrorBadRequest(err)
			}
			privacy := p.tc.APIVisToVis(apimodel.Visibility(*form.Source.Privacy))
			account.Privacy = privacy
		}

		if form.Source.StatusFormat != nil {
			if err := validate.StatusFormat(*form.Source.StatusFormat); err != nil {
				return nil, gtserror.NewErrorBadRequest(err, err.Error())
			}

			account.StatusFormat = *form.Source.StatusFormat
		}
	}

	if form.CustomCSS != nil {
		customCSS := *form.CustomCSS
		if err := validate.CustomCSS(customCSS); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		account.CustomCSS = text.SanitizePlaintext(customCSS)
	}

	if form.EnableRSS != nil {
		account.EnableRSS = form.EnableRSS
	}

	err := p.db.UpdateAccount(ctx, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("could not update account %s: %s", account.ID, err))
	}

	p.clientWorker.Queue(messages.FromClientAPI{
		APObjectType:   ap.ObjectProfile,
		APActivityType: ap.ActivityUpdate,
		GTSModel:       account,
		OriginAccount:  account,
	})

	acctSensitive, err := p.tc.AccountToAPIAccountSensitive(ctx, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("could not convert account into apisensitive account: %s", err))
	}
	return acctSensitive, nil
}

// UpdateAvatar does the dirty work of checking the avatar part of an account update form,
// parsing and checking the image, and doing the necessary updates in the database for this to become
// the account's new avatar image.
func (p *processor) UpdateAvatar(ctx context.Context, avatar *multipart.FileHeader, description *string, accountID string) (*gtsmodel.MediaAttachment, error) {
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

	processingMedia, err := p.mediaManager.ProcessMedia(ctx, dataFunc, nil, accountID, ai)
	if err != nil {
		return nil, fmt.Errorf("UpdateAvatar: error processing avatar: %s", err)
	}

	return processingMedia.LoadAttachment(ctx)
}

// UpdateHeader does the dirty work of checking the header part of an account update form,
// parsing and checking the image, and doing the necessary updates in the database for this to become
// the account's new header image.
func (p *processor) UpdateHeader(ctx context.Context, header *multipart.FileHeader, description *string, accountID string) (*gtsmodel.MediaAttachment, error) {
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

	processingMedia, err := p.mediaManager.ProcessMedia(ctx, dataFunc, nil, accountID, ai)
	if err != nil {
		return nil, fmt.Errorf("UpdateHeader: error processing header: %s", err)
	}
	if err != nil {
		return nil, fmt.Errorf("UpdateHeader: error processing header: %s", err)
	}

	return processingMedia.LoadAttachment(ctx)
}

func (p *processor) processNote(ctx context.Context, note string, account *gtsmodel.Account) (string, error) {
	if note == "" {
		return "", nil
	}

	tagStrings := util.DeriveHashtagsFromText(note)
	tags, err := p.db.TagStringsToTags(ctx, tagStrings, account.ID)
	if err != nil {
		return "", err
	}

	mentionStrings := util.DeriveMentionNamesFromText(note)
	mentions := []*gtsmodel.Mention{}
	for _, mentionString := range mentionStrings {
		mention, err := p.parseMention(ctx, mentionString, account.ID, "")
		if err != nil {
			continue
		}
		mentions = append(mentions, mention)
	}

	// TODO: support emojis in account notes
	// emojiStrings := util.DeriveEmojisFromText(note)
	// emojis, err := p.db.EmojiStringsToEmojis(ctx, emojiStrings)

	if account.StatusFormat == "markdown" {
		return p.formatter.FromMarkdown(ctx, note, mentions, tags, nil), nil
	}

	return p.formatter.FromPlain(ctx, note, mentions, tags), nil
}
