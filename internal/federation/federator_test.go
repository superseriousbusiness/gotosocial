/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"codeberg.org/gruf/go-store/kv"
	"github.com/stretchr/testify/suite"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FederatorStandardTestSuite struct {
	suite.Suite
	db            db.DB
	storage       *kv.KVStore
	typeConverter typeutils.TypeConverter
	accounts      map[string]*gtsmodel.Account
	activities    map[string]testrig.ActivityWithSignature
}

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *FederatorStandardTestSuite) SetupSuite() {
	// setup standard items
	suite.storage = testrig.NewTestStorage()
	suite.typeConverter = testrig.NewTestTypeConverter(suite.db)
	suite.accounts = testrig.NewTestAccounts()
}

func (suite *FederatorStandardTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()
	suite.db = testrig.NewTestDB()
	suite.activities = testrig.NewTestActivities(suite.accounts)
	testrig.StandardDBSetup(suite.db, suite.accounts)
}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *FederatorStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}
