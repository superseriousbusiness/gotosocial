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

package webfinger_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/api/wellknown/webfinger"
	"github.com/superseriousbusiness/gotosocial/internal/cleaner"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/filter/interaction"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/subscriptions"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type WebfingerGetTestSuite struct {
	WebfingerStandardTestSuite
}

func (suite *WebfingerGetTestSuite) finger(requestPath string) string {
	// Set up the request.
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodGet, requestPath, nil)
	ctx.Request.Header.Set("accept", "application/jrd+json")

	// Trigger the handler.
	suite.webfingerModule.WebfingerGETRequest(ctx)

	// Read the result + return it
	// as nicely indented JSON.
	result := recorder.Result()
	defer result.Body.Close()

	// Result should always use the
	// webfinger content-type.
	if ct := result.Header.Get("content-type"); ct != string(apiutil.AppJRDJSON) {
		suite.FailNow("", "expected content type %s, got %s", apiutil.AppJRDJSON, ct)
	}

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	var dst bytes.Buffer
	if err := json.Indent(&dst, b, "", "  "); err != nil {
		suite.FailNow(err.Error())
	}

	return dst.String()
}

func (suite *WebfingerGetTestSuite) funkifyAccountDomain(host string, accountDomain string) *gtsmodel.Account {
	// Reset suite structs + config
	// to new host + account domain.
	config.SetHost(host)
	config.SetAccountDomain(accountDomain)
	testrig.StopWorkers(&suite.state)
	testrig.StartNoopWorkers(&suite.state)

	suite.processor = processing.NewProcessor(
		cleaner.New(&suite.state),
		subscriptions.New(&suite.state, suite.federator.TransportController(), suite.tc),
		suite.tc,
		suite.federator,
		testrig.NewTestOauthServer(suite.db),
		testrig.NewTestMediaManager(&suite.state),
		&suite.state,
		suite.emailSender,
		testrig.NewNoopWebPushSender(),
		visibility.NewFilter(&suite.state),
		interaction.NewFilter(&suite.state),
	)

	suite.webfingerModule = webfinger.New(suite.processor)
	testrig.StartNoopWorkers(&suite.state)

	// Generate a new account for the
	// tester, which uses the new host.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	publicKey := &privateKey.PublicKey

	targetAccount := &gtsmodel.Account{
		ID:                    "01FG1K8EA7SYHEC7V6XKVNC4ZA",
		Username:              "new_account_domain_user",
		URI:                   "http://" + host + "/users/new_account_domain_user",
		URL:                   "http://" + host + "/@new_account_domain_user",
		InboxURI:              "http://" + host + "/users/new_account_domain_user/inbox",
		OutboxURI:             "http://" + host + "/users/new_account_domain_user/outbox",
		FollowingURI:          "http://" + host + "/users/new_account_domain_user/following",
		FollowersURI:          "http://" + host + "/users/new_account_domain_user/followers",
		FeaturedCollectionURI: "http://" + host + "/users/new_account_domain_user/collections/featured",
		ActorType:             ap.ActorPerson,
		PrivateKey:            privateKey,
		PublicKey:             publicKey,
		PublicKeyURI:          "http://" + host + "/users/new_account_domain_user/main-key",
	}

	if err := suite.db.PutAccount(context.Background(), targetAccount); err != nil {
		suite.FailNow(err.Error())
	}

	if err := suite.db.PutAccountSettings(context.Background(), &gtsmodel.AccountSettings{AccountID: targetAccount.ID}); err != nil {
		suite.FailNow(err.Error())
	}

	return targetAccount
}

func (suite *WebfingerGetTestSuite) TestFingerUser() {
	targetAccount := suite.testAccounts["local_account_1"]
	requestPath := fmt.Sprintf("/%s?resource=acct:%s@%s", webfinger.WebfingerBasePath, targetAccount.Username, config.GetHost())

	resp := suite.finger(requestPath)
	suite.Equal(`{
  "subject": "acct:the_mighty_zork@localhost:8080",
  "aliases": [
    "http://localhost:8080/users/the_mighty_zork",
    "http://localhost:8080/@the_mighty_zork"
  ],
  "links": [
    {
      "rel": "http://webfinger.net/rel/profile-page",
      "type": "text/html",
      "href": "http://localhost:8080/@the_mighty_zork"
    },
    {
      "rel": "self",
      "type": "application/activity+json",
      "href": "http://localhost:8080/users/the_mighty_zork"
    }
  ]
}`, resp)
}

