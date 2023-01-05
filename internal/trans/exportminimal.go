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

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("ExportMinimal: couldn't export to %s: %s", path, err)
	}

	// export all local accounts we have in the database
	localAccounts, err := e.exportAccounts(ctx, []db.Where{{Key: "domain", Value: nil}}, file)
	if err != nil {
		return fmt.Errorf("ExportMinimal: error exporting accounts: %s", err)
	}

	// export all blocks that relate to local accounts
	blocks, err := e.exportBlocks(ctx, localAccounts, file)
	if err != nil {
		return fmt.Errorf("ExportMinimal: error exporting blocks: %s", err)
	}

	// for each block, make sure we've written out the account owning it, or targeted by it --
	// this might include non-local accounts, but we need these so we don't lose anything
	for _, b := range blocks {
		_, alreadyWritten := e.writtenIDs[b.AccountID]
		if !alreadyWritten {
			_, err := e.exportAccounts(ctx, []db.Where{{Key: "id", Value: b.AccountID}}, file)
			if err != nil {
				return fmt.Errorf("ExportMinimal: error exporting block owner account: %s", err)
			}
		}

		_, alreadyWritten = e.writtenIDs[b.TargetAccountID]
		if !alreadyWritten {
			_, err := e.exportAccounts(ctx, []db.Where{{Key: "id", Value: b.TargetAccountID}}, file)
			if err != nil {
				return fmt.Errorf("ExportMinimal: error exporting block target account: %s", err)
			}
		}
	}

	// export all follows that relate to local accounts
	follows, err := e.exportFollows(ctx, localAccounts, file)
	if err != nil {
		return fmt.Errorf("ExportMinimal: error exporting follows: %s", err)
	}

	// for each follow, make sure we've written out the account owning it, or targeted by it --
	// this might include non-local accounts, but we need these so we don't lose anything
	for _, follow := range follows {
		_, alreadyWritten := e.writtenIDs[follow.AccountID]
		if !alreadyWritten {
			_, err := e.exportAccounts(ctx, []db.Where{{Key: "id", Value: follow.AccountID}}, file)
			if err != nil {
				return fmt.Errorf("ExportMinimal: error exporting follow owner account: %s", err)
			}
		}

		_, alreadyWritten = e.writtenIDs[follow.TargetAccountID]
		if !alreadyWritten {
			_, err := e.exportAccounts(ctx, []db.Where{{Key: "id", Value: follow.TargetAccountID}}, file)
			if err != nil {
				return fmt.Errorf("ExportMinimal: error exporting follow target account: %s", err)
			}
		}
	}

	// export all follow requests that relate to local accounts
	followRequests, err := e.exportFollowRequests(ctx, localAccounts, file)
	if err != nil {
		return fmt.Errorf("ExportMinimal: error exporting follow requests: %s", err)
	}

	// for each follow request, make sure we've written out the account owning it, or targeted by it --
	// this might include non-local accounts, but we need these so we don't lose anything
	for _, fr := range followRequests {
		_, alreadyWritten := e.writtenIDs[fr.AccountID]
		if !alreadyWritten {
			_, err := e.exportAccounts(ctx, []db.Where{{Key: "id", Value: fr.AccountID}}, file)
			if err != nil {
				return fmt.Errorf("ExportMinimal: error exporting follow request owner account: %s", err)
			}
		}

		_, alreadyWritten = e.writtenIDs[fr.TargetAccountID]
		if !alreadyWritten {
			_, err := e.exportAccounts(ctx, []db.Where{{Key: "id", Value: fr.TargetAccountID}}, file)
			if err != nil {
				return fmt.Errorf("ExportMinimal: error exporting follow request target account: %s", err)
			}
		}
	}

	// export all domain blocks
	if _, err := e.exportDomainBlocks(ctx, file); err != nil {
		return fmt.Errorf("ExportMinimal: error exporting domain blocks: %s", err)
	}

	// export all users
	if _, err := e.exportUsers(ctx, file); err != nil {
		return fmt.Errorf("ExportMinimal: error exporting users: %s", err)
	}

	// export all instances
	if _, err := e.exportInstances(ctx, file); err != nil {
		return fmt.Errorf("ExportMinimal: error exporting instances: %s", err)
	}

	// export all SUSPENDED accounts to make sure the suspension sticks across db migration etc
	whereSuspended := []db.Where{{
		Key:   "suspended_at",
		Not:   true,
		Value: nil,
	}}
	if _, err := e.exportAccounts(ctx, whereSuspended, file); err != nil {
		return fmt.Errorf("ExportMinimal: error exporting suspended accounts: %s", err)
	}

	return neatClose(file)
}
