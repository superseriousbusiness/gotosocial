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

package delivery_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/httpclient"
	"code.superseriousbusiness.org/gotosocial/internal/transport/delivery"
	"github.com/stretchr/testify/assert"
)

var deliveryCases = []struct {
	msg  delivery.Delivery
	data []byte
}{
	{
		msg: delivery.Delivery{
			ActorID:  "https://google.com/users/bigboy",
			ObjectID: "https://google.com/users/bigboy/follow/1",
			TargetID: "https://askjeeves.com/users/smallboy",
			Request:  toRequest("POST", "https://askjeeves.com/users/smallboy/inbox", []byte("data!"), http.Header{"Hello": {"world1", "world2"}}),
		},
		data: toJSON(map[string]any{
			"actor_id":  "https://google.com/users/bigboy",
			"object_id": "https://google.com/users/bigboy/follow/1",
			"target_id": "https://askjeeves.com/users/smallboy",
			"method":    "POST",
			"url":       "https://askjeeves.com/users/smallboy/inbox",
			"body":      []byte("data!"),
			"header":    map[string][]string{"Hello": {"world1", "world2"}},
		}),
	},
	{
		msg: delivery.Delivery{
			Request: toRequest("GET", "https://google.com", []byte("uwu im just a wittle seawch engwin"), nil),
		},
		data: toJSON(map[string]any{
			"method": "GET",
			"url":    "https://google.com",
			"body":   []byte("uwu im just a wittle seawch engwin"),
			// "header": map[string][]string{},
		}),
	},
}

func TestSerializeDelivery(t *testing.T) {
	for _, test := range deliveryCases {
		// Serialize test message to blob.
		data, err := test.msg.Serialize()
		if err != nil {
			t.Fatal(err)
		}

		// Check that serialized JSON data is as expected.
		assert.JSONEq(t, string(test.data), string(data))
	}
}

func TestDeserializeDelivery(t *testing.T) {
	for _, test := range deliveryCases {
		var msg delivery.Delivery

		// Deserialize test message blob.
		err := msg.Deserialize(test.data)
		if err != nil {
			t.Fatal(err)
		}

		// Check that delivery fields are as expected.
		assert.Equal(t, test.msg.ActorID, msg.ActorID)
		assert.Equal(t, test.msg.ObjectID, msg.ObjectID)
		assert.Equal(t, test.msg.TargetID, msg.TargetID)
		assert.Equal(t, test.msg.Request.Method, msg.Request.Method)
		assert.Equal(t, test.msg.Request.URL, msg.Request.URL)
		assert.Equal(t, readBody(test.msg.Request.Body), readBody(msg.Request.Body))
		assert.Equal(t, test.msg.Request.Header, msg.Request.Header)
	}
}

// toRequest creates httpclient.Request from HTTP method, URL and body data.
func toRequest(method string, url string, body []byte, hdr http.Header) *httpclient.Request {
	var rbody io.Reader
	if body != nil {
		rbody = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, rbody)
	if err != nil {
		panic(err)
	}
	for key, values := range hdr {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return httpclient.WrapRequest(req)
}

// readBody reads the content of body io.ReadCloser into memory as byte slice.
func readBody(r io.ReadCloser) []byte {
	if r == nil {
		return nil
	}
	b, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return b
}

// toJSON marshals input type as JSON data.
func toJSON(a any) []byte {
	b, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	return b
}
