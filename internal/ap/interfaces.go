/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package ap

import "github.com/superseriousbusiness/activity/streams/vocab"

// Accountable represents the minimum activitypub interface for representing an 'account'.
// This interface is fulfilled by: Person, Application, Organization, Service, and Group
type Accountable interface {
	WithJSONLDId
	WithTypeName

	WithPreferredUsername
	WithIcon
	WithName
	WithImage
	WithSummary
	WithDiscoverable
	WithURL
	WithPublicKey
	WithInbox
	WithOutbox
	WithFollowing
	WithFollowers
	WithFeatured
	WithManuallyApprovesFollowers
	WithEndpoints
	WithTag
}

// Statusable represents the minimum activitypub interface for representing a 'status'.
// This interface is fulfilled by: Article, Document, Image, Video, Note, Page, Event, Place, Mention, Profile
type Statusable interface {
	WithJSONLDId
	WithTypeName

	WithSummary
	WithInReplyTo
	WithPublished
	WithURL
	WithAttributedTo
	WithTo
	WithCC
	WithSensitive
	WithConversation
	WithContent
	WithAttachment
	WithTag
	WithReplies
}

// Attachmentable represents the minimum activitypub interface for representing a 'mediaAttachment'.
// This interface is fulfilled by: Audio, Document, Image, Video
type Attachmentable interface {
	WithTypeName
	WithMediaType
	WithURL
	WithName
	WithBlurhash
}

// Hashtaggable represents the minimum activitypub interface for representing a 'hashtag' tag.
type Hashtaggable interface {
	WithTypeName
	WithHref
	WithName
}

// Emojiable represents the minimum interface for an 'emoji' tag.
type Emojiable interface {
	WithJSONLDId
	WithTypeName
	WithName
	WithUpdated
	WithIcon
}

// Mentionable represents the minimum interface for a 'mention' tag.
type Mentionable interface {
	WithName
	WithHref
}

// Followable represents the minimum interface for an activitystreams 'follow' activity.
type Followable interface {
	WithJSONLDId
	WithTypeName

	WithActor
	WithObject
}

// Likeable represents the minimum interface for an activitystreams 'like' activity.
type Likeable interface {
	WithJSONLDId
	WithTypeName

	WithActor
	WithObject
}

// Blockable represents the minimum interface for an activitystreams 'block' activity.
type Blockable interface {
	WithJSONLDId
	WithTypeName

	WithActor
	WithObject
}

// Announceable represents the minimum interface for an activitystreams 'announce' activity.
type Announceable interface {
	WithJSONLDId
	WithTypeName

	WithActor
	WithObject
	WithPublished
	WithTo
	WithCC
}

// Addressable represents the minimum interface for an addressed activity.
type Addressable interface {
	WithTo
	WithCC
}

// ReplyToable represents the minimum interface for an Activity that can be InReplyTo another activity.
type ReplyToable interface {
	WithInReplyTo
}

// CollectionPageable represents the minimum interface for an activitystreams 'CollectionPage' object.
type CollectionPageable interface {
	WithJSONLDId
	WithTypeName

	WithNext
	WithPartOf
	WithItems
}

// WithJSONLDId represents an activity with JSONLDIdProperty
type WithJSONLDId interface {
	GetJSONLDId() vocab.JSONLDIdProperty
}

// WithTypeName represents an activity with a type name
type WithTypeName interface {
	GetTypeName() string
}

// WithPreferredUsername represents an activity with ActivityStreamsPreferredUsernameProperty
type WithPreferredUsername interface {
	GetActivityStreamsPreferredUsername() vocab.ActivityStreamsPreferredUsernameProperty
}

// WithIcon represents an activity with ActivityStreamsIconProperty
type WithIcon interface {
	GetActivityStreamsIcon() vocab.ActivityStreamsIconProperty
}

// WithName represents an activity with ActivityStreamsNameProperty
type WithName interface {
	GetActivityStreamsName() vocab.ActivityStreamsNameProperty
}

// WithImage represents an activity with ActivityStreamsImageProperty
type WithImage interface {
	GetActivityStreamsImage() vocab.ActivityStreamsImageProperty
}

// WithSummary represents an activity with ActivityStreamsSummaryProperty
type WithSummary interface {
	GetActivityStreamsSummary() vocab.ActivityStreamsSummaryProperty
}

// WithDiscoverable represents an activity with TootDiscoverableProperty
type WithDiscoverable interface {
	GetTootDiscoverable() vocab.TootDiscoverableProperty
}

// WithURL represents an activity with ActivityStreamsUrlProperty
type WithURL interface {
	GetActivityStreamsUrl() vocab.ActivityStreamsUrlProperty
}

// WithPublicKey represents an activity with W3IDSecurityV1PublicKeyProperty
type WithPublicKey interface {
	GetW3IDSecurityV1PublicKey() vocab.W3IDSecurityV1PublicKeyProperty
}

