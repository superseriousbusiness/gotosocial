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

//go:build noerrcaller

package gtserror

import (
	"errors"
	"fmt"
)

// Caller returns whether created errors will prepend calling function name.
const Caller = false

// newAt is the same as New() but allows specifying calldepth.
func newAt(_ int, msg string) error {
	return errors.New(msg)
}

// newfAt is the same as Newf() but allows specifying calldepth.
func newfAt(_ int, msgf string, args ...any) error {
	return fmt.Errorf(msgf, args...)
}
