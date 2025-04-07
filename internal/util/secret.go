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

package util

import (
	"crypto/rand"
	"encoding/base32"
	"io"
)

// crockfordBase32 is an encoding alphabet that misses characters I,L,O,U,
// to avoid confusion and abuse. See: http://www.crockford.com/wrmg/base32.html
const crockfordBase32 = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// base32enc is a pre-initialized CrockfordBase32 encoding without any padding.
var base32enc = base32.NewEncoding(crockfordBase32).WithPadding(base32.NoPadding)

// MustGenerateSecret returns a cryptographically-secure,
// CrockfordBase32-encoded string of 32 chars in length
// (ie., 20-bytes/160 bits of entropy), or panics on error.
//
// The source of randomness is crypto/rand.
func MustGenerateSecret() string {
	// Crockford base32 with no padding
	// encodes 20 bytes to 32 characters.
	const blen = 20
	b := make([]byte, blen)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err)
	}
	return base32enc.EncodeToString(b)
}
