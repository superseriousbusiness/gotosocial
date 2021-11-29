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

import "github.com/spf13/pflag"

// AttachCommonFlags attaches flags that are common to all commands.
func AttachCommonFlags(flags *pflag.FlagSet, values Values) {
	flags.String(FlagNames.LogLevel, values.LogLevel, FlagUsage.LogLevel)
	flags.String(FlagNames.ConfigPath, values.ConfigPath, FlagUsage.ConfigPath)
}

// AttachGeneralFlags attaches flags pertaining to general config.
func AttachGeneralFlags(flags *pflag.FlagSet, values Values) {
	flags.String(FlagNames.ApplicationName, values.ApplicationName, FlagUsage.ApplicationName)
	flags.String(FlagNames.Host, values.Host, FlagUsage.Host)
	flags.String(FlagNames.AccountDomain, values.AccountDomain, FlagUsage.AccountDomain)
	flags.String(FlagNames.Protocol, values.Protocol, FlagUsage.Protocol)
	flags.String(FlagNames.BindAddress, values.BindAddress, FlagUsage.BindAddress)
	flags.Int(FlagNames.Port, values.Port, FlagUsage.Port)
	flags.StringArray(FlagNames.TrustedProxies, values.TrustedProxies, FlagUsage.TrustedProxies)
}

// AttachDatabaseFlags attaches flags pertaining to database config.
func AttachDatabaseFlags(flags *pflag.FlagSet, values Values) {
	flags.String(FlagNames.DbType, values.DbType, FlagUsage.DbType)
	flags.String(FlagNames.DbAddress, values.DbAddress, FlagUsage.DbAddress)
	flags.Int(FlagNames.DbPort, values.DbPort, FlagUsage.DbPort)
	flags.String(FlagNames.DbUser, values.DbUser, FlagUsage.DbUser)
	flags.String(FlagNames.DbPassword, values.DbPassword, FlagUsage.DbPassword)
	flags.String(FlagNames.DbDatabase, values.DbDatabase, FlagUsage.DbDatabase)
	flags.String(FlagNames.DbTLSMode, values.DbTLSMode, FlagUsage.DbTLSMode)
	flags.String(FlagNames.DbTLSCACert, values.DbTLSCACert, FlagUsage.DbTLSCACert)
}

// AttachTemplateFlags attaches flags pertaining to templating config.
func AttachTemplateFlags(flags *pflag.FlagSet, values Values) {
	flags.String(FlagNames.TemplateBaseDir, values.TemplateBaseDir, FlagUsage.TemplateBaseDir)
	flags.String(FlagNames.AssetBaseDir, values.AssetBaseDir, FlagUsage.AssetBaseDir)
}

// AttachAccountsFlags attaches flags pertaining to account config.
func AttachAccountsFlags(flags *pflag.FlagSet, values Values) {
	flags.Bool(FlagNames.AccountsOpenRegistration, values.AccountsOpenRegistration, FlagUsage.AccountsOpenRegistration)
	flags.Bool(FlagNames.AccountsApprovalRequired, values.AccountsApprovalRequired, FlagUsage.AccountsApprovalRequired)
	flags.Bool(FlagNames.AccountsReasonRequired, values.AccountsReasonRequired, FlagUsage.AccountsReasonRequired)
}

// AttachMediaFlags attaches flags pertaining to media config.
func AttachMediaFlags(flags *pflag.FlagSet, values Values) {
	flags.Int(FlagNames.MediaMaxImageSize, values.MediaMaxImageSize, FlagUsage.MediaMaxImageSize)
	flags.Int(FlagNames.MediaMaxVideoSize, values.MediaMaxVideoSize, FlagUsage.MediaMaxVideoSize)
	flags.Int(FlagNames.MediaMinDescriptionChars, values.MediaMinDescriptionChars, FlagUsage.MediaMinDescriptionChars)
	flags.Int(FlagNames.MediaMaxDescriptionChars, values.MediaMaxDescriptionChars, FlagUsage.MediaMaxDescriptionChars)
}

// AttachStorageFlags attaches flags pertaining to storage config.
func AttachStorageFlags(flags *pflag.FlagSet, values Values) {
	flags.String(FlagNames.StorageBackend, values.StorageBackend, FlagUsage.StorageBackend)
	flags.String(FlagNames.StorageBasePath, values.StorageBasePath, FlagUsage.StorageBasePath)
	flags.String(FlagNames.StorageServeProtocol, values.StorageServeProtocol, FlagUsage.StorageServeProtocol)
	flags.String(FlagNames.StorageServeHost, values.StorageServeHost, FlagUsage.StorageServeHost)
	flags.String(FlagNames.StorageServeBasePath, values.StorageServeBasePath, FlagUsage.StorageServeBasePath)
}

// AttachStatusesFlags attaches flags pertaining to statuses config.
func AttachStatusesFlags(flags *pflag.FlagSet, values Values) {
	flags.Int(FlagNames.StatusesMaxChars, values.StatusesMaxChars, FlagUsage.StatusesMaxChars)
	flags.Int(FlagNames.StatusesCWMaxChars, values.StatusesCWMaxChars, FlagUsage.StatusesCWMaxChars)
	flags.Int(FlagNames.StatusesPollMaxOptions, values.StatusesPollMaxOptions, FlagUsage.StatusesPollMaxOptions)
	flags.Int(FlagNames.StatusesPollOptionMaxChars, values.StatusesPollOptionMaxChars, FlagUsage.StatusesPollOptionMaxChars)
	flags.Int(FlagNames.StatusesMaxMediaFiles, values.StatusesMaxMediaFiles, FlagUsage.StatusesMaxMediaFiles)
}

