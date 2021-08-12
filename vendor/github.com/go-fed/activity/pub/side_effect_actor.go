package pub

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"net/http"
	"net/url"
)

// sideEffectActor must satisfy the DelegateActor interface.
var _ DelegateActor = &sideEffectActor{}

// sideEffectActor is a DelegateActor that handles the ActivityPub
// implementation side effects, but requires a more opinionated application to
// be written.
//
// Note that when using the sideEffectActor with an application that good-faith
// implements its required interfaces, the ActivityPub specification is
// guaranteed to be correctly followed.
type sideEffectActor struct {
	common CommonBehavior
	s2s    FederatingProtocol
	c2s    SocialProtocol
	db     Database
	clock  Clock
}

// PostInboxRequestBodyHook defers to the delegate.
func (a *sideEffectActor) PostInboxRequestBodyHook(c context.Context, r *http.Request, activity Activity) (context.Context, error) {
	return a.s2s.PostInboxRequestBodyHook(c, r, activity)
}

// PostOutboxRequestBodyHook defers to the delegate.
func (a *sideEffectActor) PostOutboxRequestBodyHook(c context.Context, r *http.Request, data vocab.Type) (context.Context, error) {
	return a.c2s.PostOutboxRequestBodyHook(c, r, data)
}

// AuthenticatePostInbox defers to the delegate to authenticate the request.
func (a *sideEffectActor) AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return a.s2s.AuthenticatePostInbox(c, w, r)
}

// AuthenticateGetInbox defers to the delegate to authenticate the request.
func (a *sideEffectActor) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return a.common.AuthenticateGetInbox(c, w, r)
}

// AuthenticatePostOutbox defers to the delegate to authenticate the request.
func (a *sideEffectActor) AuthenticatePostOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return a.c2s.AuthenticatePostOutbox(c, w, r)
}

// AuthenticateGetOutbox defers to the delegate to authenticate the request.
func (a *sideEffectActor) AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return a.common.AuthenticateGetOutbox(c, w, r)
}

// GetOutbox delegates to the SocialProtocol.
func (a *sideEffectActor) GetOutbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	return a.common.GetOutbox(c, r)
}

// GetInbox delegates to the FederatingProtocol.
func (a *sideEffectActor) GetInbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	return a.s2s.GetInbox(c, r)
}

