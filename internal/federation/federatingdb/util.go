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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func sameActor(activityActor vocab.ActivityStreamsActorProperty, followActor vocab.ActivityStreamsActorProperty) bool {
	if activityActor == nil || followActor == nil {
		return false
	}
	for aIter := activityActor.Begin(); aIter != activityActor.End(); aIter = aIter.Next() {
		for fIter := followActor.Begin(); fIter != followActor.End(); fIter = fIter.Next() {
			if aIter.GetIRI() == nil {
				return false
			}
			if fIter.GetIRI() == nil {
				return false
			}
			if aIter.GetIRI().String() == fIter.GetIRI().String() {
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
	if log.Level() >= level.DEBUG {
		i, err := marshalItem(t)
		if err != nil {
			return nil, err
		}
		l := log.WithContext(ctx).
			WithField("newID", i)
		l.Debug("entering NewID")
	}

	switch t.GetTypeName() {
	case ap.ActivityFollow:
		// FOLLOW
		// ID might already be set on a follow we've created, so check it here and return it if it is
		follow, ok := t.(vocab.ActivityStreamsFollow)
		if !ok {
			return nil, errors.New("newid: follow couldn't be parsed into vocab.ActivityStreamsFollow")
		}
		idProp := follow.GetJSONLDId()
		if idProp != nil {
			if idProp.IsIRI() {
				return idProp.GetIRI(), nil
			}
		}
		// it's not set so create one based on the actor set on the follow (ie., the followER not the followEE)
		actorProp := follow.GetActivityStreamsActor()
		if actorProp != nil {
			for iter := actorProp.Begin(); iter != actorProp.End(); iter = iter.Next() {
				// take the IRI of the first actor we can find (there should only be one)
				if iter.IsIRI() {
					// if there's an error here, just use the fallback behavior -- we don't need to return an error here
					if actorAccount, err := f.db.GetAccountByURI(ctx, iter.GetIRI().String()); err == nil {
						newID, err := id.NewRandomULID()
						if err != nil {
							return nil, err
						}
						return url.Parse(uris.GenerateURIForFollow(actorAccount.Username, newID))
					}
				}
			}
		}
	case ap.ObjectNote:
		// NOTE aka STATUS
		// ID might already be set on a note we've created, so check it here and return it if it is
		note, ok := t.(vocab.ActivityStreamsNote)
		if !ok {
			return nil, errors.New("newid: note couldn't be parsed into vocab.ActivityStreamsNote")
		}
		idProp := note.GetJSONLDId()
		if idProp != nil {
			if idProp.IsIRI() {
				return idProp.GetIRI(), nil
			}
		}
	case ap.ActivityLike:
		// LIKE aka FAVE
		// ID might already be set on a fave we've created, so check it here and return it if it is
		fave, ok := t.(vocab.ActivityStreamsLike)
		if !ok {
			return nil, errors.New("newid: fave couldn't be parsed into vocab.ActivityStreamsLike")
		}
		idProp := fave.GetJSONLDId()
		if idProp != nil {
			if idProp.IsIRI() {
				return idProp.GetIRI(), nil
			}
		}
	case ap.ActivityCreate:
		// CREATE
		// ID might already be set on a Create, so check it here and return it if it is
		create, ok := t.(vocab.ActivityStreamsCreate)
		if !ok {
			return nil, errors.New("newid: create couldn't be parsed into vocab.ActivityStreamsCreate")
		}
		idProp := create.GetJSONLDId()
		if idProp != nil {
			if idProp.IsIRI() {
				return idProp.GetIRI(), nil
			}
		}
	case ap.ActivityAnnounce:
		// ANNOUNCE aka BOOST
		// ID might already be set on an announce we've created, so check it here and return it if it is
		announce, ok := t.(vocab.ActivityStreamsAnnounce)
		if !ok {
			return nil, errors.New("newid: announce couldn't be parsed into vocab.ActivityStreamsAnnounce")
		}
		idProp := announce.GetJSONLDId()
		if idProp != nil {
			if idProp.IsIRI() {
				return idProp.GetIRI(), nil
			}
		}
	case ap.ActivityUpdate:
		// UPDATE
		// ID might already be set on an update we've created, so check it here and return it if it is
		update, ok := t.(vocab.ActivityStreamsUpdate)
		if !ok {
			return nil, errors.New("newid: update couldn't be parsed into vocab.ActivityStreamsUpdate")
		}
		idProp := update.GetJSONLDId()
		if idProp != nil {
			if idProp.IsIRI() {
				return idProp.GetIRI(), nil
			}
		}
	case ap.ActivityBlock:
		// BLOCK
		// ID might already be set on a block we've created, so check it here and return it if it is
		block, ok := t.(vocab.ActivityStreamsBlock)
		if !ok {
			return nil, errors.New("newid: block couldn't be parsed into vocab.ActivityStreamsBlock")
		}
		idProp := block.GetJSONLDId()
		if idProp != nil {
			if idProp.IsIRI() {
				return idProp.GetIRI(), nil
			}
		}
	case ap.ActivityUndo:
		// UNDO
		// ID might already be set on an undo we've created, so check it here and return it if it is
		undo, ok := t.(vocab.ActivityStreamsUndo)
		if !ok {
			return nil, errors.New("newid: undo couldn't be parsed into vocab.ActivityStreamsUndo")
		}
		idProp := undo.GetJSONLDId()
		if idProp != nil {
			if idProp.IsIRI() {
				return idProp.GetIRI(), nil
			}
		}
	}

	// fallback default behavior: just return a random ULID after our protocol and host
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
		acct = &gtsmodel.Account{}
		err  error
	)

	switch {
	case uris.IsUserPath(iri):
		if acct, err = f.db.GetAccountByURI(ctx, iri.String()); err != nil {
			if err == db.ErrNoEntries {
				return nil, fmt.Errorf("no actor found that corresponds to uri %s", iri.String())
			}
			return nil, fmt.Errorf("db error searching for actor with uri %s", iri.String())
		}
		return acct, nil
	case uris.IsInboxPath(iri):
		if err = f.db.GetWhere(ctx, []db.Where{{Key: "inbox_uri", Value: iri.String()}}, acct); err != nil {
			if err == db.ErrNoEntries {
				return nil, fmt.Errorf("no actor found that corresponds to inbox %s", iri.String())
			}
			return nil, fmt.Errorf("db error searching for actor with inbox %s", iri.String())
		}
		return acct, nil
	case uris.IsOutboxPath(iri):
		if err = f.db.GetWhere(ctx, []db.Where{{Key: "outbox_uri", Value: iri.String()}}, acct); err != nil {
			if err == db.ErrNoEntries {
				return nil, fmt.Errorf("no actor found that corresponds to outbox %s", iri.String())
			}
			return nil, fmt.Errorf("db error searching for actor with outbox %s", iri.String())
		}
		return acct, nil
	case uris.IsFollowersPath(iri):
		if err = f.db.GetWhere(ctx, []db.Where{{Key: "followers_uri", Value: iri.String()}}, acct); err != nil {
			if err == db.ErrNoEntries {
				return nil, fmt.Errorf("no actor found that corresponds to followers_uri %s", iri.String())
			}
			return nil, fmt.Errorf("db error searching for actor with followers_uri %s", iri.String())
		}
		return acct, nil
	case uris.IsFollowingPath(iri):
		if err = f.db.GetWhere(ctx, []db.Where{{Key: "following_uri", Value: iri.String()}}, acct); err != nil {
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

// extractFromCtx extracts some useful values from a context passed into the federatingDB via the API:
//   - The target account that owns the inbox or URI being interacted with.
//   - The requesting account that posted to the inbox.
//   - A channel that messages for the processor can be placed into.
//
// If a value is not present, nil will be returned for it. It's up to the caller to check this and respond appropriately.
func extractFromCtx(ctx context.Context) (receivingAccount, requestingAccount *gtsmodel.Account) {
	receivingAccountI := ctx.Value(ap.ContextReceivingAccount)
	if receivingAccountI != nil {
		var ok bool
		receivingAccount, ok = receivingAccountI.(*gtsmodel.Account)
		if !ok {
			log.Panicf(ctx, "context entry with key %s could not be asserted to *gtsmodel.Account", ap.ContextReceivingAccount)
		}
	}

	requestingAcctI := ctx.Value(ap.ContextRequestingAccount)
	if requestingAcctI != nil {
		var ok bool
		requestingAccount, ok = requestingAcctI.(*gtsmodel.Account)
		if !ok {
			log.Panicf(ctx, "context entry with key %s could not be asserted to *gtsmodel.Account", ap.ContextRequestingAccount)
		}
	}

	return
}

func marshalItem(item vocab.Type) (string, error) {
	m, err := streams.Serialize(item)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
