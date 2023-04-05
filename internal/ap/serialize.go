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
