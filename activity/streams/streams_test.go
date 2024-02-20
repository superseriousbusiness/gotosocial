package streams

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"
	"testing"

	"github.com/go-test/deep"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

// IsKnownResolverError returns true if it is known that an example from
// GetTestTable will trigger a JSONResolver error.
func IsKnownResolverError(t TestTable) (isError bool, reason string) {
	isError = true

	switch t.name {
	case "Example 61":
		reason = "no 'type' property is on the root object"
	case "Example 62":
		reason = "an unknown 'type' property is on the root object"
	case "Example 153":
		reason = "no 'type' property is on the root object"
	default:
		isError = false
	}

	return
}

// SerializeForTest calls Serialize and stabilizes the @context value for
// testing purposes.
func SerializeForTest(a vocab.Type) (m map[string]interface{}, e error) {
	m, e = Serialize(a)
	if e != nil {
		return
	}
	ctx, ok := m["@context"]
	if !ok {
		return
	}
	arr, ok := ctx.([]interface{})
	if !ok {
		return
	}
	var s []string
	var o []interface{}
	for _, v := range arr {
		if str, ok := v.(string); ok {
			s = append(s, str)
		} else {
			o = append(o, v)
		}
	}
	sort.Sort(sort.StringSlice(s))
	newArr := make([]interface{}, 0, len(arr))
	for _, v := range s {
		newArr = append(newArr, v)
	}
	for _, v := range o {
		newArr = append(newArr, v)
	}
	m["@context"] = newArr
	return
}

// PostSerializationAdjustment is needed in rare cases when a test example
// requires post processing on the serialized map to match expectations.
func PostSerializationAdjustment(t TestTable, m map[string]interface{}) (map[string]interface{}, string) {
	adjustReason := ""
	switch t.name {
	case "Service w/ Multiple schema:PropertyValue Attachments":
		m["@context"] = []interface{}{
			"https://www.w3.org/ns/activitystreams",
			map[string]interface{}{
				"schema":        "https://schema.org#",
				"PropertyValue": "schema:PropertyValue",
				"value":         "schema:value",
			},
		}
		adjustReason = "go-fed has no way of knowing that schema.org types need to be in the @context"
	}
	return m, adjustReason
}

func makeResolver(t *testing.T, tc TestTable, expected []byte) (*JSONResolver, error) {
	resFn := func(s vocab.Type) error {
		return nil
	}

	if t != nil {
		resFn = func(s vocab.Type) error {
			m, err := SerializeForTest(s)
			if err != nil {
				return err
			}

			m, adjustReason := PostSerializationAdjustment(tc, m)
			if len(adjustReason) > 0 {
				t.Logf("%s: Post-serialization adjustment: %s", tc.name, adjustReason)
			}
			actual, err := json.Marshal(m)
			if err != nil {
				t.Errorf("json.Marshal returned error: %v", err)
			}
			if diff, err := GetJSONDiff(actual, expected); err == nil && diff != nil {
				t.Error("Serialize JSON equality is false")
				for _, d := range diff {
					t.Log(d)
				}
			} else if err != nil {
				t.Errorf("GetJSONDiff returned error: %v", err)
			}
			return nil
		}
	}

	return NewJSONResolver(
		func(c context.Context, x vocab.ActivityStreamsAccept) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsActivity) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsAdd) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsAnnounce) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsApplication) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsArrive) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsArticle) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsAudio) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsBlock) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsCollection) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsCollectionPage) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsCreate) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsDelete) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsDislike) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsDocument) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsEvent) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsFlag) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsFollow) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsGroup) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsIgnore) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsImage) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsIntransitiveActivity) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsInvite) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsJoin) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsLeave) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsLike) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsLink) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsListen) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsMention) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsMove) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsNote) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsObject) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsOffer) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsOrderedCollection) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsOrderedCollectionPage) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsOrganization) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsPage) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsPerson) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsPlace) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsProfile) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsQuestion) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsRead) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsReject) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsRelationship) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsRemove) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsService) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsTentativeAccept) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsTentativeReject) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsTombstone) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsTravel) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsUndo) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsUpdate) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsVideo) error {
			return resFn(x)
		},
		func(c context.Context, x vocab.ActivityStreamsView) error {
			return resFn(x)
		},
	)
}

