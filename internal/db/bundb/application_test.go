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

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type ApplicationTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *ApplicationTestSuite) TestGetApplicationBy() {
	t := suite.T()

	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Sentinel error to mark avoiding a test case.
	sentinelErr := errors.New("sentinel")

	// isEqual checks if 2 application models are equal.
	isEqual := func(a1, a2 gtsmodel.Application) bool {
		return reflect.DeepEqual(a1, a2)
	}

	for _, app := range suite.testApplications {
		for lookup, dbfunc := range map[string]func() (*gtsmodel.Application, error){
			"id": func() (*gtsmodel.Application, error) {
				return suite.db.GetApplicationByID(ctx, app.ID)
			},

			"client_id": func() (*gtsmodel.Application, error) {
				return suite.db.GetApplicationByClientID(ctx, app.ClientID)
			},
		} {
			// Clear database caches.
			suite.state.Caches.Init()

			t.Logf("checking database lookup %q", lookup)

			// Perform database function.
			checkApp, err := dbfunc()
			if err != nil {
				if err == sentinelErr {
					continue
				}

				t.Errorf("error encountered for database lookup %q: %v", lookup, err)
				continue
			}

			// Check received application data.
			if !isEqual(*checkApp, *app) {
				t.Errorf("application does not contain expected data: %+v", checkApp)
				continue
			}
		}
	}
}

func (suite *ApplicationTestSuite) TestDeleteApplicationBy() {
	t := suite.T()

	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	for _, app := range suite.testApplications {
		for lookup, dbfunc := range map[string]func() error{
			"client_id": func() error {
				return suite.db.DeleteApplicationByClientID(ctx, app.ClientID)
			},
		} {
			// Clear database caches.
			suite.state.Caches.Init()

			t.Logf("checking database lookup %q", lookup)

			// Perform database function.
			err := dbfunc()
			if err != nil {
				t.Errorf("error encountered for database lookup %q: %v", lookup, err)
				continue
			}

			// Ensure this application has been deleted and cache cleared.
			if _, err := suite.db.GetApplicationByID(ctx, app.ID); err != db.ErrNoEntries {
				t.Errorf("application does not appear to have been deleted %q: %v", lookup, err)
				continue
			}
		}
	}
}

func (suite *ApplicationTestSuite) TestGetAllTokens() {
	tokens, err := suite.db.GetAllTokens(context.Background())
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotEmpty(tokens)
}

func TestApplicationTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}
