package pub

import (
	"context"
	"fmt"
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
)

// SocialWrappedCallbacks lists the callback functions that already have some
// side effect behavior provided by the pub library.
//
// These functions are wrapped for the Social Protocol.
type SocialWrappedCallbacks struct {
	// Create handles additional side effects for the Create ActivityStreams
	// type.
	//
	// The wrapping callback copies the actor(s) to the 'attributedTo'
	// property and copies recipients between the Create activity and all
	// objects. It then saves the entry in the database.
	Create func(context.Context, vocab.ActivityStreamsCreate) error
	// Update handles additional side effects for the Update ActivityStreams
	// type.
	//
	// The wrapping callback applies new top-level values on an object to
	// the stored objects. Any top-level null literals will be deleted on
	// the stored objects as well.
	Update func(context.Context, vocab.ActivityStreamsUpdate) error
	// Delete handles additional side effects for the Delete ActivityStreams
	// type.
	//
	// The wrapping callback replaces the object(s) with tombstones in the
	// database.
	Delete func(context.Context, vocab.ActivityStreamsDelete) error
	// Follow handles additional side effects for the Follow ActivityStreams
	// type.
	//
	// The wrapping callback only ensures the 'Follow' has at least one
	// 'object' entry, but otherwise has no default side effect.
	Follow func(context.Context, vocab.ActivityStreamsFollow) error
	// Add handles additional side effects for the Add ActivityStreams
	// type.
	//
	//
	// The wrapping function will add the 'object' IRIs to a specific
	// 'target' collection if the 'target' collection(s) live on this
	// server.
	Add func(context.Context, vocab.ActivityStreamsAdd) error
	// Remove handles additional side effects for the Remove ActivityStreams
	// type.
	//
	// The wrapping function will remove all 'object' IRIs from a specific
	// 'target' collection if the 'target' collection(s) live on this
	// server.
	Remove func(context.Context, vocab.ActivityStreamsRemove) error
	// Like handles additional side effects for the Like ActivityStreams
	// type.
	//
	// The wrapping function will add the objects on the activity to the
	// "liked" collection of this actor.
	Like func(context.Context, vocab.ActivityStreamsLike) error
	// Undo handles additional side effects for the Undo ActivityStreams
	// type.
	//
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
	// type.
	//
	// The wrapping callback only ensures the 'Block' has at least one
	// 'object' entry, but otherwise has no default side effect. It is up
	// to the wrapped application function to properly enforce the new
	// blocking behavior.
	//
	// Note that go-fed does not federate 'Block' activities received in the
	// Social Protocol.
	Block func(context.Context, vocab.ActivityStreamsBlock) error

	// Sidechannel data -- this is set at request handling time. These must
	// be set before the callbacks are used.

	// db is the Database the SocialWrappedCallbacks should use. It must be
	// set before calling the callbacks.
	db Database
	// outboxIRI is the outboxIRI that is handling this callback.
	outboxIRI *url.URL
	// rawActivity is the JSON map literal received when deserializing the
	// request body.
	rawActivity map[string]interface{}
	// clock is the server's clock.
	clock Clock
	// newTransport creates a new Transport.
	newTransport func(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t Transport, err error)
	// undeliverable is a sidechannel out, indicating if the handled activity
	// should not be delivered to a peer.
	//
	// Its provided default value will always be used when a custom function
	// is called.
	undeliverable *bool
}

// callbacks returns the WrappedCallbacks members into a single interface slice
// for use in streams.Resolver callbacks.
//
// If the given functions have a type that collides with the default behavior,
// then disable our default behavior
func (w SocialWrappedCallbacks) callbacks(fns []interface{}) []interface{} {
	enableCreate := true
	enableUpdate := true
	enableDelete := true
	enableFollow := true
	enableAdd := true
	enableRemove := true
	enableLike := true
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
		case func(context.Context, vocab.ActivityStreamsAdd) error:
			enableAdd = false
		case func(context.Context, vocab.ActivityStreamsRemove) error:
			enableRemove = false
		case func(context.Context, vocab.ActivityStreamsLike) error:
			enableLike = false
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
	if enableAdd {
		fns = append(fns, w.add)
	}
	if enableRemove {
		fns = append(fns, w.remove)
	}
	if enableLike {
		fns = append(fns, w.like)
	}
	if enableUndo {
		fns = append(fns, w.undo)
	}
	if enableBlock {
		fns = append(fns, w.block)
	}
	return fns
}

