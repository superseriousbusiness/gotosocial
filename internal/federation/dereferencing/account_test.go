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

package dereferencing_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AccountTestSuite struct {
	DereferencerStandardTestSuite
}

func (suite *AccountTestSuite) TestDereferenceGroup() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	groupURL := testrig.URLMustParse("https://unknown-instance.com/groups/some_group")
	group, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername: fetchingAccount.Username,
		RemoteAccountID:    groupURL,
	})
	suite.NoError(err)
	suite.NotNil(group)

	// group values should be set
	suite.Equal("https://unknown-instance.com/groups/some_group", group.URI)
	suite.Equal("https://unknown-instance.com/@some_group", group.URL)
	suite.WithinDuration(time.Now(), group.LastWebfingeredAt, 5*time.Second)

	// group should be in the database
	dbGroup, err := suite.db.GetAccountByURI(context.Background(), group.URI)
	suite.NoError(err)
	suite.Equal(group.ID, dbGroup.ID)
	suite.Equal(ap.ActorGroup, dbGroup.ActorType)
}

func (suite *AccountTestSuite) TestDereferenceService() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	serviceURL := testrig.URLMustParse("https://owncast.example.org/federation/user/rgh")
	service, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername: fetchingAccount.Username,
		RemoteAccountID:    serviceURL,
	})
	suite.NoError(err)
	suite.NotNil(service)

	// service values should be set
	suite.Equal("https://owncast.example.org/federation/user/rgh", service.URI)
	suite.Equal("https://owncast.example.org/federation/user/rgh", service.URL)
	suite.WithinDuration(time.Now(), service.LastWebfingeredAt, 5*time.Second)

	// service should be in the database
	dbService, err := suite.db.GetAccountByURI(context.Background(), service.URI)
	suite.NoError(err)
	suite.Equal(service.ID, dbService.ID)
	suite.Equal(ap.ActorService, dbService.ActorType)
	suite.Equal("example.org", dbService.Domain)
}

/*
	We shouldn't try webfingering or making http calls to dereference local accounts
	that might be passed into GetRemoteAccount for whatever reason, so these tests are
	here to make sure that such cases are (basically) short-circuit evaluated and given
	back as-is without trying to make any calls to one's own instance.
*/

