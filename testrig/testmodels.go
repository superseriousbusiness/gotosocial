/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package testrig

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// NewTestTokens returns a map of tokens keyed according to which account the token belongs to.
func NewTestTokens() map[string]*gtsmodel.Token {
	tokens := map[string]*gtsmodel.Token{
		"local_account_1": {
			ID:              "01F8MGTQW4DKTDF8SW5CT9HYGA",
			ClientID:        "01F8MGV8AC3NGSJW0FE8W1BV70",
			UserID:          "01F8MGVGPHQ2D3P3X0454H54Z5",
			RedirectURI:     "http://localhost:8080",
			Scope:           "read write follow push",
			Access:          "NZAZOTC0OWITMDU0NC0ZODG4LWE4NJITMWUXM2M4MTRHZDEX",
			AccessCreateAt:  time.Now(),
			AccessExpiresAt: time.Now().Add(72 * time.Hour),
		},
		"local_account_2": {
			ID:              "01F8MGVVM1EDVYET710J27XY5R",
			ClientID:        "01F8MGW47HN8ZXNHNZ7E47CDMQ",
			UserID:          "01F8MH1VYJAE00TVVGMM5JNJ8X",
			RedirectURI:     "http://localhost:8080",
			Scope:           "read write follow push",
			Access:          "PIPINALKNNNFNF98717NAMNAMNFKIJKJ881818KJKJAKJJJA",
			AccessCreateAt:  time.Now(),
			AccessExpiresAt: time.Now().Add(72 * time.Hour),
		},
	}
	return tokens
}

// NewTestClients returns a map of Clients keyed according to which account they are used by.
func NewTestClients() map[string]*gtsmodel.Client {
	clients := map[string]*gtsmodel.Client{
		"admin_account": {
			ID:     "01F8MGWSJCND9BWBD4WGJXBM93",
			Secret: "dda8e835-2c9c-4bd2-9b8b-77c2e26d7a7a",
			Domain: "http://localhost:8080",
			UserID: "01F8MGWYWKVKS3VS8DV1AMYPGE", // admin_account
		},
		"local_account_1": {
			ID:     "01F8MGV8AC3NGSJW0FE8W1BV70",
			Secret: "c3724c74-dc3b-41b2-a108-0ea3d8399830",
			Domain: "http://localhost:8080",
			UserID: "01F8MGVGPHQ2D3P3X0454H54Z5", // local_account_1
		},
		"local_account_2": {
			ID:     "01F8MGW47HN8ZXNHNZ7E47CDMQ",
			Secret: "8f5603a5-c721-46cd-8f1b-2e368f51379f",
			Domain: "http://localhost:8080",
			UserID: "01F8MH1VYJAE00TVVGMM5JNJ8X", // local_account_2
		},
	}
	return clients
}

// NewTestApplications returns a map of applications keyed to which number application they are.
func NewTestApplications() map[string]*gtsmodel.Application {
	apps := map[string]*gtsmodel.Application{
		"admin_account": {
			ID:           "01F8MGXQRHYF5QPMTMXP78QC2F",
			Name:         "superseriousbusiness",
			Website:      "https://superserious.business",
			RedirectURI:  "http://localhost:8080",
			ClientID:     "01F8MGWSJCND9BWBD4WGJXBM93",           // admin client
			ClientSecret: "dda8e835-2c9c-4bd2-9b8b-77c2e26d7a7a", // admin client
			Scopes:       "read write follow push",
		},
		"application_1": {
			ID:           "01F8MGY43H3N2C8EWPR2FPYEXG",
			Name:         "really cool gts application",
			Website:      "https://reallycool.app",
			RedirectURI:  "http://localhost:8080",
			ClientID:     "01F8MGV8AC3NGSJW0FE8W1BV70",           // client_1
			ClientSecret: "c3724c74-dc3b-41b2-a108-0ea3d8399830", // client_1
			Scopes:       "read write follow push",
		},
		"application_2": {
			ID:           "01F8MGYG9E893WRHW0TAEXR8GJ",
			Name:         "kindaweird",
			Website:      "https://kindaweird.app",
			RedirectURI:  "http://localhost:8080",
			ClientID:     "01F8MGW47HN8ZXNHNZ7E47CDMQ",           // client_2
			ClientSecret: "8f5603a5-c721-46cd-8f1b-2e368f51379f", // client_2
			Scopes:       "read write follow push",
		},
	}
	return apps
}