// create implements the social Create activity side effects.
func (w SocialWrappedCallbacks) create(c context.Context, a vocab.ActivityStreamsCreate) error {
	*w.undeliverable = false
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	// Obtain all actor IRIs.
	actors := a.GetActivityStreamsActor()
	createActorIds := make(map[string]*url.URL)
	if actors != nil {
		createActorIds = make(map[string]*url.URL, actors.Len())
		for iter := actors.Begin(); iter != actors.End(); iter = iter.Next() {
			id, err := ToId(iter)
			if err != nil {
				return err
			}
			createActorIds[id.String()] = id
		}
	}
	// Obtain each object's 'attributedTo' IRIs.
	objectAttributedToIds := make([]map[string]*url.URL, op.Len())
	for i := range objectAttributedToIds {
		objectAttributedToIds[i] = make(map[string]*url.URL)
	}
	for i := 0; i < op.Len(); i++ {
		t := op.At(i).GetType()
		attrToer, ok := t.(attributedToer)
		if !ok {
			continue
		}
		attr := attrToer.GetActivityStreamsAttributedTo()
		if attr == nil {
			attr = streams.NewActivityStreamsAttributedToProperty()
			attrToer.SetActivityStreamsAttributedTo(attr)
		}
		for iter := attr.Begin(); iter != attr.End(); iter = iter.Next() {
			id, err := ToId(iter)
			if err != nil {
				return err
			}
			objectAttributedToIds[i][id.String()] = id
		}
	}
	// Put all missing actor IRIs onto all object attributedTo properties.
	for k, v := range createActorIds {
		for i, attributedToMap := range objectAttributedToIds {
			if _, ok := attributedToMap[k]; !ok {
				t := op.At(i).GetType()
				attrToer, ok := t.(attributedToer)
				if !ok {
					continue
				}
				attr := attrToer.GetActivityStreamsAttributedTo()
				attr.AppendIRI(v)
			}
		}
	}
	// Put all missing object attributedTo IRIs onto the actor property
	// if there is one.
	if actors != nil {
		for _, attributedToMap := range objectAttributedToIds {
			for k, v := range attributedToMap {
				if _, ok := createActorIds[k]; !ok {
					actors.AppendIRI(v)
				}
			}
		}
	}
	// Copy over the 'to', 'bto', 'cc', 'bcc', and 'audience' recipients
	// between the activity and all child objects and vice versa.
	if err := normalizeRecipients(a); err != nil {
		return err
	}
	// Create anonymous loop function to be able to properly scope the defer
	// for the database lock at each iteration.
	loopFn := func(i int) error {
		obj := op.At(i).GetType()
		id, err := GetId(obj)
		if err != nil {
			return err
		}
		var unlock func()
		unlock, err = w.db.Lock(c, id)
		if err != nil {
			return err
		}
		defer unlock()
		if err := w.db.Create(c, obj); err != nil {
			return err
		}
		return nil
	}
	// Persist all objects we've created, which will include sensitive
	// recipients such as 'bcc' and 'bto'.
	for i := 0; i < op.Len(); i++ {
		if err := loopFn(i); err != nil {
			return err
		}
	}
	if w.Create != nil {
		return w.Create(c, a)
	}
	return nil
}

// update implements the social Update activity side effects.
func (w SocialWrappedCallbacks) update(c context.Context, a vocab.ActivityStreamsUpdate) error {
	*w.undeliverable = false
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	// Obtain all object ids, which should be owned by this server.
	objIds := make([]*url.URL, 0, op.Len())
	for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		objIds = append(objIds, id)
	}
	// Create anonymous loop function to be able to properly scope the defer
	// for the database lock at each iteration.
	loopFn := func(idx int, loopId *url.URL) error {
		unlock, err := w.db.Lock(c, loopId)
		if err != nil {
			return err
		}
		defer unlock()
		t, err := w.db.Get(c, loopId)
		if err != nil {
			return err
		}
		m, err := t.Serialize()
		if err != nil {
			return err
		}
		// Copy over new top-level values.
		objType := op.At(idx).GetType()
		if objType == nil {
			return fmt.Errorf("object at index %d is not a literal type value", idx)
		}
		newM, err := objType.Serialize()
		if err != nil {
			return err
		}
		for k, v := range newM {
			m[k] = v
		}
		// Delete top-level values where the raw Activity had nils.
		for k, v := range w.rawActivity {
			if _, ok := m[k]; v == nil && ok {
				delete(m, k)
			}
		}
		newT, err := streams.ToType(c, m)
		if err != nil {
			return err
		}
		if err = w.db.Update(c, newT); err != nil {
			return err
		}
		return nil
	}
	for i, id := range objIds {
		if err := loopFn(i, id); err != nil {
			return err
		}
	}
	if w.Update != nil {
		return w.Update(c, a)
	}
	return nil
}

