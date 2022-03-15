/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package transport

import (
	"context"
	"crypto"
	"io"
	"net/url"
	"sync"

	"github.com/go-fed/httpsig"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Transport wraps the pub.Transport interface with some additional functionality for fetching remote media.
//
// Since the transport has the concept of 'shortcuts' for fetching data locally rather than remotely, it is
// not *always* the case that calling a Transport function does an http call, but it usually will for remote
// hosts or resources for which a shortcut isn't provided by the transport controller (also in this package).
type Transport interface {
	pub.Transport
	// DereferenceMedia fetches the given media attachment IRI, returning the reader and filesize.
	DereferenceMedia(ctx context.Context, iri *url.URL) (io.ReadCloser, int, error)
	// DereferenceInstance dereferences remote instance information, first by checking /api/v1/instance, and then by checking /.well-known/nodeinfo.
	DereferenceInstance(ctx context.Context, iri *url.URL) (*gtsmodel.Instance, error)
	// Finger performs a webfinger request with the given username and domain, and returns the bytes from the response body.
	Finger(ctx context.Context, targetUsername string, targetDomains string) ([]byte, error)
	// SigTransport returns the underlying http signature transport wrapped by the GoToSocial transport.
	SigTransport() pub.Transport
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

	// shortcuts for dereferencing things that exist on our instance without making an http call to ourself

	dereferenceFollowersShortcut func(ctx context.Context, iri *url.URL) ([]byte, error)
	dereferenceUserShortcut      func(ctx context.Context, iri *url.URL) ([]byte, error)
}

func (t *transport) SigTransport() pub.Transport {
	return t.sigTransport
}
