package ap

import (
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func NormalizeObjectContent(obj vocab.ActivityStreamsObjectProperty, rawActivity map[string]interface{}) {
	// First get the Statusable from the Object property,
	// if it exists. If it doesn't, there's nothing to do.

	if obj.Len() != 1 {
		// Not interested in Object arrays.
		return
	}

	i := obj.At(0)
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
		// Object is not a Statusable,
		// so we're not interested.
		return
	}

	NormalizeStatusable(statusable, rawActivity)
}

func NormalizeStatusable(statusable Statusable, raw map[string]interface{}) {
	// To normalize a Statusable, we need to fetch the
	// raw object "content" property from the rawActivity map.
	var objectRaw map[string]interface{}

	object, ok := raw["object"]
	if !ok {
		// No object in raw map.
		return
	}

	objectRaw, ok = object.()
	if !ok {
		// Object was something other than
		// a JSON object (a url most likely).
		return
	}

	content, ok := objectRaw["content"]
	if !ok {
		// No content in object.
		// TODO: In future we might also
		// look for "contentMap" property.
		return
	}

	contentString, ok := content.(string)
	if !ok {
		// Not interested in content arrays.
		return
	}

	// Set our normalized content property from the raw string.
	contentProp := streams.NewActivityStreamsContentProperty()
	contentProp.AppendXMLSchemaString(contentString)

	// We're done!
	statusable.SetActivityStreamsContent(contentProp)
}
