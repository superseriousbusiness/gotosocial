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
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/httpsig"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

/*
	publicKeyer is BORROWED DIRECTLY FROM https://github.com/go-fed/apcore/blob/master/ap/util.go
	Thank you @cj@mastodon.technology ! <3
*/
type publicKeyer interface {
	GetW3IDSecurityV1PublicKey() vocab.W3IDSecurityV1PublicKeyProperty
}

/*
	getPublicKeyFromResponse is BORROWED DIRECTLY FROM https://github.com/go-fed/apcore/blob/master/ap/util.go
	Thank you @cj@mastodon.technology ! <3
*/
func getPublicKeyFromResponse(c context.Context, b []byte, keyID *url.URL) (p crypto.PublicKey, err error) {
	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	var t vocab.Type
	t, err = streams.ToType(c, m)
	if err != nil {
		return
	}
	pker, ok := t.(publicKeyer)
	if !ok {
		err = fmt.Errorf("ActivityStreams type cannot be converted to one known to have publicKey property: %T", t)
		return
	}
	pkp := pker.GetW3IDSecurityV1PublicKey()
	if pkp == nil {
		err = fmt.Errorf("publicKey property is not provided")
		return
	}
	var pkpFound vocab.W3IDSecurityV1PublicKey
	for pkpIter := pkp.Begin(); pkpIter != pkp.End(); pkpIter = pkpIter.Next() {
		if !pkpIter.IsW3IDSecurityV1PublicKey() {
			continue
		}
		pkValue := pkpIter.Get()
		var pkID *url.URL
		pkID, err = pub.GetId(pkValue)
		if err != nil {
			return
		}
		if pkID.String() != keyID.String() {
			continue
		}
		pkpFound = pkValue
		break
	}
	if pkpFound == nil {
		err = fmt.Errorf("cannot find publicKey with id: %s", keyID)
		return
	}
	pkPemProp := pkpFound.GetW3IDSecurityV1PublicKeyPem()
	if pkPemProp == nil || !pkPemProp.IsXMLSchemaString() {
		err = fmt.Errorf("publicKeyPem property is not provided or it is not embedded as a value")
		return
	}
	pubKeyPem := pkPemProp.Get()
	var block *pem.Block
	block, _ = pem.Decode([]byte(pubKeyPem))
	if block == nil || block.Type != "PUBLIC KEY" {
		err = fmt.Errorf("could not decode publicKeyPem to PUBLIC KEY pem block type")
		return
	}
	p, err = x509.ParsePKIXPublicKey(block.Bytes)
	return
}

// validateInboundFederationRequest validates an incoming federation request (!!) by deriving the public key
// of the requester from the request, checking the owner of the inbox that's being requested, and doing
// some fiddling around with http signatures.
func validateInboundFederationRequest(ctx context.Context, request *http.Request, db db.DB, inboxUsername string, transportController transport.Controller) (context.Context, bool, error) {
	v, err := httpsig.NewVerifier(request)
	if err != nil {
		return ctx, false, fmt.Errorf("could not create http sig verifier: %s", err)
	}

	requesterPublicKeyID, err := url.Parse(v.KeyId())
	if err != nil {
		return ctx, false, fmt.Errorf("could not create parse key id into a url: %s", err)
	}

	acct := &gtsmodel.Account{}
	if err := db.GetWhere("username", inboxUsername, acct); err != nil {
		return ctx, false, fmt.Errorf("could not fetch username %s from the database: %s", inboxUsername, err)
	}

	transport, err := transportController.NewTransport(acct.PublicKeyURI, acct.PrivateKey)
	if err != nil {
		return ctx, false, fmt.Errorf("error creating new transport: %s", err)
	}

	b, err := transport.Dereference(ctx, requesterPublicKeyID)
	if err != nil {
		return ctx, false, fmt.Errorf("error deferencing key %s: %s", requesterPublicKeyID.String(), err)
	}

	requesterPublicKey, err := getPublicKeyFromResponse(ctx, b, requesterPublicKeyID)
	if err != nil {
		return ctx, false, fmt.Errorf("error getting key %s from response %s: %s", requesterPublicKeyID.String(), string(b), err)
	}

	algo := httpsig.RSA_SHA256
	if err := v.Verify(requesterPublicKey, algo); err != nil {
		return ctx, false, fmt.Errorf("error verifying key %s: %s", requesterPublicKeyID.String(), err)
	}

	return ctx, true, nil
}
