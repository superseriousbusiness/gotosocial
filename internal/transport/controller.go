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
	"crypto"
	"fmt"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/httpsig"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// Controller generates transports for use in making federation requests to other servers.
type Controller interface {
	NewTransport(pubKeyID string, privkey crypto.PrivateKey) (pub.Transport, error)
}

type controller struct {
	config   *config.Config
	clock    pub.Clock
	client   pub.HttpClient
	appAgent string
}

// NewController returns an implementation of the Controller interface for creating new transports
func NewController(config *config.Config, clock pub.Clock, client pub.HttpClient, log *logrus.Logger) Controller {
	return &controller{
		config:   config,
		clock:    clock,
		client:   client,
		appAgent: fmt.Sprintf("%s %s", config.ApplicationName, config.Host),
	}
}

// NewTransport returns a new http signature transport with the given public key id (a URL), and the given private key.
func (c *controller) NewTransport(pubKeyID string, privkey crypto.PrivateKey) (pub.Transport, error) {
	prefs := []httpsig.Algorithm{httpsig.RSA_SHA256, httpsig.RSA_SHA512}
	digestAlgo := httpsig.DigestSha256
	getHeaders := []string{"(request-target)", "host", "date"}
	postHeaders := []string{"(request-target)", "host", "date", "digest"}

	getSigner, _, err := httpsig.NewSigner(prefs, digestAlgo, getHeaders, httpsig.Signature, 120)
	if err != nil {
		return nil, fmt.Errorf("error creating get signer: %s", err)
	}

	postSigner, _, err := httpsig.NewSigner(prefs, digestAlgo, postHeaders, httpsig.Signature, 120)
	if err != nil {
		return nil, fmt.Errorf("error creating post signer: %s", err)
	}

	return pub.NewHttpSigTransport(c.client, c.appAgent, c.clock, getSigner, postSigner, pubKeyID, privkey), nil
}
