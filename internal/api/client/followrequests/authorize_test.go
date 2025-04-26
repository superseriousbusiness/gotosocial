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

package followrequests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/followrequests"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type AuthorizeTestSuite struct {
	FollowRequestStandardTestSuite
}

func (suite *AuthorizeTestSuite) TestAuthorize() {
	requestingAccount := suite.testAccounts["remote_account_2"]
	targetAccount := suite.testAccounts["local_account_1"]

	// put a follow request in the database
	fr := &gtsmodel.FollowRequest{
		ID:              "01FJ1S8DX3STJJ6CEYPMZ1M0R3",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		URI:             fmt.Sprintf("%s/follow/01FJ1S8DX3STJJ6CEYPMZ1M0R3", requestingAccount.URI),
		AccountID:       requestingAccount.ID,
		TargetAccountID: targetAccount.ID,
	}

	err := suite.db.Put(context.Background(), fr)
	suite.NoError(err)

	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, []byte{}, fmt.Sprintf("/api/v1/follow_requests/%s/authorize", requestingAccount.ID), "")

	ctx.Params = gin.Params{
		gin.Param{
			Key:   followrequests.IDKey,
			Value: requestingAccount.ID,
		},
	}

	// call the handler
	suite.followRequestModule.FollowRequestAuthorizePOSTHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	dst := new(bytes.Buffer)
	err = json.Indent(dst, b, "", "  ")
	suite.NoError(err)
	suite.Equal(`{
  "id": "01FHMQX3GAABWSM0S2VZEC2SWC",
  "following": false,
  "showing_reblogs": false,
  "notifying": false,
  "followed_by": true,
  "blocking": false,
  "blocked_by": false,
  "muting": false,
  "muting_notifications": false,
  "requested": false,
  "requested_by": false,
  "domain_blocking": false,
  "endorsed": false,
  "note": ""
}`, dst.String())
}

func (suite *AuthorizeTestSuite) TestAuthorizeNoFR() {
	requestingAccount := suite.testAccounts["remote_account_2"]

	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, []byte{}, fmt.Sprintf("/api/v1/follow_requests/%s/authorize", requestingAccount.ID), "")

	ctx.Params = gin.Params{
		gin.Param{
			Key:   followrequests.IDKey,
			Value: requestingAccount.ID,
		},
	}

	// call the handler
	suite.followRequestModule.FollowRequestAuthorizePOSTHandler(ctx)

	suite.Equal(http.StatusNotFound, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Not Found"}`, string(b))
}

func TestAuthorizeTestSuite(t *testing.T) {
	suite.Run(t, &AuthorizeTestSuite{})
}
