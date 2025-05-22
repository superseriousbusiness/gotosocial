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

package bundb_test

import (
	"errors"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type ErrorsTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *ErrorsTestSuite) TestErrorAlreadyExists() {
	type testType struct {
		follow   *gtsmodel.Follow
		expected error
	}

	var (
		ctx           = suite.T().Context()
		initialFollow = &gtsmodel.Follow{
			ID:              "01HD11D8JH5V64GJRFDA7VFNDX",
			URI:             "https://example.org/unique_uri",
			AccountID:       "01HD11E9Y02HVXZ4YB20SG4QR9",
			TargetAccountID: "01HD11EPJWM919DRP9JCKDZBX6",
		}
		editFollow = func(f func(*gtsmodel.Follow)) *gtsmodel.Follow {
			edited := new(gtsmodel.Follow)
			*edited = *initialFollow

			f(edited)
			return edited
		}
	)

	// Put the initial follow so we have
	// a constraint in place to fail on.
	if err := suite.db.PutFollow(ctx, initialFollow); err != nil {
		suite.FailNow(err.Error())
	}

	for i, test := range []testType{
		{
			// Try to put the initial follow in again.
			follow:   initialFollow,
			expected: db.ErrAlreadyExists,
		},
		{
			// Different ID but same URI,
			// should fail on URI constraint.
			follow: editFollow(func(f *gtsmodel.Follow) {
				f.ID = "01HD12ASEEBZ7MRY8AK51RJ20Y"
			}),
			expected: db.ErrAlreadyExists,
		},
		{
			// Different URI but same ID,
			// should fail on ID constraint.
			follow: editFollow(func(f *gtsmodel.Follow) {
				f.URI = "https://example.org/new_uri"
			}),
			expected: db.ErrAlreadyExists,
		},
		{
			// Different ID and URI, but same account and
			// target. Should fail on unique constraint.
			follow: editFollow(func(f *gtsmodel.Follow) {
				f.ID = "01HD12ASEEBZ7MRY8AK51RJ20Y"
				f.URI = "https://example.org/new_uri"
			}),
			expected: db.ErrAlreadyExists,
		},
		{
			// Put a follow with different ID, URI,
			// and AccountID. This should work fine.
			follow: editFollow(func(f *gtsmodel.Follow) {
				f.ID = "01HD12ASEEBZ7MRY8AK51RJ20Y"
				f.URI = "https://example.org/new_uri"
				f.AccountID = "01HD12DTPDR8A96S5K5M7SJAF3"
			}),
			expected: nil,
		},
	} {
		err := suite.db.PutFollow(ctx, test.follow)
		if !errors.Is(err, test.expected) {
			suite.Fail("", "test number %d expected %v, got %v", i+1, test.expected, err)
		}
	}
}

func TestErrorsTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorsTestSuite))
}
