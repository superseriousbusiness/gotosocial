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
)

// package private context key type.
type ctxkey uint

const (
	// context keys.
	_ ctxkey = iota
	barebonesKey
	fastFailKey
	pubKeyIDKey
	requestIDKey
)

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

// PublicKeyID returns the public key ID (URI) associated with context. This
// value is useful for logging situations in which a given public key URI is
// relevant, e.g. for outgoing requests being signed by the given key.
func PublicKeyID(ctx context.Context) string {
	id, _ := ctx.Value(pubKeyIDKey).(string)
	return id
}

// SetPublicKeyID stores the given public key ID value and returns the wrapped
// context. See PublicKeyID() for further information on the public key ID value.
func SetPublicKeyID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, pubKeyIDKey, id)
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
