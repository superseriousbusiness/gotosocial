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

func (ps *postgresService) GetBlocksForAccount(accountID string, maxID string, sinceID string, limit int) ([]*gtsmodel.Account, string, string, error) {
	blocks := []*gtsmodel.Block{}

	fq := ps.conn.Model(&blocks).
		Where("block.account_id = ?", accountID).
		Relation("TargetAccount").
		Order("block.id DESC")

	if maxID != "" {
		fq = fq.Where("block.id < ?", maxID)
	}

	if sinceID != "" {
		fq = fq.Where("block.id > ?", sinceID)
	}

	if limit > 0 {
		fq = fq.Limit(limit)
	}

	err := fq.Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, "", "", db.ErrNoEntries{}
		}
		return nil, "", "", err
	}

	if len(blocks) == 0 {
		return nil, "", "", db.ErrNoEntries{}
	}

	accounts := []*gtsmodel.Account{}
	for _, b := range blocks {
		accounts = append(accounts, b.TargetAccount)
	}

	nextMaxID := blocks[len(blocks)-1].ID
	prevMinID := blocks[0].ID
	return accounts, nextMaxID, prevMinID, nil
}
