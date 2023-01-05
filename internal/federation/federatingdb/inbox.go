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

package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
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

// InboxesForIRI fetches inboxes corresponding to the given iri.
// This allows your server to skip remote dereferencing of iris
// in order to speed up message delivery, if desired.
//
// It is acceptable to just return nil or an empty slice for the inboxIRIs,
// if you don't know the inbox iri, or you don't wish to use this feature.
// In this case, the library will attempt to resolve inboxes of the iri
// by remote dereferencing instead.
//
// If the input iri is the iri of an Actor, then the inbox for the actor
// should be returned as a single-entry slice.
//
// If the input iri is a Collection (such as a Collection of followers),
// then each follower inbox IRI should be returned in the inboxIRIs slice.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) InboxesForIRI(c context.Context, iri *url.URL) (inboxIRIs []*url.URL, err error) {
	// check if this is a followers collection iri for a local account...
	if iri.Host == config.GetHost() && uris.IsFollowersPath(iri) {
		localAccountUsername, err := uris.ParseFollowersPath(iri)
		if err != nil {
			return nil, fmt.Errorf("couldn't extract local account username from uri %s: %s", iri, err)
		}

		account, err := f.db.GetAccountByUsernameDomain(c, localAccountUsername, "")
		if err != nil {
			return nil, fmt.Errorf("couldn't find local account with username %s: %s", localAccountUsername, err)
		}

		follows, err := f.db.GetAccountFollowedBy(c, account.ID, false)
		if err != nil {
			return nil, fmt.Errorf("couldn't get followers of local account %s: %s", localAccountUsername, err)
		}

		for _, follow := range follows {
			// make sure we retrieved the following account from the db
			if follow.Account == nil {
				followingAccount, err := f.db.GetAccountByID(c, follow.AccountID)
				if err != nil {
					if err == db.ErrNoEntries {
						continue
					}
					return nil, fmt.Errorf("error retrieving account with id %s: %s", follow.AccountID, err)
				}
				follow.Account = followingAccount
			}

			// deliver to a shared inbox if we have that option
			var inbox string
			if config.GetInstanceDeliverToSharedInboxes() && follow.Account.SharedInboxURI != nil && *follow.Account.SharedInboxURI != "" {
				inbox = *follow.Account.SharedInboxURI
			} else {
				inbox = follow.Account.InboxURI
			}

			inboxIRI, err := url.Parse(inbox)
			if err != nil {
				return nil, fmt.Errorf("error parsing inbox uri of following account %s: %s", follow.Account.InboxURI, err)
			}
			inboxIRIs = append(inboxIRIs, inboxIRI)
		}
		return inboxIRIs, nil
	}

	// check if this is just an account IRI...
	if account, err := f.db.GetAccountByURI(c, iri.String()); err == nil {
		// deliver to a shared inbox if we have that option
		var inbox string
		if config.GetInstanceDeliverToSharedInboxes() && account.SharedInboxURI != nil && *account.SharedInboxURI != "" {
			inbox = *account.SharedInboxURI
		} else {
			inbox = account.InboxURI
		}

		inboxIRI, err := url.Parse(inbox)
		if err != nil {
			return nil, fmt.Errorf("error parsing account inbox uri %s: %s", account.InboxURI, account.InboxURI)
		}
		// we've got it
		inboxIRIs = append(inboxIRIs, inboxIRI)
		return inboxIRIs, nil
	} else if err != db.ErrNoEntries {
		// there's been a real error
		return nil, err
	}

	// no error, we just didn't find anything so let the library handle the rest
	return nil, nil
}
