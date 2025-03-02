package pub

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"codeberg.org/superseriousbusiness/activity/streams"
	"codeberg.org/superseriousbusiness/activity/streams/vocab"
)

// sideEffectActor must satisfy the DelegateActor interface.
var _ DelegateActor = &SideEffectActor{}

// SideEffectActor is a DelegateActor that handles the ActivityPub
// implementation side effects, but requires a more opinionated application to
// be written.
//
// Note that when using the SideEffectActor with an application that good-faith
// implements its required interfaces, the ActivityPub specification is
// guaranteed to be correctly followed.
type SideEffectActor struct {
	// When doing deliveries to remote servers via the s2s protocol, the side effect
	// actor will by default use the Serialize function from the streams package.
	// However, this can be overridden after the side effect actor is intantiated,
	// by setting the exposed Serialize function on the struct. For example:
	//
	//	a := NewSideEffectActor(...)
	//	a.Serialize = func(a vocab.Type) (m map[string]interface{}, e error) {
	//	  // Put your custom serializer logic here.
	//	}
	//
	// Note that you should only do this *immediately* after instantiating the side
	// effect actor -- never while your application is already running, as this will
	// likely cause race conditions or other problems! In most cases, you will never
	// need to change this; it's provided solely to allow easier customization by
	// applications.
	Serialize func(a vocab.Type) (m map[string]interface{}, e error)

	// When doing deliveries to remote servers via the s2s protocol, it may be desirable
	// for implementations to be able to pre-sort recipients so that higher-priority
	// recipients are higher up in the delivery queue, and lower-priority recipients
	// are further down. This can be achieved by setting the DeliveryRecipientPreSort
	// function on the side effect actor after it's instantiated. For example:
	//
	//	a := NewSideEffectActor(...)
	//	a.DeliveryRecipientPreSort = func(actorAndCollectionIRIs []*url.URL) []*url.URL {
	//	  // Put your sorting logic here.
	//	}
	//
	// The actorAndCollectionIRIs parameter will be the initial list of IRIs derived by
	// looking at the "to", "cc", "bto", "bcc", and "audience" properties of the activity
	// being delivered, excluding the AP public IRI, and before dereferencing of inboxes.
	// It may look something like this:
	//
	//	[
	//		"https://example.org/users/someone/followers",     // <-- collection IRI
	//		"https://another.example.org/users/someone_else",  // <-- actor IRI
	//		"[...]"                                            // <-- etc
	//	]
	//
	// In this case, implementers may wish to sort the slice so that the directly-addressed
	// actor "https://another.example.org/users/someone_else" occurs at an earlier index in
	// the slice than the followers collection "https://example.org/users/someone/followers",
	// so that "@someone_else" receives the delivery first.
	//
	// Note that you should only do this *immediately* after instantiating the side
	// effect actor -- never while your application is already running, as this will
	// likely cause race conditions or other problems! It's also completely fine to not
	// set this function at all -- in this case, no pre-sorting of recipients will be
	// performed, and delivery will occur in a non-determinate order.
	DeliveryRecipientPreSort func(actorAndCollectionIRIs []*url.URL) []*url.URL

	common CommonBehavior
	s2s    FederatingProtocol
	c2s    SocialProtocol
	db     Database
	clock  Clock
}

// NewSideEffectActor returns a new SideEffectActor, which satisfies the
// DelegateActor interface. Most of the time you will not need to call this
// function, and should instead rely on the NewSocialActor, NewFederatingActor,
// and NewActor functions, all of which use a SideEffectActor under the hood.
// Nevertheless, this function is exposed in case application developers need
// a SideEffectActor for some other reason (tests, monkey patches, etc).
//
// If you are using the returned SideEffectActor for federation, ensure that s2s
// is not nil. Likewise, if you are using it for the social protocol, ensure
// that c2s is not nil.
func NewSideEffectActor(c CommonBehavior,
	s2s FederatingProtocol,
	c2s SocialProtocol,
	db Database,
	clock Clock) *SideEffectActor {
	return &SideEffectActor{
		Serialize: streams.Serialize,
		common:    c,
		s2s:       s2s,
		c2s:       c2s,
		db:        db,
		clock:     clock,
	}
}

