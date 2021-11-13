/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"net/url"
	"sync"

	"github.com/go-fed/httpsig"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Transport wraps the pub.Transport interface with some additional
// functionality for fetching remote media.
type Transport interface {
	pub.Transport
	// DereferenceMedia fetches the bytes of the given media attachment IRI, with the expectedContentType.
	DereferenceMedia(ctx context.Context, iri *url.URL, expectedContentType string) ([]byte, error)
	// DereferenceInstance dereferences remote instance information, first by checking /api/v1/instance, and then by checking /.well-known/nodeinfo.
	DereferenceInstance(ctx context.Context, iri *url.URL) (*gtsmodel.Instance, error)
	// Finger performs a webfinger request with the given username and domain, and returns the bytes from the response body.
	Finger(ctx context.Context, targetUsername string, targetDomains string) ([]byte, error)
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
}
