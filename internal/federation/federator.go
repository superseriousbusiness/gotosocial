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
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
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
	AuthenticateFederatedRequest(username string, r *http.Request) (*url.URL, error)
	// DereferenceRemoteAccount can be used to get the representation of a remote account, based on the account ID (which is a URI).
	// The given username will be used to create a transport for making outgoing requests. See the implementation for more detailed comments.
	DereferenceRemoteAccount(username string, remoteAccountID *url.URL) (typeutils.Accountable, error)
	// DereferenceRemoteStatus can be used to get the representation of a remote status, based on its ID (which is a URI).
	// The given username will be used to create a transport for making outgoing requests. See the implementation for more detailed comments.
	DereferenceRemoteStatus(username string, remoteStatusID *url.URL) (typeutils.Statusable, error)
	// GetTransportForUser returns a new transport initialized with the key credentials belonging to the given username.
	// This can be used for making signed http requests.
	//
	// If username is an empty string, our instance user's credentials will be used instead.
	GetTransportForUser(username string) (transport.Transport, error)
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
	actor               pub.FederatingActor
	log                 *logrus.Logger
}

// NewFederator returns a new federator
func NewFederator(db db.DB, federatingDB federatingdb.DB, transportController transport.Controller, config *config.Config, log *logrus.Logger, typeConverter typeutils.TypeConverter) Federator {

	clock := &Clock{}
	f := &federator{
		config:              config,
		db:                  db,
		federatingDB:        federatingDB,
		clock:               &Clock{},
		typeConverter:       typeConverter,
		transportController: transportController,
		log:                 log,
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
