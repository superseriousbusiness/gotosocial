package admin_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/admin"
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

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.NotEmpty(b)
}

func TestEmojiCreateTestSuite(t *testing.T) {
	suite.Run(t, &EmojiCreateTestSuite{})
}
