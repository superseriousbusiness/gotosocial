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

package model

// ActivityStreamsObject refers to https://www.w3.org/TR/activitystreams-vocabulary/#object-types
type ActivityStreamsObject string

const (
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-article
	ActivityStreamsArticle ActivityStreamsObject = "Article"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-audio
	ActivityStreamsAudio ActivityStreamsObject = "Audio"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-document
	ActivityStreamsDocument ActivityStreamsObject = "Event"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-event
	ActivityStreamsEvent ActivityStreamsObject = "Event"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-image
	ActivityStreamsImage ActivityStreamsObject = "Image"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-note
	ActivityStreamsNote ActivityStreamsObject = "Note"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-page
	ActivityStreamsPage ActivityStreamsObject = "Page"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-place
	ActivityStreamsPlace ActivityStreamsObject = "Place"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-profile
	ActivityStreamsProfile ActivityStreamsObject = "Profile"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-relationship
	ActivityStreamsRelationship ActivityStreamsObject = "Relationship"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-tombstone
	ActivityStreamsTombstone ActivityStreamsObject = "Tombstone"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-video
	ActivityStreamsVideo ActivityStreamsObject = "Video"
)

// ActivityStreamsActor refers to https://www.w3.org/TR/activitystreams-vocabulary/#actor-types
type ActivityStreamsActor string

const (
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-application
	ActivityStreamsApplication ActivityStreamsActor = "Application"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-group
	ActivityStreamsGroup ActivityStreamsActor = "Group"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-organization
	ActivityStreamsOrganization ActivityStreamsActor = "Organization"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-person
	ActivityStreamsPerson ActivityStreamsActor = "Person"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-service
	ActivityStreamsService ActivityStreamsActor = "Service"
)

// ActivityStreamsActivity refers to https://www.w3.org/TR/activitystreams-vocabulary/#activity-types
type ActivityStreamsActivity string

const (
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-accept
	ActivityStreamsAccept ActivityStreamsActivity = "Accept"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-add
	ActivityStreamsAdd ActivityStreamsActivity = "Add"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-announce
	ActivityStreamsAnnounce ActivityStreamsActivity = "Announce"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-arrive
	ActivityStreamsArrive ActivityStreamsActivity = "Arrive"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-block
	ActivityStreamsBlock ActivityStreamsActivity = "Block"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-create
	ActivityStreamsCreate ActivityStreamsActivity = "Create"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-delete
	ActivityStreamsDelete ActivityStreamsActivity = "Delete"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-dislike
	ActivityStreamsDislike ActivityStreamsActivity = "Dislike"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-flag
	ActivityStreamsFlag ActivityStreamsActivity = "Flag"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-follow
	ActivityStreamsFollow ActivityStreamsActivity = "Follow"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-ignore
	ActivityStreamsIgnore ActivityStreamsActivity = "Ignore"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-invite
	ActivityStreamsInvite ActivityStreamsActivity = "Invite"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-join
	ActivityStreamsJoin ActivityStreamsActivity = "Join"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-leave
	ActivityStreamsLeave ActivityStreamsActivity = "Leave"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-like
	ActivityStreamsLike ActivityStreamsActivity = "Like"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-listen
	ActivityStreamsListen ActivityStreamsActivity = "Listen"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-move
	ActivityStreamsMove ActivityStreamsActivity = "Move"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-offer
	ActivityStreamsOffer ActivityStreamsActivity = "Offer"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-question
	ActivityStreamsQuestion ActivityStreamsActivity = "Question"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-reject
	ActivityStreamsReject ActivityStreamsActivity = "Reject"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-read
	ActivityStreamsRead ActivityStreamsActivity = "Read"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-remove
	ActivityStreamsRemove ActivityStreamsActivity = "Remove"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-tentativereject
	ActivityStreamsTentativeReject ActivityStreamsActivity = "TentativeReject"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-tentativeaccept
	ActivityStreamsTentativeAccept ActivityStreamsActivity = "TentativeAccept"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-travel
	ActivityStreamsTravel ActivityStreamsActivity = "Travel"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-undo
	ActivityStreamsUndo ActivityStreamsActivity = "Undo"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-update
	ActivityStreamsUpdate ActivityStreamsActivity = "Update"
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-view
	ActivityStreamsView ActivityStreamsActivity = "View"
)
