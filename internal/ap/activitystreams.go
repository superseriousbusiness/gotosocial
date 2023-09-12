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
	"strconv"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

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

	ObjectArticle           = "Article"           // ActivityStreamsArticle https://www.w3.org/TR/activitystreams-vocabulary/#dfn-article
	ObjectAudio             = "Audio"             // ActivityStreamsAudio https://www.w3.org/TR/activitystreams-vocabulary/#dfn-audio
	ObjectDocument          = "Document"          // ActivityStreamsDocument https://www.w3.org/TR/activitystreams-vocabulary/#dfn-document
	ObjectEvent             = "Event"             // ActivityStreamsEvent https://www.w3.org/TR/activitystreams-vocabulary/#dfn-event
	ObjectImage             = "Image"             // ActivityStreamsImage https://www.w3.org/TR/activitystreams-vocabulary/#dfn-image
	ObjectNote              = "Note"              // ActivityStreamsNote https://www.w3.org/TR/activitystreams-vocabulary/#dfn-note
	ObjectPage              = "Page"              // ActivityStreamsPage https://www.w3.org/TR/activitystreams-vocabulary/#dfn-page
	ObjectPlace             = "Place"             // ActivityStreamsPlace https://www.w3.org/TR/activitystreams-vocabulary/#dfn-place
	ObjectProfile           = "Profile"           // ActivityStreamsProfile https://www.w3.org/TR/activitystreams-vocabulary/#dfn-profile
	ObjectRelationship      = "Relationship"      // ActivityStreamsRelationship https://www.w3.org/TR/activitystreams-vocabulary/#dfn-relationship
	ObjectTombstone         = "Tombstone"         // ActivityStreamsTombstone https://www.w3.org/TR/activitystreams-vocabulary/#dfn-tombstone
	ObjectVideo             = "Video"             // ActivityStreamsVideo https://www.w3.org/TR/activitystreams-vocabulary/#dfn-video
	ObjectCollection        = "Collection"        // ActivityStreamsCollection https://www.w3.org/TR/activitystreams-vocabulary/#dfn-collection
	ObjectCollectionPage    = "CollectionPage"    // ActivityStreamsCollectionPage https://www.w3.org/TR/activitystreams-vocabulary/#dfn-collectionpage
	ObjectOrderedCollection = "OrderedCollection" // ActivityStreamsOrderedCollection https://www.w3.org/TR/activitystreams-vocabulary/#dfn-orderedcollection

	// Hashtag is not in the AS spec per se, but it tends to get used
	// as though 'Hashtag' is a named type under the Tag property.
	//
	// See https://www.w3.org/TR/activitystreams-vocabulary/#microsyntaxes
	// and https://www.w3.org/TR/activitystreams-vocabulary/#dfn-tag
	TagHashtag = "Hashtag"
)

type CollectionParams struct {
	// Containing collection
	// ID (i.e. NOT the page).
	ID *url.URL

	// Total no. items.
	Total int
}

type CollectionPageParams struct {
	// containing collection.
	CollectionParams

	// Paging details.
	Current *paging.Page
	Next    *paging.Page
	Prev    *paging.Page
	Query   url.Values

	// Item appender for each item at index.
	Append func(int, ItemsPropertyBuilder)
	Count  int
}

// CollectionPage is a simplified interface type
// that can be fulfilled by either of (where required):
// vocab.ActivityStreamsCollection
// vocab.ActivityStreamsOrderedCollection
type CollectionBuilder interface {
	SetJSONLDId(vocab.JSONLDIdProperty)
	SetActivityStreamsFirst(vocab.ActivityStreamsFirstProperty)
	SetActivityStreamsTotalItems(i vocab.ActivityStreamsTotalItemsProperty)
}

// CollectionPageBuilder is a simplified interface type
// that can be fulfilled by either of (where required):
// vocab.ActivityStreamsCollectionPage
// vocab.ActivityStreamsOrderedCollectionPage
type CollectionPageBuilder interface {
	SetJSONLDId(vocab.JSONLDIdProperty)
	SetActivityStreamsPartOf(vocab.ActivityStreamsPartOfProperty)
	SetActivityStreamsNext(vocab.ActivityStreamsNextProperty)
	SetActivityStreamsPrev(vocab.ActivityStreamsPrevProperty)
	SetActivityStreamsTotalItems(i vocab.ActivityStreamsTotalItemsProperty)
}

// ItemsPropertyBuilder is a simplified interface type
// that can be fulfilled by either of (where required):
// vocab.ActivityStreamsItemsProperty
// vocab.ActivityStreamsOrderedItemsProperty
type ItemsPropertyBuilder interface {
	AppendIRI(*url.URL)

	// NOTE: add more of the items-property-like interface
	// functions here as you require them for building pages.
}

