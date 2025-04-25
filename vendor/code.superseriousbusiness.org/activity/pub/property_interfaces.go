package pub

import (
	"net/url"

	"code.superseriousbusiness.org/activity/streams/vocab"
)

// inReplyToer is an ActivityStreams type with an 'inReplyTo' property
type inReplyToer interface {
	GetActivityStreamsInReplyTo() vocab.ActivityStreamsInReplyToProperty
}

// objecter is an ActivityStreams type with an 'object' property
type objecter interface {
	GetActivityStreamsObject() vocab.ActivityStreamsObjectProperty
}

// targeter is an ActivityStreams type with a 'target' property
type targeter interface {
	GetActivityStreamsTarget() vocab.ActivityStreamsTargetProperty
}

// tagger is an ActivityStreams type with a 'tag' property
type tagger interface {
	GetActivityStreamsTag() vocab.ActivityStreamsTagProperty
}

// hrefer is an ActivityStreams type with a 'href' property
type hrefer interface {
	GetActivityStreamsHref() vocab.ActivityStreamsHrefProperty
}

// itemser is an ActivityStreams type with an 'items' property
type itemser interface {
	GetActivityStreamsItems() vocab.ActivityStreamsItemsProperty
	SetActivityStreamsItems(vocab.ActivityStreamsItemsProperty)
}

// orderedItemser is an ActivityStreams type with an 'orderedItems' property
type orderedItemser interface {
	GetActivityStreamsOrderedItems() vocab.ActivityStreamsOrderedItemsProperty
	SetActivityStreamsOrderedItems(vocab.ActivityStreamsOrderedItemsProperty)
}

// publisheder is an ActivityStreams type with a 'published' property
type publisheder interface {
	GetActivityStreamsPublished() vocab.ActivityStreamsPublishedProperty
}

// updateder is an ActivityStreams type with an 'updateder' property
type updateder interface {
	GetActivityStreamsUpdated() vocab.ActivityStreamsUpdatedProperty
}

// toer is an ActivityStreams type with a 'to' property
type toer interface {
	GetActivityStreamsTo() vocab.ActivityStreamsToProperty
	SetActivityStreamsTo(i vocab.ActivityStreamsToProperty)
}

// btoer is an ActivityStreams type with a 'bto' property
type btoer interface {
	GetActivityStreamsBto() vocab.ActivityStreamsBtoProperty
	SetActivityStreamsBto(i vocab.ActivityStreamsBtoProperty)
}

// ccer is an ActivityStreams type with a 'cc' property
type ccer interface {
	GetActivityStreamsCc() vocab.ActivityStreamsCcProperty
	SetActivityStreamsCc(i vocab.ActivityStreamsCcProperty)
}

// bccer is an ActivityStreams type with a 'bcc' property
type bccer interface {
	GetActivityStreamsBcc() vocab.ActivityStreamsBccProperty
	SetActivityStreamsBcc(i vocab.ActivityStreamsBccProperty)
}

// audiencer is an ActivityStreams type with an 'audience' property
type audiencer interface {
	GetActivityStreamsAudience() vocab.ActivityStreamsAudienceProperty
	SetActivityStreamsAudience(i vocab.ActivityStreamsAudienceProperty)
}

// inboxer is an ActivityStreams type with an 'inbox' property
type inboxer interface {
	GetActivityStreamsInbox() vocab.ActivityStreamsInboxProperty
}

// attributedToer is an ActivityStreams type with an 'attributedTo' property
type attributedToer interface {
	GetActivityStreamsAttributedTo() vocab.ActivityStreamsAttributedToProperty
	SetActivityStreamsAttributedTo(i vocab.ActivityStreamsAttributedToProperty)
}

// likeser is an ActivityStreams type with a 'likes' property
type likeser interface {
	GetActivityStreamsLikes() vocab.ActivityStreamsLikesProperty
	SetActivityStreamsLikes(i vocab.ActivityStreamsLikesProperty)
}

// shareser is an ActivityStreams type with a 'shares' property
type shareser interface {
	GetActivityStreamsShares() vocab.ActivityStreamsSharesProperty
	SetActivityStreamsShares(i vocab.ActivityStreamsSharesProperty)
}

// actorer is an ActivityStreams type with an 'actor' property
type actorer interface {
	GetActivityStreamsActor() vocab.ActivityStreamsActorProperty
	SetActivityStreamsActor(i vocab.ActivityStreamsActorProperty)
}

// appendIRIer is an ActivityStreams type that can Append IRIs.
type appendIRIer interface {
	AppendIRI(v *url.URL)
}
