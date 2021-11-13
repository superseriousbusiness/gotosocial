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
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/go-fed/httpsig"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

/*
	publicKeyer is BORROWED DIRECTLY FROM https://github.com/go-fed/apcore/blob/master/ap/util.go
	Thank you @cj@mastodon.technology ! <3
*/
type publicKeyer interface {
	GetW3IDSecurityV1PublicKey() vocab.W3IDSecurityV1PublicKeyProperty
}

/*
	getPublicKeyFromResponse is adapted from https://github.com/go-fed/apcore/blob/master/ap/util.go
	Thank you @cj@mastodon.technology ! <3
*/
func getPublicKeyFromResponse(c context.Context, b []byte, keyID *url.URL) (vocab.W3IDSecurityV1PublicKey, error) {
	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	t, err := streams.ToType(c, m)
	if err != nil {
		return nil, err
	}

	pker, ok := t.(publicKeyer)
	if !ok {
		return nil, fmt.Errorf("ActivityStreams type cannot be converted to one known to have publicKey property: %T", t)
	}

	pkp := pker.GetW3IDSecurityV1PublicKey()
	if pkp == nil {
		return nil, errors.New("publicKey property is not provided")
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
			return nil, err
		}
		if pkID.String() != keyID.String() {
			continue
		}
		pkpFound = pkValue
		break
	}

	if pkpFound == nil {
		return nil, fmt.Errorf("cannot find publicKey with id: %s", keyID)
	}

	return pkpFound, nil
}

