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
	"github.com/superseriousbusiness/gotosocial/internal/media"
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
		return p.emojiUpdateCopy(ctx, emoji, form.Shortcode)
	case apimodel.EmojiUpdateDisable:
		return p.emojiUpdateDisable(ctx, emoji)
	case apimodel.EmojiUpdateModify:
		return p.emojiUpdateModify(ctx, emoji, form.Shortcode, form.Image, form.CategoryName)
	default:
		err := errors.New("unrecognized emoji action type")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}
}

func (p *processor) emojiUpdateCopy(ctx context.Context, emoji *gtsmodel.Emoji, shortcode *string) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if emoji.Domain == "" {
		err := fmt.Errorf("emojiUpdateCopy: emoji %s is not a remote emoji, cannot copy it to local", emoji.ID)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	return nil, nil
}

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

func (p *processor) emojiUpdateModify(ctx context.Context, emoji *gtsmodel.Emoji, shortcode *string, image *multipart.FileHeader, categoryName *string) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if emoji.Domain != "" {
		err := fmt.Errorf("emojiUpdateModify: emoji %s is not a local emoji, cannot do a modify action on it", emoji.ID)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	var updatedEmoji *gtsmodel.Emoji

	// keep existing shortcode unless a new one is defined
	var (
		updatedShortcode = emoji.Shortcode
		updateShortcode  bool
	)
	if shortcode != nil {
		updatedShortcode = *shortcode
		updateShortcode = true
	}

	// keep existing categoryID unless a new one is defined
	var (
		updatedCategoryID = emoji.CategoryID
		updateCategoryID  bool
	)
	if categoryName != nil {
		category, err := p.GetOrCreateEmojiCategory(ctx, *categoryName)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error putting id in category: %s", err), "error putting id in category")
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

		if updateShortcode {
			emoji.Shortcode = updatedShortcode
			columns = append(columns, "shortcode")
		}

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
		// new image, so we need to reprocess the emoji again, and
		// provide the other fields too as though they've changed;
		// they'll be updated inside the ProcessEmoji function
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

		processingEmoji, err := p.mediaManager.ProcessEmoji(ctx, data, nil, updatedShortcode, emoji.ID, emoji.URI, ai, true)
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
