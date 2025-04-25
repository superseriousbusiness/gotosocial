package pub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
)

// baseActor must satisfy the Actor interface.
var _ Actor = &baseActor{}

// baseActor is an application-independent ActivityPub implementation. It does
// not implement the entire protocol, and relies on a delegate to do so. It
// only implements the part of the protocol that is side-effect-free, allowing
// an existing application to write a DelegateActor that glues their application
// into the ActivityPub world.
//
// It is preferred to use a DelegateActor provided by this library, so that the
// application does not need to worry about the ActivityPub implementation.
type baseActor struct {
	// delegate contains application-specific delegation logic.
	delegate DelegateActor
	// enableSocialProtocol enables or disables the Social API, the client to
	// server part of ActivityPub. Useful if permitting remote clients to
	// act on behalf of the users of the client application.
	enableSocialProtocol bool
	// enableFederatedProtocol enables or disables the Federated Protocol, or the
	// server to server part of ActivityPub. Useful to permit integrating
	// with the rest of the federative web.
	enableFederatedProtocol bool
	// clock simply tracks the current time.
	clock Clock
}

// baseActorFederating must satisfy the FederatingActor interface.
var _ FederatingActor = &baseActorFederating{}

// baseActorFederating is a baseActor that also satisfies the FederatingActor
// interface.
//
// The baseActor is preserved as an Actor which will not successfully cast to a
// FederatingActor.
type baseActorFederating struct {
	baseActor
}

// NewSocialActor builds a new Actor concept that handles only the Social
// Protocol part of ActivityPub.
//
// This Actor can be created once in an application and reused to handle
// multiple requests concurrently and for different endpoints.
//
// It leverages as much of go-fed as possible to ensure the implementation is
// compliant with the ActivityPub specification, while providing enough freedom
// to be productive without shooting one's self in the foot.
//
// Do not try to use NewSocialActor and NewFederatingActor together to cover
// both the Social and Federating parts of the protocol. Instead, use NewActor.
func NewSocialActor(c CommonBehavior,
	c2s SocialProtocol,
	db Database,
	clock Clock) Actor {
	return &baseActor{
		// Use SideEffectActor without s2s.
		delegate:             NewSideEffectActor(c, nil, c2s, db, clock),
		enableSocialProtocol: true,
		clock:                clock,
	}
}

// NewFederatingActor builds a new Actor concept that handles only the Federating
// Protocol part of ActivityPub.
//
// This Actor can be created once in an application and reused to handle
// multiple requests concurrently and for different endpoints.
//
// It leverages as much of go-fed as possible to ensure the implementation is
// compliant with the ActivityPub specification, while providing enough freedom
// to be productive without shooting one's self in the foot.
//
// Do not try to use NewSocialActor and NewFederatingActor together to cover
// both the Social and Federating parts of the protocol. Instead, use NewActor.
func NewFederatingActor(c CommonBehavior,
	s2s FederatingProtocol,
	db Database,
	clock Clock) FederatingActor {
	return &baseActorFederating{
		baseActor{
			// Use SideEffectActor without c2s.
			delegate:                NewSideEffectActor(c, s2s, nil, db, clock),
			enableFederatedProtocol: true,
			clock:                   clock,
		},
	}
}

// NewActor builds a new Actor concept that handles both the Social and
// Federating Protocol parts of ActivityPub.
//
// This Actor can be created once in an application and reused to handle
// multiple requests concurrently and for different endpoints.
//
// It leverages as much of go-fed as possible to ensure the implementation is
// compliant with the ActivityPub specification, while providing enough freedom
// to be productive without shooting one's self in the foot.
func NewActor(c CommonBehavior,
	c2s SocialProtocol,
	s2s FederatingProtocol,
	db Database,
	clock Clock) FederatingActor {
	return &baseActorFederating{
		baseActor{
			delegate:                NewSideEffectActor(c, s2s, c2s, db, clock),
			enableSocialProtocol:    true,
			enableFederatedProtocol: true,
			clock:                   clock,
		},
	}
}

