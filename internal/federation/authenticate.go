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
	"io"
	"net/http"
	"net/url"
	"time"

	"code.superseriousbusiness.org/activity/streams"
	typepublickey "code.superseriousbusiness.org/activity/streams/impl/w3idsecurityv1/type_publickey"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/httpsig"
	"codeberg.org/gruf/go-kv"
)

var (
	errUnsigned = errors.New("http request wasn't signed or http signature was invalid")
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
	// now authenticating. This will always be set.
	OwnerURI *url.URL

	// Owner is the account corresponding to
	// OwnerURI. This will always be set UNLESS
	// the PubKeyAuth.Handshaking field is set..
	Owner *gtsmodel.Account

	// Handshaking indicates that uncached owner
	// account was NOT dereferenced due to an ongoing
	// handshake with another instance.
	Handshaking bool
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
// This function will handle dereferencing and storage of any new remote accounts
// and / or instances. The returned PubKeyAuth{}.Owner account will ONLY ever be
// nil in the case that there is an ongoing handshake involving this account.
//
// Note that it is also valid to pass in an empty string here, in which case the
// keys of the instance account will be used.
func (f *Federator) AuthenticateFederatedRequest(ctx context.Context, requestedUsername string) (*PubKeyAuth, gtserror.WithCode) {
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
		isLocal     = (pubKeyID.Host == config.GetHost())
		pubKeyAuth  *PubKeyAuth
		errWithCode gtserror.WithCode
	)

	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"requestedUsername", requestedUsername},
			{"pubKeyID", pubKeyIDStr},
		}...)

	if isLocal {
		l.Trace("public key is local, no dereference needed")
		pubKeyAuth, errWithCode = f.derefPubKeyDBOnly(ctx, pubKeyIDStr)
	} else {
		l.Trace("public key is remote, checking if we need to dereference")
		pubKeyAuth, errWithCode = f.derefPubKey(ctx, requestedUsername, pubKeyIDStr, pubKeyID)
	}

	if errWithCode != nil {
		return nil, errWithCode
	}

	if isLocal && pubKeyAuth == nil {
		// We signed this request, apparently, but
		// local lookup didn't find anything. This
		// is an almost impossible error condition!
		err := gtserror.Newf("local public key %s could not be found; "+
			"has the account been manually removed from the db?", pubKeyIDStr)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Attempt to verify auth with both fetched and cached keys.
	if !verifyAuth(&l, verifier, pubKeyAuth.CachedPubKey) &&
		!verifyAuth(&l, verifier, pubKeyAuth.FetchedPubKey) {

		const format = "authentication NOT PASSED for public key %s; tried algorithms %+v; signature value was '%s'"
		text := fmt.Sprintf(format, pubKeyIDStr, signingAlgorithms, signature)
		return nil, gtserror.NewErrorUnauthorized(errors.New(text), text)
	}

	if pubKeyAuth.Owner == nil {
		// Ensure we have instance stored in
		// database for the account at URI.
		err := f.fetchAccountInstance(ctx,
			requestedUsername,
			pubKeyAuth.OwnerURI,
		)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		// If we're currently handshaking with another instance, return
		// without derefing the owner, the only possible time we do this.
		// This prevents deadlocks when GTS instances mutually deref.
		if f.Handshaking(requestedUsername, pubKeyAuth.OwnerURI) {
			log.Warnf(ctx, "network race during %s handshake", pubKeyAuth.OwnerURI)
			pubKeyAuth.Handshaking = true
			return pubKeyAuth, nil
		}

		// Dereference the account located at owner URI.
		// Use exact URI match, not URL match.
		pubKeyAuth.Owner, _, err = f.GetAccountByURI(ctx,
			requestedUsername,
			pubKeyAuth.OwnerURI,
			false,
		)
		if err != nil {
			if gtserror.StatusCode(err) == http.StatusGone {
				// This can happen here instead of the pubkey 'gone'
				// checks due to: the server sending account deletion
				// notifications out, we start processing, the request above
				// succeeds, and *then* the profile is removed and starts
				// returning 410 Gone, at which point _this_ request fails.
				return nil, gtserror.NewErrorGone(err)
			}

			err := gtserror.Newf("error dereferencing account %s: %w", pubKeyAuth.OwnerURI, err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Catch a possible (but very rare) race condition where
		// we've fetched a key, then fetched the Actor who owns the
		// key, but the Key of the Actor has changed in the meantime.
		if !pubKeyAuth.Owner.PublicKey.Equal(pubKeyAuth.FetchedPubKey) {
			err := gtserror.Newf(
				"key mismatch: fetched key %s does not match pubkey of fetched Actor %s",
				pubKeyID, pubKeyAuth.Owner.URI,
			)
			return nil, gtserror.NewErrorUnauthorized(err)
		}
	}

	if !pubKeyAuth.Owner.SuspendedAt.IsZero() {
		const text = "requesting account suspended"
		return nil, gtserror.NewErrorForbidden(errors.New(text))
	}

	return pubKeyAuth, nil
}

// derefPubKeyDBOnly tries to dereference the given
// pubKey using only entries already in the database.
//
// In case of a db or URL error, will return the error.
//
// In case an entry for the pubKey owner just doesn't
// exist in the db (yet), will return nil, nil.
func (f *Federator) derefPubKeyDBOnly(
	ctx context.Context,
	pubKeyIDStr string,
) (
	*PubKeyAuth,
	gtserror.WithCode,
) {
	// Look for pubkey ID owner in the database.
	owner, err := f.db.GetAccountByPubkeyID(ctx, pubKeyIDStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting account with pubKeyID %s: %w", pubKeyIDStr, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if owner == nil {
		// We don't have this
		// account stored (yet).
		return nil, nil
	}

	// Parse owner account URI as URL obj.
	ownerURI, err := url.Parse(owner.URI)
	if err != nil {
		err := gtserror.Newf("error parsing account uri with pubKeyID %s: %w", pubKeyIDStr, err)
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
func (f *Federator) derefPubKey(
	ctx context.Context,
	requestedUsername string,
	pubKeyIDStr string,
	pubKeyID *url.URL,
) (
	*PubKeyAuth,
	gtserror.WithCode,
) {
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
		err := gtserror.Newf("error parsing public key (%s): %w", pubKeyID, err)
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
	owner := pubKeyAuth.Owner
	owner.PublicKey = pubKeyAuth.FetchedPubKey
	owner.PublicKeyExpiresAt = time.Time{}
	if err := f.db.UpdateAccount(
		ctx,
		owner,
		"public_key",
		"public_key_expires_at",
	); err != nil {
		err := gtserror.Newf("db error updating account with refreshed public key (%s): %w", pubKeyIDStr, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	l.Info("obtained new public key to replace expired; attempting auth with old / new")

	// Return both new and cached (now
	// expired) keys, authentication
	// will be attempted with both.
	return pubKeyAuth, nil
}

// callForPubKey handles the nitty gritty of actually
// making a request for the given pubKeyID with a
// transport created on behalf of requestedUsername.
func (f *Federator) callForPubKey(
	ctx context.Context,
	requestedUsername string,
	pubKeyID *url.URL,
) ([]byte, gtserror.WithCode) {
	// Use a transport to dereference the remote.
	trans, err := f.transport.NewTransportForUsername(

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
	rsp, err := trans.Dereference(ctx, pubKeyID)

	if err == nil {
		// Read the response body data.
		b, err := io.ReadAll(rsp.Body)
		_ = rsp.Body.Close() // done

		if err != nil {
			err := gtserror.Newf("error reading pubkey: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		return b, nil
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

// fetchAccountInstance ensures that an instance model exists in
// the database for the given account URI, deref'ing if necessary.
func (f *Federator) fetchAccountInstance(
	ctx context.Context,
	requestedUsername string,
	accountURI *url.URL,
) error {
	// Look for an existing entry for instance in database.
	instance, err := f.db.GetInstance(ctx, accountURI.Host)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("error getting instance from database: %w", err)
	}

	if instance != nil {
		// already fetched.
		return nil
	}

	// We don't have an entry for this
	// instance yet; go dereference it.
	instance, err = f.GetRemoteInstance(
		gtscontext.SetFastFail(ctx),
		requestedUsername,
		&url.URL{
			Scheme: accountURI.Scheme,
			Host:   accountURI.Host,
		},
	)
	if err != nil {
		return gtserror.Newf("error dereferencing instance %s: %w", accountURI.Host, err)
	}

	// Insert new instance into the datbase.
	err = f.db.PutInstance(ctx, instance)
	if err != nil && !errors.Is(err, db.ErrAlreadyExists) {
		return gtserror.Newf("error inserting instance %s into database: %w", accountURI.Host, err)
	}

	return nil
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

	var (
		pubKey   *rsa.PublicKey
		ownerURI *url.URL
	)

	if t, err := streams.ToType(ctx, m); err == nil {
		// See if Actor with a PublicKey attached.
		wpk, ok := t.(ap.WithPublicKey)
		if !ok {
			return nil, nil, gtserror.Newf(
				"resource at %s with type %T did not contain recognizable public key",
				pubKeyID, t,
			)
		}

		pubKey, _, ownerURI, err = ap.ExtractPubKeyFromActor(wpk)
		if err != nil {
			return nil, nil, gtserror.Newf(
				"error extracting public key from %T at %s: %w",
				t, pubKeyID, err,
			)
		}
	} else if pk, err := typepublickey.DeserializePublicKey(m, nil); err == nil {
		// Bare PublicKey.
		pubKey, _, ownerURI, err = ap.ExtractPubKeyFromKey(pk)
		if err != nil {
			return nil, nil, gtserror.Newf(
				"error extracting public key at %s: %w",
				pubKeyID, err,
			)
		}
	} else {
		return nil, nil, gtserror.Newf(
			"resource at %s did not contain recognizable public key",
			pubKeyID,
		)
	}

	return pubKey, ownerURI, nil
}

var signingAlgorithms = []httpsig.Algorithm{
	httpsig.RSA_SHA256, // Prefer common RSA_SHA256.
	httpsig.RSA_SHA512, // Fall back to less common RSA_SHA512.
	httpsig.ED25519,    // Try ED25519 as a long shot.
}

// Cheeky type to wrap a signing option with a
// description of that option for logging purposes.
type signingOption struct {
	desc   string                  // Description of this options set.
	sigOpt httpsig.SignatureOption // The options themselves.
}

var signingOptions = []signingOption{
	{
		// Prefer include query params.
		desc: "include query params",
		sigOpt: httpsig.SignatureOption{
			ExcludeQueryStringFromPathPseudoHeader: false,
		},
	},
	{
		// Fall back to exclude query params.
		desc: "exclude query params",
		sigOpt: httpsig.SignatureOption{
			ExcludeQueryStringFromPathPseudoHeader: true,
		},
	},
}

// verifyAuth verifies auth using generated verifier,
// according to pubkey, our supported signing algorithms,
// and signature options. The loops in the function are
// arranged in such a way that the most common combos are
// tried first, so that we can hopefully succeed quickly
// without wasting too many CPU cycles.
func verifyAuth(
	l *log.Entry,
	verifier httpsig.VerifierWithOptions,
	pubKey *rsa.PublicKey,
) bool {
	if pubKey == nil {
		return false
	}

	// Loop through supported algorithms.
	for _, algo := range signingAlgorithms {

		// Loop through signing options.
		for _, opt := range signingOptions {

			// Try to verify according to this pubkey,
			// algo, and signing options combination.
			err := verifier.VerifyWithOptions(pubKey, algo, opt.sigOpt)
			if err != nil {
				l.Tracef("authentication NOT PASSED with %s (%s): %v", algo, opt.desc, err)
				continue
			}

			l.Tracef("authenticated PASSED with %s (%s)", algo, opt.desc)
			return true
		}
	}

	return false
}
