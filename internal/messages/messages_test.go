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

package messages_test

import (
	"bytes"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/testrig"

	"github.com/google/go-cmp/cmp"
)

var testStatus = testrig.NewTestStatuses()["admin_account_status_1"]

var testAccount = testrig.NewTestAccounts()["admin_account"]

var fromClientAPICases = []struct {
	msg  messages.FromClientAPI
	data []byte
}{
	{
		msg: messages.FromClientAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityCreate,
			GTSModel:       testStatus,
			TargetURI:      "https://gotosocial.org",
			Origin:         &gtsmodel.Account{ID: "654321"},
			Target:         &gtsmodel.Account{ID: "123456"},
		},
		data: toJSON(map[string]any{
			"ap_object_type":   ap.ObjectNote,
			"ap_activity_type": ap.ActivityCreate,
			"gts_model":        json.RawMessage(toJSON(testStatus)),
			"gts_model_type":   "*gtsmodel.Status",
			"target_uri":       "https://gotosocial.org",
			"origin_id":        "654321",
			"target_id":        "123456",
		}),
	},
	{
		msg: messages.FromClientAPI{
			APObjectType:   ap.ObjectProfile,
			APActivityType: ap.ActivityUpdate,
			GTSModel:       testAccount,
			TargetURI:      "https://uk-queen-is-dead.org",
			Origin:         &gtsmodel.Account{ID: "123456"},
			Target:         &gtsmodel.Account{ID: "654321"},
		},
		data: toJSON(map[string]any{
			"ap_object_type":   ap.ObjectProfile,
			"ap_activity_type": ap.ActivityUpdate,
			"gts_model":        json.RawMessage(toJSON(testAccount)),
			"gts_model_type":   "*gtsmodel.Account",
			"target_uri":       "https://uk-queen-is-dead.org",
			"origin_id":        "123456",
			"target_id":        "654321",
		}),
	},
}

var fromFediAPICases = []struct {
	msg  messages.FromFediAPI
	data []byte
}{
	{
		msg: messages.FromFediAPI{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityCreate,
			GTSModel:       testStatus,
			TargetURI:      "https://gotosocial.org",
			Requesting:     &gtsmodel.Account{ID: "654321"},
			Receiving:      &gtsmodel.Account{ID: "123456"},
		},
		data: toJSON(map[string]any{
			"ap_object_type":   ap.ObjectNote,
			"ap_activity_type": ap.ActivityCreate,
			"gts_model":        json.RawMessage(toJSON(testStatus)),
			"gts_model_type":   "*gtsmodel.Status",
			"target_uri":       "https://gotosocial.org",
			"requesting_id":    "654321",
			"receiving_id":     "123456",
		}),
	},
	{
		msg: messages.FromFediAPI{
			APObjectType:   ap.ObjectProfile,
			APActivityType: ap.ActivityUpdate,
			GTSModel:       testAccount,
			TargetURI:      "https://uk-queen-is-dead.org",
			Requesting:     &gtsmodel.Account{ID: "123456"},
			Receiving:      &gtsmodel.Account{ID: "654321"},
		},
		data: toJSON(map[string]any{
			"ap_object_type":   ap.ObjectProfile,
			"ap_activity_type": ap.ActivityUpdate,
			"gts_model":        json.RawMessage(toJSON(testAccount)),
			"gts_model_type":   "*gtsmodel.Account",
			"target_uri":       "https://uk-queen-is-dead.org",
			"requesting_id":    "123456",
			"receiving_id":     "654321",
		}),
	},
}

func TestSerializeFromClientAPI(t *testing.T) {
	for _, test := range fromClientAPICases {
		// Serialize test message to blob.
		data, err := test.msg.Serialize()
		if err != nil {
			t.Fatal(err)
		}

		// Check serialized JSON data as expected.
		assertJSONEqual(t, test.data, data)
	}
}

