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

package stream

import (
	"context"
	"maps"
	"slices"
	"sync"
	"sync/atomic"
)

const (
	// EventTypeNotification -- a user
	// should be shown a notification.
	EventTypeNotification = "notification"

	// EventTypeUpdate -- a user should
	// be shown an update in their timeline.
	EventTypeUpdate = "update"

	// EventTypeDelete -- something
	// should be deleted from a user.
	EventTypeDelete = "delete"

	// EventTypeStatusUpdate -- something in the
	// user's timeline has been edited (yes this
	// is a confusing name, blame Mastodon ...).
	EventTypeStatusUpdate = "status.update"
)

const (
	// TimelineLocal -- public
	// statuses from the LOCAL timeline.
	TimelineLocal = "public:local"

	// TimelinePublic -- public
	// statuses, including federated.
	TimelinePublic = "public"

	// TimelineHome -- statuses
	// for a user's Home timeline.
	TimelineHome = "user"

	// TimelineNotifications -- notification events.
	TimelineNotifications = "user:notification"

	// TimelineDirect -- statuses
	// sent to a user directly.
	TimelineDirect = "direct"

	// TimelineList -- statuses
	// for a user's list timeline.
	TimelineList = "list"
)

// AllStatusTimelines contains all Timelines that a status could conceivably be delivered to -- useful for doing deletes.
var AllStatusTimelines = []string{
	TimelineLocal,
	TimelinePublic,
	TimelineHome,
	TimelineDirect,
	TimelineList,
}

type Streams struct {
	streams map[string][]*Stream
	mutex   sync.Mutex
}

// Open will open open a new Stream for given account ID and stream types, the given context will be passed to Stream.
func (s *Streams) Open(accountID string, streamTypes ...string) *Stream {
	if len(streamTypes) == 0 {
		panic("no stream types given")
	}

	// Prep new Stream.
	str := new(Stream)
	str.done = make(chan struct{})
	str.msgCh = make(chan Message, 50) // TODO: make configurable
	for _, streamType := range streamTypes {
		str.Subscribe(streamType)
	}

	// TODO: add configurable
	// max streams per account.

	// Acquire lock.
	s.mutex.Lock()

	if s.streams == nil {
		// Main stream-map needs allocating.
		s.streams = make(map[string][]*Stream)
	}

	// Add new stream for account.
	strs := s.streams[accountID]
	strs = append(strs, str)
	s.streams[accountID] = strs

	// Register close callback
	// to remove stream from our
	// internal map for this account.
	str.close = func() {
		s.mutex.Lock()
		strs := s.streams[accountID]
		strs = slices.DeleteFunc(strs, func(s *Stream) bool {
			return s == str // remove 'str' ptr
		})
		s.streams[accountID] = strs
		s.mutex.Unlock()
	}

	// Done with lock.
	s.mutex.Unlock()

	return str
}

// Post will post the given message to all streams of given account ID matching type.
func (s *Streams) Post(ctx context.Context, accountID string, msg Message) bool {
	// Acquire lock.
	s.mutex.Lock()

	// Look for open streams.
	strs := s.streams[accountID]

	if len(strs) == 0 {
		// No streams for
		// given account ID.
		s.mutex.Unlock()
		return true
	}

	// Create new slice of supported streams
	// which we use as a concurrency safe copy
	// of []Stream that support message type.
	support := make([]*Stream, 0, len(strs))
	for _, str := range strs {
		if str.Supports(msg.Stream...) {
			support = append(support, str)
		}
	}

	// Done with lock.
	s.mutex.Unlock()

	var ok bool

	// Send message to supported stream
	// types OUTSIDE of main Streams{} mutex
	// lock so we don't risk blocking / slow
	// access to the main Streams{} mutex.
	for _, str := range support {
		sent := str.Send(ctx, msg)
		ok = ok && sent
	}

	return ok
}

