package pub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

const (
	testMyInboxIRI            = "https://example.com/addison/inbox"
	testMyOutboxIRI           = "https://example.com/addison/outbox"
	testFederatedActivityIRI  = "https://other.example.com/activity/1"
	testFederatedActivityIRI2 = "https://other.example.com/activity/2"
	testFederatedActorIRI     = "https://other.example.com/dakota"
	testFederatedActorIRI2    = "https://other.example.com/addison"
	testFederatedActorIRI3    = "https://other.example.com/sam"
	testFederatedActorIRI4    = "https://other.example.com/jessie"
	testFederatedInboxIRI     = "https://other.example.com/dakota/inbox"
	testFederatedInboxIRI2    = "https://other.example.com/addison/inbox"
	testNoteId1               = "https://example.com/note/1"
	testNoteId2               = "https://example.com/note/2"
	testNewActivityIRI        = "https://example.com/new/1"
	testNewActivityIRI2       = "https://example.com/new/2"
	testNewActivityIRI3       = "https://example.com/new/3"
	testToIRI                 = "https://maybe.example.com/to/1"
	testToIRI2                = "https://maybe.example.com/to/2"
	testCcIRI                 = "https://maybe.example.com/cc/1"
	testCcIRI2                = "https://maybe.example.com/cc/2"
	testAudienceIRI           = "https://maybe.example.com/audience/1"
	testAudienceIRI2          = "https://maybe.example.com/audience/2"
	testPersonIRI             = "https://maybe.example.com/person"
	testServiceIRI            = "https://maybe.example.com/service"
	testTagIRI                = "https://example.com/tag/1"
	testTagIRI2               = "https://example.com/tag/2"
	inReplyToIRI              = "https://example.com/inReplyTo/1"
	inReplyToIRI2             = "https://example.com/inReplyTo/2"
)

// mustParse parses a URL or panics.
func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

// assertEqual ensures two values are equal.
func assertEqual(t *testing.T, a, b interface{}) {
	if a != b {
		t.Errorf("expected equal: %v != %v", a, b)
	}
}

// assertByteEqual ensures two byte slices are equal.
func assertByteEqual(t *testing.T, a, b []byte) {
	if string(a) != string(b) {
		t.Errorf("expected equal:\n%s\n\n%s", a, b)
	}
}

// assertNotEqual ensures two values are not equal.
func assertNotEqual(t *testing.T, a, b interface{}) {
	if a == b {
		t.Errorf("expected not equal: %v != %v", a, b)
	}
}

