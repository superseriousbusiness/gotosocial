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
	status, _, err := suite.dereferencer.GetStatus(context.Background(), fetchingAccount.Username, statusURL, false, false)
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
	suite.True(*dbStatus.Boostable)
	suite.True(*dbStatus.Replyable)
	suite.True(*dbStatus.Likeable)

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
	status, _, err := suite.dereferencer.GetStatus(context.Background(), fetchingAccount.Username, statusURL, false, false)
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
	suite.True(*dbStatus.Boostable)
	suite.True(*dbStatus.Replyable)
	suite.True(*dbStatus.Likeable)

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

func (suite *StatusTestSuite) TestDereferenceStatusWithImageAndNoContent() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	statusURL := testrig.URLMustParse("https://turnip.farm/users/turniplover6969/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042")
	status, _, err := suite.dereferencer.GetStatus(context.Background(), fetchingAccount.Username, statusURL, false, false)
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
	suite.True(*dbStatus.Boostable)
	suite.True(*dbStatus.Replyable)
	suite.True(*dbStatus.Likeable)

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

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}
