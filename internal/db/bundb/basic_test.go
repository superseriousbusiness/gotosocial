/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type BasicTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *BasicTestSuite) TestGetAccountByID() {
	testAccount := suite.testAccounts["local_account_1"]

	a := &gtsmodel.Account{}
	err := suite.db.GetByID(context.Background(), testAccount.ID, a)
	suite.NoError(err)
}

func (suite *BasicTestSuite) TestPutAccountWithBunDefaultFields() {
	testAccount := &gtsmodel.Account{
		ID:           "01GADR1AH9VCKH8YYCM86XSZ00",
		Username:     "test",
		URI:          "https://example.org/users/test",
		URL:          "https://example.org/@test",
		InboxURI:     "https://example.org/users/test/inbox",
		OutboxURI:    "https://example.org/users/test/outbox",
		ActorType:    "Person",
		PublicKeyURI: "https://example.org/test#main-key",
	}

	if err := suite.db.Put(context.Background(), testAccount); err != nil {
		suite.FailNow(err.Error())
	}

	a := &gtsmodel.Account{}
	if err := suite.db.GetByID(context.Background(), testAccount.ID, a); err != nil {
		suite.FailNow(err.Error())
	}

	// check all fields are set as expected, including database defaults
	suite.Equal(testAccount.ID, a.ID)
	suite.WithinDuration(time.Now(), a.CreatedAt, 5*time.Second)
	suite.WithinDuration(time.Now(), a.UpdatedAt, 5*time.Second)
	suite.Equal(testAccount.Username, a.Username)
	suite.Empty(a.Domain)
	suite.Empty(a.AvatarMediaAttachmentID)
	suite.Nil(a.AvatarMediaAttachment)
	suite.Empty(a.AvatarRemoteURL)
	suite.Empty(a.HeaderMediaAttachmentID)
	suite.Nil(a.HeaderMediaAttachment)
	suite.Empty(a.HeaderRemoteURL)
	suite.Empty(a.DisplayName)
	suite.Nil(a.Fields)
	suite.Empty(a.Note)
	suite.Empty(a.NoteRaw)
	suite.False(*a.Memorial)
	suite.Empty(a.AlsoKnownAs)
	suite.Empty(a.MovedToAccountID)
	suite.False(*a.Bot)
	suite.Empty(a.Reason)
	// Locked is especially important, since it's a bool that defaults
	// to true, which is why we use pointers for bools in the first place
	suite.True(*a.Locked)
	suite.False(*a.Discoverable)
	suite.Empty(a.Privacy)
	suite.False(*a.Sensitive)
	suite.Equal("en", a.Language)
	suite.Empty(a.StatusFormat)
	suite.Equal(testAccount.URI, a.URI)
	suite.Equal(testAccount.URL, a.URL)
	suite.Zero(testAccount.FetchedAt)
	suite.Equal(testAccount.InboxURI, a.InboxURI)
	suite.Equal(testAccount.OutboxURI, a.OutboxURI)
	suite.Empty(a.FollowingURI)
	suite.Empty(a.FollowersURI)
	suite.Empty(a.FeaturedCollectionURI)
	suite.Equal(testAccount.ActorType, a.ActorType)
	suite.Nil(a.PrivateKey)
	suite.Nil(a.PublicKey)
	suite.Equal(testAccount.PublicKeyURI, a.PublicKeyURI)
	suite.Zero(a.SensitizedAt)
	suite.Zero(a.SilencedAt)
	suite.Zero(a.SuspendedAt)
	suite.False(*a.HideCollections)
	suite.Empty(a.SuspensionOrigin)
}

func (suite *BasicTestSuite) TestGetAllStatuses() {
	s := []*gtsmodel.Status{}
	err := suite.db.GetAll(context.Background(), &s)
	suite.NoError(err)
	suite.Len(s, 17)
}

func (suite *BasicTestSuite) TestGetAllNotNull() {
	where := []db.Where{{
		Key:   "domain",
		Value: nil,
		Not:   true,
	}}

	a := []*gtsmodel.Account{}

	err := suite.db.GetWhere(context.Background(), where, &a)
	suite.NoError(err)
	suite.NotEmpty(a)

	for _, acct := range a {
		suite.NotEmpty(acct.Domain)
	}
}

func TestBasicTestSuite(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}