var (
	// testErr is a test error.
	testErr = errors.New("test error")
	// testFederatedNote is a test Note from a federated peer.
	testFederatedNote vocab.ActivityStreamsNote
	// testFederatedNote2 is a test Note from a federated peer.
	testFederatedNote2 vocab.ActivityStreamsNote
	// testMyNote is a test Note owned by this server.
	testMyNote vocab.ActivityStreamsNote
	// testMyNoteNoId is a test Note owned by this server.
	testMyNoteNoId vocab.ActivityStreamsNote
	// testMyCreate is a test Create Activity.
	testMyCreate vocab.ActivityStreamsCreate
	// testCreate is a test Create Activity.
	testCreate vocab.ActivityStreamsCreate
	// testCreate2 is a test Create Activity with two actors.
	testCreate2 vocab.ActivityStreamsCreate
	// testCreateNoId is a test Create Activity without an 'id' set.
	testCreateNoId vocab.ActivityStreamsCreate
	// testOrderedCollectionUniqueElems is a collection with only unique
	// ids.
	testOrderedCollectionUniqueElems vocab.ActivityStreamsOrderedCollectionPage
	// testOrderedCollectionUniqueElemsString is the JSON-LD version of the
	// testOrderedCollectionUniqueElems value
	testOrderedCollectionUniqueElemsString string
	// testOrderedCollectionDupedElems is a collection with duplicated ids.
	testOrderedCollectionDupedElems vocab.ActivityStreamsOrderedCollectionPage
	// testOrderedCollectionDedupedElemsString is the JSON-LD version of the
	// testOrderedCollectionDedupedElems value with duplicates removed
	testOrderedCollectionDedupedElemsString string
	// testEmptyOrderedCollection is an empty OrderedCollectionPage.
	testEmptyOrderedCollection vocab.ActivityStreamsOrderedCollectionPage
	// testOrderedCollectionWithNewId has the new id
	testOrderedCollectionWithNewId vocab.ActivityStreamsOrderedCollectionPage
	// testOrderedCollectionWithNewId has the second new id
	testOrderedCollectionWithNewId2 vocab.ActivityStreamsOrderedCollectionPage
	// testOrderedCollectionWithBothNewIds has both new ids.
	testOrderedCollectionWithBothNewIds vocab.ActivityStreamsOrderedCollectionPage
	// testOrderedCollectionWithFederatedId has the federated Activity id.
	testOrderedCollectionWithFederatedId vocab.ActivityStreamsOrderedCollectionPage
	// testMyListen is a test Listen C2S Activity.
	testMyListen vocab.ActivityStreamsListen
	// testMyListenNoId is a test Listen C2S Activity without an id.
	testMyListenNoId vocab.ActivityStreamsListen
	// testListen is a test Listen Activity.
	testListen vocab.ActivityStreamsListen
	// testOrderedCollectionWithFederatedId2 has the second federated
	// Activity id.
	testOrderedCollectionWithFederatedId2 vocab.ActivityStreamsOrderedCollectionPage
	// testOrderedCollectionWithBothFederatedIds has both federated Activity id.
	testOrderedCollectionWithBothFederatedIds vocab.ActivityStreamsOrderedCollectionPage
	// testPerson is a Person.
	testPerson vocab.ActivityStreamsPerson
	// testMyPerson is my Person.
	testMyPerson vocab.ActivityStreamsPerson
	// testFederatedPerson1 is a federated Person.
	testFederatedPerson1 vocab.ActivityStreamsPerson
	// testFederatedPerson2 is a federated Person.
	testFederatedPerson2 vocab.ActivityStreamsPerson
	// testService is a Service.
	testService vocab.ActivityStreamsService
	// testCollectionOfActors is a collection of actors.
	testCollectionOfActors vocab.ActivityStreamsCollectionPage
	// testOrderedCollectionOfActors is an ordered collection of actors.
	testOrderedCollectionOfActors vocab.ActivityStreamsOrderedCollectionPage
	// testNestedInReplyTo is an Activity with an 'object' with an 'inReplyTo'
	testNestedInReplyTo vocab.ActivityStreamsListen
	// testFollow is a test Follow Activity.
	testFollow vocab.ActivityStreamsFollow
	// testTombstone is a test Tombsone.
	testTombstone vocab.ActivityStreamsTombstone
)

