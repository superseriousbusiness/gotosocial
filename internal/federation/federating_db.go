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

package federation

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// FederatingDB uses the underlying DB interface to implement the go-fed pub.Database interface.
// It doesn't care what the underlying implementation of the DB interface is, as long as it works.
type federatingDB struct {
	locks         *sync.Map
	db            db.DB
	config        *config.Config
	log           *logrus.Logger
	typeConverter typeutils.TypeConverter
}

func NewFederatingDB(db db.DB, config *config.Config, log *logrus.Logger) pub.Database {
	return &federatingDB{
		locks:         new(sync.Map),
		db:            db,
		config:        config,
		log:           log,
		typeConverter: typeutils.NewConverter(config, db),
	}
}

/*
   GO-FED DB INTERFACE-IMPLEMENTING FUNCTIONS
*/

// Lock takes a lock for the object at the specified id. If an error
// is returned, the lock must not have been taken.
//
// The lock must be able to succeed for an id that does not exist in
// the database. This means acquiring the lock does not guarantee the
// entry exists in the database.
//
// Locks are encouraged to be lightweight and in the Go layer, as some
// processes require tight loops acquiring and releasing locks.
//
// Used to ensure race conditions in multiple requests do not occur.
func (f *federatingDB) Lock(c context.Context, id *url.URL) error {
	// Before any other Database methods are called, the relevant `id`
	// entries are locked to allow for fine-grained concurrency.

	// Strategy: create a new lock, if stored, continue. Otherwise, lock the
	// existing mutex.
	mu := &sync.Mutex{}
	mu.Lock() // Optimistically lock if we do store it.
	i, loaded := f.locks.LoadOrStore(id.String(), mu)
	if loaded {
		mu = i.(*sync.Mutex)
		mu.Lock()
	}
	return nil
}

// Unlock makes the lock for the object at the specified id available.
// If an error is returned, the lock must have still been freed.
//
// Used to ensure race conditions in multiple requests do not occur.
func (f *federatingDB) Unlock(c context.Context, id *url.URL) error {
	// Once Go-Fed is done calling Database methods, the relevant `id`
	// entries are unlocked.

	i, ok := f.locks.Load(id.String())
	if !ok {
		return errors.New("missing an id in unlock")
	}
	mu := i.(*sync.Mutex)
	mu.Unlock()
	return nil
}

// InboxContains returns true if the OrderedCollection at 'inbox'
// contains the specified 'id'.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) InboxContains(c context.Context, inbox, id *url.URL) (contains bool, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "InboxContains",
			"id":   id.String(),
		},
	)
	l.Debugf("entering INBOXCONTAINS function with for inbox %s and id %s", inbox.String(), id.String())

	if !util.IsInboxPath(inbox) {
		return false, fmt.Errorf("%s is not an inbox URI", inbox.String())
	}

	activityI := c.Value(util.APActivity)
	if activityI == nil {
		return false, fmt.Errorf("no activity was set for id %s", id.String())
	}
	activity, ok := activityI.(pub.Activity)
	if !ok || activity == nil {
		return false, fmt.Errorf("could not parse contextual activity for id %s", id.String())
	}

	l.Debugf("activity type %s for id %s", activity.GetTypeName(), id.String())

	return false, nil

	// if err := f.db.GetByID(statusID, &gtsmodel.Status{}); err != nil {
	// 	if _, ok := err.(db.ErrNoEntries); ok {
	// 		// we don't have it
	// 		return false, nil
	// 	}
	// 	// actual error
	// 	return false, fmt.Errorf("error getting status from db: %s", err)
	// }

	// // we must have it
	// return true, nil
}

// GetInbox returns the first ordered collection page of the outbox at
// the specified IRI, for prepending new items.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "GetInbox",
		},
	)
	l.Debugf("entering GETINBOX function with inboxIRI %s", inboxIRI.String())
	return streams.NewActivityStreamsOrderedCollectionPage(), nil
}

// SetInbox saves the inbox value given from GetInbox, with new items
// prepended. Note that the new items must not be added as independent
// database entries. Separate calls to Create will do that.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "SetInbox",
		},
	)
	l.Debug("entering SETINBOX function")
	return nil
}

