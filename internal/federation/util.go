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
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/httpsig"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
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

// AuthenticateFederatedRequest authenticates any kind of federated request from a remote server. This includes things like
// GET requests for dereferencing users or statuses etc and POST requests for delivering new Activities.
//
// Error means the request did not pass authentication. No error means it's authentic.
//
// Authenticate in this case is defined as just making sure that the http request is actually signed by whoever claims
// to have signed it, by fetching the public key from the signature and checking it against the remote public key.
//
// The provided transport will be used to dereference the public key ID of the request signature. Ideally you should pass in a transport
// with the credentials of the user *being requested*, so that the remote server can decide how to handle the request based on who's making it.
// Ie., if the request on this server is for https://example.org/users/some_username then you should pass in a transport that's been initialized with
// the keys belonging to local user 'some_username'. The remote server will then know that this is the user making the
// dereferencing request, and they can decide to allow or deny the request depending on their settings.
//
// Note that this function *does not* dereference the remote account that the signature key is associated with, but it will
// return the public key URL associated with that account, so that other functions can dereference it with that, as required.
func AuthenticateFederatedRequest(transport pub.Transport, r *http.Request) (*url.URL, error) {
	verifier, err := httpsig.NewVerifier(r)
	if err != nil {
		return nil, fmt.Errorf("could not create http sig verifier: %s", err)
	}

	// The key ID should be given in the signature so that we know where to fetch it from the remote server.
	// This will be something like https://example.org/users/whatever_requesting_user#main-key
	requestingPublicKeyID, err := url.Parse(verifier.KeyId())
	if err != nil {
		return nil, fmt.Errorf("could not parse key id into a url: %s", err)
	}

	// use the new transport to fetch the requesting public key from the remote server
	b, err := transport.Dereference(context.Background(), requestingPublicKeyID)
	if err != nil {
		return nil, fmt.Errorf("error deferencing key %s: %s", requestingPublicKeyID.String(), err)
	}

	// if the key isn't in the response, we can't authenticate the request
	requestingPublicKey, err := getPublicKeyFromResponse(context.Background(), b, requestingPublicKeyID)
	if err != nil {
		return nil, fmt.Errorf("error getting key %s from response %s: %s", requestingPublicKeyID.String(), string(b), err)
	}

	// do the actual authentication here!
	algo := httpsig.RSA_SHA256 // TODO: make this more robust
	if err := verifier.Verify(requestingPublicKey, algo); err != nil {
		return nil, fmt.Errorf("error verifying key %s: %s", requestingPublicKeyID.String(), err)
	}

	// all good!
	return requestingPublicKeyID, nil
}

func DereferenceAccount(transport pub.Transport, publicKeyID *url.URL) (vocab.ActivityStreamsPerson, error) {
	b, err := transport.Dereference(context.Background(), publicKeyID)
	if err != nil {
		return nil, fmt.Errorf("error deferencing %s: %s", publicKeyID.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		return nil, fmt.Errorf("error resolving json into ap vocab type: %s", err)
	}

	switch t.GetTypeName() {
	case string(gtsmodel.ActivityStreamsPerson):
		p, ok := t.(vocab.ActivityStreamsPerson)
		if !ok {
			return nil, errors.New("error resolving type as activitystreams person")
		}
		return p, nil
	case string(gtsmodel.ActivityStreamsApplication):
		// TODO: convert application into person
	}

	return nil, fmt.Errorf("type name %s not supported", t.GetTypeName())
}
