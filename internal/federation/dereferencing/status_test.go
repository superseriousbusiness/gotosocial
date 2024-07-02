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

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusTestSuite struct {
	DereferencerStandardTestSuite
}

func (suite *StatusTestSuite) TestDereferenceSimpleStatus() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	statusURL := testrig.URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01FE4NTHKWW7THT67EF10EB839")
	status, _, err := suite.dereferencer.GetStatusByURI(context.Background(), fetchingAccount.Username, statusURL)
	suite.NoError(err)
	suite.NotNil(status)

	// status values should be set
	suite.Equal("https://unknown-instance.com/users/brand_new_person/statuses/01FE4NTHKWW7THT67EF10EB839", status.URI)
	suite.Equal("https://unknown-instance.com/users/@brand_new_person/01FE4NTHKWW7THT67EF10EB839", status.URL)
	suite.Equal("Hello world!", status.Content)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", status.AccountURI)
	suite.False(*status.Local)
	suite.Empty(status.ContentWarning)
	suite.Equal(gtsmodel.VisibilityPublic, status.Visibility)
	suite.Equal(ap.ObjectNote, status.ActivityStreamsType)

	// status should be in the database
	dbStatus, err := suite.db.GetStatusByURI(context.Background(), status.URI)
	suite.NoError(err)
	suite.Equal(status.ID, dbStatus.ID)
	suite.True(*dbStatus.Federated)

	// account should be in the database now too
	account, err := suite.db.GetAccountByURI(context.Background(), status.AccountURI)
	suite.NoError(err)
	suite.NotNil(account)
	suite.True(*account.Discoverable)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", account.URI)
	suite.Equal("hey I'm a new person, your instance hasn't seen me yet uwu", account.Note)
	suite.Equal("Geoff Brando New Personson", account.DisplayName)
	suite.Equal("brand_new_person", account.Username)
	suite.NotNil(account.PublicKey)
	suite.Nil(account.PrivateKey)
}

func (suite *StatusTestSuite) TestDereferenceStatusWithMention() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	statusURL := testrig.URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01FE5Y30E3W4P7TRE0R98KAYQV")
	status, _, err := suite.dereferencer.GetStatusByURI(context.Background(), fetchingAccount.Username, statusURL)
	suite.NoError(err)
	suite.NotNil(status)

	// status values should be set
	suite.Equal("https://unknown-instance.com/users/brand_new_person/statuses/01FE5Y30E3W4P7TRE0R98KAYQV", status.URI)
	suite.Equal("https://unknown-instance.com/users/@brand_new_person/01FE5Y30E3W4P7TRE0R98KAYQV", status.URL)
	suite.Equal("Hey @the_mighty_zork@localhost:8080 how's it going?", status.Content)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", status.AccountURI)
	suite.False(*status.Local)
	suite.Empty(status.ContentWarning)
	suite.Equal(gtsmodel.VisibilityPublic, status.Visibility)
	suite.Equal(ap.ObjectNote, status.ActivityStreamsType)

	// status should be in the database
	dbStatus, err := suite.db.GetStatusByURI(context.Background(), status.URI)
	suite.NoError(err)
	suite.Equal(status.ID, dbStatus.ID)
	suite.True(*dbStatus.Federated)

	// account should be in the database now too
	account, err := suite.db.GetAccountByURI(context.Background(), status.AccountURI)
	suite.NoError(err)
	suite.NotNil(account)
	suite.True(*account.Discoverable)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", account.URI)
	suite.Equal("hey I'm a new person, your instance hasn't seen me yet uwu", account.Note)
	suite.Equal("Geoff Brando New Personson", account.DisplayName)
	suite.Equal("brand_new_person", account.Username)
	suite.NotNil(account.PublicKey)
	suite.Nil(account.PrivateKey)

	// we should have a mention in the database
	m := &gtsmodel.Mention{}
	err = suite.db.GetWhere(context.Background(), []db.Where{{Key: "status_id", Value: status.ID}}, m)
	suite.NoError(err)
	suite.NotNil(m)
	suite.Equal(status.ID, m.StatusID)
	suite.Equal(account.ID, m.OriginAccountID)
	suite.Equal(fetchingAccount.ID, m.TargetAccountID)
	suite.Equal(account.URI, m.OriginAccountURI)
	suite.False(*m.Silent)
}

