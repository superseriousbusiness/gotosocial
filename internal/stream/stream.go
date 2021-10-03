package stream

import "sync"

// EventType models a type of stream event.
type EventType string

const (
	// EventTypeNotification -- a user should be shown a notification
	EventTypeNotification EventType = "notification"
	// EventTypeUpdate -- a user should be shown an update in their timeline
	EventTypeUpdate EventType = "update"
	// EventTypeDelete -- something should be deleted from a user
	EventTypeDelete EventType = "delete"
)

// StreamsForAccount is a wrapper for the multiple streams that one account can have running at the same time.
// TODO: put a limit on this
type StreamsForAccount struct {
	// The currently held streams for this account
	Streams []*Stream
	// Mutex to lock/unlock when modifying the slice of streams.
	sync.Mutex
}

// Stream represents one open stream for a client.
type Stream struct {
	// ID of this stream, generated during creation.
	ID string
	// Type of this stream: user/public/etc
	Type string
	// Channel of messages for the client to read from
	Messages chan *Message
	// Channel to close when the client drops away
	Hangup chan interface{}
	// Only put messages in the stream when Connected
	Connected bool
	// Mutex to lock/unlock when inserting messages, hanging up, changing the connected state etc.
	sync.Mutex
}

// Message represents one streamed message.
type Message struct {
	// All the stream types this message should be delivered to.
	Stream []string `json:"stream"`
	// The event type of the message (update/delete/notification etc)
	Event string `json:"event"`
	// The actual payload of the message. In case of an update or notification, this will be a JSON string.
	Payload string `json:"payload"`
}
