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
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (p *processor) EmojiCategoriesGet(ctx context.Context) ([]*apimodel.EmojiCategory, gtserror.WithCode) {
	categories, err := p.db.GetEmojiCategories(ctx)
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
