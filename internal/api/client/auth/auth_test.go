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
	"net/http/httptest"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/auth"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AuthStandardTestSuite struct {
	suite.Suite
	db          db.DB
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

const (
	sessionUserID   = "userid"
	sessionClientID = "client_id"
)

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

	suite.oauthServer = testrig.NewTestOauthServer(suite.db)
	var err error
	suite.idp, err = oidc.NewIDP(context.Background())
	if err != nil {
		panic(err)
	}
	suite.authModule = auth.New(suite.db, suite.oauthServer, suite.idp).(*auth.Module)
	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *AuthStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *AuthStandardTestSuite) newContext(requestMethod string, requestPath string) (*gin.Context, *httptest.ResponseRecorder) {
	// create the recorder and gin test context
	recorder := httptest.NewRecorder()
	ctx, engine := gin.CreateTestContext(recorder)

	// load templates into the engine
	testrig.ConfigureTemplatesWithGin(engine)

	// create the request
	protocol := viper.GetString(config.Keys.Protocol)
	host := viper.GetString(config.Keys.Host)
	baseURI := fmt.Sprintf("%s://%s", protocol, host)
	requestURI := fmt.Sprintf("%s/%s", baseURI, requestPath)
	ctx.Request = httptest.NewRequest(requestMethod, requestURI, nil) // the endpoint we're hitting
	ctx.Request.Header.Set("accept", "text/html")

	// trigger the session middleware on the context
	store := memstore.NewStore(make([]byte, 32), make([]byte, 32))
	store.Options(router.SessionOptions())
	sessionMiddleware := sessions.Sessions("gotosocial-localhost", store)
	sessionMiddleware(ctx)

	return ctx, recorder
}
