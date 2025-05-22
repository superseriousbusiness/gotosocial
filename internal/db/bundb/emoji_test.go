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

package bundb_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type EmojiTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *EmojiTestSuite) TestGetUseableEmojis() {
	emojis, err := suite.db.GetUseableEmojis(suite.T().Context())

	suite.NoError(err)
	suite.Equal(1, len(emojis))
	suite.Equal("rainbow", emojis[0].Shortcode)
}

func (suite *EmojiTestSuite) TestDeleteEmojiByID() {
	testEmoji := suite.testEmojis["rainbow"]

	err := suite.db.DeleteEmojiByID(suite.T().Context(), testEmoji.ID)
	suite.NoError(err)

	dbEmoji, err := suite.db.GetEmojiByID(suite.T().Context(), testEmoji.ID)
	suite.Nil(dbEmoji)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func (suite *EmojiTestSuite) TestGetEmojiByStaticURL() {
	emoji, err := suite.db.GetEmojiByStaticURL(suite.T().Context(), "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png")
	suite.NoError(err)
	suite.NotNil(emoji)
	suite.Equal("rainbow", emoji.Shortcode)
	suite.NotNil(emoji.Category)
	suite.Equal("reactions", emoji.Category.Name)
}

func (suite *EmojiTestSuite) TestGetAllEmojis() {
	emojis, err := suite.db.GetEmojisBy(suite.T().Context(), db.EmojiAllDomains, true, true, "", "", "", 0)

	suite.NoError(err)
	suite.Equal(2, len(emojis))
	suite.Equal("rainbow", emojis[0].Shortcode)
	suite.Equal("yell", emojis[1].Shortcode)
}

func (suite *EmojiTestSuite) TestGetAllEmojisLimit1() {
	emojis, err := suite.db.GetEmojisBy(suite.T().Context(), db.EmojiAllDomains, true, true, "", "", "", 1)

	suite.NoError(err)
	suite.Equal(1, len(emojis))
	suite.Equal("rainbow", emojis[0].Shortcode)
}

func (suite *EmojiTestSuite) TestGetAllEmojisMaxID() {
	emojis, err := suite.db.GetEmojisBy(suite.T().Context(), db.EmojiAllDomains, true, true, "", "rainbow@", "", 0)

	suite.NoError(err)
	suite.Equal(1, len(emojis))
	suite.Equal("yell", emojis[0].Shortcode)
}

func (suite *EmojiTestSuite) TestGetAllEmojisMinID() {
	emojis, err := suite.db.GetEmojisBy(suite.T().Context(), db.EmojiAllDomains, true, true, "", "", "yell@fossbros-anonymous.io", 0)

	suite.NoError(err)
	suite.Equal(1, len(emojis))
	suite.Equal("rainbow", emojis[0].Shortcode)
}

func (suite *EmojiTestSuite) TestGetAllDisabledEmojis() {
	emojis, err := suite.db.GetEmojisBy(suite.T().Context(), db.EmojiAllDomains, true, false, "", "", "", 0)

	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Equal(0, len(emojis))
}

func (suite *EmojiTestSuite) TestGetAllEnabledEmojis() {
	emojis, err := suite.db.GetEmojisBy(suite.T().Context(), db.EmojiAllDomains, false, true, "", "", "", 0)

	suite.NoError(err)
	suite.Equal(2, len(emojis))
	suite.Equal("rainbow", emojis[0].Shortcode)
	suite.Equal("yell", emojis[1].Shortcode)
}

func (suite *EmojiTestSuite) TestGetLocalEnabledEmojis() {
	emojis, err := suite.db.GetEmojisBy(suite.T().Context(), "", false, true, "", "", "", 0)

	suite.NoError(err)
	suite.Equal(1, len(emojis))
	suite.Equal("rainbow", emojis[0].Shortcode)
}

func (suite *EmojiTestSuite) TestGetLocalDisabledEmojis() {
	emojis, err := suite.db.GetEmojisBy(suite.T().Context(), "", true, false, "", "", "", 0)

	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Equal(0, len(emojis))
}

func (suite *EmojiTestSuite) TestGetAllEmojisFromDomain() {
	emojis, err := suite.db.GetEmojisBy(suite.T().Context(), "peepee.poopoo", true, true, "", "", "", 0)

	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Equal(0, len(emojis))
}

func (suite *EmojiTestSuite) TestGetAllEmojisFromDomain2() {
	emojis, err := suite.db.GetEmojisBy(suite.T().Context(), "fossbros-anonymous.io", true, true, "", "", "", 0)

	suite.NoError(err)
	suite.Equal(1, len(emojis))
	suite.Equal("yell", emojis[0].Shortcode)
}

func (suite *EmojiTestSuite) TestGetSpecificEmojisFromDomain2() {
	emojis, err := suite.db.GetEmojisBy(suite.T().Context(), "fossbros-anonymous.io", true, true, "yell", "", "", 0)

	suite.NoError(err)
	suite.Equal(1, len(emojis))
	suite.Equal("yell", emojis[0].Shortcode)
}

func (suite *EmojiTestSuite) TestGetEmojiCategories() {
	categories, err := suite.db.GetEmojiCategories(suite.T().Context())
	suite.NoError(err)
	suite.Len(categories, 2)
	// check alphabetical order
	suite.Equal(categories[0].Name, "cute stuff")
	suite.Equal(categories[1].Name, "reactions")
}

func (suite *EmojiTestSuite) TestGetEmojiCategory() {
	category, err := suite.db.GetEmojiCategory(suite.T().Context(), testrig.NewTestEmojiCategories()["reactions"].ID)
	suite.NoError(err)
	suite.NotNil(category)
}

func (suite *EmojiTestSuite) TestUpdateEmojiCategory() {
	testEmoji := new(gtsmodel.Emoji)
	*testEmoji = *suite.testEmojis["rainbow"]

	testEmoji.CategoryID = ""

	err := suite.db.UpdateEmoji(suite.T().Context(), testEmoji, "category_id")
	suite.NoError(err)
}

func TestEmojiTestSuite(t *testing.T) {
	suite.Run(t, new(EmojiTestSuite))
}
