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

package admin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (p *processor) EmojiUpdate(ctx context.Context, id string, form *apimodel.EmojiUpdateRequest) (*apimodel.AdminEmoji, gtserror.WithCode) {
	emoji, err := p.db.GetEmojiByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("EmojiUpdate: no emoji with id %s found in the db", id)
			return nil, gtserror.NewErrorNotFound(err)
		}
		err := fmt.Errorf("EmojiUpdate: db error: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	switch form.Type {
	case apimodel.EmojiUpdateCopy:
		return p.emojiUpdateCopy(ctx, emoji, form.Shortcode, form.CategoryName)
	case apimodel.EmojiUpdateDisable:
		return p.emojiUpdateDisable(ctx, emoji)
	case apimodel.EmojiUpdateModify:
		return p.emojiUpdateModify(ctx, emoji, form.Image, form.CategoryName)
	default:
		err := errors.New("unrecognized emoji action type")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}
}

// copy an emoji from remote to local
func (p *processor) emojiUpdateCopy(ctx context.Context, emoji *gtsmodel.Emoji, shortcode *string, categoryName *string) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if emoji.Domain == "" {
		err := fmt.Errorf("emojiUpdateCopy: emoji %s is not a remote emoji, cannot copy it to local", emoji.ID)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	if shortcode == nil {
		err := fmt.Errorf("emojiUpdateCopy: emoji %s could not be copied, no shortcode provided", emoji.ID)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	maybeExisting, err := p.db.GetEmojiByShortcodeDomain(ctx, *shortcode, "")
	if maybeExisting != nil {
		err := fmt.Errorf("emojiUpdateCopy: emoji %s could not be copied, emoji with shortcode %s already exists on this instance", emoji.ID, *shortcode)
		return nil, gtserror.NewErrorConflict(err, err.Error())
	}

	if err != nil && err != db.ErrNoEntries {
		err := fmt.Errorf("emojiUpdateCopy: emoji %s could not be copied, error checking existence of emoji with shortcode %s: %s", emoji.ID, *shortcode, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	newEmojiID, err := id.NewRandomULID()
	if err != nil {
		err := fmt.Errorf("emojiUpdateCopy: emoji %s could not be copied, error creating id for new emoji: %s", emoji.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	newEmojiURI := uris.GenerateURIForEmoji(newEmojiID)

	data := func(ctx context.Context) (reader io.ReadCloser, fileSize int64, err error) {
		rc, err := p.storage.GetStream(ctx, emoji.ImagePath)
		return rc, int64(emoji.ImageFileSize), err
	}

	var ai *media.AdditionalEmojiInfo
	if categoryName != nil {
		category, err := p.GetOrCreateEmojiCategory(ctx, *categoryName)
		if err != nil {
			err = fmt.Errorf("emojiUpdateCopy: error getting or creating category: %s", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		ai = &media.AdditionalEmojiInfo{
			CategoryID: &category.ID,
		}
	}

	processingEmoji, err := p.mediaManager.PreProcessEmoji(ctx, data, nil, *shortcode, newEmojiID, newEmojiURI, ai, false)
	if err != nil {
		err = fmt.Errorf("emojiUpdateCopy: error processing emoji %s: %s", emoji.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	newEmoji, err := processingEmoji.LoadEmoji(ctx)
	if err != nil {
		err = fmt.Errorf("emojiUpdateCopy: error loading processed emoji %s: %s", emoji.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	adminEmoji, err := p.tc.EmojiToAdminAPIEmoji(ctx, newEmoji)
	if err != nil {
		err = fmt.Errorf("emojiUpdateCopy: error converting updated emoji %s to admin emoji: %s", emoji.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return adminEmoji, nil
}

// disable a remote emoji
func (p *processor) emojiUpdateDisable(ctx context.Context, emoji *gtsmodel.Emoji) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if emoji.Domain == "" {
		err := fmt.Errorf("emojiUpdateDisable: emoji %s is not a remote emoji, cannot disable it via this endpoint", emoji.ID)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	emojiDisabled := true
	emoji.Disabled = &emojiDisabled
	updatedEmoji, err := p.db.UpdateEmoji(ctx, emoji, "updated_at", "disabled")
	if err != nil {
		err = fmt.Errorf("emojiUpdateDisable: error updating emoji %s: %s", emoji.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	adminEmoji, err := p.tc.EmojiToAdminAPIEmoji(ctx, updatedEmoji)
	if err != nil {
		err = fmt.Errorf("emojiUpdateDisable: error converting updated emoji %s to admin emoji: %s", emoji.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return adminEmoji, nil
}

// modify a local emoji
func (p *processor) emojiUpdateModify(ctx context.Context, emoji *gtsmodel.Emoji, image *multipart.FileHeader, categoryName *string) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if emoji.Domain != "" {
		err := fmt.Errorf("emojiUpdateModify: emoji %s is not a local emoji, cannot do a modify action on it", emoji.ID)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	var updatedEmoji *gtsmodel.Emoji

	// keep existing categoryID unless a new one is defined
	var (
		updatedCategoryID = emoji.CategoryID
		updateCategoryID  bool
	)
	if categoryName != nil {
		category, err := p.GetOrCreateEmojiCategory(ctx, *categoryName)
		if err != nil {
			err = fmt.Errorf("emojiUpdateModify: error getting or creating category: %s", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		updatedCategoryID = category.ID
		updateCategoryID = true
	}

	// only update image if provided with one
	var updateImage bool
	if image != nil && image.Size != 0 {
		updateImage = true
	}

	if !updateImage {
		// only updating fields, we only need
		// to do a database update for this
		columns := []string{"updated_at"}

		if updateCategoryID {
			emoji.CategoryID = updatedCategoryID
			columns = append(columns, "category_id")
		}

		var err error
		updatedEmoji, err = p.db.UpdateEmoji(ctx, emoji, columns...)
		if err != nil {
			err = fmt.Errorf("emojiUpdateModify: error updating emoji %s: %s", emoji.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	} else {
		// new image, so we need to reprocess the emoji
		data := func(ctx context.Context) (reader io.ReadCloser, fileSize int64, err error) {
			i, err := image.Open()
			return i, image.Size, err
		}

		var ai *media.AdditionalEmojiInfo
		if updateCategoryID {
			ai = &media.AdditionalEmojiInfo{
				CategoryID: &updatedCategoryID,
			}
		}

		processingEmoji, err := p.mediaManager.PreProcessEmoji(ctx, data, nil, emoji.Shortcode, emoji.ID, emoji.URI, ai, true)
		if err != nil {
			err = fmt.Errorf("emojiUpdateModify: error processing emoji %s: %s", emoji.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		updatedEmoji, err = processingEmoji.LoadEmoji(ctx)
		if err != nil {
			err = fmt.Errorf("emojiUpdateModify: error loading processed emoji %s: %s", emoji.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	adminEmoji, err := p.tc.EmojiToAdminAPIEmoji(ctx, updatedEmoji)
	if err != nil {
		err = fmt.Errorf("emojiUpdateModify: error converting updated emoji %s to admin emoji: %s", emoji.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return adminEmoji, nil
}
