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
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
)

// NormalizeActivityObject normalizes the 'object'.'content' field of the given Activity.
//
// The rawActivity map should the freshly deserialized json representation of the Activity.
//
// This function is a noop if the type passed in is anything except a Create with a Statusable as its Object.
func NormalizeActivityObject(activity pub.Activity, rawActivity map[string]interface{}) {
	if activity.GetTypeName() != ActivityCreate {
		// Only interested in Create right now.
		return
	}

	withObject, ok := activity.(WithObject)
	if !ok {
		// Create was not a WithObject.
		return
	}

	createObject := withObject.GetActivityStreamsObject()
	if createObject == nil {
		// No object set.
		return
	}

	if createObject.Len() != 1 {
		// Not interested in Object arrays.
		return
	}

	// We now know length is 1 so get the first
	// item from the iter.  We need this to be
	// a Statusable if we're to continue.
	i := createObject.At(0)
	if i == nil {
		// This is awkward.
		return
	}

	t := i.GetType()
	if t == nil {
		// This is also awkward.
		return
	}

	statusable, ok := t.(Statusable)
	if !ok {
		// Object is not Statusable;
		// we're not interested.
		return
	}

	object, ok := rawActivity["object"]
	if !ok {
		// No object in raw map.
		return
	}

	rawStatusable, ok := object.(map[string]interface{})
	if !ok {
		// Object wasn't a json object.
		return
	}

	// Pass in the statusable and its raw JSON representation.
	NormalizeStatusableContent(statusable, rawStatusable)
}

// NormalizeStatusableContent replaces the Content of the given statusable
// with the raw 'content' value from the given json object map.
//
// noop if there was no content in the json object map or the content was
// not a plain string.
func NormalizeStatusableContent(statusable Statusable, rawStatusable map[string]interface{}) {
	content, ok := rawStatusable["content"]
	if !ok {
		// No content in rawStatusable.
		// TODO: In future we might also
		// look for "contentMap" property.
		return
	}

	rawContent, ok := content.(string)
	if !ok {
		// Not interested in content arrays.
		return
	}

	// Set normalized content property from the raw string; this
	// will replace any existing content property on the statusable.
	contentProp := streams.NewActivityStreamsContentProperty()
	contentProp.AppendXMLSchemaString(rawContent)
	statusable.SetActivityStreamsContent(contentProp)
}
