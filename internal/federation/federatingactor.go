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

package federation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// federatingActor implements the go-fed federating protocol interface
type federatingActor struct {
	actor pub.FederatingActor
}

// newFederatingProtocol returns the gotosocial implementation of the GTSFederatingProtocol interface
func newFederatingActor(c pub.CommonBehavior, s2s pub.FederatingProtocol, db pub.Database, clock pub.Clock) pub.FederatingActor {
	sideEffectActor := pub.NewSideEffectActor(c, s2s, nil, db, clock)
	sideEffectActor.Serialize = ap.Serialize // hook in our own custom Serialize function

	return &federatingActor{
		sideEffectActor: sideEffectActor,
		wrapped:         pub.NewCustomActor(sideEffectActor, false, true, clock),
	}
}

// Send a federated activity.
//
// The provided url must be the outbox of the sender. All processing of
// the activity occurs similarly to the C2S flow:
//   - If t is not an Activity, it is wrapped in a Create activity.
//   - A new ID is generated for the activity.
//   - The activity is added to the specified outbox.
//   - The activity is prepared and delivered to recipients.
//
// Note that this function will only behave as expected if the
// implementation has been constructed to support federation. This
// method will guaranteed work for non-custom Actors. For custom actors,
// care should be used to not call this method if only C2S is supported.
func (f *federatingActor) Send(c context.Context, outbox *url.URL, t vocab.Type) (pub.Activity, error) {
	log.Infof(c, "send activity %s via outbox %s", t.GetTypeName(), outbox)
	return f.actor.Send(c, outbox, t)
}

// PostInbox returns true if the request was handled as an ActivityPub
// POST to an actor's inbox. If false, the request was not an
// ActivityPub request and may still be handled by the caller in
// another way, such as serving a web page.
//
// Key differences from that implementation:
//   - More explicit debug logging when a request is not processed.
//   - Normalize content of activity object.
//   - Return code 202 instead of 200 on successful POST, to reflect
//     that we process most side effects asynchronously.
func (f *federatingActor) PostInboxScheme(ctx context.Context, w http.ResponseWriter, r *http.Request, scheme string) (bool, error) {
	l := log.
		WithContext(ctx).
		WithFields([]kv.Field{
			{"userAgent", r.UserAgent()},
			{"path", r.URL.Path},
		}...)

	// Do nothing if this is not an ActivityPub POST request.
	if !func() bool {
		if r.Method != http.MethodPost {
			l.Debugf("inbox request was %s rather than required POST", r.Method)
			return false
		}

		contentType := r.Header.Get("Content-Type")
		for _, mediaType := range activityStreamsMediaTypes {
			if strings.Contains(contentType, mediaType) {
				return true
			}
		}

		l.Debugf("inbox POST request content-type %s was not recognized", contentType)
		return false
	}() {
		return false, nil
	}

	// Check the peer request is authentic.
	ctx, authenticated, err := f.sideEffectActor.AuthenticatePostInbox(ctx, w, r)
	if err != nil {
		return true, err
	} else if !authenticated {
		return true, nil
	}

	// Begin processing the request, but note that we have
	// not yet applied authorization (ex: blocks).
	//
	// Obtain the activity and reject unknown activities.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("PostInboxScheme: error reading request body: %w", err)
		return true, err
	}

	var rawActivity map[string]interface{}
	if err := json.Unmarshal(b, &rawActivity); err != nil {
		err = fmt.Errorf("PostInboxScheme: error unmarshalling request body: %w", err)
		return true, err
	}

	t, err := streams.ToType(ctx, rawActivity)
	if err != nil {
		if !streams.IsUnmatchedErr(err) {
			// Real error.
			err = fmt.Errorf("PostInboxScheme: error matching json to type: %w", err)
			return true, err
		}
		// Respond with bad request; we just couldn't
		// match the type to one that we know about.
		l.Debug("json could not be resolved to ActivityStreams value")
		w.WriteHeader(http.StatusBadRequest)
		return true, nil
	}

	activity, ok := t.(pub.Activity)
	if !ok {
		err = fmt.Errorf("ActivityStreams value with type %T is not a pub.Activity", t)
		return true, err
	}

	if activity.GetJSONLDId() == nil {
		l.Debugf("incoming Activity %s did not have required id property set", activity.GetTypeName())
		w.WriteHeader(http.StatusBadRequest)
		return true, nil
	}

	// If activity Object is a Statusable, we'll want to replace the
	// parsed `content` value with the value from the raw JSON instead.
	// See https://github.com/superseriousbusiness/gotosocial/issues/1661
	// Likewise, if it's an Accountable, we'll normalize some fields on it.
	ap.NormalizeIncomingActivityObject(activity, rawActivity)

	// Allow server implementations to set context data with a hook.
	ctx, err = f.sideEffectActor.PostInboxRequestBodyHook(ctx, r, activity)
	if err != nil {
		return true, err
	}

	// Check authorization of the activity.
	authorized, err := f.sideEffectActor.AuthorizePostInbox(ctx, w, activity)
	if err != nil {
		return true, err
	} else if !authorized {
		return true, nil
	}

	// Copy existing URL + add request host and scheme.
	inboxID := func() *url.URL {
		id := &url.URL{}
		*id = *r.URL
		id.Host = r.Host
		id.Scheme = scheme
		return id
	}()

	// Post the activity to the actor's inbox and trigger side effects for
	// that particular Activity type. It is up to the delegate to resolve
	// the given map.
	if err := f.sideEffectActor.PostInbox(ctx, inboxID, activity); err != nil {
		// Special case: We know it is a bad request if the object or
		// target properties needed to be populated, but weren't.
		//
		// Send the rejection to the peer.
		if err == pub.ErrObjectRequired || err == pub.ErrTargetRequired {
			l.Debugf("malformed incoming Activity: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return true, nil
		}
		err = fmt.Errorf("PostInboxScheme: error calling sideEffectActor.PostInbox: %w", err)
		return true, err
	}

	// Our side effects are complete, now delegate determining whether to do inbox forwarding, as well as the action to do it.
	if err := f.sideEffectActor.InboxForwarding(ctx, inboxID, activity); err != nil {
		err = fmt.Errorf("PostInboxScheme: error calling sideEffectActor.InboxForwarding: %w", err)
		return true, err
	}

	// Request is now undergoing processing.
	// Respond with an Accepted status.
	w.WriteHeader(http.StatusAccepted)
	return true, nil
}