// AuthenticateFederatedRequest authenticates any kind of incoming federated request from a remote server. This includes things like
// GET requests for dereferencing our users or statuses etc, and POST requests for delivering new Activities. The function returns
// the URL of the owner of the public key used in the requesting http signature.
//
// Authenticate in this case is defined as making sure that the http request is actually signed by whoever claims
// to have signed it, by fetching the public key from the signature and checking it against the remote public key.
//
// The provided username will be used to generate a transport for making remote requests/derefencing the public key ID of the request signature.
// Ideally you should pass in the username of the user *being requested*, so that the remote server can decide how to handle the request based on who's making it.
// Ie., if the request on this server is for https://example.org/users/some_username then you should pass in the username 'some_username'.
// The remote server will then know that this is the user making the dereferencing request, and they can decide to allow or deny the request depending on their settings.
//
// Note that it is also valid to pass in an empty string here, in which case the keys of the instance account will be used.
//
// Also note that this function *does not* dereference the remote account that the signature key is associated with.
// Other functions should use the returned URL to dereference the remote account, if required.
func (f *federator) AuthenticateFederatedRequest(ctx context.Context, requestedUsername string) (*url.URL, bool, error) {
	l := logrus.WithField("func", "AuthenticateFederatedRequest")

	var publicKey interface{}
	var pkOwnerURI *url.URL
	var err error

	// thanks to signaturecheck.go in the security package, we should already have a signature verifier set on the context
	vi := ctx.Value(util.APRequestingPublicKeyVerifier)
	if vi == nil {
		l.Debug("request wasn't signed")
		return nil, false, nil // request wasn't signed
	}

	verifier, ok := vi.(httpsig.Verifier)
	if !ok {
		l.Debug("couldn't extract sig verifier")
		return nil, false, nil // couldn't extract the verifier
	}

	// we should have the signature itself set too
	si := ctx.Value(util.APRequestingPublicKeySignature)
	if vi == nil {
		l.Debug("request wasn't signed")
		return nil, false, nil // request wasn't signed
	}

	signature, ok := si.(string)
	if !ok {
		l.Debug("couldn't extract signature")
		return nil, false, nil // couldn't extract the signature
	}

	requestingPublicKeyID, err := url.Parse(verifier.KeyId())
	if err != nil {
		l.Debug("couldn't parse public key URL")
		return nil, false, nil // couldn't parse the public key ID url
	}

	requestingRemoteAccount := &gtsmodel.Account{}
	requestingLocalAccount := &gtsmodel.Account{}
	requestingHost := requestingPublicKeyID.Host
	if strings.EqualFold(requestingHost, f.config.Host) {
		// LOCAL ACCOUNT REQUEST
		// the request is coming from INSIDE THE HOUSE so skip the remote dereferencing
		l.Tracef("proceeding without dereference for local public key %s", requestingPublicKeyID)
		if err := f.db.GetWhere(ctx, []db.Where{{Key: "public_key_uri", Value: requestingPublicKeyID.String()}}, requestingLocalAccount); err != nil {
			return nil, false, fmt.Errorf("couldn't get local account with public key uri %s from the database: %s", requestingPublicKeyID.String(), err)
		}
		publicKey = requestingLocalAccount.PublicKey
		pkOwnerURI, err = url.Parse(requestingLocalAccount.URI)
		if err != nil {
			return nil, false, fmt.Errorf("error parsing url %s: %s", requestingLocalAccount.URI, err)
		}
	} else if err := f.db.GetWhere(ctx, []db.Where{{Key: "public_key_uri", Value: requestingPublicKeyID.String()}}, requestingRemoteAccount); err == nil {
		// REMOTE ACCOUNT REQUEST WITH KEY CACHED LOCALLY
		// this is a remote account and we already have the public key for it so use that
		l.Tracef("proceeding without dereference for cached public key %s", requestingPublicKeyID)
		publicKey = requestingRemoteAccount.PublicKey
		pkOwnerURI, err = url.Parse(requestingRemoteAccount.URI)
		if err != nil {
			return nil, false, fmt.Errorf("error parsing url %s: %s", requestingRemoteAccount.URI, err)
		}
	} else {
		// REMOTE ACCOUNT REQUEST WITHOUT KEY CACHED LOCALLY
		// the request is remote and we don't have the public key yet,
		// so we need to authenticate the request properly by dereferencing the remote key
		l.Tracef("proceeding with dereference for uncached public key %s", requestingPublicKeyID)
		transport, err := f.transportController.NewTransportForUsername(ctx, requestedUsername)
		if err != nil {
			return nil, false, fmt.Errorf("transport err: %s", err)
		}

		// The actual http call to the remote server is made right here in the Dereference function.
		b, err := transport.Dereference(ctx, requestingPublicKeyID)
		if err != nil {
			return nil, false, fmt.Errorf("error deferencing key %s: %s", requestingPublicKeyID.String(), err)
		}

		// if the key isn't in the response, we can't authenticate the request
		requestingPublicKey, err := getPublicKeyFromResponse(ctx, b, requestingPublicKeyID)
		if err != nil {
			return nil, false, fmt.Errorf("error getting key %s from response %s: %s", requestingPublicKeyID.String(), string(b), err)
		}

		// we should be able to get the actual key embedded in the vocab.W3IDSecurityV1PublicKey
		pkPemProp := requestingPublicKey.GetW3IDSecurityV1PublicKeyPem()
		if pkPemProp == nil || !pkPemProp.IsXMLSchemaString() {
			return nil, false, errors.New("publicKeyPem property is not provided or it is not embedded as a value")
		}

		// and decode the PEM so that we can parse it as a golang public key
		pubKeyPem := pkPemProp.Get()
		block, _ := pem.Decode([]byte(pubKeyPem))
		if block == nil || block.Type != "PUBLIC KEY" {
			return nil, false, errors.New("could not decode publicKeyPem to PUBLIC KEY pem block type")
		}

		publicKey, err = x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, false, fmt.Errorf("could not parse public key from block bytes: %s", err)
		}

		// all good! we just need the URI of the key owner to return
		pkOwnerProp := requestingPublicKey.GetW3IDSecurityV1Owner()
		if pkOwnerProp == nil || !pkOwnerProp.IsIRI() {
			return nil, false, errors.New("publicKeyOwner property is not provided or it is not embedded as a value")
		}
		pkOwnerURI = pkOwnerProp.GetIRI()
	}

	// after all that, public key should be defined
	if publicKey == nil {
		return nil, false, errors.New("returned public key was empty")
	}

	// do the actual authentication here!
	algos := []httpsig.Algorithm{
		httpsig.RSA_SHA512,
		httpsig.RSA_SHA256,
		httpsig.ED25519,
	}

	for _, algo := range algos {
		l.Tracef("trying algo: %s", algo)
		err := verifier.Verify(publicKey, algo)
		if err == nil {
			l.Tracef("authentication for %s PASSED with algorithm %s", pkOwnerURI, algo)
			return pkOwnerURI, true, nil
		}
		l.Tracef("authentication for %s NOT PASSED with algorithm %s: %s", pkOwnerURI, algo, err)
	}

	l.Infof("authentication not passed for public key owner %s; signature value was '%s'", pkOwnerURI, signature)
	return nil, false, nil
}