// WithInbox represents an activity with ActivityStreamsInboxProperty
type WithInbox interface {
	GetActivityStreamsInbox() vocab.ActivityStreamsInboxProperty
}

// WithOutbox represents an activity with ActivityStreamsOutboxProperty
type WithOutbox interface {
	GetActivityStreamsOutbox() vocab.ActivityStreamsOutboxProperty
}

// WithFollowing represents an activity with ActivityStreamsFollowingProperty
type WithFollowing interface {
	GetActivityStreamsFollowing() vocab.ActivityStreamsFollowingProperty
}

// WithFollowers represents an activity with ActivityStreamsFollowersProperty
type WithFollowers interface {
	GetActivityStreamsFollowers() vocab.ActivityStreamsFollowersProperty
}

// WithFeatured represents an activity with TootFeaturedProperty
type WithFeatured interface {
	GetTootFeatured() vocab.TootFeaturedProperty
}

// WithAttributedTo represents an activity with ActivityStreamsAttributedToProperty
type WithAttributedTo interface {
	GetActivityStreamsAttributedTo() vocab.ActivityStreamsAttributedToProperty
}

// WithAttachment represents an activity with ActivityStreamsAttachmentProperty
type WithAttachment interface {
	GetActivityStreamsAttachment() vocab.ActivityStreamsAttachmentProperty
}

// WithTo represents an activity with ActivityStreamsToProperty
type WithTo interface {
	GetActivityStreamsTo() vocab.ActivityStreamsToProperty
}

// WithInReplyTo represents an activity with ActivityStreamsInReplyToProperty
type WithInReplyTo interface {
	GetActivityStreamsInReplyTo() vocab.ActivityStreamsInReplyToProperty
}

// WithCC represents an activity with ActivityStreamsCcProperty
type WithCC interface {
	GetActivityStreamsCc() vocab.ActivityStreamsCcProperty
}

// WithSensitive represents an activity with ActivityStreamsSensitiveProperty
type WithSensitive interface {
	GetActivityStreamsSensitive() vocab.ActivityStreamsSensitiveProperty
}

// WithConversation ...
type WithConversation interface { // TODO
}

// WithContent represents an activity with ActivityStreamsContentProperty
type WithContent interface {
	GetActivityStreamsContent() vocab.ActivityStreamsContentProperty
}

// WithPublished represents an activity with ActivityStreamsPublishedProperty
type WithPublished interface {
	GetActivityStreamsPublished() vocab.ActivityStreamsPublishedProperty
}

// WithTag represents an activity with ActivityStreamsTagProperty
type WithTag interface {
	GetActivityStreamsTag() vocab.ActivityStreamsTagProperty
}

// WithReplies represents an activity with ActivityStreamsRepliesProperty
type WithReplies interface {
	GetActivityStreamsReplies() vocab.ActivityStreamsRepliesProperty
}

// WithMediaType represents an activity with ActivityStreamsMediaTypeProperty
type WithMediaType interface {
	GetActivityStreamsMediaType() vocab.ActivityStreamsMediaTypeProperty
}

// WithBlurhash represents an activity with TootBlurhashProperty
type WithBlurhash interface {
	GetTootBlurhash() vocab.TootBlurhashProperty
}

// type withFocalPoint interface {
// 	// TODO
// }

// WithHref represents an activity with ActivityStreamsHrefProperty
type WithHref interface {
	GetActivityStreamsHref() vocab.ActivityStreamsHrefProperty
}

// WithUpdated represents an activity with ActivityStreamsUpdatedProperty
type WithUpdated interface {
	GetActivityStreamsUpdated() vocab.ActivityStreamsUpdatedProperty
}

// WithActor represents an activity with ActivityStreamsActorProperty
type WithActor interface {
	GetActivityStreamsActor() vocab.ActivityStreamsActorProperty
}

// WithObject represents an activity with ActivityStreamsObjectProperty
type WithObject interface {
	GetActivityStreamsObject() vocab.ActivityStreamsObjectProperty
}

// WithNext represents an activity with ActivityStreamsNextProperty
type WithNext interface {
	GetActivityStreamsNext() vocab.ActivityStreamsNextProperty
}

// WithPartOf represents an activity with ActivityStreamsPartOfProperty
type WithPartOf interface {
	GetActivityStreamsPartOf() vocab.ActivityStreamsPartOfProperty
}

// WithItems represents an activity with ActivityStreamsItemsProperty
type WithItems interface {
	GetActivityStreamsItems() vocab.ActivityStreamsItemsProperty
}

// WithManuallyApprovesFollowers represents a Person or profile with the ManuallyApprovesFollowers property.
type WithManuallyApprovesFollowers interface {
	GetActivityStreamsManuallyApprovesFollowers() vocab.ActivityStreamsManuallyApprovesFollowersProperty
}

// WithEndpoints represents a Person or profile with the endpoints property
type WithEndpoints interface {
	GetActivityStreamsEndpoints() vocab.ActivityStreamsEndpointsProperty
}
