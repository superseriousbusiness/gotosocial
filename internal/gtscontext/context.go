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

package gtscontext

import (
	"context"
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/httpsig"
)

// package private context key type.
type ctxkey uint

const (
	// context keys.
	_ ctxkey = iota
	barebonesKey
	fastFailKey
	outgoingPubKeyIDKey
	requestIDKey
	receivingAccountKey
	requestingAccountKey
	otherIRIsKey
	httpSigVerifierKey
	httpSigKey
	httpSigPubKeyIDKey
	dryRunKey
	httpClientSignFnKey
)

// DryRun returns whether the "dryrun" context key has been set. This can be
// used to indicate to functions, (that support it), that only a dry-run of
// the operation should be performed. As opposed to making any permanent changes.
func DryRun(ctx context.Context) bool {
	_, ok := ctx.Value(dryRunKey).(struct{})
	return ok
}

// SetDryRun sets the "dryrun" context flag and returns this wrapped context.
// See DryRun() for further information on the "dryrun" context flag.
func SetDryRun(ctx context.Context) context.Context {
	return context.WithValue(ctx, dryRunKey, struct{}{})
}

// RequestID returns the request ID associated with context. This value will usually
// be set by the request ID middleware handler, either pulling an existing supplied
// value from request headers, or generating a unique new entry. This is useful for
// tying together log entries associated with an original incoming request.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// SetRequestID stores the given request ID value and returns the wrapped
// context. See RequestID() for further information on the request ID value.
func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// OutgoingPublicKeyID returns the public key ID (URI) associated with context. This
// value is useful for logging situations in which a given public key URI is
// relevant, e.g. for outgoing requests being signed by the given key.
func OutgoingPublicKeyID(ctx context.Context) string {
	id, _ := ctx.Value(outgoingPubKeyIDKey).(string)
	return id
}

// SetOutgoingPublicKeyID stores the given public key ID value and returns the wrapped
// context. See PublicKeyID() for further information on the public key ID value.
func SetOutgoingPublicKeyID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, outgoingPubKeyIDKey, id)
}

// ReceivingAccount returns the local account who owns the resource being
// interacted with (inbox, uri, etc) in the current ActivityPub request chain.
func ReceivingAccount(ctx context.Context) *gtsmodel.Account {
	acct, _ := ctx.Value(receivingAccountKey).(*gtsmodel.Account)
	return acct
}

// SetReceivingAccount stores the given receiving account value and returns the wrapped
// context. See ReceivingAccount() for further information on the receiving account value.
func SetReceivingAccount(ctx context.Context, acct *gtsmodel.Account) context.Context {
	return context.WithValue(ctx, receivingAccountKey, acct)
}

// RequestingAccount returns the remote account interacting with a local
// resource (inbox, uri, etc) in the current ActivityPub request chain.
func RequestingAccount(ctx context.Context) *gtsmodel.Account {
	acct, _ := ctx.Value(requestingAccountKey).(*gtsmodel.Account)
	return acct
}

// SetRequestingAccount stores the given requesting account value and returns the wrapped
// context. See RequestingAccount() for further information on the requesting account value.
func SetRequestingAccount(ctx context.Context, acct *gtsmodel.Account) context.Context {
	return context.WithValue(ctx, requestingAccountKey, acct)
}

// OtherIRIs returns other IRIs which are involved in the current ActivityPub request
// chain. This usually means: other accounts who are mentioned, CC'd, TO'd, or boosted
// by the current inbox POST request.
func OtherIRIs(ctx context.Context) []*url.URL {
	iris, _ := ctx.Value(otherIRIsKey).([]*url.URL)
	return iris
}

// SetOtherIRIs stores the given IRIs slice and returns the wrapped context.
// See OtherIRIs() for further information on the IRIs slice value.
func SetOtherIRIs(ctx context.Context, iris []*url.URL) context.Context {
	return context.WithValue(ctx, otherIRIsKey, iris)
}

// HTTPClientSignFunc returns an httpclient signing function for the current client
// request context. This can be used to resign a request as calling transport's user.
func HTTPClientSignFunc(ctx context.Context) func(*http.Request) error {
	fn, _ := ctx.Value(httpClientSignFnKey).(func(*http.Request) error)
	return fn
}

// SetHTTPClientSignFunc stores the given httpclient signing function and returns the wrapped
// context. See HTTPClientSignFunc() for further information on the signing function value.
func SetHTTPClientSignFunc(ctx context.Context, fn func(*http.Request) error) context.Context {
	return context.WithValue(ctx, httpClientSignFnKey, fn)
}

// HTTPSignatureVerifier returns an http signature verifier for the current ActivityPub
// request chain. This verifier can be called to authenticate the current request.
func HTTPSignatureVerifier(ctx context.Context) httpsig.VerifierWithOptions {
	verifier, _ := ctx.Value(httpSigVerifierKey).(httpsig.VerifierWithOptions)
	return verifier
}

// SetHTTPSignatureVerifier stores the given http signature verifier and returns the
// wrapped context. See HTTPSignatureVerifier() for further information on the verifier value.
func SetHTTPSignatureVerifier(ctx context.Context, verifier httpsig.VerifierWithOptions) context.Context {
	return context.WithValue(ctx, httpSigVerifierKey, verifier)
}

// HTTPSignature returns the http signature string
// value for the current ActivityPub request chain.
func HTTPSignature(ctx context.Context) string {
	signature, _ := ctx.Value(httpSigKey).(string)
	return signature
}

// SetHTTPSignature stores the given http signature string and returns the wrapped
// context. See HTTPSignature() for further information on the verifier value.
func SetHTTPSignature(ctx context.Context, signature string) context.Context {
	return context.WithValue(ctx, httpSigKey, signature)
}

// HTTPSignaturePubKeyID returns the public key id of the http signature
// for the current ActivityPub request chain.
func HTTPSignaturePubKeyID(ctx context.Context) *url.URL {
	pubKeyID, _ := ctx.Value(httpSigPubKeyIDKey).(*url.URL)
	return pubKeyID
}

// SetHTTPSignaturePubKeyID stores the given http signature public key id and returns
// the wrapped context. See HTTPSignaturePubKeyID() for further information on the value.
func SetHTTPSignaturePubKeyID(ctx context.Context, pubKeyID *url.URL) context.Context {
	return context.WithValue(ctx, httpSigPubKeyIDKey, pubKeyID)
}

// IsFastFail returns whether the "fastfail" context key has been set. This
// can be used to indicate to an http client, for example, that the result
// of an outgoing request is time sensitive and so not to bother with retries.
func IsFastfail(ctx context.Context) bool {
	_, ok := ctx.Value(fastFailKey).(struct{})
	return ok
}

// SetFastFail sets the "fastfail" context flag and returns this wrapped context.
// See IsFastFail() for further information on the "fastfail" context flag.
func SetFastFail(ctx context.Context) context.Context {
	return context.WithValue(ctx, fastFailKey, struct{}{})
}

// Barebones returns whether the "barebones" context key has been set. This
// can be used to indicate to the database, for example, that only a barebones
// model need be returned, Allowing it to skip populating sub models.
func Barebones(ctx context.Context) bool {
	_, ok := ctx.Value(barebonesKey).(struct{})
	return ok
}

// SetBarebones sets the "barebones" context flag and returns this wrapped context.
// See Barebones() for further information on the "barebones" context flag.
func SetBarebones(ctx context.Context) context.Context {
	return context.WithValue(ctx, barebonesKey, struct{}{})
}
