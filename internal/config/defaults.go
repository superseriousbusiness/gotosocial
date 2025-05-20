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

package config

import (
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/language"
	"codeberg.org/gruf/go-bytesize"
	"github.com/coreos/go-oidc/v3/oidc"
)

// Defaults contains a populated Configuration with reasonable defaults. Note that
// if you use this, you will still need to set Host, and, if desired, ConfigPath.
var Defaults = Configuration{
	LogLevel:           "info",
	LogTimestampFormat: "02/01/2006 15:04:05.000",
	LogDbQueries:       false,
	ApplicationName:    "gotosocial",
	LandingPageUser:    "",
	ConfigPath:         "",
	Host:               "",
	AccountDomain:      "",
	Protocol:           "https",
	BindAddress:        "0.0.0.0",
	Port:               8080,
	TrustedProxies:     []string{"127.0.0.1/32", "::1"}, // localhost

	DbType:                   "sqlite",
	DbAddress:                "db.sqlite",
	DbPort:                   5432,
	DbUser:                   "",
	DbPassword:               "",
	DbDatabase:               "gotosocial",
	DbTLSMode:                "disable",
	DbTLSCACert:              "",
	DbMaxOpenConnsMultiplier: 8,
	DbSqliteJournalMode:      "WAL",
	DbSqliteSynchronous:      "NORMAL",
	DbSqliteCacheSize:        8 * bytesize.MiB,
	DbSqliteBusyTimeout:      time.Minute * 30,

	WebTemplateBaseDir: "./web/template/",
	WebAssetBaseDir:    "./web/assets/",

	InstanceFederationMode:            InstanceFederationModeDefault,
	InstanceFederationSpamFilter:      false,
	InstanceExposePeers:               false,
	InstanceExposeBlocklist:           false,
	InstanceExposeBlocklistWeb:        false,
	InstanceDeliverToSharedInboxes:    true,
	InstanceLanguages:                 make(language.Languages, 0),
	InstanceSubscriptionsProcessFrom:  "23:00",        // 11pm,
	InstanceSubscriptionsProcessEvery: 24 * time.Hour, // 1/day.
	InstanceAllowBackdatingStatuses:   true,

	AccountsRegistrationOpen:         false,
	AccountsReasonRequired:           true,
	AccountsRegistrationDailyLimit:   10,
	AccountsRegistrationBacklogLimit: 20,
	AccountsAllowCustomCSS:           false,
	AccountsCustomCSSLength:          10000,
	AccountsMaxProfileFields:         6,

	MediaDescriptionMinChars: 0,
	MediaDescriptionMaxChars: 1500,
	MediaRemoteCacheDays:     7,
	MediaLocalMaxSize:        40 * bytesize.MiB,
	MediaRemoteMaxSize:       40 * bytesize.MiB,
	MediaEmojiLocalMaxSize:   50 * bytesize.KiB,
	MediaEmojiRemoteMaxSize:  100 * bytesize.KiB,
	MediaCleanupFrom:         "00:00",        // Midnight.
	MediaCleanupEvery:        24 * time.Hour, // 1/day.
	MediaFfmpegPoolSize:      1,

	StorageBackend:        "local",
	StorageLocalBasePath:  "/gotosocial/storage",
	StorageS3UseSSL:       true,
	StorageS3Proxy:        false,
	StorageS3RedirectURL:  "",
	StorageS3BucketLookup: "auto",

	StatusesMaxChars:           5000,
	StatusesPollMaxOptions:     6,
	StatusesPollOptionMaxChars: 50,
	StatusesMediaMaxFiles:      6,

	LetsEncryptEnabled:      false,
	LetsEncryptPort:         80,
	LetsEncryptCertDir:      "/gotosocial/storage/certs",
	LetsEncryptEmailAddress: "",

	TLSCertificateChain: "",
	TLSCertificateKey:   "",

	OIDCEnabled:          false,
	OIDCIdpName:          "",
	OIDCSkipVerification: false,
	OIDCIssuer:           "",
	OIDCClientID:         "",
	OIDCClientSecret:     "",
	OIDCScopes:           []string{oidc.ScopeOpenID, "profile", "email", "groups"},
	OIDCLinkExisting:     false,

	SMTPHost:               "",
	SMTPPort:               0,
	SMTPUsername:           "",
	SMTPPassword:           "",
	SMTPFrom:               "",
	SMTPDiscloseRecipients: false,

	TracingEnabled: false,
	MetricsEnabled: false,

	SyslogEnabled:  false,
	SyslogProtocol: "udp",
	SyslogAddress:  "localhost:514",

	Advanced: AdvancedConfig{
		SenderMultiplier: 2, // 2 senders per CPU
		CSPExtraURIs:     []string{},
		HeaderFilterMode: RequestHeaderFilterModeDisabled,
		CookiesSamesite:  "lax",

		RateLimit: RateLimitConfig{
			Requests:   300, // 1 per second per 5 minutes
			Exceptions: IPPrefixes{},
		},

		Throttling: ThrottlingConfig{
			Multiplier: 8, // 8 open requests per CPU
			RetryAfter: 30 * time.Second,
		},

		ScraperDeterrence: ScraperDeterrenceConfig{
			Enabled:    false,
			Difficulty: 4,
		},
	},

	Cache: CacheConfiguration{
		// Rough memory target that the total
		// size of all State.Caches will attempt
		// to remain with. Emphasis on *rough*.
		MemoryTarget: 100 * bytesize.MiB,

		// These ratios signal what percentage
		// of the available cache target memory
		// is allocated to each object type's
		// cache.
		//
		// These are weighted by a totally
		// assorted mixture of priority, and
		// manual twiddling to get the generated
		// cache capacity ratios within normal
		// amounts dependent size of the models.
		//
		// when TODO items in the size.go source
		// file have been addressed, these should
		// be able to make some more sense :D
		AccountMemRatio:                       5,
		AccountNoteMemRatio:                   1,
		AccountSettingsMemRatio:               0.1,
		AccountStatsMemRatio:                  2,
		ApplicationMemRatio:                   0.1,
		BlockMemRatio:                         2,
		BlockIDsMemRatio:                      3,
		BoostOfIDsMemRatio:                    3,
		ClientMemRatio:                        0.1,
		ConversationMemRatio:                  1,
		ConversationLastStatusIDsMemRatio:     2,
		DomainPermissionDraftMemRation:        0.5,
		DomainPermissionSubscriptionMemRation: 0.5,
		EmojiMemRatio:                         3,
		EmojiCategoryMemRatio:                 0.1,
		FilterMemRatio:                        0.5,
		FilterKeywordMemRatio:                 0.5,
		FilterStatusMemRatio:                  0.5,
		FollowMemRatio:                        2,
		FollowIDsMemRatio:                     4,
		FollowRequestMemRatio:                 2,
		FollowRequestIDsMemRatio:              2,
		FollowingTagIDsMemRatio:               2,
		InReplyToIDsMemRatio:                  3,
		InstanceMemRatio:                      1,
		InteractionRequestMemRatio:            1,
		ListMemRatio:                          1,
		ListIDsMemRatio:                       2,
		ListedIDsMemRatio:                     2,
		MarkerMemRatio:                        0.5,
		MediaMemRatio:                         4,
		MentionMemRatio:                       2,
		MoveMemRatio:                          0.1,
		NotificationMemRatio:                  2,
		PollMemRatio:                          1,
		PollVoteMemRatio:                      2,
		PollVoteIDsMemRatio:                   2,
		ReportMemRatio:                        1,
		SinBinStatusMemRatio:                  0.5,
		StatusMemRatio:                        5,
		StatusBookmarkMemRatio:                0.5,
		StatusBookmarkIDsMemRatio:             2,
		StatusEditMemRatio:                    2,
		StatusFaveMemRatio:                    2,
		StatusFaveIDsMemRatio:                 3,
		TagMemRatio:                           2,
		ThreadMuteMemRatio:                    0.2,
		TokenMemRatio:                         0.75,
		TombstoneMemRatio:                     0.5,
		UserMemRatio:                          0.25,
		UserMuteMemRatio:                      2,
		UserMuteIDsMemRatio:                   3,
		WebfingerMemRatio:                     0.1,
		WebPushSubscriptionMemRatio:           1,
		WebPushSubscriptionIDsMemRatio:        1,
		VisibilityMemRatio:                    2,
	},

	HTTPClient: HTTPClientConfiguration{
		AllowIPs:              make([]string, 0),
		BlockIPs:              make([]string, 0),
		Timeout:               30 * time.Second,
		TLSInsecureSkipVerify: false,
	},

	AdminMediaPruneDryRun: true,

	RequestIDHeader: "X-Request-Id",

	LogClientIP: true,
}
