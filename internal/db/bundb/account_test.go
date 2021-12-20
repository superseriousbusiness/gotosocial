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
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type AccountTestSuite struct {
	BunDBStandardTestSuite
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

func (suite *AccountTestSuite) TestUpdateAccount() {
	testAccount := suite.testAccounts["local_account_1"]

	testAccount.DisplayName = "new display name!"

	_, err := suite.db.UpdateAccount(context.Background(), testAccount)
	suite.NoError(err)

	updated, err := suite.db.GetAccountByID(context.Background(), testAccount.ID)
	suite.NoError(err)
	suite.Equal("new display name!", updated.DisplayName)
	suite.WithinDuration(time.Now(), updated.UpdatedAt, 5*time.Second)
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
	suite.False(newAccount.Memorial)
	suite.False(newAccount.Bot)
	suite.False(newAccount.Discoverable)
	suite.False(newAccount.Sensitive)
	suite.False(newAccount.HideCollections)
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
