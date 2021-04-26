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
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

// Controller generates transports for use in making federation requests to other servers.
type Controller interface {
	NewTransport(pubKeyID string, privkey crypto.PrivateKey) (pub.Transport, error)
}

type controller struct {
	config   *config.Config
	db       db.DB
	clock    pub.Clock
	client   pub.HttpClient
	appAgent string
}

// NewController returns an implementation of the Controller interface for creating new transports
func NewController(config *config.Config, db db.DB, clock pub.Clock, client pub.HttpClient, log *logrus.Logger) Controller {
	return &controller{
		config:   config,
		db:       db,
		clock:    clock,
		client:   client,
		appAgent: fmt.Sprintf("%s %s", config.ApplicationName, config.Host),
	}
}

func (c *controller) NewTransport(pubKeyID string, privkey crypto.PrivateKey) (pub.Transport, error) {
	prefs := []httpsig.Algorithm{httpsig.Algorithm("rsa-sha256"), httpsig.Algorithm("rsa-sha512")}
	digestAlgo := httpsig.DigestAlgorithm("SHA-256")
	getHeaders := []string{"(request-target)", "Date"}
	postHeaders := []string{"(request-target)", "Date", "Digest"}

	getSigner, _, err := httpsig.NewSigner(prefs, digestAlgo, getHeaders, httpsig.Signature)
	if err != nil {
		return nil, fmt.Errorf("error creating get signer: %s", err)
	}

	postSigner, _, err := httpsig.NewSigner(prefs, digestAlgo, postHeaders, httpsig.Signature)
	if err != nil {
		return nil, fmt.Errorf("error creating post signer: %s", err)
	}

	return pub.NewHttpSigTransport(c.client, c.appAgent, c.clock, getSigner, postSigner, pubKeyID, privkey), nil
}