// Owns returns true if the IRI belongs to this instance, and if
// the database has an entry for the IRI.
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Owns(c context.Context, id *url.URL) (bool, error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Owns",
			"id":   id.String(),
		},
	)
	l.Debugf("entering OWNS function with id %s", id.String())

	// if the id host isn't this instance host, we don't own this IRI
	if id.Host != f.config.Host {
		l.Debugf("we DO NOT own activity because the host is %s not %s", id.Host, f.config.Host)
		return false, nil
	}

	// apparently it belongs to this host, so what *is* it?

	// check if it's a status, eg /users/example_username/statuses/SOME_UUID_OF_A_STATUS
	if util.IsStatusesPath(id) {
		_, uid, err := util.ParseStatusesPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing statuses path for url %s: %s", id.String(), err)
		}
		if err := f.db.GetWhere("uri", uid, &gtsmodel.Status{}); err != nil {
			if _, ok := err.(db.ErrNoEntries); ok {
				// there are no entries for this status
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching status with id %s: %s", uid, err)
		}
		l.Debug("we DO own this")
		return true, nil
	}

	// check if it's a user, eg /users/example_username
	if util.IsUserPath(id) {
		username, err := util.ParseUserPath(id)
		if err != nil {
			return false, fmt.Errorf("error parsing statuses path for url %s: %s", id.String(), err)
		}
		if err := f.db.GetLocalAccountByUsername(username, &gtsmodel.Account{}); err != nil {
			if _, ok := err.(db.ErrNoEntries); ok {
				// there are no entries for this username
				return false, nil
			}
			// an actual error happened
			return false, fmt.Errorf("database error fetching account with username %s: %s", username, err)
		}
		l.Debug("we DO own this")
		return true, nil
	}

	return false, fmt.Errorf("could not match activityID: %s", id.String())
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
	if err := f.db.GetWhere("outbox_uri", outboxIRI.String(), acct); err != nil {
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
	if err := f.db.GetWhere("inbox_uri", inboxIRI.String(), acct); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return nil, fmt.Errorf("no actor found that corresponds to inbox %s", inboxIRI.String())
		}
		return nil, fmt.Errorf("db error searching for actor with inbox %s", inboxIRI.String())
	}
	return url.Parse(acct.URI)
}

// OutboxForInbox fetches the corresponding actor's outbox IRI for the
// actor's inbox IRI.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) OutboxForInbox(c context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func":     "OutboxForInbox",
			"inboxIRI": inboxIRI.String(),
		},
	)
	l.Debugf("entering OUTBOXFORINBOX function with inboxIRI %s", inboxIRI.String())

	if !util.IsInboxPath(inboxIRI) {
		return nil, fmt.Errorf("%s is not an inbox URI", inboxIRI.String())
	}
	acct := &gtsmodel.Account{}
	if err := f.db.GetWhere("inbox_uri", inboxIRI.String(), acct); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return nil, fmt.Errorf("no actor found that corresponds to inbox %s", inboxIRI.String())
		}
		return nil, fmt.Errorf("db error searching for actor with inbox %s", inboxIRI.String())
	}
	return url.Parse(acct.OutboxURI)
}

// Exists returns true if the database has an entry for the specified
// id. It may not be owned by this application instance.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Exists(c context.Context, id *url.URL) (exists bool, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Exists",
			"id":   id.String(),
		},
	)
	l.Debugf("entering EXISTS function with id %s", id.String())

	return false, nil
}

// Get returns the database entry for the specified id.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Get(c context.Context, id *url.URL) (value vocab.Type, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Get",
			"id":   id.String(),
		},
	)
	l.Debug("entering GET function")

	if util.IsUserPath(id) {
		acct := &gtsmodel.Account{}
		if err := f.db.GetWhere("uri", id.String(), acct); err != nil {
			return nil, err
		}
		return f.typeConverter.AccountToAS(acct)
	}

	return nil, nil
}

// Create adds a new entry to the database which must be able to be
// keyed by its id.
//
// Note that Activity values received from federated peers may also be
// created in the database this way if the Federating Protocol is
// enabled. The client may freely decide to store only the id instead of
// the entire value.
//
// The library makes this call only after acquiring a lock first.
//
// Under certain conditions and network activities, Create may be called
// multiple times for the same ActivityStreams object.
func (f *federatingDB) Create(c context.Context, asType vocab.Type) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func":   "Create",
			"asType": asType.GetTypeName(),
		},
	)
	l.Debugf("received CREATE asType %+v", asType)

	switch gtsmodel.ActivityStreamsActivity(asType.GetTypeName()) {
	case gtsmodel.ActivityStreamsCreate:
		create, ok := asType.(vocab.ActivityStreamsCreate)
		if !ok {
			return errors.New("could not convert type to create")
		}
		object := create.GetActivityStreamsObject()
		for objectIter := object.Begin(); objectIter != object.End(); objectIter = objectIter.Next() {
			switch gtsmodel.ActivityStreamsObject(objectIter.GetType().GetTypeName()) {
			case gtsmodel.ActivityStreamsNote:
				note := objectIter.GetActivityStreamsNote()
				status, err := f.typeConverter.ASStatusToStatus(note)
				if err != nil {
					return fmt.Errorf("error converting note to status: %s", err)
				}
				if err := f.db.Put(status); err != nil {
					return fmt.Errorf("database error inserting status: %s", err)
				}
			}
		}
	case gtsmodel.ActivityStreamsFollow:
		follow, ok := asType.(vocab.ActivityStreamsFollow)
		if !ok {
			return errors.New("could not convert type to follow")
		}

		followRequest, err := f.typeConverter.ASFollowToFollowRequest(follow)
		if err != nil {
			return fmt.Errorf("could not convert Follow to follow request: %s", err)
		}

		if err := f.db.Put(followRequest); err != nil {
			return fmt.Errorf("database error inserting follow request: %s", err)
		}
	}
	return nil
}

