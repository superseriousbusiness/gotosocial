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
	"fmt"
	"net/netip"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

type AccountTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *AccountTestSuite) TestGetAccountStatuses() {
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 20, false, false, "", "", false, false)
	suite.NoError(err)
	suite.Len(statuses, 9)
}

func (suite *AccountTestSuite) TestGetAccountStatusesPageDown() {
	// get the first page
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 3, false, false, "", "", false, false)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Len(statuses, 3)

	// get the second page
	statuses, err = suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 3, false, false, statuses[len(statuses)-1].ID, "", false, false)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Len(statuses, 3)

	// get the third page
	statuses, err = suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 3, false, false, statuses[len(statuses)-1].ID, "", false, false)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Len(statuses, 3)

	// try to get the last page (should be empty)
	statuses, err = suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 3, false, false, statuses[len(statuses)-1].ID, "", false, false)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Empty(statuses)
}

func (suite *AccountTestSuite) TestGetAccountStatusesExcludeRepliesAndReblogs() {
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 20, true, true, "", "", false, false)
	suite.NoError(err)
	suite.Len(statuses, 8)
}

func (suite *AccountTestSuite) TestGetAccountStatusesExcludeRepliesAndReblogsPublicOnly() {
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 20, true, true, "", "", false, true)
	suite.NoError(err)
	suite.Len(statuses, 4)
}

// populateTestStatus adds mandatory fields to a partially populated status.
func (suite *AccountTestSuite) populateTestStatus(testAccountKey string, status *gtsmodel.Status, inReplyTo *gtsmodel.Status) *gtsmodel.Status {
	testAccount := suite.testAccounts[testAccountKey]
	if testAccount == nil {
		suite.FailNowf("", "Missing test account: %s", testAccountKey)
		return status
	}
	if testAccount.Domain != "" {
		suite.FailNowf("", "Only local test accounts are supported: %s is remote", testAccountKey)
		return status
	}

	status.AccountID = testAccount.ID
	status.AccountURI = testAccount.URI
	status.URI = fmt.Sprintf("http://localhost:8080/users/%s/statuses/%s", testAccount.Username, status.ID)
	status.Local = util.Ptr(true)

	if status.Visibility == 0 {
		status.Visibility = gtsmodel.VisibilityDefault
	}
	if status.ActivityStreamsType == "" {
		status.ActivityStreamsType = ap.ObjectNote
	}
	if status.Federated == nil {
		status.Federated = util.Ptr(true)
	}

	if inReplyTo != nil {
		status.InReplyToAccountID = inReplyTo.AccountID
		status.InReplyToID = inReplyTo.ID
		status.InReplyToURI = inReplyTo.URI
	}

	return status
}

// Tests that we're including self-replies but excluding those that mention other accounts.
func (suite *AccountTestSuite) TestGetAccountStatusesExcludeRepliesExcludesSelfRepliesWithMentions() {
	post := suite.populateTestStatus(
		"local_account_1",
		&gtsmodel.Status{
			ID:      "01HQ1FGN679M5F81DZ18WS6JQG",
			Content: "post",
		},
		nil,
	)
	reply := suite.populateTestStatus(
		"local_account_2",
		&gtsmodel.Status{
			ID:         "01HQ1GTXMT2W6PF8MA2XG9DG6Q",
			Content:    "post",
			MentionIDs: []string{post.InReplyToAccountID},
		},
		post,
	)
	riposte := suite.populateTestStatus(
		"local_account_1",
		&gtsmodel.Status{
			ID:         "01HQ1GTXN0RWG9ZWJKRFAEF5RE",
			Content:    "riposte",
			MentionIDs: []string{reply.InReplyToAccountID},
		},
		reply,
	)
	followup := suite.populateTestStatus(
		"local_account_1",
		&gtsmodel.Status{
			ID:         "01HQ1GTXN52X7MM9Z12PNJWEHQ",
			Content:    "followup",
			MentionIDs: []string{reply.InReplyToAccountID},
		},
		riposte,
	)

	for _, status := range []*gtsmodel.Status{post, reply, riposte, followup} {
		if err := suite.db.PutStatus(context.Background(), status); err != nil {
			suite.FailNowf("", "Error while adding test status with ID %s: %v", status.ID, err)
			return
		}
	}

	testAccount := suite.testAccounts["local_account_1"]
	statuses, err := suite.db.GetAccountStatuses(context.Background(), testAccount.ID, 20, true, true, "", "", false, false)
	suite.NoError(err)
	suite.Len(statuses, 9)
	for _, status := range statuses {
		if status.InReplyToID != "" && status.InReplyToAccountID != testAccount.ID {
			suite.FailNowf("", "Status with ID %s is a non-self reply and should have been excluded", status.ID)
		}
		if len(status.Mentions) != 0 {
			suite.FailNowf("", "Status with ID %s has mentions and should have been excluded", status.ID)
		}
	}
}

