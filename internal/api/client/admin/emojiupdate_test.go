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

package admin_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/admin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type EmojiUpdateTestSuite struct {
	AdminStandardTestSuite
}

func (suite *EmojiUpdateTestSuite) TestEmojiUpdateNewCategory() {
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *suite.testEmojis["rainbow"]

	// set up the request
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"category": "New Category", // this category doesn't exist yet
			"type":     "modify",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPathWithID, w.FormDataContentType())
	ctx.AddParam(admin.IDKey, testEmoji.ID)

	// call the handler
	suite.adminModule.EmojiPATCHHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.NotEmpty(b)

	// response should be an admin model emoji
	adminEmoji := &apimodel.AdminEmoji{}
	err = json.Unmarshal(b, adminEmoji)
	suite.NoError(err)

	// appropriate fields should be set
	suite.Equal("rainbow", adminEmoji.Shortcode)
	suite.NotEmpty(adminEmoji.URL)
	suite.NotEmpty(adminEmoji.StaticURL)
	suite.True(adminEmoji.VisibleInPicker)

	// emoji should be in the db
	dbEmoji, err := suite.db.GetEmojiByShortcodeDomain(context.Background(), adminEmoji.Shortcode, "")
	suite.NoError(err)

	// check fields on the emoji
	suite.NotEmpty(dbEmoji.ID)
	suite.Equal("rainbow", dbEmoji.Shortcode)
	suite.Empty(dbEmoji.Domain)
	suite.Empty(dbEmoji.ImageRemoteURL)
	suite.Empty(dbEmoji.ImageStaticRemoteURL)
	suite.Equal(adminEmoji.URL, dbEmoji.ImageURL)
	suite.Equal(adminEmoji.StaticURL, dbEmoji.ImageStaticURL)
	suite.NotEmpty(dbEmoji.ImagePath)
	suite.NotEmpty(dbEmoji.ImageStaticPath)
	suite.Equal("image/png", dbEmoji.ImageContentType)
	suite.Equal("image/png", dbEmoji.ImageStaticContentType)
	suite.Equal(36702, dbEmoji.ImageFileSize)
	suite.Equal(10413, dbEmoji.ImageStaticFileSize)
	suite.False(*dbEmoji.Disabled)
	suite.NotEmpty(dbEmoji.URI)
	suite.True(*dbEmoji.VisibleInPicker)
	suite.NotEmpty(dbEmoji.CategoryID)

	// emoji should be in storage
	emojiBytes, err := suite.storage.Get(ctx, dbEmoji.ImagePath)
	suite.NoError(err)
	suite.Len(emojiBytes, dbEmoji.ImageFileSize)
	emojiStaticBytes, err := suite.storage.Get(ctx, dbEmoji.ImageStaticPath)
	suite.NoError(err)
	suite.Len(emojiStaticBytes, dbEmoji.ImageStaticFileSize)
}

func (suite *EmojiUpdateTestSuite) TestEmojiUpdateSwitchCategory() {
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *suite.testEmojis["rainbow"]

	// set up the request
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"type":     "modify",
			"category": "cute stuff",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPathWithID, w.FormDataContentType())
	ctx.AddParam(admin.IDKey, testEmoji.ID)

	// call the handler
	suite.adminModule.EmojiPATCHHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.NotEmpty(b)

	// response should be an admin model emoji
	adminEmoji := &apimodel.AdminEmoji{}
	err = json.Unmarshal(b, adminEmoji)
	suite.NoError(err)

	// appropriate fields should be set
	suite.Equal("rainbow", adminEmoji.Shortcode)
	suite.NotEmpty(adminEmoji.URL)
	suite.NotEmpty(adminEmoji.StaticURL)
	suite.True(adminEmoji.VisibleInPicker)

	// emoji should be in the db
	dbEmoji, err := suite.db.GetEmojiByShortcodeDomain(context.Background(), adminEmoji.Shortcode, "")
	suite.NoError(err)

	// check fields on the emoji
	suite.NotEmpty(dbEmoji.ID)
	suite.Equal("rainbow", dbEmoji.Shortcode)
	suite.Empty(dbEmoji.Domain)
	suite.Empty(dbEmoji.ImageRemoteURL)
	suite.Empty(dbEmoji.ImageStaticRemoteURL)
	suite.Equal(adminEmoji.URL, dbEmoji.ImageURL)
	suite.Equal(adminEmoji.StaticURL, dbEmoji.ImageStaticURL)
	suite.NotEmpty(dbEmoji.ImagePath)
	suite.NotEmpty(dbEmoji.ImageStaticPath)
	suite.Equal("image/png", dbEmoji.ImageContentType)
	suite.Equal("image/png", dbEmoji.ImageStaticContentType)
	suite.Equal(36702, dbEmoji.ImageFileSize)
	suite.Equal(10413, dbEmoji.ImageStaticFileSize)
	suite.False(*dbEmoji.Disabled)
	suite.NotEmpty(dbEmoji.URI)
	suite.True(*dbEmoji.VisibleInPicker)
	suite.NotEmpty(dbEmoji.CategoryID)

	// emoji should be in storage
	emojiBytes, err := suite.storage.Get(ctx, dbEmoji.ImagePath)
	suite.NoError(err)
	suite.Len(emojiBytes, dbEmoji.ImageFileSize)
	emojiStaticBytes, err := suite.storage.Get(ctx, dbEmoji.ImageStaticPath)
	suite.NoError(err)
	suite.Len(emojiStaticBytes, dbEmoji.ImageStaticFileSize)
}

