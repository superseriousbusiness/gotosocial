package pub

import (
	"context"
	"fmt"
	"net/url"

	"codeberg.org/superseriousbusiness/activity/streams"
	"codeberg.org/superseriousbusiness/activity/streams/vocab"
)

// OnFollowBehavior enumerates the different default actions that the go-fed
// library can provide when receiving a Follow Activity from a peer.
type OnFollowBehavior int

const (
	// OnFollowDoNothing does not take any action when a Follow Activity
	// is received.
	OnFollowDoNothing OnFollowBehavior = iota
	// OnFollowAutomaticallyAccept triggers the side effect of sending an
	// Accept of this Follow request in response.
	OnFollowAutomaticallyAccept
	// OnFollowAutomaticallyAccept triggers the side effect of sending a
	// Reject of this Follow request in response.
	OnFollowAutomaticallyReject
)

// FederatingWrappedCallbacks lists the callback functions that already have
// some side effect behavior provided by the pub library.
//
// These functions are wrapped for the Federating Protocol.
type FederatingWrappedCallbacks struct {
	// Create handles additional side effects for the Create ActivityStreams
	// type, specific to the application using go-fed.
	//
	// The wrapping callback for the Federating Protocol ensures the
	// 'object' property is created in the database.
	//
	// Create calls Create for each object in the federated Activity.
	Create func(context.Context, vocab.ActivityStreamsCreate) error
	// Update handles additional side effects for the Update ActivityStreams
	// type, specific to the application using go-fed.
	//
	// The wrapping callback for the Federating Protocol ensures the
	// 'object' property is updated in the database.
	//
	// Update calls Update on the federated entry from the database, with a
	// new value.
	Update func(context.Context, vocab.ActivityStreamsUpdate) error
	// Delete handles additional side effects for the Delete ActivityStreams
	// type, specific to the application using go-fed.
	//
	// Delete removes the federated entry from the database.
	Delete func(context.Context, vocab.ActivityStreamsDelete) error
	// Follow handles additional side effects for the Follow ActivityStreams
	// type, specific to the application using go-fed.
	//
	// The wrapping function can have one of several default behaviors,
	// depending on the value of the OnFollow setting.
	Follow func(context.Context, vocab.ActivityStreamsFollow) error
	// OnFollow determines what action to take for this particular callback
	// if a Follow Activity is handled.
	OnFollow OnFollowBehavior
	// Accept handles additional side effects for the Accept ActivityStreams
	// type, specific to the application using go-fed.
	//
	// The wrapping function determines if this 'Accept' is in response to a
	// 'Follow'. If so, then the 'actor' is added to the original 'actor's
	// 'following' collection.
	//
	// Otherwise, no side effects are done by go-fed.
	Accept func(context.Context, vocab.ActivityStreamsAccept) error
	// Reject handles additional side effects for the Reject ActivityStreams
	// type, specific to the application using go-fed.
	//
	// The wrapping function has no default side effects. However, if this
	// 'Reject' is in response to a 'Follow' then the client MUST NOT go
	// forward with adding the 'actor' to the original 'actor's 'following'
	// collection by the client application.
	Reject func(context.Context, vocab.ActivityStreamsReject) error
	// Add handles additional side effects for the Add ActivityStreams
	// type, specific to the application using go-fed.
	//
	// The wrapping function will add the 'object' IRIs to a specific
	// 'target' collection if the 'target' collection(s) live on this
	// server.
	Add func(context.Context, vocab.ActivityStreamsAdd) error
	// Remove handles additional side effects for the Remove ActivityStreams
	// type, specific to the application using go-fed.
	//
	// The wrapping function will remove all 'object' IRIs from a specific
	// 'target' collection if the 'target' collection(s) live on this
	// server.
	Remove func(context.Context, vocab.ActivityStreamsRemove) error
	// Like handles additional side effects for the Like ActivityStreams
	// type, specific to the application using go-fed.
	//
	// The wrapping function will add the activity to the "likes" collection
	// on all 'object' targets owned by this server.
	Like func(context.Context, vocab.ActivityStreamsLike) error
	// Announce handles additional side effects for the Announce
	// ActivityStreams type, specific to the application using go-fed.
	//
	// The wrapping function will add the activity to the "shares"
	// collection on all 'object' targets owned by this server.
	Announce func(context.Context, vocab.ActivityStreamsAnnounce) error
	// Undo handles additional side effects for the Undo ActivityStreams
	// type, specific to the application using go-fed.
	//
	// The wrapping function ensures the 'actor' on the 'Undo'
	// is be the same as the 'actor' on all Activities being undone.
	// It enforces that the actors on the Undo must correspond to all of the
	// 'object' actors in some manner.
	//
	// It is expected that the application will implement the proper
	// reversal of activities that are being undone.
	Undo func(context.Context, vocab.ActivityStreamsUndo) error
	// Block handles additional side effects for the Block ActivityStreams
	// type, specific to the application using go-fed.
	//
	// The wrapping function provides no default side effects. It simply
	// calls the wrapped function. However, note that Blocks should not be
	// received from a federated peer, as delivering Blocks explicitly
	// deviates from the original ActivityPub specification.
	Block func(context.Context, vocab.ActivityStreamsBlock) error

	// Sidechannel data -- this is set at request handling time. These must
	// be set before the callbacks are used.

	// db is the Database the FederatingWrappedCallbacks should use.
	db Database
	// inboxIRI is the inboxIRI that is handling this callback.
	inboxIRI *url.URL
	// addNewIds creates new 'id' entries on an activity and its objects if
	// it is a Create activity.
	addNewIds func(c context.Context, activity Activity) error
	// deliver delivers an outgoing message.
	deliver func(c context.Context, outboxIRI *url.URL, activity Activity) error
	// newTransport creates a new Transport.
	newTransport func(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t Transport, err error)
}

