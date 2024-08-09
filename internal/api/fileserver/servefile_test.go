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

package fileserver_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/fileserver"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ServeFileTestSuite struct {
	FileserverTestSuite
}

// GetFile is just a convenience function to save repetition in this test suite.
// It takes the required params to serve a file, calls the handler, and returns
// the http status code, the response headers, and the parsed body bytes.
func (suite *ServeFileTestSuite) GetFile(
	accountID string,
	mediaType media.Type,
	mediaSize media.Size,
	filename string,
) (code int, headers http.Header, body []byte) {
	recorder := httptest.NewRecorder()

	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodGet, "http://localhost:8080/whatever", nil)
	ctx.Request.Header.Set("accept", "*/*")
	ctx.AddParam(fileserver.AccountIDKey, accountID)
	ctx.AddParam(fileserver.MediaTypeKey, string(mediaType))
	ctx.AddParam(fileserver.MediaSizeKey, string(mediaSize))
	ctx.AddParam(fileserver.FileNameKey, filename)

	logger := middleware.Logger(false)
	suite.fileServer.ServeFile(ctx)
	logger(ctx)

	code = recorder.Code
	headers = recorder.Result().Header

	var err error
	body, err = io.ReadAll(recorder.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return
}

// UncacheAttachment is a convenience function that uncaches the targetAttachment by
// removing its associated files from storage, and updating the database.
func (suite *ServeFileTestSuite) UncacheAttachment(targetAttachment *gtsmodel.MediaAttachment) {
	ctx := context.Background()

	cached := false
	targetAttachment.Cached = &cached

	if err := suite.db.UpdateByID(ctx, targetAttachment, targetAttachment.ID, "cached"); err != nil {
		suite.FailNow(err.Error())
	}
	if err := suite.storage.Delete(ctx, targetAttachment.File.Path); err != nil {
		suite.FailNow(err.Error())
	}
	if err := suite.storage.Delete(ctx, targetAttachment.Thumbnail.Path); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *ServeFileTestSuite) TestServeOriginalLocalFileOK() {
	targetAttachment := &gtsmodel.MediaAttachment{}
	*targetAttachment = *suite.testAttachments["admin_account_status_1_attachment_1"]
	fileInStorage, err := suite.storage.Get(context.Background(), targetAttachment.File.Path)
	if err != nil {
		suite.FailNow(err.Error())
	}

	code, headers, body := suite.GetFile(
		targetAttachment.AccountID,
		media.TypeAttachment,
		media.SizeOriginal,
		targetAttachment.ID+".jpg",
	)

	suite.Equal(http.StatusOK, code)
	suite.Equal("image/jpeg", headers.Get("content-type"))
	suite.Equal(fileInStorage, body)
}

func (suite *ServeFileTestSuite) TestServeSmallLocalFileOK() {
	targetAttachment := &gtsmodel.MediaAttachment{}
	*targetAttachment = *suite.testAttachments["admin_account_status_1_attachment_1"]
	fileInStorage, err := suite.storage.Get(context.Background(), targetAttachment.Thumbnail.Path)
	if err != nil {
		suite.FailNow(err.Error())
	}

	code, headers, body := suite.GetFile(
		targetAttachment.AccountID,
		media.TypeAttachment,
		media.SizeSmall,
		targetAttachment.ID+".webp",
	)

	suite.Equal(http.StatusOK, code)
	suite.Equal("image/webp", headers.Get("content-type"))
	suite.Equal(fileInStorage, body)
}

func (suite *ServeFileTestSuite) TestServeOriginalRemoteFileOK() {
	targetAttachment := &gtsmodel.MediaAttachment{}
	*targetAttachment = *suite.testAttachments["remote_account_1_status_1_attachment_1"]
	fileInStorage, err := suite.storage.Get(context.Background(), targetAttachment.File.Path)
	if err != nil {
		suite.FailNow(err.Error())
	}

	code, headers, body := suite.GetFile(
		targetAttachment.AccountID,
		media.TypeAttachment,
		media.SizeOriginal,
		targetAttachment.ID+".jpg",
	)

	suite.Equal(http.StatusOK, code)
	suite.Equal("image/jpeg", headers.Get("content-type"))
	suite.Equal(fileInStorage, body)
}

