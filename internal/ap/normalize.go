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
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

/*
	INCOMING NORMALIZATION
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
	dataIfaces, rawData, ok := ExtractActivityData(activity, rawJSON)
	if !ok || len(dataIfaces) != len(rawData) {
		// non-equal lengths *shouldn't* happen,
		// but this is just an integrity check.
		return
	}

	// Iterate over the available data.
	for i, dataIface := range dataIfaces {
		// Try to get as vocab.Type, else
		// skip this entry for normalization.
		dataType := dataIface.GetType()
		if dataType == nil {
			continue
		}

		// Get the raw data map at index, else skip
		// this entry due to impossible normalization.
		rawData, ok := rawData[i].(map[string]any)
		if !ok {
			continue
		}

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
			NormalizeIncomingFields(accountable, rawData)
			continue
		}
	}
}

// normalizeContent normalizes the given content
// string by sanitizing its HTML and minimizing it.
//
// Noop for non-string content.
func normalizeContent(rawContent interface{}) string {
	if rawContent == nil {
		// Nothing to fix.
		return ""
	}

	content, ok := rawContent.(string)
	if !ok {
		// Not interested in
		// content slices etc.
		return ""
	}

	if content == "" {
		// Nothing to fix.
		return ""
	}

	// Content entries should be HTML encoded by default:
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-content
	//
	// TODO: sanitize differently based on mediaType.
	// https://www.w3.org/TR/activitystreams-vocabulary/#dfn-mediatype
	content = text.SanitizeToHTML(content)
	content = text.MinifyHTML(content)
	return content
}

// NormalizeIncomingContent replaces the Content property of the given
// item with the normalized versions of the raw 'content' and 'contentMap'
// values from the raw json object map.
//
// noop if there was no 'content' or 'contentMap' in the json object map.
func NormalizeIncomingContent(item WithContent, rawJSON map[string]interface{}) {
	var (
		rawContent    = rawJSON["content"]
		rawContentMap = rawJSON["contentMap"]
	)

	if rawContent == nil &&
		rawContentMap == nil {
		// Nothing to normalize,
		// leave no content on item.
		return
	}

	// Create wrapper for normalized content.
	contentProp := streams.NewActivityStreamsContentProperty()

	// Fix 'content' if applicable.
	content := normalizeContent(rawContent)
	if content != "" {
		contentProp.AppendXMLSchemaString(content)
	}

	// Fix 'contentMap' if applicable.
	contentMap, ok := rawContentMap.(map[string]interface{})
	if ok {
		rdfLangs := make(map[string]string, len(contentMap))

		for lang, rawContent := range contentMap {
			content := normalizeContent(rawContent)
			if content != "" {
				rdfLangs[lang] = content
			}
		}

		if len(rdfLangs) != 0 {
			contentProp.AppendRDFLangString(rdfLangs)
		}
	}

	// Replace any existing content property
	// on the item with normalized version.
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

// NormalizeIncomingFields sanitizes any PropertyValue fields on the
// given WithAttachment interface, by removing html completely from
// the "name" field, and sanitizing dodgy HTML out of the "value" field.
func NormalizeIncomingFields(item WithAttachment, rawJSON map[string]interface{}) {
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

		if !iter.IsSchemaPropertyValue() {
			// Not interested.
			continue
		}

		pv := iter.GetSchemaPropertyValue()
		if pv == nil {
			// Odd.
			continue
		}

		rawPv, ok := attachments[i].(map[string]interface{})
		if !ok {
			continue
		}

		NormalizeIncomingName(pv, rawPv)
		NormalizeIncomingValue(pv, rawPv)
	}
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

// NormalizeIncomingValue replaces the Value of the given
// tem with the raw 'value' from the raw json object map.
//
// noop if there was no name in the json object map or the
// value was not a plain string.
func NormalizeIncomingValue(item WithValue, rawJSON map[string]interface{}) {
	rawValue, ok := rawJSON["value"]
	if !ok {
		// No value in rawJSON.
		return
	}

	value, ok := rawValue.(string)
	if !ok {
		// Not interested in non-string name.
		return
	}

	// Value often contains links or
	// mentions or other little snippets.
	// Sanitize to HTML to allow these.
	value = text.SanitizeToHTML(value)

	// Set normalized name property from the raw string; this
	// will replace any existing value property on the item.
	valueProp := streams.NewSchemaValueProperty()
	valueProp.Set(value)
	item.SetSchemaValue(valueProp)
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

/*
	OUTGOING NORMALIZATION
	The below functions should be called to normalize the content
	of messages *GOING OUT OF* GoToSocial via the federation API,
	either as the result of delivery to a remote instance from this
	instance, or as a result of a remote instance doing an http call
	to us to dereference something.
*/