// PostInboxRequestBodyHook defers to the delegate.
func (a *SideEffectActor) PostInboxRequestBodyHook(c context.Context, r *http.Request, activity Activity) (context.Context, error) {
	return a.s2s.PostInboxRequestBodyHook(c, r, activity)
}

// PostOutboxRequestBodyHook defers to the delegate.
func (a *SideEffectActor) PostOutboxRequestBodyHook(c context.Context, r *http.Request, data vocab.Type) (context.Context, error) {
	return a.c2s.PostOutboxRequestBodyHook(c, r, data)
}

// AuthenticatePostInbox defers to the delegate to authenticate the request.
func (a *SideEffectActor) AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return a.s2s.AuthenticatePostInbox(c, w, r)
}

// AuthenticateGetInbox defers to the delegate to authenticate the request.
func (a *SideEffectActor) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return a.common.AuthenticateGetInbox(c, w, r)
}

// AuthenticatePostOutbox defers to the delegate to authenticate the request.
func (a *SideEffectActor) AuthenticatePostOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return a.c2s.AuthenticatePostOutbox(c, w, r)
}

// AuthenticateGetOutbox defers to the delegate to authenticate the request.
func (a *SideEffectActor) AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return a.common.AuthenticateGetOutbox(c, w, r)
}

// GetOutbox delegates to the SocialProtocol.
func (a *SideEffectActor) GetOutbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	return a.common.GetOutbox(c, r)
}

// GetInbox delegates to the FederatingProtocol.
func (a *SideEffectActor) GetInbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	return a.s2s.GetInbox(c, r)
}

