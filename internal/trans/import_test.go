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

package trans_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/trans"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ImportMinimalTestSuite struct {
	TransTestSuite
}

func (suite *ImportMinimalTestSuite) TestImportMinimalOK() {
	ctx := context.Background()

	// use a temporary file path
	tempFilePath := fmt.Sprintf("%s/%s", suite.T().TempDir(), uuid.NewString())

	// export to the tempFilePath
	exporter := trans.NewExporter(suite.db, suite.log)
	err := exporter.ExportMinimal(ctx, tempFilePath)
	suite.NoError(err)

	// we should have some bytes in that file now
	b, err := os.ReadFile(tempFilePath)
	suite.NoError(err)
	suite.NotEmpty(b)
	fmt.Println(string(b))

	// create a new database with just the tables created, no entries
	testrig.StandardDBTeardown(suite.db)
	newDB := testrig.NewTestDB()
	testrig.CreateTestTables(newDB)

	importer := trans.NewImporter(newDB, suite.log)
	err = importer.Import(ctx, tempFilePath)
	suite.NoError(err)

	// we should have some accounts in the database
	accounts := []*gtsmodel.Account{}
	err = newDB.GetAll(ctx, &accounts)
	suite.NoError(err)
	suite.NotEmpty(accounts)

	// we should have some blocks in the database
	blocks := []*gtsmodel.Block{}
	err = newDB.GetAll(ctx, &blocks)
	suite.NoError(err)
	suite.NotEmpty(blocks)
}

func TestImportMinimalTestSuite(t *testing.T) {
	suite.Run(t, &ImportMinimalTestSuite{})
}
