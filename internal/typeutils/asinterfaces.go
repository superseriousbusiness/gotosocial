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
	withGetTypeName

	withPreferredUsername
	withIcon
	withDisplayName
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
	withGetTypeName

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

type withJSONLDId interface {
	GetJSONLDId() vocab.JSONLDIdProperty
}

type withGetTypeName interface {
	GetTypeName() string
}

type withPreferredUsername interface {
	GetActivityStreamsPreferredUsername() vocab.ActivityStreamsPreferredUsernameProperty
}

type withIcon interface {
	GetActivityStreamsIcon() vocab.ActivityStreamsIconProperty
}

type withDisplayName interface {
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
