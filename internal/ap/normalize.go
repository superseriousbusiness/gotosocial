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
	NormalizeStatusableAttachments(statusable, rawStatusable)
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

// NormalizeStatusableAttachments normalizes all attachments (if any) of the given
// statusable, replacing the 'name' (aka content warning) field of each attachment
// with the raw 'name' value from the rawStatusable JSON.
//
// noop if there are no attachments; no replacement done for an attachment if it's
// not in a format we can understand.
func NormalizeStatusableAttachments(statusable Statusable, rawStatusable map[string]interface{}) {
	rawAttachments, ok := rawStatusable["attachment"]
	if !ok {
		// No attachments in rawStatusable.
		return
	}

	// Convert to slice if not already.
	var attachments []interface{}
	if attachments, ok = rawAttachments.([]interface{}); !ok {
		attachments = []interface{}{rawAttachments}
	}

	attachmentProperty := statusable.GetActivityStreamsAttachment()
	if attachmentProperty == nil || attachmentProperty.Len() == 0 {
		// Nothing to do here.
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

		rawAttachmentName, ok := rawAttachment["name"]
		if !ok {
			continue
		}

		attachmentNameString, ok := rawAttachmentName.(string)
		if !ok {
			continue
		}

		// We now have the attachmentable and we have
		// the name string as it came in via the json,
		// so we can proceed to normalize.
		nameProp := streams.NewActivityStreamsNameProperty()
		nameProp.AppendXMLSchemaString(attachmentNameString)
		attachmentable.SetActivityStreamsName(nameProp)
	}
}

func NormalizeAccountableSummary(accountable Accountable, rawAccountable map[string]interface{}) {
	summary, ok := rawAccountable["summary"]
	if !ok {
		// No summary in rawAccountable.
		return
	}

	rawSummary, ok := summary.(string)
	if !ok {
		// Not interested in non-string summary.
		return
	}

	// Set normalized summary property from the raw string; this
	// will replace any existing summary property on the accountable.
	summaryProp := streams.NewActivityStreamsSummaryProperty()
	summaryProp.AppendXMLSchemaString(rawSummary)
	accountable.SetActivityStreamsSummary(summaryProp)
}
