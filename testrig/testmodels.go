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
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// NewTestTokens returns a map of tokens keyed according to which account the token belongs to.
func NewTestTokens() map[string]*oauth.Token {
	tokens := map[string]*oauth.Token{
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
			UserID:          "01F8MGWAPB4GJ42M4N0TCZSQ7K",
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
func NewTestClients() map[string]*oauth.Client {
	clients := map[string]*oauth.Client{
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
			UserID: "01F8MGWAPB4GJ42M4N0TCZSQ7K", // local_account_2
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
			VapidKey:     "76ae0095-8a10-438f-9f49-522d1985b190",
		},
		"application_1": {
			ID:           "01F8MGY43H3N2C8EWPR2FPYEXG",
			Name:         "really cool gts application",
			Website:      "https://reallycool.app",
			RedirectURI:  "http://localhost:8080",
			ClientID:     "01F8MGV8AC3NGSJW0FE8W1BV70",           // client_1
			ClientSecret: "c3724c74-dc3b-41b2-a108-0ea3d8399830", // client_1
			Scopes:       "read write follow push",
			VapidKey:     "4738dfd7-ca73-4aa6-9aa9-80e946b7db36",
		},
		"application_2": {
			ID:           "01F8MGYG9E893WRHW0TAEXR8GJ",
			Name:         "kindaweird",
			Website:      "https://kindaweird.app",
			RedirectURI:  "http://localhost:8080",
			ClientID:     "01F8MGW47HN8ZXNHNZ7E47CDMQ",           // client_2
			ClientSecret: "8f5603a5-c721-46cd-8f1b-2e368f51379f", // client_2
			Scopes:       "read write follow push",
			VapidKey:     "c040a5fc-e1e2-4859-bbea-0a3efbca1c4b",
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
			CreatedByApplicationID: "",
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
			CreatedByApplicationID: "",
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
			CreatedByApplicationID: "",
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
		"instance_account": {
			ID:       "01F8MH261H1KSV3GW3016GZRY3",
			Username: "localhost:8080",
		},
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
			ActorType:               gtsmodel.ActivityStreamsPerson,
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
			ActorType:               gtsmodel.ActivityStreamsPerson,
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
			ActorType:               gtsmodel.ActivityStreamsPerson,
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
			ActorType:               gtsmodel.ActivityStreamsPerson,
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
			CreatedAt:             time.Now().Add(-190 * time.Hour),
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
			ActorType:             gtsmodel.ActivityStreamsPerson,
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
	}

	// generate keys for each account
	for _, v := range accounts {
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}
		pub := &priv.PublicKey

		// normally only local accounts get a private key (obviously)
		// but for testing purposes and signing requests, we'll give
		// remote accounts a private key as well
		v.PrivateKey = priv
		v.PublicKey = pub
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
			CreatedAt:                time.Now().Add(-71 * time.Hour),
			UpdatedAt:                time.Now().Add(-71 * time.Hour),
			Local:                    true,
			AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                false,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
			VisibilityAdvanced: &gtsmodel.VisibilityAdvanced{
				Federated: true,
				Boostable: true,
				Replyable: true,
				Likeable:  true,
			},
			ActivityStreamsType: gtsmodel.ActivityStreamsNote,
		},
		"admin_account_status_2": {
			ID:                       "01F8MHAAY43M6RJ473VQFCVH37",
			URI:                      "http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
			URL:                      "http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
			Content:                  "🐕🐕🐕🐕🐕",
			CreatedAt:                time.Now().Add(-70 * time.Hour),
			UpdatedAt:                time.Now().Add(-70 * time.Hour),
			Local:                    true,
			AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "open to see some puppies",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                true,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
			VisibilityAdvanced: &gtsmodel.VisibilityAdvanced{
				Federated: true,
				Boostable: true,
				Replyable: true,
				Likeable:  true,
			},
			ActivityStreamsType: gtsmodel.ActivityStreamsNote,
		},
		"local_account_1_status_1": {
			ID:                       "01F8MHAMCHF6Y650WCRSCP4WMY",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
			Content:                  "hello everyone!",
			CreatedAt:                time.Now().Add(-47 * time.Hour),
			UpdatedAt:                time.Now().Add(-47 * time.Hour),
			Local:                    true,
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "introduction post",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                true,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			VisibilityAdvanced: &gtsmodel.VisibilityAdvanced{
				Federated: true,
				Boostable: true,
				Replyable: true,
				Likeable:  true,
			},
			ActivityStreamsType: gtsmodel.ActivityStreamsNote,
		},
		"local_account_1_status_2": {
			ID:                       "01F8MHAYFKS4KMXF8K5Y1C0KRN",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAYFKS4KMXF8K5Y1C0KRN",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAYFKS4KMXF8K5Y1C0KRN",
			Content:                  "this is an unlocked local-only post that shouldn't federate, but it's still boostable, replyable, and likeable",
			CreatedAt:                time.Now().Add(-46 * time.Hour),
			UpdatedAt:                time.Now().Add(-46 * time.Hour),
			Local:                    true,
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "",
			Visibility:               gtsmodel.VisibilityUnlocked,
			Sensitive:                false,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			VisibilityAdvanced: &gtsmodel.VisibilityAdvanced{
				Federated: false,
				Boostable: true,
				Replyable: true,
				Likeable:  true,
			},
			ActivityStreamsType: gtsmodel.ActivityStreamsNote,
		},
		"local_account_1_status_3": {
			ID:                       "01F8MHBBN8120SYH7D5S050MGK",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHBBN8120SYH7D5S050MGK",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01F8MHBBN8120SYH7D5S050MGK",
			Content:                  "this is a very personal post that I don't want anyone to interact with at all, and i only want mutuals to see it",
			CreatedAt:                time.Now().Add(-45 * time.Hour),
			UpdatedAt:                time.Now().Add(-45 * time.Hour),
			Local:                    true,
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "test: you shouldn't be able to interact with this post in any way",
			Visibility:               gtsmodel.VisibilityMutualsOnly,
			Sensitive:                false,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			VisibilityAdvanced: &gtsmodel.VisibilityAdvanced{
				Federated: true,
				Boostable: false,
				Replyable: false,
				Likeable:  false,
			},
			ActivityStreamsType: gtsmodel.ActivityStreamsNote,
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
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "eye contact, trent reznor gif",
			Visibility:               gtsmodel.VisibilityMutualsOnly,
			Sensitive:                false,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			VisibilityAdvanced: &gtsmodel.VisibilityAdvanced{
				Federated: true,
				Boostable: true,
				Replyable: true,
				Likeable:  true,
			},
			ActivityStreamsType: gtsmodel.ActivityStreamsNote,
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
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "",
			Visibility:               gtsmodel.VisibilityMutualsOnly,
			Sensitive:                false,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			VisibilityAdvanced: &gtsmodel.VisibilityAdvanced{
				Federated: true,
				Boostable: true,
				Replyable: true,
				Likeable:  true,
			},
			ActivityStreamsType: gtsmodel.ActivityStreamsNote,
		},
		"local_account_2_status_1": {
			ID:                       "01F8MHBQCBTDKN6X5VHGMMN4MA",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01F8MHBQCBTDKN6X5VHGMMN4MA",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01F8MHBQCBTDKN6X5VHGMMN4MA",
			Content:                  "🐢 hi everyone i post about turtles 🐢",
			CreatedAt:                time.Now().Add(-189 * time.Hour),
			UpdatedAt:                time.Now().Add(-189 * time.Hour),
			Local:                    true,
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "introduction post",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                true,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			VisibilityAdvanced: &gtsmodel.VisibilityAdvanced{
				Federated: true,
				Boostable: true,
				Replyable: true,
				Likeable:  true,
			},
			ActivityStreamsType: gtsmodel.ActivityStreamsNote,
		},
		"local_account_2_status_2": {
			ID:                       "01F8MHC0H0A7XHTVH5F596ZKBM",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01F8MHC0H0A7XHTVH5F596ZKBM",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01F8MHC0H0A7XHTVH5F596ZKBM",
			Content:                  "🐢 this one is federated, likeable, and boostable but not replyable 🐢",
			CreatedAt:                time.Now().Add(-1 * time.Minute),
			UpdatedAt:                time.Now().Add(-1 * time.Minute),
			Local:                    true,
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                true,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			VisibilityAdvanced: &gtsmodel.VisibilityAdvanced{
				Federated: true,
				Boostable: true,
				Replyable: false,
				Likeable:  true,
			},
			ActivityStreamsType: gtsmodel.ActivityStreamsNote,
		},
		"local_account_2_status_3": {
			ID:                       "01F8MHC8VWDRBQR0N1BATDDEM5",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01F8MHC8VWDRBQR0N1BATDDEM5",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01F8MHC8VWDRBQR0N1BATDDEM5",
			Content:                  "🐢 i don't mind people sharing this one but I don't want likes or replies to it because cba🐢",
			CreatedAt:                time.Now().Add(-2 * time.Minute),
			UpdatedAt:                time.Now().Add(-2 * time.Minute),
			Local:                    true,
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "you won't be able to like or reply to this",
			Visibility:               gtsmodel.VisibilityUnlocked,
			Sensitive:                true,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			VisibilityAdvanced: &gtsmodel.VisibilityAdvanced{
				Federated: true,
				Boostable: true,
				Replyable: false,
				Likeable:  false,
			},
			ActivityStreamsType: gtsmodel.ActivityStreamsNote,
		},
		"local_account_2_status_4": {
			ID:                       "01F8MHCP5P2NWYQ416SBA0XSEV",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01F8MHCP5P2NWYQ416SBA0XSEV",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01F8MHCP5P2NWYQ416SBA0XSEV",
			Content:                  "🐢 this is a public status but I want it local only and not boostable 🐢",
			CreatedAt:                time.Now().Add(-1 * time.Minute),
			UpdatedAt:                time.Now().Add(-1 * time.Minute),
			Local:                    true,
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			InReplyToID:              "",
			BoostOfID:                "",
			ContentWarning:           "",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                true,
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			VisibilityAdvanced: &gtsmodel.VisibilityAdvanced{
				Federated: false,
				Boostable: false,
				Replyable: true,
				Likeable:  true,
			},
			ActivityStreamsType: gtsmodel.ActivityStreamsNote,
		},
		"local_account_2_status_5": {
			ID:                       "01FCQSQ667XHJ9AV9T27SJJSX5",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01FCQSQ667XHJ9AV9T27SJJSX5",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01FCQSQ667XHJ9AV9T27SJJSX5",
			Content:                  "🐢 @the_mighty_zork hi zork! 🐢",
			CreatedAt:                time.Now().Add(-1 * time.Minute),
			UpdatedAt:                time.Now().Add(-1 * time.Minute),
			Local:                    true,
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
			VisibilityAdvanced: &gtsmodel.VisibilityAdvanced{
				Federated: true,
				Boostable: true,
				Replyable: true,
				Likeable:  true,
			},
			ActivityStreamsType: gtsmodel.ActivityStreamsNote,
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
		URLMustParse("https://fossbros-anonymous.io/users/foss_satan/statuses/5424b153-4553-4f30-9358-7b92f7cd42f6"),
		URLMustParse("https://fossbros-anonymous.io/@foss_satan/5424b153-4553-4f30-9358-7b92f7cd42f6"),
		time.Now(),
		"hey zork here's a new private note for you",
		"new note for zork",
		URLMustParse("https://fossbros-anonymous.io/users/foss_satan"),
		[]*url.URL{URLMustParse("http://localhost:8080/users/the_mighty_zork")},
		nil,
		true,
		[]vocab.ActivityStreamsMention{})
	createDmForZork := wrapNoteInCreate(
		URLMustParse("https://fossbros-anonymous.io/users/foss_satan/statuses/5424b153-4553-4f30-9358-7b92f7cd42f6/activity"),
		URLMustParse("https://fossbros-anonymous.io/users/foss_satan"),
		time.Now(),
		dmForZork)
	sig, digest, date := getSignatureForActivity(createDmForZork, accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, URLMustParse(accounts["local_account_1"].InboxURI))

	return map[string]ActivityWithSignature{
		"dm_for_zork": {
			Activity:        createDmForZork,
			SignatureHeader: sig,
			DigestHeader:    digest,
			DateHeader:      date,
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
				URLMustParse("https://www.w3.org/ns/activitystreams#Public"),
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
				URLMustParse("https://www.w3.org/ns/activitystreams#Public"),
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
	sig, digest, date = getSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceZork := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(statuses["local_account_1_status_1"].URI + "/replies")
	sig, digest, date = getSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceLocalAccount1Status1Replies := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(statuses["local_account_1_status_1"].URI + "/replies?only_other_accounts=false&page=true")
	sig, digest, date = getSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceLocalAccount1Status1RepliesNext := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(statuses["local_account_1_status_1"].URI + "/replies?only_other_accounts=false&page=true&min_id=01FCQSQ667XHJ9AV9T27SJJSX5")
	sig, digest, date = getSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceLocalAccount1Status1RepliesLast := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	return map[string]ActivityWithSignature{
		"foss_satan_dereference_zork":                                  fossSatanDereferenceZork,
		"foss_satan_dereference_local_account_1_status_1_replies":      fossSatanDereferenceLocalAccount1Status1Replies,
		"foss_satan_dereference_local_account_1_status_1_replies_next": fossSatanDereferenceLocalAccount1Status1RepliesNext,
		"foss_satan_dereference_local_account_1_status_1_replies_last": fossSatanDereferenceLocalAccount1Status1RepliesLast,
	}
}

// getSignatureForActivity does some sneaky sneaky work with a mock http client and a test transport controller, in order to derive
// the HTTP Signature for the given activity, public key ID, private key, and destination.
func getSignatureForActivity(activity pub.Activity, pubKeyID string, privkey crypto.PrivateKey, destination *url.URL) (signatureHeader string, digestHeader string, dateHeader string) {
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

// getSignatureForDereference does some sneaky sneaky work with a mock http client and a test transport controller, in order to derive
// the HTTP Signature for the given derefence GET request using public key ID, private key, and destination.
func getSignatureForDereference(pubKeyID string, privkey crypto.PrivateKey, destination *url.URL) (signatureHeader string, digestHeader string, dateHeader string) {
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
