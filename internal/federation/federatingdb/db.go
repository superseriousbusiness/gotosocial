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
	"net/url"

	"codeberg.org/gruf/go-cache/v3/simple"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/filter/interaction"
	"github.com/superseriousbusiness/gotosocial/internal/filter/spam"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// DB wraps the pub.Database interface with
// a couple of custom functions for GoToSocial.
type DB interface {
	// Default
	// functionality.
	pub.Database

	// Federating protocol overridden callback functionality.
	Like(context.Context, vocab.ActivityStreamsLike) error
	Block(context.Context, vocab.ActivityStreamsBlock) error
	Follow(context.Context, vocab.ActivityStreamsFollow) error
	Undo(context.Context, vocab.ActivityStreamsUndo) error
	Accept(context.Context, vocab.ActivityStreamsAccept) error
	Reject(context.Context, vocab.ActivityStreamsReject) error
	Announce(context.Context, vocab.ActivityStreamsAnnounce) error
	Move(context.Context, vocab.ActivityStreamsMove) error
	Flag(context.Context, vocab.ActivityStreamsFlag) error

	/*
		Extra/convenience functionality.
	*/

	GetAccept(ctx context.Context, acceptIRI *url.URL) (vocab.ActivityStreamsAccept, error)
}

// FederatingDB uses the given state interface
// to implement the go-fed pub.Database interface.
type federatingDB struct {
	state      *state.State
	converter  *typeutils.Converter
	visFilter  *visibility.Filter
	intFilter  *interaction.Filter
	spamFilter *spam.Filter

	// tracks Activity IDs we have handled creates for,
	// for use in the Exists() function during forwarding.
	activityIDs simple.Cache[string, struct{}]
}

// New returns a DB that satisfies the pub.Database
// interface, using the given state and filters.
func New(
	state *state.State,
	converter *typeutils.Converter,
	visFilter *visibility.Filter,
	intFilter *interaction.Filter,
	spamFilter *spam.Filter,
) DB {
	fdb := federatingDB{
		state:      state,
		converter:  converter,
		visFilter:  visFilter,
		intFilter:  intFilter,
		spamFilter: spamFilter,
	}
	fdb.activityIDs.Init(0, 2048)
	return &fdb
}

// storeActivityID stores an entry in the .activityIDs cache for this
// type's JSON-LD ID, for later checks in Exist() to mark it as seen.
func (f *federatingDB) storeActivityID(asType vocab.Type) {
	f.activityIDs.Set(ap.GetJSONLDId(asType).String(), struct{}{})
}
