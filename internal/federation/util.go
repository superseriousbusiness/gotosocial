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
	"net/http"
	"net/url"
	"strings"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/httpsig"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
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
// the URL of the owner of the public key used in the http signature.
//
// Authenticate in this case is defined as just making sure that the http request is actually signed by whoever claims
// to have signed it, by fetching the public key from the signature and checking it against the remote public key. This function
// *does not* check whether the request is authorized, only whether it's authentic.
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
func (f *federator) AuthenticateFederatedRequest(username string, r *http.Request) (*url.URL, error) {
	// set this extra field for signature validation
	r.Header.Set("host", f.config.Host)

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

	var publicKey interface{}
	var pkOwnerURI *url.URL
	requestingRemoteAccount := &gtsmodel.Account{}
	requestingLocalAccount := &gtsmodel.Account{}
	requestingHost := requestingPublicKeyID.Host
	if strings.EqualFold(requestingHost, f.config.Host) {
		// LOCAL ACCOUNT REQUEST
		// the request is coming from INSIDE THE HOUSE so skip the remote dereferencing
		if err := f.db.GetWhere([]db.Where{{Key: "public_key_uri", Value: requestingPublicKeyID.String()}}, requestingLocalAccount); err != nil {
			return nil, fmt.Errorf("couldn't get local account with public key uri %s from the database: %s", requestingPublicKeyID.String(), err)
		}
		publicKey = requestingLocalAccount.PublicKey
		pkOwnerURI, err = url.Parse(requestingLocalAccount.URI)
		if err != nil {
			return nil, fmt.Errorf("error parsing url %s: %s", requestingLocalAccount.URI, err)
		}
	} else if err := f.db.GetWhere([]db.Where{{Key: "public_key_uri", Value: requestingPublicKeyID.String()}}, requestingRemoteAccount); err == nil {
		// REMOTE ACCOUNT REQUEST WITH KEY CACHED LOCALLY
		// this is a remote account and we already have the public key for it so use that
		publicKey = requestingRemoteAccount.PublicKey
		pkOwnerURI, err = url.Parse(requestingRemoteAccount.URI)
		if err != nil {
			return nil, fmt.Errorf("error parsing url %s: %s", requestingRemoteAccount.URI, err)
		}
	} else {
		// REMOTE ACCOUNT REQUEST WITHOUT KEY CACHED LOCALLY
		// the request is remote and we don't have the public key yet,
		// so we need to authenticate the request properly by dereferencing the remote key
		transport, err := f.GetTransportForUser(username)
		if err != nil {
			return nil, fmt.Errorf("transport err: %s", err)
		}

		// The actual http call to the remote server is made right here in the Dereference function.
		b, err := transport.Dereference(context.Background(), requestingPublicKeyID)
		if err != nil {
			return nil, fmt.Errorf("error deferencing key %s: %s", requestingPublicKeyID.String(), err)
		}

		// if the key isn't in the response, we can't authenticate the request
		requestingPublicKey, err := getPublicKeyFromResponse(context.Background(), b, requestingPublicKeyID)
		if err != nil {
			return nil, fmt.Errorf("error getting key %s from response %s: %s", requestingPublicKeyID.String(), string(b), err)
		}

		// we should be able to get the actual key embedded in the vocab.W3IDSecurityV1PublicKey
		pkPemProp := requestingPublicKey.GetW3IDSecurityV1PublicKeyPem()
		if pkPemProp == nil || !pkPemProp.IsXMLSchemaString() {
			return nil, errors.New("publicKeyPem property is not provided or it is not embedded as a value")
		}

		// and decode the PEM so that we can parse it as a golang public key
		pubKeyPem := pkPemProp.Get()
		block, _ := pem.Decode([]byte(pubKeyPem))
		if block == nil || block.Type != "PUBLIC KEY" {
			return nil, errors.New("could not decode publicKeyPem to PUBLIC KEY pem block type")
		}

		publicKey, err = x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse public key from block bytes: %s", err)
		}

		// all good! we just need the URI of the key owner to return
		pkOwnerProp := requestingPublicKey.GetW3IDSecurityV1Owner()
		if pkOwnerProp == nil || !pkOwnerProp.IsIRI() {
			return nil, errors.New("publicKeyOwner property is not provided or it is not embedded as a value")
		}
		pkOwnerURI = pkOwnerProp.GetIRI()
	}
	if publicKey == nil {
		return nil, errors.New("returned public key was empty")
	}

	// do the actual authentication here!
	algo := httpsig.RSA_SHA256 // TODO: make this more robust
	if err := verifier.Verify(publicKey, algo); err != nil {
		return nil, fmt.Errorf("error verifying key %s: %s", requestingPublicKeyID.String(), err)
	}

	return pkOwnerURI, nil
}

