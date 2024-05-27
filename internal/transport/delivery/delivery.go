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

package delivery

import (
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
)

// Delivery wraps an httpclient.Request{}
// to add ActivityPub ID IRI fields of the
// outgoing activity, so that deliveries may
// be indexed (and so, dropped from queue)
// by any of these possible ID IRIs.
type Delivery struct {

	// ActorID contains the ActivityPub
	// actor ID IRI (if any) of the activity
	// being sent out by this request.
	ActorID string

	// ObjectID contains the ActivityPub
	// object ID IRI (if any) of the activity
	// being sent out by this request.
	ObjectID string

	// TargetID contains the ActivityPub
	// target ID IRI (if any) of the activity
	// being sent out by this request.
	TargetID string

	// Request is the prepared (+ wrapped)
	// httpclient.Client{} request that
	// constitutes this ActivtyPub delivery.
	Request httpclient.Request

	// internal fields.
	next time.Time
}

type delivery struct {
	ActorID  string
	ObjectID string
	TargetID string
	Method   string
	Headers  map[string][]string
	URL      string
	Body     []byte
}

// Serialize ...
func (dlv *Delivery) Serialize() ([]byte, error) {
	panic("TODO")
}

// Deserialize ...
func (dlv *Delivery) Deserialize(data []byte) error {
	panic("TODO")
}

// backoff ...
func (dlv *Delivery) backoff() time.Duration {
	if dlv.next.IsZero() {
		return 0
	}
	return time.Until(dlv.next)
}
