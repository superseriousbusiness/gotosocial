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

import "context"

// package private context key type.
type ctxkey uint

const (
	// context keys.
	_ ctxkey = iota
	barebonesKey
)

// Barebones returns whether the "barebones" context key has been set. This
// can be used to indicate to the database, for example, that only a barebones
// model need be returned, Allowing it to skip populating sub models.
func Barebones(ctx context.Context) bool {
	_, ok := ctx.Value(barebonesKey).(struct{})
	return ok
}

// SetBarebones sets the "barebones" context flag and returns this wrapped context.
// See Barebones() for further information on the "barebones" context flag..
func SetBarebones(ctx context.Context) context.Context {
	return context.WithValue(ctx, barebonesKey, struct{}{})
}
