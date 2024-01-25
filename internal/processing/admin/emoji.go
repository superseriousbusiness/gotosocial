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
func (p *Processor) EmojiCreate(
	ctx context.Context,
	account *gtsmodel.Account,
	form *apimodel.EmojiCreateRequest,
) (*apimodel.Emoji, gtserror.WithCode) {
	// Ensure emoji with this shortcode
	// doesn't already exist on the instance.
	maybeExisting, err := p.state.DB.GetEmojiByShortcodeDomain(ctx, form.Shortcode, "")
	if err != nil && errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error checking existence of emoji with shortcode %s: %w", form.Shortcode, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if maybeExisting != nil {
		err := fmt.Errorf("emoji with shortcode %s already exists", form.Shortcode)
		return nil, gtserror.NewErrorConflict(err, err.Error())
	}

	// Prepare data function for emoji processing
	// (just read data from the submitted form).
	data := func(innerCtx context.Context) (io.ReadCloser, int64, error) {
		f, err := form.Image.Open()
		return f, form.Image.Size, err
	}

	// If category was supplied on the form,
	// ensure the category exists and provide
	// it as additional info to emoji processing.
	var ai *media.AdditionalEmojiInfo
	if form.CategoryName != "" {
		category, err := p.getOrCreateEmojiCategory(ctx, form.CategoryName)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		ai = &media.AdditionalEmojiInfo{
			CategoryID: &category.ID,
		}
	}

	// Generate new emoji ID and URI.
	emojiID, err := id.NewRandomULID()
	if err != nil {
		err := gtserror.Newf("error creating id for new emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	emojiURI := uris.URIForEmoji(emojiID)

	// Begin media processing.
	processingEmoji, err := p.mediaManager.PreProcessEmoji(ctx,
		data, form.Shortcode, emojiID, emojiURI, ai, false,
	)
	if err != nil {
		err := gtserror.Newf("error processing emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Complete processing immediately.
	emoji, err := processingEmoji.LoadEmoji(ctx)
	if err != nil {
		err := gtserror.Newf("error loading emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiEmoji, err := p.converter.EmojiToAPIEmoji(ctx, emoji)
	if err != nil {
		err := gtserror.Newf("error converting emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return &apiEmoji, nil
}

// emojisGetFilterParams builds extra
// query parameters to return as part
// of an Emojis pageable response.
//
// The returned string will look like:
//
// "filter=domain:all,enabled,shortcode:example"
func emojisGetFilterParams(
	shortcode string,
	domain string,
	includeDisabled bool,
	includeEnabled bool,
) string {
	var filterBuilder strings.Builder
	filterBuilder.WriteString("filter=")

	switch domain {
	case "", "local":
		// Local emojis only.
		filterBuilder.WriteString("domain:local")

	case db.EmojiAllDomains:
		// Local or remote.
		filterBuilder.WriteString("domain:all")

	default:
		// Specific domain only.
		filterBuilder.WriteString("domain:" + domain)
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
		// Specific shortcode only.
		filterBuilder.WriteString(",shortcode:" + shortcode)
	}

	return filterBuilder.String()
}

// EmojisGet returns an admin view of custom
// emojis, filtered with the given parameters.
func (p *Processor) EmojisGet(
	ctx context.Context,
	account *gtsmodel.Account,
	domain string,
	includeDisabled bool,
	includeEnabled bool,
	shortcode string,
	maxShortcodeDomain string,
	minShortcodeDomain string,
	limit int,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	emojis, err := p.state.DB.GetEmojisBy(ctx,
		domain,
		includeDisabled,
		includeEnabled,
		shortcode,
		maxShortcodeDomain,
		minShortcodeDomain,
		limit,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(emojis)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	items := make([]interface{}, 0, count)
	for _, emoji := range emojis {
		adminEmoji, err := p.converter.EmojiToAdminAPIEmoji(ctx, emoji)
		if err != nil {
			err := gtserror.Newf("error converting emoji to admin model emoji: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		items = append(items, adminEmoji)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "api/v1/admin/custom_emojis",
		NextMaxIDKey:   "max_shortcode_domain",
		NextMaxIDValue: util.ShortcodeDomain(emojis[count-1]),
		PrevMinIDKey:   "min_shortcode_domain",
		PrevMinIDValue: util.ShortcodeDomain(emojis[0]),
		Limit:          limit,
		ExtraQueryParams: []string{
			emojisGetFilterParams(
				shortcode,
				domain,
				includeDisabled,
				includeEnabled,
			),
		},
	})
}

// EmojiGet returns the admin view of
// one custom emoji with the given id.
func (p *Processor) EmojiGet(
	ctx context.Context,
	account *gtsmodel.Account,
	id string,
) (*apimodel.AdminEmoji, gtserror.WithCode) {
	emoji, err := p.state.DB.GetEmojiByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if emoji == nil {
		err := gtserror.Newf("no emoji with id %s found in the db", id)
		return nil, gtserror.NewErrorNotFound(err)
	}

	adminEmoji, err := p.converter.EmojiToAdminAPIEmoji(ctx, emoji)
	if err != nil {
		err := gtserror.Newf("error converting emoji to admin api emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return adminEmoji, nil
}

// EmojiDelete deletes one *local* emoji
// from the database, with the given id.
func (p *Processor) EmojiDelete(
	ctx context.Context,
	id string,
) (*apimodel.AdminEmoji, gtserror.WithCode) {
	emoji, err := p.state.DB.GetEmojiByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if emoji == nil {
		err := gtserror.Newf("no emoji with id %s found in the db", id)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if !emoji.IsLocal() {
		err := fmt.Errorf("emoji with id %s was not a local emoji, will not delete", id)
		return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	// Convert to admin emoji before deletion,
	// so we can return the deleted emoji.
	adminEmoji, err := p.converter.EmojiToAdminAPIEmoji(ctx, emoji)
	if err != nil {
		err := gtserror.Newf("error converting emoji to admin api emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.state.DB.DeleteEmojiByID(ctx, id); err != nil {
		err := gtserror.Newf("db error deleting emoji %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return adminEmoji, nil
}

// EmojiUpdate updates one emoji with the
// given id, using the provided form parameters.
func (p *Processor) EmojiUpdate(
	ctx context.Context,
	id string,
	form *apimodel.EmojiUpdateRequest,
) (*apimodel.AdminEmoji, gtserror.WithCode) {
	emoji, err := p.state.DB.GetEmojiByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if emoji == nil {
		err := gtserror.Newf("no emoji with id %s found in the db", id)
		return nil, gtserror.NewErrorNotFound(err)
	}

	switch t := form.Type; t {

	case apimodel.EmojiUpdateCopy:
		return p.emojiUpdateCopy(ctx, emoji, form.Shortcode, form.CategoryName)

	case apimodel.EmojiUpdateDisable:
		return p.emojiUpdateDisable(ctx, emoji)

	case apimodel.EmojiUpdateModify:
		return p.emojiUpdateModify(ctx, emoji, form.Image, form.CategoryName)

	default:
		err := fmt.Errorf("unrecognized emoji action type %s", t)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}
}

// EmojiCategoriesGet returns all custom emoji
// categories that exist on this instance.
func (p *Processor) EmojiCategoriesGet(
	ctx context.Context,
) ([]*apimodel.EmojiCategory, gtserror.WithCode) {
	categories, err := p.state.DB.GetEmojiCategories(ctx)
	if err != nil {
		err := gtserror.Newf("db error getting emoji categories: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiCategories := make([]*apimodel.EmojiCategory, 0, len(categories))
	for _, category := range categories {
		apiCategory, err := p.converter.EmojiCategoryToAPIEmojiCategory(ctx, category)
		if err != nil {
			err := gtserror.Newf("error converting emoji category to api emoji category: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		apiCategories = append(apiCategories, apiCategory)
	}

	return apiCategories, nil
}

/*
	UTIL FUNCTIONS
*/

// getOrCreateEmojiCategory either gets an existing
// category with the given name from the database,
// or, if the category doesn't yet exist, it creates
// the category and then returns it.
func (p *Processor) getOrCreateEmojiCategory(
	ctx context.Context,
	name string,
) (*gtsmodel.EmojiCategory, error) {
	category, err := p.state.DB.GetEmojiCategoryByName(ctx, name)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.Newf(
			"database error trying get emoji category %s: %w",
			name, err,
		)
	}

	if category != nil {
		// We had it already.
		return category, nil
	}

	// We don't have the category yet,
	// create it with the given name.
	categoryID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.Newf(
			"error generating id for new emoji category %s: %w",
			name, err,
		)
	}

	category = &gtsmodel.EmojiCategory{
		ID:   categoryID,
		Name: name,
	}

	if err := p.state.DB.PutEmojiCategory(ctx, category); err != nil {
		return nil, gtserror.Newf(
			"db error putting new emoji category %s: %w",
			name, err,
		)
	}

	return category, nil
}

// emojiUpdateCopy copies and stores the given
// *remote* emoji as a *local* emoji, preserving
// the same image, and using the provided shortcode.
//
// The provided emoji model must correspond to an
// emoji already stored in the database + storage.
func (p *Processor) emojiUpdateCopy(
	ctx context.Context,
	targetEmoji *gtsmodel.Emoji,
	shortcode *string,
	category *string,
) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if targetEmoji.IsLocal() {
		err := fmt.Errorf("emoji %s is not a remote emoji, cannot copy it to local", targetEmoji.ID)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	if shortcode == nil {
		err := errors.New("no shortcode provided")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	sc := *shortcode
	if sc == "" {
		err := errors.New("empty shortcode provided")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Ensure we don't already have an emoji
	// stored locally with this shortcode.
	maybeExisting, err := p.state.DB.GetEmojiByShortcodeDomain(ctx, sc, "")
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error checking for emoji with shortcode %s: %w", sc, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if maybeExisting != nil {
		err := fmt.Errorf("emoji with shortcode %s already exists on this instance", sc)
		return nil, gtserror.NewErrorConflict(err, err.Error())
	}

	// We don't have an emoji with this
	// shortcode yet! Prepare to create it.

	// Data function for copying just streams media
	// out of storage into an additional location.
	//
	// This means that data for the copy persists even
	// if the remote copied emoji gets deleted at some point.
	data := func(ctx context.Context) (io.ReadCloser, int64, error) {
		rc, err := p.state.Storage.GetStream(ctx, targetEmoji.ImagePath)
		return rc, int64(targetEmoji.ImageFileSize), err
	}

	// Generate new emoji ID and URI.
	emojiID, err := id.NewRandomULID()
	if err != nil {
		err := gtserror.Newf("error creating id for new emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	emojiURI := uris.URIForEmoji(emojiID)

	// If category was supplied, ensure the
	// category exists and provide it as
	// additional info to emoji processing.
	var ai *media.AdditionalEmojiInfo
	if category != nil && *category != "" {
		category, err := p.getOrCreateEmojiCategory(ctx, *category)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		ai = &media.AdditionalEmojiInfo{
			CategoryID: &category.ID,
		}
	}

	// Begin media processing.
	processingEmoji, err := p.mediaManager.PreProcessEmoji(ctx,
		data, sc, emojiID, emojiURI, ai, false,
	)
	if err != nil {
		err := gtserror.Newf("error processing emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Complete processing immediately.
	newEmoji, err := processingEmoji.LoadEmoji(ctx)
	if err != nil {
		err := gtserror.Newf("error loading emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	adminEmoji, err := p.converter.EmojiToAdminAPIEmoji(ctx, newEmoji)
	if err != nil {
		err := gtserror.Newf("error converting emoji %s to admin emoji: %w", newEmoji.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return adminEmoji, nil
}

// emojiUpdateDisable marks the given *remote*
// emoji as disabled by setting disabled = true.
//
// The provided emoji model must correspond to an
// emoji already stored in the database + storage.
func (p *Processor) emojiUpdateDisable(
	ctx context.Context,
	emoji *gtsmodel.Emoji,
) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if emoji.IsLocal() {
		err := fmt.Errorf("emoji %s is not a remote emoji, cannot disable it via this endpoint", emoji.ID)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Only bother with a db call
	// if emoji not already disabled.
	if !*emoji.Disabled {
		emoji.Disabled = util.Ptr(true)
		if err := p.state.DB.UpdateEmoji(ctx, emoji, "disabled"); err != nil {
			err := gtserror.Newf("db error updating emoji %s: %w", emoji.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	adminEmoji, err := p.converter.EmojiToAdminAPIEmoji(ctx, emoji)
	if err != nil {
		err := gtserror.Newf("error converting emoji %s to admin emoji: %w", emoji.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return adminEmoji, nil
}

// emojiUpdateModify modifies the given *local* emoji.
//
// Either one of image or category must be non-nil,
// otherwise there's nothing to modify. If category
// is non-nil and dereferences to an empty string,
// category will be cleared.
//
// The provided emoji model must correspond to an
// emoji already stored in the database + storage.
func (p *Processor) emojiUpdateModify(
	ctx context.Context,
	emoji *gtsmodel.Emoji,
	image *multipart.FileHeader,
	category *string,
) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if !emoji.IsLocal() {
		err := fmt.Errorf("emoji %s is not a local emoji, cannot update it via this endpoint", emoji.ID)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Ensure there's actually something to update.
	if image == nil && category == nil {
		err := errors.New("neither new category nor new image set, cannot update")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Only update category
	// if it's changed.
	var (
		newCategory      *gtsmodel.EmojiCategory
		newCategoryID    string
		updateCategoryID bool
	)

	if category != nil {
		catName := *category
		if catName != "" {
			// Set new category.
			var err error
			newCategory, err = p.getOrCreateEmojiCategory(ctx, catName)
			if err != nil {
				err := gtserror.Newf("error getting or creating category: %w", err)
				return nil, gtserror.NewErrorInternalError(err)
			}

			newCategoryID = newCategory.ID
		} else {
			// Clear existing category.
			newCategoryID = ""
		}

		updateCategoryID = emoji.CategoryID != newCategoryID
	}

	// Only update image
	// if one is provided.
	var updateImage bool
	if image != nil && image.Size != 0 {
		updateImage = true
	}

	if updateCategoryID && !updateImage {
		// Only updating category; we only
		// need to do a db update for this.
		emoji.CategoryID = newCategoryID
		emoji.Category = newCategory
		if err := p.state.DB.UpdateEmoji(ctx, emoji, "category_id"); err != nil {
			err := gtserror.Newf("db error updating emoji %s: %w", emoji.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	} else if updateImage {
		// Updating image and maybe categoryID.
		// We can do both at the same time :)

		// Set data function to provided image.
		data := func(ctx context.Context) (io.ReadCloser, int64, error) {
			i, err := image.Open()
			return i, image.Size, err
		}

		// If necessary, include
		// update to categoryID too.
		var ai *media.AdditionalEmojiInfo
		if updateCategoryID {
			ai = &media.AdditionalEmojiInfo{
				CategoryID: &newCategoryID,
			}
		}

		// Begin media processing.
		processingEmoji, err := p.mediaManager.PreProcessEmoji(ctx,
			data, emoji.Shortcode, emoji.ID, emoji.URI, ai, false,
		)
		if err != nil {
			err := gtserror.Newf("error processing emoji: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Replace emoji ptr with newly-processed version.
		emoji, err = processingEmoji.LoadEmoji(ctx)
		if err != nil {
			err := gtserror.Newf("error loading emoji: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	adminEmoji, err := p.converter.EmojiToAdminAPIEmoji(ctx, emoji)
	if err != nil {
		err := gtserror.Newf("error converting emoji %s to admin emoji: %w", emoji.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return adminEmoji, nil
}
