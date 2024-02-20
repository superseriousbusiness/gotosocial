package streams

import (
	"net/url"
	"time"

	"github.com/superseriousbusiness/activity/streams/vocab"
)

func MustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

const example1 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Object",
  "id": "http://www.test.example/object/1",
  "name": "A Simple, non-specific object"
}`

func example1Type() vocab.ActivityStreamsObject {
	example1Type := NewActivityStreamsObject()
	id := NewJSONLDIdProperty()
	id.Set(MustParseURL("http://www.test.example/object/1"))
	example1Type.SetJSONLDId(id)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("A Simple, non-specific object")
	example1Type.SetActivityStreamsName(name)
	return example1Type
}

const example2 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Link",
  "href": "http://example.org/abc",
  "hreflang": "en",
  "mediaType": "text/html",
  "name": "An example link"
}`

func example2Type() vocab.ActivityStreamsLink {
	example2Type := NewActivityStreamsLink()
	hrefUrl := MustParseURL("http://example.org/abc")
	href := NewActivityStreamsHrefProperty()
	href.Set(hrefUrl)
	example2Type.SetActivityStreamsHref(href)
	hrefLang := NewActivityStreamsHreflangProperty()
	hrefLang.Set("en")
	example2Type.SetActivityStreamsHreflang(hrefLang)
	mediaType := NewActivityStreamsMediaTypeProperty()
	mediaType.Set("text/html")
	example2Type.SetActivityStreamsMediaType(mediaType)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("An example link")
	example2Type.SetActivityStreamsName(name)
	return example2Type
}

const example3 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Activity",
  "summary": "Sally did something to a note",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Note",
    "name": "A Note"
  }
}`

func example3Type() vocab.ActivityStreamsActivity {
	example3Type := NewActivityStreamsActivity()
	person := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	note := NewActivityStreamsNote()
	aNote := NewActivityStreamsNameProperty()
	aNote.AppendXMLSchemaString("A Note")
	note.SetActivityStreamsName(aNote)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally did something to a note")
	example3Type.SetActivityStreamsSummary(summary)
	actor := NewActivityStreamsActorProperty()
	actor.AppendActivityStreamsPerson(person)
	example3Type.SetActivityStreamsActor(actor)
	object := NewActivityStreamsObjectProperty()
	object.AppendActivityStreamsNote(note)
	example3Type.SetActivityStreamsObject(object)
	return example3Type
}

const example4 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Travel",
  "summary": "Sally went to work",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "target": {
    "type": "Place",
    "name": "Work"
  }
}`

func example4Type() vocab.ActivityStreamsTravel {
	example4Type := NewActivityStreamsTravel()
	person := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	place := NewActivityStreamsPlace()
	work := NewActivityStreamsNameProperty()
	work.AppendXMLSchemaString("Work")
	place.SetActivityStreamsName(work)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally went to work")
	example4Type.SetActivityStreamsSummary(summary)
	actor := NewActivityStreamsActorProperty()
	actor.AppendActivityStreamsPerson(person)
	example4Type.SetActivityStreamsActor(actor)
	target := NewActivityStreamsTargetProperty()
	target.AppendActivityStreamsPlace(place)
	example4Type.SetActivityStreamsTarget(target)
	return example4Type
}

const example5 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally's notes",
  "type": "Collection",
  "totalItems": 2,
  "items": [
    {
      "type": "Note",
      "name": "A Simple Note"
    },
    {
      "type": "Note",
      "name": "Another Simple Note"
    }
  ]
}`

func example5Type() vocab.ActivityStreamsCollection {
	example5Type := NewActivityStreamsCollection()
	note1 := NewActivityStreamsNote()
	name1 := NewActivityStreamsNameProperty()
	name1.AppendXMLSchemaString("A Simple Note")
	note1.SetActivityStreamsName(name1)
	note2 := NewActivityStreamsNote()
	name2 := NewActivityStreamsNameProperty()
	name2.AppendXMLSchemaString("Another Simple Note")
	note2.SetActivityStreamsName(name2)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally's notes")
	example5Type.SetActivityStreamsSummary(summary)
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(2)
	example5Type.SetActivityStreamsTotalItems(totalItems)
	items := NewActivityStreamsItemsProperty()
	items.AppendActivityStreamsNote(note1)
	items.AppendActivityStreamsNote(note2)
	example5Type.SetActivityStreamsItems(items)
	return example5Type
}

const example6 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally's notes",
  "type": "OrderedCollection",
  "totalItems": 2,
  "orderedItems": [
    {
      "type": "Note",
      "name": "A Simple Note"
    },
    {
      "type": "Note",
      "name": "Another Simple Note"
    }
  ]
}`

func example6Type() vocab.ActivityStreamsOrderedCollection {
	example6Type := NewActivityStreamsOrderedCollection()
	note1 := NewActivityStreamsNote()
	name1 := NewActivityStreamsNameProperty()
	name1.AppendXMLSchemaString("A Simple Note")
	note1.SetActivityStreamsName(name1)
	note2 := NewActivityStreamsNote()
	name2 := NewActivityStreamsNameProperty()
	name2.AppendXMLSchemaString("Another Simple Note")
	note2.SetActivityStreamsName(name2)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally's notes")
	example6Type.SetActivityStreamsSummary(summary)
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(2)
	example6Type.SetActivityStreamsTotalItems(totalItems)
	orderedItems := NewActivityStreamsOrderedItemsProperty()
	orderedItems.AppendActivityStreamsNote(note1)
	orderedItems.AppendActivityStreamsNote(note2)
	example6Type.SetActivityStreamsOrderedItems(orderedItems)
	return example6Type
}

const example7 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Page 1 of Sally's notes",
  "type": "CollectionPage",
  "id": "http://example.org/foo?page=1",
  "partOf": "http://example.org/foo",
  "items": [
    {
      "type": "Note",
      "name": "A Simple Note"
    },
    {
      "type": "Note",
      "name": "Another Simple Note"
    }
  ]
}`

func example7Type() vocab.ActivityStreamsCollectionPage {
	example7Type := NewActivityStreamsCollectionPage()
	note1 := NewActivityStreamsNote()
	name1 := NewActivityStreamsNameProperty()
	name1.AppendXMLSchemaString("A Simple Note")
	note1.SetActivityStreamsName(name1)
	note2 := NewActivityStreamsNote()
	name2 := NewActivityStreamsNameProperty()
	name2.AppendXMLSchemaString("Another Simple Note")
	note2.SetActivityStreamsName(name2)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Page 1 of Sally's notes")
	example7Type.SetActivityStreamsSummary(summary)
	id := NewJSONLDIdProperty()
	id.Set(MustParseURL("http://example.org/foo?page=1"))
	example7Type.SetJSONLDId(id)
	partOf := NewActivityStreamsPartOfProperty()
	partOf.SetIRI(MustParseURL("http://example.org/foo"))
	example7Type.SetActivityStreamsPartOf(partOf)
	items := NewActivityStreamsItemsProperty()
	items.AppendActivityStreamsNote(note1)
	items.AppendActivityStreamsNote(note2)
	example7Type.SetActivityStreamsItems(items)
	return example7Type
}

const example8 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Page 1 of Sally's notes",
  "type": "OrderedCollectionPage",
  "id": "http://example.org/foo?page=1",
  "partOf": "http://example.org/foo",
  "orderedItems": [
    {
      "type": "Note",
      "name": "A Simple Note"
    },
    {
      "type": "Note",
      "name": "Another Simple Note"
    }
  ]
}`

func example8Type() vocab.ActivityStreamsOrderedCollectionPage {
	example8Type := NewActivityStreamsOrderedCollectionPage()
	note1 := NewActivityStreamsNote()
	name1 := NewActivityStreamsNameProperty()
	name1.AppendXMLSchemaString("A Simple Note")
	note1.SetActivityStreamsName(name1)
	note2 := NewActivityStreamsNote()
	name2 := NewActivityStreamsNameProperty()
	name2.AppendXMLSchemaString("Another Simple Note")
	note2.SetActivityStreamsName(name2)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Page 1 of Sally's notes")
	example8Type.SetActivityStreamsSummary(summary)
	id := NewJSONLDIdProperty()
	id.Set(MustParseURL("http://example.org/foo?page=1"))
	example8Type.SetJSONLDId(id)
	partOf := NewActivityStreamsPartOfProperty()
	partOf.SetIRI(MustParseURL("http://example.org/foo"))
	example8Type.SetActivityStreamsPartOf(partOf)
	orderedItems := NewActivityStreamsOrderedItemsProperty()
	orderedItems.AppendActivityStreamsNote(note1)
	orderedItems.AppendActivityStreamsNote(note2)
	example8Type.SetActivityStreamsOrderedItems(orderedItems)
	return example8Type
}

const example9 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally accepted an invitation to a party",
  "type": "Accept",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Invite",
    "actor": "http://john.example.org",
    "object": {
      "type": "Event",
      "name": "Going-Away Party for Jim"
    }
  }
}`

func example9Type() vocab.ActivityStreamsAccept {
	example9Type := NewActivityStreamsAccept()
	person := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	event := NewActivityStreamsEvent()
	goingAway := NewActivityStreamsNameProperty()
	goingAway.AppendXMLSchemaString("Going-Away Party for Jim")
	event.SetActivityStreamsName(goingAway)
	invite := NewActivityStreamsInvite()
	actor := NewActivityStreamsActorProperty()
	actor.AppendIRI(MustParseURL("http://john.example.org"))
	invite.SetActivityStreamsActor(actor)
	object := NewActivityStreamsObjectProperty()
	object.AppendActivityStreamsEvent(event)
	invite.SetActivityStreamsObject(object)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally accepted an invitation to a party")
	example9Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(person)
	example9Type.SetActivityStreamsActor(rootActor)
	inviteObject := NewActivityStreamsObjectProperty()
	inviteObject.AppendActivityStreamsInvite(invite)
	example9Type.SetActivityStreamsObject(inviteObject)
	return example9Type
}

const example10 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally accepted Joe into the club",
  "type": "Accept",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Person",
    "name": "Joe"
  },
  "target": {
    "type": "Group",
    "name": "The Club"
  }
}`

func example10Type() vocab.ActivityStreamsAccept {
	example10Type := NewActivityStreamsAccept()
	person := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	object := NewActivityStreamsPerson()
	joe := NewActivityStreamsNameProperty()
	joe.AppendXMLSchemaString("Joe")
	object.SetActivityStreamsName(joe)
	target := NewActivityStreamsGroup()
	club := NewActivityStreamsNameProperty()
	club.AppendXMLSchemaString("The Club")
	target.SetActivityStreamsName(club)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally accepted Joe into the club")
	example10Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(person)
	example10Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsPerson(object)
	example10Type.SetActivityStreamsObject(obj)
	tobj := NewActivityStreamsTargetProperty()
	tobj.AppendActivityStreamsGroup(target)
	example10Type.SetActivityStreamsTarget(tobj)
	return example10Type
}

const example11 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally tentatively accepted an invitation to a party",
  "type": "TentativeAccept",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Invite",
    "actor": "http://john.example.org",
    "object": {
      "type": "Event",
      "name": "Going-Away Party for Jim"
    }
  }
}`

func example11Type() vocab.ActivityStreamsTentativeAccept {
	example11Type := NewActivityStreamsTentativeAccept()
	person := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	event := NewActivityStreamsEvent()
	goingAway := NewActivityStreamsNameProperty()
	goingAway.AppendXMLSchemaString("Going-Away Party for Jim")
	event.SetActivityStreamsName(goingAway)
	invite := NewActivityStreamsInvite()
	inviteActor := NewActivityStreamsActorProperty()
	inviteActor.AppendIRI(MustParseURL("http://john.example.org"))
	invite.SetActivityStreamsActor(inviteActor)
	objInvite := NewActivityStreamsObjectProperty()
	objInvite.AppendActivityStreamsEvent(event)
	invite.SetActivityStreamsObject(objInvite)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally tentatively accepted an invitation to a party")
	example11Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(person)
	example11Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsInvite(invite)
	example11Type.SetActivityStreamsObject(obj)
	return example11Type
}

const example12 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally added an object",
  "type": "Add",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": "http://example.org/abc"
}`

func example12Type() vocab.ActivityStreamsAdd {
	example12Type := NewActivityStreamsAdd()
	person := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	link := MustParseURL("http://example.org/abc")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally added an object")
	example12Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(person)
	example12Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(link)
	example12Type.SetActivityStreamsObject(obj)
	return example12Type
}

const example13 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally added a picture of her cat to her cat picture collection",
  "type": "Add",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Image",
    "name": "A picture of my cat",
    "url": "http://example.org/img/cat.png"
  },
  "origin": {
    "type": "Collection",
    "name": "Camera Roll"
  },
  "target": {
    "type": "Collection",
    "name": "My Cat Pictures"
  }
}`

func example13Type() vocab.ActivityStreamsAdd {
	example13Type := NewActivityStreamsAdd()
	person := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	link := MustParseURL("http://example.org/img/cat.png")
	object := NewActivityStreamsImage()
	objectName := NewActivityStreamsNameProperty()
	objectName.AppendXMLSchemaString("A picture of my cat")
	object.SetActivityStreamsName(objectName)
	urlProp := NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(link)
	object.SetActivityStreamsUrl(urlProp)
	origin := NewActivityStreamsCollection()
	originName := NewActivityStreamsNameProperty()
	originName.AppendXMLSchemaString("Camera Roll")
	origin.SetActivityStreamsName(originName)
	target := NewActivityStreamsCollection()
	targetName := NewActivityStreamsNameProperty()
	targetName.AppendXMLSchemaString("My Cat Pictures")
	target.SetActivityStreamsName(targetName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally added a picture of her cat to her cat picture collection")
	example13Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(person)
	example13Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsImage(object)
	example13Type.SetActivityStreamsObject(obj)
	originProp := NewActivityStreamsOriginProperty()
	originProp.AppendActivityStreamsCollection(origin)
	example13Type.SetActivityStreamsOrigin(originProp)
	tobj := NewActivityStreamsTargetProperty()
	tobj.AppendActivityStreamsCollection(target)
	example13Type.SetActivityStreamsTarget(tobj)
	return example13Type
}

const example14 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally arrived at work",
  "type": "Arrive",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "location": {
    "type": "Place",
    "name": "Work"
  },
  "origin": {
    "type": "Place",
    "name": "Home"
  }
}`

func example14Type() vocab.ActivityStreamsArrive {
	example14Type := NewActivityStreamsArrive()
	person := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	location := NewActivityStreamsPlace()
	locationName := NewActivityStreamsNameProperty()
	locationName.AppendXMLSchemaString("Work")
	location.SetActivityStreamsName(locationName)
	origin := NewActivityStreamsPlace()
	originName := NewActivityStreamsNameProperty()
	originName.AppendXMLSchemaString("Home")
	origin.SetActivityStreamsName(originName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally arrived at work")
	example14Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(person)
	example14Type.SetActivityStreamsActor(rootActor)
	loc := NewActivityStreamsLocationProperty()
	loc.AppendActivityStreamsPlace(location)
	example14Type.SetActivityStreamsLocation(loc)
	originProp := NewActivityStreamsOriginProperty()
	originProp.AppendActivityStreamsPlace(origin)
	example14Type.SetActivityStreamsOrigin(originProp)
	return example14Type
}

const example15 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally created a note",
  "type": "Create",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Note",
    "name": "A Simple Note",
    "content": "This is a simple note"
  }
}`

func example15Type() vocab.ActivityStreamsCreate {
	example15Type := NewActivityStreamsCreate()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	object := NewActivityStreamsNote()
	objectName := NewActivityStreamsNameProperty()
	objectName.AppendXMLSchemaString("A Simple Note")
	object.SetActivityStreamsName(objectName)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("This is a simple note")
	object.SetActivityStreamsContent(content)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally created a note")
	example15Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example15Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsNote(object)
	example15Type.SetActivityStreamsObject(obj)
	return example15Type
}

const example16 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally deleted a note",
  "type": "Delete",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": "http://example.org/notes/1",
  "origin": {
    "type": "Collection",
    "name": "Sally's Notes"
  }
}`

func example16Type() vocab.ActivityStreamsDelete {
	example16Type := NewActivityStreamsDelete()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	origin := NewActivityStreamsCollection()
	originName := NewActivityStreamsNameProperty()
	originName.AppendXMLSchemaString("Sally's Notes")
	origin.SetActivityStreamsName(originName)
	link := MustParseURL("http://example.org/notes/1")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally deleted a note")
	example16Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example16Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(link)
	example16Type.SetActivityStreamsObject(obj)
	originProp := NewActivityStreamsOriginProperty()
	originProp.AppendActivityStreamsCollection(origin)
	example16Type.SetActivityStreamsOrigin(originProp)
	return example16Type
}

const example17 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally followed John",
  "type": "Follow",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Person",
    "name": "John"
  }
}`

func example17Type() vocab.ActivityStreamsFollow {
	example17Type := NewActivityStreamsFollow()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	object := NewActivityStreamsPerson()
	objectName := NewActivityStreamsNameProperty()
	objectName.AppendXMLSchemaString("John")
	object.SetActivityStreamsName(objectName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally followed John")
	example17Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example17Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsPerson(object)
	example17Type.SetActivityStreamsObject(obj)
	return example17Type
}

const example18 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally ignored a note",
  "type": "Ignore",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": "http://example.org/notes/1"
}`

func example18Type() vocab.ActivityStreamsIgnore {
	example18Type := NewActivityStreamsIgnore()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	link := MustParseURL("http://example.org/notes/1")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally ignored a note")
	example18Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example18Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(link)
	example18Type.SetActivityStreamsObject(obj)
	return example18Type
}

const example19 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally joined a group",
  "type": "Join",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Group",
    "name": "A Simple Group"
  }
}`

func example19Type() vocab.ActivityStreamsJoin {
	example19Type := NewActivityStreamsJoin()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	object := NewActivityStreamsGroup()
	objectName := NewActivityStreamsNameProperty()
	objectName.AppendXMLSchemaString("A Simple Group")
	object.SetActivityStreamsName(objectName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally joined a group")
	example19Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example19Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsGroup(object)
	example19Type.SetActivityStreamsObject(obj)
	return example19Type
}

const example20 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally left work",
  "type": "Leave",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Place",
    "name": "Work"
  }
}`

func example20Type() vocab.ActivityStreamsLeave {
	example20Type := NewActivityStreamsLeave()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	object := NewActivityStreamsPlace()
	objectName := NewActivityStreamsNameProperty()
	objectName.AppendXMLSchemaString("Work")
	object.SetActivityStreamsName(objectName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally left work")
	example20Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example20Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsPlace(object)
	example20Type.SetActivityStreamsObject(obj)
	return example20Type
}

const example21 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally left a group",
  "type": "Leave",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Group",
    "name": "A Simple Group"
  }
}`

func example21Type() vocab.ActivityStreamsLeave {
	example21Type := NewActivityStreamsLeave()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	object := NewActivityStreamsGroup()
	objectName := NewActivityStreamsNameProperty()
	objectName.AppendXMLSchemaString("A Simple Group")
	object.SetActivityStreamsName(objectName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally left a group")
	example21Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example21Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsGroup(object)
	example21Type.SetActivityStreamsObject(obj)
	return example21Type
}

const example22 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally liked a note",
  "type": "Like",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": "http://example.org/notes/1"
}`

func example22Type() vocab.ActivityStreamsLike {
	example22Type := NewActivityStreamsLike()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	link := MustParseURL("http://example.org/notes/1")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally liked a note")
	example22Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example22Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(link)
	example22Type.SetActivityStreamsObject(obj)
	return example22Type
}

const example23 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally offered 50% off to Lewis",
  "type": "Offer",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "http://www.types.example/ProductOffer",
    "name": "50% Off!"
  },
  "target": {
    "type": "Person",
    "name": "Lewis"
  }
}`

var example23Unknown = func(m map[string]interface{}) map[string]interface{} {
	m["object"] = map[string]interface{}{
		"type": "http://www.types.example/ProductOffer",
		"name": "50% Off!",
	}
	return m
}

func example23Type() vocab.ActivityStreamsOffer {
	example23Type := NewActivityStreamsOffer()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	target := NewActivityStreamsPerson()
	targetName := NewActivityStreamsNameProperty()
	targetName.AppendXMLSchemaString("Lewis")
	target.SetActivityStreamsName(targetName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally offered 50% off to Lewis")
	example23Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example23Type.SetActivityStreamsActor(rootActor)
	tobj := NewActivityStreamsTargetProperty()
	tobj.AppendActivityStreamsPerson(target)
	example23Type.SetActivityStreamsTarget(tobj)
	return example23Type
}

const example24 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally invited John and Lisa to a party",
  "type": "Invite",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Event",
    "name": "A Party"
  },
  "target": [
    {
      "type": "Person",
      "name": "John"
    },
    {
      "type": "Person",
      "name": "Lisa"
    }
  ]
}`

func example24Type() vocab.ActivityStreamsInvite {
	example24Type := NewActivityStreamsInvite()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	object := NewActivityStreamsEvent()
	objectName := NewActivityStreamsNameProperty()
	objectName.AppendXMLSchemaString("A Party")
	object.SetActivityStreamsName(objectName)
	target1 := NewActivityStreamsPerson()
	target1Name := NewActivityStreamsNameProperty()
	target1Name.AppendXMLSchemaString("John")
	target1.SetActivityStreamsName(target1Name)
	target2 := NewActivityStreamsPerson()
	target2Name := NewActivityStreamsNameProperty()
	target2Name.AppendXMLSchemaString("Lisa")
	target2.SetActivityStreamsName(target2Name)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally invited John and Lisa to a party")
	example24Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example24Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsEvent(object)
	example24Type.SetActivityStreamsObject(obj)
	tobj := NewActivityStreamsTargetProperty()
	tobj.AppendActivityStreamsPerson(target1)
	tobj.AppendActivityStreamsPerson(target2)
	example24Type.SetActivityStreamsTarget(tobj)
	return example24Type
}

