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
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type EmojiCreateTestSuite struct {
	AdminStandardTestSuite
}

func (suite *EmojiCreateTestSuite) TestEmojiCreate() {
	// set up the request
	requestBody, w, err := testrig.CreateMultipartFormData(
		"image", "../../../../testrig/media/rainbow-original.png",
		map[string]string{
			"shortcode": "new_emoji",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPath, w.FormDataContentType())

	// call the handler
	suite.adminModule.EmojiCreatePOSTHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.NotEmpty(b)

	// response should be an api model emoji
	apiEmoji := &apimodel.Emoji{}
	err = json.Unmarshal(b, apiEmoji)
	suite.NoError(err)

	// appropriate fields should be set
	suite.Equal("new_emoji", apiEmoji.Shortcode)
	suite.NotEmpty(apiEmoji.URL)
	suite.NotEmpty(apiEmoji.StaticURL)
	suite.True(apiEmoji.VisibleInPicker)

	// emoji should be in the db
	dbEmoji := &gtsmodel.Emoji{}
	err = suite.db.GetWhere(context.Background(), []db.Where{{Key: "shortcode", Value: "new_emoji"}}, dbEmoji)
	suite.NoError(err)

	// check fields on the emoji
	suite.NotEmpty(dbEmoji.ID)
	suite.Equal("new_emoji", dbEmoji.Shortcode)
	suite.Empty(dbEmoji.Domain)
	suite.Empty(dbEmoji.ImageRemoteURL)
	suite.Empty(dbEmoji.ImageStaticRemoteURL)
	suite.Equal(apiEmoji.URL, dbEmoji.ImageURL)
	suite.Equal(apiEmoji.StaticURL, dbEmoji.ImageStaticURL)
	suite.NotEmpty(dbEmoji.ImagePath)
	suite.NotEmpty(dbEmoji.ImageStaticPath)
	suite.Equal("image/png", dbEmoji.ImageContentType)
	suite.Equal("image/png", dbEmoji.ImageStaticContentType)
	suite.Equal(36702, dbEmoji.ImageFileSize)
	suite.Equal(10413, dbEmoji.ImageStaticFileSize)
	suite.False(dbEmoji.Disabled)
	suite.NotEmpty(dbEmoji.URI)
	suite.True(dbEmoji.VisibleInPicker)
	suite.Empty(dbEmoji.CategoryID)

	// emoji should be in storage
	emojiBytes, err := suite.storage.Get(dbEmoji.ImagePath)
	suite.NoError(err)
	suite.Len(emojiBytes, dbEmoji.ImageFileSize)
	emojiStaticBytes, err := suite.storage.Get(dbEmoji.ImageStaticPath)
	suite.NoError(err)
	suite.Len(emojiStaticBytes, dbEmoji.ImageStaticFileSize)
}

func (suite *EmojiCreateTestSuite) TestEmojiCreateAlreadyExists() {
	// set up the request -- use a shortcode that already exists for an emoji in the database
	requestBody, w, err := testrig.CreateMultipartFormData(
		"image", "../../../../testrig/media/rainbow-original.png",
		map[string]string{
			"shortcode": "rainbow",
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, bodyBytes, admin.EmojiPath, w.FormDataContentType())

	// call the handler
	suite.adminModule.EmojiCreatePOSTHandler(ctx)

	suite.Equal(http.StatusConflict, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.NotEmpty(b)

	suite.Equal(`{"error":"Conflict: emoji with shortcode rainbow already exists"}`, string(b))
}

func TestEmojiCreateTestSuite(t *testing.T) {
	suite.Run(t, &EmojiCreateTestSuite{})
}