func (suite *StatusTestSuite) TestDereferenceStatusWithTag() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	statusURL := testrig.URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01H641QSRS3TCXSVC10X4GPKW7")
	status, _, err := suite.dereferencer.GetStatusByURI(context.Background(), fetchingAccount.Username, statusURL)
	suite.NoError(err)
	suite.NotNil(status)

	// status values should be set
	suite.Equal("https://unknown-instance.com/users/brand_new_person/statuses/01H641QSRS3TCXSVC10X4GPKW7", status.URI)
	suite.Equal("https://unknown-instance.com/users/@brand_new_person/01H641QSRS3TCXSVC10X4GPKW7", status.URL)
	suite.Equal("<p>Babe are you okay, you've hardly touched your <a href=\"https://unknown-instance.com/tags/piss\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>piss</span></a></p>", status.Content)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", status.AccountURI)
	suite.False(*status.Local)
	suite.Empty(status.ContentWarning)
	suite.Equal(gtsmodel.VisibilityPublic, status.Visibility)
	suite.Equal(ap.ObjectNote, status.ActivityStreamsType)

	// Ensure tags set + ID'd.
	suite.Len(status.Tags, 1)
	suite.Len(status.TagIDs, 1)

	// status should be in the database
	dbStatus, err := suite.db.GetStatusByURI(context.Background(), status.URI)
	suite.NoError(err)
	suite.Equal(status.ID, dbStatus.ID)
	suite.True(*dbStatus.Federated)

	// account should be in the database now too
	account, err := suite.db.GetAccountByURI(context.Background(), status.AccountURI)
	suite.NoError(err)
	suite.NotNil(account)
	suite.True(*account.Discoverable)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", account.URI)
	suite.Equal("hey I'm a new person, your instance hasn't seen me yet uwu", account.Note)
	suite.Equal("Geoff Brando New Personson", account.DisplayName)
	suite.Equal("brand_new_person", account.Username)
	suite.NotNil(account.PublicKey)
	suite.Nil(account.PrivateKey)

	// we should have a tag in the database
	t := &gtsmodel.Tag{}
	err = suite.db.GetWhere(context.Background(), []db.Where{{Key: "name", Value: "piss"}}, t)
	suite.NoError(err)
	suite.NotNil(t)
}

func (suite *StatusTestSuite) TestDereferenceStatusWithImageAndNoContent() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	statusURL := testrig.URLMustParse("https://turnip.farm/users/turniplover6969/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042")
	status, _, err := suite.dereferencer.GetStatusByURI(context.Background(), fetchingAccount.Username, statusURL)
	suite.NoError(err)
	suite.NotNil(status)

	// status values should be set
	suite.Equal("https://turnip.farm/users/turniplover6969/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042", status.URI)
	suite.Equal("https://turnip.farm/@turniplover6969/70c53e54-3146-42d5-a630-83c8b6c7c042", status.URL)
	suite.Equal("", status.Content)
	suite.Equal("https://turnip.farm/users/turniplover6969", status.AccountURI)
	suite.False(*status.Local)
	suite.Empty(status.ContentWarning)
	suite.Equal(gtsmodel.VisibilityPublic, status.Visibility)
	suite.Equal(ap.ObjectNote, status.ActivityStreamsType)

	// status should be in the database
	dbStatus, err := suite.db.GetStatusByURI(context.Background(), status.URI)
	suite.NoError(err)
	suite.Equal(status.ID, dbStatus.ID)
	suite.True(*dbStatus.Federated)

	// account should be in the database now too
	account, err := suite.db.GetAccountByURI(context.Background(), status.AccountURI)
	suite.NoError(err)
	suite.NotNil(account)
	suite.True(*account.Discoverable)
	suite.Equal("https://turnip.farm/users/turniplover6969", account.URI)
	suite.Equal("I just think they're neat", account.Note)
	suite.Equal("Turnip Lover 6969", account.DisplayName)
	suite.Equal("turniplover6969", account.Username)
	suite.NotNil(account.PublicKey)
	suite.Nil(account.PrivateKey)

	// we should have an attachment in the database
	a := &gtsmodel.MediaAttachment{}
	err = suite.db.GetWhere(context.Background(), []db.Where{{Key: "status_id", Value: status.ID}}, a)
	suite.NoError(err)
}

func (suite *StatusTestSuite) TestDereferenceStatusWithNonMatchingURI() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	const (
		remoteURI    = "https://turnip.farm/users/turniplover6969/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042"
		remoteAltURI = "https://turnip.farm/users/turniphater420/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042"
	)

	// Create a copy of this remote account at alternative URI.
	remoteStatus := suite.client.TestRemoteStatuses[remoteURI]
	suite.client.TestRemoteStatuses[remoteAltURI] = remoteStatus

	// Attempt to fetch account at alternative URI, it should fail!
	fetchedStatus, _, err := suite.dereferencer.GetStatusByURI(
		context.Background(),
		fetchingAccount.Username,
		testrig.URLMustParse(remoteAltURI),
	)
	suite.Equal(err.Error(), fmt.Sprintf("enrichStatus: dereferenced status uri %s does not match %s", remoteURI, remoteAltURI))
	suite.Nil(fetchedStatus)
}

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}