func (suite *AccountTestSuite) TestDereferenceLocalAccountAsRemoteURL() {
	fetchingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["local_account_2"]

	fetchedAccount, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername: fetchingAccount.Username,
		RemoteAccountID:    testrig.URLMustParse(targetAccount.URI),
	})
	suite.NoError(err)
	suite.NotNil(fetchedAccount)
	suite.Empty(fetchedAccount.Domain)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountAsRemoteURLNoSharedInboxYet() {
	fetchingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["local_account_2"]

	targetAccount.SharedInboxURI = nil
	if err := suite.db.UpdateAccount(context.Background(), targetAccount); err != nil {
		suite.FailNow(err.Error())
	}

	fetchedAccount, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername: fetchingAccount.Username,
		RemoteAccountID:    testrig.URLMustParse(targetAccount.URI),
	})
	suite.NoError(err)
	suite.NotNil(fetchedAccount)
	suite.Empty(fetchedAccount.Domain)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountAsUsername() {
	fetchingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["local_account_2"]

	fetchedAccount, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername:    fetchingAccount.Username,
		RemoteAccountUsername: targetAccount.Username,
	})
	suite.NoError(err)
	suite.NotNil(fetchedAccount)
	suite.Empty(fetchedAccount.Domain)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountAsUsernameDomain() {
	fetchingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["local_account_2"]

	fetchedAccount, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername:    fetchingAccount.Username,
		RemoteAccountUsername: targetAccount.Username,
		RemoteAccountHost:     config.GetHost(),
	})
	suite.NoError(err)
	suite.NotNil(fetchedAccount)
	suite.Empty(fetchedAccount.Domain)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountAsUsernameDomainAndURL() {
	fetchingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["local_account_2"]

	fetchedAccount, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername:    fetchingAccount.Username,
		RemoteAccountID:       testrig.URLMustParse(targetAccount.URI),
		RemoteAccountUsername: targetAccount.Username,
		RemoteAccountHost:     config.GetHost(),
	})
	suite.NoError(err)
	suite.NotNil(fetchedAccount)
	suite.Empty(fetchedAccount.Domain)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountWithUnknownUsername() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	fetchedAccount, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername:    fetchingAccount.Username,
		RemoteAccountUsername: "thisaccountdoesnotexist",
	})
	var errNotRetrievable *dereferencing.ErrNotRetrievable
	suite.ErrorAs(err, &errNotRetrievable)
	suite.EqualError(err, "item could not be retrieved: GetRemoteAccount: couldn't retrieve account locally and not allowed to resolve it")
	suite.Nil(fetchedAccount)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountWithUnknownUsernameDomain() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	fetchedAccount, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername:    fetchingAccount.Username,
		RemoteAccountUsername: "thisaccountdoesnotexist",
		RemoteAccountHost:     "localhost:8080",
	})
	var errNotRetrievable *dereferencing.ErrNotRetrievable
	suite.ErrorAs(err, &errNotRetrievable)
	suite.EqualError(err, "item could not be retrieved: GetRemoteAccount: couldn't retrieve account locally and not allowed to resolve it")
	suite.Nil(fetchedAccount)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountWithUnknownUserURI() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	fetchedAccount, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername: fetchingAccount.Username,
		RemoteAccountID:    testrig.URLMustParse("http://localhost:8080/users/thisaccountdoesnotexist"),
	})
	var errNotRetrievable *dereferencing.ErrNotRetrievable
	suite.ErrorAs(err, &errNotRetrievable)
	suite.EqualError(err, "item could not be retrieved: GetRemoteAccount: couldn't retrieve account locally and not allowed to resolve it")
	suite.Nil(fetchedAccount)
}

func (suite *AccountTestSuite) TestDereferenceRemoteAccountWithPartial() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	remoteAccount := suite.testAccounts["remote_account_1"]
	remoteAccountPartial := &gtsmodel.Account{
		ID:                    remoteAccount.ID,
		ActorType:             remoteAccount.ActorType,
		Language:              remoteAccount.Language,
		CreatedAt:             remoteAccount.CreatedAt,
		UpdatedAt:             remoteAccount.UpdatedAt,
		Username:              remoteAccount.Username,
		Domain:                remoteAccount.Domain,
		DisplayName:           remoteAccount.DisplayName,
		URI:                   remoteAccount.URI,
		InboxURI:              remoteAccount.URI,
		SharedInboxURI:        remoteAccount.SharedInboxURI,
		PublicKeyURI:          remoteAccount.PublicKeyURI,
		URL:                   remoteAccount.URL,
		FollowingURI:          remoteAccount.FollowingURI,
		FollowersURI:          remoteAccount.FollowersURI,
		OutboxURI:             remoteAccount.OutboxURI,
		FeaturedCollectionURI: remoteAccount.FeaturedCollectionURI,
		Emojis: []*gtsmodel.Emoji{
			// dereference an emoji we don't have stored yet
			{
				URI:             "http://fossbros-anonymous.io/emoji/01GD5HCC2YECT012TK8PAGX4D1",
				Shortcode:       "kip_van_den_bos",
				UpdatedAt:       testrig.TimeMustParse("2022-09-13T12:13:12+02:00"),
				ImageUpdatedAt:  testrig.TimeMustParse("2022-09-13T12:13:12+02:00"),
				ImageRemoteURL:  "http://fossbros-anonymous.io/emoji/kip.gif",
				Disabled:        testrig.FalseBool(),
				VisibleInPicker: testrig.FalseBool(),
				Domain:          "fossbros-anonymous.io",
			},
		},
	}

	fetchedAccount, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername:    fetchingAccount.Username,
		RemoteAccountID:       testrig.URLMustParse(remoteAccount.URI),
		RemoteAccountHost:     remoteAccount.Domain,
		RemoteAccountUsername: remoteAccount.Username,
		PartialAccount:        remoteAccountPartial,
		Blocking:              true,
	})
	suite.NoError(err)
	suite.NotNil(fetchedAccount)
	suite.NotNil(fetchedAccount.EmojiIDs)
	suite.NotNil(fetchedAccount.Emojis)
}

