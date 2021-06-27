package transport

import (
	"context"
	"net/url"
)

func (t *transport) BatchDeliver(c context.Context, b []byte, recipients []*url.URL) error {
	return t.sigTransport.BatchDeliver(c, b, recipients)
}

func (t *transport) Deliver(c context.Context, b []byte, to *url.URL) error {
	l := t.log.WithField("func", "Deliver")
	l.Debugf("performing POST to %s", to.String())
	return t.sigTransport.Deliver(c, b, to)
}
