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

package admin_test

import (
	"context"
	"testing"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type EmojiTestSuite struct {
	AdminStandardTestSuite
}

func (suite *EmojiTestSuite) TestUpdateEmojiCategory() {
	ctx := context.Background()
	testEmoji := new(gtsmodel.Emoji)
	*testEmoji = *suite.testEmojis["rainbow"]

	// Toggle the emoji category around.
	for _, categoryName := range []string{
		"",
		"newCategory",
		"newCategory",
		"newCategory2",
		"",
		"reactions",
		"",
		"",
	} {
		emoji, err := suite.adminProcessor.EmojiUpdate(ctx,
			testEmoji.ID,
			&apimodel.EmojiUpdateRequest{
				Type:         apimodel.EmojiUpdateModify,
				CategoryName: util.Ptr(categoryName),
			},
		)
		if err != nil {
			suite.FailNow(err.Error())
		}

		suite.Equal(categoryName, emoji.Category)
	}
}

func TestEmojiTestSuite(t *testing.T) {
	suite.Run(t, new(EmojiTestSuite))
}
