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
	"time"

	"github.com/stretchr/testify/suite"
)

type MediaTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *MediaTestSuite) TestGetAttachmentByID() {
	testAttachment := suite.testAttachments["admin_account_status_1_attachment_1"]
	attachment, err := suite.db.GetAttachmentByID(suite.T().Context(), testAttachment.ID)
	suite.NoError(err)
	suite.NotNil(attachment)
}

func (suite *MediaTestSuite) TestGetOlder() {
	attachments, err := suite.db.GetCachedAttachmentsOlderThan(suite.T().Context(), time.Now(), 20)
	suite.NoError(err)
	suite.Len(attachments, 3)
}

func (suite *MediaTestSuite) TestGetCachedAttachmentsOlderThan() {
	ctx := suite.T().Context()

	attachments, err := suite.db.GetCachedAttachmentsOlderThan(ctx, time.Now(), 20)
	suite.NoError(err)
	suite.Len(attachments, 3)
}

func TestMediaTestSuite(t *testing.T) {
	suite.Run(t, new(MediaTestSuite))
}
