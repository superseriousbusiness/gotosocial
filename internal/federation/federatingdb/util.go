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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
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
func (f *federatingDB) NewID(c context.Context, t vocab.Type) (id *url.URL, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func":   "NewID",
			"asType": t.GetTypeName(),
		},
	)
	m, err := streams.Serialize(t)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	l.Debugf("received NEWID request for asType %s", string(b))

	switch t.GetTypeName() {
	case gtsmodel.ActivityStreamsFollow:
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
					actorAccount := &gtsmodel.Account{}
					if err := f.db.GetWhere([]db.Where{{Key: "uri", Value: iter.GetIRI().String()}}, actorAccount); err == nil { // if there's an error here, just use the fallback behavior -- we don't need to return an error here
						return url.Parse(util.GenerateURIForFollow(actorAccount.Username, f.config.Protocol, f.config.Host, uuid.NewString()))
					}
				}
			}
		}
	case gtsmodel.ActivityStreamsNote:
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
	case gtsmodel.ActivityStreamsLike:
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
	case gtsmodel.ActivityStreamsAnnounce:
		// ANNOUNCE aka BOOST
		// ID might already be set on an announce we've created, so check it here and return it if it is
		announce, ok := t.(vocab.ActivityStreamsAnnounce)
		if !ok {
			return nil, errors.New("newid: fave couldn't be parsed into vocab.ActivityStreamsAnnounce")
		}
		idProp := announce.GetJSONLDId()
		if idProp != nil {
			if idProp.IsIRI() {
				return idProp.GetIRI(), nil
			}
		}
	}

	// fallback default behavior: just return a random UUID after our protocol and host
	return url.Parse(fmt.Sprintf("%s://%s/%s", f.config.Protocol, f.config.Host, uuid.NewString()))
}

// ActorForOutbox fetches the actor's IRI for the given outbox IRI.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) ActorForOutbox(c context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func":     "ActorForOutbox",
			"inboxIRI": outboxIRI.String(),
		},
	)
	l.Debugf("entering ACTORFOROUTBOX function with outboxIRI %s", outboxIRI.String())

	if !util.IsOutboxPath(outboxIRI) {
		return nil, fmt.Errorf("%s is not an outbox URI", outboxIRI.String())
	}
	acct := &gtsmodel.Account{}
	if err := f.db.GetWhere([]db.Where{{Key: "outbox_uri", Value: outboxIRI.String()}}, acct); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return nil, fmt.Errorf("no actor found that corresponds to outbox %s", outboxIRI.String())
		}
		return nil, fmt.Errorf("db error searching for actor with outbox %s", outboxIRI.String())
	}
	return url.Parse(acct.URI)
}

// ActorForInbox fetches the actor's IRI for the given outbox IRI.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) ActorForInbox(c context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func":     "ActorForInbox",
			"inboxIRI": inboxIRI.String(),
		},
	)
	l.Debugf("entering ACTORFORINBOX function with inboxIRI %s", inboxIRI.String())

	if !util.IsInboxPath(inboxIRI) {
		return nil, fmt.Errorf("%s is not an inbox URI", inboxIRI.String())
	}
	acct := &gtsmodel.Account{}
	if err := f.db.GetWhere([]db.Where{{Key: "inbox_uri", Value: inboxIRI.String()}}, acct); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return nil, fmt.Errorf("no actor found that corresponds to inbox %s", inboxIRI.String())
		}
		return nil, fmt.Errorf("db error searching for actor with inbox %s", inboxIRI.String())
	}
	return url.Parse(acct.URI)
}