// NewTestUsers returns a map of Users keyed by which account belongs to them.
func NewTestUsers() map[string]*gtsmodel.User {
	users := map[string]*gtsmodel.User{
		"unconfirmed_account": {
			ID:                     "01F8MGYG9E893WRHW0TAEXR8GJ",
			Email:                  "",
			AccountID:              "01F8MH0BBE4FHXPH513MBVFHB0",
			EncryptedPassword:      "$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS", // 'password'
			CreatedAt:              time.Now(),
			SignUpIP:               net.ParseIP("199.222.111.89"),
			UpdatedAt:              time.Time{},
			CurrentSignInAt:        time.Time{},
			CurrentSignInIP:        nil,
			LastSignInAt:           time.Time{},
			LastSignInIP:           nil,
			SignInCount:            0,
			InviteID:               "",
			ChosenLanguages:        []string{},
			FilteredLanguages:      []string{},
			Locale:                 "en",
			CreatedByApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			LastEmailedAt:          time.Time{},
			ConfirmationToken:      "a5a280bd-34be-44a3-8330-a57eaf61b8dd",
			ConfirmedAt:            time.Time{},
			ConfirmationSentAt:     time.Now(),
			UnconfirmedEmail:       "weed_lord420@example.org",
			Moderator:              false,
			Admin:                  false,
			Disabled:               false,
			Approved:               false,
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
		"admin_account": {
			ID:                     "01F8MGWYWKVKS3VS8DV1AMYPGE",
			Email:                  "admin@example.org",
			AccountID:              "01F8MH17FWEB39HZJ76B6VXSKF",
			EncryptedPassword:      "$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS", // 'password'
			CreatedAt:              time.Now().Add(-72 * time.Hour),
			SignUpIP:               net.ParseIP("89.22.189.19"),
			UpdatedAt:              time.Now().Add(-72 * time.Hour),
			CurrentSignInAt:        time.Now().Add(-10 * time.Minute),
			CurrentSignInIP:        net.ParseIP("89.122.255.1"),
			LastSignInAt:           time.Now().Add(-2 * time.Hour),
			LastSignInIP:           net.ParseIP("89.122.255.1"),
			SignInCount:            78,
			InviteID:               "",
			ChosenLanguages:        []string{"en"},
			FilteredLanguages:      []string{},
			Locale:                 "en",
			CreatedByApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
			LastEmailedAt:          time.Now().Add(-30 * time.Minute),
			ConfirmationToken:      "",
			ConfirmedAt:            time.Now().Add(-72 * time.Hour),
			ConfirmationSentAt:     time.Time{},
			UnconfirmedEmail:       "",
			Moderator:              true,
			Admin:                  true,
			Disabled:               false,
			Approved:               true,
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
		"local_account_1": {
			ID:                     "01F8MGVGPHQ2D3P3X0454H54Z5",
			Email:                  "zork@example.org",
			AccountID:              "01F8MH1H7YV1Z7D2C8K2730QBF",
			EncryptedPassword:      "$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS", // 'password'
			CreatedAt:              time.Now().Add(-36 * time.Hour),
			SignUpIP:               net.ParseIP("59.99.19.172"),
			UpdatedAt:              time.Now().Add(-72 * time.Hour),
			CurrentSignInAt:        time.Now().Add(-30 * time.Minute),
			CurrentSignInIP:        net.ParseIP("88.234.118.16"),
			LastSignInAt:           time.Now().Add(-2 * time.Hour),
			LastSignInIP:           net.ParseIP("147.111.231.154"),
			SignInCount:            9,
			InviteID:               "",
			ChosenLanguages:        []string{"en"},
			FilteredLanguages:      []string{},
			Locale:                 "en",
			CreatedByApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			LastEmailedAt:          time.Now().Add(-55 * time.Minute),
			ConfirmationToken:      "",
			ConfirmedAt:            time.Now().Add(-34 * time.Hour),
			ConfirmationSentAt:     time.Now().Add(-36 * time.Hour),
			UnconfirmedEmail:       "",
			Moderator:              false,
			Admin:                  false,
			Disabled:               false,
			Approved:               true,
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
		"local_account_2": {
			ID:                     "01F8MH1VYJAE00TVVGMM5JNJ8X",
			Email:                  "tortle.dude@example.org",
			AccountID:              "01F8MH5NBDF2MV7CTC4Q5128HF",
			EncryptedPassword:      "$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS", // 'password'
			CreatedAt:              time.Now().Add(-36 * time.Hour),
			SignUpIP:               net.ParseIP("59.99.19.172"),
			UpdatedAt:              time.Now().Add(-72 * time.Hour),
			CurrentSignInAt:        time.Now().Add(-30 * time.Minute),
			CurrentSignInIP:        net.ParseIP("118.44.18.196"),
			LastSignInAt:           time.Now().Add(-2 * time.Hour),
			LastSignInIP:           net.ParseIP("198.98.21.15"),
			SignInCount:            9,
			InviteID:               "",
			ChosenLanguages:        []string{"en"},
			FilteredLanguages:      []string{},
			Locale:                 "en",
			CreatedByApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			LastEmailedAt:          time.Now().Add(-55 * time.Minute),
			ConfirmationToken:      "",
			ConfirmedAt:            time.Now().Add(-34 * time.Hour),
			ConfirmationSentAt:     time.Now().Add(-36 * time.Hour),
			UnconfirmedEmail:       "",
			Moderator:              false,
			Admin:                  false,
			Disabled:               false,
			Approved:               true,
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
	}

	return users
}

// NewTestAccounts returns a map of accounts keyed by what type of account they are.
func NewTestAccounts() map[string]*gtsmodel.Account {
	accounts := map[string]*gtsmodel.Account{
		"unconfirmed_account": {
			ID:                      "01F8MH0BBE4FHXPH513MBVFHB0",
			Username:                "weed_lord420",
			AvatarMediaAttachmentID: "",
			HeaderMediaAttachmentID: "",
			DisplayName:             "",
			Fields:                  []gtsmodel.Field{},
			Note:                    "",
			Memorial:                false,
			MovedToAccountID:        "",
			CreatedAt:               time.Now(),
			UpdatedAt:               time.Now(),
			Bot:                     false,
			Reason:                  "hi, please let me in! I'm looking for somewhere neato bombeato to hang out.",
			Locked:                  false,
			Discoverable:            false,
			Privacy:                 gtsmodel.VisibilityPublic,
			Sensitive:               false,
			Language:                "en",
			URI:                     "http://localhost:8080/users/weed_lord420",
			URL:                     "http://localhost:8080/@weed_lord420",
			LastWebfingeredAt:       time.Time{},
			InboxURI:                "http://localhost:8080/users/weed_lord420/inbox",
			OutboxURI:               "http://localhost:8080/users/weed_lord420/outbox",
			FollowersURI:            "http://localhost:8080/users/weed_lord420/followers",
			FollowingURI:            "http://localhost:8080/users/weed_lord420/following",
			FeaturedCollectionURI:   "http://localhost:8080/users/weed_lord420/collections/featured",
			ActorType:               ap.ActorPerson,
			AlsoKnownAs:             "",
			PrivateKey:              &rsa.PrivateKey{},
			PublicKey:               &rsa.PublicKey{},
			PublicKeyURI:            "http://localhost:8080/users/weed_lord420#main-key",
			SensitizedAt:            time.Time{},
			SilencedAt:              time.Time{},
			SuspendedAt:             time.Time{},
			HideCollections:         false,
			SuspensionOrigin:        "",
		},
		"admin_account": {
			ID:                      "01F8MH17FWEB39HZJ76B6VXSKF",
			Username:                "admin",
			AvatarMediaAttachmentID: "",
			HeaderMediaAttachmentID: "",
			DisplayName:             "",
			Fields:                  []gtsmodel.Field{},
			Note:                    "",
			Memorial:                false,
			MovedToAccountID:        "",
			CreatedAt:               time.Now().Add(-72 * time.Hour),
			UpdatedAt:               time.Now().Add(-72 * time.Hour),
			Bot:                     false,
			Reason:                  "",
			Locked:                  false,
			Discoverable:            true,
			Privacy:                 gtsmodel.VisibilityPublic,
			Sensitive:               false,
			Language:                "en",
			URI:                     "http://localhost:8080/users/admin",
			URL:                     "http://localhost:8080/@admin",
			PublicKeyURI:            "http://localhost:8080/users/admin#main-key",
			LastWebfingeredAt:       time.Time{},
			InboxURI:                "http://localhost:8080/users/admin/inbox",
			OutboxURI:               "http://localhost:8080/users/admin/outbox",
			FollowersURI:            "http://localhost:8080/users/admin/followers",
			FollowingURI:            "http://localhost:8080/users/admin/following",
			FeaturedCollectionURI:   "http://localhost:8080/users/admin/collections/featured",
			ActorType:               ap.ActorPerson,
			AlsoKnownAs:             "",
			PrivateKey:              &rsa.PrivateKey{},
			PublicKey:               &rsa.PublicKey{},
			SensitizedAt:            time.Time{},
			SilencedAt:              time.Time{},
			SuspendedAt:             time.Time{},
			HideCollections:         false,
			SuspensionOrigin:        "",
		},
		"local_account_1": {
			ID:                      "01F8MH1H7YV1Z7D2C8K2730QBF",
			Username:                "the_mighty_zork",
			AvatarMediaAttachmentID: "01F8MH58A357CV5K7R7TJMSH6S",
			HeaderMediaAttachmentID: "01PFPMWK2FF0D9WMHEJHR07C3Q",
			DisplayName:             "original zork (he/they)",
			Fields:                  []gtsmodel.Field{},
			Note:                    "hey yo this is my profile!",
			Memorial:                false,
			MovedToAccountID:        "",
			CreatedAt:               time.Now().Add(-48 * time.Hour),
			UpdatedAt:               time.Now().Add(-48 * time.Hour),
			Bot:                     false,
			Reason:                  "I wanna be on this damned webbed site so bad! Please! Wow",
			Locked:                  false,
			Discoverable:            true,
			Privacy:                 gtsmodel.VisibilityPublic,
			Sensitive:               false,
			Language:                "en",
			URI:                     "http://localhost:8080/users/the_mighty_zork",
			URL:                     "http://localhost:8080/@the_mighty_zork",
			LastWebfingeredAt:       time.Time{},
			InboxURI:                "http://localhost:8080/users/the_mighty_zork/inbox",
			OutboxURI:               "http://localhost:8080/users/the_mighty_zork/outbox",
			FollowersURI:            "http://localhost:8080/users/the_mighty_zork/followers",
			FollowingURI:            "http://localhost:8080/users/the_mighty_zork/following",
			FeaturedCollectionURI:   "http://localhost:8080/users/the_mighty_zork/collections/featured",
			ActorType:               ap.ActorPerson,
			AlsoKnownAs:             "",
			PrivateKey:              &rsa.PrivateKey{},
			PublicKey:               &rsa.PublicKey{},
			PublicKeyURI:            "http://localhost:8080/users/the_mighty_zork#main-key",
			SensitizedAt:            time.Time{},
			SilencedAt:              time.Time{},
			SuspendedAt:             time.Time{},
			HideCollections:         false,
			SuspensionOrigin:        "",
		},
		"local_account_2": {
			ID:                      "01F8MH5NBDF2MV7CTC4Q5128HF",
			Username:                "1happyturtle",
			AvatarMediaAttachmentID: "",
			HeaderMediaAttachmentID: "",
			DisplayName:             "happy little turtle :3",
			Fields:                  []gtsmodel.Field{},
			Note:                    "i post about things that concern me",
			Memorial:                false,
			MovedToAccountID:        "",
			CreatedAt:               time.Now().Add(-190 * time.Hour),
			UpdatedAt:               time.Now().Add(-36 * time.Hour),
			Bot:                     false,
			Reason:                  "",
			Locked:                  true,
			Discoverable:            false,
			Privacy:                 gtsmodel.VisibilityFollowersOnly,
			Sensitive:               false,
			Language:                "en",
			URI:                     "http://localhost:8080/users/1happyturtle",
			URL:                     "http://localhost:8080/@1happyturtle",
			LastWebfingeredAt:       time.Time{},
			InboxURI:                "http://localhost:8080/users/1happyturtle/inbox",
			OutboxURI:               "http://localhost:8080/users/1happyturtle/outbox",
			FollowersURI:            "http://localhost:8080/users/1happyturtle/followers",
			FollowingURI:            "http://localhost:8080/users/1happyturtle/following",
			FeaturedCollectionURI:   "http://localhost:8080/users/1happyturtle/collections/featured",
			ActorType:               ap.ActorPerson,
			AlsoKnownAs:             "",
			PrivateKey:              &rsa.PrivateKey{},
			PublicKey:               &rsa.PublicKey{},
			PublicKeyURI:            "http://localhost:8080/users/1happyturtle#main-key",
			SensitizedAt:            time.Time{},
			SilencedAt:              time.Time{},
			SuspendedAt:             time.Time{},
			HideCollections:         false,
			SuspensionOrigin:        "",
		},
		"remote_account_1": {
			ID:                    "01F8MH5ZK5VRH73AKHQM6Y9VNX",
			Username:              "foss_satan",
			Domain:                "fossbros-anonymous.io",
			DisplayName:           "big gerald",
			Fields:                []gtsmodel.Field{},
			Note:                  "i post about like, i dunno, stuff, or whatever!!!!",
			Memorial:              false,
			MovedToAccountID:      "",
			CreatedAt:             TimeMustParse("2021-09-26T12:52:36+02:00"),
			UpdatedAt:             time.Now().Add(-36 * time.Hour),
			Bot:                   false,
			Locked:                false,
			Discoverable:          true,
			Sensitive:             false,
			Language:              "en",
			URI:                   "http://fossbros-anonymous.io/users/foss_satan",
			URL:                   "http://fossbros-anonymous.io/@foss_satan",
			LastWebfingeredAt:     time.Time{},
			InboxURI:              "http://fossbros-anonymous.io/users/foss_satan/inbox",
			OutboxURI:             "http://fossbros-anonymous.io/users/foss_satan/outbox",
			FollowersURI:          "http://fossbros-anonymous.io/users/foss_satan/followers",
			FollowingURI:          "http://fossbros-anonymous.io/users/foss_satan/following",
			FeaturedCollectionURI: "http://fossbros-anonymous.io/users/foss_satan/collections/featured",
			ActorType:             ap.ActorPerson,
			AlsoKnownAs:           "",
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			PublicKeyURI:          "http://fossbros-anonymous.io/users/foss_satan/main-key",
			SensitizedAt:          time.Time{},
			SilencedAt:            time.Time{},
			SuspendedAt:           time.Time{},
			HideCollections:       false,
			SuspensionOrigin:      "",
		},
		"remote_account_2": {
			ID:                    "01FHMQX3GAABWSM0S2VZEC2SWC",
			Username:              "some_user",
			Domain:                "example.org",
			DisplayName:           "some user",
			Fields:                []gtsmodel.Field{},
			Note:                  "i'm a real son of a gun",
			Memorial:              false,
			MovedToAccountID:      "",
			CreatedAt:             TimeMustParse("2020-08-10T14:13:28+02:00"),
			UpdatedAt:             time.Now().Add(-1 * time.Hour),
			Bot:                   false,
			Locked:                true,
			Discoverable:          true,
			Sensitive:             false,
			Language:              "en",
			URI:                   "http://example.org/users/some_user",
			URL:                   "http://example.org/@some_user",
			LastWebfingeredAt:     time.Time{},
			InboxURI:              "http://example.org/users/some_user/inbox",
			OutboxURI:             "http://example.org/users/some_user/outbox",
			FollowersURI:          "http://example.org/users/some_user/followers",
			FollowingURI:          "http://example.org/users/some_user/following",
			FeaturedCollectionURI: "http://example.org/users/some_user/collections/featured",
			ActorType:             ap.ActorPerson,
			AlsoKnownAs:           "",
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			PublicKeyURI:          "http://example.org/users/some_user#main-key",
			SensitizedAt:          time.Time{},
			SilencedAt:            time.Time{},
			SuspendedAt:           time.Time{},
			HideCollections:       false,
			SuspensionOrigin:      "",
		},
	}

	preserializedKeys := []string{
		"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",
	}

	if diff := len(accounts) - len(preserializedKeys); diff > 0 {
		var keyStrings = make([]string, diff)
		for i := 0; i < diff; i++ {
			priv, _ := rsa.GenerateKey(rand.Reader, 2048)
			key, _ := x509.MarshalPKCS8PrivateKey(priv)
			keyStrings = append(keyStrings, base64.StdEncoding.EncodeToString(key))
		}
		panic(fmt.Sprintf("mismatch between number of hardcoded test RSA keys and accounts used for test data. Insert the following generated key[s]: \n%+v", keyStrings))
	}

	// generate keys for each account
	i := 0
	for _, v := range accounts {
		premadeBytes, err := base64.StdEncoding.DecodeString(preserializedKeys[i])
		if err != nil {
			panic(err)
		}
		key, err := x509.ParsePKCS8PrivateKey(premadeBytes)
		if err != nil {
			panic(err)
		}
		priv, ok := key.(*rsa.PrivateKey)
		if !ok {
			panic(fmt.Sprintf("generated key at index %d is of incorrect type", i))
		}
		v.PrivateKey = priv
		v.PublicKey = &priv.PublicKey
		i++
	}
	return accounts
}

// NewTestAttachments returns a map of attachments keyed according to which account
// and status they belong to, and which attachment number of that status they are.
func NewTestAttachments() map[string]*gtsmodel.MediaAttachment {
	return map[string]*gtsmodel.MediaAttachment{
		"admin_account_status_1_attachment_1": {
			ID:        "01F8MH6NEM8D7527KZAECTCR76",
			StatusID:  "01F8MH75CBF9JFX4ZAD54N0W0R",
			URL:       "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpeg",
			RemoteURL: "",
			CreatedAt: time.Now().Add(-71 * time.Hour),
			UpdatedAt: time.Now().Add(-71 * time.Hour),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  1200,
					Height: 630,
					Size:   756000,
					Aspect: 1.9047619047619047,
				},
				Small: gtsmodel.Small{
					Width:  256,
					Height: 134,
					Size:   34304,
					Aspect: 1.9104477611940298,
				},
			},
			AccountID:         "01F8MH17FWEB39HZJ76B6VXSKF",
			Description:       "Black and white image of some 50's style text saying: Welcome On Board",
			ScheduledStatusID: "",
			Blurhash:          "LNJRdVM{00Rj%Mayt7j[4nWBofRj",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "/gotosocial/storage/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpeg",
				ContentType: "image/jpeg",
				FileSize:    62529,
				UpdatedAt:   time.Now().Add(-71 * time.Hour),
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "/gotosocial/storage/01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.jpeg",
				ContentType: "image/jpeg",
				FileSize:    6872,
				UpdatedAt:   time.Now().Add(-71 * time.Hour),
				URL:         "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.jpeg",
				RemoteURL:   "",
			},
			Avatar: false,
			Header: false,
		},
		"local_account_1_status_4_attachment_1": {
			ID:        "01F8MH7TDVANYKWVE8VVKFPJTJ",
			StatusID:  "01F8MH82FYRXD2RC6108DAJ5HB",
			URL:       "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH7TDVANYKWVE8VVKFPJTJ.gif",
			RemoteURL: "",
			CreatedAt: time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
			Type:      gtsmodel.FileTypeGif,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  400,
					Height: 280,
					Size:   756000,
					Aspect: 1.4285714285714286,
				},
				Small: gtsmodel.Small{
					Width:  256,
					Height: 179,
					Size:   45824,
					Aspect: 1.4301675977653632,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01F8MH1H7YV1Z7D2C8K2730QBF",
			Description:       "90's Trent Reznor turning to the camera",
			ScheduledStatusID: "",
			Blurhash:          "LEDara58O=t5EMSOENEN9]}?aK%0",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "/gotosocial/storage/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH7TDVANYKWVE8VVKFPJTJ.gif",
				ContentType: "image/gif",
				FileSize:    1109138,
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "/gotosocial/storage/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01F8MH7TDVANYKWVE8VVKFPJTJ.jpeg",
				ContentType: "image/jpeg",
				FileSize:    8803,
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
				URL:         "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01F8MH7TDVANYKWVE8VVKFPJTJ.jpeg",
				RemoteURL:   "",
			},
			Avatar: false,
			Header: false,
		},
		"local_account_1_unattached_1": {
			ID:        "01F8MH8RMYQ6MSNY3JM2XT1CQ5",
			StatusID:  "", // this attachment isn't connected to a status YET
			URL:       "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH8RMYQ6MSNY3JM2XT1CQ5.jpeg",
			RemoteURL: "",
			CreatedAt: time.Now().Add(30 * time.Second),
			UpdatedAt: time.Now().Add(30 * time.Second),
			Type:      gtsmodel.FileTypeGif,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  800,
					Height: 450,
					Size:   360000,
					Aspect: 1.7777777777777777,
				},
				Small: gtsmodel.Small{
					Width:  256,
					Height: 144,
					Size:   36864,
					Aspect: 1.7777777777777777,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01F8MH1H7YV1Z7D2C8K2730QBF",
			Description:       "the oh you meme",
			ScheduledStatusID: "",
			Blurhash:          "LSAd]9ogDge-R:M|j=xWIto0xXWX",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "/gotosocial/storage/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH8RMYQ6MSNY3JM2XT1CQ5.jpeg",
				ContentType: "image/jpeg",
				FileSize:    27759,
				UpdatedAt:   time.Now().Add(30 * time.Second),
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "/gotosocial/storage/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01F8MH8RMYQ6MSNY3JM2XT1CQ5.jpeg",
				ContentType: "image/jpeg",
				FileSize:    6177,
				UpdatedAt:   time.Now().Add(30 * time.Second),
				URL:         "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01F8MH8RMYQ6MSNY3JM2XT1CQ5.jpeg",
				RemoteURL:   "",
			},
			Avatar: false,
			Header: false,
		},
		"local_account_1_avatar": {
			ID:        "01F8MH58A357CV5K7R7TJMSH6S",
			StatusID:  "", // this attachment isn't connected to a status
			URL:       "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpeg",
			RemoteURL: "",
			CreatedAt: time.Now().Add(-47 * time.Hour),
			UpdatedAt: time.Now().Add(-47 * time.Hour),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  1092,
					Height: 1800,
					Size:   1965600,
					Aspect: 0.6066666666666667,
				},
				Small: gtsmodel.Small{
					Width:  155,
					Height: 256,
					Size:   39680,
					Aspect: 0.60546875,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01F8MH1H7YV1Z7D2C8K2730QBF",
			Description:       "a green goblin looking nasty",
			ScheduledStatusID: "",
			Blurhash:          "LKK9MT,p|YSNDkJ-5rsmvnwcOoe:",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "/gotosocial/storage/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpeg",
				ContentType: "image/jpeg",
				FileSize:    457680,
				UpdatedAt:   time.Now().Add(-47 * time.Hour),
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "/gotosocial/storage/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpeg",
				ContentType: "image/jpeg",
				FileSize:    15374,
				UpdatedAt:   time.Now().Add(-47 * time.Hour),
				URL:         "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpeg",
				RemoteURL:   "",
			},
			Avatar: true,
			Header: false,
		},
		"local_account_1_header": {
			ID:        "01PFPMWK2FF0D9WMHEJHR07C3Q",
			StatusID:  "",
			URL:       "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg",
			RemoteURL: "",
			CreatedAt: time.Now().Add(-47 * time.Hour),
			UpdatedAt: time.Now().Add(-47 * time.Hour),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  1018,
					Height: 764,
					Size:   777752,
					Aspect: 1.3324607329842932,
				},
				Small: gtsmodel.Small{
					Width:  256,
					Height: 192,
					Size:   49152,
					Aspect: 1.3333333333333333,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01F8MH1H7YV1Z7D2C8K2730QBF",
			Description:       "A very old-school screenshot of the original team fortress mod for quake ",
			ScheduledStatusID: "",
			Blurhash:          "L26j{^WCs+R-N}jsxWj@4;WWxDoK",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "/gotosocial/storage/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg",
				ContentType: "image/jpeg",
				FileSize:    517226,
				UpdatedAt:   time.Now().Add(-47 * time.Hour),
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "/gotosocial/storage/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg",
				ContentType: "image/jpeg",
				FileSize:    42308,
				UpdatedAt:   time.Now().Add(-47 * time.Hour),
				URL:         "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg",
				RemoteURL:   "",
			},
			Avatar: false,
			Header: true,
		},
	}
}

