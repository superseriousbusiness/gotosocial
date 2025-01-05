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
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"codeberg.org/gruf/go-bytesize"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/language"
	"github.com/superseriousbusiness/gotosocial/internal/media/ffmpeg"
)

func init() {
	ctx := context.Background()

	// Ensure global ffmpeg WASM pool initialized.
	fmt.Println("testrig: precompiling ffmpeg WASM")
	if err := ffmpeg.InitFfmpeg(ctx, 1); err != nil {
		panic(err)
	}

	// Ensure global ffmpeg WASM pool initialized.
	fmt.Println("testrig: precompiling ffprobe WASM")
	if err := ffmpeg.InitFfprobe(ctx, 1); err != nil {
		panic(err)
	}
}

// InitTestConfig initializes viper
// configuration with test defaults.
func InitTestConfig() {
	config.Defaults = testDefaults()
	config.Reset()
}

func testDefaults() config.Configuration {
	return config.Configuration{
		LogLevel:                 envStr("GTS_LOG_LEVEL", "error"),
		LogTimestampFormat:       "02/01/2006 15:04:05.000",
		LogDbQueries:             true,
		ApplicationName:          "gotosocial",
		LandingPageUser:          "",
		ConfigPath:               "",
		Host:                     "localhost:8080",
		AccountDomain:            "localhost:8080",
		Protocol:                 "http",
		BindAddress:              "127.0.0.1",
		Port:                     8080,
		TrustedProxies:           []string{"127.0.0.1/32", "::1"},
		DbType:                   envStr("GTS_DB_TYPE", "sqlite"),
		DbAddress:                envStr("GTS_DB_ADDRESS", ":memory:"),
		DbPort:                   envInt("GTS_DB_PORT", 0),
		DbUser:                   envStr("GTS_DB_USER", ""),
		DbPassword:               envStr("GTS_DB_PASSWORD", ""),
		DbDatabase:               envStr("GTS_DB_DATABASE", ""),
		DbTLSMode:                envStr("GTS_DB_TLS_MODE", ""),
		DbTLSCACert:              envStr("GTS_DB_TLS_CA_CERT", ""),
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
		InstanceSubscriptionsProcessFrom:  "23:00",        // 11pm,
		InstanceSubscriptionsProcessEvery: 24 * time.Hour, // 1/day.

		AccountsRegistrationOpen: true,
		AccountsReasonRequired:   true,
		AccountsAllowCustomCSS:   true,
		AccountsCustomCSSLength:  10000,

		MediaDescriptionMinChars: 0,
		MediaDescriptionMaxChars: 500,
		MediaRemoteCacheDays:     7,
		MediaLocalMaxSize:        40 * bytesize.MiB,
		MediaRemoteMaxSize:       40 * bytesize.MiB,
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

		MetricsEnabled:     true,
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
}

func envInt(key string, _default int) int {
	return env(key, _default, func(value string) int {
		i, _ := strconv.Atoi(value)
		return i
	})
}

func envStr(key string, _default string) string {
	return env(key, _default, func(value string) string {
		return value
	})
}

func env[T any](key string, _default T, parse func(string) T) T {
	value, ok := os.LookupEnv(key)
	if ok {
		return parse(value)
	}
	return _default
}
