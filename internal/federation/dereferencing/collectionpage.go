package dereferencing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// DereferenceCollectionPage returns the activitystreams CollectionPage at the specified IRI, or an error if something goes wrong.
func (d *deref) DereferenceCollectionPage(username string, pageIRI *url.URL) (ap.CollectionPageable, error) {
	if blocked, err := d.blockedDomain(pageIRI.Host); blocked || err != nil {
		return nil, fmt.Errorf("DereferenceCollectionPage: domain %s is blocked", pageIRI.Host)
	}

	transport, err := d.transportController.NewTransportForUsername(username)
	if err != nil {
		return nil, fmt.Errorf("DereferenceCollectionPage: error creating transport: %s", err)
	}

	b, err := transport.Dereference(context.Background(), pageIRI)
	if err != nil {
		return nil, fmt.Errorf("DereferenceCollectionPage: error deferencing %s: %s", pageIRI.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("DereferenceCollectionPage: error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		return nil, fmt.Errorf("DereferenceCollectionPage: error resolving json into ap vocab type: %s", err)
	}

	if t.GetTypeName() != gtsmodel.ActivityStreamsCollectionPage {
		return nil, fmt.Errorf("DereferenceCollectionPage: type name %s not supported", t.GetTypeName())
	}

	p, ok := t.(vocab.ActivityStreamsCollectionPage)
	if !ok {
		return nil, errors.New("DereferenceCollectionPage: error resolving type as activitystreams collection page")
	}

	return p, nil
}
