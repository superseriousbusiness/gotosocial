package federatingdb

import (
	"context"
	"net/url"

	"github.com/sirupsen/logrus"
)

// Exists returns true if the database has an entry for the specified
// id. It may not be owned by this application instance.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Exists(c context.Context, id *url.URL) (exists bool, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Exists",
			"id":   id.String(),
		},
	)
	l.Debugf("entering EXISTS function with id %s", id.String())

	return false, nil
}