// NewASCollection builds and returns a new ActivityStreams Collection from given parameters.
func NewASCollection(params CollectionParams) vocab.ActivityStreamsCollection {
	collection := streams.NewActivityStreamsCollection()
	buildCollection(collection, params, 40)
	return collection
}

// NewASCollectionPage builds and returns a new ActivityStreams CollectionPage from given parameters (including item property appending function).
func NewASCollectionPage(params CollectionPageParams) vocab.ActivityStreamsCollectionPage {
	collectionPage := streams.NewActivityStreamsCollectionPage()
	itemsProp := streams.NewActivityStreamsItemsProperty()
	buildCollectionPage(collectionPage, itemsProp, collectionPage.SetActivityStreamsItems, params)
	return collectionPage
}

// NewASOrderedCollection builds and returns a new ActivityStreams OrderedCollection from given parameters.
func NewASOrderedCollection(params CollectionParams) vocab.ActivityStreamsOrderedCollection {
	collection := streams.NewActivityStreamsOrderedCollection()
	buildCollection(collection, params, 40)
	return collection
}

// NewASOrderedCollectionPage builds and returns a new ActivityStreams OrderedCollectionPage from given parameters (including item property appending function).
func NewASOrderedCollectionPage(params CollectionPageParams) vocab.ActivityStreamsOrderedCollectionPage {
	collectionPage := streams.NewActivityStreamsOrderedCollectionPage()
	itemsProp := streams.NewActivityStreamsOrderedItemsProperty()
	buildCollectionPage(collectionPage, itemsProp, collectionPage.SetActivityStreamsOrderedItems, params)
	return collectionPage
}

func buildCollection[C CollectionBuilder](collection C, params CollectionParams, pageLimit int) {
	// Add the collection ID property.
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(params.ID)
	collection.SetJSONLDId(idProp)

	// Add the collection totalItems count property.
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(params.Total)
	collection.SetActivityStreamsTotalItems(totalItems)

	// Clone the collection ID page
	// to add first page query data.
	firstIRI := new(url.URL)
	*firstIRI = *params.ID

	// Note that simply adding a limit signals to our
	// endpoint to use paging (which will start at beginning).
	limit := "limit=" + strconv.Itoa(pageLimit)
	firstIRI.RawQuery = appendQuery(firstIRI.RawQuery, limit)

	// Add the collection first IRI property.
	first := streams.NewActivityStreamsFirstProperty()
	first.SetIRI(firstIRI)
	collection.SetActivityStreamsFirst(first)
}

func buildCollectionPage[C CollectionPageBuilder, I ItemsPropertyBuilder](collectionPage C, itemsProp I, setItems func(I), params CollectionPageParams) {
	// Add the partOf property for its containing collection ID.
	partOfProp := streams.NewActivityStreamsPartOfProperty()
	partOfProp.SetIRI(params.ID)
	collectionPage.SetActivityStreamsPartOf(partOfProp)

	// Build the current page link IRI.
	currentIRI := params.Current.ToLinkURL(
		params.ID.Scheme,
		params.ID.Host,
		params.ID.Path,
		params.Query,
	)

	// Add the collection ID property for
	// the *current* collection page params.
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(currentIRI)
	collectionPage.SetJSONLDId(idProp)

	// Build the next page link IRI.
	nextIRI := params.Next.ToLinkURL(
		params.ID.Scheme,
		params.ID.Host,
		params.ID.Path,
		params.Query,
	)

	if nextIRI != nil {
		// Add the collection next property for the next page.
		nextProp := streams.NewActivityStreamsNextProperty()
		nextProp.SetIRI(nextIRI)
		collectionPage.SetActivityStreamsNext(nextProp)
	}

	// Build the prevocab.page link IRI.
	prevIRI := params.Prev.ToLinkURL(
		params.ID.Scheme,
		params.ID.Host,
		params.ID.Path,
		params.Query,
	)

	if prevIRI != nil {
		// Add the collection prevocab.property for the prevocab.page.
		prevProp := streams.NewActivityStreamsPrevProperty()
		prevProp.SetIRI(prevIRI)
		collectionPage.SetActivityStreamsPrev(prevProp)
	}

	// Add the collection totalItems count property.
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(params.Total)
	collectionPage.SetActivityStreamsTotalItems(totalItems)

	if params.Append == nil {
		// nil check outside the for loop.
		panic("nil params.Append function")
	}

	// Append each of the items to the provided
	// pre-allocated items property builder type.
	for i := 0; i < params.Count; i++ {
		params.Append(i, itemsProp)
	}

	// Set the collection
	// page items property.
	setItems(itemsProp)
}

// appendQuery appends part to an existing raw
// query with ampersand, else just returning part.
func appendQuery(raw, part string) string {
	if raw != "" {
		return raw + "&" + part
	}
	return part
}