func (suite *AccountTestSuite) TestDereferenceRemoteAccountWithPartial2() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	knownEmoji := suite.testEmojis["yell"]

	remoteAccount := suite.testAccounts["remote_account_1"]
	remoteAccountPartial := &gtsmodel.Account{
		ID:                    remoteAccount.ID,
		ActorType:             remoteAccount.ActorType,
		Language:              remoteAccount.Language,
		CreatedAt:             remoteAccount.CreatedAt,
		UpdatedAt:             remoteAccount.UpdatedAt,
		Username:              remoteAccount.Username,
		Domain:                remoteAccount.Domain,
		DisplayName:           remoteAccount.DisplayName,
		URI:                   remoteAccount.URI,
		InboxURI:              remoteAccount.URI,
		SharedInboxURI:        remoteAccount.SharedInboxURI,
		PublicKeyURI:          remoteAccount.PublicKeyURI,
		URL:                   remoteAccount.URL,
		FollowingURI:          remoteAccount.FollowingURI,
		FollowersURI:          remoteAccount.FollowersURI,
		OutboxURI:             remoteAccount.OutboxURI,
		FeaturedCollectionURI: remoteAccount.FeaturedCollectionURI,
		Emojis: []*gtsmodel.Emoji{
			// an emoji we already have
			{
				URI:             knownEmoji.URI,
				Shortcode:       knownEmoji.Shortcode,
				UpdatedAt:       knownEmoji.UpdatedAt,
				ImageUpdatedAt:  knownEmoji.ImageUpdatedAt,
				ImageRemoteURL:  knownEmoji.ImageRemoteURL,
				Disabled:        knownEmoji.Disabled,
				VisibleInPicker: knownEmoji.VisibleInPicker,
				Domain:          knownEmoji.Domain,
			},
		},
	}

	fetchedAccount, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername:    fetchingAccount.Username,
		RemoteAccountID:       testrig.URLMustParse(remoteAccount.URI),
		RemoteAccountHost:     remoteAccount.Domain,
		RemoteAccountUsername: remoteAccount.Username,
		PartialAccount:        remoteAccountPartial,
		Blocking:              true,
	})
	suite.NoError(err)
	suite.NotNil(fetchedAccount)
	suite.NotNil(fetchedAccount.EmojiIDs)
	suite.NotNil(fetchedAccount.Emojis)
}