const example25 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally rejected an invitation to a party",
  "type": "Reject",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Invite",
    "actor": "http://john.example.org",
    "object": {
      "type": "Event",
      "name": "Going-Away Party for Jim"
    }
  }
}`

func example25Type() vocab.ActivityStreamsReject {
	example25Type := NewActivityStreamsReject()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	inviteObject := NewActivityStreamsEvent()
	goingAway := NewActivityStreamsNameProperty()
	goingAway.AppendXMLSchemaString("Going-Away Party for Jim")
	inviteObject.SetActivityStreamsName(goingAway)
	object := NewActivityStreamsInvite()
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(MustParseURL("http://john.example.org"))
	object.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsEvent(inviteObject)
	object.SetActivityStreamsObject(obj)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally rejected an invitation to a party")
	example25Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example25Type.SetActivityStreamsActor(rootActor)
	objRoot := NewActivityStreamsObjectProperty()
	objRoot.AppendActivityStreamsInvite(object)
	example25Type.SetActivityStreamsObject(objRoot)
	return example25Type
}

const example26 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally tentatively rejected an invitation to a party",
  "type": "TentativeReject",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Invite",
    "actor": "http://john.example.org",
    "object": {
      "type": "Event",
      "name": "Going-Away Party for Jim"
    }
  }
}`

func example26Type() vocab.ActivityStreamsTentativeReject {
	example26Type := NewActivityStreamsTentativeReject()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	inviteObject := NewActivityStreamsEvent()
	goingAway := NewActivityStreamsNameProperty()
	goingAway.AppendXMLSchemaString("Going-Away Party for Jim")
	inviteObject.SetActivityStreamsName(goingAway)
	object := NewActivityStreamsInvite()
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(MustParseURL("http://john.example.org"))
	object.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsEvent(inviteObject)
	object.SetActivityStreamsObject(obj)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally tentatively rejected an invitation to a party")
	example26Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example26Type.SetActivityStreamsActor(rootActor)
	objRoot := NewActivityStreamsObjectProperty()
	objRoot.AppendActivityStreamsInvite(object)
	example26Type.SetActivityStreamsObject(objRoot)
	return example26Type
}

const example27 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally removed a note from her notes folder",
  "type": "Remove",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": "http://example.org/notes/1",
  "target": {
    "type": "Collection",
    "name": "Notes Folder"
  }
}`

func example27Type() vocab.ActivityStreamsRemove {
	example27Type := NewActivityStreamsRemove()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	link := MustParseURL("http://example.org/notes/1")
	target := NewActivityStreamsCollection()
	targetName := NewActivityStreamsNameProperty()
	targetName.AppendXMLSchemaString("Notes Folder")
	target.SetActivityStreamsName(targetName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally removed a note from her notes folder")
	example27Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example27Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(link)
	example27Type.SetActivityStreamsObject(obj)
	tobj := NewActivityStreamsTargetProperty()
	tobj.AppendActivityStreamsCollection(target)
	example27Type.SetActivityStreamsTarget(tobj)
	return example27Type
}

const example28 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "The moderator removed Sally from a group",
  "type": "Remove",
  "actor": {
    "type": "http://example.org/Role",
    "name": "The Moderator"
  },
  "object": {
    "type": "Person",
    "name": "Sally"
  },
  "origin": {
    "type": "Group",
    "name": "A Simple Group"
  }
}`

var example28Unknown = func(m map[string]interface{}) map[string]interface{} {
	m["actor"] = map[string]interface{}{
		"type": "http://example.org/Role",
		"name": "The Moderator",
	}
	return m
}

func example28Type() vocab.ActivityStreamsRemove {
	example28Type := NewActivityStreamsRemove()
	object := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	object.SetActivityStreamsName(sally)
	origin := NewActivityStreamsGroup()
	originName := NewActivityStreamsNameProperty()
	originName.AppendXMLSchemaString("A Simple Group")
	origin.SetActivityStreamsName(originName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("The moderator removed Sally from a group")
	example28Type.SetActivityStreamsSummary(summary)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsPerson(object)
	example28Type.SetActivityStreamsObject(obj)
	originProp := NewActivityStreamsOriginProperty()
	originProp.AppendActivityStreamsGroup(origin)
	example28Type.SetActivityStreamsOrigin(originProp)
	return example28Type
}

const example29 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally retracted her offer to John",
  "type": "Undo",
  "actor": "http://sally.example.org",
  "object": {
    "type": "Offer",
    "actor": "http://sally.example.org",
    "object": "http://example.org/posts/1",
    "target": "http://john.example.org"
  }
}`

func example29Type() vocab.ActivityStreamsUndo {
	example29Type := NewActivityStreamsUndo()
	link := MustParseURL("http://sally.example.org")
	objectLink := MustParseURL("http://example.org/posts/1")
	targetLink := MustParseURL("http://john.example.org")
	object := NewActivityStreamsOffer()
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(link)
	object.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(objectLink)
	object.SetActivityStreamsObject(obj)
	target := NewActivityStreamsTargetProperty()
	target.AppendIRI(targetLink)
	object.SetActivityStreamsTarget(target)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally retracted her offer to John")
	example29Type.SetActivityStreamsSummary(summary)
	actor := NewActivityStreamsActorProperty()
	actor.AppendIRI(link)
	example29Type.SetActivityStreamsActor(actor)
	objRoot := NewActivityStreamsObjectProperty()
	objRoot.AppendActivityStreamsOffer(object)
	example29Type.SetActivityStreamsObject(objRoot)
	return example29Type
}

const example30 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally updated her note",
  "type": "Update",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": "http://example.org/notes/1"
}`

func example30Type() vocab.ActivityStreamsUpdate {
	example30Type := NewActivityStreamsUpdate()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	link := MustParseURL("http://example.org/notes/1")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally updated her note")
	example30Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example30Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(link)
	example30Type.SetActivityStreamsObject(obj)
	return example30Type
}

const example31 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally read an article",
  "type": "View",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Article",
    "name": "What You Should Know About Activity Streams"
  }
}`

func example31Type() vocab.ActivityStreamsView {
	example31Type := NewActivityStreamsView()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	object := NewActivityStreamsArticle()
	objectName := NewActivityStreamsNameProperty()
	objectName.AppendXMLSchemaString("What You Should Know About Activity Streams")
	object.SetActivityStreamsName(objectName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally read an article")
	example31Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example31Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsArticle(object)
	example31Type.SetActivityStreamsObject(obj)
	return example31Type
}

const example32 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally listened to a piece of music",
  "type": "Listen",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": "http://example.org/music.mp3"
}`

func example32Type() vocab.ActivityStreamsListen {
	example32Type := NewActivityStreamsListen()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	link := MustParseURL("http://example.org/music.mp3")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally listened to a piece of music")
	example32Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example32Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(link)
	example32Type.SetActivityStreamsObject(obj)
	return example32Type
}

const example33 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally read a blog post",
  "type": "Read",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": "http://example.org/posts/1"
}`

func example33Type() vocab.ActivityStreamsRead {
	example33Type := NewActivityStreamsRead()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	link := MustParseURL("http://example.org/posts/1")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally read a blog post")
	example33Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example33Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(link)
	example33Type.SetActivityStreamsObject(obj)
	return example33Type
}

const example34 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally moved a post from List A to List B",
  "type": "Move",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": "http://example.org/posts/1",
  "target": {
    "type": "Collection",
    "name": "List B"
  },
  "origin": {
    "type": "Collection",
    "name": "List A"
  }
}`

func example34Type() vocab.ActivityStreamsMove {
	example34Type := NewActivityStreamsMove()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	link := MustParseURL("http://example.org/posts/1")
	target := NewActivityStreamsCollection()
	targetName := NewActivityStreamsNameProperty()
	targetName.AppendXMLSchemaString("List B")
	target.SetActivityStreamsName(targetName)
	origin := NewActivityStreamsCollection()
	originName := NewActivityStreamsNameProperty()
	originName.AppendXMLSchemaString("List A")
	origin.SetActivityStreamsName(originName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally moved a post from List A to List B")
	example34Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example34Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(link)
	example34Type.SetActivityStreamsObject(obj)
	tobj := NewActivityStreamsTargetProperty()
	tobj.AppendActivityStreamsCollection(target)
	example34Type.SetActivityStreamsTarget(tobj)
	originProp := NewActivityStreamsOriginProperty()
	originProp.AppendActivityStreamsCollection(origin)
	example34Type.SetActivityStreamsOrigin(originProp)
	return example34Type
}

const example35 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally went home from work",
  "type": "Travel",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "target": {
    "type": "Place",
    "name": "Home"
  },
  "origin": {
    "type": "Place",
    "name": "Work"
  }
}`

func example35Type() vocab.ActivityStreamsTravel {
	example35Type := NewActivityStreamsTravel()
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	target := NewActivityStreamsPlace()
	targetName := NewActivityStreamsNameProperty()
	targetName.AppendXMLSchemaString("Home")
	target.SetActivityStreamsName(targetName)
	origin := NewActivityStreamsPlace()
	originName := NewActivityStreamsNameProperty()
	originName.AppendXMLSchemaString("Work")
	origin.SetActivityStreamsName(originName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally went home from work")
	example35Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example35Type.SetActivityStreamsActor(rootActor)
	tobj := NewActivityStreamsTargetProperty()
	tobj.AppendActivityStreamsPlace(target)
	example35Type.SetActivityStreamsTarget(tobj)
	originProp := NewActivityStreamsOriginProperty()
	originProp.AppendActivityStreamsPlace(origin)
	example35Type.SetActivityStreamsOrigin(originProp)
	return example35Type
}

const example36 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally announced that she had arrived at work",
  "type": "Announce",
  "actor": {
    "type": "Person",
    "id": "http://sally.example.org",
    "name": "Sally"
  },
  "object": {
    "type": "Arrive",
    "actor": "http://sally.example.org",
    "location": {
      "type": "Place",
      "name": "Work"
    }
  }
}`

func example36Type() vocab.ActivityStreamsAnnounce {
	example36Type := NewActivityStreamsAnnounce()
	link := MustParseURL("http://sally.example.org")
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	id := NewJSONLDIdProperty()
	id.Set(link)
	actor.SetJSONLDId(id)
	loc := NewActivityStreamsPlace()
	locName := NewActivityStreamsNameProperty()
	locName.AppendXMLSchemaString("Work")
	loc.SetActivityStreamsName(locName)
	object := NewActivityStreamsArrive()
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(link)
	object.SetActivityStreamsActor(objectActor)
	location := NewActivityStreamsLocationProperty()
	location.AppendActivityStreamsPlace(loc)
	object.SetActivityStreamsLocation(location)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally announced that she had arrived at work")
	example36Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example36Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsArrive(object)
	example36Type.SetActivityStreamsObject(obj)
	return example36Type
}

const example37 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally blocked Joe",
  "type": "Block",
  "actor": "http://sally.example.org",
  "object": "http://joe.example.org"
}`

func example37Type() vocab.ActivityStreamsBlock {
	example37Type := NewActivityStreamsBlock()
	link := MustParseURL("http://sally.example.org")
	objLink := MustParseURL("http://joe.example.org")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally blocked Joe")
	example37Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(link)
	example37Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(objLink)
	example37Type.SetActivityStreamsObject(obj)
	return example37Type
}

const example38 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally flagged an inappropriate note",
  "type": "Flag",
  "actor": "http://sally.example.org",
  "object": {
    "type": "Note",
    "content": "An inappropriate note"
  }
}`

func example38Type() vocab.ActivityStreamsFlag {
	example38Type := NewActivityStreamsFlag()
	link := MustParseURL("http://sally.example.org")
	object := NewActivityStreamsNote()
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("An inappropriate note")
	object.SetActivityStreamsContent(content)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally flagged an inappropriate note")
	example38Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(link)
	example38Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsNote(object)
	example38Type.SetActivityStreamsObject(obj)
	return example38Type
}

const example39 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally disliked a post",
  "type": "Dislike",
  "actor": "http://sally.example.org",
  "object": "http://example.org/posts/1"
}`

func example39Type() vocab.ActivityStreamsDislike {
	example39Type := NewActivityStreamsDislike()
	link := MustParseURL("http://sally.example.org")
	objLink := MustParseURL("http://example.org/posts/1")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally disliked a post")
	example39Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(link)
	example39Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(objLink)
	example39Type.SetActivityStreamsObject(obj)
	return example39Type
}

const example40 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Question",
  "name": "What is the answer?",
  "oneOf": [
    {
      "type": "Note",
      "name": "Option A"
    },
    {
      "type": "Note",
      "name": "Option B"
    }
  ]
}`

func example40Type() vocab.ActivityStreamsQuestion {
	example40Type := NewActivityStreamsQuestion()
	note1 := NewActivityStreamsNote()
	note1Name := NewActivityStreamsNameProperty()
	note1Name.AppendXMLSchemaString("Option A")
	note1.SetActivityStreamsName(note1Name)
	note2 := NewActivityStreamsNote()
	note2Name := NewActivityStreamsNameProperty()
	note2Name.AppendXMLSchemaString("Option B")
	note2.SetActivityStreamsName(note2Name)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("What is the answer?")
	example40Type.SetActivityStreamsName(name)
	oneOf := NewActivityStreamsOneOfProperty()
	oneOf.AppendActivityStreamsNote(note1)
	oneOf.AppendActivityStreamsNote(note2)
	example40Type.SetActivityStreamsOneOf(oneOf)
	return example40Type
}

const example41 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Question",
  "name": "What is the answer?",
  "closed": "2016-05-10T00:00:00Z"
}`

func example41Type() vocab.ActivityStreamsQuestion {
	example41Type := NewActivityStreamsQuestion()
	t, err := time.Parse(time.RFC3339, "2016-05-10T00:00:00Z")
	if err != nil {
		panic(err)
	}
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("What is the answer?")
	example41Type.SetActivityStreamsName(name)
	closed := NewActivityStreamsClosedProperty()
	closed.AppendXMLSchemaDateTime(t)
	example41Type.SetActivityStreamsClosed(closed)
	return example41Type
}

const example42 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Application",
  "name": "Exampletron 3000"
}`

func example42Type() vocab.ActivityStreamsApplication {
	example42Type := NewActivityStreamsApplication()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Exampletron 3000")
	example42Type.SetActivityStreamsName(name)
	return example42Type
}

const example43 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Group",
  "name": "Big Beards of Austin"
}`

func example43Type() vocab.ActivityStreamsGroup {
	example43Type := NewActivityStreamsGroup()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Big Beards of Austin")
	example43Type.SetActivityStreamsName(name)
	return example43Type
}

const example44 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Organization",
  "name": "Example Co."
}`

func example44Type() vocab.ActivityStreamsOrganization {
	example44Type := NewActivityStreamsOrganization()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Example Co.")
	example44Type.SetActivityStreamsName(name)
	return example44Type
}

const example45 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Person",
  "name": "Sally Smith"
}`

func example45Type() vocab.ActivityStreamsPerson {
	example45Type := NewActivityStreamsPerson()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Sally Smith")
	example45Type.SetActivityStreamsName(name)
	return example45Type
}

const example46 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Service",
  "name": "Acme Web Service"
}`

func example46Type() vocab.ActivityStreamsService {
	example46Type := NewActivityStreamsService()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Acme Web Service")
	example46Type.SetActivityStreamsName(name)
	return example46Type
}

const example47 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally is an acquaintance of John",
  "type": "Relationship",
  "subject": {
    "type": "Person",
    "name": "Sally"
  },
  "relationship": "http://purl.org/vocab/relationship/acquaintanceOf",
  "object": {
    "type": "Person",
    "name": "John"
  }
}`

func example47Type() vocab.ActivityStreamsRelationship {
	example47Type := NewActivityStreamsRelationship()
	subject := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	subject.SetActivityStreamsName(sally)
	object := NewActivityStreamsPerson()
	objectName := NewActivityStreamsNameProperty()
	objectName.AppendXMLSchemaString("John")
	object.SetActivityStreamsName(objectName)
	rel := MustParseURL("http://purl.org/vocab/relationship/acquaintanceOf")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally is an acquaintance of John")
	example47Type.SetActivityStreamsSummary(summary)
	subj := NewActivityStreamsSubjectProperty()
	subj.SetActivityStreamsPerson(subject)
	example47Type.SetActivityStreamsSubject(subj)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsPerson(object)
	example47Type.SetActivityStreamsObject(obj)
	relationship := NewActivityStreamsRelationshipProperty()
	relationship.AppendIRI(rel)
	example47Type.SetActivityStreamsRelationship(relationship)
	return example47Type
}

const example48 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Article",
  "name": "What a Crazy Day I Had",
  "content": "<div>... you will never believe ...</div>",
  "attributedTo": "http://sally.example.org"
}`

func example48Type() vocab.ActivityStreamsArticle {
	example48Type := NewActivityStreamsArticle()
	att := MustParseURL("http://sally.example.org")
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("What a Crazy Day I Had")
	example48Type.SetActivityStreamsName(name)
	attr := NewActivityStreamsAttributedToProperty()
	attr.AppendIRI(att)
	example48Type.SetActivityStreamsAttributedTo(attr)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("<div>... you will never believe ...</div>")
	example48Type.SetActivityStreamsContent(content)
	return example48Type
}

const example49 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Document",
  "name": "4Q Sales Forecast",
  "url": "http://example.org/4q-sales-forecast.pdf"
}`

func example49Type() vocab.ActivityStreamsDocument {
	example49Type := NewActivityStreamsDocument()
	l := MustParseURL("http://example.org/4q-sales-forecast.pdf")
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("4Q Sales Forecast")
	example49Type.SetActivityStreamsName(name)
	urlProp := NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(l)
	example49Type.SetActivityStreamsUrl(urlProp)
	return example49Type
}

const example50 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Audio",
  "name": "Interview With A Famous Technologist",
  "url": {
    "type": "Link",
    "href": "http://example.org/podcast.mp3",
    "mediaType": "audio/mp3"
  }
}`

func example50Type() vocab.ActivityStreamsAudio {
	example50Type := NewActivityStreamsAudio()
	l := MustParseURL("http://example.org/podcast.mp3")
	link := NewActivityStreamsLink()
	href := NewActivityStreamsHrefProperty()
	href.Set(l)
	link.SetActivityStreamsHref(href)
	mediaType := NewActivityStreamsMediaTypeProperty()
	mediaType.Set("audio/mp3")
	link.SetActivityStreamsMediaType(mediaType)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Interview With A Famous Technologist")
	example50Type.SetActivityStreamsName(name)
	urlProperty := NewActivityStreamsUrlProperty()
	urlProperty.AppendActivityStreamsLink(link)
	example50Type.SetActivityStreamsUrl(urlProperty)
	return example50Type
}

const example51 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Image",
  "name": "Cat Jumping on Wagon",
  "url": [
    {
      "type": "Link",
      "href": "http://example.org/image.jpeg",
      "mediaType": "image/jpeg"
    },
    {
      "type": "Link",
      "href": "http://example.org/image.png",
      "mediaType": "image/png"
    }
  ]
}`

func example51Type() vocab.ActivityStreamsImage {
	example51Type := NewActivityStreamsImage()
	l1 := MustParseURL("http://example.org/image.jpeg")
	l2 := MustParseURL("http://example.org/image.png")
	link1 := NewActivityStreamsLink()
	href1 := NewActivityStreamsHrefProperty()
	href1.Set(l1)
	link1.SetActivityStreamsHref(href1)
	mediaType1 := NewActivityStreamsMediaTypeProperty()
	mediaType1.Set("image/jpeg")
	link1.SetActivityStreamsMediaType(mediaType1)
	link2 := NewActivityStreamsLink()
	href2 := NewActivityStreamsHrefProperty()
	href2.Set(l2)
	link2.SetActivityStreamsHref(href2)
	mediaType2 := NewActivityStreamsMediaTypeProperty()
	mediaType2.Set("image/png")
	link2.SetActivityStreamsMediaType(mediaType2)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Cat Jumping on Wagon")
	example51Type.SetActivityStreamsName(name)
	urlProperty := NewActivityStreamsUrlProperty()
	urlProperty.AppendActivityStreamsLink(link1)
	urlProperty.AppendActivityStreamsLink(link2)
	example51Type.SetActivityStreamsUrl(urlProperty)
	return example51Type
}

const example52 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Video",
  "name": "Puppy Plays With Ball",
  "url": "http://example.org/video.mkv",
  "duration": "PT2H"
}`

func example52Type() vocab.ActivityStreamsVideo {
	example52Type := NewActivityStreamsVideo()
	l := MustParseURL("http://example.org/video.mkv")
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Puppy Plays With Ball")
	example52Type.SetActivityStreamsName(name)
	urlProp := NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(l)
	example52Type.SetActivityStreamsUrl(urlProp)
	dur := NewActivityStreamsDurationProperty()
	dur.Set(time.Hour * 2)
	example52Type.SetActivityStreamsDuration(dur)
	return example52Type
}

const example53 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Note",
  "name": "A Word of Warning",
  "content": "Looks like it is going to rain today. Bring an umbrella!"
}`

func example53Type() vocab.ActivityStreamsNote {
	example53Type := NewActivityStreamsNote()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("A Word of Warning")
	example53Type.SetActivityStreamsName(name)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("Looks like it is going to rain today. Bring an umbrella!")
	example53Type.SetActivityStreamsContent(content)
	return example53Type
}

const example54 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Page",
  "name": "Omaha Weather Report",
  "url": "http://example.org/weather-in-omaha.html"
}`

func example54Type() vocab.ActivityStreamsPage {
	example54Type := NewActivityStreamsPage()
	l := MustParseURL("http://example.org/weather-in-omaha.html")
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Omaha Weather Report")
	example54Type.SetActivityStreamsName(name)
	urlProp := NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(l)
	example54Type.SetActivityStreamsUrl(urlProp)
	return example54Type
}

