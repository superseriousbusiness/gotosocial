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
	"errors"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

// SerializeOrderedCollection is a custom serializer for an ActivityStreamsOrderedCollection.
// Unlike the standard streams.Serialize function, this serializer normalizes the orderedItems
// value to always be an array/slice, regardless of how many items are contained therein.
//
// TODO: Remove this function if we can fix the underlying issue in Go-Fed.
//
// See:
//   - https://github.com/go-fed/activity/issues/139
//   - https://github.com/mastodon/mastodon/issues/24225
func SerializeOrderedCollection(orderedCollection vocab.ActivityStreamsOrderedCollection) (map[string]interface{}, error) {
	data, err := streams.Serialize(orderedCollection)
	if err != nil {
		return nil, err
	}

	return data, normalizeOrderedCollectionData(data)
}

func normalizeOrderedCollectionData(rawOrderedCollection map[string]interface{}) error {
	orderedItems, ok := rawOrderedCollection["orderedItems"]
	if !ok {
		return errors.New("no orderedItems set on OrderedCollection")
	}

	if _, ok := orderedItems.([]interface{}); ok {
		// Already slice.
		return nil
	}

	orderedItemsString, ok := orderedItems.(string)
	if !ok {
		return errors.New("orderedItems was neither slice nor string")
	}

	rawOrderedCollection["orderedItems"] = []string{orderedItemsString}

	return nil
}

// SerializeAccountable is a custom serializer for any Accountable type.
// This serializer rewrites the 'attachment' value of the Accountable, if
// present, to always be an array.
//
// While this is not strictly necessary in json-ld terms, most other fedi
// implementations look for attachment to be an array of PropertyValue (field)
// entries, and will not parse single-entry, non-array attachments on accounts
// properly.
func SerializeAccountable(accountable vocab.Type) (map[string]interface{}, error) {
	data, err := streams.Serialize(accountable)
	if err != nil {
		return nil, err
	}

	return data, normalizeAccountableAttachments(data)
}

func normalizeAccountableAttachments(rawAccountable map[string]interface{}) error {
	attachment, ok := rawAccountable["attachment"]
	if !ok {
		// 'attachment' not set,
		// so nothing to rewrite.
		return nil
	}

	if _, ok := attachment.([]interface{}); ok {
		// Already slice.
		return nil
	}

	// Coerce single-object to slice.
	rawAccountable["attachment"] = []interface{}{attachment}

	return nil
}
