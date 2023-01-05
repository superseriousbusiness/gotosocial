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

package validate_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

func happyAccount() *gtsmodel.Account {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	pub := &priv.PublicKey

	return &gtsmodel.Account{
		ID:                      "01F8MH1H7YV1Z7D2C8K2730QBF",
		CreatedAt:               time.Now().Add(-48 * time.Hour),
		UpdatedAt:               time.Now().Add(-48 * time.Hour),
		Username:                "the_mighty_zork",
		Domain:                  "",
		AvatarMediaAttachmentID: "01F8MH58A357CV5K7R7TJMSH6S",
		AvatarMediaAttachment:   nil,
		AvatarRemoteURL:         "",
		HeaderMediaAttachmentID: "01PFPMWK2FF0D9WMHEJHR07C3Q",
		HeaderMediaAttachment:   nil,
		HeaderRemoteURL:         "",
		DisplayName:             "original zork (he/they)",
		Fields:                  []gtsmodel.Field{},
		Note:                    "hey yo this is my profile!",
		Memorial:                testrig.FalseBool(),
		AlsoKnownAs:             "",
		MovedToAccountID:        "",
		Bot:                     testrig.FalseBool(),
		Reason:                  "I wanna be on this damned webbed site so bad! Please! Wow",
		Locked:                  testrig.FalseBool(),
		Discoverable:            testrig.TrueBool(),
		Privacy:                 gtsmodel.VisibilityPublic,
		Sensitive:               testrig.FalseBool(),
		Language:                "en",
		StatusFormat:            "plain",
		URI:                     "http://localhost:8080/users/the_mighty_zork",
		URL:                     "http://localhost:8080/@the_mighty_zork",
		LastWebfingeredAt:       time.Time{},
		InboxURI:                "http://localhost:8080/users/the_mighty_zork/inbox",
		OutboxURI:               "http://localhost:8080/users/the_mighty_zork/outbox",
		FollowersURI:            "http://localhost:8080/users/the_mighty_zork/followers",
		FollowingURI:            "http://localhost:8080/users/the_mighty_zork/following",
		FeaturedCollectionURI:   "http://localhost:8080/users/the_mighty_zork/collections/featured",
		ActorType:               ap.ActorPerson,
		PrivateKey:              priv,
		PublicKey:               pub,
		PublicKeyURI:            "http://localhost:8080/users/the_mighty_zork#main-key",
		SensitizedAt:            time.Time{},
		SilencedAt:              time.Time{},
		SuspendedAt:             time.Time{},
		HideCollections:         testrig.FalseBool(),
		SuspensionOrigin:        "",
	}
}

type AccountValidateTestSuite struct {
	suite.Suite
}

func (suite *AccountValidateTestSuite) TestValidateAccountHappyPath() {
	// no problem here
	a := happyAccount()
	err := validate.Struct(*a)
	suite.NoError(err)
}

