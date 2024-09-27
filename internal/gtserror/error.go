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
	notRelevantKey
	spamKey
	notPermittedKey
	limitReachedKey
)

// LimitReached indicates that this error was caused by
// some kind of limit being reached, e.g. media upload limit.
func LimitReached(err error) bool {
	_, ok := errors.Value(err, limitReachedKey).(struct{})
	return ok
}

// SetLimitReached will wrap the given error to store a "limit reached"
// flag, returning wrapped error. See LimitReached() for example use-cases.
func SetLimitReached(err error) error {
	return errors.WithValue(err, limitReachedKey, struct{}{})
}

// IsUnretrievable indicates that a call to retrieve a resource
// (account, status, attachment, etc) could not be fulfilled, either
// because it was not found locally, or because some prerequisite
// remote resource call failed, making it impossible to return it.
func IsUnretrievable(err error) bool {
	_, ok := errors.Value(err, unrtrvableKey).(struct{})
	return ok
}

// SetUnretrievable will wrap the given error to store an "unretrievable"
// flag, returning wrapped error. See Unretrievable() for example use-cases.
func SetUnretrievable(err error) error {
	return errors.WithValue(err, unrtrvableKey, struct{}{})
}

// NotPermitted indicates that some call failed due to failed permission
// or acceptibility checks. For example an attempt to dereference remote
// status in which the status author does not have permission to reply
// to the status it is intended to be replying to.
func NotPermitted(err error) bool {
	_, ok := errors.Value(err, notPermittedKey).(struct{})
	return ok
}

// SetNotPermitted will wrap the given error to store a "not permitted"
// flag, returning wrapped error. See NotPermitted() for example use-cases.
func SetNotPermitted(err error) error {
	return errors.WithValue(err, notPermittedKey, struct{}{})
}

// IsWrongType checks error for a stored "wrong type" flag.
// Wrong type indicates that an ActivityPub URI returned a
// type we weren't expecting. For example:
//
//   - HTML instead of JSON.
//   - Normal JSON instead of ActivityPub JSON.
//   - Statusable instead of Accountable.
//   - Accountable instead of Statusable.
//   - etc.
func IsWrongType(err error) bool {
	_, ok := errors.Value(err, wrongTypeKey).(struct{})
	return ok
}

// SetWrongType will wrap the given error to store a "wrong type" flag,
// returning wrapped error. See IsWrongType() for example use-cases.
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
// returning wrapped error. See IsNotFound() for example use-cases.
func SetNotFound(err error) error {
	return errors.WithValue(err, notFoundKey, struct{}{})
}

// IsSMTP checks error for a stored "smtp" flag. For
// example an error from outgoing SMTP email attempt.
func IsSMTP(err error) bool {
	_, ok := errors.Value(err, smtpKey).(struct{})
	return ok
}

// SetSMTP will wrap the given error to store an "smtp" flag,
// returning wrapped error. See IsSMTP() for example use-cases.
func SetSMTP(err error) error {
	return errors.WithValue(err, smtpKey, struct{}{})
}

// IsMalformed checks error for a stored "malformed" flag. For
// example an error from an incoming ActivityStreams type conversion.
func IsMalformed(err error) bool {
	_, ok := errors.Value(err, malformedKey).(struct{})
	return ok
}

// SetMalformed will wrap the given error to store a "malformed" flag,
// returning wrapped error. See IsMalformed() for example use-cases.
func SetMalformed(err error) error {
	return errors.WithValue(err, malformedKey, struct{}{})
}

// IsNotRelevant checks error for a stored "notRelevant" flag.
// This error is used when determining whether or not to store
// + process an incoming AP message.
func IsNotRelevant(err error) bool {
	_, ok := errors.Value(err, notRelevantKey).(struct{})
	return ok
}

// SetNotRelevant will wrap the given error to store a "notRelevant" flag,
// returning wrapped error. See IsNotRelevant() for example use-cases.
func SetNotRelevant(err error) error {
	return errors.WithValue(err, notRelevantKey, struct{}{})
}

// IsSpam checks error for a stored "spam" flag. This error is used when
// determining whether or not to store + process an incoming AP message.
func IsSpam(err error) bool {
	_, ok := errors.Value(err, spamKey).(struct{})
	return ok
}

// SetSpam will wrap the given error to store a "spam" flag,
// returning wrapped error. See IsSpam() for example use-cases.
func SetSpam(err error) error {
	return errors.WithValue(err, spamKey, struct{}{})
}
