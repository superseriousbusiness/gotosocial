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
	"time"

	"codeberg.org/gruf/go-kv"
	"github.com/go-fed/httpsig"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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

// PubKeyAuth models authorization information for a remote
// Actor making a signed HTTP request to this GtS instance
// using a public key.
type PubKeyAuth struct {
	// CachedPubKey is the public key found in the db
	// for the Actor whose request we're now authenticating.
	// Will be set only in cases where we had the Owner
	// of the key stored in the database already.
	CachedPubKey *rsa.PublicKey

	// FetchedPubKey is an up-to-date public key fetched
	// from the remote instance. Will be set in cases
	// where EITHER we hadn't seen the Actor before whose
	// request we're now authenticating, OR a CachedPubKey
	// was found in our database, but was expired.
	FetchedPubKey *rsa.PublicKey

	// OwnerURI is the ActivityPub id of the owner of
	// the public key used to sign the request we're
	// now authenticating. This will always be set
	// even if Owner isn't, so that callers can use
	// this URI to go fetch the Owner from remote.
	OwnerURI *url.URL

	// Owner is the account corresponding to OwnerURI.
	//
	// Owner will only be defined if the account who
	// owns the public key was already cached in the
	// database when we received the request we're now
	// authenticating (ie., we've seen it before).
	//
	// If it's not defined, callers should use OwnerURI
	// to go and dereference it.
	Owner *gtsmodel.Account
}

// AuthenticateFederatedRequest authenticates any kind of incoming federated
// request from a remote server. This includes things like GET requests for
// dereferencing our users or statuses etc, and POST requests for delivering
// new Activities. The function returns details of the public key(s) used to
// authenticate the requesting http signature.
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
func (f *federator) AuthenticateFederatedRequest(ctx context.Context, requestedUsername string) (*PubKeyAuth, gtserror.WithCode) {
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
		pubKeyIDStr = pubKeyID.String()
		local       = (pubKeyID.Host == config.GetHost())
		pubKeyAuth  *PubKeyAuth
		errWithCode gtserror.WithCode
	)

	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"requestedUsername", requestedUsername},
			{"pubKeyID", pubKeyIDStr},
		}...)

	if local {
		l.Trace("public key is local, no dereference needed")
		pubKeyAuth, errWithCode = f.derefPubKeyDBOnly(ctx, pubKeyIDStr)
	} else {
		l.Trace("public key is remote, checking if we need to dereference")
		pubKeyAuth, errWithCode = f.derefPubKey(ctx, requestedUsername, pubKeyIDStr, pubKeyID)
	}

	if errWithCode != nil {
		return nil, errWithCode
	}

	if local && pubKeyAuth == nil {
		// We signed this request, apparently, but
		// local lookup didn't find anything. This
		// is an almost impossible error condition!
		err := gtserror.Newf("local public key %s could not be found; "+
			"has the account been manually removed from the db?", pubKeyIDStr)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Try to authenticate using permitted algorithms in
	// order of most -> least common, checking each defined
	// pubKey for this Actor. Return OK as soon as one passes.
	for _, pubKey := range [2]*rsa.PublicKey{
		pubKeyAuth.FetchedPubKey,
		pubKeyAuth.CachedPubKey,
	} {
		if pubKey == nil {
			continue
		}

		for _, algo := range signingAlgorithms {
			l.Tracef("trying %s", algo)

			err := verifier.Verify(pubKey, algo)
			if err == nil {
				l.Tracef("authentication PASSED with %s", algo)
				return pubKeyAuth, nil
			}

			l.Tracef("authentication NOT PASSED with %s: %q", algo, err)
		}
	}

	// At this point no algorithms passed.
	err := gtserror.Newf(
		"authentication NOT PASSED for public key %s; tried algorithms %+v; signature value was '%s'",
		pubKeyIDStr, signature, signingAlgorithms,
	)

	return nil, gtserror.NewErrorUnauthorized(err, err.Error())
}