// callbacks returns the WrappedCallbacks members into a single interface slice
// for use in streams.Resolver callbacks.
//
// If the given functions have a type that collides with the default behavior,
// then disable our default behavior
func (w FederatingWrappedCallbacks) callbacks(fns []interface{}) []interface{} {
	enableCreate := true
	enableUpdate := true
	enableDelete := true
	enableFollow := true
	enableAccept := true
	enableReject := true
	enableAdd := true
	enableRemove := true
	enableLike := true
	enableAnnounce := true
	enableUndo := true
	enableBlock := true
	for _, fn := range fns {
		switch fn.(type) {
		default:
			continue
		case func(context.Context, vocab.ActivityStreamsCreate) error:
			enableCreate = false
		case func(context.Context, vocab.ActivityStreamsUpdate) error:
			enableUpdate = false
		case func(context.Context, vocab.ActivityStreamsDelete) error:
			enableDelete = false
		case func(context.Context, vocab.ActivityStreamsFollow) error:
			enableFollow = false
		case func(context.Context, vocab.ActivityStreamsAccept) error:
			enableAccept = false
		case func(context.Context, vocab.ActivityStreamsReject) error:
			enableReject = false
		case func(context.Context, vocab.ActivityStreamsAdd) error:
			enableAdd = false
		case func(context.Context, vocab.ActivityStreamsRemove) error:
			enableRemove = false
		case func(context.Context, vocab.ActivityStreamsLike) error:
			enableLike = false
		case func(context.Context, vocab.ActivityStreamsAnnounce) error:
			enableAnnounce = false
		case func(context.Context, vocab.ActivityStreamsUndo) error:
			enableUndo = false
		case func(context.Context, vocab.ActivityStreamsBlock) error:
			enableBlock = false
		}
	}
	if enableCreate {
		fns = append(fns, w.create)
	}
	if enableUpdate {
		fns = append(fns, w.update)
	}
	if enableDelete {
		fns = append(fns, w.deleteFn)
	}
	if enableFollow {
		fns = append(fns, w.follow)
	}
	if enableAccept {
		fns = append(fns, w.accept)
	}
	if enableReject {
		fns = append(fns, w.reject)
	}
	if enableAdd {
		fns = append(fns, w.add)
	}
	if enableRemove {
		fns = append(fns, w.remove)
	}
	if enableLike {
		fns = append(fns, w.like)
	}
	if enableAnnounce {
		fns = append(fns, w.announce)
	}
	if enableUndo {
		fns = append(fns, w.undo)
	}
	if enableBlock {
		fns = append(fns, w.block)
	}
	return fns
}

