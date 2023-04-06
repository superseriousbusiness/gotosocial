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

// Potential incoming Content-Type header values; be
// lenient with whitespace and quotation mark placement.
var activityStreamsMediaTypes = []string{
	"application/activity+json",
	"application/ld+json;profile=https://www.w3.org/ns/activitystreams",
	"application/ld+json;profile=\"https://www.w3.org/ns/activitystreams\"",
	"application/ld+json ;profile=https://www.w3.org/ns/activitystreams",
	"application/ld+json ;profile=\"https://www.w3.org/ns/activitystreams\"",
	"application/ld+json ; profile=https://www.w3.org/ns/activitystreams",
	"application/ld+json ; profile=\"https://www.w3.org/ns/activitystreams\"",
	"application/ld+json; profile=https://www.w3.org/ns/activitystreams",
	"application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"",
}

// federatingActor wraps the pub.FederatingActor interface
// with some custom GoToSocial-specific logic.
type federatingActor struct {
	sideEffectActor pub.DelegateActor
	wrapped         pub.FederatingActor
}

// newFederatingProtocol returns a new federatingActor, which
// implements the pub.FederatingActor interface.
func newFederatingActor(c pub.CommonBehavior, s2s pub.FederatingProtocol, db pub.Database, clock pub.Clock) pub.FederatingActor {
	sideEffectActor := pub.NewSideEffectActor(c, s2s, nil, db, clock)
	customActor := pub.NewCustomActor(sideEffectActor, false, true, clock)

	return &federatingActor{
		sideEffectActor: sideEffectActor,
		wrapped:         customActor,
	}
}

func (f *federatingActor) Send(c context.Context, outbox *url.URL, t vocab.Type) (pub.Activity, error) {
	log.Infof(c, "send activity %s via outbox %s", t.GetTypeName(), outbox)
	return f.wrapped.Send(c, outbox, t)
}

func (f *federatingActor) PostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return f.PostInboxScheme(c, w, r, "https")
}

// PostInboxScheme is a reimplementation of the default baseActor
// implementation of PostInboxScheme in pub/base_actor.go.
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
	ap.NormalizeActivityObject(activity, rawActivity)

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

func (f *federatingActor) GetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return f.wrapped.GetInbox(c, w, r)
}

func (f *federatingActor) PostOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return f.wrapped.PostOutbox(c, w, r)
}

func (f *federatingActor) PostOutboxScheme(c context.Context, w http.ResponseWriter, r *http.Request, scheme string) (bool, error) {
	return f.wrapped.PostOutboxScheme(c, w, r, scheme)
}

func (f *federatingActor) GetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return f.wrapped.GetOutbox(c, w, r)
}
