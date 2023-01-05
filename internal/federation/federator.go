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

package federation

import (
	"context"
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// Federator wraps various interfaces and functions to manage activitypub federation from gotosocial
type Federator interface {
	// FederatingActor returns the underlying pub.FederatingActor, which can be used to send activities, and serve actors at inboxes/outboxes.
	FederatingActor() pub.FederatingActor
	// FederatingDB returns the underlying FederatingDB interface.
	FederatingDB() federatingdb.DB
	// TransportController returns the underlying transport controller.
	TransportController() transport.Controller

	// AuthenticateFederatedRequest can be used to check the authenticity of incoming http-signed requests for federating resources.
	// The given username will be used to create a transport for making outgoing requests. See the implementation for more detailed comments.
	//
	// If the request is valid and passes authentication, the URL of the key owner ID will be returned, as well as true, and nil.
	//
	// If the request does not pass authentication, or there's a domain block, nil, false, nil will be returned.
	//
	// If something goes wrong during authentication, nil, false, and an error will be returned.
	AuthenticateFederatedRequest(ctx context.Context, username string) (*url.URL, gtserror.WithCode)

	/*
		dereferencing functions
	*/
	DereferenceRemoteThread(ctx context.Context, username string, statusURI *url.URL, status *gtsmodel.Status, statusable ap.Statusable)
	DereferenceAnnounce(ctx context.Context, announce *gtsmodel.Status, requestingUsername string) error
	GetAccount(ctx context.Context, params dereferencing.GetAccountParams) (*gtsmodel.Account, error)
	GetStatus(ctx context.Context, username string, remoteStatusID *url.URL, refetch, includeParent bool) (*gtsmodel.Status, ap.Statusable, error)
	EnrichRemoteStatus(ctx context.Context, username string, status *gtsmodel.Status, includeParent bool) (*gtsmodel.Status, error)
	GetRemoteInstance(ctx context.Context, username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error)

	// Handshaking returns true if the given username is currently in the process of dereferencing the remoteAccountID.
	Handshaking(ctx context.Context, username string, remoteAccountID *url.URL) bool
	pub.CommonBehavior
	pub.FederatingProtocol
}

type federator struct {
	db                  db.DB
	federatingDB        federatingdb.DB
	clock               pub.Clock
	typeConverter       typeutils.TypeConverter
	transportController transport.Controller
	dereferencer        dereferencing.Dereferencer
	mediaManager        media.Manager
	actor               pub.FederatingActor
}

// NewFederator returns a new federator
func NewFederator(db db.DB, federatingDB federatingdb.DB, transportController transport.Controller, typeConverter typeutils.TypeConverter, mediaManager media.Manager) Federator {
	dereferencer := dereferencing.NewDereferencer(db, typeConverter, transportController, mediaManager)

	clock := &Clock{}
	f := &federator{
		db:                  db,
		federatingDB:        federatingDB,
		clock:               &Clock{},
		typeConverter:       typeConverter,
		transportController: transportController,
		dereferencer:        dereferencer,
		mediaManager:        mediaManager,
	}
	actor := newFederatingActor(f, f, federatingDB, clock)
	f.actor = actor
	return f
}

func (f *federator) FederatingActor() pub.FederatingActor {
	return f.actor
}

func (f *federator) FederatingDB() federatingdb.DB {
	return f.federatingDB
}

func (f *federator) TransportController() transport.Controller {
	return f.transportController
}
