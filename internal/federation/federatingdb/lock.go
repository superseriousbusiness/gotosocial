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

package federatingdb

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"sync/atomic"
)

// Lock takes a lock for the object at the specified id. If an error
// is returned, the lock must not have been taken.
//
// The lock must be able to succeed for an id that does not exist in
// the database. This means acquiring the lock does not guarantee the
// entry exists in the database.
//
// Locks are encouraged to be lightweight and in the Go layer, as some
// processes require tight loops acquiring and releasing locks.
//
// Used to ensure race conditions in multiple requests do not occur.
func (f *federatingDB) Lock(c context.Context, id *url.URL) error {
	// Before any other Database methods are called, the relevant `id`
	// entries are locked to allow for fine-grained concurrency.

	// Strategy: create a new lock, if stored, continue. Otherwise, lock the
	// existing mutex.
	if id == nil {
		return errors.New("Lock: id was nil")
	}
	idStr := id.String()

	// Acquire map lock
	f.mutex.Lock()

	// Get mutex, or create new
	mu, ok := f.locks[idStr]
	if !ok {
		mu = f.pool.Get().(*mutex)
		f.locks[idStr] = mu
	}

	// Unlock map, acquire mutex lock
	f.mutex.Unlock()
	mu.Lock()
	return nil
}

// Unlock makes the lock for the object at the specified id available.
// If an error is returned, the lock must have still been freed.
//
// Used to ensure race conditions in multiple requests do not occur.
func (f *federatingDB) Unlock(c context.Context, id *url.URL) error {
	// Once Go-Fed is done calling Database methods, the relevant `id`
	// entries are unlocked.
	if id == nil {
		return errors.New("Unlock: id was nil")
	}
	idStr := id.String()

	// Check map for mutex
	f.mutex.Lock()
	mu, ok := f.locks[idStr]
	f.mutex.Unlock()

	if !ok {
		return errors.New("missing an id in unlock")
	}

	// Unlock the mutex
	mu.Unlock()
	return nil
}

// mutex defines a mutex we can check the lock status of.
// this is not perfect, but it's good enough for a semi
// regular mutex cleanup routine
type mutex struct {
	mu sync.Mutex
	st uint32
}

// inUse returns if the mutex is in use
func (mu *mutex) inUse() bool {
	return atomic.LoadUint32(&mu.st) == 1
}

// Lock acquire mutex lock
func (mu *mutex) Lock() {
	mu.mu.Lock()
	atomic.StoreUint32(&mu.st, 1)
}

// Unlock releases mutex lock
func (mu *mutex) Unlock() {
	mu.mu.Unlock()
	atomic.StoreUint32(&mu.st, 0)
}
