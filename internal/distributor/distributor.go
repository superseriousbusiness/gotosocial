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

package distributor

import (
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
)

// Distributor should be passed to api modules (see internal/apimodule/...). It is used for
// passing messages back and forth from the client API and the federating interface, via channels.
// It also contains logic for filtering which messages should end up where.
// It is designed to be used asynchronously: the client API and the federating API should just be able to
// fire messages into the distributor and not wait for a reply before proceeding with other work. This allows
// for clean distribution of messages without slowing down the client API and harming the user experience.
type Distributor interface {
	// FromClientAPI returns a channel for accepting messages that come from the gts client API.
	FromClientAPI() chan FromClientAPI
	// ClientAPIOut returns a channel for putting in messages that need to go to the gts client API.
	ToClientAPI() chan ToClientAPI
	// Start starts the Distributor, reading from its channels and passing messages back and forth.
	Start() error
	// Stop stops the distributor cleanly, finishing handling any remaining messages before closing down.
	Stop() error
}

// distributor just implements the Distributor interface
type distributor struct {
	// federator     pub.FederatingActor
	fromClientAPI chan FromClientAPI
	toClientAPI   chan ToClientAPI
	stop          chan interface{}
	log           *logrus.Logger
}

// New returns a new Distributor that uses the given federator and logger
func New(log *logrus.Logger) Distributor {
	return &distributor{
		// federator:     federator,
		fromClientAPI: make(chan FromClientAPI, 100),
		toClientAPI:   make(chan ToClientAPI, 100),
		stop:          make(chan interface{}),
		log:           log,
	}
}

// ClientAPIIn returns a channel for accepting messages that come from the gts client API.
func (d *distributor) FromClientAPI() chan FromClientAPI {
	return d.fromClientAPI
}

// ClientAPIOut returns a channel for putting in messages that need to go to the gts client API.
func (d *distributor) ToClientAPI() chan ToClientAPI {
	return d.toClientAPI
}

// Start starts the Distributor, reading from its channels and passing messages back and forth.
func (d *distributor) Start() error {
	go func() {
	DistLoop:
		for {
			select {
			case clientMsg := <-d.fromClientAPI:
				d.log.Infof("received message FROM client API: %+v", clientMsg)
			case clientMsg := <-d.toClientAPI:
				d.log.Infof("received message TO client API: %+v", clientMsg)
			case <-d.stop:
				break DistLoop
			}
		}
	}()
	return nil
}

// Stop stops the distributor cleanly, finishing handling any remaining messages before closing down.
// TODO: empty message buffer properly before stopping otherwise we'll lose federating messages.
func (d *distributor) Stop() error {
	close(d.stop)
	return nil
}

// FromClientAPI wraps a message that travels from the client API into the distributor
type FromClientAPI struct {
	APObjectType   gtsmodel.ActivityStreamsObject
	APActivityType gtsmodel.ActivityStreamsActivity
	Activity       interface{}
}

// ToClientAPI wraps a message that travels from the distributor into the client API
type ToClientAPI struct {
	APObjectType   gtsmodel.ActivityStreamsObject
	APActivityType gtsmodel.ActivityStreamsActivity
	Activity       interface{}
}
