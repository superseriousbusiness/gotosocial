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

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func (p *processor) GetOrCreateEmojiCategory(ctx context.Context, name string) (*gtsmodel.EmojiCategory, error) {
	category, err := p.db.GetEmojiCategoryByName(ctx, name)
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

	if err := p.db.PutEmojiCategory(ctx, category); err != nil {
		err = fmt.Errorf("GetOrCreateEmojiCategory: error putting new emoji category in the database: %s", err)
		return nil, err
	}

	return category, nil
}
