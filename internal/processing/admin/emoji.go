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

package admin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// EmojiCreate creates a custom emoji on this instance.
func (p *Processor) EmojiCreate(ctx context.Context, account *gtsmodel.Account, user *gtsmodel.User, form *apimodel.EmojiCreateRequest) (*apimodel.Emoji, gtserror.WithCode) {
	if !*user.Admin {
		return nil, gtserror.NewErrorUnauthorized(fmt.Errorf("user %s not an admin", user.ID), "user is not an admin")
	}

	maybeExisting, err := p.state.DB.GetEmojiByShortcodeDomain(ctx, form.Shortcode, "")
	if maybeExisting != nil {
		return nil, gtserror.NewErrorConflict(fmt.Errorf("emoji with shortcode %s already exists", form.Shortcode), fmt.Sprintf("emoji with shortcode %s already exists", form.Shortcode))
	}

	if err != nil && err != db.ErrNoEntries {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error checking existence of emoji with shortcode %s: %s", form.Shortcode, err))
	}

	emojiID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error creating id for new emoji: %s", err), "error creating emoji ID")
	}

	emojiURI := uris.GenerateURIForEmoji(emojiID)

	data := func(innerCtx context.Context) (io.ReadCloser, int64, error) {
		f, err := form.Image.Open()
		return f, form.Image.Size, err
	}

	var ai *media.AdditionalEmojiInfo
	if form.CategoryName != "" {
		category, err := p.getOrCreateEmojiCategory(ctx, form.CategoryName)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error putting id in category: %s", err), "error putting id in category")
		}

		ai = &media.AdditionalEmojiInfo{
			CategoryID: &category.ID,
		}
	}

	processingEmoji, err := p.mediaManager.PreProcessEmoji(ctx, data, nil, form.Shortcode, emojiID, emojiURI, ai, false)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error processing emoji: %s", err), "error processing emoji")
	}

	emoji, err := processingEmoji.LoadEmoji(ctx)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error loading emoji: %s", err), "error loading emoji")
	}

	apiEmoji, err := p.tc.EmojiToAPIEmoji(ctx, emoji)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting emoji: %s", err), "error converting emoji to api representation")
	}

	return &apiEmoji, nil
}