const example55 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Event",
  "name": "Going-Away Party for Jim",
  "startTime": "2014-12-31T23:00:00-08:00",
  "endTime": "2015-01-01T06:00:00-08:00"
}`

func example55Type() vocab.ActivityStreamsEvent {
	example55Type := NewActivityStreamsEvent()
	t1, err := time.Parse(time.RFC3339, "2014-12-31T23:00:00-08:00")
	if err != nil {
		panic(err)
	}
	t2, err := time.Parse(time.RFC3339, "2015-01-01T06:00:00-08:00")
	if err != nil {
		panic(err)
	}
	goingAway := NewActivityStreamsNameProperty()
	goingAway.AppendXMLSchemaString("Going-Away Party for Jim")
	example55Type.SetActivityStreamsName(goingAway)
	startTime := NewActivityStreamsStartTimeProperty()
	startTime.Set(t1)
	example55Type.SetActivityStreamsStartTime(startTime)
	endTime := NewActivityStreamsEndTimeProperty()
	endTime.Set(t2)
	example55Type.SetActivityStreamsEndTime(endTime)
	return example55Type
}

const example56 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Place",
  "name": "Work"
}`

func example56Type() vocab.ActivityStreamsPlace {
	example56Type := NewActivityStreamsPlace()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Work")
	example56Type.SetActivityStreamsName(name)
	return example56Type
}

const example57 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Place",
  "name": "Fresno Area",
  "latitude": 36.75,
  "longitude": 119.7667,
  "radius": 15,
  "units": "miles"
}`

func example57Type() vocab.ActivityStreamsPlace {
	example57Type := NewActivityStreamsPlace()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Fresno Area")
	example57Type.SetActivityStreamsName(name)
	lat := NewActivityStreamsLatitudeProperty()
	lat.Set(36.75)
	example57Type.SetActivityStreamsLatitude(lat)
	lon := NewActivityStreamsLongitudeProperty()
	lon.Set(119.7667)
	example57Type.SetActivityStreamsLongitude(lon)
	rad := NewActivityStreamsRadiusProperty()
	rad.Set(15)
	example57Type.SetActivityStreamsRadius(rad)
	units := NewActivityStreamsUnitsProperty()
	units.SetXMLSchemaString("miles")
	example57Type.SetActivityStreamsUnits(units)
	return example57Type
}

const example58 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Mention",
  "summary": "Mention of Joe by Carrie in her note",
  "href": "http://example.org/joe",
  "name": "Joe"
}`

func example58Type() vocab.ActivityStreamsMention {
	example58Type := NewActivityStreamsMention()
	l := MustParseURL("http://example.org/joe")
	href := NewActivityStreamsHrefProperty()
	href.Set(l)
	example58Type.SetActivityStreamsHref(href)
	joe := NewActivityStreamsNameProperty()
	joe.AppendXMLSchemaString("Joe")
	example58Type.SetActivityStreamsName(joe)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Mention of Joe by Carrie in her note")
	example58Type.SetActivityStreamsSummary(summary)
	return example58Type
}

const example59 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Profile",
  "summary": "Sally's Profile",
  "describes": {
    "type": "Person",
    "name": "Sally Smith"
  }
}`

func example59Type() vocab.ActivityStreamsProfile {
	example59Type := NewActivityStreamsProfile()
	person := NewActivityStreamsPerson()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Sally Smith")
	person.SetActivityStreamsName(name)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally's Profile")
	example59Type.SetActivityStreamsSummary(summary)
	describes := NewActivityStreamsDescribesProperty()
	describes.SetActivityStreamsPerson(person)
	example59Type.SetActivityStreamsDescribes(describes)
	return example59Type
}

// Note that the @context is missing from the spec!
const example60 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "OrderedCollection",
  "totalItems": 3,
  "name": "Vacation photos 2016",
  "orderedItems": [
    {
      "type": "Image",
      "id": "http://image.example/1"
    },
    {
      "type": "Tombstone",
      "formerType": "Image",
      "id": "http://image.example/2",
      "deleted": "2016-03-17T00:00:00Z"
    },
    {
      "type": "Image",
      "id": "http://image.example/3"
    }
  ]
}`

func example60Type() vocab.ActivityStreamsOrderedCollection {
	example60Type := NewActivityStreamsOrderedCollection()
	t, err := time.Parse(time.RFC3339, "2016-03-17T00:00:00Z")
	if err != nil {
		panic(err)
	}
	image1 := NewActivityStreamsImage()
	imgId1 := NewJSONLDIdProperty()
	imgId1.Set(MustParseURL("http://image.example/1"))
	image1.SetJSONLDId(imgId1)
	tombstone := NewActivityStreamsTombstone()
	ft := NewActivityStreamsFormerTypeProperty()
	ft.AppendXMLSchemaString("Image")
	tombstone.SetActivityStreamsFormerType(ft)
	tombId := NewJSONLDIdProperty()
	tombId.Set(MustParseURL("http://image.example/2"))
	tombstone.SetJSONLDId(tombId)
	deleted := NewActivityStreamsDeletedProperty()
	deleted.Set(t)
	tombstone.SetActivityStreamsDeleted(deleted)
	image2 := NewActivityStreamsImage()
	imgId2 := NewJSONLDIdProperty()
	imgId2.Set(MustParseURL("http://image.example/3"))
	image2.SetJSONLDId(imgId2)
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(3)
	example60Type.SetActivityStreamsTotalItems(totalItems)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Vacation photos 2016")
	example60Type.SetActivityStreamsName(name)
	orderedItems := NewActivityStreamsOrderedItemsProperty()
	orderedItems.AppendActivityStreamsImage(image1)
	orderedItems.AppendActivityStreamsTombstone(tombstone)
	orderedItems.AppendActivityStreamsImage(image2)
	example60Type.SetActivityStreamsOrderedItems(orderedItems)
	return example60Type
}

const example61 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "name": "Foo",
  "id": "http://example.org/foo"
}`

var example61Unknown = func(m map[string]interface{}) map[string]interface{} {
	m["@context"] = "https://www.w3.org/ns/activitystreams"
	m["id"] = "http://example.org/foo"
	m["name"] = "Foo"
	return m
}

const example62 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A foo",
  "type": "http://example.org/Foo"
}`

var example62Unknown = func(m map[string]interface{}) map[string]interface{} {
	m["@context"] = "https://www.w3.org/ns/activitystreams"
	m["type"] = "http://example.org/Foo"
	m["summary"] = "A foo"
	return m
}

const example63 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally offered the Foo object",
  "type": "Offer",
  "actor": "http://sally.example.org",
  "object": "http://example.org/foo"
}`

func example63Type() vocab.ActivityStreamsOffer {
	example63Type := NewActivityStreamsOffer()
	l := MustParseURL("http://sally.example.org")
	o := MustParseURL("http://example.org/foo")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally offered the Foo object")
	example63Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(l)
	example63Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	example63Type.SetActivityStreamsObject(obj)
	return example63Type
}

const example64 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally offered the Foo object",
  "type": "Offer",
  "actor": {
    "type": "Person",
    "id": "http://sally.example.org",
    "summary": "Sally"
  },
  "object": "http://example.org/foo"
}`

func example64Type() vocab.ActivityStreamsOffer {
	example64Type := NewActivityStreamsOffer()
	actor := NewActivityStreamsPerson()
	actorId := NewJSONLDIdProperty()
	actorId.Set(MustParseURL("http://sally.example.org"))
	actor.SetJSONLDId(actorId)
	summaryActor := NewActivityStreamsSummaryProperty()
	summaryActor.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsSummary(summaryActor)
	o := MustParseURL("http://example.org/foo")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally offered the Foo object")
	example64Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example64Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	example64Type.SetActivityStreamsObject(obj)
	return example64Type
}

const example65 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally and Joe offered the Foo object",
  "type": "Offer",
  "actor": [
    "http://joe.example.org",
    {
      "type": "Person",
      "id": "http://sally.example.org",
      "name": "Sally"
    }
  ],
  "object": "http://example.org/foo"
}`

func example65Type() vocab.ActivityStreamsOffer {
	example65Type := NewActivityStreamsOffer()
	actor := NewActivityStreamsPerson()
	actorId := NewJSONLDIdProperty()
	actorId.Set(MustParseURL("http://sally.example.org"))
	actor.SetJSONLDId(actorId)
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	o := MustParseURL("http://example.org/foo")
	l := MustParseURL("http://joe.example.org")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally and Joe offered the Foo object")
	example65Type.SetActivityStreamsSummary(summary)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	example65Type.SetActivityStreamsObject(obj)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendIRI(l)
	rootActor.AppendActivityStreamsPerson(actor)
	example65Type.SetActivityStreamsActor(rootActor)
	return example65Type
}

// NOTE: Changed to not be an array value for "attachment" to keep in line with other examples in spec!
const example66 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Note",
  "name": "Have you seen my cat?",
  "attachment": {
    "type": "Image",
    "content": "This is what he looks like.",
    "url": "http://example.org/cat.jpeg"
  }
}`

func example66Type() vocab.ActivityStreamsNote {
	example66Type := NewActivityStreamsNote()
	l := MustParseURL("http://example.org/cat.jpeg")
	image := NewActivityStreamsImage()
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("This is what he looks like.")
	image.SetActivityStreamsContent(content)
	imgProp := NewActivityStreamsUrlProperty()
	imgProp.AppendIRI(l)
	image.SetActivityStreamsUrl(imgProp)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Have you seen my cat?")
	example66Type.SetActivityStreamsName(name)
	attachment := NewActivityStreamsAttachmentProperty()
	attachment.AppendActivityStreamsImage(image)
	example66Type.SetActivityStreamsAttachment(attachment)
	return example66Type
}

// NOTE: Changed to not be an array value for "attributedTo" to keep in line with other examples in spec!
const example67 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Image",
  "name": "My cat taking a nap",
  "url": "http://example.org/cat.jpeg",
  "attributedTo": {
    "type": "Person",
    "name": "Sally"
  }
}`

func example67Type() vocab.ActivityStreamsImage {
	example67Type := NewActivityStreamsImage()
	l := MustParseURL("http://example.org/cat.jpeg")
	person := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("My cat taking a nap")
	example67Type.SetActivityStreamsName(name)
	urlProp := NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(l)
	example67Type.SetActivityStreamsUrl(urlProp)
	attr := NewActivityStreamsAttributedToProperty()
	attr.AppendActivityStreamsPerson(person)
	example67Type.SetActivityStreamsAttributedTo(attr)
	return example67Type
}

const example68 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Image",
  "name": "My cat taking a nap",
  "url": "http://example.org/cat.jpeg",
  "attributedTo": [
    "http://joe.example.org",
    {
      "type": "Person",
      "name": "Sally"
    }
  ]
}`

func example68Type() vocab.ActivityStreamsImage {
	example68Type := NewActivityStreamsImage()
	l := MustParseURL("http://example.org/cat.jpeg")
	a := MustParseURL("http://joe.example.org")
	person := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("My cat taking a nap")
	example68Type.SetActivityStreamsName(name)
	urlProp := NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(l)
	example68Type.SetActivityStreamsUrl(urlProp)
	attr := NewActivityStreamsAttributedToProperty()
	attr.AppendIRI(a)
	attr.AppendActivityStreamsPerson(person)
	example68Type.SetActivityStreamsAttributedTo(attr)
	return example68Type
}

const example69 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "name": "Holiday announcement",
  "type": "Note",
  "content": "Thursday will be a company-wide holiday. Enjoy your day off!",
  "audience": {
    "type": "http://example.org/Organization",
    "name": "ExampleCo LLC"
  }
}`

var example69Unknown = func(m map[string]interface{}) map[string]interface{} {
	m["audience"] = map[string]interface{}{
		"type": "http://example.org/Organization",
		"name": "ExampleCo LLC",
	}
	return m
}

func example69Type() vocab.ActivityStreamsNote {
	example69Type := NewActivityStreamsNote()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Holiday announcement")
	example69Type.SetActivityStreamsName(name)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("Thursday will be a company-wide holiday. Enjoy your day off!")
	example69Type.SetActivityStreamsContent(content)
	return example69Type
}

// NOTE: Changed to not be an array value for "bcc" to keep in line with other examples in spec!
const example70 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally offered a post to John",
  "type": "Offer",
  "actor": "http://sally.example.org",
  "object": "http://example.org/posts/1",
  "target": "http://john.example.org",
  "bcc": "http://joe.example.org"
}`

func example70Type() vocab.ActivityStreamsOffer {
	example70Type := NewActivityStreamsOffer()
	o := MustParseURL("http://example.org/posts/1")
	a := MustParseURL("http://sally.example.org")
	t := MustParseURL("http://john.example.org")
	b := MustParseURL("http://joe.example.org")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally offered a post to John")
	example70Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(a)
	example70Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	example70Type.SetActivityStreamsObject(obj)
	target := NewActivityStreamsTargetProperty()
	target.AppendIRI(t)
	example70Type.SetActivityStreamsTarget(target)
	bcc := NewActivityStreamsBccProperty()
	bcc.AppendIRI(b)
	example70Type.SetActivityStreamsBcc(bcc)
	return example70Type
}

// NOTE: Changed to not be an array value for "bto" to keep in line with other examples in spec!
const example71 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally offered a post to John",
  "type": "Offer",
  "actor": "http://sally.example.org",
  "object": "http://example.org/posts/1",
  "target": "http://john.example.org",
  "bto": "http://joe.example.org"
}`

func example71Type() vocab.ActivityStreamsOffer {
	example71Type := NewActivityStreamsOffer()
	o := MustParseURL("http://example.org/posts/1")
	a := MustParseURL("http://sally.example.org")
	t := MustParseURL("http://john.example.org")
	b := MustParseURL("http://joe.example.org")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally offered a post to John")
	example71Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(a)
	example71Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	example71Type.SetActivityStreamsObject(obj)
	target := NewActivityStreamsTargetProperty()
	target.AppendIRI(t)
	example71Type.SetActivityStreamsTarget(target)
	bto := NewActivityStreamsBtoProperty()
	bto.AppendIRI(b)
	example71Type.SetActivityStreamsBto(bto)
	return example71Type
}

// NOTE: Changed to not be an array value for "cc" to keep in line with other examples in spec!
const example72 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally offered a post to John",
  "type": "Offer",
  "actor": "http://sally.example.org",
  "object": "http://example.org/posts/1",
  "target": "http://john.example.org",
  "cc": "http://joe.example.org"
}`

func example72Type() vocab.ActivityStreamsOffer {
	example72Type := NewActivityStreamsOffer()
	o := MustParseURL("http://example.org/posts/1")
	a := MustParseURL("http://sally.example.org")
	t := MustParseURL("http://john.example.org")
	b := MustParseURL("http://joe.example.org")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally offered a post to John")
	example72Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(a)
	example72Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	example72Type.SetActivityStreamsObject(obj)
	target := NewActivityStreamsTargetProperty()
	target.AppendIRI(t)
	example72Type.SetActivityStreamsTarget(target)
	cc := NewActivityStreamsCcProperty()
	cc.AppendIRI(b)
	example72Type.SetActivityStreamsCc(cc)
	return example72Type
}

const example73 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Activities in context 1",
  "type": "Collection",
  "items": [
    {
      "type": "Offer",
      "actor": "http://sally.example.org",
      "object": "http://example.org/posts/1",
      "target": "http://john.example.org",
      "context": "http://example.org/contexts/1"
    },
    {
      "type": "Like",
      "actor": "http://joe.example.org",
      "object": "http://example.org/posts/2",
      "context": "http://example.org/contexts/1"
    }
  ]
}`

func example73Type() vocab.ActivityStreamsCollection {
	example73Type := NewActivityStreamsCollection()
	oa := MustParseURL("http://sally.example.org")
	oo := MustParseURL("http://example.org/posts/1")
	ot := MustParseURL("http://john.example.org")
	oc := MustParseURL("http://example.org/contexts/1")
	offer := NewActivityStreamsOffer()
	offerActor := NewActivityStreamsActorProperty()
	offerActor.AppendIRI(oa)
	offer.SetActivityStreamsActor(offerActor)
	objOffer := NewActivityStreamsObjectProperty()
	objOffer.AppendIRI(oo)
	offer.SetActivityStreamsObject(objOffer)
	target := NewActivityStreamsTargetProperty()
	target.AppendIRI(ot)
	offer.SetActivityStreamsTarget(target)
	ctx := NewActivityStreamsContextProperty()
	ctx.AppendIRI(oc)
	offer.SetActivityStreamsContext(ctx)
	la := MustParseURL("http://joe.example.org")
	lo := MustParseURL("http://example.org/posts/2")
	lc := MustParseURL("http://example.org/contexts/1")
	like := NewActivityStreamsLike()
	likeActor := NewActivityStreamsActorProperty()
	likeActor.AppendIRI(la)
	like.SetActivityStreamsActor(likeActor)
	objLike := NewActivityStreamsObjectProperty()
	objLike.AppendIRI(lo)
	like.SetActivityStreamsObject(objLike)
	ctxLike := NewActivityStreamsContextProperty()
	ctxLike.AppendIRI(lc)
	like.SetActivityStreamsContext(ctxLike)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Activities in context 1")
	example73Type.SetActivityStreamsSummary(summary)
	items := NewActivityStreamsItemsProperty()
	items.AppendActivityStreamsOffer(offer)
	items.AppendActivityStreamsLike(like)
	example73Type.SetActivityStreamsItems(items)
	return example73Type
}

const example74 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally's blog posts",
  "type": "Collection",
  "totalItems": 3,
  "current": "http://example.org/collection",
  "items": [
    "http://example.org/posts/1",
    "http://example.org/posts/2",
    "http://example.org/posts/3"
  ]
}`

func example74Type() vocab.ActivityStreamsCollection {
	example74Type := NewActivityStreamsCollection()
	c := MustParseURL("http://example.org/collection")
	i1 := MustParseURL("http://example.org/posts/1")
	i2 := MustParseURL("http://example.org/posts/2")
	i3 := MustParseURL("http://example.org/posts/3")
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(3)
	example74Type.SetActivityStreamsTotalItems(totalItems)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally's blog posts")
	example74Type.SetActivityStreamsSummary(summary)
	current := NewActivityStreamsCurrentProperty()
	current.SetIRI(c)
	example74Type.SetActivityStreamsCurrent(current)
	items := NewActivityStreamsItemsProperty()
	items.AppendIRI(i1)
	items.AppendIRI(i2)
	items.AppendIRI(i3)
	example74Type.SetActivityStreamsItems(items)
	return example74Type
}

const example75 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally's blog posts",
  "type": "Collection",
  "totalItems": 3,
  "current": {
    "type": "Link",
    "summary": "Most Recent Items",
    "href": "http://example.org/collection"
  },
  "items": [
    "http://example.org/posts/1",
    "http://example.org/posts/2",
    "http://example.org/posts/3"
  ]
}`

func example75Type() vocab.ActivityStreamsCollection {
	example75Type := NewActivityStreamsCollection()
	i1 := MustParseURL("http://example.org/posts/1")
	i2 := MustParseURL("http://example.org/posts/2")
	i3 := MustParseURL("http://example.org/posts/3")
	href := MustParseURL("http://example.org/collection")
	link := NewActivityStreamsLink()
	summaryLink := NewActivityStreamsSummaryProperty()
	summaryLink.AppendXMLSchemaString("Most Recent Items")
	link.SetActivityStreamsSummary(summaryLink)
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(href)
	link.SetActivityStreamsHref(hrefLink)
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(3)
	example75Type.SetActivityStreamsTotalItems(totalItems)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally's blog posts")
	example75Type.SetActivityStreamsSummary(summary)
	current := NewActivityStreamsCurrentProperty()
	current.SetActivityStreamsLink(link)
	example75Type.SetActivityStreamsCurrent(current)
	items := NewActivityStreamsItemsProperty()
	items.AppendIRI(i1)
	items.AppendIRI(i2)
	items.AppendIRI(i3)
	example75Type.SetActivityStreamsItems(items)
	return example75Type
}

const example76 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally's blog posts",
  "type": "Collection",
  "totalItems": 3,
  "first": "http://example.org/collection?page=0"
}`

func example76Type() vocab.ActivityStreamsCollection {
	example76Type := NewActivityStreamsCollection()
	f := MustParseURL("http://example.org/collection?page=0")
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(3)
	example76Type.SetActivityStreamsTotalItems(totalItems)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally's blog posts")
	example76Type.SetActivityStreamsSummary(summary)
	first := NewActivityStreamsFirstProperty()
	first.SetIRI(f)
	example76Type.SetActivityStreamsFirst(first)
	return example76Type
}

const example77 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally's blog posts",
  "type": "Collection",
  "totalItems": 3,
  "first": {
    "type": "Link",
    "summary": "First Page",
    "href": "http://example.org/collection?page=0"
  }
}`

func example77Type() vocab.ActivityStreamsCollection {
	example77Type := NewActivityStreamsCollection()
	href := MustParseURL("http://example.org/collection?page=0")
	link := NewActivityStreamsLink()
	summaryLink := NewActivityStreamsSummaryProperty()
	summaryLink.AppendXMLSchemaString("First Page")
	link.SetActivityStreamsSummary(summaryLink)
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(href)
	link.SetActivityStreamsHref(hrefLink)
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(3)
	example77Type.SetActivityStreamsTotalItems(totalItems)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally's blog posts")
	example77Type.SetActivityStreamsSummary(summary)
	first := NewActivityStreamsFirstProperty()
	first.SetActivityStreamsLink(link)
	example77Type.SetActivityStreamsFirst(first)
	return example77Type
}

const example78 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A simple note",
  "type": "Note",
  "content": "This is all there is.",
  "generator": {
    "type": "Application",
    "name": "Exampletron 3000"
  }
}`

func example78Type() vocab.ActivityStreamsNote {
	example78Type := NewActivityStreamsNote()
	app := NewActivityStreamsApplication()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Exampletron 3000")
	app.SetActivityStreamsName(name)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A simple note")
	example78Type.SetActivityStreamsSummary(summary)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("This is all there is.")
	example78Type.SetActivityStreamsContent(content)
	gen := NewActivityStreamsGeneratorProperty()
	gen.AppendActivityStreamsApplication(app)
	example78Type.SetActivityStreamsGenerator(gen)
	return example78Type
}

