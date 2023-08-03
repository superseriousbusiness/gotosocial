// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package normalize

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type NormalizationTestSuite struct {
	suite.Suite
}

func (suite *NormalizationTestSuite) TestNormalizeLanguage() {
	empty := ""
	notALanguage := "this isn't a language at all!"
	english := "en"
	// Should be all lowercase
	capitalEnglish := "EN"
	// Overlong, should be in ISO 639-1 format
	arabic3Letters := "ara"
	// Should be all lowercase
	mixedCapsEnglish := "eN"
	// Region should be capitalized
	englishUS := "en-us"
	dutch := "nl"
	german := "de"
	chinese := "zh"
	chineseSimplified := "zh-Hans"
	chineseTraditional := "zh-Hant"

	var actual string

	actual = Language(empty)
	suite.Equal(empty, actual)

	actual = Language(notALanguage)
	suite.Equal(empty, actual)

	actual = Language(english)
	suite.Equal(english, actual)

	actual = Language(capitalEnglish)
	suite.Equal(english, actual)

	actual = Language(arabic3Letters)
	suite.Equal("ar", actual)

	actual = Language(mixedCapsEnglish)
	suite.Equal(english, actual)

	actual = Language(englishUS)
	suite.Equal("en-US", actual)

	actual = Language(dutch)
	suite.Equal(dutch, actual)

	actual = Language(german)
	suite.Equal(german, actual)

	actual = Language(chinese)
	suite.Equal(chinese, actual)

	actual = Language(chineseSimplified)
	suite.Equal(chineseSimplified, actual)

	actual = Language(chineseTraditional)
	suite.Equal(chineseTraditional, actual)
}

func TestNormalizationTestSuite(t *testing.T) {
	suite.Run(t, new(NormalizationTestSuite))
}
