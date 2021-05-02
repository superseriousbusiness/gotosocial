/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package federation

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
)

// federatingActor implements the go-fed federating protocol interface
type federatingActor struct {
	actor pub.FederatingActor
}

// newFederatingProtocol returns the gotosocial implementation of the GTSFederatingProtocol interface
func newFederatingActor(c pub.CommonBehavior, s2s pub.FederatingProtocol, db pub.Database, clock pub.Clock) pub.FederatingActor {
	actor := pub.NewFederatingActor(c, s2s, db, clock)

	return &federatingActor{
		actor: actor,
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
	return f.actor.Send(c, outbox, t)
}

// PostInbox returns true if the request was handled as an ActivityPub
// POST to an actor's inbox. If false, the request was not an
// ActivityPub request and may still be handled by the caller in
// another way, such as serving a web page.
//
// If the error is nil, then the ResponseWriter's headers and response
// has already been written. If a non-nil error is returned, then no
// response has been written.
//
// If the Actor was constructed with the Federated Protocol enabled,
// side effects will occur.
//
// If the Federated Protocol is not enabled, writes the
// http.StatusMethodNotAllowed status code in the response. No side
// effects occur.
func (f *federatingActor) PostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	return f.actor.PostInbox(c, w, r)
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
