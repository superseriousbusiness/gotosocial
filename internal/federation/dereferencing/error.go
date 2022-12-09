/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package dereferencing

import (
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

// ErrDB denotes that a proper error has occurred when doing
// a database call, as opposed to a simple db.ErrNoEntries.
type ErrDB struct {
	wrapped error
}

func (err *ErrDB) Error() string {
	return fmt.Sprintf("database error during dereferencing: %v", err.wrapped)
}

func newErrDB(err error) error {
	return &ErrDB{wrapped: err}
}

// ErrNotRetrievable denotes that an item could not be dereferenced
// with the given parameters.
type ErrNotRetrievable struct {
	wrapped error
}

func (err *ErrNotRetrievable) Error() string {
	return fmt.Sprintf("item could not be retrieved: %v", err.wrapped)
}

func newErrNotRetrievable(err error) error {
	return &ErrNotRetrievable{wrapped: err}
}

// ErrBadRequest denotes that insufficient or improperly formed parameters
// were passed into one of the dereference functions.
type ErrBadRequest struct {
	wrapped error
}

func (err *ErrBadRequest) Error() string {
	return fmt.Sprintf("bad request: %v", err.wrapped)
}

func newErrBadRequest(err error) error {
	return &ErrBadRequest{wrapped: err}
}

// ErrTransportError indicates that something unforeseen went wrong creating
// a transport, or while making an http call to a remote resource with a transport.
type ErrTransportError struct {
	wrapped error
}

func (err *ErrTransportError) Error() string {
	return fmt.Sprintf("transport error: %v", err.wrapped)
}

func newErrTransportError(err error) error {
	return &ErrTransportError{wrapped: err}
}

// ErrWrongType indicates that an unexpected type was returned from a remote call;
// for example, we were served a Person when we were looking for a statusable.
type ErrWrongType struct {
	wrapped error
}

func (err *ErrWrongType) Error() string {
	return fmt.Sprintf("wrong received type: %v", err.wrapped)
}

func newErrWrongType(err error) error {
	return &ErrWrongType{wrapped: err}
}

// ErrOther denotes some other kind of weird error, perhaps from a malformed json
// or some other weird crapola.
type ErrOther struct {
	wrapped error
}

func (err *ErrOther) Error() string {
	return fmt.Sprintf("unexpected error: %v", err.wrapped)
}

func newErrOther(err error) error {
	return &ErrOther{wrapped: err}
}

func wrapDerefError(derefErr error, fluff string) error {
	var (
		err          error
		errWrongType *ErrWrongType
	)

	if fluff != "" {
		err = fmt.Errorf("%s: %w", fluff, derefErr)
	}

	switch {
	case errors.Is(derefErr, transport.ErrGone):
		err = newErrNotRetrievable(err)
	case errors.As(derefErr, &errWrongType):
		err = newErrWrongType(err)
	default:
		err = newErrTransportError(err)
	}

	return err
}
