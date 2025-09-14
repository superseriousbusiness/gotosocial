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

package id

import (
	"crypto/rand"
	"math/big"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-kv/v2"
	"github.com/oklog/ulid"
)

const (
	Highest     = "ZZZZZZZZZZZZZZZZZZZZZZZZZZ" // Highest is the highest possible ULID
	Lowest      = "00000000000000000000000000" // Lowest is the lowest possible ULID
	randomRange = 631152381                    // ~20 years in seconds
)

// bigRandomRange contains randomRange as big.Int.
var bigRandomRange = big.NewInt(randomRange)

// ULID represents a Universally Unique Lexicographically Sortable Identifier of 26 characters. See https://github.com/oklog/ulid
type ULID string

// newAt returns a new ulid.ULID from timestamp,
// else panics with caller's caller information.
func newAt(ts uint64) string {
	ulid, err := ulid.New(ts, rand.Reader)
	if err != nil {
		panic(gtserror.NewfAt(4, "error generating ulid: %w", err))
	}
	return ulid.String()
}

// NewULID returns a new ULID
// string using the current time.
func NewULID() string {
	return newAt(ulid.Now())
}

// NewULIDFromTime returns a new ULID string using
// given time, or from current time on any error.
func NewULIDFromTime(t time.Time) string {
	ts := ulid.Timestamp(t)
	if ts > ulid.MaxTime() {
		log.WarnKVs(nil, kv.Fields{
			{K: "caller", V: log.Caller(3)},
			{K: "value", V: t},
			{K: "msg", V: "invalid ulid time"},
		}...)
		ts = ulid.Now()
	}
	return newAt(ts)
}

// NewRandomULID returns a new ULID string using a random
// time in an ~80 year range around the current datetime.
func NewRandomULID() string {
	n1, err := rand.Int(rand.Reader, bigRandomRange)
	if err != nil {
		panic(gtserror.NewfAt(3, "error reading rand: %w", err))
	}

	n2, err := rand.Int(rand.Reader, bigRandomRange)
	if err != nil {
		panic(gtserror.NewfAt(3, "error reading rand: %w", err))
	}

	// Random addition and decrement durations.
	add := time.Duration(n1.Int64()) * time.Second
	dec := -time.Duration(n2.Int64()) * time.Second

	// Return new ULID string from now.
	t := time.Now().Add(add).Add(dec)
	return newAt(ulid.Timestamp(t))
}

// TimeFromULID parses a ULID string and returns the
// encoded time.Time{}, or error with caller prefix.
func TimeFromULID(id string) (time.Time, error) {
	parsed, err := ulid.ParseStrict(id)
	if err != nil {
		return time.Time{}, gtserror.NewfAt(3, "could not extract time (malformed ulid): %w", err)
	}
	return ulid.Time(parsed.Time()), nil
}
