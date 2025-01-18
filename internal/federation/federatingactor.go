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

	errorsv2 "codeberg.org/gruf/go-errors/v2"
	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

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
	if ct := r.Header.Get("Content-Type"); !apiutil.ASContentType(ct) {
		const ct1 = "application/activity+json"
		const ct2 = "application/ld+json;profile=https://w3.org/ns/activitystreams"
		err := fmt.Errorf("Content-Type %s not acceptable, this endpoint accepts: [%q %q]", ct, ct1, ct2)
		return false, gtserror.NewErrorNotAcceptable(err)
	}

	// Authenticate request by checking http signature.
	//
	// NOTE: the behaviour here is a little strange as we have
	// the competing code styles of the go-fed interface expecting
	// that any 'err' is fatal, but 'authenticated' bool is intended to
	// be the main passer of whether failed auth occurred, but we in
	// the gts codebase use errors to pass-back non-200 status codes,
	// so we specifically have to check for already wrapped with code.
	//
	ctx, authenticated, err := f.sideEffectActor.AuthenticatePostInbox(ctx, w, r)
	if errorsv2.AsV2[gtserror.WithCode](err) != nil {
		// If it was already wrapped with an
		// HTTP code then don't bother rewrapping
		// it, just return it as-is for caller to
		// handle. AuthenticatePostInbox already
		// calls WriteHeader() in some situations.
		return false, err
	} else if err != nil {
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
		if errorsv2.AsV2[*errOtherIRIBlocked](err) != nil {
			// There's no direct block between requester(s) and
			// receiver. However, one or more of the other IRIs
			// involved in the request (account replied to, note
			// boosted, etc) is blocked either at domain level or
			// by the receiver. We don't need to return 403 here,
			// instead, just return 202 accepted but don't do any
			// further processing of the activity.
			return true, nil //nolint
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
	// Post the activity to the Actor's inbox and trigger side effects.
	if err := f.sideEffectActor.PostInbox(ctx, inboxID, activity); err != nil {
		// Check if it's a bad request because the
		// object or target props weren't populated,
		// or we failed parsing activity details.
		//
		// Log such activities to help debug, then
		// return the rejection (400) to the peer.
		if gtserror.IsMalformed(err) ||
			errors.Is(err, pub.ErrObjectRequired) ||
			errors.Is(err, pub.ErrTargetRequired) {

			l = l.WithField("activity", activity)
			l.Warnf("malformed incoming activity: %v", err)

			const text = "malformed incoming activity"
			return false, gtserror.NewErrorBadRequest(errors.New(text), text)
		}

		// Check if a function in the federatingDB
		// has returned an explicit errWithCode for us.
		if errWithCode, ok := err.(gtserror.WithCode); ok {
			return false, errWithCode
		}

		// Default: there's been some real error.
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