// PostAll will post the given message to all streams with matching types.
func (s *Streams) PostAll(ctx context.Context, msg Message) bool {
	// Acquire lock.
	s.mutex.Lock()

	// Create new slice of supported streams
	// which we use as a concurrency safe copy
	// of []Stream that support message type.
	support := make([]*Stream, 0, len(s.streams))
	for _, strs := range s.streams {
		for _, str := range strs {
			if str.Supports(msg.Stream...) {
				support = append(support, str)
			}
		}
	}

	// Done with lock.
	s.mutex.Unlock()

	var ok bool

	// Send message to supported stream
	// types OUTSIDE of main Streams{} mutex
	// lock so we don't risk blocking / slow
	// access to the main Streams{} mutex.
	for _, str := range support {
		sent := str.Send(ctx, msg)
		ok = ok && sent
	}

	return ok
}

// Stream represents one
// open stream for a client.
type Stream struct {

	// atomically updated ptr to a read-only copy
	// of supported stream types in a hashmap. this
	// gets updated via CAS operations in .cas().
	types atomic.Pointer[map[string]struct{}]

	// protects stream close.
	done chan struct{}

	// inbound msg ch.
	msgCh chan Message

	// close hook to remove
	// stream from Streams{}.
	close func()
}

// Supports returns whether Stream supports given any of stream types.
func (s *Stream) Supports(streamTypes ...string) bool {
	if ptr := s.types.Load(); ptr != nil {
		for _, streamType := range streamTypes {
			if _, ok := (*ptr)[streamType]; ok {
				return true
			}
		}
	}
	return false
}

// Subscribe will add given type to given types this stream supports.
func (s *Stream) Subscribe(streamType string) {
	s.cas(func(m map[string]struct{}) bool {
		if _, ok := m[streamType]; ok {
			return false
		}
		m[streamType] = struct{}{}
		return true
	})
}

// Unsubscribe will remove given type (if found) from types this stream supports.
func (s *Stream) Unsubscribe(streamType string) {
	s.cas(func(m map[string]struct{}) bool {
		if _, ok := m[streamType]; !ok {
			return false
		}
		delete(m, streamType)
		return true
	})
}

// Send will block on posting a new Message{}, returning early with
// a false value if provided context is canceled, or stream closed.
func (s *Stream) Send(ctx context.Context, msg Message) bool {
	select {
	case <-s.done:
		return false
	case <-ctx.Done():
		return false
	case s.msgCh <- msg:
		return true
	}
}

// Recv will block on receiving Message{}, returning early with a
// false value if provided context is canceled, or stream closed.
func (s *Stream) Recv(ctx context.Context) (Message, bool) {
	select {
	case <-s.done:
		return Message{}, false
	case <-ctx.Done():
		return Message{}, false
	case msg := <-s.msgCh:
		return msg, true
	}
}

// Close will close the underlying context, finally
// removing it from the parent Streams per-account-map.
func (s *Stream) Close() {
	select {
	case <-s.done:
	default:
		close(s.done)
		s.close()
	}
}

// cas will perform a Compare And Swap operation on s.types using modifier func.
func (s *Stream) cas(fn func(map[string]struct{}) bool) {
	if fn == nil {
		panic("nil function")
	}
	for {
		var m map[string]struct{}

		// Get current value.
		ptr := s.types.Load()

		if ptr == nil {
			// Allocate new types map.
			m = make(map[string]struct{})
		} else {
			// Clone r-only map.
			m = maps.Clone(*ptr)
		}

		// Apply
		// changes.
		if !fn(m) {
			return
		}

		// Attempt to Compare And Swap ptr.
		if s.types.CompareAndSwap(ptr, &m) {
			return
		}
	}
}

// Message represents
// one streamed message.
type Message struct {

	// All the stream types this
	// message should be delivered to.
	Stream []string `json:"stream"`

	// The event type of the message
	// (update/delete/notification etc)
	Event string `json:"event"`

	// The actual payload of the message. In case of an
	// update or notification, this will be a JSON string.
	Payload string `json:"payload"`
}