// deleteFn implements the social Delete activity side effects.
func (w SocialWrappedCallbacks) deleteFn(c context.Context, a vocab.ActivityStreamsDelete) error {
	*w.undeliverable = false
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	// Obtain all object ids, which should be owned by this server.
	objIds := make([]*url.URL, 0, op.Len())
	for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
		id, err := ToId(iter)
		if err != nil {
			return err
		}
		objIds = append(objIds, id)
	}
	// Create anonymous loop function to be able to properly scope the defer
	// for the database lock at each iteration.
	loopFn := func(idx int, loopId *url.URL) error {
		unlock, err := w.db.Lock(c, loopId)
		if err != nil {
			return err
		}
		defer unlock()
		t, err := w.db.Get(c, loopId)
		if err != nil {
			return err
		}
		tomb := toTombstone(t, loopId, w.clock.Now())
		if err := w.db.Update(c, tomb); err != nil {
			return err
		}
		return nil
	}
	for i, id := range objIds {
		if err := loopFn(i, id); err != nil {
			return err
		}
	}
	if w.Delete != nil {
		return w.Delete(c, a)
	}
	return nil
}

// follow implements the social Follow activity side effects.
func (w SocialWrappedCallbacks) follow(c context.Context, a vocab.ActivityStreamsFollow) error {
	*w.undeliverable = false
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	if w.Follow != nil {
		return w.Follow(c, a)
	}
	return nil
}

// add implements the social Add activity side effects.
func (w SocialWrappedCallbacks) add(c context.Context, a vocab.ActivityStreamsAdd) error {
	*w.undeliverable = false
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

// remove implements the social Remove activity side effects.
func (w SocialWrappedCallbacks) remove(c context.Context, a vocab.ActivityStreamsRemove) error {
	*w.undeliverable = false
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

// like implements the social Like activity side effects.
func (w SocialWrappedCallbacks) like(c context.Context, a vocab.ActivityStreamsLike) error {
	*w.undeliverable = false
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	// Get this actor's IRI.
	unlock, err := w.db.Lock(c, w.outboxIRI)
	if err != nil {
		return err
	}
	// WARNING: Unlock not deferred.
	actorIRI, err := w.db.ActorForOutbox(c, w.outboxIRI)
	unlock() // unlock even on error
	if err != nil {
		return err
	}
	// Unlock must be called by now and every branch above.
	//
	// Now obtain this actor's 'liked' collection.
	unlock, err = w.db.Lock(c, actorIRI)
	if err != nil {
		return err
	}
	defer unlock()
	liked, err := w.db.Liked(c, actorIRI)
	if err != nil {
		return err
	}
	likedItems := liked.GetActivityStreamsItems()
	if likedItems == nil {
		likedItems = streams.NewActivityStreamsItemsProperty()
		liked.SetActivityStreamsItems(likedItems)
	}
	for iter := op.Begin(); iter != op.End(); iter = iter.Next() {
		objId, err := ToId(iter)
		if err != nil {
			return err
		}
		likedItems.PrependIRI(objId)
	}
	err = w.db.Update(c, liked)
	if err != nil {
		return err
	}
	if w.Like != nil {
		return w.Like(c, a)
	}
	return nil
}

// undo implements the social Undo activity side effects.
func (w SocialWrappedCallbacks) undo(c context.Context, a vocab.ActivityStreamsUndo) error {
	*w.undeliverable = false
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	actors := a.GetActivityStreamsActor()
	if err := mustHaveActivityActorsMatchObjectActors(c, actors, op, w.newTransport, w.outboxIRI); err != nil {
		return err
	}
	if w.Undo != nil {
		return w.Undo(c, a)
	}
	return nil
}

// block implements the social Block activity side effects.
func (w SocialWrappedCallbacks) block(c context.Context, a vocab.ActivityStreamsBlock) error {
	*w.undeliverable = true
	op := a.GetActivityStreamsObject()
	if op == nil || op.Len() == 0 {
		return ErrObjectRequired
	}
	if w.Block != nil {
		return w.Block(c, a)
	}
	return nil
}
