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

package fedi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

func (p *Processor) authenticateCollectionsRequest(ctx context.Context, requestedUsername string) (requestedAccount, requestingAccount *gtsmodel.Account, errWithCode gtserror.WithCode) {
	requestedAccount, err := p.db.GetAccountByUsernameDomain(ctx, requestedUsername, "")
	if err != nil {
		errWithCode = gtserror.NewErrorNotFound(fmt.Errorf("database error getting account with username %s: %s", requestedUsername, err))
		return
	}

	var requestingAccountURI *url.URL
	requestingAccountURI, errWithCode = p.federator.AuthenticateFederatedRequest(ctx, requestedUsername)
	if errWithCode != nil {
		return
	}

	if requestingAccount, err = p.federator.GetAccountByURI(transport.WithFastfail(ctx), requestedUsername, requestingAccountURI, false); err != nil {
		errWithCode = gtserror.NewErrorUnauthorized(err)
		return
	}

	blocked, err := p.db.IsBlocked(ctx, requestedAccount.ID, requestingAccount.ID, true)
	if err != nil {
		errWithCode = gtserror.NewErrorInternalError(err)
		return
	}

	if blocked {
		errWithCode = gtserror.NewErrorUnauthorized(fmt.Errorf("block exists between accounts %s and %s", requestedAccount.ID, requestingAccount.ID))
	}

	return
}

// InboxPost handles POST requests to a user's inbox for new activitypub messages.
//
// InboxPost returns true if the request was handled as an ActivityPub POST to an actor's inbox.
// If false, the request was not an ActivityPub request and may still be handled by the caller in another way, such as serving a web page.
//
// If the error is nil, then the ResponseWriter's headers and response has already been written. If a non-nil error is returned, then no response has been written.
//
// If the Actor was constructed with the Federated Protocol enabled, side effects will occur.
//
// If the Federated Protocol is not enabled, writes the http.StatusMethodNotAllowed status code in the response. No side effects occur.
func (p *Processor) InboxPost(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return p.federator.FederatingActor().PostInbox(ctx, w, r)
}

// OutboxGet returns the activitypub representation of a local user's outbox.
// This contains links to PUBLIC posts made by this user.
func (p *Processor) OutboxGet(ctx context.Context, requestedUsername string, page bool, maxID string, minID string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	requestedAccount, _, errWithCode := p.authenticateCollectionsRequest(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode
	}

	var data map[string]interface{}
	// There are two scenarios:
	// 1. we're asked for the whole collection and not a page -- we can just return the collection, with no items, but a link to 'first' page.
	// 2. we're asked for a specific page; this can be either the first page or any other page

	if !page {
		/*
			scenario 1: return the collection with no items
			we want something that looks like this:
			{
				"@context": "https://www.w3.org/ns/activitystreams",
				"id": "https://example.org/users/whatever/outbox",
				"type": "OrderedCollection",
				"first": "https://example.org/users/whatever/outbox?page=true",
				"last": "https://example.org/users/whatever/outbox?min_id=0&page=true"
			}
		*/
		collection, err := p.tc.OutboxToASCollection(ctx, requestedAccount.OutboxURI)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		data, err = streams.Serialize(collection)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		return data, nil
	}

	// scenario 2 -- get the requested page
	// limit pages to 30 entries per page
	publicStatuses, err := p.db.GetAccountStatuses(ctx, requestedAccount.ID, 30, true, true, maxID, minID, false, true)
	if err != nil && err != db.ErrNoEntries {
		return nil, gtserror.NewErrorInternalError(err)
	}

	outboxPage, err := p.tc.StatusesToASOutboxPage(ctx, requestedAccount.OutboxURI, maxID, minID, publicStatuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	data, err = streams.Serialize(outboxPage)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

// FollowersGet handles the getting of a fedi/activitypub representation of a user/account's followers, performing appropriate
// authentication before returning a JSON serializable interface to the caller.
func (p *Processor) FollowersGet(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	requestedAccount, _, errWithCode := p.authenticateCollectionsRequest(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode
	}

	requestedAccountURI, err := url.Parse(requestedAccount.URI)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error parsing url %s: %s", requestedAccount.URI, err))
	}

	requestedFollowers, err := p.federator.FederatingDB().Followers(ctx, requestedAccountURI)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error fetching followers for uri %s: %s", requestedAccountURI.String(), err))
	}

	data, err := streams.Serialize(requestedFollowers)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

// FollowingGet handles the getting of a fedi/activitypub representation of a user/account's following, performing appropriate
// authentication before returning a JSON serializable interface to the caller.
func (p *Processor) FollowingGet(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	requestedAccount, _, errWithCode := p.authenticateCollectionsRequest(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode
	}

	requestedAccountURI, err := url.Parse(requestedAccount.URI)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error parsing url %s: %s", requestedAccount.URI, err))
	}

	requestedFollowing, err := p.federator.FederatingDB().Following(ctx, requestedAccountURI)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error fetching following for uri %s: %s", requestedAccountURI.String(), err))
	}

	data, err := streams.Serialize(requestedFollowing)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

func (p *Processor) FeaturedCollectionGet(ctx context.Context, requestedUsername string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	requestedAccount, requestingAccount, errWithCode := p.authenticateCollectionsRequest(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode
	}

	statuses, err := p.db.GetAccountPinnedStatuses(ctx, requestedAccount.ID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	filtered := make([]*gtsmodel.Status, 0, len(statuses))
	for _, s := range statuses {
		visible, err := p.filter.StatusVisible(ctx, s, requestingAccount)
		if err != nil {
			log.WithContext(ctx).Errorf("error checking visibility of status %s for account %s: %v", s.ID, requestingAccount.ID, err)
			continue
		}

		if visible {
			filtered = append(filtered, s)
		}
	}

	
}
