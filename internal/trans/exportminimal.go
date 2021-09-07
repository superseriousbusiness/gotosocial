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

package trans

import (
	"context"
	"fmt"
	"os"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	transmodel "github.com/superseriousbusiness/gotosocial/internal/trans/model"
)

func (e *exporter) ExportMinimal(ctx context.Context, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	// export all local accounts we have in the database
	localAccounts, err := e.exportAccounts(ctx, []db.Where{{Key: "domain", Value: nil}}, f)
	if err != nil {
		return fmt.Errorf("ExportMinimal: error exporting accounts: %s", err)
	}

	// export all blocks that relate to those accounts
	blocks, err := e.exportBlocks(ctx, localAccounts, f)
	if err != nil {
		return fmt.Errorf("ExportMinimal: error exporting blocks: %s", err)
	}

	// for each block, make sure we've written out the account owning it, or targeted by it
	for _, b := range blocks {
		_, alreadyWritten := e.writtenIDs[b.AccountID]
		if !alreadyWritten {
			_, err := e.exportAccounts(ctx, []db.Where{{Key: "id", Value: b.AccountID}}, f)
			if err != nil {
				return fmt.Errorf("ExportMinimal: error exporting block owner account: %s", err)
			}
		}

		_, alreadyWritten = e.writtenIDs[b.TargetAccountID]
		if !alreadyWritten {
			_, err := e.exportAccounts(ctx, []db.Where{{Key: "id", Value: b.TargetAccountID}}, f)
			if err != nil {
				return fmt.Errorf("ExportMinimal: error exporting block target account: %s", err)
			}
		}
	}

	return neatClose(f)
}

func (e *exporter) exportAccounts(ctx context.Context, where []db.Where, f *os.File) ([]*transmodel.Account, error) {
	// select using the 'where' we've been provided
	accounts := []*transmodel.Account{}
	if err := e.db.GetWhere(ctx, where, &accounts); err != nil {
		return nil, fmt.Errorf("exportAccounts: error selecting accounts: %s", err)
	}

	// write any accounts found to file
	for _, a := range accounts {
		if err := e.accountEncode(ctx, f, a); err != nil {
			return nil, fmt.Errorf("exportAccounts: error encoding account: %s", err)
		}
	}

	return accounts, nil
}

func (e *exporter) exportBlocks(ctx context.Context, accounts []*transmodel.Account, f *os.File) ([]*transmodel.Block, error) {
	blocksUnique := make(map[string]*transmodel.Block)

	// for each account we want to export both where it's blocking and where it's blocked
	for _, a := range accounts {
		// 1. export blocks owned by given account
		whereBlocking := []db.Where{{Key: "account_id", Value: a.ID}}
		blocking := []*transmodel.Block{}
		if err := e.db.GetWhere(ctx, whereBlocking, &blocking); err != nil {
			return nil, fmt.Errorf("exportBlocks: error selecting blocks owned by account %s: %s", a.ID, err)
		}
		for _, b := range blocking {
			b.Type = transmodel.TransBlock
			if err := e.simpleEncode(ctx, f, b, b.ID); err != nil {
				return nil, fmt.Errorf("exportBlocks: error encoding block owned by account %s: %s", a.ID, err)
			}
			blocksUnique[b.ID] = b
		}

		// 2. export blocks that target given account
		whereBlocked := []db.Where{{Key: "target_account_id", Value: a.ID}}
		blocked := []*transmodel.Block{}
		if err := e.db.GetWhere(ctx, whereBlocked, &blocked); err != nil {
			return nil, fmt.Errorf("exportBlocks: error selecting blocks targeting account %s: %s", a.ID, err)
		}
		for _, b := range blocked {
			b.Type = transmodel.TransBlock
			if err := e.simpleEncode(ctx, f, b, b.ID); err != nil {
				return nil, fmt.Errorf("exportBlocks: error encoding block targeting account %s: %s", a.ID, err)
			}
			blocksUnique[b.ID] = b
		}
	}

	// now return all the blocks we found
	blocks := []*transmodel.Block{}
	for _, b := range blocksUnique {
		blocks = append(blocks, b)
	}

	return blocks, nil
}