// derefPubKeyDBOnly tries to dereference the given
// pubKey using only entries already in the database.
//
// In case of a db or URL error, will return the error.
//
// In case an entry for the pubKey owner just doesn't
// exist in the db (yet), will return nil, nil.
func (f *federator) derefPubKeyDBOnly(
	ctx context.Context,
	pubKeyIDStr string,
) (*PubKeyAuth, gtserror.WithCode) {
	owner, err := f.db.GetAccountByPubkeyID(ctx, pubKeyIDStr)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// We don't have this
			// account stored (yet).
			return nil, nil
		}

		err = gtserror.Newf("db error getting account with pubKeyID %s: %w", pubKeyIDStr, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	ownerURI, err := url.Parse(owner.URI)
	if err != nil {
		err = gtserror.Newf("error parsing account uri with pubKeyID %s: %w", pubKeyIDStr, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return &PubKeyAuth{
		CachedPubKey: owner.PublicKey,
		OwnerURI:     ownerURI,
		Owner:        owner,
	}, nil
}

// derefPubKey tries to dereference the given public key by first
// checking in the database, and then (if no entry found, or entry
// found but pubKey expired) calling the remote pub key URI and
// extracting the key.
func (f *federator) derefPubKey(
	ctx context.Context,
	requestedUsername string,
	pubKeyIDStr string,
	pubKeyID *url.URL,
) (*PubKeyAuth, gtserror.WithCode) {
	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"requestedUsername", requestedUsername},
			{"pubKeyID", pubKeyIDStr},
		}...)

	// Try a database only deref first. We may already
	// have the requesting account cached locally.
	pubKeyAuth, errWithCode := f.derefPubKeyDBOnly(ctx, pubKeyIDStr)
	if errWithCode != nil {
		return nil, errWithCode
	}

	var (
		// Just haven't seen this
		// Actor + their pubkey yet.
		uncached = (pubKeyAuth == nil)

		// Have seen this Actor + their
		// pubkey but latter is now expired.
		expired = (!uncached && pubKeyAuth.Owner.PubKeyExpired())
	)

	switch {
	case uncached:
		l.Trace("public key was not cached, trying dereference of public key")
	case !expired:
		l.Trace("public key cached and up to date, no dereference needed")
		return pubKeyAuth, nil
	case expired:
		// This is fairly rare and it may be helpful for
		// admins to see what's going on, so log at info.
		l.Infof(
			"public key was cached, but expired at %s, trying dereference of new public key",
			pubKeyAuth.Owner.PublicKeyExpiresAt,
		)
	}

	// If we've tried to get this account before and we
	// now have a tombstone for it (ie., it's been deleted
	// from remote), don't try to dereference it again.
	gone, err := f.CheckGone(ctx, pubKeyID)
	if err != nil {
		err := gtserror.Newf("error checking for tombstone (%s): %w", pubKeyIDStr, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if gone {
		err := gtserror.Newf("account with public key is gone (%s)", pubKeyIDStr)
		return nil, gtserror.NewErrorGone(err)
	}

	// Make an http call to get the (refreshed) pubkey.
	pubKeyBytes, errWithCode := f.callForPubKey(ctx, requestedUsername, pubKeyID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Extract the key and the owner from the response.
	pubKey, pubKeyOwner, err := parsePubKeyBytes(ctx, pubKeyBytes, pubKeyID)
	if err != nil {
		err := fmt.Errorf("error parsing public key (%s): %w", pubKeyID, err)
		return nil, gtserror.NewErrorUnauthorized(err)
	}

	if !expired {
		// PubKeyResponse was nil before because
		// we had nothing cached; return the key
		// we just fetched, and nothing else.
		return &PubKeyAuth{
			FetchedPubKey: pubKey,
			OwnerURI:      pubKeyOwner,
		}, nil
	}

	// Add newly-fetched key to response.
	pubKeyAuth.FetchedPubKey = pubKey

	// If key was expired, that means we already
	// had an owner stored for it locally. Since
	// we now successfully refreshed the pub key,
	// we should update the account to reflect that.
	ownerAcct := pubKeyAuth.Owner
	ownerAcct.PublicKey = pubKeyAuth.FetchedPubKey
	ownerAcct.PublicKeyExpiresAt = time.Time{}

	l.Info("obtained a new public key to replace expired key, caching now; " +
		"authorization for this request will be attempted with both old and new keys")

	if err := f.db.UpdateAccount(
		ctx,
		ownerAcct,
		"public_key",
		"public_key_expires_at",
	); err != nil {
		err := gtserror.Newf("db error updating account with refreshed public key (%s): %w", pubKeyIDStr, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return both new and cached (now
	// expired) keys, authentication
	// will be attempted with both.
	return pubKeyAuth, nil
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

		err := gtserror.Newf("account with public key %s is gone", pubKeyID)
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
