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
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type AccountTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *AccountTestSuite) TestGetAccountStatuses() {
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 20, false, false, "", "", false, false)
	suite.NoError(err)
	suite.Len(statuses, 5)
}

func (suite *AccountTestSuite) TestGetAccountStatusesExcludeRepliesAndReblogs() {
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 20, true, true, "", "", false, false)
	suite.NoError(err)
	suite.Len(statuses, 5)
}

func (suite *AccountTestSuite) TestGetAccountStatusesExcludeRepliesAndReblogsPublicOnly() {
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 20, true, true, "", "", false, true)
	suite.NoError(err)
	suite.Len(statuses, 1)
}

func (suite *AccountTestSuite) TestGetAccountStatusesMediaOnly() {
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 20, false, false, "", "", true, false)
	suite.NoError(err)
	suite.Len(statuses, 1)
}

func (suite *AccountTestSuite) TestGetAccountBy() {
	t := suite.T()

	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Sentinel error to mark avoiding a test case.
	sentinelErr := errors.New("sentinel")

	// isEqual checks if 2 account models are equal.
	isEqual := func(a1, a2 gtsmodel.Account) bool {
		// Clear populated sub-models.
		a1.HeaderMediaAttachment = nil
		a2.HeaderMediaAttachment = nil
		a1.AvatarMediaAttachment = nil
		a2.AvatarMediaAttachment = nil
		a1.Emojis = nil
		a2.Emojis = nil

		// Clear database-set fields.
		a1.CreatedAt = time.Time{}
		a2.CreatedAt = time.Time{}
		a1.UpdatedAt = time.Time{}
		a2.UpdatedAt = time.Time{}

		// Manually compare keys.
		pk1 := a1.PublicKey
		pv1 := a1.PrivateKey
		pk2 := a2.PublicKey
		pv2 := a2.PrivateKey
		a1.PublicKey = nil
		a1.PrivateKey = nil
		a2.PublicKey = nil
		a2.PrivateKey = nil

		return reflect.DeepEqual(a1, a2) &&
			((pk1 == nil && pk2 == nil) || pk1.Equal(pk2)) &&
			((pv1 == nil && pv2 == nil) || pv1.Equal(pv2))
	}

	for _, account := range suite.testAccounts {
		for lookup, dbfunc := range map[string]func() (*gtsmodel.Account, error){
			"id": func() (*gtsmodel.Account, error) {
				return suite.db.GetAccountByID(ctx, account.ID)
			},

			"uri": func() (*gtsmodel.Account, error) {
				return suite.db.GetAccountByURI(ctx, account.URI)
			},

			"url": func() (*gtsmodel.Account, error) {
				if account.URL == "" {
					return nil, sentinelErr
				}
				return suite.db.GetAccountByURL(ctx, account.URL)
			},

			"username@domain": func() (*gtsmodel.Account, error) {
				return suite.db.GetAccountByUsernameDomain(ctx, account.Username, account.Domain)
			},

			"username_upper@domain": func() (*gtsmodel.Account, error) {
				return suite.db.GetAccountByUsernameDomain(ctx, strings.ToUpper(account.Username), account.Domain)
			},

			"username_lower@domain": func() (*gtsmodel.Account, error) {
				return suite.db.GetAccountByUsernameDomain(ctx, strings.ToLower(account.Username), account.Domain)
			},

			"public_key_uri": func() (*gtsmodel.Account, error) {
				if account.PublicKeyURI == "" {
					return nil, sentinelErr
				}
				return suite.db.GetAccountByPubkeyID(ctx, account.PublicKeyURI)
			},

			"inbox_uri": func() (*gtsmodel.Account, error) {
				if account.InboxURI == "" {
					return nil, sentinelErr
				}
				return suite.db.GetAccountByInboxURI(ctx, account.InboxURI)
			},

			"outbox_uri": func() (*gtsmodel.Account, error) {
				if account.OutboxURI == "" {
					return nil, sentinelErr
				}
				return suite.db.GetAccountByOutboxURI(ctx, account.OutboxURI)
			},

			"following_uri": func() (*gtsmodel.Account, error) {
				if account.FollowingURI == "" {
					return nil, sentinelErr
				}
				return suite.db.GetAccountByFollowingURI(ctx, account.FollowingURI)
			},

			"followers_uri": func() (*gtsmodel.Account, error) {
				if account.FollowersURI == "" {
					return nil, sentinelErr
				}
				return suite.db.GetAccountByFollowersURI(ctx, account.FollowersURI)
			},
		} {

			// Clear database caches.
			suite.state.Caches.Init()

			t.Logf("checking database lookup %q", lookup)

			// Perform database function.
			checkAcc, err := dbfunc()
			if err != nil {
				if err == sentinelErr {
					continue
				}

				t.Errorf("error encountered for database lookup %q: %v", lookup, err)
				continue
			}

			// Check received account data.
			if !isEqual(*checkAcc, *account) {
				t.Errorf("account does not contain expected data: %+v", checkAcc)
				continue
			}

			// Check that avatar attachment populated.
			if account.AvatarMediaAttachmentID != "" &&
				(checkAcc.AvatarMediaAttachment == nil || checkAcc.AvatarMediaAttachment.ID != account.AvatarMediaAttachmentID) {
				t.Errorf("account avatar media attachment not correctly populated for: %+v", account)
				continue
			}

			// Check that header attachment populated.
			if account.HeaderMediaAttachmentID != "" &&
				(checkAcc.HeaderMediaAttachment == nil || checkAcc.HeaderMediaAttachment.ID != account.HeaderMediaAttachmentID) {
				t.Errorf("account header media attachment not correctly populated for: %+v", account)
				continue
			}
		}
	}
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

func (suite *AccountTestSuite) TestGetAccountPinnedStatusesSomeResults() {
	testAccount := suite.testAccounts["admin_account"]

	statuses, err := suite.db.GetAccountPinnedStatuses(context.Background(), testAccount.ID)
	suite.NoError(err)
	suite.Len(statuses, 2) // This account has 2 statuses pinned.
}

func (suite *AccountTestSuite) TestGetAccountPinnedStatusesNothingPinned() {
	testAccount := suite.testAccounts["local_account_1"]

	statuses, err := suite.db.GetAccountPinnedStatuses(context.Background(), testAccount.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Empty(statuses) // This account has nothing pinned.
}

func (suite *AccountTestSuite) TestCountAccountPinnedSomeResults() {
	testAccount := suite.testAccounts["admin_account"]

	pinned, err := suite.db.CountAccountPinned(context.Background(), testAccount.ID)
	suite.NoError(err)
	suite.Equal(pinned, 2) // This account has 2 statuses pinned.
}

func (suite *AccountTestSuite) TestCountAccountPinnedNothingPinned() {
	testAccount := suite.testAccounts["local_account_1"]

	pinned, err := suite.db.CountAccountPinned(context.Background(), testAccount.ID)
	suite.NoError(err)
	suite.Equal(pinned, 0) // This account has nothing pinned.
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
