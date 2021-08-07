package dereferencing

import (
	"context"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (d *deref) GetRemoteInstance(username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error) {
	if blocked, err := d.blockedDomain(remoteInstanceURI.Host); blocked || err != nil {
		return nil, fmt.Errorf("GetRemoteInstance: domain %s is blocked", remoteInstanceURI.Host)
	}

	transport, err := d.transportController.NewTransportForUsername(username)
	if err != nil {
		return nil, fmt.Errorf("transport err: %s", err)
	}

	return transport.DereferenceInstance(context.Background(), remoteInstanceURI)
}
