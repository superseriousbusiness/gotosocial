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
	"net/http"
	"net/url"
	"strings"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type errOtherIRIBlocked struct {
	account     string
	domainBlock bool
	iriStrs     []string
}

func (e *errOtherIRIBlocked) Error() string {
	iriStrsNice := "[" + strings.Join(e.iriStrs, ", ") + "]"
	if e.domainBlock {
		return "domain block exists for one or more of " + iriStrsNice
	}
	return "block exists between " + e.account + " and one or more of " + iriStrsNice
}

func newErrOtherIRIBlocked(
	account string,
	domainBlock bool,
	otherIRIs []*url.URL,
) error {
	e := errOtherIRIBlocked{
		account:     account,
		domainBlock: domainBlock,
		iriStrs:     make([]string, 0, len(otherIRIs)),
	}

	for _, iri := range otherIRIs {
		e.iriStrs = append(e.iriStrs, iri.String())
	}

	return &e
}

/*
	GO FED FEDERATING PROTOCOL INTERFACE
	FederatingProtocol contains behaviors an application needs to satisfy for the
	full ActivityPub S2S implementation to be supported by this library.
	It is only required if the client application wants to support the server-to-
	server, or federating, protocol.
	It is passed to the library as a dependency injection from the client
	application.
*/

// PostInboxRequestBodyHook callback after parsing the request body for a
// federated request to the Actor's inbox.
//
// Can be used to set contextual information based on the Activity received.
//
// Warning: Neither authentication nor authorization has taken place at
// this time. Doing anything beyond setting contextual information is
// strongly discouraged.
//
// If an error is returned, it is passed back to the caller of PostInbox.
// In this case, the DelegateActor implementation must not write a response
// to the ResponseWriter as is expected that the caller to PostInbox will
// do so when handling the error.
func (f *Federator) PostInboxRequestBodyHook(ctx context.Context, r *http.Request, activity pub.Activity) (context.Context, error) {
	// Extract any other IRIs involved in this activity.
	otherIRIs := []*url.URL{}

	// Get the ID of the Activity itslf.
	activityID, err := pub.GetId(activity)
	if err == nil {
		otherIRIs = append(otherIRIs, activityID)
	}

	// Check if the Activity has an 'inReplyTo'.
	if replyToable, ok := activity.(ap.ReplyToable); ok {
		if inReplyToURI := ap.ExtractInReplyToURI(replyToable); inReplyToURI != nil {
			otherIRIs = append(otherIRIs, inReplyToURI)
		}
	}

	// Check for TO and CC URIs on the Activity.
	if addressable, ok := activity.(ap.Addressable); ok {
		otherIRIs = append(otherIRIs, ap.ExtractToURIs(addressable)...)
		otherIRIs = append(otherIRIs, ap.ExtractCcURIs(addressable)...)
	}

	// Now perform the same checks, but for the Object(s) of the Activity.
	objectProp := activity.GetActivityStreamsObject()
	for iter := objectProp.Begin(); iter != objectProp.End(); iter = iter.Next() {
		if iter.IsIRI() {
			otherIRIs = append(otherIRIs, iter.GetIRI())
			continue
		}

		t := iter.GetType()
		if t == nil {
			continue
		}

		objectID, err := pub.GetId(t)
		if err == nil {
			otherIRIs = append(otherIRIs, objectID)
		}

		if replyToable, ok := t.(ap.ReplyToable); ok {
			if inReplyToURI := ap.ExtractInReplyToURI(replyToable); inReplyToURI != nil {
				otherIRIs = append(otherIRIs, inReplyToURI)
			}
		}

		if addressable, ok := t.(ap.Addressable); ok {
			otherIRIs = append(otherIRIs, ap.ExtractToURIs(addressable)...)
			otherIRIs = append(otherIRIs, ap.ExtractCcURIs(addressable)...)
		}
	}

	// Clean any instances of the public URI, since
	// we don't care about that in this context.
	otherIRIs = func(iris []*url.URL) []*url.URL {
		np := make([]*url.URL, 0, len(iris))

		for _, i := range iris {
			if !pub.IsPublic(i.String()) {
				np = append(np, i)
			}
		}

		return np
	}(otherIRIs)

	// OtherIRIs will likely contain some
	// duplicate entries now, so remove them.
	otherIRIs = util.DeduplicateFunc(otherIRIs,
		(*url.URL).String, // serialized URL is 'key()'
	)

	// Finished, set other IRIs on the context
	// so they can be checked for blocks later.
	ctx = gtscontext.SetOtherIRIs(ctx, otherIRIs)
	return ctx, nil
}

