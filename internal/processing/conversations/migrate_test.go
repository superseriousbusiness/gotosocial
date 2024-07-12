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

package conversations_test

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Test that we can migrate DMs to conversations.
// This test assumes that we're using the standard test fixtures, which contain some conversation-eligible DMs.
func (suite *ConversationsTestSuite) TestMigrateDMsToConversations() {
	advancedMigrationID := "20240611190733_add_conversations"
	ctx := context.Background()
	rawDB := (suite.db).(*bundb.DBService).DB()

	// Precondition: we shouldn't have any conversations yet.
	numConversations := 0
	if err := rawDB.NewSelect().
		Model((*gtsmodel.Conversation)(nil)).
		ColumnExpr("COUNT(*)").
		Scan(ctx, &numConversations); // nocollapse
	err != nil {
		suite.FailNow(err.Error())
	}
	suite.Zero(numConversations)

	// Precondition: there is no record of the conversations advanced migration.
	_, err := suite.db.GetAdvancedMigration(ctx, advancedMigrationID)
	suite.ErrorIs(err, db.ErrNoEntries)

	// Run the migration, which should not fail.
	if err := suite.conversationsProcessor.MigrateDMsToConversations(ctx); err != nil {
		suite.FailNow(err.Error())
	}

	// We should now have some conversations.
	if err := rawDB.NewSelect().
		Model((*gtsmodel.Conversation)(nil)).
		ColumnExpr("COUNT(*)").
		Scan(ctx, &numConversations); // nocollapse
	err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotZero(numConversations)

	// The advanced migration should now be marked as finished.
	advancedMigration, err := suite.db.GetAdvancedMigration(ctx, advancedMigrationID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if suite.NotNil(advancedMigration) && suite.NotNil(advancedMigration.Finished) {
		suite.True(*advancedMigration.Finished)
	}

	// Run the migration again, which should not fail.
	if err := suite.conversationsProcessor.MigrateDMsToConversations(ctx); err != nil {
		suite.FailNow(err.Error())
	}

	// However, it shouldn't have done anything, so the advanced migration should not have been updated.
	advancedMigration2, err := suite.db.GetAdvancedMigration(ctx, advancedMigrationID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal(advancedMigration.UpdatedAt, advancedMigration2.UpdatedAt)
}
