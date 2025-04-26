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

package gtserror_test

import (
	"errors"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

func TestMultiError(t *testing.T) {
	errs := gtserror.MultiError([]error{
		db.ErrNoEntries,
		errors.New("oopsie woopsie we did a fucky wucky etc"),
	})

	errs.Appendf("appended + wrapped error: %w", db.ErrAlreadyExists)

	err := errs.Combine()

	if !errors.Is(err, db.ErrNoEntries) {
		t.Error("should be db.ErrNoEntries")
	}

	if !errors.Is(err, db.ErrAlreadyExists) {
		t.Error("should be db.ErrAlreadyExists")
	}

	errString := err.Error()
	expected := `sql: no rows in result set
oopsie woopsie we did a fucky wucky etc
TestMultiError: appended + wrapped error: already exists`
	if errString != expected {
		t.Errorf("errString '%s' should be '%s'", errString, expected)
	}
}

func TestMultiErrorEmpty(t *testing.T) {
	err := new(gtserror.MultiError).Combine()
	if err != nil {
		t.Errorf("should be nil")
	}
}
