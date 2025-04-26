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

package ap_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"github.com/stretchr/testify/suite"
)

type ExtractHashtagsTestSuite struct {
	APTestSuite
}

func (suite *ExtractHashtagsTestSuite) TestExtractHashtags1() {
	note := suite.noteWithHashtags1()

	hashtags, err := ap.ExtractHashtags(note)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if l := len(hashtags); l != 4 {
		suite.FailNow("", "expected 4 hashtags, got %d", l)
	}

	hashtagFediverse := hashtags[0]
	suite.Equal("fediverse", hashtagFediverse.Name)
	suite.Equal(true, *hashtagFediverse.Useable)
	suite.Equal(true, *hashtagFediverse.Listable)

	hashtagGoToSocial := hashtags[1]
	suite.Equal("gotosocial", hashtagGoToSocial.Name)
	suite.Equal(true, *hashtagGoToSocial.Useable)
	suite.Equal(true, *hashtagGoToSocial.Listable)

	hashtagGrüvy := hashtags[2]
	suite.Equal("grüvy", hashtagGrüvy.Name)
	suite.Equal(true, *hashtagGrüvy.Useable)
	suite.Equal(true, *hashtagGrüvy.Listable)

	hashtagAngle := hashtags[3]
	suite.Equal("각", hashtagAngle.Name)
	suite.Equal(true, *hashtagAngle.Useable)
	suite.Equal(true, *hashtagAngle.Listable)
}

func TestExtractHashtagsTestSuite(t *testing.T) {
	suite.Run(t, &ExtractHashtagsTestSuite{})
}
