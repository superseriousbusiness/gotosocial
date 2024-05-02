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

	// Optional GTS database model
	// of the Activity / Object.
	GTSModel interface{}

	// Targeted object URI.
	TargetURI string

	// Origin is the account that
	// this message originated from.
	Origin *gtsmodel.Account

	// Target is the account that
	// this message is targeting.
	Target *gtsmodel.Account
}

// ClientMsgIndices defines queue indices this
// message type should be accessible / stored under.
func ClientMsgIndices() []structr.IndexConfig {
	return []structr.IndexConfig{
		{Fields: "TargetURI", Multiple: true},
		{Fields: "Origin.ID", Multiple: true},
		{Fields: "Target.ID", Multiple: true},
	}
}

// FromFediAPI wraps a message that
// travels from the federating API into the processor.
type FromFediAPI struct {

	// APObjectType ...
	APObjectType string

	// APActivityType ...
	APActivityType string

	// Optional ActivityPub ID (IRI)
	// and / or model of Activity / Object.
	APIRI    *url.URL
	APObject interface{}

	// Optional GTS database model
	// of the Activity / Object.
	GTSModel interface{}

	// Targeted object URI.
	TargetURI string

	// Remote account that posted
	// this Activity to the inbox.
	Requesting *gtsmodel.Account

	// Local account which owns the inbox
	// that this Activity was posted to.
	Receiving *gtsmodel.Account
}

// FederatorMsgIndices defines queue indices this
// message type should be accessible / stored under.
func FederatorMsgIndices() []structr.IndexConfig {
	return []structr.IndexConfig{
		{Fields: "APIRI", Multiple: true},
		{Fields: "TargetURI", Multiple: true},
		{Fields: "Requesting.ID", Multiple: true},
		{Fields: "Receiving.ID", Multiple: true},
	}
}
