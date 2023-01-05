/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package bundb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
)

type BundbNewTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *BundbNewTestSuite) TestCreateNewDB() {
	// create a new db with standard test settings
	db, err := bundb.NewBunDBService(context.Background(), nil)
	suite.NoError(err)
	suite.NotNil(db)
}

func (suite *BundbNewTestSuite) TestCreateNewSqliteDBNoAddress() {
	// create a new db with no address specified
	config.SetDbAddress("")
	config.SetDbType("sqlite")
	db, err := bundb.NewBunDBService(context.Background(), nil)
	suite.EqualError(err, "'db-address' was not set when attempting to start sqlite")
	suite.Nil(db)
}

func TestBundbNewTestSuite(t *testing.T) {
	suite.Run(t, new(BundbNewTestSuite))
}
