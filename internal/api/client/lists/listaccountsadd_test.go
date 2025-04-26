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

package lists_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/lists"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type ListAccountsAddTestSuite struct {
	ListsStandardTestSuite
}

func (suite *ListAccountsAddTestSuite) postListAccounts(
	expectedHTTPStatus int,
	listID string,
	accountIDs []string,
) ([]byte, error) {
	var (
		recorder = httptest.NewRecorder()
		ctx, _   = testrig.CreateGinTestContext(recorder, nil)
	)

	// Prepare test context.
	ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	// Inject path parameters.
	ctx.AddParam("id", listID)

	// Inject query parameters.
	requestPath := config.GetProtocol() + "://" + config.GetHost() + "/api/" + lists.BasePath + "/" + listID + "/accounts"

	// Prepare test body.
	buf, w, err := testrig.CreateMultipartFormData(nil, map[string][]string{
		"account_ids[]": accountIDs,
	})

	// Prepare test context request.
	request := httptest.NewRequest(http.MethodPost, requestPath, bytes.NewReader(buf.Bytes()))
	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", w.FormDataContentType())
	ctx.Request = request

	// trigger the handler
	suite.listsModule.ListAccountsPOSTHandler(ctx)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	// Check status code.
	if status := recorder.Code; expectedHTTPStatus != status {
		err = fmt.Errorf("expected %d got %d", expectedHTTPStatus, status)
	}

	return b, err
}

func (suite *ListAccountsAddTestSuite) TestPostListAccountNotFollowed() {
	listID := suite.testLists["local_account_1_list_1"].ID
	accountIDs := []string{
		suite.testAccounts["remote_account_1"].ID,
	}

	resp, err := suite.postListAccounts(http.StatusNotFound, listID, accountIDs)
	suite.NoError(err)
	suite.Equal(`{"error":"Not Found: account 01F8MH5ZK5VRH73AKHQM6Y9VNX not currently followed"}`, string(resp))
}

func (suite *ListAccountsAddTestSuite) TestPostListAccountOK() {
	entry := suite.testListEntries["local_account_1_list_1_entry_1"]

	// Remove turtle from the list.
	if err := suite.db.DeleteListEntry(
		context.Background(),
		entry.ListID,
		entry.FollowID,
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Add turtle back to the list.
	listID := suite.testLists["local_account_1_list_1"].ID
	accountIDs := []string{
		suite.testAccounts["local_account_2"].ID,
	}

	resp, err := suite.postListAccounts(http.StatusOK, listID, accountIDs)
	suite.NoError(err)
	suite.Equal(`{}`, string(resp))
}

func TestListAccountsAddTestSuite(t *testing.T) {
	suite.Run(t, new(ListAccountsAddTestSuite))
}
