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

package gtsmodel

// ActivityStreamsObject refers to https://www.w3.org/TR/activitystreams-vocabulary/#object-types
type ActivityStreamsObject string

const (
	// ActivityStreamsArticle https://www.w3.org/TR/activitystreams-vocabulary/#dfn-article
	ActivityStreamsArticle ActivityStreamsObject = "Article"
	// ActivityStreamsAudio https://www.w3.org/TR/activitystreams-vocabulary/#dfn-audio
	ActivityStreamsAudio ActivityStreamsObject = "Audio"
	// ActivityStreamsDocument https://www.w3.org/TR/activitystreams-vocabulary/#dfn-document
	ActivityStreamsDocument ActivityStreamsObject = "Event"
	// ActivityStreamsEvent https://www.w3.org/TR/activitystreams-vocabulary/#dfn-event
	ActivityStreamsEvent ActivityStreamsObject = "Event"
	// ActivityStreamsImage https://www.w3.org/TR/activitystreams-vocabulary/#dfn-image
	ActivityStreamsImage ActivityStreamsObject = "Image"
	// ActivityStreamsNote https://www.w3.org/TR/activitystreams-vocabulary/#dfn-note
	ActivityStreamsNote ActivityStreamsObject = "Note"
	// ActivityStreamsPage https://www.w3.org/TR/activitystreams-vocabulary/#dfn-page
	ActivityStreamsPage ActivityStreamsObject = "Page"
	// ActivityStreamsPlace https://www.w3.org/TR/activitystreams-vocabulary/#dfn-place
	ActivityStreamsPlace ActivityStreamsObject = "Place"
	// ActivityStreamsProfile https://www.w3.org/TR/activitystreams-vocabulary/#dfn-profile
	ActivityStreamsProfile ActivityStreamsObject = "Profile"
	// ActivityStreamsRelationship https://www.w3.org/TR/activitystreams-vocabulary/#dfn-relationship
	ActivityStreamsRelationship ActivityStreamsObject = "Relationship"
	// ActivityStreamsTombstone https://www.w3.org/TR/activitystreams-vocabulary/#dfn-tombstone
	ActivityStreamsTombstone ActivityStreamsObject = "Tombstone"
	// ActivityStreamsVideo https://www.w3.org/TR/activitystreams-vocabulary/#dfn-video
	ActivityStreamsVideo ActivityStreamsObject = "Video"
)

// ActivityStreamsActor refers to https://www.w3.org/TR/activitystreams-vocabulary/#actor-types
type ActivityStreamsActor string

const (
	// ActivityStreamsApplication https://www.w3.org/TR/activitystreams-vocabulary/#dfn-application
	ActivityStreamsApplication ActivityStreamsActor = "Application"
	// ActivityStreamsGroup https://www.w3.org/TR/activitystreams-vocabulary/#dfn-group
	ActivityStreamsGroup ActivityStreamsActor = "Group"
	// ActivityStreamsOrganization https://www.w3.org/TR/activitystreams-vocabulary/#dfn-organization
	ActivityStreamsOrganization ActivityStreamsActor = "Organization"
	// ActivityStreamsPerson https://www.w3.org/TR/activitystreams-vocabulary/#dfn-person
	ActivityStreamsPerson ActivityStreamsActor = "Person"
	// ActivityStreamsService https://www.w3.org/TR/activitystreams-vocabulary/#dfn-service
	ActivityStreamsService ActivityStreamsActor = "Service"
)

// ActivityStreamsActivity refers to https://www.w3.org/TR/activitystreams-vocabulary/#activity-types
type ActivityStreamsActivity string

