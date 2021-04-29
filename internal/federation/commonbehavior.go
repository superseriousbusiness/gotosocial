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
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// commonBehavior implements the go-fed common behavior interface
type commonBehavior struct {
	db                  db.DB
	log                 *logrus.Logger
	config              *config.Config
	transportController transport.Controller
}

// newCommonBehavior returns an implementation of the pub.CommonBehavior interface that uses the given db, log, config, and transportController
func newCommonBehavior(db db.DB, log *logrus.Logger, config *config.Config, transportController transport.Controller) pub.CommonBehavior {
	return &commonBehavior{
		db:                  db,
		log:                 log,
		config:              config,
		transportController: transportController,
	}
}

/*
	GOFED COMMON BEHAVIOR INTERFACE
	Contains functions required for both the Social API and Federating Protocol.
	It is passed to the library as a dependency injection from the client
	application.
*/

// AuthenticateGetInbox delegates the authentication of a GET to an
// inbox.
//
// Always called, regardless whether the Federated Protocol or Social
// API is enabled.
//
// If an error is returned, it is passed back to the caller of
// GetInbox. In this case, the implementation must not write a
// response to the ResponseWriter as is expected that the client will
// do so when handling the error. The 'authenticated' is ignored.
//
// If no error is returned, but authentication or authorization fails,
// then authenticated must be false and error nil. It is expected that
// the implementation handles writing to the ResponseWriter in this
// case.
//
// Finally, if the authentication and authorization succeeds, then
// authenticated must be true and error nil. The request will continue
// to be processed.
func (c *commonBehavior) AuthenticateGetInbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, bool, error) {
	// IMPLEMENTATION NOTE: For GoToSocial, we serve outboxes and inboxes through
	// the CLIENT API, not through the federation API, so we just do nothing here.
	return nil, false, nil
}

// AuthenticateGetOutbox delegates the authentication of a GET to an
// outbox.
//
// Always called, regardless whether the Federated Protocol or Social
// API is enabled.
//
// If an error is returned, it is passed back to the caller of
// GetOutbox. In this case, the implementation must not write a
// response to the ResponseWriter as is expected that the client will
// do so when handling the error. The 'authenticated' is ignored.
//
// If no error is returned, but authentication or authorization fails,
// then authenticated must be false and error nil. It is expected that
// the implementation handles writing to the ResponseWriter in this
// case.
//
// Finally, if the authentication and authorization succeeds, then
// authenticated must be true and error nil. The request will continue
// to be processed.
func (c *commonBehavior) AuthenticateGetOutbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, bool, error) {
	// IMPLEMENTATION NOTE: For GoToSocial, we serve outboxes and inboxes through
	// the CLIENT API, not through the federation API, so we just do nothing here.
	return nil, false, nil
}

// GetOutbox returns the OrderedCollection inbox of the actor for this
// context. It is up to the implementation to provide the correct
// collection for the kind of authorization given in the request.
//
// AuthenticateGetOutbox will be called prior to this.
//
// Always called, regardless whether the Federated Protocol or Social
// API is enabled.
func (c *commonBehavior) GetOutbox(ctx context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// IMPLEMENTATION NOTE: For GoToSocial, we serve outboxes and inboxes through
	// the CLIENT API, not through the federation API, so we just do nothing here.
	return nil, nil
}

// NewTransport returns a new Transport on behalf of a specific actor.
//
// The actorBoxIRI will be either the inbox or outbox of an actor who is
// attempting to do the dereferencing or delivery. Any authentication
// scheme applied on the request must be based on this actor. The
// request must contain some sort of credential of the user, such as a
// HTTP Signature.
//
// The gofedAgent passed in should be used by the Transport
// implementation in the User-Agent, as well as the application-specific
// user agent string. The gofedAgent will indicate this library's use as
// well as the library's version number.
//
// Any server-wide rate-limiting that needs to occur should happen in a
// Transport implementation. This factory function allows this to be
// created, so peer servers are not DOS'd.
//
// Any retry logic should also be handled by the Transport
// implementation.
//
// Note that the library will not maintain a long-lived pointer to the
// returned Transport so that any private credentials are able to be
// garbage collected.
func (c *commonBehavior) NewTransport(ctx context.Context, actorBoxIRI *url.URL, gofedAgent string) (pub.Transport, error) {

	var username string
	var err error

	if util.IsInboxPath(actorBoxIRI) {
		username, err = util.ParseInboxPath(actorBoxIRI)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse path %s as an inbox: %s", actorBoxIRI.String(), err)
		}
	} else if util.IsOutboxPath(actorBoxIRI) {
		username, err = util.ParseOutboxPath(actorBoxIRI)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse path %s as an outbox: %s", actorBoxIRI.String(), err)
		}
	} else {
		return nil, fmt.Errorf("id %s was neither an inbox path nor an outbox path", actorBoxIRI.String())
	}

	account := &gtsmodel.Account{}
	if err := c.db.GetLocalAccountByUsername(username, account); err != nil {
		return nil, fmt.Errorf("error getting account with username %s from the db: %s", username, err)
	}

	return c.transportController.NewTransport(account.PublicKeyURI, account.PrivateKey)
}
