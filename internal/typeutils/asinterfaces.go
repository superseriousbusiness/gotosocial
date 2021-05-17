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

package typeutils

import "github.com/go-fed/activity/streams/vocab"

// Accountable represents the minimum activitypub interface for representing an 'account'.
// This interface is fulfilled by: Person, Application, Organization, Service, and Group
type Accountable interface {
	withJSONLDId
	withTypeName

	withPreferredUsername
	withIcon
	withName
	withImage
	withSummary
	withDiscoverable
	withURL
	withPublicKey
	withInbox
	withOutbox
	withFollowing
	withFollowers
	withFeatured
}

// Statusable represents the minimum activitypub interface for representing a 'status'.
// This interface is fulfilled by: Article, Document, Image, Video, Note, Page, Event, Place, Mention, Profile
type Statusable interface {
	withJSONLDId
	withTypeName

	withSummary
	withInReplyTo
	withPublished
	withURL
	withAttributedTo
	withTo
	withCC
	withSensitive
	withConversation
	withContent
	withAttachment
	withTag
	withReplies
}

// Attachmentable represents the minimum activitypub interface for representing a 'mediaAttachment'.
// This interface is fulfilled by: Audio, Document, Image, Video
type Attachmentable interface {
	withTypeName
	withMediaType
	withURL
	withName
}

// Hashtaggable represents the minimum activitypub interface for representing a 'hashtag' tag.
type Hashtaggable interface {
	withTypeName
	withHref
	withName
}

// Emojiable represents the minimum interface for an 'emoji' tag.
type Emojiable interface {
	withJSONLDId
	withTypeName
	withName
	withUpdated
	withIcon
}

// Mentionable represents the minimum interface for a 'mention' tag.
type Mentionable interface {
	withName
	withHref
}

// Followable represents the minimum interface for an activitystreams 'follow' activity.
type Followable interface {
	withJSONLDId
	withTypeName

	withActor
	withObject
}

type withJSONLDId interface {
	GetJSONLDId() vocab.JSONLDIdProperty
}

type withTypeName interface {
	GetTypeName() string
}

type withPreferredUsername interface {
	GetActivityStreamsPreferredUsername() vocab.ActivityStreamsPreferredUsernameProperty
}

type withIcon interface {
	GetActivityStreamsIcon() vocab.ActivityStreamsIconProperty
}

type withName interface {
	GetActivityStreamsName() vocab.ActivityStreamsNameProperty
}

type withImage interface {
	GetActivityStreamsImage() vocab.ActivityStreamsImageProperty
}

type withSummary interface {
	GetActivityStreamsSummary() vocab.ActivityStreamsSummaryProperty
}

type withDiscoverable interface {
	GetTootDiscoverable() vocab.TootDiscoverableProperty
}

type withURL interface {
	GetActivityStreamsUrl() vocab.ActivityStreamsUrlProperty
}

type withPublicKey interface {
	GetW3IDSecurityV1PublicKey() vocab.W3IDSecurityV1PublicKeyProperty
}

type withInbox interface {
	GetActivityStreamsInbox() vocab.ActivityStreamsInboxProperty
}

type withOutbox interface {
	GetActivityStreamsOutbox() vocab.ActivityStreamsOutboxProperty
}

type withFollowing interface {
	GetActivityStreamsFollowing() vocab.ActivityStreamsFollowingProperty
}

type withFollowers interface {
	GetActivityStreamsFollowers() vocab.ActivityStreamsFollowersProperty
}

type withFeatured interface {
	GetTootFeatured() vocab.TootFeaturedProperty
}

type withAttributedTo interface {
	GetActivityStreamsAttributedTo() vocab.ActivityStreamsAttributedToProperty
}

type withAttachment interface {
	GetActivityStreamsAttachment() vocab.ActivityStreamsAttachmentProperty
}

type withTo interface {
	GetActivityStreamsTo() vocab.ActivityStreamsToProperty
}

type withInReplyTo interface {
	GetActivityStreamsInReplyTo() vocab.ActivityStreamsInReplyToProperty
}

type withCC interface {
	GetActivityStreamsCc() vocab.ActivityStreamsCcProperty
}

type withSensitive interface {
	// TODO
}

type withConversation interface {
	// TODO
}

type withContent interface {
	GetActivityStreamsContent() vocab.ActivityStreamsContentProperty
}

type withPublished interface {
	GetActivityStreamsPublished() vocab.ActivityStreamsPublishedProperty
}

type withTag interface {
	GetActivityStreamsTag() vocab.ActivityStreamsTagProperty
}

type withReplies interface {
	GetActivityStreamsReplies() vocab.ActivityStreamsRepliesProperty
}

type withMediaType interface {
	GetActivityStreamsMediaType() vocab.ActivityStreamsMediaTypeProperty
}

// type withBlurhash interface {
// 	GetTootBlurhashProperty() vocab.TootBlurhashProperty
// }

// type withFocalPoint interface {
// 	// TODO
// }

type withHref interface {
	GetActivityStreamsHref() vocab.ActivityStreamsHrefProperty
}

type withUpdated interface {
	GetActivityStreamsUpdated() vocab.ActivityStreamsUpdatedProperty
}

type withActor interface {
	GetActivityStreamsActor() vocab.ActivityStreamsActorProperty
}

type withObject interface {
	GetActivityStreamsObject() vocab.ActivityStreamsObjectProperty
}
