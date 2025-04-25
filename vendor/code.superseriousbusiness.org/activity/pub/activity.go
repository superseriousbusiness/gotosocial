package pub

import (
	"code.superseriousbusiness.org/activity/streams/vocab"
)

// Activity represents any ActivityStreams Activity type.
//
// The Activity types provided in the streams package implement this.
type Activity interface {
	// Activity is also a vocab.Type
	vocab.Type
	// GetActivityStreamsActor returns the "actor" property if it exists, and
	// nil otherwise.
	GetActivityStreamsActor() vocab.ActivityStreamsActorProperty
	// GetActivityStreamsAudience returns the "audience" property if it
	// exists, and nil otherwise.
	GetActivityStreamsAudience() vocab.ActivityStreamsAudienceProperty
	// GetActivityStreamsBcc returns the "bcc" property if it exists, and nil
	// otherwise.
	GetActivityStreamsBcc() vocab.ActivityStreamsBccProperty
	// GetActivityStreamsBto returns the "bto" property if it exists, and nil
	// otherwise.
	GetActivityStreamsBto() vocab.ActivityStreamsBtoProperty
	// GetActivityStreamsCc returns the "cc" property if it exists, and nil
	// otherwise.
	GetActivityStreamsCc() vocab.ActivityStreamsCcProperty
	// GetActivityStreamsTo returns the "to" property if it exists, and nil
	// otherwise.
	GetActivityStreamsTo() vocab.ActivityStreamsToProperty
	// GetActivityStreamsAttributedTo returns the "attributedTo" property if
	// it exists, and nil otherwise.
	GetActivityStreamsAttributedTo() vocab.ActivityStreamsAttributedToProperty
	// GetActivityStreamsObject returns the "object" property if it exists,
	// and nil otherwise.
	GetActivityStreamsObject() vocab.ActivityStreamsObjectProperty
	// SetActivityStreamsActor sets the "actor" property.
	SetActivityStreamsActor(i vocab.ActivityStreamsActorProperty)
	// SetActivityStreamsObject sets the "object" property.
	SetActivityStreamsObject(i vocab.ActivityStreamsObjectProperty)
	// SetActivityStreamsTo sets the "to" property.
	SetActivityStreamsTo(i vocab.ActivityStreamsToProperty)
	// SetActivityStreamsBto sets the "bto" property.
	SetActivityStreamsBto(i vocab.ActivityStreamsBtoProperty)
	// SetActivityStreamsBcc sets the "bcc" property.
	SetActivityStreamsBcc(i vocab.ActivityStreamsBccProperty)
	// SetActivityStreamsAttributedTo sets the "attributedTo" property.
	SetActivityStreamsAttributedTo(i vocab.ActivityStreamsAttributedToProperty)
}
