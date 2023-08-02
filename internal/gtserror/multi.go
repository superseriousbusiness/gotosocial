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

package gtserror

import (
	"errors"
	"fmt"
)

// MultiError allows encapsulating multiple
// errors under a singular instance, which
// is useful when you only want to log on
// errors, not return early / bubble up.
type MultiError struct {
	e []error
}

// NewMultiError returns a *MultiError with
// the capacity of its underlying error slice
// set to the provided value.
//
// This capacity can be exceeded if necessary,
// but it saves a teeny tiny bit of memory if
// callers set it correctly.
//
// If you don't know in advance what the capacity
// must be, just use new(MultiError) instead.
func NewMultiError(capacity int) *MultiError {
	return &MultiError{
		e: make([]error, 0, capacity),
	}
}

// Append the given error to the MultiError.
func (m *MultiError) Append(err error) {
	m.e = append(m.e, err)
}

// Append the given format string to the MultiError.
//
// It is valid to use %w in the format string
// to wrap any other errors.
func (m *MultiError) Appendf(format string, args ...any) {
	m.e = append(m.e, fmt.Errorf(format, args...))
}

// Combine the MultiError into a single error.
//
// Unwrap will work on the returned error as expected.
func (m MultiError) Combine() error {
	return errors.Join(m.e...)
}