const example79 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A simple note",
  "type": "Note",
  "content": "This is all there is.",
  "icon": {
    "type": "Image",
    "name": "Note icon",
    "url": "http://example.org/note.png",
    "width": 16,
    "height": 16
  }
}`

func example79Type() vocab.ActivityStreamsNote {
	example79Type := NewActivityStreamsNote()
	u := MustParseURL("http://example.org/note.png")
	image := NewActivityStreamsImage()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Note icon")
	image.SetActivityStreamsName(name)
	imgProp := NewActivityStreamsUrlProperty()
	imgProp.AppendIRI(u)
	image.SetActivityStreamsUrl(imgProp)
	width := NewActivityStreamsWidthProperty()
	width.Set(16)
	image.SetActivityStreamsWidth(width)
	height := NewActivityStreamsHeightProperty()
	height.Set(16)
	image.SetActivityStreamsHeight(height)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A simple note")
	example79Type.SetActivityStreamsSummary(summary)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("This is all there is.")
	example79Type.SetActivityStreamsContent(content)
	icon := NewActivityStreamsIconProperty()
	icon.AppendActivityStreamsImage(image)
	example79Type.SetActivityStreamsIcon(icon)
	return example79Type
}

const example80 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A simple note",
  "type": "Note",
  "content": "A simple note",
  "icon": [
    {
      "type": "Image",
      "summary": "Note (16x16)",
      "url": "http://example.org/note1.png",
      "width": 16,
      "height": 16
    },
    {
      "type": "Image",
      "summary": "Note (32x32)",
      "url": "http://example.org/note2.png",
      "width": 32,
      "height": 32
    }
  ]
}`

func example80Type() vocab.ActivityStreamsNote {
	example80Type := NewActivityStreamsNote()
	u1 := MustParseURL("http://example.org/note1.png")
	u2 := MustParseURL("http://example.org/note2.png")
	image1 := NewActivityStreamsImage()
	summaryImg1 := NewActivityStreamsSummaryProperty()
	summaryImg1.AppendXMLSchemaString("Note (16x16)")
	image1.SetActivityStreamsSummary(summaryImg1)
	imgProp1 := NewActivityStreamsUrlProperty()
	imgProp1.AppendIRI(u1)
	image1.SetActivityStreamsUrl(imgProp1)
	width1 := NewActivityStreamsWidthProperty()
	width1.Set(16)
	image1.SetActivityStreamsWidth(width1)
	height1 := NewActivityStreamsHeightProperty()
	height1.Set(16)
	image1.SetActivityStreamsHeight(height1)
	image2 := NewActivityStreamsImage()
	summaryImg2 := NewActivityStreamsSummaryProperty()
	summaryImg2.AppendXMLSchemaString("Note (32x32)")
	image2.SetActivityStreamsSummary(summaryImg2)
	imgProp2 := NewActivityStreamsUrlProperty()
	imgProp2.AppendIRI(u2)
	image2.SetActivityStreamsUrl(imgProp2)
	width2 := NewActivityStreamsWidthProperty()
	width2.Set(32)
	image2.SetActivityStreamsWidth(width2)
	height2 := NewActivityStreamsHeightProperty()
	height2.Set(32)
	image2.SetActivityStreamsHeight(height2)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A simple note")
	example80Type.SetActivityStreamsSummary(summary)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("A simple note")
	example80Type.SetActivityStreamsContent(content)
	icon := NewActivityStreamsIconProperty()
	icon.AppendActivityStreamsImage(image1)
	icon.AppendActivityStreamsImage(image2)
	example80Type.SetActivityStreamsIcon(icon)
	return example80Type
}

const example81 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "name": "A simple note",
  "type": "Note",
  "content": "This is all there is.",
  "image": {
    "type": "Image",
    "name": "A Cat",
    "url": "http://example.org/cat.png"
  }
}`

func example81Type() vocab.ActivityStreamsNote {
	example81Type := NewActivityStreamsNote()
	u := MustParseURL("http://example.org/cat.png")
	image := NewActivityStreamsImage()
	imageName := NewActivityStreamsNameProperty()
	imageName.AppendXMLSchemaString("A Cat")
	image.SetActivityStreamsName(imageName)
	imgProp := NewActivityStreamsUrlProperty()
	imgProp.AppendIRI(u)
	image.SetActivityStreamsUrl(imgProp)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("A simple note")
	example81Type.SetActivityStreamsName(name)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("This is all there is.")
	example81Type.SetActivityStreamsContent(content)
	imageProp := NewActivityStreamsImageProperty()
	imageProp.AppendActivityStreamsImage(image)
	example81Type.SetActivityStreamsImage(imageProp)
	return example81Type
}

const example82 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "name": "A simple note",
  "type": "Note",
  "content": "This is all there is.",
  "image": [
    {
      "type": "Image",
      "name": "Cat 1",
      "url": "http://example.org/cat1.png"
    },
    {
      "type": "Image",
      "name": "Cat 2",
      "url": "http://example.org/cat2.png"
    }
  ]
}`

func example82Type() vocab.ActivityStreamsNote {
	example82Type := NewActivityStreamsNote()
	u1 := MustParseURL("http://example.org/cat1.png")
	u2 := MustParseURL("http://example.org/cat2.png")
	image1 := NewActivityStreamsImage()
	image1Name := NewActivityStreamsNameProperty()
	image1Name.AppendXMLSchemaString("Cat 1")
	image1.SetActivityStreamsName(image1Name)
	imgProp1 := NewActivityStreamsUrlProperty()
	imgProp1.AppendIRI(u1)
	image1.SetActivityStreamsUrl(imgProp1)
	image2 := NewActivityStreamsImage()
	image2Name := NewActivityStreamsNameProperty()
	image2Name.AppendXMLSchemaString("Cat 2")
	image2.SetActivityStreamsName(image2Name)
	imgProp2 := NewActivityStreamsUrlProperty()
	imgProp2.AppendIRI(u2)
	image2.SetActivityStreamsUrl(imgProp2)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("A simple note")
	example82Type.SetActivityStreamsName(name)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("This is all there is.")
	example82Type.SetActivityStreamsContent(content)
	imageProp := NewActivityStreamsImageProperty()
	imageProp.AppendActivityStreamsImage(image1)
	imageProp.AppendActivityStreamsImage(image2)
	example82Type.SetActivityStreamsImage(imageProp)
	return example82Type
}

const example83 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A simple note",
  "type": "Note",
  "content": "This is all there is.",
  "inReplyTo": {
    "summary": "Previous note",
    "type": "Note",
    "content": "What else is there?"
  }
}`

func example83Type() vocab.ActivityStreamsNote {
	example83Type := NewActivityStreamsNote()
	note := NewActivityStreamsNote()
	summaryNote := NewActivityStreamsSummaryProperty()
	summaryNote.AppendXMLSchemaString("Previous note")
	note.SetActivityStreamsSummary(summaryNote)
	contentNote := NewActivityStreamsContentProperty()
	contentNote.AppendXMLSchemaString("What else is there?")
	note.SetActivityStreamsContent(contentNote)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A simple note")
	example83Type.SetActivityStreamsSummary(summary)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("This is all there is.")
	example83Type.SetActivityStreamsContent(content)
	inReplyTo := NewActivityStreamsInReplyToProperty()
	inReplyTo.AppendActivityStreamsNote(note)
	example83Type.SetActivityStreamsInReplyTo(inReplyTo)
	return example83Type
}

const example84 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A simple note",
  "type": "Note",
  "content": "This is all there is.",
  "inReplyTo": "http://example.org/posts/1"
}`

func example84Type() vocab.ActivityStreamsNote {
	example84Type := NewActivityStreamsNote()
	u := MustParseURL("http://example.org/posts/1")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A simple note")
	example84Type.SetActivityStreamsSummary(summary)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("This is all there is.")
	example84Type.SetActivityStreamsContent(content)
	inReplyTo := NewActivityStreamsInReplyToProperty()
	inReplyTo.AppendIRI(u)
	example84Type.SetActivityStreamsInReplyTo(inReplyTo)
	return example84Type
}

const example85 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally listened to a piece of music on the Acme Music Service",
  "type": "Listen",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": "http://example.org/foo.mp3",
  "instrument": {
    "type": "Service",
    "name": "Acme Music Service"
  }
}`

func example85Type() vocab.ActivityStreamsListen {
	example85Type := NewActivityStreamsListen()
	u := MustParseURL("http://example.org/foo.mp3")
	actor := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	actor.SetActivityStreamsName(sally)
	service := NewActivityStreamsService()
	serviceName := NewActivityStreamsNameProperty()
	serviceName.AppendXMLSchemaString("Acme Music Service")
	service.SetActivityStreamsName(serviceName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally listened to a piece of music on the Acme Music Service")
	example85Type.SetActivityStreamsSummary(summary)
	rootActor := NewActivityStreamsActorProperty()
	rootActor.AppendActivityStreamsPerson(actor)
	example85Type.SetActivityStreamsActor(rootActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(u)
	example85Type.SetActivityStreamsObject(obj)
	inst := NewActivityStreamsInstrumentProperty()
	inst.AppendActivityStreamsService(service)
	example85Type.SetActivityStreamsInstrument(inst)
	return example85Type
}

const example86 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A collection",
  "type": "Collection",
  "totalItems": 3,
  "last": "http://example.org/collection?page=1"
}`

func example86Type() vocab.ActivityStreamsCollection {
	example86Type := NewActivityStreamsCollection()
	u := MustParseURL("http://example.org/collection?page=1")
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(3)
	example86Type.SetActivityStreamsTotalItems(totalItems)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A collection")
	example86Type.SetActivityStreamsSummary(summary)
	last := NewActivityStreamsLastProperty()
	last.SetIRI(u)
	example86Type.SetActivityStreamsLast(last)
	return example86Type
}

const example87 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A collection",
  "type": "Collection",
  "totalItems": 5,
  "last": {
    "type": "Link",
    "summary": "Last Page",
    "href": "http://example.org/collection?page=1"
  }
}`

func example87Type() vocab.ActivityStreamsCollection {
	example87Type := NewActivityStreamsCollection()
	u := MustParseURL("http://example.org/collection?page=1")
	link := NewActivityStreamsLink()
	summaryLink := NewActivityStreamsSummaryProperty()
	summaryLink.AppendXMLSchemaString("Last Page")
	link.SetActivityStreamsSummary(summaryLink)
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(u)
	link.SetActivityStreamsHref(hrefLink)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A collection")
	example87Type.SetActivityStreamsSummary(summary)
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(5)
	example87Type.SetActivityStreamsTotalItems(totalItems)
	last := NewActivityStreamsLastProperty()
	last.SetActivityStreamsLink(link)
	example87Type.SetActivityStreamsLast(last)
	return example87Type
}

const example88 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Person",
  "name": "Sally",
  "location": {
    "name": "Over the Arabian Sea, east of Socotra Island Nature Sanctuary",
    "type": "Place",
    "longitude": 12.34,
    "latitude": 56.78,
    "altitude": 90,
    "units": "m"
  }
}`

func example88Type() vocab.ActivityStreamsPerson {
	example88Type := NewActivityStreamsPerson()
	place := NewActivityStreamsPlace()
	placeName := NewActivityStreamsNameProperty()
	placeName.AppendXMLSchemaString("Over the Arabian Sea, east of Socotra Island Nature Sanctuary")
	place.SetActivityStreamsName(placeName)
	lon := NewActivityStreamsLongitudeProperty()
	lon.Set(12.34)
	place.SetActivityStreamsLongitude(lon)
	lat := NewActivityStreamsLatitudeProperty()
	lat.Set(56.78)
	place.SetActivityStreamsLatitude(lat)
	alt := NewActivityStreamsAltitudeProperty()
	alt.Set(90)
	place.SetActivityStreamsAltitude(alt)
	units := NewActivityStreamsUnitsProperty()
	units.SetXMLSchemaString("m")
	place.SetActivityStreamsUnits(units)
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	example88Type.SetActivityStreamsName(sally)
	location := NewActivityStreamsLocationProperty()
	location.AppendActivityStreamsPlace(place)
	example88Type.SetActivityStreamsLocation(location)
	return example88Type
}

const example89 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally's notes",
  "type": "Collection",
  "totalItems": 2,
  "items": [
    {
      "type": "Note",
      "name": "Reminder for Going-Away Party"
    },
    {
      "type": "Note",
      "name": "Meeting 2016-11-17"
    }
  ]
}`

func example89Type() vocab.ActivityStreamsCollection {
	example89Type := NewActivityStreamsCollection()
	note1 := NewActivityStreamsNote()
	note1Name := NewActivityStreamsNameProperty()
	note1Name.AppendXMLSchemaString("Reminder for Going-Away Party")
	note1.SetActivityStreamsName(note1Name)
	note2 := NewActivityStreamsNote()
	note2Name := NewActivityStreamsNameProperty()
	note2Name.AppendXMLSchemaString("Meeting 2016-11-17")
	note2.SetActivityStreamsName(note2Name)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally's notes")
	example89Type.SetActivityStreamsSummary(summary)
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(2)
	example89Type.SetActivityStreamsTotalItems(totalItems)
	items := NewActivityStreamsItemsProperty()
	items.AppendActivityStreamsNote(note1)
	items.AppendActivityStreamsNote(note2)
	example89Type.SetActivityStreamsItems(items)
	return example89Type
}

const example90 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally's notes",
  "type": "OrderedCollection",
  "totalItems": 2,
  "orderedItems": [
    {
      "type": "Note",
      "name": "Meeting 2016-11-17"
    },
    {
      "type": "Note",
      "name": "Reminder for Going-Away Party"
    }
  ]
}`

func example90Type() vocab.ActivityStreamsOrderedCollection {
	example90Type := NewActivityStreamsOrderedCollection()
	note1 := NewActivityStreamsNote()
	note1Name := NewActivityStreamsNameProperty()
	note1Name.AppendXMLSchemaString("Meeting 2016-11-17")
	note1.SetActivityStreamsName(note1Name)
	note2 := NewActivityStreamsNote()
	note2Name := NewActivityStreamsNameProperty()
	note2Name.AppendXMLSchemaString("Reminder for Going-Away Party")
	note2.SetActivityStreamsName(note2Name)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally's notes")
	example90Type.SetActivityStreamsSummary(summary)
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(2)
	example90Type.SetActivityStreamsTotalItems(totalItems)
	orderedItems := NewActivityStreamsOrderedItemsProperty()
	orderedItems.AppendActivityStreamsNote(note1)
	orderedItems.AppendActivityStreamsNote(note2)
	example90Type.SetActivityStreamsOrderedItems(orderedItems)
	return example90Type
}

const example91 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Question",
  "name": "What is the answer?",
  "oneOf": [
    {
      "type": "Note",
      "name": "Option A"
    },
    {
      "type": "Note",
      "name": "Option B"
    }
  ]
}`

func example91Type() vocab.ActivityStreamsQuestion {
	example91Type := NewActivityStreamsQuestion()
	note1 := NewActivityStreamsNote()
	note1Name := NewActivityStreamsNameProperty()
	note1Name.AppendXMLSchemaString("Option A")
	note1.SetActivityStreamsName(note1Name)
	note2 := NewActivityStreamsNote()
	note2Name := NewActivityStreamsNameProperty()
	note2Name.AppendXMLSchemaString("Option B")
	note2.SetActivityStreamsName(note2Name)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("What is the answer?")
	example91Type.SetActivityStreamsName(name)
	oneOf := NewActivityStreamsOneOfProperty()
	oneOf.AppendActivityStreamsNote(note1)
	oneOf.AppendActivityStreamsNote(note2)
	example91Type.SetActivityStreamsOneOf(oneOf)
	return example91Type
}

const example92 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Question",
  "name": "What is the answer?",
  "anyOf": [
    {
      "type": "Note",
      "name": "Option A"
    },
    {
      "type": "Note",
      "name": "Option B"
    }
  ]
}`

func example92Type() vocab.ActivityStreamsQuestion {
	example92Type := NewActivityStreamsQuestion()
	note1 := NewActivityStreamsNote()
	note1Name := NewActivityStreamsNameProperty()
	note1Name.AppendXMLSchemaString("Option A")
	note1.SetActivityStreamsName(note1Name)
	note2 := NewActivityStreamsNote()
	note2Name := NewActivityStreamsNameProperty()
	note2Name.AppendXMLSchemaString("Option B")
	note2.SetActivityStreamsName(note2Name)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("What is the answer?")
	example92Type.SetActivityStreamsName(name)
	anyOf := NewActivityStreamsAnyOfProperty()
	anyOf.AppendActivityStreamsNote(note1)
	anyOf.AppendActivityStreamsNote(note2)
	example92Type.SetActivityStreamsAnyOf(anyOf)
	return example92Type
}

const example93 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Question",
  "name": "What is the answer?",
  "closed": "2016-05-10T00:00:00Z"
}`

func example93Type() vocab.ActivityStreamsQuestion {
	example93Type := NewActivityStreamsQuestion()
	t, err := time.Parse(time.RFC3339, "2016-05-10T00:00:00Z")
	if err != nil {
		panic(err)
	}
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("What is the answer?")
	example93Type.SetActivityStreamsName(name)
	closed := NewActivityStreamsClosedProperty()
	closed.AppendXMLSchemaDateTime(t)
	example93Type.SetActivityStreamsClosed(closed)
	return example93Type
}

const example94 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally moved a post from List A to List B",
  "type": "Move",
  "actor": "http://sally.example.org",
  "object": "http://example.org/posts/1",
  "target": {
    "type": "Collection",
    "name": "List B"
  },
  "origin": {
    "type": "Collection",
    "name": "List A"
  }
}`

func example94Type() vocab.ActivityStreamsMove {
	example94Type := NewActivityStreamsMove()
	a := MustParseURL("http://sally.example.org")
	o := MustParseURL("http://example.org/posts/1")
	target := NewActivityStreamsCollection()
	targetName := NewActivityStreamsNameProperty()
	targetName.AppendXMLSchemaString("List B")
	target.SetActivityStreamsName(targetName)
	origin := NewActivityStreamsCollection()
	originName := NewActivityStreamsNameProperty()
	originName.AppendXMLSchemaString("List A")
	origin.SetActivityStreamsName(originName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally moved a post from List A to List B")
	example94Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(a)
	example94Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	example94Type.SetActivityStreamsObject(obj)
	tobj := NewActivityStreamsTargetProperty()
	tobj.AppendActivityStreamsCollection(target)
	example94Type.SetActivityStreamsTarget(tobj)
	originProp := NewActivityStreamsOriginProperty()
	originProp.AppendActivityStreamsCollection(origin)
	example94Type.SetActivityStreamsOrigin(originProp)
	return example94Type
}

const example95 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Page 2 of Sally's blog posts",
  "type": "CollectionPage",
  "next": "http://example.org/collection?page=2",
  "items": [
    "http://example.org/posts/1",
    "http://example.org/posts/2",
    "http://example.org/posts/3"
  ]
}`

func example95Type() vocab.ActivityStreamsCollectionPage {
	example95Type := NewActivityStreamsCollectionPage()
	i := MustParseURL("http://example.org/collection?page=2")
	u1 := MustParseURL("http://example.org/posts/1")
	u2 := MustParseURL("http://example.org/posts/2")
	u3 := MustParseURL("http://example.org/posts/3")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Page 2 of Sally's blog posts")
	example95Type.SetActivityStreamsSummary(summary)
	next := NewActivityStreamsNextProperty()
	next.SetIRI(i)
	example95Type.SetActivityStreamsNext(next)
	items := NewActivityStreamsItemsProperty()
	items.AppendIRI(u1)
	items.AppendIRI(u2)
	items.AppendIRI(u3)
	example95Type.SetActivityStreamsItems(items)
	return example95Type
}

const example96 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Page 2 of Sally's blog posts",
  "type": "CollectionPage",
  "next": {
    "type": "Link",
    "name": "Next Page",
    "href": "http://example.org/collection?page=2"
  },
  "items": [
    "http://example.org/posts/1",
    "http://example.org/posts/2",
    "http://example.org/posts/3"
  ]
}`

func example96Type() vocab.ActivityStreamsCollectionPage {
	example96Type := NewActivityStreamsCollectionPage()
	href := MustParseURL("http://example.org/collection?page=2")
	u1 := MustParseURL("http://example.org/posts/1")
	u2 := MustParseURL("http://example.org/posts/2")
	u3 := MustParseURL("http://example.org/posts/3")
	link := NewActivityStreamsLink()
	linkName := NewActivityStreamsNameProperty()
	linkName.AppendXMLSchemaString("Next Page")
	link.SetActivityStreamsName(linkName)
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(href)
	link.SetActivityStreamsHref(hrefLink)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Page 2 of Sally's blog posts")
	example96Type.SetActivityStreamsSummary(summary)
	next := NewActivityStreamsNextProperty()
	next.SetActivityStreamsLink(link)
	example96Type.SetActivityStreamsNext(next)
	items := NewActivityStreamsItemsProperty()
	items.AppendIRI(u1)
	items.AppendIRI(u2)
	items.AppendIRI(u3)
	example96Type.SetActivityStreamsItems(items)
	return example96Type
}

const example97 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally liked a post",
  "type": "Like",
  "actor": "http://sally.example.org",
  "object": "http://example.org/posts/1"
}`

func example97Type() vocab.ActivityStreamsLike {
	example97Type := NewActivityStreamsLike()
	a := MustParseURL("http://sally.example.org")
	o := MustParseURL("http://example.org/posts/1")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally liked a post")
	example97Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(a)
	example97Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	example97Type.SetActivityStreamsObject(obj)
	return example97Type
}

const example98 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Like",
  "actor": "http://sally.example.org",
  "object": {
    "type": "Note",
    "content": "A simple note"
  }
}`

func example98Type() vocab.ActivityStreamsLike {
	example98Type := NewActivityStreamsLike()
	a := MustParseURL("http://sally.example.org")
	note := NewActivityStreamsNote()
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("A simple note")
	note.SetActivityStreamsContent(content)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(a)
	example98Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsNote(note)
	example98Type.SetActivityStreamsObject(obj)
	return example98Type
}

