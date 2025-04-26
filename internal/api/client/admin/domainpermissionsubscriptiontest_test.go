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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/admin"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type DomainPermissionSubscriptionTestTestSuite struct {
	AdminStandardTestSuite
}

func (suite *DomainPermissionSubscriptionTestTestSuite) TestDomainPermissionSubscriptionTestCSV() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["admin_account"]
		permSub     = &gtsmodel.DomainPermissionSubscription{
			ID:                 "01JGE681TQSBPAV59GZXPKE62H",
			Priority:           255,
			Title:              "whatever!",
			PermissionType:     gtsmodel.DomainPermissionBlock,
			AsDraft:            util.Ptr(false),
			AdoptOrphans:       util.Ptr(true),
			CreatedByAccountID: testAccount.ID,
			CreatedByAccount:   testAccount,
			URI:                "https://lists.example.org/baddies.csv",
			ContentType:        gtsmodel.DomainPermSubContentTypeCSV,
		}
	)

	// Create a subscription for a CSV list of baddies.
	err := suite.state.DB.PutDomainPermissionSubscription(ctx, permSub)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Prepare the request to the /test endpoint.
	subPath := strings.ReplaceAll(
		admin.DomainPermissionSubscriptionTestPath,
		":id", permSub.ID,
	)
	path := "/api" + subPath
	recorder := httptest.NewRecorder()
	ginCtx := suite.newContext(recorder, http.MethodPost, nil, path, "application/json")
	ginCtx.Params = gin.Params{
		gin.Param{
			Key:   apiutil.IDKey,
			Value: permSub.ID,
		},
	}

	// Trigger the handler.
	suite.adminModule.DomainPermissionSubscriptionTestPOSTHandler(ginCtx)
	suite.Equal(http.StatusOK, recorder.Code)

	// Read the body back.
	b, err := io.ReadAll(recorder.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	dst := new(bytes.Buffer)
	if err := json.Indent(dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	// Ensure expected.
	suite.Equal(`[
  {
    "domain": "bumfaces.net",
    "public_comment": "big jerks",
    "obfuscate": false,
    "private_comment": ""
  },
  {
    "domain": "peepee.poopoo",
    "public_comment": "harassment",
    "obfuscate": false,
    "private_comment": ""
  },
  {
    "domain": "nothanks.com",
    "public_comment": "",
    "obfuscate": false,
    "private_comment": ""
  }
]`, dst.String())

	// No permissions should be created
	// since this is a dry run / test.
	blocked, err := suite.state.DB.AreDomainsBlocked(
		ctx,
		[]string{"bumfaces.net", "peepee.poopoo", "nothanks.com"},
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.False(blocked)
}

func (suite *DomainPermissionSubscriptionTestTestSuite) TestDomainPermissionSubscriptionTestText() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["admin_account"]
		permSub     = &gtsmodel.DomainPermissionSubscription{
			ID:                 "01JGE681TQSBPAV59GZXPKE62H",
			Priority:           255,
			Title:              "whatever!",
			PermissionType:     gtsmodel.DomainPermissionBlock,
			AsDraft:            util.Ptr(false),
			AdoptOrphans:       util.Ptr(true),
			CreatedByAccountID: testAccount.ID,
			CreatedByAccount:   testAccount,
			URI:                "https://lists.example.org/baddies.txt",
			ContentType:        gtsmodel.DomainPermSubContentTypePlain,
		}
	)

	// Create a subscription for a plaintext list of baddies.
	err := suite.state.DB.PutDomainPermissionSubscription(ctx, permSub)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Prepare the request to the /test endpoint.
	subPath := strings.ReplaceAll(
		admin.DomainPermissionSubscriptionTestPath,
		":id", permSub.ID,
	)
	path := "/api" + subPath
	recorder := httptest.NewRecorder()
	ginCtx := suite.newContext(recorder, http.MethodPost, nil, path, "application/json")
	ginCtx.Params = gin.Params{
		gin.Param{
			Key:   apiutil.IDKey,
			Value: permSub.ID,
		},
	}

	// Trigger the handler.
	suite.adminModule.DomainPermissionSubscriptionTestPOSTHandler(ginCtx)
	suite.Equal(http.StatusOK, recorder.Code)

	// Read the body back.
	b, err := io.ReadAll(recorder.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	dst := new(bytes.Buffer)
	if err := json.Indent(dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	// Ensure expected.
	suite.Equal(`[
  {
    "domain": "bumfaces.net",
    "public_comment": "",
    "obfuscate": false,
    "private_comment": ""
  },
  {
    "domain": "peepee.poopoo",
    "public_comment": "",
    "obfuscate": false,
    "private_comment": ""
  },
  {
    "domain": "nothanks.com",
    "public_comment": "",
    "obfuscate": false,
    "private_comment": ""
  }
]`, dst.String())

	// No permissions should be created
	// since this is a dry run / test.
	blocked, err := suite.state.DB.AreDomainsBlocked(
		ctx,
		[]string{"bumfaces.net", "peepee.poopoo", "nothanks.com"},
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.False(blocked)
}

func TestDomainPermissionSubscriptionTestTestSuite(t *testing.T) {
	suite.Run(t, &DomainPermissionSubscriptionTestTestSuite{})
}
