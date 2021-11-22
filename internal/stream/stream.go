package stream

import "sync"

const (
	// EventTypeNotification -- a user should be shown a notification
	EventTypeNotification string = "notification"
	// EventTypeUpdate -- a user should be shown an update in their timeline
	EventTypeUpdate string = "update"
	// EventTypeDelete -- something should be deleted from a user
	EventTypeDelete string = "delete"
)

const (
	// TimelineLocal -- public statuses from the LOCAL timeline.
	TimelineLocal string = "public:local"
	// TimelinePublic -- public statuses, including federated ones.
	TimelinePublic string = "public"
	// TimelineHome -- statuses for a user's Home timeline.
	TimelineHome string = "user"
	// TimelineNotifications -- notification events.
	TimelineNotifications string = "user:notification"
	// TimelineDirect -- statuses sent to a user directly.
	TimelineDirect string = "direct"
)

// AllStatusTimelines contains all Timelines that a status could conceivably be delivered to -- useful for doing deletes.
var AllStatusTimelines = []string{
	TimelineLocal,
	TimelinePublic,
	TimelineHome,
	TimelineDirect,
}

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
	// Timeline of this stream: user/public/etc
	Timeline string
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
