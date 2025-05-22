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

package trans_test

import (
	"fmt"
	"os"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/trans"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ImportMinimalTestSuite struct {
	TransTestSuite
}

func (suite *ImportMinimalTestSuite) TestImportMinimalOK() {
	ctx := suite.T().Context()

	testAccountBefore, err := suite.db.GetAccountByID(ctx, suite.testAccounts["local_account_1"].ID)
	if err != nil {
		suite.FailNow("couldn't get testAccountBefore")
	}

	// use a temporary file path
	tempFilePath := fmt.Sprintf("%s/%s", suite.T().TempDir(), uuid.NewString())

	// export to the tempFilePath
	exporter := trans.NewExporter(suite.db)
	err = exporter.ExportMinimal(ctx, tempFilePath)
	suite.NoError(err)

	// we should have some bytes in that file now
	b, err := os.ReadFile(tempFilePath)
	suite.NoError(err)
	suite.NotEmpty(b)
	fmt.Println(string(b))

	var state state.State
	state.Caches.Init()

	// create a new database with just the tables created, no entries
	newDB := testrig.NewTestDB(&state)

	importer := trans.NewImporter(newDB)
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

	// we should have some follows in the database
	follows := []*gtsmodel.Follow{}
	err = newDB.GetAll(ctx, &follows)
	suite.NoError(err)
	suite.NotEmpty(follows)

	// we should have some domain blocks in the database
	domainBlocks := []*gtsmodel.DomainBlock{}
	err = newDB.GetAll(ctx, &domainBlocks)
	suite.NoError(err)
	suite.NotEmpty(domainBlocks)

	// compare test account before + after
	testAccountAfter, err := newDB.GetAccountByID(ctx, suite.testAccounts["local_account_1"].ID)
	if err != nil {
		suite.FailNow("couldn't get testAccountAfter")
	}

	suite.Equal(testAccountBefore.ID, testAccountAfter.ID)
	suite.Equal(testAccountBefore.Username, testAccountAfter.Username)
	suite.Equal(testAccountBefore.Domain, testAccountAfter.Domain)
	suite.Equal(testAccountBefore.DisplayName, testAccountAfter.DisplayName)
	suite.Equal(testAccountBefore.Note, testAccountAfter.Note)
	suite.Equal(testAccountBefore.NoteRaw, testAccountAfter.NoteRaw)
	suite.Equal(testAccountBefore.MemorializedAt, testAccountAfter.MemorializedAt)
	suite.Equal(testAccountBefore.Locked, testAccountAfter.Locked)
	suite.Equal(testAccountBefore.URI, testAccountAfter.URI)
	suite.Equal(testAccountBefore.URL, testAccountAfter.URL)
	suite.Equal(testAccountBefore.InboxURI, testAccountAfter.InboxURI)
	suite.Equal(testAccountBefore.OutboxURI, testAccountAfter.OutboxURI)
	suite.Equal(testAccountBefore.FollowingURI, testAccountAfter.FollowingURI)
	suite.Equal(testAccountBefore.FollowersURI, testAccountAfter.FollowersURI)
	suite.Equal(testAccountBefore.FeaturedCollectionURI, testAccountAfter.FeaturedCollectionURI)
	suite.Equal(testAccountBefore.ActorType, testAccountAfter.ActorType)
	suite.Equal(testAccountBefore.PrivateKey, testAccountAfter.PrivateKey)
	suite.Equal(testAccountBefore.PublicKey, testAccountAfter.PublicKey)
	suite.Equal(testAccountBefore.PublicKeyURI, testAccountAfter.PublicKeyURI)
	suite.Equal(testAccountBefore.SuspendedAt, testAccountAfter.SuspendedAt)
	suite.Equal(testAccountBefore.SuspensionOrigin, testAccountAfter.SuspensionOrigin)
}

func TestImportMinimalTestSuite(t *testing.T) {
	suite.Run(t, &ImportMinimalTestSuite{})
}