const example99 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally liked a note",
  "type": "Like",
  "actor": "http://sally.example.org",
  "object": [
    "http://example.org/posts/1",
    {
      "type": "Note",
      "summary": "A simple note",
      "content": "That is a tree."
    }
  ]
}`

func example99Type() vocab.ActivityStreamsLike {
	example99Type := NewActivityStreamsLike()
	a := MustParseURL("http://sally.example.org")
	o := MustParseURL("http://example.org/posts/1")
	note := NewActivityStreamsNote()
	summaryNote := NewActivityStreamsSummaryProperty()
	summaryNote.AppendXMLSchemaString("A simple note")
	note.SetActivityStreamsSummary(summaryNote)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("That is a tree.")
	note.SetActivityStreamsContent(content)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally liked a note")
	example99Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(a)
	example99Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	obj.AppendActivityStreamsNote(note)
	example99Type.SetActivityStreamsObject(obj)
	return example99Type
}

const example100 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Page 1 of Sally's blog posts",
  "type": "CollectionPage",
  "prev": "http://example.org/collection?page=1",
  "items": [
    "http://example.org/posts/1",
    "http://example.org/posts/2",
    "http://example.org/posts/3"
  ]
}`

func example100Type() vocab.ActivityStreamsCollectionPage {
	example100Type := NewActivityStreamsCollectionPage()
	p := MustParseURL("http://example.org/collection?page=1")
	u1 := MustParseURL("http://example.org/posts/1")
	u2 := MustParseURL("http://example.org/posts/2")
	u3 := MustParseURL("http://example.org/posts/3")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Page 1 of Sally's blog posts")
	example100Type.SetActivityStreamsSummary(summary)
	prev := NewActivityStreamsPrevProperty()
	prev.SetIRI(p)
	example100Type.SetActivityStreamsPrev(prev)
	items := NewActivityStreamsItemsProperty()
	items.AppendIRI(u1)
	items.AppendIRI(u2)
	items.AppendIRI(u3)
	example100Type.SetActivityStreamsItems(items)
	return example100Type
}

const example101 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Page 1 of Sally's blog posts",
  "type": "CollectionPage",
  "prev": {
    "type": "Link",
    "name": "Previous Page",
    "href": "http://example.org/collection?page=1"
  },
  "items": [
    "http://example.org/posts/1",
    "http://example.org/posts/2",
    "http://example.org/posts/3"
  ]
}`

func example101Type() vocab.ActivityStreamsCollectionPage {
	example101Type := NewActivityStreamsCollectionPage()
	p := MustParseURL("http://example.org/collection?page=1")
	u1 := MustParseURL("http://example.org/posts/1")
	u2 := MustParseURL("http://example.org/posts/2")
	u3 := MustParseURL("http://example.org/posts/3")
	link := NewActivityStreamsLink()
	linkName := NewActivityStreamsNameProperty()
	linkName.AppendXMLSchemaString("Previous Page")
	link.SetActivityStreamsName(linkName)
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(p)
	link.SetActivityStreamsHref(hrefLink)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Page 1 of Sally's blog posts")
	example101Type.SetActivityStreamsSummary(summary)
	prev := NewActivityStreamsPrevProperty()
	prev.SetActivityStreamsLink(link)
	example101Type.SetActivityStreamsPrev(prev)
	items := NewActivityStreamsItemsProperty()
	items.AppendIRI(u1)
	items.AppendIRI(u2)
	items.AppendIRI(u3)
	example101Type.SetActivityStreamsItems(items)
	return example101Type
}

// NOTE: The 'url' field has added the 'type' property
const example102 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Video",
  "name": "Cool New Movie",
  "duration": "PT2H30M",
  "preview": {
    "type": "Video",
    "name": "Trailer",
    "duration": "PT1M",
    "url": {
      "type": "Link",
      "href": "http://example.org/trailer.mkv",
      "mediaType": "video/mkv"
    }
  }
}`

func example102Type() vocab.ActivityStreamsVideo {
	example102Type := NewActivityStreamsVideo()
	u := MustParseURL("http://example.org/trailer.mkv")
	link := NewActivityStreamsLink()
	mediaType := NewActivityStreamsMediaTypeProperty()
	mediaType.Set("video/mkv")
	link.SetActivityStreamsMediaType(mediaType)
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(u)
	link.SetActivityStreamsHref(hrefLink)
	video := NewActivityStreamsVideo()
	videoName := NewActivityStreamsNameProperty()
	videoName.AppendXMLSchemaString("Trailer")
	video.SetActivityStreamsName(videoName)
	durVideo := NewActivityStreamsDurationProperty()
	durVideo.Set(time.Minute)
	video.SetActivityStreamsDuration(durVideo)
	urlProperty := NewActivityStreamsUrlProperty()
	urlProperty.AppendActivityStreamsLink(link)
	video.SetActivityStreamsUrl(urlProperty)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Cool New Movie")
	example102Type.SetActivityStreamsName(name)
	dur := NewActivityStreamsDurationProperty()
	dur.Set(time.Hour*2 + time.Minute*30)
	example102Type.SetActivityStreamsDuration(dur)
	preview := NewActivityStreamsPreviewProperty()
	preview.AppendActivityStreamsVideo(video)
	example102Type.SetActivityStreamsPreview(preview)
	return example102Type
}

const example103 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally checked that her flight was on time",
  "type": ["Activity", "http://www.verbs.example/Check"],
  "actor": "http://sally.example.org",
  "object": "http://example.org/flights/1",
  "result": {
    "type": "http://www.types.example/flightstatus",
    "name": "On Time"
  }
}`

var example103Unknown = func(m map[string]interface{}) map[string]interface{} {
	m["type"] = []interface{}{
		m["type"],
		"http://www.verbs.example/Check",
	}
	m["result"] = map[string]interface{}{
		"type": "http://www.types.example/flightstatus",
		"name": "On Time",
	}
	return m
}

func example103Type() vocab.ActivityStreamsActivity {
	example103Type := NewActivityStreamsActivity()
	o := MustParseURL("http://example.org/flights/1")
	a := MustParseURL("http://sally.example.org")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally checked that her flight was on time")
	example103Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(a)
	example103Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	example103Type.SetActivityStreamsObject(obj)
	return example103Type
}

// NOTE: Changed to not be an array value for "items" to keep in line with other examples in spec!
const example104 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A simple note",
  "type": "Note",
  "id": "http://www.test.example/notes/1",
  "content": "I am fine.",
  "replies": {
    "type": "Collection",
    "totalItems": 1,
    "items": {
      "summary": "A response to the note",
      "type": "Note",
      "content": "I am glad to hear it.",
      "inReplyTo": "http://www.test.example/notes/1"
    }
  }
}`

func example104Type() vocab.ActivityStreamsNote {
	example104Type := NewActivityStreamsNote()
	i := MustParseURL("http://www.test.example/notes/1")
	note := NewActivityStreamsNote()
	summaryNote := NewActivityStreamsSummaryProperty()
	summaryNote.AppendXMLSchemaString("A response to the note")
	note.SetActivityStreamsSummary(summaryNote)
	contentNote := NewActivityStreamsContentProperty()
	contentNote.AppendXMLSchemaString("I am glad to hear it.")
	note.SetActivityStreamsContent(contentNote)
	inReplyTo := NewActivityStreamsInReplyToProperty()
	inReplyTo.AppendIRI(i)
	note.SetActivityStreamsInReplyTo(inReplyTo)
	replies := NewActivityStreamsCollection()
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(1)
	replies.SetActivityStreamsTotalItems(totalItems)
	items := NewActivityStreamsItemsProperty()
	items.AppendActivityStreamsNote(note)
	replies.SetActivityStreamsItems(items)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A simple note")
	example104Type.SetActivityStreamsSummary(summary)
	id := NewJSONLDIdProperty()
	id.Set(MustParseURL("http://www.test.example/notes/1"))
	example104Type.SetJSONLDId(id)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("I am fine.")
	example104Type.SetActivityStreamsContent(content)
	reply := NewActivityStreamsRepliesProperty()
	reply.SetActivityStreamsCollection(replies)
	example104Type.SetActivityStreamsReplies(reply)
	return example104Type
}

// NOTE: Changed to not be an array value for "tag" to keep in line with other examples in spec!
const example105 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Image",
  "summary": "Picture of Sally",
  "url": "http://example.org/sally.jpg",
  "tag": {
    "type": "Person",
    "id": "http://sally.example.org",
    "name": "Sally"
  }
}`

func example105Type() vocab.ActivityStreamsImage {
	example105Type := NewActivityStreamsImage()
	u := MustParseURL("http://example.org/sally.jpg")
	person := NewActivityStreamsPerson()
	id := NewJSONLDIdProperty()
	id.Set(MustParseURL("http://sally.example.org"))
	person.SetJSONLDId(id)
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Picture of Sally")
	example105Type.SetActivityStreamsSummary(summary)
	urlProp := NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(u)
	example105Type.SetActivityStreamsUrl(urlProp)
	tag := NewActivityStreamsTagProperty()
	tag.AppendActivityStreamsPerson(person)
	example105Type.SetActivityStreamsTag(tag)
	return example105Type
}

const example106 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally offered the post to John",
  "type": "Offer",
  "actor": "http://sally.example.org",
  "object": "http://example.org/posts/1",
  "target": "http://john.example.org"
}`

func example106Type() vocab.ActivityStreamsOffer {
	example106Type := NewActivityStreamsOffer()
	a := MustParseURL("http://sally.example.org")
	o := MustParseURL("http://example.org/posts/1")
	t := MustParseURL("http://john.example.org")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally offered the post to John")
	example106Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(a)
	example106Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	example106Type.SetActivityStreamsObject(obj)
	target := NewActivityStreamsTargetProperty()
	target.AppendIRI(t)
	example106Type.SetActivityStreamsTarget(target)
	return example106Type
}

const example107 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally offered the post to John",
  "type": "Offer",
  "actor": "http://sally.example.org",
  "object": "http://example.org/posts/1",
  "target": {
    "type": "Person",
    "name": "John"
  }
}`

func example107Type() vocab.ActivityStreamsOffer {
	example107Type := NewActivityStreamsOffer()
	a := MustParseURL("http://sally.example.org")
	o := MustParseURL("http://example.org/posts/1")
	person := NewActivityStreamsPerson()
	personName := NewActivityStreamsNameProperty()
	personName.AppendXMLSchemaString("John")
	person.SetActivityStreamsName(personName)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally offered the post to John")
	example107Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(a)
	example107Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	example107Type.SetActivityStreamsObject(obj)
	tobj := NewActivityStreamsTargetProperty()
	tobj.AppendActivityStreamsPerson(person)
	example107Type.SetActivityStreamsTarget(tobj)
	return example107Type
}

// NOTE: Changed to not be an array value for "to" to keep in line with other examples in spec!
const example108 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally offered the post to John",
  "type": "Offer",
  "actor": "http://sally.example.org",
  "object": "http://example.org/posts/1",
  "target": "http://john.example.org",
  "to": "http://joe.example.org"
}`

func example108Type() vocab.ActivityStreamsOffer {
	example108Type := NewActivityStreamsOffer()
	a := MustParseURL("http://sally.example.org")
	o := MustParseURL("http://example.org/posts/1")
	t := MustParseURL("http://john.example.org")
	z := MustParseURL("http://joe.example.org")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally offered the post to John")
	example108Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(a)
	example108Type.SetActivityStreamsActor(objectActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(o)
	example108Type.SetActivityStreamsObject(obj)
	target := NewActivityStreamsTargetProperty()
	target.AppendIRI(t)
	example108Type.SetActivityStreamsTarget(target)
	to := NewActivityStreamsToProperty()
	to.AppendIRI(z)
	example108Type.SetActivityStreamsTo(to)
	return example108Type
}

const example109 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Document",
  "name": "4Q Sales Forecast",
  "url": "http://example.org/4q-sales-forecast.pdf"
}`

func example109Type() vocab.ActivityStreamsDocument {
	example109Type := NewActivityStreamsDocument()
	u := MustParseURL("http://example.org/4q-sales-forecast.pdf")
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("4Q Sales Forecast")
	example109Type.SetActivityStreamsName(name)
	urlProp := NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(u)
	example109Type.SetActivityStreamsUrl(urlProp)
	return example109Type
}

const example110 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Document",
  "name": "4Q Sales Forecast",
  "url": {
    "type": "Link",
    "href": "http://example.org/4q-sales-forecast.pdf"
  }
}`

func example110Type() vocab.ActivityStreamsDocument {
	example110Type := NewActivityStreamsDocument()
	u := MustParseURL("http://example.org/4q-sales-forecast.pdf")
	link := NewActivityStreamsLink()
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(u)
	link.SetActivityStreamsHref(hrefLink)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("4Q Sales Forecast")
	example110Type.SetActivityStreamsName(name)
	urlProperty := NewActivityStreamsUrlProperty()
	urlProperty.AppendActivityStreamsLink(link)
	example110Type.SetActivityStreamsUrl(urlProperty)
	return example110Type
}

const example111 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Document",
  "name": "4Q Sales Forecast",
  "url": [
    {
      "type": "Link",
      "href": "http://example.org/4q-sales-forecast.pdf",
      "mediaType": "application/pdf"
    },
    {
      "type": "Link",
      "href": "http://example.org/4q-sales-forecast.html",
      "mediaType": "text/html"
    }
  ]
}`

func example111Type() vocab.ActivityStreamsDocument {
	example111Type := NewActivityStreamsDocument()
	u1 := MustParseURL("http://example.org/4q-sales-forecast.pdf")
	u2 := MustParseURL("http://example.org/4q-sales-forecast.html")
	link1 := NewActivityStreamsLink()
	hrefLink1 := NewActivityStreamsHrefProperty()
	hrefLink1.Set(u1)
	link1.SetActivityStreamsHref(hrefLink1)
	mediaType1 := NewActivityStreamsMediaTypeProperty()
	mediaType1.Set("application/pdf")
	link1.SetActivityStreamsMediaType(mediaType1)
	link2 := NewActivityStreamsLink()
	hrefLink2 := NewActivityStreamsHrefProperty()
	hrefLink2.Set(u2)
	link2.SetActivityStreamsHref(hrefLink2)
	mediaType2 := NewActivityStreamsMediaTypeProperty()
	mediaType2.Set("text/html")
	link2.SetActivityStreamsMediaType(mediaType2)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("4Q Sales Forecast")
	example111Type.SetActivityStreamsName(name)
	urlProperty := NewActivityStreamsUrlProperty()
	urlProperty.AppendActivityStreamsLink(link1)
	urlProperty.AppendActivityStreamsLink(link2)
	example111Type.SetActivityStreamsUrl(urlProperty)
	return example111Type
}

const example112 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "name": "Liu Gu Lu Cun, Pingdu, Qingdao, Shandong, China",
  "type": "Place",
  "latitude": 36.75,
  "longitude": 119.7667,
  "accuracy": 94.5
}`

func example112Type() vocab.ActivityStreamsPlace {
	example112Type := NewActivityStreamsPlace()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Liu Gu Lu Cun, Pingdu, Qingdao, Shandong, China")
	example112Type.SetActivityStreamsName(name)
	lat := NewActivityStreamsLatitudeProperty()
	lat.Set(36.75)
	example112Type.SetActivityStreamsLatitude(lat)
	lon := NewActivityStreamsLongitudeProperty()
	lon.Set(119.7667)
	example112Type.SetActivityStreamsLongitude(lon)
	acc := NewActivityStreamsAccuracyProperty()
	acc.Set(94.5)
	example112Type.SetActivityStreamsAccuracy(acc)
	return example112Type
}

const example113 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Place",
  "name": "Fresno Area",
  "altitude": 15.0,
  "latitude": 36.75,
  "longitude": 119.7667,
  "units": "miles"
}`

func example113Type() vocab.ActivityStreamsPlace {
	example113Type := NewActivityStreamsPlace()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Fresno Area")
	example113Type.SetActivityStreamsName(name)
	alt := NewActivityStreamsAltitudeProperty()
	alt.Set(15.0)
	example113Type.SetActivityStreamsAltitude(alt)
	lat := NewActivityStreamsLatitudeProperty()
	lat.Set(36.75)
	example113Type.SetActivityStreamsLatitude(lat)
	lon := NewActivityStreamsLongitudeProperty()
	lon.Set(119.7667)
	example113Type.SetActivityStreamsLongitude(lon)
	units := NewActivityStreamsUnitsProperty()
	units.SetXMLSchemaString("miles")
	example113Type.SetActivityStreamsUnits(units)
	return example113Type
}

const example114 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A simple note",
  "type": "Note",
  "content": "A <em>simple</em> note"
}`

func example114Type() vocab.ActivityStreamsNote {
	example114Type := NewActivityStreamsNote()
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A simple note")
	example114Type.SetActivityStreamsSummary(summary)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("A <em>simple</em> note")
	example114Type.SetActivityStreamsContent(content)
	return example114Type
}

const example115 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A simple note",
  "type": "Note",
  "contentMap": {
    "en": "A <em>simple</em> note",
    "es": "Una nota <em>sencilla</em>",
    "zh-Hans": "一段<em>简单的</em>笔记"
  }
}`

func example115Type() vocab.ActivityStreamsNote {
	example115Type := NewActivityStreamsNote()
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A simple note")
	example115Type.SetActivityStreamsSummary(summary)
	content := NewActivityStreamsContentProperty()
	content.AppendRDFLangString(map[string]string{
		"en":      "A <em>simple</em> note",
		"es":      "Una nota <em>sencilla</em>",
		"zh-Hans": "一段<em>简单的</em>笔记",
	})
	example115Type.SetActivityStreamsContent(content)
	return example115Type
}

const example116 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A simple note",
  "type": "Note",
  "mediaType": "text/markdown",
  "content": "## A simple note\nA simple markdown ` + "`note`" + `"
}`

func example116Type() vocab.ActivityStreamsNote {
	example116Type := NewActivityStreamsNote()
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A simple note")
	example116Type.SetActivityStreamsSummary(summary)
	mediaType := NewActivityStreamsMediaTypeProperty()
	mediaType.Set("text/markdown")
	example116Type.SetActivityStreamsMediaType(mediaType)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("## A simple note\nA simple markdown `note`")
	example116Type.SetActivityStreamsContent(content)
	return example116Type
}

const example117 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Note",
  "name": "A simple note"
}`

func example117Type() vocab.ActivityStreamsNote {
	example117Type := NewActivityStreamsNote()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("A simple note")
	example117Type.SetActivityStreamsName(name)
	return example117Type
}

const example118 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Note",
  "nameMap": {
    "en": "A simple note",
    "es": "Una nota sencilla",
    "zh-Hans": "一段简单的笔记"
  }
}`

func example118Type() vocab.ActivityStreamsNote {
	example118Type := NewActivityStreamsNote()
	name := NewActivityStreamsNameProperty()
	name.AppendRDFLangString(map[string]string{
		"en":      "A simple note",
		"es":      "Una nota sencilla",
		"zh-Hans": "一段简单的笔记",
	})
	example118Type.SetActivityStreamsName(name)
	return example118Type
}

const example119 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Video",
  "name": "Birds Flying",
  "url": "http://example.org/video.mkv",
  "duration": "PT2H"
}`

func example119Type() vocab.ActivityStreamsVideo {
	example119Type := NewActivityStreamsVideo()
	u := MustParseURL("http://example.org/video.mkv")
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Birds Flying")
	example119Type.SetActivityStreamsName(name)
	urlProp := NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(u)
	example119Type.SetActivityStreamsUrl(urlProp)
	dur := NewActivityStreamsDurationProperty()
	dur.Set(time.Hour * 2)
	example119Type.SetActivityStreamsDuration(dur)
	return example119Type
}

const example120 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Link",
  "href": "http://example.org/image.png",
  "height": 100,
  "width": 100
}`

func example120Type() vocab.ActivityStreamsLink {
	example120Type := NewActivityStreamsLink()
	u := MustParseURL("http://example.org/image.png")
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(u)
	example120Type.SetActivityStreamsHref(hrefLink)
	width := NewActivityStreamsWidthProperty()
	width.Set(100)
	example120Type.SetActivityStreamsWidth(width)
	height := NewActivityStreamsHeightProperty()
	height.Set(100)
	example120Type.SetActivityStreamsHeight(height)
	return example120Type
}

const example121 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Link",
  "href": "http://example.org/abc",
  "mediaType": "text/html",
  "name": "Previous"
}`

func example121Type() vocab.ActivityStreamsLink {
	example121Type := NewActivityStreamsLink()
	u := MustParseURL("http://example.org/abc")
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(u)
	example121Type.SetActivityStreamsHref(hrefLink)
	mediaType := NewActivityStreamsMediaTypeProperty()
	mediaType.Set("text/html")
	example121Type.SetActivityStreamsMediaType(mediaType)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Previous")
	example121Type.SetActivityStreamsName(name)
	return example121Type
}

const example122 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Link",
  "href": "http://example.org/abc",
  "hreflang": "en",
  "mediaType": "text/html",
  "name": "Previous"
}`

func example122Type() vocab.ActivityStreamsLink {
	example122Type := NewActivityStreamsLink()
	u := MustParseURL("http://example.org/abc")
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(u)
	example122Type.SetActivityStreamsHref(hrefLink)
	mediaType := NewActivityStreamsMediaTypeProperty()
	mediaType.Set("text/html")
	example122Type.SetActivityStreamsMediaType(mediaType)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Previous")
	example122Type.SetActivityStreamsName(name)
	hreflang := NewActivityStreamsHreflangProperty()
	hreflang.Set("en")
	example122Type.SetActivityStreamsHreflang(hreflang)
	return example122Type
}

