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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/admin"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type MediaCleanupTestSuite struct {
	AdminStandardTestSuite
}

func (suite *MediaCleanupTestSuite) TestMediaCleanup() {
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	suite.True(*testAttachment.Cached)

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, []byte("{\"remote_cache_days\": 1}"), admin.MediaCleanupPath, "application/json")

	// call the handler
	suite.adminModule.MediaCleanupPOSTHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// the attachment should be updated in the database
	if !testrig.WaitFor(func() bool {
		if prunedAttachment, _ := suite.db.GetAttachmentByID(suite.T().Context(), testAttachment.ID); prunedAttachment != nil {
			return !*prunedAttachment.Cached
		}
		return false
	}) {
		suite.FailNow("timed out waiting for attachment to be pruned")
	}
}

func (suite *MediaCleanupTestSuite) TestMediaCleanupNoArg() {
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	suite.True(*testAttachment.Cached)
	println("TIME: ", testAttachment.CreatedAt.String())

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, []byte("{}"), admin.MediaCleanupPath, "application/json")

	// call the handler
	suite.adminModule.MediaCleanupPOSTHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	if !testrig.WaitFor(func() bool {
		if prunedAttachment, _ := suite.db.GetAttachmentByID(suite.T().Context(), testAttachment.ID); prunedAttachment != nil {
			return !*prunedAttachment.Cached
		}
		return false
	}) {
		suite.FailNow("timed out waiting for attachment to be pruned")
	}
}

func (suite *MediaCleanupTestSuite) TestMediaCleanupNotOldEnough() {
	testAttachment := suite.testAttachments["remote_account_1_status_1_attachment_1"]
	suite.True(*testAttachment.Cached)

	// set up the request
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, []byte("{\"remote_cache_days\": 10000}"), admin.MediaCleanupPath, "application/json")

	// call the handler
	suite.adminModule.MediaCleanupPOSTHandler(ctx)

	// we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// Wait for async task to finish
	time.Sleep(1 * time.Second)

	// Get media we pruned
	prunedAttachment, err := suite.db.GetAttachmentByID(suite.T().Context(), testAttachment.ID)
	suite.NoError(err)

	// the media should still be cached
	suite.True(*prunedAttachment.Cached)
}

func TestMediaCleanupTestSuite(t *testing.T) {
	suite.Run(t, &MediaCleanupTestSuite{})
}