// AuthenticatePostInbox delegates the authentication of a POST to an
// inbox.
//
// If an error is returned, it is passed back to the caller of
// PostInbox. In this case, the implementation must not write a
// response to the ResponseWriter as is expected that the client will
// do so when handling the error. The 'authenticated' is ignored.
//
// If no error is returned, but authentication or authorization fails,
// then authenticated must be false and error nil. It is expected that
// the implementation handles writing to the ResponseWriter in this
// case.
//
// Finally, if the authentication and authorization succeeds, then
// authenticated must be true and error nil. The request will continue
// to be processed.
func (f *Federator) AuthenticatePostInbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, bool, error) {
	log.Tracef(ctx, "received request to authenticate inbox %s", r.URL.String())

	// Ensure this is an inbox path, and fetch the inbox owner
	// account by parsing username from `/users/{username}/inbox`.
	username, err := uris.ParseInboxPath(r.URL)
	if err != nil {
		err = gtserror.Newf("could not parse %s as inbox path: %w", r.URL.String(), err)
		return nil, false, err
	}

	if username == "" {
		err = gtserror.New("inbox username was empty")
		return nil, false, err
	}

	receivingAccount, err := f.db.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		err = gtserror.Newf("could not fetch receiving account %s: %w", username, err)
		return nil, false, err
	}

	// Check who's trying to deliver to us by inspecting the http signature.
	pubKeyAuth, errWithCode := f.AuthenticateFederatedRequest(ctx, receivingAccount.Username)
	if errWithCode != nil {
		switch errWithCode.Code() {
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest:
			// If codes 400, 401, or 403, obey the go-fed
			// interface by writing the header and bailing.
			w.WriteHeader(errWithCode.Code())
		case http.StatusGone:
			// If the requesting account's key has gone
			// (410) then likely inbox post was a delete.
			//
			// We can just write 202 and leave: we didn't
			// know about the account anyway, so we can't
			// do any further processing.
			w.WriteHeader(http.StatusAccepted)
		}

		// We still return the error
		// for later request logging.
		return ctx, false, errWithCode
	}

	if pubKeyAuth.Handshaking {
		// There is a mutal handshake occurring between us and
		// the owner URI. Return 202 and leave as we can't do
		// much else until the handshake procedure has finished.
		w.WriteHeader(http.StatusAccepted)
		return ctx, false, nil
	}

	// We have everything we need now, set the requesting
	// and receiving accounts on the context for later use.
	ctx = gtscontext.SetRequestingAccount(ctx, pubKeyAuth.Owner)
	ctx = gtscontext.SetReceivingAccount(ctx, receivingAccount)
	return ctx, true, nil
}

