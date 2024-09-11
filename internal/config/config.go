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
	"reflect"
	"time"

	"codeberg.org/gruf/go-bytesize"
	"github.com/mitchellh/mapstructure"
	"github.com/superseriousbusiness/gotosocial/internal/language"
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
	LogLevel           string   `name:"log-level" usage:"Log level to run at: [trace, debug, info, warn, fatal]"`
	LogTimestampFormat string   `name:"log-timestamp-format" usage:"Format to use for the log timestamp, as supported by Go's time.Layout"`
	LogDbQueries       bool     `name:"log-db-queries" usage:"Log database queries verbosely when log-level is trace or debug"`
	LogClientIP        bool     `name:"log-client-ip" usage:"Include the client IP in logs"`
	ApplicationName    string   `name:"application-name" usage:"Name of the application, used in various places internally"`
	LandingPageUser    string   `name:"landing-page-user" usage:"the user that should be shown on the instance's landing page"`
	ConfigPath         string   `name:"config-path" usage:"Path to a file containing gotosocial configuration. Values set in this file will be overwritten by values set as env vars or arguments"`
	Host               string   `name:"host" usage:"Hostname to use for the server (eg., example.org, gotosocial.whatever.com). DO NOT change this on a server that's already run!"`
	AccountDomain      string   `name:"account-domain" usage:"Domain to use in account names (eg., example.org, whatever.com). If not set, will default to the setting for host. DO NOT change this on a server that's already run!"`
	Protocol           string   `name:"protocol" usage:"Protocol to use for the REST api of the server (only use http if you are debugging or behind a reverse proxy!)"`
	BindAddress        string   `name:"bind-address" usage:"Bind address to use for the GoToSocial server (eg., 0.0.0.0, 172.138.0.9, [::], localhost). For ipv6, enclose the address in square brackets, eg [2001:db8::fed1]. Default binds to all interfaces."`
	Port               int      `name:"port" usage:"Port to use for GoToSocial. Change this to 443 if you're running the binary directly on the host machine."`
	TrustedProxies     []string `name:"trusted-proxies" usage:"Proxies to trust when parsing x-forwarded headers into real IPs."`
	SoftwareVersion    string   `name:"software-version" usage:""`

	DbType                     string        `name:"db-type" usage:"Database type: eg., postgres"`
	DbAddress                  string        `name:"db-address" usage:"Database ipv4 address, hostname, or filename"`
	DbPort                     int           `name:"db-port" usage:"Database port"`
	DbUser                     string        `name:"db-user" usage:"Database username"`
	DbPassword                 string        `name:"db-password" usage:"Database password"`
	DbDatabase                 string        `name:"db-database" usage:"Database name"`
	DbTLSMode                  string        `name:"db-tls-mode" usage:"Database tls mode"`
	DbTLSCACert                string        `name:"db-tls-ca-cert" usage:"Path to CA cert for db tls connection"`
	DbMaxOpenConnsMultiplier   int           `name:"db-max-open-conns-multiplier" usage:"Multiplier to use per cpu for max open database connections. 0 or less is normalized to 1."`
	DbSqliteJournalMode        string        `name:"db-sqlite-journal-mode" usage:"Sqlite only: see https://www.sqlite.org/pragma.html#pragma_journal_mode"`
	DbSqliteSynchronous        string        `name:"db-sqlite-synchronous" usage:"Sqlite only: see https://www.sqlite.org/pragma.html#pragma_synchronous"`
	DbSqliteCacheSize          bytesize.Size `name:"db-sqlite-cache-size" usage:"Sqlite only: see https://www.sqlite.org/pragma.html#pragma_cache_size"`
	DbSqliteBusyTimeout        time.Duration `name:"db-sqlite-busy-timeout" usage:"Sqlite only: see https://www.sqlite.org/pragma.html#pragma_busy_timeout"`
	DbPostgresConnectionString string        `name:"db-postgres-connection-string" usage:"Full Database URL for connection to postgres"`

	WebTemplateBaseDir string `name:"web-template-base-dir" usage:"Basedir for html templating files for rendering pages and composing emails."`
	WebAssetBaseDir    string `name:"web-asset-base-dir" usage:"Directory to serve static assets from, accessible at example.org/assets/"`

	InstanceFederationMode         string             `name:"instance-federation-mode" usage:"Set instance federation mode."`
	InstanceFederationSpamFilter   bool               `name:"instance-federation-spam-filter" usage:"Enable basic spam filter heuristics for messages coming from other instances, and drop messages identified as spam"`
	InstanceExposePeers            bool               `name:"instance-expose-peers" usage:"Allow unauthenticated users to query /api/v1/instance/peers?filter=open"`
	InstanceExposeSuspended        bool               `name:"instance-expose-suspended" usage:"Expose suspended instances via web UI, and allow unauthenticated users to query /api/v1/instance/peers?filter=suspended"`
	InstanceExposeSuspendedWeb     bool               `name:"instance-expose-suspended-web" usage:"Expose list of suspended instances as webpage on /about/suspended"`
	InstanceExposePublicTimeline   bool               `name:"instance-expose-public-timeline" usage:"Allow unauthenticated users to query /api/v1/timelines/public"`
	InstanceDeliverToSharedInboxes bool               `name:"instance-deliver-to-shared-inboxes" usage:"Deliver federated messages to shared inboxes, if they're available."`
	InstanceInjectMastodonVersion  bool               `name:"instance-inject-mastodon-version" usage:"This injects a Mastodon compatible version in /api/v1/instance to help Mastodon clients that use that version for feature detection"`
	InstanceLanguages              language.Languages `name:"instance-languages" usage:"BCP47 language tags for the instance. Used to indicate the preferred languages of instance residents (in order from most-preferred to least-preferred)."`

	AccountsRegistrationOpen bool `name:"accounts-registration-open" usage:"Allow anyone to submit an account signup request. If false, server will be invite-only."`
	AccountsReasonRequired   bool `name:"accounts-reason-required" usage:"Do new account signups require a reason to be submitted on registration?"`
	AccountsAllowCustomCSS   bool `name:"accounts-allow-custom-css" usage:"Allow accounts to enable custom CSS for their profile pages and statuses."`
	AccountsCustomCSSLength  int  `name:"accounts-custom-css-length" usage:"Maximum permitted length (characters) of custom CSS for accounts."`

	MediaDescriptionMinChars int           `name:"media-description-min-chars" usage:"Min required chars for an image description"`
	MediaDescriptionMaxChars int           `name:"media-description-max-chars" usage:"Max permitted chars for an image description"`
	MediaRemoteCacheDays     int           `name:"media-remote-cache-days" usage:"Number of days to locally cache media from remote instances. If set to 0, remote media will be kept indefinitely."`
	MediaEmojiLocalMaxSize   bytesize.Size `name:"media-emoji-local-max-size" usage:"Max size in bytes of emojis uploaded to this instance via the admin API."`
	MediaEmojiRemoteMaxSize  bytesize.Size `name:"media-emoji-remote-max-size" usage:"Max size in bytes of emojis to download from other instances."`
	MediaLocalMaxSize        bytesize.Size `name:"media-local-max-size" usage:"Max size in bytes of media uploaded to this instance via API"`
	MediaRemoteMaxSize       bytesize.Size `name:"media-remote-max-size" usage:"Max size in bytes of media to download from other instances"`
	MediaCleanupFrom         string        `name:"media-cleanup-from" usage:"Time of day from which to start running media cleanup/prune jobs. Should be in the format 'hh:mm:ss', eg., '15:04:05'."`
	MediaCleanupEvery        time.Duration `name:"media-cleanup-every" usage:"Period to elapse between cleanups, starting from media-cleanup-at."`
	MediaFfmpegPoolSize      int           `name:"media-ffmpeg-pool-size" usage:"Number of instances of the embedded ffmpeg WASM binary to add to the media processing pool. 0 or less uses GOMAXPROCS."`

	StorageBackend       string `name:"storage-backend" usage:"Storage backend to use for media attachments"`
	StorageLocalBasePath string `name:"storage-local-base-path" usage:"Full path to an already-created directory where gts should store/retrieve media files. Subfolders will be created within this dir."`
	StorageS3Endpoint    string `name:"storage-s3-endpoint" usage:"S3 Endpoint URL (e.g 'minio.example.org:9000')"`
	StorageS3AccessKey   string `name:"storage-s3-access-key" usage:"S3 Access Key"`
	StorageS3SecretKey   string `name:"storage-s3-secret-key" usage:"S3 Secret Key"`
	StorageS3UseSSL      bool   `name:"storage-s3-use-ssl" usage:"Use SSL for S3 connections. Only set this to 'false' when testing locally"`
	StorageS3BucketName  string `name:"storage-s3-bucket" usage:"Place blobs in this bucket"`
	StorageS3Proxy       bool   `name:"storage-s3-proxy" usage:"Proxy S3 contents through GoToSocial instead of redirecting to a presigned URL"`
	StorageS3RedirectURL string `name:"storage-s3-redirect-url" usage:"Custom URL to use for redirecting S3 media links. If set, this will be used instead of the S3 bucket URL."`

	StatusesMaxChars           int `name:"statuses-max-chars" usage:"Max permitted characters for posted statuses, including content warning"`
	StatusesPollMaxOptions     int `name:"statuses-poll-max-options" usage:"Max amount of options permitted on a poll"`
	StatusesPollOptionMaxChars int `name:"statuses-poll-option-max-chars" usage:"Max amount of characters for a poll option"`
	StatusesMediaMaxFiles      int `name:"statuses-media-max-files" usage:"Maximum number of media files/attachments per status"`

	LetsEncryptEnabled      bool   `name:"letsencrypt-enabled" usage:"Enable letsencrypt TLS certs for this server. If set to true, then cert dir also needs to be set (or take the default)."`
	LetsEncryptPort         int    `name:"letsencrypt-port" usage:"Port to listen on for letsencrypt certificate challenges. Must not be the same as the GtS webserver/API port."`
	LetsEncryptCertDir      string `name:"letsencrypt-cert-dir" usage:"Directory to store acquired letsencrypt certificates."`
	LetsEncryptEmailAddress string `name:"letsencrypt-email-address" usage:"Email address to use when requesting letsencrypt certs. Will receive updates on cert expiry etc."`

	TLSCertificateChain string `name:"tls-certificate-chain" usage:"Filesystem path to the certificate chain including any intermediate CAs and the TLS public key"`
	TLSCertificateKey   string `name:"tls-certificate-key" usage:"Filesystem path to the TLS private key"`

	OIDCEnabled          bool     `name:"oidc-enabled" usage:"Enabled OIDC authorization for this instance. If set to true, then the other OIDC flags must also be set."`
	OIDCIdpName          string   `name:"oidc-idp-name" usage:"Name of the OIDC identity provider. Will be shown to the user when logging in."`
	OIDCSkipVerification bool     `name:"oidc-skip-verification" usage:"Skip verification of tokens returned by the OIDC provider. Should only be set to 'true' for testing purposes, never in a production environment!"`
	OIDCIssuer           string   `name:"oidc-issuer" usage:"Address of the OIDC issuer. Should be the web address, including protocol, at which the issuer can be reached. Eg., 'https://example.org/auth'"`
	OIDCClientID         string   `name:"oidc-client-id" usage:"ClientID of GoToSocial, as registered with the OIDC provider."`
	OIDCClientSecret     string   `name:"oidc-client-secret" usage:"ClientSecret of GoToSocial, as registered with the OIDC provider."`
	OIDCScopes           []string `name:"oidc-scopes" usage:"OIDC scopes."`
	OIDCLinkExisting     bool     `name:"oidc-link-existing" usage:"link existing user accounts to OIDC logins based on the stored email value"`
	OIDCAllowedGroups    []string `name:"oidc-allowed-groups" usage:"Membership of one of the listed groups allows access to GtS. If this is empty, all groups are allowed."`
	OIDCAdminGroups      []string `name:"oidc-admin-groups" usage:"Membership of one of the listed groups makes someone a GtS admin"`

	TracingEnabled           bool   `name:"tracing-enabled" usage:"Enable OTLP Tracing"`
	TracingTransport         string `name:"tracing-transport" usage:"grpc or http"`
	TracingEndpoint          string `name:"tracing-endpoint" usage:"Endpoint of your trace collector. Eg., 'localhost:4317' for gRPC, 'localhost:4318' for http"`
	TracingInsecureTransport bool   `name:"tracing-insecure-transport" usage:"Disable TLS for the gRPC or HTTP transport protocol"`

	MetricsEnabled      bool   `name:"metrics-enabled" usage:"Enable OpenTelemetry based metrics support."`
	MetricsAuthEnabled  bool   `name:"metrics-auth-enabled" usage:"Enable HTTP Basic Authentication for Prometheus metrics endpoint"`
	MetricsAuthUsername string `name:"metrics-auth-username" usage:"Username for Prometheus metrics endpoint"`
	MetricsAuthPassword string `name:"metrics-auth-password" usage:"Password for Prometheus metrics endpoint"`

	SMTPHost               string `name:"smtp-host" usage:"Host of the smtp server. Eg., 'smtp.eu.mailgun.org'"`
	SMTPPort               int    `name:"smtp-port" usage:"Port of the smtp server. Eg., 587"`
	SMTPUsername           string `name:"smtp-username" usage:"Username to authenticate with the smtp server as. Eg., 'postmaster@mail.example.org'"`
	SMTPPassword           string `name:"smtp-password" usage:"Password to pass to the smtp server."`
	SMTPFrom               string `name:"smtp-from" usage:"Address to use as the 'from' field of the email. Eg., 'gotosocial@example.org'"`
	SMTPDiscloseRecipients bool   `name:"smtp-disclose-recipients" usage:"If true, email notifications sent to multiple recipients will be To'd to every recipient at once. If false, recipients will not be disclosed"`

	SyslogEnabled  bool   `name:"syslog-enabled" usage:"Enable the syslog logging hook. Logs will be mirrored to the configured destination."`
	SyslogProtocol string `name:"syslog-protocol" usage:"Protocol to use when directing logs to syslog. Leave empty to connect to local syslog."`
	SyslogAddress  string `name:"syslog-address" usage:"Address:port to send syslog logs to. Leave empty to connect to local syslog."`

	AdvancedCookiesSamesite      string        `name:"advanced-cookies-samesite" usage:"'strict' or 'lax', see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite"`
	AdvancedRateLimitRequests    int           `name:"advanced-rate-limit-requests" usage:"Amount of HTTP requests to permit within a 5 minute window. 0 or less turns rate limiting off."`
	AdvancedRateLimitExceptions  []string      `name:"advanced-rate-limit-exceptions" usage:"Slice of CIDRs to exclude from rate limit restrictions."`
	AdvancedThrottlingMultiplier int           `name:"advanced-throttling-multiplier" usage:"Multiplier to use per cpu for http request throttling. 0 or less turns throttling off."`
	AdvancedThrottlingRetryAfter time.Duration `name:"advanced-throttling-retry-after" usage:"Retry-After duration response to send for throttled requests."`
	AdvancedSenderMultiplier     int           `name:"advanced-sender-multiplier" usage:"Multiplier to use per cpu for batching outgoing fedi messages. 0 or less turns batching off (not recommended)."`
	AdvancedCSPExtraURIs         []string      `name:"advanced-csp-extra-uris" usage:"Additional URIs to allow when building content-security-policy for media + images."`
	AdvancedHeaderFilterMode     string        `name:"advanced-header-filter-mode" usage:"Set incoming request header filtering mode."`

	// HTTPClient configuration vars.
	HTTPClient HTTPClientConfiguration `name:"http-client"`

	// Cache configuration vars.
	Cache CacheConfiguration `name:"cache"`

	// TODO: move these elsewhere, these are more ephemeral vs long-running flags like above
	AdminAccountUsername     string `name:"username" usage:"the username to create/delete/etc"`
	AdminAccountEmail        string `name:"email" usage:"the email address of this account"`
	AdminAccountPassword     string `name:"password" usage:"the password to set for this account"`
	AdminTransPath           string `name:"path" usage:"the path of the file to import from/export to"`
	AdminMediaPruneDryRun    bool   `name:"dry-run" usage:"perform a dry run and only log number of items eligible for pruning"`
	AdminMediaListLocalOnly  bool   `name:"local-only" usage:"list only local attachments/emojis; if specified then remote-only cannot also be true"`
	AdminMediaListRemoteOnly bool   `name:"remote-only" usage:"list only remote attachments/emojis; if specified then local-only cannot also be true"`

	RequestIDHeader string `name:"request-id-header" usage:"Header to extract the Request ID from. Eg.,'X-Request-Id'."`
}