// NewTestEmojis returns a map of gts emojis, keyed by the emoji shortcode
func NewTestEmojis() map[string]*gtsmodel.Emoji {
	return map[string]*gtsmodel.Emoji{
		"rainbow": {
			ID:                     "01F8MH9H8E4VG3KDYJR9EGPXCQ",
			Shortcode:              "rainbow",
			Domain:                 "",
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
			ImageRemoteURL:         "",
			ImageStaticRemoteURL:   "",
			ImageURL:               "http://localhost:8080/fileserver/01F8MH261H1KSV3GW3016GZRY3/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
			ImagePath:              "/tmp/gotosocial/01F8MH261H1KSV3GW3016GZRY3/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
			ImageStaticURL:         "http://localhost:8080/fileserver/01F8MH261H1KSV3GW3016GZRY3/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
			ImageStaticPath:        "/tmp/gotosocial/01F8MH261H1KSV3GW3016GZRY3/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
			ImageContentType:       "image/png",
			ImageStaticContentType: "image/png",
			ImageFileSize:          36702,
			ImageStaticFileSize:    10413,
			ImageUpdatedAt:         time.Now(),
			Disabled:               false,
			URI:                    "http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ",
			VisibleInPicker:        true,
			CategoryID:             "",
		},
	}
}

func NewTestDomainBlocks() map[string]*gtsmodel.DomainBlock {
	return map[string]*gtsmodel.DomainBlock{
		"replyguys.com": {
			ID:                 "01FF22EQM7X8E3RX1XGPN7S87D",
			Domain:             "replyguys.com",
			CreatedByAccountID: "01F8MH17FWEB39HZJ76B6VXSKF",
			PrivateComment:     "i blocked this domain because they keep replying with pushy + unwarranted linux advice",
			PublicComment:      "reply-guying to tech posts",
			Obfuscate:          false,
		},
	}
}

