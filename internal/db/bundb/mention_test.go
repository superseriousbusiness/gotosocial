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

	"github.com/stretchr/testify/suite"
)

type MentionTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *MentionTestSuite) TestGetMention() {
	m := suite.testMentions["local_user_2_mention_zork"]

	dbMention, err := suite.db.GetMention(suite.T().Context(), m.ID)
	suite.NoError(err)
	suite.NotNil(dbMention)
	suite.Equal(m.ID, dbMention.ID)
	suite.Equal(m.OriginAccountID, dbMention.OriginAccountID)
	suite.NotNil(dbMention.OriginAccount)
	suite.Equal(m.TargetAccountID, dbMention.TargetAccountID)
	suite.NotNil(dbMention.TargetAccount)
	suite.Equal(m.StatusID, dbMention.StatusID)
	suite.NotNil(dbMention.Status)
}

func (suite *MentionTestSuite) TestGetMentions() {
	m := suite.testMentions["local_user_2_mention_zork"]

	dbMentions, err := suite.db.GetMentions(suite.T().Context(), []string{m.ID})
	suite.NoError(err)
	suite.Len(dbMentions, 1)
	dbMention := dbMentions[0]
	suite.Equal(m.ID, dbMention.ID)
	suite.Equal(m.OriginAccountID, dbMention.OriginAccountID)
	suite.NotNil(dbMention.OriginAccount)
	suite.Equal(m.TargetAccountID, dbMention.TargetAccountID)
	suite.NotNil(dbMention.TargetAccount)
	suite.Equal(m.StatusID, dbMention.StatusID)
	suite.NotNil(dbMention.Status)
}

func TestMentionTestSuite(t *testing.T) {
	suite.Run(t, new(MentionTestSuite))
}
