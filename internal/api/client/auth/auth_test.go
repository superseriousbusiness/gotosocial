/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/auth"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AuthStandardTestSuite struct {
	suite.Suite
	db          db.DB
	tc          typeutils.TypeConverter
	idp         oidc.IDP
	oauthServer oauth.Server

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account

	// module being tested
	authModule *auth.Module
}

func (suite *AuthStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
}

func (suite *AuthStandardTestSuite) SetupTest() {
	testrig.InitTestConfig()
	suite.db = testrig.NewTestDB()
	testrig.InitTestLog()
	// suite.sentEmails = make(map[string]string)
	// suite.emailSender = testrig.NewEmailSender("../../../../web/template/", suite.sentEmails)

	suite.oauthServer = testrig.NewTestOauthServer(suite.db)
	var err error
	suite.idp, err = oidc.NewIDP(context.Background())
	if err != nil {
		panic(err)
	}
	suite.authModule = auth.New(suite.db, suite.oauthServer, suite.idp).(*auth.Module)
	testrig.StandardDBSetup(suite.db, nil)
	//testrig.New
}

func (suite *AuthStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *AuthStandardTestSuite) newContext(recorder *httptest.ResponseRecorder, requestMethod string, requestPath string) *gin.Context {

	protocol := viper.GetString(config.Keys.Protocol)
	host := viper.GetString(config.Keys.Host)

	baseURI := fmt.Sprintf("%s://%s", protocol, host)
	requestURI := fmt.Sprintf("%s/%s", baseURI, requestPath)

	request := httptest.NewRequest(http.MethodPatch, requestURI, nil) // the endpoint we're hitting
	request.Header.Set("accept", "text/html")

	ctx, _ := testrig.CreateTestContextWithTemplatesAndSessions(recorder)
	ctx.Request = request
	//suite.testSessionsMiddleware(ctx)

	ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])

	return ctx
}
