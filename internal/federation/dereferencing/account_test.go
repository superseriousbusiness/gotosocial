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

package dereferencing_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AccountTestSuite struct {
	DereferencerStandardTestSuite
}

func (suite *AccountTestSuite) TestDereferenceGroup() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	groupURL := testrig.URLMustParse("https://unknown-instance.com/groups/some_group")
	group, _, err := suite.dereferencer.GetAccountByURI(
		context.Background(),
		fetchingAccount.Username,
		groupURL,
	)
	suite.NoError(err)
	suite.NotNil(group)

	// group values should be set
	suite.Equal("https://unknown-instance.com/groups/some_group", group.URI)
	suite.Equal("https://unknown-instance.com/@some_group", group.URL)
	suite.WithinDuration(time.Now(), group.FetchedAt, 5*time.Second)

	// group should be in the database
	dbGroup, err := suite.db.GetAccountByURI(context.Background(), group.URI)
	suite.NoError(err)
	suite.Equal(group.ID, dbGroup.ID)
	suite.Equal(ap.ActorGroup, dbGroup.ActorType)
}

func (suite *AccountTestSuite) TestDereferenceService() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	serviceURL := testrig.URLMustParse("https://owncast.example.org/federation/user/rgh")
	service, _, err := suite.dereferencer.GetAccountByURI(
		context.Background(),
		fetchingAccount.Username,
		serviceURL,
	)
	suite.NoError(err)
	suite.NotNil(service)

	// service values should be set
	suite.Equal("https://owncast.example.org/federation/user/rgh", service.URI)
	suite.Equal("https://owncast.example.org/federation/user/rgh", service.URL)
	suite.WithinDuration(time.Now(), service.FetchedAt, 5*time.Second)

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

	fetchedAccount, _, err := suite.dereferencer.GetAccountByURI(
		context.Background(),
		fetchingAccount.Username,
		testrig.URLMustParse(targetAccount.URI),
	)
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

	fetchedAccount, _, err := suite.dereferencer.GetAccountByURI(
		context.Background(),
		fetchingAccount.Username,
		testrig.URLMustParse(targetAccount.URI),
	)
	suite.NoError(err)
	suite.NotNil(fetchedAccount)
	suite.Empty(fetchedAccount.Domain)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountAsUsername() {
	fetchingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["local_account_2"]

	fetchedAccount, _, err := suite.dereferencer.GetAccountByURI(
		context.Background(),
		fetchingAccount.Username,
		testrig.URLMustParse(targetAccount.URI),
	)
	suite.NoError(err)
	suite.NotNil(fetchedAccount)
	suite.Empty(fetchedAccount.Domain)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountAsUsernameDomain() {
	fetchingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["local_account_2"]

	fetchedAccount, _, err := suite.dereferencer.GetAccountByURI(
		context.Background(),
		fetchingAccount.Username,
		testrig.URLMustParse(targetAccount.URI),
	)
	suite.NoError(err)
	suite.NotNil(fetchedAccount)
	suite.Empty(fetchedAccount.Domain)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountAsUsernameDomainAndURL() {
	fetchingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["local_account_2"]

	fetchedAccount, _, err := suite.dereferencer.GetAccountByUsernameDomain(
		context.Background(),
		fetchingAccount.Username,
		targetAccount.Username,
		config.GetHost(),
	)
	suite.NoError(err)
	suite.NotNil(fetchedAccount)
	suite.Empty(fetchedAccount.Domain)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountWithUnknownUsername() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	fetchedAccount, _, err := suite.dereferencer.GetAccountByUsernameDomain(
		context.Background(),
		fetchingAccount.Username,
		"thisaccountdoesnotexist",
		config.GetHost(),
	)
	suite.True(gtserror.IsUnretrievable(err))
	suite.EqualError(err, db.ErrNoEntries.Error())
	suite.Nil(fetchedAccount)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountWithUnknownUsernameDomain() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	fetchedAccount, _, err := suite.dereferencer.GetAccountByUsernameDomain(
		context.Background(),
		fetchingAccount.Username,
		"thisaccountdoesnotexist",
		"localhost:8080",
	)
	suite.True(gtserror.IsUnretrievable(err))
	suite.EqualError(err, db.ErrNoEntries.Error())
	suite.Nil(fetchedAccount)
}

func (suite *AccountTestSuite) TestDereferenceLocalAccountWithUnknownUserURI() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	fetchedAccount, _, err := suite.dereferencer.GetAccountByURI(
		context.Background(),
		fetchingAccount.Username,
		testrig.URLMustParse("http://localhost:8080/users/thisaccountdoesnotexist"),
	)
	suite.True(gtserror.IsUnretrievable(err))
	suite.EqualError(err, db.ErrNoEntries.Error())
	suite.Nil(fetchedAccount)
}

func (suite *AccountTestSuite) TestDereferenceRemoteAccountWithNonMatchingURI() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	const (
		remoteURI    = "https://turnip.farm/users/turniplover6969"
		remoteAltURI = "https://turnip.farm/users/turniphater420"
	)

	// Create a copy of this remote account at alternative URI.
	remotePerson := suite.client.TestRemotePeople[remoteURI]
	suite.client.TestRemotePeople[remoteAltURI] = remotePerson

	// Attempt to fetch account at alternative URI, it should fail!
	fetchedAccount, _, err := suite.dereferencer.GetAccountByURI(
		context.Background(),
		fetchingAccount.Username,
		testrig.URLMustParse(remoteAltURI),
	)
	suite.Equal(err.Error(), fmt.Sprintf("enrichAccount: account uri %s does not match %s", remoteURI, remoteAltURI))
	suite.Nil(fetchedAccount)
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
