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

// Package federation provides ActivityPub/federation functionality for GoToSocial
package federation

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/gotosocial/gotosocial/internal/db"
)

// New returns a go-fed compatible federating actor
func New(db db.DB) pub.FederatingActor {
	f := &Federator{}
	return pub.NewFederatingActor(f, f, db.Federation(), f)
}

// Federator implements several go-fed interfaces in one convenient location
type Federator struct {
}

// AuthenticateGetInbox determines whether the request is for a GET call to the Actor's Inbox.
func (f *Federator) AuthenticateGetInbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, bool, error) {
	// TODO
	return nil, false, nil
}

// AuthenticateGetOutbox determines whether the request is for a GET call to the Actor's Outbox.
func (f *Federator) AuthenticateGetOutbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, bool, error) {
	// TODO
	return nil, false, nil
}

// GetOutbox returns a proper paginated view of the Outbox for serving in a response.
func (f *Federator) GetOutbox(ctx context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// TODO
	return nil, nil
}

// NewTransport returns a new pub.Transport for federating with peer software.
func (f *Federator) NewTransport(ctx context.Context, actorBoxIRI *url.URL, gofedAgent string) (pub.Transport, error) {
	// TODO
	return nil, nil
}

func (f *Federator) PostInboxRequestBodyHook(ctx context.Context, r *http.Request, activity pub.Activity) (context.Context, error) {
	// TODO
	return nil, nil
}

func (f *Federator) AuthenticatePostInbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, bool, error) {
	// TODO
	return nil, false, nil
}

func (f *Federator) Blocked(ctx context.Context, actorIRIs []*url.URL) (bool, error) {
	// TODO
	return false, nil
}

func (f *Federator) FederatingCallbacks(ctx context.Context) (pub.FederatingWrappedCallbacks, []interface{}, error) {
	// TODO
	return pub.FederatingWrappedCallbacks{}, nil, nil
}

func (f *Federator) DefaultCallback(ctx context.Context, activity pub.Activity) error {
	// TODO
	return nil
}

func (f *Federator) MaxInboxForwardingRecursionDepth(ctx context.Context) int {
	// TODO
	return 0
}

func (f *Federator) MaxDeliveryRecursionDepth(ctx context.Context) int {
	// TODO
	return 0
}

func (f *Federator) FilterForwarding(ctx context.Context, potentialRecipients []*url.URL, a pub.Activity) ([]*url.URL, error) {
	// TODO
	return nil, nil
}

func (f *Federator) GetInbox(ctx context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// TODO
	return nil, nil
}

func (f *Federator) Now() time.Time {
	return time.Now()
}
