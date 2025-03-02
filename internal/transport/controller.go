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
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-cache/v3"
	"codeberg.org/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

// Controller generates transports for use in making federation requests to other servers.
type Controller interface {
	// NewTransport returns an http signature transport with the given public key ID (URL location of pubkey), and the given private key.
	NewTransport(pubKeyID string, privkey *rsa.PrivateKey) (Transport, error)

	// NewTransportForUsername searches for account with username, and returns result of .NewTransport().
	NewTransportForUsername(ctx context.Context, username string) (Transport, error)
}

type controller struct {
	state     *state.State
	fedDB     federatingdb.DB
	clock     pub.Clock
	client    pub.HttpClient
	trspCache cache.TTLCache[string, *transport]
	userAgent string
}

// NewController returns an implementation of the Controller interface for creating new transports
func NewController(state *state.State, federatingDB federatingdb.DB, clock pub.Clock, client pub.HttpClient) Controller {
	var (
		host    = config.GetHost()
		proto   = config.GetProtocol()
		version = config.GetSoftwareVersion()
	)

	c := &controller{
		state:     state,
		fedDB:     federatingDB,
		clock:     clock,
		client:    client,
		trspCache: cache.NewTTL[string, *transport](0, 100, 0),
		userAgent: fmt.Sprintf("gotosocial/%s (+%s://%s)", version, proto, host),
	}

	return c
}

func (c *controller) NewTransport(pubKeyID string, privkey *rsa.PrivateKey) (Transport, error) {
	// Generate public key string for cache key
	//
	// NOTE: it is safe to use the public key as the cache
	// key here as we are generating it ourselves from the
	// private key. If we were simply using a public key
	// provided as argument that would absolutely NOT be safe.
	pubStr := privkeyToPublicStr(privkey)

	// First check for cached transport
	transp, ok := c.trspCache.Get(pubStr)
	if ok {
		return transp, nil
	}

	// Create the transport
	transp = &transport{
		controller: c,
		pubKeyID:   pubKeyID,
		privkey:    privkey,
	}

	// Cache this transport under pubkey
	if !c.trspCache.Add(pubStr, transp) {
		var cached *transport

		cached, ok = c.trspCache.Get(pubStr)
		if !ok {
			// Some ridiculous race cond.
			c.trspCache.Set(pubStr, transp)
		} else {
			// Use already cached
			transp = cached
		}
	}

	return transp, nil
}

func (c *controller) NewTransportForUsername(ctx context.Context, username string) (Transport, error) {
	// We need an account to use to create a transport for dereferecing something.
	// If a username has been given, we can fetch the account with that username and use it.
	// Otherwise, we can take the instance account and use those credentials to make the request.
	var u string
	if username == "" {
		u = config.GetHost()
	} else {
		u = username
	}

	ourAccount, err := c.state.DB.GetAccountByUsernameDomain(ctx, u, "")
	if err != nil {
		return nil, fmt.Errorf("error getting account %s from db: %s", username, err)
	}

	transport, err := c.NewTransport(ourAccount.PublicKeyURI, ourAccount.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating transport for user %s: %s", username, err)
	}

	return transport, nil
}

// dereferenceLocalFollowers is a shortcut to dereference followers of an
// account on this instance, without making any external api/http calls.
//
// It is passed to new transports, and should only be invoked when the iri.Host == this host.
func (c *controller) dereferenceLocalFollowers(ctx context.Context, iri *url.URL) (*http.Response, error) {
	followers, err := c.fedDB.Followers(ctx, iri)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, err
	}

	if followers == nil {
		// Return a generic 404 not found response.
		rsp := craftResponse(iri, http.StatusNotFound)
		return rsp, nil
	}

	i, err := ap.Serialize(followers)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	// Return a response with AS data as body.
	rsp := craftResponse(iri, http.StatusOK)
	rsp.Body = io.NopCloser(bytes.NewReader(b))
	return rsp, nil
}

// dereferenceLocalUser is a shortcut to dereference followers an account on
// this instance, without making any external api/http calls.
//
// It is passed to new transports, and should only be invoked when the iri.Host == this host.
func (c *controller) dereferenceLocalUser(ctx context.Context, iri *url.URL) (*http.Response, error) {
	user, err := c.fedDB.Get(ctx, iri)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, err
	}

	if user == nil {
		// Return a generic 404 not found response.
		rsp := craftResponse(iri, http.StatusNotFound)
		return rsp, nil
	}

	i, err := ap.Serialize(user)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	// Return a response with AS data as body.
	rsp := craftResponse(iri, http.StatusOK)
	rsp.Body = io.NopCloser(bytes.NewReader(b))
	return rsp, nil
}

// dereferenceLocalAccept is a shortcut to dereference an accept created
// by an account on this instance, without making any external api/http calls.
//
// It is passed to new transports, and should only be invoked when the iri.Host == this host.
func (c *controller) dereferenceLocalAccept(ctx context.Context, iri *url.URL) (*http.Response, error) {
	accept, err := c.fedDB.GetAccept(ctx, iri)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, err
	}

	if accept == nil {
		// Return a generic 404 not found response.
		rsp := craftResponse(iri, http.StatusNotFound)
		return rsp, nil
	}

	i, err := ap.Serialize(accept)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	// Return a response with AS data as body.
	rsp := craftResponse(iri, http.StatusOK)
	rsp.Body = io.NopCloser(bytes.NewReader(b))
	return rsp, nil
}

func craftResponse(url *url.URL, code int) *http.Response {
	rsp := new(http.Response)
	rsp.Request = new(http.Request)
	rsp.Request.URL = url
	rsp.Status = http.StatusText(code)
	rsp.StatusCode = code
	return rsp
}

// privkeyToPublicStr will create a string representation of RSA public key from private.
func privkeyToPublicStr(privkey *rsa.PrivateKey) string {
	b := x509.MarshalPKCS1PublicKey(&privkey.PublicKey)
	return byteutil.B2S(b)
}