type filenames struct {
	Original string
	Small    string
	Static   string
}

// newTestStoredAttachments returns a map of filenames, keyed according to which attachment they pertain to.
func newTestStoredAttachments() map[string]filenames {
	return map[string]filenames{
		"admin_account_status_1_attachment_1": {
			Original: "welcome-original.jpeg",
			Small:    "welcome-small.jpeg",
		},
		"local_account_1_status_4_attachment_1": {
			Original: "trent-original.gif",
			Small:    "trent-small.jpeg",
		},
		"local_account_1_unattached_1": {
			Original: "ohyou-original.jpeg",
			Small:    "ohyou-small.jpeg",
		},
		"local_account_1_avatar": {
			Original: "zork-original.jpeg",
			Small:    "zork-small.jpeg",
		},
		"local_account_1_header": {
			Original: "team-fortress-original.jpeg",
			Small:    "team-fortress-small.jpeg",
		},
	}
}

// newTestStoredEmoji returns a map of filenames, keyed according to which emoji they pertain to
func newTestStoredEmoji() map[string]filenames {
	return map[string]filenames{
		"rainbow": {
			Original: "rainbow-original.png",
			Static:   "rainbow-static.png",
		},
	}
}

// NewTestStatuses returns a map of statuses keyed according to which account
// and status they are.
func NewTestStatuses() map[string]*gtsmodel.Status {
	return map[string]*gtsmodel.Status{
		"admin_account_status_1": {
			ID:                       "01F8MH75CBF9JFX4ZAD54N0W0R",
			URI:                      "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
			URL:                      "http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
			Content:                  "hello world! #welcome ! first post on the instance :rainbow: !",
			AttachmentIDs:            []string{"01F8MH6NEM8D7527KZAECTCR76"},
			TagIDs:                   []string{"01F8MHA1A2NF9MJ3WCCQ3K8BSZ"},
			MentionIDs:               []string{},
			EmojiIDs:                 []string{"01F8MH9H8E4VG3KDYJR9EGPXCQ"},
			CreatedAt:                TimeMustParse("2021-10-20T11:36:45Z"),
			UpdatedAt:                TimeMustParse("2021-10-20T11:36:45Z"),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/admin",
			AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                false,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
			Federated:                true,
			Boostable:                true,
			Replyable:                true,
			Likeable:                 true,
			ActivityStreamsType:      ap.ObjectNote,
		},
		"admin_account_status_2": {
			ID:                       "01F8MHAAY43M6RJ473VQFCVH37",
			URI:                      "http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
			URL:                      "http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
			Content:                  "🐕🐕🐕🐕🐕",
			CreatedAt:                TimeMustParse("2021-10-20T12:36:45Z"),
			UpdatedAt:                TimeMustParse("2021-10-20T12:36:45Z"),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/admin",
			AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "open to see some puppies",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                true,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
			Federated:                true,
			Boostable:                true,
			Replyable:                true,
			Likeable:                 true,
			ActivityStreamsType:      ap.ObjectNote,
		},
		"admin_account_status_3": {
			ID:                       "01FF25D5Q0DH7CHD57CTRS6WK0",
			URI:                      "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
			URL:                      "http://localhost:8080/@admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
			Content:                  "hi @the_mighty_zork welcome to the instance!",
			CreatedAt:                time.Now().Add(-46 * time.Hour),
			UpdatedAt:                time.Now().Add(-46 * time.Hour),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/admin",
			MentionIDs:               []string{"01FF26A6BGEKCZFWNEHXB2ZZ6M"},
			AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
			InReplyToID:              "01F8MHAMCHF6Y650WCRSCP4WMY",
			InReplyToAccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToURI:             "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
			BoostOfID:                "",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                false,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
			Federated:                true,
			Boostable:                true,
			Replyable:                true,
			Likeable:                 true,
			ActivityStreamsType:      ap.ObjectNote,
		},
		"local_account_1_status_1": {
			ID:                       "01F8MHAMCHF6Y650WCRSCP4WMY",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
			Content:                  "hello everyone!",
			CreatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			UpdatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "introduction post",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                true,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                true,
			Boostable:                true,
			Replyable:                true,
			Likeable:                 true,
			ActivityStreamsType:      ap.ObjectNote,
		},
		"local_account_1_status_2": {
			ID:                       "01F8MHAYFKS4KMXF8K5Y1C0KRN",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAYFKS4KMXF8K5Y1C0KRN",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAYFKS4KMXF8K5Y1C0KRN",
			Content:                  "this is an unlocked local-only post that shouldn't federate, but it's still boostable, replyable, and likeable",
			CreatedAt:                time.Now().Add(-46 * time.Hour),
			UpdatedAt:                time.Now().Add(-46 * time.Hour),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "",
			Visibility:               gtsmodel.VisibilityUnlocked,
			Sensitive:                false,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                false,
			Boostable:                true,
			Replyable:                true,
			Likeable:                 true,
			ActivityStreamsType:      ap.ObjectNote,
		},
		"local_account_1_status_3": {
			ID:                       "01F8MHBBN8120SYH7D5S050MGK",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHBBN8120SYH7D5S050MGK",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01F8MHBBN8120SYH7D5S050MGK",
			Content:                  "this is a very personal post that I don't want anyone to interact with at all, and i only want mutuals to see it",
			CreatedAt:                time.Now().Add(-45 * time.Hour),
			UpdatedAt:                time.Now().Add(-45 * time.Hour),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "test: you shouldn't be able to interact with this post in any way",
			Visibility:               gtsmodel.VisibilityMutualsOnly,
			Sensitive:                false,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                true,
			Boostable:                false,
			Replyable:                false,
			Likeable:                 false,
			ActivityStreamsType:      ap.ObjectNote,
		},
		"local_account_1_status_4": {
			ID:                       "01F8MH82FYRXD2RC6108DAJ5HB",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01F8MH82FYRXD2RC6108DAJ5HB",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01F8MH82FYRXD2RC6108DAJ5HB",
			Content:                  "here's a little gif of trent",
			AttachmentIDs:            []string{"01F8MH7TDVANYKWVE8VVKFPJTJ"},
			CreatedAt:                time.Now().Add(-1 * time.Hour),
			UpdatedAt:                time.Now().Add(-1 * time.Hour),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "eye contact, trent reznor gif",
			Visibility:               gtsmodel.VisibilityMutualsOnly,
			Sensitive:                false,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                true,
			Boostable:                true,
			Replyable:                true,
			Likeable:                 true,
			ActivityStreamsType:      ap.ObjectNote,
		},
		"local_account_1_status_5": {
			ID:                       "01FCTA44PW9H1TB328S9AQXKDS",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01FCTA44PW9H1TB328S9AQXKDS",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01FCTA44PW9H1TB328S9AQXKDS",
			Content:                  "hi!",
			AttachmentIDs:            []string{},
			CreatedAt:                time.Now().Add(-1 * time.Minute),
			UpdatedAt:                time.Now().Add(-1 * time.Minute),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "",
			Visibility:               gtsmodel.VisibilityMutualsOnly,
			Sensitive:                false,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                true,
			Boostable:                true,
			Replyable:                true,
			Likeable:                 true,
			ActivityStreamsType:      ap.ObjectNote,
		},
		"local_account_2_status_1": {
			ID:                       "01F8MHBQCBTDKN6X5VHGMMN4MA",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01F8MHBQCBTDKN6X5VHGMMN4MA",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01F8MHBQCBTDKN6X5VHGMMN4MA",
			Content:                  "🐢 hi everyone i post about turtles 🐢",
			CreatedAt:                time.Now().Add(-189 * time.Hour),
			UpdatedAt:                time.Now().Add(-189 * time.Hour),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "introduction post",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                true,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                true,
			Boostable:                true,
			Replyable:                true,
			Likeable:                 true,
			ActivityStreamsType:      ap.ObjectNote,
		},
		"local_account_2_status_2": {
			ID:                       "01F8MHC0H0A7XHTVH5F596ZKBM",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01F8MHC0H0A7XHTVH5F596ZKBM",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01F8MHC0H0A7XHTVH5F596ZKBM",
			Content:                  "🐢 this one is federated, likeable, and boostable but not replyable 🐢",
			CreatedAt:                time.Now().Add(-1 * time.Minute),
			UpdatedAt:                time.Now().Add(-1 * time.Minute),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                true,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                true,
			Boostable:                true,
			Replyable:                false,
			Likeable:                 true,
			ActivityStreamsType:      ap.ObjectNote,
		},
		"local_account_2_status_3": {
			ID:                       "01F8MHC8VWDRBQR0N1BATDDEM5",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01F8MHC8VWDRBQR0N1BATDDEM5",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01F8MHC8VWDRBQR0N1BATDDEM5",
			Content:                  "🐢 i don't mind people sharing this one but I don't want likes or replies to it because cba🐢",
			CreatedAt:                time.Now().Add(-2 * time.Minute),
			UpdatedAt:                time.Now().Add(-2 * time.Minute),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "you won't be able to like or reply to this",
			Visibility:               gtsmodel.VisibilityUnlocked,
			Sensitive:                true,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                true,
			Boostable:                true,
			Replyable:                false,
			Likeable:                 false,
			ActivityStreamsType:      ap.ObjectNote,
		},
		"local_account_2_status_4": {
			ID:                       "01F8MHCP5P2NWYQ416SBA0XSEV",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01F8MHCP5P2NWYQ416SBA0XSEV",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01F8MHCP5P2NWYQ416SBA0XSEV",
			Content:                  "🐢 this is a public status but I want it local only and not boostable 🐢",
			CreatedAt:                time.Now().Add(-1 * time.Minute),
			UpdatedAt:                time.Now().Add(-1 * time.Minute),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                true,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                false,
			Boostable:                false,
			Replyable:                true,
			Likeable:                 true,

			ActivityStreamsType: ap.ObjectNote,
		},
		"local_account_2_status_5": {
			ID:                       "01FCQSQ667XHJ9AV9T27SJJSX5",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01FCQSQ667XHJ9AV9T27SJJSX5",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01FCQSQ667XHJ9AV9T27SJJSX5",
			Content:                  "🐢 @the_mighty_zork hi zork! 🐢",
			CreatedAt:                time.Now().Add(-1 * time.Minute),
			UpdatedAt:                time.Now().Add(-1 * time.Minute),
			Local:                    true,
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			MentionIDs:               []string{"01FDF2HM2NF6FSRZCDEDV451CN"},
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			InReplyToID:              "01F8MHAMCHF6Y650WCRSCP4WMY",
			InReplyToAccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToURI:             "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
			BoostOfID:                "",
			ContentWarning:           "",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                false,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                true,
			Boostable:                true,
			Replyable:                true,
			Likeable:                 true,
			ActivityStreamsType:      ap.ObjectNote,
		},
	}
}