const example123 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Page 1 of Sally's notes",
  "type": "CollectionPage",
  "id": "http://example.org/collection?page=1",
  "partOf": "http://example.org/collection",
  "items": [
    {
      "type": "Note",
      "name": "Pizza Toppings to Try"
    },
    {
      "type": "Note",
      "name": "Thought about California"
    }
  ]
}`

func example123Type() vocab.ActivityStreamsCollectionPage {
	example123Type := NewActivityStreamsCollectionPage()
	u := MustParseURL("http://example.org/collection")
	note1 := NewActivityStreamsNote()
	name1 := NewActivityStreamsNameProperty()
	name1.AppendXMLSchemaString("Pizza Toppings to Try")
	note1.SetActivityStreamsName(name1)
	note2 := NewActivityStreamsNote()
	name2 := NewActivityStreamsNameProperty()
	name2.AppendXMLSchemaString("Thought about California")
	note2.SetActivityStreamsName(name2)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Page 1 of Sally's notes")
	example123Type.SetActivityStreamsSummary(summary)
	id := NewJSONLDIdProperty()
	id.Set(MustParseURL("http://example.org/collection?page=1"))
	example123Type.SetJSONLDId(id)
	partOf := NewActivityStreamsPartOfProperty()
	partOf.SetIRI(u)
	example123Type.SetActivityStreamsPartOf(partOf)
	items := NewActivityStreamsItemsProperty()
	items.AppendActivityStreamsNote(note1)
	items.AppendActivityStreamsNote(note2)
	example123Type.SetActivityStreamsItems(items)
	return example123Type
}

const example124 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Place",
  "name": "Fresno Area",
  "latitude": 36.75,
  "longitude": 119.7667,
  "radius": 15,
  "units": "miles"
}`

func example124Type() vocab.ActivityStreamsPlace {
	example124Type := NewActivityStreamsPlace()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Fresno Area")
	example124Type.SetActivityStreamsName(name)
	lat := NewActivityStreamsLatitudeProperty()
	lat.Set(36.75)
	example124Type.SetActivityStreamsLatitude(lat)
	lon := NewActivityStreamsLongitudeProperty()
	lon.Set(119.7667)
	example124Type.SetActivityStreamsLongitude(lon)
	rad := NewActivityStreamsRadiusProperty()
	rad.Set(15)
	example124Type.SetActivityStreamsRadius(rad)
	units := NewActivityStreamsUnitsProperty()
	units.SetXMLSchemaString("miles")
	example124Type.SetActivityStreamsUnits(units)
	return example124Type
}

const example125 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Place",
  "name": "Fresno Area",
  "latitude": 36.75,
  "longitude": 119.7667,
  "radius": 15,
  "units": "miles"
}`

func example125Type() vocab.ActivityStreamsPlace {
	example125Type := NewActivityStreamsPlace()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Fresno Area")
	example125Type.SetActivityStreamsName(name)
	lat := NewActivityStreamsLatitudeProperty()
	lat.Set(36.75)
	example125Type.SetActivityStreamsLatitude(lat)
	lon := NewActivityStreamsLongitudeProperty()
	lon.Set(119.7667)
	example125Type.SetActivityStreamsLongitude(lon)
	rad := NewActivityStreamsRadiusProperty()
	rad.Set(15)
	example125Type.SetActivityStreamsRadius(rad)
	units := NewActivityStreamsUnitsProperty()
	units.SetXMLSchemaString("miles")
	example125Type.SetActivityStreamsUnits(units)
	return example125Type
}

const example126 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Link",
  "href": "http://example.org/abc",
  "hreflang": "en",
  "mediaType": "text/html",
  "name": "Next"
}`

func example126Type() vocab.ActivityStreamsLink {
	example126Type := NewActivityStreamsLink()
	u := MustParseURL("http://example.org/abc")
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(u)
	example126Type.SetActivityStreamsHref(hrefLink)
	hreflang := NewActivityStreamsHreflangProperty()
	hreflang.Set("en")
	example126Type.SetActivityStreamsHreflang(hreflang)
	mediaType := NewActivityStreamsMediaTypeProperty()
	mediaType.Set("text/html")
	example126Type.SetActivityStreamsMediaType(mediaType)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Next")
	example126Type.SetActivityStreamsName(name)
	return example126Type
}

const example127 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Event",
  "name": "Going-Away Party for Jim",
  "startTime": "2014-12-31T23:00:00-08:00",
  "endTime": "2015-01-01T06:00:00-08:00"
}`

func example127Type() vocab.ActivityStreamsEvent {
	example127Type := NewActivityStreamsEvent()
	t1, err := time.Parse(time.RFC3339, "2014-12-31T23:00:00-08:00")
	if err != nil {
		panic(err)
	}
	t2, err := time.Parse(time.RFC3339, "2015-01-01T06:00:00-08:00")
	if err != nil {
		panic(err)
	}
	goingAway := NewActivityStreamsNameProperty()
	goingAway.AppendXMLSchemaString("Going-Away Party for Jim")
	example127Type.SetActivityStreamsName(goingAway)
	startTime := NewActivityStreamsStartTimeProperty()
	startTime.Set(t1)
	example127Type.SetActivityStreamsStartTime(startTime)
	endTime := NewActivityStreamsEndTimeProperty()
	endTime.Set(t2)
	example127Type.SetActivityStreamsEndTime(endTime)
	return example127Type
}

const example128 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "A simple note",
  "type": "Note",
  "content": "Fish swim.",
  "published": "2014-12-12T12:12:12Z"
}`

func example128Type() vocab.ActivityStreamsNote {
	example128Type := NewActivityStreamsNote()
	t, err := time.Parse(time.RFC3339, "2014-12-12T12:12:12Z")
	if err != nil {
		panic(err)
	}
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A simple note")
	example128Type.SetActivityStreamsSummary(summary)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("Fish swim.")
	example128Type.SetActivityStreamsContent(content)
	published := NewActivityStreamsPublishedProperty()
	published.Set(t)
	example128Type.SetActivityStreamsPublished(published)
	return example128Type
}

const example129 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Event",
  "name": "Going-Away Party for Jim",
  "startTime": "2014-12-31T23:00:00-08:00",
  "endTime": "2015-01-01T06:00:00-08:00"
}`

func example129Type() vocab.ActivityStreamsEvent {
	example129Type := NewActivityStreamsEvent()
	t1, err := time.Parse(time.RFC3339, "2014-12-31T23:00:00-08:00")
	if err != nil {
		panic(err)
	}
	t2, err := time.Parse(time.RFC3339, "2015-01-01T06:00:00-08:00")
	if err != nil {
		panic(err)
	}
	goingAway := NewActivityStreamsNameProperty()
	goingAway.AppendXMLSchemaString("Going-Away Party for Jim")
	example129Type.SetActivityStreamsName(goingAway)
	startTime := NewActivityStreamsStartTimeProperty()
	startTime.Set(t1)
	example129Type.SetActivityStreamsStartTime(startTime)
	endTime := NewActivityStreamsEndTimeProperty()
	endTime.Set(t2)
	example129Type.SetActivityStreamsEndTime(endTime)
	return example129Type
}

const example130 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Place",
  "name": "Fresno Area",
  "latitude": 36.75,
  "longitude": 119.7667,
  "radius": 15,
  "units": "miles"
}`

func example130Type() vocab.ActivityStreamsPlace {
	example130Type := NewActivityStreamsPlace()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Fresno Area")
	example130Type.SetActivityStreamsName(name)
	lat := NewActivityStreamsLatitudeProperty()
	lat.Set(36.75)
	example130Type.SetActivityStreamsLatitude(lat)
	lon := NewActivityStreamsLongitudeProperty()
	lon.Set(119.7667)
	example130Type.SetActivityStreamsLongitude(lon)
	rad := NewActivityStreamsRadiusProperty()
	rad.Set(15)
	example130Type.SetActivityStreamsRadius(rad)
	units := NewActivityStreamsUnitsProperty()
	units.SetXMLSchemaString("miles")
	example130Type.SetActivityStreamsUnits(units)
	return example130Type
}

const example131 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Link",
  "href": "http://example.org/abc",
  "hreflang": "en",
  "mediaType": "text/html",
  "name": "Preview",
  "rel": ["canonical", "preview"]
}`

func example131Type() vocab.ActivityStreamsLink {
	example131Type := NewActivityStreamsLink()
	u := MustParseURL("http://example.org/abc")
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(u)
	example131Type.SetActivityStreamsHref(hrefLink)
	hreflang := NewActivityStreamsHreflangProperty()
	hreflang.Set("en")
	example131Type.SetActivityStreamsHreflang(hreflang)
	mediaType := NewActivityStreamsMediaTypeProperty()
	mediaType.Set("text/html")
	example131Type.SetActivityStreamsMediaType(mediaType)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Preview")
	example131Type.SetActivityStreamsName(name)
	rel := NewActivityStreamsRelProperty()
	rel.AppendRFCRfc5988("canonical")
	rel.AppendRFCRfc5988("preview")
	example131Type.SetActivityStreamsRel(rel)
	return example131Type
}

const example132 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Page 1 of Sally's notes",
  "type": "OrderedCollectionPage",
  "startIndex": 0,
  "orderedItems": [
    {
      "type": "Note",
      "name": "Density of Water"
    },
    {
      "type": "Note",
      "name": "Air Mattress Idea"
    }
  ]
}`

func example132Type() vocab.ActivityStreamsOrderedCollectionPage {
	example132Type := NewActivityStreamsOrderedCollectionPage()
	note1 := NewActivityStreamsNote()
	name1 := NewActivityStreamsNameProperty()
	name1.AppendXMLSchemaString("Density of Water")
	note1.SetActivityStreamsName(name1)
	note2 := NewActivityStreamsNote()
	name2 := NewActivityStreamsNameProperty()
	name2.AppendXMLSchemaString("Air Mattress Idea")
	note2.SetActivityStreamsName(name2)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Page 1 of Sally's notes")
	example132Type.SetActivityStreamsSummary(summary)
	start := NewActivityStreamsStartIndexProperty()
	start.Set(0)
	example132Type.SetActivityStreamsStartIndex(start)
	orderedItems := NewActivityStreamsOrderedItemsProperty()
	orderedItems.AppendActivityStreamsNote(note1)
	orderedItems.AppendActivityStreamsNote(note2)
	example132Type.SetActivityStreamsOrderedItems(orderedItems)
	return example132Type
}

const example133 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "name": "Cane Sugar Processing",
  "type": "Note",
  "summary": "A simple <em>note</em>"
}`

func example133Type() vocab.ActivityStreamsNote {
	example133Type := NewActivityStreamsNote()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Cane Sugar Processing")
	example133Type.SetActivityStreamsName(name)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("A simple <em>note</em>")
	example133Type.SetActivityStreamsSummary(summary)
	return example133Type
}

const example134 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "name": "Cane Sugar Processing",
  "type": "Note",
  "summaryMap": {
    "en": "A simple <em>note</em>",
    "es": "Una <em>nota</em> sencilla",
    "zh-Hans": "一段<em>简单的</em>笔记"
  }
}`

func example134Type() vocab.ActivityStreamsNote {
	example134Type := NewActivityStreamsNote()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Cane Sugar Processing")
	example134Type.SetActivityStreamsName(name)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendRDFLangString(map[string]string{
		"en":      "A simple <em>note</em>",
		"es":      "Una <em>nota</em> sencilla",
		"zh-Hans": "一段<em>简单的</em>笔记",
	})
	example134Type.SetActivityStreamsSummary(summary)
	return example134Type
}

const example135 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally's notes",
  "type": "Collection",
  "totalItems": 2,
  "items": [
    {
      "type": "Note",
      "name": "Which Staircase Should I Use"
    },
    {
      "type": "Note",
      "name": "Something to Remember"
    }
  ]
}`

func example135Type() vocab.ActivityStreamsCollection {
	example135Type := NewActivityStreamsCollection()
	note1 := NewActivityStreamsNote()
	name1 := NewActivityStreamsNameProperty()
	name1.AppendXMLSchemaString("Which Staircase Should I Use")
	note1.SetActivityStreamsName(name1)
	note2 := NewActivityStreamsNote()
	name2 := NewActivityStreamsNameProperty()
	name2.AppendXMLSchemaString("Something to Remember")
	note2.SetActivityStreamsName(name2)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally's notes")
	example135Type.SetActivityStreamsSummary(summary)
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(2)
	example135Type.SetActivityStreamsTotalItems(totalItems)
	items := NewActivityStreamsItemsProperty()
	items.AppendActivityStreamsNote(note1)
	items.AppendActivityStreamsNote(note2)
	example135Type.SetActivityStreamsItems(items)
	return example135Type
}

const example136 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Place",
  "name": "Fresno Area",
  "latitude": 36.75,
  "longitude": 119.7667,
  "radius": 15,
  "units": "miles"
}`

func example136Type() vocab.ActivityStreamsPlace {
	example136Type := NewActivityStreamsPlace()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Fresno Area")
	example136Type.SetActivityStreamsName(name)
	lat := NewActivityStreamsLatitudeProperty()
	lat.Set(36.75)
	example136Type.SetActivityStreamsLatitude(lat)
	lon := NewActivityStreamsLongitudeProperty()
	lon.Set(119.7667)
	example136Type.SetActivityStreamsLongitude(lon)
	rad := NewActivityStreamsRadiusProperty()
	rad.Set(15)
	example136Type.SetActivityStreamsRadius(rad)
	units := NewActivityStreamsUnitsProperty()
	units.SetXMLSchemaString("miles")
	example136Type.SetActivityStreamsUnits(units)
	return example136Type
}

const example137 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "name": "Cranberry Sauce Idea",
  "type": "Note",
  "content": "Mush it up so it does not have the same shape as the can.",
  "updated": "2014-12-12T12:12:12Z"
}`

func example137Type() vocab.ActivityStreamsNote {
	example137Type := NewActivityStreamsNote()
	t, err := time.Parse(time.RFC3339, "2014-12-12T12:12:12Z")
	if err != nil {
		panic(err)
	}
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Cranberry Sauce Idea")
	example137Type.SetActivityStreamsName(name)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("Mush it up so it does not have the same shape as the can.")
	example137Type.SetActivityStreamsContent(content)
	updated := NewActivityStreamsUpdatedProperty()
	updated.Set(t)
	example137Type.SetActivityStreamsUpdated(updated)
	return example137Type
}

const example138 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Link",
  "href": "http://example.org/image.png",
  "height": 100,
  "width": 100
}`

func example138Type() vocab.ActivityStreamsLink {
	example138Type := NewActivityStreamsLink()
	u := MustParseURL("http://example.org/image.png")
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(u)
	example138Type.SetActivityStreamsHref(hrefLink)
	width := NewActivityStreamsWidthProperty()
	width.Set(100)
	example138Type.SetActivityStreamsWidth(width)
	height := NewActivityStreamsHeightProperty()
	height.Set(100)
	example138Type.SetActivityStreamsHeight(height)
	return example138Type
}

const example139 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally is an acquaintance of John's",
  "type": "Relationship",
  "subject": {
    "type": "Person",
    "name": "Sally"
  },
  "relationship": "http://purl.org/vocab/relationship/acquaintanceOf",
  "object": {
    "type": "Person",
    "name": "John"
  }
}`

func example139Type() vocab.ActivityStreamsRelationship {
	example139Type := NewActivityStreamsRelationship()
	u := MustParseURL("http://purl.org/vocab/relationship/acquaintanceOf")
	subject := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	subject.SetActivityStreamsName(sally)
	object := NewActivityStreamsPerson()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("John")
	object.SetActivityStreamsName(name)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally is an acquaintance of John's")
	example139Type.SetActivityStreamsSummary(summary)
	subj := NewActivityStreamsSubjectProperty()
	subj.SetActivityStreamsPerson(subject)
	example139Type.SetActivityStreamsSubject(subj)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsPerson(object)
	example139Type.SetActivityStreamsObject(obj)
	relationship := NewActivityStreamsRelationshipProperty()
	relationship.AppendIRI(u)
	example139Type.SetActivityStreamsRelationship(relationship)
	return example139Type
}

const example140 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally is an acquaintance of John's",
  "type": "Relationship",
  "subject": {
    "type": "Person",
    "name": "Sally"
  },
  "relationship": "http://purl.org/vocab/relationship/acquaintanceOf",
  "object": {
    "type": "Person",
    "name": "John"
  }
}`

func example140Type() vocab.ActivityStreamsRelationship {
	example140Type := NewActivityStreamsRelationship()
	u := MustParseURL("http://purl.org/vocab/relationship/acquaintanceOf")
	subject := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	subject.SetActivityStreamsName(sally)
	object := NewActivityStreamsPerson()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("John")
	object.SetActivityStreamsName(name)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally is an acquaintance of John's")
	example140Type.SetActivityStreamsSummary(summary)
	subj := NewActivityStreamsSubjectProperty()
	subj.SetActivityStreamsPerson(subject)
	example140Type.SetActivityStreamsSubject(subj)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsPerson(object)
	example140Type.SetActivityStreamsObject(obj)
	relationship := NewActivityStreamsRelationshipProperty()
	relationship.AppendIRI(u)
	example140Type.SetActivityStreamsRelationship(relationship)
	return example140Type
}

const example141 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally's profile",
  "type": "Profile",
  "describes": {
    "type": "Person",
    "name": "Sally"
  },
  "url": "http://sally.example.org"
}`

func example141Type() vocab.ActivityStreamsProfile {
	example141Type := NewActivityStreamsProfile()
	u := MustParseURL("http://sally.example.org")
	person := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally's profile")
	example141Type.SetActivityStreamsSummary(summary)
	urlProp := NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(u)
	example141Type.SetActivityStreamsUrl(urlProp)
	describes := NewActivityStreamsDescribesProperty()
	describes.SetActivityStreamsPerson(person)
	example141Type.SetActivityStreamsDescribes(describes)
	return example141Type
}

const example142 = `{
"@context": "https://www.w3.org/ns/activitystreams",
"summary": "This image has been deleted",
"type": "Tombstone",
"formerType": "Image",
"url": "http://example.org/image/2"
}`

func example142Type() vocab.ActivityStreamsTombstone {
	example142Type := NewActivityStreamsTombstone()
	u := MustParseURL("http://example.org/image/2")
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("This image has been deleted")
	example142Type.SetActivityStreamsSummary(summary)
	former := NewActivityStreamsFormerTypeProperty()
	former.AppendXMLSchemaString("Image")
	example142Type.SetActivityStreamsFormerType(former)
	urlProp := NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(u)
	example142Type.SetActivityStreamsUrl(urlProp)
	return example142Type
}

const example143 = `{
"@context": "https://www.w3.org/ns/activitystreams",
"summary": "This image has been deleted",
"type": "Tombstone",
"deleted": "2016-05-03T00:00:00Z"
}`

func example143Type() vocab.ActivityStreamsTombstone {
	example143Type := NewActivityStreamsTombstone()
	t, err := time.Parse(time.RFC3339, "2016-05-03T00:00:00Z")
	if err != nil {
		panic(err)
	}
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("This image has been deleted")
	example143Type.SetActivityStreamsSummary(summary)
	deleted := NewActivityStreamsDeletedProperty()
	deleted.Set(t)
	example143Type.SetActivityStreamsDeleted(deleted)
	return example143Type
}

const example144 = `{
 "@context": "https://www.w3.org/ns/activitystreams",
 "summary": "Activities in Project XYZ",
 "type": "Collection",
 "items": [
   {
     "summary": "Sally created a note",
     "type": "Create",
     "id": "http://activities.example.com/1",
     "actor": "http://sally.example.org",
     "object": {
       "summary": "A note",
       "type": "Note",
       "id": "http://notes.example.com/1",
       "content": "A note"
     },
     "context": {
       "type": "http://example.org/Project",
       "name": "Project XYZ"
     },
     "audience": {
       "type": "Group",
       "name": "Project XYZ Working Group"
     },
     "to": "http://john.example.org"
   },
   {
     "summary": "John liked Sally's note",
     "type": "Like",
     "id": "http://activities.example.com/1",
     "actor": "http://john.example.org",
     "object": "http://notes.example.com/1",
     "context": {
       "type": "http://example.org/Project",
       "name": "Project XYZ"
     },
     "audience": {
       "type": "Group",
       "name": "Project XYZ Working Group"
     },
     "to": "http://sally.example.org"
   }
 ]
}`

var example144Unknown = func(m map[string]interface{}) map[string]interface{} {
	items := m["items"].([]interface{})
	create := items[0].(map[string]interface{})
	like := items[1].(map[string]interface{})
	create["context"] = map[string]interface{}{
		"type": "http://example.org/Project",
		"name": "Project XYZ",
	}
	like["context"] = map[string]interface{}{
		"type": "http://example.org/Project",
		"name": "Project XYZ",
	}
	m["items"] = []interface{}{create, like}
	return m
}

func example144Type() vocab.ActivityStreamsCollection {
	example144Type := NewActivityStreamsCollection()
	sally := MustParseURL("http://sally.example.org")
	john := MustParseURL("http://john.example.org")
	o := MustParseURL("http://notes.example.com/1")
	audience := NewActivityStreamsGroup()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Project XYZ Working Group")
	audience.SetActivityStreamsName(name)
	note := NewActivityStreamsNote()
	summaryNote := NewActivityStreamsSummaryProperty()
	summaryNote.AppendXMLSchemaString("A note")
	note.SetActivityStreamsSummary(summaryNote)
	noteId := NewJSONLDIdProperty()
	noteId.Set(MustParseURL("http://notes.example.com/1"))
	note.SetJSONLDId(noteId)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("A note")
	note.SetActivityStreamsContent(content)
	create := NewActivityStreamsCreate()
	summaryCreate := NewActivityStreamsSummaryProperty()
	summaryCreate.AppendXMLSchemaString("Sally created a note")
	create.SetActivityStreamsSummary(summaryCreate)
	createId := NewJSONLDIdProperty()
	createId.Set(MustParseURL("http://activities.example.com/1"))
	create.SetJSONLDId(createId)
	createActor := NewActivityStreamsActorProperty()
	createActor.AppendIRI(sally)
	create.SetActivityStreamsActor(createActor)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsNote(note)
	create.SetActivityStreamsObject(obj)
	aud1 := NewActivityStreamsAudienceProperty()
	aud1.AppendActivityStreamsGroup(audience)
	create.SetActivityStreamsAudience(aud1)
	toCreate := NewActivityStreamsToProperty()
	toCreate.AppendIRI(john)
	create.SetActivityStreamsTo(toCreate)
	like := NewActivityStreamsLike()
	summaryLike := NewActivityStreamsSummaryProperty()
	summaryLike.AppendXMLSchemaString("John liked Sally's note")
	like.SetActivityStreamsSummary(summaryLike)
	likeId := NewJSONLDIdProperty()
	likeId.Set(MustParseURL("http://activities.example.com/1"))
	like.SetJSONLDId(likeId)
	likeActor := NewActivityStreamsActorProperty()
	likeActor.AppendIRI(john)
	like.SetActivityStreamsActor(likeActor)
	objLike := NewActivityStreamsObjectProperty()
	objLike.AppendIRI(o)
	like.SetActivityStreamsObject(objLike)
	aud2 := NewActivityStreamsAudienceProperty()
	aud2.AppendActivityStreamsGroup(audience)
	like.SetActivityStreamsAudience(aud2)
	toLike := NewActivityStreamsToProperty()
	toLike.AppendIRI(sally)
	like.SetActivityStreamsTo(toLike)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Activities in Project XYZ")
	example144Type.SetActivityStreamsSummary(summary)
	items := NewActivityStreamsItemsProperty()
	items.AppendActivityStreamsCreate(create)
	items.AppendActivityStreamsLike(like)
	example144Type.SetActivityStreamsItems(items)
	return example144Type
}

const example145 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally's friends list",
  "type": "Collection",
  "items": [
    {
      "summary": "Sally is influenced by Joe",
      "type": "Relationship",
      "subject": {
        "type": "Person",
        "name": "Sally"
      },
      "relationship": "http://purl.org/vocab/relationship/influencedBy",
      "object": {
        "type": "Person",
        "name": "Joe"
      }
    },
    {
      "summary": "Sally is a friend of Jane",
      "type": "Relationship",
      "subject": {
        "type": "Person",
        "name": "Sally"
      },
      "relationship": "http://purl.org/vocab/relationship/friendOf",
      "object": {
        "type": "Person",
        "name": "Jane"
      }
    }
  ]
}`

