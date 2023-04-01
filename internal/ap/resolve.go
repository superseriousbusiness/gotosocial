package ap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func ResolveStatusable(ctx context.Context, b []byte) (Statusable, error) {
	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("ResolveStatusable: error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("ResolveStatusable: error resolving json into ap vocab type: %s", err)
	}

	// Article, Document, Image, Video, Note, Page, Event, Place, Profile
	switch t.GetTypeName() {
	case ObjectArticle:
		p, ok := t.(vocab.ActivityStreamsArticle)
		if !ok {
			return nil, errors.New("ResolveStatusable: error resolving type as ActivityStreamsArticle")
		}
		return p, nil
	case ObjectDocument:
		p, ok := t.(vocab.ActivityStreamsDocument)
		if !ok {
			return nil, errors.New("ResolveStatusable: error resolving type as ActivityStreamsDocument")
		}
		return p, nil
	case ObjectImage:
		p, ok := t.(vocab.ActivityStreamsImage)
		if !ok {
			return nil, errors.New("ResolveStatusable: error resolving type as ActivityStreamsImage")
		}
		return p, nil
	case ObjectVideo:
		p, ok := t.(vocab.ActivityStreamsVideo)
		if !ok {
			return nil, errors.New("ResolveStatusable: error resolving type as ActivityStreamsVideo")
		}
		return p, nil
	case ObjectNote:
		p, ok := t.(vocab.ActivityStreamsNote)
		if !ok {
			return nil, errors.New("ResolveStatusable: error resolving type as ActivityStreamsNote")
		}
		return p, nil
	case ObjectPage:
		p, ok := t.(vocab.ActivityStreamsPage)
		if !ok {
			return nil, errors.New("ResolveStatusable: error resolving type as ActivityStreamsPage")
		}
		return p, nil
	case ObjectEvent:
		p, ok := t.(vocab.ActivityStreamsEvent)
		if !ok {
			return nil, errors.New("ResolveStatusable: error resolving type as ActivityStreamsEvent")
		}
		return p, nil
	case ObjectPlace:
		p, ok := t.(vocab.ActivityStreamsPlace)
		if !ok {
			return nil, errors.New("ResolveStatusable: error resolving type as ActivityStreamsPlace")
		}
		return p, nil
	case ObjectProfile:
		p, ok := t.(vocab.ActivityStreamsProfile)
		if !ok {
			return nil, errors.New("ResolveStatusable: error resolving type as ActivityStreamsProfile")
		}
		return p, nil
	default:
		return nil, fmt.Errorf("ResolveStatusable: could not resolve %T to Statusable", t)
	}
}
