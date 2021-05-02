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

package federation

import (
	"github.com/go-fed/activity/pub"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/distributor"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// Federator wraps various interfaces and functions to manage activitypub federation from gotosocial
type Federator interface {
	FederatingActor() pub.FederatingActor
	TransportController() transport.Controller
	FederatingProtocol() pub.FederatingProtocol
	CommonBehavior() pub.CommonBehavior
}

type federator struct {
	actor               pub.FederatingActor
	distributor         distributor.Distributor
	federatingProtocol  pub.FederatingProtocol
	commonBehavior      pub.CommonBehavior
	clock               pub.Clock
	transportController transport.Controller
}

// NewFederator returns a new federator
func NewFederator(db db.DB, transportController transport.Controller, config *config.Config, log *logrus.Logger, distributor distributor.Distributor, typeConverter typeutils.TypeConverter) Federator {

	clock := &Clock{}
	federatingProtocol := newFederatingProtocol(db, log, config, transportController, typeConverter)
	commonBehavior := newCommonBehavior(db, log, config, transportController)
	actor := newFederatingActor(commonBehavior, federatingProtocol, db.Federation(), clock)

	return &federator{
		actor:               actor,
		distributor:         distributor,
		federatingProtocol:  federatingProtocol,
		commonBehavior:      commonBehavior,
		clock:               clock,
		transportController: transportController,
	}
}

func (f *federator) FederatingActor() pub.FederatingActor {
	return f.actor
}

func (f *federator) TransportController() transport.Controller {
	return f.transportController
}

func (f *federator) FederatingProtocol() pub.FederatingProtocol {
	return f.federatingProtocol
}

func (f *federator) CommonBehavior() pub.CommonBehavior {
	return f.commonBehavior
}