// The test data cannot be created at init time since that is when the hooks of
// the `streams` package are set up. So initialize the data in this call instead
// of at init time.
func setupData() {
	// testFederatedNote
	func() {
		testFederatedNote = streams.NewActivityStreamsNote()
		name := streams.NewActivityStreamsNameProperty()
		name.AppendXMLSchemaString("A Federated Note")
		testFederatedNote.SetActivityStreamsName(name)
		content := streams.NewActivityStreamsContentProperty()
		content.AppendXMLSchemaString("This is a simple note being federated.")
		testFederatedNote.SetActivityStreamsContent(content)
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testNoteId1))
		testFederatedNote.SetJSONLDId(id)
	}()
	// testFederatedNote2
	func() {
		testFederatedNote2 = streams.NewActivityStreamsNote()
		name := streams.NewActivityStreamsNameProperty()
		name.AppendXMLSchemaString("A second federated note")
		testFederatedNote2.SetActivityStreamsName(name)
		content := streams.NewActivityStreamsContentProperty()
		content.AppendXMLSchemaString("This is a simple second note being federated.")
		testFederatedNote2.SetActivityStreamsContent(content)
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testNoteId2))
		testFederatedNote2.SetJSONLDId(id)
	}()
	// testMyNote
	func() {
		testMyNote = streams.NewActivityStreamsNote()
		name := streams.NewActivityStreamsNameProperty()
		name.AppendXMLSchemaString("My Note")
		testMyNote.SetActivityStreamsName(name)
		content := streams.NewActivityStreamsContentProperty()
		content.AppendXMLSchemaString("This is a simple note of mine.")
		testMyNote.SetActivityStreamsContent(content)
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testNoteId1))
		testMyNote.SetJSONLDId(id)
	}()
	// testMyNoteNoId
	func() {
		testMyNoteNoId = streams.NewActivityStreamsNote()
		name := streams.NewActivityStreamsNameProperty()
		name.AppendXMLSchemaString("My Note")
		testMyNoteNoId.SetActivityStreamsName(name)
		content := streams.NewActivityStreamsContentProperty()
		content.AppendXMLSchemaString("This is a simple note of mine.")
		testMyNoteNoId.SetActivityStreamsContent(content)
	}()
	// testMyCreate
	func() {
		testMyCreate = streams.NewActivityStreamsCreate()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testNewActivityIRI))
		testMyCreate.SetJSONLDId(id)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testMyNote)
		testMyCreate.SetActivityStreamsObject(op)
	}()
	// testCreate
	func() {
		testCreate = streams.NewActivityStreamsCreate()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		testCreate.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		testCreate.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testFederatedNote)
		testCreate.SetActivityStreamsObject(op)
	}()
	// testCreate2
	func() {
		testCreate2 = streams.NewActivityStreamsCreate()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		testCreate2.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		actor.AppendIRI(mustParse(testFederatedActorIRI2))
		testCreate2.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testFederatedNote)
		testCreate2.SetActivityStreamsObject(op)
	}()
	// testCreateNoId
	func() {
		testCreateNoId = streams.NewActivityStreamsCreate()
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		testCreateNoId.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testFederatedNote)
		testCreateNoId.SetActivityStreamsObject(op)
	}()
	// testOrderedCollectionUniqueElems and
	// testOrderedCollectionUniqueElemsString
	func() {
		testOrderedCollectionUniqueElems = streams.NewActivityStreamsOrderedCollectionPage()
		oi := streams.NewActivityStreamsOrderedItemsProperty()
		oi.AppendIRI(mustParse(testNoteId1))
		oi.AppendIRI(mustParse(testNoteId2))
		testOrderedCollectionUniqueElems.SetActivityStreamsOrderedItems(oi)
		testOrderedCollectionUniqueElemsString = `{"@context":"https://www.w3.org/ns/activitystreams","orderedItems":["https://example.com/note/1","https://example.com/note/2"],"type":"OrderedCollectionPage"}`
	}()
	// testOrderedCollectionDupedElems and
	// testOrderedCollectionDedupedElemsString
	func() {
		testOrderedCollectionDupedElems = streams.NewActivityStreamsOrderedCollectionPage()
		oi := streams.NewActivityStreamsOrderedItemsProperty()
		oi.AppendIRI(mustParse(testNoteId1))
		oi.AppendIRI(mustParse(testNoteId1))
		testOrderedCollectionDupedElems.SetActivityStreamsOrderedItems(oi)
		testOrderedCollectionDedupedElemsString = `{"@context":"https://www.w3.org/ns/activitystreams","orderedItems":"https://example.com/note/1","type":"OrderedCollectionPage"}`
	}()
	// testEmptyOrderedCollection
	func() {
		testEmptyOrderedCollection = streams.NewActivityStreamsOrderedCollectionPage()
	}()
	// testOrderedCollectionWithNewId
	func() {
		testOrderedCollectionWithNewId = streams.NewActivityStreamsOrderedCollectionPage()
		oi := streams.NewActivityStreamsOrderedItemsProperty()
		oi.AppendIRI(mustParse(testNewActivityIRI))
		testOrderedCollectionWithNewId.SetActivityStreamsOrderedItems(oi)
	}()
	// testOrderedCollectionWithNewId2
	func() {
		testOrderedCollectionWithNewId2 = streams.NewActivityStreamsOrderedCollectionPage()
		oi := streams.NewActivityStreamsOrderedItemsProperty()
		oi.AppendIRI(mustParse(testNewActivityIRI2))
		testOrderedCollectionWithNewId2.SetActivityStreamsOrderedItems(oi)
	}()
	// testOrderedCollectionWithBothNewIds
	func() {
		testOrderedCollectionWithBothNewIds = streams.NewActivityStreamsOrderedCollectionPage()
		oi := streams.NewActivityStreamsOrderedItemsProperty()
		oi.AppendIRI(mustParse(testNewActivityIRI))
		oi.AppendIRI(mustParse(testNewActivityIRI2))
		testOrderedCollectionWithBothNewIds.SetActivityStreamsOrderedItems(oi)
	}()
	// testOrderedCollectionWithFederatedId
	func() {
		testOrderedCollectionWithFederatedId = streams.NewActivityStreamsOrderedCollectionPage()
		oi := streams.NewActivityStreamsOrderedItemsProperty()
		oi.AppendIRI(mustParse(testFederatedActivityIRI))
		testOrderedCollectionWithFederatedId.SetActivityStreamsOrderedItems(oi)
	}()
	// testMyListen
	func() {
		testMyListen = streams.NewActivityStreamsListen()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testNewActivityIRI))
		testMyListen.SetJSONLDId(id)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testMyNote)
		testMyListen.SetActivityStreamsObject(op)
	}()
	// testMyListenNoId
	func() {
		testMyListenNoId = streams.NewActivityStreamsListen()
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testMyNoteNoId)
		testMyListenNoId.SetActivityStreamsObject(op)
	}()
	// testListen
	func() {
		testListen = streams.NewActivityStreamsListen()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		testListen.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		testListen.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsNote(testFederatedNote)
		testListen.SetActivityStreamsObject(op)
	}()
	// testOrderedCollectionWithFederatedId2
	func() {
		testOrderedCollectionWithFederatedId2 = streams.NewActivityStreamsOrderedCollectionPage()
		oi := streams.NewActivityStreamsOrderedItemsProperty()
		oi.AppendIRI(mustParse(testFederatedActivityIRI2))
		testOrderedCollectionWithFederatedId2.SetActivityStreamsOrderedItems(oi)
	}()
	// testOrderedCollectionWithBothFederatedIds
	func() {
		testOrderedCollectionWithBothFederatedIds = streams.NewActivityStreamsOrderedCollectionPage()
		oi := streams.NewActivityStreamsOrderedItemsProperty()
		oi.AppendIRI(mustParse(testFederatedActivityIRI))
		oi.AppendIRI(mustParse(testFederatedActivityIRI2))
		testOrderedCollectionWithBothFederatedIds.SetActivityStreamsOrderedItems(oi)
	}()
	// testPerson
	func() {
		testPerson = streams.NewActivityStreamsPerson()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testPersonIRI))
		testPerson.SetJSONLDId(id)
	}()
	// testMyPerson
	func() {
		testMyPerson = streams.NewActivityStreamsPerson()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testPersonIRI))
		testMyPerson.SetJSONLDId(id)
		inbox := streams.NewActivityStreamsInboxProperty()
		inbox.SetIRI(mustParse(testMyInboxIRI))
		testMyPerson.SetActivityStreamsInbox(inbox)
		outbox := streams.NewActivityStreamsOutboxProperty()
		outbox.SetIRI(mustParse(testMyOutboxIRI))
		testMyPerson.SetActivityStreamsOutbox(outbox)
	}()
	// testFederatedPerson1
	func() {
		testFederatedPerson1 = streams.NewActivityStreamsPerson()
		id := streams.NewJSONLDIdProperty()
		id.SetIRI(mustParse(testFederatedActorIRI))
		testFederatedPerson1.SetJSONLDId(id)
		inbox := streams.NewActivityStreamsInboxProperty()
		inbox.SetIRI(mustParse(testFederatedInboxIRI))
		testFederatedPerson1.SetActivityStreamsInbox(inbox)
	}()
	// testFederatedPerson2
	func() {
		testFederatedPerson2 = streams.NewActivityStreamsPerson()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActorIRI2))
		testFederatedPerson2.SetJSONLDId(id)
		inbox := streams.NewActivityStreamsInboxProperty()
		inbox.SetIRI(mustParse(testFederatedInboxIRI2))
		testFederatedPerson2.SetActivityStreamsInbox(inbox)
	}()
	// testService
	func() {
		testService = streams.NewActivityStreamsService()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testServiceIRI))
		testService.SetJSONLDId(id)
	}()
	// testCollectionOfActors
	func() {
		testCollectionOfActors = streams.NewActivityStreamsCollectionPage()
		i := streams.NewActivityStreamsItemsProperty()
		i.AppendIRI(mustParse(testFederatedActorIRI))
		i.AppendIRI(mustParse(testFederatedActorIRI2))
		testCollectionOfActors.SetActivityStreamsItems(i)
	}()
	// testOrderedCollectionOfActors
	func() {
		testOrderedCollectionOfActors = streams.NewActivityStreamsOrderedCollectionPage()
		oi := streams.NewActivityStreamsOrderedItemsProperty()
		oi.AppendIRI(mustParse(testFederatedActorIRI3))
		oi.AppendIRI(mustParse(testFederatedActorIRI4))
		testOrderedCollectionOfActors.SetActivityStreamsOrderedItems(oi)
	}()
	// testNestedInReplyTo
	func() {
		testNestedInReplyTo = streams.NewActivityStreamsListen()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		testNestedInReplyTo.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI))
		testNestedInReplyTo.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		// Note
		note := streams.NewActivityStreamsNote()
		name := streams.NewActivityStreamsNameProperty()
		name.AppendXMLSchemaString("A Federated Note")
		note.SetActivityStreamsName(name)
		content := streams.NewActivityStreamsContentProperty()
		content.AppendXMLSchemaString("This is a simple note being federated.")
		note.SetActivityStreamsContent(content)
		noteId := streams.NewJSONLDIdProperty()
		noteId.Set(mustParse(testNoteId1))
		note.SetJSONLDId(noteId)
		irt := streams.NewActivityStreamsInReplyToProperty()
		irt.AppendIRI(mustParse(inReplyToIRI))
		irt.AppendIRI(mustParse(inReplyToIRI2))
		note.SetActivityStreamsInReplyTo(irt)
		// Listen
		op.AppendActivityStreamsNote(note)
		testNestedInReplyTo.SetActivityStreamsObject(op)
	}()
	// testFollow
	func() {
		testFollow = streams.NewActivityStreamsFollow()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		testFollow.SetJSONLDId(id)
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(mustParse(testFederatedActorIRI2))
		testFollow.SetActivityStreamsActor(actor)
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendIRI(mustParse(testFederatedActorIRI))
		testFollow.SetActivityStreamsObject(op)
	}()
	// testTombstone
	func() {
		testTombstone = streams.NewActivityStreamsTombstone()
		id := streams.NewJSONLDIdProperty()
		id.Set(mustParse(testFederatedActivityIRI))
		testTombstone.SetJSONLDId(id)
	}()
}

