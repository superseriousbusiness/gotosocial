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

package db

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-pg/pg/v10"
)

type postgresFederation struct {
	locks *sync.Map
	conn  *pg.DB
}

func newPostgresFederation(conn *pg.DB) pub.Database {
	return &postgresFederation{
		locks: new(sync.Map),
		conn:  conn,
	}
}

/*
   GO-FED DB INTERFACE-IMPLEMENTING FUNCTIONS
*/
func (pf *postgresFederation) Lock(ctx context.Context, id *url.URL) error {
	// Before any other Database methods are called, the relevant `id`
	// entries are locked to allow for fine-grained concurrency.

	// Strategy: create a new lock, if stored, continue. Otherwise, lock the
	// existing mutex.
	mu := &sync.Mutex{}
	mu.Lock() // Optimistically lock if we do store it.
	i, loaded := pf.locks.LoadOrStore(id.String(), mu)
	if loaded {
		mu = i.(*sync.Mutex)
		mu.Lock()
	}
	return nil
}

func (pf *postgresFederation) Unlock(ctx context.Context, id *url.URL) error {
	// Once Go-Fed is done calling Database methods, the relevant `id`
	// entries are unlocked.

	i, ok := pf.locks.Load(id.String())
	if !ok {
		return errors.New("missing an id in unlock")
	}
	mu := i.(*sync.Mutex)
	mu.Unlock()
	return nil
}

func (pf *postgresFederation) InboxContains(ctx context.Context, inbox *url.URL, id *url.URL) (bool, error) {
	return false, nil
}

func (pf *postgresFederation) GetInbox(ctx context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	return nil, nil
}

func (pf *postgresFederation) SetInbox(ctx context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	return nil
}

func (pf *postgresFederation) Owns(ctx context.Context, id *url.URL) (owns bool, err error) {
	return false, nil
}

func (pf *postgresFederation) ActorForOutbox(ctx context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return nil, nil
}

func (pf *postgresFederation) ActorForInbox(ctx context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return nil, nil
}

func (pf *postgresFederation) OutboxForInbox(ctx context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	return nil, nil
}

func (pf *postgresFederation) Exists(ctx context.Context, id *url.URL) (exists bool, err error) {
	return false, nil
}

func (pf *postgresFederation) Get(ctx context.Context, id *url.URL) (value vocab.Type, err error) {
	return nil, nil
}

func (pf *postgresFederation) Create(ctx context.Context, asType vocab.Type) error {
	t, err := streams.NewTypeResolver()
	if err != nil {
		return err
	}
	if err := t.Resolve(ctx, asType); err != nil {
		return err
	}
	asType.GetTypeName()
	return nil
}

func (pf *postgresFederation) Update(ctx context.Context, asType vocab.Type) error {
	return nil
}

func (pf *postgresFederation) Delete(ctx context.Context, id *url.URL) error {
	return nil
}

func (pf *postgresFederation) GetOutbox(ctx context.Context, outboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	return nil, nil
}

func (pf *postgresFederation) SetOutbox(ctx context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error {
	return nil
}

func (pf *postgresFederation) NewID(ctx context.Context, t vocab.Type) (id *url.URL, err error) {
	return nil, nil
}

func (pf *postgresFederation) Followers(ctx context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}

func (pf *postgresFederation) Following(ctx context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}

func (pf *postgresFederation) Liked(ctx context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}
