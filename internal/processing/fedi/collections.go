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

	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/util"
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

// OutboxGet returns the serialized ActivityPub
// collection of a local account's outbox, which
// contains links to PUBLIC posts by this account.
func (p *Processor) OutboxGet(
	ctx context.Context,
	requestedUser string,
	page *paging.Page,
) (interface{}, gtserror.WithCode) {
	// Authenticate incoming request, getting related accounts.
	auth, errWithCode := p.authenticate(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}
	receivingAcct := auth.receivingAcct

	// Parse the collection ID object from account's followers URI.
	collectionID, err := url.Parse(receivingAcct.OutboxURI)
	if err != nil {
		err := gtserror.Newf("error parsing account outbox uri %s: %w", receivingAcct.OutboxURI, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Ensure we have stats for this account.
	if err := p.state.DB.PopulateAccountStats(ctx, receivingAcct); err != nil {
		err := gtserror.Newf("error getting stats for account %s: %w", receivingAcct.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	var obj vocab.Type

	// Start the AS collection params.
	var params ap.CollectionParams
	params.ID = collectionID

	switch {

	case receivingAcct.IsInstance() ||
		*receivingAcct.Settings.HideCollections:
		// If account that hides collections, or instance
		// account (ie., can't post / have relationships),
		// just return barest stub of collection.
		obj = ap.NewASOrderedCollection(params)

	case page == nil || auth.handshakingURI != nil:
		// If paging disabled, or we're currently handshaking
		// the requester, just return collection that links
		// to first page (i.e. path below), with no items.
		params.Total = util.Ptr(*receivingAcct.Stats.StatusesCount)
		params.First = new(paging.Page)
		params.Query = make(url.Values, 1)
		params.Query.Set("limit", "40") // enables paging
		obj = ap.NewASOrderedCollection(params)

	default:
		// Paging enabled.
		// Get page of full public statuses.
		statuses, err := p.state.DB.GetAccountStatuses(
			ctx,
			receivingAcct.ID,
			page.GetLimit(), // limit
			true,            // excludeReplies
			true,            // excludeReblogs
			page.GetMax(),   // maxID
			page.GetMin(),   // minID
			false,           // mediaOnly
			true,            // publicOnly
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := gtserror.Newf("error getting statuses: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// page ID values.
		var lo, hi string

		if len(statuses) > 0 {
			// Get the lowest and highest
			// ID values, used for paging.
			lo = statuses[len(statuses)-1].ID
			hi = statuses[0].ID
		}

		// Reslice statuses dropping all those invisible to requester
		// (eg., local-only statuses, if the requester is remote).
		statuses, err = p.visFilter.StatusesVisible(
			ctx,
			auth.requestingAcct,
			statuses,
		)
		if err != nil {
			err := gtserror.Newf("error filtering statuses: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Start building AS collection page params.
		params.Total = util.Ptr(*receivingAcct.Stats.StatusesCount)
		var pageParams ap.CollectionPageParams
		pageParams.CollectionParams = params

		// Current page details.
		pageParams.Current = page
		pageParams.Count = len(statuses)

		// Set linked next/prev parameters.
		pageParams.Next = page.Next(lo, hi)
		pageParams.Prev = page.Prev(lo, hi)

		// Set the collection item property builder function.
		pageParams.Append = func(i int, itemsProp ap.ItemsPropertyBuilder) {
			// Get status at index.
			status := statuses[i]

			// Derive statusable from status.
			statusable, err := p.converter.StatusToAS(ctx, status)
			if err != nil {
				log.Errorf(ctx, "error converting %s to statusable: %v", status.URI, err)
				return
			}

			// Derive create from statusable, using the IRI only.
			create := typeutils.WrapStatusableInCreate(statusable, true)

			// Add to item property.
			itemsProp.AppendActivityStreamsCreate(create)
		}

		// Build AS collection page object from params.
		obj = ap.NewASOrderedCollectionPage(pageParams)
	}

	// Serialize the prepared object.
	data, err := ap.Serialize(obj)
	if err != nil {
		err := gtserror.Newf("error serializing: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

// FollowersGet returns the serialized ActivityPub
// collection of a local account's followers collection,
// which contains links to accounts following this account.
func (p *Processor) FollowersGet(
	ctx context.Context,
	requestedUser string,
	page *paging.Page,
) (interface{}, gtserror.WithCode) {
	// Authenticate incoming request, getting related accounts.
	auth, errWithCode := p.authenticate(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}
	receivingAcct := auth.receivingAcct

	// Parse the collection ID object from account's followers URI.
	collectionID, err := url.Parse(receivingAcct.FollowersURI)
	if err != nil {
		err := gtserror.Newf("error parsing account followers uri %s: %w", receivingAcct.FollowersURI, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Ensure we have stats for this account.
	if err := p.state.DB.PopulateAccountStats(ctx, receivingAcct); err != nil {
		err := gtserror.Newf("error getting stats for account %s: %w", receivingAcct.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	var obj vocab.Type

	// Start the AS collection params.
	var params ap.CollectionParams
	params.ID = collectionID

	switch {

	case receivingAcct.IsInstance() ||
		*receivingAcct.Settings.HideCollections:
		// If account that hides collections, or instance
		// account (ie., can't post / have relationships),
		// just return barest stub of collection.
		obj = ap.NewASOrderedCollection(params)

	case page == nil || auth.handshakingURI != nil:
		// If paging disabled, or we're currently handshaking
		// the requester, just return collection that links
		// to first page (i.e. path below), with no items.
		params.Total = util.Ptr(*receivingAcct.Stats.FollowersCount)
		params.First = new(paging.Page)
		params.Query = make(url.Values, 1)
		params.Query.Set("limit", "40") // enables paging
		obj = ap.NewASOrderedCollection(params)

	default:
		// Paging enabled.
		// Get page of full follower objects with attached accounts.
		followers, err := p.state.DB.GetAccountFollowers(ctx, receivingAcct.ID, page)
		if err != nil {
			err := gtserror.Newf("error getting followers: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// page ID values.
		var lo, hi string

		if len(followers) > 0 {
			// Get the lowest and highest
			// ID values, used for paging.
			lo = followers[len(followers)-1].ID
			hi = followers[0].ID
		}

		// Start building AS collection page params.
		params.Total = util.Ptr(*receivingAcct.Stats.FollowersCount)
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

	// Serialize the prepared object.
	data, err := ap.Serialize(obj)
	if err != nil {
		err := gtserror.Newf("error serializing: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

// FollowingGet returns the serialized ActivityPub
// collection of a local account's following collection,
// which contains links to accounts followed by this account.
func (p *Processor) FollowingGet(ctx context.Context, requestedUser string, page *paging.Page) (interface{}, gtserror.WithCode) {
	// Authenticate incoming request, getting related accounts.
	auth, errWithCode := p.authenticate(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}
	receivingAcct := auth.receivingAcct

	// Parse collection ID from account's following URI.
	collectionID, err := url.Parse(receivingAcct.FollowingURI)
	if err != nil {
		err := gtserror.Newf("error parsing account following uri %s: %w", receivingAcct.FollowingURI, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Ensure we have stats for this account.
	if err := p.state.DB.PopulateAccountStats(ctx, receivingAcct); err != nil {
		err := gtserror.Newf("error getting stats for account %s: %w", receivingAcct.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	var obj vocab.Type

	// Start AS collection params.
	var params ap.CollectionParams
	params.ID = collectionID

	switch {
	case receivingAcct.IsInstance() ||
		*receivingAcct.Settings.HideCollections:
		// If account that hides collections, or instance
		// account (ie., can't post / have relationships),
		// just return barest stub of collection.
		obj = ap.NewASOrderedCollection(params)

	case page == nil || auth.handshakingURI != nil:
		// If paging disabled, or we're currently handshaking
		// the requester, just return collection that links
		// to first page (i.e. path below), with no items.
		params.Total = util.Ptr(*receivingAcct.Stats.FollowingCount)
		params.First = new(paging.Page)
		params.Query = make(url.Values, 1)
		params.Query.Set("limit", "40") // enables paging
		obj = ap.NewASOrderedCollection(params)

	default:
		// Paging enabled.
		// Get page of full follower objects with attached accounts.
		follows, err := p.state.DB.GetAccountFollows(ctx, receivingAcct.ID, page)
		if err != nil {
			err := gtserror.Newf("error getting follows: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// page ID values.
		var lo, hi string

		if len(follows) > 0 {
			// Get the lowest and highest
			// ID values, used for paging.
			lo = follows[len(follows)-1].ID
			hi = follows[0].ID
		}

		// Start AS collection page params.
		params.Total = util.Ptr(*receivingAcct.Stats.FollowingCount)
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
			// Get followed URI at index.
			follow := follows[i]
			accURI := follow.TargetAccount.URI

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

	// Serialize the prepared object.
	data, err := ap.Serialize(obj)
	if err != nil {
		err := gtserror.Newf("error serializing: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

// FeaturedCollectionGet returns an ordered collection of the requested username's Pinned posts.
// The returned collection have an `items` property which contains an ordered list of status URIs.
func (p *Processor) FeaturedCollectionGet(ctx context.Context, requestedUser string) (interface{}, gtserror.WithCode) {
	// Authenticate incoming request, getting related accounts.
	auth, errWithCode := p.authenticate(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}
	receivingAcct := auth.receivingAcct

	statuses, err := p.state.DB.GetAccountPinnedStatuses(ctx, receivingAcct.ID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	collection, err := p.converter.StatusesToASFeaturedCollection(ctx, receivingAcct.FeaturedCollectionURI, statuses)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.Serialize(collection)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
