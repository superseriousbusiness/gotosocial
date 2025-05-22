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

package processing_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type PreferencesTestSuite struct {
	ProcessingStandardTestSuite
}

func (suite *PreferencesTestSuite) TestPreferencesGet() {
	ctx := suite.T().Context()
	tests := []struct {
		act   *gtsmodel.Account
		prefs *model.Preferences
	}{
		{
			act: suite.testAccounts["local_account_1"],
			prefs: &model.Preferences{
				PostingDefaultVisibility: "public",
				PostingDefaultSensitive:  false,
				PostingDefaultLanguage:   "en",
				ReadingExpandMedia:       "default",
				ReadingExpandSpoilers:    false,
				ReadingAutoPlayGifs:      false,
			},
		},
		{
			act: suite.testAccounts["local_account_2"],
			prefs: &model.Preferences{
				PostingDefaultVisibility: "private",
				PostingDefaultSensitive:  true,
				PostingDefaultLanguage:   "fr",
				ReadingExpandMedia:       "default",
				ReadingExpandSpoilers:    false,
				ReadingAutoPlayGifs:      false,
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.act.ID, func() {
			prefs, err := suite.processor.PreferencesGet(ctx, tt.act.ID)
			suite.NoError(err)
			suite.Equal(tt.prefs, prefs)
		})
	}
}

func TestPreferencesTestSuite(t *testing.T) {
	suite.Run(t, &PreferencesTestSuite{})
}
