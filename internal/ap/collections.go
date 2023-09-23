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
	"fmt"
	"net/url"
	"strconv"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// ToCollectionPageIterator attempts to resolve the given vocab type as a CollectionPage
// like object and wrap in a standardised interface in order to iterate its contents.
func ToCollectionPageIterator(t vocab.Type) (CollectionPageIterator, error) {
	switch name := t.GetTypeName(); name {
	case ObjectCollectionPage:
		t := t.(vocab.ActivityStreamsCollectionPage) //nolint:forcetypeassert
		return WrapCollectionPage(t), nil
	case ObjectOrderedCollectionPage:
		t := t.(vocab.ActivityStreamsOrderedCollectionPage) //nolint:forcetypeassert
		return WrapOrderedCollectionPage(t), nil
	default:
		return nil, fmt.Errorf("%T(%s) was not CollectionPage-like", t, name)
	}
}

// WrapCollectionPage wraps an ActivityStreamsCollectionPage in a standardised collection page interface.
func WrapCollectionPage(page vocab.ActivityStreamsCollectionPage) CollectionPageIterator {
	return &regularCollectionPageIterator{ActivityStreamsCollectionPage: page}
}

// WrapOrderedCollectionPage wraps an ActivityStreamsOrderedCollectionPage in a standardised collection page interface.
func WrapOrderedCollectionPage(page vocab.ActivityStreamsOrderedCollectionPage) CollectionPageIterator {
	return &orderedCollectionPageIterator{ActivityStreamsOrderedCollectionPage: page}
}

// regularCollectionPageIterator implements CollectionPageIterator
// for the vocab.ActivitiyStreamsCollectionPage type.
type regularCollectionPageIterator struct {
	vocab.ActivityStreamsCollectionPage
	items vocab.ActivityStreamsItemsPropertyIterator
	once  bool // only init items once
}

func (iter *regularCollectionPageIterator) NextPage() WithIRI {
	if iter.ActivityStreamsCollectionPage == nil {
		return nil
	}
	return iter.GetActivityStreamsNext()
}

func (iter *regularCollectionPageIterator) PrevPage() WithIRI {
	if iter.ActivityStreamsCollectionPage == nil {
		return nil
	}
	return iter.GetActivityStreamsPrev()
}

func (iter *regularCollectionPageIterator) NextItem() IteratorItemable {
	if !iter.initItems() {
		return nil
	}
	cur := iter.items
	iter.items = iter.items.Next()
	return cur
}

func (iter *regularCollectionPageIterator) PrevItem() IteratorItemable {
	if !iter.initItems() {
		return nil
	}
	cur := iter.items
	iter.items = iter.items.Prev()
	return cur
}

func (iter *regularCollectionPageIterator) initItems() bool {
	if iter.once {
		return (iter.items != nil)
	}
	iter.once = true
	if iter.ActivityStreamsCollectionPage == nil {
		return false // no page set
	}
	items := iter.GetActivityStreamsItems()
	if items == nil {
		return false // no items found
	}
	iter.items = items.Begin()
	return (iter.items != nil)
}

// orderedCollectionPageIterator implements CollectionPageIterator
// for the vocab.ActivitiyStreamsOrderedCollectionPage type.
type orderedCollectionPageIterator struct {
	vocab.ActivityStreamsOrderedCollectionPage
	items vocab.ActivityStreamsOrderedItemsPropertyIterator
	once  bool // only init items once
}

func (iter *orderedCollectionPageIterator) NextPage() WithIRI {
	if iter.ActivityStreamsOrderedCollectionPage == nil {
		return nil
	}
	return iter.GetActivityStreamsNext()
}

func (iter *orderedCollectionPageIterator) PrevPage() WithIRI {
	if iter.ActivityStreamsOrderedCollectionPage == nil {
		return nil
	}
	return iter.GetActivityStreamsPrev()
}

func (iter *orderedCollectionPageIterator) NextItem() IteratorItemable {
	if !iter.initItems() {
		return nil
	}
	cur := iter.items
	iter.items = iter.items.Next()
	return cur
}

func (iter *orderedCollectionPageIterator) PrevItem() IteratorItemable {
	if !iter.initItems() {
		return nil
	}
	cur := iter.items
	iter.items = iter.items.Prev()
	return cur
}

func (iter *orderedCollectionPageIterator) initItems() bool {
	if iter.once {
		return (iter.items != nil)
	}
	iter.once = true
	if iter.ActivityStreamsOrderedCollectionPage == nil {
		return false // no page set
	}
	items := iter.GetActivityStreamsOrderedItems()
	if items == nil {
		return false // no items found
	}
	iter.items = items.Begin()
	return (iter.items != nil)
}

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

	// Build the prev page link IRI.
	prevIRI := params.Prev.ToLinkURL(
		params.ID.Scheme,
		params.ID.Host,
		params.ID.Path,
		params.Query,
	)

	if prevIRI != nil {
		// Add the collection prev property for the prev page.
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