// PostInboxScheme is similar to PostInbox, except clients are able to
// specify which protocol scheme to handle the incoming request and the
// data stored within the application (HTTP, HTTPS, etc).
func (f *federatingActor) PostInboxScheme(c context.Context, w http.ResponseWriter, r *http.Request, scheme string) (bool, error) {
	return f.actor.PostInboxScheme(c, w, r, scheme)
}

// GetInbox returns true if the request was handled as an ActivityPub
// GET to an actor's inbox. If false, the request was not an ActivityPub
// request and may still be handled by the caller in another way, such
// as serving a web page.
//
// If the error is nil, then the ResponseWriter's headers and response
// has already been written. If a non-nil error is returned, then no
// response has been written.
//
// If the request is an ActivityPub request, the Actor will defer to the
// application to determine the correct authorization of the request and
// the resulting OrderedCollection to respond with. The Actor handles
// serializing this OrderedCollection and responding with the correct
// headers and http.StatusOK.
func (f *federatingActor) GetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return f.actor.GetInbox(c, w, r)
}

// PostOutbox returns true if the request was handled as an ActivityPub
// POST to an actor's outbox. If false, the request was not an
// ActivityPub request and may still be handled by the caller in another
// way, such as serving a web page.
//
// If the error is nil, then the ResponseWriter's headers and response
// has already been written. If a non-nil error is returned, then no
// response has been written.
//
// If the Actor was constructed with the Social Protocol enabled, side
// effects will occur.
//
// If the Social Protocol is not enabled, writes the
// http.StatusMethodNotAllowed status code in the response. No side
// effects occur.
//
// If the Social and Federated Protocol are both enabled, it will handle
// the side effects of receiving an ActivityStream Activity, and then
// federate the Activity to peers.
func (f *federatingActor) PostOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return f.actor.PostOutbox(c, w, r)
}

// PostOutboxScheme is similar to PostOutbox, except clients are able to
// specify which protocol scheme to handle the incoming request and the
// data stored within the application (HTTP, HTTPS, etc).
func (f *federatingActor) PostOutboxScheme(c context.Context, w http.ResponseWriter, r *http.Request, scheme string) (bool, error) {
	return f.actor.PostOutboxScheme(c, w, r, scheme)
}

// GetOutbox returns true if the request was handled as an ActivityPub
// GET to an actor's outbox. If false, the request was not an
// ActivityPub request.
//
// If the error is nil, then the ResponseWriter's headers and response
// has already been written. If a non-nil error is returned, then no
// response has been written.
//
// If the request is an ActivityPub request, the Actor will defer to the
// application to determine the correct authorization of the request and
// the resulting OrderedCollection to respond with. The Actor handles
// serializing this OrderedCollection and responding with the correct
// headers and http.StatusOK.
func (f *federatingActor) GetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return f.actor.GetOutbox(c, w, r)
}
