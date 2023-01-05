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

package util_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type TimeSuite struct {
	suite.Suite
}

func (suite *TimeSuite) TestISO8601Format1() {
	testTime := testrig.TimeMustParse("2022-05-17T13:10:59Z")
	testTimeString := util.FormatISO8601(testTime)
	suite.Equal("2022-05-17T13:10:59.000Z", testTimeString)
}

func (suite *TimeSuite) TestISO8601Format2() {
	testTime := testrig.TimeMustParse("2022-05-09T07:34:35+02:00")
	testTimeString := util.FormatISO8601(testTime)
	suite.Equal("2022-05-09T05:34:35.000Z", testTimeString)
}

func (suite *TimeSuite) TestISO8601Format3() {
	testTime := testrig.TimeMustParse("2021-10-04T10:52:36+02:00")
	testTimeString := util.FormatISO8601(testTime)
	suite.Equal("2021-10-04T08:52:36.000Z", testTimeString)
}

func TestTimeSuite(t *testing.T) {
	suite.Run(t, &TimeSuite{})
}
