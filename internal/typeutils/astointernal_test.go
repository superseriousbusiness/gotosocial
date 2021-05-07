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

package typeutils_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ASToInternalTestSuite struct {
	ConverterStandardTestSuite
}

func (suite *ASToInternalTestSuite) SetupSuite() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.accounts = testrig.NewTestAccounts()
	suite.people = testrig.NewTestFediPeople()
	suite.typeconverter = typeutils.NewConverter(suite.config, suite.db)
}

func (suite *ASToInternalTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db)
}

func (suite *ASToInternalTestSuite) TestASRepresentationToAccount() {

	testPerson := suite.people["new_person_1"]

	acct, err := suite.typeconverter.ASRepresentationToAccount(testPerson)
	assert.NoError(suite.T(), err)

	fmt.Printf("%+v", acct)
	// TODO: write assertions here, rn we're just eyeballing the output

}

func (suite *ASToInternalTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func TestASToInternalTestSuite(t *testing.T) {
	suite.Run(t, new(ASToInternalTestSuite))
}
