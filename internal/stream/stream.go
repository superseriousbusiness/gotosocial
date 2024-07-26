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

	// EventTypeFiltersChanged -- the user's filters
	// (including keywords and statuses) have changed.
	EventTypeFiltersChanged = "filters_changed"

	// EventTypeConversation -- a user
	// should be shown an updated conversation.
	EventTypeConversation = "conversation"
)

const (
	// TimelineLocal:
	// All public posts originating from this
	// server. Analogous to the local timeline.
	TimelineLocal = "public:local"

	// TimelinePublic:
	// All public posts known to the server.
	// Analogous to the federated timeline.
	TimelinePublic = "public"

	// TimelineHome:
	// Events related to the current user, such
	// as home feed updates and notifications.
	TimelineHome = "user"

	// TimelineNotifications:
	// Notifications for the current user.
	TimelineNotifications = "user:notification"

	// TimelineDirect:
	// Updates to direct conversations.
	TimelineDirect = "direct"

	// TimelineList:
	// Updates to a specific list.
	TimelineList = "list"
)

// AllStatusTimelines contains all Timelines
// that a status could conceivably be delivered
// to, useful for sending out status deletes.
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
	var deferred []func() bool

	// Acquire lock.
	s.mutex.Lock()

	// Iterate all streams stored for account.
	for _, str := range s.streams[accountID] {

		// Check whether stream supports any of our message targets.
		if stype := str.getStreamType(msg.Stream...); stype != "" {

			// Rescope var
			// to prevent
			// ptr reuse.
			stream := str

			// Use a message copy to *only*
			// include the supported stream.
			msgCopy := Message{
				Stream:  []string{stype},
				Event:   msg.Event,
				Payload: msg.Payload,
			}

			// Send message to supported stream
			// DEFERRED (i.e. OUTSIDE OF MAIN MUTEX).
			// This prevents deadlocks between each
			// msg channel and main Streams{} mutex.
			deferred = append(deferred, func() bool {
				return stream.send(ctx, msgCopy)
			})
		}
	}

	// Done with lock.
	s.mutex.Unlock()

	var ok bool

	// Execute deferred outside lock.
	for _, deferfn := range deferred {
		v := deferfn()
		ok = ok && v
	}

	return ok
}

// PostAll will post the given message to all streams with matching types.
func (s *Streams) PostAll(ctx context.Context, msg Message) bool {
	var deferred []func() bool

	// Acquire lock.
	s.mutex.Lock()

	// Iterate ALL stored streams.
	for _, strs := range s.streams {
		for _, str := range strs {

			// Check whether stream supports any of our message targets.
			if stype := str.getStreamType(msg.Stream...); stype != "" {

				// Rescope var
				// to prevent
				// ptr reuse.
				stream := str

				// Use a message copy to *only*
				// include the supported stream.
				msgCopy := Message{
					Stream:  []string{stype},
					Event:   msg.Event,
					Payload: msg.Payload,
				}

				// Send message to supported stream
				// DEFERRED (i.e. OUTSIDE OF MAIN MUTEX).
				// This prevents deadlocks between each
				// msg channel and main Streams{} mutex.
				deferred = append(deferred, func() bool {
					return stream.send(ctx, msgCopy)
				})
			}
		}
	}

	// Done with lock.
	s.mutex.Unlock()

	var ok bool

	// Execute deferred outside lock.
	for _, deferfn := range deferred {
		v := deferfn()
		ok = ok && v
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

// getStreamType returns the first stream type in given list that stream supports.
func (s *Stream) getStreamType(streamTypes ...string) string {
	if ptr := s.types.Load(); ptr != nil {
		for _, streamType := range streamTypes {
			if _, ok := (*ptr)[streamType]; ok {
				return streamType
			}
		}
	}
	return ""
}

// send will block on posting a new Message{}, returning early with
// a false value if provided context is canceled, or stream closed.
func (s *Stream) send(ctx context.Context, msg Message) bool {
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
