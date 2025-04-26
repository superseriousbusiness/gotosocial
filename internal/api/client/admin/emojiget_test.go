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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/admin"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"github.com/stretchr/testify/suite"
)

type EmojiGetTestSuite struct {
	AdminStandardTestSuite
}

func (suite *EmojiGetTestSuite) TestEmojiGet1() {
	recorder := httptest.NewRecorder()
	testEmoji := suite.testEmojis["rainbow"]

	path := admin.EmojiPathWithID
	ctx := suite.newContext(recorder, http.MethodGet, nil, path, "application/json")
	ctx.AddParam(apiutil.IDKey, testEmoji.ID)

	suite.adminModule.EmojiGETHandler(ctx)
	suite.Equal(http.StatusOK, recorder.Code)

	b, err := io.ReadAll(recorder.Body)
	suite.NoError(err)
	suite.NotNil(b)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`{
  "shortcode": "rainbow",
  "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
  "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
  "visible_in_picker": true,
  "category": "reactions",
  "id": "01F8MH9H8E4VG3KDYJR9EGPXCQ",
  "disabled": false,
  "updated_at": "2021-09-20T10:40:37.000Z",
  "total_file_size": 42794,
  "content_type": "image/png",
  "uri": "http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ"
}`, dst.String())
}

func (suite *EmojiGetTestSuite) TestEmojiGet2() {
	recorder := httptest.NewRecorder()
	testEmoji := suite.testEmojis["yell"]

	path := admin.EmojiPathWithID
	ctx := suite.newContext(recorder, http.MethodGet, nil, path, "application/json")
	ctx.AddParam(apiutil.IDKey, testEmoji.ID)

	suite.adminModule.EmojiGETHandler(ctx)
	suite.Equal(http.StatusOK, recorder.Code)

	b, err := io.ReadAll(recorder.Body)
	suite.NoError(err)
	suite.NotNil(b)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`{
  "shortcode": "yell",
  "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01GD5KP5CQEE1R3X43Y1EHS2CW.png",
  "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01GD5KP5CQEE1R3X43Y1EHS2CW.png",
  "visible_in_picker": false,
  "id": "01GD5KP5CQEE1R3X43Y1EHS2CW",
  "disabled": false,
  "domain": "fossbros-anonymous.io",
  "updated_at": "2020-03-18T12:12:00.000Z",
  "total_file_size": 19854,
  "content_type": "image/png",
  "uri": "http://fossbros-anonymous.io/emoji/01GD5KP5CQEE1R3X43Y1EHS2CW"
}`, dst.String())
}

func (suite *EmojiGetTestSuite) TestEmojiGetNotFound() {
	recorder := httptest.NewRecorder()

	path := admin.EmojiPathWithID
	ctx := suite.newContext(recorder, http.MethodGet, nil, path, "application/json")
	ctx.AddParam(apiutil.IDKey, "01GF8VRXX1R00X7XH8973Z29R1")

	suite.adminModule.EmojiGETHandler(ctx)
	suite.Equal(http.StatusNotFound, recorder.Code)

	b, err := io.ReadAll(recorder.Body)
	suite.NoError(err)
	suite.NotNil(b)
	suite.Equal(`{"error":"Not Found"}`, string(b))
}

func TestEmojiGetTestSuite(t *testing.T) {
	suite.Run(t, &EmojiGetTestSuite{})
}
