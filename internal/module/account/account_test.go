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
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/gotosocial/gotosocial/internal/db"
	"github.com/gotosocial/gotosocial/internal/db/model"
	"github.com/gotosocial/gotosocial/internal/module/oauth"
	"github.com/gotosocial/gotosocial/internal/router"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type AccountTestSuite struct {
	suite.Suite
	db                db.DB
	testAccountLocal  *model.Account
	testAccountRemote *model.Account
	testUser          *model.User
	config            *config.Config
}

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *AccountTestSuite) SetupSuite() {
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
	suite.config = c

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		logrus.Panicf("error encrypting user pass: %s", err)
	}

	localAvatar, err := url.Parse("https://localhost:8080/media/aaaaaaaaa.png")
	if err != nil {
		logrus.Panicf("error parsing localavatar url: %s", err)
	}
	localHeader, err := url.Parse("https://localhost:8080/media/ffffffffff.png")
	if err != nil {
		logrus.Panicf("error parsing localheader url: %s", err)
	}

	acctID := uuid.NewString()
	suite.testAccountLocal = &model.Account{
		ID:              acctID,
		Username:        "local_account_of_some_kind",
		AvatarRemoteURL: localAvatar,
		HeaderRemoteURL: localHeader,
		DisplayName:     "michael caine",
		Fields: []model.Field{
			{
				Name:  "come and ave a go",
				Value: "if you think you're hard enough",
			},
			{
				Name:       "website",
				Value:      "https://imdb.com",
				VerifiedAt: time.Now(),
			},
		},
		Note:         "My name is Michael Caine and i'm a local user.",
		Discoverable: true,
	}

	avatarURL, err := url.Parse("http://example.org/accounts/avatars/000/207/122/original/089-1098-09.png")
	if err != nil {
		logrus.Panicf("error parsing avatarURL: %s", err)
	}

	headerURL, err := url.Parse("http://example.org/accounts/headers/000/207/122/original/111111111111.png")
	if err != nil {
		logrus.Panicf("error parsing avatarURL: %s", err)
	}
	suite.testAccountRemote = &model.Account{
		ID:       uuid.NewString(),
		Username: "neato_bombeato",
		Domain:   "example.org",

		AvatarFileName:    "avatar.png",
		AvatarContentType: "image/png",
		AvatarFileSize:    1024,
		AvatarUpdatedAt:   time.Now(),
		AvatarRemoteURL:   avatarURL,

		HeaderFileName:    "avatar.png",
		HeaderContentType: "image/png",
		HeaderFileSize:    1024,
		HeaderUpdatedAt:   time.Now(),
		HeaderRemoteURL:   headerURL,

		DisplayName: "one cool dude 420",
		Fields: []model.Field{
			{
				Name:  "pronouns",
				Value: "he/they",
			},
			{
				Name:       "website",
				Value:      "https://imcool.edu",
				VerifiedAt: time.Now(),
			},
		},
		Note:                  "<p>I'm cool as heck!</p>",
		Discoverable:          true,
		URI:                   "https://example.org/users/neato_bombeato",
		URL:                   "https://example.org/@neato_bombeato",
		LastWebfingeredAt:     time.Now(),
		InboxURL:              "https://example.org/users/neato_bombeato/inbox",
		OutboxURL:             "https://example.org/users/neato_bombeato/outbox",
		SharedInboxURL:        "https://example.org/inbox",
		FollowersURL:          "https://example.org/users/neato_bombeato/followers",
		FeaturedCollectionURL: "https://example.org/users/neato_bombeato/collections/featured",
	}
	suite.testUser = &model.User{
		ID:                uuid.NewString(),
		EncryptedPassword: string(encryptedPassword),
		Email:             "user@example.org",
		AccountID:         acctID,
	}
}

// SetupTest creates a postgres connection and creates the oauth_clients table before each test
func (suite *AccountTestSuite) SetupTest() {

	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	db, err := db.New(context.Background(), suite.config, log)
	if err != nil {
		logrus.Panicf("error creating database connection: %s", err)
	}

	suite.db = db

	models := []interface{}{
		&model.User{},
		&model.Account{},
		&model.Follow{},
		&model.Status{},
	}

	for _, m := range models {
		if err := suite.db.CreateTable(m); err != nil {
			logrus.Panicf("db connection error: %s", err)
		}
	}

	if err := suite.db.Put(suite.testAccountLocal); err != nil {
		logrus.Panicf("could not insert test account into db: %s", err)
	}
	if err := suite.db.Put(suite.testUser); err != nil {
		logrus.Panicf("could not insert test user into db: %s", err)
	}

}

// TearDownTest drops the oauth_clients table and closes the pg connection after each test
func (suite *AccountTestSuite) TearDownTest() {
	models := []interface{}{
		&model.User{},
		&model.Account{},
		&model.Follow{},
		&model.Status{},
	}
	for _, m := range models {
		if err := suite.db.DropTable(m); err != nil {
			logrus.Panicf("error dropping table: %s", err)
		}
	}
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
	suite.db = nil
}

func (suite *AccountTestSuite) TestAPIInitialize() {
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)

	r, err := router.New(suite.config, log)
	if err != nil {
		suite.FailNow(fmt.Sprintf("error mapping routes onto router: %s", err))
	}

	r.AttachMiddleware(func(c *gin.Context) {
		c.Set(oauth.SessionAuthorizedUser, suite.testUser.ID)
	})

	acct := New(suite.config, suite.db)
	acct.Route(r)

	r.Start()
	defer r.Stop(context.Background())
	time.Sleep(10 * time.Second)

}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
