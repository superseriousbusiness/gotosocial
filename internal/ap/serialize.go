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
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// Serialize is a custom serializer for ActivityStreams types.
//
// In most cases, it will simply call the go-fed streams.Serialize function under the hood.
// However, if custom serialization is required on a specific type (eg for inter-implementation
// compatibility), it can be inserted into the switch as necessary.
//
// Callers should always call this function instead of streams.Serialize, unless there's a
// very good reason to do otherwise.
//
// Currently, the following things will be custom serialized:
//
//   - OrderedCollection:       'orderedItems' property will always be made into an array.
//   - OrderedCollectionPage:   'orderedItems' property will always be made into an array.
//   - Any Accountable type:    'attachment' property will always be made into an array.
//   - Any Statusable type:     'attachment' property will always be made into an array; 'content', 'contentMap', and 'interactionPolicy' will be normalized.
//   - Any Activityable type:   any 'object's set on an activity will be custom serialized as above.
func Serialize(t vocab.Type) (m map[string]interface{}, e error) {
	switch tn := t.GetTypeName(); {
	case tn == ObjectOrderedCollection ||
		tn == ObjectOrderedCollectionPage:
		return serializeWithOrderedItems(t)
	case IsAccountable(tn):
		return serializeAccountable(t, true)
	case IsStatusable(tn):
		return serializeStatusable(t, true)
	case IsActivityable(tn):
		return serializeActivityable(t, true)
	default:
		// No custom serializer necessary.
		return streams.Serialize(t)
	}
}

// serializeWithOrderedItems is a custom serializer
// for any type that has an `orderedItems` property.
// Unlike the standard streams.Serialize function,
// this serializer normalizes the orderedItems
// value to always be an array/slice, regardless
// of how many items are contained therein.
//
// See:
//   - https://github.com/go-fed/activity/issues/139
//   - https://github.com/mastodon/mastodon/issues/24225
func serializeWithOrderedItems(t vocab.Type) (map[string]interface{}, error) {
	data, err := streams.Serialize(t)
	if err != nil {
		return nil, err
	}

	orderedItems, ok := data["orderedItems"]
	if !ok {
		// No 'orderedItems', nothing to change.
		return data, nil
	}

	if _, ok := orderedItems.([]interface{}); ok {
		// Already slice.
		return data, nil
	}

	// Coerce single-object to slice.
	data["orderedItems"] = []interface{}{orderedItems}

	return data, nil
}

// SerializeAccountable is a custom serializer for any Accountable type.
// This serializer rewrites certain values of the Accountable, if present,
// to always be an array/slice.
//
// While this may not always be strictly necessary in json-ld terms, most other
// fedi implementations look for certain fields to be an array and will not parse
// single-entry, non-array fields on accounts properly.
//
// If the accountable is being serialized as a top-level object (eg., for serving
// in response to an account dereference request), then includeContext should be
// set to true, so as to include the json-ld '@context' entries in the data.
// If the accountable is being serialized as part of another object (eg., as the
// object of an activity), then includeContext should be set to false, as the
// @context entry should be included on the top-level/wrapping activity/object.
func serializeAccountable(t vocab.Type, includeContext bool) (map[string]interface{}, error) {
	accountable, ok := t.(Accountable)
	if !ok {
		return nil, gtserror.Newf("vocab.Type %T not accountable", t)
	}

	var (
		data map[string]interface{}
		err  error
	)

	if includeContext {
		data, err = streams.Serialize(accountable)
	} else {
		data, err = accountable.Serialize()
	}

	if err != nil {
		return nil, err
	}

	NormalizeOutgoingAttachmentProp(accountable, data)
	NormalizeOutgoingAlsoKnownAsProp(accountable, data)

	return data, nil
}

func serializeStatusable(t vocab.Type, includeContext bool) (map[string]interface{}, error) {
	statusable, ok := t.(Statusable)
	if !ok {
		return nil, gtserror.Newf("vocab.Type %T not statusable", t)
	}

	var (
		data map[string]interface{}
		err  error
	)

	if includeContext {
		data, err = streams.Serialize(statusable)
	} else {
		data, err = statusable.Serialize()
	}

	if err != nil {
		return nil, err
	}

	NormalizeOutgoingAttachmentProp(statusable, data)
	NormalizeOutgoingContentProp(statusable, data)
	NormalizeOutgoingInteractionPolicyProp(statusable, data)

	return data, nil
}

func serializeActivityable(t vocab.Type, includeContext bool) (map[string]interface{}, error) {
	activityable, ok := t.(Activityable)
	if !ok {
		return nil, gtserror.Newf("vocab.Type %T not activityable", t)
	}

	var (
		data map[string]interface{}
		err  error
	)

	if includeContext {
		data, err = streams.Serialize(activityable)
	} else {
		data, err = activityable.Serialize()
	}

	if err != nil {
		return nil, err
	}

	if err := NormalizeOutgoingObjectProp(activityable, data); err != nil {
		return nil, err
	}

	return data, nil
}
