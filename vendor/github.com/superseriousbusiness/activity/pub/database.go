package pub

import (
	"context"
	"net/url"

	"github.com/superseriousbusiness/activity/streams/vocab"
)

type Database interface {
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
	Lock(c context.Context, id *url.URL) error
	// Unlock makes the lock for the object at the specified id available.
	// If an error is returned, the lock must have still been freed.
	//
	// Used to ensure race conditions in multiple requests do not occur.
	Unlock(c context.Context, id *url.URL) error
	// InboxContains returns true if the OrderedCollection at 'inbox'
	// contains the specified 'id'.
	//
	// The library makes this call only after acquiring a lock first.
	InboxContains(c context.Context, inbox, id *url.URL) (contains bool, err error)
	// GetInbox returns the first ordered collection page of the outbox at
	// the specified IRI, for prepending new items.
	//
	// The library makes this call only after acquiring a lock first.
	GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error)
	// SetInbox saves the inbox value given from GetInbox, with new items
	// prepended. Note that the new items must not be added as independent
	// database entries. Separate calls to Create will do that.
	//
	// The library makes this call only after acquiring a lock first.
	SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error
	// Owns returns true if the database has an entry for the IRI and it
	// exists in the database.
	//
	// The library makes this call only after acquiring a lock first.
	Owns(c context.Context, id *url.URL) (owns bool, err error)
	// ActorForOutbox fetches the actor's IRI for the given outbox IRI.
	//
	// The library makes this call only after acquiring a lock first.
	ActorForOutbox(c context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error)
	// ActorForInbox fetches the actor's IRI for the given outbox IRI.
	//
	// The library makes this call only after acquiring a lock first.
	ActorForInbox(c context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error)
	// OutboxForInbox fetches the corresponding actor's outbox IRI for the
	// actor's inbox IRI.
	//
	// The library makes this call only after acquiring a lock first.
	OutboxForInbox(c context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error)
	// InboxForActor fetches the inbox corresponding to the given actorIRI.
	//
	// It is acceptable to just return nil for the inboxIRI. In this case, the library will
	// attempt to resolve the inbox of the actor by remote dereferencing instead.
	//
	// The library makes this call only after acquiring a lock first.
	InboxForActor(c context.Context, actorIRI *url.URL) (inboxIRI *url.URL, err error)
	// Exists returns true if the database has an entry for the specified
	// id. It may not be owned by this application instance.
	//
	// The library makes this call only after acquiring a lock first.
	Exists(c context.Context, id *url.URL) (exists bool, err error)
	// Get returns the database entry for the specified id.
	//
	// The library makes this call only after acquiring a lock first.
	Get(c context.Context, id *url.URL) (value vocab.Type, err error)
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
	Create(c context.Context, asType vocab.Type) error
	// Update sets an existing entry to the database based on the value's
	// id.
	//
	// Note that Activity values received from federated peers may also be
	// updated in the database this way if the Federating Protocol is
	// enabled. The client may freely decide to store only the id instead of
	// the entire value.
	//
	// The library makes this call only after acquiring a lock first.
	Update(c context.Context, asType vocab.Type) error
	// Delete removes the entry with the given id.
	//
	// Delete is only called for federated objects. Deletes from the Social
	// Protocol instead call Update to create a Tombstone.
	//
	// The library makes this call only after acquiring a lock first.
	Delete(c context.Context, id *url.URL) error
	// GetOutbox returns the first ordered collection page of the outbox
	// at the specified IRI, for prepending new items.
	//
	// The library makes this call only after acquiring a lock first.
	GetOutbox(c context.Context, outboxIRI *url.URL) (outbox vocab.ActivityStreamsOrderedCollectionPage, err error)
	// SetOutbox saves the outbox value given from GetOutbox, with new items
	// prepended. Note that the new items must not be added as independent
	// database entries. Separate calls to Create will do that.
	//
	// The library makes this call only after acquiring a lock first.
	SetOutbox(c context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error
	// NewID creates a new IRI id for the provided activity or object. The
	// implementation does not need to set the 'id' property and simply
	// needs to determine the value.
	//
	// The go-fed library will handle setting the 'id' property on the
	// activity or object provided with the value returned.
	NewID(c context.Context, t vocab.Type) (id *url.URL, err error)
	// Followers obtains the Followers Collection for an actor with the
	// given id.
	//
	// If modified, the library will then call Update.
	//
	// The library makes this call only after acquiring a lock first.
	Followers(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error)
	// Following obtains the Following Collection for an actor with the
	// given id.
	//
	// If modified, the library will then call Update.
	//
	// The library makes this call only after acquiring a lock first.
	Following(c context.Context, actorIRI *url.URL) (following vocab.ActivityStreamsCollection, err error)
	// Liked obtains the Liked Collection for an actor with the
	// given id.
	//
	// If modified, the library will then call Update.
	//
	// The library makes this call only after acquiring a lock first.
	Liked(c context.Context, actorIRI *url.URL) (liked vocab.ActivityStreamsCollection, err error)
}
