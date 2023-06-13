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

package federation

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"codeberg.org/gruf/go-kv"
	"github.com/go-fed/httpsig"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

var (
	errUnsigned       = errors.New("http request wasn't signed or http signature was invalid")
	signingAlgorithms = []httpsig.Algorithm{
		httpsig.RSA_SHA256, // Prefer common RSA_SHA256.
		httpsig.RSA_SHA512, // Fall back to less common RSA_SHA512.
		httpsig.ED25519,    // Try ED25519 as a long shot.
	}
)

// AuthenticateFederatedRequest authenticates any kind of incoming federated
// request from a remote server. This includes things like GET requests for
// dereferencing our users or statuses etc, and POST requests for delivering
// new Activities. The function returns the URL of the owner of the public key
// used in the requesting http signature.
//
// 'Authenticate' in this case is defined as making sure that the http request
// is actually signed by whoever claims to have signed it, by fetching the public
// key from the signature and checking it against the remote public key.
//
// The provided username will be used to generate a transport for making remote
// requests/derefencing the public key ID of the request signature. Ideally you
// should pass in the username of the user *being requested*, so that the remote
// server can decide how to handle the request based on who's making it. Ie., if
// the request on this server is for https://example.org/users/some_username then
// you should pass in the username 'some_username'. The remote server will then
// know that this is the user making the dereferencing request, and they can decide
// to allow or deny the request depending on their settings.
//
// Note that it is also valid to pass in an empty string here, in which case the
// keys of the instance account will be used.
//
// Also note that this function *does not* dereference the remote account that
// the signature key is associated with. Other functions should use the returned
// URL to dereference the remote account, if required.
func (f *federator) AuthenticateFederatedRequest(ctx context.Context, requestedUsername string) (*url.URL, gtserror.WithCode) {
	// Thanks to the signature check middleware,
	// we should already have an http signature
	// verifier set on the context. If we don't,
	// this is an unsigned request.
	verifier := gtscontext.HTTPSignatureVerifier(ctx)
	if verifier == nil {
		err := gtserror.Newf("%w", errUnsigned)
		errWithCode := gtserror.NewErrorUnauthorized(err, errUnsigned.Error(), "(verifier)")
		return nil, errWithCode
	}

	// We should have the signature itself set too.
	signature := gtscontext.HTTPSignature(ctx)
	if signature == "" {
		err := gtserror.Newf("%w", errUnsigned)
		errWithCode := gtserror.NewErrorUnauthorized(err, errUnsigned.Error(), "(signature)")
		return nil, errWithCode
	}

	// And finally the public key ID URI.
	pubKeyID := gtscontext.HTTPSignaturePubKeyID(ctx)
	if pubKeyID == nil {
		err := gtserror.Newf("%w", errUnsigned)
		errWithCode := gtserror.NewErrorUnauthorized(err, errUnsigned.Error(), "(pubKeyID)")
		return nil, errWithCode
	}

	// At this point we know the request was signed,
	// so now we need to validate the signature.

	var (
		pubKeyIDStr          = pubKeyID.String()
		requestingAccountURI *url.URL
		pubKey               interface{}
		errWithCode          gtserror.WithCode
	)

	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"requestedUsername", requestedUsername},
			{"pubKeyID", pubKeyIDStr},
		}...)

	if pubKeyID.Host == config.GetHost() {
		l.Trace("public key is ours, no dereference needed")
		requestingAccountURI, pubKey, errWithCode = f.derefDBOnly(ctx, pubKeyIDStr)
	} else {
		l.Trace("public key is not ours, checking if we need to dereference")
		requestingAccountURI, pubKey, errWithCode = f.deref(ctx, requestedUsername, pubKeyIDStr, pubKeyID)
	}

	if errWithCode != nil {
		return nil, errWithCode
	}

	// Ensure public key now defined.
	if pubKey == nil {
		err := gtserror.New("public key was nil")
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Try to authenticate using permitted algorithms in
	// order of most -> least common. Return OK as soon
	// as one passes.
	for _, algo := range signingAlgorithms {
		l.Tracef("trying %s", algo)

		err := verifier.Verify(pubKey, algo)
		if err == nil {
			l.Tracef("authentication PASSED with %s", algo)
			return requestingAccountURI, nil
		}

		l.Tracef("authentication NOT PASSED with %s: %q", algo, err)
	}

	// At this point no algorithms passed.
	err := fmt.Errorf(
		"authentication NOT PASSED for public key %s; tried algorithms %+v; signature value was '%s'",
		pubKeyIDStr, signature, signingAlgorithms,
	)

	return nil, gtserror.NewErrorUnauthorized(err, err.Error())
}

