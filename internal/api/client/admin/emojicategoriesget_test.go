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
	"github.com/stretchr/testify/suite"
)

type EmojiCategoriesGetTestSuite struct {
	AdminStandardTestSuite
}

func (suite *EmojiCategoriesGetTestSuite) TestEmojiCategoriesGet() {
	recorder := httptest.NewRecorder()

	path := admin.EmojiCategoriesPath
	ctx := suite.newContext(recorder, http.MethodGet, nil, path, "application/json")

	suite.adminModule.EmojiCategoriesGETHandler(ctx)
	suite.Equal(http.StatusOK, recorder.Code)

	b, err := io.ReadAll(recorder.Body)
	suite.NoError(err)
	suite.NotNil(b)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`[
  {
    "id": "01GGQ989PTT9PMRN4FZ1WWK2B9",
    "name": "cute stuff"
  },
  {
    "id": "01GGQ8V4993XK67B2JB396YFB7",
    "name": "reactions"
  }
]`, dst.String())
}

func TestEmojiCategoriesGetTestSuite(t *testing.T) {
	suite.Run(t, &EmojiCategoriesGetTestSuite{})
}
