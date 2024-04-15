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

	"codeberg.org/gruf/go-structr"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// FromClientAPI wraps a message that
// travels from the client API into the processor.
type FromClientAPI struct {

	// APObjectType ...
	APObjectType string

	// APActivityType ...
	APActivityType string

	// Optional GTS model of
	// the Activity or Object.
	GTSModel interface{}

	// Origin ...
	Origin *gtsmodel.Account

	// Target ...
	Target *gtsmodel.Account
}

// FromFediAPI wraps a message that
// travels from the federating API into the processor.
type FromFediAPI struct {

	// APObjectType ...
	APObjectType string

	// APActivityType ...
	APActivityType string

	// APIRI ...
	APIRI *url.URL

	// Optional AP model of the Object of the
	// Activity. Likely Accountable or Statusable.
	APObject interface{}

	// Optional GTS model of
	// the Activity or Object.
	GTSModel interface{}

	// Remote account that posted
	// this Activity to the inbox.
	Requesting *gtsmodel.Account

	// Local account which owns the inbox
	// that this Activity was posted to.
	Receiving *gtsmodel.Account
}

// ClientMsgIndices ...
func ClientMsgIndices() []structr.IndexConfig {
	return []structr.IndexConfig{
		{Fields: "Origin.ID", Multiple: true},
		{Fields: "Target.ID", Multiple: true},
	}
}

// FederatorMsgIndices ...
func FederatorMsgIndices() []structr.IndexConfig {
	return []structr.IndexConfig{
		{Fields: "APIRI", Multiple: true},
		{Fields: "Requesting.ID", Multiple: true},
		{Fields: "Receiving.ID", Multiple: true},
	}
}