// NewTestTags returns a map of gts model tags keyed by their name
func NewTestTags() map[string]*gtsmodel.Tag {
	return map[string]*gtsmodel.Tag{
		"welcome": {
			ID:                     "01F8MHA1A2NF9MJ3WCCQ3K8BSZ",
			URL:                    "http://localhost:8080/tags/welcome",
			Name:                   "welcome",
			FirstSeenFromAccountID: "",
			CreatedAt:              time.Now().Add(-71 * time.Hour),
			UpdatedAt:              time.Now().Add(-71 * time.Hour),
			Useable:                true,
			Listable:               true,
			LastStatusAt:           time.Now().Add(-71 * time.Hour),
		},
		"Hashtag": {
			ID:                     "01FCT9SGYA71487N8D0S1M638G",
			URL:                    "http://localhost:8080/tags/Hashtag",
			Name:                   "Hashtag",
			FirstSeenFromAccountID: "",
			CreatedAt:              time.Now().Add(-71 * time.Hour),
			UpdatedAt:              time.Now().Add(-71 * time.Hour),
			Useable:                true,
			Listable:               true,
			LastStatusAt:           time.Now().Add(-71 * time.Hour),
		},
	}
}

// NewTestMentions returns a map of gts model mentions keyed by their name.
func NewTestMentions() map[string]*gtsmodel.Mention {
	return map[string]*gtsmodel.Mention{
		"zork_mention_foss_satan": {
			ID:               "01FCTA2Y6FGHXQA4ZE6N5NMNEX",
			StatusID:         "01FCTA44PW9H1TB328S9AQXKDS",
			CreatedAt:        time.Now().Add(-1 * time.Minute),
			UpdatedAt:        time.Now().Add(-1 * time.Minute),
			OriginAccountID:  "01F8MH1H7YV1Z7D2C8K2730QBF",
			OriginAccountURI: "http://localhost:8080/users/the_mighty_zork",
			TargetAccountID:  "01F8MH5ZK5VRH73AKHQM6Y9VNX",
			NameString:       "@foss_satan@fossbros-anonymous.io",
			TargetAccountURI: "http://fossbros-anonymous.io/users/foss_satan",
			TargetAccountURL: "http://fossbros-anonymous.io/@foss_satan",
		},
		"local_user_2_mention_zork": {
			ID:               "01FDF2HM2NF6FSRZCDEDV451CN",
			StatusID:         "01FCQSQ667XHJ9AV9T27SJJSX5",
			CreatedAt:        time.Now().Add(-1 * time.Minute),
			UpdatedAt:        time.Now().Add(-1 * time.Minute),
			OriginAccountID:  "01F8MH5NBDF2MV7CTC4Q5128HF",
			OriginAccountURI: "http://localhost:8080/users/1happyturtle",
			TargetAccountID:  "01F8MH1H7YV1Z7D2C8K2730QBF",
			NameString:       "@the_mighty_zork",
			TargetAccountURI: "http://localhost:8080/users/the_mighty_zork",
			TargetAccountURL: "http://localhost:8080/@the_mighty_zork",
		},
		"admin_account_mention_zork": {
			ID:               "01FF26A6BGEKCZFWNEHXB2ZZ6M",
			StatusID:         "01FF25D5Q0DH7CHD57CTRS6WK0",
			CreatedAt:        time.Now().Add(-46 * time.Hour),
			UpdatedAt:        time.Now().Add(-46 * time.Hour),
			OriginAccountID:  "01F8MH17FWEB39HZJ76B6VXSKF",
			OriginAccountURI: "http://localhost:8080/users/admin",
			TargetAccountID:  "01F8MH1H7YV1Z7D2C8K2730QBF",
			NameString:       "@the_mighty_zork",
			TargetAccountURI: "http://localhost:8080/users/the_mighty_zork",
			TargetAccountURL: "http://localhost:8080/@the_mighty_zork",
		},
	}
}

// NewTestFaves returns a map of gts model faves, keyed in the format [faving_account]_[target_status]
func NewTestFaves() map[string]*gtsmodel.StatusFave {
	return map[string]*gtsmodel.StatusFave{
		"local_account_1_admin_account_status_1": {
			ID:              "01F8MHD2QCZSZ6WQS2ATVPEYJ9",
			CreatedAt:       time.Now().Add(-47 * time.Hour),
			AccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF", // local account 1
			TargetAccountID: "01F8MH17FWEB39HZJ76B6VXSKF", // admin account
			StatusID:        "01F8MH75CBF9JFX4ZAD54N0W0R", // admin account status 1
			URI:             "http://localhost:8080/users/the_mighty_zork/liked/01F8MHD2QCZSZ6WQS2ATVPEYJ9",
		},
		"admin_account_local_account_1_status_1": {
			ID:              "01F8Q0486ANTDWKG02A7DS1Q24",
			CreatedAt:       time.Now().Add(-46 * time.Hour),
			AccountID:       "01F8MH17FWEB39HZJ76B6VXSKF", // admin account
			TargetAccountID: "01F8MH1H7YV1Z7D2C8K2730QBF", // local account 1
			StatusID:        "01F8MHAMCHF6Y650WCRSCP4WMY", // local account status 1
			URI:             "http://localhost:8080/users/admin/liked/01F8Q0486ANTDWKG02A7DS1Q24",
		},
	}
}

// NewTestNotifications returns some notifications for use in testing.
func NewTestNotifications() map[string]*gtsmodel.Notification {
	return map[string]*gtsmodel.Notification{
		"local_account_1_like": {
			ID:               "01F8Q0ANPTWW10DAKTX7BRPBJP",
			NotificationType: gtsmodel.NotificationFave,
			CreatedAt:        time.Now().Add(-46 * time.Hour),
			TargetAccountID:  "01F8MH1H7YV1Z7D2C8K2730QBF",
			OriginAccountID:  "01F8MH17FWEB39HZJ76B6VXSKF",
			StatusID:         "01F8MHAMCHF6Y650WCRSCP4WMY",
			Read:             false,
		},
	}
}

// NewTestFollows returns some follows for use in testing.
func NewTestFollows() map[string]*gtsmodel.Follow {
	return map[string]*gtsmodel.Follow{
		"local_account_1_admin_account": {
			ID:              "01F8PY8RHWRQZV038T4E8T9YK8",
			CreatedAt:       time.Now().Add(-46 * time.Hour),
			UpdatedAt:       time.Now().Add(-46 * time.Hour),
			AccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF",
			TargetAccountID: "01F8MH17FWEB39HZJ76B6VXSKF",
			ShowReblogs:     true,
			URI:             "http://localhost:8080/users/the_mighty_zork/follow/01F8PY8RHWRQZV038T4E8T9YK8",
			Notify:          false,
		},
		"local_account_1_local_account_2": {
			ID:              "01F8PYDCE8XE23GRE5DPZJDZDP",
			CreatedAt:       time.Now().Add(-1 * time.Hour),
			UpdatedAt:       time.Now().Add(-1 * time.Hour),
			AccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF",
			TargetAccountID: "01F8MH5NBDF2MV7CTC4Q5128HF",
			ShowReblogs:     true,
			URI:             "http://localhost:8080/users/the_mighty_zork/follow/01F8PYDCE8XE23GRE5DPZJDZDP",
			Notify:          false,
		},
	}
}

