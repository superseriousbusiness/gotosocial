/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package admin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/admin"
)

type MediaCleanupTestSuite struct {
	AdminStandardTestSuite
}

func (suite *MediaCleanupTestSuite) TestMediaCleanup() {
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_2"]
	suite.True(testAttachment.Cached)

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, []byte("{\"remote_cache_days\": 1}"), admin.EmojiPath, "application/json")

	// call the handler
	suite.adminModule.MediaCleanupPOSTHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// Wait for async task to finish
	time.Sleep(1 * time.Second)

	// Get media we prunes
	prunedAttachment, err := suite.db.GetAttachmentByID(context.Background(), testAttachment.ID)
	suite.NoError(err)

	// the media should no longer be cached
	suite.False(prunedAttachment.Cached)
}

func (suite *MediaCleanupTestSuite) TestMediaCleanupNoArg() {
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_2"]
	suite.True(testAttachment.Cached)
	println("TIME: ", testAttachment.CreatedAt.String())

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, []byte("{}"), admin.EmojiPath, "application/json")

	// call the handler
	suite.adminModule.MediaCleanupPOSTHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// Wait for async task to finish
	time.Sleep(1 * time.Second)

	// Get media we prunes
	prunedAttachment, err := suite.db.GetAttachmentByID(context.Background(), testAttachment.ID)
	suite.NoError(err)

	// the media should no longer be cached
	suite.False(prunedAttachment.Cached)
}

func (suite *MediaCleanupTestSuite) TestMediaCleanupNotOldEnough() {
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_2"]
	suite.True(testAttachment.Cached)

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, []byte("{\"remote_cache_days\": 10000}"), admin.EmojiPath, "application/json")

	// call the handler
	suite.adminModule.MediaCleanupPOSTHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// Wait for async task to finish
	time.Sleep(1 * time.Second)

	// Get media we prunes
	prunedAttachment, err := suite.db.GetAttachmentByID(context.Background(), testAttachment.ID)
	suite.NoError(err)

	// the media should still be cached
	suite.True(prunedAttachment.Cached)
}

func TestMediaCleanupTestSuite(t *testing.T) {
	suite.Run(t, &MediaCleanupTestSuite{})
}