func (f *federator) DereferenceRemoteAccount(username string, remoteAccountID *url.URL) (typeutils.Accountable, error) {
	f.startHandshake(username, remoteAccountID)
	defer f.stopHandshake(username, remoteAccountID)

	transport, err := f.GetTransportForUser(username)
	if err != nil {
		return nil, fmt.Errorf("transport err: %s", err)
	}

	b, err := transport.Dereference(context.Background(), remoteAccountID)
	if err != nil {
		return nil, fmt.Errorf("error deferencing %s: %s", remoteAccountID.String(), err)
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
		p, ok := t.(vocab.ActivityStreamsApplication)
		if !ok {
			return nil, errors.New("error resolving type as activitystreams application")
		}
		return p, nil
	case string(gtsmodel.ActivityStreamsService):
		p, ok := t.(vocab.ActivityStreamsService)
		if !ok {
			return nil, errors.New("error resolving type as activitystreams service")
		}
		return p, nil
	}

	return nil, fmt.Errorf("type name %s not supported", t.GetTypeName())
}

func (f *federator) DereferenceRemoteStatus(username string, remoteStatusID *url.URL) (typeutils.Statusable, error) {
	transport, err := f.GetTransportForUser(username)
	if err != nil {
		return nil, fmt.Errorf("transport err: %s", err)
	}

	b, err := transport.Dereference(context.Background(), remoteStatusID)
	if err != nil {
		return nil, fmt.Errorf("error deferencing %s: %s", remoteStatusID.String(), err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("error unmarshalling bytes into json: %s", err)
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		return nil, fmt.Errorf("error resolving json into ap vocab type: %s", err)
	}

	// Article, Document, Image, Video, Note, Page, Event, Place, Mention, Profile
	switch t.GetTypeName() {
	case gtsmodel.ActivityStreamsArticle:
		p, ok := t.(vocab.ActivityStreamsArticle)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsArticle")
		}
		return p, nil
	case gtsmodel.ActivityStreamsDocument:
		p, ok := t.(vocab.ActivityStreamsDocument)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsDocument")
		}
		return p, nil
	case gtsmodel.ActivityStreamsImage:
		p, ok := t.(vocab.ActivityStreamsImage)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsImage")
		}
		return p, nil
	case gtsmodel.ActivityStreamsVideo:
		p, ok := t.(vocab.ActivityStreamsVideo)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsVideo")
		}
		return p, nil
	case gtsmodel.ActivityStreamsNote:
		p, ok := t.(vocab.ActivityStreamsNote)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsNote")
		}
		return p, nil
	case gtsmodel.ActivityStreamsPage:
		p, ok := t.(vocab.ActivityStreamsPage)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsPage")
		}
		return p, nil
	case gtsmodel.ActivityStreamsEvent:
		p, ok := t.(vocab.ActivityStreamsEvent)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsEvent")
		}
		return p, nil
	case gtsmodel.ActivityStreamsPlace:
		p, ok := t.(vocab.ActivityStreamsPlace)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsPlace")
		}
		return p, nil
	case gtsmodel.ActivityStreamsProfile:
		p, ok := t.(vocab.ActivityStreamsProfile)
		if !ok {
			return nil, errors.New("error resolving type as ActivityStreamsProfile")
		}
		return p, nil
	}

	return nil, fmt.Errorf("type name %s not supported", t.GetTypeName())
}

func (f *federator) DereferenceRemoteInstance(username string, remoteInstanceURI *url.URL) (*apimodel.Instance, error) {
	transport, err := f.GetTransportForUser(username)
	if err != nil {
		return nil, fmt.Errorf("transport err: %s", err)
	}

	return transport.DereferenceInstance(context.Background(), remoteInstanceURI)
}

func (f *federator) GetTransportForUser(username string) (transport.Transport, error) {
	// We need an account to use to create a transport for dereferecing the signature.
	// If a username has been given, we can fetch the account with that username and use it.
	// Otherwise, we can take the instance account and use those credentials to make the request.
	ourAccount := &gtsmodel.Account{}
	var u string
	if username == "" {
		u = f.config.Host
	} else {
		u = username
	}
	if err := f.db.GetLocalAccountByUsername(u, ourAccount); err != nil {
		return nil, fmt.Errorf("error getting account %s from db: %s", username, err)
	}

	transport, err := f.transportController.NewTransport(ourAccount.PublicKeyURI, ourAccount.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating transport for user %s: %s", username, err)
	}
	return transport, nil
}
