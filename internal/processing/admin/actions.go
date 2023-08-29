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

package admin

import (
	"context"
	"sync"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"golang.org/x/exp/slices"
)

func errActionConflict(action *gtsmodel.AdminAction) gtserror.WithCode {
	err := gtserror.NewfAt(
		4, // Include caller's function name.
		"an action (%s) is currently running (duration %s) which conflicts with the attempted action",
		action.Key(), time.Since(action.CreatedAt),
	)

	const help = "wait until this action is complete and try again"
	return gtserror.NewErrorConflict(err, err.Error(), help)
}

type Actions struct {
	r     map[string]*gtsmodel.AdminAction
	state *state.State

	// Not embedded struct,
	// to shield from access
	// by outside packages.
	m sync.Mutex
}

// Run runs the given admin action by executing the supplied function.
//
// Run handles locking, action insertion and updating, so you don't have to!
//
// If an action is already running which overlaps/conflicts with the
// given action, an ErrorWithCode 409 will be returned.
//
// If execution of the provided function returns errors, the errors
// will be updated on the provided admin action in the database.
func (a *Actions) Run(
	ctx context.Context,
	action *gtsmodel.AdminAction,
	f func(context.Context) gtserror.MultiError,
) gtserror.WithCode {
	key := action.Key()

	// Check if an action with
	// this key is already running.
	a.m.Lock()
	running, ok := a.r[key]
	a.m.Unlock()

	if ok {
		return errActionConflict(running)
	}

	// Action with this key not
	// yet running, lock it in.
	a.m.Lock()

	if err := a.state.DB.PutAdminAction(ctx, action); err != nil {
		err = gtserror.Newf("db error putting admin action %s: %w", key, err)

		// Don't store in map
		// if there's an error.
		a.m.Unlock()
		return gtserror.NewErrorInternalError(err)
	}

	// Action was inserted,
	// store in map.
	a.r[key] = action
	a.m.Unlock()

	// Do the rest of the work asynchronously.
	a.state.Workers.ClientAPI.Enqueue(func(ctx context.Context) {
		// Run the thing and collect errors.
		if errs := f(ctx); errs != nil {
			action.Errors = make([]string, 0, len(errs))
			for _, err := range errs {
				action.Errors = append(action.Errors, err.Error())
			}
		}

		// Action is no longer running:
		// remove from running map.
		a.m.Lock()
		delete(a.r, key)
		a.m.Unlock()

		// Mark as completed in the db,
		// storing errors for later review.
		action.CompletedAt = time.Now()
		if err := a.state.DB.UpdateAdminAction(ctx, action, "completed_at", "errors"); err != nil {
			log.Errorf(ctx, "db error marking action %s as completed: %q", key, err)
		}
	})

	return nil
}

// GetRunning sounds like a threat, but it actually just
// returns all of the currently running actions held by
// the Actions struct, ordered by ID descending.
func (a *Actions) GetRunning() []*gtsmodel.AdminAction {
	a.m.Lock()
	defer a.m.Unlock()

	// Assemble all currently running actions.
	running := make([]*gtsmodel.AdminAction, 0, len(a.r))
	for _, action := range a.r {
		running = append(running, action)
	}

	// Order by ID descending (creation date).
	slices.SortFunc(
		running,
		func(a *gtsmodel.AdminAction, b *gtsmodel.AdminAction) bool {
			return a.ID > b.ID
		},
	)

	return running
}

// TotalRunning is a sequel to the classic
// 1972 environmental-themed science fiction
// film Silent Running, starring Bruce Dern.
func (a *Actions) TotalRunning() int {
	a.m.Lock()
	defer a.m.Unlock()

	return len(a.r)
}
