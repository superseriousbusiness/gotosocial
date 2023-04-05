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

package fedi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

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
func (p *Processor) OutboxGet(ctx context.Context, requestedUsername string, page bool, maxID string, minID string) (interface{}, gtserror.WithCode) {
	requestedAccount, _, errWithCode := p.authenticate(ctx, requestedUsername)
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
	publicStatuses, err := p.state.DB.GetAccountStatuses(ctx, requestedAccount.ID, 30, true, true, maxID, minID, false, true)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
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
func (p *Processor) FollowersGet(ctx context.Context, requestedUsername string) (interface{}, gtserror.WithCode) {
	requestedAccount, _, errWithCode := p.authenticate(ctx, requestedUsername)
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
func (p *Processor) FollowingGet(ctx context.Context, requestedUsername string) (interface{}, gtserror.WithCode) {
	requestedAccount, _, errWithCode := p.authenticate(ctx, requestedUsername)
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

// FeaturedCollectionGet returns an ordered collection of the requested username's Pinned posts.
// The returned collection have an `items` property which contains an ordered list of status URIs.
func (p *Processor) FeaturedCollectionGet(ctx context.Context, requestedUsername string) (interface{}, gtserror.WithCode) {
	requestedAccount, _, errWithCode := p.authenticate(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode
	}

	statuses, err := p.state.DB.GetAccountPinnedStatuses(ctx, requestedAccount.ID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	collection, err := p.tc.StatusesToASFeaturedCollection(ctx, requestedAccount.FeaturedCollectionURI, statuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.SerializeOrderedCollection(collection)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