// AuthorizePostInbox defers to the federating protocol whether the peer request
// is authorized based on the actors' ids.
func (a *SideEffectActor) AuthorizePostInbox(c context.Context, w http.ResponseWriter, activity Activity) (authorized bool, err error) {
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
func (a *SideEffectActor) PostInbox(c context.Context, inboxIRI *url.URL, activity Activity) error {
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
func (a *SideEffectActor) InboxForwarding(c context.Context, inboxIRI *url.URL, activity Activity) error {
	// 1. Must be first time we have seen this Activity.
	//
	// Obtain the id of the activity
	id := activity.GetJSONLDId()
	// Acquire a lock for the id. To be held for the rest of execution.
	unlock, err := a.db.Lock(c, id.Get())
	if err != nil {
		return err
	}
	// WARNING: Unlock is not deferred
	//
	// If the database already contains the activity, exit early.
	exists, err := a.db.Exists(c, id.Get())
	if err != nil {
		unlock()
		return err
	} else if exists {
		unlock()
		return nil
	}
	// Attempt to create the activity entry.
	err = a.db.Create(c, activity)
	unlock() // unlock even on error return
	if err != nil {
		return err
	}
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
		var unlock func()
		unlock, err = a.db.Lock(c, iri)
		if err != nil {
			return err
		}
		// WARNING: Unlock is not deferred
		owns, err := a.db.Owns(c, iri)
		unlock() // unlock even on error
		if err != nil {
			return err
		} else if !owns {
			continue
		}
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
		var unlock func()
		unlock, err = a.db.Lock(c, iri)
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
				defer unlock()
			} else {
				unlock() // unlock instantly
			}
		} else if streams.IsOrExtendsActivityStreamsCollection(t) {
			if im, ok := t.(itemser); ok {
				col[iri.String()] = im
				colIRIs = append(colIRIs, iri)
				defer unlock()
			} else {
				unlock() // unlock instantly
			}
		} else {
			unlock() // unlock instantly
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
func (a *SideEffectActor) PostOutbox(c context.Context, activity Activity, outboxIRI *url.URL, rawJSON map[string]interface{}) (deliverable bool, err error) {
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
func (a *SideEffectActor) AddNewIDs(c context.Context, activity Activity) error {
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
func (a *SideEffectActor) Deliver(c context.Context, outboxIRI *url.URL, activity Activity) error {
	recipients, err := a.prepare(c, outboxIRI, activity)
	if err != nil {
		return err
	}
	return a.deliverToRecipients(c, outboxIRI, activity, recipients)
}

// WrapInCreate wraps an object with a Create activity.
func (a *SideEffectActor) WrapInCreate(c context.Context, obj vocab.Type, outboxIRI *url.URL) (create vocab.ActivityStreamsCreate, err error) {
	var unlock func()
	unlock, err = a.db.Lock(c, outboxIRI)
	if err != nil {
		return
	}
	// WARNING: No deferring the Unlock
	actorIRI, err := a.db.ActorForOutbox(c, outboxIRI)
	unlock() // unlock after regardless
	if err != nil {
		return
	}
	// Unlock the lock at this point and every branch above
	return wrapInCreate(c, obj, actorIRI)
}

// deliverToRecipients will take a prepared Activity and send it to specific
// recipients on behalf of an actor.
func (a *SideEffectActor) deliverToRecipients(c context.Context, boxIRI *url.URL, activity Activity, recipients []*url.URL) error {
	// Call whichever serializer is
	// set on the side effect actor.
	m, err := a.Serialize(activity)
	if err != nil {
		return err
	}

	tp, err := a.common.NewTransport(c, boxIRI, goFedUserAgent())
	if err != nil {
		return err
	}

	return tp.BatchDeliver(c, m, recipients)
}

// addToOutbox adds the activity to the outbox and creates the activity in the
// internal database as its own entry.
func (a *SideEffectActor) addToOutbox(c context.Context, outboxIRI *url.URL, activity Activity) error {
	// Set the activity in the database first.
	id := activity.GetJSONLDId()
	unlock, err := a.db.Lock(c, id.Get())
	if err != nil {
		return err
	}
	// WARNING: Unlock not deferred
	err = a.db.Create(c, activity)
	unlock() // unlock after regardless
	if err != nil {
		return err
	}
	// WARNING: Unlock(c, id) should be called by this point and in every
	// return before here.
	//
	// Acquire a lock to read the outbox. Defer release.
	unlock, err = a.db.Lock(c, outboxIRI)
	if err != nil {
		return err
	}
	defer unlock()
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
func (a *SideEffectActor) addToInboxIfNew(c context.Context, inboxIRI *url.URL, activity Activity) (isNew bool, err error) {
	// Acquire a lock to read the inbox. Defer release.
	var unlock func()
	unlock, err = a.db.Lock(c, inboxIRI)
	if err != nil {
		return
	}
	defer unlock()
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
func (a *SideEffectActor) hasInboxForwardingValues(c context.Context, inboxIRI *url.URL, val vocab.Type, maxDepth, currDepth int) (bool, error) {
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
		unlock, err := a.db.Lock(c, iri)
		if err != nil {
			return false, err
		}
		// WARNING: Unlock is not deferred
		owns, err := a.db.Owns(c, iri)
		unlock() // unlock after regardless
		if err != nil {
			return false, err
		} else if owns {
			return true, nil
		}
		// Unlock by this point and in every branch above
	}
	// For embedded literals, check the id.
	for _, val := range types {
		id, err := GetId(val)
		if err != nil {
			return false, err
		}
		var unlock func()
		unlock, err = a.db.Lock(c, id)
		if err != nil {
			return false, err
		}
		// WARNING: Unlock is not deferred
		owns, err := a.db.Owns(c, id)
		unlock() // unlock after regardless
		if err != nil {
			return false, err
		} else if owns {
			return true, nil
		}
		// Unlock by this point and in every branch above
	}
	// Recur Preparation: Try fetching the IRIs so we can recur into them.
	for _, iri := range iris {
		// Dereferencing the IRI.
		tport, err := a.common.NewTransport(c, inboxIRI, goFedUserAgent())
		if err != nil {
			return false, err
		}
		resp, err := tport.Dereference(c, iri)
		if err != nil {
			// Do not fail the entire process if the data is
			// missing.
			continue
		}
		m, err := readActivityPubResp(resp)
		if err != nil {
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

// prepare takes a deliverableObject and returns a list of the final
// recipient inbox IRIs. Additionally, the deliverableObject will have
// any hidden hidden recipients ("bto" and "bcc") stripped from it.
//
// Only call if both the social and federated protocol are supported.
func (a *SideEffectActor) prepare(
	ctx context.Context,
	outboxIRI *url.URL,
	activity Activity,
) ([]*url.URL, error) {
	// Iterate through to, bto, cc, bcc, and audience
	// to extract a slice of addressee IRIs / IDs.
	//
	// The resulting slice might look something like:
	//
	//	[
	//		"https://example.org/users/someone/followers",     // <-- collection IRI
	//		"https://another.example.org/users/someone_else",  // <-- actor IRI
	//		"[...]"                                            // <-- etc
	//	]
	var actorsAndCollections []*url.URL
	if to := activity.GetActivityStreamsTo(); to != nil {
		for iter := to.Begin(); iter != to.End(); iter = iter.Next() {
			var err error
			actorsAndCollections, err = appendToActorsAndCollectionsIRIs(
				iter, actorsAndCollections,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	if bto := activity.GetActivityStreamsBto(); bto != nil {
		for iter := bto.Begin(); iter != bto.End(); iter = iter.Next() {
			var err error
			actorsAndCollections, err = appendToActorsAndCollectionsIRIs(
				iter, actorsAndCollections,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	if cc := activity.GetActivityStreamsCc(); cc != nil {
		for iter := cc.Begin(); iter != cc.End(); iter = iter.Next() {
			var err error
			actorsAndCollections, err = appendToActorsAndCollectionsIRIs(
				iter, actorsAndCollections,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	if bcc := activity.GetActivityStreamsBcc(); bcc != nil {
		for iter := bcc.Begin(); iter != bcc.End(); iter = iter.Next() {
			var err error
			actorsAndCollections, err = appendToActorsAndCollectionsIRIs(
				iter, actorsAndCollections,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	if audience := activity.GetActivityStreamsAudience(); audience != nil {
		for iter := audience.Begin(); iter != audience.End(); iter = iter.Next() {
			var err error
			actorsAndCollections, err = appendToActorsAndCollectionsIRIs(
				iter, actorsAndCollections,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	// PRE-SORTING

	// If the pre-delivery sort function is defined,
	// call it now so that implementations can sort
	// delivery order to their preferences.
	if a.DeliveryRecipientPreSort != nil {
		actorsAndCollections = a.DeliveryRecipientPreSort(actorsAndCollections)
	}

	// We now need to dereference the actor or collection
	// IRIs to derive inboxes that we can POST requests to.
	var (
		inboxes       = make([]*url.URL, 0, len(actorsAndCollections))
		derefdEntries = make(map[string]struct{}, len(actorsAndCollections))
	)

	// First check if the implemented database logic
	// can return any of these inboxes without having
	// to make remote dereference calls (much cheaper).
	for _, actorOrCollection := range actorsAndCollections {
		actorOrCollectionStr := actorOrCollection.String()
		if _, derefd := derefdEntries[actorOrCollectionStr]; derefd {
			// Ignore potential duplicates
			// we've already derefd to inbox(es).
			continue
		}

		// BEGIN LOCK
		unlock, err := a.db.Lock(ctx, actorOrCollection)
		if err != nil {
			return nil, err
		}

		// Try to get inbox(es) for this actor or collection.
		gotInboxes, err := a.db.InboxesForIRI(ctx, actorOrCollection)

		// END LOCK
		unlock()

		if err != nil {
			return nil, err
		}

		if len(gotInboxes) == 0 {
			// No hit(s).
			continue
		}

		// We have one or more hits.
		inboxes = append(inboxes, gotInboxes...)

		// Mark this actor or collection as deref'd.
		derefdEntries[actorOrCollectionStr] = struct{}{}
	}

	// Now look for any remaining actors/collections
	// that weren't already dereferenced into inboxes
	// with db calls; find these by making deref calls
	// to remote instances.
	//
	// First get a transport to do the http calls.
	t, err := a.common.NewTransport(ctx, outboxIRI, goFedUserAgent())
	if err != nil {
		return nil, err
	}

	// Make HTTP calls to unpack collection IRIs into
	// Actor IRIs and then into Actor types, ignoring
	// actors or collections we've already deref'd.
	actorsFromRemote, err := a.resolveActors(
		ctx,
		t,
		actorsAndCollections,
		derefdEntries,
		0, a.s2s.MaxDeliveryRecursionDepth(ctx),
	)
	if err != nil {
		return nil, err
	}

	// Release no-longer-needed collections.
	clear(derefdEntries)
	clear(actorsAndCollections)

	// Extract inbox IRI from each deref'd Actor (if any).
	inboxesFromRemote, err := actorsToInboxIRIs(actorsFromRemote)
	if err != nil {
		return nil, err
	}

	// Combine db-discovered inboxes and remote-discovered
	// inboxes into a final list of destination inboxes.
	inboxes = append(inboxes, inboxesFromRemote...)

	// POST FILTERING

	// Do a final pass of the inboxes to:
	//
	// 1. Deduplicate entries.
	// 2. Ensure that the list of inboxes doesn't
	// contain the inbox of whoever the outbox
	// belongs to, no point delivering to oneself.
	//
	// To do this we first need to get the
	// inbox IRI of this outbox's Actor.

	// BEGIN LOCK
	unlock, err := a.db.Lock(ctx, outboxIRI)
	if err != nil {
		return nil, err
	}

	// Get the IRI of the Actor who owns this outbox.
	outboxActorIRI, err := a.db.ActorForOutbox(ctx, outboxIRI)

	// END LOCK
	unlock()

	if err != nil {
		return nil, err
	}

	// BEGIN LOCK
	unlock, err = a.db.Lock(ctx, outboxActorIRI)
	if err != nil {
		return nil, err
	}

	// Now get the Actor who owns this outbox.
	outboxActor, err := a.db.Get(ctx, outboxActorIRI)

	// END LOCK
	unlock()

	if err != nil {
		return nil, err
	}

	// Extract the inbox IRI for the outbox Actor.
	inboxOfOutboxActor, err := getInbox(outboxActor)
	if err != nil {
		return nil, err
	}

	// Deduplicate the final inboxes slice, and filter
	// out of the inbox of this outbox actor (if present).
	inboxes = filterInboxIRIs(inboxes, inboxOfOutboxActor)

	// Now that we've derived inboxes to deliver
	// the activity to, strip off any bto or bcc
	// recipients, as per the AP spec requirements.
	stripHiddenRecipients(activity)

	// All done!
	return inboxes, nil
}

// resolveActors takes a list of Actor id URIs and returns them as concrete
// instances of actorObject. It attempts to apply recursively when it encounters
// a target that is a Collection or OrderedCollection.
//
// Any IRI strings in the ignores map will be skipped (use this when
// you've already dereferenced some of the actorAndCollectionIRIs).
//
// If maxDepth is zero or negative, then recursion is infinitely applied.
//
// If a recipient is a Collection or OrderedCollection, then the server MUST
// dereference the collection, WITH the user's credentials.
//
// Note that this also applies to CollectionPage and OrderedCollectionPage.
func (a *SideEffectActor) resolveActors(
	ctx context.Context,
	t Transport,
	actorAndCollectionIRIs []*url.URL,
	ignores map[string]struct{},
	depth, maxDepth int,
) ([]vocab.Type, error) {
	if maxDepth > 0 && depth >= maxDepth {
		// Hit our max depth.
		return nil, nil
	}

	if len(actorAndCollectionIRIs) == 0 {
		// Nothing to do.
		return nil, nil
	}

	// Optimistically assume 1:1 mapping of IRIs to actors.
	actors := make([]vocab.Type, 0, len(actorAndCollectionIRIs))

	// Deref each actorOrCollectionIRI if not ignored.
	for _, actorOrCollectionIRI := range actorAndCollectionIRIs {
		_, ignore := ignores[actorOrCollectionIRI.String()]
		if ignore {
			// Don't try to
			// deref this one.
			continue
		}

		// TODO: Determine if more logic is needed here for
		// inaccessible collections owned by peer servers.
		actor, more, err := a.dereferenceForResolvingInboxes(ctx, t, actorOrCollectionIRI)
		if err != nil {
			// Missing recipient -- skip.
			continue
		}

		if actor != nil {
			// Got a hit.
			actors = append(actors, actor)
		}

		// If this was a collection, get more.
		recurActors, err := a.resolveActors(
			ctx,
			t,
			more,
			ignores,
			depth+1, maxDepth,
		)
		if err != nil {
			return nil, err
		}

		actors = append(actors, recurActors...)
	}

	return actors, nil
}

// dereferenceForResolvingInboxes dereferences an IRI solely for finding an
// actor's inbox IRI to deliver to.
//
// The returned actor could be nil, if it wasn't an actor (ex: a Collection or
// OrderedCollection).
func (a *SideEffectActor) dereferenceForResolvingInboxes(c context.Context, t Transport, actorIRI *url.URL) (actor vocab.Type, moreActorIRIs []*url.URL, err error) {
	var resp *http.Response
	resp, err = t.Dereference(c, actorIRI)
	if err != nil {
		return
	}
	var m map[string]interface{}
	m, err = readActivityPubResp(resp)
	if err != nil {
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