// NewCustomActor allows clients to create a custom ActivityPub implementation
// for the Social Protocol, Federating Protocol, or both.
//
// It still uses the library as a high-level scaffold, which has the benefit of
// allowing applications to grow into a custom ActivityPub solution without
// having to refactor the code that passes HTTP requests into the Actor.
//
// It is possible to create a DelegateActor that is not ActivityPub compliant.
// Use with due care.
//
// If you find yourself passing a SideEffectActor in as the DelegateActor,
// consider using NewActor, NewFederatingActor, or NewSocialActor instead.
func NewCustomActor(delegate DelegateActor,
	enableSocialProtocol, enableFederatedProtocol bool,
	clock Clock) FederatingActor {
	return &baseActorFederating{
		baseActor{
			delegate:                delegate,
			enableSocialProtocol:    enableSocialProtocol,
			enableFederatedProtocol: enableFederatedProtocol,
			clock:                   clock,
		},
	}
}

// PostInbox implements the generic algorithm for handling a POST request to an
// actor's inbox independent on an application. It relies on a delegate to
// implement application specific functionality.
//
// Only supports serving data with identifiers having the HTTPS scheme.
func (b *baseActor) PostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return b.PostInboxScheme(c, w, r, "https")
}

// PostInbox implements the generic algorithm for handling a POST request to an
// actor's inbox independent on an application. It relies on a delegate to
// implement application specific functionality.
//
// Specifying the "scheme" allows for retrieving ActivityStreams content with
// identifiers such as HTTP, HTTPS, or other protocol schemes.
func (b *baseActor) PostInboxScheme(c context.Context, w http.ResponseWriter, r *http.Request, scheme string) (bool, error) {
	// Do nothing if it is not an ActivityPub POST request.
	if !isActivityPubPost(r) {
		return false, nil
	}
	// If the Federated Protocol is not enabled, then this endpoint is not
	// enabled.
	if !b.enableFederatedProtocol {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return true, nil
	}
	// Check the peer request is authentic.
	c, authenticated, err := b.delegate.AuthenticatePostInbox(c, w, r)
	if err != nil {
		return true, err
	} else if !authenticated {
		return true, nil
	}
	// Begin processing the request, but have not yet applied
	// authorization (ex: blocks). Obtain the activity reject unknown
	// activities.
	m, err := readActivityPubReq(r)
	if err != nil {
		return true, err
	}
	asValue, err := streams.ToType(c, m)
	if err != nil && !streams.IsUnmatchedErr(err) {
		return true, err
	} else if streams.IsUnmatchedErr(err) {
		// Respond with bad request -- we do not understand the type.
		w.WriteHeader(http.StatusBadRequest)
		return true, nil
	}
	activity, ok := asValue.(Activity)
	if !ok {
		return true, fmt.Errorf("activity streams value is not an Activity: %T", asValue)
	}
	if activity.GetJSONLDId() == nil {
		w.WriteHeader(http.StatusBadRequest)
		return true, nil
	}
	// Allow server implementations to set context data with a hook.
	c, err = b.delegate.PostInboxRequestBodyHook(c, r, activity)
	if err != nil {
		return true, err
	}
	// Check authorization of the activity.
	authorized, err := b.delegate.AuthorizePostInbox(c, w, activity)
	if err != nil {
		return true, err
	} else if !authorized {
		return true, nil
	}
	// Post the activity to the actor's inbox and trigger side effects for
	// that particular Activity type. It is up to the delegate to resolve
	// the given map.
	inboxId := requestId(r, scheme)
	err = b.delegate.PostInbox(c, inboxId, activity)
	if err != nil {
		// Special case: We know it is a bad request if the object or
		// target properties needed to be populated, but weren't.
		//
		// Send the rejection to the peer.
		if err == ErrObjectRequired || err == ErrTargetRequired {
			w.WriteHeader(http.StatusBadRequest)
			return true, nil
		}
		return true, err
	}
	// Our side effects are complete, now delegate determining whether to
	// do inbox forwarding, as well as the action to do it.
	if err := b.delegate.InboxForwarding(c, inboxId, activity); err != nil {
		return true, err
	}
	// Request has been processed. Begin responding to the request.
	//
	// Simply respond with an OK status to the peer.
	w.WriteHeader(http.StatusOK)
	return true, nil
}