// create implements the federating Create activity side effects.
func (w FederatingWrappedCallbacks) create(c context.Context, a vocab.ActivityStreamsCreate) error {
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	// Create anonymous loop function to be able to properly scope the defer
	// for the database lock at each iteration.
	loopFn := func(iter vocab.ActivityStreamsObjectPropertyIterator) error {
		t := iter.GetType()
		if t == nil && iter.IsIRI() {
			// Attempt to dereference the IRI instead
			tport, err := w.newTransport(c, w.inboxIRI, goFedUserAgent())
			if err != nil {
				return err
			}
			resp, err := tport.Dereference(c, iter.GetIRI())
			if err != nil {
				return err
			}
			m, err := readActivityPubResp(resp)
			if err != nil {
				return err
			}
			t, err = streams.ToType(c, m)
			if err != nil {
				return err
			}
		} else if t == nil {
			return fmt.Errorf("cannot handle federated create: object is neither a value nor IRI")
		}
		id, err := GetId(t)
		if err != nil {
			return err
		}
		var unlock func()
		unlock, err = w.db.Lock(c, id)
		if err != nil {
			return err
		}
		defer unlock()
		if err := w.db.Create(c, t); err != nil {
			return err
		}
		return nil
	}
	for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
		if err := loopFn(iter); err != nil {
			return err
		}
	}
	if w.Create != nil {
		return w.Create(c, a)
	}
	return nil
}

// update implements the federating Update activity side effects.
func (w FederatingWrappedCallbacks) update(c context.Context, a vocab.ActivityStreamsUpdate) error {
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	if err := mustHaveActivityOriginMatchObjects(a); err != nil {
		return err
	}
	// Create anonymous loop function to be able to properly scope the defer
	// for the database lock at each iteration.
	loopFn := func(iter vocab.ActivityStreamsObjectPropertyIterator) error {
		t := iter.GetType()
		if t == nil {
			return fmt.Errorf("update requires an object to be wholly provided")
		}
		id, err := GetId(t)
		if err != nil {
			return err
		}
		var unlock func()
		unlock, err = w.db.Lock(c, id)
		if err != nil {
			return err
		}
		defer unlock()
		if err := w.db.Update(c, t); err != nil {
			return err
		}
		return nil
	}
	for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
		if err := loopFn(iter); err != nil {
			return err
		}
	}
	if w.Update != nil {
		return w.Update(c, a)
	}
	return nil
}

// deleteFn implements the federating Delete activity side effects.
func (w FederatingWrappedCallbacks) deleteFn(c context.Context, a vocab.ActivityStreamsDelete) error {
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	if err := mustHaveActivityOriginMatchObjects(a); err != nil {
		return err
	}
	// Create anonymous loop function to be able to properly scope the defer
	// for the database lock at each iteration.
	loopFn := func(iter vocab.ActivityStreamsObjectPropertyIterator) error {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		var unlock func()
		unlock, err = w.db.Lock(c, id)
		if err != nil {
			return err
		}
		defer unlock()
		if err := w.db.Delete(c, id); err != nil {
			return err
		}
		return nil
	}
	for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
		if err := loopFn(iter); err != nil {
			return err
		}
	}
	if w.Delete != nil {
		return w.Delete(c, a)
	}
	return nil
}

