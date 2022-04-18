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

package flag

import "github.com/superseriousbusiness/gotosocial/internal/config"

var usage = config.KeyNames{
	LogLevel:                   "Log level to run at: [trace, debug, info, warn, fatal]",
	LogDbQueries:               "Log database queries verbosely when log-level is trace or debug",
	ApplicationName:            "Name of the application, used in various places internally",
	ConfigPath:                 "Path to a file containing gotosocial configuration. Values set in this file will be overwritten by values set as env vars or arguments",
	Host:                       "Hostname to use for the server (eg., example.org, gotosocial.whatever.com). DO NOT change this on a server that's already run!",
	AccountDomain:              "Domain to use in account names (eg., example.org, whatever.com). If not set, will default to the setting for host. DO NOT change this on a server that's already run!",
	Protocol:                   "Protocol to use for the REST api of the server (only use http for debugging and tests!)",
	BindAddress:                "Bind address to use for the GoToSocial server (eg., 0.0.0.0, 172.138.0.9, [::], localhost). For ipv6, enclose the address in square brackets, eg [2001:db8::fed1]. Default binds to all interfaces.",
	Port:                       "Port to use for GoToSocial. Change this to 443 if you're running the binary directly on the host machine.",
	TrustedProxies:             "Proxies to trust when parsing x-forwarded headers into real IPs.",
	DbType:                     "Database type: eg., postgres",
	DbAddress:                  "Database ipv4 address, hostname, or filename",
	DbPort:                     "Database port",
	DbUser:                     "Database username",
	DbPassword:                 "Database password",
	DbDatabase:                 "Database name",
	DbTLSMode:                  "Database tls mode",
	DbTLSCACert:                "Path to CA cert for db tls connection",
	WebTemplateBaseDir:         "Basedir for html templating files for rendering pages and composing emails.",
	WebAssetBaseDir:            "Directory to serve static assets from, accessible at example.org/assets/",
	AccountsRegistrationOpen:   "Allow anyone to submit an account signup request. If false, server will be invite-only.",
	AccountsApprovalRequired:   "Do account signups require approval by an admin or moderator before user can log in? If false, new registrations will be automatically approved.",
	AccountsReasonRequired:     "Do new account signups require a reason to be submitted on registration?",
	MediaImageMaxSize:          "Max size of accepted images in bytes",
	MediaVideoMaxSize:          "Max size of accepted videos in bytes",
	MediaDescriptionMinChars:   "Min required chars for an image description",
	MediaDescriptionMaxChars:   "Max permitted chars for an image description",
	MediaRemoteCacheDays:       "Number of days to locally cache media from remote instances. If set to 0, remote media will be kept indefinitely.",
	StorageBackend:             "Storage backend to use for media attachments",
	StorageLocalBasePath:       "Full path to an already-created directory where gts should store/retrieve media files. Subfolders will be created within this dir.",
	StatusesMaxChars:           "Max permitted characters for posted statuses",
	StatusesCWMaxChars:         "Max permitted characters for content/spoiler warnings on statuses",
	StatusesPollMaxOptions:     "Max amount of options permitted on a poll",
	StatusesPollOptionMaxChars: "Max amount of characters for a poll option",
	StatusesMediaMaxFiles:      "Maximum number of media files/attachments per status",
	LetsEncryptEnabled:         "Enable letsencrypt TLS certs for this server. If set to true, then cert dir also needs to be set (or take the default).",
	LetsEncryptPort:            "Port to listen on for letsencrypt certificate challenges. Must not be the same as the GtS webserver/API port.",
	LetsEncryptCertDir:         "Directory to store acquired letsencrypt certificates.",
	LetsEncryptEmailAddress:    "Email address to use when requesting letsencrypt certs. Will receive updates on cert expiry etc.",
	OIDCEnabled:                "Enabled OIDC authorization for this instance. If set to true, then the other OIDC flags must also be set.",
	OIDCIdpName:                "Name of the OIDC identity provider. Will be shown to the user when logging in.",
	OIDCSkipVerification:       "Skip verification of tokens returned by the OIDC provider. Should only be set to 'true' for testing purposes, never in a production environment!",
	OIDCIssuer:                 "Address of the OIDC issuer. Should be the web address, including protocol, at which the issuer can be reached. Eg., 'https://example.org/auth'",
	OIDCClientID:               "ClientID of GoToSocial, as registered with the OIDC provider.",
	OIDCClientSecret:           "ClientSecret of GoToSocial, as registered with the OIDC provider.",
	OIDCScopes:                 "OIDC scopes.",
	SMTPHost:                   "Host of the smtp server. Eg., 'smtp.eu.mailgun.org'",
	SMTPPort:                   "Port of the smtp server. Eg., 587",
	SMTPUsername:               "Username to authenticate with the smtp server as. Eg., 'postmaster@mail.example.org'",
	SMTPPassword:               "Password to pass to the smtp server.",
	SMTPFrom:                   "Address to use as the 'from' field of the email. Eg., 'gotosocial@example.org'",
	SyslogEnabled:              "Enable the syslog logging hook. Logs will be mirrored to the configured destination.",
	SyslogProtocol:             "Protocol to use when directing logs to syslog. Leave empty to connect to local syslog.",
	SyslogAddress:              "Address:port to send syslog logs to. Leave empty to connect to local syslog.",
	AdminAccountUsername:       "the username to create/delete/etc",
	AdminAccountEmail:          "the email address of this account",
	AdminAccountPassword:       "the password to set for this account",
	AdminTransPath:             "the path of the file to import from/export to",
}