// GetInbox implements the generic algorithm for handling a GET request to an
// actor's inbox independent on an application. It relies on a delegate to
// implement application specific functionality.
func (b *baseActor) GetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	// Do nothing if it is not an ActivityPub GET request.
	if !isActivityPubGet(r) {
		return false, nil
	}
	// Delegate authenticating and authorizing the request.
	c, authenticated, err := b.delegate.AuthenticateGetInbox(c, w, r)
	if err != nil {
		return true, err
	} else if !authenticated {
		return true, nil
	}
	// Everything is good to begin processing the request.
	oc, err := b.delegate.GetInbox(c, r)
	if err != nil {
		return true, err
	}
	// Deduplicate the 'orderedItems' property by ID.
	err = dedupeOrderedItems(oc)
	if err != nil {
		return true, err
	}
	// Request has been processed. Begin responding to the request.
	//
	// Serialize the OrderedCollection.
	m, err := streams.Serialize(oc)
	if err != nil {
		return true, err
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return true, err
	}
	// Write the response.
	addResponseHeaders(w.Header(), b.clock, raw)
	w.WriteHeader(http.StatusOK)
	n, err := w.Write(raw)
	if err != nil {
		return true, err
	} else if n != len(raw) {
		return true, fmt.Errorf("ResponseWriter.Write wrote %d of %d bytes", n, len(raw))
	}
	return true, nil
}

// PostOutbox implements the generic algorithm for handling a POST request to an
// actor's outbox independent on an application. It relies on a delegate to
// implement application specific functionality.
//
// Only supports serving data with identifiers having the HTTPS scheme.
func (b *baseActor) PostOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return b.PostOutboxScheme(c, w, r, "https")
}

// PostOutbox implements the generic algorithm for handling a POST request to an
// actor's outbox independent on an application. It relies on a delegate to
// implement application specific functionality.
//
// Specifying the "scheme" allows for retrieving ActivityStreams content with
// identifiers such as HTTP, HTTPS, or other protocol schemes.
func (b *baseActor) PostOutboxScheme(c context.Context, w http.ResponseWriter, r *http.Request, scheme string) (bool, error) {
	// Do nothing if it is not an ActivityPub POST request.
	if !isActivityPubPost(r) {
		return false, nil
	}
	// If the Social API is not enabled, then this endpoint is not enabled.
	if !b.enableSocialProtocol {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return true, nil
	}
	// Delegate authenticating and authorizing the request.
	c, authenticated, err := b.delegate.AuthenticatePostOutbox(c, w, r)
	if err != nil {
		return true, err
	} else if !authenticated {
		return true, nil
	}
	// Everything is good to begin processing the request.
	m, err := readActivityPubReq(r)
	if err != nil {
		return true, err
	}
	// Note that converting to a Type will NOT successfully convert types
	// not known to go-fed. This prevents accidentally wrapping an Activity
	// type unknown to go-fed in a Create below. Instead,
	// streams.ErrUnhandledType will be returned here.
	asValue, err := streams.ToType(c, m)
	if err != nil && !streams.IsUnmatchedErr(err) {
		return true, err
	} else if streams.IsUnmatchedErr(err) {
		// Respond with bad request -- we do not understand the type.
		w.WriteHeader(http.StatusBadRequest)
		return true, nil
	}
	// Allow server implementations to set context data with a hook.
	c, err = b.delegate.PostOutboxRequestBodyHook(c, r, asValue)
	if err != nil {
		return true, err
	}
	// The HTTP request steps are complete, complete the rest of the outbox
	// and delivery process.
	outboxId := requestId(r, scheme)
	activity, err := b.deliver(c, outboxId, asValue, m)
	// Special case: We know it is a bad request if the object or
	// target properties needed to be populated, but weren't.
	//
	// Send the rejection to the client.
	if err == ErrObjectRequired || err == ErrTargetRequired {
		w.WriteHeader(http.StatusBadRequest)
		return true, nil
	} else if err != nil {
		return true, err
	}
	// Respond to the request with the new Activity's IRI location.
	w.Header().Set(locationHeader, activity.GetJSONLDId().Get().String())
	w.WriteHeader(http.StatusCreated)
	return true, nil
}

