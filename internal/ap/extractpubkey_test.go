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

package ap_test

import (
	"context"
	"encoding/json"
	"testing"

	"codeberg.org/superseriousbusiness/activity/streams"
	typepublickey "codeberg.org/superseriousbusiness/activity/streams/impl/w3idsecurityv1/type_publickey"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
)

const (
	stubActor = `{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    "https://w3id.org/security/v1"
  ],
  "id": "https://gts.superseriousbusiness.org/users/dumpsterqueer",
  "preferredUsername": "dumpsterqueer",
  "publicKey": {
    "id": "https://gts.superseriousbusiness.org/users/dumpsterqueer/main-key",
    "owner": "https://gts.superseriousbusiness.org/users/dumpsterqueer",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAt7cDz2XfTJXbmmmVXZ3o\nQGB1zu1yP+2/QZZFbLCeM0bMm5cfjJ/olli6kpdcGLh1lFpSgyLE0PlAVNYdSke9\nzcxDao6N16wavFx/bOYhh8HJPPXzlFpNeQQ+EBQ1ivzuLQyzIFTMV4TyZzOREoG9\nizuXuuKDaH/ENDE6qlIDuqtICIjnURjpxnBLldPUxfUvuSO3zY+jTidsxhjUjqkK\nC7RtEVi/D6/CzktVevz5bE/gcAYgKmK0dmkJ9HH6LzOlvkM4Wrq5h/hrM+H1z5e5\nPpdJsl3KlRT4wusuM1Z5xqLQ0oIP4mX/Kd3ypCe150i+jaoCsqBk8OPtl/zKMw1a\nYQIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "type": "Person"
}`

	key = `{
  "@context": "https://w3id.org/security/v1",
  "id": "https://gts.superseriousbusiness.org/users/dumpsterqueer/main-key",
  "owner": "https://gts.superseriousbusiness.org/users/dumpsterqueer",
  "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAt7cDz2XfTJXbmmmVXZ3o\nQGB1zu1yP+2/QZZFbLCeM0bMm5cfjJ/olli6kpdcGLh1lFpSgyLE0PlAVNYdSke9\nzcxDao6N16wavFx/bOYhh8HJPPXzlFpNeQQ+EBQ1ivzuLQyzIFTMV4TyZzOREoG9\nizuXuuKDaH/ENDE6qlIDuqtICIjnURjpxnBLldPUxfUvuSO3zY+jTidsxhjUjqkK\nC7RtEVi/D6/CzktVevz5bE/gcAYgKmK0dmkJ9HH6LzOlvkM4Wrq5h/hrM+H1z5e5\nPpdJsl3KlRT4wusuM1Z5xqLQ0oIP4mX/Kd3ypCe150i+jaoCsqBk8OPtl/zKMw1a\nYQIDAQAB\n-----END PUBLIC KEY-----\n"
}`
)

type ExtractPubKeyTestSuite struct {
	APTestSuite
}

func (suite *ExtractPubKeyTestSuite) TestExtractPubKeyFromStub() {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(stubActor), &m); err != nil {
		suite.FailNow(err.Error())
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		suite.FailNow(err.Error())
	}

	wpk, ok := t.(ap.WithPublicKey)
	if !ok {
		suite.FailNow("", "could not parse %T as WithPublicKey", t)
	}

	pubKey, pubKeyID, ownerURI, err := ap.ExtractPubKeyFromActor(wpk)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotNil(pubKey)
	suite.Equal("https://gts.superseriousbusiness.org/users/dumpsterqueer/main-key", pubKeyID.String())
	suite.Equal("https://gts.superseriousbusiness.org/users/dumpsterqueer", ownerURI.String())
}

func (suite *ExtractPubKeyTestSuite) TestExtractPubKeyFromKey() {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(key), &m); err != nil {
		suite.FailNow(err.Error())
	}

	pk, err := typepublickey.DeserializePublicKey(m, nil)
	if err != nil {
		suite.FailNow(err.Error())
	}

	pubKey, pubKeyID, ownerURI, err := ap.ExtractPubKeyFromKey(pk)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotNil(pubKey)
	suite.Equal("https://gts.superseriousbusiness.org/users/dumpsterqueer/main-key", pubKeyID.String())
	suite.Equal("https://gts.superseriousbusiness.org/users/dumpsterqueer", ownerURI.String())
}

func TestExtractPubKeyTestSuite(t *testing.T) {
	suite.Run(t, &ExtractPubKeyTestSuite{})
}
