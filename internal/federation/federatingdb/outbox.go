package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// GetOutbox returns the first ordered collection page of the outbox
// at the specified IRI, for prepending new items.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) GetOutbox(c context.Context, outboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "GetOutbox",
		},
	)
	l.Debug("entering GETOUTBOX function")

	return streams.NewActivityStreamsOrderedCollectionPage(), nil
}

// SetOutbox saves the outbox value given from GetOutbox, with new items
// prepended. Note that the new items must not be added as independent
// database entries. Separate calls to Create will do that.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) SetOutbox(c context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "SetOutbox",
		},
	)
	l.Debug("entering SETOUTBOX function")

	return nil
}

// OutboxForInbox fetches the corresponding actor's outbox IRI for the
// actor's inbox IRI.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) OutboxForInbox(c context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func":     "OutboxForInbox",
			"inboxIRI": inboxIRI.String(),
		},
	)
	l.Debugf("entering OUTBOXFORINBOX function with inboxIRI %s", inboxIRI.String())

	if !util.IsInboxPath(inboxIRI) {
		return nil, fmt.Errorf("%s is not an inbox URI", inboxIRI.String())
	}
	acct := &gtsmodel.Account{}
	if err := f.db.GetWhere([]db.Where{{Key: "inbox_uri", Value: inboxIRI.String()}}, acct); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return nil, fmt.Errorf("no actor found that corresponds to inbox %s", inboxIRI.String())
		}
		return nil, fmt.Errorf("db error searching for actor with inbox %s", inboxIRI.String())
	}
	return url.Parse(acct.OutboxURI)
}