// follow implements the federating Follow activity side effects.
func (w FederatingWrappedCallbacks) follow(c context.Context, a vocab.ActivityStreamsFollow) error {
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	// Check that we own at least one of the 'object' properties, and ensure
	// it is to the actor that owns this inbox.
	//
	// If not then don't send a response. It was federated to us as an FYI,
	// by mistake, or some other reason.
	unlock, err := w.db.Lock(c, w.inboxIRI)
	if err != nil {
		return err
	}
	// WARNING: Unlock not deferred.
	actorIRI, err := w.db.ActorForInbox(c, w.inboxIRI)
	unlock() // unlock even on error
	if err != nil {
		return err
	}
	// Unlock must be called by now and every branch above.
	isMe := false
	if w.OnFollow != OnFollowDoNothing {
		for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
			id, err := ToId(iter)
			if err != nil {
				return err
			}
			if id.String() == actorIRI.String() {
				isMe = true
				break
			}
		}
	}
	if isMe {
		// Prepare the response.
		var response Activity
		if w.OnFollow == OnFollowAutomaticallyAccept {
			response = streams.NewActivityStreamsAccept()
		} else if w.OnFollow == OnFollowAutomaticallyReject {
			response = streams.NewActivityStreamsReject()
		} else {
			return fmt.Errorf("unknown OnFollowBehavior: %d", w.OnFollow)
		}
		// Set us as the 'actor'.
		me := streams.NewActivityStreamsActorProperty()
		response.SetActivityStreamsActor(me)
		me.AppendIRI(actorIRI)
		// Set the Follow as the 'object' property.
		op := streams.NewActivityStreamsObjectProperty()
		response.SetActivityStreamsObject(op)
		op.AppendActivityStreamsFollow(a)
		// Add all actors on the original Follow to the 'to' property.
		recipients := make([]*url.URL, 0)
		to := streams.NewActivityStreamsToProperty()
		response.SetActivityStreamsTo(to)
		followActors := a.GetActivityStreamsActor()
		for iter := followActors.Begin(); iter != followActors.End(); iter = iter.Next() {
			id, err := ToId(iter)
			if err != nil {
				return err
			}
			to.AppendIRI(id)
			recipients = append(recipients, id)
		}
		if w.OnFollow == OnFollowAutomaticallyAccept {
			// If automatically accepting, then also update our
			// followers collection with the new actors.
			//
			// If automatically rejecting, do not update the
			// followers collection.
			unlock, err := w.db.Lock(c, actorIRI)
			if err != nil {
				return err
			}
			// WARNING: Unlock not deferred.
			followers, err := w.db.Followers(c, actorIRI)
			if err != nil {
				unlock()
				return err
			}
			items := followers.GetActivityStreamsItems()
			if items == nil {
				items = streams.NewActivityStreamsItemsProperty()
				followers.SetActivityStreamsItems(items)
			}
			for _, elem := range recipients {
				items.PrependIRI(elem)
			}
			err = w.db.Update(c, followers)
			unlock() // unlock even on error
			if err != nil {
				return err
			}
			// Unlock must be called by now and every branch above.
		}
		// Lock without defer!
		unlock, err := w.db.Lock(c, w.inboxIRI)
		if err != nil {
			return err
		}
		outboxIRI, err := w.db.OutboxForInbox(c, w.inboxIRI)
		unlock() // unlock after, regardless
		if err != nil {
			return err
		}
		// Everything must be unlocked by now.
		if err := w.addNewIds(c, response); err != nil {
			return err
		} else if err := w.deliver(c, outboxIRI, response); err != nil {
			return err
		}
	}
	if w.Follow != nil {
		return w.Follow(c, a)
	}
	return nil
}

