/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package pg

import (
	"github.com/go-pg/pg/v10"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (ps *postgresService) GetNotificationsForAccount(accountID string, limit int, maxID string, sinceID string) ([]*gtsmodel.Notification, db.Error) {
	notifications := []*gtsmodel.Notification{}

	q := ps.conn.Model(&notifications).Where("target_account_id = ?", accountID)

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if sinceID != "" {
		q = q.Where("id > ?", sinceID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	q = q.Order("created_at DESC")

	if err := q.Select(); err != nil {
		if err != pg.ErrNoRows {
			return nil, err
		}

	}
	return notifications, nil
}