func (suite *EmojiUpdateTestSuite) TestEmojiUpdateCopyRemoteToLocal() {
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *suite.testEmojis["yell"]

	// set up the request
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"type":      "copy",
			"category":  "emojis i stole",
			"shortcode": "yell",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPathWithID, w.FormDataContentType())
	ctx.AddParam(admin.IDKey, testEmoji.ID)

	// call the handler
	suite.adminModule.EmojiPATCHHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.NotEmpty(b)

	// response should be an admin model emoji
	adminEmoji := &apimodel.AdminEmoji{}
	err = json.Unmarshal(b, adminEmoji)
	suite.NoError(err)

	// appropriate fields should be set
	suite.Equal("yell", adminEmoji.Shortcode)
	suite.NotEmpty(adminEmoji.URL)
	suite.NotEmpty(adminEmoji.StaticURL)
	suite.True(adminEmoji.VisibleInPicker)

	// emoji should be in the db
	dbEmoji, err := suite.db.GetEmojiByShortcodeDomain(context.Background(), adminEmoji.Shortcode, "")
	suite.NoError(err)

	// check fields on the emoji
	suite.NotEmpty(dbEmoji.ID)
	suite.Equal("yell", dbEmoji.Shortcode)
	suite.Empty(dbEmoji.Domain)
	suite.Empty(dbEmoji.ImageRemoteURL)
	suite.Empty(dbEmoji.ImageStaticRemoteURL)
	suite.Equal(adminEmoji.URL, dbEmoji.ImageURL)
	suite.Equal(adminEmoji.StaticURL, dbEmoji.ImageStaticURL)
	suite.NotEmpty(dbEmoji.ImagePath)
	suite.NotEmpty(dbEmoji.ImageStaticPath)
	suite.Equal("image/png", dbEmoji.ImageContentType)
	suite.Equal("image/png", dbEmoji.ImageStaticContentType)
	suite.Equal(10889, dbEmoji.ImageFileSize)
	suite.Equal(10672, dbEmoji.ImageStaticFileSize)
	suite.False(*dbEmoji.Disabled)
	suite.NotEmpty(dbEmoji.URI)
	suite.True(*dbEmoji.VisibleInPicker)
	suite.NotEmpty(dbEmoji.CategoryID)

	// emoji should be in storage
	emojiBytes, err := suite.storage.Get(ctx, dbEmoji.ImagePath)
	suite.NoError(err)
	suite.Len(emojiBytes, dbEmoji.ImageFileSize)
	emojiStaticBytes, err := suite.storage.Get(ctx, dbEmoji.ImageStaticPath)
	suite.NoError(err)
	suite.Len(emojiStaticBytes, dbEmoji.ImageStaticFileSize)
}

func (suite *EmojiUpdateTestSuite) TestEmojiUpdateDisableEmoji() {
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *suite.testEmojis["yell"]

	// set up the request
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"type": "disable",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPathWithID, w.FormDataContentType())
	ctx.AddParam(admin.IDKey, testEmoji.ID)

	// call the handler
	suite.adminModule.EmojiPATCHHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.NotEmpty(b)

	// response should be an admin model emoji
	adminEmoji := &apimodel.AdminEmoji{}
	err = json.Unmarshal(b, adminEmoji)
	suite.NoError(err)

	suite.True(adminEmoji.Disabled)
}

