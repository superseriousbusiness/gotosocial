/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package federation_test

import (
	"github.com/stretchr/testify/suite"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FederatorStandardTestSuite struct {
	suite.Suite
	db             db.DB
	storage        *storage.Driver
	tc             typeutils.TypeConverter
	testAccounts   map[string]*gtsmodel.Account
	testStatuses   map[string]*gtsmodel.Status
	testActivities map[string]testrig.ActivityWithSignature
	testTombstones map[string]*gtsmodel.Tombstone
}

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *FederatorStandardTestSuite) SetupSuite() {
	// setup standard items
	suite.storage = testrig.NewInMemoryStorage()
	suite.tc = testrig.NewTestTypeConverter(suite.db)
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testTombstones = testrig.NewTestTombstones()
}

func (suite *FederatorStandardTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()
	suite.db = testrig.NewTestDB()
	suite.testActivities = testrig.NewTestActivities(suite.testAccounts)
	testrig.StandardDBSetup(suite.db, suite.testAccounts)
}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *FederatorStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}
