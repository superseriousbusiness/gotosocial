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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	errorsv2 "codeberg.org/gruf/go-errors/v2"
	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// IsASMediaType will return whether the given content-type string
// matches one of the 2 possible ActivityStreams incoming content types:
// - application/activity+json
// - application/ld+json;profile=https://w3.org/ns/activitystreams
//
// Where for the above we are leniant with whitespace, quotes, and charset.
func IsASMediaType(ct string) bool {
	var (
		// First content-type part,
		// contains the application/...
		p1 string = ct //nolint:revive

		// Second content-type part,
		// contains AS IRI or charset
		// if provided.
		p2 string
	)

	// Split content-type by semi-colon.
	sep := strings.IndexByte(ct, ';')
	if sep >= 0 {
		p1 = ct[:sep]

		// Trim all start/end
		// space of second part.
		p2 = ct[sep+1:]
		p2 = strings.Trim(p2, " ")
	}

	// Trim any ending space from the
	// main content-type part of string.
	p1 = strings.TrimRight(p1, " ")

	switch p1 {
	case "application/activity+json":
		// Accept with or without charset.
		// This should be case insensitive.
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type#charset
		return p2 == "" || strings.EqualFold(p2, "charset=utf-8")

	case "application/ld+json":
		// Drop any quotes around the URI str.
		p2 = strings.ReplaceAll(p2, "\"", "")

		// End part must be a ref to the main AS namespace IRI.
		return p2 == "profile=https://www.w3.org/ns/activitystreams"

	default:
		return false
	}
}

// federatingActor wraps the pub.FederatingActor
// with some custom GoToSocial-specific logic.
type federatingActor struct {
	sideEffectActor pub.DelegateActor
	wrapped         pub.FederatingActor
}

// newFederatingActor returns a federatingActor.
func newFederatingActor(c pub.CommonBehavior, s2s pub.FederatingProtocol, db pub.Database, clock pub.Clock) pub.FederatingActor {
	sideEffectActor := pub.NewSideEffectActor(c, s2s, nil, db, clock)
	sideEffectActor.Serialize = ap.Serialize // hook in our own custom Serialize function

	return &federatingActor{
		sideEffectActor: sideEffectActor,
		wrapped:         pub.NewCustomActor(sideEffectActor, false, true, clock),
	}
}