func example145Type() vocab.ActivityStreamsCollection {
	example145Type := NewActivityStreamsCollection()
	friend := MustParseURL("http://purl.org/vocab/relationship/friendOf")
	influenced := MustParseURL("http://purl.org/vocab/relationship/influencedBy")
	sally := NewActivityStreamsPerson()
	sallyName := NewActivityStreamsNameProperty()
	sallyName.AppendXMLSchemaString("Sally")
	sally.SetActivityStreamsName(sallyName)
	jane := NewActivityStreamsPerson()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("Jane")
	jane.SetActivityStreamsName(name)
	joe := NewActivityStreamsPerson()
	joeName := NewActivityStreamsNameProperty()
	joeName.AppendXMLSchemaString("Joe")
	joe.SetActivityStreamsName(joeName)
	joeRel := NewActivityStreamsRelationship()
	summaryJoe := NewActivityStreamsSummaryProperty()
	summaryJoe.AppendXMLSchemaString("Sally is influenced by Joe")
	joeRel.SetActivityStreamsSummary(summaryJoe)
	subj1 := NewActivityStreamsSubjectProperty()
	subj1.SetActivityStreamsPerson(sally)
	joeRel.SetActivityStreamsSubject(subj1)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsPerson(joe)
	joeRel.SetActivityStreamsObject(obj)
	relationship := NewActivityStreamsRelationshipProperty()
	relationship.AppendIRI(influenced)
	joeRel.SetActivityStreamsRelationship(relationship)
	janeRel := NewActivityStreamsRelationship()
	summaryJane := NewActivityStreamsSummaryProperty()
	summaryJane.AppendXMLSchemaString("Sally is a friend of Jane")
	janeRel.SetActivityStreamsSummary(summaryJane)
	subj2 := NewActivityStreamsSubjectProperty()
	subj2.SetActivityStreamsPerson(sally)
	janeRel.SetActivityStreamsSubject(subj2)
	objJane := NewActivityStreamsObjectProperty()
	objJane.AppendActivityStreamsPerson(jane)
	janeRel.SetActivityStreamsObject(objJane)
	relationshipJane := NewActivityStreamsRelationshipProperty()
	relationshipJane.AppendIRI(friend)
	janeRel.SetActivityStreamsRelationship(relationshipJane)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally's friends list")
	example145Type.SetActivityStreamsSummary(summary)
	items := NewActivityStreamsItemsProperty()
	items.AppendActivityStreamsRelationship(joeRel)
	items.AppendActivityStreamsRelationship(janeRel)
	example145Type.SetActivityStreamsItems(items)
	return example145Type
}

// NOTE: Added `Z` to `startTime` to make align to spec!
const example146 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally became a friend of Matt",
  "type": "Create",
  "actor": "http://sally.example.org",
  "object": {
    "type": "Relationship",
    "subject": "http://sally.example.org",
    "relationship": "http://purl.org/vocab/relationship/friendOf",
    "object": "http://matt.example.org",
    "startTime": "2015-04-21T12:34:56Z"
  }
}`

func example146Type() vocab.ActivityStreamsCreate {
	example146Type := NewActivityStreamsCreate()
	friend := MustParseURL("http://purl.org/vocab/relationship/friendOf")
	m := MustParseURL("http://matt.example.org")
	s := MustParseURL("http://sally.example.org")
	t, err := time.Parse(time.RFC3339, "2015-04-21T12:34:56Z")
	if err != nil {
		panic(err)
	}
	relationship := NewActivityStreamsRelationship()
	subj := NewActivityStreamsSubjectProperty()
	subj.SetIRI(s)
	relationship.SetActivityStreamsSubject(subj)
	friendRel := NewActivityStreamsRelationshipProperty()
	friendRel.AppendIRI(friend)
	relationship.SetActivityStreamsRelationship(friendRel)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(m)
	relationship.SetActivityStreamsObject(obj)
	startTime := NewActivityStreamsStartTimeProperty()
	startTime.Set(t)
	relationship.SetActivityStreamsStartTime(startTime)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally became a friend of Matt")
	example146Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(s)
	example146Type.SetActivityStreamsActor(objectActor)
	objRoot := NewActivityStreamsObjectProperty()
	objRoot.AppendActivityStreamsRelationship(relationship)
	example146Type.SetActivityStreamsObject(objRoot)
	return example146Type
}

const example147 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://example.org/connection-requests/123",
  "summary": "Sally requested to be a friend of John",
  "type": "Offer",
  "actor": "acct:sally@example.org",
  "object": {
    "summary": "Sally and John's friendship",
    "id": "http://example.org/connections/123",
    "type": "Relationship",
    "subject": "acct:sally@example.org",
    "relationship": "http://purl.org/vocab/relationship/friendOf",
    "object": "acct:john@example.org"
  },
  "target": "acct:john@example.org"
}`

func example147Type() vocab.ActivityStreamsOffer {
	example147Type := NewActivityStreamsOffer()
	friend := MustParseURL("http://purl.org/vocab/relationship/friendOf")
	s := MustParseURL("acct:sally@example.org")
	t := MustParseURL("acct:john@example.org")
	rel := NewActivityStreamsRelationship()
	summaryRel := NewActivityStreamsSummaryProperty()
	summaryRel.AppendXMLSchemaString("Sally and John's friendship")
	rel.SetActivityStreamsSummary(summaryRel)
	relId := NewJSONLDIdProperty()
	relId.Set(MustParseURL("http://example.org/connections/123"))
	rel.SetJSONLDId(relId)
	subj := NewActivityStreamsSubjectProperty()
	subj.SetIRI(s)
	rel.SetActivityStreamsSubject(subj)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(t)
	rel.SetActivityStreamsObject(obj)
	relationship := NewActivityStreamsRelationshipProperty()
	relationship.AppendIRI(friend)
	rel.SetActivityStreamsRelationship(relationship)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally requested to be a friend of John")
	example147Type.SetActivityStreamsSummary(summary)
	id := NewJSONLDIdProperty()
	id.Set(MustParseURL("http://example.org/connection-requests/123"))
	example147Type.SetJSONLDId(id)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(s)
	example147Type.SetActivityStreamsActor(objectActor)
	objRel := NewActivityStreamsObjectProperty()
	objRel.AppendActivityStreamsRelationship(rel)
	example147Type.SetActivityStreamsObject(objRel)
	objRoot := NewActivityStreamsObjectProperty()
	objRoot.AppendActivityStreamsRelationship(rel)
	example147Type.SetActivityStreamsObject(objRoot)
	target := NewActivityStreamsTargetProperty()
	target.AppendIRI(t)
	example147Type.SetActivityStreamsTarget(target)
	return example147Type
}

const example148 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally and John's relationship history",
  "type": "Collection",
  "items": [
    {
      "summary": "John accepted Sally's friend request",
      "id": "http://example.org/activities/122",
      "type": "Accept",
      "actor": "acct:john@example.org",
      "object": "http://example.org/connection-requests/123",
      "inReplyTo": "http://example.org/connection-requests/123",
      "context": "http://example.org/connections/123",
      "result": [
        "http://example.org/activities/123",
        "http://example.org/activities/124",
        "http://example.org/activities/125",
        "http://example.org/activities/126"
      ]
    },
    {
      "summary": "John followed Sally",
      "id": "http://example.org/activities/123",
      "type": "Follow",
      "actor": "acct:john@example.org",
      "object": "acct:sally@example.org",
      "context": "http://example.org/connections/123"
    },
    {
      "summary": "Sally followed John",
      "id": "http://example.org/activities/124",
      "type": "Follow",
      "actor": "acct:sally@example.org",
      "object": "acct:john@example.org",
      "context": "http://example.org/connections/123"
    },
    {
      "summary": "John added Sally to his friends list",
      "id": "http://example.org/activities/125",
      "type": "Add",
      "actor": "acct:john@example.org",
      "object": "http://example.org/connections/123",
      "target": {
        "type": "Collection",
        "summary": "John's Connections"
      },
      "context": "http://example.org/connections/123"
    },
    {
      "summary": "Sally added John to her friends list",
      "id": "http://example.org/activities/126",
      "type": "Add",
      "actor": "acct:sally@example.org",
      "object": "http://example.org/connections/123",
      "target": {
        "type": "Collection",
        "summary": "Sally's Connections"
      },
      "context": "http://example.org/connections/123"
    }
  ]
}`

func example148Type() vocab.ActivityStreamsCollection {
	example148Type := NewActivityStreamsCollection()
	john := MustParseURL("acct:john@example.org")
	sally := MustParseURL("acct:sally@example.org")
	req123 := MustParseURL("http://example.org/connection-requests/123")
	conn123 := MustParseURL("http://example.org/connections/123")
	a123 := MustParseURL("http://example.org/activities/123")
	a124 := MustParseURL("http://example.org/activities/124")
	a125 := MustParseURL("http://example.org/activities/125")
	a126 := MustParseURL("http://example.org/activities/126")
	jc := NewActivityStreamsCollection()
	summaryJc := NewActivityStreamsSummaryProperty()
	summaryJc.AppendXMLSchemaString("John's Connections")
	jc.SetActivityStreamsSummary(summaryJc)
	sc := NewActivityStreamsCollection()
	summarySc := NewActivityStreamsSummaryProperty()
	summarySc.AppendXMLSchemaString("Sally's Connections")
	sc.SetActivityStreamsSummary(summarySc)
	o1 := NewActivityStreamsAccept()
	oId1 := NewJSONLDIdProperty()
	oId1.Set(MustParseURL("http://example.org/activities/122"))
	o1.SetJSONLDId(oId1)
	summary1 := NewActivityStreamsSummaryProperty()
	summary1.AppendXMLSchemaString("John accepted Sally's friend request")
	o1.SetActivityStreamsSummary(summary1)
	obj1 := NewActivityStreamsObjectProperty()
	obj1.AppendIRI(req123)
	o1.SetActivityStreamsObject(obj1)
	inReplyTo := NewActivityStreamsInReplyToProperty()
	inReplyTo.AppendIRI(req123)
	o1.SetActivityStreamsInReplyTo(inReplyTo)
	ctx1 := NewActivityStreamsContextProperty()
	ctx1.AppendIRI(conn123)
	o1.SetActivityStreamsContext(ctx1)
	result := NewActivityStreamsResultProperty()
	result.AppendIRI(a123)
	result.AppendIRI(a124)
	result.AppendIRI(a125)
	result.AppendIRI(a126)
	o1.SetActivityStreamsResult(result)
	objectActor1 := NewActivityStreamsActorProperty()
	objectActor1.AppendIRI(john)
	o1.SetActivityStreamsActor(objectActor1)
	o2 := NewActivityStreamsFollow()
	oId2 := NewJSONLDIdProperty()
	oId2.Set(MustParseURL("http://example.org/activities/123"))
	o2.SetJSONLDId(oId2)
	objectActor2 := NewActivityStreamsActorProperty()
	objectActor2.AppendIRI(john)
	o2.SetActivityStreamsActor(objectActor2)
	obj2 := NewActivityStreamsObjectProperty()
	obj2.AppendIRI(sally)
	o2.SetActivityStreamsObject(obj2)
	ctx2 := NewActivityStreamsContextProperty()
	ctx2.AppendIRI(conn123)
	o2.SetActivityStreamsContext(ctx2)
	summary2 := NewActivityStreamsSummaryProperty()
	summary2.AppendXMLSchemaString("John followed Sally")
	o2.SetActivityStreamsSummary(summary2)
	o3 := NewActivityStreamsFollow()
	oId3 := NewJSONLDIdProperty()
	oId3.Set(MustParseURL("http://example.org/activities/124"))
	o3.SetJSONLDId(oId3)
	objectActor3 := NewActivityStreamsActorProperty()
	objectActor3.AppendIRI(sally)
	o3.SetActivityStreamsActor(objectActor3)
	obj3 := NewActivityStreamsObjectProperty()
	obj3.AppendIRI(john)
	o3.SetActivityStreamsObject(obj3)
	ctx3 := NewActivityStreamsContextProperty()
	ctx3.AppendIRI(conn123)
	o3.SetActivityStreamsContext(ctx3)
	summary3 := NewActivityStreamsSummaryProperty()
	summary3.AppendXMLSchemaString("Sally followed John")
	o3.SetActivityStreamsSummary(summary3)
	o4 := NewActivityStreamsAdd()
	oId4 := NewJSONLDIdProperty()
	oId4.Set(MustParseURL("http://example.org/activities/125"))
	o4.SetJSONLDId(oId4)
	summary4 := NewActivityStreamsSummaryProperty()
	summary4.AppendXMLSchemaString("John added Sally to his friends list")
	o4.SetActivityStreamsSummary(summary4)
	objectActor4 := NewActivityStreamsActorProperty()
	objectActor4.AppendIRI(john)
	o4.SetActivityStreamsActor(objectActor4)
	obj4 := NewActivityStreamsObjectProperty()
	obj4.AppendIRI(conn123)
	o4.SetActivityStreamsObject(obj4)
	ctx4 := NewActivityStreamsContextProperty()
	ctx4.AppendIRI(conn123)
	o4.SetActivityStreamsContext(ctx4)
	tobj4 := NewActivityStreamsTargetProperty()
	tobj4.AppendActivityStreamsCollection(jc)
	o4.SetActivityStreamsTarget(tobj4)
	o5 := NewActivityStreamsAdd()
	oId5 := NewJSONLDIdProperty()
	oId5.Set(MustParseURL("http://example.org/activities/126"))
	o5.SetJSONLDId(oId5)
	summary5 := NewActivityStreamsSummaryProperty()
	summary5.AppendXMLSchemaString("Sally added John to her friends list")
	o5.SetActivityStreamsSummary(summary5)
	objectActor5 := NewActivityStreamsActorProperty()
	objectActor5.AppendIRI(sally)
	o5.SetActivityStreamsActor(objectActor5)
	obj5 := NewActivityStreamsObjectProperty()
	obj5.AppendIRI(conn123)
	o5.SetActivityStreamsObject(obj5)
	ctx5 := NewActivityStreamsContextProperty()
	ctx5.AppendIRI(conn123)
	o5.SetActivityStreamsContext(ctx5)
	tobj5 := NewActivityStreamsTargetProperty()
	tobj5.AppendActivityStreamsCollection(sc)
	o5.SetActivityStreamsTarget(tobj5)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally and John's relationship history")
	example148Type.SetActivityStreamsSummary(summary)
	items := NewActivityStreamsItemsProperty()
	items.AppendActivityStreamsAccept(o1)
	items.AppendActivityStreamsFollow(o2)
	items.AppendActivityStreamsFollow(o3)
	items.AppendActivityStreamsAdd(o4)
	items.AppendActivityStreamsAdd(o5)
	example148Type.SetActivityStreamsItems(items)
	return example148Type
}

const example149 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Place",
  "name": "San Francisco, CA"
}`

func example149Type() vocab.ActivityStreamsPlace {
	example149Type := NewActivityStreamsPlace()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("San Francisco, CA")
	example149Type.SetActivityStreamsName(name)
	return example149Type
}

// NOTE: Un-stringified the longitude and latitude values.
const example150 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Place",
  "name": "San Francisco, CA",
  "longitude": 122.4167,
  "latitude": 37.7833
}`

func example150Type() vocab.ActivityStreamsPlace {
	example150Type := NewActivityStreamsPlace()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("San Francisco, CA")
	example150Type.SetActivityStreamsName(name)
	lon := NewActivityStreamsLongitudeProperty()
	lon.Set(122.4167)
	example150Type.SetActivityStreamsLongitude(lon)
	lat := NewActivityStreamsLatitudeProperty()
	lat.Set(37.7833)
	example150Type.SetActivityStreamsLatitude(lat)
	return example150Type
}

const example151 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "name": "A question about robots",
  "id": "http://help.example.org/question/1",
  "type": "Question",
  "content": "I'd like to build a robot to feed my cat. Should I use Arduino or Raspberry Pi?"
}`

func example151Type() vocab.ActivityStreamsQuestion {
	example151Type := NewActivityStreamsQuestion()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("A question about robots")
	example151Type.SetActivityStreamsName(name)
	id := NewJSONLDIdProperty()
	id.Set(MustParseURL("http://help.example.org/question/1"))
	example151Type.SetJSONLDId(id)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("I'd like to build a robot to feed my cat. Should I use Arduino or Raspberry Pi?")
	example151Type.SetActivityStreamsContent(content)
	return example151Type
}

const example152 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://polls.example.org/question/1",
  "name": "A question about robots",
  "type": "Question",
  "content": "I'd like to build a robot to feed my cat. Which platform is best?",
  "oneOf": [
    {"name": "arduino"},
    {"name": "raspberry pi"}
  ]
}`

var example152Unknown = func(m map[string]interface{}) map[string]interface{} {
	ard := make(map[string]interface{})
	ard["name"] = "arduino"
	ras := make(map[string]interface{})
	ras["name"] = "raspberry pi"
	m["oneOf"] = []interface{}{ard, ras}
	return m
}

func example152Type() vocab.ActivityStreamsQuestion {
	example152Type := NewActivityStreamsQuestion()
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("A question about robots")
	example152Type.SetActivityStreamsName(name)
	id := NewJSONLDIdProperty()
	id.Set(MustParseURL("http://polls.example.org/question/1"))
	example152Type.SetJSONLDId(id)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("I'd like to build a robot to feed my cat. Which platform is best?")
	example152Type.SetActivityStreamsContent(content)
	return example152Type
}

const example153 = `{
 "@context": "https://www.w3.org/ns/activitystreams",
 "attributedTo": "http://sally.example.org",
 "inReplyTo": "http://polls.example.org/question/1",
 "name": "arduino"
}`

var example153Unknown = func(m map[string]interface{}) map[string]interface{} {
	m["@context"] = "https://www.w3.org/ns/activitystreams"
	m["attributedTo"] = "http://sally.example.org"
	m["inReplyTo"] = "http://polls.example.org/question/1"
	m["name"] = "arduino"
	return m
}

const example154 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "name": "A question about robots",
  "id": "http://polls.example.org/question/1",
  "type": "Question",
  "content": "I'd like to build a robot to feed my cat. Which platform is best?",
  "oneOf": [
    {"name": "arduino"},
    {"name": "raspberry pi"}
  ],
  "replies": {
    "type": "Collection",
    "totalItems": 3,
    "items": [
      {
        "attributedTo": "http://sally.example.org",
        "inReplyTo": "http://polls.example.org/question/1",
        "name": "arduino"
      },
      {
        "attributedTo": "http://joe.example.org",
        "inReplyTo": "http://polls.example.org/question/1",
        "name": "arduino"
      },
      {
        "attributedTo": "http://john.example.org",
        "inReplyTo": "http://polls.example.org/question/1",
        "name": "raspberry pi"
      }
    ]
  },
  "result": {
    "type": "Note",
    "content": "Users are favoriting &quot;arduino&quot; by a 33% margin."
  }
}`

var example154Unknown = func(m map[string]interface{}) map[string]interface{} {
	ard := make(map[string]interface{})
	ard["name"] = "arduino"
	ras := make(map[string]interface{})
	ras["name"] = "raspberry pi"
	m["oneOf"] = []interface{}{ard, ras}
	// replies
	one := make(map[string]interface{})
	one["attributedTo"] = "http://sally.example.org"
	one["inReplyTo"] = "http://polls.example.org/question/1"
	one["name"] = "arduino"
	two := make(map[string]interface{})
	two["attributedTo"] = "http://joe.example.org"
	two["inReplyTo"] = "http://polls.example.org/question/1"
	two["name"] = "arduino"
	three := make(map[string]interface{})
	three["attributedTo"] = "http://john.example.org"
	three["inReplyTo"] = "http://polls.example.org/question/1"
	three["name"] = "raspberry pi"
	items := []interface{}{one, two, three}
	replies := m["replies"].(map[string]interface{})
	replies["items"] = items
	return m
}

