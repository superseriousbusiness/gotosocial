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

package media_test

import (
	"testing"

	"codeberg.org/gruf/go-store/kv"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type MediaStandardTestSuite struct {
	suite.Suite

	db      db.DB
	storage *kv.KVStore
	manager media.Manager
}

func (suite *MediaStandardTestSuite) SetupSuite() {
	testrig.InitTestLog()
	testrig.InitTestConfig()

	suite.db = testrig.NewTestDB()
	suite.storage = testrig.NewTestStorage()
}

func (suite *MediaStandardTestSuite) SetupTest() {
	testrig.StandardStorageSetup(suite.storage, "../../testrig/media")
	testrig.StandardDBSetup(suite.db, nil)

	m, err := media.New(suite.db, suite.storage)
	if err != nil {
		panic(err)
	}
	suite.manager = m
}

func (suite *MediaStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
}

func TestMediaStandardTestSuite(t *testing.T) {
	suite.Run(t, &MediaStandardTestSuite{})
}
