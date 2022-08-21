/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package httpclient

import (
	"strings"
	"sync"

	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type requestQueue struct {
	hostQueues   sync.Map // map of `hostQueue`
	maxOpenConns int      // max open conns per host per request method
}

type hostQueue struct {
	slotsByMethod sync.Map
}

// getWaitSpot returns a wait channel and done function for http clients
// that want to do requests politely: that is, wait for their turn.
//
// To wait, a caller should do a select on an attempted insert into the
// returned wait channel. Once the insert succeeds, then the caller should
// proceed with the http request that pertains to the given host + method.
// It doesn't matter what's put into the wait channel, just any interface{}.
//
// When the caller is done with their http request, they should free up the
// slot they were occupying in the wait queue, by calling the done function.
//
// The reason for the caller needing to provide host and method, is that each
// remote host has a separate wait queue, and there's a separate wait queue
// per method for that host as well. This ensures that outgoing requests can still
// proceed for others hosts and methods while other requests are undergoing,
// while also preventing one host from being spammed with, for example, a
// shitload of GET requests all at once.
func (rc *requestQueue) getWaitSpot(host string, method string) (wait chan<- interface{}, done func()) {
	hostQueueI, _ := rc.hostQueues.LoadOrStore(host, new(hostQueue))
	hostQueue, ok := hostQueueI.(*hostQueue)
	if !ok {
		log.Panic("hostQueueI was not a *hostQueue")
	}

	waitSlotI, _ := hostQueue.slotsByMethod.LoadOrStore(strings.ToUpper(method), make(chan interface{}, rc.maxOpenConns))
	methodQueue, ok := waitSlotI.(chan interface{})
	if !ok {
		log.Panic("waitSlotI was not a chan interface{}")
	}

	return methodQueue, func() { <-methodQueue }
}
