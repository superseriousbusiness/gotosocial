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

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// TODO: replace must of this logic with just
// using extractIRIs() on the iterator types.

// ToCollectionIterator attempts to resolve the given vocab type as a Collection
// like object and wrap in a standardised interface in order to iterate its contents.
func ToCollectionIterator(t vocab.Type) (CollectionIterator, error) {
	switch name := t.GetTypeName(); name {
	case ObjectCollection:
		t := t.(vocab.ActivityStreamsCollection)
		return WrapCollection(t), nil
	case ObjectOrderedCollection:
		t := t.(vocab.ActivityStreamsOrderedCollection)
		return WrapOrderedCollection(t), nil
	default:
		return nil, fmt.Errorf("%T(%s) was not Collection-like", t, name)
	}
}

// ToCollectionPageIterator attempts to resolve the given vocab type as a CollectionPage
// like object and wrap in a standardised interface in order to iterate its contents.
func ToCollectionPageIterator(t vocab.Type) (CollectionPageIterator, error) {
	switch name := t.GetTypeName(); name {
	case ObjectCollectionPage:
		t := t.(vocab.ActivityStreamsCollectionPage)
		return WrapCollectionPage(t), nil
	case ObjectOrderedCollectionPage:
		t := t.(vocab.ActivityStreamsOrderedCollectionPage)
		return WrapOrderedCollectionPage(t), nil
	default:
		return nil, fmt.Errorf("%T(%s) was not CollectionPage-like", t, name)
	}
}

// WrapCollection wraps an ActivityStreamsCollection in a standardised collection interface.
func WrapCollection(collection vocab.ActivityStreamsCollection) CollectionIterator {
	return &regularCollectionIterator{ActivityStreamsCollection: collection}
}

// WrapOrderedCollection wraps an ActivityStreamsOrderedCollection in a standardised collection interface.
func WrapOrderedCollection(collection vocab.ActivityStreamsOrderedCollection) CollectionIterator {
	return &orderedCollectionIterator{ActivityStreamsOrderedCollection: collection}
}

// WrapCollectionPage wraps an ActivityStreamsCollectionPage in a standardised collection page interface.
func WrapCollectionPage(page vocab.ActivityStreamsCollectionPage) CollectionPageIterator {
	return &regularCollectionPageIterator{ActivityStreamsCollectionPage: page}
}

// WrapOrderedCollectionPage wraps an ActivityStreamsOrderedCollectionPage in a standardised collection page interface.
func WrapOrderedCollectionPage(page vocab.ActivityStreamsOrderedCollectionPage) CollectionPageIterator {
	return &orderedCollectionPageIterator{ActivityStreamsOrderedCollectionPage: page}
}

// regularCollectionIterator implements CollectionIterator
// for the vocab.ActivitiyStreamsCollection type.
type regularCollectionIterator struct {
	vocab.ActivityStreamsCollection
	items vocab.ActivityStreamsItemsPropertyIterator
	once  bool // only init items once
}

func (iter *regularCollectionIterator) NextItem() TypeOrIRI {
	if !iter.initItems() {
		return nil
	}
	cur := iter.items
	iter.items = iter.items.Next()
	return cur
}

func (iter *regularCollectionIterator) PrevItem() TypeOrIRI {
	if !iter.initItems() {
		return nil
	}
	cur := iter.items
	iter.items = iter.items.Prev()
	return cur
}

func (iter *regularCollectionIterator) TotalItems() int {
	totalItems := iter.GetActivityStreamsTotalItems()
	if totalItems == nil || !totalItems.IsXMLSchemaNonNegativeInteger() {
		return -1
	}

	return totalItems.Get()
}

func (iter *regularCollectionIterator) initItems() bool {
	if iter.once {
		return (iter.items != nil)
	}
	iter.once = true
	if iter.ActivityStreamsCollection == nil {
		return false // no page set
	}
	items := iter.GetActivityStreamsItems()
	if items == nil {
		return false // no items found
	}
	iter.items = items.Begin()
	return (iter.items != nil)
}

// orderedCollectionIterator implements CollectionIterator
// for the vocab.ActivitiyStreamsOrderedCollection type.
type orderedCollectionIterator struct {
	vocab.ActivityStreamsOrderedCollection
	items vocab.ActivityStreamsOrderedItemsPropertyIterator
	once  bool // only init items once
}

func (iter *orderedCollectionIterator) NextItem() TypeOrIRI {
	if !iter.initItems() {
		return nil
	}
	cur := iter.items
	iter.items = iter.items.Next()
	return cur
}

func (iter *orderedCollectionIterator) PrevItem() TypeOrIRI {
	if !iter.initItems() {
		return nil
	}
	cur := iter.items
	iter.items = iter.items.Prev()
	return cur
}

func (iter *orderedCollectionIterator) TotalItems() int {
	totalItems := iter.GetActivityStreamsTotalItems()
	if totalItems == nil || !totalItems.IsXMLSchemaNonNegativeInteger() {
		return -1
	}

	return totalItems.Get()
}

