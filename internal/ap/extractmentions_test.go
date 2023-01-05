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

package ap_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
)

type ExtractMentionsTestSuite struct {
	ExtractTestSuite
}

func (suite *ExtractMentionsTestSuite) TestExtractMentions() {
	note := suite.noteWithMentions1

	mentions, err := ap.ExtractMentions(note)
	suite.NoError(err)
	suite.Len(mentions, 2)

	m1 := mentions[0]
	suite.Equal("@dumpsterqueer@superseriousbusiness.org", m1.NameString)
	suite.Equal("https://gts.superseriousbusiness.org/users/dumpsterqueer", m1.TargetAccountURI)

	m2 := mentions[1]
	suite.Equal("@f0x@superseriousbusiness.org", m2.NameString)
	suite.Equal("https://gts.superseriousbusiness.org/users/f0x", m2.TargetAccountURI)
}

func TestExtractMentionsTestSuite(t *testing.T) {
	suite.Run(t, &ExtractMentionsTestSuite{})
}
