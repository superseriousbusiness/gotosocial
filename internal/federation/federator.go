/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"sync"

	"github.com/go-fed/activity/pub"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
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

	// AuthenticateFederatedRequest can be used to check the authenticity of incoming http-signed requests for federating resources.
	// The given username will be used to create a transport for making outgoing requests. See the implementation for more detailed comments.
	//
	// If the request is valid and passes authentication, the URL of the key owner ID will be returned, as well as true, and nil.
	//
	// If the request does not pass authentication, or there's a domain block, nil, false, nil will be returned.
	//
	// If something goes wrong during authentication, nil, false, and an error will be returned.
	AuthenticateFederatedRequest(ctx context.Context, username string) (*url.URL, bool, error)

	// FingerRemoteAccount performs a webfinger lookup for a remote account, using the .well-known path. It will return the ActivityPub URI for that
	// account, or an error if it doesn't exist or can't be retrieved.
	FingerRemoteAccount(requestingUsername string, targetUsername string, targetDomain string) (*url.URL, error)
	// DereferenceRemoteAccount can be used to get the representation of a remote account, based on the account ID (which is a URI).
	// The given username will be used to create a transport for making outgoing requests. See the implementation for more detailed comments.
	DereferenceRemoteAccount(username string, remoteAccountID *url.URL) (typeutils.Accountable, error)
	// DereferenceRemoteStatus can be used to get the representation of a remote status, based on its ID (which is a URI).
	// The given username will be used to create a transport for making outgoing requests. See the implementation for more detailed comments.
	DereferenceRemoteStatus(username string, remoteStatusID *url.URL) (typeutils.Statusable, error)
	// DereferenceRemoteInstance takes the URL of a remote instance, and a username (optional) to spin up a transport with. It then
	// does its damnedest to get some kind of information back about the instance, trying /api/v1/instance, then /.well-known/nodeinfo
	DereferenceRemoteInstance(username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error)
	// DereferenceRemoteThread takes a statusable (something that has withReplies and withInReplyTo),
	// and dereferences statusables in the conversation, putting them in the database.
	//
	// This process involves working up and down the chain of replies, and parsing through the collections of IDs
	// presented by remote instances as part of their replies collections, and will likely involve making several calls to
	// multiple different hosts.
	DereferenceRemoteThread(username string, statusURI *url.URL) error
	// DereferenceCollectionPage returns the activitystreams CollectionPage at the specified IRI, or an error if something goes wrong.
	DereferenceCollectionPage(username string, pageIRI *url.URL) (typeutils.CollectionPageable, error)

	// DereferenceStatusFields does further dereferencing on a status.
	DereferenceStatusFields(status *gtsmodel.Status, requestingUsername string) error
	// DereferenceAccountFields does further dereferencing on an account.
	DereferenceAccountFields(account *gtsmodel.Account, requestingUsername string, refresh bool) error
	// DereferenceAnnounce does further dereferencing on an announce.
	DereferenceAnnounce(announce *gtsmodel.Status, requestingUsername string) error

	// Handshaking returns true if the given username is currently in the process of dereferencing the remoteAccountID.
	Handshaking(username string, remoteAccountID *url.URL) bool
	pub.CommonBehavior
	pub.FederatingProtocol
}

type federator struct {
	config              *config.Config
	db                  db.DB
	federatingDB        federatingdb.DB
	clock               pub.Clock
	typeConverter       typeutils.TypeConverter
	transportController transport.Controller
	dereferencer        dereferencing.Dereferencer
	mediaHandler        media.Handler
	actor               pub.FederatingActor
	log                 *logrus.Logger
	handshakes          map[string][]*url.URL
	handshakeSync       *sync.Mutex // mutex to lock/unlock when checking or updating the handshakes map
}

// NewFederator returns a new federator
func NewFederator(db db.DB, federatingDB federatingdb.DB, transportController transport.Controller, config *config.Config, log *logrus.Logger, typeConverter typeutils.TypeConverter, mediaHandler media.Handler) Federator {

	dereferencer := dereferencing.NewDereferencer(config, db, typeConverter, transportController, mediaHandler, log)

	clock := &Clock{}
	f := &federator{
		config:              config,
		db:                  db,
		federatingDB:        federatingDB,
		clock:               &Clock{},
		typeConverter:       typeConverter,
		transportController: transportController,
		dereferencer:        dereferencer,
		mediaHandler:        mediaHandler,
		log:                 log,
		handshakeSync:       &sync.Mutex{},
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
