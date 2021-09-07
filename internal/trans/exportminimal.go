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
	"errors"
	"fmt"
	"os"

	"github.com/superseriousbusiness/gotosocial/internal/db"
)

func (e *exporter) ExportMinimal(ctx context.Context, path string) error {
	if path == "" {
		return errors.New("ExportMinimal: path empty")
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("ExportMinimal: couldn't export to %s: %s", path, err)
	}

	// export all local accounts we have in the database
	localAccounts, err := e.exportAccounts(ctx, []db.Where{{Key: "domain", Value: nil}}, f)
	if err != nil {
		return fmt.Errorf("ExportMinimal: error exporting accounts: %s", err)
	}

	// export all blocks that relate to local accounts
	blocks, err := e.exportBlocks(ctx, localAccounts, f)
	if err != nil {
		return fmt.Errorf("ExportMinimal: error exporting blocks: %s", err)
	}

	// for each block, make sure we've written out the account owning it, or targeted by it --
	// this might include non-local accounts, but we need these so we don't lose anything
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

	// export all follows that relate to local accounts
	follows, err := e.exportFollows(ctx, localAccounts, f)
	if err != nil {
		return fmt.Errorf("ExportMinimal: error exporting follows: %s", err)
	}

	// for each follow, make sure we've written out the account owning it, or targeted by it --
	// this might include non-local accounts, but we need these so we don't lose anything
	for _, follow := range follows {
		_, alreadyWritten := e.writtenIDs[follow.AccountID]
		if !alreadyWritten {
			_, err := e.exportAccounts(ctx, []db.Where{{Key: "id", Value: follow.AccountID}}, f)
			if err != nil {
				return fmt.Errorf("ExportMinimal: error exporting follow owner account: %s", err)
			}
		}

		_, alreadyWritten = e.writtenIDs[follow.TargetAccountID]
		if !alreadyWritten {
			_, err := e.exportAccounts(ctx, []db.Where{{Key: "id", Value: follow.TargetAccountID}}, f)
			if err != nil {
				return fmt.Errorf("ExportMinimal: error exporting follow target account: %s", err)
			}
		}
	}

	// export all follow requests that relate to local accounts
	frs, err := e.exportFollowRequests(ctx, localAccounts, f)
	if err != nil {
		return fmt.Errorf("ExportMinimal: error exporting follow requests: %s", err)
	}

	// for each follow request, make sure we've written out the account owning it, or targeted by it --
	// this might include non-local accounts, but we need these so we don't lose anything
	for _, fr := range frs {
		_, alreadyWritten := e.writtenIDs[fr.AccountID]
		if !alreadyWritten {
			_, err := e.exportAccounts(ctx, []db.Where{{Key: "id", Value: fr.AccountID}}, f)
			if err != nil {
				return fmt.Errorf("ExportMinimal: error exporting follow request owner account: %s", err)
			}
		}

		_, alreadyWritten = e.writtenIDs[fr.TargetAccountID]
		if !alreadyWritten {
			_, err := e.exportAccounts(ctx, []db.Where{{Key: "id", Value: fr.TargetAccountID}}, f)
			if err != nil {
				return fmt.Errorf("ExportMinimal: error exporting follow request target account: %s", err)
			}
		}
	}

	// export all domain blocks
	if _, err := e.exportDomainBlocks(ctx, f); err != nil {
		return fmt.Errorf("ExportMinimal: error exporting domain blocks: %s", err)
	}

	// export all users
	if _, err := e.exportUsers(ctx, f); err != nil {
		return fmt.Errorf("ExportMinimal: error exporting users: %s", err)
	}

	// export all instances
	if _, err := e.exportInstances(ctx, f); err != nil {
		return fmt.Errorf("ExportMinimal: error exporting instances: %s", err)
	}

	return neatClose(f)
}
