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
	"fmt"
	"io"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (p *processor) EmojiCreate(ctx context.Context, account *gtsmodel.Account, user *gtsmodel.User, form *apimodel.EmojiCreateRequest) (*apimodel.Emoji, gtserror.WithCode) {
	if !*user.Admin {
		return nil, gtserror.NewErrorUnauthorized(fmt.Errorf("user %s not an admin", user.ID), "user is not an admin")
	}

	maybeExisting, err := p.db.GetEmojiByShortcodeDomain(ctx, form.Shortcode, "")
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
		category, err := p.GetOrCreateEmojiCategory(ctx, form.CategoryName)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error putting id in category: %s", err), "error putting id in category")
		}

		ai = &media.AdditionalEmojiInfo{
			CategoryID: &category.ID,
		}
	}

	processingEmoji, err := p.mediaManager.ProcessEmoji(ctx, data, nil, form.Shortcode, emojiID, emojiURI, ai, false)
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
