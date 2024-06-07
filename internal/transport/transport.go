// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package transport

import (
	"bytes"
	"context"
	"crypto"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
	"github.com/superseriousbusiness/httpsig"
)

// Transport implements the pub.Transport interface with some additional functionality for fetching remote media.
//
// Since the transport has the concept of 'shortcuts' for fetching data locally rather than remotely, it is
// not *always* the case that calling a Transport function does an http call, but it usually will for remote
// hosts or resources for which a shortcut isn't provided by the transport controller (also in this package).
//
// For any of the transport functions, if a Fastfail context is passed in as the first parameter, the function
// will return after the first transport failure, instead of retrying + backing off.
type Transport interface {
	/*
		POST functions
	*/

	// POST will perform given the http request using
	// transport client, retrying on certain preset errors.
	POST(*http.Request, []byte) (*http.Response, error)

	// Deliver sends an ActivityStreams object.
	Deliver(ctx context.Context, obj map[string]interface{}, to *url.URL) error

	// BatchDeliver sends an ActivityStreams object to multiple recipients.
	BatchDeliver(ctx context.Context, obj map[string]interface{}, recipients []*url.URL) error

	/*
		GET functions
	*/

	// GET will perform the given http request using
	// transport client, retrying on certain preset errors.
	GET(*http.Request) (*http.Response, error)

	// Dereference fetches the ActivityStreams object located at this IRI with a GET request.
	Dereference(ctx context.Context, iri *url.URL) (*http.Response, error)

	// DereferenceMedia fetches the given media attachment IRI, returning the reader and filesize.
	DereferenceMedia(ctx context.Context, iri *url.URL) (io.ReadCloser, int64, error)

	// DereferenceInstance dereferences remote instance information, first by checking /api/v1/instance, and then by checking /.well-known/nodeinfo.
	DereferenceInstance(ctx context.Context, iri *url.URL) (*gtsmodel.Instance, error)

	// Finger performs a webfinger request with the given username and domain, and returns the bytes from the response body.
	Finger(ctx context.Context, targetUsername string, targetDomain string) ([]byte, error)
}

// transport implements
// the Transport interface.
type transport struct {
	controller *controller
	pubKeyID   string
	privkey    crypto.PrivateKey

	signerExp  time.Time
	getSigner  httpsig.SignerWithOptions
	postSigner httpsig.SignerWithOptions
	signerMu   sync.Mutex
}

func (t *transport) GET(r *http.Request) (*http.Response, error) {
	if r.Method != http.MethodGet {
		return nil, errors.New("must be GET request")
	}

	// Prepare HTTP GET signing func with opts.
	sign := t.signGET(httpsig.SignatureOption{
		ExcludeQueryStringFromPathPseudoHeader: false,
	})

	ctx := r.Context() // update with signing details.
	ctx = gtscontext.SetOutgoingPublicKeyID(ctx, t.pubKeyID)
	ctx = gtscontext.SetHTTPClientSignFunc(ctx, sign)
	r = r.WithContext(ctx) // replace request ctx.

	// Set our predefined controller user-agent.
	r.Header.Set("User-Agent", t.controller.userAgent)

	// Pass to underlying HTTP client.
	resp, err := t.controller.client.Do(r)
	if err != nil || resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}

	// Ignore this response.
	_ = resp.Body.Close()

	// Try again without the path included in
	// the HTTP signature for better compatibility.
	sign = t.signGET(httpsig.SignatureOption{
		ExcludeQueryStringFromPathPseudoHeader: true,
	})

	ctx = r.Context() // update with signing details.
	ctx = gtscontext.SetHTTPClientSignFunc(ctx, sign)
	r = r.WithContext(ctx) // replace request ctx.

	// Pass to underlying HTTP client.
	return t.controller.client.Do(r)
}

func (t *transport) POST(r *http.Request, body []byte) (*http.Response, error) {
	if r.Method != http.MethodPost {
		return nil, errors.New("must be POST request")
	}

	// Prepare POST signer.
	sign := t.signPOST(body)

	ctx := r.Context() // update with signing details.
	ctx = gtscontext.SetOutgoingPublicKeyID(ctx, t.pubKeyID)
	ctx = gtscontext.SetHTTPClientSignFunc(ctx, sign)
	r = r.WithContext(ctx) // replace request ctx.

	// Set our predefined controller user-agent.
	r.Header.Set("User-Agent", t.controller.userAgent)

	// Pass to underlying HTTP client.
	return t.controller.client.Do(r)
}

// newGET prepares a new get request relying on context, to recipient at URL with given signing options.
func (t *transport) newGET(ctx context.Context, url string, opts httpsig.SignatureOption) (*http.Request, error) {
	// Prepare GET signer.
	sign := t.signGET(opts)

	// Update to-be-used request context with signing details.
	ctx = gtscontext.SetOutgoingPublicKeyID(ctx, t.pubKeyID)
	ctx = gtscontext.SetHTTPClientSignFunc(ctx, sign)

	// Prepare a new request to fetch data from URL.
	r, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, gtserror.Newf("error preparing request: %w", err)
	}

	// Set our predefined controller user-agent.
	r.Header.Set("User-Agent", t.controller.userAgent)

	return r, nil
}

// newPOST prepares a new post request relying on context, to recipient at URL with given data body.
func (t *transport) newPOST(ctx context.Context, url string, data []byte) (*http.Request, error) {
	// Prepare POST signer.
	sign := t.signPOST(data)

	// Use *bytes.Reader for request body,
	// as NewRequest() automatically will
	// set .GetBody and content-length.
	body := bytes.NewReader(data)

	// Update to-be-used request context with signing details.
	ctx = gtscontext.SetOutgoingPublicKeyID(ctx, t.pubKeyID)
	ctx = gtscontext.SetHTTPClientSignFunc(ctx, sign)

	// Prepare a new request with data body directed at URL.
	r, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, gtserror.Newf("error preparing request: %w", err)
	}

	// Set our predefined controller user-agent.
	r.Header.Set("User-Agent", t.controller.userAgent)

	return r, nil
}

// signGET will safely sign an HTTP GET request.
func (t *transport) signGET(opts httpsig.SignatureOption) httpclient.SignFunc {
	return func(r *http.Request) (err error) {
		t.safesign(func() {
			err = t.getSigner.SignRequestWithOptions(t.privkey, t.pubKeyID, r, nil, opts)
		})
		return
	}
}

// signPOST will safely sign an HTTP POST request for given body.
func (t *transport) signPOST(body []byte) httpclient.SignFunc {
	return func(r *http.Request) (err error) {
		t.safesign(func() {
			err = t.postSigner.SignRequest(t.privkey, t.pubKeyID, r, body)
		})
		return
	}
}

// safesign will perform sign function within mutex protection,
// and ensured that httpsig.Signers are up-to-date.
func (t *transport) safesign(sign func()) {
	// Perform within mu safety
	t.signerMu.Lock()
	defer t.signerMu.Unlock()

	if now := time.Now(); now.After(t.signerExp) {
		const expiry = 120

		// Signers have expired and require renewal
		t.getSigner, _ = NewGETSigner(expiry)
		t.postSigner, _ = NewPOSTSigner(expiry)
		t.signerExp = now.Add(time.Second * expiry)
	}

	// Perform signing
	sign()
}
