package federation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

func (f *federator) DereferenceRemoteAccount(username string, remoteAccountID *url.URL) (typeutils.Accountable, error) {
	f.startHandshake(username, remoteAccountID)
	defer f.stopHandshake(username, remoteAccountID)

	transport, err := f.GetTransportForUser(username)
	if err != nil {
		return nil, fmt.Errorf("transport err: %s", err)
	}

	b, err := transport.Dereference(context.Background(), remoteAccountID)
	if err != nil {
		return nil, fmt.Errorf("error deferencing %s: %s", remoteAccountID.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		return nil, fmt.Errorf("error resolving json into ap vocab type: %s", err)
	}

	switch t.GetTypeName() {
	case string(gtsmodel.ActivityStreamsPerson):
		p, ok := t.(vocab.ActivityStreamsPerson)
		if !ok {
			return nil, errors.New("error resolving type as activitystreams person")
		}
		return p, nil
	case string(gtsmodel.ActivityStreamsApplication):
		p, ok := t.(vocab.ActivityStreamsApplication)
		if !ok {
			return nil, errors.New("error resolving type as activitystreams application")
		}
		return p, nil
	case string(gtsmodel.ActivityStreamsService):
		p, ok := t.(vocab.ActivityStreamsService)
		if !ok {
			return nil, errors.New("error resolving type as activitystreams service")
		}
		return p, nil
	}

	return nil, fmt.Errorf("type name %s not supported", t.GetTypeName())
}

func (f *federator) DereferenceRemoteStatus(username string, remoteStatusID *url.URL) (typeutils.Statusable, error) {
	transport, err := f.GetTransportForUser(username)
	if err != nil {
		return nil, fmt.Errorf("transport err: %s", err)
	}

	b, err := transport.Dereference(context.Background(), remoteStatusID)
	if err != nil {
		return nil, fmt.Errorf("error deferencing %s: %s", remoteStatusID.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		return nil, fmt.Errorf("error resolving json into ap vocab type: %s", err)
	}

	// Article, Document, Image, Video, Note, Page, Event, Place, Mention, Profile
	switch t.GetTypeName() {
	case gtsmodel.ActivityStreamsArticle:
		p, ok := t.(vocab.ActivityStreamsArticle)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsArticle")
		}
		return p, nil
	case gtsmodel.ActivityStreamsDocument:
		p, ok := t.(vocab.ActivityStreamsDocument)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsDocument")
		}
		return p, nil
	case gtsmodel.ActivityStreamsImage:
		p, ok := t.(vocab.ActivityStreamsImage)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsImage")
		}
		return p, nil
	case gtsmodel.ActivityStreamsVideo:
		p, ok := t.(vocab.ActivityStreamsVideo)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsVideo")
		}
		return p, nil
	case gtsmodel.ActivityStreamsNote:
		p, ok := t.(vocab.ActivityStreamsNote)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsNote")
		}
		return p, nil
	case gtsmodel.ActivityStreamsPage:
		p, ok := t.(vocab.ActivityStreamsPage)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsPage")
		}
		return p, nil
	case gtsmodel.ActivityStreamsEvent:
		p, ok := t.(vocab.ActivityStreamsEvent)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsEvent")
		}
		return p, nil
	case gtsmodel.ActivityStreamsPlace:
		p, ok := t.(vocab.ActivityStreamsPlace)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsPlace")
		}
		return p, nil
	case gtsmodel.ActivityStreamsProfile:
		p, ok := t.(vocab.ActivityStreamsProfile)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsProfile")
		}
		return p, nil
	}

	return nil, fmt.Errorf("type name %s not supported", t.GetTypeName())
}

func (f *federator) DereferenceRemoteInstance(username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error) {
	transport, err := f.GetTransportForUser(username)
	if err != nil {
		return nil, fmt.Errorf("transport err: %s", err)
	}

	return transport.DereferenceInstance(context.Background(), remoteInstanceURI)
}