func example154Type() vocab.ActivityStreamsQuestion {
	example154Type := NewActivityStreamsQuestion()
	replies := NewActivityStreamsCollection()
	totalItems := NewActivityStreamsTotalItemsProperty()
	totalItems.Set(3)
	replies.SetActivityStreamsTotalItems(totalItems)
	note := NewActivityStreamsNote()
	contentNote := NewActivityStreamsContentProperty()
	contentNote.AppendXMLSchemaString("Users are favoriting &quot;arduino&quot; by a 33% margin.")
	note.SetActivityStreamsContent(contentNote)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("A question about robots")
	example154Type.SetActivityStreamsName(name)
	id := NewJSONLDIdProperty()
	id.Set(MustParseURL("http://polls.example.org/question/1"))
	example154Type.SetJSONLDId(id)
	content := NewActivityStreamsContentProperty()
	content.AppendXMLSchemaString("I'd like to build a robot to feed my cat. Which platform is best?")
	example154Type.SetActivityStreamsContent(content)
	reply := NewActivityStreamsRepliesProperty()
	reply.SetActivityStreamsCollection(replies)
	example154Type.SetActivityStreamsReplies(reply)
	result := NewActivityStreamsResultProperty()
	result.AppendActivityStreamsNote(note)
	example154Type.SetActivityStreamsResult(result)
	return example154Type
}

const example155 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "History of John's note",
  "type": "Collection",
  "items": [
    {
      "summary": "Sally liked John's note",
      "type": "Like",
      "actor": "http://sally.example.org",
      "id": "http://activities.example.com/1",
      "published": "2015-11-12T12:34:56Z",
      "object": {
        "summary": "John's note",
        "type": "Note",
        "id": "http://notes.example.com/1",
        "attributedTo": "http://john.example.org",
        "content": "My note"
      }
    },
    {
      "summary": "Sally disliked John's note",
      "type": "Dislike",
      "actor": "http://sally.example.org",
      "id": "http://activities.example.com/2",
      "published": "2015-12-11T21:43:56Z",
      "object": {
        "summary": "John's note",
        "type": "Note",
        "id": "http://notes.example.com/1",
        "attributedTo": "http://john.example.org",
        "content": "My note"
      }
    }
  ]
}`

func example155Type() vocab.ActivityStreamsCollection {
	example155Type := NewActivityStreamsCollection()
	john := MustParseURL("http://john.example.org")
	sally := MustParseURL("http://sally.example.org")
	t1, err := time.Parse(time.RFC3339, "2015-11-12T12:34:56Z")
	if err != nil {
		panic(err)
	}
	t2, err := time.Parse(time.RFC3339, "2015-12-11T21:43:56Z")
	if err != nil {
		panic(err)
	}
	note := NewActivityStreamsNote()
	summaryNote := NewActivityStreamsSummaryProperty()
	summaryNote.AppendXMLSchemaString("John's note")
	note.SetActivityStreamsSummary(summaryNote)
	noteId := NewJSONLDIdProperty()
	noteId.Set(MustParseURL("http://notes.example.com/1"))
	note.SetJSONLDId(noteId)
	contentNote := NewActivityStreamsContentProperty()
	contentNote.AppendXMLSchemaString("My note")
	note.SetActivityStreamsContent(contentNote)
	attr := NewActivityStreamsAttributedToProperty()
	attr.AppendIRI(john)
	note.SetActivityStreamsAttributedTo(attr)
	like := NewActivityStreamsLike()
	summaryLike := NewActivityStreamsSummaryProperty()
	summaryLike.AppendXMLSchemaString("Sally liked John's note")
	like.SetActivityStreamsSummary(summaryLike)
	likeId := NewJSONLDIdProperty()
	likeId.Set(MustParseURL("http://activities.example.com/1"))
	like.SetJSONLDId(likeId)
	likeActor := NewActivityStreamsActorProperty()
	likeActor.AppendIRI(sally)
	like.SetActivityStreamsActor(likeActor)
	published1 := NewActivityStreamsPublishedProperty()
	published1.Set(t1)
	like.SetActivityStreamsPublished(published1)
	objLike := NewActivityStreamsObjectProperty()
	objLike.AppendActivityStreamsNote(note)
	like.SetActivityStreamsObject(objLike)
	dislike := NewActivityStreamsDislike()
	summaryDislike := NewActivityStreamsSummaryProperty()
	summaryDislike.AppendXMLSchemaString("Sally disliked John's note")
	dislike.SetActivityStreamsSummary(summaryDislike)
	dislikeId := NewJSONLDIdProperty()
	dislikeId.Set(MustParseURL("http://activities.example.com/2"))
	dislike.SetJSONLDId(dislikeId)
	dislikeActor := NewActivityStreamsActorProperty()
	dislikeActor.AppendIRI(sally)
	dislike.SetActivityStreamsActor(dislikeActor)
	published2 := NewActivityStreamsPublishedProperty()
	published2.Set(t2)
	dislike.SetActivityStreamsPublished(published2)
	objDislike := NewActivityStreamsObjectProperty()
	objDislike.AppendActivityStreamsNote(note)
	dislike.SetActivityStreamsObject(objDislike)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("History of John's note")
	example155Type.SetActivityStreamsSummary(summary)
	items := NewActivityStreamsItemsProperty()
	items.AppendActivityStreamsLike(like)
	items.AppendActivityStreamsDislike(dislike)
	example155Type.SetActivityStreamsItems(items)
	return example155Type
}

const example156 = `{
 "@context": "https://www.w3.org/ns/activitystreams",
 "summary": "History of John's note",
 "type": "Collection",
 "items": [
   {
     "summary": "Sally liked John's note",
     "type": "Like",
     "id": "http://activities.example.com/1",
     "actor": "http://sally.example.org",
     "published": "2015-11-12T12:34:56Z",
     "object": {
       "summary": "John's note",
       "type": "Note",
       "id": "http://notes.example.com/1",
       "attributedTo": "http://john.example.org",
       "content": "My note"
     }
   },
   {
     "summary": "Sally no longer likes John's note",
     "type": "Undo",
     "id": "http://activities.example.com/2",
     "actor": "http://sally.example.org",
     "published": "2015-12-11T21:43:56Z",
     "object": "http://activities.example.com/1"
   }
 ]
}`

func example156Type() vocab.ActivityStreamsCollection {
	example156Type := NewActivityStreamsCollection()
	john := MustParseURL("http://john.example.org")
	sally := MustParseURL("http://sally.example.org")
	a := MustParseURL("http://activities.example.com/1")
	t1, err := time.Parse(time.RFC3339, "2015-11-12T12:34:56Z")
	if err != nil {
		panic(err)
	}
	t2, err := time.Parse(time.RFC3339, "2015-12-11T21:43:56Z")
	if err != nil {
		panic(err)
	}
	note := NewActivityStreamsNote()
	noteId := NewJSONLDIdProperty()
	noteId.Set(MustParseURL("http://notes.example.com/1"))
	note.SetJSONLDId(noteId)
	summaryNote := NewActivityStreamsSummaryProperty()
	summaryNote.AppendXMLSchemaString("John's note")
	note.SetActivityStreamsSummary(summaryNote)
	attr := NewActivityStreamsAttributedToProperty()
	attr.AppendIRI(john)
	note.SetActivityStreamsAttributedTo(attr)
	contentNote := NewActivityStreamsContentProperty()
	contentNote.AppendXMLSchemaString("My note")
	note.SetActivityStreamsContent(contentNote)
	like := NewActivityStreamsLike()
	likeId := NewJSONLDIdProperty()
	likeId.Set(MustParseURL("http://activities.example.com/1"))
	like.SetJSONLDId(likeId)
	summaryLike := NewActivityStreamsSummaryProperty()
	summaryLike.AppendXMLSchemaString("Sally liked John's note")
	like.SetActivityStreamsSummary(summaryLike)
	likeActor := NewActivityStreamsActorProperty()
	likeActor.AppendIRI(sally)
	like.SetActivityStreamsActor(likeActor)
	published1 := NewActivityStreamsPublishedProperty()
	published1.Set(t1)
	like.SetActivityStreamsPublished(published1)
	objLike := NewActivityStreamsObjectProperty()
	objLike.AppendActivityStreamsNote(note)
	like.SetActivityStreamsObject(objLike)
	undo := NewActivityStreamsUndo()
	undoId := NewJSONLDIdProperty()
	undoId.Set(MustParseURL("http://activities.example.com/2"))
	undo.SetJSONLDId(undoId)
	summaryUndo := NewActivityStreamsSummaryProperty()
	summaryUndo.AppendXMLSchemaString("Sally no longer likes John's note")
	undo.SetActivityStreamsSummary(summaryUndo)
	undoActor := NewActivityStreamsActorProperty()
	undoActor.AppendIRI(sally)
	undo.SetActivityStreamsActor(undoActor)
	published2 := NewActivityStreamsPublishedProperty()
	published2.Set(t2)
	undo.SetActivityStreamsPublished(published2)
	obj := NewActivityStreamsObjectProperty()
	obj.AppendIRI(a)
	undo.SetActivityStreamsObject(obj)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("History of John's note")
	example156Type.SetActivityStreamsSummary(summary)
	items := NewActivityStreamsItemsProperty()
	items.AppendActivityStreamsLike(like)
	items.AppendActivityStreamsUndo(undo)
	example156Type.SetActivityStreamsItems(items)
	return example156Type
}

// NOTE: The `content` field has been inlined to keep within JSON spec.
const example157 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "name": "A thank-you note",
  "type": "Note",
  "content": "Thank you <a href='http://sally.example.org'>@sally</a> for all your hard work! <a href='http://example.org/tags/givingthanks'>#givingthanks</a>",
  "to": {
    "name": "Sally",
    "type": "Person",
    "id": "http://sally.example.org"
  },
  "tag": {
    "id": "http://example.org/tags/givingthanks",
    "name": "#givingthanks"
  }
}`

var example157Unknown = func(m map[string]interface{}) map[string]interface{} {
	m["tag"] = map[string]interface{}{
		"id":   "http://example.org/tags/givingthanks",
		"name": "#givingthanks",
	}
	return m
}

func example157Type() vocab.ActivityStreamsNote {
	example157Type := NewActivityStreamsNote()
	person := NewActivityStreamsPerson()
	sally := NewActivityStreamsNameProperty()
	sally.AppendXMLSchemaString("Sally")
	person.SetActivityStreamsName(sally)
	id := NewJSONLDIdProperty()
	id.Set(MustParseURL("http://sally.example.org"))
	person.SetJSONLDId(id)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("A thank-you note")
	example157Type.SetActivityStreamsName(name)
	contentNote := NewActivityStreamsContentProperty()
	contentNote.AppendXMLSchemaString("Thank you <a href='http://sally.example.org'>@sally</a> for all your hard work! <a href='http://example.org/tags/givingthanks'>#givingthanks</a>")
	example157Type.SetActivityStreamsContent(contentNote)
	to := NewActivityStreamsToProperty()
	to.AppendActivityStreamsPerson(person)
	example157Type.SetActivityStreamsTo(to)
	return example157Type
}

const example158 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "name": "A thank-you note",
  "type": "Note",
  "content": "Thank you @sally for all your hard work! #givingthanks",
  "tag": [
    {
      "type": "Mention",
      "href": "http://example.org/people/sally",
      "name": "@sally"
    },
    {
      "id": "http://example.org/tags/givingthanks",
      "name": "#givingthanks"
    }
  ]
}`

var example158Unknown = func(m map[string]interface{}) map[string]interface{} {
	existing := m["tag"]
	next := map[string]interface{}{
		"id":   "http://example.org/tags/givingthanks",
		"name": "#givingthanks",
	}
	m["tag"] = []interface{}{existing, next}
	return m
}

func example158Type() vocab.ActivityStreamsNote {
	example158Type := NewActivityStreamsNote()
	u := MustParseURL("http://example.org/people/sally")
	mention := NewActivityStreamsMention()
	hrefLink := NewActivityStreamsHrefProperty()
	hrefLink.Set(u)
	mention.SetActivityStreamsHref(hrefLink)
	mentionName := NewActivityStreamsNameProperty()
	mentionName.AppendXMLSchemaString("@sally")
	mention.SetActivityStreamsName(mentionName)
	name := NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString("A thank-you note")
	example158Type.SetActivityStreamsName(name)
	contentNote := NewActivityStreamsContentProperty()
	contentNote.AppendXMLSchemaString("Thank you @sally for all your hard work! #givingthanks")
	example158Type.SetActivityStreamsContent(contentNote)
	tag := NewActivityStreamsTagProperty()
	tag.AppendActivityStreamsMention(mention)
	example158Type.SetActivityStreamsTag(tag)
	return example158Type
}

const example159 = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally moved the sales figures from Folder A to Folder B",
  "type": "Move",
  "actor": "http://sally.example.org",
  "object": {
    "type": "Document",
    "name": "sales figures"
  },
  "origin": {
    "type": "Collection",
    "name": "Folder A"
  },
  "target": {
    "type": "Collection",
    "name": "Folder B"
  }
}`

func example159Type() vocab.ActivityStreamsMove {
	example159Type := NewActivityStreamsMove()
	sally := MustParseURL("http://sally.example.org")
	obj := NewActivityStreamsDocument()
	nameObj := NewActivityStreamsNameProperty()
	nameObj.AppendXMLSchemaString("sales figures")
	obj.SetActivityStreamsName(nameObj)
	origin := NewActivityStreamsCollection()
	nameOrigin := NewActivityStreamsNameProperty()
	nameOrigin.AppendXMLSchemaString("Folder A")
	origin.SetActivityStreamsName(nameOrigin)
	target := NewActivityStreamsCollection()
	nameTarget := NewActivityStreamsNameProperty()
	nameTarget.AppendXMLSchemaString("Folder B")
	target.SetActivityStreamsName(nameTarget)
	summary := NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString("Sally moved the sales figures from Folder A to Folder B")
	example159Type.SetActivityStreamsSummary(summary)
	objectActor := NewActivityStreamsActorProperty()
	objectActor.AppendIRI(sally)
	example159Type.SetActivityStreamsActor(objectActor)
	object := NewActivityStreamsObjectProperty()
	object.AppendActivityStreamsDocument(obj)
	example159Type.SetActivityStreamsObject(object)
	originProp := NewActivityStreamsOriginProperty()
	originProp.AppendActivityStreamsCollection(origin)
	example159Type.SetActivityStreamsOrigin(originProp)
	tobj := NewActivityStreamsTargetProperty()
	tobj.AppendActivityStreamsCollection(target)
	example159Type.SetActivityStreamsTarget(tobj)
	return example159Type
}

const personExampleWithPublicKey = `{
  "@context": [
    "https://w3id.org/security/v1",
    "https://www.w3.org/ns/activitystreams"
  ],
  "id":"https://mastodon.technology/users/cj",
  "type":"Person",
  "following":"https://mastodon.technology/users/cj/following",
  "followers":"https://mastodon.technology/users/cj/followers",
  "inbox":"https://mastodon.technology/users/cj/inbox",
  "outbox":"https://mastodon.technology/users/cj/outbox",
  "preferredUsername":"cj",
  "name":"cj",
  "url":"https://mastodon.technology/@cj",
  "publicKey": {
    "id":"https://mastodon.technology/users/cj#main-key",
    "owner":"https://mastodon.technology/users/cj",
    "publicKeyPem":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArYzMmldblHfnAPbwfVIo\nFpV6ej3JUS9boZHJbYh9c3IpumDoDXJyThTx19wM8M04fJljJ74aTp+fZdIng6l3\nswT24dvgahMUoD4/NLrPOjulhIOGfYzGYfTduh6wT+aaxV5w+OPO5plOVgrgS+RK\n9mv5SOQIaSLQGNCc5RSKuea8H4fG/bUiPpCXlqq2iVp0hoc3rI3K2NOErxOeex1B\nzcBLupBiXBB5a4hxwTVMfjmxqEZSuC1xnx2c8R9FpZPmhGovqlVzK/JlTPBU53f/\nDT773fOqr2jLTLfS0VNrI0jYpz0GG687O/FDRi2YR91D9NSs9WFHeuVC/A5mhuMH\npQIDAQAB\n-----END PUBLIC KEY-----\n"
  }
}`

func personExampleWithPublicKeyType() vocab.ActivityStreamsPerson {
	personExampleType := NewActivityStreamsPerson()
	idIRI := MustParseURL("https://mastodon.technology/users/cj")
	idProp := NewJSONLDIdProperty()
	idProp.Set(idIRI)
	personExampleType.SetJSONLDId(idProp)
	followingIRI := MustParseURL("https://mastodon.technology/users/cj/following")
	followingProp := NewActivityStreamsFollowingProperty()
	followingProp.SetIRI(followingIRI)
	personExampleType.SetActivityStreamsFollowing(followingProp)
	followersIRI := MustParseURL("https://mastodon.technology/users/cj/followers")
	followersProp := NewActivityStreamsFollowersProperty()
	followersProp.SetIRI(followersIRI)
	personExampleType.SetActivityStreamsFollowers(followersProp)
	inboxIRI := MustParseURL("https://mastodon.technology/users/cj/inbox")
	inboxProp := NewActivityStreamsInboxProperty()
	inboxProp.SetIRI(inboxIRI)
	personExampleType.SetActivityStreamsInbox(inboxProp)
	outboxIRI := MustParseURL("https://mastodon.technology/users/cj/outbox")
	outboxProp := NewActivityStreamsOutboxProperty()
	outboxProp.SetIRI(outboxIRI)
	personExampleType.SetActivityStreamsOutbox(outboxProp)
	preferredUsernameProp := NewActivityStreamsPreferredUsernameProperty()
	preferredUsernameProp.SetXMLSchemaString("cj")
	personExampleType.SetActivityStreamsPreferredUsername(preferredUsernameProp)
	nameProp := NewActivityStreamsNameProperty()
	nameProp.AppendXMLSchemaString("cj")
	personExampleType.SetActivityStreamsName(nameProp)
	urlIRI := MustParseURL("https://mastodon.technology/@cj")
	urlProp := NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(urlIRI)
	personExampleType.SetActivityStreamsUrl(urlProp)
	publicKeyProp := NewW3IDSecurityV1PublicKeyProperty()
	publicKeyType := NewW3IDSecurityV1PublicKey()
	pubKeyIdIRI := MustParseURL("https://mastodon.technology/users/cj#main-key")
	pubKeyIdProp := NewJSONLDIdProperty()
	pubKeyIdProp.Set(pubKeyIdIRI)
	publicKeyType.SetJSONLDId(pubKeyIdProp)
	ownerIRI := MustParseURL("https://mastodon.technology/users/cj")
	ownerProp := NewW3IDSecurityV1OwnerProperty()
	ownerProp.SetIRI(ownerIRI)
	publicKeyType.SetW3IDSecurityV1Owner(ownerProp)
	publicKeyPemProp := NewW3IDSecurityV1PublicKeyPemProperty()
	publicKeyPemProp.Set("-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArYzMmldblHfnAPbwfVIo\nFpV6ej3JUS9boZHJbYh9c3IpumDoDXJyThTx19wM8M04fJljJ74aTp+fZdIng6l3\nswT24dvgahMUoD4/NLrPOjulhIOGfYzGYfTduh6wT+aaxV5w+OPO5plOVgrgS+RK\n9mv5SOQIaSLQGNCc5RSKuea8H4fG/bUiPpCXlqq2iVp0hoc3rI3K2NOErxOeex1B\nzcBLupBiXBB5a4hxwTVMfjmxqEZSuC1xnx2c8R9FpZPmhGovqlVzK/JlTPBU53f/\nDT773fOqr2jLTLfS0VNrI0jYpz0GG687O/FDRi2YR91D9NSs9WFHeuVC/A5mhuMH\npQIDAQAB\n-----END PUBLIC KEY-----\n")
	publicKeyType.SetW3IDSecurityV1PublicKeyPem(publicKeyPemProp)
	publicKeyProp.AppendW3IDSecurityV1PublicKey(publicKeyType)
	personExampleType.SetW3IDSecurityV1PublicKey(publicKeyProp)
	return personExampleType
}

type testContextWrapper struct {
	vocab.ActivityStreamsObject
}

func (a *testContextWrapper) JSONLDContext() map[string]string {
	m := a.ActivityStreamsObject.JSONLDContext()
	m["https://schema.org#"] = "schema"
	m["schema:PropertyValue"] = "PropertyValue"
	m["schema:value"] = "value"
	return m
}

const serviceHasAttachmentWithUnknown = `{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    {
      "schema": "https://schema.org#",
      "PropertyValue": "schema:PropertyValue",
      "value": "schema:value"
    }
  ],
  "id": "https://example.com/service",
  "type": "Service",
  "attachment": [
    {
      "type": "PropertyValue",
      "name": "First Object",
      "value": "test value on first object"
    },
    {
      "type": "PropertyValue",
      "name": "Second Object",
      "value": "test value on second object"
    }
  ]
}`

func serviceHasAttachmentWithUnknownType() vocab.ActivityStreamsService {
	serviceType := NewActivityStreamsService()
	idProp := NewJSONLDIdProperty()
	idIRI := MustParseURL("https://example.com/service")
	idProp.Set(idIRI)
	serviceType.SetJSONLDId(idProp)
	attachmentProp := NewActivityStreamsAttachmentProperty()
	firstObject := NewActivityStreamsObject()
	firstObjectTypeProp := NewJSONLDTypeProperty()
	firstObjectTypeProp.AppendXMLSchemaString("PropertyValue")
	firstObject.SetJSONLDType(firstObjectTypeProp)
	firstObjectNameProp := NewActivityStreamsNameProperty()
	firstObjectNameProp.AppendXMLSchemaString("First Object")
	firstObject.SetActivityStreamsName(firstObjectNameProp)
	firstObject.GetUnknownProperties()["value"] = "test value on first object"
	attachmentProp.AppendType(&testContextWrapper{firstObject})
	secondObject := NewActivityStreamsObject()
	secondObjectTypeProp := NewJSONLDTypeProperty()
	secondObjectTypeProp.AppendXMLSchemaString("PropertyValue")
	secondObject.SetJSONLDType(secondObjectTypeProp)
	secondObjectNameProp := NewActivityStreamsNameProperty()
	secondObjectNameProp.AppendXMLSchemaString("Second Object")
	secondObject.SetActivityStreamsName(secondObjectNameProp)
	secondObject.GetUnknownProperties()["value"] = "test value on second object"
	attachmentProp.AppendType(secondObject)
	serviceType.SetActivityStreamsAttachment(attachmentProp)
	return serviceType
}