func TestJSONResolver(t *testing.T) {
	for _, example := range GetTestTable() {
		example := example // shadow loop variable
		t.Run(example.name, func(t *testing.T) {
			if skip, reason := IsKnownResolverError(example); skip {
				t.Skipf("it is known an error will be returned because %q", reason)
				return
			}

			ex := []byte(example.expectedJSON)
			r, err := makeResolver(t, example, ex)
			if err != nil {
				t.Errorf("Cannot create JSONResolver: %v", err)
				return
			}

			var m map[string]interface{}
			err = json.Unmarshal(ex, &m)
			if err != nil {
				t.Errorf("Cannot json.Unmarshal: %v", err)
				return
			}
			err = r.Resolve(context.Background(), m)
			if err != nil {
				t.Errorf("Cannot JSONResolver.Deserialize: %v", err)
				return
			}
		})
	}
}

func TestJSONResolverErrors(t *testing.T) {
	for _, example := range GetTestTable() {
		example := example // shadow loop variable
		t.Run(example.name, func(t *testing.T) {
			isError, reason := IsKnownResolverError(example)
			if !isError {
				t.Skip("no expected error")
				return
			}

			ex := []byte(example.expectedJSON)
			r, err := makeResolver(nil, example, nil)
			if err != nil {
				t.Errorf("Cannot create JSONResolver: %v", err)
				return
			}

			var m map[string]interface{}
			err = json.Unmarshal([]byte(ex), &m)
			if err != nil {
				t.Errorf("Cannot json.Unmarshal: %v", err)
				return
			}

			t.Logf("Expecting error because %q", reason)

			err = r.Resolve(context.Background(), m)
			if err == nil {
				t.Error("Expected error, but nil was returned.")
			} else {
				t.Logf("Returned error was %v", err)
			}
		})
	}
}