// EmojisGet returns an admin view of custom emojis, filtered with the given parameters.
func (p *Processor) EmojisGet(
	ctx context.Context,
	account *gtsmodel.Account,
	user *gtsmodel.User,
	domain string,
	includeDisabled bool,
	includeEnabled bool,
	shortcode string,
	maxShortcodeDomain string,
	minShortcodeDomain string,
	limit int,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	if !*user.Admin {
		return nil, gtserror.NewErrorUnauthorized(fmt.Errorf("user %s not an admin", user.ID), "user is not an admin")
	}

	emojis, err := p.state.DB.GetEmojis(ctx, domain, includeDisabled, includeEnabled, shortcode, maxShortcodeDomain, minShortcodeDomain, limit)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := fmt.Errorf("EmojisGet: db error: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(emojis)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	items := make([]interface{}, 0, count)
	for _, emoji := range emojis {
		adminEmoji, err := p.tc.EmojiToAdminAPIEmoji(ctx, emoji)
		if err != nil {
			err := fmt.Errorf("EmojisGet: error converting emoji to admin model emoji: %s", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		items = append(items, adminEmoji)
	}

	filterBuilder := strings.Builder{}
	filterBuilder.WriteString("filter=")

	switch domain {
	case "", "local":
		filterBuilder.WriteString("domain:local")
	case db.EmojiAllDomains:
		filterBuilder.WriteString("domain:all")
	default:
		filterBuilder.WriteString("domain:")
		filterBuilder.WriteString(domain)
	}

	if includeDisabled != includeEnabled {
		if includeDisabled {
			filterBuilder.WriteString(",disabled")
		}
		if includeEnabled {
			filterBuilder.WriteString(",enabled")
		}
	}

	if shortcode != "" {
		filterBuilder.WriteString(",shortcode:")
		filterBuilder.WriteString(shortcode)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:            items,
		Path:             "api/v1/admin/custom_emojis",
		NextMaxIDKey:     "max_shortcode_domain",
		NextMaxIDValue:   util.ShortcodeDomain(emojis[count-1]),
		PrevMinIDKey:     "min_shortcode_domain",
		PrevMinIDValue:   util.ShortcodeDomain(emojis[0]),
		Limit:            limit,
		ExtraQueryParams: []string{filterBuilder.String()},
	})
}

// EmojiGet returns the admin view of one custom emoji with the given id.
func (p *Processor) EmojiGet(ctx context.Context, account *gtsmodel.Account, user *gtsmodel.User, id string) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if !*user.Admin {
		return nil, gtserror.NewErrorUnauthorized(fmt.Errorf("user %s not an admin", user.ID), "user is not an admin")
	}

	emoji, err := p.state.DB.GetEmojiByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("EmojiGet: no emoji with id %s found in the db", id)
			return nil, gtserror.NewErrorNotFound(err)
		}
		err := fmt.Errorf("EmojiGet: db error: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	adminEmoji, err := p.tc.EmojiToAdminAPIEmoji(ctx, emoji)
	if err != nil {
		err = fmt.Errorf("EmojiGet: error converting emoji to admin api emoji: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return adminEmoji, nil
}

// EmojiDelete deletes one emoji from the database, with the given id.
func (p *Processor) EmojiDelete(ctx context.Context, id string) (*apimodel.AdminEmoji, gtserror.WithCode) {
	emoji, err := p.state.DB.GetEmojiByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("EmojiDelete: no emoji with id %s found in the db", id)
			return nil, gtserror.NewErrorNotFound(err)
		}
		err := fmt.Errorf("EmojiDelete: db error: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if emoji.Domain != "" {
		err = fmt.Errorf("EmojiDelete: emoji with id %s was not a local emoji, will not delete", id)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	adminEmoji, err := p.tc.EmojiToAdminAPIEmoji(ctx, emoji)
	if err != nil {
		err = fmt.Errorf("EmojiDelete: error converting emoji to admin api emoji: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.state.DB.DeleteEmojiByID(ctx, id); err != nil {
		err := fmt.Errorf("EmojiDelete: db error: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return adminEmoji, nil
}

// EmojiUpdate updates one emoji with the given id, using the provided form parameters.
func (p *Processor) EmojiUpdate(ctx context.Context, id string, form *apimodel.EmojiUpdateRequest) (*apimodel.AdminEmoji, gtserror.WithCode) {
	emoji, err := p.state.DB.GetEmojiByID(ctx, id)
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

// EmojiCategoriesGet returns all custom emoji categories that exist on this instance.
func (p *Processor) EmojiCategoriesGet(ctx context.Context) ([]*apimodel.EmojiCategory, gtserror.WithCode) {
	categories, err := p.state.DB.GetEmojiCategories(ctx)
	if err != nil {
		err := fmt.Errorf("EmojiCategoriesGet: db error: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiCategories := make([]*apimodel.EmojiCategory, 0, len(categories))
	for _, category := range categories {
		apiCategory, err := p.tc.EmojiCategoryToAPIEmojiCategory(ctx, category)
		if err != nil {
			err := fmt.Errorf("EmojiCategoriesGet: error converting emoji category to api emoji category: %s", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		apiCategories = append(apiCategories, apiCategory)
	}

	return apiCategories, nil
}

/*
	UTIL FUNCTIONS
*/

func (p *Processor) getOrCreateEmojiCategory(ctx context.Context, name string) (*gtsmodel.EmojiCategory, error) {
	category, err := p.state.DB.GetEmojiCategoryByName(ctx, name)
	if err == nil {
		return category, nil
	}

	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("GetOrCreateEmojiCategory: database error trying get emoji category by name: %s", err)
		return nil, err
	}

	// we don't have the category yet, just create it with the given name
	categoryID, err := id.NewRandomULID()
	if err != nil {
		err = fmt.Errorf("GetOrCreateEmojiCategory: error generating id for new emoji category: %s", err)
		return nil, err
	}

	category = &gtsmodel.EmojiCategory{
		ID:   categoryID,
		Name: name,
	}

	if err := p.state.DB.PutEmojiCategory(ctx, category); err != nil {
		err = fmt.Errorf("GetOrCreateEmojiCategory: error putting new emoji category in the database: %s", err)
		return nil, err
	}

	return category, nil
}

// copy an emoji from remote to local
func (p *Processor) emojiUpdateCopy(ctx context.Context, emoji *gtsmodel.Emoji, shortcode *string, categoryName *string) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if emoji.Domain == "" {
		err := fmt.Errorf("emojiUpdateCopy: emoji %s is not a remote emoji, cannot copy it to local", emoji.ID)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	if shortcode == nil {
		err := fmt.Errorf("emojiUpdateCopy: emoji %s could not be copied, no shortcode provided", emoji.ID)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	maybeExisting, err := p.state.DB.GetEmojiByShortcodeDomain(ctx, *shortcode, "")
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
		rc, err := p.state.Storage.GetStream(ctx, emoji.ImagePath)
		return rc, int64(emoji.ImageFileSize), err
	}

	var ai *media.AdditionalEmojiInfo
	if categoryName != nil {
		category, err := p.getOrCreateEmojiCategory(ctx, *categoryName)
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
func (p *Processor) emojiUpdateDisable(ctx context.Context, emoji *gtsmodel.Emoji) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if emoji.Domain == "" {
		err := fmt.Errorf("emojiUpdateDisable: emoji %s is not a remote emoji, cannot disable it via this endpoint", emoji.ID)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	emojiDisabled := true
	emoji.Disabled = &emojiDisabled
	updatedEmoji, err := p.state.DB.UpdateEmoji(ctx, emoji, "updated_at", "disabled")
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
func (p *Processor) emojiUpdateModify(ctx context.Context, emoji *gtsmodel.Emoji, image *multipart.FileHeader, categoryName *string) (*apimodel.AdminEmoji, gtserror.WithCode) {
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
		category, err := p.getOrCreateEmojiCategory(ctx, *categoryName)
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
		updatedEmoji, err = p.state.DB.UpdateEmoji(ctx, emoji, columns...)
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