func (suite *EmojiUpdateTestSuite) TestEmojiUpdateDisableLocalEmoji() {
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *suite.testEmojis["rainbow"]

	// set up the request
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"type": "disable",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPathWithID, w.FormDataContentType())
	ctx.AddParam(admin.IDKey, testEmoji.ID)

	// call the handler
	suite.adminModule.EmojiPATCHHandler(ctx)
	suite.Equal(http.StatusBadRequest, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: emojiUpdateDisable: emoji 01F8MH9H8E4VG3KDYJR9EGPXCQ is not a remote emoji, cannot disable it via this endpoint"}`, string(b))
}

func (suite *EmojiUpdateTestSuite) TestEmojiUpdateModifyRemoteEmoji() {
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *suite.testEmojis["yell"]

	// set up the request
	requestBody, w, err := testrig.CreateMultipartFormData(
		"image", "../../../../testrig/media/kip-original.gif",
		map[string]string{
			"type": "modify",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPathWithID, w.FormDataContentType())
	ctx.AddParam(admin.IDKey, testEmoji.ID)

	// call the handler
	suite.adminModule.EmojiPATCHHandler(ctx)
	suite.Equal(http.StatusBadRequest, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: emojiUpdateModify: emoji 01GD5KP5CQEE1R3X43Y1EHS2CW is not a local emoji, cannot do a modify action on it"}`, string(b))
}

func (suite *EmojiUpdateTestSuite) TestEmojiUpdateModifyNoParams() {
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *suite.testEmojis["rainbow"]

	// set up the request
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"type": "modify",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPathWithID, w.FormDataContentType())
	ctx.AddParam(admin.IDKey, testEmoji.ID)

	// call the handler
	suite.adminModule.EmojiPATCHHandler(ctx)
	suite.Equal(http.StatusBadRequest, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: emoji action type was 'modify' but no image or category name was provided"}`, string(b))
}

func (suite *EmojiUpdateTestSuite) TestEmojiUpdateCopyLocalToLocal() {
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *suite.testEmojis["rainbow"]

	// set up the request
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"type":      "copy",
			"shortcode": "bottoms",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPathWithID, w.FormDataContentType())
	ctx.AddParam(admin.IDKey, testEmoji.ID)

	// call the handler
	suite.adminModule.EmojiPATCHHandler(ctx)
	suite.Equal(http.StatusBadRequest, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: emojiUpdateCopy: emoji 01F8MH9H8E4VG3KDYJR9EGPXCQ is not a remote emoji, cannot copy it to local"}`, string(b))
}

func (suite *EmojiUpdateTestSuite) TestEmojiUpdateCopyEmptyShortcode() {
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *suite.testEmojis["yell"]

	// set up the request
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"type":      "copy",
			"shortcode": "",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPathWithID, w.FormDataContentType())
	ctx.AddParam(admin.IDKey, testEmoji.ID)

	// call the handler
	suite.adminModule.EmojiPATCHHandler(ctx)
	suite.Equal(http.StatusBadRequest, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: shortcode  did not pass validation, must be between 2 and 30 characters, lowercase letters, numbers, and underscores only"}`, string(b))
}

func (suite *EmojiUpdateTestSuite) TestEmojiUpdateCopyNoShortcode() {
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *suite.testEmojis["yell"]

	// set up the request
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"type": "copy",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPathWithID, w.FormDataContentType())
	ctx.AddParam(admin.IDKey, testEmoji.ID)

	// call the handler
	suite.adminModule.EmojiPATCHHandler(ctx)
	suite.Equal(http.StatusBadRequest, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: emoji action type was 'copy' but no shortcode was provided"}`, string(b))
}

func (suite *EmojiUpdateTestSuite) TestEmojiUpdateCopyShortcodeAlreadyInUse() {
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *suite.testEmojis["yell"]

	// set up the request
	requestBody, w, err := testrig.CreateMultipartFormData(
		"", "",
		map[string]string{
			"type":      "copy",
			"shortcode": "rainbow",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPathWithID, w.FormDataContentType())
	ctx.AddParam(admin.IDKey, testEmoji.ID)

	// call the handler
	suite.adminModule.EmojiPATCHHandler(ctx)
	suite.Equal(http.StatusConflict, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Conflict: emojiUpdateCopy: emoji 01GD5KP5CQEE1R3X43Y1EHS2CW could not be copied, emoji with shortcode rainbow already exists on this instance"}`, string(b))
}

func TestEmojiUpdateTestSuite(t *testing.T) {
	suite.Run(t, &EmojiUpdateTestSuite{})
}
