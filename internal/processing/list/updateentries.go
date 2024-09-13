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
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// AddToList adds targetAccountIDs to the given list, if valid.
func (p *Processor) AddToList(ctx context.Context, account *gtsmodel.Account, listID string, targetAccountIDs []string) gtserror.WithCode {

	// Ensure this list exists + account owns it.
	_, errWithCode := p.getList(ctx, account.ID, listID)
	if errWithCode != nil {
		return errWithCode
	}

	// Get all follows that are entries in list.
	follows, err := p.state.DB.GetFollowsInList(

		// We only need barebones model.
		gtscontext.SetBarebones(ctx),
		listID,
		nil,
	)
	if err != nil {
		err := gtserror.Newf("error getting list follows: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	// Convert the follows to a hash set containing the target account IDs.
	inFollows := util.ToSetFunc(follows, func(follow *gtsmodel.Follow) string {
		return follow.TargetAccountID
	})

	// Preallocate a slice of expected list entries, we specifically
	// gather and add all the target accounts in one go rather than
	// individually, to ensure we don't end up with partial updates.
	entries := make([]*gtsmodel.ListEntry, 0, len(targetAccountIDs))

	// Iterate all the account IDs in given target list.
	for _, targetAccountID := range targetAccountIDs {

		// Look for follow to target account.
		if inFollows.Has(targetAccountID) {
			text := fmt.Sprintf("account %s is already in list %s", targetAccountID, listID)
			return gtserror.NewErrorUnprocessableEntity(errors.New(text), text)
		}

		// Get the actual follow to target.
		follow, err := p.state.DB.GetFollow(

			// We don't need any sub-models.
			gtscontext.SetBarebones(ctx),
			account.ID,
			targetAccountID,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := gtserror.Newf("db error getting follow: %w", err)
			return gtserror.NewErrorInternalError(err)
		}

		if follow == nil {
			text := fmt.Sprintf("account %s not currently followed", targetAccountID)
			return gtserror.NewErrorNotFound(errors.New(text), text)
		}

		// Generate new entry for this follow in list.
		entries = append(entries, &gtsmodel.ListEntry{
			ID:       id.NewULID(),
			ListID:   listID,
			FollowID: follow.ID,
		})
	}

	// Add all of the gathered list entries to the database.
	switch err := p.state.DB.PutListEntries(ctx, entries); {
	case err == nil:

	case errors.Is(err, db.ErrAlreadyExists):
		err := gtserror.Newf("conflict adding list entry: %w", err)
		return gtserror.NewErrorUnprocessableEntity(err)

	default:
		err := gtserror.Newf("db error inserting list entries: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}

// RemoveFromList removes targetAccountIDs from the given list, if valid.
func (p *Processor) RemoveFromList(
	ctx context.Context,
	account *gtsmodel.Account,
	listID string,
	targetAccountIDs []string,
) gtserror.WithCode {
	// Ensure this list exists + account owns it.
	_, errWithCode := p.getList(ctx, account.ID, listID)
	if errWithCode != nil {
		return errWithCode
	}

	// Get all follows that are entries in list.
	follows, err := p.state.DB.GetFollowsInList(

		// We only need barebones model.
		gtscontext.SetBarebones(ctx),
		listID,
		nil,
	)
	if err != nil {
		err := gtserror.Newf("error getting list follows: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	// Convert the follows to a map keyed by the target account ID.
	followsMap := util.KeyBy(follows, func(follow *gtsmodel.Follow) string {
		return follow.TargetAccountID
	})

	var errs gtserror.MultiError

	// Iterate all the account IDs in given target list.
	for _, targetAccountID := range targetAccountIDs {

		// Look for follow targetting this account.
		follow, ok := followsMap[targetAccountID]

		if !ok {
			// not in list.
			continue
		}

		// Delete the list entry containing follow ID in list.
		err := p.state.DB.DeleteListEntry(ctx, listID, follow.ID)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			errs.Appendf("error removing list entry: %w", err)
			continue
		}
	}

	// Wrap errors in errWithCode if set.
	if err := errs.Combine(); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}