// accept implements the federating Accept activity side effects.
func (w FederatingWrappedCallbacks) accept(c context.Context, a vocab.ActivityStreamsAccept) error {
	op := a.GetActivityStreamsObject()
	if op != nil && op.Len() > 0 {
		// Get this actor's id.
		unlock, err := w.db.Lock(c, w.inboxIRI)
		if err != nil {
			return err
		}
		// WARNING: Unlock not deferred.
		actorIRI, err := w.db.ActorForInbox(c, w.inboxIRI)
		unlock() // unlock after regardless
		if err != nil {
			return err
		}
		// Unlock must be called by now and every branch above.
		//
		// Determine if we are in a follow on the 'object' property.
		//
		// TODO: Handle Accept multiple Follow.
		var maybeMyFollowIRI *url.URL
		for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
			t := iter.GetType()
			if t == nil && iter.IsIRI() {
				// Attempt to dereference the IRI instead
				tport, err := w.newTransport(c, w.inboxIRI, goFedUserAgent())
				if err != nil {
					return err
				}
				resp, err := tport.Dereference(c, iter.GetIRI())
				if err != nil {
					return err
				}
				m, err := readActivityPubResp(resp)
				if err != nil {
					return err
				}
				t, err = streams.ToType(c, m)
				if err != nil {
					return err
				}
			} else if t == nil {
				return fmt.Errorf("cannot handle federated create: object is neither a value nor IRI")
			}
			// Ensure it is a Follow.
			if !streams.IsOrExtendsActivityStreamsFollow(t) {
				continue
			}
			follow, ok := t.(Activity)
			if !ok {
				return fmt.Errorf("a Follow in an Accept does not satisfy the Activity interface")
			}
			followId, err := GetId(follow)
			if err != nil {
				return err
			}
			// Ensure that we are one of the actors on the Follow.
			actors := follow.GetActivityStreamsActor()
			for iter := actors.Begin(); iter != actors.End(); iter = iter.Next() {
				id, err := ToId(iter)
				if err != nil {
					return err
				}
				if id.String() == actorIRI.String() {
					maybeMyFollowIRI = followId
					break
				}
			}
			// Continue breaking if we found ourselves
			if maybeMyFollowIRI != nil {
				break
			}
		}
		// If we received an Accept whose 'object' is a Follow with an
		// Accept that we sent, add to the following collection.
		if maybeMyFollowIRI != nil {
			// Verify our Follow request exists and the peer didn't
			// fabricate it.
			activityActors := a.GetActivityStreamsActor()
			if activityActors == nil || activityActors.Len() == 0 {
				return fmt.Errorf("an Accept with a Follow has no actors")
			}
			// This may be a duplicate check if we dereferenced the
			// Follow above. TODO: Separate this logic to avoid
			// redundancy.
			//
			// Use an anonymous function to properly scope the
			// database lock, immediately call it.
			err = func() error {
				unlock, err := w.db.Lock(c, maybeMyFollowIRI)
				if err != nil {
					return err
				}
				defer unlock()
				t, err := w.db.Get(c, maybeMyFollowIRI)
				if err != nil {
					return err
				}
				if !streams.IsOrExtendsActivityStreamsFollow(t) {
					return fmt.Errorf("peer gave an Accept wrapping a Follow but provided a non-Follow id")
				}
				follow, ok := t.(Activity)
				if !ok {
					return fmt.Errorf("a Follow in an Accept does not satisfy the Activity interface")
				}
				// Ensure that we are one of the actors on the Follow.
				ok = false
				actors := follow.GetActivityStreamsActor()
				for iter := actors.Begin(); iter != actors.End(); iter = iter.Next() {
					id, err := ToId(iter)
					if err != nil {
						return err
					}
					if id.String() == actorIRI.String() {
						ok = true
						break
					}
				}
				if !ok {
					return fmt.Errorf("peer gave an Accept wrapping a Follow but we are not the actor on that Follow")
				}
				// Build map of original Accept actors
				acceptActors := make(map[string]bool)
				for iter := activityActors.Begin(); iter != activityActors.End(); iter = iter.Next() {
					id, err := ToId(iter)
					if err != nil {
						return err
					}
					acceptActors[id.String()] = false
				}
				// Verify all actor(s) were on the original Follow.
				followObj := follow.GetActivityStreamsObject()
				for iter := followObj.Begin(); iter != followObj.End(); iter = iter.Next() {
					id, err := ToId(iter)
					if err != nil {
						return err
					}
					if _, ok := acceptActors[id.String()]; ok {
						acceptActors[id.String()] = true
					}
				}
				for _, found := range acceptActors {
					if !found {
						return fmt.Errorf("peer gave an Accept wrapping a Follow but was not an object in the original Follow")
					}
				}
				return nil
			}()
			if err != nil {
				return err
			}
			// Add the peer to our following collection.
			unlock, err := w.db.Lock(c, actorIRI)
			if err != nil {
				return err
			}
			// WARNING: Unlock not deferred.
			following, err := w.db.Following(c, actorIRI)
			if err != nil {
				unlock()
				return err
			}
			items := following.GetActivityStreamsItems()
			if items == nil {
				items = streams.NewActivityStreamsItemsProperty()
				following.SetActivityStreamsItems(items)
			}
			for iter := activityActors.Begin(); iter != activityActors.End(); iter = iter.Next() {
				id, err := ToId(iter)
				if err != nil {
					unlock()
					return err
				}
				items.PrependIRI(id)
			}
			err = w.db.Update(c, following)
			unlock() // unlock after regardless
			if err != nil {
				return err
			}
			// Unlock must be called by now and every branch above.
		}
	}
	if w.Accept != nil {
		return w.Accept(c, a)
	}
	return nil
}

