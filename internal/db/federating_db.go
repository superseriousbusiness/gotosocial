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
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// FederatingDB uses the underlying DB interface to implement the go-fed pub.Database interface.
// It doesn't care what the underlying implementation of the DB interface is, as long as it works.
type federatingDB struct {
	locks  *sync.Map
	db     DB
	config *config.Config
}

func newFederatingDB(db DB, config *config.Config) pub.Database {
	return &federatingDB{
		locks:  new(sync.Map),
		db:     db,
		config: config,
	}
}

/*
   GO-FED DB INTERFACE-IMPLEMENTING FUNCTIONS
*/
func (f *federatingDB) Lock(ctx context.Context, id *url.URL) error {
	// Before any other Database methods are called, the relevant `id`
	// entries are locked to allow for fine-grained concurrency.

	// Strategy: create a new lock, if stored, continue. Otherwise, lock the
	// existing mutex.
	mu := &sync.Mutex{}
	mu.Lock() // Optimistically lock if we do store it.
	i, loaded := f.locks.LoadOrStore(id.String(), mu)
	if loaded {
		mu = i.(*sync.Mutex)
		mu.Lock()
	}
	return nil
}

func (f *federatingDB) Unlock(ctx context.Context, id *url.URL) error {
	// Once Go-Fed is done calling Database methods, the relevant `id`
	// entries are unlocked.

	i, ok := f.locks.Load(id.String())
	if !ok {
		return errors.New("missing an id in unlock")
	}
	mu := i.(*sync.Mutex)
	mu.Unlock()
	return nil
}

func (f *federatingDB) InboxContains(ctx context.Context, inbox *url.URL, id *url.URL) (bool, error) {
	return false, nil
}

func (f *federatingDB) GetInbox(ctx context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	return nil, nil
}

func (f *federatingDB) SetInbox(ctx context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	return nil
}

func (f *federatingDB) Owns(ctx context.Context, id *url.URL) (owns bool, err error) {
	return id.Host == f.config.Host, nil
}

func (f *federatingDB) ActorForOutbox(ctx context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return nil, nil
}

func (f *federatingDB) ActorForInbox(ctx context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return nil, nil
}

func (f *federatingDB) OutboxForInbox(ctx context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	return nil, nil
}

func (f *federatingDB) Exists(ctx context.Context, id *url.URL) (exists bool, err error) {
	return false, nil
}

func (f *federatingDB) Get(ctx context.Context, id *url.URL) (value vocab.Type, err error) {
	return nil, nil
}

func (f *federatingDB) Create(ctx context.Context, asType vocab.Type) error {
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

func (f *federatingDB) Update(ctx context.Context, asType vocab.Type) error {
	return nil
}

func (f *federatingDB) Delete(ctx context.Context, id *url.URL) error {
	return nil
}

func (f *federatingDB) GetOutbox(ctx context.Context, outboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	return nil, nil
}

func (f *federatingDB) SetOutbox(ctx context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error {
	return nil
}

func (f *federatingDB) NewID(ctx context.Context, t vocab.Type) (id *url.URL, err error) {
	return nil, nil
}

func (f *federatingDB) Followers(ctx context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}

func (f *federatingDB) Following(ctx context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}

func (f *federatingDB) Liked(ctx context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}
