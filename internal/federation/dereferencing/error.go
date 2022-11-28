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

// Error aliases common error types that occur when doing dereferencing.
type Error error

// ErrDB denotes that a proper error has occurred when doing
// a database call, as opposed to a simple db.ErrNoEntries.
type ErrDB struct {
	wrapped error
}

func (err *ErrDB) Error() string {
	return err.wrapped.Error()
}

func newErrDB(err error) Error {
	return &ErrDB{
		wrapped: fmt.Errorf("database error during dereferencing: %w", err),
	}
}

// ErrNotRetrievable denotes that an item could not be dereferenced
// with the given parameters.
type ErrNotRetrievable struct {
	wrapped error
}

func (err *ErrNotRetrievable) Error() string {
	return err.wrapped.Error()
}

func newErrNotRetrievable(err error) Error {
	return &ErrNotRetrievable{
		wrapped: fmt.Errorf("item could not be retrieved: %w", err),
	}
}

// ErrBadRequest denotes that insufficient or improperly formed parameters
// were passed into one of the dereference functions.
type ErrBadRequest struct {
	wrapped error
}

func (err *ErrBadRequest) Error() string {
	return err.wrapped.Error()
}

func newErrBadRequest(err error) Error {
	return &ErrBadRequest{
		wrapped: fmt.Errorf("bad request: %w", err),
	}
}

// ErrTransportError indicates that something unforeseen went wrong creating
// a transport, or while making an http call to a remote resource with a transport.
type ErrTransportError struct {
	wrapped error
}

func (err *ErrTransportError) Error() string {
	return err.wrapped.Error()
}

func newErrTransportError(err error) Error {
	return &ErrTransportError{
		wrapped: fmt.Errorf("transport error: %w", err),
	}
}

// ErrWrongType indicates that an unexpected type was returned from a remote call;
// for example, we were served a Person when we were looking for a statusable.
type ErrWrongType struct {
	wrapped error
}

func (err *ErrWrongType) Error() string {
	return err.wrapped.Error()
}

func newErrWrongType(err error) Error {
	return &ErrWrongType{
		wrapped: fmt.Errorf("wrong type: %w", err),
	}
}

// ErrOther denotes some other kind of weird error, perhaps from a malformed json
// or some other weird crapola.
type ErrOther struct {
	wrapped error
}

func (err *ErrOther) Error() string {
	return err.wrapped.Error()
}

func newErrOther(err error) Error {
	return &ErrOther{
		wrapped: fmt.Errorf("other error: %w", err),
	}
}

func wrapDerefError(derefErr error, fluff string) Error {
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
