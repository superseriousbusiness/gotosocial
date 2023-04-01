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
	"net/http"
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

// sideEffectActor wraps the go-fed pub sideEffectActor, which implements
// the DelegateActor interface, with a few tweaks specific to GoToSocial.
//
// sideEffectActor is NOT INTENDED FOR USE AS A C2S ACTOR.
type sideEffectActor struct {
	wrapped *pub.SideEffectActor
}

// newSideEffectActor returns a GoToSocial SideEffectActor, sans the c2s
// part. The returned actor should never be used as a C2S actor, because
// it will break things.
func newSideEffectActor(c pub.CommonBehavior, s2s pub.FederatingProtocol, db pub.Database, clock pub.Clock) *sideEffectActor {
	return &sideEffectActor{
		wrapped: pub.NewSideEffectActor(c, s2s, nil, db, clock),
	}
}

func (s *sideEffectActor) PostInboxRequestBodyHook(c context.Context, r *http.Request, activity pub.Activity) (context.Context, error) {
	return s.wrapped.PostInboxRequestBodyHook(c, r, activity)
}

func (s *sideEffectActor) PostOutboxRequestBodyHook(c context.Context, r *http.Request, data vocab.Type) (context.Context, error) {
	return s.wrapped.PostOutboxRequestBodyHook(c, r, data)
}

func (s *sideEffectActor) AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return s.wrapped.AuthenticatePostInbox(c, w, r)
}

func (s *sideEffectActor) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return s.wrapped.AuthenticateGetInbox(c, w, r)
}

func (s *sideEffectActor) AuthorizePostInbox(c context.Context, w http.ResponseWriter, activity pub.Activity) (authorized bool, err error) {
	return s.wrapped.AuthorizePostInbox(c, w, activity)
}

func (s *sideEffectActor) PostInbox(c context.Context, inboxIRI *url.URL, activity pub.Activity) error {
	return s.wrapped.PostInbox(c, inboxIRI, activity)
}

func (s *sideEffectActor) InboxForwarding(c context.Context, inboxIRI *url.URL, activity pub.Activity) error {
	return s.wrapped.InboxForwarding(c, inboxIRI, activity)
}

func (s *sideEffectActor) PostOutbox(c context.Context, a pub.Activity, outboxIRI *url.URL, rawJSON map[string]interface{}) (deliverable bool, e error) {
	return s.wrapped.PostOutbox(c, a, outboxIRI, rawJSON)
}

func (s *sideEffectActor) AddNewIDs(c context.Context, a pub.Activity) error {
	return s.wrapped.AddNewIDs(c, a)
}

func (s *sideEffectActor) Deliver(c context.Context, outbox *url.URL, activity pub.Activity) error {
	return s.wrapped.Deliver(c, outbox, activity)
}

func (s *sideEffectActor) AuthenticatePostOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return s.wrapped.AuthenticatePostOutbox(c, w, r)
}

func (s *sideEffectActor) AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return s.wrapped.AuthenticateGetOutbox(c, w, r)
}

func (s *sideEffectActor) WrapInCreate(c context.Context, value vocab.Type, outboxIRI *url.URL) (vocab.ActivityStreamsCreate, error) {
	return s.wrapped.WrapInCreate(c, value, outboxIRI)
}

func (s *sideEffectActor) GetOutbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	return s.wrapped.GetOutbox(c, r)
}

func (s *sideEffectActor) GetInbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	return s.wrapped.GetInbox(c, r)
}