// Update sets an existing entry to the database based on the value's
// id.
//
// Note that Activity values received from federated peers may also be
// updated in the database this way if the Federating Protocol is
// enabled. The client may freely decide to store only the id instead of
// the entire value.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Update(c context.Context, asType vocab.Type) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func":   "Update",
			"asType": asType.GetTypeName(),
		},
	)
	l.Debugf("received UPDATE asType %+v", asType)
	return nil
}

// Delete removes the entry with the given id.
//
// Delete is only called for federated objects. Deletes from the Social
// Protocol instead call Update to create a Tombstone.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Delete(c context.Context, id *url.URL) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Delete",
			"id":   id.String(),
		},
	)
	l.Debugf("received DELETE id %s", id.String())
	return nil
}

// GetOutbox returns the first ordered collection page of the outbox
// at the specified IRI, for prepending new items.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) GetOutbox(c context.Context, outboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "GetOutbox",
		},
	)
	l.Debug("entering GETOUTBOX function")

	return nil, nil
}

// SetOutbox saves the outbox value given from GetOutbox, with new items
// prepended. Note that the new items must not be added as independent
// database entries. Separate calls to Create will do that.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) SetOutbox(c context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "SetOutbox",
		},
	)
	l.Debug("entering SETOUTBOX function")

	return nil
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
	l.Debugf("received NEWID request for asType %+v", t)

	return url.Parse(fmt.Sprintf("%s://%s/", f.config.Protocol, uuid.NewString()))
}

// Followers obtains the Followers Collection for an actor with the
// given id.
//
// If modified, the library will then call Update.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Followers(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Followers",
			"actorIRI":   actorIRI.String(),
		},
	)
	l.Debugf("entering FOLLOWERS function with actorIRI %s", actorIRI.String())

	acct := &gtsmodel.Account{}
	if err := f.db.GetWhere("uri", actorIRI.String(), acct); err != nil {
		return nil, fmt.Errorf("db error getting account with uri %s: %s", actorIRI.String(), err)
	}

	acctFollowers := []gtsmodel.Follow{}
	if err := f.db.GetFollowersByAccountID(acct.ID, &acctFollowers); err != nil {
		return nil, fmt.Errorf("db error getting followers for account id %s: %s", acct.ID, err)
	}

	followers = streams.NewActivityStreamsCollection()
	items := streams.NewActivityStreamsItemsProperty()
	for _, follow := range acctFollowers {
		gtsFollower := &gtsmodel.Account{}
		if err := f.db.GetByID(follow.AccountID, gtsFollower); err != nil {
			return nil, fmt.Errorf("db error getting account id %s: %s", follow.AccountID, err)
		}
		uri, err := url.Parse(gtsFollower.URI)
		if err != nil {
			return nil, fmt.Errorf("error parsing %s as url: %s", gtsFollower.URI, err)
		}
		items.AppendIRI(uri)
	}
	followers.SetActivityStreamsItems(items)
	return
}

// Following obtains the Following Collection for an actor with the
// given id.
//
// If modified, the library will then call Update.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Following(c context.Context, actorIRI *url.URL) (following vocab.ActivityStreamsCollection, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Following",
			"actorIRI":   actorIRI.String(),
		},
	)
	l.Debugf("entering FOLLOWING function with actorIRI %s", actorIRI.String())

	acct := &gtsmodel.Account{}
	if err := f.db.GetWhere("uri", actorIRI.String(), acct); err != nil {
		return nil, fmt.Errorf("db error getting account with uri %s: %s", actorIRI.String(), err)
	}

	acctFollowing := []gtsmodel.Follow{}
	if err := f.db.GetFollowingByAccountID(acct.ID, &acctFollowing); err != nil {
		return nil, fmt.Errorf("db error getting following for account id %s: %s", acct.ID, err)
	}

	following = streams.NewActivityStreamsCollection()
	items := streams.NewActivityStreamsItemsProperty()
	for _, follow := range acctFollowing {
		gtsFollowing := &gtsmodel.Account{}
		if err := f.db.GetByID(follow.AccountID, gtsFollowing); err != nil {
			return nil, fmt.Errorf("db error getting account id %s: %s", follow.AccountID, err)
		}
		uri, err := url.Parse(gtsFollowing.URI)
		if err != nil {
			return nil, fmt.Errorf("error parsing %s as url: %s", gtsFollowing.URI, err)
		}
		items.AppendIRI(uri)
	}
	following.SetActivityStreamsItems(items)
	return
}

// Liked obtains the Liked Collection for an actor with the
// given id.
//
// If modified, the library will then call Update.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Liked(c context.Context, actorIRI *url.URL) (liked vocab.ActivityStreamsCollection, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Liked",
			"actorIRI":   actorIRI.String(),
		},
	)
	l.Debugf("entering LIKED function with actorIRI %s", actorIRI.String())
	return nil, nil
}