func NewTestBlocks() map[string]*gtsmodel.Block {
	return map[string]*gtsmodel.Block{
		"local_account_2_block_remote_account_1": {
			ID:              "01FEXXET6XXMF7G2V3ASZP3YQW",
			CreatedAt:       time.Now().Add(-1 * time.Hour),
			UpdatedAt:       time.Now().Add(-1 * time.Hour),
			URI:             "http://localhost:8080/users/1happyturtle/blocks/01FEXXET6XXMF7G2V3ASZP3YQW",
			AccountID:       "01F8MH5NBDF2MV7CTC4Q5128HF",
			TargetAccountID: "01F8MH5ZK5VRH73AKHQM6Y9VNX",
		},
	}
}

// ActivityWithSignature wraps a pub.Activity along with its signature headers, for testing.
type ActivityWithSignature struct {
	Activity        pub.Activity
	SignatureHeader string
	DigestHeader    string
	DateHeader      string
}

// NewTestActivities returns a bunch of pub.Activity types for use in testing the federation protocols.
// A struct of accounts needs to be passed in because the activities will also be bundled along with
// their requesting signatures.
func NewTestActivities(accounts map[string]*gtsmodel.Account) map[string]ActivityWithSignature {
	dmForZork := newNote(
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan/statuses/5424b153-4553-4f30-9358-7b92f7cd42f6"),
		URLMustParse("http://fossbros-anonymous.io/@foss_satan/5424b153-4553-4f30-9358-7b92f7cd42f6"),
		time.Now(),
		"hey zork here's a new private note for you",
		"new note for zork",
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan"),
		[]*url.URL{URLMustParse("http://localhost:8080/users/the_mighty_zork")},
		nil,
		true,
		[]vocab.ActivityStreamsMention{})
	createDmForZork := wrapNoteInCreate(
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan/statuses/5424b153-4553-4f30-9358-7b92f7cd42f6/activity"),
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan"),
		time.Now(),
		dmForZork)
	createDmForZorkSig, createDmForZorkDigest, creatDmForZorkDate := GetSignatureForActivity(createDmForZork, accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, URLMustParse(accounts["local_account_1"].InboxURI))

	forwardedMessage := newNote(
		URLMustParse("http://example.org/users/some_user/statuses/afaba698-5740-4e32-a702-af61aa543bc1"),
		URLMustParse("http://example.org/@some_user/afaba698-5740-4e32-a702-af61aa543bc1"),
		time.Now(),
		"this is a public status, please forward it!",
		"",
		URLMustParse("http://example.org/users/some_user"),
		[]*url.URL{URLMustParse(pub.PublicActivityPubIRI)},
		nil,
		false,
		[]vocab.ActivityStreamsMention{})
	createForwardedMessage := wrapNoteInCreate(
		URLMustParse("http://example.org/users/some_user/statuses/afaba698-5740-4e32-a702-af61aa543bc1/activity"),
		URLMustParse("http://example.org/users/some_user"),
		time.Now(),
		forwardedMessage)
	createForwardedMessageSig, createForwardedMessageDigest, createForwardedMessageDate := GetSignatureForActivity(createForwardedMessage, accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, URLMustParse(accounts["local_account_1"].InboxURI))

	return map[string]ActivityWithSignature{
		"dm_for_zork": {
			Activity:        createDmForZork,
			SignatureHeader: createDmForZorkSig,
			DigestHeader:    createDmForZorkDigest,
			DateHeader:      creatDmForZorkDate,
		},
		"forwarded_message": {
			Activity:        createForwardedMessage,
			SignatureHeader: createForwardedMessageSig,
			DigestHeader:    createForwardedMessageDigest,
			DateHeader:      createForwardedMessageDate,
		},
	}
}

// NewTestFediPeople returns a bunch of activity pub Person representations for testing converters and so on.
func NewTestFediPeople() map[string]vocab.ActivityStreamsPerson {
	newPerson1Priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	newPerson1Pub := &newPerson1Priv.PublicKey

	return map[string]vocab.ActivityStreamsPerson{
		"https://unknown-instance.com/users/brand_new_person": newPerson(
			URLMustParse("https://unknown-instance.com/users/brand_new_person"),
			URLMustParse("https://unknown-instance.com/users/brand_new_person/following"),
			URLMustParse("https://unknown-instance.com/users/brand_new_person/followers"),
			URLMustParse("https://unknown-instance.com/users/brand_new_person/inbox"),
			URLMustParse("https://unknown-instance.com/users/brand_new_person/outbox"),
			URLMustParse("https://unknown-instance.com/users/brand_new_person/collections/featured"),
			"brand_new_person",
			"Geoff Brando New Personson",
			"hey I'm a new person, your instance hasn't seen me yet uwu",
			URLMustParse("https://unknown-instance.com/@brand_new_person"),
			true,
			URLMustParse("https://unknown-instance.com/users/brand_new_person#main-key"),
			newPerson1Pub,
			nil,
			"image/jpeg",
			nil,
			"image/png",
			false,
		),
	}
}

func NewTestFediGroups() map[string]vocab.ActivityStreamsGroup {
	newGroup1Priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	newGroup1Pub := &newGroup1Priv.PublicKey

	return map[string]vocab.ActivityStreamsGroup{
		"https://unknown-instance.com/groups/some_group": newGroup(
			URLMustParse("https://unknown-instance.com/groups/some_group"),
			URLMustParse("https://unknown-instance.com/groups/some_group/following"),
			URLMustParse("https://unknown-instance.com/groups/some_group/followers"),
			URLMustParse("https://unknown-instance.com/groups/some_group/inbox"),
			URLMustParse("https://unknown-instance.com/groups/some_group/outbox"),
			URLMustParse("https://unknown-instance.com/groups/some_group/collections/featured"),
			"some_group",
			"This is a group about... something?",
			"",
			URLMustParse("https://unknown-instance.com/@some_group"),
			true,
			URLMustParse("https://unknown-instance.com/groups/some_group#main-key"),
			newGroup1Pub,
			nil,
			"image/jpeg",
			nil,
			"image/png",
			false,
		),
	}
}

// RemoteAttachmentFile mimics a remote (federated) attachment
type RemoteAttachmentFile struct {
	Data        []byte
	ContentType string
}

func NewTestFediAttachments(relativePath string) map[string]RemoteAttachmentFile {
	beeBytes, err := os.ReadFile(fmt.Sprintf("%s/beeplushie.jpg", relativePath))
	if err != nil {
		panic(err)
	}
	return map[string]RemoteAttachmentFile{
		"https://s3-us-west-2.amazonaws.com/plushcity/media_attachments/files/106/867/380/219/163/828/original/88e8758c5f011439.jpg": {
			Data:        beeBytes,
			ContentType: "image/jpeg",
		},
	}
}

func NewTestFediStatuses() map[string]vocab.ActivityStreamsNote {
	return map[string]vocab.ActivityStreamsNote{
		"https://unknown-instance.com/users/brand_new_person/statuses/01FE4NTHKWW7THT67EF10EB839": newNote(
			URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01FE4NTHKWW7THT67EF10EB839"),
			URLMustParse("https://unknown-instance.com/users/@brand_new_person/01FE4NTHKWW7THT67EF10EB839"),
			time.Now(),
			"Hello world!",
			"",
			URLMustParse("https://unknown-instance.com/users/brand_new_person"),
			[]*url.URL{
				URLMustParse(pub.PublicActivityPubIRI),
			},
			[]*url.URL{},
			false,
			[]vocab.ActivityStreamsMention{},
		),
		"https://unknown-instance.com/users/brand_new_person/statuses/01FE5Y30E3W4P7TRE0R98KAYQV": newNote(
			URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01FE5Y30E3W4P7TRE0R98KAYQV"),
			URLMustParse("https://unknown-instance.com/users/@brand_new_person/01FE5Y30E3W4P7TRE0R98KAYQV"),
			time.Now(),
			"Hey @the_mighty_zork@localhost:8080 how's it going?",
			"",
			URLMustParse("https://unknown-instance.com/users/brand_new_person"),
			[]*url.URL{
				URLMustParse(pub.PublicActivityPubIRI),
			},
			[]*url.URL{},
			false,
			[]vocab.ActivityStreamsMention{
				newMention(
					URLMustParse("http://localhost:8080/users/the_mighty_zork"),
					"@the_mighty_zork@localhost:8080",
				),
			},
		),
	}
}