func (iter *orderedCollectionIterator) initItems() bool {
	if iter.once {
		return (iter.items != nil)
	}
	iter.once = true
	if iter.ActivityStreamsOrderedCollection == nil {
		return false // no page set
	}
	items := iter.GetActivityStreamsOrderedItems()
	if items == nil {
		return false // no items found
	}
	iter.items = items.Begin()
	return (iter.items != nil)
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

func (iter *regularCollectionPageIterator) NextItem() TypeOrIRI {
	if !iter.initItems() {
		return nil
	}
	cur := iter.items
	iter.items = iter.items.Next()
	return cur
}

func (iter *regularCollectionPageIterator) PrevItem() TypeOrIRI {
	if !iter.initItems() {
		return nil
	}
	cur := iter.items
	iter.items = iter.items.Prev()
	return cur
}

func (iter *regularCollectionPageIterator) TotalItems() int {
	totalItems := iter.GetActivityStreamsTotalItems()
	if totalItems == nil || !totalItems.IsXMLSchemaNonNegativeInteger() {
		return -1
	}

	return totalItems.Get()
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

func (iter *orderedCollectionPageIterator) NextItem() TypeOrIRI {
	if !iter.initItems() {
		return nil
	}
	cur := iter.items
	iter.items = iter.items.Next()
	return cur
}

func (iter *orderedCollectionPageIterator) PrevItem() TypeOrIRI {
	if !iter.initItems() {
		return nil
	}
	cur := iter.items
	iter.items = iter.items.Prev()
	return cur
}

func (iter *orderedCollectionPageIterator) TotalItems() int {
	totalItems := iter.GetActivityStreamsTotalItems()
	if totalItems == nil || !totalItems.IsXMLSchemaNonNegativeInteger() {
		return -1
	}

	return totalItems.Get()
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

	// First page details.
	First *paging.Page
	Query url.Values

	// Total no. items.
	// Omitted if nil.
	Total *int
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
	AppendActivityStreamsCreate(vocab.ActivityStreamsCreate)

	// NOTE: add more of the items-property-like interface
	// functions here as you require them for building pages.
}

// NewASCollection builds and returns a new ActivityStreams Collection from given parameters.
func NewASCollection(params CollectionParams) vocab.ActivityStreamsCollection {
	collection := streams.NewActivityStreamsCollection()
	buildCollection(collection, params)
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
	buildCollection(collection, params)
	return collection
}

// NewASOrderedCollectionPage builds and returns a new ActivityStreams OrderedCollectionPage from given parameters (including item property appending function).
func NewASOrderedCollectionPage(params CollectionPageParams) vocab.ActivityStreamsOrderedCollectionPage {
	collectionPage := streams.NewActivityStreamsOrderedCollectionPage()
	itemsProp := streams.NewActivityStreamsOrderedItemsProperty()
	buildCollectionPage(collectionPage, itemsProp, collectionPage.SetActivityStreamsOrderedItems, params)
	return collectionPage
}

func buildCollection[C CollectionBuilder](collection C, params CollectionParams) {
	// Add the collection ID property.
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(params.ID)
	collection.SetJSONLDId(idProp)

	// Add the collection totalItems count property.
	if params.Total != nil {
		totalItems := streams.NewActivityStreamsTotalItemsProperty()
		totalItems.Set(*params.Total)
		collection.SetActivityStreamsTotalItems(totalItems)
	}

	// No First page means we're done.
	if params.First == nil {
		return
	}

	// Append paging query params
	// to those already in ID prop.
	pageQueryParams := appendQuery(
		params.Query,
		params.ID.Query(),
	)

	// Build the first page link IRI.
	firstIRI := params.First.ToLinkURL(
		params.ID.Scheme,
		params.ID.Host,
		params.ID.Path,
		pageQueryParams,
	)

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

	// Append paging query params
	// to those already in ID prop.
	pageQueryParams := appendQuery(
		params.Query,
		params.ID.Query(),
	)

	// Build the current page link IRI.
	currentIRI := params.Current.ToLinkURL(
		params.ID.Scheme,
		params.ID.Host,
		params.ID.Path,
		pageQueryParams,
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
		pageQueryParams,
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
		pageQueryParams,
	)

	if prevIRI != nil {
		// Add the collection prev property for the prev page.
		prevProp := streams.NewActivityStreamsPrevProperty()
		prevProp.SetIRI(prevIRI)
		collectionPage.SetActivityStreamsPrev(prevProp)
	}

	// Add the collection totalItems count property.
	if params.Total != nil {
		totalItems := streams.NewActivityStreamsTotalItemsProperty()
		totalItems.Set(*params.Total)
		collectionPage.SetActivityStreamsTotalItems(totalItems)
	}

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

// appendQuery appends query values in 'src' to 'dst', returning 'dst'.
func appendQuery(dst, src url.Values) url.Values {
	if dst == nil {
		return src
	}
	for k, vs := range src {
		dst[k] = append(dst[k], vs...)
	}
	return dst
}
