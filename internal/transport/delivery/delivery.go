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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
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
	Request *httpclient.Request

	// internal fields.
	next time.Time
}

// delivery is an internal type
// for Delivery{} that provides
// a json serialize / deserialize
// able shape that minimizes data.
type delivery struct {
	ActorID  string              `json:"actor_id,omitempty"`
	ObjectID string              `json:"object_id,omitempty"`
	TargetID string              `json:"target_id,omitempty"`
	Method   string              `json:"method,omitempty"`
	Header   map[string][]string `json:"header,omitempty"`
	URL      string              `json:"url,omitempty"`
	Body     []byte              `json:"body,omitempty"`
}

// Serialize will serialize the delivery data as data blob for storage,
// note that this will flatten some of the data, dropping signing funcs.
func (dlv *Delivery) Serialize() ([]byte, error) {
	var body []byte

	if dlv.Request.GetBody != nil {
		// Fetch a fresh copy of request body.
		rbody, err := dlv.Request.GetBody()
		if err != nil {
			return nil, err
		}

		// Read request body into memory.
		body, err = io.ReadAll(rbody)

		// Done with body.
		_ = rbody.Close()

		if err != nil {
			return nil, err
		}
	}

	// Marshal as internal JSON type.
	return json.Marshal(delivery{
		ActorID:  dlv.ActorID,
		ObjectID: dlv.ObjectID,
		TargetID: dlv.TargetID,
		Method:   dlv.Request.Method,
		Header:   dlv.Request.Header,
		URL:      dlv.Request.URL.String(),
		Body:     body,
	})
}

// Deserialize will attempt to deserialize a blob of task data,
// which will involve unflattening previously serialized data and
// leave delivery incomplete, still requiring signing func setup.
func (dlv *Delivery) Deserialize(data []byte) error {
	var idlv delivery

	// Unmarshal as internal JSON type.
	err := json.Unmarshal(data, &idlv)
	if err != nil {
		return err
	}

	// Copy over simplest fields.
	dlv.ActorID = idlv.ActorID
	dlv.ObjectID = idlv.ObjectID
	dlv.TargetID = idlv.TargetID

	var body io.Reader

	if idlv.Body != nil {
		// Create new body reader from data.
		body = bytes.NewReader(idlv.Body)
	}

	// Create a new request object from unmarshaled details.
	r, err := http.NewRequest(idlv.Method, idlv.URL, body)
	if err != nil {
		return err
	}

	// Copy over any stored header values.
	for key, values := range idlv.Header {
		for _, value := range values {
			r.Header.Add(key, value)
		}
	}

	// Wrap request in httpclient type.
	dlv.Request = httpclient.WrapRequest(r)

	return nil
}

// backoff returns a valid (>= 0) backoff duration.
func (dlv *Delivery) backoff() time.Duration {
	if dlv.next.IsZero() {
		return 0
	}
	return time.Until(dlv.next)
}