func (suite *WebfingerGetTestSuite) TestFingerUserActorURI() {
	targetAccount := suite.testAccounts["local_account_1"]
	host := config.GetHost()

	tests := []struct {
		resource string
	}{
		{resource: fmt.Sprintf("https://%s/@%s", host, targetAccount.Username)},
		{resource: fmt.Sprintf("https://%s/users/%s", host, targetAccount.Username)},
	}

	for _, tt := range tests {
		tt := tt
		suite.Run(tt.resource, func() {
			requestPath := fmt.Sprintf("/%s?resource=%s", webfinger.WebfingerBasePath, tt.resource)
			resp := suite.finger(requestPath)
			suite.Equal(`{
  "subject": "acct:the_mighty_zork@localhost:8080",
  "aliases": [
    "http://localhost:8080/users/the_mighty_zork",
    "http://localhost:8080/@the_mighty_zork"
  ],
  "links": [
    {
      "rel": "http://webfinger.net/rel/profile-page",
      "type": "text/html",
      "href": "http://localhost:8080/@the_mighty_zork"
    },
    {
      "rel": "self",
      "type": "application/activity+json",
      "href": "http://localhost:8080/users/the_mighty_zork"
    }
  ]
}`, resp)
		})
	}
}

func (suite *WebfingerGetTestSuite) TestFingerUserWithDifferentAccountDomainByHost() {
	targetAccount := suite.funkifyAccountDomain("gts.example.org", "example.org")
	requestPath := fmt.Sprintf("/%s?resource=acct:%s@%s", webfinger.WebfingerBasePath, targetAccount.Username, config.GetHost())

	resp := suite.finger(requestPath)
	suite.Equal(`{
  "subject": "acct:new_account_domain_user@example.org",
  "aliases": [
    "http://gts.example.org/users/new_account_domain_user",
    "http://gts.example.org/@new_account_domain_user"
  ],
  "links": [
    {
      "rel": "http://webfinger.net/rel/profile-page",
      "type": "text/html",
      "href": "http://gts.example.org/@new_account_domain_user"
    },
    {
      "rel": "self",
      "type": "application/activity+json",
      "href": "http://gts.example.org/users/new_account_domain_user"
    }
  ]
}`, resp)
}

func (suite *WebfingerGetTestSuite) TestFingerUserWithDifferentAccountDomainByAccountDomain() {
	targetAccount := suite.funkifyAccountDomain("gts.example.org", "example.org")
	requestPath := fmt.Sprintf("/%s?resource=acct:%s@%s", webfinger.WebfingerBasePath, targetAccount.Username, config.GetAccountDomain())

	resp := suite.finger(requestPath)
	suite.Equal(`{
  "subject": "acct:new_account_domain_user@example.org",
  "aliases": [
    "http://gts.example.org/users/new_account_domain_user",
    "http://gts.example.org/@new_account_domain_user"
  ],
  "links": [
    {
      "rel": "http://webfinger.net/rel/profile-page",
      "type": "text/html",
      "href": "http://gts.example.org/@new_account_domain_user"
    },
    {
      "rel": "self",
      "type": "application/activity+json",
      "href": "http://gts.example.org/users/new_account_domain_user"
    }
  ]
}`, resp)
}

func (suite *WebfingerGetTestSuite) TestFingerUserWithoutAcct() {
	// Leave out the 'acct:' part in the request path;
	// the handler should be generous + still work OK.
	targetAccount := suite.testAccounts["local_account_1"]
	requestPath := fmt.Sprintf("/%s?resource=%s@%s", webfinger.WebfingerBasePath, targetAccount.Username, config.GetHost())

	resp := suite.finger(requestPath)
	suite.Equal(`{
  "subject": "acct:the_mighty_zork@localhost:8080",
  "aliases": [
    "http://localhost:8080/users/the_mighty_zork",
    "http://localhost:8080/@the_mighty_zork"
  ],
  "links": [
    {
      "rel": "http://webfinger.net/rel/profile-page",
      "type": "text/html",
      "href": "http://localhost:8080/@the_mighty_zork"
    },
    {
      "rel": "self",
      "type": "application/activity+json",
      "href": "http://localhost:8080/users/the_mighty_zork"
    }
  ]
}`, resp)
}

func TestWebfingerGetTestSuite(t *testing.T) {
	suite.Run(t, new(WebfingerGetTestSuite))
}
