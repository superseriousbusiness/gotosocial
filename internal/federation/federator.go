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

package federation

import (
	"codeberg.org/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
	"github.com/superseriousbusiness/gotosocial/internal/filter/interaction"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

var _ interface {
	pub.CommonBehavior
	pub.FederatingProtocol
} = (*Federator)(nil)

type Federator struct {
	db           db.DB
	federatingDB federatingdb.DB
	clock        pub.Clock
	converter    *typeutils.Converter
	transport    transport.Controller
	mediaManager *media.Manager
	actor        pub.FederatingActor
	dereferencing.Dereferencer

	// store result of FederatingCallbacks() ahead
	// of time since it's called in every PostInbox().
	wrapped  pub.FederatingWrappedCallbacks
	callback []any
}

// NewFederator returns a new federator instance.
func NewFederator(
	state *state.State,
	federatingDB federatingdb.DB,
	transportController transport.Controller,
	converter *typeutils.Converter,
	visFilter *visibility.Filter,
	intFilter *interaction.Filter,
	mediaManager *media.Manager,
) *Federator {
	clock := &Clock{}
	f := &Federator{
		db:           state.DB,
		federatingDB: federatingDB,
		clock:        clock,
		converter:    converter,
		transport:    transportController,
		mediaManager: mediaManager,

		Dereferencer: dereferencing.NewDereferencer(
			state,
			converter,
			transportController,
			visFilter,
			intFilter,
			mediaManager,
		),

		// prepared response to FederatingCallbacks()
		wrapped: pub.FederatingWrappedCallbacks{

			// OnFollow determines what action to take for this
			// particular callback if a Follow Activity is handled.
			//
			// For our implementation, we always want to do nothing
			// because we have internal logic for handling follows.
			OnFollow: pub.OnFollowDoNothing,
		},
		callback: []any{
			federatingDB.Like,
			federatingDB.Block,
			federatingDB.Follow,
			federatingDB.Undo,
			federatingDB.Accept,
			federatingDB.Reject,
			federatingDB.Announce,
			federatingDB.Move,
			federatingDB.Flag,
		},
	}
	actor := newFederatingActor(f, f, federatingDB, clock)
	f.actor = actor
	return f
}

// FederatingActor returns the underlying pub.FederatingActor, which can be used to send activities, and serve actors at inboxes/outboxes.
func (f *Federator) FederatingActor() pub.FederatingActor {
	return f.actor
}

// FederatingDB returns the underlying FederatingDB interface.
func (f *Federator) FederatingDB() federatingdb.DB {
	return f.federatingDB
}

// TransportController returns the underlying transport controller.
func (f *Federator) TransportController() transport.Controller {
	return f.transport
}