// PostInboxScheme is a reimplementation of the default baseActor
// implementation of PostInboxScheme in pub/base_actor.go.
//
// Key differences from that implementation:
//   - More explicit debug logging when a request is not processed.
//   - Normalize content of activity object.
//   - *ALWAYS* return gtserror.WithCode if there's an issue, to
//     provide more helpful messages to remote callers.
//   - Return code 202 instead of 200 on successful POST, to reflect
//     that we process most side effects asynchronously.
func (f *federatingActor) PostInboxScheme(ctx context.Context, w http.ResponseWriter, r *http.Request, scheme string) (bool, error) {
	l := log.WithContext(ctx).
		WithFields([]kv.Field{
			{"userAgent", r.UserAgent()},
			{"path", r.URL.Path},
		}...)

	// Ensure valid ActivityPub Content-Type.
	// https://www.w3.org/TR/activitypub/#server-to-server-interactions
	if ct := r.Header.Get("Content-Type"); !IsASMediaType(ct) {
		const ct1 = "application/activity+json"
		const ct2 = "application/activity+json;charset=utf-8"
		const ct3 = "application/ld+json;profile=https://w3.org/ns/activitystreams"
		err := fmt.Errorf("Content-Type %s not acceptable, this endpoint accepts: [%q %q]", ct, ct1, ct2)
		return false, gtserror.NewErrorNotAcceptable(err)
	}

	// Authenticate request by checking http signature.
	ctx, authenticated, err := f.sideEffectActor.AuthenticatePostInbox(ctx, w, r)
	if err != nil {
		err := gtserror.Newf("error authenticating post inbox: %w", err)
		return false, gtserror.NewErrorInternalError(err)
	}

	if !authenticated {
		const text = "not authenticated"
		return false, gtserror.NewErrorUnauthorized(errors.New(text), text)
	}

	/*
		Begin processing the request, but note that we
		have not yet applied authorization (ie., blocks).
	*/

	// Resolve the activity, rejecting badly formatted / transient.
	activity, ok, errWithCode := ap.ResolveIncomingActivity(r)
	if errWithCode != nil {
		return false, errWithCode
	} else if !ok { // transient
		return false, nil
	}

	// Set additional context data. Primarily this means
	// looking at the Activity and seeing which IRIs are
	// involved in it tangentially.
	ctx, err = f.sideEffectActor.PostInboxRequestBodyHook(ctx, r, activity)
	if err != nil {
		err := gtserror.Newf("error during post inbox request body hook: %w", err)
		return false, gtserror.NewErrorInternalError(err)
	}

	// Check authorization of the activity; this will include blocks.
	authorized, err := f.sideEffectActor.AuthorizePostInbox(ctx, w, activity)
	if err != nil {
		if errors.As(err, new(errOtherIRIBlocked)) {
			// There's no direct block between requester(s) and
			// receiver. However, one or more of the other IRIs
			// involved in the request (account replied to, note
			// boosted, etc) is blocked either at domain level or
			// by the receiver. We don't need to return 403 here,
			// instead, just return 202 accepted but don't do any
			// further processing of the activity.
			return true, nil
		}

		// Real error has occurred.
		err := gtserror.Newf("error authorizing post inbox: %w", err)
		return false, gtserror.NewErrorInternalError(err)
	}

	if !authorized {
		// Block exists either from this instance against
		// one or more directly involved actors, or between
		// receiving account and one of those actors.
		const text = "blocked"
		return false, gtserror.NewErrorForbidden(errors.New(text), text)
	}

	// Copy existing URL + add request host and scheme.
	inboxID := func() *url.URL {
		u := new(url.URL)
		*u = *r.URL
		u.Host = r.Host
		u.Scheme = scheme
		return u
	}()

	// At this point we have everything we need, and have verified that
	// the POST request is authentic (properly signed) and authorized
	// (permitted to interact with the target inbox).
	//
	// Post the activity to the Actor's inbox and trigger side effects .
	if err := f.sideEffectActor.PostInbox(ctx, inboxID, activity); err != nil {
		// Special case: We know it is a bad request if the object or target
		// props needed to be populated, or we failed parsing activity details.
		// Send the rejection to the peer.
		if errors.Is(err, pub.ErrObjectRequired) ||
			errors.Is(err, pub.ErrTargetRequired) ||
			gtserror.IsMalformed(err) {

			// Log malformed activities to help debug.
			l = l.WithField("activity", activity)
			l.Warnf("malformed incoming activity: %v", err)

			const text = "malformed incoming activity"
			return false, gtserror.NewErrorBadRequest(errors.New(text), text)
		}

		// There's been some real error.
		err := gtserror.Newf("error calling sideEffectActor.PostInbox: %w", err)
		return false, gtserror.NewErrorInternalError(err)
	}

	// Side effects are complete. Now delegate determining whether
	// to do inbox forwarding, as well as the action to do it.
	if err := f.sideEffectActor.InboxForwarding(ctx, inboxID, activity); err != nil {
		// As a not-ideal side-effect, InboxForwarding will try
		// to create entries if the federatingDB returns `false`
		// when calling `Exists()` to determine whether the Activity
		// is in the database.
		//
		// Since our `Exists()` function currently *always*
		// returns false, it will *always* attempt to parse
		// out and insert the Activity, trying to fetch other
		// items from the DB in the process, which may or may
		// not exist yet. Therefore, we should expect some
		// errors coming from this function, and only warn log
		// on certain ones.
		//
		// This check may be removed when the `Exists()` func
		// is updated, and/or federating callbacks are handled
		// properly.
		if !errorsv2.IsV2(
			err,
			db.ErrAlreadyExists,
			db.ErrNoEntries,
		) {
			// Failed inbox forwarding is not a show-stopper,
			// and doesn't even necessarily denote a real error.
			l.Warnf("error calling sideEffectActor.InboxForwarding: %v", err)
		}
	}

	// Request is now undergoing processing. Caller
	// of this function will handle writing Accepted.
	return true, nil
}

/*
	Functions below are just lightly wrapped versions
	of the original go-fed federatingActor functions.
*/

func (f *federatingActor) PostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return f.PostInboxScheme(c, w, r, "https")
}

func (f *federatingActor) Send(c context.Context, outbox *url.URL, t vocab.Type) (pub.Activity, error) {
	log.Infof(c, "send activity %s via outbox %s", t.GetTypeName(), outbox)
	return f.wrapped.Send(c, outbox, t)
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