// NewTestDereferenceRequests returns a map of incoming dereference requests, with their signatures.
func NewTestDereferenceRequests(accounts map[string]*gtsmodel.Account) map[string]ActivityWithSignature {
	var sig, digest, date string
	var target *url.URL
	statuses := NewTestStatuses()

	target = URLMustParse(accounts["local_account_1"].URI)
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceZork := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(statuses["local_account_1_status_1"].URI + "/replies")
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceLocalAccount1Status1Replies := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(statuses["local_account_1_status_1"].URI + "/replies?only_other_accounts=false&page=true")
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceLocalAccount1Status1RepliesNext := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(statuses["local_account_1_status_1"].URI + "/replies?only_other_accounts=false&page=true&min_id=01FF25D5Q0DH7CHD57CTRS6WK0")
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceLocalAccount1Status1RepliesLast := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(accounts["local_account_1"].OutboxURI)
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceZorkOutbox := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(accounts["local_account_1"].OutboxURI + "?page=true")
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceZorkOutboxFirst := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(accounts["local_account_1"].OutboxURI + "?page=true&max_id=01F8MHAMCHF6Y650WCRSCP4WMY")
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceZorkOutboxNext := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	return map[string]ActivityWithSignature{
		"foss_satan_dereference_zork":                                  fossSatanDereferenceZork,
		"foss_satan_dereference_local_account_1_status_1_replies":      fossSatanDereferenceLocalAccount1Status1Replies,
		"foss_satan_dereference_local_account_1_status_1_replies_next": fossSatanDereferenceLocalAccount1Status1RepliesNext,
		"foss_satan_dereference_local_account_1_status_1_replies_last": fossSatanDereferenceLocalAccount1Status1RepliesLast,
		"foss_satan_dereference_zork_outbox":                           fossSatanDereferenceZorkOutbox,
		"foss_satan_dereference_zork_outbox_first":                     fossSatanDereferenceZorkOutboxFirst,
		"foss_satan_dereference_zork_outbox_next":                      fossSatanDereferenceZorkOutboxNext,
	}
}

