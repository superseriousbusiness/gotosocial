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

package testrig

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"codeberg.org/superseriousbusiness/activity/pub"
	"codeberg.org/superseriousbusiness/activity/streams"
	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// NewTestTokens returns a map of tokens keyed according to which account the token belongs to.
func NewTestTokens() map[string]*gtsmodel.Token {
	tokens := map[string]*gtsmodel.Token{
		"local_account_1": {
			ID:              "01F8MGTQW4DKTDF8SW5CT9HYGA",
			ClientID:        "01F8MGV8AC3NGSJW0FE8W1BV70",
			UserID:          "01F8MGVGPHQ2D3P3X0454H54Z5",
			RedirectURI:     "http://localhost:8080",
			Scope:           "read write push",
			Access:          "NZAZOTC0OWITMDU0NC0ZODG4LWE4NJITMWUXM2M4MTRHZDEX",
			AccessCreateAt:  TimeMustParse("2022-06-10T15:22:08Z"),
			AccessExpiresAt: TimeMustParse("2050-01-01T15:22:08Z"),
		},
		"local_account_1_push_only": {
			ID:              "01JN0X2D9GJTZQ5KYPYFWN16QW",
			ClientID:        "01F8MGV8AC3NGSJW0FE8W1BV70",
			UserID:          "01F8MGVGPHQ2D3P3X0454H54Z5",
			RedirectURI:     "http://localhost:8080",
			Scope:           "push",
			Access:          "01JN0X49RYKMP6G9X0HJAP317101JN0X49RYKMP6G9X0HJAP",
			AccessCreateAt:  TimeMustParse("2022-06-10T15:22:08Z"),
			AccessExpiresAt: TimeMustParse("2050-01-01T15:22:08Z"),
		},
		"local_account_1_client_application_token": {
			ID:              "01P9SVWS9J3SPHZQ3KCMBEN70N",
			ClientID:        "01F8MGV8AC3NGSJW0FE8W1BV70",
			RedirectURI:     "http://localhost:8080",
			Scope:           "read write push",
			Access:          "ZTK1MWMWZDGTMGMXOS0ZY2UXLWI5ZWETMWEZYZZIYTLHMZI4",
			AccessCreateAt:  TimeMustParse("2022-06-10T15:22:08Z"),
			AccessExpiresAt: TimeMustParse("2050-01-01T15:22:08Z"),
		},
		"local_account_1_user_authorization_token": {
			ID:            "01G574M2VTV66YZBC9AZ7HY3FV",
			ClientID:      "01F8MGV8AC3NGSJW0FE8W1BV70",
			UserID:        "01F8MGVGPHQ2D3P3X0454H54Z5",
			RedirectURI:   "http://localhost:8080",
			Scope:         "read write push",
			Code:          "ZJYYMZQ0MTQTZTU1NC0ZNJK4LWE2ZWITYTM1MDHHOTAXNJHL",
			CodeCreateAt:  TimeMustParse("2022-06-10T15:22:08Z"),
			CodeExpiresAt: TimeMustParse("2050-01-01T15:22:08Z"),
		},
		"local_account_2": {
			ID:              "01F8MGVVM1EDVYET710J27XY5R",
			ClientID:        "01F8MGW47HN8ZXNHNZ7E47CDMQ",
			UserID:          "01F8MH1VYJAE00TVVGMM5JNJ8X",
			RedirectURI:     "http://localhost:8080",
			Scope:           "read write push",
			Access:          "PIPINALKNNNFNF98717NAMNAMNFKIJKJ881818KJKJAKJJJA",
			AccessCreateAt:  TimeMustParse("2022-06-10T15:22:08Z"),
			AccessExpiresAt: TimeMustParse("2050-01-01T15:22:08Z"),
		},
		"local_account_3": {
			ID:             "01JPCMGR09M8VGARPSBABXNZFQ",
			ClientID:       "01F8MGV8AC3NGSJW0FE8W1BV70",
			UserID:         "01JPCMFRTQ0B6R8SXPM7RS80Q4",
			RedirectURI:    "http://localhost:8080",
			Scope:          "read write push",
			Access:         "01JPCMK0YQ24FFVZ98PYZGJCC901JPCMK32ZKZMM737HGSWMW",
			AccessCreateAt: TimeMustParse("2025-03-15T11:08:00Z"),
		},
		"admin_account": {
			ID:              "01FS4TP8ANA5VE92EAPA9E0M7Q",
			ClientID:        "01F8MGWSJCND9BWBD4WGJXBM93",
			UserID:          "01F8MGWYWKVKS3VS8DV1AMYPGE",
			RedirectURI:     "http://localhost:8080",
			Scope:           "read write push admin",
			Access:          "AININALKNENFNF98717NAMG4LWE4NJITMWUXM2M4MTRHZDEX",
			AccessCreateAt:  TimeMustParse("2022-06-10T15:22:08Z"),
			AccessExpiresAt: TimeMustParse("2050-01-01T15:22:08Z"),
		},
	}
	return tokens
}

