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
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type StatusEditTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *StatusEditTestSuite) TestGetStatusEditBy() {
	t := suite.T()

	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Sentinel error to mark avoiding a test case.
	sentinelErr := errors.New("sentinel")

	// isEqual checks if 2 status edit models are equal.
	isEqual := func(e1, e2 gtsmodel.StatusEdit) bool {

		// Clear populated sub-models.
		e1.Attachments = nil
		e2.Attachments = nil

		// Clear database-set fields.
		e1.CreatedAt = time.Time{}
		e2.CreatedAt = time.Time{}

		return reflect.DeepEqual(e1, e2)
	}

	for _, edit := range suite.testStatusEdits {
		for lookup, dbfunc := range map[string]func() (*gtsmodel.StatusEdit, error){
			"id": func() (*gtsmodel.StatusEdit, error) {
				return suite.db.GetStatusEditByID(ctx, edit.ID)
			},
		} {
			// Clear database caches.
			suite.state.Caches.Init()

			t.Logf("checking database lookup %q", lookup)

			// Perform database function.
			checkEdit, err := dbfunc()
			if err != nil {
				if err == sentinelErr {
					continue
				}

				t.Errorf("error encountered for database lookup %q: %v", lookup, err)
				continue
			}

			// Check received account data.
			if !isEqual(*checkEdit, *edit) {
				t.Errorf("edit does not contain expected data: %+v", checkEdit)
				continue
			}
		}
	}
}

func (suite *StatusEditTestSuite) TestDeleteStatusEdits() {

}

func TestStatusEditTestSuite(t *testing.T) {
	suite.Run(t, new(StatusEditTestSuite))
}