// reject implements the federating Reject activity side effects.
func (w FederatingWrappedCallbacks) reject(c context.Context, a vocab.ActivityStreamsReject) error {
	if w.Reject != nil {
		return w.Reject(c, a)
	}
	return nil
}

// add implements the federating Add activity side effects.
func (w FederatingWrappedCallbacks) add(c context.Context, a vocab.ActivityStreamsAdd) error {
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	target := a.GetActivityStreamsTarget()
	if target == nil || target.Len() == 0 {
		return ErrTargetRequired
	}
	if err := add(c, op, target, w.db); err != nil {
		return err
	}
	if w.Add != nil {
		return w.Add(c, a)
	}
	return nil
}

// remove implements the federating Remove activity side effects.
func (w FederatingWrappedCallbacks) remove(c context.Context, a vocab.ActivityStreamsRemove) error {
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	target := a.GetActivityStreamsTarget()
	if target == nil || target.Len() == 0 {
		return ErrTargetRequired
	}
	if err := remove(c, op, target, w.db); err != nil {
		return err
	}
	if w.Remove != nil {
		return w.Remove(c, a)
	}
	return nil
}

// like implements the federating Like activity side effects.
func (w FederatingWrappedCallbacks) like(c context.Context, a vocab.ActivityStreamsLike) error {
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	id, err := GetId(a)
	if err != nil {
		return err
	}
	// Create anonymous loop function to be able to properly scope the defer
	// for the database lock at each iteration.
	loopFn := func(iter vocab.ActivityStreamsObjectPropertyIterator) error {
		objId, err := ToId(iter)
		if err != nil {
			return err
		}
		unlock, err := w.db.Lock(c, objId)
		if err != nil {
			return err
		}
		defer unlock()
		if owns, err := w.db.Owns(c, objId); err != nil {
			return err
		} else if !owns {
			return nil
		}
		t, err := w.db.Get(c, objId)
		if err != nil {
			return err
		}
		l, ok := t.(likeser)
		if !ok {
			return fmt.Errorf("cannot add Like to likes collection for type %T", t)
		}
		// Get 'likes' property on the object, creating default if
		// necessary.
		likes := l.GetActivityStreamsLikes()
		if likes == nil {
			likes = streams.NewActivityStreamsLikesProperty()
			l.SetActivityStreamsLikes(likes)
		}
		// Get 'likes' value, defaulting to a collection.
		likesT := likes.GetType()
		if likesT == nil {
			col := streams.NewActivityStreamsCollection()
			likesT = col
			likes.SetActivityStreamsCollection(col)
		}
		// Prepend the activity's 'id' on the 'likes' Collection or
		// OrderedCollection.
		if col, ok := likesT.(itemser); ok {
			items := col.GetActivityStreamsItems()
			if items == nil {
				items = streams.NewActivityStreamsItemsProperty()
				col.SetActivityStreamsItems(items)
			}
			items.PrependIRI(id)
		} else if oCol, ok := likesT.(orderedItemser); ok {
			oItems := oCol.GetActivityStreamsOrderedItems()
			if oItems == nil {
				oItems = streams.NewActivityStreamsOrderedItemsProperty()
				oCol.SetActivityStreamsOrderedItems(oItems)
			}
			oItems.PrependIRI(id)
		} else {
			return fmt.Errorf("likes type is neither a Collection nor an OrderedCollection: %T", likesT)
		}
		err = w.db.Update(c, t)
		if err != nil {
			return err
		}
		return nil
	}
	for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
		if err := loopFn(iter); err != nil {
			return err
		}
	}
	if w.Like != nil {
		return w.Like(c, a)
	}
	return nil
}

