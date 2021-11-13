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

package federatingdb

import (
	"context"
	"net/url"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

// InboxContains returns true if the OrderedCollection at 'inbox'
// contains the specified 'id'.
//
// The library makes this call only after acquiring a lock first.
//
// Implementation note: we have our own logic for inboxes so always return false here.
func (f *federatingDB) InboxContains(c context.Context, inbox, id *url.URL) (contains bool, err error) {
	return false, nil
}

// GetInbox returns the first ordered collection page of the outbox at
// the specified IRI, for prepending new items.
//
// The library makes this call only after acquiring a lock first.
//
// Implementation note: we don't (yet) serve inboxes, so just return empty and nil here.
func (f *federatingDB) GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	return streams.NewActivityStreamsOrderedCollectionPage(), nil
}

// SetInbox saves the inbox value given from GetInbox, with new items
// prepended. Note that the new items must not be added as independent
// database entries. Separate calls to Create will do that.
//
// The library makes this call only after acquiring a lock first.
//
// Implementation note: we don't allow inbox setting so just return nil here.
func (f *federatingDB) SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	return nil
}

// InboxForActor fetches the inbox corresponding to the given actorIRI.
//
// It is acceptable to just return nil for the inboxIRI. In this case, the library will
// attempt to resolve the inbox of the actor by remote dereferencing instead.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) InboxForActor(c context.Context, actorIRI *url.URL) (inboxIRI *url.URL, err error) {
	account, err := f.db.GetAccountByURI(c, actorIRI.String())
	if err != nil {
		// if there are just no entries for this account yet it's fine, return nil
		// and go-fed will try to dereference it instead
		if err == db.ErrNoEntries {
			return nil, nil
		}
		// there's been an actual error...
		return nil, err
	}

	// we got it!
	return url.Parse(account.InboxURI)
}
