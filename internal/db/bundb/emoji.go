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

package bundb

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type emojiDB struct {
	conn *DBConn
}

func (e emojiDB) GetCustomEmojis(ctx context.Context) ([]*gtsmodel.Emoji, db.Error) {
	emojis := []*gtsmodel.Emoji{}

	q := e.conn.
		NewSelect().
		Model(&emojis).
		Where("visible_in_picker = true").
		Where("disabled = false").
		Order("shortcode ASC")

	if err := q.Scan(ctx); err != nil {
		return nil, e.conn.ProcessError(err)
	}
	return emojis, nil
}
