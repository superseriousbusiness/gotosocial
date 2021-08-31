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

package messages

import "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

// FromClientAPI wraps a message that travels from the client API into the processor.
type FromClientAPI struct {
	APObjectType   string
	APActivityType string
	GTSModel       interface{}
	OriginAccount  *gtsmodel.Account
	TargetAccount  *gtsmodel.Account
}

// FromFederator wraps a message that travels from the federator into the processor.
type FromFederator struct {
	APObjectType     string
	APActivityType   string
	GTSModel         interface{}
	ReceivingAccount *gtsmodel.Account
}