func (suite *ServeFileTestSuite) TestServeSmallRemoteFileOK() {
	targetAttachment := &gtsmodel.MediaAttachment{}
	*targetAttachment = *suite.testAttachments["remote_account_1_status_1_attachment_1"]
	fileInStorage, err := suite.storage.Get(context.Background(), targetAttachment.Thumbnail.Path)
	if err != nil {
		suite.FailNow(err.Error())
	}

	code, headers, body := suite.GetFile(
		targetAttachment.AccountID,
		media.TypeAttachment,
		media.SizeSmall,
		targetAttachment.ID+".webp",
	)

	suite.Equal(http.StatusOK, code)
	suite.Equal("image/jpeg", headers.Get("content-type"))
	suite.Equal(fileInStorage, body)
}

func (suite *ServeFileTestSuite) TestServeOriginalRemoteFileRecache() {
	targetAttachment := &gtsmodel.MediaAttachment{}
	*targetAttachment = *suite.testAttachments["remote_account_1_status_1_attachment_1"]
	fileInStorage, err := suite.storage.Get(context.Background(), targetAttachment.File.Path)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// uncache the attachment so we'll have to refetch it from the 'remote' instance
	suite.UncacheAttachment(targetAttachment)

	code, headers, body := suite.GetFile(
		targetAttachment.AccountID,
		media.TypeAttachment,
		media.SizeOriginal,
		targetAttachment.ID+".jpg",
	)

	suite.Equal(http.StatusOK, code)
	suite.Equal("image/jpeg", headers.Get("content-type"))
	suite.Equal(fileInStorage, body)
}

func (suite *ServeFileTestSuite) TestServeSmallRemoteFileRecache() {
	targetAttachment := &gtsmodel.MediaAttachment{}
	*targetAttachment = *suite.testAttachments["remote_account_1_status_1_attachment_1"]
	fileInStorage, err := suite.storage.Get(context.Background(), targetAttachment.Thumbnail.Path)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// uncache the attachment so we'll have to refetch it from the 'remote' instance
	suite.UncacheAttachment(targetAttachment)

	code, headers, body := suite.GetFile(
		targetAttachment.AccountID,
		media.TypeAttachment,
		media.SizeSmall,
		targetAttachment.ID+".webp",
	)

	suite.Equal(http.StatusOK, code)
	suite.Equal("image/jpeg", headers.Get("content-type"))
	suite.Equal(fileInStorage, body)
}

func (suite *ServeFileTestSuite) TestServeOriginalRemoteFileRecacheNotFound() {
	targetAttachment := &gtsmodel.MediaAttachment{}
	*targetAttachment = *suite.testAttachments["remote_account_1_status_1_attachment_1"]

	// uncache the attachment *and* set the remote URL to something that will return a 404
	suite.UncacheAttachment(targetAttachment)
	targetAttachment.RemoteURL = "http://nothing.at.this.url/weeeeeeeee"
	if err := suite.db.UpdateByID(context.Background(), targetAttachment, targetAttachment.ID, "remote_url"); err != nil {
		suite.FailNow(err.Error())
	}

	code, _, _ := suite.GetFile(
		targetAttachment.AccountID,
		media.TypeAttachment,
		media.SizeOriginal,
		targetAttachment.ID+".jpg",
	)

	suite.Equal(http.StatusNotFound, code)
}

func (suite *ServeFileTestSuite) TestServeSmallRemoteFileRecacheNotFound() {
	targetAttachment := &gtsmodel.MediaAttachment{}
	*targetAttachment = *suite.testAttachments["remote_account_1_status_1_attachment_1"]

	// uncache the attachment *and* set the remote URL to something that will return a 404
	suite.UncacheAttachment(targetAttachment)
	targetAttachment.RemoteURL = "http://nothing.at.this.url/weeeeeeeee"
	if err := suite.db.UpdateByID(context.Background(), targetAttachment, targetAttachment.ID, "remote_url"); err != nil {
		suite.FailNow(err.Error())
	}

	code, _, _ := suite.GetFile(
		targetAttachment.AccountID,
		media.TypeAttachment,
		media.SizeSmall,
		targetAttachment.ID+".webp",
	)

	suite.Equal(http.StatusNotFound, code)
}

// Callers trying to get some random-ass file that doesn't exist should just get a 404
func (suite *ServeFileTestSuite) TestServeFileNotFound() {
	code, _, _ := suite.GetFile(
		"01GMMY4G9B0QEG0PQK5Q5JGJWZ",
		media.TypeAttachment,
		media.SizeOriginal,
		"01GMMY68Y7E5DJ3CA3Y9SS8524.jpg",
	)

	suite.Equal(http.StatusNotFound, code)
}

func TestServeFileTestSuite(t *testing.T) {
	suite.Run(t, new(ServeFileTestSuite))
}
