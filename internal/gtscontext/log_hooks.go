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

	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-kv"
)

func init() {
	// Add our required logging hooks on application initialization.
	//
	// Request ID middleware hook.
	log.Hook(func(ctx context.Context, kvs []kv.Field) []kv.Field {
		if id := RequestID(ctx); id != "" {
			return append(kvs, kv.Field{K: "requestID", V: id})
		}
		return kvs
	})
	// Public Key ID middleware hook.
	log.Hook(func(ctx context.Context, kvs []kv.Field) []kv.Field {
		if id := OutgoingPublicKeyID(ctx); id != "" {
			return append(kvs, kv.Field{K: "pubKeyID", V: id})
		}
		return kvs
	})
}
