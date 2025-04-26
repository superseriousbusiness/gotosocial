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

package bundb_test

import (
	"context"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db/bundb"
	"github.com/stretchr/testify/suite"
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
