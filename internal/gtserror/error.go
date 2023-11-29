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
	"codeberg.org/gruf/go-errors/v2"
)

// package private error key type.
type errkey int

// ErrorType denotes the type of an error, if set.
type ErrorType string

const (
	// error value keys.
	_ errkey = iota
	statusCodeKey
	notFoundKey
	errorTypeKey
	unrtrvableKey
	wrongTypeKey
	smtpKey
	malformedKey
)

// IsUnretrievable indicates that a call to retrieve a resource
// (account, status, attachment, etc) could not be fulfilled,
// either because it was not found locally, or because some
// prerequisite remote resource call failed, making it impossible
// to return the item.
func IsUnretrievable(err error) bool {
	_, ok := errors.Value(err, unrtrvableKey).(struct{})
	return ok
}

// SetUnretrievable will wrap the given error to store an "unretrievable"
// flag, returning wrapped error. See "Unretrievable" for example use-cases.
func SetUnretrievable(err error) error {
	return errors.WithValue(err, unrtrvableKey, struct{}{})
}

// IsWrongType checks error for a stored "wrong type" flag. Wrong type
// indicates that an ActivityPub URI returned a type we weren't expecting:
// Statusable instead of Accountable, or vice versa, for example.
func IsWrongType(err error) bool {
	_, ok := errors.Value(err, wrongTypeKey).(struct{})
	return ok
}

// SetWrongType will wrap the given error to store a "wrong type" flag,
// returning wrapped error. See "WrongType" for example use-cases.
func SetWrongType(err error) error {
	return errors.WithValue(err, wrongTypeKey, struct{}{})
}

// StatusCode checks error for a stored status code value. For example
// an error from an outgoing HTTP request may be stored, or an API handler
// expected response status code may be stored.
func StatusCode(err error) int {
	i, _ := errors.Value(err, statusCodeKey).(int)
	return i
}

// WithStatusCode will wrap the given error to store provided status code,
// returning wrapped error. See StatusCode() for example use-cases.
func WithStatusCode(err error, code int) error {
	return errors.WithValue(err, statusCodeKey, code)
}

// IsNotFound checks error for a stored "not found" flag. For
// example an error from an outgoing HTTP request due to DNS lookup.
func IsNotFound(err error) bool {
	_, ok := errors.Value(err, notFoundKey).(struct{})
	return ok
}

// SetNotFound will wrap the given error to store a "not found" flag,
// returning wrapped error. See NotFound() for example use-cases.
func SetNotFound(err error) error {
	return errors.WithValue(err, notFoundKey, struct{}{})
}

// IsSMTP ...
func IsSMTP(err error) bool {
	_, ok := errors.Value(err, smtpKey).(struct{})
	return ok
}

// SetSMTP ...
func SetSMTP(err error) error {
	return errors.WithValue(err, smtpKey, struct{}{})
}

// IsMalformed ...
func IsMalformed(err error) bool {
	_, ok := errors.Value(err, malformedKey).(struct{})
	return ok
}

// SetMalformed ...
func SetMalformed(err error) error {
	return errors.WithValue(err, malformedKey, struct{}{})
}
