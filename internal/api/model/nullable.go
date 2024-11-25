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

package model

import (
	"bytes"
	"encoding/json"
	"errors"
)

// Nullable is a generic type, which implements a field that can be one of three states:
//
// - field is not set in the request
// - field is explicitly set to `null` in the request
// - field is explicitly set to a valid value in the request
//
// Nullable is intended to be used with JSON unmarshalling.
//
// Adapted from https://github.com/oapi-codegen/nullable/blob/main/nullable.go
type Nullable[T any] struct {
	state nullableState
	value T
}

type nullableState uint8

const (
	nullableStateUnspecified nullableState = 0
	nullableStateNull        nullableState = 1
	nullableStateSet         nullableState = 2
)

// Get retrieves the underlying value, if present,
// and returns an error if the value was not present.
func (t Nullable[T]) Get() (T, error) {
	var empty T
	if t.IsNull() {
		return empty, errors.New("value is null")
	}

	if !t.IsSpecified() {
		return empty, errors.New("value is not specified")
	}

	return t.value, nil
}

// IsNull indicates whether the field
// was sent, and had a value of `null`
func (t Nullable[T]) IsNull() bool {
	return t.state == nullableStateNull
}

// IsSpecified indicates whether the field
// was sent either as a value or as `null`.
func (t Nullable[T]) IsSpecified() bool {
	return t.state != nullableStateUnspecified
}

// If field is unspecified,
// UnmarshalJSON won't be called.
func (t *Nullable[T]) UnmarshalJSON(data []byte) error {
	// If field is specified as `null`.
	if bytes.Equal(data, []byte("null")) {
		t.setNull()
		return nil
	}

	// Otherwise, we have an
	// actual value, so parse it.
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	t.set(v)
	return nil
}

// setNull indicates that the field
// was sent, and had a value of `null`
func (t *Nullable[T]) setNull() {
	*t = Nullable[T]{state: nullableStateNull}
}

// set the underlying value to given value.
func (t *Nullable[T]) set(value T) {
	*t = Nullable[T]{
		state: nullableStateSet,
		value: value,
	}
}
