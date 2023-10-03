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

package ap

import (
	"net/url"

	"github.com/superseriousbusiness/activity/streams/vocab"
)

// IsAccountable returns whether AS vocab type name is acceptable as Accountable.
func IsAccountable(typeName string) bool {
	switch typeName {
	case ActorPerson,
		ActorApplication,
		ActorOrganization,
		ActorService,
		ActorGroup:
		return true
	default:
		return false
	}
}

// ToAccountable safely tries to cast vocab.Type as Accountable, also checking for expected AS type names.
func ToAccountable(t vocab.Type) (Accountable, bool) {
	accountable, ok := t.(Accountable)
	if !ok || !IsAccountable(t.GetTypeName()) {
		return nil, false
	}
	return accountable, true
}

// IsStatusable returns whether AS vocab type name is acceptable as Statusable.
func IsStatusable(typeName string) bool {
	switch typeName {
	case ObjectArticle,
		ObjectDocument,
		ObjectImage,
		ObjectVideo,
		ObjectNote,
		ObjectPage,
		ObjectEvent,
		ObjectPlace,
		ObjectProfile,
		ActivityQuestion:
		return true
	default:
		return false
	}
}

// ToStatusable safely tries to cast vocab.Type as Statusable, also checking for expected  AS type names.
func ToStatusable(t vocab.Type) (Statusable, bool) {
	statusable, ok := t.(Statusable)
	if !ok || !IsStatusable(t.GetTypeName()) {
		return nil, false
	}
	return statusable, true
}

// IsPollable returns whether AS vocab type name is acceptable as Pollable.
func IsPollable(typeName string) bool {
	return typeName == ActivityQuestion
}

// ToPollable safely tries to cast vocab.Type as Pollable, also checking for expected AS type names.
func ToPollable(t vocab.Type) (Pollable, bool) {
	pollable, ok := t.(Pollable)
	if !ok || !IsPollable(t.GetTypeName()) {
		return nil, false
	}
	return pollable, true
}