// wrappedInCreate returns a Create activity wrapping the given type.
func wrappedInCreate(t vocab.Type) vocab.ActivityStreamsCreate {
	create := streams.NewActivityStreamsCreate()
	op := streams.NewActivityStreamsObjectProperty()
	op.AppendType(t)
	create.SetActivityStreamsObject(op)
	return create
}

// wrappedInResponse wraps a vocab.Type as serialized JSON response body with IRI.
func wrappedInResponse(iri *url.URL, t vocab.Type) *http.Response {
	r := new(http.Response)
	r.Request = new(http.Request)
	r.Request.URL = iri
	r.Status = http.StatusText(http.StatusOK)
	r.StatusCode = http.StatusOK
	body := bytes.NewReader(mustSerializeToBytes(t))
	r.Body = io.NopCloser(body)
	return r
}

// mustSerializeToBytes serializes a type to bytes or panics.
func mustSerializeToBytes(t vocab.Type) []byte {
	m := mustSerialize(t)
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

// mustSerialize serializes a type or panics.
func mustSerialize(t vocab.Type) map[string]interface{} {
	m, err := streams.Serialize(t)
	if err != nil {
		panic(err)
	}
	return m
}

// toDeserializedForm serializes and deserializes a type so that it works with
// mock expectations.
func toDeserializedForm(t vocab.Type) vocab.Type {
	m := mustSerialize(t)
	asValue, err := streams.ToType(context.Background(), m)
	if err != nil {
		panic(err)
	}
	return asValue
}

// withNewId sets a new id property on the activity
func withNewId(t vocab.Type) Activity {
	a, ok := t.(Activity)
	if !ok {
		panic("activity streams value is not an Activity")
	}
	id := streams.NewJSONLDIdProperty()
	id.Set(mustParse(testNewActivityIRI))
	a.SetJSONLDId(id)
	return a
}

// now returns the "current" time for tests.
func now() time.Time {
	l, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	return time.Date(2000, 2, 3, 4, 5, 6, 7, l)
}

// nowDateHeader returns the "current" time formatted in a form expected by the
// Date header in HTTP responses.
func nowDateHeader() string {
	return now().UTC().Format("Mon, 02 Jan 2006 15:04:05") + " GMT"
}

// toAPRequests adds the appropriate Content-Type or Accept headers to indicate
// that the HTTP request is an ActivityPub one. Also sets the Date header with
// the "current" test time.
func toAPRequest(r *http.Request) *http.Request {
	if r.Method == "POST" {
		existing, ok := r.Header[contentTypeHeader]
		if ok {
			r.Header[contentTypeHeader] = append(existing, activityStreamsMediaTypes[0])
		} else {
			r.Header[contentTypeHeader] = []string{activityStreamsMediaTypes[0]}
		}
	} else if r.Method == "GET" {
		existing, ok := r.Header[acceptHeader]
		if ok {
			r.Header[acceptHeader] = append(existing, activityStreamsMediaTypes[0])
		} else {
			r.Header[acceptHeader] = []string{activityStreamsMediaTypes[0]}
		}
	} else {
		panic("cannot toAPRequest with method " + r.Method)
	}
	r.Header[dateHeader] = []string{now().UTC().Format("Mon, 02 Jan 2006 15:04:05") + " GMT"}
	return r
}

// toPostInboxRequest creates a new POST HTTP request with the given type as
// the payload.
func toPostInboxRequest(t vocab.Type) *http.Request {
	m, err := streams.Serialize(t)
	if err != nil {
		panic(err)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer(b)
	return httptest.NewRequest("POST", testMyInboxIRI, buf)
}

// toPostOutboxRequest creates a new POST HTTP request with the given type as
// the payload.
func toPostOutboxRequest(t vocab.Type) *http.Request {
	m, err := streams.Serialize(t)
	if err != nil {
		panic(err)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer(b)
	return httptest.NewRequest("POST", testMyOutboxIRI, buf)
}

// toPostOutboxUnknownRequest creates a new POST HTTP request with an unknown
// type in the payload.
func toPostOutboxUnknownRequest() *http.Request {
	s := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "http://www.types.example/ProductOffer",
  "id": "http://www.example.com/spam"
}`
	b := []byte(s)
	buf := bytes.NewBuffer(b)
	return httptest.NewRequest("POST", testMyOutboxIRI, buf)
}

// toPostInboxUnknownRequest creates a new POST HTTP request with an unknown
// type in the payload.
func toPostInboxUnknownRequest() *http.Request {
	s := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "http://www.types.example/ProductOffer",
  "id": "http://www.example.com/spam"
}`
	b := []byte(s)
	buf := bytes.NewBuffer(b)
	return httptest.NewRequest("POST", testMyInboxIRI, buf)
}

// toGetInboxRequest creates a new GET HTTP request.
func toGetInboxRequest() *http.Request {
	return httptest.NewRequest("GET", testMyInboxIRI, nil)
}

// toGetOutboxRequest creates a new GET HTTP request.
func toGetOutboxRequest() *http.Request {
	return httptest.NewRequest("GET", testMyOutboxIRI, nil)
}

// addToIds adds two IRIs to the 'to' property
func addToIds(t Activity) Activity {
	to := streams.NewActivityStreamsToProperty()
	to.AppendIRI(mustParse(testToIRI))
	to.AppendIRI(mustParse(testToIRI2))
	t.SetActivityStreamsTo(to)
	return t
}

// mustAddCcIds adds two IRIs to the 'cc' property
func mustAddCcIds(t Activity) Activity {
	if ccer, ok := t.(ccer); ok {
		cc := streams.NewActivityStreamsCcProperty()
		cc.AppendIRI(mustParse(testCcIRI))
		cc.AppendIRI(mustParse(testCcIRI2))
		ccer.SetActivityStreamsCc(cc)
	} else {
		panic(fmt.Sprintf("activity is not ccer: %T", t))
	}
	return t
}

// mustAddAudienceIds adds two IRIs to the 'audience' property
func mustAddAudienceIds(t Activity) Activity {
	if audiencer, ok := t.(audiencer); ok {
		aud := streams.NewActivityStreamsAudienceProperty()
		aud.AppendIRI(mustParse(testAudienceIRI))
		aud.AppendIRI(mustParse(testAudienceIRI2))
		audiencer.SetActivityStreamsAudience(aud)
	} else {
		panic(fmt.Sprintf("activity is not audiencer: %T", t))
	}
	return t
}

// setTagger is an ActivityStreams type with a 'tag' property
type setTagger interface {
	SetActivityStreamsTag(vocab.ActivityStreamsTagProperty)
}

// mustAddTagIds adds two IRIs to the 'tag' property
func mustAddTagIds(t Activity) Activity {
	if st, ok := t.(setTagger); ok {
		tag := streams.NewActivityStreamsTagProperty()
		tag.AppendIRI(mustParse(testTagIRI))
		tag.AppendIRI(mustParse(testTagIRI2))
		st.SetActivityStreamsTag(tag)
	} else {
		panic(fmt.Sprintf("activity is not setTagger: %T", t))
	}
	return t
}

// setInReplyToer is an ActivityStreams type with a 'inReplyTo' property
type setInReplyToer interface {
	SetActivityStreamsInReplyTo(vocab.ActivityStreamsInReplyToProperty)
}

// mustAddInReplyToIds adds two IRIs to the 'inReplyTo' property
func mustAddInReplyToIds(t Activity) Activity {
	if st, ok := t.(setInReplyToer); ok {
		irt := streams.NewActivityStreamsInReplyToProperty()
		irt.AppendIRI(mustParse(inReplyToIRI))
		irt.AppendIRI(mustParse(inReplyToIRI2))
		st.SetActivityStreamsInReplyTo(irt)
	} else {
		panic(fmt.Sprintf("activity is not setInReplyToer: %T", t))
	}
	return t
}

// newObjectWithId creates a generic object with a given id.
func newObjectWithId(id string) vocab.ActivityStreamsObject {
	obj := streams.NewActivityStreamsObject()
	i := streams.NewJSONLDIdProperty()
	i.Set(mustParse(id))
	obj.SetJSONLDId(i)
	return obj
}

// newActivityWithId creates a generic Activity with a given id.
func newActivityWithId(id string) vocab.ActivityStreamsActivity {
	a := streams.NewActivityStreamsActivity()
	i := streams.NewJSONLDIdProperty()
	i.Set(mustParse(id))
	a.SetJSONLDId(i)
	return a
}
