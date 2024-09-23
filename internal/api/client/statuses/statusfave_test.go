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

package statuses_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/statuses"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusFaveTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusFaveTestSuite) postStatusFave(
	targetStatusID string,
	app *gtsmodel.Application,
	token *gtsmodel.Token,
	user *gtsmodel.User,
	account *gtsmodel.Account,
) (string, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Set(oauth.SessionAuthorizedApplication, app)
	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(token))
	ctx.Set(oauth.SessionAuthorizedUser, user)
	ctx.Set(oauth.SessionAuthorizedAccount, account)

	const pathBase = "http://localhost:8080/api" + statuses.FavouritePath
	path := strings.ReplaceAll(pathBase, ":"+apiutil.IDKey, targetStatusID)
	ctx.Request = httptest.NewRequest(http.MethodPost, path, nil)
	ctx.Request.Header.Set("accept", "application/json")

	// Populate target status ID.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   apiutil.IDKey,
			Value: targetStatusID,
		},
	}

	// Trigger handler.
	suite.statusModule.StatusFavePOSTHandler(ctx)
	return suite.parseStatusResponse(recorder)
}

// Fave a status we haven't faved yet.
func (suite *StatusFaveTestSuite) TestPostFave() {
	var (
		targetStatus = suite.testStatuses["admin_account_status_2"]
		app          = suite.testApplications["application_1"]
		token        = suite.testTokens["local_account_1"]
		user         = suite.testUsers["local_account_1"]
		account      = suite.testAccounts["local_account_1"]
	)

	out, recorder := suite.postStatusFave(
		targetStatus.ID,
		app,
		token,
		user,
		account,
	)

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Target status should now
	// be "favourited" by us.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "superseriousbusiness",
    "website": "https://superserious.business"
  },
  "bookmarked": false,
  "card": null,
  "content": "🐕🐕🐕🐕🐕",
  "created_at": "right the hell just now babyee",
  "emojis": [],
  "favourited": true,
  "favourites_count": 1,
  "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
  "in_reply_to_account_id": null,
  "in_reply_to_id": null,
  "interaction_policy": {
    "can_favourite": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    }
  },
  "language": "en",
  "media_attachments": [],
  "mentions": [],
  "muted": false,
  "pinned": false,
  "poll": null,
  "reblog": null,
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": true,
  "spoiler_text": "open to see some puppies",
  "tags": [],
  "text": "🐕🐕🐕🐕🐕",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "public"
}`, out)
}

// Try to fave a status
// that's not faveable by us.
func (suite *StatusFaveTestSuite) TestPostUnfaveable() {
	var (
		targetStatus = suite.testStatuses["local_account_1_status_3"]
		app          = suite.testApplications["application_1"]
		token        = suite.testTokens["admin_account"]
		user         = suite.testUsers["admin_account"]
		account      = suite.testAccounts["admin_account"]
	)

	out, recorder := suite.postStatusFave(
		targetStatus.ID,
		app,
		token,
		user,
		account,
	)

	// We should have 403 from
	// our call to the function.
	suite.Equal(http.StatusForbidden, recorder.Code)

	// We should get a helpful error.
	suite.Equal(`{
  "error": "Forbidden: you do not have permission to fave this status"
}`, out)
}

// Fave a status that's pending approval by us.
func (suite *StatusFaveTestSuite) TestPostFaveImplicitAccept() {
	var (
		targetStatus = suite.testStatuses["admin_account_status_5"]
		app          = suite.testApplications["application_1"]
		token        = suite.testTokens["local_account_2"]
		user         = suite.testUsers["local_account_2"]
		account      = suite.testAccounts["local_account_2"]
	)

	out, recorder := suite.postStatusFave(
		targetStatus.ID,
		app,
		token,
		user,
		account,
	)

	// We should have OK from
	// our call to the function.
	suite.Equal(http.StatusOK, recorder.Code)

	// Target status should now
	// be "favourited" by us.
	suite.Equal(`{
  "account": "yeah this is my account, what about it punk",
  "application": {
    "name": "superseriousbusiness",
    "website": "https://superserious.business"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>Hi <span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span>, can I reply?</p>",
  "created_at": "right the hell just now babyee",
  "emojis": [],
  "favourited": true,
  "favourites_count": 1,
  "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
  "in_reply_to_account_id": "01F8MH5NBDF2MV7CTC4Q5128HF",
  "in_reply_to_id": "01F8MHC8VWDRBQR0N1BATDDEM5",
  "interaction_policy": {
    "can_favourite": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    },
    "can_reblog": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    },
    "can_reply": {
      "always": [
        "public",
        "me"
      ],
      "with_approval": []
    }
  },
  "language": null,
  "media_attachments": [],
  "mentions": [
    {
      "acct": "1happyturtle",
      "id": "ZZZZZZZZZZZZZZZZZZZZZZZZZZ",
      "url": "http://localhost:8080/@1happyturtle",
      "username": "1happyturtle"
    }
  ],
  "muted": false,
  "pinned": false,
  "poll": null,
  "reblog": null,
  "reblogged": false,
  "reblogs_count": 0,
  "replies_count": 0,
  "sensitive": false,
  "spoiler_text": "",
  "tags": [],
  "text": "Hi @1happyturtle, can I reply?",
  "uri": "http://localhost:8080/some/determinate/url",
  "url": "http://localhost:8080/some/determinate/url",
  "visibility": "unlisted"
}`, out)

	// Target status should no
	// longer be pending approval.
	dbStatus, err := suite.state.DB.GetStatusByID(
		context.Background(),
		targetStatus.ID,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.False(*dbStatus.PendingApproval)

	// There should be an Accept
	// stored for the target status.
	intReq, err := suite.state.DB.GetInteractionRequestByInteractionURI(
		context.Background(), targetStatus.URI,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotZero(intReq.AcceptedAt)
	suite.NotEmpty(intReq.URI)
}

func TestStatusFaveTestSuite(t *testing.T) {
	suite.Run(t, new(StatusFaveTestSuite))
}