// Blocked should determine whether to permit a set of actors given by
// their ids are able to interact with this particular end user due to
// being blocked or other application-specific logic.
func (f *Federator) Blocked(ctx context.Context, actorIRIs []*url.URL) (bool, error) {
	// Fetch relevant items from request context.
	// These should have been set further up the flow.
	receivingAccount := gtscontext.ReceivingAccount(ctx)
	if receivingAccount == nil {
		err := gtserror.New("couldn't determine blocks (receiving account not set on request context)")
		return false, err
	}

	requestingAccount := gtscontext.RequestingAccount(ctx)
	if requestingAccount == nil {
		err := gtserror.New("couldn't determine blocks (requesting account not set on request context)")
		return false, err
	}

	otherIRIs := gtscontext.OtherIRIs(ctx)
	if otherIRIs == nil {
		err := gtserror.New("couldn't determine blocks (otherIRIs not set on request context)")
		return false, err
	}

	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"actorIRIs", actorIRIs},
			{"receivingAccount", receivingAccount.URI},
			{"requestingAccount", requestingAccount.URI},
			{"otherIRIs", otherIRIs},
		}...)
	l.Trace("checking blocks")

	// Start broad by checking domain-level blocks first for
	// the given actor IRIs; if any of them are domain blocked
	// then we can save some work.
	blocked, err := f.db.AreURIsBlocked(ctx, actorIRIs)
	if err != nil {
		err = gtserror.Newf("error checking domain blocks of actorIRIs: %w", err)
		return false, err
	}

	if blocked {
		l.Trace("one or more actorIRIs are domain blocked")
		return blocked, nil
	}

	// Now user level blocks. Receiver should not block requester.
	blocked, err = f.db.IsBlocked(ctx, receivingAccount.ID, requestingAccount.ID)
	if err != nil {
		err = gtserror.Newf("db error checking block between receiver and requester: %w", err)
		return false, err
	}

	if blocked {
		l.Trace("receiving account blocks requesting account")
		return blocked, nil
	}

	// We've established that no blocks exist between directly
	// involved actors, but what about IRIs of other actors and
	// objects which are tangentially involved in the activity
	// (ie., replied to, boosted)?
	//
	// If one or more of these other IRIs is domain blocked, or
	// blocked by the receiving account, this shouldn't return
	// blocked=true to send a 403, since that would be rather
	// silly behavior. Instead, we should indicate to the caller
	// that we should stop processing the activity and just write
	// 202 Accepted instead.
	//
	// For this, we can use the errOtherIRIBlocked type, which
	// will be checked for

	// Check high-level domain blocks first.
	blocked, err = f.db.AreURIsBlocked(ctx, otherIRIs)
	if err != nil {
		err := gtserror.Newf("error checking domain block of otherIRIs: %w", err)
		return false, err
	}

	if blocked {
		err := newErrOtherIRIBlocked(receivingAccount.URI, true, otherIRIs)
		l.Trace(err.Error())
		return false, err
	}

	// For each other IRI, check whether the IRI points to an
	// account or a status, and try to get (an) accountID(s)
	// from it to do further checks on.
	//
	// We use a map for this instead of a slice in order to
	// deduplicate entries and avoid doing the same check twice.
	// The map value is the host of the otherIRI.
	accountIDs := make(map[string]string, len(otherIRIs))
	for _, iri := range otherIRIs {
		// Assemble iri string just once.
		iriStr := iri.String()

		account, err := f.db.GetAccountByURI(
			// We're on a hot path, fetch bare minimum.
			gtscontext.SetBarebones(ctx),
			iriStr,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			// Real db error.
			err = gtserror.Newf("db error trying to get %s as account: %w", iriStr, err)
			return false, err
		} else if err == nil {
			// IRI is for an account.
			accountIDs[account.ID] = iri.Host
			continue
		}

		status, err := f.db.GetStatusByURI(
			// We're on a hot path, fetch bare minimum.
			gtscontext.SetBarebones(ctx),
			iriStr,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			// Real db error.
			err = gtserror.Newf("db error trying to get %s as status: %w", iriStr, err)
			return false, err
		} else if err == nil {
			// IRI is for a status.
			accountIDs[status.AccountID] = iri.Host
			continue
		}
	}

	// Get our own host value just once outside the loop.
	ourHost := config.GetHost()

	for accountID, iriHost := range accountIDs {
		// Receiver shouldn't block other IRI owner.
		//
		// This check protects against cases where someone on our
		// instance is receiving a boost from someone they don't
		// block, but the boost target is the status of an account
		// they DO have blocked, or the boosted status mentions an
		// account they have blocked. In this case, it's v. unlikely
		// they care to see the boost in their timeline, so there's
		// no point in us processing it.
		blocked, err = f.db.IsBlocked(ctx, receivingAccount.ID, accountID)
		if err != nil {
			err = gtserror.Newf("db error checking block between receiver and other account: %w", err)
			return false, err
		}

		if blocked {
			l.Trace("receiving account blocks one or more otherIRIs")
			err := newErrOtherIRIBlocked(receivingAccount.URI, false, otherIRIs)
			return false, err
		}

		// If other account is from our instance (indicated by the
		// host of the URI stored in the map), ensure they don't block
		// the requester.
		//
		// This check protects against cases where one of our users
		// might be mentioned by the requesting account, and therefore
		// appear in otherIRIs, but the activity itself has been sent
		// to a different account on our instance. In other words, two
		// accounts are gossiping about + trying to tag a third account
		// who has one or the other of them blocked.
		if iriHost == ourHost {
			blocked, err = f.db.IsBlocked(ctx, accountID, requestingAccount.ID)
			if err != nil {
				err = gtserror.Newf("db error checking block between other account and requester: %w", err)
				return false, err
			}

			if blocked {
				l.Trace("one or more otherIRIs belonging to us blocks requesting account")
				err := newErrOtherIRIBlocked(requestingAccount.URI, false, otherIRIs)
				return false, err
			}
		}
	}

	return false, nil
}

