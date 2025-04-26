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

	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-kv"
	"github.com/oklog/ulid"
)

const (
	Highest     = "ZZZZZZZZZZZZZZZZZZZZZZZZZZ" // Highest is the highest possible ULID
	Lowest      = "00000000000000000000000000" // Lowest is the lowest possible ULID
	randomRange = 631152381                    // ~20 years in seconds
)

// ULID represents a Universally Unique Lexicographically Sortable Identifier of 26 characters. See https://github.com/oklog/ulid
type ULID string

// NewULID returns a new ULID string using the current time.
func NewULID() string {
	ulid, err := ulid.New(
		ulid.Timestamp(time.Now()), rand.Reader,
	)
	if err != nil {
		panic(err)
	}
	return ulid.String()
}

// NewULIDFromTime returns a new ULID string using
// given time, or from current time on any error.
func NewULIDFromTime(t time.Time) string {
	ts := ulid.Timestamp(t)
	if ts > ulid.MaxTime() {
		log.WarnKVs(nil, kv.Fields{
			{K: "caller", V: log.Caller(2)},
			{K: "value", V: t},
			{K: "msg", V: "invalid ulid time"},
		}...)
		ts = ulid.Now()
	}
	return ulid.MustNew(ts, rand.Reader).String()
}

// NewRandomULID returns a new ULID string using a random time in an ~80 year range around the current datetime, or an error if something goes wrong.
func NewRandomULID() (string, error) {
	b1, err := rand.Int(rand.Reader, big.NewInt(randomRange))
	if err != nil {
		return "", err
	}
	r1 := time.Duration(int(b1.Int64()))

	b2, err := rand.Int(rand.Reader, big.NewInt(randomRange))
	if err != nil {
		return "", err
	}
	r2 := -time.Duration(int(b2.Int64()))

	arbitraryTime := time.Now().Add(r1 * time.Second).Add(r2 * time.Second)
	newUlid, err := ulid.New(ulid.Timestamp(arbitraryTime), rand.Reader)
	if err != nil {
		return "", err
	}
	return newUlid.String(), nil
}

func TimeFromULID(id string) (time.Time, error) {
	parsed, err := ulid.ParseStrict(id)
	if err != nil {
		return time.Time{}, err
	}

	return ulid.Time(parsed.Time()), nil
}
