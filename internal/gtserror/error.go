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

	// error types
	TypeSMTP ErrorType = "smtp" // smtp (mail) error
)

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

// NotFound checks error for a stored "not found" flag. For example
// an error from an outgoing HTTP request due to DNS lookup.
func NotFound(err error) bool {
	_, ok := errors.Value(err, notFoundKey).(struct{})
	return ok
}

// SetNotFound will wrap the given error to store a "not found" flag,
// returning wrapped error. See NotFound() for example use-cases.
func SetNotFound(err error) error {
	return errors.WithValue(err, notFoundKey, struct{}{})
}

// Type checks error for a stored "type" value. For example
// an error from sending an email may set a value of "smtp"
// to indicate this was an SMTP error.
func Type(err error) ErrorType {
	s, _ := errors.Value(err, errorTypeKey).(ErrorType)
	return s
}

// SetType will wrap the given error to store a "type" value,
// returning wrapped error. See Type() for example use-cases.
func SetType(err error, errType ErrorType) error {
	return errors.WithValue(err, errorTypeKey, errType)
}
