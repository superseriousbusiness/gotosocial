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

package bundb_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type AccountTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *AccountTestSuite) TestGetAccountStatuses() {
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 20, false, false, "", "", false, false, false)
	suite.NoError(err)
	suite.Len(statuses, 5)
}

func (suite *AccountTestSuite) TestGetAccountStatusesExcludeRepliesAndReblogs() {
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 20, true, true, "", "", false, false, false)
	suite.NoError(err)
	suite.Len(statuses, 5)
}

func (suite *AccountTestSuite) TestGetAccountStatusesExcludeRepliesAndReblogsPublicOnly() {
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 20, true, true, "", "", false, false, true)
	suite.NoError(err)
	suite.Len(statuses, 1)
}

func (suite *AccountTestSuite) TestGetAccountStatusesMediaOnly() {
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 20, false, false, "", "", false, true, false)
	suite.NoError(err)
	suite.Len(statuses, 1)
}

func (suite *AccountTestSuite) TestGetAccountByIDWithExtras() {
	account, err := suite.db.GetAccountByID(context.Background(), suite.testAccounts["local_account_1"].ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(account)
	suite.NotNil(account.AvatarMediaAttachment)
	suite.NotEmpty(account.AvatarMediaAttachment.URL)
	suite.NotNil(account.HeaderMediaAttachment)
	suite.NotEmpty(account.HeaderMediaAttachment.URL)
}

func (suite *AccountTestSuite) TestGetAccountByUsernameDomain() {
	testAccount1 := suite.testAccounts["local_account_1"]
	account1, err := suite.db.GetAccountByUsernameDomain(context.Background(), testAccount1.Username, testAccount1.Domain)
	suite.NoError(err)
	suite.NotNil(account1)

	testAccount2 := suite.testAccounts["remote_account_1"]
	account2, err := suite.db.GetAccountByUsernameDomain(context.Background(), testAccount2.Username, testAccount2.Domain)
	suite.NoError(err)
	suite.NotNil(account2)
}

func (suite *AccountTestSuite) TestUpdateAccount() {
	ctx := context.Background()

	testAccount := suite.testAccounts["local_account_1"]

	testAccount.DisplayName = "new display name!"
	testAccount.EmojiIDs = []string{"01GD36ZKWTKY3T1JJ24JR7KY1Q", "01GD36ZV904SHBHNAYV6DX5QEF"}

	err := suite.db.UpdateAccount(ctx, testAccount)
	suite.NoError(err)

	updated, err := suite.db.GetAccountByID(ctx, testAccount.ID)
	suite.NoError(err)
	suite.Equal("new display name!", updated.DisplayName)
	suite.Equal([]string{"01GD36ZKWTKY3T1JJ24JR7KY1Q", "01GD36ZV904SHBHNAYV6DX5QEF"}, updated.EmojiIDs)
	suite.WithinDuration(time.Now(), updated.UpdatedAt, 5*time.Second)

	// get account without cache + make sure it's really in the db as desired
	dbService, ok := suite.db.(*bundb.DBService)
	if !ok {
		panic("db was not *bundb.DBService")
	}

	noCache := &gtsmodel.Account{}
	err = dbService.GetConn().
		NewSelect().
		Model(noCache).
		Where("? = ?", bun.Ident("account.id"), testAccount.ID).
		Relation("AvatarMediaAttachment").
		Relation("HeaderMediaAttachment").
		Relation("Emojis").
		Scan(ctx)

	suite.NoError(err)
	suite.Equal("new display name!", noCache.DisplayName)
	suite.Equal([]string{"01GD36ZKWTKY3T1JJ24JR7KY1Q", "01GD36ZV904SHBHNAYV6DX5QEF"}, noCache.EmojiIDs)
	suite.WithinDuration(time.Now(), noCache.UpdatedAt, 5*time.Second)
	suite.NotNil(noCache.AvatarMediaAttachment)
	suite.NotNil(noCache.HeaderMediaAttachment)

	// update again to remove emoji associations
	testAccount.EmojiIDs = []string{}

	err = suite.db.UpdateAccount(ctx, testAccount)
	suite.NoError(err)

	updated, err = suite.db.GetAccountByID(ctx, testAccount.ID)
	suite.NoError(err)
	suite.Equal("new display name!", updated.DisplayName)
	suite.Empty(updated.EmojiIDs)
	suite.WithinDuration(time.Now(), updated.UpdatedAt, 5*time.Second)

	err = dbService.GetConn().
		NewSelect().
		Model(noCache).
		Where("? = ?", bun.Ident("account.id"), testAccount.ID).
		Relation("AvatarMediaAttachment").
		Relation("HeaderMediaAttachment").
		Relation("Emojis").
		Scan(ctx)

	suite.NoError(err)
	suite.Equal("new display name!", noCache.DisplayName)
	suite.Empty(noCache.EmojiIDs)
	suite.WithinDuration(time.Now(), noCache.UpdatedAt, 5*time.Second)
}

func (suite *AccountTestSuite) TestGetAccountLastPosted() {
	lastPosted, err := suite.db.GetAccountLastPosted(context.Background(), suite.testAccounts["local_account_1"].ID, false)
	suite.NoError(err)
	suite.EqualValues(1653046675, lastPosted.Unix())
}

func (suite *AccountTestSuite) TestGetAccountLastPostedWebOnly() {
	lastPosted, err := suite.db.GetAccountLastPosted(context.Background(), suite.testAccounts["local_account_1"].ID, true)
	suite.NoError(err)
	suite.EqualValues(1634726437, lastPosted.Unix())
}

func (suite *AccountTestSuite) TestInsertAccountWithDefaults() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	suite.NoError(err)

	newAccount := &gtsmodel.Account{
		ID:           "01FGP5P4VJ9SPFB0T3E36Q60DW",
		Username:     "test_service",
		Domain:       "example.org",
		URI:          "https://example.org/users/test_service",
		URL:          "https://example.org/@test_service",
		ActorType:    ap.ActorService,
		PublicKey:    &key.PublicKey,
		PublicKeyURI: "https://example.org/users/test_service#main-key",
	}

	err = suite.db.Put(context.Background(), newAccount)
	suite.NoError(err)

	suite.Equal("en", newAccount.Language)
	suite.WithinDuration(time.Now(), newAccount.CreatedAt, 30*time.Second)
	suite.WithinDuration(time.Now(), newAccount.UpdatedAt, 30*time.Second)
	suite.False(*newAccount.Memorial)
	suite.False(*newAccount.Bot)
	suite.False(*newAccount.Discoverable)
	suite.False(*newAccount.Sensitive)
	suite.False(*newAccount.HideCollections)
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
