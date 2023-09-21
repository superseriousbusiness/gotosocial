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
	"net/http"
	"net/url"

	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
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
		collection, err := p.converter.OutboxToASCollection(ctx, requestedAccount.OutboxURI)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		data, err = ap.Serialize(collection)
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

	outboxPage, err := p.converter.StatusesToASOutboxPage(ctx, requestedAccount.OutboxURI, maxID, minID, publicStatuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	data, err = ap.Serialize(outboxPage)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

// FollowersGet handles the getting of a fedi/activitypub representation of a user/account's followers, performing appropriate
// authentication before returning a JSON serializable interface to the caller.
func (p *Processor) FollowersGet(ctx context.Context, requestedUsername string, page *paging.Page) (interface{}, gtserror.WithCode) {
	requestedAccount, _, errWithCode := p.authenticate(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Parse the collection ID object from account's followers URI.
	collectionID, err := url.Parse(requestedAccount.FollowersURI)
	if err != nil {
		err := gtserror.Newf("error parsing account followers uri %s: %w", requestedAccount.FollowersURI, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Calculate total number of followers available for account.
	total, err := p.state.DB.CountAccountFollowers(ctx, requestedAccount.ID)
	if err != nil {
		err := gtserror.Newf("error counting followers: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	var obj vocab.Type

	// Start building AS collection params.
	var params ap.CollectionParams
	params.ID = collectionID
	params.Total = total

	if page == nil {
		// i.e. paging disabled, the simplest case.
		//
		// Just build collection object from params.
		obj = ap.NewASOrderedCollection(params)
	} else {
		// i.e. paging enabled

		// Get the request page of full follower objects with attached accounts.
		followers, err := p.state.DB.GetAccountFollowers(ctx, requestedAccount.ID, page)
		if err != nil {
			err := gtserror.Newf("error getting followers: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Get the lowest and highest
		// ID values, used for paging.
		lo := followers[len(followers)-1].ID
		hi := followers[0].ID

		// Start building AS collection page params.
		var pageParams ap.CollectionPageParams
		pageParams.CollectionParams = params

		// Current page details.
		pageParams.Current = page
		pageParams.Count = len(followers)

		// Set linked next/prev parameters.
		pageParams.Next = page.Next(lo, hi)
		pageParams.Prev = page.Prev(lo, hi)

		// Set the collection item property builder function.
		pageParams.Append = func(i int, itemsProp ap.ItemsPropertyBuilder) {
			// Get follower URI at index.
			follow := followers[i]
			accURI := follow.Account.URI

			// Parse URL object from URI.
			iri, err := url.Parse(accURI)
			if err != nil {
				log.Errorf(ctx, "error parsing account uri %s: %v", accURI, err)
				return
			}

			// Add to item property.
			itemsProp.AppendIRI(iri)
		}

		// Build AS collection page object from params.
		obj = ap.NewASOrderedCollectionPage(pageParams)
	}

	// Serialized the prepared object.
	data, err := ap.Serialize(obj)
	if err != nil {
		err := gtserror.Newf("error serializing: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

// FollowingGet handles the getting of a fedi/activitypub representation of a user/account's following, performing appropriate
// authentication before returning a JSON serializable interface to the caller.
func (p *Processor) FollowingGet(ctx context.Context, requestedUsername string, page *paging.Page) (interface{}, gtserror.WithCode) {
	requestedAccount, _, errWithCode := p.authenticate(ctx, requestedUsername)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Parse the collection ID object from account's following URI.
	collectionID, err := url.Parse(requestedAccount.FollowingURI)
	if err != nil {
		err := gtserror.Newf("error parsing account following uri %s: %w", requestedAccount.FollowingURI, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Calculate total number of following available for account.
	total, err := p.state.DB.CountAccountFollows(ctx, requestedAccount.ID)
	if err != nil {
		err := gtserror.Newf("error counting follows: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	var obj vocab.Type

	// Start building AS collection params.
	var params ap.CollectionParams
	params.ID = collectionID
	params.Total = total

	if page == nil {
		// i.e. paging disabled, the simplest case.
		//
		// Just build collection object from params.
		obj = ap.NewASOrderedCollection(params)
	} else {
		// i.e. paging enabled

		// Get the request page of full follower objects with attached accounts.
		follows, err := p.state.DB.GetAccountFollows(ctx, requestedAccount.ID, page)
		if err != nil {
			err := gtserror.Newf("error getting follows: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Get the lowest and highest
		// ID values, used for paging.
		lo := follows[len(follows)-1].ID
		hi := follows[0].ID

		// Start building AS collection page params.
		var pageParams ap.CollectionPageParams
		pageParams.CollectionParams = params

		// Current page details.
		pageParams.Current = page
		pageParams.Count = len(follows)

		// Set linked next/prev parameters.
		pageParams.Next = page.Next(lo, hi)
		pageParams.Prev = page.Prev(lo, hi)

		// Set the collection item property builder function.
		pageParams.Append = func(i int, itemsProp ap.ItemsPropertyBuilder) {
			// Get follower URI at index.
			follow := follows[i]
			accURI := follow.Account.URI

			// Parse URL object from URI.
			iri, err := url.Parse(accURI)
			if err != nil {
				log.Errorf(ctx, "error parsing account uri %s: %v", accURI, err)
				return
			}

			// Add to item property.
			itemsProp.AppendIRI(iri)
		}

		// Build AS collection page object from params.
		obj = ap.NewASOrderedCollectionPage(pageParams)
	}

	// Serialized the prepared object.
	data, err := ap.Serialize(obj)
	if err != nil {
		err := gtserror.Newf("error serializing: %w", err)
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

	collection, err := p.converter.StatusesToASFeaturedCollection(ctx, requestedAccount.FeaturedCollectionURI, statuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.Serialize(collection)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
