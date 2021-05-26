package federatingdb

import (
	"context"
	"net/url"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
)

// Liked obtains the Liked Collection for an actor with the
// given id.
//
// If modified, the library will then call Update.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Liked(c context.Context, actorIRI *url.URL) (liked vocab.ActivityStreamsCollection, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func":     "Liked",
			"actorIRI": actorIRI.String(),
		},
	)
	l.Debugf("entering LIKED function with actorIRI %s", actorIRI.String())
	return nil, nil
}
