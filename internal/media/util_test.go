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

package media

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type MediaUtilTestSuite struct {
	suite.Suite
	log *logrus.Logger
}

/*
	TEST INFRASTRUCTURE
*/

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *MediaUtilTestSuite) SetupSuite() {
	// some of our subsequent entities need a log so create this here
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	suite.log = log
}

func (suite *MediaUtilTestSuite) TearDownSuite() {

}

// SetupTest creates a db connection and creates necessary tables before each test
func (suite *MediaUtilTestSuite) SetupTest() {

}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *MediaUtilTestSuite) TearDownTest() {

}

/*
	ACTUAL TESTS
*/

func (suite *MediaUtilTestSuite) TestParseContentType() {
	f, err := os.Open("./test/test-jpeg.jpg")
	if err != nil {
		suite.FailNow(err.Error())
	}
	ct, err := parseContentType(f)
	suite.log.Debug(ct)
}

func TestMediaUtilTestSuite(t *testing.T) {
	suite.Run(t, new(MediaUtilTestSuite))
}
