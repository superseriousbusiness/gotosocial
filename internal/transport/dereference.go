package transport

import (
	"context"
	"net/url"
)

func (t *transport) Dereference(c context.Context, iri *url.URL) ([]byte, error) {
	l := t.log.WithField("func", "Dereference")
	l.Debugf("performing GET to %s", iri.String())
	return t.sigTransport.Dereference(c, iri)
}
