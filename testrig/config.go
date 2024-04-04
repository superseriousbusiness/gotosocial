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
	"os"
	"time"

	"codeberg.org/gruf/go-bytesize"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/language"
)

// InitTestConfig initializes viper configuration with test defaults.
func InitTestConfig() {
	config.Config(func(cfg *config.Configuration) {
		*cfg = testDefaults
	})
}

func logLevel() string {
	level := "error"
	if lv := os.Getenv("GTS_LOG_LEVEL"); lv != "" {
		level = lv
	}
	return level
}

var testDefaults = config.Configuration{
	LogLevel:           logLevel(),
	LogTimestampFormat: "02/01/2006 15:04:05.000",
	LogDbQueries:       true,
	ApplicationName:    "gotosocial",
	LandingPageUser:    "",
	ConfigPath:         "",
	Host:               "localhost:8080",
	AccountDomain:      "localhost:8080",
	Protocol:           "http",
	BindAddress:        "127.0.0.1",
	Port:               8080,
	TrustedProxies:     []string{"127.0.0.1/32", "::1"},

	DbType:                   "sqlite",
	DbAddress:                ":memory:",
	DbPort:                   5432,
	DbUser:                   "postgres",
	DbPassword:               "postgres",
	DbDatabase:               "postgres",
	DbTLSMode:                "disable",
	DbTLSCACert:              "",
	DbMaxOpenConnsMultiplier: 8,
	DbSqliteJournalMode:      "WAL",
	DbSqliteSynchronous:      "NORMAL",
	DbSqliteCacheSize:        8 * bytesize.MiB,
	DbSqliteBusyTimeout:      time.Minute * 5,

	WebTemplateBaseDir: "./web/template/",
	WebAssetBaseDir:    "./web/assets/",

	InstanceFederationMode:         config.InstanceFederationModeDefault,
	InstanceFederationSpamFilter:   true,
	InstanceExposePeers:            true,
	InstanceExposeSuspended:        true,
	InstanceExposeSuspendedWeb:     true,
	InstanceDeliverToSharedInboxes: true,
	InstanceLanguages: language.Languages{
		{
			TagStr: "nl",
		},
		{
			TagStr: "en-gb",
		},
	},

	AccountsRegistrationOpen: true,
	AccountsApprovalRequired: true,
	AccountsReasonRequired:   true,
	AccountsAllowCustomCSS:   true,
	AccountsCustomCSSLength:  10000,

	MediaImageMaxSize:        10485760, // 10MiB
	MediaVideoMaxSize:        41943040, // 40MiB
	MediaDescriptionMinChars: 0,
	MediaDescriptionMaxChars: 500,
	MediaRemoteCacheDays:     7,
	MediaEmojiLocalMaxSize:   51200,          // 50KiB
	MediaEmojiRemoteMaxSize:  102400,         // 100KiB
	MediaCleanupFrom:         "00:00",        // midnight.
	MediaCleanupEvery:        24 * time.Hour, // 1/day.

	// the testrig only uses in-memory storage, so we can
	// safely set this value to 'test' to avoid running storage
	// migrations, and other silly things like that
	StorageBackend:       "test",
	StorageLocalBasePath: "",

	StatusesMaxChars:           5000,
	StatusesPollMaxOptions:     6,
	StatusesPollOptionMaxChars: 50,
	StatusesMediaMaxFiles:      6,

	LetsEncryptEnabled:      false,
	LetsEncryptPort:         0,
	LetsEncryptCertDir:      "",
	LetsEncryptEmailAddress: "",

	OIDCEnabled:          false,
	OIDCIdpName:          "",
	OIDCSkipVerification: false,
	OIDCIssuer:           "",
	OIDCClientID:         "",
	OIDCClientSecret:     "",
	OIDCScopes:           []string{oidc.ScopeOpenID, "profile", "email", "groups"},
	OIDCLinkExisting:     false,
	OIDCAdminGroups:      []string{"adminRole"},
	OIDCAllowedGroups:    []string{"allowedRole"},

	SMTPHost:               "",
	SMTPPort:               0,
	SMTPUsername:           "",
	SMTPPassword:           "",
	SMTPFrom:               "GoToSocial",
	SMTPDiscloseRecipients: false,

	TracingEnabled:           false,
	TracingEndpoint:          "localhost:4317",
	TracingTransport:         "grpc",
	TracingInsecureTransport: true,

	MetricsEnabled:     false,
	MetricsAuthEnabled: false,

	SyslogEnabled:  false,
	SyslogProtocol: "udp",
	SyslogAddress:  "localhost:514",

	AdvancedCookiesSamesite:      "lax",
	AdvancedRateLimitRequests:    0, // disabled
	AdvancedThrottlingMultiplier: 0, // disabled
	AdvancedSenderMultiplier:     0, // 1 sender only, regardless of CPU

	SoftwareVersion: "0.0.0-testrig",

	// simply use cache defaults.
	Cache: config.Defaults.Cache,
}
