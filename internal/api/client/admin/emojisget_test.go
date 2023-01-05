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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/admin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

type EmojisGetTestSuite struct {
	AdminStandardTestSuite
}

func (suite *EmojisGetTestSuite) TestEmojiGet() {
	recorder := httptest.NewRecorder()

	path := admin.EmojiPath + "?filter=domain:all&limit=1"
	ctx := suite.newContext(recorder, http.MethodGet, nil, path, "application/json")

	suite.adminModule.EmojisGETHandler(ctx)
	suite.Equal(http.StatusOK, recorder.Code)

	b, err := io.ReadAll(recorder.Body)
	suite.NoError(err)
	suite.NotNil(b)

	apiEmojis := []*apimodel.AdminEmoji{}
	if err := json.Unmarshal(b, &apiEmojis); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(apiEmojis, 1)
	suite.Equal("rainbow", apiEmojis[0].Shortcode)
	suite.Equal("", apiEmojis[0].Domain)

	suite.Equal(`<http://localhost:8080/api/v1/admin/custom_emojis?limit=1&max_shortcode_domain=rainbow@&filter=domain:all>; rel="next", <http://localhost:8080/api/v1/admin/custom_emojis?limit=1&min_shortcode_domain=rainbow@&filter=domain:all>; rel="prev"`, recorder.Header().Get("link"))
}

func (suite *EmojisGetTestSuite) TestEmojiGet2() {
	recorder := httptest.NewRecorder()

	path := admin.EmojiPath + "?filter=domain:all&limit=1&max_shortcode_domain=rainbow@"
	ctx := suite.newContext(recorder, http.MethodGet, nil, path, "application/json")

	suite.adminModule.EmojisGETHandler(ctx)
	suite.Equal(http.StatusOK, recorder.Code)

	b, err := io.ReadAll(recorder.Body)
	suite.NoError(err)
	suite.NotNil(b)

	apiEmojis := []*apimodel.AdminEmoji{}
	if err := json.Unmarshal(b, &apiEmojis); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(apiEmojis, 1)
	suite.Equal("yell", apiEmojis[0].Shortcode)
	suite.Equal("fossbros-anonymous.io", apiEmojis[0].Domain)

	suite.Equal(`<http://localhost:8080/api/v1/admin/custom_emojis?limit=1&max_shortcode_domain=yell@fossbros-anonymous.io&filter=domain:all>; rel="next", <http://localhost:8080/api/v1/admin/custom_emojis?limit=1&min_shortcode_domain=yell@fossbros-anonymous.io&filter=domain:all>; rel="prev"`, recorder.Header().Get("link"))
}

func (suite *EmojisGetTestSuite) TestEmojiGet3() {
	recorder := httptest.NewRecorder()

	path := admin.EmojiPath + "?filter=domain:all&limit=1&min_shortcode_domain=yell@fossbros-anonymous.io"
	ctx := suite.newContext(recorder, http.MethodGet, nil, path, "application/json")

	suite.adminModule.EmojisGETHandler(ctx)
	suite.Equal(http.StatusOK, recorder.Code)

	b, err := io.ReadAll(recorder.Body)
	suite.NoError(err)
	suite.NotNil(b)

	apiEmojis := []*apimodel.AdminEmoji{}
	if err := json.Unmarshal(b, &apiEmojis); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(apiEmojis, 1)
	suite.Equal("rainbow", apiEmojis[0].Shortcode)
	suite.Equal("", apiEmojis[0].Domain)

	suite.Equal(`<http://localhost:8080/api/v1/admin/custom_emojis?limit=1&max_shortcode_domain=rainbow@&filter=domain:all>; rel="next", <http://localhost:8080/api/v1/admin/custom_emojis?limit=1&min_shortcode_domain=rainbow@&filter=domain:all>; rel="prev"`, recorder.Header().Get("link"))
}

func TestEmojisGetTestSuite(t *testing.T) {
	suite.Run(t, &EmojisGetTestSuite{})
}
