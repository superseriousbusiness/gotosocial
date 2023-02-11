/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package bundb_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type MediaTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *MediaTestSuite) TestGetAttachmentByID() {
	testAttachment := suite.testAttachments["admin_account_status_1_attachment_1"]
	attachment, err := suite.db.GetAttachmentByID(context.Background(), testAttachment.ID)
	suite.NoError(err)
	suite.NotNil(attachment)
}

func (suite *MediaTestSuite) TestGetOlder() {
	attachments, err := suite.db.GetRemoteOlderThan(context.Background(), time.Now(), 20)
	suite.NoError(err)
	suite.Len(attachments, 2)
}

func (suite *MediaTestSuite) TestGetAvisAndHeaders() {
	ctx := context.Background()

	attachments, err := suite.db.GetAvatarsAndHeaders(ctx, "", 20)
	suite.NoError(err)
	suite.Len(attachments, 3)
}

func (suite *MediaTestSuite) TestGetLocalUnattachedOlderThan() {
	ctx := context.Background()

	attachments, err := suite.db.GetLocalUnattachedOlderThan(ctx, testrig.TimeMustParse("2090-06-04T13:12:00Z"), 10)
	suite.NoError(err)
	suite.Len(attachments, 1)
}

func TestMediaTestSuite(t *testing.T) {
	suite.Run(t, new(MediaTestSuite))
}
