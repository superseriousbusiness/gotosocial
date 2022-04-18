/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package config

// KeyNames is a struct that just contains the names of configuration keys.
type KeyNames struct {
	// root
	LogLevel     string
	LogDbQueries string
	ConfigPath   string

	// general
	ApplicationName string
	Host            string
	AccountDomain   string
	Protocol        string
	BindAddress     string
	Port            string
	TrustedProxies  string
	SoftwareVersion string

	// database
	DbType      string
	DbAddress   string
	DbPort      string
	DbUser      string
	DbPassword  string
	DbDatabase  string
	DbTLSMode   string
	DbTLSCACert string

	// template
	WebTemplateBaseDir string
	WebAssetBaseDir    string

	// accounts
	AccountsRegistrationOpen string
	AccountsApprovalRequired string
	AccountsReasonRequired   string

	// media
	MediaImageMaxSize        string
	MediaVideoMaxSize        string
	MediaDescriptionMinChars string
	MediaDescriptionMaxChars string
	MediaRemoteCacheDays     string

	// storage
	StorageBackend       string
	StorageLocalBasePath string

	// statuses
	StatusesMaxChars           string
	StatusesCWMaxChars         string
	StatusesPollMaxOptions     string
	StatusesPollOptionMaxChars string
	StatusesMediaMaxFiles      string

	// letsencrypt
	LetsEncryptEnabled      string
	LetsEncryptCertDir      string
	LetsEncryptEmailAddress string
	LetsEncryptPort         string

	// oidc
	OIDCEnabled          string
	OIDCIdpName          string
	OIDCSkipVerification string
	OIDCIssuer           string
	OIDCClientID         string
	OIDCClientSecret     string
	OIDCScopes           string

	// smtp
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// syslog
	SyslogEnabled  string
	SyslogProtocol string
	SyslogAddress  string

	// admin
	AdminAccountUsername string
	AdminAccountEmail    string
	AdminAccountPassword string
	AdminTransPath       string
}

// Keys contains the names of the various keys used for initializing and storing flag variables,
// and retrieving values from the viper config store.
var Keys = KeyNames{
	LogLevel:        "log-level",
	LogDbQueries:    "log-db-queries",
	ApplicationName: "application-name",
	ConfigPath:      "config-path",
	Host:            "host",
	AccountDomain:   "account-domain",
	Protocol:        "protocol",
	BindAddress:     "bind-address",
	Port:            "port",
	TrustedProxies:  "trusted-proxies",
	SoftwareVersion: "software-version",

	DbType:      "db-type",
	DbAddress:   "db-address",
	DbPort:      "db-port",
	DbUser:      "db-user",
	DbPassword:  "db-password",
	DbDatabase:  "db-database",
	DbTLSMode:   "db-tls-mode",
	DbTLSCACert: "db-tls-ca-cert",

	WebTemplateBaseDir: "web-template-base-dir",
	WebAssetBaseDir:    "web-asset-base-dir",

	AccountsRegistrationOpen: "accounts-registration-open",
	AccountsApprovalRequired: "accounts-approval-required",
	AccountsReasonRequired:   "accounts-reason-required",

	MediaImageMaxSize:        "media-image-max-size",
	MediaVideoMaxSize:        "media-video-max-size",
	MediaDescriptionMinChars: "media-description-min-chars",
	MediaDescriptionMaxChars: "media-description-max-chars",
	MediaRemoteCacheDays:     "media-remote-cache-days",

	StorageBackend:       "storage-backend",
	StorageLocalBasePath: "storage-local-base-path",

	StatusesMaxChars:           "statuses-max-chars",
	StatusesCWMaxChars:         "statuses-cw-max-chars",
	StatusesPollMaxOptions:     "statuses-poll-max-options",
	StatusesPollOptionMaxChars: "statuses-poll-option-max-chars",
	StatusesMediaMaxFiles:      "statuses-media-max-files",

	LetsEncryptEnabled:      "letsencrypt-enabled",
	LetsEncryptPort:         "letsencrypt-port",
	LetsEncryptCertDir:      "letsencrypt-cert-dir",
	LetsEncryptEmailAddress: "letsencrypt-email-address",

	OIDCEnabled:          "oidc-enabled",
	OIDCIdpName:          "oidc-idp-name",
	OIDCSkipVerification: "oidc-skip-verification",
	OIDCIssuer:           "oidc-issuer",
	OIDCClientID:         "oidc-client-id",
	OIDCClientSecret:     "oidc-client-secret",
	OIDCScopes:           "oidc-scopes",

	SMTPHost:     "smtp-host",
	SMTPPort:     "smtp-port",
	SMTPUsername: "smtp-username",
	SMTPPassword: "smtp-password",
	SMTPFrom:     "smtp-from",

	SyslogEnabled:  "syslog-enabled",
	SyslogProtocol: "syslog-protocol",
	SyslogAddress:  "syslog-address",

	AdminAccountUsername: "username",
	AdminAccountEmail:    "email",
	AdminAccountPassword: "password",
	AdminTransPath:       "path",
}
