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

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
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
//   - OrderedCollection: 'orderedItems' property will always be made into an array.
//   - Any Accountable type: 'attachment' property will always be made into an array.
//   - Update: any Accountable 'object's set on an update will be custom serialized as above.
func Serialize(t vocab.Type) (m map[string]interface{}, e error) {
	switch t.GetTypeName() {
	case ObjectOrderedCollection:
		return serializeOrderedCollection(t)
	case ActorApplication, ActorGroup, ActorOrganization, ActorPerson, ActorService:
		return serializeAccountable(t, true)
	case ActivityUpdate:
		return serializeWithObject(t)
	default:
		// No custom serializer necessary.
		return streams.Serialize(t)
	}
}

// serializeOrderedCollection is a custom serializer for an ActivityStreamsOrderedCollection.
// Unlike the standard streams.Serialize function, this serializer normalizes the orderedItems
// value to always be an array/slice, regardless of how many items are contained therein.
//
// TODO: Remove this function if we can fix the underlying issue in Go-Fed.
//
// See:
//   - https://github.com/go-fed/activity/issues/139
//   - https://github.com/mastodon/mastodon/issues/24225
func serializeOrderedCollection(orderedCollection vocab.Type) (map[string]interface{}, error) {
	data, err := streams.Serialize(orderedCollection)
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
// This serializer rewrites the 'attachment' value of the Accountable, if
// present, to always be an array/slice.
//
// While this is not strictly necessary in json-ld terms, most other fedi
// implementations look for attachment to be an array of PropertyValue (field)
// entries, and will not parse single-entry, non-array attachments on accounts
// properly.
//
// If the accountable is being serialized as a top-level object (eg., for serving
// in response to an account dereference request), then includeContext should be
// set to true, so as to include the json-ld '@context' entries in the data.
// If the accountable is being serialized as part of another object (eg., as the
// object of an activity), then includeContext should be set to false, as the
// @context entry should be included on the top-level/wrapping activity/object.
func serializeAccountable(accountable vocab.Type, includeContext bool) (map[string]interface{}, error) {
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

	attachment, ok := data["attachment"]
	if !ok {
		// No 'attachment', nothing to change.
		return data, nil
	}

	if _, ok := attachment.([]interface{}); ok {
		// Already slice.
		return data, nil
	}

	// Coerce single-object to slice.
	data["attachment"] = []interface{}{attachment}

	return data, nil
}

func serializeWithObject(t vocab.Type) (map[string]interface{}, error) {
	withObject, ok := t.(WithObject)
	if !ok {
		return nil, fmt.Errorf("serializeWithObject: could not resolve %T to WithObject", t)
	}

	data, err := streams.Serialize(t)
	if err != nil {
		return nil, err
	}

	object := withObject.GetActivityStreamsObject()
	if object == nil {
		// Nothing to do, bail early.
		return data, nil
	}

	objectLen := object.Len()
	if objectLen == 0 {
		// Nothing to do, bail early.
		return data, nil
	}

	// The thing we already serialized has objects
	// on it, so we should see if we need to custom
	// serialize any of those objects, and replace
	// them on the data map as necessary.
	objects := make([]interface{}, 0, objectLen)
	for iter := object.Begin(); iter != object.End(); iter = iter.Next() {
		if iter.IsIRI() {
			// Plain IRIs don't need custom serialization.
			objects = append(objects, iter.GetIRI().String())
			continue
		}

		var (
			objectType = iter.GetType()
			objectSer  map[string]interface{}
		)

		if objectType == nil {
			// This is awkward.
			return nil, fmt.Errorf("serializeWithObject: could not resolve object iter %T to vocab.Type", iter)
		}

		switch objectType.GetTypeName() {
		case ActorApplication, ActorGroup, ActorOrganization, ActorPerson, ActorService:
			// @context will be included in wrapping type already,
			// we don't need to include it in the object itself.
			objectSer, err = serializeAccountable(objectType, false)
		default:
			// No custom serializer for this type; serialize as normal.
			objectSer, err = objectType.Serialize()
		}

		if err != nil {
			return nil, err
		}

		objects = append(objects, objectSer)
	}

	if objectLen == 1 {
		// Unnest single object.
		data["object"] = objects[0]
	} else {
		// Array of objects.
		data["object"] = objects
	}

	return data, nil
}