// derefDBOnly tries to dereference the given public
// key using only entries already in the database.
func (f *federator) derefDBOnly(
	ctx context.Context,
	pubKeyIDStr string,
) (*url.URL, interface{}, gtserror.WithCode) {
	reqAcct, err := f.db.GetAccountByPubkeyID(ctx, pubKeyIDStr)
	if err != nil {
		err = gtserror.Newf("db error getting account with pubKeyID %s: %w", pubKeyIDStr, err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	reqAcctURI, err := url.Parse(reqAcct.URI)
	if err != nil {
		err = gtserror.Newf("error parsing account uri with pubKeyID %s: %w", pubKeyIDStr, err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	return reqAcctURI, reqAcct.PublicKey, nil
}

// deref tries to dereference the given public key by first
// checking in the database, and then (if no entries found)
// calling the remote pub key URI and extracting the key.
func (f *federator) deref(
	ctx context.Context,
	requestedUsername string,
	pubKeyIDStr string,
	pubKeyID *url.URL,
) (*url.URL, interface{}, gtserror.WithCode) {
	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"requestedUsername", requestedUsername},
			{"pubKeyID", pubKeyIDStr},
		}...)

	// Try a database only deref first. We may already
	// have the requesting account cached locally.
	reqAcctURI, pubKey, errWithCode := f.derefDBOnly(ctx, pubKeyIDStr)
	if errWithCode == nil {
		l.Trace("public key cached, no dereference needed")
		return reqAcctURI, pubKey, nil
	}

	l.Trace("public key not cached, trying dereference")

	// If we've tried to get this account before and we
	// now have a tombstone for it (ie., it's been deleted
	// from remote), don't try to dereference it again.
	gone, err := f.CheckGone(ctx, pubKeyID)
	if err != nil {
		err := gtserror.Newf("error checking for tombstone for %s: %w", pubKeyIDStr, err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	if gone {
		err := gtserror.Newf("account with public key %s is gone", pubKeyIDStr)
		return nil, nil, gtserror.NewErrorGone(err)
	}

	// Make an http call to get the pubkey.
	pubKeyBytes, errWithCode := f.callForPubKey(ctx, requestedUsername, pubKeyID)
	if errWithCode != nil {
		return nil, nil, errWithCode
	}

	// Extract the key and the owner from the response.
	pubKey, pubKeyOwner, err := parsePubKeyBytes(ctx, pubKeyBytes, pubKeyID)
	if err != nil {
		err := fmt.Errorf("error parsing public key %s: %w", pubKeyID, err)
		return nil, nil, gtserror.NewErrorUnauthorized(err)
	}

	return pubKeyOwner, pubKey, nil
}

// callForPubKey handles the nitty gritty of actually
// making a request for the given pubKeyID with a
// transport created on behalf of requestedUsername.
func (f *federator) callForPubKey(
	ctx context.Context,
	requestedUsername string,
	pubKeyID *url.URL,
) ([]byte, gtserror.WithCode) {
	// Use a transport to dereference the remote.
	trans, err := f.transportController.NewTransportForUsername(
		// We're on a hot path: don't retry if req fails.
		gtscontext.SetFastFail(ctx),
		requestedUsername,
	)
	if err != nil {
		err = gtserror.Newf("error creating transport for %s: %w", requestedUsername, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// The actual http call to the remote server is
	// made right here by the Dereference function.
	pubKeyBytes, err := trans.Dereference(ctx, pubKeyID)
	if err == nil {
		// No problem.
		return pubKeyBytes, nil
	}

	if gtserror.StatusCode(err) == http.StatusGone {
		// 410 indicates remote public key no longer exists
		// (account deleted, moved, etc). Add a tombstone
		// to our database so that we can avoid trying to
		// dereference it in future.
		if err := f.HandleGone(ctx, pubKeyID); err != nil {
			err := gtserror.Newf("error marking public key %s as gone: %w", pubKeyID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		err := fmt.Errorf("account with public key %s is gone", pubKeyID)
		return nil, gtserror.NewErrorGone(err)
	}

	// Fall back to generic error.
	err = gtserror.Newf("error dereferencing public key %s: %w", pubKeyID, err)
	return nil, gtserror.NewErrorInternalError(err)
}

// parsePubKeyBytes extracts an rsa public key from the
// given pubKeyBytes by trying to parse the pubKeyBytes
// as an ActivityPub type. It will return the public key
// itself, and the URI of the public key owner.
func parsePubKeyBytes(
	ctx context.Context,
	pubKeyBytes []byte,
	pubKeyID *url.URL,
) (*rsa.PublicKey, *url.URL, error) {
	m := make(map[string]interface{})
	if err := json.Unmarshal(pubKeyBytes, &m); err != nil {
		return nil, nil, err
	}

	t, err := streams.ToType(ctx, m)
	if err != nil {
		return nil, nil, err
	}

	withPublicKey, ok := t.(ap.WithPublicKey)
	if !ok {
		err = gtserror.Newf("resource at %s with type %T could not be converted to ap.WithPublicKey", pubKeyID, t)
		return nil, nil, err
	}

	pubKey, _, pubKeyOwnerID, err := ap.ExtractPublicKey(withPublicKey)
	if err != nil {
		err = gtserror.Newf("resource at %s with type %T did not contain recognizable public key", pubKeyID, t)
		return nil, nil, err
	}

	return pubKey, pubKeyOwnerID, nil
}
