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

package federatingdb

import (
	"context"

	"codeberg.org/gruf/go-mutexes"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// DB wraps the pub.Database interface with a couple of custom functions for GoToSocial.
type DB interface {
	pub.Database
	Undo(ctx context.Context, undo vocab.ActivityStreamsUndo) error
	Accept(ctx context.Context, accept vocab.ActivityStreamsAccept) error
	Reject(ctx context.Context, reject vocab.ActivityStreamsReject) error
	Announce(ctx context.Context, announce vocab.ActivityStreamsAnnounce) error
}

// FederatingDB uses the underlying DB interface to implement the go-fed pub.Database interface.
// It doesn't care what the underlying implementation of the DB interface is, as long as it works.
type federatingDB struct {
	locks         mutexes.MutexMap
	state         *state.State
	typeConverter typeutils.TypeConverter
}

// New returns a DB interface using the given database and config
func New(state *state.State, tc typeutils.TypeConverter) DB {
	fdb := federatingDB{
		locks:         mutexes.NewMap(-1, -1), // use defaults
		state:         state,
		typeConverter: tc,
	}
	return &fdb
}
