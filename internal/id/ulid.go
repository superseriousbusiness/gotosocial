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

package id

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/oklog/ulid"
)

const (
	Highest     = "ZZZZZZZZZZZZZZZZZZZZZZZZZZ" // Highest is the highest possible ULID
	Lowest      = "00000000000000000000000000" // Lowest is the lowest possible ULID
	randomRange = 631152381                    // ~20 years in seconds
)

// ULID represents a Universally Unique Lexicographically Sortable Identifier of 26 characters. See https://github.com/oklog/ulid
type ULID string

// NewULID returns a new ULID string using the current time, or an error if something goes wrong.
func NewULID() (string, error) {
	newUlid, err := ulid.New(ulid.Timestamp(time.Now()), rand.Reader)
	if err != nil {
		return "", err
	}
	return newUlid.String(), nil
}

// NewULIDFromTime returns a new ULID string using the given time, or an error if something goes wrong.
func NewULIDFromTime(t time.Time) (string, error) {
	newUlid, err := ulid.New(ulid.Timestamp(t), rand.Reader)
	if err != nil {
		return "", err
	}
	return newUlid.String(), nil
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