// GetSignatureForActivity does some sneaky sneaky work with a mock http client and a test transport controller, in order to derive
// the HTTP Signature for the given activity, public key ID, private key, and destination.
func GetSignatureForActivity(activity pub.Activity, pubKeyID string, privkey crypto.PrivateKey, destination *url.URL) (signatureHeader string, digestHeader string, dateHeader string) {
	// create a client that basically just pulls the signature out of the request and sets it
	client := &mockHTTPClient{
		do: func(req *http.Request) (*http.Response, error) {
			signatureHeader = req.Header.Get("Signature")
			digestHeader = req.Header.Get("Digest")
			dateHeader = req.Header.Get("Date")
			r := ioutil.NopCloser(bytes.NewReader([]byte{})) // we only need this so the 'close' func doesn't nil out
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	}

	// use the client to create a new transport
	c := NewTestTransportController(client, NewTestDB())
	tp, err := c.NewTransport(pubKeyID, privkey)
	if err != nil {
		panic(err)
	}

	// convert the activity into json bytes
	m, err := activity.Serialize()
	if err != nil {
		panic(err)
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	// trigger the delivery function, which will trigger the 'do' function of the recorder above
	if err := tp.Deliver(context.Background(), bytes, destination); err != nil {
		panic(err)
	}

	// headers should now be populated
	return
}

// GetSignatureForDereference does some sneaky sneaky work with a mock http client and a test transport controller, in order to derive
// the HTTP Signature for the given derefence GET request using public key ID, private key, and destination.
func GetSignatureForDereference(pubKeyID string, privkey crypto.PrivateKey, destination *url.URL) (signatureHeader string, digestHeader string, dateHeader string) {
	// create a client that basically just pulls the signature out of the request and sets it
	client := &mockHTTPClient{
		do: func(req *http.Request) (*http.Response, error) {
			signatureHeader = req.Header.Get("Signature")
			dateHeader = req.Header.Get("Date")
			r := ioutil.NopCloser(bytes.NewReader([]byte{})) // we only need this so the 'close' func doesn't nil out
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	}

	// use the client to create a new transport
	c := NewTestTransportController(client, NewTestDB())
	tp, err := c.NewTransport(pubKeyID, privkey)
	if err != nil {
		panic(err)
	}

	// trigger the delivery function, which will trigger the 'do' function of the recorder above
	if _, err := tp.Dereference(context.Background(), destination); err != nil {
		panic(err)
	}

	// headers should now be populated
	return
}

func newPerson(
	profileIDURI *url.URL,
	followingURI *url.URL,
	followersURI *url.URL,
	inboxURI *url.URL,
	outboxURI *url.URL,
	featuredURI *url.URL,
	username string,
	displayName string,
	note string,
	profileURL *url.URL,
	discoverable bool,
	publicKeyURI *url.URL,
	pkey *rsa.PublicKey,
	avatarURL *url.URL,
	avatarContentType string,
	headerURL *url.URL,
	headerContentType string,
	manuallyApprovesFollowers bool) vocab.ActivityStreamsPerson {
	person := streams.NewActivityStreamsPerson()

	// id should be the activitypub URI of this user
	// something like https://example.org/users/example_user
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(profileIDURI)
	person.SetJSONLDId(idProp)

	// following
	// The URI for retrieving a list of accounts this user is following
	followingProp := streams.NewActivityStreamsFollowingProperty()
	followingProp.SetIRI(followingURI)
	person.SetActivityStreamsFollowing(followingProp)

	// followers
	// The URI for retrieving a list of this user's followers
	followersProp := streams.NewActivityStreamsFollowersProperty()
	followersProp.SetIRI(followersURI)
	person.SetActivityStreamsFollowers(followersProp)

	// inbox
	// the activitypub inbox of this user for accepting messages
	inboxProp := streams.NewActivityStreamsInboxProperty()
	inboxProp.SetIRI(inboxURI)
	person.SetActivityStreamsInbox(inboxProp)

	// outbox
	// the activitypub outbox of this user for serving messages
	outboxProp := streams.NewActivityStreamsOutboxProperty()
	outboxProp.SetIRI(outboxURI)
	person.SetActivityStreamsOutbox(outboxProp)

	// featured posts
	// Pinned posts.
	featuredProp := streams.NewTootFeaturedProperty()
	featuredProp.SetIRI(featuredURI)
	person.SetTootFeatured(featuredProp)

	// featuredTags
	// NOT IMPLEMENTED

	// preferredUsername
	// Used for Webfinger lookup. Must be unique on the domain, and must correspond to a Webfinger acct: URI.
	preferredUsernameProp := streams.NewActivityStreamsPreferredUsernameProperty()
	preferredUsernameProp.SetXMLSchemaString(username)
	person.SetActivityStreamsPreferredUsername(preferredUsernameProp)

	// name
	// Used as profile display name.
	nameProp := streams.NewActivityStreamsNameProperty()
	if displayName != "" {
		nameProp.AppendXMLSchemaString(displayName)
	} else {
		nameProp.AppendXMLSchemaString(username)
	}
	person.SetActivityStreamsName(nameProp)

	// summary
	// Used as profile bio.
	if note != "" {
		summaryProp := streams.NewActivityStreamsSummaryProperty()
		summaryProp.AppendXMLSchemaString(note)
		person.SetActivityStreamsSummary(summaryProp)
	}

	// url
	// Used as profile link.
	urlProp := streams.NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(profileURL)
	person.SetActivityStreamsUrl(urlProp)

	// manuallyApprovesFollowers
	manuallyApprovesFollowersProp := streams.NewActivityStreamsManuallyApprovesFollowersProperty()
	manuallyApprovesFollowersProp.Set(manuallyApprovesFollowers)
	person.SetActivityStreamsManuallyApprovesFollowers(manuallyApprovesFollowersProp)

	// discoverable
	// Will be shown in the profile directory.
	discoverableProp := streams.NewTootDiscoverableProperty()
	discoverableProp.Set(discoverable)
	person.SetTootDiscoverable(discoverableProp)

	// devices
	// NOT IMPLEMENTED, probably won't implement

	// alsoKnownAs
	// Required for Move activity.
	// TODO: NOT IMPLEMENTED **YET** -- this needs to be added as an activitypub extension to https://github.com/go-fed/activity, see https://github.com/go-fed/activity/tree/master/astool

	// publicKey
	// Required for signatures.
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
	encodedPublicKey, err := x509.MarshalPKIXPublicKey(pkey)
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

	// tag
	// TODO: Any tags used in the summary of this profile

	// attachment
	// Used for profile fields.
	// TODO: The PropertyValue type has to be added: https://schema.org/PropertyValue

	// endpoints
	// NOT IMPLEMENTED -- this is for shared inbox which we don't use

	// icon
	// Used as profile avatar.
	iconProperty := streams.NewActivityStreamsIconProperty()
	iconImage := streams.NewActivityStreamsImage()
	mediaType := streams.NewActivityStreamsMediaTypeProperty()
	mediaType.Set(avatarContentType)
	iconImage.SetActivityStreamsMediaType(mediaType)
	avatarURLProperty := streams.NewActivityStreamsUrlProperty()
	avatarURLProperty.AppendIRI(avatarURL)
	iconImage.SetActivityStreamsUrl(avatarURLProperty)
	iconProperty.AppendActivityStreamsImage(iconImage)
	person.SetActivityStreamsIcon(iconProperty)

	// image
	// Used as profile header.
	headerProperty := streams.NewActivityStreamsImageProperty()
	headerImage := streams.NewActivityStreamsImage()
	headerMediaType := streams.NewActivityStreamsMediaTypeProperty()
	mediaType.Set(headerContentType)
	headerImage.SetActivityStreamsMediaType(headerMediaType)
	headerURLProperty := streams.NewActivityStreamsUrlProperty()
	headerURLProperty.AppendIRI(headerURL)
	headerImage.SetActivityStreamsUrl(headerURLProperty)
	headerProperty.AppendActivityStreamsImage(headerImage)
	person.SetActivityStreamsImage(headerProperty)

	return person
}

func newGroup(
	profileIDURI *url.URL,
	followingURI *url.URL,
	followersURI *url.URL,
	inboxURI *url.URL,
	outboxURI *url.URL,
	featuredURI *url.URL,
	username string,
	displayName string,
	note string,
	profileURL *url.URL,
	discoverable bool,
	publicKeyURI *url.URL,
	pkey *rsa.PublicKey,
	avatarURL *url.URL,
	avatarContentType string,
	headerURL *url.URL,
	headerContentType string,
	manuallyApprovesFollowers bool) vocab.ActivityStreamsGroup {
	group := streams.NewActivityStreamsGroup()

	// id should be the activitypub URI of this group
	// something like https://example.org/users/example_group
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(profileIDURI)
	group.SetJSONLDId(idProp)

	// following
	// The URI for retrieving a list of accounts this group is following
	followingProp := streams.NewActivityStreamsFollowingProperty()
	followingProp.SetIRI(followingURI)
	group.SetActivityStreamsFollowing(followingProp)

	// followers
	// The URI for retrieving a list of this user's followers
	followersProp := streams.NewActivityStreamsFollowersProperty()
	followersProp.SetIRI(followersURI)
	group.SetActivityStreamsFollowers(followersProp)

	// inbox
	// the activitypub inbox of this user for accepting messages
	inboxProp := streams.NewActivityStreamsInboxProperty()
	inboxProp.SetIRI(inboxURI)
	group.SetActivityStreamsInbox(inboxProp)

	// outbox
	// the activitypub outbox of this user for serving messages
	outboxProp := streams.NewActivityStreamsOutboxProperty()
	outboxProp.SetIRI(outboxURI)
	group.SetActivityStreamsOutbox(outboxProp)

	// featured posts
	// Pinned posts.
	featuredProp := streams.NewTootFeaturedProperty()
	featuredProp.SetIRI(featuredURI)
	group.SetTootFeatured(featuredProp)

	// featuredTags
	// NOT IMPLEMENTED

	// preferredUsername
	// Used for Webfinger lookup. Must be unique on the domain, and must correspond to a Webfinger acct: URI.
	preferredUsernameProp := streams.NewActivityStreamsPreferredUsernameProperty()
	preferredUsernameProp.SetXMLSchemaString(username)
	group.SetActivityStreamsPreferredUsername(preferredUsernameProp)

	// name
	// Used as profile display name.
	nameProp := streams.NewActivityStreamsNameProperty()
	if displayName != "" {
		nameProp.AppendXMLSchemaString(displayName)
	} else {
		nameProp.AppendXMLSchemaString(username)
	}
	group.SetActivityStreamsName(nameProp)

	// summary
	// Used as profile bio.
	if note != "" {
		summaryProp := streams.NewActivityStreamsSummaryProperty()
		summaryProp.AppendXMLSchemaString(note)
		group.SetActivityStreamsSummary(summaryProp)
	}

	// url
	// Used as profile link.
	urlProp := streams.NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(profileURL)
	group.SetActivityStreamsUrl(urlProp)

	// manuallyApprovesFollowers
	manuallyApprovesFollowersProp := streams.NewActivityStreamsManuallyApprovesFollowersProperty()
	manuallyApprovesFollowersProp.Set(manuallyApprovesFollowers)
	group.SetActivityStreamsManuallyApprovesFollowers(manuallyApprovesFollowersProp)

	// discoverable
	// Will be shown in the profile directory.
	discoverableProp := streams.NewTootDiscoverableProperty()
	discoverableProp.Set(discoverable)
	group.SetTootDiscoverable(discoverableProp)

	// devices
	// NOT IMPLEMENTED, probably won't implement

	// alsoKnownAs
	// Required for Move activity.
	// TODO: NOT IMPLEMENTED **YET** -- this needs to be added as an activitypub extension to https://github.com/go-fed/activity, see https://github.com/go-fed/activity/tree/master/astool

	// publicKey
	// Required for signatures.
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
	encodedPublicKey, err := x509.MarshalPKIXPublicKey(pkey)
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
	group.SetW3IDSecurityV1PublicKey(publicKeyProp)

	// tag
	// TODO: Any tags used in the summary of this profile

	// attachment
	// Used for profile fields.
	// TODO: The PropertyValue type has to be added: https://schema.org/PropertyValue

	// endpoints
	// NOT IMPLEMENTED -- this is for shared inbox which we don't use

	// icon
	// Used as profile avatar.
	iconProperty := streams.NewActivityStreamsIconProperty()
	iconImage := streams.NewActivityStreamsImage()
	mediaType := streams.NewActivityStreamsMediaTypeProperty()
	mediaType.Set(avatarContentType)
	iconImage.SetActivityStreamsMediaType(mediaType)
	avatarURLProperty := streams.NewActivityStreamsUrlProperty()
	avatarURLProperty.AppendIRI(avatarURL)
	iconImage.SetActivityStreamsUrl(avatarURLProperty)
	iconProperty.AppendActivityStreamsImage(iconImage)
	group.SetActivityStreamsIcon(iconProperty)

	// image
	// Used as profile header.
	headerProperty := streams.NewActivityStreamsImageProperty()
	headerImage := streams.NewActivityStreamsImage()
	headerMediaType := streams.NewActivityStreamsMediaTypeProperty()
	mediaType.Set(headerContentType)
	headerImage.SetActivityStreamsMediaType(headerMediaType)
	headerURLProperty := streams.NewActivityStreamsUrlProperty()
	headerURLProperty.AppendIRI(headerURL)
	headerImage.SetActivityStreamsUrl(headerURLProperty)
	headerProperty.AppendActivityStreamsImage(headerImage)
	group.SetActivityStreamsImage(headerProperty)

	return group
}

func newMention(uri *url.URL, namestring string) vocab.ActivityStreamsMention {
	mention := streams.NewActivityStreamsMention()

	hrefProp := streams.NewActivityStreamsHrefProperty()
	hrefProp.SetIRI(uri)
	mention.SetActivityStreamsHref(hrefProp)

	nameProp := streams.NewActivityStreamsNameProperty()
	nameProp.AppendXMLSchemaString(namestring)
	mention.SetActivityStreamsName(nameProp)

	return mention
}

// newNote returns a new activity streams note for the given parameters
func newNote(
	noteID *url.URL,
	noteURL *url.URL,
	noteCreatedAt time.Time,
	noteContent string,
	noteSummary string,
	noteAttributedTo *url.URL,
	noteTo []*url.URL,
	noteCC []*url.URL,
	noteSensitive bool,
	noteMentions []vocab.ActivityStreamsMention) vocab.ActivityStreamsNote {

	// create the note itself
	note := streams.NewActivityStreamsNote()

	// set id
	if noteID != nil {
		id := streams.NewJSONLDIdProperty()
		id.Set(noteID)
		note.SetJSONLDId(id)
	}

	// set noteURL
	if noteURL != nil {
		url := streams.NewActivityStreamsUrlProperty()
		url.AppendIRI(noteURL)
		note.SetActivityStreamsUrl(url)
	}

	if noteCreatedAt.IsZero() {
		noteCreatedAt = time.Now()
	}
	published := streams.NewActivityStreamsPublishedProperty()
	published.Set(noteCreatedAt)
	note.SetActivityStreamsPublished(published)

	// set noteContent
	if noteContent != "" {
		content := streams.NewActivityStreamsContentProperty()
		content.AppendXMLSchemaString(noteContent)
		note.SetActivityStreamsContent(content)
	}

	// set noteSummary (aka content warning)
	if noteSummary != "" {
		summary := streams.NewActivityStreamsSummaryProperty()
		summary.AppendXMLSchemaString(noteSummary)
		note.SetActivityStreamsSummary(summary)
	}

	// set noteAttributedTo (the url of the author of the note)
	if noteAttributedTo != nil {
		attributedTo := streams.NewActivityStreamsAttributedToProperty()
		attributedTo.AppendIRI(noteAttributedTo)
		note.SetActivityStreamsAttributedTo(attributedTo)
	}

	// set noteTO
	if noteTo != nil {
		to := streams.NewActivityStreamsToProperty()
		for _, r := range noteTo {
			to.AppendIRI(r)
		}
		note.SetActivityStreamsTo(to)
	}

	// set noteCC
	if noteCC != nil {
		cc := streams.NewActivityStreamsCcProperty()
		for _, r := range noteCC {
			cc.AppendIRI(r)
		}
		note.SetActivityStreamsCc(cc)
	}

	// set note tags
	tag := streams.NewActivityStreamsTagProperty()

	// mentions
	for _, m := range noteMentions {
		tag.AppendActivityStreamsMention(m)
	}

	note.SetActivityStreamsTag(tag)

	return note
}

// wrapNoteInCreate wraps the given activity streams note in a Create activity streams action
func wrapNoteInCreate(createID *url.URL, createActor *url.URL, createPublished time.Time, createNote vocab.ActivityStreamsNote) vocab.ActivityStreamsCreate {
	// create the.... create
	create := streams.NewActivityStreamsCreate()

	// set createID
	if createID != nil {
		id := streams.NewJSONLDIdProperty()
		id.Set(createID)
		create.SetJSONLDId(id)
	}

	// set createActor
	if createActor != nil {
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(createActor)
		create.SetActivityStreamsActor(actor)
	}

	// set createPublished (time)
	if !createPublished.IsZero() {
		published := streams.NewActivityStreamsPublishedProperty()
		published.Set(createPublished)
		create.SetActivityStreamsPublished(published)
	}

	// setCreateTo
	if createNote.GetActivityStreamsTo() != nil {
		create.SetActivityStreamsTo(createNote.GetActivityStreamsTo())
	}

	// setCreateCC
	if createNote.GetActivityStreamsCc() != nil {
		create.SetActivityStreamsCc(createNote.GetActivityStreamsCc())
	}

	// set createNote
	if createNote != nil {
		note := streams.NewActivityStreamsObjectProperty()
		note.AppendActivityStreamsNote(createNote)
		create.SetActivityStreamsObject(note)
	}

	return create
}
