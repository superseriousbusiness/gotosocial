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
	"fmt"
	"sync"

	"github.com/go-fed/httpsig"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

// Controller generates transports for use in making federation requests to other servers.
type Controller interface {
	NewTransport(pubKeyID string, privkey crypto.PrivateKey) (Transport, error)
	NewTransportForUsername(ctx context.Context, username string) (Transport, error)
}

type controller struct {
	config   *config.Config
	db       db.DB
	clock    pub.Clock
	client   pub.HttpClient
	appAgent string
}

// NewController returns an implementation of the Controller interface for creating new transports
func NewController(config *config.Config, db db.DB, clock pub.Clock, client pub.HttpClient) Controller {
	return &controller{
		config:   config,
		db:       db,
		clock:    clock,
		client:   client,
		appAgent: fmt.Sprintf("%s %s", config.ApplicationName, config.Host),
	}
}

// NewTransport returns a new http signature transport with the given public key id (a URL), and the given private key.
func (c *controller) NewTransport(pubKeyID string, privkey crypto.PrivateKey) (Transport, error) {
	prefs := []httpsig.Algorithm{httpsig.RSA_SHA256}
	digestAlgo := httpsig.DigestSha256
	getHeaders := []string{httpsig.RequestTarget, "host", "date"}
	postHeaders := []string{httpsig.RequestTarget, "host", "date", "digest"}

	getSigner, _, err := httpsig.NewSigner(prefs, digestAlgo, getHeaders, httpsig.Signature, 120)
	if err != nil {
		return nil, fmt.Errorf("error creating get signer: %s", err)
	}

	postSigner, _, err := httpsig.NewSigner(prefs, digestAlgo, postHeaders, httpsig.Signature, 120)
	if err != nil {
		return nil, fmt.Errorf("error creating post signer: %s", err)
	}

	sigTransport := pub.NewHttpSigTransport(c.client, c.appAgent, c.clock, getSigner, postSigner, pubKeyID, privkey)

	return &transport{
		client:       c.client,
		appAgent:     c.appAgent,
		gofedAgent:   "(go-fed/activity v1.0.0)",
		clock:        c.clock,
		pubKeyID:     pubKeyID,
		privkey:      privkey,
		sigTransport: sigTransport,
		getSigner:    getSigner,
		getSignerMu:  &sync.Mutex{},
	}, nil
}

func (c *controller) NewTransportForUsername(ctx context.Context, username string) (Transport, error) {
	// We need an account to use to create a transport for dereferecing something.
	// If a username has been given, we can fetch the account with that username and use it.
	// Otherwise, we can take the instance account and use those credentials to make the request.
	var u string
	if username == "" {
		u = c.config.Host
	} else {
		u = username
	}

	ourAccount, err := c.db.GetLocalAccountByUsername(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("error getting account %s from db: %s", username, err)
	}

	transport, err := c.NewTransport(ourAccount.PublicKeyURI, ourAccount.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating transport for user %s: %s", username, err)
	}
	return transport, nil
}