func (suite *AccountTestSuite) TestDereferenceRemoteAccountWithPartial3() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	knownEmoji := suite.testEmojis["yell"]

	remoteAccount := suite.testAccounts["remote_account_1"]
	remoteAccountPartial := &gtsmodel.Account{
		ID:                    remoteAccount.ID,
		ActorType:             remoteAccount.ActorType,
		Language:              remoteAccount.Language,
		CreatedAt:             remoteAccount.CreatedAt,
		UpdatedAt:             remoteAccount.UpdatedAt,
		Username:              remoteAccount.Username,
		Domain:                remoteAccount.Domain,
		DisplayName:           remoteAccount.DisplayName,
		URI:                   remoteAccount.URI,
		InboxURI:              remoteAccount.URI,
		SharedInboxURI:        remoteAccount.SharedInboxURI,
		PublicKeyURI:          remoteAccount.PublicKeyURI,
		URL:                   remoteAccount.URL,
		FollowingURI:          remoteAccount.FollowingURI,
		FollowersURI:          remoteAccount.FollowersURI,
		OutboxURI:             remoteAccount.OutboxURI,
		FeaturedCollectionURI: remoteAccount.FeaturedCollectionURI,
		Emojis: []*gtsmodel.Emoji{
			// an emoji we already have
			{
				URI:             knownEmoji.URI,
				Shortcode:       knownEmoji.Shortcode,
				UpdatedAt:       knownEmoji.UpdatedAt,
				ImageUpdatedAt:  knownEmoji.ImageUpdatedAt,
				ImageRemoteURL:  knownEmoji.ImageRemoteURL,
				Disabled:        knownEmoji.Disabled,
				VisibleInPicker: knownEmoji.VisibleInPicker,
				Domain:          knownEmoji.Domain,
			},
		},
	}

	fetchedAccount, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername:    fetchingAccount.Username,
		RemoteAccountID:       testrig.URLMustParse(remoteAccount.URI),
		RemoteAccountHost:     remoteAccount.Domain,
		RemoteAccountUsername: remoteAccount.Username,
		PartialAccount:        remoteAccountPartial,
		Blocking:              true,
	})
	suite.NoError(err)
	suite.NotNil(fetchedAccount)
	suite.NotNil(fetchedAccount.EmojiIDs)
	suite.NotNil(fetchedAccount.Emojis)
	suite.Equal(knownEmoji.URI, fetchedAccount.Emojis[0].URI)

	remoteAccountPartial2 := &gtsmodel.Account{
		ID:                    remoteAccount.ID,
		ActorType:             remoteAccount.ActorType,
		Language:              remoteAccount.Language,
		CreatedAt:             remoteAccount.CreatedAt,
		UpdatedAt:             remoteAccount.UpdatedAt,
		Username:              remoteAccount.Username,
		Domain:                remoteAccount.Domain,
		DisplayName:           remoteAccount.DisplayName,
		URI:                   remoteAccount.URI,
		InboxURI:              remoteAccount.URI,
		SharedInboxURI:        remoteAccount.SharedInboxURI,
		PublicKeyURI:          remoteAccount.PublicKeyURI,
		URL:                   remoteAccount.URL,
		FollowingURI:          remoteAccount.FollowingURI,
		FollowersURI:          remoteAccount.FollowersURI,
		OutboxURI:             remoteAccount.OutboxURI,
		FeaturedCollectionURI: remoteAccount.FeaturedCollectionURI,
		Emojis: []*gtsmodel.Emoji{
			// dereference an emoji we don't have stored yet
			{
				URI:             "http://fossbros-anonymous.io/emoji/01GD5HCC2YECT012TK8PAGX4D1",
				Shortcode:       "kip_van_den_bos",
				UpdatedAt:       testrig.TimeMustParse("2022-09-13T12:13:12+02:00"),
				ImageUpdatedAt:  testrig.TimeMustParse("2022-09-13T12:13:12+02:00"),
				ImageRemoteURL:  "http://fossbros-anonymous.io/emoji/kip.gif",
				Disabled:        testrig.FalseBool(),
				VisibleInPicker: testrig.FalseBool(),
				Domain:          "fossbros-anonymous.io",
			},
		},
	}

	fetchedAccount2, err := suite.dereferencer.GetAccount(context.Background(), dereferencing.GetAccountParams{
		RequestingUsername:    fetchingAccount.Username,
		RemoteAccountID:       testrig.URLMustParse(remoteAccount.URI),
		RemoteAccountHost:     remoteAccount.Domain,
		RemoteAccountUsername: remoteAccount.Username,
		PartialAccount:        remoteAccountPartial2,
		Blocking:              true,
	})
	suite.NoError(err)
	suite.NotNil(fetchedAccount2)
	suite.NotNil(fetchedAccount2.EmojiIDs)
	suite.NotNil(fetchedAccount2.Emojis)
	suite.Equal("http://fossbros-anonymous.io/emoji/01GD5HCC2YECT012TK8PAGX4D1", fetchedAccount2.Emojis[0].URI)
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
