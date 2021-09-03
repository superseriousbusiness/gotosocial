package id

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/oklog/ulid"
)

const randomRange = 631152381 // ~20 years in seconds

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
