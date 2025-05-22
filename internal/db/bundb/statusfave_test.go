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

type StatusFaveTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *StatusFaveTestSuite) TestGetStatusFaves() {
	testStatus := suite.testStatuses["admin_account_status_1"]

	faves, err := suite.db.GetStatusFaves(suite.T().Context(), testStatus.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotEmpty(faves)
	for _, fave := range faves {
		suite.NotNil(fave.Account)
		suite.NotNil(fave.TargetAccount)
		suite.NotNil(fave.Status)
	}
}

func (suite *StatusFaveTestSuite) TestGetStatusFavesNone() {
	testStatus := suite.testStatuses["admin_account_status_4"]

	faves, err := suite.db.GetStatusFaves(suite.T().Context(), testStatus.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Empty(faves)
}

func (suite *StatusFaveTestSuite) TestGetStatusFaveByAccountID() {
	testAccount := suite.testAccounts["local_account_1"]
	testStatus := suite.testStatuses["admin_account_status_1"]

	fave, err := suite.db.GetStatusFave(suite.T().Context(), testAccount.ID, testStatus.ID)
	suite.NoError(err)
	suite.NotNil(fave)
}

func (suite *StatusFaveTestSuite) TestDeleteStatusFavesOriginatingFromAccount() {
	testAccount := suite.testAccounts["local_account_1"]

	if err := suite.db.DeleteStatusFaves(suite.T().Context(), "", testAccount.ID); err != nil {
		suite.FailNow(err.Error())
	}

	faves := []*gtsmodel.StatusFave{}
	if err := suite.db.GetAll(suite.T().Context(), &faves); err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	for _, b := range faves {
		if b.AccountID == testAccount.ID {
			suite.FailNowf("", "no StatusFaves with account id %s should remain", testAccount.ID)
		}
	}
}

func (suite *StatusFaveTestSuite) TestDeleteStatusFavesTargetingAccount() {
	testAccount := suite.testAccounts["local_account_1"]

	if err := suite.db.DeleteStatusFaves(suite.T().Context(), testAccount.ID, ""); err != nil {
		suite.FailNow(err.Error())
	}

	faves := []*gtsmodel.StatusFave{}
	if err := suite.db.GetAll(suite.T().Context(), &faves); err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	for _, b := range faves {
		if b.TargetAccountID == testAccount.ID {
			suite.FailNowf("", "no StatusFaves with target account id %s should remain", testAccount.ID)
		}
	}
}

func (suite *StatusFaveTestSuite) TestDeleteStatusFavesTargetingStatus() {
	testStatus := suite.testStatuses["local_account_1_status_1"]

	if err := suite.db.DeleteStatusFavesForStatus(suite.T().Context(), testStatus.ID); err != nil {
		suite.FailNow(err.Error())
	}

	faves := []*gtsmodel.StatusFave{}
	if err := suite.db.GetAll(suite.T().Context(), &faves); err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	for _, b := range faves {
		if b.StatusID == testStatus.ID {
			suite.FailNowf("", "no StatusFaves with status id %s should remain", testStatus.ID)
		}
	}
}

func (suite *StatusFaveTestSuite) TestDeleteStatusFave() {
	testFave := suite.testFaves["local_account_1_admin_account_status_1"]
	ctx := suite.T().Context()

	if err := suite.db.DeleteStatusFaveByID(ctx, testFave.ID); err != nil {
		suite.FailNow(err.Error())
	}

	fave, err := suite.db.GetStatusFaveByID(ctx, testFave.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(fave)
}

func (suite *StatusFaveTestSuite) TestDeleteStatusFaveNonExisting() {
	err := suite.db.DeleteStatusFaveByID(suite.T().Context(), "01GVAV715K6Y2SG9ZKS9ZA8G7G")
	suite.NoError(err)
}

func TestStatusFaveTestSuite(t *testing.T) {
	suite.Run(t, new(StatusFaveTestSuite))
}
