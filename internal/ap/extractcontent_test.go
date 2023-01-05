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

type ExtractContentTestSuite struct {
	ExtractTestSuite
}

func (suite *ExtractContentTestSuite) TestExtractContent1() {
	note := suite.noteWithMentions1

	content := ap.ExtractContent(note)

	suite.Equal("hey @f0x and @dumpsterqueer", content)
}

func TestExtractContentTestSuite(t *testing.T) {
	suite.Run(t, &ExtractContentTestSuite{})
}