func TestDeserializeFromClientAPI(t *testing.T) {
	for _, test := range fromClientAPICases {
		var msg messages.FromClientAPI

		// Deserialize test message blob.
		err := msg.Deserialize(test.data)
		if err != nil {
			t.Fatal(err)
		}

		// Check that msg is as expected.
		assertEqual(t, test.msg.APActivityType, msg.APActivityType)
		assertEqual(t, test.msg.APObjectType, msg.APObjectType)
		assertEqual(t, test.msg.GTSModel, msg.GTSModel)
		assertEqual(t, test.msg.TargetURI, msg.TargetURI)
		assertEqual(t, accountID(test.msg.Origin), accountID(msg.Origin))
		assertEqual(t, accountID(test.msg.Target), accountID(msg.Target))

		// Perform final check to ensure
		// account model keys deserialized.
		assertEqualRSA(t, test.msg.GTSModel, msg.GTSModel)
	}
}

func TestSerializeFromFediAPI(t *testing.T) {
	for _, test := range fromFediAPICases {
		// Serialize test message to blob.
		data, err := test.msg.Serialize()
		if err != nil {
			t.Fatal(err)
		}

		// Check serialized JSON data as expected.
		assertJSONEqual(t, test.data, data)
	}
}

func TestDeserializeFromFediAPI(t *testing.T) {
	for _, test := range fromFediAPICases {
		var msg messages.FromFediAPI

		// Deserialize test message blob.
		err := msg.Deserialize(test.data)
		if err != nil {
			t.Fatal(err)
		}

		// Check that msg is as expected.
		assertEqual(t, test.msg.APActivityType, msg.APActivityType)
		assertEqual(t, test.msg.APObjectType, msg.APObjectType)
		assertEqual(t, urlStr(test.msg.APIRI), urlStr(msg.APIRI))
		assertEqual(t, test.msg.APObject, msg.APObject)
		assertEqual(t, test.msg.GTSModel, msg.GTSModel)
		assertEqual(t, test.msg.TargetURI, msg.TargetURI)
		assertEqual(t, accountID(test.msg.Receiving), accountID(msg.Receiving))
		assertEqual(t, accountID(test.msg.Requesting), accountID(msg.Requesting))

		// Perform final check to ensure
		// account model keys deserialized.
		assertEqualRSA(t, test.msg.GTSModel, msg.GTSModel)
	}
}

// assertEqualRSA asserts that test account model RSA keys are equal.
func assertEqualRSA(t *testing.T, expect, receive any) bool {
	t.Helper()

	account1, ok1 := expect.(*gtsmodel.Account)

	account2, ok2 := receive.(*gtsmodel.Account)

	if ok1 != ok2 {
		t.Errorf("different model types: expect=%T receive=%T", expect, receive)
		return false
	} else if !ok1 {
		return true
	}

	if !account1.PublicKey.Equal(account2.PublicKey) {
		t.Error("public keys do not match")
		return false
	}

	t.Logf("publickey=%v", account1.PublicKey)

	if !account1.PrivateKey.Equal(account2.PrivateKey) {
		t.Error("private keys do not match")
		return false
	}

	t.Logf("privatekey=%v", account1.PrivateKey)

	return true
}

// assertEqual asserts that two values (of any type!) are equal,
// note we use the 'cmp' library here as it's much more useful in
// outputting debug information than testify, and handles more complex
// types like rsa public / private key comparisons correctly.
func assertEqual(t *testing.T, expect, receive any) bool {
	t.Helper()
	if diff := cmp.Diff(expect, receive); diff != "" {
		t.Error(diff)
		return false
	}
	return true
}

// assertJSONEqual asserts that two slices of JSON data are equal.
func assertJSONEqual(t *testing.T, expect, receive []byte) bool {
	t.Helper()
	return assertEqual(t, fromJSON(expect), fromJSON(receive))
}

// urlStr returns url as string, or empty.
func urlStr(url *url.URL) string {
	if url == nil {
		return ""
	}
	return url.String()
}

// accountID returns account's ID, or empty.
func accountID(account *gtsmodel.Account) string {
	if account == nil {
		return ""
	}
	return account.ID
}

// fromJSON unmarshals input data as JSON.
func fromJSON(b []byte) any {
	r := bytes.NewReader(b)
	d := json.NewDecoder(r)
	d.UseNumber()
	var a any
	err := d.Decode(&a)
	if err != nil {
		panic(err)
	}
	if d.More() {
		panic("multiple json values in b")
	}
	return a
}

// toJSON marshals input type as JSON data.
func toJSON(a any) []byte {
	b, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	return b
}
