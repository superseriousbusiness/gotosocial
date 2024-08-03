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

	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-iotools"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// EmojiCreate creates a custom emoji on this instance.
func (p *Processor) EmojiCreate(
	ctx context.Context,
	account *gtsmodel.Account,
	form *apimodel.EmojiCreateRequest,
) (*apimodel.Emoji, gtserror.WithCode) {

	// Get maximum supported local emoji size.
	maxsz := config.GetMediaEmojiLocalMaxSize()

	// Ensure media within size bounds.
	if form.Image.Size > int64(maxsz) {
		text := fmt.Sprintf("emoji exceeds configured max size: %s", maxsz)
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Open multipart file reader.
	mpfile, err := form.Image.Open()
	if err != nil {
		err := gtserror.Newf("error opening multipart file: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Wrap the multipart file reader to ensure is limited to max.
	rc, _, _ := iotools.UpdateReadCloserLimit(mpfile, int64(maxsz))
	data := func(context.Context) (io.ReadCloser, error) {
		return rc, nil
	}

	// Attempt to create the new local emoji.
	emoji, errWithCode := p.createEmoji(ctx,
		form.Shortcode,
		form.CategoryName,
		data,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	apiEmoji, err := p.converter.EmojiToAPIEmoji(ctx, emoji)
	if err != nil {
		err := gtserror.Newf("error converting emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return &apiEmoji, nil
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
		NextMaxIDValue: emojis[count-1].ShortcodeDomain(),
		PrevMinIDKey:   "min_shortcode_domain",
		PrevMinIDValue: emojis[0].ShortcodeDomain(),
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
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
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
	emojiID string,
	form *apimodel.EmojiUpdateRequest,
) (*apimodel.AdminEmoji, gtserror.WithCode) {

	// Get the emoji with given ID from the database.
	emoji, err := p.state.DB.GetEmojiByID(ctx, emojiID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error fetching emoji from db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check found.
	if emoji == nil {
		const text = "emoji not found"
		return nil, gtserror.NewErrorNotFound(errors.New(text), text)
	}

	switch form.Type {

	case apimodel.EmojiUpdateCopy:
		return p.emojiUpdateCopy(ctx, emoji, form.Shortcode, form.CategoryName)

	case apimodel.EmojiUpdateDisable:
		return p.emojiUpdateDisable(ctx, emoji)

	case apimodel.EmojiUpdateModify:
		return p.emojiUpdateModify(ctx, emoji, form.Image, form.CategoryName)

	default:
		const text = "unrecognized emoji update action type"
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
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

// emojiUpdateCopy copies and stores the given
// *remote* emoji as a *local* emoji, preserving
// the same image, and using the provided shortcode.
//
// The provided emoji model must correspond to an
// emoji already stored in the database + storage.
func (p *Processor) emojiUpdateCopy(
	ctx context.Context,
	target *gtsmodel.Emoji,
	shortcode *string,
	categoryName *string,
) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if target.IsLocal() {
		const text = "target emoji is not remote; cannot copy to local"
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Ensure target emoji is locally cached.
	target, err := p.federator.RecacheEmoji(ctx,
		target,
	)
	if err != nil {
		err := gtserror.Newf("error recaching emoji %s: %w", target.ImageRemoteURL, err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Get maximum supported local emoji size.
	maxsz := config.GetMediaEmojiLocalMaxSize()

	// Ensure target emoji image within size bounds.
	if bytesize.Size(target.ImageFileSize) > maxsz {
		text := fmt.Sprintf("emoji exceeds configured max size: %s", maxsz)
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Data function for copying just streams media
	// out of storage into an additional location.
	//
	// This means that data for the copy persists even
	// if the remote copied emoji gets deleted at some point.
	data := func(ctx context.Context) (io.ReadCloser, error) {
		rc, err := p.state.Storage.GetStream(ctx, target.ImagePath)
		return rc, err
	}

	// Attempt to create the new local emoji.
	emoji, errWithCode := p.createEmoji(ctx,
		util.PtrOrZero(shortcode),
		util.PtrOrZero(categoryName),
		data,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	apiEmoji, err := p.converter.EmojiToAdminAPIEmoji(ctx, emoji)
	if err != nil {
		err := gtserror.Newf("error converting emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiEmoji, nil
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
		err := gtserror.Newf("error converting emoji: %w", err)
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
	categoryName *string,
) (*apimodel.AdminEmoji, gtserror.WithCode) {
	if !emoji.IsLocal() {
		const text = "cannot modify remote emoji"
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Ensure there's actually something to update.
	if image == nil && categoryName == nil {
		const text = "no changes were provided"
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Check if we need to
	// set a new category ID.
	var newCategoryID *string
	switch {
	case categoryName == nil:
		// No changes.

	case *categoryName == "":
		// Emoji category was unset.
		newCategoryID = util.Ptr("")
		emoji.CategoryID = ""
		emoji.Category = nil

	case *categoryName != "":
		// A category was provided, get or create relevant emoji category.
		category, errWithCode := p.mustGetEmojiCategory(ctx, *categoryName)
		if errWithCode != nil {
			return nil, errWithCode
		}

		// Update emoji category if
		// it's different from before.
		if category.ID != emoji.CategoryID {
			newCategoryID = &category.ID
			emoji.CategoryID = category.ID
			emoji.Category = category
		}
	}

	// Check whether any image changes were requested.
	imageUpdated := (image != nil && image.Size > 0)

	if !imageUpdated && newCategoryID != nil {
		// Only updating category; only a single database update required.
		if err := p.state.DB.UpdateEmoji(ctx, emoji, "category_id"); err != nil {
			err := gtserror.Newf("error updating emoji in db: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	} else if imageUpdated {
		var err error

		// Updating image and maybe categoryID.
		// We can do both at the same time :)

		// Get maximum supported local emoji size.
		maxsz := config.GetMediaEmojiLocalMaxSize()

		// Ensure media within size bounds.
		if image.Size > int64(maxsz) {
			text := fmt.Sprintf("emoji exceeds configured max size: %s", maxsz)
			return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
		}

		// Open multipart file reader.
		mpfile, err := image.Open()
		if err != nil {
			err := gtserror.Newf("error opening multipart file: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Wrap the multipart file reader to ensure is limited to max.
		rc, _, _ := iotools.UpdateReadCloserLimit(mpfile, int64(maxsz))
		data := func(context.Context) (io.ReadCloser, error) {
			return rc, nil
		}

		// Include category ID
		// update if necessary.
		ai := media.AdditionalEmojiInfo{}
		ai.CategoryID = newCategoryID

		// Prepare emoji model for update+recache from new data.
		processing, err := p.media.UpdateEmoji(ctx, emoji, data, ai)
		if err != nil {
			err := gtserror.Newf("error preparing recache: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Load to trigger update + write.
		emoji, err = processing.Load(ctx)
		if err != nil {
			err := gtserror.Newf("error processing emoji %s: %w", emoji.Shortcode, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	adminEmoji, err := p.converter.EmojiToAdminAPIEmoji(ctx, emoji)
	if err != nil {
		err := gtserror.Newf("error converting emoji: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return adminEmoji, nil
}

// createEmoji will create a new local emoji
// with the given shortcode, attached category
// name (if any) and data source function.
func (p *Processor) createEmoji(
	ctx context.Context,
	shortcode string,
	categoryName string,
	data media.DataFunc,
) (
	*gtsmodel.Emoji,
	gtserror.WithCode,
) {
	// Validate shortcode.
	if shortcode == "" {
		const text = "empty shortcode name"
		return nil, gtserror.NewErrorBadRequest(errors.New(text), text)
	}

	// Look for an existing local emoji with shortcode to ensure this is new.
	existing, err := p.state.DB.GetEmojiByShortcodeDomain(ctx, shortcode, "")
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error fetching emoji from db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	} else if existing != nil {
		const text = "emoji with shortcode already exists"
		return nil, gtserror.NewErrorConflict(errors.New(text), text)
	}

	var categoryID *string

	if categoryName != "" {
		// A category was provided, get / create relevant emoji category.
		category, errWithCode := p.mustGetEmojiCategory(ctx, categoryName)
		if errWithCode != nil {
			return nil, errWithCode
		}

		// Set category ID for emoji.
		categoryID = &category.ID
	}

	// Store to instance storage.
	return p.c.StoreLocalEmoji(
		ctx,
		shortcode,
		data,
		media.AdditionalEmojiInfo{
			CategoryID: categoryID,
		},
	)
}

// mustGetEmojiCategory either gets an existing
// category with the given name from the database,
// or, if the category doesn't yet exist, it creates
// the category and then returns it.
func (p *Processor) mustGetEmojiCategory(
	ctx context.Context,
	name string,
) (
	*gtsmodel.EmojiCategory,
	gtserror.WithCode,
) {
	// Look for an existing emoji category with name.
	category, err := p.state.DB.GetEmojiCategoryByName(ctx, name)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error fetching emoji category from db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if category != nil {
		// We had it already.
		return category, nil
	}

	// Create new ID.
	id := id.NewULID()

	// Prepare new category for insertion.
	category = &gtsmodel.EmojiCategory{
		ID:   id,
		Name: name,
	}

	// Insert new category into the database.
	err = p.state.DB.PutEmojiCategory(ctx, category)
	if err != nil {
		err := gtserror.Newf("error inserting emoji category into db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return category, nil
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
