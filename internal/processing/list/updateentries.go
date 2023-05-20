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

package list

import (
	"context"
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

// AddToList adds targetAccountIDs to the given list, if valid.
func (p *Processor) AddToList(ctx context.Context, account *gtsmodel.Account, listID string, targetAccountIDs []string) gtserror.WithCode {
	// Ensure this list exists + account owns it.
	list, errWithCode := p.getList(ctx, account.ID, listID)
	if errWithCode != nil {
		return errWithCode
	}

	// Pre-assemble list of entries to add. We *could* add these
	// one by one as we iterate through accountIDs, but according
	// to the Mastodon API we should only add them all once we know
	// they're all valid, no partial updates.
	listEntries := make([]*gtsmodel.ListEntry, 0, len(targetAccountIDs))

	// Check each targetAccountID is valid.
	//   - Follow must exist.
	//   - Follow must not already be in the given list.
	for _, targetAccountID := range targetAccountIDs {
		// Ensure follow exists.
		follow, err := p.state.DB.GetFollow(ctx, account.ID, targetAccountID)
		if err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				err = fmt.Errorf("you do not follow account %s", targetAccountID)
				return gtserror.NewErrorNotFound(err, err.Error())
			}
			return gtserror.NewErrorInternalError(err)
		}

		// Ensure followID not already in list.
		// This particular call to isInList will
		// never error, so just check entryID.
		entryID, _ := isInList(
			list,
			follow.ID,
			func(listEntry *gtsmodel.ListEntry) (string, error) {
				// Looking for the listEntry follow ID.
				return listEntry.FollowID, nil
			},
		)

		// Empty entryID means entry with given
		// followID wasn't found in the list.
		if entryID != "" {
			err = fmt.Errorf("account with id %s is already in list %s with entryID %s", targetAccountID, listID, entryID)
			return gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}

		// Entry wasn't in the list, we can add it.
		listEntries = append(listEntries, &gtsmodel.ListEntry{
			ID:       id.NewULID(),
			ListID:   listID,
			FollowID: follow.ID,
		})
	}

	// If we get to here we can assume all
	// entries are valid, so try to add them.
	if err := p.state.DB.PutListEntries(ctx, listEntries); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			err = fmt.Errorf("one or more errors inserting list entries: %w", err)
			return gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}

// RemoveFromList removes targetAccountIDs from the given list, if valid.
func (p *Processor) RemoveFromList(ctx context.Context, account *gtsmodel.Account, listID string, targetAccountIDs []string) gtserror.WithCode {
	// Ensure this list exists + account owns it.
	list, errWithCode := p.getList(ctx, account.ID, listID)
	if errWithCode != nil {
		return errWithCode
	}

	// For each targetAccountID, we want to check if
	// a follow with that targetAccountID is in the
	// given list. If it is in there, we want to remove
	// it from the list.
	for _, targetAccountID := range targetAccountIDs {
		// Check if targetAccountID is
		// on a follow in the list.
		entryID, err := isInList(
			list,
			targetAccountID,
			func(listEntry *gtsmodel.ListEntry) (string, error) {
				// We need the follow so populate this
				// entry, if it's not already populated.
				if err := p.state.DB.PopulateListEntry(ctx, listEntry); err != nil {
					return "", err
				}

				// Looking for the list entry targetAccountID.
				return listEntry.Follow.TargetAccountID, nil
			},
		)

		// Error may be returned here if there was an issue
		// populating the list entry. We only return on proper
		// DB errors, we can just skip no entry errors.
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("error checking if targetAccountID %s was in list %s: %w", targetAccountID, listID, err)
			return gtserror.NewErrorInternalError(err)
		}

		if entryID == "" {
			// There was an errNoEntries or targetAccount
			// wasn't in this list anyway, so we can skip it.
			continue
		}

		// TargetAccount was in the list, remove the entry.
		if err := p.state.DB.DeleteListEntry(ctx, entryID); err != nil && !errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("error removing list entry %s from list %s: %w", entryID, listID, err)
			return gtserror.NewErrorInternalError(err)
		}
	}

	return nil
}
