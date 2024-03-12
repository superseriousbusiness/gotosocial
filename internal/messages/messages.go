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

package messages

import (
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// FromClientAPI wraps a message that travels from the client API into the processor.
type FromClientAPI struct {
	APObjectType   string
	APActivityType string
	GTSModel       interface{}
	OriginAccount  *gtsmodel.Account
	TargetAccount  *gtsmodel.Account
}

// FromFediAPI wraps a message that travels from the federating API into the processor.
type FromFediAPI struct {
	APObjectType      string
	APActivityType    string
	APIri             *url.URL
	APObjectModel     interface{}       // Optional AP model of the Object of the Activity. Should be Accountable or Statusable.
	GTSModel          interface{}       // Optional GTS model of the Activity or Object.
	RequestingAccount *gtsmodel.Account // Remote account that posted this Activity to the inbox.
	ReceivingAccount  *gtsmodel.Account // Local account which owns the inbox that this Activity was posted to.
}
