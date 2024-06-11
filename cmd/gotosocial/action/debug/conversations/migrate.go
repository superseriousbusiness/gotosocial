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

package conversations

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/processing/stream"
	"github.com/superseriousbusiness/gotosocial/internal/processing/workers"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

func initState(ctx context.Context) (*state.State, error) {
	var state state.State
	state.Caches.Init()
	state.Caches.Start()

	// Set the state DB connection
	dbConn, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return nil, fmt.Errorf("error creating dbConn: %w", err)
	}
	state.DB = dbConn

	return &state, nil
}

func stopState(state *state.State) error {
	err := state.DB.Close()
	state.Caches.Stop()
	return err
}

// Migrate processes every DM to create conversations.
var Migrate action.GTSAction = func(ctx context.Context) (err error) {
	state, err := initState(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure state gets stopped on return.
		if err := stopState(state); err != nil {
			log.Error(ctx, err)
		}
	}()

	streamProcessor := stream.New(state, oauth.New(ctx, state.DB))
	surface := workers.Surface{
		State:     state,
		Converter: typeutils.NewConverter(state),
		Stream:    &streamProcessor,
		Filter:    visibility.NewFilter(state),
	}
	if surface.EmailSender, err = email.NewNoopSender(func(toAddress string, message string) {}); err != nil {
		return nil
	}

	return state.DB.MigrateConversations(ctx, func(ctx context.Context, status *gtsmodel.Status) error {
		return surface.UpdateConversationsForStatus(ctx, status, false)
	})
}
