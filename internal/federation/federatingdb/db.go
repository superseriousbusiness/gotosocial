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

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/filter/spam"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// DB wraps the pub.Database interface with
// a couple of custom functions for GoToSocial.
type DB interface {
	// Default functionality.
	pub.Database

	/*
		Overridden functionality for calling from federatingProtocol.
	*/

	Undo(ctx context.Context, undo vocab.ActivityStreamsUndo) error
	Accept(ctx context.Context, accept vocab.ActivityStreamsAccept) error
	Reject(ctx context.Context, reject vocab.ActivityStreamsReject) error
	Announce(ctx context.Context, announce vocab.ActivityStreamsAnnounce) error
	Move(ctx context.Context, move vocab.ActivityStreamsMove) error

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
	spamFilter *spam.Filter
}

// New returns a DB that satisfies the pub.Database
// interface, using the given state and filters.
func New(
	state *state.State,
	converter *typeutils.Converter,
	visFilter *visibility.Filter,
	spamFilter *spam.Filter,
) DB {
	fdb := federatingDB{
		state:      state,
		converter:  converter,
		visFilter:  visFilter,
		spamFilter: spamFilter,
	}
	return &fdb
}