// AuthorizePostInbox defers to the federating protocol whether the peer request
// is authorized based on the actors' ids.
func (a *sideEffectActor) AuthorizePostInbox(c context.Context, w http.ResponseWriter, activity Activity) (authorized bool, err error) {
	authorized = false
	actor := activity.GetActivityStreamsActor()
	if actor == nil {
		err = fmt.Errorf("no actors in post to inbox")
		return
	}
	var iris []*url.URL
	for i := 0; i < actor.Len(); i++ {
		iter := actor.At(i)
		if iter.IsIRI() {
			iris = append(iris, iter.GetIRI())
		} else if t := iter.GetType(); t != nil {
			iris = append(iris, activity.GetJSONLDId().Get())
		} else {
			err = fmt.Errorf("actor at index %d is missing an id", i)
			return
		}
	}
	// Determine if the actor(s) sending this request are blocked.
	var blocked bool
	if blocked, err = a.s2s.Blocked(c, iris); err != nil {
		return
	} else if blocked {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	authorized = true
	return
}

// PostInbox handles the side effects of determining whether to block the peer's
// request, adding the activity to the actor's inbox, and triggering side
// effects based on the activity's type.
func (a *sideEffectActor) PostInbox(c context.Context, inboxIRI *url.URL, activity Activity) error {
	isNew, err := a.addToInboxIfNew(c, inboxIRI, activity)
	if err != nil {
		return err
	}
	if isNew {
		wrapped, other, err := a.s2s.FederatingCallbacks(c)
		if err != nil {
			return err
		}
		// Populate side channels.
		wrapped.db = a.db
		wrapped.inboxIRI = inboxIRI
		wrapped.newTransport = a.common.NewTransport
		wrapped.deliver = a.Deliver
		wrapped.addNewIds = a.AddNewIDs
		res, err := streams.NewTypeResolver(wrapped.callbacks(other)...)
		if err != nil {
			return err
		}
		if err = res.Resolve(c, activity); err != nil && !streams.IsUnmatchedErr(err) {
			return err
		} else if streams.IsUnmatchedErr(err) {
			err = a.s2s.DefaultCallback(c, activity)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// InboxForwarding implements the 3-part inbox forwarding algorithm specified in
// the ActivityPub specification. Does not modify the Activity, but may send
// outbound requests as a side effect.
//
// InboxForwarding sets the federated data in the database.
func (a *sideEffectActor) InboxForwarding(c context.Context, inboxIRI *url.URL, activity Activity) error {
	// 1. Must be first time we have seen this Activity.
	//
	// Obtain the id of the activity
	id := activity.GetJSONLDId()
	// Acquire a lock for the id. To be held for the rest of execution.
	err := a.db.Lock(c, id.Get())
	if err != nil {
		return err
	}
	// WARNING: Unlock is not deferred
	//
	// If the database already contains the activity, exit early.
	exists, err := a.db.Exists(c, id.Get())
	if err != nil {
		a.db.Unlock(c, id.Get())
		return err
	} else if exists {
		a.db.Unlock(c, id.Get())
		return nil
	}
	// Attempt to create the activity entry.
	err = a.db.Create(c, activity)
	if err != nil {
		a.db.Unlock(c, id.Get())
		return err
	}
	a.db.Unlock(c, id.Get())
	// Unlock by this point and in every branch above.
	//
	// 2. The values of 'to', 'cc', or 'audience' are Collections owned by
	//    this server.
	var r []*url.URL
	to := activity.GetActivityStreamsTo()
	if to != nil {
		for iter := to.Begin(); iter != to.End(); iter = iter.Next() {
			val, err := ToId(iter)
			if err != nil {
				return err
			}
			r = append(r, val)
		}
	}
	cc := activity.GetActivityStreamsCc()
	if cc != nil {
		for iter := cc.Begin(); iter != cc.End(); iter = iter.Next() {
			val, err := ToId(iter)
			if err != nil {
				return err
			}
			r = append(r, val)
		}
	}
	audience := activity.GetActivityStreamsAudience()
	if audience != nil {
		for iter := audience.Begin(); iter != audience.End(); iter = iter.Next() {
			val, err := ToId(iter)
			if err != nil {
				return err
			}
			r = append(r, val)
		}
	}
	// Find all IRIs owned by this server. We need to find all of them so
	// that forwarding can properly occur.
	var myIRIs []*url.URL
	for _, iri := range r {
		if err != nil {
			return err
		}
		err = a.db.Lock(c, iri)
		if err != nil {
			return err
		}
		// WARNING: Unlock is not deferred
		if owns, err := a.db.Owns(c, iri); err != nil {
			a.db.Unlock(c, iri)
			return err
		} else if !owns {
			a.db.Unlock(c, iri)
			continue
		}
		a.db.Unlock(c, iri)
		// Unlock by this point and in every branch above.
		myIRIs = append(myIRIs, iri)
	}
	// Finally, load our IRIs to determine if they are a Collection or
	// OrderedCollection.
	//
	// Load the unfiltered IRIs.
	var colIRIs []*url.URL
	col := make(map[string]itemser)
	oCol := make(map[string]orderedItemser)
	for _, iri := range myIRIs {
		err = a.db.Lock(c, iri)
		if err != nil {
			return err
		}
		// WARNING: Not Unlocked
		t, err := a.db.Get(c, iri)
		if err != nil {
			return err
		}
		if streams.IsOrExtendsActivityStreamsOrderedCollection(t) {
			if im, ok := t.(orderedItemser); ok {
				oCol[iri.String()] = im
				colIRIs = append(colIRIs, iri)
				defer a.db.Unlock(c, iri)
			} else {
				a.db.Unlock(c, iri)
			}
		} else if streams.IsOrExtendsActivityStreamsCollection(t) {
			if im, ok := t.(itemser); ok {
				col[iri.String()] = im
				colIRIs = append(colIRIs, iri)
				defer a.db.Unlock(c, iri)
			} else {
				a.db.Unlock(c, iri)
			}
		} else {
			a.db.Unlock(c, iri)
		}
	}
	// If we own none of the Collection IRIs in 'to', 'cc', or 'audience'
	// then no need to do inbox forwarding. We have nothing to forward to.
	if len(colIRIs) == 0 {
		return nil
	}
	// 3. The values of 'inReplyTo', 'object', 'target', or 'tag' are owned
	//    by this server. This is only a boolean trigger: As soon as we get
	//    a hit that we own something, then we should do inbox forwarding.
	maxDepth := a.s2s.MaxInboxForwardingRecursionDepth(c)
	ownsValue, err := a.hasInboxForwardingValues(c, inboxIRI, activity, maxDepth, 0)
	if err != nil {
		return err
	}
	// If we don't own any of the 'inReplyTo', 'object', 'target', or 'tag'
	// values, then no need to do inbox forwarding.
	if !ownsValue {
		return nil
	}
	// Do the inbox forwarding since the above conditions hold true. Support
	// the behavior of letting the application filter out the resulting
	// collections to be targeted.
	toSend, err := a.s2s.FilterForwarding(c, colIRIs, activity)
	if err != nil {
		return err
	}
	recipients := make([]*url.URL, 0, len(toSend))
	for _, iri := range toSend {
		if c, ok := col[iri.String()]; ok {
			if it := c.GetActivityStreamsItems(); it != nil {
				for iter := it.Begin(); iter != it.End(); iter = iter.Next() {
					id, err := ToId(iter)
					if err != nil {
						return err
					}
					recipients = append(recipients, id)
				}
			}
		} else if oc, ok := oCol[iri.String()]; ok {
			if oit := oc.GetActivityStreamsOrderedItems(); oit != nil {
				for iter := oit.Begin(); iter != oit.End(); iter = iter.Next() {
					id, err := ToId(iter)
					if err != nil {
						return err
					}
					recipients = append(recipients, id)
				}
			}
		}
	}
	return a.deliverToRecipients(c, inboxIRI, activity, recipients)
}

// PostOutbox handles the side effects of adding the activity to the actor's
// outbox, and triggering side effects based on the activity's type.
//
// This implementation assumes all types are meant to be delivered except for
// the ActivityStreams Block type.
func (a *sideEffectActor) PostOutbox(c context.Context, activity Activity, outboxIRI *url.URL, rawJSON map[string]interface{}) (deliverable bool, err error) {
	// TODO: Determine this if c2s is nil
	deliverable = true
	if a.c2s != nil {
		var wrapped SocialWrappedCallbacks
		var other []interface{}
		wrapped, other, err = a.c2s.SocialCallbacks(c)
		if err != nil {
			return
		}
		// Populate side channels.
		wrapped.db = a.db
		wrapped.outboxIRI = outboxIRI
		wrapped.rawActivity = rawJSON
		wrapped.clock = a.clock
		wrapped.newTransport = a.common.NewTransport
		undeliverable := false
		wrapped.undeliverable = &undeliverable
		var res *streams.TypeResolver
		res, err = streams.NewTypeResolver(wrapped.callbacks(other)...)
		if err != nil {
			return
		}
		if err = res.Resolve(c, activity); err != nil && !streams.IsUnmatchedErr(err) {
			return
		} else if streams.IsUnmatchedErr(err) {
			deliverable = true
			err = a.c2s.DefaultCallback(c, activity)
			if err != nil {
				return
			}
		} else {
			deliverable = !undeliverable
		}
	}
	err = a.addToOutbox(c, outboxIRI, activity)
	return
}

// AddNewIDs creates new 'id' entries on an activity and its objects if it is a
// Create activity.
func (a *sideEffectActor) AddNewIDs(c context.Context, activity Activity) error {
	id, err := a.db.NewID(c, activity)
	if err != nil {
		return err
	}
	activityId := streams.NewJSONLDIdProperty()
	activityId.Set(id)
	activity.SetJSONLDId(activityId)
	if streams.IsOrExtendsActivityStreamsCreate(activity) {
		o, ok := activity.(objecter)
		if !ok {
			return fmt.Errorf("cannot add new id for Create: %T has no object property", activity)
		}
		if oProp := o.GetActivityStreamsObject(); oProp != nil {
			for iter := oProp.Begin(); iter != oProp.End(); iter = iter.Next() {
				t := iter.GetType()
				if t == nil {
					return fmt.Errorf("cannot add new id for object in Create: object is not embedded as a value literal")
				}
				id, err = a.db.NewID(c, t)
				if err != nil {
					return err
				}
				objId := streams.NewJSONLDIdProperty()
				objId.Set(id)
				t.SetJSONLDId(objId)
			}
		}
	}
	return nil
}

// deliver will complete the peer-to-peer sending of a federated message to
// another server.
//
// Must be called if at least the federated protocol is supported.
func (a *sideEffectActor) Deliver(c context.Context, outboxIRI *url.URL, activity Activity) error {
	recipients, err := a.prepare(c, outboxIRI, activity)
	if err != nil {
		return err
	}
	return a.deliverToRecipients(c, outboxIRI, activity, recipients)
}

// WrapInCreate wraps an object with a Create activity.
func (a *sideEffectActor) WrapInCreate(c context.Context, obj vocab.Type, outboxIRI *url.URL) (create vocab.ActivityStreamsCreate, err error) {
	err = a.db.Lock(c, outboxIRI)
	if err != nil {
		return
	}
	// WARNING: No deferring the Unlock
	actorIRI, err := a.db.ActorForOutbox(c, outboxIRI)
	if err != nil {
		a.db.Unlock(c, outboxIRI)
		return
	}
	a.db.Unlock(c, outboxIRI)
	// Unlock the lock at this point and every branch above
	return wrapInCreate(c, obj, actorIRI)
}

// deliverToRecipients will take a prepared Activity and send it to specific
// recipients on behalf of an actor.
func (a *sideEffectActor) deliverToRecipients(c context.Context, boxIRI *url.URL, activity Activity, recipients []*url.URL) error {
	m, err := streams.Serialize(activity)
	if err != nil {
		return err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	tp, err := a.common.NewTransport(c, boxIRI, goFedUserAgent())
	if err != nil {
		return err
	}
	return tp.BatchDeliver(c, b, recipients)
}

// addToOutbox adds the activity to the outbox and creates the activity in the
// internal database as its own entry.
func (a *sideEffectActor) addToOutbox(c context.Context, outboxIRI *url.URL, activity Activity) error {
	// Set the activity in the database first.
	id := activity.GetJSONLDId()
	err := a.db.Lock(c, id.Get())
	if err != nil {
		return err
	}
	// WARNING: Unlock not deferred
	err = a.db.Create(c, activity)
	if err != nil {
		a.db.Unlock(c, id.Get())
		return err
	}
	a.db.Unlock(c, id.Get())
	// WARNING: Unlock(c, id) should be called by this point and in every
	// return before here.
	//
	// Acquire a lock to read the outbox. Defer release.
	err = a.db.Lock(c, outboxIRI)
	if err != nil {
		return err
	}
	defer a.db.Unlock(c, outboxIRI)
	outbox, err := a.db.GetOutbox(c, outboxIRI)
	if err != nil {
		return err
	}
	// Prepend the activity to the list of 'orderedItems'.
	oi := outbox.GetActivityStreamsOrderedItems()
	if oi == nil {
		oi = streams.NewActivityStreamsOrderedItemsProperty()
	}
	oi.PrependIRI(id.Get())
	outbox.SetActivityStreamsOrderedItems(oi)
	// Save in the database.
	err = a.db.SetOutbox(c, outbox)
	return err
}

// addToInboxIfNew will add the activity to the inbox at the specified IRI if
// the activity's ID has not yet been added to the inbox.
//
// It does not add the activity to this database's know federated data.
//
// Returns true when the activity is novel.
func (a *sideEffectActor) addToInboxIfNew(c context.Context, inboxIRI *url.URL, activity Activity) (isNew bool, err error) {
	// Acquire a lock to read the inbox. Defer release.
	err = a.db.Lock(c, inboxIRI)
	if err != nil {
		return
	}
	defer a.db.Unlock(c, inboxIRI)
	// Obtain the id of the activity
	id := activity.GetJSONLDId()
	// If the inbox already contains the URL, early exit.
	contains, err := a.db.InboxContains(c, inboxIRI, id.Get())
	if err != nil {
		return
	} else if contains {
		return
	}
	// It is a new id, acquire the inbox.
	isNew = true
	inbox, err := a.db.GetInbox(c, inboxIRI)
	if err != nil {
		return
	}
	// Prepend the activity to the list of 'orderedItems'.
	oi := inbox.GetActivityStreamsOrderedItems()
	if oi == nil {
		oi = streams.NewActivityStreamsOrderedItemsProperty()
	}
	oi.PrependIRI(id.Get())
	inbox.SetActivityStreamsOrderedItems(oi)
	// Save in the database.
	err = a.db.SetInbox(c, inbox)
	return
}

// Given an ActivityStreams value, recursively examines ownership of the id or
// href and the ones on properties applicable to inbox forwarding.
//
// Recursion may be limited by providing a 'maxDepth' greater than zero. A
// value of zero or a negative number will result in infinite recursion.
func (a *sideEffectActor) hasInboxForwardingValues(c context.Context, inboxIRI *url.URL, val vocab.Type, maxDepth, currDepth int) (bool, error) {
	// Stop recurring if we are exceeding the maximum depth and the maximum
	// is a positive number.
	if maxDepth > 0 && currDepth >= maxDepth {
		return false, nil
	}
	// Determine if we own the 'id' of any values on the properties we care
	// about.
	types, iris := getInboxForwardingValues(val)
	// For IRIs, simply check if we own them.
	for _, iri := range iris {
		err := a.db.Lock(c, iri)
		if err != nil {
			return false, err
		}
		// WARNING: Unlock is not deferred
		if owns, err := a.db.Owns(c, iri); err != nil {
			a.db.Unlock(c, iri)
			return false, err
		} else if owns {
			a.db.Unlock(c, iri)
			return true, nil
		}
		a.db.Unlock(c, iri)
		// Unlock by this point and in every branch above
	}
	// For embedded literals, check the id.
	for _, val := range types {
		id, err := GetId(val)
		if err != nil {
			return false, err
		}
		err = a.db.Lock(c, id)
		if err != nil {
			return false, err
		}
		// WARNING: Unlock is not deferred
		if owns, err := a.db.Owns(c, id); err != nil {
			a.db.Unlock(c, id)
			return false, err
		} else if owns {
			a.db.Unlock(c, id)
			return true, nil
		}
		a.db.Unlock(c, id)
		// Unlock by this point and in every branch above
	}
	// Recur Preparation: Try fetching the IRIs so we can recur into them.
	for _, iri := range iris {
		// Dereferencing the IRI.
		tport, err := a.common.NewTransport(c, inboxIRI, goFedUserAgent())
		if err != nil {
			return false, err
		}
		b, err := tport.Dereference(c, iri)
		if err != nil {
			// Do not fail the entire process if the data is
			// missing.
			continue
		}
		var m map[string]interface{}
		if err = json.Unmarshal(b, &m); err != nil {
			return false, err
		}
		t, err := streams.ToType(c, m)
		if err != nil {
			// Do not fail the entire process if we cannot handle
			// the type.
			continue
		}
		types = append(types, t)
	}
	// Recur.
	for _, nextVal := range types {
		if has, err := a.hasInboxForwardingValues(c, inboxIRI, nextVal, maxDepth, currDepth+1); err != nil {
			return false, err
		} else if has {
			return true, nil
		}
	}
	return false, nil
}

// prepare takes a deliverableObject and returns a list of the proper recipient
// target URIs. Additionally, the deliverableObject will have any hidden
// hidden recipients ("bto" and "bcc") stripped from it.
//
// Only call if both the social and federated protocol are supported.
func (a *sideEffectActor) prepare(c context.Context, outboxIRI *url.URL, activity Activity) (r []*url.URL, err error) {
	// Get inboxes of recipients
	if to := activity.GetActivityStreamsTo(); to != nil {
		for iter := to.Begin(); iter != to.End(); iter = iter.Next() {
			var val *url.URL
			val, err = ToId(iter)
			if err != nil {
				return
			}
			r = append(r, val)
		}
	}
	if bto := activity.GetActivityStreamsBto(); bto != nil {
		for iter := bto.Begin(); iter != bto.End(); iter = iter.Next() {
			var val *url.URL
			val, err = ToId(iter)
			if err != nil {
				return
			}
			r = append(r, val)
		}
	}
	if cc := activity.GetActivityStreamsCc(); cc != nil {
		for iter := cc.Begin(); iter != cc.End(); iter = iter.Next() {
			var val *url.URL
			val, err = ToId(iter)
			if err != nil {
				return
			}
			r = append(r, val)
		}
	}
	if bcc := activity.GetActivityStreamsBcc(); bcc != nil {
		for iter := bcc.Begin(); iter != bcc.End(); iter = iter.Next() {
			var val *url.URL
			val, err = ToId(iter)
			if err != nil {
				return
			}
			r = append(r, val)
		}
	}
	if audience := activity.GetActivityStreamsAudience(); audience != nil {
		for iter := audience.Begin(); iter != audience.End(); iter = iter.Next() {
			var val *url.URL
			val, err = ToId(iter)
			if err != nil {
				return
			}
			r = append(r, val)
		}
	}
	// 1. When an object is being delivered to the originating actor's
	//    followers, a server MAY reduce the number of receiving actors
	//    delivered to by identifying all followers which share the same
	//    sharedInbox who would otherwise be individual recipients and
	//    instead deliver objects to said sharedInbox.
	// 2. If an object is addressed to the Public special collection, a
	//    server MAY deliver that object to all known sharedInbox endpoints
	//    on the network.
	r = filterURLs(r, IsPublic)
	t, err := a.common.NewTransport(c, outboxIRI, goFedUserAgent())
	if err != nil {
		return nil, err
	}
	receiverActors, err := a.resolveInboxes(c, t, r, 0, a.s2s.MaxDeliveryRecursionDepth(c))
	if err != nil {
		return nil, err
	}
	targets, err := getInboxes(receiverActors)
	if err != nil {
		return nil, err
	}
	// Get inboxes of sender.
	err = a.db.Lock(c, outboxIRI)
	if err != nil {
		return
	}
	// WARNING: No deferring the Unlock
	actorIRI, err := a.db.ActorForOutbox(c, outboxIRI)
	if err != nil {
		a.db.Unlock(c, outboxIRI)
		return
	}
	a.db.Unlock(c, outboxIRI)
	// Get the inbox on the sender.
	err = a.db.Lock(c, actorIRI)
	if err != nil {
		return nil, err
	}
	// BEGIN LOCK
	thisActor, err := a.db.Get(c, actorIRI)
	a.db.Unlock(c, actorIRI)
	// END LOCK -- Still need to handle err
	if err != nil {
		return nil, err
	}
	// Post-processing
	var ignore *url.URL
	ignore, err = getInbox(thisActor)
	if err != nil {
		return nil, err
	}
	r = dedupeIRIs(targets, []*url.URL{ignore})
	stripHiddenRecipients(activity)
	return r, nil
}

// resolveInboxes takes a list of Actor id URIs and returns them as concrete
// instances of actorObject. It attempts to apply recursively when it encounters
// a target that is a Collection or OrderedCollection.
//
// If maxDepth is zero or negative, then recursion is infinitely applied.
//
// If a recipient is a Collection or OrderedCollection, then the server MUST
// dereference the collection, WITH the user's credentials.
//
// Note that this also applies to CollectionPage and OrderedCollectionPage.
func (a *sideEffectActor) resolveInboxes(c context.Context, t Transport, r []*url.URL, depth, maxDepth int) (actors []vocab.Type, err error) {
	if maxDepth > 0 && depth >= maxDepth {
		return
	}
	for _, u := range r {
		var act vocab.Type
		var more []*url.URL
		// TODO: Determine if more logic is needed here for inaccessible
		// collections owned by peer servers.
		act, more, err = a.dereferenceForResolvingInboxes(c, t, u)
		if err != nil {
			// Missing recipient -- skip.
			continue
		}
		var recurActors []vocab.Type
		recurActors, err = a.resolveInboxes(c, t, more, depth+1, maxDepth)
		if err != nil {
			return
		}
		if act != nil {
			actors = append(actors, act)
		}
		actors = append(actors, recurActors...)
	}
	return
}

// dereferenceForResolvingInboxes dereferences an IRI solely for finding an
// actor's inbox IRI to deliver to.
//
// The returned actor could be nil, if it wasn't an actor (ex: a Collection or
// OrderedCollection).
func (a *sideEffectActor) dereferenceForResolvingInboxes(c context.Context, t Transport, actorIRI *url.URL) (actor vocab.Type, moreActorIRIs []*url.URL, err error) {
	var resp []byte
	resp, err = t.Dereference(c, actorIRI)
	if err != nil {
		return
	}
	var m map[string]interface{}
	if err = json.Unmarshal(resp, &m); err != nil {
		return
	}
	actor, err = streams.ToType(c, m)
	if err != nil {
		return
	}
	// Attempt to see if the 'actor' is really some sort of type that has
	// an 'items' or 'orderedItems' property.
	if v, ok := actor.(itemser); ok {
		if i := v.GetActivityStreamsItems(); i != nil {
			for iter := i.Begin(); iter != i.End(); iter = iter.Next() {
				var id *url.URL
				id, err = ToId(iter)
				if err != nil {
					return
				}
				moreActorIRIs = append(moreActorIRIs, id)
			}
		}
		actor = nil
	} else if v, ok := actor.(orderedItemser); ok {
		if i := v.GetActivityStreamsOrderedItems(); i != nil {
			for iter := i.Begin(); iter != i.End(); iter = iter.Next() {
				var id *url.URL
				id, err = ToId(iter)
				if err != nil {
					return
				}
				moreActorIRIs = append(moreActorIRIs, id)
			}
		}
		actor = nil
	}
	return
}