// FederatingCallbacks returns the application logic that handles
// ActivityStreams received from federating peers.
//
// Note that certain types of callbacks will be 'wrapped' with default
// behaviors supported natively by the library. Other callbacks
// compatible with streams.TypeResolver can be specified by 'other'.
//
// For example, setting the 'Create' field in the
// FederatingWrappedCallbacks lets an application dependency inject
// additional behaviors they want to take place, including the default
// behavior supplied by this library. This is guaranteed to be compliant
// with the ActivityPub Social protocol.
//
// To override the default behavior, instead supply the function in
// 'other', which does not guarantee the application will be compliant
// with the ActivityPub Social Protocol.
//
// Applications are not expected to handle every single ActivityStreams
// type and extension. The unhandled ones are passed to DefaultCallback.
func (f *Federator) FederatingCallbacks(ctx context.Context) (
	wrapped pub.FederatingWrappedCallbacks,
	other []any,
	err error,
) {
	wrapped = pub.FederatingWrappedCallbacks{
		// OnFollow determines what action to take for this
		// particular callback if a Follow Activity is handled.
		//
		// For our implementation, we always want to do nothing
		// because we have internal logic for handling follows.
		OnFollow: pub.OnFollowDoNothing,
	}

	// Override some default behaviors to trigger our own side effects.
	other = []any{
		func(ctx context.Context, undo vocab.ActivityStreamsUndo) error {
			return f.FederatingDB().Undo(ctx, undo)
		},
		func(ctx context.Context, accept vocab.ActivityStreamsAccept) error {
			return f.FederatingDB().Accept(ctx, accept)
		},
		func(ctx context.Context, reject vocab.ActivityStreamsReject) error {
			return f.FederatingDB().Reject(ctx, reject)
		},
		func(ctx context.Context, announce vocab.ActivityStreamsAnnounce) error {
			return f.FederatingDB().Announce(ctx, announce)
		},
	}

	// Define some of our own behaviors which are not
	// overrides of the default pub.FederatingWrappedCallbacks.
	other = append(other, []any{
		func(ctx context.Context, move vocab.ActivityStreamsMove) error {
			return f.FederatingDB().Move(ctx, move)
		},
	}...)

	return
}

// DefaultCallback is called for types that go-fed can deserialize but
// are not handled by the application's callbacks returned in the
// Callbacks method.
//
// Applications are not expected to handle every single ActivityStreams
// type and extension, so the unhandled ones are passed to
// DefaultCallback.
func (f *Federator) DefaultCallback(ctx context.Context, activity pub.Activity) error {
	log.Debugf(ctx, "received unhandle-able activity type (%s) so ignoring it", activity.GetTypeName())
	return nil
}

// MaxInboxForwardingRecursionDepth determines how deep to search within
// an activity to determine if inbox forwarding needs to occur.
//
// Zero or negative numbers indicate infinite recursion.
func (f *Federator) MaxInboxForwardingRecursionDepth(ctx context.Context) int {
	// TODO
	return 4
}

// MaxDeliveryRecursionDepth determines how deep to search within
// collections owned by peers when they are targeted to receive a
// delivery.
//
// Zero or negative numbers indicate infinite recursion.
func (f *Federator) MaxDeliveryRecursionDepth(ctx context.Context) int {
	// TODO
	return 4
}

// FilterForwarding allows the implementation to apply business logic
// such as blocks, spam filtering, and so on to a list of potential
// Collections and OrderedCollections of recipients when inbox
// forwarding has been triggered.
//
// The activity is provided as a reference for more intelligent
// logic to be used, but the implementation must not modify it.
func (f *Federator) FilterForwarding(ctx context.Context, potentialRecipients []*url.URL, a pub.Activity) ([]*url.URL, error) {
	// TODO
	return []*url.URL{}, nil
}

// GetInbox returns the OrderedCollection inbox of the actor for this
// context. It is up to the implementation to provide the correct
// collection for the kind of authorization given in the request.
//
// AuthenticateGetInbox will be called prior to this.
//
// Always called, regardless whether the Federated Protocol or Social
// API is enabled.
func (f *Federator) GetInbox(ctx context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// IMPLEMENTATION NOTE: For GoToSocial, we serve GETS to outboxes and inboxes through
	// the CLIENT API, not through the federation API, so we just do nothing here.
	return streams.NewActivityStreamsOrderedCollectionPage(), nil
}
