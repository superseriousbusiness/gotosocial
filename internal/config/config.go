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

import (
	"reflect"

	"codeberg.org/gruf/go-bytesize"
	"github.com/mitchellh/mapstructure"
)

// cfgtype is the reflected type information of Configuration{}.
var cfgtype = reflect.TypeOf(Configuration{})

// fieldtag will fetch the string value for the given tag name
// on the given field name in the Configuration{} struct.
func fieldtag(field, tag string) string {
	sfield, ok := cfgtype.FieldByName(field)
	if !ok {
		panic("unknown struct field")
	}
	return sfield.Tag.Get(tag)
}

// Configuration represents global GTS server runtime configuration.
//
// Please note that if you update this struct's fields or tags, you
// will need to regenerate the global Getter/Setter helpers by running:
// `go run ./internal/config/gen/ -out ./internal/config/helpers.gen.go`
type Configuration struct {
	LogLevel        string   `name:"log-level" usage:"Log level to run at: [trace, debug, info, warn, fatal]"`
	LogDbQueries    bool     `name:"log-db-queries" usage:"Log database queries verbosely when log-level is trace or debug"`
	ApplicationName string   `name:"application-name" usage:"Name of the application, used in various places internally"`
	LandingPageUser string   `name:"landing-page-user" usage:"the user that should be shown on the instance's landing page"`
	ConfigPath      string   `name:"config-path" usage:"Path to a file containing gotosocial configuration. Values set in this file will be overwritten by values set as env vars or arguments"`
	Host            string   `name:"host" usage:"Hostname to use for the server (eg., example.org, gotosocial.whatever.com). DO NOT change this on a server that's already run!"`
	AccountDomain   string   `name:"account-domain" usage:"Domain to use in account names (eg., example.org, whatever.com). If not set, will default to the setting for host. DO NOT change this on a server that's already run!"`
	Protocol        string   `name:"protocol" usage:"Protocol to use for the REST api of the server (only use http if you are debugging or behind a reverse proxy!)"`
	BindAddress     string   `name:"bind-address" usage:"Bind address to use for the GoToSocial server (eg., 0.0.0.0, 172.138.0.9, [::], localhost). For ipv6, enclose the address in square brackets, eg [2001:db8::fed1]. Default binds to all interfaces."`
	Port            int      `name:"port" usage:"Port to use for GoToSocial. Change this to 443 if you're running the binary directly on the host machine."`
	TrustedProxies  []string `name:"trusted-proxies" usage:"Proxies to trust when parsing x-forwarded headers into real IPs."`
	SoftwareVersion string   `name:"software-version" usage:""`

	DbType      string `name:"db-type" usage:"Database type: eg., postgres"`
	DbAddress   string `name:"db-address" usage:"Database ipv4 address, hostname, or filename"`
	DbPort      int    `name:"db-port" usage:"Database port"`
	DbUser      string `name:"db-user" usage:"Database username"`
	DbPassword  string `name:"db-password" usage:"Database password"`
	DbDatabase  string `name:"db-database" usage:"Database name"`
	DbTLSMode   string `name:"db-tls-mode" usage:"Database tls mode"`
	DbTLSCACert string `name:"db-tls-ca-cert" usage:"Path to CA cert for db tls connection"`

	WebTemplateBaseDir string `name:"web-template-base-dir" usage:"Basedir for html templating files for rendering pages and composing emails."`
	WebAssetBaseDir    string `name:"web-asset-base-dir" usage:"Directory to serve static assets from, accessible at example.org/assets/"`

	InstanceExposePeers            bool `name:"instance-expose-peers" usage:"Allow unauthenticated users to query /api/v1/instance/peers?filter=open"`
	InstanceExposeSuspended        bool `name:"instance-expose-suspended" usage:"Expose suspended instances via web UI, and allow unauthenticated users to query /api/v1/instance/peers?filter=suspended"`
	InstanceExposePublicTimeline   bool `name:"instance-expose-public-timeline" usage:"Allow unauthenticated users to query /api/v1/timelines/public"`
	InstanceDeliverToSharedInboxes bool `name:"instance-deliver-to-shared-inboxes" usage:"Deliver federated messages to shared inboxes, if they're available."`

	AccountsRegistrationOpen bool `name:"accounts-registration-open" usage:"Allow anyone to submit an account signup request. If false, server will be invite-only."`
	AccountsApprovalRequired bool `name:"accounts-approval-required" usage:"Do account signups require approval by an admin or moderator before user can log in? If false, new registrations will be automatically approved."`
	AccountsReasonRequired   bool `name:"accounts-reason-required" usage:"Do new account signups require a reason to be submitted on registration?"`
	AccountsAllowCustomCSS   bool `name:"accounts-allow-custom-css" usage:"Allow accounts to enable custom CSS for their profile pages and statuses."`

	MediaImageMaxSize        bytesize.Size `name:"media-image-max-size" usage:"Max size of accepted images in bytes"`
	MediaVideoMaxSize        bytesize.Size `name:"media-video-max-size" usage:"Max size of accepted videos in bytes"`
	MediaDescriptionMinChars int           `name:"media-description-min-chars" usage:"Min required chars for an image description"`
	MediaDescriptionMaxChars int           `name:"media-description-max-chars" usage:"Max permitted chars for an image description"`
	MediaRemoteCacheDays     int           `name:"media-remote-cache-days" usage:"Number of days to locally cache media from remote instances. If set to 0, remote media will be kept indefinitely."`
	MediaEmojiLocalMaxSize   bytesize.Size `name:"media-emoji-local-max-size" usage:"Max size in bytes of emojis uploaded to this instance via the admin API."`
	MediaEmojiRemoteMaxSize  bytesize.Size `name:"media-emoji-remote-max-size" usage:"Max size in bytes of emojis to download from other instances."`

	StorageBackend       string `name:"storage-backend" usage:"Storage backend to use for media attachments"`
	StorageLocalBasePath string `name:"storage-local-base-path" usage:"Full path to an already-created directory where gts should store/retrieve media files. Subfolders will be created within this dir."`
	StorageS3Endpoint    string `name:"storage-s3-endpoint" usage:"S3 Endpoint URL (e.g 'minio.example.org:9000')"`
	StorageS3AccessKey   string `name:"storage-s3-access-key" usage:"S3 Access Key"`
	StorageS3SecretKey   string `name:"storage-s3-secret-key" usage:"S3 Secret Key"`
	StorageS3UseSSL      bool   `name:"storage-s3-use-ssl" usage:"Use SSL for S3 connections. Only set this to 'false' when testing locally"`
	StorageS3BucketName  string `name:"storage-s3-bucket" usage:"Place blobs in this bucket"`
	StorageS3Proxy       bool   `name:"storage-s3-proxy" usage:"Proxy S3 contents through GoToSocial instead of redirecting to a presigned URL"`

	StatusesMaxChars           int `name:"statuses-max-chars" usage:"Max permitted characters for posted statuses"`
	StatusesCWMaxChars         int `name:"statuses-cw-max-chars" usage:"Max permitted characters for content/spoiler warnings on statuses"`
	StatusesPollMaxOptions     int `name:"statuses-poll-max-options" usage:"Max amount of options permitted on a poll"`
	StatusesPollOptionMaxChars int `name:"statuses-poll-option-max-chars" usage:"Max amount of characters for a poll option"`
	StatusesMediaMaxFiles      int `name:"statuses-media-max-files" usage:"Maximum number of media files/attachments per status"`

	LetsEncryptEnabled      bool   `name:"letsencrypt-enabled" usage:"Enable letsencrypt TLS certs for this server. If set to true, then cert dir also needs to be set (or take the default)."`
	LetsEncryptPort         int    `name:"letsencrypt-port" usage:"Port to listen on for letsencrypt certificate challenges. Must not be the same as the GtS webserver/API port."`
	LetsEncryptCertDir      string `name:"letsencrypt-cert-dir" usage:"Directory to store acquired letsencrypt certificates."`
	LetsEncryptEmailAddress string `name:"letsencrypt-email-address" usage:"Email address to use when requesting letsencrypt certs. Will receive updates on cert expiry etc."`

	OIDCEnabled          bool     `name:"oidc-enabled" usage:"Enabled OIDC authorization for this instance. If set to true, then the other OIDC flags must also be set."`
	OIDCIdpName          string   `name:"oidc-idp-name" usage:"Name of the OIDC identity provider. Will be shown to the user when logging in."`
	OIDCSkipVerification bool     `name:"oidc-skip-verification" usage:"Skip verification of tokens returned by the OIDC provider. Should only be set to 'true' for testing purposes, never in a production environment!"`
	OIDCIssuer           string   `name:"oidc-issuer" usage:"Address of the OIDC issuer. Should be the web address, including protocol, at which the issuer can be reached. Eg., 'https://example.org/auth'"`
	OIDCClientID         string   `name:"oidc-client-id" usage:"ClientID of GoToSocial, as registered with the OIDC provider."`
	OIDCClientSecret     string   `name:"oidc-client-secret" usage:"ClientSecret of GoToSocial, as registered with the OIDC provider."`
	OIDCScopes           []string `name:"oidc-scopes" usage:"OIDC scopes."`

	SMTPHost     string `name:"smtp-host" usage:"Host of the smtp server. Eg., 'smtp.eu.mailgun.org'"`
	SMTPPort     int    `name:"smtp-port" usage:"Port of the smtp server. Eg., 587"`
	SMTPUsername string `name:"smtp-username" usage:"Username to authenticate with the smtp server as. Eg., 'postmaster@mail.example.org'"`
	SMTPPassword string `name:"smtp-password" usage:"Password to pass to the smtp server."`
	SMTPFrom     string `name:"smtp-from" usage:"Address to use as the 'from' field of the email. Eg., 'gotosocial@example.org'"`

	SyslogEnabled  bool   `name:"syslog-enabled" usage:"Enable the syslog logging hook. Logs will be mirrored to the configured destination."`
	SyslogProtocol string `name:"syslog-protocol" usage:"Protocol to use when directing logs to syslog. Leave empty to connect to local syslog."`
	SyslogAddress  string `name:"syslog-address" usage:"Address:port to send syslog logs to. Leave empty to connect to local syslog."`

	// TODO: move these elsewhere, these are more ephemeral vs long-running flags like above
	AdminAccountUsername string `name:"username" usage:"the username to create/delete/etc"`
	AdminAccountEmail    string `name:"email" usage:"the email address of this account"`
	AdminAccountPassword string `name:"password" usage:"the password to set for this account"`
	AdminTransPath       string `name:"path" usage:"the path of the file to import from/export to"`

	AdvancedCookiesSamesite   string `name:"advanced-cookies-samesite" usage:"'strict' or 'lax', see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite"`
	AdvancedRateLimitRequests int    `name:"advanced-rate-limit-requests" usage:"Amount of HTTP requests to permit within a 5 minute window. 0 or less turns rate limiting off."`
}

// MarshalMap will marshal current Configuration into a map structure (useful for JSON).
func (cfg *Configuration) MarshalMap() (map[string]interface{}, error) {
	var dst map[string]interface{}
	dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "name",
		Result:  &dst,
	})
	if err := dec.Decode(cfg); err != nil {
		return nil, err
	}
	return dst, nil
}
