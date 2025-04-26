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

package account_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/processing/account"
	"github.com/stretchr/testify/suite"
)

type ThemesTestSuite struct {
	AccountStandardTestSuite
}

func (suite *ThemesTestSuite) TestPopulateThemes() {
	config.SetWebAssetBaseDir("../../../web/assets")

	themes := account.PopulateThemes()
	if themes == nil {
		suite.FailNow("themes was nil")
	}

	suite.NotEmpty(themes.SortedByTitle)
	theme := themes.ByFileName["blurple-light.css"]
	if theme == nil {
		suite.FailNow("theme was nil")
	}
	suite.Equal("Blurple (light)", theme.Title)
	suite.Equal("Official light blurple theme", theme.Description)
	suite.Equal("blurple-light.css", theme.FileName)
}

func TestThemesTestSuite(t *testing.T) {
	suite.Run(t, new(ThemesTestSuite))
}
