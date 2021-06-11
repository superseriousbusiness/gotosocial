package typeutils

import (
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (c *converter) WrapPersonInUpdate(person vocab.ActivityStreamsPerson, originAccount *gtsmodel.Account) (vocab.ActivityStreamsUpdate, error) {

	update := streams.NewActivityStreamsUpdate()

	// set the actor
	actorURI, err := url.Parse(originAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("WrapPersonInUpdate: error parsing url %s: %s", originAccount.URI, err)
	}
	actorProp := streams.NewActivityStreamsActorProperty()
	actorProp.AppendIRI(actorURI)
	update.SetActivityStreamsActor(actorProp)

	// set the ID

	newID, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	idString := util.GenerateURIForUpdate(originAccount.Username, c.config.Protocol, c.config.Host, newID)
	idURI, err := url.Parse(idString)
	if err != nil {
		return nil, fmt.Errorf("WrapPersonInUpdate: error parsing url %s: %s", idString, err)
	}
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(idURI)
	update.SetJSONLDId(idProp)

	// set the person as the object here
	objectProp := streams.NewActivityStreamsObjectProperty()
	objectProp.AppendActivityStreamsPerson(person)
	update.SetActivityStreamsObject(objectProp)

	// to should be public
	toURI, err := url.Parse(asPublicURI)
	if err != nil {
		return nil, fmt.Errorf("WrapPersonInUpdate: error parsing url %s: %s", asPublicURI, err)
	}
	toProp := streams.NewActivityStreamsToProperty()
	toProp.AppendIRI(toURI)
	update.SetActivityStreamsTo(toProp)

	// bcc followers
	followersURI, err := url.Parse(originAccount.FollowersURI)
	if err != nil {
		return nil, fmt.Errorf("WrapPersonInUpdate: error parsing url %s: %s", originAccount.FollowersURI, err)
	}
	bccProp := streams.NewActivityStreamsBccProperty()
	bccProp.AppendIRI(followersURI)
	update.SetActivityStreamsBcc(bccProp)

	return update, nil
}
