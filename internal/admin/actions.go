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
	"slices"
	"sync"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/workers"
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
	// Map of running actions.
	running map[string]*gtsmodel.AdminAction

	// Lock for running admin actions.
	//
	// Not embedded struct, to shield
	// from access by outside packages.
	m sync.Mutex

	// DB for storing, updating,
	// deleting admin actions etc.
	db db.DB

	// Workers for queuing
	// admin action side effects.
	workers *workers.Workers
}

func New(db db.DB, workers *workers.Workers) *Actions {
	return &Actions{
		running: make(map[string]*gtsmodel.AdminAction),
		db:      db,
		workers: workers,
	}
}

type ActionF func(context.Context) gtserror.MultiError

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
	adminAction *gtsmodel.AdminAction,
	f ActionF,
) gtserror.WithCode {
	actionKey := adminAction.Key()

	// LOCK THE MAP HERE, since we're
	// going to do some operations on it.
	a.m.Lock()

	// Bail if an action with
	// this key is already running.
	running, ok := a.running[actionKey]
	if ok {
		a.m.Unlock()
		return errActionConflict(running)
	}

	// Action with this key not
	// yet running, create it.
	if err := a.db.PutAdminAction(ctx, adminAction); err != nil {
		err = gtserror.Newf("db error putting admin action %s: %w", actionKey, err)

		// Don't store in map
		// if there's an error.
		a.m.Unlock()
		return gtserror.NewErrorInternalError(err)
	}

	// Action was inserted,
	// store in map.
	a.running[actionKey] = adminAction

	// UNLOCK THE MAP HERE, since
	// we're done modifying it for now.
	a.m.Unlock()

	go func() {
		// Use a background context with existing values.
		ctx = gtscontext.WithValues(context.Background(), ctx)

		// Run the thing and collect errors.
		if errs := f(ctx); errs != nil {
			adminAction.Errors = make([]string, 0, len(errs))
			for _, err := range errs {
				adminAction.Errors = append(adminAction.Errors, err.Error())
			}
		}

		// Action is no longer running:
		// remove from running map.
		a.m.Lock()
		delete(a.running, actionKey)
		a.m.Unlock()

		// Mark as completed in the db,
		// storing errors for later review.
		adminAction.CompletedAt = time.Now()
		if err := a.db.UpdateAdminAction(ctx, adminAction, "completed_at", "errors"); err != nil {
			log.Errorf(ctx, "db error marking action %s as completed: %q", actionKey, err)
		}
	}()

	return nil
}

// GetRunning sounds like a threat, but it actually just
// returns all of the currently running actions held by
// the Actions struct, ordered by ID descending.
func (a *Actions) GetRunning() []*gtsmodel.AdminAction {
	a.m.Lock()
	defer a.m.Unlock()

	// Assemble all currently running actions.
	running := make([]*gtsmodel.AdminAction, 0, len(a.running))
	for _, action := range a.running {
		running = append(running, action)
	}

	// Order by ID descending (creation date).
	slices.SortFunc(
		running,
		func(a *gtsmodel.AdminAction, b *gtsmodel.AdminAction) int {
			const k = -1
			switch {
			case a.ID > b.ID:
				return +k
			case a.ID < b.ID:
				return -k
			default:
				return 0
			}
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

	return len(a.running)
}
