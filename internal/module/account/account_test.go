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

package account

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/gotosocial/gotosocial/internal/db"
	"github.com/gotosocial/gotosocial/internal/db/model"
	"github.com/gotosocial/gotosocial/internal/oauth"
	"github.com/gotosocial/oauth2/v4"
	oauthmodels "github.com/gotosocial/oauth2/v4/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type AccountTestSuite struct {
	suite.Suite
	log               *logrus.Logger
	testAccountLocal  *model.Account
	testAccountRemote *model.Account
	testUser          *model.User
	testApplication   *model.Application
	testToken         oauth2.TokenInfo
	db                db.DB
	accountModule     *accountModule
}

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *AccountTestSuite) SetupSuite() {
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	suite.log = log

	c := config.Empty()
	c.DBConfig = &config.DBConfig{
		Type:            "postgres",
		Address:         "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "postgres",
		ApplicationName: "gotosocial",
	}
	c.AccountsConfig = &config.AccountsConfig{
		OpenRegistration: true,
		RequireApproval:  true,
		ReasonRequired:   true,
	}

	database, err := db.New(context.Background(), c, log)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.db = database

	suite.accountModule = &accountModule{
		config: c,
		db:     database,
		log:    log,
	}

	suite.testApplication = &model.Application{
		ID:           "weeweeeeeeeeeeeeee",
		Name:         "a test application",
		Website:      "https://some-application-website.com",
		RedirectURI:  "http://localhost:8080",
		ClientID:     "a-known-client-id",
		ClientSecret: "some-secret",
		Scopes:       "read",
		VapidKey:     "aaaaaa-aaaaaaaa-aaaaaaaaaaa",
	}

	suite.testToken = &oauthmodels.Token{
		ClientID:      "a-known-client-id",
		RedirectURI:   "http://localhost:8080",
		Scope:         "read",
		Code:          "123456789",
		CodeCreateAt:  time.Now(),
		CodeExpiresIn: time.Duration(10 * time.Minute),
	}

	// encryptedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	// if err != nil {
	// 	logrus.Panicf("error encrypting user pass: %s", err)
	// }

	// localAvatar, err := url.Parse("https://localhost:8080/media/aaaaaaaaa.png")
	// if err != nil {
	// 	logrus.Panicf("error parsing localavatar url: %s", err)
	// }
	// localHeader, err := url.Parse("https://localhost:8080/media/ffffffffff.png")
	// if err != nil {
	// 	logrus.Panicf("error parsing localheader url: %s", err)
	// }

	// acctID := uuid.NewString()
	// suite.testAccountLocal = &model.Account{
	// 	ID:              acctID,
	// 	Username:        "local_account_of_some_kind",
	// 	AvatarRemoteURL: localAvatar,
	// 	HeaderRemoteURL: localHeader,
	// 	DisplayName:     "michael caine",
	// 	Fields: []model.Field{
	// 		{
	// 			Name:  "come and ave a go",
	// 			Value: "if you think you're hard enough",
	// 		},
	// 		{
	// 			Name:       "website",
	// 			Value:      "https://imdb.com",
	// 			VerifiedAt: time.Now(),
	// 		},
	// 	},
	// 	Note:         "My name is Michael Caine and i'm a local user.",
	// 	Discoverable: true,
	// }

	// avatarURL, err := url.Parse("http://example.org/accounts/avatars/000/207/122/original/089-1098-09.png")
	// if err != nil {
	// 	logrus.Panicf("error parsing avatarURL: %s", err)
	// }

	// headerURL, err := url.Parse("http://example.org/accounts/headers/000/207/122/original/111111111111.png")
	// if err != nil {
	// 	logrus.Panicf("error parsing avatarURL: %s", err)
	// }
	// suite.testAccountRemote = &model.Account{
	// 	ID:       uuid.NewString(),
	// 	Username: "neato_bombeato",
	// 	Domain:   "example.org",

	// 	AvatarFileName:    "avatar.png",
	// 	AvatarContentType: "image/png",
	// 	AvatarFileSize:    1024,
	// 	AvatarUpdatedAt:   time.Now(),
	// 	AvatarRemoteURL:   avatarURL,

	// 	HeaderFileName:    "avatar.png",
	// 	HeaderContentType: "image/png",
	// 	HeaderFileSize:    1024,
	// 	HeaderUpdatedAt:   time.Now(),
	// 	HeaderRemoteURL:   headerURL,

	// 	DisplayName: "one cool dude 420",
	// 	Fields: []model.Field{
	// 		{
	// 			Name:  "pronouns",
	// 			Value: "he/they",
	// 		},
	// 		{
	// 			Name:       "website",
	// 			Value:      "https://imcool.edu",
	// 			VerifiedAt: time.Now(),
	// 		},
	// 	},
	// 	Note:                  "<p>I'm cool as heck!</p>",
	// 	Discoverable:          true,
	// 	URI:                   "https://example.org/users/neato_bombeato",
	// 	URL:                   "https://example.org/@neato_bombeato",
	// 	LastWebfingeredAt:     time.Now(),
	// 	InboxURL:              "https://example.org/users/neato_bombeato/inbox",
	// 	OutboxURL:             "https://example.org/users/neato_bombeato/outbox",
	// 	SharedInboxURL:        "https://example.org/inbox",
	// 	FollowersURL:          "https://example.org/users/neato_bombeato/followers",
	// 	FeaturedCollectionURL: "https://example.org/users/neato_bombeato/collections/featured",
	// }
	// suite.testUser = &model.User{
	// 	ID:                uuid.NewString(),
	// 	EncryptedPassword: string(encryptedPassword),
	// 	Email:             "user@example.org",
	// 	AccountID:         acctID,
	// }
}

func (suite *AccountTestSuite) TearDownSuite() {
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
}

// SetupTest creates a db connection and creates necessary tables before each test
func (suite *AccountTestSuite) SetupTest() {
	models := []interface{}{
		&model.User{},
		&model.Account{},
		&model.Follow{},
		&model.Status{},
		&model.Application{},
		&model.EmailDomainBlock{},
	}

	for _, m := range models {
		if err := suite.db.CreateTable(m); err != nil {
			logrus.Panicf("db connection error: %s", err)
		}
	}
}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *AccountTestSuite) TearDownTest() {
	models := []interface{}{
		&model.User{},
		&model.Account{},
		&model.Follow{},
		&model.Status{},
		&model.Application{},
		&model.EmailDomainBlock{},
	}
	for _, m := range models {
		if err := suite.db.DropTable(m); err != nil {
			logrus.Panicf("error dropping table: %s", err)
		}
	}
}

func (suite *AccountTestSuite) TestAccountCreatePOSTHandler() {
	// TODO: figure out how to test this properly
	recorder := httptest.NewRecorder()
	recorder.Header().Set("X-Forwarded-For", "127.0.0.1")
	recorder.Header().Set("Content-Type", "application/json")
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplication)
	ctx.Set(oauth.SessionAuthorizedToken, suite.testToken)
	ctx.Request = httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/accounts", nil)
	ctx.Request.Form = url.Values{
		"reason":    []string{"a very good reason that's at least 40 characters i swear"},
		"username":  []string{"test_user"},
		"email":     []string{"user@example.org"},
		"password":  []string{"very-strong-password"},
		"agreement": []string{"true"},
		"locale":    []string{"en"},
	}
	suite.accountModule.accountCreatePOSTHandler(ctx)
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
