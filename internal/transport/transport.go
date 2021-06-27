package transport

import (
	"context"
	"crypto"
	"net/url"
	"sync"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/httpsig"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Transport wraps the pub.Transport interface with some additional
// functionality for fetching remote media.
type Transport interface {
	pub.Transport
	// DereferenceMedia fetches the bytes of the given media attachment IRI, with the expectedContentType.
	DereferenceMedia(c context.Context, iri *url.URL, expectedContentType string) ([]byte, error)
	// DereferenceInstance dereferences remote instance information, first by checking /api/v1/instance, and then by checking /.well-known/nodeinfo.
	DereferenceInstance(c context.Context, iri *url.URL) (*gtsmodel.Instance, error)
	// Finger performs a webfinger request with the given username and domain, and returns the bytes from the response body.
	Finger(c context.Context, targetUsername string, targetDomains string) ([]byte, error)
}

// transport implements the Transport interface
type transport struct {
	client       pub.HttpClient
	appAgent     string
	gofedAgent   string
	clock        pub.Clock
	pubKeyID     string
	privkey      crypto.PrivateKey
	sigTransport *pub.HttpSigTransport
	getSigner    httpsig.Signer
	getSignerMu  *sync.Mutex
	log          *logrus.Logger
}
