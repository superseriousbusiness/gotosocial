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

// https://www.w3.org/TR/activitystreams-vocabulary
const (
	ActivityAccept          = "Accept"          // ActivityStreamsAccept https://www.w3.org/TR/activitystreams-vocabulary/#dfn-accept
	ActivityAdd             = "Add"             // ActivityStreamsAdd https://www.w3.org/TR/activitystreams-vocabulary/#dfn-add
	ActivityAnnounce        = "Announce"        // ActivityStreamsAnnounce https://www.w3.org/TR/activitystreams-vocabulary/#dfn-announce
	ActivityArrive          = "Arrive"          // ActivityStreamsArrive https://www.w3.org/TR/activitystreams-vocabulary/#dfn-arrive
	ActivityBlock           = "Block"           // ActivityStreamsBlock https://www.w3.org/TR/activitystreams-vocabulary/#dfn-block
	ActivityCreate          = "Create"          // ActivityStreamsCreate https://www.w3.org/TR/activitystreams-vocabulary/#dfn-create
	ActivityDelete          = "Delete"          // ActivityStreamsDelete https://www.w3.org/TR/activitystreams-vocabulary/#dfn-delete
	ActivityDislike         = "Dislike"         // ActivityStreamsDislike https://www.w3.org/TR/activitystreams-vocabulary/#dfn-dislike
	ActivityFlag            = "Flag"            // ActivityStreamsFlag https://www.w3.org/TR/activitystreams-vocabulary/#dfn-flag
	ActivityFollow          = "Follow"          // ActivityStreamsFollow https://www.w3.org/TR/activitystreams-vocabulary/#dfn-follow
	ActivityIgnore          = "Ignore"          // ActivityStreamsIgnore https://www.w3.org/TR/activitystreams-vocabulary/#dfn-ignore
	ActivityInvite          = "Invite"          // ActivityStreamsInvite https://www.w3.org/TR/activitystreams-vocabulary/#dfn-invite
	ActivityJoin            = "Join"            // ActivityStreamsJoin https://www.w3.org/TR/activitystreams-vocabulary/#dfn-join
	ActivityLeave           = "Leave"           // ActivityStreamsLeave https://www.w3.org/TR/activitystreams-vocabulary/#dfn-leave
	ActivityLike            = "Like"            // ActivityStreamsLike https://www.w3.org/TR/activitystreams-vocabulary/#dfn-like
	ActivityListen          = "Listen"          // ActivityStreamsListen https://www.w3.org/TR/activitystreams-vocabulary/#dfn-listen
	ActivityMove            = "Move"            // ActivityStreamsMove https://www.w3.org/TR/activitystreams-vocabulary/#dfn-move
	ActivityOffer           = "Offer"           // ActivityStreamsOffer https://www.w3.org/TR/activitystreams-vocabulary/#dfn-offer
	ActivityQuestion        = "Question"        // ActivityStreamsQuestion https://www.w3.org/TR/activitystreams-vocabulary/#dfn-question
	ActivityReject          = "Reject"          // ActivityStreamsReject https://www.w3.org/TR/activitystreams-vocabulary/#dfn-reject
	ActivityRead            = "Read"            // ActivityStreamsRead https://www.w3.org/TR/activitystreams-vocabulary/#dfn-read
	ActivityRemove          = "Remove"          // ActivityStreamsRemove https://www.w3.org/TR/activitystreams-vocabulary/#dfn-remove
	ActivityTentativeReject = "TentativeReject" // ActivityStreamsTentativeReject https://www.w3.org/TR/activitystreams-vocabulary/#dfn-tentativereject
	ActivityTentativeAccept = "TentativeAccept" // ActivityStreamsTentativeAccept https://www.w3.org/TR/activitystreams-vocabulary/#dfn-tentativeaccept
	ActivityTravel          = "Travel"          // ActivityStreamsTravel https://www.w3.org/TR/activitystreams-vocabulary/#dfn-travel
	ActivityUndo            = "Undo"            // ActivityStreamsUndo https://www.w3.org/TR/activitystreams-vocabulary/#dfn-undo
	ActivityUpdate          = "Update"          // ActivityStreamsUpdate https://www.w3.org/TR/activitystreams-vocabulary/#dfn-update
	ActivityView            = "View"            // ActivityStreamsView https://www.w3.org/TR/activitystreams-vocabulary/#dfn-view

	ActorApplication  = "Application"  // ActivityStreamsApplication https://www.w3.org/TR/activitystreams-vocabulary/#dfn-application
	ActorGroup        = "Group"        // ActivityStreamsGroup https://www.w3.org/TR/activitystreams-vocabulary/#dfn-group
	ActorOrganization = "Organization" // ActivityStreamsOrganization https://www.w3.org/TR/activitystreams-vocabulary/#dfn-organization
	ActorPerson       = "Person"       // ActivityStreamsPerson https://www.w3.org/TR/activitystreams-vocabulary/#dfn-person
	ActorService      = "Service"      // ActivityStreamsService https://www.w3.org/TR/activitystreams-vocabulary/#dfn-service

	ObjectArticle               = "Article"               // ActivityStreamsArticle https://www.w3.org/TR/activitystreams-vocabulary/#dfn-article
	ObjectAudio                 = "Audio"                 // ActivityStreamsAudio https://www.w3.org/TR/activitystreams-vocabulary/#dfn-audio
	ObjectDocument              = "Document"              // ActivityStreamsDocument https://www.w3.org/TR/activitystreams-vocabulary/#dfn-document
	ObjectEvent                 = "Event"                 // ActivityStreamsEvent https://www.w3.org/TR/activitystreams-vocabulary/#dfn-event
	ObjectImage                 = "Image"                 // ActivityStreamsImage https://www.w3.org/TR/activitystreams-vocabulary/#dfn-image
	ObjectNote                  = "Note"                  // ActivityStreamsNote https://www.w3.org/TR/activitystreams-vocabulary/#dfn-note
	ObjectPage                  = "Page"                  // ActivityStreamsPage https://www.w3.org/TR/activitystreams-vocabulary/#dfn-page
	ObjectPlace                 = "Place"                 // ActivityStreamsPlace https://www.w3.org/TR/activitystreams-vocabulary/#dfn-place
	ObjectProfile               = "Profile"               // ActivityStreamsProfile https://www.w3.org/TR/activitystreams-vocabulary/#dfn-profile
	ObjectRelationship          = "Relationship"          // ActivityStreamsRelationship https://www.w3.org/TR/activitystreams-vocabulary/#dfn-relationship
	ObjectTombstone             = "Tombstone"             // ActivityStreamsTombstone https://www.w3.org/TR/activitystreams-vocabulary/#dfn-tombstone
	ObjectVideo                 = "Video"                 // ActivityStreamsVideo https://www.w3.org/TR/activitystreams-vocabulary/#dfn-video
	ObjectCollection            = "Collection"            // ActivityStreamsCollection https://www.w3.org/TR/activitystreams-vocabulary/#dfn-collection
	ObjectCollectionPage        = "CollectionPage"        // ActivityStreamsCollectionPage https://www.w3.org/TR/activitystreams-vocabulary/#dfn-collectionpage
	ObjectOrderedCollection     = "OrderedCollection"     // ActivityStreamsOrderedCollection https://www.w3.org/TR/activitystreams-vocabulary/#dfn-orderedcollection
	ObjectOrderedCollectionPage = "OrderedCollectionPage" // ActivityStreamsOrderedCollectionPage https://www.w3.org/TR/activitystreams-vocabulary/#dfn-orderedcollectionPage

	// Hashtag is not in the AS spec per se, but it tends to get used
	// as though 'Hashtag' is a named type under the Tag property.
	//
	// See https://www.w3.org/TR/activitystreams-vocabulary/#microsyntaxes
	// and https://www.w3.org/TR/activitystreams-vocabulary/#dfn-tag
	TagHashtag = "Hashtag"

	// Not in the AS spec, just used internally to indicate
	// that we don't *yet* know what type of Object something is.
	ObjectUnknown = "Unknown"
)

// isActivity returns whether AS type name is of an Activity (NOT IntransitiveActivity).
func isActivity(typeName string) bool {
	switch typeName {
	case ActivityAccept,
		ActivityTentativeAccept,
		ActivityAdd,
		ActivityCreate,
		ActivityDelete,
		ActivityFollow,
		ActivityIgnore,
		ActivityJoin,
		ActivityLeave,
		ActivityLike,
		ActivityOffer,
		ActivityInvite,
		ActivityReject,
		ActivityTentativeReject,
		ActivityRemove,
		ActivityUndo,
		ActivityUpdate,
		ActivityView,
		ActivityListen,
		ActivityRead,
		ActivityMove,
		ActivityAnnounce,
		ActivityBlock,
		ActivityFlag,
		ActivityDislike:
		return true
	default:
		return false
	}
}

// isIntransitiveActivity returns whether AS type name is of an IntransitiveActivity.
func isIntransitiveActivity(typeName string) bool {
	switch typeName {
	case ActivityArrive,
		ActivityTravel,
		ActivityQuestion:
		return true
	default:
		return false
	}
}
