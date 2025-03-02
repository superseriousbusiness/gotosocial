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

package dereferencing

import (
	"context"
	"net/url"

	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// dereferenceCollectionPage returns the activitystreams Collection at the specified IRI, or an error if something goes wrong.
func (d *Dereferencer) dereferenceCollection(ctx context.Context, username string, pageIRI *url.URL) (ap.CollectionIterator, error) {
	if blocked, err := d.state.DB.IsDomainBlocked(ctx, pageIRI.Host); blocked || err != nil {
		return nil, gtserror.Newf("domain %s is blocked", pageIRI.Host)
	}

	transport, err := d.transportController.NewTransportForUsername(ctx, username)
	if err != nil {
		return nil, gtserror.Newf("error creating transport: %w", err)
	}

	rsp, err := transport.Dereference(ctx, pageIRI)
	if err != nil {
		return nil, gtserror.Newf("error dereferencing %s: %w", pageIRI.String(), err)
	}

	collect, err := ap.ResolveCollection(ctx, rsp.Body)

	// Tidy up rsp body.
	_ = rsp.Body.Close()

	if err != nil {
		return nil, gtserror.Newf("error resolving collection %s: %w", pageIRI.String(), err)
	}

	return collect, nil
}

// dereferenceCollectionPage returns the activitystreams CollectionPage at the specified IRI, or an error if something goes wrong.
func (d *Dereferencer) dereferenceCollectionPage(ctx context.Context, username string, pageIRI *url.URL) (ap.CollectionPageIterator, error) {
	if blocked, err := d.state.DB.IsDomainBlocked(ctx, pageIRI.Host); blocked || err != nil {
		return nil, gtserror.Newf("domain %s is blocked", pageIRI.Host)
	}

	transport, err := d.transportController.NewTransportForUsername(ctx, username)
	if err != nil {
		return nil, gtserror.Newf("error creating transport: %w", err)
	}

	rsp, err := transport.Dereference(ctx, pageIRI)
	if err != nil {
		return nil, gtserror.Newf("error deferencing %s: %w", pageIRI.String(), err)
	}

	page, err := ap.ResolveCollectionPage(ctx, rsp.Body)

	// Tidy up rsp body.
	_ = rsp.Body.Close()

	if err != nil {
		return nil, gtserror.Newf("error resolving collection page %s: %w", pageIRI.String(), err)
	}

	return page, nil
}

// getAttachedStatusCollection is a small utility function to fetch the first page of an
// attached activity streams collection from a provided statusable object, along with a URI.
func getAttachedStatusCollectionPage(status ap.Statusable) (ap.CollectionPageIterator, string) { //nolint:gocritic
	// Look for an attached status replies (as collection)
	replies := status.GetActivityStreamsReplies()
	if replies == nil {
		return nil, ""
	}

	// Look for an attached collection page, wrap and return.
	if page := getRepliesCollectionPage(replies); page != nil {
		return ap.WrapCollectionPage(page), getIDString(page)
	}

	// Look for an attached ordered collection page, wrap and return.
	if page := getRepliesOrderedCollectionPage(replies); page != nil {
		return ap.WrapOrderedCollectionPage(page), getIDString(page)
	}

	log.Warnf(nil, "replies without collection page: %s", getIDString(status))
	return nil, ""
}

func getRepliesCollectionPage(replies vocab.ActivityStreamsRepliesProperty) vocab.ActivityStreamsCollectionPage {
	// Get the status replies collection
	collection := replies.GetActivityStreamsCollection()
	if collection == nil {
		return nil
	}

	// Get the "first" property of the replies collection
	first := collection.GetActivityStreamsFirst()
	if first == nil {
		return nil
	}

	// Return the first activity stream collection page
	return first.GetActivityStreamsCollectionPage()
}

func getRepliesOrderedCollectionPage(replies vocab.ActivityStreamsRepliesProperty) vocab.ActivityStreamsOrderedCollectionPage {
	// Get the status replies collection
	collection := replies.GetActivityStreamsOrderedCollection()
	if collection == nil {
		return nil
	}

	// Get the "first" property of the replies collection
	first := collection.GetActivityStreamsFirst()
	if first == nil {
		return nil
	}

	// Return the first activity stream collection page
	return first.GetActivityStreamsOrderedCollectionPage()
}

// getIDString is shorthand to fetch an ID URI string from AP type with attached JSONLDId.
func getIDString(a ap.WithJSONLDId) string {
	id := a.GetJSONLDId()
	if id == nil {
		return ""
	}
	uri := id.Get()
	if uri == nil {
		return ""
	}
	return uri.String()
}
