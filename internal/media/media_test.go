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
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

type MediaTestSuite struct {
	suite.Suite
	config       *config.Config
	log          *logrus.Logger
	db           db.DB
	mediaHandler *mediaHandler
	mockStorage  storage.Storage
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
	suite.config.MediaConfig.MaxImageSize = 2 << 20 // 2 megabits

	// use an actual database for this, because it's just easier than mocking one out
	database, err := db.New(context.Background(), c, log)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.db = database

	suite.mockStorage = &storage.MockStorage{}

	// and finally here's the thing we're actually testing!
	suite.mediaHandler = &mediaHandler{
		config:  suite.config,
		db:      suite.db,
		storage: &storage.MockStorage{},
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
		&model.Account{},
		&model.MediaAttachment{},
	}
	for _, m := range models {
		if err := suite.db.CreateTable(m); err != nil {
			logrus.Panicf("db connection error: %s", err)
		}
	}
}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *MediaTestSuite) TearDownTest() {

	// remove all the tables we might have used so it's clear for the next test
	models := []interface{}{
		&model.Account{},
		&model.MediaAttachment{},
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

func TestMediaTestSuite(t *testing.T) {
	suite.Run(t, new(MediaTestSuite))
}