// NormalizeOutgoingAttachmentProp replaces single-entry Attachment objects with
// single-entry arrays, for better compatibility with other AP implementations.
//
// Ie:
//
//	"attachment": {
//	  ...
//	}
//
// becomes:
//
//	"attachment": [
//	  {
//	    ...
//	  }
//	]
//
// Noop for items with no attachments, or with attachments that are already a slice.
func NormalizeOutgoingAttachmentProp(item WithAttachment, rawJSON map[string]interface{}) {
	attachment, ok := rawJSON["attachment"]
	if !ok {
		// No 'attachment',
		// nothing to change.
		return
	}

	if _, ok := attachment.([]interface{}); ok {
		// Already slice,
		// nothing to change.
		return
	}

	// Coerce single-object to slice.
	rawJSON["attachment"] = []interface{}{attachment}
}

// NormalizeOutgoingAlsoKnownAsProp replaces single-entry alsoKnownAs values with
// single-entry arrays, for better compatibility with other AP implementations.
//
// Ie:
//
//	"alsoKnownAs": "https://example.org/users/some_user"
//
// becomes:
//
//	"alsoKnownAs": ["https://example.org/users/some_user"]
//
// Noop for items with no attachments, or with attachments that are already a slice.
func NormalizeOutgoingAlsoKnownAsProp(item WithAlsoKnownAs, rawJSON map[string]interface{}) {
	alsoKnownAs, ok := rawJSON["alsoKnownAs"]
	if !ok {
		// No 'alsoKnownAs',
		// nothing to change.
		return
	}

	if _, ok := alsoKnownAs.([]interface{}); ok {
		// Already slice,
		// nothing to change.
		return
	}

	// Coerce single-object to slice.
	rawJSON["alsoKnownAs"] = []interface{}{alsoKnownAs}
}

// NormalizeOutgoingContentProp normalizes go-fed's funky formatting of content and
// contentMap properties to a format better understood by other AP implementations.
//
// Ie., incoming "content" property like this:
//
//	"content": [
//	  "hello world!",
//	  {
//	    "en": "hello world!"
//	  }
//	]
//
// Is unpacked to:
//
//	"content": "hello world!",
//	"contentMap": {
//	  "en": "hello world!"
//	}
//
// Noop if neither content nor contentMap are set.
func NormalizeOutgoingContentProp(item WithContent, rawJSON map[string]interface{}) {
	contentProp := item.GetActivityStreamsContent()
	if contentProp == nil {
		// Nothing to do,
		// bail early.
		return
	}

	contentPropLen := contentProp.Len()
	if contentPropLen == 0 {
		// Nothing to do,
		// bail early.
		return
	}

	var (
		content    string
		contentMap map[string]string
	)

	for iter := contentProp.Begin(); iter != contentProp.End(); iter = iter.Next() {
		switch {
		case iter.IsRDFLangString() &&
			contentMap == nil:
			contentMap = iter.GetRDFLangString()

		case content == "" &&
			iter.IsXMLSchemaString():
			content = iter.GetXMLSchemaString()
		}
	}

	if content != "" {
		rawJSON["content"] = content
	} else {
		delete(rawJSON, "content")
	}

	if contentMap != nil {
		rawJSON["contentMap"] = contentMap
	} else {
		delete(rawJSON, "contentMap")
	}
}