// Accountable represents the minimum activitypub interface for representing an 'account'.
// (see: IsAccountable() for types implementing this, though you MUST make sure to check
// the typeName as this bare interface may be implementable by non-Accountable types).
type Accountable interface {
	vocab.Type

	WithPreferredUsername
	WithIcon
	WithName
	WithImage
	WithSummary
	WithAttachment
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
// (see: IsStatusable() for types implementing this, though you MUST make sure to check
// the typeName as this bare interface may be implementable by non-Statusable types).
type Statusable interface {
	vocab.Type

	WithSummary
	WithName
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

// Pollable represents the minimum activitypub interface for representing a 'poll' (it's a subset of a status).
// (see: IsPollable() for types implementing this, though you MUST make sure to check
// the typeName as this bare interface may be implementable by non-Pollable types).
type Pollable interface {
	WithOneOf
	WithAnyOf
	WithEndTime
	WithClosed
	WithVotersCount

	// base-interface
	Statusable
}

// PollOptionable represents the minimum activitypub interface for representing a poll 'option'.
// (see: IsPollOptionable() for types implementing this).
type PollOptionable interface {
	WithTypeName
	WithName
	WithReplies
}

// Attachmentable represents the minimum activitypub interface for representing a 'mediaAttachment'. (see: IsAttachmentable).
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

// CollectionPageIterator represents the minimum interface for interacting with a wrapped
// CollectionPage or OrderedCollectionPage in order to access both next / prev pages and items.
type CollectionPageIterator interface {
	vocab.Type

	NextPage() WithIRI
	PrevPage() WithIRI

	NextItem() IteratorItemable
	PrevItem() IteratorItemable
}

// IteratorItemable represents the minimum interface for an item in an iterator.
type IteratorItemable interface {
	WithIRI
	WithType
}

// Flaggable represents the minimum interface for an activitystreams 'Flag' activity.
type Flaggable interface {
	WithJSONLDId
	WithTypeName

	WithActor
	WithContent
	WithObject
}

// WithJSONLDId represents an activity with JSONLDIdProperty.
type WithJSONLDId interface {
	GetJSONLDId() vocab.JSONLDIdProperty
	SetJSONLDId(vocab.JSONLDIdProperty)
}

// WithIRI represents an object (possibly) representable as an IRI.
type WithIRI interface {
	GetIRI() *url.URL
	IsIRI() bool
	SetIRI(*url.URL)
}

// WithType ...
type WithType interface {
	GetType() vocab.Type
}

// WithTypeName represents an activity with a type name
type WithTypeName interface {
	GetTypeName() string
}

// WithPreferredUsername represents an activity with ActivityStreamsPreferredUsernameProperty
type WithPreferredUsername interface {
	GetActivityStreamsPreferredUsername() vocab.ActivityStreamsPreferredUsernameProperty
	SetActivityStreamsPreferredUsername(vocab.ActivityStreamsPreferredUsernameProperty)
}

// WithIcon represents an activity with ActivityStreamsIconProperty
type WithIcon interface {
	GetActivityStreamsIcon() vocab.ActivityStreamsIconProperty
	SetActivityStreamsIcon(vocab.ActivityStreamsIconProperty)
}

// WithName represents an activity with ActivityStreamsNameProperty
type WithName interface {
	GetActivityStreamsName() vocab.ActivityStreamsNameProperty
	SetActivityStreamsName(vocab.ActivityStreamsNameProperty)
}

// WithImage represents an activity with ActivityStreamsImageProperty
type WithImage interface {
	GetActivityStreamsImage() vocab.ActivityStreamsImageProperty
}

// WithSummary represents an activity with ActivityStreamsSummaryProperty
type WithSummary interface {
	GetActivityStreamsSummary() vocab.ActivityStreamsSummaryProperty
	SetActivityStreamsSummary(vocab.ActivityStreamsSummaryProperty)
}

// WithDiscoverable represents an activity with TootDiscoverableProperty
type WithDiscoverable interface {
	GetTootDiscoverable() vocab.TootDiscoverableProperty
	SetTootDiscoverable(vocab.TootDiscoverableProperty)
}

// WithURL represents an activity with ActivityStreamsUrlProperty
type WithURL interface {
	GetActivityStreamsUrl() vocab.ActivityStreamsUrlProperty
	SetActivityStreamsUrl(vocab.ActivityStreamsUrlProperty)
}

// WithPublicKey represents an activity with W3IDSecurityV1PublicKeyProperty
type WithPublicKey interface {
	GetW3IDSecurityV1PublicKey() vocab.W3IDSecurityV1PublicKeyProperty
	SetW3IDSecurityV1PublicKey(vocab.W3IDSecurityV1PublicKeyProperty)
}

// WithInbox represents an activity with ActivityStreamsInboxProperty
type WithInbox interface {
	GetActivityStreamsInbox() vocab.ActivityStreamsInboxProperty
	SetActivityStreamsInbox(vocab.ActivityStreamsInboxProperty)
}

// WithOutbox represents an activity with ActivityStreamsOutboxProperty
type WithOutbox interface {
	GetActivityStreamsOutbox() vocab.ActivityStreamsOutboxProperty
	SetActivityStreamsOutbox(vocab.ActivityStreamsOutboxProperty)
}

// WithFollowing represents an activity with ActivityStreamsFollowingProperty
type WithFollowing interface {
	GetActivityStreamsFollowing() vocab.ActivityStreamsFollowingProperty
	SetActivityStreamsFollowing(vocab.ActivityStreamsFollowingProperty)
}

// WithFollowers represents an activity with ActivityStreamsFollowersProperty
type WithFollowers interface {
	GetActivityStreamsFollowers() vocab.ActivityStreamsFollowersProperty
	SetActivityStreamsFollowers(vocab.ActivityStreamsFollowersProperty)
}

// WithFeatured represents an activity with TootFeaturedProperty
type WithFeatured interface {
	GetTootFeatured() vocab.TootFeaturedProperty
	SetTootFeatured(vocab.TootFeaturedProperty)
}

// WithAttributedTo represents an activity with ActivityStreamsAttributedToProperty
type WithAttributedTo interface {
	GetActivityStreamsAttributedTo() vocab.ActivityStreamsAttributedToProperty
	SetActivityStreamsAttributedTo(vocab.ActivityStreamsAttributedToProperty)
}

// WithAttachment represents an activity with ActivityStreamsAttachmentProperty
type WithAttachment interface {
	GetActivityStreamsAttachment() vocab.ActivityStreamsAttachmentProperty
	SetActivityStreamsAttachment(vocab.ActivityStreamsAttachmentProperty)
}

// WithTo represents an activity with ActivityStreamsToProperty
type WithTo interface {
	GetActivityStreamsTo() vocab.ActivityStreamsToProperty
	SetActivityStreamsTo(vocab.ActivityStreamsToProperty)
}

// WithInReplyTo represents an activity with ActivityStreamsInReplyToProperty
type WithInReplyTo interface {
	GetActivityStreamsInReplyTo() vocab.ActivityStreamsInReplyToProperty
	SetActivityStreamsInReplyTo(vocab.ActivityStreamsInReplyToProperty)
}

// WithCC represents an activity with ActivityStreamsCcProperty
type WithCC interface {
	GetActivityStreamsCc() vocab.ActivityStreamsCcProperty
	SetActivityStreamsCc(vocab.ActivityStreamsCcProperty)
}

// WithSensitive represents an activity with ActivityStreamsSensitiveProperty
type WithSensitive interface {
	GetActivityStreamsSensitive() vocab.ActivityStreamsSensitiveProperty
	SetActivityStreamsSensitive(vocab.ActivityStreamsSensitiveProperty)
}

// WithConversation ...
type WithConversation interface { // TODO
}

// WithContent represents an activity with ActivityStreamsContentProperty
type WithContent interface {
	GetActivityStreamsContent() vocab.ActivityStreamsContentProperty
	SetActivityStreamsContent(vocab.ActivityStreamsContentProperty)
}

// WithPublished represents an activity with ActivityStreamsPublishedProperty
type WithPublished interface {
	GetActivityStreamsPublished() vocab.ActivityStreamsPublishedProperty
	SetActivityStreamsPublished(vocab.ActivityStreamsPublishedProperty)
}

// WithTag represents an activity with ActivityStreamsTagProperty
type WithTag interface {
	GetActivityStreamsTag() vocab.ActivityStreamsTagProperty
	SetActivityStreamsTag(vocab.ActivityStreamsTagProperty)
}

// WithReplies represents an activity with ActivityStreamsRepliesProperty
type WithReplies interface {
	GetActivityStreamsReplies() vocab.ActivityStreamsRepliesProperty
	SetActivityStreamsReplies(vocab.ActivityStreamsRepliesProperty)
}

// WithMediaType represents an activity with ActivityStreamsMediaTypeProperty
type WithMediaType interface {
	GetActivityStreamsMediaType() vocab.ActivityStreamsMediaTypeProperty
	SetActivityStreamsMediaType(vocab.ActivityStreamsMediaTypeProperty)
}

// WithBlurhash represents an activity with TootBlurhashProperty
type WithBlurhash interface {
	GetTootBlurhash() vocab.TootBlurhashProperty
	SetTootBlurhash(vocab.TootBlurhashProperty)
}

// type withFocalPoint interface {
// 	// TODO
// }

// WithHref represents an activity with ActivityStreamsHrefProperty
type WithHref interface {
	GetActivityStreamsHref() vocab.ActivityStreamsHrefProperty
	SetActivityStreamsHref(vocab.ActivityStreamsHrefProperty)
}

// WithUpdated represents an activity with ActivityStreamsUpdatedProperty
type WithUpdated interface {
	GetActivityStreamsUpdated() vocab.ActivityStreamsUpdatedProperty
	SetActivityStreamsUpdated(vocab.ActivityStreamsUpdatedProperty)
}

// WithActor represents an activity with ActivityStreamsActorProperty
type WithActor interface {
	GetActivityStreamsActor() vocab.ActivityStreamsActorProperty
	SetActivityStreamsActor(vocab.ActivityStreamsActorProperty)
}

// WithObject represents an activity with ActivityStreamsObjectProperty
type WithObject interface {
	GetActivityStreamsObject() vocab.ActivityStreamsObjectProperty
	SetActivityStreamsObject(vocab.ActivityStreamsObjectProperty)
}

// WithNext represents an activity with ActivityStreamsNextProperty
type WithNext interface {
	GetActivityStreamsNext() vocab.ActivityStreamsNextProperty
	SetActivityStreamsNext(vocab.ActivityStreamsNextProperty)
}

// WithPartOf represents an activity with ActivityStreamsPartOfProperty
type WithPartOf interface {
	GetActivityStreamsPartOf() vocab.ActivityStreamsPartOfProperty
	SetActivityStreamsPartOf(vocab.ActivityStreamsPartOfProperty)
}

// WithItems represents an activity with ActivityStreamsItemsProperty
type WithItems interface {
	GetActivityStreamsItems() vocab.ActivityStreamsItemsProperty
	SetActivityStreamsItems(vocab.ActivityStreamsItemsProperty)
}

// WithManuallyApprovesFollowers represents a Person or profile with the ManuallyApprovesFollowers property.
type WithManuallyApprovesFollowers interface {
	GetActivityStreamsManuallyApprovesFollowers() vocab.ActivityStreamsManuallyApprovesFollowersProperty
	SetActivityStreamsManuallyApprovesFollowers(vocab.ActivityStreamsManuallyApprovesFollowersProperty)
}

// WithEndpoints represents a Person or profile with the endpoints property
type WithEndpoints interface {
	GetActivityStreamsEndpoints() vocab.ActivityStreamsEndpointsProperty
	SetActivityStreamsEndpoints(vocab.ActivityStreamsEndpointsProperty)
}

// WithOneOf represents an activity with the oneOf property.
type WithOneOf interface {
	GetActivityStreamsOneOf() vocab.ActivityStreamsOneOfProperty
	SetActivityStreamsOneOf(vocab.ActivityStreamsOneOfProperty)
}

// WithOneOf represents an activity with the oneOf property.
type WithAnyOf interface {
	GetActivityStreamsAnyOf() vocab.ActivityStreamsAnyOfProperty
	SetActivityStreamsAnyOf(vocab.ActivityStreamsAnyOfProperty)
}

// WithEndTime represents an activity with the endTime property.
type WithEndTime interface {
	GetActivityStreamsEndTime() vocab.ActivityStreamsEndTimeProperty
	SetActivityStreamsEndTime(vocab.ActivityStreamsEndTimeProperty)
}

// WithClosed represents an activity with the closed property.
type WithClosed interface {
	GetActivityStreamsClosed() vocab.ActivityStreamsClosedProperty
	SetActivityStreamsClosed(vocab.ActivityStreamsClosedProperty)
}

// WithVotersCount represents an activity with the votersCount property.
type WithVotersCount interface {
	GetTootVotersCount() vocab.TootVotersCountProperty
	SetTootVotersCount(vocab.TootVotersCountProperty)
}
