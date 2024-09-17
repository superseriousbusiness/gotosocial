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

package federatingdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"codeberg.org/gruf/go-byteutil"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func typeNames(objects []ap.TypeOrIRI) []string {
	names := make([]string, len(objects))
	for i, object := range objects {
		if object.IsIRI() {
			names[i] = "IRI"
		} else if t := object.GetType(); t != nil {
			names[i] = t.GetTypeName()
		} else {
			names[i] = "nil"
		}
	}
	return names
}

// isSender returns whether an object with AttributedTo property comes from the given requesting account.
func isSender(with ap.WithAttributedTo, requester *gtsmodel.Account) bool {
	for _, uri := range ap.GetAttributedTo(with) {
		if uri.String() == requester.URI {
			return true
		}
	}
	return false
}

func sameActor(actor1 vocab.ActivityStreamsActorProperty, actor2 vocab.ActivityStreamsActorProperty) bool {
	if actor1 == nil || actor2 == nil {
		return false
	}

	for a1Iter := actor1.Begin(); a1Iter != actor1.End(); a1Iter = a1Iter.Next() {
		for a2Iter := actor2.Begin(); a2Iter != actor2.End(); a2Iter = a2Iter.Next() {
			if a1Iter.GetIRI() == nil {
				return false
			}

			if a2Iter.GetIRI() == nil {
				return false
			}

			if a1Iter.GetIRI().String() == a2Iter.GetIRI().String() {
				return true
			}
		}
	}

	return false
}

// NewID creates a new IRI id for the provided activity or object. The
// implementation does not need to set the 'id' property and simply
// needs to determine the value.
//
// The go-fed library will handle setting the 'id' property on the
// activity or object provided with the value returned.
func (f *federatingDB) NewID(ctx context.Context, t vocab.Type) (idURL *url.URL, err error) {
	log.DebugKV(ctx, "newID", serialize{t})

	// Most of our types set an ID already
	// by this point, return this if found.
	idProp := t.GetJSONLDId()
	if idProp != nil && idProp.IsIRI() {
		return idProp.GetIRI(), nil
	}

	if t.GetTypeName() == ap.ActivityFollow {
		follow, _ := t.(vocab.ActivityStreamsFollow)

		// If an actor URI has been set, create a new ID
		// based on actor (i.e. followER not the followEE).
		if uri := ap.GetActorIRIs(follow); len(uri) == 1 {
			if actorAccount, err := f.state.DB.GetAccountByURI(ctx, uri[0].String()); err == nil {
				newID, err := id.NewRandomULID()
				if err != nil {
					return nil, err
				}
				return url.Parse(uris.GenerateURIForFollow(actorAccount.Username, newID))
			}
		}
	}

	// Default fallback behaviour:
	// {proto}://{host}/{randomID}
	newID, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	return url.Parse(fmt.Sprintf("%s://%s/%s", config.GetProtocol(), config.GetHost(), newID))
}

// ActorForOutbox fetches the actor's IRI for the given outbox IRI.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) ActorForOutbox(ctx context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	acct, err := f.getAccountForIRI(ctx, outboxIRI)
	if err != nil {
		return nil, err
	}
	return url.Parse(acct.URI)
}

// ActorForInbox fetches the actor's IRI for the given outbox IRI.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) ActorForInbox(ctx context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	acct, err := f.getAccountForIRI(ctx, inboxIRI)
	if err != nil {
		return nil, err
	}
	return url.Parse(acct.URI)
}