const (
	// ActivityStreamsAccept https://www.w3.org/TR/activitystreams-vocabulary/#dfn-accept
	ActivityStreamsAccept ActivityStreamsActivity = "Accept"
	// ActivityStreamsAdd https://www.w3.org/TR/activitystreams-vocabulary/#dfn-add
	ActivityStreamsAdd ActivityStreamsActivity = "Add"
	// ActivityStreamsAnnounce https://www.w3.org/TR/activitystreams-vocabulary/#dfn-announce
	ActivityStreamsAnnounce ActivityStreamsActivity = "Announce"
	// ActivityStreamsArrive https://www.w3.org/TR/activitystreams-vocabulary/#dfn-arrive
	ActivityStreamsArrive ActivityStreamsActivity = "Arrive"
	// ActivityStreamsBlock https://www.w3.org/TR/activitystreams-vocabulary/#dfn-block
	ActivityStreamsBlock ActivityStreamsActivity = "Block"
	// ActivityStreamsCreate https://www.w3.org/TR/activitystreams-vocabulary/#dfn-create
	ActivityStreamsCreate ActivityStreamsActivity = "Create"
	// ActivityStreamsDelete https://www.w3.org/TR/activitystreams-vocabulary/#dfn-delete
	ActivityStreamsDelete ActivityStreamsActivity = "Delete"
	// ActivityStreamsDislike https://www.w3.org/TR/activitystreams-vocabulary/#dfn-dislike
	ActivityStreamsDislike ActivityStreamsActivity = "Dislike"
	// ActivityStreamsFlag https://www.w3.org/TR/activitystreams-vocabulary/#dfn-flag
	ActivityStreamsFlag ActivityStreamsActivity = "Flag"
	// ActivityStreamsFollow https://www.w3.org/TR/activitystreams-vocabulary/#dfn-follow
	ActivityStreamsFollow ActivityStreamsActivity = "Follow"
	// ActivityStreamsIgnore https://www.w3.org/TR/activitystreams-vocabulary/#dfn-ignore
	ActivityStreamsIgnore ActivityStreamsActivity = "Ignore"
	// ActivityStreamsInvite https://www.w3.org/TR/activitystreams-vocabulary/#dfn-invite
	ActivityStreamsInvite ActivityStreamsActivity = "Invite"
	// ActivityStreamsJoin https://www.w3.org/TR/activitystreams-vocabulary/#dfn-join
	ActivityStreamsJoin ActivityStreamsActivity = "Join"
	// ActivityStreamsLeave https://www.w3.org/TR/activitystreams-vocabulary/#dfn-leave
	ActivityStreamsLeave ActivityStreamsActivity = "Leave"
	// ActivityStreamsLike https://www.w3.org/TR/activitystreams-vocabulary/#dfn-like
	ActivityStreamsLike ActivityStreamsActivity = "Like"
	// ActivityStreamsListen https://www.w3.org/TR/activitystreams-vocabulary/#dfn-listen
	ActivityStreamsListen ActivityStreamsActivity = "Listen"
	// ActivityStreamsMove https://www.w3.org/TR/activitystreams-vocabulary/#dfn-move
	ActivityStreamsMove ActivityStreamsActivity = "Move"
	// ActivityStreamsOffer https://www.w3.org/TR/activitystreams-vocabulary/#dfn-offer
	ActivityStreamsOffer ActivityStreamsActivity = "Offer"
	// ActivityStreamsQuestion https://www.w3.org/TR/activitystreams-vocabulary/#dfn-question
	ActivityStreamsQuestion ActivityStreamsActivity = "Question"
	// ActivityStreamsReject https://www.w3.org/TR/activitystreams-vocabulary/#dfn-reject
	ActivityStreamsReject ActivityStreamsActivity = "Reject"
	// ActivityStreamsRead https://www.w3.org/TR/activitystreams-vocabulary/#dfn-read
	ActivityStreamsRead ActivityStreamsActivity = "Read"
	// ActivityStreamsRemove https://www.w3.org/TR/activitystreams-vocabulary/#dfn-remove
	ActivityStreamsRemove ActivityStreamsActivity = "Remove"
	// ActivityStreamsTentativeReject https://www.w3.org/TR/activitystreams-vocabulary/#dfn-tentativereject
	ActivityStreamsTentativeReject ActivityStreamsActivity = "TentativeReject"
	// ActivityStreamsTentativeAccept https://www.w3.org/TR/activitystreams-vocabulary/#dfn-tentativeaccept
	ActivityStreamsTentativeAccept ActivityStreamsActivity = "TentativeAccept"
	// ActivityStreamsTravel https://www.w3.org/TR/activitystreams-vocabulary/#dfn-travel
	ActivityStreamsTravel ActivityStreamsActivity = "Travel"
	// ActivityStreamsUndo https://www.w3.org/TR/activitystreams-vocabulary/#dfn-undo
	ActivityStreamsUndo ActivityStreamsActivity = "Undo"
	// ActivityStreamsUpdate https://www.w3.org/TR/activitystreams-vocabulary/#dfn-update
	ActivityStreamsUpdate ActivityStreamsActivity = "Update"
	// ActivityStreamsView https://www.w3.org/TR/activitystreams-vocabulary/#dfn-view
	ActivityStreamsView ActivityStreamsActivity = "View"
)
