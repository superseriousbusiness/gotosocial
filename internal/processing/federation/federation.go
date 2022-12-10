/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"net/http"
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

// Processor wraps functions for processing federation API requests.
type Processor interface {
	// GetUser handles the getting of a fedi/activitypub representation of a user/account, performing appropriate authentication
	// before returning a JSON serializable interface to the caller.
	GetUser(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode)

	// GetFollowers handles the getting of a fedi/activitypub representation of a user/account's followers, performing appropriate
	// authentication before returning a JSON serializable interface to the caller.
	GetFollowers(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode)

	// GetFollowing handles the getting of a fedi/activitypub representation of a user/account's following, performing appropriate
	// authentication before returning a JSON serializable interface to the caller.
	GetFollowing(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode)

	// GetStatus handles the getting of a fedi/activitypub representation of a particular status, performing appropriate
	// authentication before returning a JSON serializable interface to the caller.
	GetStatus(ctx context.Context, requestedUsername string, requestedStatusID string, requestURL *url.URL) (interface{}, gtserror.WithCode)

	// GetStatus handles the getting of a fedi/activitypub representation of replies to a status, performing appropriate
	// authentication before returning a JSON serializable interface to the caller.
	GetStatusReplies(ctx context.Context, requestedUsername string, requestedStatusID string, page bool, onlyOtherAccounts bool, minID string, requestURL *url.URL) (interface{}, gtserror.WithCode)

	// GetWebfingerAccount handles the GET for a webfinger resource. Most commonly, it will be used for returning account lookups.
	GetWebfingerAccount(ctx context.Context, requestedUsername string) (*apimodel.WellKnownResponse, gtserror.WithCode)

	// GetFediEmoji handles the GET for a federated emoji originating from this instance.
	GetEmoji(ctx context.Context, requestedEmojiID string, requestURL *url.URL) (interface{}, gtserror.WithCode)

	// GetNodeInfoRel returns a well known response giving the path to node info.
	GetNodeInfoRel(ctx context.Context) (*apimodel.WellKnownResponse, gtserror.WithCode)

	// GetNodeInfo returns a node info struct in response to a node info request.
	GetNodeInfo(ctx context.Context) (*apimodel.Nodeinfo, gtserror.WithCode)

	// GetOutbox returns the activitypub representation of a local user's outbox.
	// This contains links to PUBLIC posts made by this user.
	GetOutbox(ctx context.Context, requestedUsername string, page bool, maxID string, minID string, requestURL *url.URL) (interface{}, gtserror.WithCode)

	// PostInbox handles POST requests to a user's inbox for new activitypub messages.
	//
	// PostInbox returns true if the request was handled as an ActivityPub POST to an actor's inbox.
	// If false, the request was not an ActivityPub request and may still be handled by the caller in another way, such as serving a web page.
	//
	// If the error is nil, then the ResponseWriter's headers and response has already been written. If a non-nil error is returned, then no response has been written.
	//
	// If the Actor was constructed with the Federated Protocol enabled, side effects will occur.
	//
	// If the Federated Protocol is not enabled, writes the http.StatusMethodNotAllowed status code in the response. No side effects occur.
	PostInbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, error)
}

type processor struct {
	db        db.DB
	federator federation.Federator
	tc        typeutils.TypeConverter
	filter    visibility.Filter
}

// New returns a new federation processor.
func New(db db.DB, tc typeutils.TypeConverter, federator federation.Federator) Processor {
	return &processor{
		db:        db,
		federator: federator,
		tc:        tc,
		filter:    visibility.NewFilter(db),
	}
}