func TestNulls(t *testing.T) {
	makeIRI := func(path string) *url.URL {
		return &url.URL{
			Scheme: "https",
			Host:   "example.com",
			Path:   path,
		}
	}
	var (
		samIRIInbox = makeIRI("/sam/inbox")
		samIRI      = makeIRI("/sam")
		noteIRI     = makeIRI("/note/123")
		sallyIRI    = makeIRI("/sally")
		activityIRI = makeIRI("/test/new/iri")
	)

	noteIdProperty := NewJSONLDIdProperty()
	noteIdProperty.SetIRI(noteIRI)
	expectedNote := NewActivityStreamsNote()
	expectedNote.SetJSONLDId(noteIdProperty)
	noteNameProperty := NewActivityStreamsNameProperty()
	noteNameProperty.AppendXMLSchemaString("A Note")
	expectedNote.SetActivityStreamsName(noteNameProperty)
	noteContentProperty := NewActivityStreamsContentProperty()
	noteContentProperty.AppendXMLSchemaString("This is a simple note")
	expectedNote.SetActivityStreamsContent(noteContentProperty)
	noteToProperty := NewActivityStreamsToProperty()
	expectedSamActor := NewActivityStreamsPerson()
	samInboxProperty := NewActivityStreamsInboxProperty()
	samInboxProperty.SetIRI(samIRIInbox)
	expectedSamActor.SetActivityStreamsInbox(samInboxProperty)
	samIdProperty := NewJSONLDIdProperty()
	samIdProperty.SetIRI(samIRI)
	expectedSamActor.SetJSONLDId(samIdProperty)
	noteToProperty.AppendActivityStreamsPerson(expectedSamActor)
	expectedNote.SetActivityStreamsTo(noteToProperty)
	expectedUpdate := NewActivityStreamsUpdate()
	sallyIdProperty := NewJSONLDIdProperty()
	sallyIdProperty.SetIRI(sallyIRI)
	sallyPerson := NewActivityStreamsPerson()
	sallyPerson.SetJSONLDId(sallyIdProperty)
	sallyActor := NewActivityStreamsActorProperty()
	sallyActor.AppendActivityStreamsPerson(sallyPerson)
	expectedUpdate.SetActivityStreamsActor(sallyActor)
	summaryProperty := NewActivityStreamsSummaryProperty()
	summaryProperty.AppendXMLSchemaString("Sally updated her note")
	expectedUpdate.SetActivityStreamsSummary(summaryProperty)
	updateIdProperty := NewJSONLDIdProperty()
	updateIdProperty.SetIRI(activityIRI)
	expectedUpdate.SetJSONLDId(updateIdProperty)
	objectNote := NewActivityStreamsObjectProperty()
	objectNote.AppendActivityStreamsNote(expectedNote)
	expectedUpdate.SetActivityStreamsObject(objectNote)

	// Variable to aid in deserialization in tests
	tables := []struct {
		name              string
		expected          vocab.Type
		callback          func(*vocab.Type) interface{}
		input             string
		inputWithoutNulls string
	}{
		{
			name:     "JSON nulls are not preserved",
			expected: expectedUpdate,
			callback: func(actual *vocab.Type) interface{} {
				return func(c context.Context, v vocab.ActivityStreamsUpdate) error {
					*actual = v
					return nil
				}
			},
			input: `
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally updated her note",
  "type": "Update",
  "actor": "https://example.com/sally",
  "id": "https://example.com/test/new/iri",
  "object": {
    "id": "https://example.com/note/123",
    "type": "Note",
    "to": {
      "id": "https://example.com/sam",
      "inbox": "https://example.com/sam/inbox",
      "type": "Person",
      "name": null
    }
  }
}
`,
			inputWithoutNulls: `
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally updated her note",
  "type": "Update",
  "actor": "https://example.com/sally",
  "id": "https://example.com/test/new/iri",
  "object": {
    "id": "https://example.com/note/123",
    "type": "Note",
    "to": {
      "id": "https://example.com/sam",
      "inbox": "https://example.com/sam/inbox",
      "type": "Person"
    }
  }
}
`,
		},
	}
	for _, r := range tables {
		r := r // shadow loop variable
		t.Run(r.name, func(t *testing.T) {
			var actual vocab.Type
			res, err := NewJSONResolver(r.callback(&actual))
			if err != nil {
				t.Errorf("cannot create resolver: %s", err)
				return
			}

			var m map[string]interface{}
			err = json.Unmarshal([]byte(r.input), &m)
			if err != nil {
				t.Errorf("Cannot json.Unmarshal: %v", err)
				return
			}
			err = res.Resolve(context.Background(), m)
			if err != nil {
				t.Errorf("Cannot Deserialize: %v", err)
				return
			}
			if diff := deep.Equal(actual, r.expected); diff != nil {
				t.Error("Deserialize deep equal is false")
				for _, d := range diff {
					t.Log(d)
				}
			}
			m, err = SerializeForTest(actual)
			if err != nil {
				t.Errorf("Cannot Serialize: %v", err)
				return
			}
			reser, err := json.Marshal(m)
			if err != nil {
				t.Errorf("Cannot json.Marshal: %v", err)
				return
			}
			if diff, err := GetJSONDiff(reser, []byte(r.inputWithoutNulls)); err == nil && diff != nil {
				t.Error("Serialize JSON equality is false")
				for _, d := range diff {
					t.Log(d)
				}
			} else if err != nil {
				t.Errorf("GetJSONDiff returned error: %v", err)
			}
		})
	}
}

func GetJSONDiff(str1, str2 []byte) ([]string, error) {
	var i1 interface{}
	var i2 interface{}

	err := json.Unmarshal(str1, &i1)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(str2, &i2)
	if err != nil {
		return nil, err
	}
	return deep.Equal(i1, i2), nil
}
