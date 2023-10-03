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
	"github.com/superseriousbusiness/gotosocial/internal/text"
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
func NormalizeIncomingActivity(activity pub.Activity, rawJSON map[string]interface{}) {
	// From the activity extract the data vocab.Type + its "raw" JSON.
	dataTypes, rawData, ok := ExtractActivityData(activity, rawJSON)
	if !ok || len(dataTypes) != len(rawData) {
		return
	}

	// Iterate over the available data.
	for i, dataType := range dataTypes {

		// Get the raw data map at type index.
		rawData, _ := rawData[i].(map[string]any)

		if statusable, ok := ToStatusable(dataType); ok {
			if pollable, ok := ToPollable(dataType); ok {
				// Normalize the Pollable specific properties.
				NormalizeIncomingPollOptions(pollable, rawData)
			}

			// Normalize everything we can on the statusable.
			NormalizeIncomingContent(statusable, rawData)
			NormalizeIncomingAttachments(statusable, rawData)
			NormalizeIncomingSummary(statusable, rawData)
			NormalizeIncomingName(statusable, rawData)
			continue
		}

		if accountable, ok := ToAccountable(dataType); ok {
			// Normalize everything we can on the accountable.
			NormalizeIncomingSummary(accountable, rawData)
			continue
		}
	}
}

// NormalizeIncomingContent replaces the Content of the given item
// with the sanitized version of the raw 'content' value from the
// raw json object map.
//
// noop if there was no content in the json object map or the
// content was not a plain string.
func NormalizeIncomingContent(item WithContent, rawJSON map[string]interface{}) {
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

	// Content should be HTML encoded by default:
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-content
	//
	// TODO: sanitize differently based on mediaType.
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-mediatype
	content = text.SanitizeToHTML(content)
	content = text.MinifyHTML(content)

	// Set normalized content property from the raw string;
	// this replaces any existing content property on the item.
	contentProp := streams.NewActivityStreamsContentProperty()
	contentProp.AppendXMLSchemaString(content)
	item.SetActivityStreamsContent(contentProp)
}

// NormalizeIncomingAttachments normalizes all attachments (if any) of the given
// item, replacing the 'name' (aka content warning) field of each attachment
// with the raw 'name' value from the raw json object map, and doing sanitization
// on the result.
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
// with the sanitized version of the raw 'summary' value from the
// raw json object map.
//
// noop if there was no summary in the json object map or the
// summary was not a plain string.
func NormalizeIncomingSummary(item WithSummary, rawJSON map[string]interface{}) {
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

	// Summary should be HTML encoded:
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-summary
	summary = text.SanitizeToHTML(summary)
	summary = text.MinifyHTML(summary)

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
func NormalizeIncomingName(item WithName, rawJSON map[string]interface{}) {
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

	// Name *must not* include any HTML markup:
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-name
	//
	// todo: We probably want to update this to allow
	// *escaped* HTML markup, but for now just nuke it.
	name = text.SanitizeToPlaintext(name)

	// Set normalized name property from the raw string; this
	// will replace any existing name property on the item.
	nameProp := streams.NewActivityStreamsNameProperty()
	nameProp.AppendXMLSchemaString(name)
	item.SetActivityStreamsName(nameProp)
}

// NormalizeIncomingOneOf normalizes all oneOf (if any) of the given
// item, replacing the 'name' field of each oneOf with the raw 'name'
// value from the raw json object map, and doing sanitization
// on the result.
//
// noop if there are no oneOf; noop if oneOf is not expected format.
func NormalizeIncomingPollOptions(item WithOneOf, rawJSON map[string]interface{}) {
	var oneOf []interface{}

	// Get the raw one-of JSON data.
	rawOneOf, ok := rawJSON["oneOf"]
	if !ok {
		return
	}

	// Convert to slice if not already, so we can iterate.
	if oneOf, ok = rawOneOf.([]interface{}); !ok {
		oneOf = []interface{}{rawOneOf}
	}

	// Extract the one-of property from interface.
	oneOfProp := item.GetActivityStreamsOneOf()
	if oneOfProp == nil {
		return
	}

	// Check we have useable one-of JSON-vs-unmarshaled data.
	if l := oneOfProp.Len(); l == 0 || l != len(oneOf) {
		return
	}

	// Get start and end of iter.
	start := oneOfProp.Begin()
	end := oneOfProp.End()

	// Iterate a counter, from start through to end iter item.
	for i, iter := 0, start; iter != end; i, iter = i+1, iter.Next() {
		// Get item type.
		t := iter.GetType()

		// Check fulfills Choiceable type
		// (this accounts for nil input type).
		choiceable, ok := t.(PollOptionable)
		if !ok {
			continue
		}

		// Get the corresponding raw one-of data.
		rawChoice, ok := oneOf[i].(map[string]interface{})
		if !ok {
			continue
		}

		NormalizeIncomingName(choiceable, rawChoice)
	}
}
