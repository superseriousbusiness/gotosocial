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
)

type NamestringSuite struct {
	suite.Suite
}

func (suite *NamestringSuite) TestExtractWebfingerParts1() {
	webfinger := "acct:stonerkitty.monster@stonerkitty.monster"
	username, host, err := util.ExtractWebfingerParts(webfinger)
	suite.NoError(err)

	suite.Equal("stonerkitty.monster", username)
	suite.Equal("stonerkitty.monster", host)
}

func (suite *NamestringSuite) TestExtractWebfingerParts2() {
	webfinger := "@stonerkitty.monster@stonerkitty.monster"
	username, host, err := util.ExtractWebfingerParts(webfinger)
	suite.NoError(err)

	suite.Equal("stonerkitty.monster", username)
	suite.Equal("stonerkitty.monster", host)
}

func (suite *NamestringSuite) TestExtractWebfingerParts3() {
	webfinger := "acct:someone@somewhere"
	username, host, err := util.ExtractWebfingerParts(webfinger)
	suite.NoError(err)

	suite.Equal("someone", username)
	suite.Equal("somewhere", host)
}

func (suite *NamestringSuite) TestExtractWebfingerParts4() {
	webfinger := "@stoner-kitty.monster@stonerkitty.monster"
	username, host, err := util.ExtractWebfingerParts(webfinger)
	suite.NoError(err)

	suite.Equal("stoner-kitty.monster", username)
	suite.Equal("stonerkitty.monster", host)
}

func (suite *NamestringSuite) TestExtractWebfingerParts5() {
	webfinger := "@stonerkitty.monster"
	username, host, err := util.ExtractWebfingerParts(webfinger)
	suite.NoError(err)

	suite.Equal("stonerkitty.monster", username)
	suite.Empty(host)
}

func (suite *NamestringSuite) TestExtractWebfingerParts6() {
	webfinger := "@@stonerkitty.monster"
	_, _, err := util.ExtractWebfingerParts(webfinger)
	suite.EqualError(err, "couldn't match mention @@stonerkitty.monster")
}

func (suite *NamestringSuite) TestExtractNamestringParts1() {
	namestring := "@stonerkitty.monster@stonerkitty.monster"
	username, host, err := util.ExtractNamestringParts(namestring)
	suite.NoError(err)

	suite.Equal("stonerkitty.monster", username)
	suite.Equal("stonerkitty.monster", host)
}

func (suite *NamestringSuite) TestExtractNamestringParts2() {
	namestring := "@stonerkitty.monster"
	username, host, err := util.ExtractNamestringParts(namestring)
	suite.NoError(err)

	suite.Equal("stonerkitty.monster", username)
	suite.Empty(host)
}

func (suite *NamestringSuite) TestExtractNamestringParts3() {
	namestring := "@someone@somewhere"
	username, host, err := util.ExtractWebfingerParts(namestring)
	suite.NoError(err)

	suite.Equal("someone", username)
	suite.Equal("somewhere", host)
}

func (suite *NamestringSuite) TestExtractNamestringParts4() {
	namestring := ""
	_, _, err := util.ExtractNamestringParts(namestring)
	suite.EqualError(err, "couldn't match mention ")
}

func TestNamestringSuite(t *testing.T) {
	suite.Run(t, &NamestringSuite{})
}
