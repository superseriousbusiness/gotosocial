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
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
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

func (suite *AccountTestSuite) TestDereferenceRemoteAccountWithUnexpectedKeyChange() {
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	fetchingAcc := suite.testAccounts["local_account_1"]
	remoteURI := "https://turnip.farm/users/turniplover6969"

	// Fetch the remote account to load into the database.
	remoteAcc, _, err := suite.dereferencer.GetAccountByURI(ctx,
		fetchingAcc.Username,
		testrig.URLMustParse(remoteURI),
	)
	suite.NoError(err)
	suite.NotNil(remoteAcc)

	// Mark account as requiring a refetch.
	remoteAcc.FetchedAt = time.Time{}
	err = suite.state.DB.UpdateAccount(ctx, remoteAcc, "fetched_at")
	suite.NoError(err)

	// Update remote to have an unexpected different key.
	remotePerson := suite.client.TestRemotePeople[remoteURI]
	setPublicKey(remotePerson,
		remoteURI,
		fetchingAcc.PublicKeyURI+".unique",
		fetchingAcc.PublicKey,
	)

	// Force refresh account expecting key change error.
	_, _, err = suite.dereferencer.RefreshAccount(ctx,
		fetchingAcc.Username,
		remoteAcc,
		nil,
		nil,
	)
	suite.Equal(err.Error(), fmt.Sprintf("RefreshAccount: enrichAccount: account %s pubkey has changed (key rotation required?)", remoteURI))
}

func (suite *AccountTestSuite) TestDereferenceRemoteAccountWithExpectedKeyChange() {
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	fetchingAcc := suite.testAccounts["local_account_1"]
	remoteURI := "https://turnip.farm/users/turniplover6969"

	// Fetch the remote account to load into the database.
	remoteAcc, _, err := suite.dereferencer.GetAccountByURI(ctx,
		fetchingAcc.Username,
		testrig.URLMustParse(remoteURI),
	)
	suite.NoError(err)
	suite.NotNil(remoteAcc)

	// Expire the remote account's public key.
	remoteAcc.PublicKeyExpiresAt = time.Now()
	remoteAcc.FetchedAt = time.Time{} // force fetch
	err = suite.state.DB.UpdateAccount(ctx, remoteAcc, "fetched_at", "public_key_expires_at")
	suite.NoError(err)

	// Update remote to have a different stored public key.
	remotePerson := suite.client.TestRemotePeople[remoteURI]
	setPublicKey(remotePerson,
		remoteURI,
		fetchingAcc.PublicKeyURI+".unique",
		fetchingAcc.PublicKey,
	)

	// Refresh account expecting a succesful refresh with changed keys!
	updatedAcc, apAcc, err := suite.dereferencer.RefreshAccount(ctx,
		fetchingAcc.Username,
		remoteAcc,
		nil,
		nil,
	)
	suite.NoError(err)
	suite.NotNil(apAcc)
	suite.True(updatedAcc.PublicKey.Equal(fetchingAcc.PublicKey))
}

func (suite *AccountTestSuite) TestRefreshFederatedRemoteAccountWithKeyChange() {
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	fetchingAcc := suite.testAccounts["local_account_1"]
	remoteURI := "https://turnip.farm/users/turniplover6969"

	// Fetch the remote account to load into the database.
	remoteAcc, _, err := suite.dereferencer.GetAccountByURI(ctx,
		fetchingAcc.Username,
		testrig.URLMustParse(remoteURI),
	)
	suite.NoError(err)
	suite.NotNil(remoteAcc)

	// Update remote to have a different stored public key.
	remotePerson := suite.client.TestRemotePeople[remoteURI]
	setPublicKey(remotePerson,
		remoteURI,
		fetchingAcc.PublicKeyURI+".unique",
		fetchingAcc.PublicKey,
	)

	// Refresh account expecting a succesful refresh with changed keys!
	// By passing in the remote person model this indicates that the data
	// was received via the federator, which should trust any key change.
	updatedAcc, apAcc, err := suite.dereferencer.RefreshAccount(ctx,
		fetchingAcc.Username,
		remoteAcc,
		remotePerson,
		nil,
	)
	suite.NoError(err)
	suite.NotNil(apAcc)
	suite.True(updatedAcc.PublicKey.Equal(fetchingAcc.PublicKey))
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}

func setPublicKey(person vocab.ActivityStreamsPerson, ownerURI, keyURI string, key *rsa.PublicKey) {
	profileIDURI, err := url.Parse(ownerURI)
	if err != nil {
		panic(err)
	}

	publicKeyURI, err := url.Parse(keyURI)
	if err != nil {
		panic(err)
	}

	publicKeyProp := streams.NewW3IDSecurityV1PublicKeyProperty()

	// create the public key
	publicKey := streams.NewW3IDSecurityV1PublicKey()

	// set ID for the public key
	publicKeyIDProp := streams.NewJSONLDIdProperty()
	publicKeyIDProp.SetIRI(publicKeyURI)
	publicKey.SetJSONLDId(publicKeyIDProp)

	// set owner for the public key
	publicKeyOwnerProp := streams.NewW3IDSecurityV1OwnerProperty()
	publicKeyOwnerProp.SetIRI(profileIDURI)
	publicKey.SetW3IDSecurityV1Owner(publicKeyOwnerProp)

	// set the pem key itself
	encodedPublicKey, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		panic(err)
	}
	publicKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: encodedPublicKey,
	})
	publicKeyPEMProp := streams.NewW3IDSecurityV1PublicKeyPemProperty()
	publicKeyPEMProp.Set(string(publicKeyBytes))
	publicKey.SetW3IDSecurityV1PublicKeyPem(publicKeyPEMProp)

	// append the public key to the public key property
	publicKeyProp.AppendW3IDSecurityV1PublicKey(publicKey)

	// set the public key property on the Person
	person.SetW3IDSecurityV1PublicKey(publicKeyProp)
}