// ID must be set and be valid ULID
func (suite *AccountValidateTestSuite) TestValidateAccountBadID() {
	a := happyAccount()

	a.ID = ""
	err := validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	a.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

// CreatedAt can be set or not -- it will be set in the database anyway
func (suite *AccountValidateTestSuite) TestValidateAccountNoCreatedAt() {
	a := happyAccount()

	a.CreatedAt = time.Time{}
	err := validate.Struct(*a)
	suite.NoError(err)
}

// LastWebfingeredAt must be defined if remote account
func (suite *AccountValidateTestSuite) TestValidateAccountNoWebfingeredAt() {
	a := happyAccount()

	a.Domain = "example.org"
	a.LastWebfingeredAt = time.Time{}
	err := validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.LastWebfingeredAt' Error:Field validation for 'LastWebfingeredAt' failed on the 'required_with' tag")
}

// Username must be set
func (suite *AccountValidateTestSuite) TestValidateAccountUsername() {
	a := happyAccount()

	a.Username = ""
	err := validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.Username' Error:Field validation for 'Username' failed on the 'required' tag")
}

// Domain must be either empty (for local accounts) or proper fqdn (for remote accounts)
func (suite *AccountValidateTestSuite) TestValidateAccountDomain() {
	a := happyAccount()
	a.LastWebfingeredAt = time.Now()

	a.Domain = ""
	err := validate.Struct(*a)
	suite.NoError(err)

	a.Domain = "localhost:8080"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.Domain' Error:Field validation for 'Domain' failed on the 'fqdn' tag")

	a.Domain = "ahhhhh"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.Domain' Error:Field validation for 'Domain' failed on the 'fqdn' tag")

	a.Domain = "https://www.example.org"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.Domain' Error:Field validation for 'Domain' failed on the 'fqdn' tag")

	a.Domain = "example.org:8080"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.Domain' Error:Field validation for 'Domain' failed on the 'fqdn' tag")

	a.Domain = "example.org"
	err = validate.Struct(*a)
	suite.NoError(err)
}

// Attachment IDs must either be not set, or must be valid ULID
func (suite *AccountValidateTestSuite) TestValidateAttachmentIDs() {
	a := happyAccount()

	a.AvatarMediaAttachmentID = ""
	a.HeaderMediaAttachmentID = ""
	err := validate.Struct(*a)
	suite.NoError(err)

	a.AvatarMediaAttachmentID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	a.HeaderMediaAttachmentID = "aaaa"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.AvatarMediaAttachmentID' Error:Field validation for 'AvatarMediaAttachmentID' failed on the 'ulid' tag\nKey: 'Account.HeaderMediaAttachmentID' Error:Field validation for 'HeaderMediaAttachmentID' failed on the 'ulid' tag")
}

// Attachment remote URLs must either not be set, or be valid URLs
func (suite *AccountValidateTestSuite) TestValidateAttachmentRemoteURLs() {
	a := happyAccount()

	a.AvatarRemoteURL = ""
	a.HeaderRemoteURL = ""
	err := validate.Struct(*a)
	suite.NoError(err)

	a.AvatarRemoteURL = "-------------"
	a.HeaderRemoteURL = "https://valid-url.com"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.AvatarRemoteURL' Error:Field validation for 'AvatarRemoteURL' failed on the 'url' tag")

	a.AvatarRemoteURL = "https://valid-url.com"
	a.HeaderRemoteURL = ""
	err = validate.Struct(*a)
	suite.NoError(err)
}

// Default privacy must be set if account is local
func (suite *AccountValidateTestSuite) TestValidatePrivacy() {
	a := happyAccount()
	a.LastWebfingeredAt = time.Now()

	a.Privacy = ""
	err := validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.Privacy' Error:Field validation for 'Privacy' failed on the 'required_without' tag")

	a.Privacy = "not valid"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.Privacy' Error:Field validation for 'Privacy' failed on the 'oneof' tag")

	a.Privacy = gtsmodel.VisibilityFollowersOnly
	err = validate.Struct(*a)
	suite.NoError(err)

	a.Privacy = ""
	a.Domain = "example.org"
	err = validate.Struct(*a)
	suite.NoError(err)

	a.Privacy = "invalid"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.Privacy' Error:Field validation for 'Privacy' failed on the 'oneof' tag")
}

// If set, language must be a valid language
func (suite *AccountValidateTestSuite) TestValidateLanguage() {
	a := happyAccount()

	a.Language = ""
	err := validate.Struct(*a)
	suite.NoError(err)

	a.Language = "not valid"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.Language' Error:Field validation for 'Language' failed on the 'bcp47_language_tag' tag")

	a.Language = "en-uk"
	err = validate.Struct(*a)
	suite.NoError(err)
}

// Account URI must be set and must be valid
func (suite *AccountValidateTestSuite) TestValidateAccountURI() {
	a := happyAccount()

	a.URI = "invalid-uri"
	err := validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.URI' Error:Field validation for 'URI' failed on the 'url' tag")

	a.URI = ""
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.URI' Error:Field validation for 'URI' failed on the 'required' tag")
}

// ActivityPub URIs must be set on account if it's local
func (suite *AccountValidateTestSuite) TestValidateAccountURIs() {
	a := happyAccount()
	a.LastWebfingeredAt = time.Now()

	a.InboxURI = "invalid-uri"
	a.OutboxURI = "invalid-uri"
	a.FollowersURI = "invalid-uri"
	a.FollowingURI = "invalid-uri"
	a.FeaturedCollectionURI = "invalid-uri"
	a.PublicKeyURI = "invalid-uri"
	err := validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.InboxURI' Error:Field validation for 'InboxURI' failed on the 'url' tag\nKey: 'Account.OutboxURI' Error:Field validation for 'OutboxURI' failed on the 'url' tag\nKey: 'Account.FollowingURI' Error:Field validation for 'FollowingURI' failed on the 'url' tag\nKey: 'Account.FollowersURI' Error:Field validation for 'FollowersURI' failed on the 'url' tag\nKey: 'Account.FeaturedCollectionURI' Error:Field validation for 'FeaturedCollectionURI' failed on the 'url' tag\nKey: 'Account.PublicKeyURI' Error:Field validation for 'PublicKeyURI' failed on the 'url' tag")

	a.InboxURI = ""
	a.OutboxURI = ""
	a.FollowersURI = ""
	a.FollowingURI = ""
	a.FeaturedCollectionURI = ""
	a.PublicKeyURI = ""
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.InboxURI' Error:Field validation for 'InboxURI' failed on the 'required_without' tag\nKey: 'Account.OutboxURI' Error:Field validation for 'OutboxURI' failed on the 'required_without' tag\nKey: 'Account.FollowingURI' Error:Field validation for 'FollowingURI' failed on the 'required_without' tag\nKey: 'Account.FollowersURI' Error:Field validation for 'FollowersURI' failed on the 'required_without' tag\nKey: 'Account.FeaturedCollectionURI' Error:Field validation for 'FeaturedCollectionURI' failed on the 'required_without' tag\nKey: 'Account.PublicKeyURI' Error:Field validation for 'PublicKeyURI' failed on the 'required' tag")

	a.Domain = "example.org"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.PublicKeyURI' Error:Field validation for 'PublicKeyURI' failed on the 'required' tag")

	a.InboxURI = "invalid-uri"
	a.OutboxURI = "invalid-uri"
	a.FollowersURI = "invalid-uri"
	a.FollowingURI = "invalid-uri"
	a.FeaturedCollectionURI = "invalid-uri"
	a.PublicKeyURI = "invalid-uri"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.InboxURI' Error:Field validation for 'InboxURI' failed on the 'url' tag\nKey: 'Account.OutboxURI' Error:Field validation for 'OutboxURI' failed on the 'url' tag\nKey: 'Account.FollowingURI' Error:Field validation for 'FollowingURI' failed on the 'url' tag\nKey: 'Account.FollowersURI' Error:Field validation for 'FollowersURI' failed on the 'url' tag\nKey: 'Account.FeaturedCollectionURI' Error:Field validation for 'FeaturedCollectionURI' failed on the 'url' tag\nKey: 'Account.PublicKeyURI' Error:Field validation for 'PublicKeyURI' failed on the 'url' tag")
}

// Actor type must be set and valid
func (suite *AccountValidateTestSuite) TestValidateActorType() {
	a := happyAccount()

	a.ActorType = ""
	err := validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.ActorType' Error:Field validation for 'ActorType' failed on the 'oneof' tag")

	a.ActorType = "not valid"
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.ActorType' Error:Field validation for 'ActorType' failed on the 'oneof' tag")

	a.ActorType = ap.ActivityArrive
	err = validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.ActorType' Error:Field validation for 'ActorType' failed on the 'oneof' tag")

	a.ActorType = ap.ActorOrganization
	err = validate.Struct(*a)
	suite.NoError(err)
}

// Private key must be set on local accounts
func (suite *AccountValidateTestSuite) TestValidatePrivateKey() {
	a := happyAccount()
	a.LastWebfingeredAt = time.Now()

	a.PrivateKey = nil
	err := validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.PrivateKey' Error:Field validation for 'PrivateKey' failed on the 'required_without' tag")

	a.Domain = "example.org"
	err = validate.Struct(*a)
	suite.NoError(err)
}

// Public key must be set
func (suite *AccountValidateTestSuite) TestValidatePublicKey() {
	a := happyAccount()

	a.PublicKey = nil
	err := validate.Struct(*a)
	suite.EqualError(err, "Key: 'Account.PublicKey' Error:Field validation for 'PublicKey' failed on the 'required' tag")
}

func TestAccountValidateTestSuite(t *testing.T) {
	suite.Run(t, new(AccountValidateTestSuite))
}
