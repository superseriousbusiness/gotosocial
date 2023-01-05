/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package transport

import "context"

// ctxkey is our own unique context key type to prevent setting outside package.
type ctxkey string

// fastfailkey is our unique context key to indicate fast-fail is enabled.
var fastfailkey = ctxkey("ff")

// WithFastfail returns a Context which indicates that any http requests made
// with it should return after the first failed attempt, instead of retrying.
//
// This can be used to fail quickly when you're making an outgoing http request
// inside the context of an incoming http request, and you want to be able to
// provide a snappy response to the user, instead of retrying + backing off.
func WithFastfail(parent context.Context) context.Context {
	return context.WithValue(parent, fastfailkey, struct{}{})
}

// IsFastfail returns true if the given context was created by WithFastfail.
func IsFastfail(ctx context.Context) bool {
	_, ok := ctx.Value(fastfailkey).(struct{})
	return ok
}
