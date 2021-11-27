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

package config

import (
	"github.com/coreos/go-oidc/v3/oidc"
)

// Flags and usage strings for configuration.
const (
	UsernameFlag  = "username"
	UsernameUsage = "the username to create/delete/etc"

	EmailFlag  = "email"
	EmailUsage = "the email address of this account"

	PasswordFlag  = "password"
	PasswordUsage = "the password to set for this account"

	TransPathFlag  = "path"
	TransPathUsage = "the path of the file to import from/export to"
)

// Flags is used for storing the names of the various flags used for
// initializing and storing flag variables.
type Flags struct {
	LogLevel        string
	ApplicationName string
	ConfigPath      string
	Host            string
	AccountDomain   string
	Protocol        string
	BindAddress     string
	Port            string
	TrustedProxies  string

	DbType      string
	DbAddress   string
	DbPort      string
	DbUser      string
	DbPassword  string
	DbDatabase  string
	DbTLSMode   string
	DbTLSCACert string

	TemplateBaseDir string
	AssetBaseDir    string

	AccountsOpenRegistration string
	AccountsApprovalRequired string
	AccountsReasonRequired   string

	MediaMaxImageSize        string
	MediaMaxVideoSize        string
	MediaMinDescriptionChars string
	MediaMaxDescriptionChars string

	StorageBackend       string
	StorageBasePath      string
	StorageServeProtocol string
	StorageServeHost     string
	StorageServeBasePath string

	StatusesMaxChars           string
	StatusesCWMaxChars         string
	StatusesPollMaxOptions     string
	StatusesPollOptionMaxChars string
	StatusesMaxMediaFiles      string

	LetsEncryptEnabled      string
	LetsEncryptCertDir      string
	LetsEncryptEmailAddress string
	LetsEncryptPort         string

	OIDCEnabled          string
	OIDCIdpName          string
	OIDCSkipVerification string
	OIDCIssuer           string
	OIDCClientID         string
	OIDCClientSecret     string
	OIDCScopes           string

	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

// Values contains all the default values for a gotosocial config
type Values struct {
	LogLevel        string
	ApplicationName string
	ConfigPath      string
	Host            string
	AccountDomain   string
	Protocol        string
	BindAddress     string
	Port            int
	TrustedProxies  []string
	SoftwareVersion string

	DbType      string
	DbAddress   string
	DbPort      int
	DbUser      string
	DbPassword  string
	DbDatabase  string
	DBTlsMode   string
	DBTlsCACert string

	TemplateBaseDir string
	AssetBaseDir    string

	AccountsOpenRegistration bool
	AccountsRequireApproval  bool
	AccountsReasonRequired   bool

	MediaMaxImageSize        int
	MediaMaxVideoSize        int
	MediaMinDescriptionChars int
	MediaMaxDescriptionChars int

	StorageBackend       string
	StorageBasePath      string
	StorageServeProtocol string
	StorageServeHost     string
	StorageServeBasePath string

	StatusesMaxChars           int
	StatusesCWMaxChars         int
	StatusesPollMaxOptions     int
	StatusesPollOptionMaxChars int
	StatusesMaxMediaFiles      int

	LetsEncryptEnabled      bool
	LetsEncryptCertDir      string
	LetsEncryptEmailAddress string
	LetsEncryptPort         int

	OIDCEnabled          bool
	OIDCIdpName          string
	OIDCSkipVerification bool
	OIDCIssuer           string
	OIDCClientID         string
	OIDCClientSecret     string
	OIDCScopes           []string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

// FlagNames returns a struct containing the names of the various flags used for
// initializing and storing flag variables.
func FlagNames() Flags {
	return Flags{
		LogLevel:        "log-level",
		ApplicationName: "application-name",
		ConfigPath:      "config-path",
		Host:            "host",
		AccountDomain:   "account-domain",
		Protocol:        "protocol",
		BindAddress:     "bind-address",
		Port:            "port",
		TrustedProxies:  "trusted-proxies",

		DbType:      "db-type",
		DbAddress:   "db-address",
		DbPort:      "db-port",
		DbUser:      "db-user",
		DbPassword:  "db-password",
		DbDatabase:  "db-database",
		DbTLSMode:   "db-tls-mode",
		DbTLSCACert: "db-tls-ca-cert",

		TemplateBaseDir: "template-basedir",
		AssetBaseDir:    "asset-basedir",

		AccountsOpenRegistration: "accounts-open-registration",
		AccountsApprovalRequired: "accounts-approval-required",
		AccountsReasonRequired:   "accounts-reason-required",

		MediaMaxImageSize:        "media-max-image-size",
		MediaMaxVideoSize:        "media-max-video-size",
		MediaMinDescriptionChars: "media-min-description-chars",
		MediaMaxDescriptionChars: "media-max-description-chars",

		StorageBackend:       "storage-backend",
		StorageBasePath:      "storage-base-path",
		StorageServeProtocol: "storage-serve-protocol",
		StorageServeHost:     "storage-serve-host",
		StorageServeBasePath: "storage-serve-base-path",

		StatusesMaxChars:           "statuses-max-chars",
		StatusesCWMaxChars:         "statuses-cw-max-chars",
		StatusesPollMaxOptions:     "statuses-poll-max-options",
		StatusesPollOptionMaxChars: "statuses-poll-option-max-chars",
		StatusesMaxMediaFiles:      "statuses-max-media-files",

		LetsEncryptEnabled:      "letsencrypt-enabled",
		LetsEncryptPort:         "letsencrypt-port",
		LetsEncryptCertDir:      "letsencrypt-cert-dir",
		LetsEncryptEmailAddress: "letsencrypt-email",

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
	}
}

func FlagUsage() Flags {
	return Flags{
		LogLevel:        "Log level to run at: debug, info, warn, fatal",
		ApplicationName: "Name of the application, used in various places internally",
		ConfigPath:      "Path to a yaml file containing gotosocial configuration. Values set in this file will be overwritten by values set as env vars or arguments",
		Host:            "Hostname to use for the server (eg., example.org, gotosocial.whatever.com). DO NOT change this on a server that's already run!",
		AccountDomain:   "Domain to use in account names (eg., example.org, whatever.com). If not set, will default to the setting for host. DO NOT change this on a server that's already run!",
		Protocol:        "Protocol to use for the REST api of the server (only use http for debugging and tests!)",
		BindAddress:     "Bind address to use for the GoToSocial server (eg., 0.0.0.0, 172.138.0.9, [::], localhost). For ipv6, enclose the address in square brackets, eg [2001:db8::fed1]. Default binds to all interfaces.",
		Port:            "Port to use for GoToSocial. Change this to 443 if you're running the binary directly on the host machine.",
		TrustedProxies:  "Proxies to trust when parsing x-forwarded headers into real IPs.",

		DbType:      "Database type: eg., postgres",
		DbAddress:   "Database ipv4 address, hostname, or filename",
		DbPort:      "Database port",
		DbUser:      "Database username",
		DbPassword:  "Database password",
		DbDatabase:  "Database name",
		DbTLSMode:   "Database tls mode",
		DbTLSCACert: "Path to CA cert for db tls connection",

		TemplateBaseDir: "Basedir for html templating files for rendering pages and composing emails.",
		AssetBaseDir:    "Directory to serve static assets from, accessible at example.org/assets/",

		AccountsOpenRegistration: "Allow anyone to submit an account signup request. If false, server will be invite-only.",
		AccountsApprovalRequired: "Do account signups require approval by an admin or moderator before user can log in? If false, new registrations will be automatically approved.",
		AccountsReasonRequired:   "Do new account signups require a reason to be submitted on registration?",

		MediaMaxImageSize:        "Max size of accepted images in bytes",
		MediaMaxVideoSize:        "Max size of accepted videos in bytes",
		MediaMinDescriptionChars: "Min required chars for an image description",
		MediaMaxDescriptionChars: "Max permitted chars for an image description",

		StorageBackend:       "Storage backend to use for media attachments",
		StorageBasePath:      "Full path to an already-created directory where gts should store/retrieve media files. Subfolders will be created within this dir.",
		StorageServeProtocol: "Protocol to use for serving media attachments (use https if storage is local)",
		StorageServeHost:     "Hostname to serve media attachments from (use the same value as host if storage is local)",
		StorageServeBasePath: "Path to append to protocol and hostname to create the base path from which media files will be served (default will mostly be fine)",

		StatusesMaxChars:           "Max permitted characters for posted statuses",
		StatusesCWMaxChars:         "Max permitted characters for content/spoiler warnings on statuses",
		StatusesPollMaxOptions:     "Max amount of options permitted on a poll",
		StatusesPollOptionMaxChars: "Max amount of characters for a poll option",
		StatusesMaxMediaFiles:      "Maximum number of media files/attachments per status",

		LetsEncryptEnabled:      "Enable letsencrypt TLS certs for this server. If set to true, then cert dir also needs to be set (or take the default).",
		LetsEncryptPort:         "Port to listen on for letsencrypt certificate challenges. Must not be the same as the GtS webserver/API port.",
		LetsEncryptCertDir:      "Directory to store acquired letsencrypt certificates.",
		LetsEncryptEmailAddress: "Email address to use when requesting letsencrypt certs. Will receive updates on cert expiry etc.",

		OIDCEnabled:          "Enabled OIDC authorization for this instance. If set to true, then the other OIDC flags must also be set.",
		OIDCIdpName:          "Name of the OIDC identity provider. Will be shown to the user when logging in.",
		OIDCSkipVerification: "Skip verification of tokens returned by the OIDC provider. Should only be set to 'true' for testing purposes, never in a production environment!",
		OIDCIssuer:           "Address of the OIDC issuer. Should be the web address, including protocol, at which the issuer can be reached. Eg., 'https://example.org/auth'",
		OIDCClientID:         "ClientID of GoToSocial, as registered with the OIDC provider.",
		OIDCClientSecret:     "ClientSecret of GoToSocial, as registered with the OIDC provider.",
		OIDCScopes:           "OIDC scopes.",

		SMTPHost:     "Host of the smtp server. Eg., 'smtp.eu.mailgun.org'",
		SMTPPort:     "Port of the smtp server. Eg., 587",
		SMTPUsername: "Username to authenticate with the smtp server as. Eg., 'postmaster@mail.example.org'",
		SMTPPassword: "Password to pass to the smtp server.",
		SMTPFrom:     "Address to use as the 'from' field of the email. Eg., 'gotosocial@example.org'",
	}
}

// Defaults returns a populated Defaults struct with most of the values set to reasonable defaults.
// Note that if you use this function, you still need to set Host and, if desired, ConfigPath.
func Defaults() Values {
	return Values{
		LogLevel:        "info",
		ApplicationName: "gotosocial",
		ConfigPath:      "",
		Host:            "",
		AccountDomain:   "",
		Protocol:        "https",
		BindAddress:     "0.0.0.0",
		Port:            8080,
		TrustedProxies:  []string{"127.0.0.1/32"}, // localhost

		DbType:      "postgres",
		DbAddress:   "localhost",
		DbPort:      5432,
		DbUser:      "postgres",
		DbPassword:  "postgres",
		DbDatabase:  "postgres",
		DBTlsMode:   "disable",
		DBTlsCACert: "",

		TemplateBaseDir: "./web/template/",
		AssetBaseDir:    "./web/assets/",

		AccountsOpenRegistration: true,
		AccountsRequireApproval:  true,
		AccountsReasonRequired:   true,

		MediaMaxImageSize:        2097152,  // 2mb
		MediaMaxVideoSize:        10485760, // 10mb
		MediaMinDescriptionChars: 0,
		MediaMaxDescriptionChars: 500,

		StorageBackend:       "local",
		StorageBasePath:      "/gotosocial/storage",
		StorageServeProtocol: "https",
		StorageServeHost:     "localhost",
		StorageServeBasePath: "/fileserver",

		StatusesMaxChars:           5000,
		StatusesCWMaxChars:         100,
		StatusesPollMaxOptions:     6,
		StatusesPollOptionMaxChars: 50,
		StatusesMaxMediaFiles:      6,

		LetsEncryptEnabled:      true,
		LetsEncryptPort:         80,
		LetsEncryptCertDir:      "/gotosocial/storage/certs",
		LetsEncryptEmailAddress: "",

		OIDCEnabled:          false,
		OIDCIdpName:          "",
		OIDCSkipVerification: false,
		OIDCIssuer:           "",
		OIDCClientID:         "",
		OIDCClientSecret:     "",
		OIDCScopes:           []string{oidc.ScopeOpenID, "profile", "email", "groups"},

		SMTPHost:     "",
		SMTPPort:     0,
		SMTPUsername: "",
		SMTPPassword: "",
		SMTPFrom:     "GoToSocial",
	}
}

// TestDefaults returns a Defaults struct with values set that are suitable for local testing.
func TestDefaults() Values {
	return Values{
		LogLevel:        "trace",
		ApplicationName: "gotosocial",
		ConfigPath:      "",
		Host:            "localhost:8080",
		AccountDomain:   "localhost:8080",
		Protocol:        "http",
		BindAddress:     "127.0.0.1",
		Port:            8080,
		TrustedProxies:  []string{"127.0.0.1/32"},

		DbType:     "sqlite",
		DbAddress:  ":memory:",
		DbPort:     5432,
		DbUser:     "postgres",
		DbPassword: "postgres",
		DbDatabase: "postgres",

		TemplateBaseDir: "./web/template/",
		AssetBaseDir:    "./web/assets/",

		AccountsOpenRegistration: true,
		AccountsRequireApproval:  true,
		AccountsReasonRequired:   true,

		MediaMaxImageSize:        1048576, // 1mb
		MediaMaxVideoSize:        5242880, // 5mb
		MediaMinDescriptionChars: 0,
		MediaMaxDescriptionChars: 500,

		StorageBackend:       "local",
		StorageBasePath:      "/gotosocial/storage",
		StorageServeProtocol: "http",
		StorageServeHost:     "localhost:8080",
		StorageServeBasePath: "/fileserver",

		StatusesMaxChars:           5000,
		StatusesCWMaxChars:         100,
		StatusesPollMaxOptions:     6,
		StatusesPollOptionMaxChars: 50,
		StatusesMaxMediaFiles:      6,

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

		SMTPHost:     "",
		SMTPPort:     0,
		SMTPUsername: "",
		SMTPPassword: "",
		SMTPFrom:     "GoToSocial",
	}
}