// NewTestApplications returns a map of applications keyed to which number application they are.
func NewTestApplications() map[string]*gtsmodel.Application {
	apps := map[string]*gtsmodel.Application{
		"instance_application": {
			ID:           "01HT5P2YHDMPAAD500NDAY8JW1",
			Name:         "localhost:8080 instance application",
			Website:      "http://localhost:8080",
			RedirectURIs: []string{"http://localhost:8080"},
			ClientID:     "01AY6P665V14JJR0AFVRT7311Y", // instance account ID
			ClientSecret: "baedee87-6d00-4cf5-87b9-4d78ee58ef01",
			Scopes:       "write:accounts",
		},
		"admin_account": {
			ID:           "01F8MGXQRHYF5QPMTMXP78QC2F",
			Name:         "superseriousbusiness",
			Website:      "https://superserious.business",
			RedirectURIs: []string{"http://localhost:8080"},
			ClientID:     "01F8MGWSJCND9BWBD4WGJXBM93",           // admin client
			ClientSecret: "dda8e835-2c9c-4bd2-9b8b-77c2e26d7a7a", // admin client
			Scopes:       "read write push",
		},
		"application_1": {
			ID:           "01F8MGY43H3N2C8EWPR2FPYEXG",
			Name:         "really cool gts application",
			Website:      "https://reallycool.app",
			RedirectURIs: []string{"http://localhost:8080"},
			ClientID:     "01F8MGV8AC3NGSJW0FE8W1BV70",           // client_1
			ClientSecret: "c3724c74-dc3b-41b2-a108-0ea3d8399830", // client_1
			Scopes:       "read write push",
		},
		"application_2": {
			ID:           "01F8MGYG9E893WRHW0TAEXR8GJ",
			Name:         "kindaweird",
			Website:      "https://kindaweird.app",
			RedirectURIs: []string{"http://localhost:8080"},
			ClientID:     "01F8MGW47HN8ZXNHNZ7E47CDMQ",           // client_2
			ClientSecret: "8f5603a5-c721-46cd-8f1b-2e368f51379f", // client_2
			Scopes:       "read write push",
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
			CreatedAt:              TimeMustParse("2022-06-04T13:12:00Z"),
			SignUpIP:               net.ParseIP("199.222.111.89"),
			UpdatedAt:              time.Time{},
			InviteID:               "",
			Reason:                 "hi, please let me in! I'm looking for somewhere neato bombeato to hang out.",
			Locale:                 "en",
			CreatedByApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			LastEmailedAt:          time.Time{},
			ConfirmationToken:      "a5a280bd-34be-44a3-8330-a57eaf61b8dd",
			ConfirmedAt:            time.Time{},
			ConfirmationSentAt:     TimeMustParse("2022-06-04T13:12:00Z"),
			UnconfirmedEmail:       "weed_lord420@example.org",
			Moderator:              util.Ptr(false),
			Admin:                  util.Ptr(false),
			Disabled:               util.Ptr(false),
			Approved:               util.Ptr(false),
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
		"admin_account": {
			ID:                     "01F8MGWYWKVKS3VS8DV1AMYPGE",
			Email:                  "admin@example.org",
			AccountID:              "01F8MH17FWEB39HZJ76B6VXSKF",
			EncryptedPassword:      "$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS", // 'password'
			CreatedAt:              TimeMustParse("2022-06-01T13:12:00Z"),
			SignUpIP:               nil,
			UpdatedAt:              TimeMustParse("2022-06-01T13:12:00Z"),
			InviteID:               "",
			Locale:                 "en",
			CreatedByApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
			LastEmailedAt:          TimeMustParse("2022-06-03T13:12:00Z"),
			ConfirmationToken:      "",
			ConfirmedAt:            TimeMustParse("2022-06-02T13:12:00Z"),
			ConfirmationSentAt:     time.Time{},
			UnconfirmedEmail:       "",
			Moderator:              util.Ptr(true),
			Admin:                  util.Ptr(true),
			Disabled:               util.Ptr(false),
			Approved:               util.Ptr(true),
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
		"local_account_1": {
			ID:                     "01F8MGVGPHQ2D3P3X0454H54Z5",
			Email:                  "zork@example.org",
			AccountID:              "01F8MH1H7YV1Z7D2C8K2730QBF",
			EncryptedPassword:      "$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS", // 'password'
			CreatedAt:              TimeMustParse("2022-06-01T13:12:00Z"),
			SignUpIP:               nil,
			UpdatedAt:              TimeMustParse("2022-06-01T13:12:00Z"),
			InviteID:               "",
			Reason:                 "I wanna be on this damned webbed site so bad! Please! Wow",
			Locale:                 "en",
			CreatedByApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			LastEmailedAt:          TimeMustParse("2022-06-02T13:12:00Z"),
			ConfirmationToken:      "",
			ConfirmedAt:            TimeMustParse("2022-06-02T13:12:00Z"),
			ConfirmationSentAt:     TimeMustParse("2022-06-02T13:12:00Z"),
			UnconfirmedEmail:       "",
			Moderator:              util.Ptr(false),
			Admin:                  util.Ptr(false),
			Disabled:               util.Ptr(false),
			Approved:               util.Ptr(true),
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
		"local_account_2": {
			ID:                     "01F8MH1VYJAE00TVVGMM5JNJ8X",
			Email:                  "tortle.dude@example.org",
			AccountID:              "01F8MH5NBDF2MV7CTC4Q5128HF",
			EncryptedPassword:      "$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS", // 'password'
			CreatedAt:              TimeMustParse("2022-05-23T13:12:00Z"),
			SignUpIP:               nil,
			UpdatedAt:              TimeMustParse("2022-05-23T13:12:00Z"),
			InviteID:               "",
			Locale:                 "en",
			CreatedByApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			LastEmailedAt:          TimeMustParse("2022-06-06T13:12:00Z"),
			ConfirmationToken:      "",
			ConfirmedAt:            TimeMustParse("2022-05-24T13:12:00Z"),
			ConfirmationSentAt:     TimeMustParse("2022-05-23T13:12:00Z"),
			UnconfirmedEmail:       "",
			Moderator:              util.Ptr(false),
			Admin:                  util.Ptr(false),
			Disabled:               util.Ptr(false),
			Approved:               util.Ptr(true),
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
		"local_account_3": {
			ID:                     "01JPCMFRTQ0B6R8SXPM7RS80Q4",
			Email:                  "media.mogul@example.org",
			AccountID:              "01JPCMD83Y4WR901094YES3QC5",
			EncryptedPassword:      "$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS", // 'password'
			CreatedAt:              TimeMustParse("2025-03-15T11:08:00Z"),
			SignUpIP:               nil,
			UpdatedAt:              TimeMustParse("2025-03-15T11:08:00Z"),
			InviteID:               "",
			Locale:                 "en",
			CreatedByApplicationID: "01HT5P2YHDMPAAD500NDAY8JW1",
			LastEmailedAt:          TimeMustParse("2025-03-15T11:08:00Z"),
			ConfirmationToken:      "",
			ConfirmedAt:            TimeMustParse("2025-03-15T11:08:00Z"),
			ConfirmationSentAt:     TimeMustParse("2025-03-15T11:08:00Z"),
			UnconfirmedEmail:       "",
			Moderator:              util.Ptr(false),
			Admin:                  util.Ptr(false),
			Disabled:               util.Ptr(false),
			Approved:               util.Ptr(true),
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
	}

	return users
}

// NewTestAccounts returns a map of accounts keyed by what type of account they are.
func NewTestAccounts() map[string]*gtsmodel.Account {
	settings := NewTestAccountSettings()

	accounts := map[string]*gtsmodel.Account{
		"instance_account": {
			ID:                    "01AY6P665V14JJR0AFVRT7311Y",
			Username:              "localhost:8080",
			CreatedAt:             TimeMustParse("2020-05-17T13:10:59Z"),
			UpdatedAt:             TimeMustParse("2020-05-17T13:10:59Z"),
			Locked:                util.Ptr(false),
			Discoverable:          util.Ptr(true),
			URI:                   "http://localhost:8080/users/localhost:8080",
			URL:                   "http://localhost:8080/@localhost:8080",
			PublicKeyURI:          "http://localhost:8080/users/localhost:8080#main-key",
			InboxURI:              "http://localhost:8080/users/localhost:8080/inbox",
			OutboxURI:             "http://localhost:8080/users/localhost:8080/outbox",
			FollowersURI:          "http://localhost:8080/users/localhost:8080/followers",
			FollowingURI:          "http://localhost:8080/users/localhost:8080/following",
			FeaturedCollectionURI: "http://localhost:8080/users/localhost:8080/collections/featured",
			ActorType:             gtsmodel.AccountActorTypeService,
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
		},
		"unconfirmed_account": {
			ID:                    "01F8MH0BBE4FHXPH513MBVFHB0",
			Username:              "weed_lord420",
			CreatedAt:             TimeMustParse("2022-06-04T13:12:00Z"),
			UpdatedAt:             TimeMustParse("2022-06-04T13:12:00Z"),
			Locked:                util.Ptr(false),
			Discoverable:          util.Ptr(false),
			URI:                   "http://localhost:8080/users/weed_lord420",
			URL:                   "http://localhost:8080/@weed_lord420",
			InboxURI:              "http://localhost:8080/users/weed_lord420/inbox",
			OutboxURI:             "http://localhost:8080/users/weed_lord420/outbox",
			FollowersURI:          "http://localhost:8080/users/weed_lord420/followers",
			FollowingURI:          "http://localhost:8080/users/weed_lord420/following",
			FeaturedCollectionURI: "http://localhost:8080/users/weed_lord420/collections/featured",
			ActorType:             gtsmodel.AccountActorTypePerson,
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			PublicKeyURI:          "http://localhost:8080/users/weed_lord420#main-key",
			Settings:              settings["unconfirmed_account"],
		},
		"admin_account": {
			ID:                    "01F8MH17FWEB39HZJ76B6VXSKF",
			Username:              "admin",
			CreatedAt:             TimeMustParse("2022-05-17T13:10:59Z"),
			UpdatedAt:             TimeMustParse("2022-05-17T13:10:59Z"),
			Locked:                util.Ptr(false),
			Discoverable:          util.Ptr(true),
			URI:                   "http://localhost:8080/users/admin",
			URL:                   "http://localhost:8080/@admin",
			PublicKeyURI:          "http://localhost:8080/users/admin#main-key",
			InboxURI:              "http://localhost:8080/users/admin/inbox",
			OutboxURI:             "http://localhost:8080/users/admin/outbox",
			FollowersURI:          "http://localhost:8080/users/admin/followers",
			FollowingURI:          "http://localhost:8080/users/admin/following",
			FeaturedCollectionURI: "http://localhost:8080/users/admin/collections/featured",
			ActorType:             gtsmodel.AccountActorTypePerson,
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			Settings:              settings["admin_account"],
		},
		"local_account_1": {
			ID:                      "01F8MH1H7YV1Z7D2C8K2730QBF",
			Username:                "the_mighty_zork",
			AvatarMediaAttachmentID: "01F8MH58A357CV5K7R7TJMSH6S",
			HeaderMediaAttachmentID: "01PFPMWK2FF0D9WMHEJHR07C3Q",
			DisplayName:             "original zork (he/they)",
			Note:                    "<p>hey yo this is my profile!</p>",
			NoteRaw:                 "hey yo this is my profile!",
			CreatedAt:               TimeMustParse("2022-05-20T11:09:18Z"),
			UpdatedAt:               TimeMustParse("2022-05-20T11:09:18Z"),
			Locked:                  util.Ptr(false),
			Discoverable:            util.Ptr(true),
			URI:                     "http://localhost:8080/users/the_mighty_zork",
			URL:                     "http://localhost:8080/@the_mighty_zork",
			InboxURI:                "http://localhost:8080/users/the_mighty_zork/inbox",
			OutboxURI:               "http://localhost:8080/users/the_mighty_zork/outbox",
			FollowersURI:            "http://localhost:8080/users/the_mighty_zork/followers",
			FollowingURI:            "http://localhost:8080/users/the_mighty_zork/following",
			FeaturedCollectionURI:   "http://localhost:8080/users/the_mighty_zork/collections/featured",
			ActorType:               gtsmodel.AccountActorTypePerson,
			PrivateKey:              &rsa.PrivateKey{},
			PublicKey:               &rsa.PublicKey{},
			PublicKeyURI:            "http://localhost:8080/users/the_mighty_zork/main-key",
			Settings:                settings["local_account_1"],
		},
		"local_account_2": {
			ID:          "01F8MH5NBDF2MV7CTC4Q5128HF",
			Username:    "1happyturtle",
			DisplayName: "happy little turtle :3",
			Fields: []*gtsmodel.Field{
				{
					Name:  "should you follow me?",
					Value: "maybe!",
				},
				{
					Name:  "age",
					Value: "120",
				},
			},
			FieldsRaw: []*gtsmodel.Field{
				{
					Name:  "should you follow me?",
					Value: "maybe!",
				},
				{
					Name:  "age",
					Value: "120",
				},
			},
			Note:                  "<p>i post about things that concern me</p>",
			NoteRaw:               "i post about things that concern me",
			CreatedAt:             TimeMustParse("2022-06-04T13:12:00Z"),
			UpdatedAt:             TimeMustParse("2022-06-04T13:12:00Z"),
			Locked:                util.Ptr(true),
			Discoverable:          util.Ptr(false),
			URI:                   "http://localhost:8080/users/1happyturtle",
			URL:                   "http://localhost:8080/@1happyturtle",
			InboxURI:              "http://localhost:8080/users/1happyturtle/inbox",
			OutboxURI:             "http://localhost:8080/users/1happyturtle/outbox",
			FollowersURI:          "http://localhost:8080/users/1happyturtle/followers",
			FollowingURI:          "http://localhost:8080/users/1happyturtle/following",
			FeaturedCollectionURI: "http://localhost:8080/users/1happyturtle/collections/featured",
			ActorType:             gtsmodel.AccountActorTypePerson,
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			PublicKeyURI:          "http://localhost:8080/users/1happyturtle#main-key",
			Settings:              settings["local_account_2"],
		},
		"local_account_3": {
			ID:                      "01JPCMD83Y4WR901094YES3QC5",
			Username:                "media_mogul",
			AvatarMediaAttachmentID: "01JPHQZ0ZHC2AXJK1JQNXRXQZN",
			HeaderMediaAttachmentID: "01JPHRB7F2RXPTEQFRYC85EPD9",
			Fields: []*gtsmodel.Field{
				{
					Name:  "I'm going to post a lot of",
					Value: "media!",
				},
				{
					Name:  "and there's nothing",
					Value: "you can do about it",
				},
			},
			FieldsRaw: []*gtsmodel.Field{
				{
					Name:  "I'm going to post a lot of",
					Value: "media!",
				},
				{
					Name:  "and there's nothing",
					Value: "you can do about it",
				},
			},
			Note:                  "<p>I'm a test account that posts a shitload of media and I have my account rendered in \"gallery\" mode</p>",
			NoteRaw:               "I'm a test account that posts a shitload of media and I have my account rendered in \"gallery\" mode",
			CreatedAt:             TimeMustParse("2025-03-15T11:08:00Z"),
			UpdatedAt:             TimeMustParse("2025-03-15T11:08:00Z"),
			Locked:                util.Ptr(false),
			Discoverable:          util.Ptr(false),
			URI:                   "http://localhost:8080/users/media_mogul",
			URL:                   "http://localhost:8080/@media_mogul",
			InboxURI:              "http://localhost:8080/users/media_mogul/inbox",
			OutboxURI:             "http://localhost:8080/users/media_mogul/outbox",
			FollowersURI:          "http://localhost:8080/users/media_mogul/followers",
			FollowingURI:          "http://localhost:8080/users/media_mogul/following",
			FeaturedCollectionURI: "http://localhost:8080/users/media_mogul/collections/featured",
			ActorType:             gtsmodel.AccountActorTypePerson,
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			PublicKeyURI:          "http://localhost:8080/users/media_mogul#main-key",
			Settings:              settings["local_account_3"],
		},
		"remote_account_1": {
			ID:                    "01F8MH5ZK5VRH73AKHQM6Y9VNX",
			Username:              "foss_satan",
			Domain:                "fossbros-anonymous.io",
			DisplayName:           "big gerald",
			Note:                  "i post about like, i dunno, stuff, or whatever!!!!",
			CreatedAt:             TimeMustParse("2021-09-26T12:52:36+02:00"),
			UpdatedAt:             TimeMustParse("2022-06-04T13:12:00Z"),
			Locked:                util.Ptr(false),
			Discoverable:          util.Ptr(true),
			URI:                   "http://fossbros-anonymous.io/users/foss_satan",
			URL:                   "http://fossbros-anonymous.io/@foss_satan",
			InboxURI:              "http://fossbros-anonymous.io/users/foss_satan/inbox",
			SharedInboxURI:        util.Ptr("http://fossbros-anonymous.io/inbox"),
			OutboxURI:             "http://fossbros-anonymous.io/users/foss_satan/outbox",
			FollowersURI:          "http://fossbros-anonymous.io/users/foss_satan/followers",
			FollowingURI:          "http://fossbros-anonymous.io/users/foss_satan/following",
			FeaturedCollectionURI: "http://fossbros-anonymous.io/users/foss_satan/collections/featured",
			ActorType:             gtsmodel.AccountActorTypePerson,
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			PublicKeyURI:          "http://fossbros-anonymous.io/users/foss_satan/main-key",
		},
		"remote_account_2": {
			ID:                    "01FHMQX3GAABWSM0S2VZEC2SWC",
			Username:              "Some_User",
			Domain:                "example.org",
			DisplayName:           "some user",
			Note:                  "i'm a real son of a gun",
			CreatedAt:             TimeMustParse("2020-08-10T14:13:28+02:00"),
			UpdatedAt:             TimeMustParse("2022-06-04T13:12:00Z"),
			Locked:                util.Ptr(true),
			Discoverable:          util.Ptr(true),
			URI:                   "http://example.org/users/Some_User",
			URL:                   "http://example.org/@Some_User",
			InboxURI:              "http://example.org/users/Some_User/inbox",
			SharedInboxURI:        util.Ptr(""),
			OutboxURI:             "http://example.org/users/Some_User/outbox",
			FollowersURI:          "http://example.org/users/Some_User/followers",
			FollowingURI:          "http://example.org/users/Some_User/following",
			FeaturedCollectionURI: "http://example.org/users/Some_User/collections/featured",
			ActorType:             gtsmodel.AccountActorTypePerson,
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			PublicKeyURI:          "http://example.org/users/Some_User#main-key",
		},
		"remote_account_3": {
			ID:                      "062G5WYKY35KKD12EMSM3F8PJ8",
			Username:                "her_fuckin_maj",
			Domain:                  "thequeenisstillalive.technology",
			DisplayName:             "lizzzieeeeeeeeeeee",
			Note:                    "if i die blame charles don't let that fuck become king",
			CreatedAt:               TimeMustParse("2020-08-10T14:13:28+02:00"),
			UpdatedAt:               TimeMustParse("2022-06-04T13:12:00Z"),
			Locked:                  util.Ptr(true),
			Discoverable:            util.Ptr(true),
			URI:                     "http://thequeenisstillalive.technology/users/her_fuckin_maj",
			URL:                     "http://thequeenisstillalive.technology/@her_fuckin_maj",
			InboxURI:                "http://thequeenisstillalive.technology/users/her_fuckin_maj/inbox",
			SharedInboxURI:          util.Ptr(""),
			OutboxURI:               "http://thequeenisstillalive.technology/users/her_fuckin_maj/outbox",
			FollowersURI:            "http://thequeenisstillalive.technology/users/her_fuckin_maj/followers",
			FollowingURI:            "http://thequeenisstillalive.technology/users/her_fuckin_maj/following",
			FeaturedCollectionURI:   "http://thequeenisstillalive.technology/users/her_fuckin_maj/collections/featured",
			ActorType:               gtsmodel.AccountActorTypePerson,
			PrivateKey:              &rsa.PrivateKey{},
			PublicKey:               &rsa.PublicKey{},
			PublicKeyURI:            "http://thequeenisstillalive.technology/users/her_fuckin_maj#main-key",
			HeaderMediaAttachmentID: "01PFPMWK2FF0D9WMHEJHR07C3R",
		},
		"remote_account_4": {
			ID:                    "07GZRBAEMBNKGZ8Z9VSKSXKR98",
			Username:              "üser",
			Domain:                "xn--xample-ova.org",
			CreatedAt:             TimeMustParse("2020-08-10T14:13:28+02:00"),
			UpdatedAt:             TimeMustParse("2022-06-04T13:12:00Z"),
			Locked:                util.Ptr(false),
			Discoverable:          util.Ptr(false),
			URI:                   "https://xn--xample-ova.org/users/%C3%BCser",
			URL:                   "https://xn--xample-ova.org/users/@%C3%BCser",
			FetchedAt:             time.Time{},
			InboxURI:              "https://xn--xample-ova.org/users/%C3%BCser/inbox",
			SharedInboxURI:        util.Ptr(""),
			OutboxURI:             "https://xn--xample-ova.org/users/%C3%BCser/outbox",
			FollowersURI:          "https://xn--xample-ova.org/users/%C3%BCser/followers",
			FollowingURI:          "https://xn--xample-ova.org/users/%C3%BCser/following",
			FeaturedCollectionURI: "https://xn--xample-ova.org/users/%C3%BCser/collections/featured",
			ActorType:             gtsmodel.AccountActorTypePerson,
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			PublicKeyURI:          "https://xn--xample-ova.org/users/%C3%BCser#main-key",
		},
	}

	var accountsSorted []*gtsmodel.Account
	for _, a := range accounts {
		accountsSorted = append(accountsSorted, a)
	}

	sort.Slice(accountsSorted, func(i, j int) bool {
		return accountsSorted[i].ID > accountsSorted[j].ID
	})

	preserializedKeys := []string{
		"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",
	}

	if diff := len(accountsSorted) - len(preserializedKeys); diff > 0 {
		keyStrings := make([]string, 0, diff)
		for i := 0; i < diff; i++ {
			priv, _ := rsa.GenerateKey(rand.Reader, 2048)
			key, _ := x509.MarshalPKCS8PrivateKey(priv)
			keyStrings = append(keyStrings, base64.StdEncoding.EncodeToString(key))
		}
		panic(fmt.Sprintf("mismatch between number of hardcoded test RSA keys and accounts used for test data. Insert the following generated key[s]: \n%+v", keyStrings))
	}

	for i, v := range accountsSorted {
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
	}

	return accounts
}

func NewTestAccountSettings() map[string]*gtsmodel.AccountSettings {
	return map[string]*gtsmodel.AccountSettings{
		"unconfirmed_account": {
			AccountID:       "01F8MH0BBE4FHXPH513MBVFHB0",
			CreatedAt:       TimeMustParse("2022-06-04T13:12:00Z"),
			UpdatedAt:       TimeMustParse("2022-06-04T13:12:00Z"),
			Privacy:         gtsmodel.VisibilityPublic,
			Sensitive:       util.Ptr(false),
			Language:        "en",
			EnableRSS:       util.Ptr(false),
			HideCollections: util.Ptr(false),
			WebVisibility:   gtsmodel.VisibilityPublic,
			WebLayout:       gtsmodel.WebLayoutMicroblog,
		},
		"admin_account": {
			AccountID:       "01F8MH17FWEB39HZJ76B6VXSKF",
			CreatedAt:       TimeMustParse("2022-05-17T13:10:59Z"),
			UpdatedAt:       TimeMustParse("2022-05-17T13:10:59Z"),
			Privacy:         gtsmodel.VisibilityPublic,
			Sensitive:       util.Ptr(false),
			Language:        "en",
			EnableRSS:       util.Ptr(true),
			HideCollections: util.Ptr(false),
			WebVisibility:   gtsmodel.VisibilityPublic,
			WebLayout:       gtsmodel.WebLayoutMicroblog,
		},
		"local_account_1": {
			AccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF",
			CreatedAt:       TimeMustParse("2022-05-20T11:09:18Z"),
			UpdatedAt:       TimeMustParse("2022-05-20T11:09:18Z"),
			Privacy:         gtsmodel.VisibilityPublic,
			Sensitive:       util.Ptr(false),
			Language:        "en",
			EnableRSS:       util.Ptr(true),
			HideCollections: util.Ptr(false),
			WebVisibility:   gtsmodel.VisibilityUnlocked,
			WebLayout:       gtsmodel.WebLayoutMicroblog,
		},
		"local_account_2": {
			AccountID:       "01F8MH5NBDF2MV7CTC4Q5128HF",
			CreatedAt:       TimeMustParse("2022-06-04T13:12:00Z"),
			UpdatedAt:       TimeMustParse("2022-06-04T13:12:00Z"),
			Privacy:         gtsmodel.VisibilityFollowersOnly,
			Sensitive:       util.Ptr(true),
			Language:        "fr",
			EnableRSS:       util.Ptr(false),
			HideCollections: util.Ptr(true),
			WebVisibility:   gtsmodel.VisibilityPublic,
			WebLayout:       gtsmodel.WebLayoutMicroblog,
		},
		"local_account_3": {
			AccountID:       "01JPCMD83Y4WR901094YES3QC5",
			CreatedAt:       TimeMustParse("2025-03-15T11:08:00Z"),
			UpdatedAt:       TimeMustParse("2025-03-15T11:08:00Z"),
			Privacy:         gtsmodel.VisibilityPublic,
			Sensitive:       util.Ptr(true),
			Language:        "en",
			EnableRSS:       util.Ptr(true),
			HideCollections: util.Ptr(false),
			WebVisibility:   gtsmodel.VisibilityUnlocked,
			WebLayout:       gtsmodel.WebLayoutGallery,
		},
	}
}

func NewTestTombstones() map[string]*gtsmodel.Tombstone {
	return map[string]*gtsmodel.Tombstone{
		"https://somewhere.mysterious/users/rest_in_piss#main-key": {
			ID:        "01GHBTVE9HQPPBDH2W5VH2DGN4",
			CreatedAt: TimeMustParse("2021-11-09T19:33:45Z"),
			UpdatedAt: TimeMustParse("2021-11-09T19:33:45Z"),
			Domain:    "somewhere.mysterious",
			URI:       "https://somewhere.mysterious/users/rest_in_piss#main-key",
		},
	}
}

// NewTestAttachments returns a map of attachments keyed according to which account
// and status they belong to, and which attachment number of that status they are.
func NewTestAttachments() map[string]*gtsmodel.MediaAttachment {
	return map[string]*gtsmodel.MediaAttachment{
		"admin_account_status_1_attachment_1": {
			ID:        "01F8MH6NEM8D7527KZAECTCR76",
			StatusID:  "01F8MH75CBF9JFX4ZAD54N0W0R",
			URL:       "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2022-06-04T13:12:00Z"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  1200,
					Height: 630,
					Size:   756000,
					Aspect: 1.9047619,
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 268,
					Size:   137216,
					Aspect: 1.9104477,
				},
			},
			AccountID:         "01F8MH17FWEB39HZJ76B6VXSKF",
			Description:       "Black and white image of some 50's style text saying: Welcome On Board",
			ScheduledStatusID: "",
			Blurhash:          "LIIE|gRj00WB-;j[t7j[4nWBj[Rj",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpeg",
				ContentType: "image/jpeg",
				FileSize:    62529,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.jpeg",
				ContentType: "image/webp",
				FileSize:    17605,
				URL:         "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/small/01F8MH6NEM8D7527KZAECTCR76.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		"local_account_1_status_4_attachment_1": {
			ID:        "01F8MH7TDVANYKWVE8VVKFPJTJ",
			StatusID:  "01F8MH82FYRXD2RC6108DAJ5HB",
			URL:       "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH7TDVANYKWVE8VVKFPJTJ.gif",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2022-06-09T13:12:00Z"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  400,
					Height: 280,
					Size:   112000,
					Aspect: 1.4285715,
				},
				Small: gtsmodel.Small{
					Width:  400,
					Height: 280,
					Size:   112000,
					Aspect: 1.4285715,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01F8MH1H7YV1Z7D2C8K2730QBF",
			Description:       "90's Trent Reznor turning to the camera",
			ScheduledStatusID: "",
			Blurhash:          "LCDRH758KOxsEMNxENEM9]}?aKxZ",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH7TDVANYKWVE8VVKFPJTJ.gif",
				ContentType: "image/gif",
				FileSize:    1109138,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01F8MH7TDVANYKWVE8VVKFPJTJ.jpeg",
				ContentType: "image/jpeg",
				FileSize:    10270,
				URL:         "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01F8MH7TDVANYKWVE8VVKFPJTJ.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		"local_account_1_status_4_attachment_2": {
			ID:        "01CDR64G398ADCHXK08WWTHEZ5",
			StatusID:  "01F8MH82FYRXD2RC6108DAJ5HB",
			URL:       "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01CDR64G398ADCHXK08WWTHEZ5.mp4",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2022-06-09T13:12:00Z"),
			Type:      gtsmodel.FileTypeVideo,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:     720,
					Height:    404,
					Size:      290880,
					Aspect:    1.7821782,
					Duration:  util.Ptr[float32](15.034),
					Framerate: util.Ptr[float32](30.0),
					Bitrate:   util.Ptr[uint64](1209808),
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 287,
					Size:   146944,
					Aspect: 1.7839721,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01F8MH1H7YV1Z7D2C8K2730QBF",
			Description:       "A cow adorably licking another cow!",
			ScheduledStatusID: "",
			Blurhash:          "L9B|BBY8yZtS~AxZV@t6,njEjZV@",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01CDR64G398ADCHXK08WWTHEZ5.mp4",
				ContentType: "video/mp4",
				FileSize:    2273532,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01CDR64G398ADCHXK08WWTHEZ5.webp",
				ContentType: "image/webp",
				FileSize:    11570,
				URL:         "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01CDR64G398ADCHXK08WWTHEZ5.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		"local_account_1_unattached_1": {
			ID:        "01F8MH8RMYQ6MSNY3JM2XT1CQ5",
			StatusID:  "", // this attachment isn't connected to a status YET
			URL:       "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH8RMYQ6MSNY3JM2XT1CQ5.jpg",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2022-06-09T13:12:00Z"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  800,
					Height: 450,
					Size:   360000,
					Aspect: 1.7777778,
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 288,
					Size:   147456,
					Aspect: 1.7777778,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01F8MH1H7YV1Z7D2C8K2730QBF",
			Description:       "the oh you meme",
			ScheduledStatusID: "",
			Blurhash:          "LNABP8o#Dge,S6M}axxVEQjYxWbH",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01F8MH8RMYQ6MSNY3JM2XT1CQ5.jpg",
				ContentType: "image/jpeg",
				FileSize:    27759,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01F8MH8RMYQ6MSNY3JM2XT1CQ5.jpeg",
				ContentType: "image/jpeg",
				FileSize:    14665,
				URL:         "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01F8MH8RMYQ6MSNY3JM2XT1CQ5.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		"local_account_1_avatar": {
			ID:        "01F8MH58A357CV5K7R7TJMSH6S",
			StatusID:  "", // this attachment isn't connected to a status
			URL:       "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2022-06-09T13:12:00Z"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  1092,
					Height: 1800,
					Size:   1965600,
					Aspect: 0.6066667,
				},
				Small: gtsmodel.Small{
					Width:  310,
					Height: 512,
					Size:   158720,
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
			Blurhash:          "LHI:dk=G|rj]H[J-5roJvnr@Opag",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpeg",
				ContentType: "image/jpeg",
				FileSize:    457680,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.jpeg",
				ContentType: "image/jpeg",
				FileSize:    50381,
				URL:         "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(true),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		"local_account_1_header": {
			ID:        "01PFPMWK2FF0D9WMHEJHR07C3Q",
			StatusID:  "",
			URL:       "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2022-06-09T13:12:00Z"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  1018,
					Height: 764,
					Size:   777752,
					Aspect: 1.3324608,
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 384,
					Size:   196608,
					Aspect: 1.3333334,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01F8MH1H7YV1Z7D2C8K2730QBF",
			Description:       "A very old-school screenshot of the original team fortress mod for quake",
			ScheduledStatusID: "",
			Blurhash:          "L17KPDs:$ykDJroJ-RoJ0fR+xVjY",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg",
				ContentType: "image/jpeg",
				FileSize:    517226,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg",
				ContentType: "image/jpeg",
				FileSize:    26794,
				URL:         "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(true),
			Cached: util.Ptr(true),
		},
		"local_account_1_status_8_attachment_1": {
			ID:        "01J2M20K6K9XQC4WSB961YJHV6",
			StatusID:  "01J2M1HPFSS54S60Y0KYV23KJE",
			URL:       "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01J2M20K6K9XQC4WSB961YJHV6.mp3",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2024-01-10T11:24:00+02:00"),
			Type:      gtsmodel.FileTypeAudio,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:     500,
					Height:    500,
					Size:      250000,
					Aspect:    1,
					Duration:  util.Ptr[float32](185.62613),
					Framerate: util.Ptr[float32](90000),
					Bitrate:   util.Ptr[uint64](322537),
				},
				Small: gtsmodel.Small{
					Width:  500,
					Height: 500,
					Size:   250000,
					Aspect: 1,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01F8MH1H7YV1Z7D2C8K2730QBF",
			Description:       "This is a track from Nine Inch Nail's \"Ghosts I-V\" album. This is the third track from \"Ghosts II\".",
			ScheduledStatusID: "",
			Blurhash:          "LZDJO?ayIUof01j[xuayxuayayj[",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01F8MH1H7YV1Z7D2C8K2730QBF/attachment/original/01J2M20K6K9XQC4WSB961YJHV6.mp3",
				ContentType: "audio/mpeg",
				FileSize:    7483917,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01J2M20K6K9XQC4WSB961YJHV6.webp",
				ContentType: "image/webp",
				FileSize:    11624,
				URL:         "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01J2M20K6K9XQC4WSB961YJHV6.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		"local_account_2_status_9_attachment_1": {
			ID:          "01JDQ164HM08SGJ7ZEK9003Z4B",
			StatusID:    "01JDPZEZ77X1NX0TY9M10BK1HM",
			URL:         "http://localhost:8080/fileserver/01FHMQX3GAABWSM0S2VZEC2SWC/attachment/original/01HE88YG74PVAB81PX2XA9F3FG.mp3",
			RemoteURL:   "http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE892Y8ZS68TQCNPX7J888P3.mp3",
			CreatedAt:   TimeMustParse("2024-11-01T10:01:00+02:00"),
			Type:        gtsmodel.FileTypeUnknown,
			FileMeta:    gtsmodel.FileMeta{},
			AccountID:   "01F8MH5NBDF2MV7CTC4Q5128HF",
			Description: "Jolly salsa song, public domain.",
			Blurhash:    "",
			Processing:  gtsmodel.ProcessingStatusProcessed,
			File:        gtsmodel.File{},
			Thumbnail:   gtsmodel.Thumbnail{RemoteURL: ""},
			Avatar:      util.Ptr(false),
			Header:      util.Ptr(false),
			Cached:      util.Ptr(false),
		},
		"local_account_3_avatar": {
			ID:        "01JPHQZ0ZHC2AXJK1JQNXRXQZN",
			StatusID:  "", // this attachment isn't connected to a status
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/avatar/original/01JPHQZ0ZHC2AXJK1JQNXRXQZN.jpeg",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-17T10:46:37+01:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  1280,
					Height: 720,
					Size:   921600,
					Aspect: 1.777778,
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 288,
					Size:   147456,
					Aspect: 1.777778,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "LRF~2LIU0esp-qRjR*aeJ$s;iwW.",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/avatar/original/01JPHQZ0ZHC2AXJK1JQNXRXQZN.jpeg",
				ContentType: "image/jpeg",
				FileSize:    291230,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/avatar/small/01JPHQZ0ZHC2AXJK1JQNXRXQZN.jpeg",
				ContentType: "image/jpeg",
				FileSize:    24486,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/avatar/small/01JPHQZ0ZHC2AXJK1JQNXRXQZN.jpeg",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(true),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		"local_account_3_header": {
			ID:        "01JPHRB7F2RXPTEQFRYC85EPD9",
			StatusID:  "", // this attachment isn't connected to a status
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/header/original/01JPHRB7F2RXPTEQFRYC85EPD9.png",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-17T10:53:17+01:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  725,
					Height: 307,
					Size:   222575,
					Aspect: 2.361563,
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 216,
					Size:   110592,
					Aspect: 2.361563,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "L9I5h:%M%M?a~os:D*bFMybFM{jI",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/header/original/01JPHRB7F2RXPTEQFRYC85EPD9.png",
				ContentType: "image/png",
				FileSize:    405238,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/header/small/01JPHRB7F2RXPTEQFRYC85EPD9.webp",
				ContentType: "image/webp",
				FileSize:    26478,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/header/small/01JPHRB7F2RXPTEQFRYC85EPD9.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(true),
			Cached: util.Ptr(true),
		},
		// sickos
		"local_account_3_status_1_attachment_1": {
			ID:        "01JPCPRMPPGWKBCAE7X81XA0PK",
			StatusID:  "01JPCNB4417JG3XHHP0WS60RM3",
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPCPRMPPGWKBCAE7X81XA0PK.jpeg",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-15T11:49:28+01:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  1920,
					Height: 1200,
					Size:   2304000,
					Aspect: 1.600000,
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 320,
					Size:   163840,
					Aspect: 1.600000,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "L~EqXWX5t6og%jW=owa~N1WFjYWC",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPCPRMPPGWKBCAE7X81XA0PK.jpeg",
				ContentType: "image/jpeg",
				FileSize:    513277,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPCPRMPPGWKBCAE7X81XA0PK.jpeg",
				ContentType: "image/jpeg",
				FileSize:    23550,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPCPRMPPGWKBCAE7X81XA0PK.jpeg",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		// marge
		"local_account_3_status_1_attachment_2": {
			ID:        "01JPCPTSFNQDAGTHP49DXSD0BM",
			StatusID:  "01JPCNB4417JG3XHHP0WS60RM3",
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPCPTSFNQDAGTHP49DXSD0BM.png",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-15T11:50:38+01:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  976,
					Height: 741,
					Size:   723216,
					Aspect: 1.317139,
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 388,
					Size:   198656,
					Aspect: 1.317139,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "LGH1i6RpD;-,0DoZaIogA2N3xZI]",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPCPTSFNQDAGTHP49DXSD0BM.png",
				ContentType: "image/png",
				FileSize:    380878,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPCPTSFNQDAGTHP49DXSD0BM.webp",
				ContentType: "image/webp",
				FileSize:    51882,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPCPTSFNQDAGTHP49DXSD0BM.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		// sloth-gear
		"local_account_3_status_1_attachment_3": {
			ID:        "01JPCPYJ6N2E2R7GAJ1XECXNV5",
			StatusID:  "01JPCNB4417JG3XHHP0WS60RM3",
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPCPYJ6N2E2R7GAJ1XECXNV5.webp",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-15T11:52:42+01:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  2830,
					Height: 1472,
					Size:   4165760,
					Aspect: 1.922554,
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 266,
					Size:   136192,
					Aspect: 1.922554,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "LOE.|bxZx]j[~pt7WWWW%Lj@%Mj[",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPCPYJ6N2E2R7GAJ1XECXNV5.webp",
				ContentType: "image/webp",
				FileSize:    366592,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPCPYJ6N2E2R7GAJ1XECXNV5.jpeg",
				ContentType: "image/jpeg",
				FileSize:    15461,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPCPYJ6N2E2R7GAJ1XECXNV5.jpeg",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		// you-posted
		"local_account_3_status_1_attachment_4": {
			ID:        "01JPCQ4WXEA52VVR9V1HN7E0RS",
			StatusID:  "01JPCNB4417JG3XHHP0WS60RM3",
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPCQ4WXEA52VVR9V1HN7E0RS.png",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-15T11:56:09+01:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  1920,
					Height: 1080,
					Size:   2073600,
					Aspect: 1.777778,
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 288,
					Size:   147456,
					Aspect: 1.777778,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "L00+zhoLNubHj[fQa|fQ9tWVw{jZ",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPCQ4WXEA52VVR9V1HN7E0RS.png",
				ContentType: "image/png",
				FileSize:    80917,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPCQ4WXEA52VVR9V1HN7E0RS.webp",
				ContentType: "image/webp",
				FileSize:    5344,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPCQ4WXEA52VVR9V1HN7E0RS.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		// buscemi
		"local_account_3_status_1_attachment_5": {
			ID:        "01JPCQ9VBZBMSTVN56QN3R5188",
			StatusID:  "01JPCNB4417JG3XHHP0WS60RM3",
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPCQ9VBZBMSTVN56QN3R5188.jpeg",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-15T11:58:51+01:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  1077,
					Height: 525,
					Size:   565425,
					Aspect: 2.051429,
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 249,
					Size:   127488,
					Aspect: 2.051429,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "L5A9A=}?J*5m56Rk={$%O?Nb$M$i",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPCQ9VBZBMSTVN56QN3R5188.jpeg",
				ContentType: "image/jpeg",
				FileSize:    42899,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPCQ9VBZBMSTVN56QN3R5188.jpeg",
				ContentType: "image/jpeg",
				FileSize:    17341,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPCQ9VBZBMSTVN56QN3R5188.jpeg",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		// butt
		"local_account_3_status_1_attachment_6": {
			ID:        "01JPG1RZPRH3Y00VSA3RQ2SJWP",
			StatusID:  "01JPCNB4417JG3XHHP0WS60RM3",
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPG1RZPRH3Y00VSA3RQ2SJWP.gif",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-16T18:59:36+01:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  31,
					Height: 25,
					Size:   775,
					Aspect: 1.240000,
				},
				Small: gtsmodel.Small{
					Width:  31,
					Height: 25,
					Size:   775,
					Aspect: 1.240000,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "LWLN.4~q00ofxuxu-;%M9F-;-;xu",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPG1RZPRH3Y00VSA3RQ2SJWP.gif",
				ContentType: "image/gif",
				FileSize:    636,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPG1RZPRH3Y00VSA3RQ2SJWP.webp",
				ContentType: "image/webp",
				FileSize:    406,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPG1RZPRH3Y00VSA3RQ2SJWP.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		// bunny
		"local_account_3_status_2_attachment_1": {
			ID:        "01JPHFKQ86GT9W76SWPHE9P8JB",
			StatusID:  "01JPCNJAPHJKJC4EXWA6N9BXDD",
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPHFKQ86GT9W76SWPHE9P8JB.webm",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-17T08:20:38+01:00"),
			Type:      gtsmodel.FileTypeVideo,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:    640,
					Height:   360,
					Size:     230400,
					Aspect:   1.777778,
					Duration: util.Ptr[float32](32.480000),
					Bitrate:  util.Ptr[uint64](533294),
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 288,
					Size:   147456,
					Aspect: 1.777778,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "LEQcn{?bfQ?b~qoffQoffQfQfQfQ",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPHFKQ86GT9W76SWPHE9P8JB.webm",
				ContentType: "video/webm",
				FileSize:    2165608,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPHFKQ86GT9W76SWPHE9P8JB.webp",
				ContentType: "image/webp",
				FileSize:    324,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPHFKQ86GT9W76SWPHE9P8JB.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		// computerbye
		"local_account_3_status_2_attachment_2": {
			ID:        "01JPHFSCVGGH02FX9VJMXGXN45",
			StatusID:  "01JPCNJAPHJKJC4EXWA6N9BXDD",
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPHFSCVGGH02FX9VJMXGXN45.gif",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-17T08:23:44+01:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:     442,
					Height:    332,
					Size:      146744,
					Aspect:    1.331325,
					Duration:  util.Ptr[float32](3.750000),
					Framerate: util.Ptr[float32](20.000000),
					Bitrate:   util.Ptr[uint64](4078150),
				},
				Small: gtsmodel.Small{
					Width:  442,
					Height: 332,
					Size:   146744,
					Aspect: 1.331325,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "LLHUzr-;o#_2~q-:IV%Mxu%MM{M{",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPHFSCVGGH02FX9VJMXGXN45.gif",
				ContentType: "image/gif",
				FileSize:    1911633,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPHFSCVGGH02FX9VJMXGXN45.webp",
				ContentType: "image/webp",
				FileSize:    10056,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPHFSCVGGH02FX9VJMXGXN45.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		// diarrhea
		"local_account_3_status_2_attachment_3": {
			ID:        "01JPHFW5HKFWQNQ954P5KNXWSR",
			StatusID:  "01JPCNJAPHJKJC4EXWA6N9BXDD",
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPHFW5HKFWQNQ954P5KNXWSR.gif",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-17T08:25:15+01:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:     320,
					Height:    214,
					Size:      68480,
					Aspect:    1.495327,
					Duration:  util.Ptr[float32](3.100000),
					Framerate: util.Ptr[float32](10.000000),
					Bitrate:   util.Ptr[uint64](2011086),
				},
				Small: gtsmodel.Small{
					Width:  320,
					Height: 214,
					Size:   68480,
					Aspect: 1.495327,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "L78qTmNG00xZkWxsIURQ01s;?aR*",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPHFW5HKFWQNQ954P5KNXWSR.gif",
				ContentType: "image/gif",
				FileSize:    779296,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPHFW5HKFWQNQ954P5KNXWSR.webp",
				ContentType: "image/webp",
				FileSize:    10238,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPHFW5HKFWQNQ954P5KNXWSR.webp",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		// ffmpreg
		"local_account_3_status_2_attachment_4": {
			ID:        "01JPHFZP2VNS1M2RQ646BXBZQG",
			StatusID:  "01JPCNJAPHJKJC4EXWA6N9BXDD",
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPHFZP2VNS1M2RQ646BXBZQG.jpeg",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-17T08:27:10+01:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  1280,
					Height: 720,
					Size:   921600,
					Aspect: 1.777778,
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 288,
					Size:   147456,
					Aspect: 1.777778,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "LOCX.y}rIpE3,?w{S4W;9vENX8t6",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPHFZP2VNS1M2RQ646BXBZQG.jpeg",
				ContentType: "image/jpeg",
				FileSize:    137328,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPHFZP2VNS1M2RQ646BXBZQG.jpeg",
				ContentType: "image/jpeg",
				FileSize:    19775,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPHFZP2VNS1M2RQ646BXBZQG.jpeg",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		// notabug
		"local_account_3_status_2_attachment_5": {
			ID:        "01JPHG32F7M6F084WKEGAYJ40X",
			StatusID:  "01JPCNJAPHJKJC4EXWA6N9BXDD",
			URL:       "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPHG32F7M6F084WKEGAYJ40X.jpeg",
			RemoteURL: "",
			CreatedAt: TimeMustParse("2025-03-17T08:29:01+01:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  500,
					Height: 739,
					Size:   369500,
					Aspect: 0.676590,
				},
				Small: gtsmodel.Small{
					Width:  346,
					Height: 512,
					Size:   177152,
					Aspect: 0.676590,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01JPCMD83Y4WR901094YES3QC5",
			Description:       "DESCRIPTION_GOES_HERE",
			ScheduledStatusID: "",
			Blurhash:          "LTGbrRxAE1og0OR:xve-OFs6kCWY",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/original/01JPHG32F7M6F084WKEGAYJ40X.jpeg",
				ContentType: "image/jpeg",
				FileSize:    106636,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPHG32F7M6F084WKEGAYJ40X.jpeg",
				ContentType: "image/jpeg",
				FileSize:    27483,
				URL:         "http://localhost:8080/fileserver/01JPCMD83Y4WR901094YES3QC5/attachment/small/01JPHG32F7M6F084WKEGAYJ40X.jpeg",
				RemoteURL:   "",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		"remote_account_1_status_1_attachment_1": {
			ID:        "01FVW7RXPQ8YJHTEXYPE7Q8ZY0",
			StatusID:  "01FVW7JHQFSFK166WWKR8CBA6M",
			URL:       "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpg",
			RemoteURL: "http://fossbros-anonymous.io/attachments/original/13bbc3f8-2b5e-46ea-9531-40b4974d9912.jpg",
			CreatedAt: TimeMustParse("2021-09-20T12:40:37+02:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  472,
					Height: 291,
					Size:   137352,
					Aspect: 1.6219932,
				},
				Small: gtsmodel.Small{
					Width:  472,
					Height: 291,
					Size:   137352,
					Aspect: 1.6219932,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "01F8MH5ZK5VRH73AKHQM6Y9VNX",
			Description:       "tweet from thoughts of dog: i drank. all the water. in my bowl. earlier. but just now. i returned. to the same bowl. and it was. full again.. the bowl. is haunted",
			ScheduledStatusID: "",
			Blurhash:          "L3Q9_@4n9E?axW4mD$Mx~q00Di%L",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/original/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpeg",
				ContentType: "image/jpeg",
				FileSize:    19310,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/small/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.jpeg",
				ContentType: "image/jpeg",
				FileSize:    20395,
				URL:         "http://localhost:8080/fileserver/01F8MH5ZK5VRH73AKHQM6Y9VNX/attachment/small/01FVW7RXPQ8YJHTEXYPE7Q8ZY0.webp",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		"remote_account_3_header": {
			ID:        "01PFPMWK2FF0D9WMHEJHR07C3R",
			StatusID:  "",
			URL:       "http://localhost:8080/fileserver/062G5WYKY35KKD12EMSM3F8PJ8/header/original/01PFPMWK2FF0D9WMHEJHR07C3R.jpg",
			RemoteURL: "http://fossbros-anonymous.io/attachments/small/a499f55b-2d1e-4acd-98d2-1ac2ba6d79b9.jpg",
			CreatedAt: TimeMustParse("2022-06-09T13:12:00Z"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  472,
					Height: 291,
					Size:   137352,
					Aspect: 1.6219932,
				},
				Small: gtsmodel.Small{
					Width:  472,
					Height: 291,
					Size:   137352,
					Aspect: 1.6219932,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:         "062G5WYKY35KKD12EMSM3F8PJ8",
			Description:       "tweet from thoughts of dog: i drank. all the water. in my bowl. earlier. but just now. i returned. to the same bowl. and it was. full again.. the bowl. is haunted",
			ScheduledStatusID: "",
			Blurhash:          "L3Q9_@4n9E?axW4mD$Mx~q00Di%L",
			Processing:        2,
			File: gtsmodel.File{
				Path:        "062G5WYKY35KKD12EMSM3F8PJ8/attachment/original/01PFPMWK2FF0D9WMHEJHR07C3R.jpeg",
				ContentType: "image/jpeg",
				FileSize:    19310,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "062G5WYKY35KKD12EMSM3F8PJ8/attachment/small/01PFPMWK2FF0D9WMHEJHR07C3R.jpeg",
				ContentType: "image/webp",
				FileSize:    20395,
				URL:         "http://localhost:8080/fileserver/062G5WYKY35KKD12EMSM3F8PJ8/header/small/01PFPMWK2FF0D9WMHEJHR07C3R.webp",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(true),
			Cached: util.Ptr(true),
		},
		"remote_account_2_status_1_attachment_1": {
			ID:        "01HE7Y3C432WRSNS10EZM86SA5",
			StatusID:  "01HE7XJ1CG84TBKH5V9XKBVGF5",
			URL:       "http://localhost:8080/fileserver/01FHMQX3GAABWSM0S2VZEC2SWC/attachment/original/01HE7Y3C432WRSNS10EZM86SA5.jpg",
			RemoteURL: "http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE7Y6G0EMCKST3Q0914WW0MS.jpg",
			CreatedAt: TimeMustParse("2023-11-02T12:44:25+02:00"),
			Type:      gtsmodel.FileTypeImage,
			FileMeta: gtsmodel.FileMeta{
				Original: gtsmodel.Original{
					Width:  3000,
					Height: 2000,
					Size:   6000000,
					Aspect: 1.5,
				},
				Small: gtsmodel.Small{
					Width:  512,
					Height: 341,
					Size:   174592,
					Aspect: 1.5014663,
				},
				Focus: gtsmodel.Focus{
					X: 0,
					Y: 0,
				},
			},
			AccountID:   "01FHMQX3GAABWSM0S2VZEC2SWC",
			Description: "Photograph of a sloth, Public Domain.",
			Blurhash:    "LKE3VIw}0KD%a2o{M|t7NFWps:t7",
			Processing:  2,
			File: gtsmodel.File{
				Path:        "01FHMQX3GAABWSM0S2VZEC2SWC/attachment/original/01HE7Y3C432WRSNS10EZM86SA5.jpg",
				ContentType: "image/jpg",
				FileSize:    5450054,
			},
			Thumbnail: gtsmodel.Thumbnail{
				Path:        "01FHMQX3GAABWSM0S2VZEC2SWC/attachment/small/01HE7Y3C432WRSNS10EZM86SA5.webp",
				ContentType: "image/webp",
				FileSize:    55966,
				URL:         "http://localhost:8080/fileserver/01FHMQX3GAABWSM0S2VZEC2SWC/attachment/small/01HE7Y3C432WRSNS10EZM86SA5.webp",
			},
			Avatar: util.Ptr(false),
			Header: util.Ptr(false),
			Cached: util.Ptr(true),
		},
		"remote_account_2_status_1_attachment_2": {
			ID:          "01HE7ZFX9GKA5ZZVD4FACABSS9",
			StatusID:    "01HE7XJ1CG84TBKH5V9XKBVGF5",
			URL:         "http://localhost:8080/fileserver/01FHMQX3GAABWSM0S2VZEC2SWC/attachment/original/01HE7ZFX9GKA5ZZVD4FACABSS9.svg",
			RemoteURL:   "http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE7ZGJYTSYMXF927GF9353KR.svg",
			CreatedAt:   TimeMustParse("2023-11-02T12:44:25+02:00"),
			Type:        gtsmodel.FileTypeUnknown,
			FileMeta:    gtsmodel.FileMeta{},
			AccountID:   "01FHMQX3GAABWSM0S2VZEC2SWC",
			Description: "SVG line art of a sloth, public domain",
			Blurhash:    "L26*j+~qE1RP?wxut7ofRlM{R*of",
			Processing:  2,
			File:        gtsmodel.File{},
			Thumbnail:   gtsmodel.Thumbnail{RemoteURL: ""},
			Avatar:      util.Ptr(false),
			Header:      util.Ptr(false),
			Cached:      util.Ptr(false),
		},
		"remote_account_2_status_1_attachment_3": {
			ID:          "01HE88YG74PVAB81PX2XA9F3FG",
			StatusID:    "01HE7XJ1CG84TBKH5V9XKBVGF5",
			URL:         "http://localhost:8080/fileserver/01FHMQX3GAABWSM0S2VZEC2SWC/attachment/original/01HE88YG74PVAB81PX2XA9F3FG.mp3",
			RemoteURL:   "http://example.org/fileserver/01HE7Y659ZWZ02JM4AWYJZ176Q/attachment/original/01HE892Y8ZS68TQCNPX7J888P3.mp3",
			CreatedAt:   TimeMustParse("2023-11-02T12:44:25+02:00"),
			Type:        gtsmodel.FileTypeUnknown,
			FileMeta:    gtsmodel.FileMeta{},
			AccountID:   "01FHMQX3GAABWSM0S2VZEC2SWC",
			Description: "Jolly salsa song, public domain.",
			Blurhash:    "",
			Processing:  2,
			File:        gtsmodel.File{},
			Thumbnail:   gtsmodel.Thumbnail{RemoteURL: ""},
			Avatar:      util.Ptr(false),
			Header:      util.Ptr(false),
			Cached:      util.Ptr(false),
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
			CreatedAt:              TimeMustParse("2021-09-20T12:40:37+02:00"),
			UpdatedAt:              TimeMustParse("2021-09-20T12:40:37+02:00"),
			ImageRemoteURL:         "",
			ImageStaticRemoteURL:   "",
			ImageURL:               "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
			ImagePath:              "01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
			ImageStaticURL:         "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
			ImageStaticPath:        "01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
			ImageContentType:       "image/png",
			ImageStaticContentType: "image/png",
			ImageFileSize:          36702,
			ImageStaticFileSize:    6092,
			Disabled:               util.Ptr(false),
			URI:                    "http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ",
			VisibleInPicker:        util.Ptr(true),
			CategoryID:             "01GGQ8V4993XK67B2JB396YFB7",
			Cached:                 util.Ptr(true),
		},
		"yell": {
			ID:                     "01GD5KP5CQEE1R3X43Y1EHS2CW",
			Shortcode:              "yell",
			Domain:                 "fossbros-anonymous.io",
			CreatedAt:              TimeMustParse("2020-03-18T13:12:00+01:00"),
			UpdatedAt:              TimeMustParse("2020-03-18T13:12:00+01:00"),
			ImageRemoteURL:         "http://fossbros-anonymous.io/emoji/yell.gif",
			ImageStaticRemoteURL:   "",
			ImageURL:               "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01GD5KP5CQEE1R3X43Y1EHS2CW.png",
			ImagePath:              "01AY6P665V14JJR0AFVRT7311Y/emoji/original/01GD5KP5CQEE1R3X43Y1EHS2CW.png",
			ImageStaticURL:         "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01GD5KP5CQEE1R3X43Y1EHS2CW.png",
			ImageStaticPath:        "01AY6P665V14JJR0AFVRT7311Y/emoji/static/01GD5KP5CQEE1R3X43Y1EHS2CW.png",
			ImageContentType:       "image/png",
			ImageStaticContentType: "image/png",
			ImageFileSize:          10889,
			ImageStaticFileSize:    8965,
			Disabled:               util.Ptr(false),
			URI:                    "http://fossbros-anonymous.io/emoji/01GD5KP5CQEE1R3X43Y1EHS2CW",
			VisibleInPicker:        util.Ptr(false),
			CategoryID:             "",
			Cached:                 util.Ptr(false),
		},
	}
}

func NewTestEmojiCategories() map[string]*gtsmodel.EmojiCategory {
	return map[string]*gtsmodel.EmojiCategory{
		"reactions": {
			ID:        "01GGQ8V4993XK67B2JB396YFB7",
			Name:      "reactions",
			CreatedAt: TimeMustParse("2020-03-18T11:40:55+02:00"),
			UpdatedAt: TimeMustParse("2020-03-19T12:35:12+02:00"),
		},
		"cute stuff": {
			ID:        "01GGQ989PTT9PMRN4FZ1WWK2B9",
			Name:      "cute stuff",
			CreatedAt: TimeMustParse("2020-03-20T11:40:55+02:00"),
			UpdatedAt: TimeMustParse("2020-03-21T12:35:12+02:00"),
		},
	}
}

func NewTestStatusToEmojis() map[string]*gtsmodel.StatusToEmoji {
	return map[string]*gtsmodel.StatusToEmoji{
		"admin_account_status_1_rainbow": {
			StatusID: "01F8MH75CBF9JFX4ZAD54N0W0R",
			EmojiID:  "01F8MH9H8E4VG3KDYJR9EGPXCQ",
		},
	}
}

func NewTestInstances() map[string]*gtsmodel.Instance {
	return map[string]*gtsmodel.Instance{
		"localhost:8080": {
			ID:                     "01G774F5TSHJ2ZSF7XRC5EMT6K",
			CreatedAt:              TimeMustParse("2020-01-20T13:12:00+02:00"),
			UpdatedAt:              TimeMustParse("2020-01-20T13:12:00+02:00"),
			Domain:                 "localhost:8080",
			URI:                    "http://localhost:8080",
			Title:                  "GoToSocial Testrig Instance",
			ShortDescription:       "<p>This is the GoToSocial testrig. It doesn't federate or anything.</p><p>When the testrig is shut down, all data on it will be deleted.</p><p>Don't use this in production!</p>",
			ShortDescriptionText:   "This is the GoToSocial testrig. It doesn't federate or anything.\n\nWhen the testrig is shut down, all data on it will be deleted.\n\nDon't use this in production!",
			Description:            "<p>Here's a fuller description of the GoToSocial testrig instance.</p><p>This instance is for testing purposes only. It doesn't federate at all. Go check out <a href=\"https://github.com/superseriousbusiness/gotosocial/tree/main/testrig\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/tree/main/testrig</a> and <a href=\"https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing</a></p><p>Users on this instance:</p><ul><li><span class=\"h-card\"><a href=\"http://localhost:8080/@admin\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>admin</span></a></span> (admin!).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span> (posts about turtles, we don't know why).</li><li><span class=\"h-card\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>the_mighty_zork</span></a></span> (who knows).</li></ul><p>If you need to edit the models for the testrig, you can do so at <code>internal/testmodels.go</code>.</p>",
			DescriptionText:        "Here's a fuller description of the GoToSocial testrig instance.\n\nThis instance is for testing purposes only. It doesn't federate at all. Go check out https://github.com/superseriousbusiness/gotosocial/tree/main/testrig and https://github.com/superseriousbusiness/gotosocial/blob/main/CONTRIBUTING.md#testing\n\nUsers on this instance:\n\n- @admin (admin!).\n- @1happyturtle (posts about turtles, we don't know why).\n- @the_mighty_zork (who knows).\n\nIf you need to edit the models for the testrig, you can do so at `internal/testmodels.go`.",
			Terms:                  "<p>This is where a list of terms and conditions might go.</p><p>For example:</p><p>If you want to sign up on this instance, you oughta know that we:</p><ol><li>Will sell your data to whoever offers.</li><li>Secure the server with password <code>password</code> wherever possible.</li></ol>",
			TermsText:              "This is where a list of terms and conditions might go.\n\nFor example:\n\nIf you want to sign up on this instance, you oughta know that we:\n\n1. Will sell your data to whoever offers.\n2. Secure the server with password `password` wherever possible.",
			ContactEmail:           "admin@example.org",
			ContactAccountUsername: "admin",
			ContactAccountID:       "01F8MH17FWEB39HZJ76B6VXSKF",
		},
		"fossbros-anonymous.io": {
			ID:        "01G5H6YMJQKR86QZKXXQ2S95FZ",
			CreatedAt: TimeMustParse("2021-09-20T12:40:37+02:00"),
			UpdatedAt: TimeMustParse("2021-09-20T12:40:37+02:00"),
			Domain:    "fossbros-anonymous.io",
			URI:       "http://fossbros-anonymous.io",
		},
		"example.org": {
			ID:        "01G5H71G52DJKVBYKXPNPNDN1G",
			CreatedAt: TimeMustParse("2020-05-13T15:29:12+02:00"),
			UpdatedAt: TimeMustParse("2020-05-13T15:29:12+02:00"),
			Domain:    "example.org",
			URI:       "http://example.org",
		},
	}
}

func NewTestDomainBlocks() map[string]*gtsmodel.DomainBlock {
	return map[string]*gtsmodel.DomainBlock{
		"replyguys.com": {
			ID:                 "01FF22EQM7X8E3RX1XGPN7S87D",
			CreatedAt:          TimeMustParse("2020-05-13T15:29:12+02:00"),
			UpdatedAt:          TimeMustParse("2020-05-13T15:29:12+02:00"),
			Domain:             "replyguys.com",
			CreatedByAccountID: "01F8MH17FWEB39HZJ76B6VXSKF",
			PrivateComment:     "i blocked this domain because they keep replying with pushy + unwarranted linux advice",
			PublicComment:      "reply-guying to tech posts",
			Obfuscate:          util.Ptr(false),
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
			Original: "welcome-original.jpg",
			Small:    "welcome-small.jpeg",
		},
		"local_account_1_status_4_attachment_1": {
			Original: "trent-original.gif",
			Small:    "trent-small.jpeg",
		},
		"local_account_1_status_4_attachment_2": {
			Original: "cowlick-original.mp4",
			Small:    "cowlick-small.webp",
		},
		"local_account_1_unattached_1": {
			Original: "ohyou-original.jpg",
			Small:    "ohyou-small.jpeg",
		},
		"local_account_1_avatar": {
			Original: "zork-original.jpg",
			Small:    "zork-small.jpeg",
		},
		"local_account_1_header": {
			Original: "team-fortress-original.jpg",
			Small:    "team-fortress-small.jpeg",
		},
		"local_account_1_status_8_attachment_1": {
			Original: "ghosts-original.mp3",
			Small:    "ghosts-small.webp",
		},
		"local_account_3_status_1_attachment_1": {
			Original: "sickos-original.jpeg",
			Small:    "sickos-small.jpeg",
		},
		"local_account_3_status_1_attachment_2": {
			Original: "marge-original.png",
			Small:    "marge-small.webp",
		},
		"local_account_3_status_1_attachment_3": {
			Original: "sloth-gear-original.webp",
			Small:    "sloth-gear-small.jpeg",
		},
		"local_account_3_status_1_attachment_4": {
			Original: "you-posted-original.webp",
			Small:    "you-posted-small.webp",
		},
		"local_account_3_status_1_attachment_5": {
			Original: "buscemi-original.jpeg",
			Small:    "buscemi-small.jpeg",
		},
		"local_account_3_avatar": {
			Original: "dollar-original.jpeg",
			Small:    "dollar-small.jpeg",
		},
		"local_account_3_header": {
			Original: "dollar2-original.png",
			Small:    "dollar2-small.webp",
		},
		"local_account_3_status_1_attachment_6": {
			Original: "butt-original.gif",
			Small:    "butt-small.webp",
		},
		"local_account_3_status_2_attachment_1": {
			Original: "bunny-original.webm",
			Small:    "bunny-small.webp",
		},
		"local_account_3_status_2_attachment_2": {
			Original: "computerbye-original.gif",
			Small:    "computerbye-small.webp",
		},
		"local_account_3_status_2_attachment_3": {
			Original: "diarrhea-original.gif",
			Small:    "diarrhea-small.webp",
		},
		"local_account_3_status_2_attachment_4": {
			Original: "ffmpreg-original.jpeg",
			Small:    "ffmpreg-small.jpeg",
		},
		"local_account_3_status_2_attachment_5": {
			Original: "notabug-original.jpeg",
			Small:    "notabug-small.jpeg",
		},
		"remote_account_1_status_1_attachment_1": {
			Original: "thoughtsofdog-original.jpg",
			Small:    "thoughtsofdog-small.jpeg",
		},
		"remote_account_2_status_1_attachment_1": {
			Original: "sloth-original.jpg",
			Small:    "sloth-small.jpeg",
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
		"yell": {
			Original: "yell-original.png",
			Static:   "yell-static.png",
		},
	}
}

// NewTestStatuses returns a map of statuses keyed according to which account
// and status they are.
func NewTestStatuses() map[string]*gtsmodel.Status {
	return map[string]*gtsmodel.Status{
		"admin_account_status_1": {
			ID:                       "01F8MH75CBF9JFX4ZAD54N0W0R",
			PinnedAt:                 TimeMustParse("2022-05-14T13:21:09+02:00"),
			URI:                      "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
			URL:                      "http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
			Content:                  "<p>hello world! <a href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>welcome</span></a> ! first post on the instance :rainbow: !</p>",
			Text:                     "hello world! #welcome ! first post on the instance :rainbow: !",
			ContentType:              gtsmodel.StatusContentTypePlain,
			AttachmentIDs:            []string{"01F8MH6NEM8D7527KZAECTCR76"},
			TagIDs:                   []string{"01F8MHA1A2NF9MJ3WCCQ3K8BSZ"},
			EmojiIDs:                 []string{"01F8MH9H8E4VG3KDYJR9EGPXCQ"},
			CreatedAt:                TimeMustParse("2021-10-20T11:36:45Z"),
			EditedAt:                 time.Time{},
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/admin",
			AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
			ThreadID:                 "01HCWDF2Q4HV5QC161C4TGQ0M3",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"admin_account_status_2": {
			ID:                       "01F8MHAAY43M6RJ473VQFCVH37",
			PinnedAt:                 TimeMustParse("2022-05-14T14:21:09+02:00"),
			URI:                      "http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
			URL:                      "http://localhost:8080/@admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
			Content:                  "<p>🐕🐕🐕🐕🐕</p>",
			Text:                     "🐕🐕🐕🐕🐕",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2021-10-20T12:36:45Z"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/admin",
			AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
			ThreadID:                 "01HCWDQ1C7APSEY34B1HFVHVX7",
			ContentWarning:           "open to see some <strong>puppies</strong>",
			ContentWarningText:       "open to see some **puppies**",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(true),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"admin_account_status_3": {
			ID:                       "01FF25D5Q0DH7CHD57CTRS6WK0",
			URI:                      "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
			URL:                      "http://localhost:8080/@admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
			Content:                  "<p>hi <span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>the_mighty_zork</span></a></span> welcome to the instance!</p>",
			Text:                     "hi @the_mighty_zork welcome to the instance!",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2021-11-20T13:32:16Z"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/admin",
			MentionIDs:               []string{"01FF26A6BGEKCZFWNEHXB2ZZ6M"},
			AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
			InReplyToID:              "01F8MHAMCHF6Y650WCRSCP4WMY",
			InReplyToAccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToURI:             "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
			ThreadID:                 "01HCWDKKBWECZJQ93E262N36VN",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"admin_account_status_4": {
			ID:                       "01G36SF3V6Y6V5BF9P4R7PQG7G",
			URI:                      "http://localhost:8080/users/admin/statuses/01G36SF3V6Y6V5BF9P4R7PQG7G",
			URL:                      "http://localhost:8080/@admin/statuses/01G36SF3V6Y6V5BF9P4R7PQG7G",
			CreatedAt:                TimeMustParse("2021-10-20T12:41:37+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/admin",
			AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
			BoostOfID:                "01F8MHAMCHF6Y650WCRSCP4WMY",
			BoostOfAccountID:         "01F8MH1H7YV1Z7D2C8K2730QBF",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(false),
			CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"admin_account_status_5": {
			ID:                       "01J5QVB9VC76NPPRQ207GG4DRZ",
			URI:                      "http://localhost:8080/users/admin/statuses/01J5QVB9VC76NPPRQ207GG4DRZ",
			URL:                      "http://localhost:8080/@admin/statuses/01J5QVB9VC76NPPRQ207GG4DRZ",
			Content:                  `<p>Hi <span class="h-card"><a href="http://localhost:8080/@1happyturtle" class="u-url mention" rel="nofollow noreferrer noopener" target="_blank">@<span>1happyturtle</span></a></span>, can I reply?</p>`,
			Text:                     "Hi @1happyturtle, can I reply?",
			ContentType:              gtsmodel.StatusContentTypeMarkdown,
			CreatedAt:                TimeMustParse("2024-02-20T12:41:37+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/admin",
			MentionIDs:               []string{"01J5QVP69ANF1K4WHES6GA4WXP"},
			AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
			InReplyToID:              "01F8MHC8VWDRBQR0N1BATDDEM5",
			InReplyToAccountID:       "01F8MH5NBDF2MV7CTC4Q5128HF",
			InReplyToURI:             "http://localhost:8080/users/1happyturtle/statuses/01F8MHC8VWDRBQR0N1BATDDEM5",
			ThreadID:                 "01HCWE4P0EW9HBA5WHW97D5YV0",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(false),
			CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(true),
		},
		"local_account_1_status_1": {
			ID:                       "01F8MHAMCHF6Y650WCRSCP4WMY",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
			Content:                  "<p>hello everyone!</p>",
			Text:                     "hello everyone!",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			ThreadID:                 "01HCWDKKBWECZJQ93E262N36VN",
			ContentWarning:           "introduction post",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(true),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"local_account_1_status_2": {
			ID:                       "01F8MHAYFKS4KMXF8K5Y1C0KRN",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAYFKS4KMXF8K5Y1C0KRN",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAYFKS4KMXF8K5Y1C0KRN",
			Content:                  "<p>this is a Public local-only post that shouldn't federate, but it's still boostable, replyable, and likeable</p>",
			Text:                     "this is a Public local-only post that shouldn't federate, but it's still boostable, replyable, and likeable",
			ContentType:              0,
			CreatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			ThreadID:                 "01HCWDVTW3HQWSX66VJQ91Z1RH",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                util.Ptr(false),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"local_account_1_status_3": {
			ID:                       "01F8MHBBN8120SYH7D5S050MGK",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHBBN8120SYH7D5S050MGK",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01F8MHBBN8120SYH7D5S050MGK",
			Content:                  "<p>this is a very personal post that I don't want anyone to interact with at all, and i only want mutuals to see it</p>",
			Text:                     "this is a very personal post that I don't want anyone to interact with at all, and i only want mutuals to see it",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			ThreadID:                 "01HCWDY9PDNHDBDBBFTJKJY8XE",
			ContentWarning:           "test: you shouldn't be able to interact with this post in any way",
			Visibility:               gtsmodel.VisibilityMutualsOnly,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                util.Ptr(true),
			InteractionPolicy: &gtsmodel.InteractionPolicy{
				CanLike: gtsmodel.PolicyRules{
					Always: gtsmodel.PolicyValues{gtsmodel.PolicyValueAuthor},
				},
				CanReply: gtsmodel.PolicyRules{
					Always: gtsmodel.PolicyValues{gtsmodel.PolicyValueAuthor},
				},
				CanAnnounce: gtsmodel.PolicyRules{
					Always: gtsmodel.PolicyValues{gtsmodel.PolicyValueAuthor},
				},
			},
			ActivityStreamsType: ap.ObjectNote,
			PendingApproval:     util.Ptr(false),
		},
		"local_account_1_status_4": {
			ID:                       "01F8MH82FYRXD2RC6108DAJ5HB",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01F8MH82FYRXD2RC6108DAJ5HB",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01F8MH82FYRXD2RC6108DAJ5HB",
			Content:                  "<p>here's a little gif of trent.... and also a cow</p>",
			Text:                     "here's a little gif of trent.... and also a cow",
			ContentType:              gtsmodel.StatusContentTypePlain,
			AttachmentIDs:            []string{"01F8MH7TDVANYKWVE8VVKFPJTJ", "01CDR64G398ADCHXK08WWTHEZ5"},
			CreatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			EditedAt:                 time.Time{},
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			ThreadID:                 "01HCWE0H2GKH794Q7GDPANH91Q",
			ContentWarning:           "eye contact, trent reznor gif, cow",
			Visibility:               gtsmodel.VisibilityMutualsOnly,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"local_account_1_status_5": {
			ID:                       "01FCTA44PW9H1TB328S9AQXKDS",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01FCTA44PW9H1TB328S9AQXKDS",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01FCTA44PW9H1TB328S9AQXKDS",
			Content:                  "<p>hi!</p>",
			Text:                     "hi!",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2022-05-20T11:37:55Z"),
			EditedAt:                 time.Time{},
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			ThreadID:                 "01HCWE1ERQSMMVWDD0BE491E2P",
			Visibility:               gtsmodel.VisibilityFollowersOnly,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"local_account_1_status_6": {
			ID:                       "01HEN2RZ8BG29Y5Z9VJC73HZW7",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/065TKBPE0H2AH8S5X8JCK4XC58",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/065TKBPE0H2AH8S5X8JCK4XC58",
			Content:                  "<p>what do you think of sloths?</p>",
			Text:                     "what do you think of sloths?",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2022-05-20T11:41:10Z"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			Visibility:               gtsmodel.VisibilityFollowersOnly,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ActivityQuestion,
			PollID:                   "01HEN2RKT1YTEZ80SA8HGP105F",
			PendingApproval:          util.Ptr(false),
		},
		"local_account_1_status_7": {
			ID:                       "01HH9KYNQPA416TNJ53NSATP40",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01HH9KYNQPA416TNJ53NSATP40",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01HH9KYNQPA416TNJ53NSATP40",
			Content:                  "<p>Here's a bunch of HTML, read it and weep, weep then!</p><pre><code class=\"language-html\">&lt;section class=&#34;about-user&#34;&gt;\n    &lt;div class=&#34;col-header&#34;&gt;\n        &lt;h2&gt;About&lt;/h2&gt;\n    &lt;/div&gt;            \n    &lt;div class=&#34;fields&#34;&gt;\n        &lt;h3 class=&#34;sr-only&#34;&gt;Fields&lt;/h3&gt;\n        &lt;dl&gt;\n            &lt;div class=&#34;field&#34;&gt;\n                &lt;dt&gt;should you follow me?&lt;/dt&gt;\n                &lt;dd&gt;maybe!&lt;/dd&gt;\n            &lt;/div&gt;\n            &lt;div class=&#34;field&#34;&gt;\n                &lt;dt&gt;age&lt;/dt&gt;\n                &lt;dd&gt;120&lt;/dd&gt;\n            &lt;/div&gt;\n        &lt;/dl&gt;\n    &lt;/div&gt;\n    &lt;div class=&#34;bio&#34;&gt;\n        &lt;h3 class=&#34;sr-only&#34;&gt;Bio&lt;/h3&gt;\n        &lt;p&gt;i post about things that concern me&lt;/p&gt;\n    &lt;/div&gt;\n    &lt;div class=&#34;sr-only&#34; role=&#34;group&#34;&gt;\n        &lt;h3 class=&#34;sr-only&#34;&gt;Stats&lt;/h3&gt;\n        &lt;span&gt;Joined in Jun, 2022.&lt;/span&gt;\n        &lt;span&gt;8 posts.&lt;/span&gt;\n        &lt;span&gt;Followed by 1.&lt;/span&gt;\n        &lt;span&gt;Following 1.&lt;/span&gt;\n    &lt;/div&gt;\n    &lt;div class=&#34;accountstats&#34; aria-hidden=&#34;true&#34;&gt;\n        &lt;b&gt;Joined&lt;/b&gt;&lt;time datetime=&#34;2022-06-04T13:12:00.000Z&#34;&gt;Jun, 2022&lt;/time&gt;\n        &lt;b&gt;Posts&lt;/b&gt;&lt;span&gt;8&lt;/span&gt;\n        &lt;b&gt;Followed by&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;\n        &lt;b&gt;Following&lt;/b&gt;&lt;span&gt;1&lt;/span&gt;\n    &lt;/div&gt;\n&lt;/section&gt;\n</code></pre><p>There, hope you liked that!</p>",
			Text:                     "Here's a bunch of HTML, read it and weep, weep then!\n\n```html\n<section class=\"about-user\">\n <div class=\"col-header\">\n <h2>About</h2>\n </div> \n <div class=\"fields\">\n <h3 class=\"sr-only\">Fields</h3>\n <dl>\n <div class=\"field\">\n <dt>should you follow me?</dt>\n <dd>maybe!</dd>\n </div>\n <div class=\"field\">\n <dt>age</dt>\n <dd>120</dd>\n </div>… <h3 class=\"sr-only\">Stats</h3>\n <span>Joined in Jun, 2022.</span>\n <span>8 posts.</span>\n <span>Followed by 1.</span>\n <span>Following 1.</span>\n </div>\n <div class=\"accountstats\" aria-hidden=\"true\">\n <b>Joined</b><time datetime=\"2022-06-04T13:12:00.000Z\">Jun, 2022</time>\n <b>Posts</b><span>8</span>\n <b>Followed by</b><span>1</span>\n <b>Following</b><span>1</span>\n </div>\n</section>\n```\n\nThere, hope you liked that!",
			ContentType:              gtsmodel.StatusContentTypeMarkdown,
			CreatedAt:                TimeMustParse("2023-12-10T11:24:00+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			ThreadID:                 "01HH9M3FVSF5J7120X9T6PG4GF",
			ContentWarning:           "HTML in post",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(true),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"local_account_1_status_8": {
			ID:                       "01J2M1HPFSS54S60Y0KYV23KJE",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01J2M1HPFSS54S60Y0KYV23KJE",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01J2M1HPFSS54S60Y0KYV23KJE",
			Content:                  "<p>Thanks! Here's a NIN track</p>",
			Text:                     "Thanks! Here's a NIN track",
			ContentType:              gtsmodel.StatusContentTypeMarkdown,
			AttachmentIDs:            []string{"01J2M20K6K9XQC4WSB961YJHV6"},
			CreatedAt:                TimeMustParse("2024-01-10T11:24:00+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToID:              "01FF25D5Q0DH7CHD57CTRS6WK0",
			InReplyToAccountID:       "01F8MH17FWEB39HZJ76B6VXSKF",
			InReplyToURI:             "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
			ThreadID:                 "01HCWDKKBWECZJQ93E262N36VN",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
		},
		"local_account_1_status_9": {
			ID:                       "01JDPZC707CKDN8N4QVWM4Z1NR",
			URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01JDPZC707CKDN8N4QVWM4Z1NR",
			URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01JDPZC707CKDN8N4QVWM4Z1NR",
			Content:                  "<p>this is the latest revision of the status, with a content-warning</p>",
			Text:                     "this is the latest revision of the status, with a content-warning",
			ContentType:              gtsmodel.StatusContentTypeMarkdown,
			ContentWarning:           "edited status",
			CreatedAt:                TimeMustParse("2024-11-01T11:00:00+02:00"),
			EditedAt:                 TimeMustParse("2024-11-01T11:02:00+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH1H7YV1Z7D2C8K2730QBF",
			EditIDs:                  []string{"01JDPZCZ2Y9KSGZW0R7ZG8T8Y2", "01JDPZDADMD1T9HKF94RECF7PP"},
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
		},
		"local_account_2_status_1": {
			ID:                       "01F8MHBQCBTDKN6X5VHGMMN4MA",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01F8MHBQCBTDKN6X5VHGMMN4MA",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01F8MHBQCBTDKN6X5VHGMMN4MA",
			Content:                  "<p>🐢 hi everyone i post about turtles 🐢</p>",
			Text:                     "🐢 hi everyone i post about turtles 🐢",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			ThreadID:                 "01HCWE2Q24FWCZE41AS77SDFRZ",
			ContentWarning:           "introduction post",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(true),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"local_account_2_status_2": {
			ID:                       "01F8MHC0H0A7XHTVH5F596ZKBM",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01F8MHC0H0A7XHTVH5F596ZKBM",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01F8MHC0H0A7XHTVH5F596ZKBM",
			Content:                  "<p>🐢 this one is federated, likeable, and boostable but not replyable 🐢</p>",
			Text:                     "🐢 this one is federated, likeable, and boostable but not replyable 🐢",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			ThreadID:                 "01HCWE3P291Z3NJEJVFPW0K9ZQ",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(true),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                util.Ptr(true),
			InteractionPolicy: &gtsmodel.InteractionPolicy{
				CanLike: gtsmodel.PolicyRules{
					Always: gtsmodel.PolicyValues{gtsmodel.PolicyValuePublic},
				},
				CanReply: gtsmodel.PolicyRules{
					Always: gtsmodel.PolicyValues{gtsmodel.PolicyValueAuthor},
				},
				CanAnnounce: gtsmodel.PolicyRules{
					Always: gtsmodel.PolicyValues{gtsmodel.PolicyValuePublic},
				},
			},
			ActivityStreamsType: ap.ObjectNote,
			PendingApproval:     util.Ptr(false),
		},
		"local_account_2_status_3": {
			ID:                       "01F8MHC8VWDRBQR0N1BATDDEM5",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01F8MHC8VWDRBQR0N1BATDDEM5",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01F8MHC8VWDRBQR0N1BATDDEM5",
			Content:                  "<p>🐢 i don't mind people sharing and liking this one but I want to moderate replies to it 🐢</p>",
			Text:                     "🐢 i don't mind people sharing and liking this one but I want to moderate replies to it 🐢",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			ThreadID:                 "01HCWE4P0EW9HBA5WHW97D5YV0",
			ContentWarning:           "you won't be able to reply to this without my approval",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(true),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                util.Ptr(true),
			InteractionPolicy: &gtsmodel.InteractionPolicy{
				CanLike: gtsmodel.PolicyRules{
					Always: gtsmodel.PolicyValues{gtsmodel.PolicyValuePublic},
				},
				CanReply: gtsmodel.PolicyRules{
					Always:       gtsmodel.PolicyValues{gtsmodel.PolicyValueAuthor},
					WithApproval: gtsmodel.PolicyValues{gtsmodel.PolicyValuePublic},
				},
				CanAnnounce: gtsmodel.PolicyRules{
					Always: gtsmodel.PolicyValues{gtsmodel.PolicyValuePublic},
				},
			},
			ActivityStreamsType: ap.ObjectNote,
			PendingApproval:     util.Ptr(false),
		},
		"local_account_2_status_4": {
			ID:                       "01F8MHCP5P2NWYQ416SBA0XSEV",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01F8MHCP5P2NWYQ416SBA0XSEV",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01F8MHCP5P2NWYQ416SBA0XSEV",
			Content:                  "<p>🐢 this is a public status but I want it local only and not boostable 🐢</p>",
			Text:                     "🐢 this is a public status but I want it local only and not boostable 🐢",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			ThreadID:                 "01HCWE5JXFPFP3P5W2QNHVVV27",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(true),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                util.Ptr(false),
			InteractionPolicy: &gtsmodel.InteractionPolicy{
				CanLike: gtsmodel.PolicyRules{
					Always: gtsmodel.PolicyValues{gtsmodel.PolicyValuePublic},
				},
				CanReply: gtsmodel.PolicyRules{
					Always: gtsmodel.PolicyValues{gtsmodel.PolicyValuePublic},
				},
				CanAnnounce: gtsmodel.PolicyRules{
					Always: gtsmodel.PolicyValues{gtsmodel.PolicyValueAuthor},
				},
			},
			ActivityStreamsType: ap.ObjectNote,
			PendingApproval:     util.Ptr(false),
		},
		"local_account_2_status_5": {
			ID:                       "01FCQSQ667XHJ9AV9T27SJJSX5",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01FCQSQ667XHJ9AV9T27SJJSX5",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01FCQSQ667XHJ9AV9T27SJJSX5",
			Content:                  "<p>🐢 <span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>the_mighty_zork</span></a></span> hi zork! 🐢</p>",
			Text:                     "🐢 @the_mighty_zork hi zork! 🐢",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			MentionIDs:               []string{"01FDF2HM2NF6FSRZCDEDV451CN"},
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			InReplyToID:              "01F8MHAMCHF6Y650WCRSCP4WMY",
			InReplyToAccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF",
			InReplyToURI:             "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
			ThreadID:                 "01HCWDKKBWECZJQ93E262N36VN",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"local_account_2_status_6": {
			ID:                       "01FN3VJGFH10KR7S2PB0GFJZYG",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01FN3VJGFH10KR7S2PB0GFJZYG",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01FN3VJGFH10KR7S2PB0GFJZYG",
			Content:                  "<p>🐢 <span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>the_mighty_zork</span></a></span> hi zork, this is a direct message, shhhhhh! 🐢</p>",
			Text:                     "🐢 @the_mighty_zork hi zork, this is a direct message, shhhhhh! 🐢",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			MentionIDs:               []string{"01FDF2HM2NF6FSRZCDEDV451CN"},
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			ThreadID:                 "01HCWE71MGRRDSHBKXFD5DDSWR",
			Visibility:               gtsmodel.VisibilityDirect,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"local_account_2_status_7": {
			ID:                       "01G20ZM733MGN8J344T4ZDDFY1",
			PinnedAt:                 TimeMustParse("2021-03-18T09:13:55+02:00"),
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01G20ZM733MGN8J344T4ZDDFY1",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01G20ZM733MGN8J344T4ZDDFY1",
			Content:                  "<p>🐢 hi followers! did u know i'm a turtle? 🐢</p>",
			Text:                     "🐢 hi followers! did u know i'm a turtle? 🐢",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2021-10-20T12:40:37+02:00"),
			EditedAt:                 time.Time{},
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			ThreadID:                 "01HCWE7ZNC2SS4P05WA5QYED23",
			Visibility:               gtsmodel.VisibilityFollowersOnly,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PendingApproval:          util.Ptr(false),
		},
		"local_account_2_status_8": {
			ID:                       "01HEN2PRXT0TF4YDRA64FZZRN7",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01HEN2PRXT0TF4YDRA64FZZRN7",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01HEN2PRXT0TF4YDRA64FZZRN7",
			Content:                  "<p>hey everyone i got stuck in a shed. any ideas for how to get out?</p>",
			Text:                     "hey everyone i got stuck in a shed. any ideas for how to get out?",
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2021-07-28T10:40:37+02:00"),
			EditedAt:                 time.Time{},
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/1happyturtle",
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ActivityQuestion,
			PollID:                   "01HEN2QB5NR4NCEHGYC3HN84K6",
			PendingApproval:          util.Ptr(false),
		},
		"local_account_2_status_9": {
			ID:                       "01JDPZEZ77X1NX0TY9M10BK1HM",
			URI:                      "http://localhost:8080/users/1happyturtle/statuses/01JDPZEZ77X1NX0TY9M10BK1HM",
			URL:                      "http://localhost:8080/@1happyturtle/statuses/01JDPZEZ77X1NX0TY9M10BK1HM",
			Content:                  "<p>now edited to bring back the previous edit's media!</p>",
			Text:                     "now edited to bring back the previous edit's media!",
			ContentType:              gtsmodel.StatusContentTypeMarkdown,
			ContentWarning:           "edit with media attachments",
			AttachmentIDs:            []string{"01JDQ164HM08SGJ7ZEK9003Z4B"},
			CreatedAt:                TimeMustParse("2024-11-01T10:00:00+02:00"),
			EditedAt:                 TimeMustParse("2024-11-01T10:03:00+02:00"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/the_mighty_zork",
			AccountID:                "01F8MH5NBDF2MV7CTC4Q5128HF",
			EditIDs:                  []string{"01JDPZPBXAX0M02YSEPB21KX4R", "01JDPZPJHKP7E3M0YQXEXPS1YT", "01JDPZPY3F85Y7B78ETRXEMWD9"},
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
		},
		"local_account_3_status_1": {
			ID:  "01JPCNB4417JG3XHHP0WS60RM3",
			URI: "http://localhost:8080/users/media_mogul/statuses/01JPCNB4417JG3XHHP0WS60RM3",
			URL: "http://localhost:8080/@media_mogul/statuses/01JPCNB4417JG3XHHP0WS60RM3",
			AttachmentIDs: []string{
				"01JPCPRMPPGWKBCAE7X81XA0PK",
				"01JPCPTSFNQDAGTHP49DXSD0BM",
				"01JPCPYJ6N2E2R7GAJ1XECXNV5",
				"01JPCQ4WXEA52VVR9V1HN7E0RS",
				"01JPCQ9VBZBMSTVN56QN3R5188",
				"01JPG1RZPRH3Y00VSA3RQ2SJWP",
			},
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2025-03-15T11:26:17Z"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/media_mogul",
			AccountID:                "01JPCMD83Y4WR901094YES3QC5",
			Visibility:               gtsmodel.VisibilityUnlocked,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
			PinnedAt:                 TimeMustParse("2025-03-15T11:27:00Z"),
		},
		"local_account_3_status_2": {
			ID:  "01JPCNJAPHJKJC4EXWA6N9BXDD",
			URI: "http://localhost:8080/users/media_mogul/statuses/01JPCNJAPHJKJC4EXWA6N9BXDD",
			URL: "http://localhost:8080/@media_mogul/statuses/01JPCNJAPHJKJC4EXWA6N9BXDD",
			AttachmentIDs: []string{
				"01JPHFKQ86GT9W76SWPHE9P8JB",
				"01JPHFSCVGGH02FX9VJMXGXN45",
				"01JPHFW5HKFWQNQ954P5KNXWSR",
				"01JPHFZP2VNS1M2RQ646BXBZQG",
				"01JPHG32F7M6F084WKEGAYJ40X",
			},
			ContentType:              gtsmodel.StatusContentTypePlain,
			CreatedAt:                TimeMustParse("2025-03-15T11:28:42Z"),
			Local:                    util.Ptr(true),
			AccountURI:               "http://localhost:8080/users/media_mogul",
			AccountID:                "01JPCMD83Y4WR901094YES3QC5",
			Visibility:               gtsmodel.VisibilityUnlocked,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGY43H3N2C8EWPR2FPYEXG",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
		},
		"remote_account_1_status_1": {
			ID:                  "01FVW7JHQFSFK166WWKR8CBA6M",
			URI:                 "http://fossbros-anonymous.io/users/foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M",
			URL:                 "http://fossbros-anonymous.io/@foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M",
			Content:             "<p>dark souls status bot: \"thoughts of dog\"</p>",
			AttachmentIDs:       []string{"01FVW7RXPQ8YJHTEXYPE7Q8ZY0"},
			CreatedAt:           TimeMustParse("2021-09-20T12:40:37+02:00"),
			Local:               util.Ptr(false),
			AccountURI:          "http://fossbros-anonymous.io/users/foss_satan",
			AccountID:           "01F8MH5ZK5VRH73AKHQM6Y9VNX",
			Visibility:          gtsmodel.VisibilityUnlocked,
			Sensitive:           util.Ptr(false),
			Language:            "en",
			Federated:           util.Ptr(true),
			ActivityStreamsType: ap.ObjectNote,
			PendingApproval:     util.Ptr(false),
		},
		"remote_account_1_status_2": {
			ID:                  "01HEN2QRFA8H3C6QPN7RD4KSR6",
			URI:                 "http://fossbros-anonymous.io/users/foss_satan/statuses/01HEN2QRFA8H3C6QPN7RD4KSR6",
			URL:                 "http://fossbros-anonymous.io/@foss_satan/statuses/01HEN2QRFA8H3C6QPN7RD4KSR6",
			Content:             "<p>what products should i buy at the grocery store?</p>",
			AttachmentIDs:       []string{"01FVW7RXPQ8YJHTEXYPE7Q8ZY0"},
			CreatedAt:           TimeMustParse("2021-09-11T11:40:37+02:00"),
			Local:               util.Ptr(false),
			AccountURI:          "http://fossbros-anonymous.io/users/foss_satan",
			AccountID:           "01F8MH5ZK5VRH73AKHQM6Y9VNX",
			Visibility:          gtsmodel.VisibilityUnlocked,
			Sensitive:           util.Ptr(false),
			Language:            "en",
			Federated:           util.Ptr(true),
			ActivityStreamsType: ap.ActivityQuestion,
			PollID:              "01HEN2R65468ZG657C4ZPHJ4EX",
			PendingApproval:     util.Ptr(false),
		},
		"remote_account_1_status_3": {
			ID:                  "01HEWV37MHV8BAC8ANFGVRRM5D",
			URI:                 "http://fossbros-anonymous.io/users/foss_satan/statuses/01HEWV37MHV8BAC8ANFGVRRM5D",
			URL:                 "http://fossbros-anonymous.io/@foss_satan/statuses/01HEWV37MHV8BAC8ANFGVRRM5D",
			Content:             "<p>what products should i buy at the grocery store? (now an endless poll!)</p>",
			AttachmentIDs:       []string{"01FVW7RXPQ8YJHTEXYPE7Q8ZY0"},
			CreatedAt:           TimeMustParse("2021-09-11T11:40:37+02:00"),
			Local:               util.Ptr(false),
			AccountURI:          "http://fossbros-anonymous.io/users/foss_satan",
			AccountID:           "01F8MH5ZK5VRH73AKHQM6Y9VNX",
			Visibility:          gtsmodel.VisibilityUnlocked,
			Sensitive:           util.Ptr(false),
			Language:            "en",
			Federated:           util.Ptr(true),
			ActivityStreamsType: ap.ActivityQuestion,
			PollID:              "01HEWV1GW2D49R919NPEDXPTZ5",
			PendingApproval:     util.Ptr(false),
		},
		"remote_account_1_status_4": {
			ID:                       "01JDQ07JZTX9CMDJP67CNA71YD",
			URI:                      "http://fossbros-anonymous.io/users/foss_satan/statuses/______",
			URL:                      "http://fossbros-anonymous.io/@foss_satan/statuses/______",
			Content:                  "<p>this is the latest status edit without poll change</p>",
			Text:                     "this is the latest status edit without poll change",
			ContentType:              gtsmodel.StatusContentTypeMarkdown,
			CreatedAt:                TimeMustParse("2024-11-01T09:00:00+02:00"),
			EditedAt:                 TimeMustParse("2024-11-01T09:02:00+02:00"),
			Local:                    util.Ptr(false),
			AccountURI:               "http://fossbros-anonymous.io/users/foss_satan",
			AccountID:                "01F8MH5ZK5VRH73AKHQM6Y9VNX",
			EditIDs:                  []string{"01JDQ07ZZ4FGP13YN8TF63P5A6", "01JDQ08AYQC0G6413VAHA51CV9"},
			PollID:                   "01JDQ0EZ5HM9T4WXRQ5WSVD40J",
			Visibility:               gtsmodel.VisibilityPublic,
			Sensitive:                util.Ptr(false),
			Language:                 "en",
			CreatedWithApplicationID: "01F8MGYG9E893WRHW0TAEXR8GJ",
			Federated:                util.Ptr(true),
			ActivityStreamsType:      ap.ObjectNote,
		},
		"remote_account_2_status_1": {
			ID:                  "01HE7XJ1CG84TBKH5V9XKBVGF5",
			URI:                 "http://example.org/users/Some_User/statuses/01HE7XJ1CG84TBKH5V9XKBVGF5",
			URL:                 "http://example.org/@Some_User/statuses/01HE7XJ1CG84TBKH5V9XKBVGF5",
			Content:             `<p>hi <span class="h-card"><a href="http://localhost:8080/@admin" class="u-url mention" rel="nofollow noreferrer noopener" target="_blank">@<span>admin</span></a></span> here's some media for ya</p>`,
			AttachmentIDs:       []string{"01HE7Y3C432WRSNS10EZM86SA5", "01HE7ZFX9GKA5ZZVD4FACABSS9", "01HE88YG74PVAB81PX2XA9F3FG"},
			CreatedAt:           TimeMustParse("2023-11-02T12:44:25+02:00"),
			Local:               util.Ptr(false),
			AccountURI:          "http://example.org/users/Some_User",
			MentionIDs:          []string{"01HE7XQNMKTVC8MNPCE1JGK4J3"},
			AccountID:           "01FHMQX3GAABWSM0S2VZEC2SWC",
			InReplyToID:         "01F8MH75CBF9JFX4ZAD54N0W0R",
			InReplyToAccountID:  "01F8MH17FWEB39HZJ76B6VXSKF",
			InReplyToURI:        "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
			ContentWarning:      "some unknown media included",
			Visibility:          gtsmodel.VisibilityPublic,
			Sensitive:           util.Ptr(true),
			Language:            "en",
			Federated:           util.Ptr(true),
			ActivityStreamsType: ap.ObjectNote,
			PendingApproval:     util.Ptr(false),
		},
	}
}

func NewTestPolls() map[string]*gtsmodel.Poll {
	return map[string]*gtsmodel.Poll{
		"local_account_1_status_6_poll": {
			ID:         "01HEN2RKT1YTEZ80SA8HGP105F",
			Multiple:   util.Ptr(false),
			HideCounts: util.Ptr(true),
			Options:    []string{"good", "bad", "meh"},
			Votes:      []int{2, 0, 0}, // needs to match stored poll votes
			Voters:     util.Ptr(2),    // needs to match stored poll votes
			StatusID:   "01HEN2RZ8BG29Y5Z9VJC73HZW7",
			Status:     nil,
			ExpiresAt:  TimeMustParse("2022-05-21T11:41:10Z"),
			ClosedAt:   time.Time{},
			Closing:    false,
		},
		"local_account_2_status_8_poll": {
			ID:         "01HEN2QB5NR4NCEHGYC3HN84K6",
			Multiple:   util.Ptr(false),
			HideCounts: util.Ptr(false),
			Options:    []string{"50:50", "phone a friend", "ask the audience"},
			Votes:      []int{0, 1, 1}, // needs to match stored poll votes
			Voters:     util.Ptr(2),    // needs to match stored poll votes
			StatusID:   "01HEN2PRXT0TF4YDRA64FZZRN7",
			Status:     nil,
			ExpiresAt:  TimeMustParse("2021-08-28T10:40:37+02:00"),
			ClosedAt:   TimeMustParse("2021-08-28T10:40:37+02:00"),
			Closing:    false,
		},
		"remote_account_1_status_2_poll": {
			ID:         "01HEN2R65468ZG657C4ZPHJ4EX",
			Multiple:   util.Ptr(true),
			HideCounts: util.Ptr(false),
			Options:    []string{"vaseline", "tissues", "financial times"},
			Votes:      []int{3, 2, 18},
			Voters:     util.Ptr(6),
			StatusID:   "01HEN2QRFA8H3C6QPN7RD4KSR6",
			Status:     nil,
			ExpiresAt:  TimeMustParse("2021-09-11T12:40:37+02:00"),
			ClosedAt:   TimeMustParse("2021-09-11T12:40:37+02:00"),
			Closing:    false,
		},
		"remote_account_1_status_3_poll": {
			ID:         "01HEWV1GW2D49R919NPEDXPTZ5",
			Multiple:   util.Ptr(true),
			HideCounts: util.Ptr(false),
			Options:    []string{"vaseline", "tissues", "financial times"},
			Votes:      []int{0, 0, 0},
			Voters:     util.Ptr(0),
			StatusID:   "01HEWV37MHV8BAC8ANFGVRRM5D",
			Status:     nil,
			// nil expiry AND closed date, i.e. no end
			ExpiresAt: time.Time{},
			ClosedAt:  time.Time{},
			Closing:   false,
		},
		"remote_account_1_status_4_poll": {
			ID:         "01JDQ0EZ5HM9T4WXRQ5WSVD40J",
			Multiple:   util.Ptr(false),
			HideCounts: util.Ptr(false),
			Options:    []string{"yes", "no", "maybe", "i don't know", "can you repeat the question"},
			Votes:      []int{0, 0, 0, 0, 2},
			Voters:     util.Ptr(2),
			StatusID:   "01JDQ07JZTX9CMDJP67CNA71YD",
			// empty expiry AND closed date, i.e. no end
			ExpiresAt: time.Time{},
			ClosedAt:  time.Time{},
			Closing:   false,
		},
	}
}

func NewTestPollVotes() map[string]*gtsmodel.PollVote {
	return map[string]*gtsmodel.PollVote{
		"local_account_1_status_6_poll_vote_local_account_2": {
			ID:        "01HEN2VN4DZ4ENCK6AS4PKM5B3",
			Choices:   []int{0},
			AccountID: "01F8MH5NBDF2MV7CTC4Q5128HF",
			Account:   nil,
			PollID:    "01HEN2RKT1YTEZ80SA8HGP105F",
			Poll:      nil,
			CreatedAt: TimeMustParse("2022-05-20T14:41:10Z"),
		},
		"local_account_1_status_6_poll_vote_remote_account_1": {
			ID:        "01HEN2VM975JG8N9KPFQ597KGF",
			Choices:   []int{0},
			AccountID: "01F8MH5ZK5VRH73AKHQM6Y9VNX",
			Account:   nil,
			PollID:    "01HEN2RKT1YTEZ80SA8HGP105F",
			Poll:      nil,
			CreatedAt: TimeMustParse("2022-05-20T15:41:10Z"),
		},
		"local_account_2_status_8_poll_vote_local_account_1": {
			ID:        "01HEN2VK9TX5BTD3B0CSRBWE89",
			Choices:   []int{2},
			AccountID: "01F8MH1H7YV1Z7D2C8K2730QBF",
			Account:   nil,
			PollID:    "01HEN2QB5NR4NCEHGYC3HN84K6",
			Poll:      nil,
			CreatedAt: TimeMustParse("2021-07-29T10:40:37+02:00"),
		},
		"local_account_2_status_8_poll_vote_remote_account_1": {
			ID:        "01HEN2VHW4HAHBM4YH3N55794D",
			Choices:   []int{1},
			AccountID: "01F8MH5ZK5VRH73AKHQM6Y9VNX",
			Account:   nil,
			PollID:    "01HEN2QB5NR4NCEHGYC3HN84K6",
			Poll:      nil,
			CreatedAt: TimeMustParse("2021-08-10T10:40:37+02:00"),
		},
		"remote_account_1_status_2_poll_vote_local_account_1": {
			ID:        "01HEN2VH077W1QY7VKQFPKD6B6",
			Choices:   []int{1, 2},
			AccountID: "01F8MH1H7YV1Z7D2C8K2730QBF",
			Account:   nil,
			PollID:    "01HEN2R65468ZG657C4ZPHJ4EX",
			Poll:      nil,
			CreatedAt: TimeMustParse("2021-09-11T11:45:37+02:00"),
		},
		"remote_account_1_status_2_poll_vote_local_account_2": {
			ID:        "01HEN2VG6EP3GJA208586H356K",
			Choices:   []int{0, 2},
			AccountID: "01F8MH5NBDF2MV7CTC4Q5128HF",
			Account:   nil,
			PollID:    "01HEN2R65468ZG657C4ZPHJ4EX",
			Poll:      nil,
			CreatedAt: TimeMustParse("2021-09-11T11:47:37+02:00"),
		},
		"remote_account_1_status_4_poll_vote_local_account_1": {
			ID:        "01JDQ0SX9QVVFHS7P8M1PA3SVG",
			Choices:   []int{4},
			AccountID: "01F8MH1H7YV1Z7D2C8K2730QBF",
			Account:   nil,
			PollID:    "01JDQ0EZ5HM9T4WXRQ5WSVD40J",
			Poll:      nil,
			CreatedAt: TimeMustParse("2024-11-01T09:01:30+02:00"),
		},
		"remote_account_1_status_4_poll_vote_local_account_2": {
			ID:        "01JDQ0T3EEDN7SAVBQMQP4PR12",
			Choices:   []int{4},
			AccountID: "01F8MH5NBDF2MV7CTC4Q5128HF",
			Account:   nil,
			PollID:    "01JDQ0EZ5HM9T4WXRQ5WSVD40J",
			Poll:      nil,
			CreatedAt: TimeMustParse("2024-11-01T09:02:30+02:00"),
		},
	}
}

// NewTestTags returns a map of gts model tags keyed by their name
func NewTestTags() map[string]*gtsmodel.Tag {
	return map[string]*gtsmodel.Tag{
		"welcome": {
			ID:        "01F8MHA1A2NF9MJ3WCCQ3K8BSZ",
			Name:      "welcome",
			CreatedAt: TimeMustParse("2022-05-14T13:21:09+02:00"),
			UpdatedAt: TimeMustParse("2022-05-14T13:21:09+02:00"),
			Useable:   util.Ptr(true),
			Listable:  util.Ptr(true),
		},
		"Hashtag": {
			ID:        "01FCT9SGYA71487N8D0S1M638G",
			Name:      "hashtag",
			CreatedAt: TimeMustParse("2022-05-14T13:21:09+02:00"),
			UpdatedAt: TimeMustParse("2022-05-14T13:21:09+02:00"),
			Useable:   util.Ptr(true),
			Listable:  util.Ptr(true),
		},
	}
}

func NewTestStatusToTags() map[string]*gtsmodel.StatusToTag {
	return map[string]*gtsmodel.StatusToTag{
		"admin_account_status_1_welcome": {
			StatusID: "01F8MH75CBF9JFX4ZAD54N0W0R",
			TagID:    "01F8MHA1A2NF9MJ3WCCQ3K8BSZ",
		},
	}
}

func NewTestThreads() map[string]*gtsmodel.Thread {
	return map[string]*gtsmodel.Thread{
		"admin_account_status_1": {
			ID: "01HCWDF2Q4HV5QC161C4TGQ0M3",
		},
		"admin_account_status_2": {
			ID: "01HCWDQ1C7APSEY34B1HFVHVX7",
		},
		"local_account_1_status_1": {
			ID: "01HCWDKKBWECZJQ93E262N36VN",
		},
		"local_account_1_status_2": {
			ID: "01HCWDVTW3HQWSX66VJQ91Z1RH",
		},
		"local_account_1_status_3": {
			ID: "01HCWDY9PDNHDBDBBFTJKJY8XE",
		},
		"local_account_1_status_4": {
			ID: "01HCWE0H2GKH794Q7GDPANH91Q",
		},
		"local_account_1_status_5": {
			ID: "01HCWE1ERQSMMVWDD0BE491E2P",
		},
		"local_account_1_status_7": {
			ID: "01HH9M3FVSF5J7120X9T6PG4GF",
		},
		"local_account_2_status_1": {
			ID: "01HCWE2Q24FWCZE41AS77SDFRZ",
		},
		"local_account_2_status_2": {
			ID: "01HCWE3P291Z3NJEJVFPW0K9ZQ",
		},
		"local_account_2_status_3": {
			ID: "01HCWE4P0EW9HBA5WHW97D5YV0",
		},
		"local_account_2_status_4": {
			ID: "01HCWE5JXFPFP3P5W2QNHVVV27",
		},
		"local_account_2_status_6": {
			ID: "01HCWE71MGRRDSHBKXFD5DDSWR",
		},
		"local_account_2_status_7": {
			ID: "01HCWE7ZNC2SS4P05WA5QYED23",
		},
	}
}

func NewTestThreadToStatus() []*gtsmodel.ThreadToStatus {
	return []*gtsmodel.ThreadToStatus{
		{
			ThreadID: "01HCWDF2Q4HV5QC161C4TGQ0M3",
			StatusID: "01F8MH75CBF9JFX4ZAD54N0W0R",
		},
		{
			ThreadID: "01HCWDQ1C7APSEY34B1HFVHVX7",
			StatusID: "01F8MHAAY43M6RJ473VQFCVH37",
		},
		{
			ThreadID: "01HCWDKKBWECZJQ93E262N36VN",
			StatusID: "01FF25D5Q0DH7CHD57CTRS6WK0",
		},
		{
			ThreadID: "01HCWDKKBWECZJQ93E262N36VN",
			StatusID: "01F8MHAMCHF6Y650WCRSCP4WMY",
		},
		{
			ThreadID: "01HCWDVTW3HQWSX66VJQ91Z1RH",
			StatusID: "01F8MHAYFKS4KMXF8K5Y1C0KRN",
		},
		{
			ThreadID: "01HCWDY9PDNHDBDBBFTJKJY8XE",
			StatusID: "01F8MHBBN8120SYH7D5S050MGK",
		},
		{
			ThreadID: "01HCWE0H2GKH794Q7GDPANH91Q",
			StatusID: "01F8MH82FYRXD2RC6108DAJ5HB",
		},
		{
			ThreadID: "01HCWE1ERQSMMVWDD0BE491E2P",
			StatusID: "01FCTA44PW9H1TB328S9AQXKDS",
		},
		{
			ThreadID: "01HCWE2Q24FWCZE41AS77SDFRZ",
			StatusID: "01F8MHBQCBTDKN6X5VHGMMN4MA",
		},
		{
			ThreadID: "01HCWE3P291Z3NJEJVFPW0K9ZQ",
			StatusID: "01F8MHC0H0A7XHTVH5F596ZKBM",
		},
		{
			ThreadID: "01HCWE4P0EW9HBA5WHW97D5YV0",
			StatusID: "01F8MHC8VWDRBQR0N1BATDDEM5",
		},
		{
			ThreadID: "01HCWDKKBWECZJQ93E262N36VN",
			StatusID: "01FCQSQ667XHJ9AV9T27SJJSX5",
		},
		{
			ThreadID: "01HCWDKKBWECZJQ93E262N36VN",
			StatusID: "01J2M1HPFSS54S60Y0KYV23KJE",
		},
		{
			ThreadID: "01HCWE71MGRRDSHBKXFD5DDSWR",
			StatusID: "01FN3VJGFH10KR7S2PB0GFJZYG",
		},
		{
			ThreadID: "01HCWE7ZNC2SS4P05WA5QYED23",
			StatusID: "01G20ZM733MGN8J344T4ZDDFY1",
		},
		{
			ThreadID: "01HCWE4P0EW9HBA5WHW97D5YV0",
			StatusID: "01J5QVB9VC76NPPRQ207GG4DRZ",
		},
	}
}

// NewTestMentions returns a map of gts model mentions keyed by their name.
func NewTestMentions() map[string]*gtsmodel.Mention {
	return map[string]*gtsmodel.Mention{
		"zork_mention_foss_satan": {
			ID:               "01FCTA2Y6FGHXQA4ZE6N5NMNEX",
			StatusID:         "01FCTA44PW9H1TB328S9AQXKDS",
			CreatedAt:        TimeMustParse("2022-05-14T13:21:09+02:00"),
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
			CreatedAt:        TimeMustParse("2022-05-14T13:21:09+02:00"),
			OriginAccountID:  "01F8MH5NBDF2MV7CTC4Q5128HF",
			OriginAccountURI: "http://localhost:8080/users/1happyturtle",
			TargetAccountID:  "01F8MH1H7YV1Z7D2C8K2730QBF",
			NameString:       "@the_mighty_zork",
			TargetAccountURI: "http://localhost:8080/users/the_mighty_zork",
			TargetAccountURL: "http://localhost:8080/@the_mighty_zork",
		},
		"local_user_2_mention_zork_direct_message": {
			ID:               "01FN3VKDEF4CN2W9TKX339BEHB",
			StatusID:         "01FN3VJGFH10KR7S2PB0GFJZYG",
			CreatedAt:        TimeMustParse("2022-05-14T13:21:09+02:00"),
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
			CreatedAt:        TimeMustParse("2022-05-14T13:21:09+02:00"),
			OriginAccountID:  "01F8MH17FWEB39HZJ76B6VXSKF",
			OriginAccountURI: "http://localhost:8080/users/admin",
			TargetAccountID:  "01F8MH1H7YV1Z7D2C8K2730QBF",
			NameString:       "@the_mighty_zork",
			TargetAccountURI: "http://localhost:8080/users/the_mighty_zork",
			TargetAccountURL: "http://localhost:8080/@the_mighty_zork",
		},
		"admin_account_mention_turtle": {
			ID:               "01J5QVP69ANF1K4WHES6GA4WXP",
			StatusID:         "01J5QVB9VC76NPPRQ207GG4DRZ",
			CreatedAt:        TimeMustParse("2024-02-20T12:41:37+02:00"),
			OriginAccountID:  "01F8MH17FWEB39HZJ76B6VXSKF",
			OriginAccountURI: "http://localhost:8080/users/admin",
			TargetAccountID:  "01F8MH5NBDF2MV7CTC4Q5128HF",
			NameString:       "@1happyturtle",
			TargetAccountURI: "http://localhost:8080/users/1happyturtle",
			TargetAccountURL: "http://localhost:8080/@1happyturtle",
		},
		"remote_account_2_mention_admin": {
			ID:               "01HE7XQNMKTVC8MNPCE1JGK4J3",
			StatusID:         "01HE7XJ1CG84TBKH5V9XKBVGF5",
			CreatedAt:        TimeMustParse("2023-11-02T12:44:25+02:00"),
			OriginAccountID:  "01FHMQX3GAABWSM0S2VZEC2SWC",
			OriginAccountURI: "http://example.org/users/Some_User",
			TargetAccountID:  "01F8MH17FWEB39HZJ76B6VXSKF",
			NameString:       "@admin@localhost:8080",
			TargetAccountURI: "http://localhost:8080/users/admin",
			TargetAccountURL: "http://localhost:8080/@admin",
		},
	}
}

// NewTestFaves returns a map of gts model faves, keyed in the format [faving_account]_[target_status]
func NewTestFaves() map[string]*gtsmodel.StatusFave {
	return map[string]*gtsmodel.StatusFave{
		"local_account_1_admin_account_status_1": {
			ID:              "01F8MHD2QCZSZ6WQS2ATVPEYJ9",
			CreatedAt:       TimeMustParse("2022-05-14T13:21:09+02:00"),
			AccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF", // local account 1
			TargetAccountID: "01F8MH17FWEB39HZJ76B6VXSKF", // admin account
			StatusID:        "01F8MH75CBF9JFX4ZAD54N0W0R", // admin account status 1
			URI:             "http://localhost:8080/users/the_mighty_zork/liked/01F8MHD2QCZSZ6WQS2ATVPEYJ9",
		},
		"local_account_1_admin_account_status_3": {
			ID:              "01GM435XERVPXXRK6NBAHK5HCZ",
			CreatedAt:       TimeMustParse("2022-12-12T20:17:56+02:00"),
			AccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF", // local account 1
			TargetAccountID: "01F8MH17FWEB39HZJ76B6VXSKF", // admin account
			StatusID:        "01F8MHAAY43M6RJ473VQFCVH37", // admin account status 1
			URI:             "http://localhost:8080/users/the_mighty_zork/liked/01GM435XERVPXXRK6NBAHK5HCZ",
		},
		"local_account_1_local_account_2_status_1": {
			ID:              "01GM43AKBMN4YNXQ1HZHVC1SGB",
			CreatedAt:       TimeMustParse("2022-12-12T20:19:49+02:00"),
			AccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF", // local account 1
			TargetAccountID: "01F8MH5NBDF2MV7CTC4Q5128HF", // admin account
			StatusID:        "01F8MHBQCBTDKN6X5VHGMMN4MA", // admin account status 1
			URI:             "http://localhost:8080/users/the_mighty_zork/liked/01GM43AKBMN4YNXQ1HZHVC1SGB",
		},
		"local_account_1_local_account_2_status_4": {
			ID:              "01GM43CC47DRPNZZ7BD04BS1YZ",
			CreatedAt:       TimeMustParse("2022-12-12T20:20:47+02:00"),
			AccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF", // local account 1
			TargetAccountID: "01F8MH5NBDF2MV7CTC4Q5128HF", // admin account
			StatusID:        "01F8MHCP5P2NWYQ416SBA0XSEV", // admin account status 1
			URI:             "http://localhost:8080/users/the_mighty_zork/liked/01GM43CC47DRPNZZ7BD04BS1YZ",
		},
		"admin_account_local_account_1_status_1": {
			ID:              "01F8Q0486ANTDWKG02A7DS1Q24",
			CreatedAt:       TimeMustParse("2022-05-14T13:21:09+02:00"),
			AccountID:       "01F8MH17FWEB39HZJ76B6VXSKF", // admin account
			TargetAccountID: "01F8MH1H7YV1Z7D2C8K2730QBF", // local account 1
			StatusID:        "01F8MHAMCHF6Y650WCRSCP4WMY", // local account status 1
			URI:             "http://localhost:8080/users/admin/liked/01F8Q0486ANTDWKG02A7DS1Q24",
		},
	}
}

// NewTestAccountNotes returns some account notes for use in testing.
func NewTestAccountNotes() map[string]*gtsmodel.AccountNote {
	return map[string]*gtsmodel.AccountNote{
		"local_account_2_note_on_1": {
			ID:              "01H53TM628GNC4ZDNRGQGPK8S0",
			AccountID:       "01F8MH5NBDF2MV7CTC4Q5128HF",
			TargetAccountID: "01F8MH1H7YV1Z7D2C8K2730QBF",
			Comment:         "extremely average poster",
		},
	}
}

// NewTestNotifications returns some notifications for use in testing.
func NewTestNotifications() map[string]*gtsmodel.Notification {
	return map[string]*gtsmodel.Notification{
		"local_account_1_like": {
			ID:               "01F8Q0ANPTWW10DAKTX7BRPBJP",
			NotificationType: gtsmodel.NotificationFavourite,
			CreatedAt:        TimeMustParse("2022-05-14T13:21:09+02:00"),
			TargetAccountID:  "01F8MH1H7YV1Z7D2C8K2730QBF",
			OriginAccountID:  "01F8MH17FWEB39HZJ76B6VXSKF",
			StatusID:         "01F8MHAMCHF6Y650WCRSCP4WMY",
			Read:             util.Ptr(false),
		},
		"local_account_2_like": {
			ID:               "01GTS6PRPXJYZBPFFQ56PP0XR8",
			NotificationType: gtsmodel.NotificationFavourite,
			CreatedAt:        TimeMustParse("2022-01-13T12:45:01+02:00"),
			TargetAccountID:  "01F8MH17FWEB39HZJ76B6VXSKF",
			OriginAccountID:  "01F8MH5NBDF2MV7CTC4Q5128HF",
			StatusID:         "01F8MH75CBF9JFX4ZAD54N0W0R",
			Read:             util.Ptr(false),
		},
		"new_signup": {
			ID:               "01HTM9TETMB3YQCBKZ7KD4KV02",
			NotificationType: gtsmodel.NotificationAdminSignup,
			CreatedAt:        TimeMustParse("2022-06-04T13:12:00Z"),
			TargetAccountID:  "01F8MH17FWEB39HZJ76B6VXSKF",
			OriginAccountID:  "01F8MH0BBE4FHXPH513MBVFHB0",
			StatusID:         "",
			Read:             util.Ptr(false),
		},
	}
}

// NewTestFollows returns some follows for use in testing.
func NewTestFollows() map[string]*gtsmodel.Follow {
	return map[string]*gtsmodel.Follow{
		"local_account_1_admin_account": {
			ID:              "01F8PY8RHWRQZV038T4E8T9YK8",
			CreatedAt:       TimeMustParse("2022-05-14T16:21:09+02:00"),
			UpdatedAt:       TimeMustParse("2022-05-14T16:21:09+02:00"),
			AccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF",
			TargetAccountID: "01F8MH17FWEB39HZJ76B6VXSKF",
			ShowReblogs:     util.Ptr(true),
			URI:             "http://localhost:8080/users/the_mighty_zork/follow/01F8PY8RHWRQZV038T4E8T9YK8",
			Notify:          util.Ptr(false),
		},
		"local_account_1_local_account_2": {
			ID:              "01F8PYDCE8XE23GRE5DPZJDZDP",
			CreatedAt:       TimeMustParse("2022-05-14T15:21:09+02:00"),
			UpdatedAt:       TimeMustParse("2022-05-14T15:21:09+02:00"),
			AccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF",
			TargetAccountID: "01F8MH5NBDF2MV7CTC4Q5128HF",
			ShowReblogs:     util.Ptr(true),
			URI:             "http://localhost:8080/users/the_mighty_zork/follow/01F8PYDCE8XE23GRE5DPZJDZDP",
			Notify:          util.Ptr(false),
		},
		"local_account_2_local_account_1": {
			ID:              "01G1TK1RS4K3E0MSFTXBFWAH9Q",
			CreatedAt:       TimeMustParse("2022-05-14T14:21:09+02:00"),
			UpdatedAt:       TimeMustParse("2022-05-14T14:21:09+02:00"),
			AccountID:       "01F8MH5NBDF2MV7CTC4Q5128HF",
			TargetAccountID: "01F8MH1H7YV1Z7D2C8K2730QBF",
			ShowReblogs:     util.Ptr(true),
			URI:             "http://localhost:8080/users/1happyturtle/follow/01F8PYDCE8XE23GRE5DPZJDZDP",
			Notify:          util.Ptr(false),
		},
		"admin_account_local_account_1": {
			ID:              "01G1TK3PQKFW1BQZ9WVYRTFECK",
			CreatedAt:       TimeMustParse("2022-05-14T13:21:09+02:00"),
			UpdatedAt:       TimeMustParse("2022-05-14T13:21:09+02:00"),
			AccountID:       "01F8MH17FWEB39HZJ76B6VXSKF",
			TargetAccountID: "01F8MH1H7YV1Z7D2C8K2730QBF",
			ShowReblogs:     util.Ptr(true),
			URI:             "http://localhost:8080/users/admin/follow/01G1TK3PQKFW1BQZ9WVYRTFECK",
			Notify:          util.Ptr(false),
		},
	}
}

func NewTestLists() map[string]*gtsmodel.List {
	return map[string]*gtsmodel.List{
		"local_account_1_list_1": {
			ID:            "01H0G8E4Q2J3FE3JDWJVWEDCD1",
			CreatedAt:     TimeMustParse("2022-05-14T13:21:09+02:00"),
			UpdatedAt:     TimeMustParse("2022-05-14T13:21:09+02:00"),
			Title:         "Cool Ass Posters From This Instance",
			AccountID:     "01F8MH1H7YV1Z7D2C8K2730QBF",
			RepliesPolicy: gtsmodel.RepliesPolicyFollowed,
			Exclusive:     util.Ptr(false),
		},
	}
}

func NewTestListEntries() map[string]*gtsmodel.ListEntry {
	return map[string]*gtsmodel.ListEntry{
		"local_account_1_list_1_entry_1": {
			ID:        "01H0G89MWVQE0M58VD2HQYMQWH",
			CreatedAt: TimeMustParse("2022-05-14T13:21:09+02:00"),
			UpdatedAt: TimeMustParse("2022-05-14T13:21:09+02:00"),
			ListID:    "01H0G8E4Q2J3FE3JDWJVWEDCD1",
			FollowID:  "01F8PYDCE8XE23GRE5DPZJDZDP",
		},
		"local_account_1_list_1_entry_2": {
			ID:        "01H0G8FFM1AGQDRNGBGGX8CYJQ",
			CreatedAt: TimeMustParse("2022-05-14T13:21:09+02:00"),
			UpdatedAt: TimeMustParse("2022-05-14T13:21:09+02:00"),
			ListID:    "01H0G8E4Q2J3FE3JDWJVWEDCD1",
			FollowID:  "01F8PY8RHWRQZV038T4E8T9YK8",
		},
	}
}

func NewTestMarkers() map[string]*gtsmodel.Marker {
	return map[string]*gtsmodel.Marker{
		"local_account_1_home_marker": {
			AccountID:  "01F8MH1H7YV1Z7D2C8K2730QBF",
			Name:       gtsmodel.MarkerNameHome,
			UpdatedAt:  TimeMustParse("2022-05-14T13:21:09+02:00"),
			Version:    0,
			LastReadID: "01F8MH82FYRXD2RC6108DAJ5HB",
		},
		"local_account_1_notification_marker": {
			AccountID:  "01F8MH1H7YV1Z7D2C8K2730QBF",
			Name:       gtsmodel.MarkerNameNotifications,
			UpdatedAt:  TimeMustParse("2022-05-14T13:21:09+02:00"),
			Version:    4,
			LastReadID: "01F8Q0ANPTWW10DAKTX7BRPBJP",
		},
	}
}

func NewTestBlocks() map[string]*gtsmodel.Block {
	return map[string]*gtsmodel.Block{
		"local_account_2_block_remote_account_1": {
			ID:              "01FEXXET6XXMF7G2V3ASZP3YQW",
			CreatedAt:       TimeMustParse("2022-05-14T13:21:09+02:00"),
			UpdatedAt:       TimeMustParse("2022-05-14T13:21:09+02:00"),
			URI:             "http://localhost:8080/users/1happyturtle/blocks/01FEXXET6XXMF7G2V3ASZP3YQW",
			AccountID:       "01F8MH5NBDF2MV7CTC4Q5128HF",
			TargetAccountID: "01F8MH5ZK5VRH73AKHQM6Y9VNX",
		},
	}
}

func NewTestReports() map[string]*gtsmodel.Report {
	return map[string]*gtsmodel.Report{
		"local_account_2_report_remote_account_1": {
			ID:              "01GP3AWY4CRDVRNZKW0TEAMB5R",
			CreatedAt:       TimeMustParse("2022-05-14T12:20:03+02:00"),
			UpdatedAt:       TimeMustParse("2022-05-14T12:20:03+02:00"),
			URI:             "http://localhost:8080/reports/01GP3AWY4CRDVRNZKW0TEAMB5R",
			AccountID:       "01F8MH5NBDF2MV7CTC4Q5128HF",
			TargetAccountID: "01F8MH5ZK5VRH73AKHQM6Y9VNX",
			Comment:         "dark souls sucks, please yeet this nerd",
			StatusIDs:       []string{"01FVW7JHQFSFK166WWKR8CBA6M"},
			Forwarded:       util.Ptr(true),
			RuleIDs:         []string{"01GP3AWY4CRDVRNZKW0TEAMB51", "01GP3DFY9XQ1TJMZT5BGAZPXX3"},
		},
		"remote_account_1_report_local_account_2": {
			ID:                     "01GP3DFY9XQ1TJMZT5BGAZPXX7",
			CreatedAt:              TimeMustParse("2022-05-15T16:20:12+02:00"),
			UpdatedAt:              TimeMustParse("2022-05-15T16:20:12+02:00"),
			URI:                    "http://fossbros-anonymous.io/87fb1478-ac46-406a-8463-96ce05645219",
			AccountID:              "01F8MH5ZK5VRH73AKHQM6Y9VNX",
			TargetAccountID:        "01F8MH5NBDF2MV7CTC4Q5128HF",
			Comment:                "this is a turtle, not a person, therefore should not be a poster",
			StatusIDs:              []string{},
			RuleIDs:                []string{},
			Forwarded:              util.Ptr(true),
			ActionTaken:            "user was warned not to be a turtle anymore",
			ActionTakenAt:          TimeMustParse("2022-05-15T17:01:56+02:00"),
			ActionTakenByAccountID: "01F8MH17FWEB39HZJ76B6VXSKF",
		},
	}
}

func NewTestRules() map[string]*gtsmodel.Rule {
	return map[string]*gtsmodel.Rule{
		"rule1": {
			ID:        "01GP3AWY4CRDVRNZKW0TEAMB51",
			CreatedAt: TimeMustParse("2022-05-14T12:20:03+02:00"),
			UpdatedAt: TimeMustParse("2022-05-14T12:20:03+02:00"),
			Text:      "Be gay",
			Deleted:   util.Ptr(false),
			Order:     util.Ptr(uint(0)),
		},
		"deleted_rule": {
			ID:        "01GP3DFY9XQ1TJMZT5BGAZPXX2",
			CreatedAt: TimeMustParse("2022-05-15T16:20:12+02:00"),
			UpdatedAt: TimeMustParse("2022-05-15T16:20:12+02:00"),
			Text:      "Deleted",
			Deleted:   util.Ptr(true),
			Order:     util.Ptr(uint(1)),
		},
		"rule2": {
			ID:        "01GP3DFY9XQ1TJMZT5BGAZPXX3",
			CreatedAt: TimeMustParse("2022-05-15T16:20:12+02:00"),
			UpdatedAt: TimeMustParse("2022-05-15T16:20:12+02:00"),
			Text:      "Do crime",
			Deleted:   util.Ptr(false),
			Order:     util.Ptr(uint(2)),
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
	dmForZork := NewAPNote(
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan/statuses/5424b153-4553-4f30-9358-7b92f7cd42f6"),
		URLMustParse("http://fossbros-anonymous.io/@foss_satan/5424b153-4553-4f30-9358-7b92f7cd42f6"),
		TimeMustParse("2022-07-13T12:13:12+02:00"),
		"@the_mighty_zork@localhost:8080 hey zork here's a new private note for you",
		"new note for zork",
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan"),
		[]*url.URL{URLMustParse("http://localhost:8080/users/the_mighty_zork")},
		nil,
		true,
		[]vocab.ActivityStreamsMention{newAPMention(
			URLMustParse("http://localhost:8080/users/the_mighty_zork"),
			"@the_mighty_zork@localhost:8080",
		)},
		[]vocab.TootHashtag{},
		nil,
	)
	createDmForZork := WrapAPNoteInCreate(
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan/statuses/5424b153-4553-4f30-9358-7b92f7cd42f6/activity"),
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan"),
		TimeMustParse("2022-07-13T12:13:12+02:00"),
		dmForZork)
	createDmForZorkSig, createDmForZorkDigest, creatDmForZorkDate := GetSignatureForActivity(createDmForZork, accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, URLMustParse(accounts["local_account_1"].InboxURI))

	replyToTurtle := NewAPNote(
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan/statuses/2f1195a6-5cb0-4475-adf5-92ab9a0147fe"),
		URLMustParse("http://fossbros-anonymous.io/@foss_satan/2f1195a6-5cb0-4475-adf5-92ab9a0147fe"),
		TimeMustParse("2022-07-13T12:13:12+02:00"),
		"@1happyturtle@localhost:8080 u suck lol",
		"",
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan"),
		[]*url.URL{URLMustParse("http://fossbros-anonymous.io/users/foss_satan/followers")},
		[]*url.URL{URLMustParse("http://localhost:8080/users/1happyturtle")},
		false,
		[]vocab.ActivityStreamsMention{newAPMention(
			URLMustParse("http://localhost:8080/users/1happyturtle"),
			"@1happyturtle@localhost:8080",
		)},
		[]vocab.TootHashtag{},
		nil,
	)
	createReplyToTurtle := WrapAPNoteInCreate(
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan/statuses/2f1195a6-5cb0-4475-adf5-92ab9a0147fe"),
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan"),
		TimeMustParse("2022-07-13T12:13:12+02:00"),
		replyToTurtle)
	createReplyToTurtleForZorkSig, createReplyToTurtleForZorkDigest, createReplyToTurtleForZorkDate := GetSignatureForActivity(createReplyToTurtle, accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, URLMustParse(accounts["local_account_1"].InboxURI))
	createReplyToTurtleForTurtleSig, createReplyToTurtleForTurtleDigest, createReplyToTurtleForTurtleDate := GetSignatureForActivity(createReplyToTurtle, accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, URLMustParse(accounts["local_account_2"].InboxURI))

	forwardedMessage := NewAPNote(
		URLMustParse("http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1"),
		URLMustParse("http://example.org/@Some_User/afaba698-5740-4e32-a702-af61aa543bc1"),
		TimeMustParse("2022-07-13T12:13:12+02:00"),
		"this is a public status, please forward it!",
		"",
		URLMustParse("http://example.org/users/Some_User"),
		[]*url.URL{ap.PublicURI()},
		nil,
		false,
		[]vocab.ActivityStreamsMention{},
		[]vocab.TootHashtag{},
		[]vocab.ActivityStreamsImage{
			newAPImage(
				URLMustParse("http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1/attachment1.jpg"),
				"image/jpeg",
				"trent reznor looking handsome as balls",
				"LEDara58O=t5EMSOENEN9]}?aK%0"),
		},
	)
	createForwardedMessage := WrapAPNoteInCreate(
		URLMustParse("http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1/activity"),
		URLMustParse("http://example.org/users/Some_User"),
		TimeMustParse("2022-07-13T12:13:12+02:00"),
		forwardedMessage)
	createForwardedMessageSig, createForwardedMessageDigest, createForwardedMessageDate := GetSignatureForActivity(createForwardedMessage, accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, URLMustParse(accounts["local_account_1"].InboxURI))

	announceForwarded1Zork := newAPAnnounce(
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan/first_announce"),
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan"),
		TimeMustParse("2022-07-13T12:13:12+02:00"),
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan/followers"),
		forwardedMessage,
	)
	announceForwarded1ZorkSig, announceForwarded1ZorkDigest, announceForwarded1ZorkDate := GetSignatureForActivity(announceForwarded1Zork, accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, URLMustParse(accounts["local_account_1"].InboxURI))

	announceForwarded1Turtle := newAPAnnounce(
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan/first_announce"),
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan"),
		TimeMustParse("2022-07-13T12:13:12+02:00"),
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan/followers"),
		forwardedMessage,
	)
	announceForwarded1TurtleSig, announceForwarded1TurtleDigest, announceForwarded1TurtleDate := GetSignatureForActivity(announceForwarded1Turtle, accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, URLMustParse(accounts["local_account_2"].InboxURI))

	announceForwarded2Zork := newAPAnnounce(
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan/second_announce"),
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan"),
		TimeMustParse("2022-07-13T12:13:12+02:00"),
		URLMustParse("http://fossbros-anonymous.io/users/foss_satan/followers"),
		forwardedMessage,
	)
	announceForwarded2ZorkSig, announceForwarded2ZorkDigest, announceForwarded2ZorkDate := GetSignatureForActivity(announceForwarded2Zork, accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, URLMustParse(accounts["local_account_1"].InboxURI))

	deleteForRemoteAccount3 := newAPDelete(
		URLMustParse("https://somewhere.mysterious/users/rest_in_piss"),
		URLMustParse("https://somewhere.mysterious/users/rest_in_piss"),
		TimeMustParse("2022-07-13T12:13:12+02:00"),
		URLMustParse(accounts["local_account_1"].URI),
	)
	// it doesn't really matter what key we use to sign this, since we're not going to be able to verify if anyway
	keyToSignDelete := accounts["remote_account_1"].PrivateKey
	deleteForRemoteAccount3Sig, deleteForRemoteAccount3Digest, deleteForRemoteAccount3Date := GetSignatureForActivity(deleteForRemoteAccount3, "https://somewhere.mysterious/users/rest_in_piss#main-key", keyToSignDelete, URLMustParse(accounts["local_account_1"].InboxURI))

	return map[string]ActivityWithSignature{
		"dm_for_zork": {
			Activity:        createDmForZork,
			SignatureHeader: createDmForZorkSig,
			DigestHeader:    createDmForZorkDigest,
			DateHeader:      creatDmForZorkDate,
		},
		"reply_to_turtle_for_zork": {
			Activity:        createReplyToTurtle,
			SignatureHeader: createReplyToTurtleForZorkSig,
			DigestHeader:    createReplyToTurtleForZorkDigest,
			DateHeader:      createReplyToTurtleForZorkDate,
		},
		"reply_to_turtle_for_turtle": {
			Activity:        createReplyToTurtle,
			SignatureHeader: createReplyToTurtleForTurtleSig,
			DigestHeader:    createReplyToTurtleForTurtleDigest,
			DateHeader:      createReplyToTurtleForTurtleDate,
		},
		"forwarded_message": {
			Activity:        createForwardedMessage,
			SignatureHeader: createForwardedMessageSig,
			DigestHeader:    createForwardedMessageDigest,
			DateHeader:      createForwardedMessageDate,
		},
		"announce_forwarded_1_zork": {
			Activity:        announceForwarded1Zork,
			SignatureHeader: announceForwarded1ZorkSig,
			DigestHeader:    announceForwarded1ZorkDigest,
			DateHeader:      announceForwarded1ZorkDate,
		},
		"announce_forwarded_1_turtle": {
			Activity:        announceForwarded1Turtle,
			SignatureHeader: announceForwarded1TurtleSig,
			DigestHeader:    announceForwarded1TurtleDigest,
			DateHeader:      announceForwarded1TurtleDate,
		},
		"announce_forwarded_2_zork": {
			Activity:        announceForwarded2Zork,
			SignatureHeader: announceForwarded2ZorkSig,
			DigestHeader:    announceForwarded2ZorkDigest,
			DateHeader:      announceForwarded2ZorkDate,
		},
		"delete_https://somewhere.mysterious/users/rest_in_piss#main-key": {
			Activity:        deleteForRemoteAccount3,
			SignatureHeader: deleteForRemoteAccount3Sig,
			DigestHeader:    deleteForRemoteAccount3Digest,
			DateHeader:      deleteForRemoteAccount3Date,
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

	turnipLover6969Priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	turnipLover6969Pub := &turnipLover6969Priv.PublicKey

	someUserPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	someUserPub := &someUserPriv.PublicKey

	return map[string]vocab.ActivityStreamsPerson{
		"https://unknown-instance.com/users/brand_new_person": newAPPerson(
			URLMustParse("https://unknown-instance.com/users/brand_new_person"),
			URLMustParse("https://unknown-instance.com/users/brand_new_person/following"),
			URLMustParse("https://unknown-instance.com/users/brand_new_person/followers"),
			URLMustParse("https://unknown-instance.com/users/brand_new_person/inbox"),
			nil,
			URLMustParse("https://unknown-instance.com/users/brand_new_person/outbox"),
			URLMustParse("https://unknown-instance.com/users/brand_new_person/collections/featured"),
			nil,
			nil,
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
		"https://turnip.farm/users/turniplover6969": newAPPerson(
			URLMustParse("https://turnip.farm/users/turniplover6969"),
			URLMustParse("https://turnip.farm/users/turniplover6969/following"),
			URLMustParse("https://turnip.farm/users/turniplover6969/followers"),
			URLMustParse("https://turnip.farm/users/turniplover6969/inbox"),
			URLMustParse("https://turnip.farm/sharedInbox"),
			URLMustParse("https://turnip.farm/users/turniplover6969/outbox"),
			URLMustParse("https://turnip.farm/users/turniplover6969/collections/featured"),
			nil,
			nil,
			"turniplover6969",
			"Turnip Lover 6969",
			"I just think they're neat",
			URLMustParse("https://turnip.farm/@turniplover6969"),
			true,
			URLMustParse("https://turnip.farm/users/turniplover6969#main-key"),
			turnipLover6969Pub,
			nil,
			"image/jpeg",
			nil,
			"image/png",
			false,
		),
		"http://example.org/users/Some_User": newAPPerson(
			URLMustParse("http://example.org/users/Some_User"),
			URLMustParse("http://example.org/users/Some_User/following"),
			URLMustParse("http://example.org/users/Some_User/followers"),
			URLMustParse("http://example.org/users/Some_User/inbox"),
			URLMustParse("http://example.org/sharedInbox"),
			URLMustParse("http://example.org/users/Some_User/outbox"),
			URLMustParse("http://example.org/users/Some_User/collections/featured"),
			nil,
			nil,
			"Some_User",
			"just some user, don't mind me",
			"Peepee poo poo",
			URLMustParse("http://example.org/@Some_User"),
			true,
			URLMustParse("http://example.org/users/Some_User#main-key"),
			someUserPub,
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
		"https://unknown-instance.com/groups/some_group": newAPGroup(
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

func NewTestFediServices() map[string]vocab.ActivityStreamsService {
	newService1Priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	newService1Pub := &newService1Priv.PublicKey

	return map[string]vocab.ActivityStreamsService{
		"https://owncast.example.org/federation/user/rgh": newAPService(
			URLMustParse("https://owncast.example.org/federation/user/rgh"),
			nil,
			URLMustParse("https://owncast.example.org/federation/user/rgh/followers"),
			URLMustParse("https://owncast.example.org/federation/user/rgh/inbox"),
			URLMustParse("https://owncast.example.org/federation/user/rgh/outbox"),
			nil,
			"rgh",
			"linux audio stuff ",
			"",
			URLMustParse("https://owncast.example.org/federation/user/rgh"),
			true,
			URLMustParse("https://owncast.example.org/federation/user/rgh#main-key"),
			newService1Pub,
			nil,
			"image/jpeg",
			nil,
			"image/png",
			false,
		),
	}
}

func NewTestFediEmojis() map[string]vocab.TootEmoji {
	return map[string]vocab.TootEmoji{
		"http://fossbros-anonymous.io/emoji/01GD5HCC2YECT012TK8PAGX4D1": newAPEmoji(
			URLMustParse("http://fossbros-anonymous.io/emoji/01GD5HCC2YECT012TK8PAGX4D1"),
			"kip_van_den_bos",
			TimeMustParse("2022-09-13T12:13:12+02:00"),
			newAPImage(
				URLMustParse("http://fossbros-anonymous.io/emoji/kip.gif"),
				"image/gif",
				"",
				"",
			),
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

	thoughtsOfDogBytes, err := os.ReadFile(fmt.Sprintf("%s/thoughtsofdog-original.jpg", relativePath))
	if err != nil {
		panic(err)
	}

	massiveFuckingTurnipBytes, err := os.ReadFile(fmt.Sprintf("%s/giant-turnip-world-record.jpg", relativePath))
	if err != nil {
		panic(err)
	}

	peglinBytes, err := os.ReadFile(fmt.Sprintf("%s/peglin.gif", relativePath))
	if err != nil {
		panic(err)
	}

	kipBytes, err := os.ReadFile(fmt.Sprintf("%s/kip-original.gif", relativePath))
	if err != nil {
		panic(err)
	}

	yellBytes, err := os.ReadFile(fmt.Sprintf("%s/yell-original.png", relativePath))
	if err != nil {
		panic(err)
	}

	return map[string]RemoteAttachmentFile{
		"https://s3-us-west-2.amazonaws.com/plushcity/media_attachments/files/106/867/380/219/163/828/original/88e8758c5f011439.jpg": {
			Data:        beeBytes,
			ContentType: "image/jpeg",
		},
		"http://fossbros-anonymous.io/attachments/original/13bbc3f8-2b5e-46ea-9531-40b4974d9912.jpg": {
			Data:        thoughtsOfDogBytes,
			ContentType: "image/jpeg",
		},
		"https://turnip.farm/attachments/f17843c7-015e-4251-9b5a-91389c49ee57.jpg": {
			Data:        massiveFuckingTurnipBytes,
			ContentType: "image/jpeg",
		},
		"http://example.org/media/emojis/1781772.gif": {
			Data:        peglinBytes,
			ContentType: "image/gif",
		},
		"http://fossbros-anonymous.io/emoji/kip.gif": {
			Data:        kipBytes,
			ContentType: "image/gif",
		},
		"http://fossbros-anonymous.io/emoji/yell.gif": {
			Data:        yellBytes,
			ContentType: "image/png",
		},
	}
}

func NewTestFediStatuses() map[string]vocab.ActivityStreamsNote {
	return map[string]vocab.ActivityStreamsNote{
		"http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1": NewAPNote(
			URLMustParse("http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1"),
			URLMustParse("http://example.org/@Some_User/afaba698-5740-4e32-a702-af61aa543bc1"),
			TimeMustParse("2022-07-13T12:13:12+02:00"),
			"this is a public status, please forward it!",
			"",
			URLMustParse("http://example.org/users/Some_User"),
			[]*url.URL{ap.PublicURI()},
			nil,
			false,
			[]vocab.ActivityStreamsMention{},
			[]vocab.TootHashtag{},
			[]vocab.ActivityStreamsImage{
				newAPImage(
					URLMustParse("http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1/attachment1.jpg"),
					"image/jpeg",
					"trent reznor looking handsome as balls",
					"LEDara58O=t5EMSOENEN9]}?aK%0"),
			},
		),
		"https://unknown-instance.com/users/brand_new_person/statuses/01FE4NTHKWW7THT67EF10EB839": NewAPNote(
			URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01FE4NTHKWW7THT67EF10EB839"),
			URLMustParse("https://unknown-instance.com/users/@brand_new_person/01FE4NTHKWW7THT67EF10EB839"),
			TimeMustParse("2022-07-13T12:13:12+02:00"),
			"Hello world!",
			"",
			URLMustParse("https://unknown-instance.com/users/brand_new_person"),
			[]*url.URL{
				ap.PublicURI(),
			},
			[]*url.URL{},
			false,
			nil,
			[]vocab.TootHashtag{},
			nil,
		),
		"https://unknown-instance.com/users/brand_new_person/statuses/01FE5Y30E3W4P7TRE0R98KAYQV": NewAPNote(
			URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01FE5Y30E3W4P7TRE0R98KAYQV"),
			URLMustParse("https://unknown-instance.com/users/@brand_new_person/01FE5Y30E3W4P7TRE0R98KAYQV"),
			TimeMustParse("2022-07-13T12:13:12+02:00"),
			"Hey @the_mighty_zork@localhost:8080 how's it going?",
			"",
			URLMustParse("https://unknown-instance.com/users/brand_new_person"),
			[]*url.URL{
				ap.PublicURI(),
			},
			[]*url.URL{},
			false,
			[]vocab.ActivityStreamsMention{
				newAPMention(
					URLMustParse("http://localhost:8080/users/the_mighty_zork"),
					"@the_mighty_zork@localhost:8080",
				),
			},
			[]vocab.TootHashtag{},
			nil,
		),
		"https://unknown-instance.com/users/brand_new_person/statuses/01H641QSRS3TCXSVC10X4GPKW7": NewAPNote(
			URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01H641QSRS3TCXSVC10X4GPKW7"),
			URLMustParse("https://unknown-instance.com/users/@brand_new_person/01H641QSRS3TCXSVC10X4GPKW7"),
			TimeMustParse("2023-04-12T12:13:12+02:00"),
			"<p>Babe are you okay, you've hardly touched your <a href=\"https://unknown-instance.com/tags/piss\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>piss</span></a></p>",
			"",
			URLMustParse("https://unknown-instance.com/users/brand_new_person"),
			[]*url.URL{
				ap.PublicURI(),
			},
			[]*url.URL{},
			false,
			[]vocab.ActivityStreamsMention{},
			[]vocab.TootHashtag{
				newAPHashtag(
					URLMustParse("https://unknown-instance.com/tags/piss"),
					"#piss",
				),
			},
			nil,
		),
		"https://turnip.farm/users/turniplover6969/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042": NewAPNote(
			URLMustParse("https://turnip.farm/users/turniplover6969/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042"),
			URLMustParse("https://turnip.farm/@turniplover6969/70c53e54-3146-42d5-a630-83c8b6c7c042"),
			TimeMustParse("2022-07-13T12:13:12+02:00"),
			"",
			"",
			URLMustParse("https://turnip.farm/users/turniplover6969"),
			[]*url.URL{
				ap.PublicURI(),
			},
			[]*url.URL{},
			false,
			nil,
			[]vocab.TootHashtag{},
			[]vocab.ActivityStreamsImage{
				newAPImage(
					URLMustParse("https://turnip.farm/attachments/f17843c7-015e-4251-9b5a-91389c49ee57.jpg"),
					"image/jpeg",
					"",
					"",
				),
			},
		),
		"http://fossbros-anonymous.io/users/foss_satan/statuses/106221634728637552": NewAPNote(
			URLMustParse("http://fossbros-anonymous.io/users/foss_satan/statuses/106221634728637552"),
			URLMustParse("http://fossbros-anonymous.io/@foss_satan/106221634728637552"),
			TimeMustParse("2022-07-13T12:13:12+02:00"),
			`<p><span class="h-card"><a href="http://localhost:8080/@the_mighty_zork" class="u-url mention">@<span>the_mighty_zork</span></a></span> nice there it is:</p><p><a href="http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/activity" rel="nofollow noopener noreferrer" target="_blank"><span class="invisible">https://</span><span class="ellipsis">social.pixie.town/users/f0x/st</span><span class="invisible">atuses/106221628567855262/activity</span></a></p>`,
			"",
			URLMustParse("http://fossbros-anonymous.io/users/foss_satan"),
			[]*url.URL{
				ap.PublicURI(),
			},
			[]*url.URL{},
			false,
			[]vocab.ActivityStreamsMention{
				newAPMention(
					URLMustParse("http://localhost:8080/users/the_mighty_zork"),
					"@the_mighty_zork@localhost:8080",
				),
			},
			[]vocab.TootHashtag{},
			nil,
		),
	}
}

// NewTestBookmarks returns a map of gts model bookmarks, keyed in the format [bookmarking_account]_[target_status]
func NewTestBookmarks() map[string]*gtsmodel.StatusBookmark {
	return map[string]*gtsmodel.StatusBookmark{
		"local_account_1_admin_account_status_1": {
			ID:              "01F8MHD2QCZSZ6WQS2ATVPEYJ9",
			CreatedAt:       TimeMustParse("2022-05-14T13:21:09+02:00"),
			AccountID:       "01F8MH1H7YV1Z7D2C8K2730QBF", // local account 1
			TargetAccountID: "01F8MH17FWEB39HZJ76B6VXSKF", // admin account
			StatusID:        "01F8MH75CBF9JFX4ZAD54N0W0R", // admin account status 1
		},
		"admin_account_local_account_1_status_1": {
			ID:              "01F8Q0486ANTDWKG02A7DS1Q24",
			CreatedAt:       TimeMustParse("2022-05-14T13:21:09+02:00"),
			AccountID:       "01F8MH17FWEB39HZJ76B6VXSKF", // admin account
			TargetAccountID: "01F8MH1H7YV1Z7D2C8K2730QBF", // local account 1
			StatusID:        "01F8MHAMCHF6Y650WCRSCP4WMY", // local account status 1
		},
	}
}

// NewTestDereferenceRequests returns a map of incoming dereference requests, with their signatures.
func NewTestDereferenceRequests(accounts map[string]*gtsmodel.Account) map[string]ActivityWithSignature {
	var sig, digest, date string
	var target *url.URL
	statuses := NewTestStatuses()
	emojis := NewTestEmojis()

	target = URLMustParse(accounts["local_account_1"].URI)
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceZork := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(accounts["local_account_1"].PublicKeyURI)
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceZorkPublicKey := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(statuses["local_account_1_status_1"].URI)
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceLocalAccount1Status1 := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(strings.ToLower(statuses["local_account_1_status_1"].URI))
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceLocalAccount1Status1Lowercase := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(statuses["local_account_1_status_1"].URI + "/replies?only_other_accounts=false")
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

	target = URLMustParse(statuses["local_account_1_status_1"].URI + "/replies?min_id=01FF25D5Q0DH7CHD57CTRS6WK0&only_other_accounts=false")
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

	target = URLMustParse(accounts["local_account_1"].OutboxURI + "?limit=40")
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceZorkOutboxFirst := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(accounts["local_account_1"].OutboxURI + "?limit=40&max_id=01F8MHAMCHF6Y650WCRSCP4WMY")
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceZorkOutboxNext := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	target = URLMustParse(emojis["rainbow"].URI)
	sig, digest, date = GetSignatureForDereference(accounts["remote_account_1"].PublicKeyURI, accounts["remote_account_1"].PrivateKey, target)
	fossSatanDereferenceEmoji := ActivityWithSignature{
		SignatureHeader: sig,
		DigestHeader:    digest,
		DateHeader:      date,
	}

	return map[string]ActivityWithSignature{
		"foss_satan_dereference_zork":                                  fossSatanDereferenceZork,
		"foss_satan_dereference_zork_public_key":                       fossSatanDereferenceZorkPublicKey,
		"foss_satan_dereference_local_account_1_status_1":              fossSatanDereferenceLocalAccount1Status1,
		"foss_satan_dereference_local_account_1_status_1_lowercase":    fossSatanDereferenceLocalAccount1Status1Lowercase,
		"foss_satan_dereference_local_account_1_status_1_replies":      fossSatanDereferenceLocalAccount1Status1Replies,
		"foss_satan_dereference_local_account_1_status_1_replies_next": fossSatanDereferenceLocalAccount1Status1RepliesNext,
		"foss_satan_dereference_local_account_1_status_1_replies_last": fossSatanDereferenceLocalAccount1Status1RepliesLast,
		"foss_satan_dereference_zork_outbox":                           fossSatanDereferenceZorkOutbox,
		"foss_satan_dereference_zork_outbox_first":                     fossSatanDereferenceZorkOutboxFirst,
		"foss_satan_dereference_zork_outbox_next":                      fossSatanDereferenceZorkOutboxNext,
		"foss_satan_dereference_emoji":                                 fossSatanDereferenceEmoji,
	}
}

func NewTestFilters() map[string]*gtsmodel.Filter {
	return map[string]*gtsmodel.Filter{
		"local_account_1_filter_1": {
			ID:            "01HN26VM6KZTW1ANNRVSBMA461",
			CreatedAt:     TimeMustParse("2024-01-25T12:20:03+02:00"),
			UpdatedAt:     TimeMustParse("2024-01-25T12:20:03+02:00"),
			AccountID:     "01F8MH1H7YV1Z7D2C8K2730QBF",
			Title:         "fnord",
			Action:        gtsmodel.FilterActionWarn,
			ContextHome:   util.Ptr(true),
			ContextPublic: util.Ptr(true),
		},
		"local_account_1_filter_2": {
			ID:            "01HN277FSPQAWXZXK92QPPYF79",
			CreatedAt:     TimeMustParse("2024-01-25T12:20:03+02:00"),
			UpdatedAt:     TimeMustParse("2024-01-25T12:20:03+02:00"),
			AccountID:     "01F8MH1H7YV1Z7D2C8K2730QBF",
			Title:         "metasyntactic variables",
			Action:        gtsmodel.FilterActionWarn,
			ContextHome:   util.Ptr(true),
			ContextPublic: util.Ptr(true),
		},
		"local_account_1_filter_3": {
			ID:            "01HWXQDXE4QX4R9EGMG729Y76C",
			CreatedAt:     TimeMustParse("2024-01-25T12:20:03+02:00"),
			UpdatedAt:     TimeMustParse("2024-01-25T12:20:03+02:00"),
			AccountID:     "01F8MH1H7YV1Z7D2C8K2730QBF",
			Title:         "puppies",
			Action:        gtsmodel.FilterActionWarn,
			ContextHome:   util.Ptr(true),
			ContextPublic: util.Ptr(true),
		},
		"local_account_1_filter_4": {
			ID:            "01HZ55WWWP82WYP2A1BKWK8Y9Q",
			CreatedAt:     TimeMustParse("2024-01-25T12:20:03+02:00"),
			UpdatedAt:     TimeMustParse("2024-01-25T12:20:03+02:00"),
			AccountID:     "01F8MH1H7YV1Z7D2C8K2730QBF",
			Title:         "empty filter with no keywords or statuses",
			Action:        gtsmodel.FilterActionWarn,
			ContextHome:   util.Ptr(true),
			ContextPublic: util.Ptr(true),
		},
		"local_account_2_filter_1": {
			ID:            "01HNGFYJBED9FS0VWRVMY4TKXH",
			CreatedAt:     TimeMustParse("2024-01-25T12:20:03+02:00"),
			UpdatedAt:     TimeMustParse("2024-01-25T12:20:03+02:00"),
			AccountID:     "01F8MH1VYJAE00TVVGMM5JNJ8X",
			Title:         "gamer words",
			Action:        gtsmodel.FilterActionWarn,
			ContextHome:   util.Ptr(true),
			ContextPublic: util.Ptr(true),
		},
	}
}

func NewTestFilterKeywords() map[string]*gtsmodel.FilterKeyword {
	return map[string]*gtsmodel.FilterKeyword{
		"local_account_1_filter_1_keyword_1": {
			ID:        "01HN272TAVWAXX72ZX4M8JZ0PS",
			CreatedAt: TimeMustParse("2024-01-25T12:20:03+02:00"),
			UpdatedAt: TimeMustParse("2024-01-25T12:20:03+02:00"),
			AccountID: "01F8MH1H7YV1Z7D2C8K2730QBF",
			FilterID:  "01HN26VM6KZTW1ANNRVSBMA461",
			Keyword:   "fnord",
			WholeWord: util.Ptr(true),
		},
		"local_account_1_filter_2_keyword_1": {
			ID:        "01HN277Y11ENG4EC1ERMAC9FH4",
			CreatedAt: TimeMustParse("2024-01-25T12:20:03+02:00"),
			UpdatedAt: TimeMustParse("2024-01-25T12:20:03+02:00"),
			AccountID: "01F8MH1H7YV1Z7D2C8K2730QBF",
			FilterID:  "01HN277FSPQAWXZXK92QPPYF79",
			Keyword:   "foo",
			WholeWord: util.Ptr(true),
		},
		"local_account_1_filter_2_keyword_2": {
			ID:        "01HN278494N88BA2FY4DZ5JTNS",
			CreatedAt: TimeMustParse("2024-01-25T12:20:03+02:00"),
			UpdatedAt: TimeMustParse("2024-01-25T12:20:03+02:00"),
			AccountID: "01F8MH1H7YV1Z7D2C8K2730QBF",
			FilterID:  "01HN277FSPQAWXZXK92QPPYF79",
			Keyword:   "bar",
			WholeWord: util.Ptr(true),
		},
		"local_account_1_filter_2_keyword_3": {
			ID:        "01HXATJTGYT4BTG2YASE5M7GSD",
			CreatedAt: TimeMustParse("2024-01-25T12:20:03+02:00"),
			UpdatedAt: TimeMustParse("2024-01-25T12:20:03+02:00"),
			AccountID: "01F8MH1H7YV1Z7D2C8K2730QBF",
			FilterID:  "01HN277FSPQAWXZXK92QPPYF79",
			Keyword:   "quux",
			WholeWord: util.Ptr(true),
		},
		"local_account_2_filter_1_keyword_1": {
			ID:        "01HNGG51HV2JT67XQ5MQ7RA1WE",
			CreatedAt: TimeMustParse("2024-01-25T12:20:03+02:00"),
			UpdatedAt: TimeMustParse("2024-01-25T12:20:03+02:00"),
			AccountID: "01F8MH1VYJAE00TVVGMM5JNJ8X",
			FilterID:  "01HNGFYJBED9FS0VWRVMY4TKXH",
			Keyword:   "Virtual Boy",
			WholeWord: util.Ptr(true),
		},
	}
}

func NewTestFilterStatuses() map[string]*gtsmodel.FilterStatus {
	return map[string]*gtsmodel.FilterStatus{
		"local_account_1_filter_3_status_1": {
			ID:        "01HWXQDY8EE182AWQKS45JV50W",
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			AccountID: "01F8MH1H7YV1Z7D2C8K2730QBF",
			FilterID:  "01HWXQDXE4QX4R9EGMG729Y76C",
			StatusID:  "01F8MHAAY43M6RJ473VQFCVH37",
		},
		"local_account_2_filter_1_status_1": {
			ID:        "01HX9WXVEH05E78ABR81FZFFFY",
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			AccountID: "01F8MH1VYJAE00TVVGMM5JNJ8X",
			FilterID:  "01HNGFYJBED9FS0VWRVMY4TKXH",
			StatusID:  "01FVW7JHQFSFK166WWKR8CBA6M",
		},
	}
}

func NewTestUserMutes() map[string]*gtsmodel.UserMute {
	// Not currently used.
	return map[string]*gtsmodel.UserMute{}
}

func NewTestWebPushSubscriptions() map[string]*gtsmodel.WebPushSubscription {
	return map[string]*gtsmodel.WebPushSubscription{
		"local_account_1_token_1": {
			ID:        "01G65Z755AFWAKHE12NY0CQ9FH",
			AccountID: "01F8MH1H7YV1Z7D2C8K2730QBF",
			TokenID:   "01F8MGTQW4DKTDF8SW5CT9HYGA",
			Endpoint:  "https://example.test/push",
			Auth:      "cgna/fzrYLDQyPf5hD7IsA==",
			P256dh:    "BMYVItYVOX+AHBdtA62Q0i6c+F7MV2Gia3aoDr8mvHkuPBNIOuTLDfmFcnBqoZcQk6BtLcIONbxhHpy2R+mYIUY=",
			NotificationFlags: gtsmodel.WebPushSubscriptionNotificationFlagsFromSlice([]gtsmodel.NotificationType{
				gtsmodel.NotificationFollow,
				gtsmodel.NotificationFollowRequest,
				gtsmodel.NotificationFavourite,
				gtsmodel.NotificationMention,
				gtsmodel.NotificationReblog,
				gtsmodel.NotificationPoll,
				gtsmodel.NotificationStatus,
				gtsmodel.NotificationUpdate,
				gtsmodel.NotificationAdminSignup,
				gtsmodel.NotificationAdminReport,
				gtsmodel.NotificationPendingFave,
				gtsmodel.NotificationPendingReply,
				gtsmodel.NotificationPendingReblog,
			}),
			Policy: gtsmodel.WebPushNotificationPolicyFollowed,
		},
	}
}

func NewTestInteractionRequests() map[string]*gtsmodel.InteractionRequest {
	return map[string]*gtsmodel.InteractionRequest{
		"admin_account_reply_turtle": {
			ID:                   "01J5QVXCCEATJYSXM9H6MZT4JR",
			CreatedAt:            TimeMustParse("2024-02-20T12:41:37+02:00"),
			StatusID:             "01F8MHC8VWDRBQR0N1BATDDEM5",
			TargetAccountID:      "01F8MH5NBDF2MV7CTC4Q5128HF",
			InteractingAccountID: "01F8MH17FWEB39HZJ76B6VXSKF",
			InteractionURI:       "http://localhost:8080/users/admin/statuses/01J5QVB9VC76NPPRQ207GG4DRZ",
			InteractionType:      gtsmodel.InteractionReply,
		},
	}
}

func NewTestStatusEdits() map[string]*gtsmodel.StatusEdit {
	return map[string]*gtsmodel.StatusEdit{
		"local_account_1_status_9_edit_1": {
			ID:          "01JDPZCZ2Y9KSGZW0R7ZG8T8Y2",
			Content:     "<p>this is the original status</p>",
			Text:        "this is the original status",
			ContentType: gtsmodel.StatusContentTypeMarkdown,
			Language:    "en",
			Sensitive:   util.Ptr(false),
			StatusID:    "01JDPZC707CKDN8N4QVWM4Z1NR",
			CreatedAt:   TimeMustParse("2024-11-01T11:00:00+02:00"),
		},
		"local_account_1_status_9_edit_2": {
			ID:             "01JDPZDADMD1T9HKF94RECF7PP",
			Content:        "<p>this is the first status edit! now with content-warning</p>",
			ContentWarning: "edited status",
			Text:           "this is the first status edit! now with content-warning",
			ContentType:    gtsmodel.StatusContentTypeMarkdown,
			Language:       "en",
			Sensitive:      util.Ptr(false),
			StatusID:       "01JDPZC707CKDN8N4QVWM4Z1NR",
			CreatedAt:      TimeMustParse("2024-11-01T11:01:00+02:00"),
		},
		"local_account_2_status_9_edit_1": {
			ID:          "01JDPZPBXAX0M02YSEPB21KX4R",
			Content:     "<p>this is the original status</p>",
			Text:        "this is the original status",
			ContentType: gtsmodel.StatusContentTypeMarkdown,
			Language:    "en",
			Sensitive:   util.Ptr(false),
			StatusID:    "01JDPZEZ77X1NX0TY9M10BK1HM",
			CreatedAt:   TimeMustParse("2024-11-01T10:00:00+02:00"),
		},
		"local_account_2_status_9_edit_2": {
			ID:             "01JDPZPJHKP7E3M0YQXEXPS1YT",
			Content:        "<p>now edited to have some media!</p>",
			ContentWarning: "edit with media attachments",
			Text:           "now edited to have some media!",
			ContentType:    gtsmodel.StatusContentTypeMarkdown,
			Language:       "en",
			Sensitive:      util.Ptr(true),
			AttachmentIDs:  []string{"01JDQ164HM08SGJ7ZEK9003Z4B"},
			StatusID:       "01JDPZEZ77X1NX0TY9M10BK1HM",
			CreatedAt:      TimeMustParse("2024-11-01T10:01:00+02:00"),
		},
		"local_account_2_status_9_edit_3": {
			ID:             "01JDPZPY3F85Y7B78ETRXEMWD9",
			Content:        "<p>now edited to remove the media</p>",
			ContentWarning: "edit missing previous media attachments",
			Text:           "now edited to remove the media",
			ContentType:    gtsmodel.StatusContentTypeMarkdown,
			Language:       "en",
			Sensitive:      util.Ptr(false),
			StatusID:       "01JDPZEZ77X1NX0TY9M10BK1HM",
			CreatedAt:      TimeMustParse("2024-11-01T10:02:00+02:00"),
		},
		"remote_account_1_status_4_edit_1": {
			ID:          "01JDQ07ZZ4FGP13YN8TF63P5A6",
			Content:     "<p>this is the original status, with a poll!</p>",
			Text:        "this is the original status, with a poll!",
			ContentType: gtsmodel.StatusContentTypeMarkdown,
			Language:    "en",
			Sensitive:   util.Ptr(false),
			PollOptions: []string{"yes", "no", "spiderman"},
			PollVotes:   []int{42, 42, 69},
			StatusID:    "01JDQ07JZTX9CMDJP67CNA71YD",
			CreatedAt:   TimeMustParse("2024-11-01T09:00:00+02:00"),
		},
		"remote_account_1_status_4_edit_2": {
			ID:             "01JDQ08AYQC0G6413VAHA51CV9",
			Content:        "<p>this is the first status edit! now with a different poll!</p>",
			ContentWarning: "edited status",
			Text:           "this is the first status edit! now with a different poll!",
			ContentType:    gtsmodel.StatusContentTypeMarkdown,
			Language:       "en",
			Sensitive:      util.Ptr(false),
			PollOptions:    []string{"yes", "no", "maybe", "i don't know", "can you repeat the question"},
			PollVotes:      []int{0, 0, 0, 0, 1},
			StatusID:       "01JDQ07JZTX9CMDJP67CNA71YD",
			CreatedAt:      TimeMustParse("2024-11-01T09:01:00+02:00"),
		},
	}
}

// GetSignatureForActivity prepares a mock HTTP request as if it were going to deliver activity to destination signed for privkey and pubKeyID, signs the request and returns the header values.
func GetSignatureForActivity(activity pub.Activity, pubKeyID string, privkey *rsa.PrivateKey, destination *url.URL) (signatureHeader string, digestHeader string, dateHeader string) {
	// convert the activity into json bytes
	m, err := activity.Serialize()
	if err != nil {
		panic(err)
	}
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	// Prepare HTTP request signer
	sig, err := transport.NewPOSTSigner(120)
	if err != nil {
		panic(err)
	}

	// Prepare a mock request ready for signing
	r, err := http.NewRequest("POST", destination.String(), bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	r.Header.Set("Host", destination.Host)
	r.Header.Set("Date", time.Now().Format("Mon, 02 Jan 2006 15:04:05")+" GMT")

	// Sign this new HTTP request
	if err := sig.SignRequest(privkey, pubKeyID, r, b); err != nil {
		panic(err)
	}

	// Load signed data from request
	signatureHeader = r.Header.Get("Signature")
	digestHeader = r.Header.Get("Digest")
	dateHeader = r.Header.Get("Date")

	// headers should now be populated
	return
}

// GetSignatureForDereference prepares a mock HTTP request as if it were going to dereference destination signed for privkey and pubKeyID, signs the request and returns the header values.
func GetSignatureForDereference(pubKeyID string, privkey *rsa.PrivateKey, destination *url.URL) (signatureHeader string, digestHeader string, dateHeader string) {
	// Prepare HTTP request signer
	sig, err := transport.NewGETSigner(120)
	if err != nil {
		panic(err)
	}

	// Prepare a mock request ready for signing
	r, err := http.NewRequest("GET", destination.String(), nil)
	if err != nil {
		panic(err)
	}
	r.Header.Set("Host", destination.Host)
	r.Header.Set("Date", time.Now().Format("Mon, 02 Jan 2006 15:04:05")+" GMT")

	// Sign this new HTTP request
	if err := sig.SignRequest(privkey, pubKeyID, r, nil); err != nil {
		panic(err)
	}

	// Load signed data from request
	signatureHeader = r.Header.Get("Signature")
	digestHeader = r.Header.Get("Digest")
	dateHeader = r.Header.Get("Date")

	// headers should now be populated
	return
}

func newAPPerson(
	profileIDURI *url.URL,
	followingURI *url.URL,
	followersURI *url.URL,
	inboxURI *url.URL,
	sharedInboxIRI *url.URL,
	outboxURI *url.URL,
	featuredURI *url.URL,
	movedToURI *url.URL,
	alsoKnownAsURIs []*url.URL,
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
	manuallyApprovesFollowers bool,
) vocab.ActivityStreamsPerson {
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

	// shared inbox
	if sharedInboxIRI != nil {
		endpointsProp := streams.NewActivityStreamsEndpointsProperty()
		endpoints := streams.NewActivityStreamsEndpoints()
		sharedInboxProp := streams.NewActivityStreamsSharedInboxProperty()
		sharedInboxProp.SetIRI(sharedInboxIRI)
		endpoints.SetActivityStreamsSharedInbox(sharedInboxProp)
		endpointsProp.AppendActivityStreamsEndpoints(endpoints)
		person.SetActivityStreamsEndpoints(endpointsProp)
	}

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

	// alsoKnownAs, movedTo
	// Required for Move activity.
	if len(alsoKnownAsURIs) != 0 {
		ap.SetAlsoKnownAs(person, alsoKnownAsURIs)
	}

	if movedToURI != nil {
		ap.SetMovedTo(person, movedToURI)
	}

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

func newAPGroup(
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
	manuallyApprovesFollowers bool,
) vocab.ActivityStreamsGroup {
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

	// AlsoKnownAsURI
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

func newAPService(
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
	manuallyApprovesFollowers bool,
) vocab.ActivityStreamsService {
	service := streams.NewActivityStreamsService()

	// id should be the activitypub URI of this group
	// something like https://example.org/users/example_group
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(profileIDURI)
	service.SetJSONLDId(idProp)

	// following
	// The URI for retrieving a list of accounts this group is following
	followingProp := streams.NewActivityStreamsFollowingProperty()
	followingProp.SetIRI(followingURI)
	service.SetActivityStreamsFollowing(followingProp)

	// followers
	// The URI for retrieving a list of this user's followers
	followersProp := streams.NewActivityStreamsFollowersProperty()
	followersProp.SetIRI(followersURI)
	service.SetActivityStreamsFollowers(followersProp)

	// inbox
	// the activitypub inbox of this user for accepting messages
	inboxProp := streams.NewActivityStreamsInboxProperty()
	inboxProp.SetIRI(inboxURI)
	service.SetActivityStreamsInbox(inboxProp)

	// outbox
	// the activitypub outbox of this user for serving messages
	outboxProp := streams.NewActivityStreamsOutboxProperty()
	outboxProp.SetIRI(outboxURI)
	service.SetActivityStreamsOutbox(outboxProp)

	// featured posts
	// Pinned posts.
	featuredProp := streams.NewTootFeaturedProperty()
	featuredProp.SetIRI(featuredURI)
	service.SetTootFeatured(featuredProp)

	// featuredTags
	// NOT IMPLEMENTED

	// preferredUsername
	// Used for Webfinger lookup. Must be unique on the domain, and must correspond to a Webfinger acct: URI.
	preferredUsernameProp := streams.NewActivityStreamsPreferredUsernameProperty()
	preferredUsernameProp.SetXMLSchemaString(username)
	service.SetActivityStreamsPreferredUsername(preferredUsernameProp)

	// name
	// Used as profile display name.
	nameProp := streams.NewActivityStreamsNameProperty()
	if displayName != "" {
		nameProp.AppendXMLSchemaString(displayName)
	} else {
		nameProp.AppendXMLSchemaString(username)
	}
	service.SetActivityStreamsName(nameProp)

	// summary
	// Used as profile bio.
	if note != "" {
		summaryProp := streams.NewActivityStreamsSummaryProperty()
		summaryProp.AppendXMLSchemaString(note)
		service.SetActivityStreamsSummary(summaryProp)
	}

	// url
	// Used as profile link.
	urlProp := streams.NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(profileURL)
	service.SetActivityStreamsUrl(urlProp)

	// manuallyApprovesFollowers
	manuallyApprovesFollowersProp := streams.NewActivityStreamsManuallyApprovesFollowersProperty()
	manuallyApprovesFollowersProp.Set(manuallyApprovesFollowers)
	service.SetActivityStreamsManuallyApprovesFollowers(manuallyApprovesFollowersProp)

	// discoverable
	// Will be shown in the profile directory.
	discoverableProp := streams.NewTootDiscoverableProperty()
	discoverableProp.Set(discoverable)
	service.SetTootDiscoverable(discoverableProp)

	// devices
	// NOT IMPLEMENTED, probably won't implement

	// AlsoKnownAsURI
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
	service.SetW3IDSecurityV1PublicKey(publicKeyProp)

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
	service.SetActivityStreamsIcon(iconProperty)

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
	service.SetActivityStreamsImage(headerProperty)

	return service
}

func newAPMention(uri *url.URL, namestring string) vocab.ActivityStreamsMention {
	mention := streams.NewActivityStreamsMention()

	hrefProp := streams.NewActivityStreamsHrefProperty()
	hrefProp.SetIRI(uri)
	mention.SetActivityStreamsHref(hrefProp)

	nameProp := streams.NewActivityStreamsNameProperty()
	nameProp.AppendXMLSchemaString(namestring)
	mention.SetActivityStreamsName(nameProp)

	return mention
}

func newAPHashtag(href *url.URL, name string) vocab.TootHashtag {
	hashtag := streams.NewTootHashtag()

	hrefProp := streams.NewActivityStreamsHrefProperty()
	hrefProp.SetIRI(href)
	hashtag.SetActivityStreamsHref(hrefProp)

	nameProp := streams.NewActivityStreamsNameProperty()
	nameProp.AppendXMLSchemaString(name)
	hashtag.SetActivityStreamsName(nameProp)

	return hashtag
}

func newAPImage(url *url.URL, mediaType string, imageDescription string, blurhash string) vocab.ActivityStreamsImage {
	image := streams.NewActivityStreamsImage()

	if url != nil {
		urlProp := streams.NewActivityStreamsUrlProperty()
		urlProp.AppendIRI(url)
		image.SetActivityStreamsUrl(urlProp)
	}

	if mediaType != "" {
		mediaTypeProp := streams.NewActivityStreamsMediaTypeProperty()
		mediaTypeProp.Set(mediaType)
		image.SetActivityStreamsMediaType(mediaTypeProp)
	}

	if imageDescription != "" {
		nameProp := streams.NewActivityStreamsNameProperty()
		nameProp.AppendXMLSchemaString(imageDescription)
		image.SetActivityStreamsName(nameProp)
	}

	if blurhash != "" {
		blurhashProp := streams.NewTootBlurhashProperty()
		blurhashProp.Set(blurhash)
		image.SetTootBlurhash(blurhashProp)
	}

	return image
}

func newAPEmoji(id *url.URL, name string, updated time.Time, image vocab.ActivityStreamsImage) vocab.TootEmoji {
	emoji := streams.NewTootEmoji()

	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(id)
	emoji.SetJSONLDId(idProp)

	nameProp := streams.NewActivityStreamsNameProperty()
	nameProp.AppendXMLSchemaString(`:` + strings.Trim(name, ":") + `:`)
	emoji.SetActivityStreamsName(nameProp)

	updatedProp := streams.NewActivityStreamsUpdatedProperty()
	updatedProp.Set(updated)
	emoji.SetActivityStreamsUpdated(updatedProp)

	iconProp := streams.NewActivityStreamsIconProperty()
	iconProp.AppendActivityStreamsImage(image)
	emoji.SetActivityStreamsIcon(iconProp)

	return emoji
}

// NewAPNote returns a new activity streams note for the given parameters
func NewAPNote(
	noteID *url.URL,
	noteURL *url.URL,
	noteCreatedAt time.Time,
	noteContent string,
	noteSummary string,
	noteAttributedTo *url.URL,
	noteTo []*url.URL,
	noteCC []*url.URL,
	noteSensitive bool,
	noteMentions []vocab.ActivityStreamsMention,
	noteTags []vocab.TootHashtag,
	noteAttachments []vocab.ActivityStreamsImage,
) vocab.ActivityStreamsNote {
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

	// Tag entries
	tag := streams.NewActivityStreamsTagProperty()

	// mentions
	for _, m := range noteMentions {
		tag.AppendActivityStreamsMention(m)
	}
	note.SetActivityStreamsTag(tag)

	// hashtags
	for _, t := range noteTags {
		tag.AppendTootHashtag(t)
	}

	// append any attachments as ActivityStreamsImage
	if noteAttachments != nil {
		attachmentProperty := streams.NewActivityStreamsAttachmentProperty()
		for _, a := range noteAttachments {
			attachmentProperty.AppendActivityStreamsImage(a)
		}
		note.SetActivityStreamsAttachment(attachmentProperty)
	}

	return note
}

// WrapAPNoteInCreate wraps the given activity streams note in a Create activity streams action
func WrapAPNoteInCreate(createID *url.URL, createActor *url.URL, createPublished time.Time, createNote vocab.ActivityStreamsNote) vocab.ActivityStreamsCreate {
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

func newAPAnnounce(announceID *url.URL, announceActor *url.URL, announcePublished time.Time, announceTo *url.URL, announceNote vocab.ActivityStreamsNote) vocab.ActivityStreamsAnnounce {
	announce := streams.NewActivityStreamsAnnounce()

	if announceID != nil {
		id := streams.NewJSONLDIdProperty()
		id.Set(announceID)
		announce.SetJSONLDId(id)
	}

	if announceActor != nil {
		actor := streams.NewActivityStreamsActorProperty()
		actor.AppendIRI(announceActor)
		announce.SetActivityStreamsActor(actor)
	}

	if !announcePublished.IsZero() {
		published := streams.NewActivityStreamsPublishedProperty()
		published.Set(announcePublished)
		announce.SetActivityStreamsPublished(published)
	}

	to := streams.NewActivityStreamsToProperty()
	to.AppendIRI(announceTo)
	announce.SetActivityStreamsTo(announceNote.GetActivityStreamsTo())

	cc := streams.NewActivityStreamsCcProperty()
	cc.AppendIRI(announceNote.GetActivityStreamsAttributedTo().Begin().GetIRI())
	announce.SetActivityStreamsCc(cc)

	if announceNote != nil {
		noteIRI := streams.NewActivityStreamsObjectProperty()
		noteIRI.AppendIRI(announceNote.GetJSONLDId().Get())
		announce.SetActivityStreamsObject(noteIRI)
	}

	return announce
}

func newAPDelete(deleteTarget *url.URL, deleteActor *url.URL, deletePublished time.Time, deleteTo *url.URL) vocab.ActivityStreamsDelete {
	delete := streams.NewActivityStreamsDelete()

	objectProp := streams.NewActivityStreamsObjectProperty()
	objectProp.AppendIRI(deleteTarget)
	delete.SetActivityStreamsObject(objectProp)

	to := streams.NewActivityStreamsToProperty()
	to.AppendIRI(deleteTo)
	delete.SetActivityStreamsTo(to)

	actor := streams.NewActivityStreamsActorProperty()
	actor.AppendIRI(deleteActor)
	delete.SetActivityStreamsActor(actor)

	published := streams.NewActivityStreamsPublishedProperty()
	published.Set(deletePublished)
	delete.SetActivityStreamsPublished(published)

	return delete
}
