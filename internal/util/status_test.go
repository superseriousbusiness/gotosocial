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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StatusTestSuite struct {
	suite.Suite
}

func (suite *StatusTestSuite) TestDeriveMentionsOK() {
	statusText := `@dumpsterqueer@example.org testing testing

	is this thing on?

	@someone_else@testing.best-horse.com can you confirm? @hello@test.lgbt

	@thiswontwork though! @NORWILL@THIS.one!!

	here is a duplicate mention: @hello@test.lgbt
	`

	menchies := DeriveMentions(statusText)
	assert.Len(suite.T(), menchies, 3)
	assert.Equal(suite.T(), "@dumpsterqueer@example.org", menchies[0])
	assert.Equal(suite.T(), "@someone_else@testing.best-horse.com", menchies[1])
	assert.Equal(suite.T(), "@hello@test.lgbt", menchies[2])
}

func (suite *StatusTestSuite) TestDeriveMentionsEmpty() {
	statusText := ``
	menchies := DeriveMentions(statusText)
	assert.Len(suite.T(), menchies, 0)
}

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}
