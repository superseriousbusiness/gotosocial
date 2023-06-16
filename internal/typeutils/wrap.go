package typeutils

import (
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (c *converter) WrapPersonInUpdate(person vocab.ActivityStreamsPerson, originAccount *gtsmodel.Account) (vocab.ActivityStreamsUpdate, error) {
	update := streams.NewActivityStreamsUpdate()

	// set the actor
	actorURI, err := url.Parse(originAccount.URI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", originAccount.URI, err)
	}
	actorProp := streams.NewActivityStreamsActorProperty()
	actorProp.AppendIRI(actorURI)
	update.SetActivityStreamsActor(actorProp)

	// set the ID

	newID, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	idString := uris.GenerateURIForUpdate(originAccount.Username, newID)
	idURI, err := url.Parse(idString)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", idString, err)
	}
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(idURI)
	update.SetJSONLDId(idProp)

	// set the person as the object here
	objectProp := streams.NewActivityStreamsObjectProperty()
	objectProp.AppendActivityStreamsPerson(person)
	update.SetActivityStreamsObject(objectProp)

	// to should be public
	toURI, err := url.Parse(pub.PublicActivityPubIRI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", pub.PublicActivityPubIRI, err)
	}
	toProp := streams.NewActivityStreamsToProperty()
	toProp.AppendIRI(toURI)
	update.SetActivityStreamsTo(toProp)

	// bcc followers
	followersURI, err := url.Parse(originAccount.FollowersURI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", originAccount.FollowersURI, err)
	}
	bccProp := streams.NewActivityStreamsBccProperty()
	bccProp.AppendIRI(followersURI)
	update.SetActivityStreamsBcc(bccProp)

	return update, nil
}

func (c *converter) WrapNoteInCreate(note vocab.ActivityStreamsNote, objectIRIOnly bool) (vocab.ActivityStreamsCreate, error) {
	create := streams.NewActivityStreamsCreate()

	// Object property
	objectProp := streams.NewActivityStreamsObjectProperty()
	if objectIRIOnly {
		objectProp.AppendIRI(note.GetJSONLDId().GetIRI())
	} else {
		objectProp.AppendActivityStreamsNote(note)
	}
	create.SetActivityStreamsObject(objectProp)

	// ID property
	idProp := streams.NewJSONLDIdProperty()
	createID := note.GetJSONLDId().GetIRI().String() + "/activity"
	createIDIRI, err := url.Parse(createID)
	if err != nil {
		return nil, err
	}
	idProp.SetIRI(createIDIRI)
	create.SetJSONLDId(idProp)

	// Actor Property
	actorProp := streams.NewActivityStreamsActorProperty()
	actorIRI, err := ap.ExtractAttributedToURI(note)
	if err != nil {
		return nil, gtserror.Newf("couldn't extract AttributedTo: %w", err)
	}
	actorProp.AppendIRI(actorIRI)
	create.SetActivityStreamsActor(actorProp)

	// Published Property
	publishedProp := streams.NewActivityStreamsPublishedProperty()
	published, err := ap.ExtractPublished(note)
	if err != nil {
		return nil, gtserror.Newf("couldn't extract Published: %w", err)
	}
	publishedProp.Set(published)
	create.SetActivityStreamsPublished(publishedProp)

	// To Property
	toProp := streams.NewActivityStreamsToProperty()
	if toURIs := ap.ExtractToURIs(note); len(toURIs) != 0 {
		for _, toURI := range toURIs {
			toProp.AppendIRI(toURI)
		}
		create.SetActivityStreamsTo(toProp)
	}

	// Cc Property
	ccProp := streams.NewActivityStreamsCcProperty()
	if ccURIs := ap.ExtractCcURIs(note); len(ccURIs) != 0 {
		for _, ccURI := range ccURIs {
			ccProp.AppendIRI(ccURI)
		}
		create.SetActivityStreamsCc(ccProp)
	}

	return create, nil
}
