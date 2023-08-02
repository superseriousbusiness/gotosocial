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
	"testing"

	"github.com/superseriousbusiness/gotosocial/internal/db"
)

func TestMultiError(t *testing.T) {
	errs := MultiError{
		e: []error{
			db.ErrNoEntries,
			errors.New("oopsie woopsie we did a fucky wucky etc"),
		},
	}
	errs.Appendf("appended + wrapped error: %w", db.ErrAlreadyExists)

	err := errs.Combine()

	if !errors.Is(err, db.ErrNoEntries) {
		t.Error("should be db.ErrNoEntries")
	}

	if !errors.Is(err, db.ErrAlreadyExists) {
		t.Error("should be db.ErrAlreadyExists")
	}

	if errors.Is(err, db.ErrBusyTimeout) {
		t.Error("should not be db.ErrBusyTimeout")
	}

	errString := err.Error()
	expected := `sql: no rows in result set
oopsie woopsie we did a fucky wucky etc
appended + wrapped error: already exists`
	if errString != expected {
		t.Errorf("errString '%s' should be '%s'", errString, expected)
	}
}

func TestMultiErrorEmpty(t *testing.T) {
	err := new(MultiError).Combine()
	if err != nil {
		t.Errorf("should be nil")
	}
}
