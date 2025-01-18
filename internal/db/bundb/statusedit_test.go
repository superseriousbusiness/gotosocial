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
	"slices"
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
			if !areEditsEqual(edit, checkEdit) {
				t.Errorf("edit does not contain expected data: %+v", checkEdit)
				continue
			}
		}
	}
}

func (suite *StatusEditTestSuite) TestGetStatusEditsByIDs() {
	t := suite.T()

	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// editsByStatus returns all test edits by the given status with ID.
	editsByStatus := func(status *gtsmodel.Status) []*gtsmodel.StatusEdit {
		var edits []*gtsmodel.StatusEdit
		for _, edit := range suite.testStatusEdits {
			if edit.StatusID == status.ID {
				edits = append(edits, edit)
			}
		}
		return edits
	}

	for _, status := range suite.testStatuses {
		// Get test status edit models
		// that should be found for status.
		check := editsByStatus(status)

		// Fetch edits for the slice of IDs attached to status from database.
		edits, err := suite.state.DB.GetStatusEditsByIDs(ctx, status.EditIDs)
		suite.NoError(err)

		// Ensure both slices
		// sorted the same.
		sortEdits(check)
		sortEdits(edits)

		// Check whether slices of status edits match.
		if !slices.EqualFunc(check, edits, areEditsEqual) {
			t.Error("status edit slices do not match")
		}
	}
}

func (suite *StatusEditTestSuite) TestDeleteStatusEdits() {
	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	for _, status := range suite.testStatuses {
		// Delete all edits for status with given IDs from database.
		err := suite.state.DB.DeleteStatusEdits(ctx, status.EditIDs)
		suite.NoError(err)

		// Now attempt to fetch these edits from database, should be empty.
		edits, err := suite.state.DB.GetStatusEditsByIDs(ctx, status.EditIDs)
		suite.NoError(err)
		suite.Empty(edits)
	}
}

func TestStatusEditTestSuite(t *testing.T) {
	suite.Run(t, new(StatusEditTestSuite))
}

func areEditsEqual(e1, e2 *gtsmodel.StatusEdit) bool {
	// Clone the 1st status edit.
	e1Copy := new(gtsmodel.StatusEdit)
	*e1Copy = *e1
	e1 = e1Copy

	// Clone the 2nd status edit.
	e2Copy := new(gtsmodel.StatusEdit)
	*e2Copy = *e2
	e2 = e2Copy

	// Clear populated sub-models.
	e1.Attachments = nil
	e2.Attachments = nil

	// Clear database-set fields.
	e1.CreatedAt = time.Time{}
	e2.CreatedAt = time.Time{}

	return reflect.DeepEqual(*e1, *e2)
}

func sortEdits(edits []*gtsmodel.StatusEdit) {
	slices.SortFunc(edits, func(a, b *gtsmodel.StatusEdit) int {
		if a.CreatedAt.Before(b.CreatedAt) {
			return +1
		} else if b.CreatedAt.Before(a.CreatedAt) {
			return -1
		}
		return 0
	})
}
