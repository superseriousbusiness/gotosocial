/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package user_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/processing/user"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type UserStandardTestSuite struct {
	suite.Suite
	config *config.Config
	db     db.DB

	testUsers map[string]*gtsmodel.User

	user user.Processor
}

func (suite *UserStandardTestSuite) SetupTest() {
	testrig.InitTestLog()
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.testUsers = testrig.NewTestUsers()
	suite.user = user.New(suite.db, suite.config)

	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *UserStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}