// AttachLetsEncryptFlags attaches flags pertaining to letsencrypt config.
func AttachLetsEncryptFlags(flags *pflag.FlagSet, values Values) {
	flags.Bool(FlagNames.LetsEncryptEnabled, values.LetsEncryptEnabled, FlagUsage.LetsEncryptEnabled)
	flags.Int(FlagNames.LetsEncryptPort, values.LetsEncryptPort, FlagUsage.LetsEncryptPort)
	flags.String(FlagNames.LetsEncryptCertDir, values.LetsEncryptCertDir, FlagUsage.LetsEncryptCertDir)
	flags.String(FlagNames.LetsEncryptEmailAddress, values.LetsEncryptEmailAddress, FlagUsage.LetsEncryptEmailAddress)
}

// AttachOIDCFlags attaches flags pertaining to oidc config.
func AttachOIDCFlags(flags *pflag.FlagSet, values Values) {
	flags.Bool(FlagNames.OIDCEnabled, values.OIDCEnabled, FlagUsage.OIDCEnabled)
	flags.String(FlagNames.OIDCIdpName, values.OIDCIdpName, FlagUsage.OIDCIdpName)
	flags.Bool(FlagNames.OIDCSkipVerification, values.OIDCSkipVerification, FlagUsage.OIDCSkipVerification)
	flags.String(FlagNames.OIDCIssuer, values.OIDCIssuer, FlagUsage.OIDCIssuer)
	flags.String(FlagNames.OIDCClientID, values.OIDCClientID, FlagUsage.OIDCClientID)
	flags.String(FlagNames.OIDCClientSecret, values.OIDCClientSecret, FlagUsage.OIDCClientSecret)
	flags.StringArray(FlagNames.OIDCScopes, values.OIDCScopes, FlagUsage.OIDCScopes)
}

// AttachSMTPFlags attaches flags pertaining to smtp/email config.
func AttachSMTPFlags(flags *pflag.FlagSet, values Values) {
	flags.String(FlagNames.SMTPHost, values.SMTPHost, FlagUsage.SMTPHost)
	flags.Int(FlagNames.SMTPPort, values.SMTPPort, FlagUsage.SMTPPort)
	flags.String(FlagNames.SMTPUsername, values.SMTPUsername, FlagUsage.SMTPUsername)
	flags.String(FlagNames.SMTPPassword, values.SMTPPassword, FlagUsage.SMTPPassword)
	flags.String(FlagNames.SMTPFrom, values.SMTPFrom, FlagUsage.SMTPFrom)
}

// AttachServerFlags attaches all flags pertaining to running the GtS server or testrig.
func AttachServerFlags(flags *pflag.FlagSet, values Values) {
	AttachGeneralFlags(flags, values)
	AttachDatabaseFlags(flags, values)
	AttachTemplateFlags(flags, values)
	AttachAccountsFlags(flags, values)
	AttachMediaFlags(flags, values)
	AttachStorageFlags(flags, values)
	AttachStatusesFlags(flags, values)
	AttachLetsEncryptFlags(flags, values)
	AttachOIDCFlags(flags, values)
	AttachSMTPFlags(flags, values)
}

// Flags is used for storing the names of the various flags used for
// initializing and storing flag variables.
type Flags struct {
	// root flags
	LogLevel   string
	ConfigPath string

	// general flags
	ApplicationName string
	Host            string
	AccountDomain   string
	Protocol        string
	BindAddress     string
	Port            string
	TrustedProxies  string
	SoftwareVersion string

	// database flags
	DbType      string
	DbAddress   string
	DbPort      string
	DbUser      string
	DbPassword  string
	DbDatabase  string
	DbTLSMode   string
	DbTLSCACert string

	// template flags
	TemplateBaseDir string
	AssetBaseDir    string

	// accounts flags
	AccountsOpenRegistration string
	AccountsApprovalRequired string
	AccountsReasonRequired   string

	// media flags
	MediaMaxImageSize        string
	MediaMaxVideoSize        string
	MediaMinDescriptionChars string
	MediaMaxDescriptionChars string

	// storage flags
	StorageBackend       string
	StorageBasePath      string
	StorageServeProtocol string
	StorageServeHost     string
	StorageServeBasePath string

	// statuses flags
	StatusesMaxChars           string
	StatusesCWMaxChars         string
	StatusesPollMaxOptions     string
	StatusesPollOptionMaxChars string
	StatusesMaxMediaFiles      string

	// letsencrypt flags
	LetsEncryptEnabled      string
	LetsEncryptCertDir      string
	LetsEncryptEmailAddress string
	LetsEncryptPort         string

	// oidc flags
	OIDCEnabled          string
	OIDCIdpName          string
	OIDCSkipVerification string
	OIDCIssuer           string
	OIDCClientID         string
	OIDCClientSecret     string
	OIDCScopes           string

	// smtp flags
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// admin flags
	AdminAccountUsername string
	AdminAccountEmail    string
	AdminAccountPassword string
	AdminTransPath       string
}

// FlagNames contains the names of the various flags used for initializing and storing flag variables.
var FlagNames = Flags{
	LogLevel:        "log-level",
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

	AdminAccountUsername: "username",
	AdminAccountEmail:    "email",
	AdminAccountPassword: "password",
	AdminTransPath:       "path",
}

// FlagUsage contains the usage text for all flags.
var FlagUsage = Flags{
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

	AdminAccountUsername: "the username to create/delete/etc",
	AdminAccountEmail:    "the email address of this account",
	AdminAccountPassword: "the password to set for this account",

	AdminTransPath: "the path of the file to import from/export to",
}
