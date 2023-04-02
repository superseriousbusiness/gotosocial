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

package ap

import "fmt"

// ErrWrongType indicates that we tried to resolve a type into
// an interface that it's not compatible with, eg a Person into
// a Statusable.
type ErrWrongType struct {
	wrapped error
}

func (err *ErrWrongType) Error() string {
	return fmt.Sprintf("wrong received type: %v", err.wrapped)
}

func newErrWrongType(err error) error {
	return &ErrWrongType{wrapped: err}
}
