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
	"crypto/x509"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-fed/httpsig"
	"github.com/sirupsen/logrus"
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
}

// transport implements the Transport interface
type transport struct {
	controller   *controller
	userAgent    string
	pubKeyID     string
	privkey      crypto.PrivateKey
	getSigner    httpsig.Signer
	getSignerMu  sync.Mutex
	postSigner   httpsig.Signer
	postSignerMu sync.Mutex
}

// GET will perform given http request using transport client, retrying on certain preset errors, or if status code is among retryOn.
func (t *transport) GET(r *http.Request, retryOn ...int) (*http.Response, error) {
	if r.Method != "GET" {
		return nil, errors.New("must be GET request")
	}
	return t.do(r, func(r *http.Request) error {
		t.getSignerMu.Lock()
		defer t.getSignerMu.Unlock()
		err := t.getSigner.SignRequest(t.privkey, t.pubKeyID, r, nil)
		return err
	}, retryOn...)
}

// POST will perform given http request using transport client, retrying on certain preset errors, or if status code is among retryOn.
func (t *transport) POST(r *http.Request, body []byte, retryOn ...int) (*http.Response, error) {
	if r.Method != "POST" {
		return nil, errors.New("must be POST request")
	}
	return t.do(r, func(r *http.Request) error {
		t.postSignerMu.Lock()
		defer t.postSignerMu.Unlock()
		err := t.postSigner.SignRequest(t.privkey, t.pubKeyID, r, body)
		return err
	}, retryOn...)
}

func (t *transport) do(r *http.Request, signer func(*http.Request) error, retryOn ...int) (*http.Response, error) {
	const maxRetries = 5
	backoff := time.Second * 2

	// Start a log entry for this request
	l := logrus.WithFields(logrus.Fields{
		"pubKeyID": t.pubKeyID,
		"method":   r.Method,
		"url":      r.URL.String(),
	})

	for i := 0; i < maxRetries; i++ {
		// Set updated request time
		now := t.controller.clock.Now().UTC()
		r.Header.Set("Date", now.Format("Mon, 02 Jan 2006 15:04:05")+" GMT")

		// Perform request signing
		if err := signer(r); err != nil {
			return nil, err
		}

		l.Infof("performing request #%d", i)

		// Attempt to perform request
		rsp, err := t.controller.client.Do(r)
		if err == nil {
			// TooManyRequest means we need to slow
			// down and retry our request. Codes over
			// 500 generally indicate temp. outages.
			if code := rsp.StatusCode; code < 500 &&
				code != http.StatusTooManyRequests &&
				!containsInt(retryOn, rsp.StatusCode) {
				return rsp, nil
			}

			// Generate error from status code for logging
			err = errors.New(`http response "` + rsp.Status + `"`)
		} else if strings.Contains(err.Error(), "stopped after 10 redirects") {
			// Don't bother if net/http returned after too many redirects
			return nil, err
		} else if errors.As(err, &x509.UnknownAuthorityError{}) {
			// Unknown authority errors we do NOT recover from
			return nil, err
		}

		l.Errorf("backing off for %s after http request error: %v", backoff.String(), err)

		// Backoff for some time
		time.Sleep(backoff)
		backoff *= 2
		continue
	}

	return nil, errors.New("transport reached max retries")
}

// containsInt checks if slice contains check.
func containsInt(slice []int, check int) bool {
	for _, i := range slice {
		if i == check {
			return true
		}
	}
	return false
}
