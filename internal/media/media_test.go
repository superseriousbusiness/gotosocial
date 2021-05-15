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

package media

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/pg"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

type MediaTestSuite struct {
	suite.Suite
	config       *config.Config
	log          *logrus.Logger
	db           db.DB
	mediaHandler *mediaHandler
	mockStorage  *storage.MockStorage
}

/*
	TEST INFRASTRUCTURE
*/

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *MediaTestSuite) SetupSuite() {
	// some of our subsequent entities need a log so create this here
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	suite.log = log

	// Direct config to local postgres instance
	c := config.Empty()
	c.Protocol = "http"
	c.Host = "localhost"
	c.DBConfig = &config.DBConfig{
		Type:            "postgres",
		Address:         "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "postgres",
		ApplicationName: "gotosocial",
	}
	c.MediaConfig = &config.MediaConfig{
		MaxImageSize: 2 << 20,
	}
	c.StorageConfig = &config.StorageConfig{
		Backend:       "local",
		BasePath:      "/tmp",
		ServeProtocol: "http",
		ServeHost:     "localhost",
		ServeBasePath: "/fileserver/media",
	}
	suite.config = c
	// use an actual database for this, because it's just easier than mocking one out
	database, err := pg.NewPostgresService(context.Background(), c, log)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.db = database

	suite.mockStorage = &storage.MockStorage{}
	// We don't need storage to do anything for these tests, so just simulate a success and do nothing
	suite.mockStorage.On("StoreFileAt", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	// and finally here's the thing we're actually testing!
	suite.mediaHandler = &mediaHandler{
		config:  suite.config,
		db:      suite.db,
		storage: suite.mockStorage,
		log:     log,
	}
}

func (suite *MediaTestSuite) TearDownSuite() {
	if err := suite.db.Stop(context.Background()); err != nil {
		logrus.Panicf("error closing db connection: %s", err)
	}
}

// SetupTest creates a db connection and creates necessary tables before each test
func (suite *MediaTestSuite) SetupTest() {
	// create all the tables we might need in thie suite
	models := []interface{}{
		&gtsmodel.Account{},
		&gtsmodel.MediaAttachment{},
	}
	for _, m := range models {
		if err := suite.db.CreateTable(m); err != nil {
			logrus.Panicf("db connection error: %s", err)
		}
	}

	err := suite.db.CreateInstanceAccount()
	if err != nil {
		logrus.Panic(err)
	}
}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *MediaTestSuite) TearDownTest() {

	// remove all the tables we might have used so it's clear for the next test
	models := []interface{}{
		&gtsmodel.Account{},
		&gtsmodel.MediaAttachment{},
	}
	for _, m := range models {
		if err := suite.db.DropTable(m); err != nil {
			logrus.Panicf("error dropping table: %s", err)
		}
	}
}

/*
	ACTUAL TESTS
*/

func (suite *MediaTestSuite) TestSetHeaderOrAvatarForAccountID() {
	// load test image
	f, err := ioutil.ReadFile("./test/test-jpeg.jpg")
	assert.Nil(suite.T(), err)

	ma, err := suite.mediaHandler.ProcessHeaderOrAvatar(f, "weeeeeee", "header")
	assert.Nil(suite.T(), err)
	suite.log.Debugf("%+v", ma)

	// attachment should have....
	assert.Equal(suite.T(), "weeeeeee", ma.AccountID)
	assert.Equal(suite.T(), "LjCZnlvyRkRn_NvzRjWF?urqV@f9", ma.Blurhash)
	//TODO: add more checks here, cba right now!
}

func (suite *MediaTestSuite) TestProcessLocalEmoji() {
	f, err := ioutil.ReadFile("./test/rainbow-original.png")
	assert.NoError(suite.T(), err)

	emoji, err := suite.mediaHandler.ProcessLocalEmoji(f, "rainbow")
	assert.NoError(suite.T(), err)
	suite.log.Debugf("%+v", emoji)
}

// TODO: add tests for sad path, gif, png....

func TestMediaTestSuite(t *testing.T) {
	suite.Run(t, new(MediaTestSuite))
}
