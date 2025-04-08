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

package timeline

import (
	"sync"
	"sync/atomic"

	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// preloader provides a means of synchronising the
// initial fill, or "preload", of a timeline cache.
// it has 4 possible states in the atomic pointer:
// - preloading    = &(interface{}(*sync.WaitGroup))
// - preloaded     = &(interface{}(nil))
// - needs preload = &(interface{}(false))
// - brand-new     = nil (functionally same as 'needs preload')
type preloader struct{ p atomic.Pointer[any] }

// Check will concurrency-safely check the preload
// state, and if needed call the provided function.
// if a preload is in progress, it will wait until complete.
func (p *preloader) Check(preload func()) {
	for {
		// Get state ptr.
		ptr := p.p.Load()

		if ptr == nil || *ptr == false {
			// Needs preloading, start it.
			ok := p.start(ptr, preload)

			if !ok {
				// Failed to acquire start,
				// other thread beat us to it.
				continue
			}

			// Success!
			return
		}

		// Check for a preload currently in progress.
		if wg, _ := (*ptr).(*sync.WaitGroup); wg != nil {
			wg.Wait()
			continue
		}

		// Anything else
		// means success.
		return
	}
}

// start attempts to start the given preload function, by
// performing a CAS operation with 'old'. return is success.
func (p *preloader) start(old *any, preload func()) bool {

	// Optimistically setup a
	// new waitgroup to set as
	// the preload waiter.
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Done()

	// Wrap waitgroup in
	// 'any' for pointer.
	new := any(&wg)

	// Attempt CAS operation to claim start.
	started := p.p.CompareAndSwap(old, &new)
	if !started {
		return false
	}

	// Start.
	preload()
	return true
}

// done marks state as preloaded,
// i.e. no more preload required.
func (p *preloader) done() {
	old := p.p.Swap(new(any))
	if old == nil { // was brand-new
		return
	}
	switch t := (*old).(type) {
	case *sync.WaitGroup: // was preloading
	default:
		log.Errorf(nil, "BUG: invalid preloader state: %#v", t)
	}
}

// clear will clear the state, marking a "preload" as required.
// i.e. next call to Check() will call provided preload func.
func (p *preloader) clear() {
	b := false
	a := any(b)
	for {
		old := p.p.Swap(&a)
		if old == nil { // was brand-new
			return
		}
		switch t := (*old).(type) {
		case nil: // was preloaded
			return
		case bool: // was cleared
			return
		case *sync.WaitGroup: // was preloading
			t.Wait()
		}
	}
}