// announce implements the federating Announce activity side effects.
func (w FederatingWrappedCallbacks) announce(c context.Context, a vocab.ActivityStreamsAnnounce) error {
	id, err := GetId(a)
	if err != nil {
		return err
	}
	op := a.GetActivityStreamsObject()
	// Create anonymous loop function to be able to properly scope the defer
	// for the database lock at each iteration.
	loopFn := func(iter vocab.ActivityStreamsObjectPropertyIterator) error {
		objId, err := ToId(iter)
		if err != nil {
			return err
		}
		unlock, err := w.db.Lock(c, objId)
		if err != nil {
			return err
		}
		defer unlock()
		if owns, err := w.db.Owns(c, objId); err != nil {
			return err
		} else if !owns {
			return nil
		}
		t, err := w.db.Get(c, objId)
		if err != nil {
			return err
		}
		s, ok := t.(shareser)
		if !ok {
			return fmt.Errorf("cannot add Announce to Shares collection for type %T", t)
		}
		// Get 'shares' property on the object, creating default if
		// necessary.
		shares := s.GetActivityStreamsShares()
		if shares == nil {
			shares = streams.NewActivityStreamsSharesProperty()
			s.SetActivityStreamsShares(shares)
		}
		// Get 'shares' value, defaulting to a collection.
		sharesT := shares.GetType()
		if sharesT == nil {
			col := streams.NewActivityStreamsCollection()
			sharesT = col
			shares.SetActivityStreamsCollection(col)
		}
		// Prepend the activity's 'id' on the 'shares' Collection or
		// OrderedCollection.
		if col, ok := sharesT.(itemser); ok {
			items := col.GetActivityStreamsItems()
			if items == nil {
				items = streams.NewActivityStreamsItemsProperty()
				col.SetActivityStreamsItems(items)
			}
			items.PrependIRI(id)
		} else if oCol, ok := sharesT.(orderedItemser); ok {
			oItems := oCol.GetActivityStreamsOrderedItems()
			if oItems == nil {
				oItems = streams.NewActivityStreamsOrderedItemsProperty()
				oCol.SetActivityStreamsOrderedItems(oItems)
			}
			oItems.PrependIRI(id)
		} else {
			return fmt.Errorf("shares type is neither a Collection nor an OrderedCollection: %T", sharesT)
		}
		err = w.db.Update(c, t)
		if err != nil {
			return err
		}
		return nil
	}
	if op != nil {
		for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
			if err := loopFn(iter); err != nil {
				return err
			}
		}
	}
	if w.Announce != nil {
		return w.Announce(c, a)
	}
	return nil
}

// undo implements the federating Undo activity side effects.
func (w FederatingWrappedCallbacks) undo(c context.Context, a vocab.ActivityStreamsUndo) error {
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	actors := a.GetActivityStreamsActor()
	if err := mustHaveActivityActorsMatchObjectActors(c, actors, op, w.newTransport, w.inboxIRI); err != nil {
		return err
	}
	if w.Undo != nil {
		return w.Undo(c, a)
	}
	return nil
}

// block implements the federating Block activity side effects.
func (w FederatingWrappedCallbacks) block(c context.Context, a vocab.ActivityStreamsBlock) error {
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	if w.Block != nil {
		return w.Block(c, a)
	}
	return nil
}