// NormalizeOutgoingInteractionPolicyProp replaces single-entry interactionPolicy values
// with single-entry arrays, for better compatibility with other AP implementations.
//
// Ie:
//
//	"interactionPolicy": {
//		"canAnnounce": {
//			"always": "https://www.w3.org/ns/activitystreams#Public",
//			"approvalRequired": []
//		},
//		"canLike": {
//			"always": "https://www.w3.org/ns/activitystreams#Public",
//			"approvalRequired": []
//		},
//		"canReply": {
//			"always": "https://www.w3.org/ns/activitystreams#Public",
//			"approvalRequired": []
//		}
//	}
//
// becomes:
//
//	"interactionPolicy": {
//		"canAnnounce": {
//			"always": [
//				"https://www.w3.org/ns/activitystreams#Public"
//			],
//			"approvalRequired": []
//		},
//		"canLike": {
//			"always": [
//				"https://www.w3.org/ns/activitystreams#Public"
//			],
//			"approvalRequired": []
//		},
//		"canReply": {
//			"always": [
//				"https://www.w3.org/ns/activitystreams#Public"
//			],
//			"approvalRequired": []
//		}
//	}
//
// Noop for items with no attachments, or with attachments that are already a slice.
func NormalizeOutgoingInteractionPolicyProp(item WithInteractionPolicy, rawJSON map[string]interface{}) {
	policy, ok := rawJSON["interactionPolicy"]
	if !ok {
		// No 'interactionPolicy',
		// nothing to change.
		return
	}

	policyMap, ok := policy.(map[string]interface{})
	if !ok {
		// Malformed 'interactionPolicy',
		// nothing to change.
		return
	}

	for _, rulesKey := range []string{
		"canLike",
		"canReply",
		"canAnnounce",
	} {
		// Either "canAnnounce",
		// "canLike", or "canApprove"
		rulesVal, ok := policyMap[rulesKey]
		if !ok {
			// Not set.
			return
		}

		rulesValMap, ok := rulesVal.(map[string]interface{})
		if !ok {
			// Malformed or not
			// present skip.
			return
		}

		for _, PolicyValuesKey := range []string{
			"always",
			"approvalRequired",
		} {
			PolicyValuesVal, ok := rulesValMap[PolicyValuesKey]
			if !ok {
				// Not set.
				continue
			}

			if _, ok := PolicyValuesVal.([]interface{}); ok {
				// Already slice,
				// nothing to change.
				continue
			}

			// Coerce single-object to slice.
			rulesValMap[PolicyValuesKey] = []interface{}{PolicyValuesVal}
		}
	}
}

// NormalizeOutgoingObjectProp normalizes each Object entry in the rawJSON of the given
// item by calling custom serialization / normalization functions on them in turn.
//
// This function also unnests single-entry arrays, so that:
//
//	"object": [
//	  {
//	    ...
//	  }
//	]
//
// Becomes:
//
//	"object": {
//	  ...
//	}
//
// Noop for each Object entry that isn't an Accountable or Statusable.
func NormalizeOutgoingObjectProp(item WithObject, rawJSON map[string]interface{}) error {
	objectProp := item.GetActivityStreamsObject()
	if objectProp == nil {
		// Nothing to do,
		// bail early.
		return nil
	}

	objectPropLen := objectProp.Len()
	if objectPropLen == 0 {
		// Nothing to do,
		// bail early.
		return nil
	}

	// The thing we already serialized has objects
	// on it, so we should see if we need to custom
	// serialize any of those objects, and replace
	// them on the data map as necessary.
	objects := make([]interface{}, 0, objectPropLen)
	for iter := objectProp.Begin(); iter != objectProp.End(); iter = iter.Next() {
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
			return gtserror.Newf("could not resolve object iter %T to vocab.Type", iter)
		}

		var err error

		// In the below accountable and statusable serialization,
		// `@context` will be included in the wrapping type already,
		// so we shouldn't also include it in the object itself.
		switch tn := objectType.GetTypeName(); {
		case IsAccountable(tn):
			objectSer, err = serializeAccountable(objectType, false)

		case IsStatusable(tn):
			// IsStatusable includes Pollable as well.
			objectSer, err = serializeStatusable(objectType, false)

		default:
			// No custom serializer for this type; serialize as normal.
			objectSer, err = objectType.Serialize()
		}

		if err != nil {
			return err
		}

		objects = append(objects, objectSer)
	}

	if objectPropLen == 1 {
		// Unnest single object.
		rawJSON["object"] = objects[0]
	} else {
		// Array of objects.
		rawJSON["object"] = objects
	}

	return nil
}