// GetOutbox implements the generic algorithm for handling a Get request to an
// actor's outbox independent on an application. It relies on a delegate to
// implement application specific functionality.
func (b *baseActor) GetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	// Do nothing if it is not an ActivityPub GET request.
	if !isActivityPubGet(r) {
		return false, nil
	}
	// Delegate authenticating and authorizing the request.
	c, authenticated, err := b.delegate.AuthenticateGetOutbox(c, w, r)
	if err != nil {
		return true, err
	} else if !authenticated {
		return true, nil
	}
	// Everything is good to begin processing the request.
	oc, err := b.delegate.GetOutbox(c, r)
	if err != nil {
		return true, err
	}
	// Request has been processed. Begin responding to the request.
	//
	// Serialize the OrderedCollection.
	m, err := streams.Serialize(oc)
	if err != nil {
		return true, err
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return true, err
	}
	// Write the response.
	addResponseHeaders(w.Header(), b.clock, raw)
	w.WriteHeader(http.StatusOK)
	n, err := w.Write(raw)
	if err != nil {
		return true, err
	} else if n != len(raw) {
		return true, fmt.Errorf("ResponseWriter.Write wrote %d of %d bytes", n, len(raw))
	}
	return true, nil
}

// deliver delegates all outbox handling steps and optionally will federate the
// activity if the federated protocol is enabled.
//
// This function is not exported so an Actor that only supports C2S cannot be
// type casted to a FederatingActor. It doesn't exactly fit the Send method
// signature anyways.
//
// Note: 'm' is nilable.
func (b *baseActor) deliver(c context.Context, outbox *url.URL, asValue vocab.Type, m map[string]interface{}) (activity Activity, err error) {
	// If the value is not an Activity or type extending from Activity, then
	// we need to wrap it in a Create Activity.
	if !streams.IsOrExtendsActivityStreamsActivity(asValue) {
		asValue, err = b.delegate.WrapInCreate(c, asValue, outbox)
		if err != nil {
			return
		}
	}
	// At this point, this should be a safe conversion. If this error is
	// triggered, then there is either a bug in the delegation of
	// WrapInCreate, behavior is not lining up in the generated ExtendedBy
	// code, or something else is incorrect with the type system.
	var ok bool
	activity, ok = asValue.(Activity)
	if !ok {
		err = fmt.Errorf("activity streams value is not an Activity: %T", asValue)
		return
	}
	// Delegate generating new IDs for the activity and all new objects.
	if err = b.delegate.AddNewIDs(c, activity); err != nil {
		return
	}
	// Post the activity to the actor's outbox and trigger side effects for
	// that particular Activity type.
	//
	// Since 'm' is nil-able and side effects may need access to literal nil
	// values, such as for Update activities, ensure 'm' is non-nil.
	if m == nil {
		m, err = asValue.Serialize()
		if err != nil {
			return
		}
	}
	deliverable, err := b.delegate.PostOutbox(c, activity, outbox, m)
	if err != nil {
		return
	}
	// Request has been processed and all side effects internal to this
	// application server have finished. Begin side effects affecting other
	// servers and/or the client who sent this request.
	//
	// If we are federating and the type is a deliverable one, then deliver
	// the activity to federating peers.
	if b.enableFederatedProtocol && deliverable {
		if err = b.delegate.Deliver(c, outbox, activity); err != nil {
			return
		}
	}
	return
}

// Send is programmatically accessible if the federated protocol is enabled.
func (b *baseActorFederating) Send(c context.Context, outbox *url.URL, t vocab.Type) (Activity, error) {
	return b.deliver(c, outbox, t, nil)
}
