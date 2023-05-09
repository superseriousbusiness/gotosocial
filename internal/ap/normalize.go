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

/*
	NORMALIZE INCOMING
	The below functions should be called to normalize the content
	of messages *COMING INTO* GoToSocial via the federation API,
	either as the result of delivery from a remote instance to this
	instance, or as a result of this instance doing an http call to
	another instance to dereference something.
*/

// NormalizeIncomingActivityObject normalizes the 'object'.'content' field of the given Activity.
//
// The rawActivity map should the freshly deserialized json representation of the Activity.
//
// This function is a noop if the type passed in is anything except a Create or Update with a Statusable or Accountable as its Object.
func NormalizeIncomingActivityObject(activity pub.Activity, rawJSON map[string]interface{}) {
	if typeName := activity.GetTypeName(); typeName != ActivityCreate && typeName != ActivityUpdate {
		// Only interested in Create or Update right now.
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
	// item from the iter. We need this to be
	// a Statusable or Accountable if we're to continue.
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

	switch t.GetTypeName() {
	case ObjectArticle, ObjectDocument, ObjectImage, ObjectVideo, ObjectNote, ObjectPage, ObjectEvent, ObjectPlace, ObjectProfile:
		statusable, ok := t.(Statusable)
		if !ok {
			// Object is not Statusable;
			// we're not interested.
			return
		}

		rawObject, ok := rawJSON["object"]
		if !ok {
			// No object in raw map.
			return
		}

		rawStatusableJSON, ok := rawObject.(map[string]interface{})
		if !ok {
			// Object wasn't a json object.
			return
		}

		// Normalize everything we can on the statusable.
		NormalizeIncomingContent(statusable, rawStatusableJSON)
		NormalizeIncomingAttachments(statusable, rawStatusableJSON)
		NormalizeIncomingSummary(statusable, rawStatusableJSON)
		NormalizeIncomingName(statusable, rawStatusableJSON)
	case ActorApplication, ActorGroup, ActorOrganization, ActorPerson, ActorService:
		accountable, ok := t.(Accountable)
		if !ok {
			// Object is not Accountable;
			// we're not interested.
			return
		}

		rawObject, ok := rawJSON["object"]
		if !ok {
			// No object in raw map.
			return
		}

		rawAccountableJSON, ok := rawObject.(map[string]interface{})
		if !ok {
			// Object wasn't a json object.
			return
		}

		// Normalize everything we can on the accountable.
		NormalizeIncomingSummary(accountable, rawAccountableJSON)
	}
}

// NormalizeIncomingContent replaces the Content of the given item
// with the raw 'content' value from the raw json object map.
//
// noop if there was no content in the json object map or the
// content was not a plain string.
func NormalizeIncomingContent(item WithSetContent, rawJSON map[string]interface{}) {
	rawContent, ok := rawJSON["content"]
	if !ok {
		// No content in rawJSON.
		// TODO: In future we might also
		// look for "contentMap" property.
		return
	}

	content, ok := rawContent.(string)
	if !ok {
		// Not interested in content arrays.
		return
	}

	// Set normalized content property from the raw string;
	// this replaces any existing content property on the item.
	contentProp := streams.NewActivityStreamsContentProperty()
	contentProp.AppendXMLSchemaString(content)
	item.SetActivityStreamsContent(contentProp)
}

// NormalizeIncomingAttachments normalizes all attachments (if any) of the given
// item, replacing the 'name' (aka content warning) field of each attachment
// with the raw 'name' value from the raw json object map.
//
// noop if there are no attachments; noop if attachment is not a format
// we can understand.
func NormalizeIncomingAttachments(item WithAttachment, rawJSON map[string]interface{}) {
	rawAttachments, ok := rawJSON["attachment"]
	if !ok {
		// No attachments in rawJSON.
		return
	}

	// Convert to slice if not already,
	// so we can iterate through it.
	var attachments []interface{}
	if attachments, ok = rawAttachments.([]interface{}); !ok {
		attachments = []interface{}{rawAttachments}
	}

	attachmentProperty := item.GetActivityStreamsAttachment()
	if attachmentProperty == nil {
		// Nothing to do here.
		return
	}

	if l := attachmentProperty.Len(); l == 0 || l != len(attachments) {
		// Mismatch between item and
		// JSON, can't normalize.
		return
	}

	// Keep an index of where we are in the iter;
	// we need this so we can modify the correct
	// attachment, in case of multiples.
	i := -1

	for iter := attachmentProperty.Begin(); iter != attachmentProperty.End(); iter = iter.Next() {
		i++

		t := iter.GetType()
		if t == nil {
			continue
		}

		attachmentable, ok := t.(Attachmentable)
		if !ok {
			continue
		}

		rawAttachment, ok := attachments[i].(map[string]interface{})
		if !ok {
			continue
		}

		NormalizeIncomingName(attachmentable, rawAttachment)
	}
}

// NormalizeIncomingSummary replaces the Summary of the given item
// with the raw 'summary' value from the raw json object map.
//
// noop if there was no summary in the json object map or the
// summary was not a plain string.
func NormalizeIncomingSummary(item WithSetSummary, rawJSON map[string]interface{}) {
	rawSummary, ok := rawJSON["summary"]
	if !ok {
		// No summary in rawJSON.
		return
	}

	summary, ok := rawSummary.(string)
	if !ok {
		// Not interested in non-string summary.
		return
	}

	// Set normalized summary property from the raw string; this
	// will replace any existing summary property on the item.
	summaryProp := streams.NewActivityStreamsSummaryProperty()
	summaryProp.AppendXMLSchemaString(summary)
	item.SetActivityStreamsSummary(summaryProp)
}

// NormalizeIncomingName replaces the Name of the given item
// with the raw 'name' value from the raw json object map.
//
// noop if there was no name in the json object map or the
// name was not a plain string.
func NormalizeIncomingName(item WithSetName, rawJSON map[string]interface{}) {
	rawName, ok := rawJSON["name"]
	if !ok {
		// No name in rawJSON.
		return
	}

	name, ok := rawName.(string)
	if !ok {
		// Not interested in non-string name.
		return
	}

	// Set normalized name property from the raw string; this
	// will replace any existing name property on the item.
	nameProp := streams.NewActivityStreamsNameProperty()
	nameProp.AppendXMLSchemaString(name)
	item.SetActivityStreamsName(nameProp)
}