func (suite *AccountTestSuite) TestGetAccountStatusesMediaOnly() {
	statuses, err := suite.db.GetAccountStatuses(context.Background(), suite.testAccounts["local_account_1"].ID, 20, false, false, "", "", true, false)
	suite.NoError(err)
	suite.Len(statuses, 2)
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
		a1.Settings = nil
		a2.Settings = nil
		a1.Stats = nil
		a2.Stats = nil

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
	err = dbService.DB().
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

	err = dbService.DB().
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

	suite.WithinDuration(time.Now(), newAccount.CreatedAt, 30*time.Second)
	suite.WithinDuration(time.Now(), newAccount.UpdatedAt, 30*time.Second)
	suite.True(*newAccount.Locked)
	suite.False(*newAccount.Bot)
	suite.False(*newAccount.Discoverable)
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

func (suite *AccountTestSuite) TestPopulateAccountWithUnknownMovedToURI() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"]

	// Set test account MovedToURI to something we don't have in the database.
	// We should not get an error when populating.
	testAccount.MovedToURI = "https://unknown-instance.example.org/users/someone_we_dont_know"
	err := suite.db.PopulateAccount(context.Background(), testAccount)
	suite.NoError(err)
}

func (suite *AccountTestSuite) TestGetAccountsAll() {
	var (
		ctx         = context.Background()
		origin      = ""
		status      = ""
		mods        = false
		invitedBy   = ""
		username    = ""
		displayName = ""
		domain      = ""
		email       = ""
		ip          netip.Addr
		page        *paging.Page = nil
	)

	accounts, err := suite.db.GetAccounts(
		ctx,
		origin,
		status,
		mods,
		invitedBy,
		username,
		displayName,
		domain,
		email,
		ip,
		page,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(accounts, 9)
}

func (suite *AccountTestSuite) TestGetAccountsMaxID() {
	var (
		ctx         = context.Background()
		origin      = ""
		status      = ""
		mods        = false
		invitedBy   = ""
		username    = ""
		displayName = ""
		domain      = ""
		email       = ""
		ip          netip.Addr
		// Get accounts with `[domain]/@[username]`
		// later in the alphabet than `/@the_mighty_zork`.
		page = &paging.Page{Max: paging.MaxID("/@the_mighty_zork")}
	)

	accounts, err := suite.db.GetAccounts(
		ctx,
		origin,
		status,
		mods,
		invitedBy,
		username,
		displayName,
		domain,
		email,
		ip,
		page,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(accounts, 5)
}

func (suite *AccountTestSuite) TestGetAccountsMinID() {
	var (
		ctx         = context.Background()
		origin      = ""
		status      = ""
		mods        = false
		invitedBy   = ""
		username    = ""
		displayName = ""
		domain      = ""
		email       = ""
		ip          netip.Addr
		// Get accounts with `[domain]/@[username]`
		// earlier in the alphabet than `/@the_mighty_zork`.
		page = &paging.Page{Min: paging.MinID("/@the_mighty_zork")}
	)

	accounts, err := suite.db.GetAccounts(
		ctx,
		origin,
		status,
		mods,
		invitedBy,
		username,
		displayName,
		domain,
		email,
		ip,
		page,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(accounts, 3)
}

func (suite *AccountTestSuite) TestGetAccountsModsOnly() {
	var (
		ctx         = context.Background()
		origin      = ""
		status      = ""
		mods        = true
		invitedBy   = ""
		username    = ""
		displayName = ""
		domain      = ""
		email       = ""
		ip          netip.Addr
		page        = &paging.Page{
			Limit: 100,
		}
	)

	accounts, err := suite.db.GetAccounts(
		ctx,
		origin,
		status,
		mods,
		invitedBy,
		username,
		displayName,
		domain,
		email,
		ip,
		page,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(accounts, 1)
}

func (suite *AccountTestSuite) TestGetAccountsLocalWithEmail() {
	var (
		ctx         = context.Background()
		origin      = "local"
		status      = ""
		mods        = false
		invitedBy   = ""
		username    = ""
		displayName = ""
		domain      = ""
		email       = "tortle.dude@example.org"
		ip          netip.Addr
		page        = &paging.Page{
			Limit: 100,
		}
	)

	accounts, err := suite.db.GetAccounts(
		ctx,
		origin,
		status,
		mods,
		invitedBy,
		username,
		displayName,
		domain,
		email,
		ip,
		page,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(accounts, 1)
}

func (suite *AccountTestSuite) TestGetAccountsWithIP() {
	var (
		ctx         = context.Background()
		origin      = ""
		status      = ""
		mods        = false
		invitedBy   = ""
		username    = ""
		displayName = ""
		domain      = ""
		email       = ""
		ip          = netip.MustParseAddr("199.222.111.89")
		page        = &paging.Page{
			Limit: 100,
		}
	)

	accounts, err := suite.db.GetAccounts(
		ctx,
		origin,
		status,
		mods,
		invitedBy,
		username,
		displayName,
		domain,
		email,
		ip,
		page,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(accounts, 1)
}

func (suite *AccountTestSuite) TestGetPendingAccounts() {
	var (
		ctx         = context.Background()
		origin      = ""
		status      = "pending"
		mods        = false
		invitedBy   = ""
		username    = ""
		displayName = ""
		domain      = ""
		email       = ""
		ip          netip.Addr
		page        = &paging.Page{
			Limit: 100,
		}
	)

	accounts, err := suite.db.GetAccounts(
		ctx,
		origin,
		status,
		mods,
		invitedBy,
		username,
		displayName,
		domain,
		email,
		ip,
		page,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(accounts, 1)
}

func (suite *AccountTestSuite) TestAccountStatsAll() {
	ctx := context.Background()
	for _, account := range suite.testAccounts {
		// Get stats for the first time. They
		// should all be generated now since
		// they're not stored in the test rig.
		if err := suite.db.PopulateAccountStats(ctx, account); err != nil {
			suite.FailNow(err.Error())
		}
		stats := account.Stats
		suite.NotNil(stats)
		suite.WithinDuration(time.Now(), stats.RegeneratedAt, 5*time.Second)

		// Get stats a second time. They shouldn't
		// be regenerated since we just did it.
		if err := suite.db.PopulateAccountStats(ctx, account); err != nil {
			suite.FailNow(err.Error())
		}
		stats2 := account.Stats
		suite.NotNil(stats2)
		suite.Equal(stats2.RegeneratedAt, stats.RegeneratedAt)

		// Update the stats to indicate they're out of date.
		stats2.RegeneratedAt = time.Now().Add(-72 * time.Hour)
		if err := suite.db.UpdateAccountStats(ctx, stats2, "regenerated_at"); err != nil {
			suite.FailNow(err.Error())
		}

		// Nil out account stats to allow
		// db to refetch + regenerate them.
		account.Stats = nil

		// Get stats for a third time, they
		// should get regenerated now, but
		// only for local accounts.
		if err := suite.db.PopulateAccountStats(ctx, account); err != nil {
			suite.FailNow(err.Error())
		}
		stats3 := account.Stats
		suite.NotNil(stats3)
		if account.IsLocal() {
			suite.True(stats3.RegeneratedAt.After(stats.RegeneratedAt))
		} else {
			suite.False(stats3.RegeneratedAt.After(stats.RegeneratedAt))
		}

		// Now delete the stats.
		if err := suite.db.DeleteAccountStats(ctx, account.ID); err != nil {
			suite.FailNow(err.Error())
		}
	}
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