// getAccountForIRI returns the account that corresponds to or owns the given IRI.
func (f *federatingDB) getAccountForIRI(ctx context.Context, iri *url.URL) (*gtsmodel.Account, error) {
	var (
		acct *gtsmodel.Account
		err  error
	)

	switch {
	case uris.IsUserPath(iri):
		if acct, err = f.state.DB.GetAccountByURI(ctx, iri.String()); err != nil {
			if err == db.ErrNoEntries {
				return nil, fmt.Errorf("no actor found that corresponds to uri %s", iri.String())
			}
			return nil, fmt.Errorf("db error searching for actor with uri %s", iri.String())
		}
		return acct, nil
	case uris.IsInboxPath(iri):
		if acct, err = f.state.DB.GetAccountByInboxURI(ctx, iri.String()); err != nil {
			if err == db.ErrNoEntries {
				return nil, fmt.Errorf("no actor found that corresponds to inbox %s", iri.String())
			}
			return nil, fmt.Errorf("db error searching for actor with inbox %s", iri.String())
		}
		return acct, nil
	case uris.IsOutboxPath(iri):
		if acct, err = f.state.DB.GetAccountByOutboxURI(ctx, iri.String()); err != nil {
			if err == db.ErrNoEntries {
				return nil, fmt.Errorf("no actor found that corresponds to outbox %s", iri.String())
			}
			return nil, fmt.Errorf("db error searching for actor with outbox %s", iri.String())
		}
		return acct, nil
	case uris.IsFollowersPath(iri):
		if acct, err = f.state.DB.GetAccountByFollowersURI(ctx, iri.String()); err != nil {
			if err == db.ErrNoEntries {
				return nil, fmt.Errorf("no actor found that corresponds to followers_uri %s", iri.String())
			}
			return nil, fmt.Errorf("db error searching for actor with followers_uri %s", iri.String())
		}
		return acct, nil
	case uris.IsFollowingPath(iri):
		if acct, err = f.state.DB.GetAccountByFollowingURI(ctx, iri.String()); err != nil {
			if err == db.ErrNoEntries {
				return nil, fmt.Errorf("no actor found that corresponds to following_uri %s", iri.String())
			}
			return nil, fmt.Errorf("db error searching for actor with following_uri %s", iri.String())
		}
		return acct, nil
	default:
		return nil, fmt.Errorf("getActorForIRI: iri %s not recognised", iri)
	}
}

// collectFollows takes a slice of iris and converts them into ActivityStreamsCollection of IRIs.
func (f *federatingDB) collectIRIs(ctx context.Context, iris []*url.URL) (vocab.ActivityStreamsCollection, error) {
	collection := streams.NewActivityStreamsCollection()
	items := streams.NewActivityStreamsItemsProperty()
	for _, i := range iris {
		items.AppendIRI(i)
	}
	collection.SetActivityStreamsItems(items)
	return collection, nil
}

// activityContext represents the context in
// which a call to one of the federatingdb
// functions is taking place, including the
// account who initiated the request via POST
// to an inbox, and the account who received
// the request in their inbox.
type activityContext struct {
	// The account that owns the inbox
	// or URI being interacted with.
	receivingAcct *gtsmodel.Account

	// The account whose keyId was used
	// to POST a request to the inbox.
	requestingAcct *gtsmodel.Account

	// Whether this is an internal request,
	// ie., one originating not from the
	// API but from inside the instance.
	//
	// If the request is internal, it's
	// safe to assume that the activity
	// has already been processed elsewhere,
	// and we can return with no action.
	internal bool
}

// getActivityContext extracts the context in
// which an Activity is taking place from the
// context.Context passed in to one of the
// federatingdb functions.
func getActivityContext(ctx context.Context) activityContext {
	receivingAcct := gtscontext.ReceivingAccount(ctx)
	requestingAcct := gtscontext.RequestingAccount(ctx)

	// If the receiving account wasn't set on
	// the context, that means this request
	// didn't pass through the fedi API, but
	// came from inside the instance as the
	// result of a local activity.
	internal := receivingAcct == nil

	return activityContext{
		receivingAcct:  receivingAcct,
		requestingAcct: requestingAcct,
		internal:       internal,
	}
}

// serialize wraps a vocab.Type to provide
// lazy-serialization along with error output.
type serialize struct{ item vocab.Type }

func (s serialize) String() string {
	m, err := ap.Serialize(s.item)
	if err != nil {
		return "!(error serializing item: " + err.Error() + ")"
	}

	b, err := json.Marshal(m)
	if err != nil {
		return "!(error json marshaling item: " + err.Error() + ")"
	}

	return byteutil.B2S(b)
}
