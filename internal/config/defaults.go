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

	"codeberg.org/gruf/go-bytesize"
	"github.com/coreos/go-oidc/v3/oidc"
)

// Defaults contains a populated Configuration with reasonable defaults. Note that
// if you use this, you will still need to set Host, and, if desired, ConfigPath.
var Defaults = Configuration{
	LogLevel:        "info",
	LogDbQueries:    false,
	ApplicationName: "gotosocial",
	LandingPageUser: "",
	ConfigPath:      "",
	Host:            "",
	AccountDomain:   "",
	Protocol:        "https",
	BindAddress:     "0.0.0.0",
	Port:            8080,
	TrustedProxies:  []string{"127.0.0.1/32", "::1"}, // localhost

	DbType:                   "postgres",
	DbAddress:                "",
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

	InstanceExposePeers:            false,
	InstanceExposeSuspended:        false,
	InstanceExposeSuspendedWeb:     false,
	InstanceDeliverToSharedInboxes: true,

	AccountsRegistrationOpen: true,
	AccountsApprovalRequired: true,
	AccountsReasonRequired:   true,
	AccountsAllowCustomCSS:   false,
	AccountsCustomCSSLength:  10000,

	MediaImageMaxSize:        10 * bytesize.MiB,
	MediaVideoMaxSize:        40 * bytesize.MiB,
	MediaDescriptionMinChars: 0,
	MediaDescriptionMaxChars: 500,
	MediaRemoteCacheDays:     30,
	MediaEmojiLocalMaxSize:   50 * bytesize.KiB,
	MediaEmojiRemoteMaxSize:  100 * bytesize.KiB,

	StorageBackend:       "local",
	StorageLocalBasePath: "/gotosocial/storage",
	StorageS3UseSSL:      true,
	StorageS3Proxy:       false,

	StatusesMaxChars:           5000,
	StatusesCWMaxChars:         100,
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
	SMTPFrom:               "GoToSocial",
	SMTPDiscloseRecipients: false,

	TracingEnabled:           false,
	TracingTransport:         "grpc",
	TracingEndpoint:          "",
	TracingInsecureTransport: false,

	SyslogEnabled:  false,
	SyslogProtocol: "udp",
	SyslogAddress:  "localhost:514",

	AdvancedCookiesSamesite:      "lax",
	AdvancedRateLimitRequests:    300, // 1 per second per 5 minutes
	AdvancedThrottlingMultiplier: 8,   // 8 open requests per CPU
	AdvancedSenderMultiplier:     2,   // 2 senders per CPU

	Cache: CacheConfiguration{
		GTS: GTSCacheConfiguration{
			AccountMaxSize:   2000,
			AccountTTL:       time.Minute * 30,
			AccountSweepFreq: time.Minute,

			BlockMaxSize:   1000,
			BlockTTL:       time.Minute * 30,
			BlockSweepFreq: time.Minute,

			DomainBlockMaxSize:   2000,
			DomainBlockTTL:       time.Hour * 24,
			DomainBlockSweepFreq: time.Minute,

			EmojiMaxSize:   2000,
			EmojiTTL:       time.Minute * 30,
			EmojiSweepFreq: time.Minute,

			EmojiCategoryMaxSize:   100,
			EmojiCategoryTTL:       time.Minute * 30,
			EmojiCategorySweepFreq: time.Minute,

			FollowMaxSize:   2000,
			FollowTTL:       time.Minute * 30,
			FollowSweepFreq: time.Minute,

			FollowRequestMaxSize:   2000,
			FollowRequestTTL:       time.Minute * 30,
			FollowRequestSweepFreq: time.Minute,

			InstanceMaxSize:   2000,
			InstanceTTL:       time.Minute * 30,
			InstanceSweepFreq: time.Minute,

			ListMaxSize:   2000,
			ListTTL:       time.Minute * 30,
			ListSweepFreq: time.Minute,

			ListEntryMaxSize:   2000,
			ListEntryTTL:       time.Minute * 30,
			ListEntrySweepFreq: time.Minute,

			MediaMaxSize:   1000,
			MediaTTL:       time.Minute * 30,
			MediaSweepFreq: time.Minute,

			MentionMaxSize:   2000,
			MentionTTL:       time.Minute * 30,
			MentionSweepFreq: time.Minute,

			NotificationMaxSize:   1000,
			NotificationTTL:       time.Minute * 30,
			NotificationSweepFreq: time.Minute,

			ReportMaxSize:   100,
			ReportTTL:       time.Minute * 30,
			ReportSweepFreq: time.Minute,

			StatusMaxSize:   2000,
			StatusTTL:       time.Minute * 30,
			StatusSweepFreq: time.Minute,

			StatusFaveMaxSize:   2000,
			StatusFaveTTL:       time.Minute * 30,
			StatusFaveSweepFreq: time.Minute,

			TombstoneMaxSize:   500,
			TombstoneTTL:       time.Minute * 30,
			TombstoneSweepFreq: time.Minute,

			UserMaxSize:   500,
			UserTTL:       time.Minute * 30,
			UserSweepFreq: time.Minute,

			WebfingerMaxSize:   250,
			WebfingerTTL:       time.Hour * 24,
			WebfingerSweepFreq: time.Minute * 15,
		},

		VisibilityMaxSize:   2000,
		VisibilityTTL:       time.Minute * 30,
		VisibilitySweepFreq: time.Minute,
	},

	AdminMediaPruneDryRun: true,

	RequestIDHeader: "X-Request-Id",

	LogClientIP: true,
}