type HTTPClientConfiguration struct {
	AllowIPs              []string      `name:"allow-ips"`
	BlockIPs              []string      `name:"block-ips"`
	Timeout               time.Duration `name:"timeout"`
	TLSInsecureSkipVerify bool          `name:"tls-insecure-skip-verify"`
}

type CacheConfiguration struct {
	MemoryTarget                      bytesize.Size `name:"memory-target"`
	AccountMemRatio                   float64       `name:"account-mem-ratio"`
	AccountNoteMemRatio               float64       `name:"account-note-mem-ratio"`
	AccountSettingsMemRatio           float64       `name:"account-settings-mem-ratio"`
	AccountStatsMemRatio              float64       `name:"account-stats-mem-ratio"`
	ApplicationMemRatio               float64       `name:"application-mem-ratio"`
	BlockMemRatio                     float64       `name:"block-mem-ratio"`
	BlockIDsMemRatio                  float64       `name:"block-ids-mem-ratio"`
	BoostOfIDsMemRatio                float64       `name:"boost-of-ids-mem-ratio"`
	ClientMemRatio                    float64       `name:"client-mem-ratio"`
	ConversationMemRatio              float64       `name:"conversation-mem-ratio"`
	ConversationLastStatusIDsMemRatio float64       `name:"conversation-last-status-ids-mem-ratio"`
	EmojiMemRatio                     float64       `name:"emoji-mem-ratio"`
	EmojiCategoryMemRatio             float64       `name:"emoji-category-mem-ratio"`
	FilterMemRatio                    float64       `name:"filter-mem-ratio"`
	FilterKeywordMemRatio             float64       `name:"filter-keyword-mem-ratio"`
	FilterStatusMemRatio              float64       `name:"filter-status-mem-ratio"`
	FollowMemRatio                    float64       `name:"follow-mem-ratio"`
	FollowIDsMemRatio                 float64       `name:"follow-ids-mem-ratio"`
	FollowRequestMemRatio             float64       `name:"follow-request-mem-ratio"`
	FollowRequestIDsMemRatio          float64       `name:"follow-request-ids-mem-ratio"`
	FollowingTagIDsMemRatio           float64       `name:"following-tag-ids-mem-ratio"`
	InReplyToIDsMemRatio              float64       `name:"in-reply-to-ids-mem-ratio"`
	InstanceMemRatio                  float64       `name:"instance-mem-ratio"`
	InteractionRequestMemRatio        float64       `name:"interaction-request-mem-ratio"`
	ListMemRatio                      float64       `name:"list-mem-ratio"`
	ListIDsMemRatio                   float64       `name:"list-ids-mem-ratio"`
	ListedIDsMemRatio                 float64       `name:"listed-ids-mem-ratio"`
	MarkerMemRatio                    float64       `name:"marker-mem-ratio"`
	MediaMemRatio                     float64       `name:"media-mem-ratio"`
	MentionMemRatio                   float64       `name:"mention-mem-ratio"`
	MoveMemRatio                      float64       `name:"move-mem-ratio"`
	NotificationMemRatio              float64       `name:"notification-mem-ratio"`
	PollMemRatio                      float64       `name:"poll-mem-ratio"`
	PollVoteMemRatio                  float64       `name:"poll-vote-mem-ratio"`
	PollVoteIDsMemRatio               float64       `name:"poll-vote-ids-mem-ratio"`
	ReportMemRatio                    float64       `name:"report-mem-ratio"`
	SinBinStatusMemRatio              float64       `name:"sin-bin-status-mem-ratio"`
	StatusMemRatio                    float64       `name:"status-mem-ratio"`
	StatusBookmarkMemRatio            float64       `name:"status-bookmark-mem-ratio"`
	StatusBookmarkIDsMemRatio         float64       `name:"status-bookmark-ids-mem-ratio"`
	StatusFaveMemRatio                float64       `name:"status-fave-mem-ratio"`
	StatusFaveIDsMemRatio             float64       `name:"status-fave-ids-mem-ratio"`
	TagMemRatio                       float64       `name:"tag-mem-ratio"`
	ThreadMuteMemRatio                float64       `name:"thread-mute-mem-ratio"`
	TokenMemRatio                     float64       `name:"token-mem-ratio"`
	TombstoneMemRatio                 float64       `name:"tombstone-mem-ratio"`
	UserMemRatio                      float64       `name:"user-mem-ratio"`
	UserMuteMemRatio                  float64       `name:"user-mute-mem-ratio"`
	UserMuteIDsMemRatio               float64       `name:"user-mute-ids-mem-ratio"`
	WebfingerMemRatio                 float64       `name:"webfinger-mem-ratio"`
	VisibilityMemRatio                float64       `name:"visibility-mem-ratio"`
}

// MarshalMap will marshal current Configuration into a map structure (useful for JSON/TOML/YAML).
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
