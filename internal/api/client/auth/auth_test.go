/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"golang.org/x/crypto/bcrypt"
)

type AuthTestSuite struct {
	suite.Suite
	oauthServer     oauth.Server
	db              db.DB
	testAccount     *gtsmodel.Account
	testApplication *gtsmodel.Application
	testUser        *gtsmodel.User
	testClient      *gtsmodel.Client
	config          *config.Config
}

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *AuthTestSuite) SetupSuite() {
	c := config.Empty()
	// we're running on localhost without https so set the protocol to http
	c.Protocol = "http"
	// just for testing
	c.Host = "localhost:8080"
	// because go tests are run within the test package directory, we need to fiddle with the templateconfig
	// basedir in a way that we wouldn't normally have to do when running the binary, in order to make
	// the templates actually load
	c.TemplateConfig.BaseDir = "../../../web/template/"
	c.DBConfig = &config.DBConfig{
		Type:            "postgres",
		Address:         "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "postgres",
		ApplicationName: "gotosocial",
	}
	suite.config = c

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		logrus.Panicf("error encrypting user pass: %s", err)
	}

	acctID := uuid.NewString()

	suite.testAccount = &gtsmodel.Account{
		ID:       acctID,
		Username: "test_user",
	}
	suite.testUser = &gtsmodel.User{
		EncryptedPassword: string(encryptedPassword),
		Email:             "user@example.org",
		AccountID:         acctID,
	}
	suite.testClient = &gtsmodel.Client{
		ID:     "a-known-client-id",
		Secret: "some-secret",
		Domain: fmt.Sprintf("%s://%s", c.Protocol, c.Host),
	}
	suite.testApplication = &gtsmodel.Application{
		Name:         "a test application",
		Website:      "https://some-application-website.com",
		RedirectURI:  "http://localhost:8080",
		ClientID:     "a-known-client-id",
		ClientSecret: "some-secret",
		Scopes:       "read",
	}
}

// SetupTest creates a postgres connection and creates the oauth_clients table before each test
func (suite *AuthTestSuite) SetupTest() {

	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	db, err := bundb.NewBunDBService(context.Background(), suite.config)
	if err != nil {
		logrus.Panicf("error creating database connection: %s", err)
	}

	suite.db = db
	suite.oauthServer = oauth.New(context.Background(), suite.db)

	if err := suite.db.Put(context.Background(), suite.testAccount); err != nil {
		logrus.Panicf("could not insert test account into db: %s", err)
	}
	if err := suite.db.Put(context.Background(), suite.testUser); err != nil {
		logrus.Panicf("could not insert test user into db: %s", err)
	}
	if err := suite.db.Put(context.Background(), suite.testClient); err != nil {
		logrus.Panicf("could not insert test client into db: %s", err)
	}
	if err := suite.db.Put(context.Background(), suite.testApplication); err != nil {
		logrus.Panicf("could not insert test application into db: %s", err)
	}

}

// TearDownTest drops the oauth_clients table and closes the pg connection after each test
func (suite *AuthTestSuite) TearDownTest() {
	models := []interface{}{
		&gtsmodel.Client{},
		&gtsmodel.Token{},
		&gtsmodel.User{},
		&gtsmodel.Account{},
		&gtsmodel.Application{},
	}
	for _, m := range models {
		if err := suite.db.DropTable(context.Background(), m); err != nil {
			logrus.Panicf("error dropping table: %s", err)
		}
	}
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
	suite.db = nil
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
