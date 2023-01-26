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

package testrig

import (
	"time"

	"codeberg.org/gruf/go-bytesize"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// InitTestConfig initializes viper configuration with test defaults.
func InitTestConfig() {
	config.Config(func(cfg *config.Configuration) {
		*cfg = testDefaults
	})
}

var testDefaults = config.Configuration{
	LogLevel:        "trace",
	LogDbQueries:    true,
	ApplicationName: "gotosocial",
	LandingPageUser: "",
	ConfigPath:      "",
	Host:            "localhost:8080",
	AccountDomain:   "localhost:8080",
	Protocol:        "http",
	BindAddress:     "127.0.0.1",
	Port:            8080,
	TrustedProxies:  []string{"127.0.0.1/32", "::1"},

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

	InstanceExposePeers:            true,
	InstanceExposeSuspended:        true,
	InstanceExposeSuspendedWeb:     true,
	InstanceDeliverToSharedInboxes: true,

	AccountsRegistrationOpen: true,
	AccountsApprovalRequired: true,
	AccountsReasonRequired:   true,
	AccountsAllowCustomCSS:   true,

	MediaImageMaxSize:        10485760, // 10mb
	MediaVideoMaxSize:        41943040, // 40mb
	MediaDescriptionMinChars: 0,
	MediaDescriptionMaxChars: 500,
	MediaRemoteCacheDays:     30,
	MediaEmojiLocalMaxSize:   51200,  // 50kb
	MediaEmojiRemoteMaxSize:  102400, // 100kb

	// the testrig only uses in-memory storage, so we can
	// safely set this value to 'test' to avoid running storage
	// migrations, and other silly things like that
	StorageBackend:       "test",
	StorageLocalBasePath: "",

	StatusesMaxChars:           5000,
	StatusesCWMaxChars:         100,
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

	SMTPHost:     "",
	SMTPPort:     0,
	SMTPUsername: "",
	SMTPPassword: "",
	SMTPFrom:     "GoToSocial",

	SyslogEnabled:  false,
	SyslogProtocol: "udp",
	SyslogAddress:  "localhost:514",

	AdvancedCookiesSamesite:      "lax",
	AdvancedRateLimitRequests:    0, // disabled
	AdvancedThrottlingMultiplier: 0, // disabled

	SoftwareVersion: "0.0.0-testrig",

	// simply use cache defaults.
	Cache: config.Defaults.Cache,
}
