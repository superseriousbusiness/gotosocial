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
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type ExtractVisibilityTestSuite struct {
	APTestSuite
}

func (suite *ExtractVisibilityTestSuite) TestExtractVisibilityPublic() {
	a := suite.addressable1
	visibility, err := ap.ExtractVisibility(a, "http://localhost:8080/users/the_mighty_zork/followers")
	suite.NoError(err)
	suite.Equal(visibility, gtsmodel.VisibilityPublic)
}

func (suite *ExtractVisibilityTestSuite) TestExtractVisibilityUnlocked() {
	a := suite.addressable2
	visibility, err := ap.ExtractVisibility(a, "http://localhost:8080/users/the_mighty_zork/followers")
	suite.NoError(err)
	suite.Equal(visibility, gtsmodel.VisibilityUnlocked)
}

func (suite *ExtractVisibilityTestSuite) TestExtractVisibilityFollowersOnly() {
	a := suite.addressable3
	visibility, err := ap.ExtractVisibility(a, "http://localhost:8080/users/the_mighty_zork/followers")
	suite.NoError(err)
	suite.Equal(visibility, gtsmodel.VisibilityFollowersOnly)
}

func (suite *ExtractVisibilityTestSuite) TestExtractVisibilityFollowersOnlyAnnounce() {
	// https://codeberg.org/superseriousbusiness/gotosocial/issues/267
	a := suite.addressable4
	visibility, err := ap.ExtractVisibility(a, "https://example.org/users/someone/followers")
	suite.NoError(err)
	suite.Equal(visibility, gtsmodel.VisibilityFollowersOnly)
}

func (suite *ExtractVisibilityTestSuite) TestExtractVisibilityDirect() {
	a := suite.addressable5
	visibility, err := ap.ExtractVisibility(a, "http://localhost:8080/users/the_mighty_zork/followers")
	suite.NoError(err)
	suite.Equal(visibility, gtsmodel.VisibilityDirect)
}

func TestExtractVisibilityTestSuite(t *testing.T) {
	suite.Run(t, &ExtractVisibilityTestSuite{})
}
