// THIS IS A GENERATED FILE, DO NOT EDIT BY HAND
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
	"fmt"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/language"
	"codeberg.org/gruf/go-bytesize"
	"github.com/spf13/cast"
	"github.com/spf13/pflag"
)

const (
	LogLevelFlag                                   = "log-level"
	LogFormatFlag                                  = "log-format"
	LogTimestampFormatFlag                         = "log-timestamp-format"
	LogDbQueriesFlag                               = "log-db-queries"
	LogClientIPFlag                                = "log-client-ip"
	RequestIDHeaderFlag                            = "request-id-header"
	ConfigPathFlag                                 = "config-path"
	ApplicationNameFlag                            = "application-name"
	LandingPageUserFlag                            = "landing-page-user"
	HostFlag                                       = "host"
	AccountDomainFlag                              = "account-domain"
	ProtocolFlag                                   = "protocol"
	BindAddressFlag                                = "bind-address"
	PortFlag                                       = "port"
	TrustedProxiesFlag                             = "trusted-proxies"
	SoftwareVersionFlag                            = "software-version"
	DbTypeFlag                                     = "db-type"
	DbAddressFlag                                  = "db-address"
	DbPortFlag                                     = "db-port"
	DbUserFlag                                     = "db-user"
	DbPasswordFlag                                 = "db-password"
	DbDatabaseFlag                                 = "db-database"
	DbTLSModeFlag                                  = "db-tls-mode"
	DbTLSCACertFlag                                = "db-tls-ca-cert"
	DbMaxOpenConnsMultiplierFlag                   = "db-max-open-conns-multiplier"
	DbSqliteJournalModeFlag                        = "db-sqlite-journal-mode"
	DbSqliteSynchronousFlag                        = "db-sqlite-synchronous"
	DbSqliteCacheSizeFlag                          = "db-sqlite-cache-size"
	DbSqliteBusyTimeoutFlag                        = "db-sqlite-busy-timeout"
	DbPostgresConnectionStringFlag                 = "db-postgres-connection-string"
	WebTemplateBaseDirFlag                         = "web-template-base-dir"
	WebAssetBaseDirFlag                            = "web-asset-base-dir"
	InstanceFederationModeFlag                     = "instance-federation-mode"
	InstanceFederationSpamFilterFlag               = "instance-federation-spam-filter"
	InstanceExposePeersFlag                        = "instance-expose-peers"
	InstanceExposeBlocklistFlag                    = "instance-expose-blocklist"
	InstanceExposeBlocklistWebFlag                 = "instance-expose-blocklist-web"
	InstanceExposeAllowlistFlag                    = "instance-expose-allowlist"
	InstanceExposeAllowlistWebFlag                 = "instance-expose-allowlist-web"
	InstanceExposePublicTimelineFlag               = "instance-expose-public-timeline"
	InstanceExposeCustomEmojisFlag                 = "instance-expose-custom-emojis"
	InstanceDeliverToSharedInboxesFlag             = "instance-deliver-to-shared-inboxes"
	InstanceInjectMastodonVersionFlag              = "instance-inject-mastodon-version"
	InstanceLanguagesFlag                          = "instance-languages"
	InstanceSubscriptionsProcessFromFlag           = "instance-subscriptions-process-from"
	InstanceSubscriptionsProcessEveryFlag          = "instance-subscriptions-process-every"
	InstanceStatsModeFlag                          = "instance-stats-mode"
	InstanceAllowBackdatingStatusesFlag            = "instance-allow-backdating-statuses"
	InstanceAllowEmptyUserAgentsFlag               = "instance-allow-empty-user-agents"
	AccountsRegistrationOpenFlag                   = "accounts-registration-open"
	AccountsReasonRequiredFlag                     = "accounts-reason-required"
	AccountsRegistrationDailyLimitFlag             = "accounts-registration-daily-limit"
	AccountsRegistrationBacklogLimitFlag           = "accounts-registration-backlog-limit"
	AccountsAllowCustomCSSFlag                     = "accounts-allow-custom-css"
	AccountsCustomCSSLengthFlag                    = "accounts-custom-css-length"
	AccountsMaxProfileFieldsFlag                   = "accounts-max-profile-fields"
	StorageBackendFlag                             = "storage-backend"
	StorageLocalBasePathFlag                       = "storage-local-base-path"
	StorageS3EndpointFlag                          = "storage-s3-endpoint"
	StorageS3AccessKeyFlag                         = "storage-s3-access-key"
	StorageS3SecretKeyFlag                         = "storage-s3-secret-key"
	StorageS3UseSSLFlag                            = "storage-s3-use-ssl"
	StorageS3BucketNameFlag                        = "storage-s3-bucket"
	StorageS3ProxyFlag                             = "storage-s3-proxy"
	StorageS3RedirectURLFlag                       = "storage-s3-redirect-url"
	StorageS3BucketLookupFlag                      = "storage-s3-bucket-lookup"
	StorageS3KeyPrefixFlag                         = "storage-s3-key-prefix"
	StatusesMaxCharsFlag                           = "statuses-max-chars"
	StatusesPollMaxOptionsFlag                     = "statuses-poll-max-options"
	StatusesPollOptionMaxCharsFlag                 = "statuses-poll-option-max-chars"
	StatusesMediaMaxFilesFlag                      = "statuses-media-max-files"
	ScheduledStatusesMaxTotalFlag                  = "scheduled-statuses-max-total"
	ScheduledStatusesMaxDailyFlag                  = "scheduled-statuses-max-daily"
	LetsEncryptEnabledFlag                         = "letsencrypt-enabled"
	LetsEncryptPortFlag                            = "letsencrypt-port"
	LetsEncryptCertDirFlag                         = "letsencrypt-cert-dir"
	LetsEncryptEmailAddressFlag                    = "letsencrypt-email-address"
	TLSCertificateChainFlag                        = "tls-certificate-chain"
	TLSCertificateKeyFlag                          = "tls-certificate-key"
	OIDCEnabledFlag                                = "oidc-enabled"
	OIDCIdpNameFlag                                = "oidc-idp-name"
	OIDCSkipVerificationFlag                       = "oidc-skip-verification"
	OIDCIssuerFlag                                 = "oidc-issuer"
	OIDCClientIDFlag                               = "oidc-client-id"
	OIDCClientSecretFlag                           = "oidc-client-secret"
	OIDCScopesFlag                                 = "oidc-scopes"
	OIDCLinkExistingFlag                           = "oidc-link-existing"
	OIDCAllowedGroupsFlag                          = "oidc-allowed-groups"
	OIDCAdminGroupsFlag                            = "oidc-admin-groups"
	TracingEnabledFlag                             = "tracing-enabled"
	MetricsEnabledFlag                             = "metrics-enabled"
	SMTPHostFlag                                   = "smtp-host"
	SMTPPortFlag                                   = "smtp-port"
	SMTPUsernameFlag                               = "smtp-username"
	SMTPPasswordFlag                               = "smtp-password"
	SMTPFromFlag                                   = "smtp-from"
	SMTPDiscloseRecipientsFlag                     = "smtp-disclose-recipients"
	SyslogEnabledFlag                              = "syslog-enabled"
	SyslogProtocolFlag                             = "syslog-protocol"
	SyslogAddressFlag                              = "syslog-address"
	AdvancedCookiesSamesiteFlag                    = "advanced-cookies-samesite"
	AdvancedSenderMultiplierFlag                   = "advanced-sender-multiplier"
	AdvancedCSPExtraURIsFlag                       = "advanced-csp-extra-uris"
	AdvancedHeaderFilterModeFlag                   = "advanced-header-filter-mode"
	AdvancedRateLimitRequestsFlag                  = "advanced-rate-limit-requests"
	AdvancedRateLimitExceptionsFlag                = "advanced-rate-limit-exceptions"
	AdvancedThrottlingMultiplierFlag               = "advanced-throttling-multiplier"
	AdvancedThrottlingRetryAfterFlag               = "advanced-throttling-retry-after"
	AdvancedScraperDeterrenceEnabledFlag           = "advanced-scraper-deterrence-enabled"
	AdvancedScraperDeterrenceDifficultyFlag        = "advanced-scraper-deterrence-difficulty"
	HTTPClientAllowIPsFlag                         = "http-client-allow-ips"
	HTTPClientBlockIPsFlag                         = "http-client-block-ips"
	HTTPClientTimeoutFlag                          = "http-client-timeout"
	HTTPClientTLSInsecureSkipVerifyFlag            = "http-client-tls-insecure-skip-verify"
	HTTPClientInsecureOutgoingFlag                 = "http-client-insecure-outgoing"
	MediaDescriptionMinCharsFlag                   = "media-description-min-chars"
	MediaDescriptionMaxCharsFlag                   = "media-description-max-chars"
	MediaRemoteCacheDaysFlag                       = "media-remote-cache-days"
	MediaEmojiLocalMaxSizeFlag                     = "media-emoji-local-max-size"
	MediaEmojiRemoteMaxSizeFlag                    = "media-emoji-remote-max-size"
	MediaImageSizeHintFlag                         = "media-image-size-hint"
	MediaVideoSizeHintFlag                         = "media-video-size-hint"
	MediaLocalMaxSizeFlag                          = "media-local-max-size"
	MediaRemoteMaxSizeFlag                         = "media-remote-max-size"
	MediaCleanupFromFlag                           = "media-cleanup-from"
	MediaCleanupEveryFlag                          = "media-cleanup-every"
	MediaFfmpegPoolSizeFlag                        = "media-ffmpeg-pool-size"
	MediaThumbMaxPixelsFlag                        = "media-thumb-max-pixels"
	CacheMemoryTargetFlag                          = "cache-memory-target"
	CacheAccountMemRatioFlag                       = "cache-account-mem-ratio"
	CacheAccountNoteMemRatioFlag                   = "cache-account-note-mem-ratio"
	CacheAccountSettingsMemRatioFlag               = "cache-account-settings-mem-ratio"
	CacheAccountStatsMemRatioFlag                  = "cache-account-stats-mem-ratio"
	CacheApplicationMemRatioFlag                   = "cache-application-mem-ratio"
	CacheBlockMemRatioFlag                         = "cache-block-mem-ratio"
	CacheBlockIDsMemRatioFlag                      = "cache-block-ids-mem-ratio"
	CacheBoostOfIDsMemRatioFlag                    = "cache-boost-of-ids-mem-ratio"
	CacheClientMemRatioFlag                        = "cache-client-mem-ratio"
	CacheConversationMemRatioFlag                  = "cache-conversation-mem-ratio"
	CacheConversationLastStatusIDsMemRatioFlag     = "cache-conversation-last-status-ids-mem-ratio"
	CacheDomainPermissionDraftMemRationFlag        = "cache-domain-permission-draft-mem-ratio"
	CacheDomainPermissionSubscriptionMemRationFlag = "cache-domain-permission-subscription-mem-ratio"
	CacheEmojiMemRatioFlag                         = "cache-emoji-mem-ratio"
	CacheEmojiCategoryMemRatioFlag                 = "cache-emoji-category-mem-ratio"
	CacheFilterMemRatioFlag                        = "cache-filter-mem-ratio"
	CacheFilterIDsMemRatioFlag                     = "cache-filter-ids-mem-ratio"
	CacheFilterKeywordMemRatioFlag                 = "cache-filter-keyword-mem-ratio"
	CacheFilterStatusMemRatioFlag                  = "cache-filter-status-mem-ratio"
	CacheFollowMemRatioFlag                        = "cache-follow-mem-ratio"
	CacheFollowIDsMemRatioFlag                     = "cache-follow-ids-mem-ratio"
	CacheFollowRequestMemRatioFlag                 = "cache-follow-request-mem-ratio"
	CacheFollowRequestIDsMemRatioFlag              = "cache-follow-request-ids-mem-ratio"
	CacheFollowingTagIDsMemRatioFlag               = "cache-following-tag-ids-mem-ratio"
	CacheInReplyToIDsMemRatioFlag                  = "cache-in-reply-to-ids-mem-ratio"
	CacheInstanceMemRatioFlag                      = "cache-instance-mem-ratio"
	CacheInteractionRequestMemRatioFlag            = "cache-interaction-request-mem-ratio"
	CacheListMemRatioFlag                          = "cache-list-mem-ratio"
	CacheListIDsMemRatioFlag                       = "cache-list-ids-mem-ratio"
	CacheListedIDsMemRatioFlag                     = "cache-listed-ids-mem-ratio"
	CacheMarkerMemRatioFlag                        = "cache-marker-mem-ratio"
	CacheMediaMemRatioFlag                         = "cache-media-mem-ratio"
	CacheMentionMemRatioFlag                       = "cache-mention-mem-ratio"
	CacheMoveMemRatioFlag                          = "cache-move-mem-ratio"
	CacheNotificationMemRatioFlag                  = "cache-notification-mem-ratio"
	CachePollMemRatioFlag                          = "cache-poll-mem-ratio"
	CachePollVoteMemRatioFlag                      = "cache-poll-vote-mem-ratio"
	CachePollVoteIDsMemRatioFlag                   = "cache-poll-vote-ids-mem-ratio"
	CacheReportMemRatioFlag                        = "cache-report-mem-ratio"
	CacheScheduledStatusMemRatioFlag               = "cache-scheduled-status-mem-ratio"
	CacheSinBinStatusMemRatioFlag                  = "cache-sin-bin-status-mem-ratio"
	CacheStatusMemRatioFlag                        = "cache-status-mem-ratio"
	CacheStatusBookmarkMemRatioFlag                = "cache-status-bookmark-mem-ratio"
	CacheStatusBookmarkIDsMemRatioFlag             = "cache-status-bookmark-ids-mem-ratio"
	CacheStatusEditMemRatioFlag                    = "cache-status-edit-mem-ratio"
	CacheStatusFaveMemRatioFlag                    = "cache-status-fave-mem-ratio"
	CacheStatusFaveIDsMemRatioFlag                 = "cache-status-fave-ids-mem-ratio"
	CacheTagMemRatioFlag                           = "cache-tag-mem-ratio"
	CacheThreadMuteMemRatioFlag                    = "cache-thread-mute-mem-ratio"
	CacheTokenMemRatioFlag                         = "cache-token-mem-ratio"
	CacheTombstoneMemRatioFlag                     = "cache-tombstone-mem-ratio"
	CacheUserMemRatioFlag                          = "cache-user-mem-ratio"
	CacheUserMuteMemRatioFlag                      = "cache-user-mute-mem-ratio"
	CacheUserMuteIDsMemRatioFlag                   = "cache-user-mute-ids-mem-ratio"
	CacheWebfingerMemRatioFlag                     = "cache-webfinger-mem-ratio"
	CacheWebPushSubscriptionMemRatioFlag           = "cache-web-push-subscription-mem-ratio"
	CacheWebPushSubscriptionIDsMemRatioFlag        = "cache-web-push-subscription-ids-mem-ratio"
	CacheMutesMemRatioFlag                         = "cache-mutes-mem-ratio"
	CacheStatusFilterMemRatioFlag                  = "cache-status-filter-mem-ratio"
	CacheVisibilityMemRatioFlag                    = "cache-visibility-mem-ratio"
	AdminAccountUsernameFlag                       = "username"
	AdminAccountEmailFlag                          = "email"
	AdminAccountPasswordFlag                       = "password"
	AdminTransPathFlag                             = "path"
	AdminMediaPruneDryRunFlag                      = "dry-run"
	AdminMediaListLocalOnlyFlag                    = "local-only"
	AdminMediaListRemoteOnlyFlag                   = "remote-only"
	TestrigSkipDBSetupFlag                         = "skip-db-setup"
	TestrigSkipDBTeardownFlag                      = "skip-db-teardown"
)

func (cfg *Configuration) RegisterFlags(flags *pflag.FlagSet) {
	flags.String("log-level", cfg.LogLevel, "Log level to run at: [trace, debug, info, warn, fatal]")
	flags.String("log-format", cfg.LogFormat, "Log output format: [logfmt, json]")
	flags.String("log-timestamp-format", cfg.LogTimestampFormat, "Format to use for the log timestamp, as supported by Go's time.Layout")
	flags.Bool("log-db-queries", cfg.LogDbQueries, "Log database queries verbosely when log-level is trace or debug")
	flags.Bool("log-client-ip", cfg.LogClientIP, "Include the client IP in logs")
	flags.String("request-id-header", cfg.RequestIDHeader, "Header to extract the Request ID from. Eg.,'X-Request-Id'.")
	flags.String("config-path", cfg.ConfigPath, "Path to a file containing gotosocial configuration. Values set in this file will be overwritten by values set as env vars or arguments")
	flags.String("application-name", cfg.ApplicationName, "Name of the application, used in various places internally")
	flags.String("landing-page-user", cfg.LandingPageUser, "the user that should be shown on the instance's landing page")
	flags.String("host", cfg.Host, "Hostname to use for the server (eg., example.org, gotosocial.whatever.com). DO NOT change this on a server that's already run!")
	flags.String("account-domain", cfg.AccountDomain, "Domain to use in account names (eg., example.org, whatever.com). If not set, will default to the setting for host. DO NOT change this on a server that's already run!")
	flags.String("protocol", cfg.Protocol, "Protocol to use for the REST api of the server (only use http if you are debugging; https should be used even if running behind a reverse proxy!)")
	flags.String("bind-address", cfg.BindAddress, "Bind address to use for the GoToSocial server (eg., 0.0.0.0, 172.138.0.9, [::], localhost). For ipv6, enclose the address in square brackets, eg [2001:db8::fed1]. Default binds to all interfaces.")
	flags.Int("port", cfg.Port, "Port to use for GoToSocial. Change this to 443 if you're running the binary directly on the host machine.")
	flags.StringSlice("trusted-proxies", cfg.TrustedProxies, "Proxies to trust when parsing x-forwarded headers into real IPs.")
	flags.String("software-version", cfg.SoftwareVersion, "")
	flags.String("db-type", cfg.DbType, "Database type: eg., postgres")
	flags.String("db-address", cfg.DbAddress, "Database ipv4 address, hostname, or filename")
	flags.Int("db-port", cfg.DbPort, "Database port")
	flags.String("db-user", cfg.DbUser, "Database username")
	flags.String("db-password", cfg.DbPassword, "Database password")
	flags.String("db-database", cfg.DbDatabase, "Database name")
	flags.String("db-tls-mode", cfg.DbTLSMode, "Database tls mode")
	flags.String("db-tls-ca-cert", cfg.DbTLSCACert, "Path to CA cert for db tls connection")
	flags.Int("db-max-open-conns-multiplier", cfg.DbMaxOpenConnsMultiplier, "Multiplier to use per cpu for max open database connections. 0 or less is normalized to 1.")
	flags.String("db-sqlite-journal-mode", cfg.DbSqliteJournalMode, "Sqlite only: see https://www.sqlite.org/pragma.html#pragma_journal_mode")
	flags.String("db-sqlite-synchronous", cfg.DbSqliteSynchronous, "Sqlite only: see https://www.sqlite.org/pragma.html#pragma_synchronous")
	flags.String("db-sqlite-cache-size", cfg.DbSqliteCacheSize.String(), "Sqlite only: see https://www.sqlite.org/pragma.html#pragma_cache_size")
	flags.Duration("db-sqlite-busy-timeout", cfg.DbSqliteBusyTimeout, "Sqlite only: see https://www.sqlite.org/pragma.html#pragma_busy_timeout")
	flags.String("db-postgres-connection-string", cfg.DbPostgresConnectionString, "Full Database URL for connection to postgres")
	flags.String("web-template-base-dir", cfg.WebTemplateBaseDir, "Basedir for html templating files for rendering pages and composing emails.")
	flags.String("web-asset-base-dir", cfg.WebAssetBaseDir, "Directory to serve static assets from, accessible at example.org/assets/")
	flags.String("instance-federation-mode", cfg.InstanceFederationMode, "Set instance federation mode.")
	flags.Bool("instance-federation-spam-filter", cfg.InstanceFederationSpamFilter, "Enable basic spam filter heuristics for messages coming from other instances, and drop messages identified as spam")
	flags.Bool("instance-expose-peers", cfg.InstanceExposePeers, "Allow unauthenticated users to query /api/v1/instance/peers?filter=open")
	flags.Bool("instance-expose-blocklist", cfg.InstanceExposeBlocklist, "Expose list of blocked domains via web UI, and allow unauthenticated users to query /api/v1/instance/peers?filter=blocked and /api/v1/instance/domain_blocks")
	flags.Bool("instance-expose-blocklist-web", cfg.InstanceExposeBlocklistWeb, "Expose list of explicitly blocked domains as webpage on /about/domain_blocks")
	flags.Bool("instance-expose-allowlist", cfg.InstanceExposeAllowlist, "Expose list of allowed domains via web UI, and allow unauthenticated users to query /api/v1/instance/peers?filter=allowed and /api/v1/instance/domain_allows")
	flags.Bool("instance-expose-allowlist-web", cfg.InstanceExposeAllowlistWeb, "Expose list of explicitly allowed domains as webpage on /about/domain_allows")
	flags.Bool("instance-expose-public-timeline", cfg.InstanceExposePublicTimeline, "Allow unauthenticated users to query /api/v1/timelines/public")
	flags.Bool("instance-expose-custom-emojis", cfg.InstanceExposeCustomEmojis, "Allow unauthenticated access to /api/v1/custom_emojis")
	flags.Bool("instance-deliver-to-shared-inboxes", cfg.InstanceDeliverToSharedInboxes, "Deliver federated messages to shared inboxes, if they're available.")
	flags.Bool("instance-inject-mastodon-version", cfg.InstanceInjectMastodonVersion, "This injects a Mastodon compatible version in /api/v1/instance to help Mastodon clients that use that version for feature detection")
	flags.StringSlice("instance-languages", cfg.InstanceLanguages.Strings(), "BCP47 language tags for the instance. Used to indicate the preferred languages of instance residents (in order from most-preferred to least-preferred).")
	flags.String("instance-subscriptions-process-from", cfg.InstanceSubscriptionsProcessFrom, "Time of day from which to start running instance subscriptions processing jobs. Should be in the format 'hh:mm:ss', eg., '15:04:05'.")
	flags.Duration("instance-subscriptions-process-every", cfg.InstanceSubscriptionsProcessEvery, "Period to elapse between instance subscriptions processing jobs, starting from instance-subscriptions-process-from.")
	flags.String("instance-stats-mode", cfg.InstanceStatsMode, "Allows you to customize the way stats are served to crawlers: one of '', 'serve', 'zero', 'baffle'. Home page stats remain unchanged.")
	flags.Bool("instance-allow-backdating-statuses", cfg.InstanceAllowBackdatingStatuses, "Allow local accounts to backdate statuses using the scheduled_at param to /api/v1/statuses")
	flags.Bool("instance-allow-empty-user-agents", cfg.InstanceAllowEmptyUserAgents, "Allow incoming HTTP requests that do not have a User-Agent header set")
	flags.Bool("accounts-registration-open", cfg.AccountsRegistrationOpen, "Allow anyone to submit an account signup request. If false, server will be invite-only.")
	flags.Bool("accounts-reason-required", cfg.AccountsReasonRequired, "Do new account signups require a reason to be submitted on registration?")
	flags.Int("accounts-registration-daily-limit", cfg.AccountsRegistrationDailyLimit, "Limit amount of approved account sign-ups allowed per 24hrs before registration is closed. 0 or less = no limit.")
	flags.Int("accounts-registration-backlog-limit", cfg.AccountsRegistrationBacklogLimit, "Limit how big the 'accounts pending approval' queue can grow before registration is closed. 0 or less = no limit.")
	flags.Bool("accounts-allow-custom-css", cfg.AccountsAllowCustomCSS, "Allow accounts to enable custom CSS for their profile pages and statuses.")
	flags.Int("accounts-custom-css-length", cfg.AccountsCustomCSSLength, "Maximum permitted length (characters) of custom CSS for accounts.")
	flags.Int("accounts-max-profile-fields", cfg.AccountsMaxProfileFields, "Maximum number of profile fields allowed for each account.")
	flags.String("storage-backend", cfg.StorageBackend, "Storage backend to use for media attachments")
	flags.String("storage-local-base-path", cfg.StorageLocalBasePath, "Full path to an already-created directory where gts should store/retrieve media files. Subfolders will be created within this dir.")
	flags.String("storage-s3-endpoint", cfg.StorageS3Endpoint, "S3 Endpoint URL (e.g 'minio.example.org:9000')")
	flags.String("storage-s3-access-key", cfg.StorageS3AccessKey, "S3 Access Key")
	flags.String("storage-s3-secret-key", cfg.StorageS3SecretKey, "S3 Secret Key")
	flags.Bool("storage-s3-use-ssl", cfg.StorageS3UseSSL, "Use SSL for S3 connections. Only set this to 'false' when testing locally")
	flags.String("storage-s3-bucket", cfg.StorageS3BucketName, "Place blobs in this bucket")
	flags.Bool("storage-s3-proxy", cfg.StorageS3Proxy, "Proxy S3 contents through GoToSocial instead of redirecting to a presigned URL")
	flags.String("storage-s3-redirect-url", cfg.StorageS3RedirectURL, "Custom URL to use for redirecting S3 media links. If set, this will be used instead of the S3 bucket URL.")
	flags.String("storage-s3-bucket-lookup", cfg.StorageS3BucketLookup, "S3 bucket lookup type to use. Can be 'auto', 'dns' or 'path'. Defaults to 'auto'.")
	flags.String("storage-s3-key-prefix", cfg.StorageS3KeyPrefix, "Prefix to use for S3 keys. This is useful for separating multiple instances sharing the same S3 bucket.")
	flags.Int("statuses-max-chars", cfg.StatusesMaxChars, "Max permitted characters for posted statuses, including content warning")
	flags.Int("statuses-poll-max-options", cfg.StatusesPollMaxOptions, "Max amount of options permitted on a poll")
	flags.Int("statuses-poll-option-max-chars", cfg.StatusesPollOptionMaxChars, "Max amount of characters for a poll option")
	flags.Int("statuses-media-max-files", cfg.StatusesMediaMaxFiles, "Maximum number of media files/attachments per status")
	flags.Int("scheduled-statuses-max-total", cfg.ScheduledStatusesMaxTotal, "Maximum number of scheduled statuses per user")
	flags.Int("scheduled-statuses-max-daily", cfg.ScheduledStatusesMaxDaily, "Maximum number of scheduled statuses per user for a single day")
	flags.Bool("letsencrypt-enabled", cfg.LetsEncryptEnabled, "Enable letsencrypt TLS certs for this server. If set to true, then cert dir also needs to be set (or take the default).")
	flags.Int("letsencrypt-port", cfg.LetsEncryptPort, "Port to listen on for letsencrypt certificate challenges. Must not be the same as the GtS webserver/API port.")
	flags.String("letsencrypt-cert-dir", cfg.LetsEncryptCertDir, "Directory to store acquired letsencrypt certificates.")
	flags.String("letsencrypt-email-address", cfg.LetsEncryptEmailAddress, "Email address to use when requesting letsencrypt certs. Will receive updates on cert expiry etc.")
	flags.String("tls-certificate-chain", cfg.TLSCertificateChain, "Filesystem path to the certificate chain including any intermediate CAs and the TLS public key")
	flags.String("tls-certificate-key", cfg.TLSCertificateKey, "Filesystem path to the TLS private key")
	flags.Bool("oidc-enabled", cfg.OIDCEnabled, "Enabled OIDC authorization for this instance. If set to true, then the other OIDC flags must also be set.")
	flags.String("oidc-idp-name", cfg.OIDCIdpName, "Name of the OIDC identity provider. Will be shown to the user when logging in.")
	flags.Bool("oidc-skip-verification", cfg.OIDCSkipVerification, "Skip verification of tokens returned by the OIDC provider. Should only be set to 'true' for testing purposes, never in a production environment!")
	flags.String("oidc-issuer", cfg.OIDCIssuer, "Address of the OIDC issuer. Should be the web address, including protocol, at which the issuer can be reached. Eg., 'https://example.org/auth'")
	flags.String("oidc-client-id", cfg.OIDCClientID, "ClientID of GoToSocial, as registered with the OIDC provider.")
	flags.String("oidc-client-secret", cfg.OIDCClientSecret, "ClientSecret of GoToSocial, as registered with the OIDC provider.")
	flags.StringSlice("oidc-scopes", cfg.OIDCScopes, "OIDC scopes.")
	flags.Bool("oidc-link-existing", cfg.OIDCLinkExisting, "link existing user accounts to OIDC logins based on the stored email value")
	flags.StringSlice("oidc-allowed-groups", cfg.OIDCAllowedGroups, "Membership of one of the listed groups allows access to GtS. If this is empty, all groups are allowed.")
	flags.StringSlice("oidc-admin-groups", cfg.OIDCAdminGroups, "Membership of one of the listed groups makes someone a GtS admin")
	flags.Bool("tracing-enabled", cfg.TracingEnabled, "Enable OTLP Tracing")
	flags.Bool("metrics-enabled", cfg.MetricsEnabled, "Enable OpenTelemetry based metrics support.")
	flags.String("smtp-host", cfg.SMTPHost, "Host of the smtp server. Eg., 'smtp.eu.mailgun.org'")
	flags.Int("smtp-port", cfg.SMTPPort, "Port of the smtp server. Eg., 587")
	flags.String("smtp-username", cfg.SMTPUsername, "Username to authenticate with the smtp server as. Eg., 'postmaster@mail.example.org'")
	flags.String("smtp-password", cfg.SMTPPassword, "Password to pass to the smtp server.")
	flags.String("smtp-from", cfg.SMTPFrom, "Address to use as the 'from' field of the email. Eg., 'gotosocial@example.org'")
	flags.Bool("smtp-disclose-recipients", cfg.SMTPDiscloseRecipients, "If true, email notifications sent to multiple recipients will be To'd to every recipient at once. If false, recipients will not be disclosed")
	flags.Bool("syslog-enabled", cfg.SyslogEnabled, "Enable the syslog logging hook. Logs will be mirrored to the configured destination.")
	flags.String("syslog-protocol", cfg.SyslogProtocol, "Protocol to use when directing logs to syslog. Leave empty to connect to local syslog.")
	flags.String("syslog-address", cfg.SyslogAddress, "Address:port to send syslog logs to. Leave empty to connect to local syslog.")
	flags.String("advanced-cookies-samesite", cfg.Advanced.CookiesSamesite, "'strict' or 'lax', see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite")
	flags.Int("advanced-sender-multiplier", cfg.Advanced.SenderMultiplier, "Multiplier to use per cpu for batching outgoing fedi messages. 0 or less turns batching off (not recommended).")
	flags.StringSlice("advanced-csp-extra-uris", cfg.Advanced.CSPExtraURIs, "Additional URIs to allow when building content-security-policy for media + images.")
	flags.String("advanced-header-filter-mode", cfg.Advanced.HeaderFilterMode, "Set incoming request header filtering mode.")
	flags.Int("advanced-rate-limit-requests", cfg.Advanced.RateLimit.Requests, "Amount of HTTP requests to permit within a 5 minute window. 0 or less turns rate limiting off.")
	flags.StringSlice("advanced-rate-limit-exceptions", cfg.Advanced.RateLimit.Exceptions.Strings(), "Slice of CIDRs to exclude from rate limit restrictions.")
	flags.Int("advanced-throttling-multiplier", cfg.Advanced.Throttling.Multiplier, "Multiplier to use per cpu for http request throttling. 0 or less turns throttling off.")
	flags.Duration("advanced-throttling-retry-after", cfg.Advanced.Throttling.RetryAfter, "Retry-After duration response to send for throttled requests.")
	flags.Bool("advanced-scraper-deterrence-enabled", cfg.Advanced.ScraperDeterrence.Enabled, "Enable proof-of-work based scraper deterrence on profile / status pages")
	flags.Uint32("advanced-scraper-deterrence-difficulty", cfg.Advanced.ScraperDeterrence.Difficulty, "The proof-of-work difficulty, which determines roughly how many hash-encode rounds required of each client.")
	flags.StringSlice("http-client-allow-ips", cfg.HTTPClient.AllowIPs, "")
	flags.StringSlice("http-client-block-ips", cfg.HTTPClient.BlockIPs, "")
	flags.Duration("http-client-timeout", cfg.HTTPClient.Timeout, "")
	flags.Bool("http-client-tls-insecure-skip-verify", cfg.HTTPClient.TLSInsecureSkipVerify, "")
	flags.Bool("http-client-insecure-outgoing", cfg.HTTPClient.InsecureOutgoing, "")
	flags.Int("media-description-min-chars", cfg.Media.DescriptionMinChars, "Min required chars for an image description")
	flags.Int("media-description-max-chars", cfg.Media.DescriptionMaxChars, "Max permitted chars for an image description")
	flags.Int("media-remote-cache-days", cfg.Media.RemoteCacheDays, "Number of days to locally cache media from remote instances. If set to 0, remote media will be kept indefinitely.")
	flags.String("media-emoji-local-max-size", cfg.Media.EmojiLocalMaxSize.String(), "Max size in bytes of emojis uploaded to this instance via the admin API.")
	flags.String("media-emoji-remote-max-size", cfg.Media.EmojiRemoteMaxSize.String(), "Max size in bytes of emojis to download from other instances.")
	flags.String("media-image-size-hint", cfg.Media.ImageSizeHint.String(), "Size in bytes of max image size referred to on /api/v_/instance endpoints (else, local max size)")
	flags.String("media-video-size-hint", cfg.Media.VideoSizeHint.String(), "Size in bytes of max video size referred to on /api/v_/instance endpoints (else, local max size)")
	flags.String("media-local-max-size", cfg.Media.LocalMaxSize.String(), "Max size in bytes of media uploaded to this instance via API")
	flags.String("media-remote-max-size", cfg.Media.RemoteMaxSize.String(), "Max size in bytes of media to download from other instances")
	flags.String("media-cleanup-from", cfg.Media.CleanupFrom, "Time of day from which to start running media cleanup/prune jobs. Should be in the format 'hh:mm:ss', eg., '15:04:05'.")
	flags.Duration("media-cleanup-every", cfg.Media.CleanupEvery, "Period to elapse between cleanups, starting from media-cleanup-at.")
	flags.Int("media-ffmpeg-pool-size", cfg.Media.FfmpegPoolSize, "Number of instances of the embedded ffmpeg WASM binary to add to the media processing pool. 0 or less uses GOMAXPROCS.")
	flags.Int("media-thumb-max-pixels", cfg.Media.ThumbMaxPixels, "Max size in pixels of any one dimension of a thumbnail (as input media ratio is preserved).")
	flags.String("cache-memory-target", cfg.Cache.MemoryTarget.String(), "")
	flags.Float64("cache-account-mem-ratio", cfg.Cache.AccountMemRatio, "")
	flags.Float64("cache-account-note-mem-ratio", cfg.Cache.AccountNoteMemRatio, "")
	flags.Float64("cache-account-settings-mem-ratio", cfg.Cache.AccountSettingsMemRatio, "")
	flags.Float64("cache-account-stats-mem-ratio", cfg.Cache.AccountStatsMemRatio, "")
	flags.Float64("cache-application-mem-ratio", cfg.Cache.ApplicationMemRatio, "")
	flags.Float64("cache-block-mem-ratio", cfg.Cache.BlockMemRatio, "")
	flags.Float64("cache-block-ids-mem-ratio", cfg.Cache.BlockIDsMemRatio, "")
	flags.Float64("cache-boost-of-ids-mem-ratio", cfg.Cache.BoostOfIDsMemRatio, "")
	flags.Float64("cache-client-mem-ratio", cfg.Cache.ClientMemRatio, "")
	flags.Float64("cache-conversation-mem-ratio", cfg.Cache.ConversationMemRatio, "")
	flags.Float64("cache-conversation-last-status-ids-mem-ratio", cfg.Cache.ConversationLastStatusIDsMemRatio, "")
	flags.Float64("cache-domain-permission-draft-mem-ratio", cfg.Cache.DomainPermissionDraftMemRation, "")
	flags.Float64("cache-domain-permission-subscription-mem-ratio", cfg.Cache.DomainPermissionSubscriptionMemRation, "")
	flags.Float64("cache-emoji-mem-ratio", cfg.Cache.EmojiMemRatio, "")
	flags.Float64("cache-emoji-category-mem-ratio", cfg.Cache.EmojiCategoryMemRatio, "")
	flags.Float64("cache-filter-mem-ratio", cfg.Cache.FilterMemRatio, "")
	flags.Float64("cache-filter-ids-mem-ratio", cfg.Cache.FilterIDsMemRatio, "")
	flags.Float64("cache-filter-keyword-mem-ratio", cfg.Cache.FilterKeywordMemRatio, "")
	flags.Float64("cache-filter-status-mem-ratio", cfg.Cache.FilterStatusMemRatio, "")
	flags.Float64("cache-follow-mem-ratio", cfg.Cache.FollowMemRatio, "")
	flags.Float64("cache-follow-ids-mem-ratio", cfg.Cache.FollowIDsMemRatio, "")
	flags.Float64("cache-follow-request-mem-ratio", cfg.Cache.FollowRequestMemRatio, "")
	flags.Float64("cache-follow-request-ids-mem-ratio", cfg.Cache.FollowRequestIDsMemRatio, "")
	flags.Float64("cache-following-tag-ids-mem-ratio", cfg.Cache.FollowingTagIDsMemRatio, "")
	flags.Float64("cache-in-reply-to-ids-mem-ratio", cfg.Cache.InReplyToIDsMemRatio, "")
	flags.Float64("cache-instance-mem-ratio", cfg.Cache.InstanceMemRatio, "")
	flags.Float64("cache-interaction-request-mem-ratio", cfg.Cache.InteractionRequestMemRatio, "")
	flags.Float64("cache-list-mem-ratio", cfg.Cache.ListMemRatio, "")
	flags.Float64("cache-list-ids-mem-ratio", cfg.Cache.ListIDsMemRatio, "")
	flags.Float64("cache-listed-ids-mem-ratio", cfg.Cache.ListedIDsMemRatio, "")
	flags.Float64("cache-marker-mem-ratio", cfg.Cache.MarkerMemRatio, "")
	flags.Float64("cache-media-mem-ratio", cfg.Cache.MediaMemRatio, "")
	flags.Float64("cache-mention-mem-ratio", cfg.Cache.MentionMemRatio, "")
	flags.Float64("cache-move-mem-ratio", cfg.Cache.MoveMemRatio, "")
	flags.Float64("cache-notification-mem-ratio", cfg.Cache.NotificationMemRatio, "")
	flags.Float64("cache-poll-mem-ratio", cfg.Cache.PollMemRatio, "")
	flags.Float64("cache-poll-vote-mem-ratio", cfg.Cache.PollVoteMemRatio, "")
	flags.Float64("cache-poll-vote-ids-mem-ratio", cfg.Cache.PollVoteIDsMemRatio, "")
	flags.Float64("cache-report-mem-ratio", cfg.Cache.ReportMemRatio, "")
	flags.Float64("cache-scheduled-status-mem-ratio", cfg.Cache.ScheduledStatusMemRatio, "")
	flags.Float64("cache-sin-bin-status-mem-ratio", cfg.Cache.SinBinStatusMemRatio, "")
	flags.Float64("cache-status-mem-ratio", cfg.Cache.StatusMemRatio, "")
	flags.Float64("cache-status-bookmark-mem-ratio", cfg.Cache.StatusBookmarkMemRatio, "")
	flags.Float64("cache-status-bookmark-ids-mem-ratio", cfg.Cache.StatusBookmarkIDsMemRatio, "")
	flags.Float64("cache-status-edit-mem-ratio", cfg.Cache.StatusEditMemRatio, "")
	flags.Float64("cache-status-fave-mem-ratio", cfg.Cache.StatusFaveMemRatio, "")
	flags.Float64("cache-status-fave-ids-mem-ratio", cfg.Cache.StatusFaveIDsMemRatio, "")
	flags.Float64("cache-tag-mem-ratio", cfg.Cache.TagMemRatio, "")
	flags.Float64("cache-thread-mute-mem-ratio", cfg.Cache.ThreadMuteMemRatio, "")
	flags.Float64("cache-token-mem-ratio", cfg.Cache.TokenMemRatio, "")
	flags.Float64("cache-tombstone-mem-ratio", cfg.Cache.TombstoneMemRatio, "")
	flags.Float64("cache-user-mem-ratio", cfg.Cache.UserMemRatio, "")
	flags.Float64("cache-user-mute-mem-ratio", cfg.Cache.UserMuteMemRatio, "")
	flags.Float64("cache-user-mute-ids-mem-ratio", cfg.Cache.UserMuteIDsMemRatio, "")
	flags.Float64("cache-webfinger-mem-ratio", cfg.Cache.WebfingerMemRatio, "")
	flags.Float64("cache-web-push-subscription-mem-ratio", cfg.Cache.WebPushSubscriptionMemRatio, "")
	flags.Float64("cache-web-push-subscription-ids-mem-ratio", cfg.Cache.WebPushSubscriptionIDsMemRatio, "")
	flags.Float64("cache-mutes-mem-ratio", cfg.Cache.MutesMemRatio, "")
	flags.Float64("cache-status-filter-mem-ratio", cfg.Cache.StatusFilterMemRatio, "")
	flags.Float64("cache-visibility-mem-ratio", cfg.Cache.VisibilityMemRatio, "")
}

func (cfg *Configuration) MarshalMap() map[string]any {
	cfgmap := make(map[string]any, 198)
	cfgmap["log-level"] = cfg.LogLevel
	cfgmap["log-format"] = cfg.LogFormat
	cfgmap["log-timestamp-format"] = cfg.LogTimestampFormat
	cfgmap["log-db-queries"] = cfg.LogDbQueries
	cfgmap["log-client-ip"] = cfg.LogClientIP
	cfgmap["request-id-header"] = cfg.RequestIDHeader
	cfgmap["config-path"] = cfg.ConfigPath
	cfgmap["application-name"] = cfg.ApplicationName
	cfgmap["landing-page-user"] = cfg.LandingPageUser
	cfgmap["host"] = cfg.Host
	cfgmap["account-domain"] = cfg.AccountDomain
	cfgmap["protocol"] = cfg.Protocol
	cfgmap["bind-address"] = cfg.BindAddress
	cfgmap["port"] = cfg.Port
	cfgmap["trusted-proxies"] = cfg.TrustedProxies
	cfgmap["software-version"] = cfg.SoftwareVersion
	cfgmap["db-type"] = cfg.DbType
	cfgmap["db-address"] = cfg.DbAddress
	cfgmap["db-port"] = cfg.DbPort
	cfgmap["db-user"] = cfg.DbUser
	cfgmap["db-password"] = cfg.DbPassword
	cfgmap["db-database"] = cfg.DbDatabase
	cfgmap["db-tls-mode"] = cfg.DbTLSMode
	cfgmap["db-tls-ca-cert"] = cfg.DbTLSCACert
	cfgmap["db-max-open-conns-multiplier"] = cfg.DbMaxOpenConnsMultiplier
	cfgmap["db-sqlite-journal-mode"] = cfg.DbSqliteJournalMode
	cfgmap["db-sqlite-synchronous"] = cfg.DbSqliteSynchronous
	cfgmap["db-sqlite-cache-size"] = cfg.DbSqliteCacheSize.String()
	cfgmap["db-sqlite-busy-timeout"] = cfg.DbSqliteBusyTimeout
	cfgmap["db-postgres-connection-string"] = cfg.DbPostgresConnectionString
	cfgmap["web-template-base-dir"] = cfg.WebTemplateBaseDir
	cfgmap["web-asset-base-dir"] = cfg.WebAssetBaseDir
	cfgmap["instance-federation-mode"] = cfg.InstanceFederationMode
	cfgmap["instance-federation-spam-filter"] = cfg.InstanceFederationSpamFilter
	cfgmap["instance-expose-peers"] = cfg.InstanceExposePeers
	cfgmap["instance-expose-blocklist"] = cfg.InstanceExposeBlocklist
	cfgmap["instance-expose-blocklist-web"] = cfg.InstanceExposeBlocklistWeb
	cfgmap["instance-expose-allowlist"] = cfg.InstanceExposeAllowlist
	cfgmap["instance-expose-allowlist-web"] = cfg.InstanceExposeAllowlistWeb
	cfgmap["instance-expose-public-timeline"] = cfg.InstanceExposePublicTimeline
	cfgmap["instance-expose-custom-emojis"] = cfg.InstanceExposeCustomEmojis
	cfgmap["instance-deliver-to-shared-inboxes"] = cfg.InstanceDeliverToSharedInboxes
	cfgmap["instance-inject-mastodon-version"] = cfg.InstanceInjectMastodonVersion
	cfgmap["instance-languages"] = cfg.InstanceLanguages.Strings()
	cfgmap["instance-subscriptions-process-from"] = cfg.InstanceSubscriptionsProcessFrom
	cfgmap["instance-subscriptions-process-every"] = cfg.InstanceSubscriptionsProcessEvery
	cfgmap["instance-stats-mode"] = cfg.InstanceStatsMode
	cfgmap["instance-allow-backdating-statuses"] = cfg.InstanceAllowBackdatingStatuses
	cfgmap["instance-allow-empty-user-agents"] = cfg.InstanceAllowEmptyUserAgents
	cfgmap["accounts-registration-open"] = cfg.AccountsRegistrationOpen
	cfgmap["accounts-reason-required"] = cfg.AccountsReasonRequired
	cfgmap["accounts-registration-daily-limit"] = cfg.AccountsRegistrationDailyLimit
	cfgmap["accounts-registration-backlog-limit"] = cfg.AccountsRegistrationBacklogLimit
	cfgmap["accounts-allow-custom-css"] = cfg.AccountsAllowCustomCSS
	cfgmap["accounts-custom-css-length"] = cfg.AccountsCustomCSSLength
	cfgmap["accounts-max-profile-fields"] = cfg.AccountsMaxProfileFields
	cfgmap["storage-backend"] = cfg.StorageBackend
	cfgmap["storage-local-base-path"] = cfg.StorageLocalBasePath
	cfgmap["storage-s3-endpoint"] = cfg.StorageS3Endpoint
	cfgmap["storage-s3-access-key"] = cfg.StorageS3AccessKey
	cfgmap["storage-s3-secret-key"] = cfg.StorageS3SecretKey
	cfgmap["storage-s3-use-ssl"] = cfg.StorageS3UseSSL
	cfgmap["storage-s3-bucket"] = cfg.StorageS3BucketName
	cfgmap["storage-s3-proxy"] = cfg.StorageS3Proxy
	cfgmap["storage-s3-redirect-url"] = cfg.StorageS3RedirectURL
	cfgmap["storage-s3-bucket-lookup"] = cfg.StorageS3BucketLookup
	cfgmap["storage-s3-key-prefix"] = cfg.StorageS3KeyPrefix
	cfgmap["statuses-max-chars"] = cfg.StatusesMaxChars
	cfgmap["statuses-poll-max-options"] = cfg.StatusesPollMaxOptions
	cfgmap["statuses-poll-option-max-chars"] = cfg.StatusesPollOptionMaxChars
	cfgmap["statuses-media-max-files"] = cfg.StatusesMediaMaxFiles
	cfgmap["scheduled-statuses-max-total"] = cfg.ScheduledStatusesMaxTotal
	cfgmap["scheduled-statuses-max-daily"] = cfg.ScheduledStatusesMaxDaily
	cfgmap["letsencrypt-enabled"] = cfg.LetsEncryptEnabled
	cfgmap["letsencrypt-port"] = cfg.LetsEncryptPort
	cfgmap["letsencrypt-cert-dir"] = cfg.LetsEncryptCertDir
	cfgmap["letsencrypt-email-address"] = cfg.LetsEncryptEmailAddress
	cfgmap["tls-certificate-chain"] = cfg.TLSCertificateChain
	cfgmap["tls-certificate-key"] = cfg.TLSCertificateKey
	cfgmap["oidc-enabled"] = cfg.OIDCEnabled
	cfgmap["oidc-idp-name"] = cfg.OIDCIdpName
	cfgmap["oidc-skip-verification"] = cfg.OIDCSkipVerification
	cfgmap["oidc-issuer"] = cfg.OIDCIssuer
	cfgmap["oidc-client-id"] = cfg.OIDCClientID
	cfgmap["oidc-client-secret"] = cfg.OIDCClientSecret
	cfgmap["oidc-scopes"] = cfg.OIDCScopes
	cfgmap["oidc-link-existing"] = cfg.OIDCLinkExisting
	cfgmap["oidc-allowed-groups"] = cfg.OIDCAllowedGroups
	cfgmap["oidc-admin-groups"] = cfg.OIDCAdminGroups
	cfgmap["tracing-enabled"] = cfg.TracingEnabled
	cfgmap["metrics-enabled"] = cfg.MetricsEnabled
	cfgmap["smtp-host"] = cfg.SMTPHost
	cfgmap["smtp-port"] = cfg.SMTPPort
	cfgmap["smtp-username"] = cfg.SMTPUsername
	cfgmap["smtp-password"] = cfg.SMTPPassword
	cfgmap["smtp-from"] = cfg.SMTPFrom
	cfgmap["smtp-disclose-recipients"] = cfg.SMTPDiscloseRecipients
	cfgmap["syslog-enabled"] = cfg.SyslogEnabled
	cfgmap["syslog-protocol"] = cfg.SyslogProtocol
	cfgmap["syslog-address"] = cfg.SyslogAddress
	cfgmap["advanced-cookies-samesite"] = cfg.Advanced.CookiesSamesite
	cfgmap["advanced-sender-multiplier"] = cfg.Advanced.SenderMultiplier
	cfgmap["advanced-csp-extra-uris"] = cfg.Advanced.CSPExtraURIs
	cfgmap["advanced-header-filter-mode"] = cfg.Advanced.HeaderFilterMode
	cfgmap["advanced-rate-limit-requests"] = cfg.Advanced.RateLimit.Requests
	cfgmap["advanced-rate-limit-exceptions"] = cfg.Advanced.RateLimit.Exceptions.Strings()
	cfgmap["advanced-throttling-multiplier"] = cfg.Advanced.Throttling.Multiplier
	cfgmap["advanced-throttling-retry-after"] = cfg.Advanced.Throttling.RetryAfter
	cfgmap["advanced-scraper-deterrence-enabled"] = cfg.Advanced.ScraperDeterrence.Enabled
	cfgmap["advanced-scraper-deterrence-difficulty"] = cfg.Advanced.ScraperDeterrence.Difficulty
	cfgmap["http-client-allow-ips"] = cfg.HTTPClient.AllowIPs
	cfgmap["http-client-block-ips"] = cfg.HTTPClient.BlockIPs
	cfgmap["http-client-timeout"] = cfg.HTTPClient.Timeout
	cfgmap["http-client-tls-insecure-skip-verify"] = cfg.HTTPClient.TLSInsecureSkipVerify
	cfgmap["http-client-insecure-outgoing"] = cfg.HTTPClient.InsecureOutgoing
	cfgmap["media-description-min-chars"] = cfg.Media.DescriptionMinChars
	cfgmap["media-description-max-chars"] = cfg.Media.DescriptionMaxChars
	cfgmap["media-remote-cache-days"] = cfg.Media.RemoteCacheDays
	cfgmap["media-emoji-local-max-size"] = cfg.Media.EmojiLocalMaxSize.String()
	cfgmap["media-emoji-remote-max-size"] = cfg.Media.EmojiRemoteMaxSize.String()
	cfgmap["media-image-size-hint"] = cfg.Media.ImageSizeHint.String()
	cfgmap["media-video-size-hint"] = cfg.Media.VideoSizeHint.String()
	cfgmap["media-local-max-size"] = cfg.Media.LocalMaxSize.String()
	cfgmap["media-remote-max-size"] = cfg.Media.RemoteMaxSize.String()
	cfgmap["media-cleanup-from"] = cfg.Media.CleanupFrom
	cfgmap["media-cleanup-every"] = cfg.Media.CleanupEvery
	cfgmap["media-ffmpeg-pool-size"] = cfg.Media.FfmpegPoolSize
	cfgmap["media-thumb-max-pixels"] = cfg.Media.ThumbMaxPixels
	cfgmap["cache-memory-target"] = cfg.Cache.MemoryTarget.String()
	cfgmap["cache-account-mem-ratio"] = cfg.Cache.AccountMemRatio
	cfgmap["cache-account-note-mem-ratio"] = cfg.Cache.AccountNoteMemRatio
	cfgmap["cache-account-settings-mem-ratio"] = cfg.Cache.AccountSettingsMemRatio
	cfgmap["cache-account-stats-mem-ratio"] = cfg.Cache.AccountStatsMemRatio
	cfgmap["cache-application-mem-ratio"] = cfg.Cache.ApplicationMemRatio
	cfgmap["cache-block-mem-ratio"] = cfg.Cache.BlockMemRatio
	cfgmap["cache-block-ids-mem-ratio"] = cfg.Cache.BlockIDsMemRatio
	cfgmap["cache-boost-of-ids-mem-ratio"] = cfg.Cache.BoostOfIDsMemRatio
	cfgmap["cache-client-mem-ratio"] = cfg.Cache.ClientMemRatio
	cfgmap["cache-conversation-mem-ratio"] = cfg.Cache.ConversationMemRatio
	cfgmap["cache-conversation-last-status-ids-mem-ratio"] = cfg.Cache.ConversationLastStatusIDsMemRatio
	cfgmap["cache-domain-permission-draft-mem-ratio"] = cfg.Cache.DomainPermissionDraftMemRation
	cfgmap["cache-domain-permission-subscription-mem-ratio"] = cfg.Cache.DomainPermissionSubscriptionMemRation
	cfgmap["cache-emoji-mem-ratio"] = cfg.Cache.EmojiMemRatio
	cfgmap["cache-emoji-category-mem-ratio"] = cfg.Cache.EmojiCategoryMemRatio
	cfgmap["cache-filter-mem-ratio"] = cfg.Cache.FilterMemRatio
	cfgmap["cache-filter-ids-mem-ratio"] = cfg.Cache.FilterIDsMemRatio
	cfgmap["cache-filter-keyword-mem-ratio"] = cfg.Cache.FilterKeywordMemRatio
	cfgmap["cache-filter-status-mem-ratio"] = cfg.Cache.FilterStatusMemRatio
	cfgmap["cache-follow-mem-ratio"] = cfg.Cache.FollowMemRatio
	cfgmap["cache-follow-ids-mem-ratio"] = cfg.Cache.FollowIDsMemRatio
	cfgmap["cache-follow-request-mem-ratio"] = cfg.Cache.FollowRequestMemRatio
	cfgmap["cache-follow-request-ids-mem-ratio"] = cfg.Cache.FollowRequestIDsMemRatio
	cfgmap["cache-following-tag-ids-mem-ratio"] = cfg.Cache.FollowingTagIDsMemRatio
	cfgmap["cache-in-reply-to-ids-mem-ratio"] = cfg.Cache.InReplyToIDsMemRatio
	cfgmap["cache-instance-mem-ratio"] = cfg.Cache.InstanceMemRatio
	cfgmap["cache-interaction-request-mem-ratio"] = cfg.Cache.InteractionRequestMemRatio
	cfgmap["cache-list-mem-ratio"] = cfg.Cache.ListMemRatio
	cfgmap["cache-list-ids-mem-ratio"] = cfg.Cache.ListIDsMemRatio
	cfgmap["cache-listed-ids-mem-ratio"] = cfg.Cache.ListedIDsMemRatio
	cfgmap["cache-marker-mem-ratio"] = cfg.Cache.MarkerMemRatio
	cfgmap["cache-media-mem-ratio"] = cfg.Cache.MediaMemRatio
	cfgmap["cache-mention-mem-ratio"] = cfg.Cache.MentionMemRatio
	cfgmap["cache-move-mem-ratio"] = cfg.Cache.MoveMemRatio
	cfgmap["cache-notification-mem-ratio"] = cfg.Cache.NotificationMemRatio
	cfgmap["cache-poll-mem-ratio"] = cfg.Cache.PollMemRatio
	cfgmap["cache-poll-vote-mem-ratio"] = cfg.Cache.PollVoteMemRatio
	cfgmap["cache-poll-vote-ids-mem-ratio"] = cfg.Cache.PollVoteIDsMemRatio
	cfgmap["cache-report-mem-ratio"] = cfg.Cache.ReportMemRatio
	cfgmap["cache-scheduled-status-mem-ratio"] = cfg.Cache.ScheduledStatusMemRatio
	cfgmap["cache-sin-bin-status-mem-ratio"] = cfg.Cache.SinBinStatusMemRatio
	cfgmap["cache-status-mem-ratio"] = cfg.Cache.StatusMemRatio
	cfgmap["cache-status-bookmark-mem-ratio"] = cfg.Cache.StatusBookmarkMemRatio
	cfgmap["cache-status-bookmark-ids-mem-ratio"] = cfg.Cache.StatusBookmarkIDsMemRatio
	cfgmap["cache-status-edit-mem-ratio"] = cfg.Cache.StatusEditMemRatio
	cfgmap["cache-status-fave-mem-ratio"] = cfg.Cache.StatusFaveMemRatio
	cfgmap["cache-status-fave-ids-mem-ratio"] = cfg.Cache.StatusFaveIDsMemRatio
	cfgmap["cache-tag-mem-ratio"] = cfg.Cache.TagMemRatio
	cfgmap["cache-thread-mute-mem-ratio"] = cfg.Cache.ThreadMuteMemRatio
	cfgmap["cache-token-mem-ratio"] = cfg.Cache.TokenMemRatio
	cfgmap["cache-tombstone-mem-ratio"] = cfg.Cache.TombstoneMemRatio
	cfgmap["cache-user-mem-ratio"] = cfg.Cache.UserMemRatio
	cfgmap["cache-user-mute-mem-ratio"] = cfg.Cache.UserMuteMemRatio
	cfgmap["cache-user-mute-ids-mem-ratio"] = cfg.Cache.UserMuteIDsMemRatio
	cfgmap["cache-webfinger-mem-ratio"] = cfg.Cache.WebfingerMemRatio
	cfgmap["cache-web-push-subscription-mem-ratio"] = cfg.Cache.WebPushSubscriptionMemRatio
	cfgmap["cache-web-push-subscription-ids-mem-ratio"] = cfg.Cache.WebPushSubscriptionIDsMemRatio
	cfgmap["cache-mutes-mem-ratio"] = cfg.Cache.MutesMemRatio
	cfgmap["cache-status-filter-mem-ratio"] = cfg.Cache.StatusFilterMemRatio
	cfgmap["cache-visibility-mem-ratio"] = cfg.Cache.VisibilityMemRatio
	cfgmap["username"] = cfg.AdminAccountUsername
	cfgmap["email"] = cfg.AdminAccountEmail
	cfgmap["password"] = cfg.AdminAccountPassword
	cfgmap["path"] = cfg.AdminTransPath
	cfgmap["dry-run"] = cfg.AdminMediaPruneDryRun
	cfgmap["local-only"] = cfg.AdminMediaListLocalOnly
	cfgmap["remote-only"] = cfg.AdminMediaListRemoteOnly
	cfgmap["skip-db-setup"] = cfg.TestrigSkipDBSetup
	cfgmap["skip-db-teardown"] = cfg.TestrigSkipDBTeardown
	return cfgmap
}

func (cfg *Configuration) UnmarshalMap(cfgmap map[string]any) error {
	// VERY IMPORTANT FIRST STEP!
	// flatten to normalize map to
	// entirely un-nested key values
	flattenConfigMap(cfgmap)

	if ival, ok := cfgmap["log-level"]; ok {
		var err error
		cfg.LogLevel, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'log-level': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["log-format"]; ok {
		var err error
		cfg.LogFormat, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'log-format': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["log-timestamp-format"]; ok {
		var err error
		cfg.LogTimestampFormat, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'log-timestamp-format': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["log-db-queries"]; ok {
		var err error
		cfg.LogDbQueries, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'log-db-queries': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["log-client-ip"]; ok {
		var err error
		cfg.LogClientIP, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'log-client-ip': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["request-id-header"]; ok {
		var err error
		cfg.RequestIDHeader, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'request-id-header': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["config-path"]; ok {
		var err error
		cfg.ConfigPath, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'config-path': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["application-name"]; ok {
		var err error
		cfg.ApplicationName, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'application-name': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["landing-page-user"]; ok {
		var err error
		cfg.LandingPageUser, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'landing-page-user': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["host"]; ok {
		var err error
		cfg.Host, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'host': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["account-domain"]; ok {
		var err error
		cfg.AccountDomain, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'account-domain': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["protocol"]; ok {
		var err error
		cfg.Protocol, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'protocol': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["bind-address"]; ok {
		var err error
		cfg.BindAddress, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'bind-address': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["port"]; ok {
		var err error
		cfg.Port, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'port': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["trusted-proxies"]; ok {
		var err error
		cfg.TrustedProxies, err = toStringSlice(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> []string for 'trusted-proxies': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["software-version"]; ok {
		var err error
		cfg.SoftwareVersion, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'software-version': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-type"]; ok {
		var err error
		cfg.DbType, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'db-type': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-address"]; ok {
		var err error
		cfg.DbAddress, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'db-address': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-port"]; ok {
		var err error
		cfg.DbPort, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'db-port': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-user"]; ok {
		var err error
		cfg.DbUser, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'db-user': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-password"]; ok {
		var err error
		cfg.DbPassword, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'db-password': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-database"]; ok {
		var err error
		cfg.DbDatabase, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'db-database': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-tls-mode"]; ok {
		var err error
		cfg.DbTLSMode, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'db-tls-mode': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-tls-ca-cert"]; ok {
		var err error
		cfg.DbTLSCACert, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'db-tls-ca-cert': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-max-open-conns-multiplier"]; ok {
		var err error
		cfg.DbMaxOpenConnsMultiplier, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'db-max-open-conns-multiplier': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-sqlite-journal-mode"]; ok {
		var err error
		cfg.DbSqliteJournalMode, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'db-sqlite-journal-mode': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-sqlite-synchronous"]; ok {
		var err error
		cfg.DbSqliteSynchronous, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'db-sqlite-synchronous': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-sqlite-cache-size"]; ok {
		t, err := cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'db-sqlite-cache-size': %w", ival, err)
		}
		cfg.DbSqliteCacheSize = 0x0
		if err := cfg.DbSqliteCacheSize.Set(t); err != nil {
			return fmt.Errorf("error parsing %#v for 'db-sqlite-cache-size': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-sqlite-busy-timeout"]; ok {
		var err error
		cfg.DbSqliteBusyTimeout, err = cast.ToDurationE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> time.Duration for 'db-sqlite-busy-timeout': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["db-postgres-connection-string"]; ok {
		var err error
		cfg.DbPostgresConnectionString, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'db-postgres-connection-string': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["web-template-base-dir"]; ok {
		var err error
		cfg.WebTemplateBaseDir, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'web-template-base-dir': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["web-asset-base-dir"]; ok {
		var err error
		cfg.WebAssetBaseDir, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'web-asset-base-dir': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-federation-mode"]; ok {
		var err error
		cfg.InstanceFederationMode, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'instance-federation-mode': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-federation-spam-filter"]; ok {
		var err error
		cfg.InstanceFederationSpamFilter, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'instance-federation-spam-filter': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-expose-peers"]; ok {
		var err error
		cfg.InstanceExposePeers, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'instance-expose-peers': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-expose-blocklist"]; ok {
		var err error
		cfg.InstanceExposeBlocklist, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'instance-expose-blocklist': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-expose-blocklist-web"]; ok {
		var err error
		cfg.InstanceExposeBlocklistWeb, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'instance-expose-blocklist-web': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-expose-allowlist"]; ok {
		var err error
		cfg.InstanceExposeAllowlist, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'instance-expose-allowlist': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-expose-allowlist-web"]; ok {
		var err error
		cfg.InstanceExposeAllowlistWeb, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'instance-expose-allowlist-web': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-expose-public-timeline"]; ok {
		var err error
		cfg.InstanceExposePublicTimeline, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'instance-expose-public-timeline': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-expose-custom-emojis"]; ok {
		var err error
		cfg.InstanceExposeCustomEmojis, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'instance-expose-custom-emojis': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-deliver-to-shared-inboxes"]; ok {
		var err error
		cfg.InstanceDeliverToSharedInboxes, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'instance-deliver-to-shared-inboxes': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-inject-mastodon-version"]; ok {
		var err error
		cfg.InstanceInjectMastodonVersion, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'instance-inject-mastodon-version': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-languages"]; ok {
		t, err := toStringSlice(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> []string for 'instance-languages': %w", ival, err)
		}
		cfg.InstanceLanguages = language.Languages{}
		for _, in := range t {
			if err := cfg.InstanceLanguages.Set(in); err != nil {
				return fmt.Errorf("error parsing %#v for 'instance-languages': %w", ival, err)
			}
		}
	}

	if ival, ok := cfgmap["instance-subscriptions-process-from"]; ok {
		var err error
		cfg.InstanceSubscriptionsProcessFrom, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'instance-subscriptions-process-from': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-subscriptions-process-every"]; ok {
		var err error
		cfg.InstanceSubscriptionsProcessEvery, err = cast.ToDurationE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> time.Duration for 'instance-subscriptions-process-every': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-stats-mode"]; ok {
		var err error
		cfg.InstanceStatsMode, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'instance-stats-mode': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-allow-backdating-statuses"]; ok {
		var err error
		cfg.InstanceAllowBackdatingStatuses, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'instance-allow-backdating-statuses': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["instance-allow-empty-user-agents"]; ok {
		var err error
		cfg.InstanceAllowEmptyUserAgents, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'instance-allow-empty-user-agents': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["accounts-registration-open"]; ok {
		var err error
		cfg.AccountsRegistrationOpen, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'accounts-registration-open': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["accounts-reason-required"]; ok {
		var err error
		cfg.AccountsReasonRequired, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'accounts-reason-required': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["accounts-registration-daily-limit"]; ok {
		var err error
		cfg.AccountsRegistrationDailyLimit, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'accounts-registration-daily-limit': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["accounts-registration-backlog-limit"]; ok {
		var err error
		cfg.AccountsRegistrationBacklogLimit, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'accounts-registration-backlog-limit': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["accounts-allow-custom-css"]; ok {
		var err error
		cfg.AccountsAllowCustomCSS, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'accounts-allow-custom-css': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["accounts-custom-css-length"]; ok {
		var err error
		cfg.AccountsCustomCSSLength, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'accounts-custom-css-length': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["accounts-max-profile-fields"]; ok {
		var err error
		cfg.AccountsMaxProfileFields, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'accounts-max-profile-fields': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["storage-backend"]; ok {
		var err error
		cfg.StorageBackend, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'storage-backend': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["storage-local-base-path"]; ok {
		var err error
		cfg.StorageLocalBasePath, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'storage-local-base-path': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["storage-s3-endpoint"]; ok {
		var err error
		cfg.StorageS3Endpoint, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'storage-s3-endpoint': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["storage-s3-access-key"]; ok {
		var err error
		cfg.StorageS3AccessKey, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'storage-s3-access-key': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["storage-s3-secret-key"]; ok {
		var err error
		cfg.StorageS3SecretKey, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'storage-s3-secret-key': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["storage-s3-use-ssl"]; ok {
		var err error
		cfg.StorageS3UseSSL, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'storage-s3-use-ssl': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["storage-s3-bucket"]; ok {
		var err error
		cfg.StorageS3BucketName, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'storage-s3-bucket': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["storage-s3-proxy"]; ok {
		var err error
		cfg.StorageS3Proxy, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'storage-s3-proxy': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["storage-s3-redirect-url"]; ok {
		var err error
		cfg.StorageS3RedirectURL, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'storage-s3-redirect-url': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["storage-s3-bucket-lookup"]; ok {
		var err error
		cfg.StorageS3BucketLookup, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'storage-s3-bucket-lookup': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["storage-s3-key-prefix"]; ok {
		var err error
		cfg.StorageS3KeyPrefix, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'storage-s3-key-prefix': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["statuses-max-chars"]; ok {
		var err error
		cfg.StatusesMaxChars, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'statuses-max-chars': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["statuses-poll-max-options"]; ok {
		var err error
		cfg.StatusesPollMaxOptions, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'statuses-poll-max-options': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["statuses-poll-option-max-chars"]; ok {
		var err error
		cfg.StatusesPollOptionMaxChars, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'statuses-poll-option-max-chars': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["statuses-media-max-files"]; ok {
		var err error
		cfg.StatusesMediaMaxFiles, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'statuses-media-max-files': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["scheduled-statuses-max-total"]; ok {
		var err error
		cfg.ScheduledStatusesMaxTotal, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'scheduled-statuses-max-total': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["scheduled-statuses-max-daily"]; ok {
		var err error
		cfg.ScheduledStatusesMaxDaily, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'scheduled-statuses-max-daily': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["letsencrypt-enabled"]; ok {
		var err error
		cfg.LetsEncryptEnabled, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'letsencrypt-enabled': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["letsencrypt-port"]; ok {
		var err error
		cfg.LetsEncryptPort, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'letsencrypt-port': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["letsencrypt-cert-dir"]; ok {
		var err error
		cfg.LetsEncryptCertDir, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'letsencrypt-cert-dir': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["letsencrypt-email-address"]; ok {
		var err error
		cfg.LetsEncryptEmailAddress, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'letsencrypt-email-address': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["tls-certificate-chain"]; ok {
		var err error
		cfg.TLSCertificateChain, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'tls-certificate-chain': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["tls-certificate-key"]; ok {
		var err error
		cfg.TLSCertificateKey, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'tls-certificate-key': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["oidc-enabled"]; ok {
		var err error
		cfg.OIDCEnabled, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'oidc-enabled': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["oidc-idp-name"]; ok {
		var err error
		cfg.OIDCIdpName, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'oidc-idp-name': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["oidc-skip-verification"]; ok {
		var err error
		cfg.OIDCSkipVerification, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'oidc-skip-verification': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["oidc-issuer"]; ok {
		var err error
		cfg.OIDCIssuer, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'oidc-issuer': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["oidc-client-id"]; ok {
		var err error
		cfg.OIDCClientID, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'oidc-client-id': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["oidc-client-secret"]; ok {
		var err error
		cfg.OIDCClientSecret, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'oidc-client-secret': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["oidc-scopes"]; ok {
		var err error
		cfg.OIDCScopes, err = toStringSlice(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> []string for 'oidc-scopes': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["oidc-link-existing"]; ok {
		var err error
		cfg.OIDCLinkExisting, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'oidc-link-existing': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["oidc-allowed-groups"]; ok {
		var err error
		cfg.OIDCAllowedGroups, err = toStringSlice(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> []string for 'oidc-allowed-groups': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["oidc-admin-groups"]; ok {
		var err error
		cfg.OIDCAdminGroups, err = toStringSlice(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> []string for 'oidc-admin-groups': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["tracing-enabled"]; ok {
		var err error
		cfg.TracingEnabled, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'tracing-enabled': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["metrics-enabled"]; ok {
		var err error
		cfg.MetricsEnabled, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'metrics-enabled': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["smtp-host"]; ok {
		var err error
		cfg.SMTPHost, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'smtp-host': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["smtp-port"]; ok {
		var err error
		cfg.SMTPPort, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'smtp-port': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["smtp-username"]; ok {
		var err error
		cfg.SMTPUsername, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'smtp-username': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["smtp-password"]; ok {
		var err error
		cfg.SMTPPassword, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'smtp-password': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["smtp-from"]; ok {
		var err error
		cfg.SMTPFrom, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'smtp-from': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["smtp-disclose-recipients"]; ok {
		var err error
		cfg.SMTPDiscloseRecipients, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'smtp-disclose-recipients': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["syslog-enabled"]; ok {
		var err error
		cfg.SyslogEnabled, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'syslog-enabled': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["syslog-protocol"]; ok {
		var err error
		cfg.SyslogProtocol, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'syslog-protocol': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["syslog-address"]; ok {
		var err error
		cfg.SyslogAddress, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'syslog-address': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["advanced-cookies-samesite"]; ok {
		var err error
		cfg.Advanced.CookiesSamesite, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'advanced-cookies-samesite': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["advanced-sender-multiplier"]; ok {
		var err error
		cfg.Advanced.SenderMultiplier, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'advanced-sender-multiplier': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["advanced-csp-extra-uris"]; ok {
		var err error
		cfg.Advanced.CSPExtraURIs, err = toStringSlice(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> []string for 'advanced-csp-extra-uris': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["advanced-header-filter-mode"]; ok {
		var err error
		cfg.Advanced.HeaderFilterMode, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'advanced-header-filter-mode': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["advanced-rate-limit-requests"]; ok {
		var err error
		cfg.Advanced.RateLimit.Requests, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'advanced-rate-limit-requests': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["advanced-rate-limit-exceptions"]; ok {
		t, err := toStringSlice(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> []string for 'advanced-rate-limit-exceptions': %w", ival, err)
		}
		cfg.Advanced.RateLimit.Exceptions = IPPrefixes{}
		for _, in := range t {
			if err := cfg.Advanced.RateLimit.Exceptions.Set(in); err != nil {
				return fmt.Errorf("error parsing %#v for 'advanced-rate-limit-exceptions': %w", ival, err)
			}
		}
	}

	if ival, ok := cfgmap["advanced-throttling-multiplier"]; ok {
		var err error
		cfg.Advanced.Throttling.Multiplier, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'advanced-throttling-multiplier': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["advanced-throttling-retry-after"]; ok {
		var err error
		cfg.Advanced.Throttling.RetryAfter, err = cast.ToDurationE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> time.Duration for 'advanced-throttling-retry-after': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["advanced-scraper-deterrence-enabled"]; ok {
		var err error
		cfg.Advanced.ScraperDeterrence.Enabled, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'advanced-scraper-deterrence-enabled': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["advanced-scraper-deterrence-difficulty"]; ok {
		var err error
		cfg.Advanced.ScraperDeterrence.Difficulty, err = cast.ToUint32E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> uint32 for 'advanced-scraper-deterrence-difficulty': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["http-client-allow-ips"]; ok {
		var err error
		cfg.HTTPClient.AllowIPs, err = toStringSlice(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> []string for 'http-client-allow-ips': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["http-client-block-ips"]; ok {
		var err error
		cfg.HTTPClient.BlockIPs, err = toStringSlice(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> []string for 'http-client-block-ips': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["http-client-timeout"]; ok {
		var err error
		cfg.HTTPClient.Timeout, err = cast.ToDurationE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> time.Duration for 'http-client-timeout': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["http-client-tls-insecure-skip-verify"]; ok {
		var err error
		cfg.HTTPClient.TLSInsecureSkipVerify, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'http-client-tls-insecure-skip-verify': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["http-client-insecure-outgoing"]; ok {
		var err error
		cfg.HTTPClient.InsecureOutgoing, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'http-client-insecure-outgoing': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-description-min-chars"]; ok {
		var err error
		cfg.Media.DescriptionMinChars, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'media-description-min-chars': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-description-max-chars"]; ok {
		var err error
		cfg.Media.DescriptionMaxChars, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'media-description-max-chars': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-remote-cache-days"]; ok {
		var err error
		cfg.Media.RemoteCacheDays, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'media-remote-cache-days': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-emoji-local-max-size"]; ok {
		t, err := cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'media-emoji-local-max-size': %w", ival, err)
		}
		cfg.Media.EmojiLocalMaxSize = 0x0
		if err := cfg.Media.EmojiLocalMaxSize.Set(t); err != nil {
			return fmt.Errorf("error parsing %#v for 'media-emoji-local-max-size': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-emoji-remote-max-size"]; ok {
		t, err := cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'media-emoji-remote-max-size': %w", ival, err)
		}
		cfg.Media.EmojiRemoteMaxSize = 0x0
		if err := cfg.Media.EmojiRemoteMaxSize.Set(t); err != nil {
			return fmt.Errorf("error parsing %#v for 'media-emoji-remote-max-size': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-image-size-hint"]; ok {
		t, err := cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'media-image-size-hint': %w", ival, err)
		}
		cfg.Media.ImageSizeHint = 0x0
		if err := cfg.Media.ImageSizeHint.Set(t); err != nil {
			return fmt.Errorf("error parsing %#v for 'media-image-size-hint': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-video-size-hint"]; ok {
		t, err := cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'media-video-size-hint': %w", ival, err)
		}
		cfg.Media.VideoSizeHint = 0x0
		if err := cfg.Media.VideoSizeHint.Set(t); err != nil {
			return fmt.Errorf("error parsing %#v for 'media-video-size-hint': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-local-max-size"]; ok {
		t, err := cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'media-local-max-size': %w", ival, err)
		}
		cfg.Media.LocalMaxSize = 0x0
		if err := cfg.Media.LocalMaxSize.Set(t); err != nil {
			return fmt.Errorf("error parsing %#v for 'media-local-max-size': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-remote-max-size"]; ok {
		t, err := cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'media-remote-max-size': %w", ival, err)
		}
		cfg.Media.RemoteMaxSize = 0x0
		if err := cfg.Media.RemoteMaxSize.Set(t); err != nil {
			return fmt.Errorf("error parsing %#v for 'media-remote-max-size': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-cleanup-from"]; ok {
		var err error
		cfg.Media.CleanupFrom, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'media-cleanup-from': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-cleanup-every"]; ok {
		var err error
		cfg.Media.CleanupEvery, err = cast.ToDurationE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> time.Duration for 'media-cleanup-every': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-ffmpeg-pool-size"]; ok {
		var err error
		cfg.Media.FfmpegPoolSize, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'media-ffmpeg-pool-size': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["media-thumb-max-pixels"]; ok {
		var err error
		cfg.Media.ThumbMaxPixels, err = cast.ToIntE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> int for 'media-thumb-max-pixels': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-memory-target"]; ok {
		t, err := cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'cache-memory-target': %w", ival, err)
		}
		cfg.Cache.MemoryTarget = 0x0
		if err := cfg.Cache.MemoryTarget.Set(t); err != nil {
			return fmt.Errorf("error parsing %#v for 'cache-memory-target': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-account-mem-ratio"]; ok {
		var err error
		cfg.Cache.AccountMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-account-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-account-note-mem-ratio"]; ok {
		var err error
		cfg.Cache.AccountNoteMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-account-note-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-account-settings-mem-ratio"]; ok {
		var err error
		cfg.Cache.AccountSettingsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-account-settings-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-account-stats-mem-ratio"]; ok {
		var err error
		cfg.Cache.AccountStatsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-account-stats-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-application-mem-ratio"]; ok {
		var err error
		cfg.Cache.ApplicationMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-application-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-block-mem-ratio"]; ok {
		var err error
		cfg.Cache.BlockMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-block-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-block-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.BlockIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-block-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-boost-of-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.BoostOfIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-boost-of-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-client-mem-ratio"]; ok {
		var err error
		cfg.Cache.ClientMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-client-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-conversation-mem-ratio"]; ok {
		var err error
		cfg.Cache.ConversationMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-conversation-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-conversation-last-status-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.ConversationLastStatusIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-conversation-last-status-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-domain-permission-draft-mem-ratio"]; ok {
		var err error
		cfg.Cache.DomainPermissionDraftMemRation, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-domain-permission-draft-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-domain-permission-subscription-mem-ratio"]; ok {
		var err error
		cfg.Cache.DomainPermissionSubscriptionMemRation, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-domain-permission-subscription-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-emoji-mem-ratio"]; ok {
		var err error
		cfg.Cache.EmojiMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-emoji-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-emoji-category-mem-ratio"]; ok {
		var err error
		cfg.Cache.EmojiCategoryMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-emoji-category-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-filter-mem-ratio"]; ok {
		var err error
		cfg.Cache.FilterMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-filter-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-filter-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.FilterIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-filter-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-filter-keyword-mem-ratio"]; ok {
		var err error
		cfg.Cache.FilterKeywordMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-filter-keyword-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-filter-status-mem-ratio"]; ok {
		var err error
		cfg.Cache.FilterStatusMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-filter-status-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-follow-mem-ratio"]; ok {
		var err error
		cfg.Cache.FollowMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-follow-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-follow-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.FollowIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-follow-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-follow-request-mem-ratio"]; ok {
		var err error
		cfg.Cache.FollowRequestMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-follow-request-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-follow-request-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.FollowRequestIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-follow-request-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-following-tag-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.FollowingTagIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-following-tag-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-in-reply-to-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.InReplyToIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-in-reply-to-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-instance-mem-ratio"]; ok {
		var err error
		cfg.Cache.InstanceMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-instance-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-interaction-request-mem-ratio"]; ok {
		var err error
		cfg.Cache.InteractionRequestMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-interaction-request-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-list-mem-ratio"]; ok {
		var err error
		cfg.Cache.ListMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-list-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-list-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.ListIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-list-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-listed-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.ListedIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-listed-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-marker-mem-ratio"]; ok {
		var err error
		cfg.Cache.MarkerMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-marker-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-media-mem-ratio"]; ok {
		var err error
		cfg.Cache.MediaMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-media-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-mention-mem-ratio"]; ok {
		var err error
		cfg.Cache.MentionMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-mention-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-move-mem-ratio"]; ok {
		var err error
		cfg.Cache.MoveMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-move-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-notification-mem-ratio"]; ok {
		var err error
		cfg.Cache.NotificationMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-notification-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-poll-mem-ratio"]; ok {
		var err error
		cfg.Cache.PollMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-poll-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-poll-vote-mem-ratio"]; ok {
		var err error
		cfg.Cache.PollVoteMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-poll-vote-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-poll-vote-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.PollVoteIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-poll-vote-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-report-mem-ratio"]; ok {
		var err error
		cfg.Cache.ReportMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-report-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-scheduled-status-mem-ratio"]; ok {
		var err error
		cfg.Cache.ScheduledStatusMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-scheduled-status-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-sin-bin-status-mem-ratio"]; ok {
		var err error
		cfg.Cache.SinBinStatusMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-sin-bin-status-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-status-mem-ratio"]; ok {
		var err error
		cfg.Cache.StatusMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-status-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-status-bookmark-mem-ratio"]; ok {
		var err error
		cfg.Cache.StatusBookmarkMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-status-bookmark-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-status-bookmark-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.StatusBookmarkIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-status-bookmark-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-status-edit-mem-ratio"]; ok {
		var err error
		cfg.Cache.StatusEditMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-status-edit-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-status-fave-mem-ratio"]; ok {
		var err error
		cfg.Cache.StatusFaveMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-status-fave-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-status-fave-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.StatusFaveIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-status-fave-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-tag-mem-ratio"]; ok {
		var err error
		cfg.Cache.TagMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-tag-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-thread-mute-mem-ratio"]; ok {
		var err error
		cfg.Cache.ThreadMuteMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-thread-mute-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-token-mem-ratio"]; ok {
		var err error
		cfg.Cache.TokenMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-token-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-tombstone-mem-ratio"]; ok {
		var err error
		cfg.Cache.TombstoneMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-tombstone-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-user-mem-ratio"]; ok {
		var err error
		cfg.Cache.UserMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-user-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-user-mute-mem-ratio"]; ok {
		var err error
		cfg.Cache.UserMuteMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-user-mute-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-user-mute-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.UserMuteIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-user-mute-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-webfinger-mem-ratio"]; ok {
		var err error
		cfg.Cache.WebfingerMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-webfinger-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-web-push-subscription-mem-ratio"]; ok {
		var err error
		cfg.Cache.WebPushSubscriptionMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-web-push-subscription-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-web-push-subscription-ids-mem-ratio"]; ok {
		var err error
		cfg.Cache.WebPushSubscriptionIDsMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-web-push-subscription-ids-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-mutes-mem-ratio"]; ok {
		var err error
		cfg.Cache.MutesMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-mutes-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-status-filter-mem-ratio"]; ok {
		var err error
		cfg.Cache.StatusFilterMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-status-filter-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["cache-visibility-mem-ratio"]; ok {
		var err error
		cfg.Cache.VisibilityMemRatio, err = cast.ToFloat64E(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> float64 for 'cache-visibility-mem-ratio': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["username"]; ok {
		var err error
		cfg.AdminAccountUsername, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'username': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["email"]; ok {
		var err error
		cfg.AdminAccountEmail, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'email': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["password"]; ok {
		var err error
		cfg.AdminAccountPassword, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'password': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["path"]; ok {
		var err error
		cfg.AdminTransPath, err = cast.ToStringE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> string for 'path': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["dry-run"]; ok {
		var err error
		cfg.AdminMediaPruneDryRun, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'dry-run': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["local-only"]; ok {
		var err error
		cfg.AdminMediaListLocalOnly, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'local-only': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["remote-only"]; ok {
		var err error
		cfg.AdminMediaListRemoteOnly, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'remote-only': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["skip-db-setup"]; ok {
		var err error
		cfg.TestrigSkipDBSetup, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'skip-db-setup': %w", ival, err)
		}
	}

	if ival, ok := cfgmap["skip-db-teardown"]; ok {
		var err error
		cfg.TestrigSkipDBTeardown, err = cast.ToBoolE(ival)
		if err != nil {
			return fmt.Errorf("error casting %#v -> bool for 'skip-db-teardown': %w", ival, err)
		}
	}

	return nil
}

// GetLogLevel safely fetches the Configuration value for state's 'LogLevel' field
func (st *ConfigState) GetLogLevel() (v string) {
	st.mutex.RLock()
	v = st.config.LogLevel
	st.mutex.RUnlock()
	return
}

// SetLogLevel safely sets the Configuration value for state's 'LogLevel' field
func (st *ConfigState) SetLogLevel(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LogLevel = v
	st.reloadToViper()
}

// GetLogLevel safely fetches the value for global configuration 'LogLevel' field
func GetLogLevel() string { return global.GetLogLevel() }

// SetLogLevel safely sets the value for global configuration 'LogLevel' field
func SetLogLevel(v string) { global.SetLogLevel(v) }

// GetLogFormat safely fetches the Configuration value for state's 'LogFormat' field
func (st *ConfigState) GetLogFormat() (v string) {
	st.mutex.RLock()
	v = st.config.LogFormat
	st.mutex.RUnlock()
	return
}

// SetLogFormat safely sets the Configuration value for state's 'LogFormat' field
func (st *ConfigState) SetLogFormat(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LogFormat = v
	st.reloadToViper()
}

// GetLogFormat safely fetches the value for global configuration 'LogFormat' field
func GetLogFormat() string { return global.GetLogFormat() }

// SetLogFormat safely sets the value for global configuration 'LogFormat' field
func SetLogFormat(v string) { global.SetLogFormat(v) }

// GetLogTimestampFormat safely fetches the Configuration value for state's 'LogTimestampFormat' field
func (st *ConfigState) GetLogTimestampFormat() (v string) {
	st.mutex.RLock()
	v = st.config.LogTimestampFormat
	st.mutex.RUnlock()
	return
}

// SetLogTimestampFormat safely sets the Configuration value for state's 'LogTimestampFormat' field
func (st *ConfigState) SetLogTimestampFormat(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LogTimestampFormat = v
	st.reloadToViper()
}

// GetLogTimestampFormat safely fetches the value for global configuration 'LogTimestampFormat' field
func GetLogTimestampFormat() string { return global.GetLogTimestampFormat() }

// SetLogTimestampFormat safely sets the value for global configuration 'LogTimestampFormat' field
func SetLogTimestampFormat(v string) { global.SetLogTimestampFormat(v) }

// GetLogDbQueries safely fetches the Configuration value for state's 'LogDbQueries' field
func (st *ConfigState) GetLogDbQueries() (v bool) {
	st.mutex.RLock()
	v = st.config.LogDbQueries
	st.mutex.RUnlock()
	return
}

// SetLogDbQueries safely sets the Configuration value for state's 'LogDbQueries' field
func (st *ConfigState) SetLogDbQueries(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LogDbQueries = v
	st.reloadToViper()
}

// GetLogDbQueries safely fetches the value for global configuration 'LogDbQueries' field
func GetLogDbQueries() bool { return global.GetLogDbQueries() }

// SetLogDbQueries safely sets the value for global configuration 'LogDbQueries' field
func SetLogDbQueries(v bool) { global.SetLogDbQueries(v) }

// GetLogClientIP safely fetches the Configuration value for state's 'LogClientIP' field
func (st *ConfigState) GetLogClientIP() (v bool) {
	st.mutex.RLock()
	v = st.config.LogClientIP
	st.mutex.RUnlock()
	return
}

// SetLogClientIP safely sets the Configuration value for state's 'LogClientIP' field
func (st *ConfigState) SetLogClientIP(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LogClientIP = v
	st.reloadToViper()
}

// GetLogClientIP safely fetches the value for global configuration 'LogClientIP' field
func GetLogClientIP() bool { return global.GetLogClientIP() }

// SetLogClientIP safely sets the value for global configuration 'LogClientIP' field
func SetLogClientIP(v bool) { global.SetLogClientIP(v) }

// GetRequestIDHeader safely fetches the Configuration value for state's 'RequestIDHeader' field
func (st *ConfigState) GetRequestIDHeader() (v string) {
	st.mutex.RLock()
	v = st.config.RequestIDHeader
	st.mutex.RUnlock()
	return
}

// SetRequestIDHeader safely sets the Configuration value for state's 'RequestIDHeader' field
func (st *ConfigState) SetRequestIDHeader(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.RequestIDHeader = v
	st.reloadToViper()
}

// GetRequestIDHeader safely fetches the value for global configuration 'RequestIDHeader' field
func GetRequestIDHeader() string { return global.GetRequestIDHeader() }

// SetRequestIDHeader safely sets the value for global configuration 'RequestIDHeader' field
func SetRequestIDHeader(v string) { global.SetRequestIDHeader(v) }

// GetConfigPath safely fetches the Configuration value for state's 'ConfigPath' field
func (st *ConfigState) GetConfigPath() (v string) {
	st.mutex.RLock()
	v = st.config.ConfigPath
	st.mutex.RUnlock()
	return
}

// SetConfigPath safely sets the Configuration value for state's 'ConfigPath' field
func (st *ConfigState) SetConfigPath(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.ConfigPath = v
	st.reloadToViper()
}

// GetConfigPath safely fetches the value for global configuration 'ConfigPath' field
func GetConfigPath() string { return global.GetConfigPath() }

// SetConfigPath safely sets the value for global configuration 'ConfigPath' field
func SetConfigPath(v string) { global.SetConfigPath(v) }

// GetApplicationName safely fetches the Configuration value for state's 'ApplicationName' field
func (st *ConfigState) GetApplicationName() (v string) {
	st.mutex.RLock()
	v = st.config.ApplicationName
	st.mutex.RUnlock()
	return
}

// SetApplicationName safely sets the Configuration value for state's 'ApplicationName' field
func (st *ConfigState) SetApplicationName(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.ApplicationName = v
	st.reloadToViper()
}

// GetApplicationName safely fetches the value for global configuration 'ApplicationName' field
func GetApplicationName() string { return global.GetApplicationName() }

// SetApplicationName safely sets the value for global configuration 'ApplicationName' field
func SetApplicationName(v string) { global.SetApplicationName(v) }

// GetLandingPageUser safely fetches the Configuration value for state's 'LandingPageUser' field
func (st *ConfigState) GetLandingPageUser() (v string) {
	st.mutex.RLock()
	v = st.config.LandingPageUser
	st.mutex.RUnlock()
	return
}

// SetLandingPageUser safely sets the Configuration value for state's 'LandingPageUser' field
func (st *ConfigState) SetLandingPageUser(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LandingPageUser = v
	st.reloadToViper()
}

// GetLandingPageUser safely fetches the value for global configuration 'LandingPageUser' field
func GetLandingPageUser() string { return global.GetLandingPageUser() }

// SetLandingPageUser safely sets the value for global configuration 'LandingPageUser' field
func SetLandingPageUser(v string) { global.SetLandingPageUser(v) }

// GetHost safely fetches the Configuration value for state's 'Host' field
func (st *ConfigState) GetHost() (v string) {
	st.mutex.RLock()
	v = st.config.Host
	st.mutex.RUnlock()
	return
}

// SetHost safely sets the Configuration value for state's 'Host' field
func (st *ConfigState) SetHost(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Host = v
	st.reloadToViper()
}

// GetHost safely fetches the value for global configuration 'Host' field
func GetHost() string { return global.GetHost() }

// SetHost safely sets the value for global configuration 'Host' field
func SetHost(v string) { global.SetHost(v) }

// GetAccountDomain safely fetches the Configuration value for state's 'AccountDomain' field
func (st *ConfigState) GetAccountDomain() (v string) {
	st.mutex.RLock()
	v = st.config.AccountDomain
	st.mutex.RUnlock()
	return
}

// SetAccountDomain safely sets the Configuration value for state's 'AccountDomain' field
func (st *ConfigState) SetAccountDomain(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountDomain = v
	st.reloadToViper()
}

// GetAccountDomain safely fetches the value for global configuration 'AccountDomain' field
func GetAccountDomain() string { return global.GetAccountDomain() }

// SetAccountDomain safely sets the value for global configuration 'AccountDomain' field
func SetAccountDomain(v string) { global.SetAccountDomain(v) }

// GetProtocol safely fetches the Configuration value for state's 'Protocol' field
func (st *ConfigState) GetProtocol() (v string) {
	st.mutex.RLock()
	v = st.config.Protocol
	st.mutex.RUnlock()
	return
}

// SetProtocol safely sets the Configuration value for state's 'Protocol' field
func (st *ConfigState) SetProtocol(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Protocol = v
	st.reloadToViper()
}

// GetProtocol safely fetches the value for global configuration 'Protocol' field
func GetProtocol() string { return global.GetProtocol() }

// SetProtocol safely sets the value for global configuration 'Protocol' field
func SetProtocol(v string) { global.SetProtocol(v) }

// GetBindAddress safely fetches the Configuration value for state's 'BindAddress' field
func (st *ConfigState) GetBindAddress() (v string) {
	st.mutex.RLock()
	v = st.config.BindAddress
	st.mutex.RUnlock()
	return
}

// SetBindAddress safely sets the Configuration value for state's 'BindAddress' field
func (st *ConfigState) SetBindAddress(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.BindAddress = v
	st.reloadToViper()
}

// GetBindAddress safely fetches the value for global configuration 'BindAddress' field
func GetBindAddress() string { return global.GetBindAddress() }

// SetBindAddress safely sets the value for global configuration 'BindAddress' field
func SetBindAddress(v string) { global.SetBindAddress(v) }

// GetPort safely fetches the Configuration value for state's 'Port' field
func (st *ConfigState) GetPort() (v int) {
	st.mutex.RLock()
	v = st.config.Port
	st.mutex.RUnlock()
	return
}

// SetPort safely sets the Configuration value for state's 'Port' field
func (st *ConfigState) SetPort(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Port = v
	st.reloadToViper()
}

// GetPort safely fetches the value for global configuration 'Port' field
func GetPort() int { return global.GetPort() }

// SetPort safely sets the value for global configuration 'Port' field
func SetPort(v int) { global.SetPort(v) }

// GetTrustedProxies safely fetches the Configuration value for state's 'TrustedProxies' field
func (st *ConfigState) GetTrustedProxies() (v []string) {
	st.mutex.RLock()
	v = st.config.TrustedProxies
	st.mutex.RUnlock()
	return
}

// SetTrustedProxies safely sets the Configuration value for state's 'TrustedProxies' field
func (st *ConfigState) SetTrustedProxies(v []string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.TrustedProxies = v
	st.reloadToViper()
}

// GetTrustedProxies safely fetches the value for global configuration 'TrustedProxies' field
func GetTrustedProxies() []string { return global.GetTrustedProxies() }

// SetTrustedProxies safely sets the value for global configuration 'TrustedProxies' field
func SetTrustedProxies(v []string) { global.SetTrustedProxies(v) }

// GetSoftwareVersion safely fetches the Configuration value for state's 'SoftwareVersion' field
func (st *ConfigState) GetSoftwareVersion() (v string) {
	st.mutex.RLock()
	v = st.config.SoftwareVersion
	st.mutex.RUnlock()
	return
}

// SetSoftwareVersion safely sets the Configuration value for state's 'SoftwareVersion' field
func (st *ConfigState) SetSoftwareVersion(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SoftwareVersion = v
	st.reloadToViper()
}

// GetSoftwareVersion safely fetches the value for global configuration 'SoftwareVersion' field
func GetSoftwareVersion() string { return global.GetSoftwareVersion() }

// SetSoftwareVersion safely sets the value for global configuration 'SoftwareVersion' field
func SetSoftwareVersion(v string) { global.SetSoftwareVersion(v) }

// GetDbType safely fetches the Configuration value for state's 'DbType' field
func (st *ConfigState) GetDbType() (v string) {
	st.mutex.RLock()
	v = st.config.DbType
	st.mutex.RUnlock()
	return
}

// SetDbType safely sets the Configuration value for state's 'DbType' field
func (st *ConfigState) SetDbType(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbType = v
	st.reloadToViper()
}

// GetDbType safely fetches the value for global configuration 'DbType' field
func GetDbType() string { return global.GetDbType() }

// SetDbType safely sets the value for global configuration 'DbType' field
func SetDbType(v string) { global.SetDbType(v) }

// GetDbAddress safely fetches the Configuration value for state's 'DbAddress' field
func (st *ConfigState) GetDbAddress() (v string) {
	st.mutex.RLock()
	v = st.config.DbAddress
	st.mutex.RUnlock()
	return
}

// SetDbAddress safely sets the Configuration value for state's 'DbAddress' field
func (st *ConfigState) SetDbAddress(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbAddress = v
	st.reloadToViper()
}

// GetDbAddress safely fetches the value for global configuration 'DbAddress' field
func GetDbAddress() string { return global.GetDbAddress() }

// SetDbAddress safely sets the value for global configuration 'DbAddress' field
func SetDbAddress(v string) { global.SetDbAddress(v) }

// GetDbPort safely fetches the Configuration value for state's 'DbPort' field
func (st *ConfigState) GetDbPort() (v int) {
	st.mutex.RLock()
	v = st.config.DbPort
	st.mutex.RUnlock()
	return
}

// SetDbPort safely sets the Configuration value for state's 'DbPort' field
func (st *ConfigState) SetDbPort(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbPort = v
	st.reloadToViper()
}

// GetDbPort safely fetches the value for global configuration 'DbPort' field
func GetDbPort() int { return global.GetDbPort() }

// SetDbPort safely sets the value for global configuration 'DbPort' field
func SetDbPort(v int) { global.SetDbPort(v) }

// GetDbUser safely fetches the Configuration value for state's 'DbUser' field
func (st *ConfigState) GetDbUser() (v string) {
	st.mutex.RLock()
	v = st.config.DbUser
	st.mutex.RUnlock()
	return
}

// SetDbUser safely sets the Configuration value for state's 'DbUser' field
func (st *ConfigState) SetDbUser(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbUser = v
	st.reloadToViper()
}

// GetDbUser safely fetches the value for global configuration 'DbUser' field
func GetDbUser() string { return global.GetDbUser() }

// SetDbUser safely sets the value for global configuration 'DbUser' field
func SetDbUser(v string) { global.SetDbUser(v) }

// GetDbPassword safely fetches the Configuration value for state's 'DbPassword' field
func (st *ConfigState) GetDbPassword() (v string) {
	st.mutex.RLock()
	v = st.config.DbPassword
	st.mutex.RUnlock()
	return
}

// SetDbPassword safely sets the Configuration value for state's 'DbPassword' field
func (st *ConfigState) SetDbPassword(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbPassword = v
	st.reloadToViper()
}

// GetDbPassword safely fetches the value for global configuration 'DbPassword' field
func GetDbPassword() string { return global.GetDbPassword() }

// SetDbPassword safely sets the value for global configuration 'DbPassword' field
func SetDbPassword(v string) { global.SetDbPassword(v) }

// GetDbDatabase safely fetches the Configuration value for state's 'DbDatabase' field
func (st *ConfigState) GetDbDatabase() (v string) {
	st.mutex.RLock()
	v = st.config.DbDatabase
	st.mutex.RUnlock()
	return
}

// SetDbDatabase safely sets the Configuration value for state's 'DbDatabase' field
func (st *ConfigState) SetDbDatabase(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbDatabase = v
	st.reloadToViper()
}

// GetDbDatabase safely fetches the value for global configuration 'DbDatabase' field
func GetDbDatabase() string { return global.GetDbDatabase() }

// SetDbDatabase safely sets the value for global configuration 'DbDatabase' field
func SetDbDatabase(v string) { global.SetDbDatabase(v) }

// GetDbTLSMode safely fetches the Configuration value for state's 'DbTLSMode' field
func (st *ConfigState) GetDbTLSMode() (v string) {
	st.mutex.RLock()
	v = st.config.DbTLSMode
	st.mutex.RUnlock()
	return
}

// SetDbTLSMode safely sets the Configuration value for state's 'DbTLSMode' field
func (st *ConfigState) SetDbTLSMode(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbTLSMode = v
	st.reloadToViper()
}

// GetDbTLSMode safely fetches the value for global configuration 'DbTLSMode' field
func GetDbTLSMode() string { return global.GetDbTLSMode() }

// SetDbTLSMode safely sets the value for global configuration 'DbTLSMode' field
func SetDbTLSMode(v string) { global.SetDbTLSMode(v) }

// GetDbTLSCACert safely fetches the Configuration value for state's 'DbTLSCACert' field
func (st *ConfigState) GetDbTLSCACert() (v string) {
	st.mutex.RLock()
	v = st.config.DbTLSCACert
	st.mutex.RUnlock()
	return
}

// SetDbTLSCACert safely sets the Configuration value for state's 'DbTLSCACert' field
func (st *ConfigState) SetDbTLSCACert(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbTLSCACert = v
	st.reloadToViper()
}

// GetDbTLSCACert safely fetches the value for global configuration 'DbTLSCACert' field
func GetDbTLSCACert() string { return global.GetDbTLSCACert() }

// SetDbTLSCACert safely sets the value for global configuration 'DbTLSCACert' field
func SetDbTLSCACert(v string) { global.SetDbTLSCACert(v) }

// GetDbMaxOpenConnsMultiplier safely fetches the Configuration value for state's 'DbMaxOpenConnsMultiplier' field
func (st *ConfigState) GetDbMaxOpenConnsMultiplier() (v int) {
	st.mutex.RLock()
	v = st.config.DbMaxOpenConnsMultiplier
	st.mutex.RUnlock()
	return
}

// SetDbMaxOpenConnsMultiplier safely sets the Configuration value for state's 'DbMaxOpenConnsMultiplier' field
func (st *ConfigState) SetDbMaxOpenConnsMultiplier(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbMaxOpenConnsMultiplier = v
	st.reloadToViper()
}

// GetDbMaxOpenConnsMultiplier safely fetches the value for global configuration 'DbMaxOpenConnsMultiplier' field
func GetDbMaxOpenConnsMultiplier() int { return global.GetDbMaxOpenConnsMultiplier() }

// SetDbMaxOpenConnsMultiplier safely sets the value for global configuration 'DbMaxOpenConnsMultiplier' field
func SetDbMaxOpenConnsMultiplier(v int) { global.SetDbMaxOpenConnsMultiplier(v) }

// GetDbSqliteJournalMode safely fetches the Configuration value for state's 'DbSqliteJournalMode' field
func (st *ConfigState) GetDbSqliteJournalMode() (v string) {
	st.mutex.RLock()
	v = st.config.DbSqliteJournalMode
	st.mutex.RUnlock()
	return
}

// SetDbSqliteJournalMode safely sets the Configuration value for state's 'DbSqliteJournalMode' field
func (st *ConfigState) SetDbSqliteJournalMode(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbSqliteJournalMode = v
	st.reloadToViper()
}

// GetDbSqliteJournalMode safely fetches the value for global configuration 'DbSqliteJournalMode' field
func GetDbSqliteJournalMode() string { return global.GetDbSqliteJournalMode() }

// SetDbSqliteJournalMode safely sets the value for global configuration 'DbSqliteJournalMode' field
func SetDbSqliteJournalMode(v string) { global.SetDbSqliteJournalMode(v) }

// GetDbSqliteSynchronous safely fetches the Configuration value for state's 'DbSqliteSynchronous' field
func (st *ConfigState) GetDbSqliteSynchronous() (v string) {
	st.mutex.RLock()
	v = st.config.DbSqliteSynchronous
	st.mutex.RUnlock()
	return
}

// SetDbSqliteSynchronous safely sets the Configuration value for state's 'DbSqliteSynchronous' field
func (st *ConfigState) SetDbSqliteSynchronous(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbSqliteSynchronous = v
	st.reloadToViper()
}

// GetDbSqliteSynchronous safely fetches the value for global configuration 'DbSqliteSynchronous' field
func GetDbSqliteSynchronous() string { return global.GetDbSqliteSynchronous() }

// SetDbSqliteSynchronous safely sets the value for global configuration 'DbSqliteSynchronous' field
func SetDbSqliteSynchronous(v string) { global.SetDbSqliteSynchronous(v) }

// GetDbSqliteCacheSize safely fetches the Configuration value for state's 'DbSqliteCacheSize' field
func (st *ConfigState) GetDbSqliteCacheSize() (v bytesize.Size) {
	st.mutex.RLock()
	v = st.config.DbSqliteCacheSize
	st.mutex.RUnlock()
	return
}

// SetDbSqliteCacheSize safely sets the Configuration value for state's 'DbSqliteCacheSize' field
func (st *ConfigState) SetDbSqliteCacheSize(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbSqliteCacheSize = v
	st.reloadToViper()
}

// GetDbSqliteCacheSize safely fetches the value for global configuration 'DbSqliteCacheSize' field
func GetDbSqliteCacheSize() bytesize.Size { return global.GetDbSqliteCacheSize() }

// SetDbSqliteCacheSize safely sets the value for global configuration 'DbSqliteCacheSize' field
func SetDbSqliteCacheSize(v bytesize.Size) { global.SetDbSqliteCacheSize(v) }

// GetDbSqliteBusyTimeout safely fetches the Configuration value for state's 'DbSqliteBusyTimeout' field
func (st *ConfigState) GetDbSqliteBusyTimeout() (v time.Duration) {
	st.mutex.RLock()
	v = st.config.DbSqliteBusyTimeout
	st.mutex.RUnlock()
	return
}

// SetDbSqliteBusyTimeout safely sets the Configuration value for state's 'DbSqliteBusyTimeout' field
func (st *ConfigState) SetDbSqliteBusyTimeout(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbSqliteBusyTimeout = v
	st.reloadToViper()
}

// GetDbSqliteBusyTimeout safely fetches the value for global configuration 'DbSqliteBusyTimeout' field
func GetDbSqliteBusyTimeout() time.Duration { return global.GetDbSqliteBusyTimeout() }

// SetDbSqliteBusyTimeout safely sets the value for global configuration 'DbSqliteBusyTimeout' field
func SetDbSqliteBusyTimeout(v time.Duration) { global.SetDbSqliteBusyTimeout(v) }

// GetDbPostgresConnectionString safely fetches the Configuration value for state's 'DbPostgresConnectionString' field
func (st *ConfigState) GetDbPostgresConnectionString() (v string) {
	st.mutex.RLock()
	v = st.config.DbPostgresConnectionString
	st.mutex.RUnlock()
	return
}

// SetDbPostgresConnectionString safely sets the Configuration value for state's 'DbPostgresConnectionString' field
func (st *ConfigState) SetDbPostgresConnectionString(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbPostgresConnectionString = v
	st.reloadToViper()
}

// GetDbPostgresConnectionString safely fetches the value for global configuration 'DbPostgresConnectionString' field
func GetDbPostgresConnectionString() string { return global.GetDbPostgresConnectionString() }

// SetDbPostgresConnectionString safely sets the value for global configuration 'DbPostgresConnectionString' field
func SetDbPostgresConnectionString(v string) { global.SetDbPostgresConnectionString(v) }

// GetWebTemplateBaseDir safely fetches the Configuration value for state's 'WebTemplateBaseDir' field
func (st *ConfigState) GetWebTemplateBaseDir() (v string) {
	st.mutex.RLock()
	v = st.config.WebTemplateBaseDir
	st.mutex.RUnlock()
	return
}

// SetWebTemplateBaseDir safely sets the Configuration value for state's 'WebTemplateBaseDir' field
func (st *ConfigState) SetWebTemplateBaseDir(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.WebTemplateBaseDir = v
	st.reloadToViper()
}

// GetWebTemplateBaseDir safely fetches the value for global configuration 'WebTemplateBaseDir' field
func GetWebTemplateBaseDir() string { return global.GetWebTemplateBaseDir() }

// SetWebTemplateBaseDir safely sets the value for global configuration 'WebTemplateBaseDir' field
func SetWebTemplateBaseDir(v string) { global.SetWebTemplateBaseDir(v) }

// GetWebAssetBaseDir safely fetches the Configuration value for state's 'WebAssetBaseDir' field
func (st *ConfigState) GetWebAssetBaseDir() (v string) {
	st.mutex.RLock()
	v = st.config.WebAssetBaseDir
	st.mutex.RUnlock()
	return
}

// SetWebAssetBaseDir safely sets the Configuration value for state's 'WebAssetBaseDir' field
func (st *ConfigState) SetWebAssetBaseDir(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.WebAssetBaseDir = v
	st.reloadToViper()
}

// GetWebAssetBaseDir safely fetches the value for global configuration 'WebAssetBaseDir' field
func GetWebAssetBaseDir() string { return global.GetWebAssetBaseDir() }

// SetWebAssetBaseDir safely sets the value for global configuration 'WebAssetBaseDir' field
func SetWebAssetBaseDir(v string) { global.SetWebAssetBaseDir(v) }

// GetInstanceFederationMode safely fetches the Configuration value for state's 'InstanceFederationMode' field
func (st *ConfigState) GetInstanceFederationMode() (v string) {
	st.mutex.RLock()
	v = st.config.InstanceFederationMode
	st.mutex.RUnlock()
	return
}

// SetInstanceFederationMode safely sets the Configuration value for state's 'InstanceFederationMode' field
func (st *ConfigState) SetInstanceFederationMode(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceFederationMode = v
	st.reloadToViper()
}

// GetInstanceFederationMode safely fetches the value for global configuration 'InstanceFederationMode' field
func GetInstanceFederationMode() string { return global.GetInstanceFederationMode() }

// SetInstanceFederationMode safely sets the value for global configuration 'InstanceFederationMode' field
func SetInstanceFederationMode(v string) { global.SetInstanceFederationMode(v) }

// GetInstanceFederationSpamFilter safely fetches the Configuration value for state's 'InstanceFederationSpamFilter' field
func (st *ConfigState) GetInstanceFederationSpamFilter() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceFederationSpamFilter
	st.mutex.RUnlock()
	return
}

// SetInstanceFederationSpamFilter safely sets the Configuration value for state's 'InstanceFederationSpamFilter' field
func (st *ConfigState) SetInstanceFederationSpamFilter(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceFederationSpamFilter = v
	st.reloadToViper()
}

// GetInstanceFederationSpamFilter safely fetches the value for global configuration 'InstanceFederationSpamFilter' field
func GetInstanceFederationSpamFilter() bool { return global.GetInstanceFederationSpamFilter() }

// SetInstanceFederationSpamFilter safely sets the value for global configuration 'InstanceFederationSpamFilter' field
func SetInstanceFederationSpamFilter(v bool) { global.SetInstanceFederationSpamFilter(v) }

// GetInstanceExposePeers safely fetches the Configuration value for state's 'InstanceExposePeers' field
func (st *ConfigState) GetInstanceExposePeers() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceExposePeers
	st.mutex.RUnlock()
	return
}

// SetInstanceExposePeers safely sets the Configuration value for state's 'InstanceExposePeers' field
func (st *ConfigState) SetInstanceExposePeers(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceExposePeers = v
	st.reloadToViper()
}

// GetInstanceExposePeers safely fetches the value for global configuration 'InstanceExposePeers' field
func GetInstanceExposePeers() bool { return global.GetInstanceExposePeers() }

// SetInstanceExposePeers safely sets the value for global configuration 'InstanceExposePeers' field
func SetInstanceExposePeers(v bool) { global.SetInstanceExposePeers(v) }

// GetInstanceExposeBlocklist safely fetches the Configuration value for state's 'InstanceExposeBlocklist' field
func (st *ConfigState) GetInstanceExposeBlocklist() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceExposeBlocklist
	st.mutex.RUnlock()
	return
}

// SetInstanceExposeBlocklist safely sets the Configuration value for state's 'InstanceExposeBlocklist' field
func (st *ConfigState) SetInstanceExposeBlocklist(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceExposeBlocklist = v
	st.reloadToViper()
}

// GetInstanceExposeBlocklist safely fetches the value for global configuration 'InstanceExposeBlocklist' field
func GetInstanceExposeBlocklist() bool { return global.GetInstanceExposeBlocklist() }

// SetInstanceExposeBlocklist safely sets the value for global configuration 'InstanceExposeBlocklist' field
func SetInstanceExposeBlocklist(v bool) { global.SetInstanceExposeBlocklist(v) }

// GetInstanceExposeBlocklistWeb safely fetches the Configuration value for state's 'InstanceExposeBlocklistWeb' field
func (st *ConfigState) GetInstanceExposeBlocklistWeb() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceExposeBlocklistWeb
	st.mutex.RUnlock()
	return
}

// SetInstanceExposeBlocklistWeb safely sets the Configuration value for state's 'InstanceExposeBlocklistWeb' field
func (st *ConfigState) SetInstanceExposeBlocklistWeb(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceExposeBlocklistWeb = v
	st.reloadToViper()
}

// GetInstanceExposeBlocklistWeb safely fetches the value for global configuration 'InstanceExposeBlocklistWeb' field
func GetInstanceExposeBlocklistWeb() bool { return global.GetInstanceExposeBlocklistWeb() }

// SetInstanceExposeBlocklistWeb safely sets the value for global configuration 'InstanceExposeBlocklistWeb' field
func SetInstanceExposeBlocklistWeb(v bool) { global.SetInstanceExposeBlocklistWeb(v) }

// GetInstanceExposeAllowlist safely fetches the Configuration value for state's 'InstanceExposeAllowlist' field
func (st *ConfigState) GetInstanceExposeAllowlist() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceExposeAllowlist
	st.mutex.RUnlock()
	return
}

// SetInstanceExposeAllowlist safely sets the Configuration value for state's 'InstanceExposeAllowlist' field
func (st *ConfigState) SetInstanceExposeAllowlist(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceExposeAllowlist = v
	st.reloadToViper()
}

// GetInstanceExposeAllowlist safely fetches the value for global configuration 'InstanceExposeAllowlist' field
func GetInstanceExposeAllowlist() bool { return global.GetInstanceExposeAllowlist() }

// SetInstanceExposeAllowlist safely sets the value for global configuration 'InstanceExposeAllowlist' field
func SetInstanceExposeAllowlist(v bool) { global.SetInstanceExposeAllowlist(v) }

// GetInstanceExposeAllowlistWeb safely fetches the Configuration value for state's 'InstanceExposeAllowlistWeb' field
func (st *ConfigState) GetInstanceExposeAllowlistWeb() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceExposeAllowlistWeb
	st.mutex.RUnlock()
	return
}

// SetInstanceExposeAllowlistWeb safely sets the Configuration value for state's 'InstanceExposeAllowlistWeb' field
func (st *ConfigState) SetInstanceExposeAllowlistWeb(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceExposeAllowlistWeb = v
	st.reloadToViper()
}

// GetInstanceExposeAllowlistWeb safely fetches the value for global configuration 'InstanceExposeAllowlistWeb' field
func GetInstanceExposeAllowlistWeb() bool { return global.GetInstanceExposeAllowlistWeb() }

// SetInstanceExposeAllowlistWeb safely sets the value for global configuration 'InstanceExposeAllowlistWeb' field
func SetInstanceExposeAllowlistWeb(v bool) { global.SetInstanceExposeAllowlistWeb(v) }

// GetInstanceExposePublicTimeline safely fetches the Configuration value for state's 'InstanceExposePublicTimeline' field
func (st *ConfigState) GetInstanceExposePublicTimeline() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceExposePublicTimeline
	st.mutex.RUnlock()
	return
}

// SetInstanceExposePublicTimeline safely sets the Configuration value for state's 'InstanceExposePublicTimeline' field
func (st *ConfigState) SetInstanceExposePublicTimeline(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceExposePublicTimeline = v
	st.reloadToViper()
}

// GetInstanceExposePublicTimeline safely fetches the value for global configuration 'InstanceExposePublicTimeline' field
func GetInstanceExposePublicTimeline() bool { return global.GetInstanceExposePublicTimeline() }

// SetInstanceExposePublicTimeline safely sets the value for global configuration 'InstanceExposePublicTimeline' field
func SetInstanceExposePublicTimeline(v bool) { global.SetInstanceExposePublicTimeline(v) }

// GetInstanceExposeCustomEmojis safely fetches the Configuration value for state's 'InstanceExposeCustomEmojis' field
func (st *ConfigState) GetInstanceExposeCustomEmojis() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceExposeCustomEmojis
	st.mutex.RUnlock()
	return
}

// SetInstanceExposeCustomEmojis safely sets the Configuration value for state's 'InstanceExposeCustomEmojis' field
func (st *ConfigState) SetInstanceExposeCustomEmojis(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceExposeCustomEmojis = v
	st.reloadToViper()
}

// GetInstanceExposeCustomEmojis safely fetches the value for global configuration 'InstanceExposeCustomEmojis' field
func GetInstanceExposeCustomEmojis() bool { return global.GetInstanceExposeCustomEmojis() }

// SetInstanceExposeCustomEmojis safely sets the value for global configuration 'InstanceExposeCustomEmojis' field
func SetInstanceExposeCustomEmojis(v bool) { global.SetInstanceExposeCustomEmojis(v) }

// GetInstanceDeliverToSharedInboxes safely fetches the Configuration value for state's 'InstanceDeliverToSharedInboxes' field
func (st *ConfigState) GetInstanceDeliverToSharedInboxes() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceDeliverToSharedInboxes
	st.mutex.RUnlock()
	return
}

// SetInstanceDeliverToSharedInboxes safely sets the Configuration value for state's 'InstanceDeliverToSharedInboxes' field
func (st *ConfigState) SetInstanceDeliverToSharedInboxes(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceDeliverToSharedInboxes = v
	st.reloadToViper()
}

// GetInstanceDeliverToSharedInboxes safely fetches the value for global configuration 'InstanceDeliverToSharedInboxes' field
func GetInstanceDeliverToSharedInboxes() bool { return global.GetInstanceDeliverToSharedInboxes() }

// SetInstanceDeliverToSharedInboxes safely sets the value for global configuration 'InstanceDeliverToSharedInboxes' field
func SetInstanceDeliverToSharedInboxes(v bool) { global.SetInstanceDeliverToSharedInboxes(v) }

// GetInstanceInjectMastodonVersion safely fetches the Configuration value for state's 'InstanceInjectMastodonVersion' field
func (st *ConfigState) GetInstanceInjectMastodonVersion() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceInjectMastodonVersion
	st.mutex.RUnlock()
	return
}

// SetInstanceInjectMastodonVersion safely sets the Configuration value for state's 'InstanceInjectMastodonVersion' field
func (st *ConfigState) SetInstanceInjectMastodonVersion(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceInjectMastodonVersion = v
	st.reloadToViper()
}

// GetInstanceInjectMastodonVersion safely fetches the value for global configuration 'InstanceInjectMastodonVersion' field
func GetInstanceInjectMastodonVersion() bool { return global.GetInstanceInjectMastodonVersion() }

// SetInstanceInjectMastodonVersion safely sets the value for global configuration 'InstanceInjectMastodonVersion' field
func SetInstanceInjectMastodonVersion(v bool) { global.SetInstanceInjectMastodonVersion(v) }

// GetInstanceLanguages safely fetches the Configuration value for state's 'InstanceLanguages' field
func (st *ConfigState) GetInstanceLanguages() (v language.Languages) {
	st.mutex.RLock()
	v = st.config.InstanceLanguages
	st.mutex.RUnlock()
	return
}

// SetInstanceLanguages safely sets the Configuration value for state's 'InstanceLanguages' field
func (st *ConfigState) SetInstanceLanguages(v language.Languages) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceLanguages = v
	st.reloadToViper()
}

// GetInstanceLanguages safely fetches the value for global configuration 'InstanceLanguages' field
func GetInstanceLanguages() language.Languages { return global.GetInstanceLanguages() }

// SetInstanceLanguages safely sets the value for global configuration 'InstanceLanguages' field
func SetInstanceLanguages(v language.Languages) { global.SetInstanceLanguages(v) }

// GetInstanceSubscriptionsProcessFrom safely fetches the Configuration value for state's 'InstanceSubscriptionsProcessFrom' field
func (st *ConfigState) GetInstanceSubscriptionsProcessFrom() (v string) {
	st.mutex.RLock()
	v = st.config.InstanceSubscriptionsProcessFrom
	st.mutex.RUnlock()
	return
}

// SetInstanceSubscriptionsProcessFrom safely sets the Configuration value for state's 'InstanceSubscriptionsProcessFrom' field
func (st *ConfigState) SetInstanceSubscriptionsProcessFrom(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceSubscriptionsProcessFrom = v
	st.reloadToViper()
}

// GetInstanceSubscriptionsProcessFrom safely fetches the value for global configuration 'InstanceSubscriptionsProcessFrom' field
func GetInstanceSubscriptionsProcessFrom() string {
	return global.GetInstanceSubscriptionsProcessFrom()
}

// SetInstanceSubscriptionsProcessFrom safely sets the value for global configuration 'InstanceSubscriptionsProcessFrom' field
func SetInstanceSubscriptionsProcessFrom(v string) { global.SetInstanceSubscriptionsProcessFrom(v) }

// GetInstanceSubscriptionsProcessEvery safely fetches the Configuration value for state's 'InstanceSubscriptionsProcessEvery' field
func (st *ConfigState) GetInstanceSubscriptionsProcessEvery() (v time.Duration) {
	st.mutex.RLock()
	v = st.config.InstanceSubscriptionsProcessEvery
	st.mutex.RUnlock()
	return
}

// SetInstanceSubscriptionsProcessEvery safely sets the Configuration value for state's 'InstanceSubscriptionsProcessEvery' field
func (st *ConfigState) SetInstanceSubscriptionsProcessEvery(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceSubscriptionsProcessEvery = v
	st.reloadToViper()
}

// GetInstanceSubscriptionsProcessEvery safely fetches the value for global configuration 'InstanceSubscriptionsProcessEvery' field
func GetInstanceSubscriptionsProcessEvery() time.Duration {
	return global.GetInstanceSubscriptionsProcessEvery()
}

// SetInstanceSubscriptionsProcessEvery safely sets the value for global configuration 'InstanceSubscriptionsProcessEvery' field
func SetInstanceSubscriptionsProcessEvery(v time.Duration) {
	global.SetInstanceSubscriptionsProcessEvery(v)
}

// GetInstanceStatsMode safely fetches the Configuration value for state's 'InstanceStatsMode' field
func (st *ConfigState) GetInstanceStatsMode() (v string) {
	st.mutex.RLock()
	v = st.config.InstanceStatsMode
	st.mutex.RUnlock()
	return
}

// SetInstanceStatsMode safely sets the Configuration value for state's 'InstanceStatsMode' field
func (st *ConfigState) SetInstanceStatsMode(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceStatsMode = v
	st.reloadToViper()
}

// GetInstanceStatsMode safely fetches the value for global configuration 'InstanceStatsMode' field
func GetInstanceStatsMode() string { return global.GetInstanceStatsMode() }

// SetInstanceStatsMode safely sets the value for global configuration 'InstanceStatsMode' field
func SetInstanceStatsMode(v string) { global.SetInstanceStatsMode(v) }

// GetInstanceAllowBackdatingStatuses safely fetches the Configuration value for state's 'InstanceAllowBackdatingStatuses' field
func (st *ConfigState) GetInstanceAllowBackdatingStatuses() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceAllowBackdatingStatuses
	st.mutex.RUnlock()
	return
}

// SetInstanceAllowBackdatingStatuses safely sets the Configuration value for state's 'InstanceAllowBackdatingStatuses' field
func (st *ConfigState) SetInstanceAllowBackdatingStatuses(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceAllowBackdatingStatuses = v
	st.reloadToViper()
}

// GetInstanceAllowBackdatingStatuses safely fetches the value for global configuration 'InstanceAllowBackdatingStatuses' field
func GetInstanceAllowBackdatingStatuses() bool { return global.GetInstanceAllowBackdatingStatuses() }

// SetInstanceAllowBackdatingStatuses safely sets the value for global configuration 'InstanceAllowBackdatingStatuses' field
func SetInstanceAllowBackdatingStatuses(v bool) { global.SetInstanceAllowBackdatingStatuses(v) }

// GetInstanceAllowEmptyUserAgents safely fetches the Configuration value for state's 'InstanceAllowEmptyUserAgents' field
func (st *ConfigState) GetInstanceAllowEmptyUserAgents() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceAllowEmptyUserAgents
	st.mutex.RUnlock()
	return
}

// SetInstanceAllowEmptyUserAgents safely sets the Configuration value for state's 'InstanceAllowEmptyUserAgents' field
func (st *ConfigState) SetInstanceAllowEmptyUserAgents(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceAllowEmptyUserAgents = v
	st.reloadToViper()
}

// GetInstanceAllowEmptyUserAgents safely fetches the value for global configuration 'InstanceAllowEmptyUserAgents' field
func GetInstanceAllowEmptyUserAgents() bool { return global.GetInstanceAllowEmptyUserAgents() }

// SetInstanceAllowEmptyUserAgents safely sets the value for global configuration 'InstanceAllowEmptyUserAgents' field
func SetInstanceAllowEmptyUserAgents(v bool) { global.SetInstanceAllowEmptyUserAgents(v) }

// GetAccountsRegistrationOpen safely fetches the Configuration value for state's 'AccountsRegistrationOpen' field
func (st *ConfigState) GetAccountsRegistrationOpen() (v bool) {
	st.mutex.RLock()
	v = st.config.AccountsRegistrationOpen
	st.mutex.RUnlock()
	return
}

// SetAccountsRegistrationOpen safely sets the Configuration value for state's 'AccountsRegistrationOpen' field
func (st *ConfigState) SetAccountsRegistrationOpen(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountsRegistrationOpen = v
	st.reloadToViper()
}

// GetAccountsRegistrationOpen safely fetches the value for global configuration 'AccountsRegistrationOpen' field
func GetAccountsRegistrationOpen() bool { return global.GetAccountsRegistrationOpen() }

// SetAccountsRegistrationOpen safely sets the value for global configuration 'AccountsRegistrationOpen' field
func SetAccountsRegistrationOpen(v bool) { global.SetAccountsRegistrationOpen(v) }

// GetAccountsReasonRequired safely fetches the Configuration value for state's 'AccountsReasonRequired' field
func (st *ConfigState) GetAccountsReasonRequired() (v bool) {
	st.mutex.RLock()
	v = st.config.AccountsReasonRequired
	st.mutex.RUnlock()
	return
}

// SetAccountsReasonRequired safely sets the Configuration value for state's 'AccountsReasonRequired' field
func (st *ConfigState) SetAccountsReasonRequired(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountsReasonRequired = v
	st.reloadToViper()
}

// GetAccountsReasonRequired safely fetches the value for global configuration 'AccountsReasonRequired' field
func GetAccountsReasonRequired() bool { return global.GetAccountsReasonRequired() }

// SetAccountsReasonRequired safely sets the value for global configuration 'AccountsReasonRequired' field
func SetAccountsReasonRequired(v bool) { global.SetAccountsReasonRequired(v) }

// GetAccountsRegistrationDailyLimit safely fetches the Configuration value for state's 'AccountsRegistrationDailyLimit' field
func (st *ConfigState) GetAccountsRegistrationDailyLimit() (v int) {
	st.mutex.RLock()
	v = st.config.AccountsRegistrationDailyLimit
	st.mutex.RUnlock()
	return
}

// SetAccountsRegistrationDailyLimit safely sets the Configuration value for state's 'AccountsRegistrationDailyLimit' field
func (st *ConfigState) SetAccountsRegistrationDailyLimit(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountsRegistrationDailyLimit = v
	st.reloadToViper()
}

// GetAccountsRegistrationDailyLimit safely fetches the value for global configuration 'AccountsRegistrationDailyLimit' field
func GetAccountsRegistrationDailyLimit() int { return global.GetAccountsRegistrationDailyLimit() }

// SetAccountsRegistrationDailyLimit safely sets the value for global configuration 'AccountsRegistrationDailyLimit' field
func SetAccountsRegistrationDailyLimit(v int) { global.SetAccountsRegistrationDailyLimit(v) }

// GetAccountsRegistrationBacklogLimit safely fetches the Configuration value for state's 'AccountsRegistrationBacklogLimit' field
func (st *ConfigState) GetAccountsRegistrationBacklogLimit() (v int) {
	st.mutex.RLock()
	v = st.config.AccountsRegistrationBacklogLimit
	st.mutex.RUnlock()
	return
}

// SetAccountsRegistrationBacklogLimit safely sets the Configuration value for state's 'AccountsRegistrationBacklogLimit' field
func (st *ConfigState) SetAccountsRegistrationBacklogLimit(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountsRegistrationBacklogLimit = v
	st.reloadToViper()
}

// GetAccountsRegistrationBacklogLimit safely fetches the value for global configuration 'AccountsRegistrationBacklogLimit' field
func GetAccountsRegistrationBacklogLimit() int { return global.GetAccountsRegistrationBacklogLimit() }

// SetAccountsRegistrationBacklogLimit safely sets the value for global configuration 'AccountsRegistrationBacklogLimit' field
func SetAccountsRegistrationBacklogLimit(v int) { global.SetAccountsRegistrationBacklogLimit(v) }

// GetAccountsAllowCustomCSS safely fetches the Configuration value for state's 'AccountsAllowCustomCSS' field
func (st *ConfigState) GetAccountsAllowCustomCSS() (v bool) {
	st.mutex.RLock()
	v = st.config.AccountsAllowCustomCSS
	st.mutex.RUnlock()
	return
}

// SetAccountsAllowCustomCSS safely sets the Configuration value for state's 'AccountsAllowCustomCSS' field
func (st *ConfigState) SetAccountsAllowCustomCSS(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountsAllowCustomCSS = v
	st.reloadToViper()
}

// GetAccountsAllowCustomCSS safely fetches the value for global configuration 'AccountsAllowCustomCSS' field
func GetAccountsAllowCustomCSS() bool { return global.GetAccountsAllowCustomCSS() }

// SetAccountsAllowCustomCSS safely sets the value for global configuration 'AccountsAllowCustomCSS' field
func SetAccountsAllowCustomCSS(v bool) { global.SetAccountsAllowCustomCSS(v) }

// GetAccountsCustomCSSLength safely fetches the Configuration value for state's 'AccountsCustomCSSLength' field
func (st *ConfigState) GetAccountsCustomCSSLength() (v int) {
	st.mutex.RLock()
	v = st.config.AccountsCustomCSSLength
	st.mutex.RUnlock()
	return
}

// SetAccountsCustomCSSLength safely sets the Configuration value for state's 'AccountsCustomCSSLength' field
func (st *ConfigState) SetAccountsCustomCSSLength(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountsCustomCSSLength = v
	st.reloadToViper()
}

// GetAccountsCustomCSSLength safely fetches the value for global configuration 'AccountsCustomCSSLength' field
func GetAccountsCustomCSSLength() int { return global.GetAccountsCustomCSSLength() }

// SetAccountsCustomCSSLength safely sets the value for global configuration 'AccountsCustomCSSLength' field
func SetAccountsCustomCSSLength(v int) { global.SetAccountsCustomCSSLength(v) }

// GetAccountsMaxProfileFields safely fetches the Configuration value for state's 'AccountsMaxProfileFields' field
func (st *ConfigState) GetAccountsMaxProfileFields() (v int) {
	st.mutex.RLock()
	v = st.config.AccountsMaxProfileFields
	st.mutex.RUnlock()
	return
}

// SetAccountsMaxProfileFields safely sets the Configuration value for state's 'AccountsMaxProfileFields' field
func (st *ConfigState) SetAccountsMaxProfileFields(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountsMaxProfileFields = v
	st.reloadToViper()
}

// GetAccountsMaxProfileFields safely fetches the value for global configuration 'AccountsMaxProfileFields' field
func GetAccountsMaxProfileFields() int { return global.GetAccountsMaxProfileFields() }

// SetAccountsMaxProfileFields safely sets the value for global configuration 'AccountsMaxProfileFields' field
func SetAccountsMaxProfileFields(v int) { global.SetAccountsMaxProfileFields(v) }

// GetStorageBackend safely fetches the Configuration value for state's 'StorageBackend' field
func (st *ConfigState) GetStorageBackend() (v string) {
	st.mutex.RLock()
	v = st.config.StorageBackend
	st.mutex.RUnlock()
	return
}

// SetStorageBackend safely sets the Configuration value for state's 'StorageBackend' field
func (st *ConfigState) SetStorageBackend(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageBackend = v
	st.reloadToViper()
}

// GetStorageBackend safely fetches the value for global configuration 'StorageBackend' field
func GetStorageBackend() string { return global.GetStorageBackend() }

// SetStorageBackend safely sets the value for global configuration 'StorageBackend' field
func SetStorageBackend(v string) { global.SetStorageBackend(v) }

// GetStorageLocalBasePath safely fetches the Configuration value for state's 'StorageLocalBasePath' field
func (st *ConfigState) GetStorageLocalBasePath() (v string) {
	st.mutex.RLock()
	v = st.config.StorageLocalBasePath
	st.mutex.RUnlock()
	return
}

// SetStorageLocalBasePath safely sets the Configuration value for state's 'StorageLocalBasePath' field
func (st *ConfigState) SetStorageLocalBasePath(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageLocalBasePath = v
	st.reloadToViper()
}

// GetStorageLocalBasePath safely fetches the value for global configuration 'StorageLocalBasePath' field
func GetStorageLocalBasePath() string { return global.GetStorageLocalBasePath() }

// SetStorageLocalBasePath safely sets the value for global configuration 'StorageLocalBasePath' field
func SetStorageLocalBasePath(v string) { global.SetStorageLocalBasePath(v) }

// GetStorageS3Endpoint safely fetches the Configuration value for state's 'StorageS3Endpoint' field
func (st *ConfigState) GetStorageS3Endpoint() (v string) {
	st.mutex.RLock()
	v = st.config.StorageS3Endpoint
	st.mutex.RUnlock()
	return
}

// SetStorageS3Endpoint safely sets the Configuration value for state's 'StorageS3Endpoint' field
func (st *ConfigState) SetStorageS3Endpoint(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3Endpoint = v
	st.reloadToViper()
}

// GetStorageS3Endpoint safely fetches the value for global configuration 'StorageS3Endpoint' field
func GetStorageS3Endpoint() string { return global.GetStorageS3Endpoint() }

// SetStorageS3Endpoint safely sets the value for global configuration 'StorageS3Endpoint' field
func SetStorageS3Endpoint(v string) { global.SetStorageS3Endpoint(v) }

// GetStorageS3AccessKey safely fetches the Configuration value for state's 'StorageS3AccessKey' field
func (st *ConfigState) GetStorageS3AccessKey() (v string) {
	st.mutex.RLock()
	v = st.config.StorageS3AccessKey
	st.mutex.RUnlock()
	return
}

// SetStorageS3AccessKey safely sets the Configuration value for state's 'StorageS3AccessKey' field
func (st *ConfigState) SetStorageS3AccessKey(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3AccessKey = v
	st.reloadToViper()
}

// GetStorageS3AccessKey safely fetches the value for global configuration 'StorageS3AccessKey' field
func GetStorageS3AccessKey() string { return global.GetStorageS3AccessKey() }

// SetStorageS3AccessKey safely sets the value for global configuration 'StorageS3AccessKey' field
func SetStorageS3AccessKey(v string) { global.SetStorageS3AccessKey(v) }

// GetStorageS3SecretKey safely fetches the Configuration value for state's 'StorageS3SecretKey' field
func (st *ConfigState) GetStorageS3SecretKey() (v string) {
	st.mutex.RLock()
	v = st.config.StorageS3SecretKey
	st.mutex.RUnlock()
	return
}

// SetStorageS3SecretKey safely sets the Configuration value for state's 'StorageS3SecretKey' field
func (st *ConfigState) SetStorageS3SecretKey(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3SecretKey = v
	st.reloadToViper()
}

// GetStorageS3SecretKey safely fetches the value for global configuration 'StorageS3SecretKey' field
func GetStorageS3SecretKey() string { return global.GetStorageS3SecretKey() }

// SetStorageS3SecretKey safely sets the value for global configuration 'StorageS3SecretKey' field
func SetStorageS3SecretKey(v string) { global.SetStorageS3SecretKey(v) }

// GetStorageS3UseSSL safely fetches the Configuration value for state's 'StorageS3UseSSL' field
func (st *ConfigState) GetStorageS3UseSSL() (v bool) {
	st.mutex.RLock()
	v = st.config.StorageS3UseSSL
	st.mutex.RUnlock()
	return
}

// SetStorageS3UseSSL safely sets the Configuration value for state's 'StorageS3UseSSL' field
func (st *ConfigState) SetStorageS3UseSSL(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3UseSSL = v
	st.reloadToViper()
}

// GetStorageS3UseSSL safely fetches the value for global configuration 'StorageS3UseSSL' field
func GetStorageS3UseSSL() bool { return global.GetStorageS3UseSSL() }

// SetStorageS3UseSSL safely sets the value for global configuration 'StorageS3UseSSL' field
func SetStorageS3UseSSL(v bool) { global.SetStorageS3UseSSL(v) }

// GetStorageS3BucketName safely fetches the Configuration value for state's 'StorageS3BucketName' field
func (st *ConfigState) GetStorageS3BucketName() (v string) {
	st.mutex.RLock()
	v = st.config.StorageS3BucketName
	st.mutex.RUnlock()
	return
}

// SetStorageS3BucketName safely sets the Configuration value for state's 'StorageS3BucketName' field
func (st *ConfigState) SetStorageS3BucketName(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3BucketName = v
	st.reloadToViper()
}

// GetStorageS3BucketName safely fetches the value for global configuration 'StorageS3BucketName' field
func GetStorageS3BucketName() string { return global.GetStorageS3BucketName() }

// SetStorageS3BucketName safely sets the value for global configuration 'StorageS3BucketName' field
func SetStorageS3BucketName(v string) { global.SetStorageS3BucketName(v) }

// GetStorageS3Proxy safely fetches the Configuration value for state's 'StorageS3Proxy' field
func (st *ConfigState) GetStorageS3Proxy() (v bool) {
	st.mutex.RLock()
	v = st.config.StorageS3Proxy
	st.mutex.RUnlock()
	return
}

// SetStorageS3Proxy safely sets the Configuration value for state's 'StorageS3Proxy' field
func (st *ConfigState) SetStorageS3Proxy(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3Proxy = v
	st.reloadToViper()
}

// GetStorageS3Proxy safely fetches the value for global configuration 'StorageS3Proxy' field
func GetStorageS3Proxy() bool { return global.GetStorageS3Proxy() }

// SetStorageS3Proxy safely sets the value for global configuration 'StorageS3Proxy' field
func SetStorageS3Proxy(v bool) { global.SetStorageS3Proxy(v) }

// GetStorageS3RedirectURL safely fetches the Configuration value for state's 'StorageS3RedirectURL' field
func (st *ConfigState) GetStorageS3RedirectURL() (v string) {
	st.mutex.RLock()
	v = st.config.StorageS3RedirectURL
	st.mutex.RUnlock()
	return
}

// SetStorageS3RedirectURL safely sets the Configuration value for state's 'StorageS3RedirectURL' field
func (st *ConfigState) SetStorageS3RedirectURL(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3RedirectURL = v
	st.reloadToViper()
}

// GetStorageS3RedirectURL safely fetches the value for global configuration 'StorageS3RedirectURL' field
func GetStorageS3RedirectURL() string { return global.GetStorageS3RedirectURL() }

// SetStorageS3RedirectURL safely sets the value for global configuration 'StorageS3RedirectURL' field
func SetStorageS3RedirectURL(v string) { global.SetStorageS3RedirectURL(v) }

// GetStorageS3BucketLookup safely fetches the Configuration value for state's 'StorageS3BucketLookup' field
func (st *ConfigState) GetStorageS3BucketLookup() (v string) {
	st.mutex.RLock()
	v = st.config.StorageS3BucketLookup
	st.mutex.RUnlock()
	return
}

// SetStorageS3BucketLookup safely sets the Configuration value for state's 'StorageS3BucketLookup' field
func (st *ConfigState) SetStorageS3BucketLookup(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3BucketLookup = v
	st.reloadToViper()
}

// GetStorageS3BucketLookup safely fetches the value for global configuration 'StorageS3BucketLookup' field
func GetStorageS3BucketLookup() string { return global.GetStorageS3BucketLookup() }

// SetStorageS3BucketLookup safely sets the value for global configuration 'StorageS3BucketLookup' field
func SetStorageS3BucketLookup(v string) { global.SetStorageS3BucketLookup(v) }

// GetStorageS3KeyPrefix safely fetches the Configuration value for state's 'StorageS3KeyPrefix' field
func (st *ConfigState) GetStorageS3KeyPrefix() (v string) {
	st.mutex.RLock()
	v = st.config.StorageS3KeyPrefix
	st.mutex.RUnlock()
	return
}

// SetStorageS3KeyPrefix safely sets the Configuration value for state's 'StorageS3KeyPrefix' field
func (st *ConfigState) SetStorageS3KeyPrefix(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3KeyPrefix = v
	st.reloadToViper()
}

// GetStorageS3KeyPrefix safely fetches the value for global configuration 'StorageS3KeyPrefix' field
func GetStorageS3KeyPrefix() string { return global.GetStorageS3KeyPrefix() }

// SetStorageS3KeyPrefix safely sets the value for global configuration 'StorageS3KeyPrefix' field
func SetStorageS3KeyPrefix(v string) { global.SetStorageS3KeyPrefix(v) }

// GetStatusesMaxChars safely fetches the Configuration value for state's 'StatusesMaxChars' field
func (st *ConfigState) GetStatusesMaxChars() (v int) {
	st.mutex.RLock()
	v = st.config.StatusesMaxChars
	st.mutex.RUnlock()
	return
}

// SetStatusesMaxChars safely sets the Configuration value for state's 'StatusesMaxChars' field
func (st *ConfigState) SetStatusesMaxChars(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StatusesMaxChars = v
	st.reloadToViper()
}

// GetStatusesMaxChars safely fetches the value for global configuration 'StatusesMaxChars' field
func GetStatusesMaxChars() int { return global.GetStatusesMaxChars() }

// SetStatusesMaxChars safely sets the value for global configuration 'StatusesMaxChars' field
func SetStatusesMaxChars(v int) { global.SetStatusesMaxChars(v) }

// GetStatusesPollMaxOptions safely fetches the Configuration value for state's 'StatusesPollMaxOptions' field
func (st *ConfigState) GetStatusesPollMaxOptions() (v int) {
	st.mutex.RLock()
	v = st.config.StatusesPollMaxOptions
	st.mutex.RUnlock()
	return
}

// SetStatusesPollMaxOptions safely sets the Configuration value for state's 'StatusesPollMaxOptions' field
func (st *ConfigState) SetStatusesPollMaxOptions(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StatusesPollMaxOptions = v
	st.reloadToViper()
}

// GetStatusesPollMaxOptions safely fetches the value for global configuration 'StatusesPollMaxOptions' field
func GetStatusesPollMaxOptions() int { return global.GetStatusesPollMaxOptions() }

// SetStatusesPollMaxOptions safely sets the value for global configuration 'StatusesPollMaxOptions' field
func SetStatusesPollMaxOptions(v int) { global.SetStatusesPollMaxOptions(v) }

// GetStatusesPollOptionMaxChars safely fetches the Configuration value for state's 'StatusesPollOptionMaxChars' field
func (st *ConfigState) GetStatusesPollOptionMaxChars() (v int) {
	st.mutex.RLock()
	v = st.config.StatusesPollOptionMaxChars
	st.mutex.RUnlock()
	return
}

// SetStatusesPollOptionMaxChars safely sets the Configuration value for state's 'StatusesPollOptionMaxChars' field
func (st *ConfigState) SetStatusesPollOptionMaxChars(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StatusesPollOptionMaxChars = v
	st.reloadToViper()
}

// GetStatusesPollOptionMaxChars safely fetches the value for global configuration 'StatusesPollOptionMaxChars' field
func GetStatusesPollOptionMaxChars() int { return global.GetStatusesPollOptionMaxChars() }

// SetStatusesPollOptionMaxChars safely sets the value for global configuration 'StatusesPollOptionMaxChars' field
func SetStatusesPollOptionMaxChars(v int) { global.SetStatusesPollOptionMaxChars(v) }

// GetStatusesMediaMaxFiles safely fetches the Configuration value for state's 'StatusesMediaMaxFiles' field
func (st *ConfigState) GetStatusesMediaMaxFiles() (v int) {
	st.mutex.RLock()
	v = st.config.StatusesMediaMaxFiles
	st.mutex.RUnlock()
	return
}

// SetStatusesMediaMaxFiles safely sets the Configuration value for state's 'StatusesMediaMaxFiles' field
func (st *ConfigState) SetStatusesMediaMaxFiles(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StatusesMediaMaxFiles = v
	st.reloadToViper()
}

// GetStatusesMediaMaxFiles safely fetches the value for global configuration 'StatusesMediaMaxFiles' field
func GetStatusesMediaMaxFiles() int { return global.GetStatusesMediaMaxFiles() }

// SetStatusesMediaMaxFiles safely sets the value for global configuration 'StatusesMediaMaxFiles' field
func SetStatusesMediaMaxFiles(v int) { global.SetStatusesMediaMaxFiles(v) }

// GetScheduledStatusesMaxTotal safely fetches the Configuration value for state's 'ScheduledStatusesMaxTotal' field
func (st *ConfigState) GetScheduledStatusesMaxTotal() (v int) {
	st.mutex.RLock()
	v = st.config.ScheduledStatusesMaxTotal
	st.mutex.RUnlock()
	return
}

// SetScheduledStatusesMaxTotal safely sets the Configuration value for state's 'ScheduledStatusesMaxTotal' field
func (st *ConfigState) SetScheduledStatusesMaxTotal(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.ScheduledStatusesMaxTotal = v
	st.reloadToViper()
}

// GetScheduledStatusesMaxTotal safely fetches the value for global configuration 'ScheduledStatusesMaxTotal' field
func GetScheduledStatusesMaxTotal() int { return global.GetScheduledStatusesMaxTotal() }

// SetScheduledStatusesMaxTotal safely sets the value for global configuration 'ScheduledStatusesMaxTotal' field
func SetScheduledStatusesMaxTotal(v int) { global.SetScheduledStatusesMaxTotal(v) }

// GetScheduledStatusesMaxDaily safely fetches the Configuration value for state's 'ScheduledStatusesMaxDaily' field
func (st *ConfigState) GetScheduledStatusesMaxDaily() (v int) {
	st.mutex.RLock()
	v = st.config.ScheduledStatusesMaxDaily
	st.mutex.RUnlock()
	return
}

// SetScheduledStatusesMaxDaily safely sets the Configuration value for state's 'ScheduledStatusesMaxDaily' field
func (st *ConfigState) SetScheduledStatusesMaxDaily(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.ScheduledStatusesMaxDaily = v
	st.reloadToViper()
}

// GetScheduledStatusesMaxDaily safely fetches the value for global configuration 'ScheduledStatusesMaxDaily' field
func GetScheduledStatusesMaxDaily() int { return global.GetScheduledStatusesMaxDaily() }

// SetScheduledStatusesMaxDaily safely sets the value for global configuration 'ScheduledStatusesMaxDaily' field
func SetScheduledStatusesMaxDaily(v int) { global.SetScheduledStatusesMaxDaily(v) }

// GetLetsEncryptEnabled safely fetches the Configuration value for state's 'LetsEncryptEnabled' field
func (st *ConfigState) GetLetsEncryptEnabled() (v bool) {
	st.mutex.RLock()
	v = st.config.LetsEncryptEnabled
	st.mutex.RUnlock()
	return
}

// SetLetsEncryptEnabled safely sets the Configuration value for state's 'LetsEncryptEnabled' field
func (st *ConfigState) SetLetsEncryptEnabled(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LetsEncryptEnabled = v
	st.reloadToViper()
}

// GetLetsEncryptEnabled safely fetches the value for global configuration 'LetsEncryptEnabled' field
func GetLetsEncryptEnabled() bool { return global.GetLetsEncryptEnabled() }

// SetLetsEncryptEnabled safely sets the value for global configuration 'LetsEncryptEnabled' field
func SetLetsEncryptEnabled(v bool) { global.SetLetsEncryptEnabled(v) }

// GetLetsEncryptPort safely fetches the Configuration value for state's 'LetsEncryptPort' field
func (st *ConfigState) GetLetsEncryptPort() (v int) {
	st.mutex.RLock()
	v = st.config.LetsEncryptPort
	st.mutex.RUnlock()
	return
}

// SetLetsEncryptPort safely sets the Configuration value for state's 'LetsEncryptPort' field
func (st *ConfigState) SetLetsEncryptPort(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LetsEncryptPort = v
	st.reloadToViper()
}

// GetLetsEncryptPort safely fetches the value for global configuration 'LetsEncryptPort' field
func GetLetsEncryptPort() int { return global.GetLetsEncryptPort() }

// SetLetsEncryptPort safely sets the value for global configuration 'LetsEncryptPort' field
func SetLetsEncryptPort(v int) { global.SetLetsEncryptPort(v) }

// GetLetsEncryptCertDir safely fetches the Configuration value for state's 'LetsEncryptCertDir' field
func (st *ConfigState) GetLetsEncryptCertDir() (v string) {
	st.mutex.RLock()
	v = st.config.LetsEncryptCertDir
	st.mutex.RUnlock()
	return
}

// SetLetsEncryptCertDir safely sets the Configuration value for state's 'LetsEncryptCertDir' field
func (st *ConfigState) SetLetsEncryptCertDir(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LetsEncryptCertDir = v
	st.reloadToViper()
}

// GetLetsEncryptCertDir safely fetches the value for global configuration 'LetsEncryptCertDir' field
func GetLetsEncryptCertDir() string { return global.GetLetsEncryptCertDir() }

// SetLetsEncryptCertDir safely sets the value for global configuration 'LetsEncryptCertDir' field
func SetLetsEncryptCertDir(v string) { global.SetLetsEncryptCertDir(v) }

// GetLetsEncryptEmailAddress safely fetches the Configuration value for state's 'LetsEncryptEmailAddress' field
func (st *ConfigState) GetLetsEncryptEmailAddress() (v string) {
	st.mutex.RLock()
	v = st.config.LetsEncryptEmailAddress
	st.mutex.RUnlock()
	return
}

// SetLetsEncryptEmailAddress safely sets the Configuration value for state's 'LetsEncryptEmailAddress' field
func (st *ConfigState) SetLetsEncryptEmailAddress(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LetsEncryptEmailAddress = v
	st.reloadToViper()
}

// GetLetsEncryptEmailAddress safely fetches the value for global configuration 'LetsEncryptEmailAddress' field
func GetLetsEncryptEmailAddress() string { return global.GetLetsEncryptEmailAddress() }

// SetLetsEncryptEmailAddress safely sets the value for global configuration 'LetsEncryptEmailAddress' field
func SetLetsEncryptEmailAddress(v string) { global.SetLetsEncryptEmailAddress(v) }

// GetTLSCertificateChain safely fetches the Configuration value for state's 'TLSCertificateChain' field
func (st *ConfigState) GetTLSCertificateChain() (v string) {
	st.mutex.RLock()
	v = st.config.TLSCertificateChain
	st.mutex.RUnlock()
	return
}

// SetTLSCertificateChain safely sets the Configuration value for state's 'TLSCertificateChain' field
func (st *ConfigState) SetTLSCertificateChain(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.TLSCertificateChain = v
	st.reloadToViper()
}

// GetTLSCertificateChain safely fetches the value for global configuration 'TLSCertificateChain' field
func GetTLSCertificateChain() string { return global.GetTLSCertificateChain() }

// SetTLSCertificateChain safely sets the value for global configuration 'TLSCertificateChain' field
func SetTLSCertificateChain(v string) { global.SetTLSCertificateChain(v) }

// GetTLSCertificateKey safely fetches the Configuration value for state's 'TLSCertificateKey' field
func (st *ConfigState) GetTLSCertificateKey() (v string) {
	st.mutex.RLock()
	v = st.config.TLSCertificateKey
	st.mutex.RUnlock()
	return
}

// SetTLSCertificateKey safely sets the Configuration value for state's 'TLSCertificateKey' field
func (st *ConfigState) SetTLSCertificateKey(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.TLSCertificateKey = v
	st.reloadToViper()
}

// GetTLSCertificateKey safely fetches the value for global configuration 'TLSCertificateKey' field
func GetTLSCertificateKey() string { return global.GetTLSCertificateKey() }

// SetTLSCertificateKey safely sets the value for global configuration 'TLSCertificateKey' field
func SetTLSCertificateKey(v string) { global.SetTLSCertificateKey(v) }

// GetOIDCEnabled safely fetches the Configuration value for state's 'OIDCEnabled' field
func (st *ConfigState) GetOIDCEnabled() (v bool) {
	st.mutex.RLock()
	v = st.config.OIDCEnabled
	st.mutex.RUnlock()
	return
}

// SetOIDCEnabled safely sets the Configuration value for state's 'OIDCEnabled' field
func (st *ConfigState) SetOIDCEnabled(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCEnabled = v
	st.reloadToViper()
}

// GetOIDCEnabled safely fetches the value for global configuration 'OIDCEnabled' field
func GetOIDCEnabled() bool { return global.GetOIDCEnabled() }

// SetOIDCEnabled safely sets the value for global configuration 'OIDCEnabled' field
func SetOIDCEnabled(v bool) { global.SetOIDCEnabled(v) }

// GetOIDCIdpName safely fetches the Configuration value for state's 'OIDCIdpName' field
func (st *ConfigState) GetOIDCIdpName() (v string) {
	st.mutex.RLock()
	v = st.config.OIDCIdpName
	st.mutex.RUnlock()
	return
}

// SetOIDCIdpName safely sets the Configuration value for state's 'OIDCIdpName' field
func (st *ConfigState) SetOIDCIdpName(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCIdpName = v
	st.reloadToViper()
}

// GetOIDCIdpName safely fetches the value for global configuration 'OIDCIdpName' field
func GetOIDCIdpName() string { return global.GetOIDCIdpName() }

// SetOIDCIdpName safely sets the value for global configuration 'OIDCIdpName' field
func SetOIDCIdpName(v string) { global.SetOIDCIdpName(v) }

// GetOIDCSkipVerification safely fetches the Configuration value for state's 'OIDCSkipVerification' field
func (st *ConfigState) GetOIDCSkipVerification() (v bool) {
	st.mutex.RLock()
	v = st.config.OIDCSkipVerification
	st.mutex.RUnlock()
	return
}

// SetOIDCSkipVerification safely sets the Configuration value for state's 'OIDCSkipVerification' field
func (st *ConfigState) SetOIDCSkipVerification(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCSkipVerification = v
	st.reloadToViper()
}

// GetOIDCSkipVerification safely fetches the value for global configuration 'OIDCSkipVerification' field
func GetOIDCSkipVerification() bool { return global.GetOIDCSkipVerification() }

// SetOIDCSkipVerification safely sets the value for global configuration 'OIDCSkipVerification' field
func SetOIDCSkipVerification(v bool) { global.SetOIDCSkipVerification(v) }

// GetOIDCIssuer safely fetches the Configuration value for state's 'OIDCIssuer' field
func (st *ConfigState) GetOIDCIssuer() (v string) {
	st.mutex.RLock()
	v = st.config.OIDCIssuer
	st.mutex.RUnlock()
	return
}

// SetOIDCIssuer safely sets the Configuration value for state's 'OIDCIssuer' field
func (st *ConfigState) SetOIDCIssuer(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCIssuer = v
	st.reloadToViper()
}

// GetOIDCIssuer safely fetches the value for global configuration 'OIDCIssuer' field
func GetOIDCIssuer() string { return global.GetOIDCIssuer() }

// SetOIDCIssuer safely sets the value for global configuration 'OIDCIssuer' field
func SetOIDCIssuer(v string) { global.SetOIDCIssuer(v) }

// GetOIDCClientID safely fetches the Configuration value for state's 'OIDCClientID' field
func (st *ConfigState) GetOIDCClientID() (v string) {
	st.mutex.RLock()
	v = st.config.OIDCClientID
	st.mutex.RUnlock()
	return
}

// SetOIDCClientID safely sets the Configuration value for state's 'OIDCClientID' field
func (st *ConfigState) SetOIDCClientID(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCClientID = v
	st.reloadToViper()
}

// GetOIDCClientID safely fetches the value for global configuration 'OIDCClientID' field
func GetOIDCClientID() string { return global.GetOIDCClientID() }

// SetOIDCClientID safely sets the value for global configuration 'OIDCClientID' field
func SetOIDCClientID(v string) { global.SetOIDCClientID(v) }

// GetOIDCClientSecret safely fetches the Configuration value for state's 'OIDCClientSecret' field
func (st *ConfigState) GetOIDCClientSecret() (v string) {
	st.mutex.RLock()
	v = st.config.OIDCClientSecret
	st.mutex.RUnlock()
	return
}

// SetOIDCClientSecret safely sets the Configuration value for state's 'OIDCClientSecret' field
func (st *ConfigState) SetOIDCClientSecret(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCClientSecret = v
	st.reloadToViper()
}

// GetOIDCClientSecret safely fetches the value for global configuration 'OIDCClientSecret' field
func GetOIDCClientSecret() string { return global.GetOIDCClientSecret() }

// SetOIDCClientSecret safely sets the value for global configuration 'OIDCClientSecret' field
func SetOIDCClientSecret(v string) { global.SetOIDCClientSecret(v) }

// GetOIDCScopes safely fetches the Configuration value for state's 'OIDCScopes' field
func (st *ConfigState) GetOIDCScopes() (v []string) {
	st.mutex.RLock()
	v = st.config.OIDCScopes
	st.mutex.RUnlock()
	return
}

// SetOIDCScopes safely sets the Configuration value for state's 'OIDCScopes' field
func (st *ConfigState) SetOIDCScopes(v []string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCScopes = v
	st.reloadToViper()
}

// GetOIDCScopes safely fetches the value for global configuration 'OIDCScopes' field
func GetOIDCScopes() []string { return global.GetOIDCScopes() }

// SetOIDCScopes safely sets the value for global configuration 'OIDCScopes' field
func SetOIDCScopes(v []string) { global.SetOIDCScopes(v) }

// GetOIDCLinkExisting safely fetches the Configuration value for state's 'OIDCLinkExisting' field
func (st *ConfigState) GetOIDCLinkExisting() (v bool) {
	st.mutex.RLock()
	v = st.config.OIDCLinkExisting
	st.mutex.RUnlock()
	return
}

// SetOIDCLinkExisting safely sets the Configuration value for state's 'OIDCLinkExisting' field
func (st *ConfigState) SetOIDCLinkExisting(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCLinkExisting = v
	st.reloadToViper()
}

// GetOIDCLinkExisting safely fetches the value for global configuration 'OIDCLinkExisting' field
func GetOIDCLinkExisting() bool { return global.GetOIDCLinkExisting() }

// SetOIDCLinkExisting safely sets the value for global configuration 'OIDCLinkExisting' field
func SetOIDCLinkExisting(v bool) { global.SetOIDCLinkExisting(v) }

// GetOIDCAllowedGroups safely fetches the Configuration value for state's 'OIDCAllowedGroups' field
func (st *ConfigState) GetOIDCAllowedGroups() (v []string) {
	st.mutex.RLock()
	v = st.config.OIDCAllowedGroups
	st.mutex.RUnlock()
	return
}

// SetOIDCAllowedGroups safely sets the Configuration value for state's 'OIDCAllowedGroups' field
func (st *ConfigState) SetOIDCAllowedGroups(v []string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCAllowedGroups = v
	st.reloadToViper()
}

// GetOIDCAllowedGroups safely fetches the value for global configuration 'OIDCAllowedGroups' field
func GetOIDCAllowedGroups() []string { return global.GetOIDCAllowedGroups() }

// SetOIDCAllowedGroups safely sets the value for global configuration 'OIDCAllowedGroups' field
func SetOIDCAllowedGroups(v []string) { global.SetOIDCAllowedGroups(v) }

// GetOIDCAdminGroups safely fetches the Configuration value for state's 'OIDCAdminGroups' field
func (st *ConfigState) GetOIDCAdminGroups() (v []string) {
	st.mutex.RLock()
	v = st.config.OIDCAdminGroups
	st.mutex.RUnlock()
	return
}

// SetOIDCAdminGroups safely sets the Configuration value for state's 'OIDCAdminGroups' field
func (st *ConfigState) SetOIDCAdminGroups(v []string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCAdminGroups = v
	st.reloadToViper()
}

// GetOIDCAdminGroups safely fetches the value for global configuration 'OIDCAdminGroups' field
func GetOIDCAdminGroups() []string { return global.GetOIDCAdminGroups() }

// SetOIDCAdminGroups safely sets the value for global configuration 'OIDCAdminGroups' field
func SetOIDCAdminGroups(v []string) { global.SetOIDCAdminGroups(v) }

// GetTracingEnabled safely fetches the Configuration value for state's 'TracingEnabled' field
func (st *ConfigState) GetTracingEnabled() (v bool) {
	st.mutex.RLock()
	v = st.config.TracingEnabled
	st.mutex.RUnlock()
	return
}

// SetTracingEnabled safely sets the Configuration value for state's 'TracingEnabled' field
func (st *ConfigState) SetTracingEnabled(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.TracingEnabled = v
	st.reloadToViper()
}

// GetTracingEnabled safely fetches the value for global configuration 'TracingEnabled' field
func GetTracingEnabled() bool { return global.GetTracingEnabled() }

// SetTracingEnabled safely sets the value for global configuration 'TracingEnabled' field
func SetTracingEnabled(v bool) { global.SetTracingEnabled(v) }

// GetMetricsEnabled safely fetches the Configuration value for state's 'MetricsEnabled' field
func (st *ConfigState) GetMetricsEnabled() (v bool) {
	st.mutex.RLock()
	v = st.config.MetricsEnabled
	st.mutex.RUnlock()
	return
}

// SetMetricsEnabled safely sets the Configuration value for state's 'MetricsEnabled' field
func (st *ConfigState) SetMetricsEnabled(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MetricsEnabled = v
	st.reloadToViper()
}

// GetMetricsEnabled safely fetches the value for global configuration 'MetricsEnabled' field
func GetMetricsEnabled() bool { return global.GetMetricsEnabled() }

// SetMetricsEnabled safely sets the value for global configuration 'MetricsEnabled' field
func SetMetricsEnabled(v bool) { global.SetMetricsEnabled(v) }

// GetSMTPHost safely fetches the Configuration value for state's 'SMTPHost' field
func (st *ConfigState) GetSMTPHost() (v string) {
	st.mutex.RLock()
	v = st.config.SMTPHost
	st.mutex.RUnlock()
	return
}

// SetSMTPHost safely sets the Configuration value for state's 'SMTPHost' field
func (st *ConfigState) SetSMTPHost(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SMTPHost = v
	st.reloadToViper()
}

// GetSMTPHost safely fetches the value for global configuration 'SMTPHost' field
func GetSMTPHost() string { return global.GetSMTPHost() }

// SetSMTPHost safely sets the value for global configuration 'SMTPHost' field
func SetSMTPHost(v string) { global.SetSMTPHost(v) }

// GetSMTPPort safely fetches the Configuration value for state's 'SMTPPort' field
func (st *ConfigState) GetSMTPPort() (v int) {
	st.mutex.RLock()
	v = st.config.SMTPPort
	st.mutex.RUnlock()
	return
}

// SetSMTPPort safely sets the Configuration value for state's 'SMTPPort' field
func (st *ConfigState) SetSMTPPort(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SMTPPort = v
	st.reloadToViper()
}

// GetSMTPPort safely fetches the value for global configuration 'SMTPPort' field
func GetSMTPPort() int { return global.GetSMTPPort() }

// SetSMTPPort safely sets the value for global configuration 'SMTPPort' field
func SetSMTPPort(v int) { global.SetSMTPPort(v) }

// GetSMTPUsername safely fetches the Configuration value for state's 'SMTPUsername' field
func (st *ConfigState) GetSMTPUsername() (v string) {
	st.mutex.RLock()
	v = st.config.SMTPUsername
	st.mutex.RUnlock()
	return
}

// SetSMTPUsername safely sets the Configuration value for state's 'SMTPUsername' field
func (st *ConfigState) SetSMTPUsername(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SMTPUsername = v
	st.reloadToViper()
}

// GetSMTPUsername safely fetches the value for global configuration 'SMTPUsername' field
func GetSMTPUsername() string { return global.GetSMTPUsername() }

// SetSMTPUsername safely sets the value for global configuration 'SMTPUsername' field
func SetSMTPUsername(v string) { global.SetSMTPUsername(v) }

// GetSMTPPassword safely fetches the Configuration value for state's 'SMTPPassword' field
func (st *ConfigState) GetSMTPPassword() (v string) {
	st.mutex.RLock()
	v = st.config.SMTPPassword
	st.mutex.RUnlock()
	return
}

// SetSMTPPassword safely sets the Configuration value for state's 'SMTPPassword' field
func (st *ConfigState) SetSMTPPassword(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SMTPPassword = v
	st.reloadToViper()
}

// GetSMTPPassword safely fetches the value for global configuration 'SMTPPassword' field
func GetSMTPPassword() string { return global.GetSMTPPassword() }

// SetSMTPPassword safely sets the value for global configuration 'SMTPPassword' field
func SetSMTPPassword(v string) { global.SetSMTPPassword(v) }

// GetSMTPFrom safely fetches the Configuration value for state's 'SMTPFrom' field
func (st *ConfigState) GetSMTPFrom() (v string) {
	st.mutex.RLock()
	v = st.config.SMTPFrom
	st.mutex.RUnlock()
	return
}

// SetSMTPFrom safely sets the Configuration value for state's 'SMTPFrom' field
func (st *ConfigState) SetSMTPFrom(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SMTPFrom = v
	st.reloadToViper()
}

// GetSMTPFrom safely fetches the value for global configuration 'SMTPFrom' field
func GetSMTPFrom() string { return global.GetSMTPFrom() }

// SetSMTPFrom safely sets the value for global configuration 'SMTPFrom' field
func SetSMTPFrom(v string) { global.SetSMTPFrom(v) }

// GetSMTPDiscloseRecipients safely fetches the Configuration value for state's 'SMTPDiscloseRecipients' field
func (st *ConfigState) GetSMTPDiscloseRecipients() (v bool) {
	st.mutex.RLock()
	v = st.config.SMTPDiscloseRecipients
	st.mutex.RUnlock()
	return
}

// SetSMTPDiscloseRecipients safely sets the Configuration value for state's 'SMTPDiscloseRecipients' field
func (st *ConfigState) SetSMTPDiscloseRecipients(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SMTPDiscloseRecipients = v
	st.reloadToViper()
}

// GetSMTPDiscloseRecipients safely fetches the value for global configuration 'SMTPDiscloseRecipients' field
func GetSMTPDiscloseRecipients() bool { return global.GetSMTPDiscloseRecipients() }

// SetSMTPDiscloseRecipients safely sets the value for global configuration 'SMTPDiscloseRecipients' field
func SetSMTPDiscloseRecipients(v bool) { global.SetSMTPDiscloseRecipients(v) }

// GetSyslogEnabled safely fetches the Configuration value for state's 'SyslogEnabled' field
func (st *ConfigState) GetSyslogEnabled() (v bool) {
	st.mutex.RLock()
	v = st.config.SyslogEnabled
	st.mutex.RUnlock()
	return
}

// SetSyslogEnabled safely sets the Configuration value for state's 'SyslogEnabled' field
func (st *ConfigState) SetSyslogEnabled(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SyslogEnabled = v
	st.reloadToViper()
}

// GetSyslogEnabled safely fetches the value for global configuration 'SyslogEnabled' field
func GetSyslogEnabled() bool { return global.GetSyslogEnabled() }

// SetSyslogEnabled safely sets the value for global configuration 'SyslogEnabled' field
func SetSyslogEnabled(v bool) { global.SetSyslogEnabled(v) }

// GetSyslogProtocol safely fetches the Configuration value for state's 'SyslogProtocol' field
func (st *ConfigState) GetSyslogProtocol() (v string) {
	st.mutex.RLock()
	v = st.config.SyslogProtocol
	st.mutex.RUnlock()
	return
}

// SetSyslogProtocol safely sets the Configuration value for state's 'SyslogProtocol' field
func (st *ConfigState) SetSyslogProtocol(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SyslogProtocol = v
	st.reloadToViper()
}

// GetSyslogProtocol safely fetches the value for global configuration 'SyslogProtocol' field
func GetSyslogProtocol() string { return global.GetSyslogProtocol() }

// SetSyslogProtocol safely sets the value for global configuration 'SyslogProtocol' field
func SetSyslogProtocol(v string) { global.SetSyslogProtocol(v) }

// GetSyslogAddress safely fetches the Configuration value for state's 'SyslogAddress' field
func (st *ConfigState) GetSyslogAddress() (v string) {
	st.mutex.RLock()
	v = st.config.SyslogAddress
	st.mutex.RUnlock()
	return
}

// SetSyslogAddress safely sets the Configuration value for state's 'SyslogAddress' field
func (st *ConfigState) SetSyslogAddress(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SyslogAddress = v
	st.reloadToViper()
}

// GetSyslogAddress safely fetches the value for global configuration 'SyslogAddress' field
func GetSyslogAddress() string { return global.GetSyslogAddress() }

// SetSyslogAddress safely sets the value for global configuration 'SyslogAddress' field
func SetSyslogAddress(v string) { global.SetSyslogAddress(v) }

// GetAdvancedCookiesSamesite safely fetches the Configuration value for state's 'Advanced.CookiesSamesite' field
func (st *ConfigState) GetAdvancedCookiesSamesite() (v string) {
	st.mutex.RLock()
	v = st.config.Advanced.CookiesSamesite
	st.mutex.RUnlock()
	return
}

// SetAdvancedCookiesSamesite safely sets the Configuration value for state's 'Advanced.CookiesSamesite' field
func (st *ConfigState) SetAdvancedCookiesSamesite(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Advanced.CookiesSamesite = v
	st.reloadToViper()
}

// GetAdvancedCookiesSamesite safely fetches the value for global configuration 'Advanced.CookiesSamesite' field
func GetAdvancedCookiesSamesite() string { return global.GetAdvancedCookiesSamesite() }

// SetAdvancedCookiesSamesite safely sets the value for global configuration 'Advanced.CookiesSamesite' field
func SetAdvancedCookiesSamesite(v string) { global.SetAdvancedCookiesSamesite(v) }

// GetAdvancedSenderMultiplier safely fetches the Configuration value for state's 'Advanced.SenderMultiplier' field
func (st *ConfigState) GetAdvancedSenderMultiplier() (v int) {
	st.mutex.RLock()
	v = st.config.Advanced.SenderMultiplier
	st.mutex.RUnlock()
	return
}

// SetAdvancedSenderMultiplier safely sets the Configuration value for state's 'Advanced.SenderMultiplier' field
func (st *ConfigState) SetAdvancedSenderMultiplier(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Advanced.SenderMultiplier = v
	st.reloadToViper()
}

// GetAdvancedSenderMultiplier safely fetches the value for global configuration 'Advanced.SenderMultiplier' field
func GetAdvancedSenderMultiplier() int { return global.GetAdvancedSenderMultiplier() }

// SetAdvancedSenderMultiplier safely sets the value for global configuration 'Advanced.SenderMultiplier' field
func SetAdvancedSenderMultiplier(v int) { global.SetAdvancedSenderMultiplier(v) }

// GetAdvancedCSPExtraURIs safely fetches the Configuration value for state's 'Advanced.CSPExtraURIs' field
func (st *ConfigState) GetAdvancedCSPExtraURIs() (v []string) {
	st.mutex.RLock()
	v = st.config.Advanced.CSPExtraURIs
	st.mutex.RUnlock()
	return
}

// SetAdvancedCSPExtraURIs safely sets the Configuration value for state's 'Advanced.CSPExtraURIs' field
func (st *ConfigState) SetAdvancedCSPExtraURIs(v []string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Advanced.CSPExtraURIs = v
	st.reloadToViper()
}

// GetAdvancedCSPExtraURIs safely fetches the value for global configuration 'Advanced.CSPExtraURIs' field
func GetAdvancedCSPExtraURIs() []string { return global.GetAdvancedCSPExtraURIs() }

// SetAdvancedCSPExtraURIs safely sets the value for global configuration 'Advanced.CSPExtraURIs' field
func SetAdvancedCSPExtraURIs(v []string) { global.SetAdvancedCSPExtraURIs(v) }

// GetAdvancedHeaderFilterMode safely fetches the Configuration value for state's 'Advanced.HeaderFilterMode' field
func (st *ConfigState) GetAdvancedHeaderFilterMode() (v string) {
	st.mutex.RLock()
	v = st.config.Advanced.HeaderFilterMode
	st.mutex.RUnlock()
	return
}

// SetAdvancedHeaderFilterMode safely sets the Configuration value for state's 'Advanced.HeaderFilterMode' field
func (st *ConfigState) SetAdvancedHeaderFilterMode(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Advanced.HeaderFilterMode = v
	st.reloadToViper()
}

// GetAdvancedHeaderFilterMode safely fetches the value for global configuration 'Advanced.HeaderFilterMode' field
func GetAdvancedHeaderFilterMode() string { return global.GetAdvancedHeaderFilterMode() }

// SetAdvancedHeaderFilterMode safely sets the value for global configuration 'Advanced.HeaderFilterMode' field
func SetAdvancedHeaderFilterMode(v string) { global.SetAdvancedHeaderFilterMode(v) }

// GetAdvancedRateLimitRequests safely fetches the Configuration value for state's 'Advanced.RateLimit.Requests' field
func (st *ConfigState) GetAdvancedRateLimitRequests() (v int) {
	st.mutex.RLock()
	v = st.config.Advanced.RateLimit.Requests
	st.mutex.RUnlock()
	return
}

// SetAdvancedRateLimitRequests safely sets the Configuration value for state's 'Advanced.RateLimit.Requests' field
func (st *ConfigState) SetAdvancedRateLimitRequests(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Advanced.RateLimit.Requests = v
	st.reloadToViper()
}

// GetAdvancedRateLimitRequests safely fetches the value for global configuration 'Advanced.RateLimit.Requests' field
func GetAdvancedRateLimitRequests() int { return global.GetAdvancedRateLimitRequests() }

// SetAdvancedRateLimitRequests safely sets the value for global configuration 'Advanced.RateLimit.Requests' field
func SetAdvancedRateLimitRequests(v int) { global.SetAdvancedRateLimitRequests(v) }

// GetAdvancedRateLimitExceptions safely fetches the Configuration value for state's 'Advanced.RateLimit.Exceptions' field
func (st *ConfigState) GetAdvancedRateLimitExceptions() (v IPPrefixes) {
	st.mutex.RLock()
	v = st.config.Advanced.RateLimit.Exceptions
	st.mutex.RUnlock()
	return
}

// SetAdvancedRateLimitExceptions safely sets the Configuration value for state's 'Advanced.RateLimit.Exceptions' field
func (st *ConfigState) SetAdvancedRateLimitExceptions(v IPPrefixes) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Advanced.RateLimit.Exceptions = v
	st.reloadToViper()
}

// GetAdvancedRateLimitExceptions safely fetches the value for global configuration 'Advanced.RateLimit.Exceptions' field
func GetAdvancedRateLimitExceptions() IPPrefixes { return global.GetAdvancedRateLimitExceptions() }

// SetAdvancedRateLimitExceptions safely sets the value for global configuration 'Advanced.RateLimit.Exceptions' field
func SetAdvancedRateLimitExceptions(v IPPrefixes) { global.SetAdvancedRateLimitExceptions(v) }

// GetAdvancedThrottlingMultiplier safely fetches the Configuration value for state's 'Advanced.Throttling.Multiplier' field
func (st *ConfigState) GetAdvancedThrottlingMultiplier() (v int) {
	st.mutex.RLock()
	v = st.config.Advanced.Throttling.Multiplier
	st.mutex.RUnlock()
	return
}

// SetAdvancedThrottlingMultiplier safely sets the Configuration value for state's 'Advanced.Throttling.Multiplier' field
func (st *ConfigState) SetAdvancedThrottlingMultiplier(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Advanced.Throttling.Multiplier = v
	st.reloadToViper()
}

// GetAdvancedThrottlingMultiplier safely fetches the value for global configuration 'Advanced.Throttling.Multiplier' field
func GetAdvancedThrottlingMultiplier() int { return global.GetAdvancedThrottlingMultiplier() }

// SetAdvancedThrottlingMultiplier safely sets the value for global configuration 'Advanced.Throttling.Multiplier' field
func SetAdvancedThrottlingMultiplier(v int) { global.SetAdvancedThrottlingMultiplier(v) }

// GetAdvancedThrottlingRetryAfter safely fetches the Configuration value for state's 'Advanced.Throttling.RetryAfter' field
func (st *ConfigState) GetAdvancedThrottlingRetryAfter() (v time.Duration) {
	st.mutex.RLock()
	v = st.config.Advanced.Throttling.RetryAfter
	st.mutex.RUnlock()
	return
}

// SetAdvancedThrottlingRetryAfter safely sets the Configuration value for state's 'Advanced.Throttling.RetryAfter' field
func (st *ConfigState) SetAdvancedThrottlingRetryAfter(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Advanced.Throttling.RetryAfter = v
	st.reloadToViper()
}

// GetAdvancedThrottlingRetryAfter safely fetches the value for global configuration 'Advanced.Throttling.RetryAfter' field
func GetAdvancedThrottlingRetryAfter() time.Duration { return global.GetAdvancedThrottlingRetryAfter() }

// SetAdvancedThrottlingRetryAfter safely sets the value for global configuration 'Advanced.Throttling.RetryAfter' field
func SetAdvancedThrottlingRetryAfter(v time.Duration) { global.SetAdvancedThrottlingRetryAfter(v) }

// GetAdvancedScraperDeterrenceEnabled safely fetches the Configuration value for state's 'Advanced.ScraperDeterrence.Enabled' field
func (st *ConfigState) GetAdvancedScraperDeterrenceEnabled() (v bool) {
	st.mutex.RLock()
	v = st.config.Advanced.ScraperDeterrence.Enabled
	st.mutex.RUnlock()
	return
}

// SetAdvancedScraperDeterrenceEnabled safely sets the Configuration value for state's 'Advanced.ScraperDeterrence.Enabled' field
func (st *ConfigState) SetAdvancedScraperDeterrenceEnabled(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Advanced.ScraperDeterrence.Enabled = v
	st.reloadToViper()
}

// GetAdvancedScraperDeterrenceEnabled safely fetches the value for global configuration 'Advanced.ScraperDeterrence.Enabled' field
func GetAdvancedScraperDeterrenceEnabled() bool { return global.GetAdvancedScraperDeterrenceEnabled() }

// SetAdvancedScraperDeterrenceEnabled safely sets the value for global configuration 'Advanced.ScraperDeterrence.Enabled' field
func SetAdvancedScraperDeterrenceEnabled(v bool) { global.SetAdvancedScraperDeterrenceEnabled(v) }

// GetAdvancedScraperDeterrenceDifficulty safely fetches the Configuration value for state's 'Advanced.ScraperDeterrence.Difficulty' field
func (st *ConfigState) GetAdvancedScraperDeterrenceDifficulty() (v uint32) {
	st.mutex.RLock()
	v = st.config.Advanced.ScraperDeterrence.Difficulty
	st.mutex.RUnlock()
	return
}

// SetAdvancedScraperDeterrenceDifficulty safely sets the Configuration value for state's 'Advanced.ScraperDeterrence.Difficulty' field
func (st *ConfigState) SetAdvancedScraperDeterrenceDifficulty(v uint32) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Advanced.ScraperDeterrence.Difficulty = v
	st.reloadToViper()
}

// GetAdvancedScraperDeterrenceDifficulty safely fetches the value for global configuration 'Advanced.ScraperDeterrence.Difficulty' field
func GetAdvancedScraperDeterrenceDifficulty() uint32 {
	return global.GetAdvancedScraperDeterrenceDifficulty()
}

// SetAdvancedScraperDeterrenceDifficulty safely sets the value for global configuration 'Advanced.ScraperDeterrence.Difficulty' field
func SetAdvancedScraperDeterrenceDifficulty(v uint32) {
	global.SetAdvancedScraperDeterrenceDifficulty(v)
}

// GetHTTPClientAllowIPs safely fetches the Configuration value for state's 'HTTPClient.AllowIPs' field
func (st *ConfigState) GetHTTPClientAllowIPs() (v []string) {
	st.mutex.RLock()
	v = st.config.HTTPClient.AllowIPs
	st.mutex.RUnlock()
	return
}

// SetHTTPClientAllowIPs safely sets the Configuration value for state's 'HTTPClient.AllowIPs' field
func (st *ConfigState) SetHTTPClientAllowIPs(v []string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.HTTPClient.AllowIPs = v
	st.reloadToViper()
}

// GetHTTPClientAllowIPs safely fetches the value for global configuration 'HTTPClient.AllowIPs' field
func GetHTTPClientAllowIPs() []string { return global.GetHTTPClientAllowIPs() }

// SetHTTPClientAllowIPs safely sets the value for global configuration 'HTTPClient.AllowIPs' field
func SetHTTPClientAllowIPs(v []string) { global.SetHTTPClientAllowIPs(v) }

// GetHTTPClientBlockIPs safely fetches the Configuration value for state's 'HTTPClient.BlockIPs' field
func (st *ConfigState) GetHTTPClientBlockIPs() (v []string) {
	st.mutex.RLock()
	v = st.config.HTTPClient.BlockIPs
	st.mutex.RUnlock()
	return
}

// SetHTTPClientBlockIPs safely sets the Configuration value for state's 'HTTPClient.BlockIPs' field
func (st *ConfigState) SetHTTPClientBlockIPs(v []string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.HTTPClient.BlockIPs = v
	st.reloadToViper()
}

// GetHTTPClientBlockIPs safely fetches the value for global configuration 'HTTPClient.BlockIPs' field
func GetHTTPClientBlockIPs() []string { return global.GetHTTPClientBlockIPs() }

// SetHTTPClientBlockIPs safely sets the value for global configuration 'HTTPClient.BlockIPs' field
func SetHTTPClientBlockIPs(v []string) { global.SetHTTPClientBlockIPs(v) }

// GetHTTPClientTimeout safely fetches the Configuration value for state's 'HTTPClient.Timeout' field
func (st *ConfigState) GetHTTPClientTimeout() (v time.Duration) {
	st.mutex.RLock()
	v = st.config.HTTPClient.Timeout
	st.mutex.RUnlock()
	return
}

// SetHTTPClientTimeout safely sets the Configuration value for state's 'HTTPClient.Timeout' field
func (st *ConfigState) SetHTTPClientTimeout(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.HTTPClient.Timeout = v
	st.reloadToViper()
}

// GetHTTPClientTimeout safely fetches the value for global configuration 'HTTPClient.Timeout' field
func GetHTTPClientTimeout() time.Duration { return global.GetHTTPClientTimeout() }

// SetHTTPClientTimeout safely sets the value for global configuration 'HTTPClient.Timeout' field
func SetHTTPClientTimeout(v time.Duration) { global.SetHTTPClientTimeout(v) }

// GetHTTPClientTLSInsecureSkipVerify safely fetches the Configuration value for state's 'HTTPClient.TLSInsecureSkipVerify' field
func (st *ConfigState) GetHTTPClientTLSInsecureSkipVerify() (v bool) {
	st.mutex.RLock()
	v = st.config.HTTPClient.TLSInsecureSkipVerify
	st.mutex.RUnlock()
	return
}

// SetHTTPClientTLSInsecureSkipVerify safely sets the Configuration value for state's 'HTTPClient.TLSInsecureSkipVerify' field
func (st *ConfigState) SetHTTPClientTLSInsecureSkipVerify(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.HTTPClient.TLSInsecureSkipVerify = v
	st.reloadToViper()
}

// GetHTTPClientTLSInsecureSkipVerify safely fetches the value for global configuration 'HTTPClient.TLSInsecureSkipVerify' field
func GetHTTPClientTLSInsecureSkipVerify() bool { return global.GetHTTPClientTLSInsecureSkipVerify() }

// SetHTTPClientTLSInsecureSkipVerify safely sets the value for global configuration 'HTTPClient.TLSInsecureSkipVerify' field
func SetHTTPClientTLSInsecureSkipVerify(v bool) { global.SetHTTPClientTLSInsecureSkipVerify(v) }

// GetHTTPClientInsecureOutgoing safely fetches the Configuration value for state's 'HTTPClient.InsecureOutgoing' field
func (st *ConfigState) GetHTTPClientInsecureOutgoing() (v bool) {
	st.mutex.RLock()
	v = st.config.HTTPClient.InsecureOutgoing
	st.mutex.RUnlock()
	return
}

// SetHTTPClientInsecureOutgoing safely sets the Configuration value for state's 'HTTPClient.InsecureOutgoing' field
func (st *ConfigState) SetHTTPClientInsecureOutgoing(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.HTTPClient.InsecureOutgoing = v
	st.reloadToViper()
}

// GetHTTPClientInsecureOutgoing safely fetches the value for global configuration 'HTTPClient.InsecureOutgoing' field
func GetHTTPClientInsecureOutgoing() bool { return global.GetHTTPClientInsecureOutgoing() }

// SetHTTPClientInsecureOutgoing safely sets the value for global configuration 'HTTPClient.InsecureOutgoing' field
func SetHTTPClientInsecureOutgoing(v bool) { global.SetHTTPClientInsecureOutgoing(v) }

// GetMediaDescriptionMinChars safely fetches the Configuration value for state's 'Media.DescriptionMinChars' field
func (st *ConfigState) GetMediaDescriptionMinChars() (v int) {
	st.mutex.RLock()
	v = st.config.Media.DescriptionMinChars
	st.mutex.RUnlock()
	return
}

// SetMediaDescriptionMinChars safely sets the Configuration value for state's 'Media.DescriptionMinChars' field
func (st *ConfigState) SetMediaDescriptionMinChars(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.DescriptionMinChars = v
	st.reloadToViper()
}

// GetMediaDescriptionMinChars safely fetches the value for global configuration 'Media.DescriptionMinChars' field
func GetMediaDescriptionMinChars() int { return global.GetMediaDescriptionMinChars() }

// SetMediaDescriptionMinChars safely sets the value for global configuration 'Media.DescriptionMinChars' field
func SetMediaDescriptionMinChars(v int) { global.SetMediaDescriptionMinChars(v) }

// GetMediaDescriptionMaxChars safely fetches the Configuration value for state's 'Media.DescriptionMaxChars' field
func (st *ConfigState) GetMediaDescriptionMaxChars() (v int) {
	st.mutex.RLock()
	v = st.config.Media.DescriptionMaxChars
	st.mutex.RUnlock()
	return
}

// SetMediaDescriptionMaxChars safely sets the Configuration value for state's 'Media.DescriptionMaxChars' field
func (st *ConfigState) SetMediaDescriptionMaxChars(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.DescriptionMaxChars = v
	st.reloadToViper()
}

// GetMediaDescriptionMaxChars safely fetches the value for global configuration 'Media.DescriptionMaxChars' field
func GetMediaDescriptionMaxChars() int { return global.GetMediaDescriptionMaxChars() }

// SetMediaDescriptionMaxChars safely sets the value for global configuration 'Media.DescriptionMaxChars' field
func SetMediaDescriptionMaxChars(v int) { global.SetMediaDescriptionMaxChars(v) }

// GetMediaRemoteCacheDays safely fetches the Configuration value for state's 'Media.RemoteCacheDays' field
func (st *ConfigState) GetMediaRemoteCacheDays() (v int) {
	st.mutex.RLock()
	v = st.config.Media.RemoteCacheDays
	st.mutex.RUnlock()
	return
}

// SetMediaRemoteCacheDays safely sets the Configuration value for state's 'Media.RemoteCacheDays' field
func (st *ConfigState) SetMediaRemoteCacheDays(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.RemoteCacheDays = v
	st.reloadToViper()
}

// GetMediaRemoteCacheDays safely fetches the value for global configuration 'Media.RemoteCacheDays' field
func GetMediaRemoteCacheDays() int { return global.GetMediaRemoteCacheDays() }

// SetMediaRemoteCacheDays safely sets the value for global configuration 'Media.RemoteCacheDays' field
func SetMediaRemoteCacheDays(v int) { global.SetMediaRemoteCacheDays(v) }

// GetMediaEmojiLocalMaxSize safely fetches the Configuration value for state's 'Media.EmojiLocalMaxSize' field
func (st *ConfigState) GetMediaEmojiLocalMaxSize() (v bytesize.Size) {
	st.mutex.RLock()
	v = st.config.Media.EmojiLocalMaxSize
	st.mutex.RUnlock()
	return
}

// SetMediaEmojiLocalMaxSize safely sets the Configuration value for state's 'Media.EmojiLocalMaxSize' field
func (st *ConfigState) SetMediaEmojiLocalMaxSize(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.EmojiLocalMaxSize = v
	st.reloadToViper()
}

// GetMediaEmojiLocalMaxSize safely fetches the value for global configuration 'Media.EmojiLocalMaxSize' field
func GetMediaEmojiLocalMaxSize() bytesize.Size { return global.GetMediaEmojiLocalMaxSize() }

// SetMediaEmojiLocalMaxSize safely sets the value for global configuration 'Media.EmojiLocalMaxSize' field
func SetMediaEmojiLocalMaxSize(v bytesize.Size) { global.SetMediaEmojiLocalMaxSize(v) }

// GetMediaEmojiRemoteMaxSize safely fetches the Configuration value for state's 'Media.EmojiRemoteMaxSize' field
func (st *ConfigState) GetMediaEmojiRemoteMaxSize() (v bytesize.Size) {
	st.mutex.RLock()
	v = st.config.Media.EmojiRemoteMaxSize
	st.mutex.RUnlock()
	return
}

// SetMediaEmojiRemoteMaxSize safely sets the Configuration value for state's 'Media.EmojiRemoteMaxSize' field
func (st *ConfigState) SetMediaEmojiRemoteMaxSize(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.EmojiRemoteMaxSize = v
	st.reloadToViper()
}

// GetMediaEmojiRemoteMaxSize safely fetches the value for global configuration 'Media.EmojiRemoteMaxSize' field
func GetMediaEmojiRemoteMaxSize() bytesize.Size { return global.GetMediaEmojiRemoteMaxSize() }

// SetMediaEmojiRemoteMaxSize safely sets the value for global configuration 'Media.EmojiRemoteMaxSize' field
func SetMediaEmojiRemoteMaxSize(v bytesize.Size) { global.SetMediaEmojiRemoteMaxSize(v) }

// GetMediaImageSizeHint safely fetches the Configuration value for state's 'Media.ImageSizeHint' field
func (st *ConfigState) GetMediaImageSizeHint() (v bytesize.Size) {
	st.mutex.RLock()
	v = st.config.Media.ImageSizeHint
	st.mutex.RUnlock()
	return
}

// SetMediaImageSizeHint safely sets the Configuration value for state's 'Media.ImageSizeHint' field
func (st *ConfigState) SetMediaImageSizeHint(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.ImageSizeHint = v
	st.reloadToViper()
}

// GetMediaImageSizeHint safely fetches the value for global configuration 'Media.ImageSizeHint' field
func GetMediaImageSizeHint() bytesize.Size { return global.GetMediaImageSizeHint() }

// SetMediaImageSizeHint safely sets the value for global configuration 'Media.ImageSizeHint' field
func SetMediaImageSizeHint(v bytesize.Size) { global.SetMediaImageSizeHint(v) }

// GetMediaVideoSizeHint safely fetches the Configuration value for state's 'Media.VideoSizeHint' field
func (st *ConfigState) GetMediaVideoSizeHint() (v bytesize.Size) {
	st.mutex.RLock()
	v = st.config.Media.VideoSizeHint
	st.mutex.RUnlock()
	return
}

// SetMediaVideoSizeHint safely sets the Configuration value for state's 'Media.VideoSizeHint' field
func (st *ConfigState) SetMediaVideoSizeHint(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.VideoSizeHint = v
	st.reloadToViper()
}

// GetMediaVideoSizeHint safely fetches the value for global configuration 'Media.VideoSizeHint' field
func GetMediaVideoSizeHint() bytesize.Size { return global.GetMediaVideoSizeHint() }

// SetMediaVideoSizeHint safely sets the value for global configuration 'Media.VideoSizeHint' field
func SetMediaVideoSizeHint(v bytesize.Size) { global.SetMediaVideoSizeHint(v) }

// GetMediaLocalMaxSize safely fetches the Configuration value for state's 'Media.LocalMaxSize' field
func (st *ConfigState) GetMediaLocalMaxSize() (v bytesize.Size) {
	st.mutex.RLock()
	v = st.config.Media.LocalMaxSize
	st.mutex.RUnlock()
	return
}

// SetMediaLocalMaxSize safely sets the Configuration value for state's 'Media.LocalMaxSize' field
func (st *ConfigState) SetMediaLocalMaxSize(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.LocalMaxSize = v
	st.reloadToViper()
}

// GetMediaLocalMaxSize safely fetches the value for global configuration 'Media.LocalMaxSize' field
func GetMediaLocalMaxSize() bytesize.Size { return global.GetMediaLocalMaxSize() }

// SetMediaLocalMaxSize safely sets the value for global configuration 'Media.LocalMaxSize' field
func SetMediaLocalMaxSize(v bytesize.Size) { global.SetMediaLocalMaxSize(v) }

// GetMediaRemoteMaxSize safely fetches the Configuration value for state's 'Media.RemoteMaxSize' field
func (st *ConfigState) GetMediaRemoteMaxSize() (v bytesize.Size) {
	st.mutex.RLock()
	v = st.config.Media.RemoteMaxSize
	st.mutex.RUnlock()
	return
}

// SetMediaRemoteMaxSize safely sets the Configuration value for state's 'Media.RemoteMaxSize' field
func (st *ConfigState) SetMediaRemoteMaxSize(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.RemoteMaxSize = v
	st.reloadToViper()
}

// GetMediaRemoteMaxSize safely fetches the value for global configuration 'Media.RemoteMaxSize' field
func GetMediaRemoteMaxSize() bytesize.Size { return global.GetMediaRemoteMaxSize() }

// SetMediaRemoteMaxSize safely sets the value for global configuration 'Media.RemoteMaxSize' field
func SetMediaRemoteMaxSize(v bytesize.Size) { global.SetMediaRemoteMaxSize(v) }

// GetMediaCleanupFrom safely fetches the Configuration value for state's 'Media.CleanupFrom' field
func (st *ConfigState) GetMediaCleanupFrom() (v string) {
	st.mutex.RLock()
	v = st.config.Media.CleanupFrom
	st.mutex.RUnlock()
	return
}

// SetMediaCleanupFrom safely sets the Configuration value for state's 'Media.CleanupFrom' field
func (st *ConfigState) SetMediaCleanupFrom(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.CleanupFrom = v
	st.reloadToViper()
}

// GetMediaCleanupFrom safely fetches the value for global configuration 'Media.CleanupFrom' field
func GetMediaCleanupFrom() string { return global.GetMediaCleanupFrom() }

// SetMediaCleanupFrom safely sets the value for global configuration 'Media.CleanupFrom' field
func SetMediaCleanupFrom(v string) { global.SetMediaCleanupFrom(v) }

// GetMediaCleanupEvery safely fetches the Configuration value for state's 'Media.CleanupEvery' field
func (st *ConfigState) GetMediaCleanupEvery() (v time.Duration) {
	st.mutex.RLock()
	v = st.config.Media.CleanupEvery
	st.mutex.RUnlock()
	return
}

// SetMediaCleanupEvery safely sets the Configuration value for state's 'Media.CleanupEvery' field
func (st *ConfigState) SetMediaCleanupEvery(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.CleanupEvery = v
	st.reloadToViper()
}

// GetMediaCleanupEvery safely fetches the value for global configuration 'Media.CleanupEvery' field
func GetMediaCleanupEvery() time.Duration { return global.GetMediaCleanupEvery() }

// SetMediaCleanupEvery safely sets the value for global configuration 'Media.CleanupEvery' field
func SetMediaCleanupEvery(v time.Duration) { global.SetMediaCleanupEvery(v) }

// GetMediaFfmpegPoolSize safely fetches the Configuration value for state's 'Media.FfmpegPoolSize' field
func (st *ConfigState) GetMediaFfmpegPoolSize() (v int) {
	st.mutex.RLock()
	v = st.config.Media.FfmpegPoolSize
	st.mutex.RUnlock()
	return
}

// SetMediaFfmpegPoolSize safely sets the Configuration value for state's 'Media.FfmpegPoolSize' field
func (st *ConfigState) SetMediaFfmpegPoolSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.FfmpegPoolSize = v
	st.reloadToViper()
}

// GetMediaFfmpegPoolSize safely fetches the value for global configuration 'Media.FfmpegPoolSize' field
func GetMediaFfmpegPoolSize() int { return global.GetMediaFfmpegPoolSize() }

// SetMediaFfmpegPoolSize safely sets the value for global configuration 'Media.FfmpegPoolSize' field
func SetMediaFfmpegPoolSize(v int) { global.SetMediaFfmpegPoolSize(v) }

// GetMediaThumbMaxPixels safely fetches the Configuration value for state's 'Media.ThumbMaxPixels' field
func (st *ConfigState) GetMediaThumbMaxPixels() (v int) {
	st.mutex.RLock()
	v = st.config.Media.ThumbMaxPixels
	st.mutex.RUnlock()
	return
}

// SetMediaThumbMaxPixels safely sets the Configuration value for state's 'Media.ThumbMaxPixels' field
func (st *ConfigState) SetMediaThumbMaxPixels(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Media.ThumbMaxPixels = v
	st.reloadToViper()
}

// GetMediaThumbMaxPixels safely fetches the value for global configuration 'Media.ThumbMaxPixels' field
func GetMediaThumbMaxPixels() int { return global.GetMediaThumbMaxPixels() }

// SetMediaThumbMaxPixels safely sets the value for global configuration 'Media.ThumbMaxPixels' field
func SetMediaThumbMaxPixels(v int) { global.SetMediaThumbMaxPixels(v) }

// GetCacheMemoryTarget safely fetches the Configuration value for state's 'Cache.MemoryTarget' field
func (st *ConfigState) GetCacheMemoryTarget() (v bytesize.Size) {
	st.mutex.RLock()
	v = st.config.Cache.MemoryTarget
	st.mutex.RUnlock()
	return
}

// SetCacheMemoryTarget safely sets the Configuration value for state's 'Cache.MemoryTarget' field
func (st *ConfigState) SetCacheMemoryTarget(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.MemoryTarget = v
	st.reloadToViper()
}

// GetCacheMemoryTarget safely fetches the value for global configuration 'Cache.MemoryTarget' field
func GetCacheMemoryTarget() bytesize.Size { return global.GetCacheMemoryTarget() }

// SetCacheMemoryTarget safely sets the value for global configuration 'Cache.MemoryTarget' field
func SetCacheMemoryTarget(v bytesize.Size) { global.SetCacheMemoryTarget(v) }

// GetCacheAccountMemRatio safely fetches the Configuration value for state's 'Cache.AccountMemRatio' field
func (st *ConfigState) GetCacheAccountMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.AccountMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheAccountMemRatio safely sets the Configuration value for state's 'Cache.AccountMemRatio' field
func (st *ConfigState) SetCacheAccountMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.AccountMemRatio = v
	st.reloadToViper()
}

// GetCacheAccountMemRatio safely fetches the value for global configuration 'Cache.AccountMemRatio' field
func GetCacheAccountMemRatio() float64 { return global.GetCacheAccountMemRatio() }

// SetCacheAccountMemRatio safely sets the value for global configuration 'Cache.AccountMemRatio' field
func SetCacheAccountMemRatio(v float64) { global.SetCacheAccountMemRatio(v) }

// GetCacheAccountNoteMemRatio safely fetches the Configuration value for state's 'Cache.AccountNoteMemRatio' field
func (st *ConfigState) GetCacheAccountNoteMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.AccountNoteMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheAccountNoteMemRatio safely sets the Configuration value for state's 'Cache.AccountNoteMemRatio' field
func (st *ConfigState) SetCacheAccountNoteMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.AccountNoteMemRatio = v
	st.reloadToViper()
}

// GetCacheAccountNoteMemRatio safely fetches the value for global configuration 'Cache.AccountNoteMemRatio' field
func GetCacheAccountNoteMemRatio() float64 { return global.GetCacheAccountNoteMemRatio() }

// SetCacheAccountNoteMemRatio safely sets the value for global configuration 'Cache.AccountNoteMemRatio' field
func SetCacheAccountNoteMemRatio(v float64) { global.SetCacheAccountNoteMemRatio(v) }

// GetCacheAccountSettingsMemRatio safely fetches the Configuration value for state's 'Cache.AccountSettingsMemRatio' field
func (st *ConfigState) GetCacheAccountSettingsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.AccountSettingsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheAccountSettingsMemRatio safely sets the Configuration value for state's 'Cache.AccountSettingsMemRatio' field
func (st *ConfigState) SetCacheAccountSettingsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.AccountSettingsMemRatio = v
	st.reloadToViper()
}

// GetCacheAccountSettingsMemRatio safely fetches the value for global configuration 'Cache.AccountSettingsMemRatio' field
func GetCacheAccountSettingsMemRatio() float64 { return global.GetCacheAccountSettingsMemRatio() }

// SetCacheAccountSettingsMemRatio safely sets the value for global configuration 'Cache.AccountSettingsMemRatio' field
func SetCacheAccountSettingsMemRatio(v float64) { global.SetCacheAccountSettingsMemRatio(v) }

// GetCacheAccountStatsMemRatio safely fetches the Configuration value for state's 'Cache.AccountStatsMemRatio' field
func (st *ConfigState) GetCacheAccountStatsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.AccountStatsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheAccountStatsMemRatio safely sets the Configuration value for state's 'Cache.AccountStatsMemRatio' field
func (st *ConfigState) SetCacheAccountStatsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.AccountStatsMemRatio = v
	st.reloadToViper()
}

// GetCacheAccountStatsMemRatio safely fetches the value for global configuration 'Cache.AccountStatsMemRatio' field
func GetCacheAccountStatsMemRatio() float64 { return global.GetCacheAccountStatsMemRatio() }

// SetCacheAccountStatsMemRatio safely sets the value for global configuration 'Cache.AccountStatsMemRatio' field
func SetCacheAccountStatsMemRatio(v float64) { global.SetCacheAccountStatsMemRatio(v) }

// GetCacheApplicationMemRatio safely fetches the Configuration value for state's 'Cache.ApplicationMemRatio' field
func (st *ConfigState) GetCacheApplicationMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.ApplicationMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheApplicationMemRatio safely sets the Configuration value for state's 'Cache.ApplicationMemRatio' field
func (st *ConfigState) SetCacheApplicationMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.ApplicationMemRatio = v
	st.reloadToViper()
}

// GetCacheApplicationMemRatio safely fetches the value for global configuration 'Cache.ApplicationMemRatio' field
func GetCacheApplicationMemRatio() float64 { return global.GetCacheApplicationMemRatio() }

// SetCacheApplicationMemRatio safely sets the value for global configuration 'Cache.ApplicationMemRatio' field
func SetCacheApplicationMemRatio(v float64) { global.SetCacheApplicationMemRatio(v) }

// GetCacheBlockMemRatio safely fetches the Configuration value for state's 'Cache.BlockMemRatio' field
func (st *ConfigState) GetCacheBlockMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.BlockMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheBlockMemRatio safely sets the Configuration value for state's 'Cache.BlockMemRatio' field
func (st *ConfigState) SetCacheBlockMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.BlockMemRatio = v
	st.reloadToViper()
}

// GetCacheBlockMemRatio safely fetches the value for global configuration 'Cache.BlockMemRatio' field
func GetCacheBlockMemRatio() float64 { return global.GetCacheBlockMemRatio() }

// SetCacheBlockMemRatio safely sets the value for global configuration 'Cache.BlockMemRatio' field
func SetCacheBlockMemRatio(v float64) { global.SetCacheBlockMemRatio(v) }

// GetCacheBlockIDsMemRatio safely fetches the Configuration value for state's 'Cache.BlockIDsMemRatio' field
func (st *ConfigState) GetCacheBlockIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.BlockIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheBlockIDsMemRatio safely sets the Configuration value for state's 'Cache.BlockIDsMemRatio' field
func (st *ConfigState) SetCacheBlockIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.BlockIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheBlockIDsMemRatio safely fetches the value for global configuration 'Cache.BlockIDsMemRatio' field
func GetCacheBlockIDsMemRatio() float64 { return global.GetCacheBlockIDsMemRatio() }

// SetCacheBlockIDsMemRatio safely sets the value for global configuration 'Cache.BlockIDsMemRatio' field
func SetCacheBlockIDsMemRatio(v float64) { global.SetCacheBlockIDsMemRatio(v) }

// GetCacheBoostOfIDsMemRatio safely fetches the Configuration value for state's 'Cache.BoostOfIDsMemRatio' field
func (st *ConfigState) GetCacheBoostOfIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.BoostOfIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheBoostOfIDsMemRatio safely sets the Configuration value for state's 'Cache.BoostOfIDsMemRatio' field
func (st *ConfigState) SetCacheBoostOfIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.BoostOfIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheBoostOfIDsMemRatio safely fetches the value for global configuration 'Cache.BoostOfIDsMemRatio' field
func GetCacheBoostOfIDsMemRatio() float64 { return global.GetCacheBoostOfIDsMemRatio() }

// SetCacheBoostOfIDsMemRatio safely sets the value for global configuration 'Cache.BoostOfIDsMemRatio' field
func SetCacheBoostOfIDsMemRatio(v float64) { global.SetCacheBoostOfIDsMemRatio(v) }

// GetCacheClientMemRatio safely fetches the Configuration value for state's 'Cache.ClientMemRatio' field
func (st *ConfigState) GetCacheClientMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.ClientMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheClientMemRatio safely sets the Configuration value for state's 'Cache.ClientMemRatio' field
func (st *ConfigState) SetCacheClientMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.ClientMemRatio = v
	st.reloadToViper()
}

// GetCacheClientMemRatio safely fetches the value for global configuration 'Cache.ClientMemRatio' field
func GetCacheClientMemRatio() float64 { return global.GetCacheClientMemRatio() }

// SetCacheClientMemRatio safely sets the value for global configuration 'Cache.ClientMemRatio' field
func SetCacheClientMemRatio(v float64) { global.SetCacheClientMemRatio(v) }

// GetCacheConversationMemRatio safely fetches the Configuration value for state's 'Cache.ConversationMemRatio' field
func (st *ConfigState) GetCacheConversationMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.ConversationMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheConversationMemRatio safely sets the Configuration value for state's 'Cache.ConversationMemRatio' field
func (st *ConfigState) SetCacheConversationMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.ConversationMemRatio = v
	st.reloadToViper()
}

// GetCacheConversationMemRatio safely fetches the value for global configuration 'Cache.ConversationMemRatio' field
func GetCacheConversationMemRatio() float64 { return global.GetCacheConversationMemRatio() }

// SetCacheConversationMemRatio safely sets the value for global configuration 'Cache.ConversationMemRatio' field
func SetCacheConversationMemRatio(v float64) { global.SetCacheConversationMemRatio(v) }

// GetCacheConversationLastStatusIDsMemRatio safely fetches the Configuration value for state's 'Cache.ConversationLastStatusIDsMemRatio' field
func (st *ConfigState) GetCacheConversationLastStatusIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.ConversationLastStatusIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheConversationLastStatusIDsMemRatio safely sets the Configuration value for state's 'Cache.ConversationLastStatusIDsMemRatio' field
func (st *ConfigState) SetCacheConversationLastStatusIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.ConversationLastStatusIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheConversationLastStatusIDsMemRatio safely fetches the value for global configuration 'Cache.ConversationLastStatusIDsMemRatio' field
func GetCacheConversationLastStatusIDsMemRatio() float64 {
	return global.GetCacheConversationLastStatusIDsMemRatio()
}

// SetCacheConversationLastStatusIDsMemRatio safely sets the value for global configuration 'Cache.ConversationLastStatusIDsMemRatio' field
func SetCacheConversationLastStatusIDsMemRatio(v float64) {
	global.SetCacheConversationLastStatusIDsMemRatio(v)
}

// GetCacheDomainPermissionDraftMemRation safely fetches the Configuration value for state's 'Cache.DomainPermissionDraftMemRation' field
func (st *ConfigState) GetCacheDomainPermissionDraftMemRation() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.DomainPermissionDraftMemRation
	st.mutex.RUnlock()
	return
}

// SetCacheDomainPermissionDraftMemRation safely sets the Configuration value for state's 'Cache.DomainPermissionDraftMemRation' field
func (st *ConfigState) SetCacheDomainPermissionDraftMemRation(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.DomainPermissionDraftMemRation = v
	st.reloadToViper()
}

// GetCacheDomainPermissionDraftMemRation safely fetches the value for global configuration 'Cache.DomainPermissionDraftMemRation' field
func GetCacheDomainPermissionDraftMemRation() float64 {
	return global.GetCacheDomainPermissionDraftMemRation()
}

// SetCacheDomainPermissionDraftMemRation safely sets the value for global configuration 'Cache.DomainPermissionDraftMemRation' field
func SetCacheDomainPermissionDraftMemRation(v float64) {
	global.SetCacheDomainPermissionDraftMemRation(v)
}

// GetCacheDomainPermissionSubscriptionMemRation safely fetches the Configuration value for state's 'Cache.DomainPermissionSubscriptionMemRation' field
func (st *ConfigState) GetCacheDomainPermissionSubscriptionMemRation() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.DomainPermissionSubscriptionMemRation
	st.mutex.RUnlock()
	return
}

// SetCacheDomainPermissionSubscriptionMemRation safely sets the Configuration value for state's 'Cache.DomainPermissionSubscriptionMemRation' field
func (st *ConfigState) SetCacheDomainPermissionSubscriptionMemRation(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.DomainPermissionSubscriptionMemRation = v
	st.reloadToViper()
}

// GetCacheDomainPermissionSubscriptionMemRation safely fetches the value for global configuration 'Cache.DomainPermissionSubscriptionMemRation' field
func GetCacheDomainPermissionSubscriptionMemRation() float64 {
	return global.GetCacheDomainPermissionSubscriptionMemRation()
}

// SetCacheDomainPermissionSubscriptionMemRation safely sets the value for global configuration 'Cache.DomainPermissionSubscriptionMemRation' field
func SetCacheDomainPermissionSubscriptionMemRation(v float64) {
	global.SetCacheDomainPermissionSubscriptionMemRation(v)
}

// GetCacheEmojiMemRatio safely fetches the Configuration value for state's 'Cache.EmojiMemRatio' field
func (st *ConfigState) GetCacheEmojiMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.EmojiMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheEmojiMemRatio safely sets the Configuration value for state's 'Cache.EmojiMemRatio' field
func (st *ConfigState) SetCacheEmojiMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.EmojiMemRatio = v
	st.reloadToViper()
}

// GetCacheEmojiMemRatio safely fetches the value for global configuration 'Cache.EmojiMemRatio' field
func GetCacheEmojiMemRatio() float64 { return global.GetCacheEmojiMemRatio() }

// SetCacheEmojiMemRatio safely sets the value for global configuration 'Cache.EmojiMemRatio' field
func SetCacheEmojiMemRatio(v float64) { global.SetCacheEmojiMemRatio(v) }

// GetCacheEmojiCategoryMemRatio safely fetches the Configuration value for state's 'Cache.EmojiCategoryMemRatio' field
func (st *ConfigState) GetCacheEmojiCategoryMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.EmojiCategoryMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheEmojiCategoryMemRatio safely sets the Configuration value for state's 'Cache.EmojiCategoryMemRatio' field
func (st *ConfigState) SetCacheEmojiCategoryMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.EmojiCategoryMemRatio = v
	st.reloadToViper()
}

// GetCacheEmojiCategoryMemRatio safely fetches the value for global configuration 'Cache.EmojiCategoryMemRatio' field
func GetCacheEmojiCategoryMemRatio() float64 { return global.GetCacheEmojiCategoryMemRatio() }

// SetCacheEmojiCategoryMemRatio safely sets the value for global configuration 'Cache.EmojiCategoryMemRatio' field
func SetCacheEmojiCategoryMemRatio(v float64) { global.SetCacheEmojiCategoryMemRatio(v) }

// GetCacheFilterMemRatio safely fetches the Configuration value for state's 'Cache.FilterMemRatio' field
func (st *ConfigState) GetCacheFilterMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.FilterMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheFilterMemRatio safely sets the Configuration value for state's 'Cache.FilterMemRatio' field
func (st *ConfigState) SetCacheFilterMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.FilterMemRatio = v
	st.reloadToViper()
}

// GetCacheFilterMemRatio safely fetches the value for global configuration 'Cache.FilterMemRatio' field
func GetCacheFilterMemRatio() float64 { return global.GetCacheFilterMemRatio() }

// SetCacheFilterMemRatio safely sets the value for global configuration 'Cache.FilterMemRatio' field
func SetCacheFilterMemRatio(v float64) { global.SetCacheFilterMemRatio(v) }

// GetCacheFilterIDsMemRatio safely fetches the Configuration value for state's 'Cache.FilterIDsMemRatio' field
func (st *ConfigState) GetCacheFilterIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.FilterIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheFilterIDsMemRatio safely sets the Configuration value for state's 'Cache.FilterIDsMemRatio' field
func (st *ConfigState) SetCacheFilterIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.FilterIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheFilterIDsMemRatio safely fetches the value for global configuration 'Cache.FilterIDsMemRatio' field
func GetCacheFilterIDsMemRatio() float64 { return global.GetCacheFilterIDsMemRatio() }

// SetCacheFilterIDsMemRatio safely sets the value for global configuration 'Cache.FilterIDsMemRatio' field
func SetCacheFilterIDsMemRatio(v float64) { global.SetCacheFilterIDsMemRatio(v) }

// GetCacheFilterKeywordMemRatio safely fetches the Configuration value for state's 'Cache.FilterKeywordMemRatio' field
func (st *ConfigState) GetCacheFilterKeywordMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.FilterKeywordMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheFilterKeywordMemRatio safely sets the Configuration value for state's 'Cache.FilterKeywordMemRatio' field
func (st *ConfigState) SetCacheFilterKeywordMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.FilterKeywordMemRatio = v
	st.reloadToViper()
}

// GetCacheFilterKeywordMemRatio safely fetches the value for global configuration 'Cache.FilterKeywordMemRatio' field
func GetCacheFilterKeywordMemRatio() float64 { return global.GetCacheFilterKeywordMemRatio() }

// SetCacheFilterKeywordMemRatio safely sets the value for global configuration 'Cache.FilterKeywordMemRatio' field
func SetCacheFilterKeywordMemRatio(v float64) { global.SetCacheFilterKeywordMemRatio(v) }

// GetCacheFilterStatusMemRatio safely fetches the Configuration value for state's 'Cache.FilterStatusMemRatio' field
func (st *ConfigState) GetCacheFilterStatusMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.FilterStatusMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheFilterStatusMemRatio safely sets the Configuration value for state's 'Cache.FilterStatusMemRatio' field
func (st *ConfigState) SetCacheFilterStatusMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.FilterStatusMemRatio = v
	st.reloadToViper()
}

// GetCacheFilterStatusMemRatio safely fetches the value for global configuration 'Cache.FilterStatusMemRatio' field
func GetCacheFilterStatusMemRatio() float64 { return global.GetCacheFilterStatusMemRatio() }

// SetCacheFilterStatusMemRatio safely sets the value for global configuration 'Cache.FilterStatusMemRatio' field
func SetCacheFilterStatusMemRatio(v float64) { global.SetCacheFilterStatusMemRatio(v) }

// GetCacheFollowMemRatio safely fetches the Configuration value for state's 'Cache.FollowMemRatio' field
func (st *ConfigState) GetCacheFollowMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.FollowMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheFollowMemRatio safely sets the Configuration value for state's 'Cache.FollowMemRatio' field
func (st *ConfigState) SetCacheFollowMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.FollowMemRatio = v
	st.reloadToViper()
}

// GetCacheFollowMemRatio safely fetches the value for global configuration 'Cache.FollowMemRatio' field
func GetCacheFollowMemRatio() float64 { return global.GetCacheFollowMemRatio() }

// SetCacheFollowMemRatio safely sets the value for global configuration 'Cache.FollowMemRatio' field
func SetCacheFollowMemRatio(v float64) { global.SetCacheFollowMemRatio(v) }

// GetCacheFollowIDsMemRatio safely fetches the Configuration value for state's 'Cache.FollowIDsMemRatio' field
func (st *ConfigState) GetCacheFollowIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.FollowIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheFollowIDsMemRatio safely sets the Configuration value for state's 'Cache.FollowIDsMemRatio' field
func (st *ConfigState) SetCacheFollowIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.FollowIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheFollowIDsMemRatio safely fetches the value for global configuration 'Cache.FollowIDsMemRatio' field
func GetCacheFollowIDsMemRatio() float64 { return global.GetCacheFollowIDsMemRatio() }

// SetCacheFollowIDsMemRatio safely sets the value for global configuration 'Cache.FollowIDsMemRatio' field
func SetCacheFollowIDsMemRatio(v float64) { global.SetCacheFollowIDsMemRatio(v) }

// GetCacheFollowRequestMemRatio safely fetches the Configuration value for state's 'Cache.FollowRequestMemRatio' field
func (st *ConfigState) GetCacheFollowRequestMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.FollowRequestMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheFollowRequestMemRatio safely sets the Configuration value for state's 'Cache.FollowRequestMemRatio' field
func (st *ConfigState) SetCacheFollowRequestMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.FollowRequestMemRatio = v
	st.reloadToViper()
}

// GetCacheFollowRequestMemRatio safely fetches the value for global configuration 'Cache.FollowRequestMemRatio' field
func GetCacheFollowRequestMemRatio() float64 { return global.GetCacheFollowRequestMemRatio() }

// SetCacheFollowRequestMemRatio safely sets the value for global configuration 'Cache.FollowRequestMemRatio' field
func SetCacheFollowRequestMemRatio(v float64) { global.SetCacheFollowRequestMemRatio(v) }

// GetCacheFollowRequestIDsMemRatio safely fetches the Configuration value for state's 'Cache.FollowRequestIDsMemRatio' field
func (st *ConfigState) GetCacheFollowRequestIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.FollowRequestIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheFollowRequestIDsMemRatio safely sets the Configuration value for state's 'Cache.FollowRequestIDsMemRatio' field
func (st *ConfigState) SetCacheFollowRequestIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.FollowRequestIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheFollowRequestIDsMemRatio safely fetches the value for global configuration 'Cache.FollowRequestIDsMemRatio' field
func GetCacheFollowRequestIDsMemRatio() float64 { return global.GetCacheFollowRequestIDsMemRatio() }

// SetCacheFollowRequestIDsMemRatio safely sets the value for global configuration 'Cache.FollowRequestIDsMemRatio' field
func SetCacheFollowRequestIDsMemRatio(v float64) { global.SetCacheFollowRequestIDsMemRatio(v) }

// GetCacheFollowingTagIDsMemRatio safely fetches the Configuration value for state's 'Cache.FollowingTagIDsMemRatio' field
func (st *ConfigState) GetCacheFollowingTagIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.FollowingTagIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheFollowingTagIDsMemRatio safely sets the Configuration value for state's 'Cache.FollowingTagIDsMemRatio' field
func (st *ConfigState) SetCacheFollowingTagIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.FollowingTagIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheFollowingTagIDsMemRatio safely fetches the value for global configuration 'Cache.FollowingTagIDsMemRatio' field
func GetCacheFollowingTagIDsMemRatio() float64 { return global.GetCacheFollowingTagIDsMemRatio() }

// SetCacheFollowingTagIDsMemRatio safely sets the value for global configuration 'Cache.FollowingTagIDsMemRatio' field
func SetCacheFollowingTagIDsMemRatio(v float64) { global.SetCacheFollowingTagIDsMemRatio(v) }

// GetCacheInReplyToIDsMemRatio safely fetches the Configuration value for state's 'Cache.InReplyToIDsMemRatio' field
func (st *ConfigState) GetCacheInReplyToIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.InReplyToIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheInReplyToIDsMemRatio safely sets the Configuration value for state's 'Cache.InReplyToIDsMemRatio' field
func (st *ConfigState) SetCacheInReplyToIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.InReplyToIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheInReplyToIDsMemRatio safely fetches the value for global configuration 'Cache.InReplyToIDsMemRatio' field
func GetCacheInReplyToIDsMemRatio() float64 { return global.GetCacheInReplyToIDsMemRatio() }

// SetCacheInReplyToIDsMemRatio safely sets the value for global configuration 'Cache.InReplyToIDsMemRatio' field
func SetCacheInReplyToIDsMemRatio(v float64) { global.SetCacheInReplyToIDsMemRatio(v) }

// GetCacheInstanceMemRatio safely fetches the Configuration value for state's 'Cache.InstanceMemRatio' field
func (st *ConfigState) GetCacheInstanceMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.InstanceMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheInstanceMemRatio safely sets the Configuration value for state's 'Cache.InstanceMemRatio' field
func (st *ConfigState) SetCacheInstanceMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.InstanceMemRatio = v
	st.reloadToViper()
}

// GetCacheInstanceMemRatio safely fetches the value for global configuration 'Cache.InstanceMemRatio' field
func GetCacheInstanceMemRatio() float64 { return global.GetCacheInstanceMemRatio() }

// SetCacheInstanceMemRatio safely sets the value for global configuration 'Cache.InstanceMemRatio' field
func SetCacheInstanceMemRatio(v float64) { global.SetCacheInstanceMemRatio(v) }

// GetCacheInteractionRequestMemRatio safely fetches the Configuration value for state's 'Cache.InteractionRequestMemRatio' field
func (st *ConfigState) GetCacheInteractionRequestMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.InteractionRequestMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheInteractionRequestMemRatio safely sets the Configuration value for state's 'Cache.InteractionRequestMemRatio' field
func (st *ConfigState) SetCacheInteractionRequestMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.InteractionRequestMemRatio = v
	st.reloadToViper()
}

// GetCacheInteractionRequestMemRatio safely fetches the value for global configuration 'Cache.InteractionRequestMemRatio' field
func GetCacheInteractionRequestMemRatio() float64 { return global.GetCacheInteractionRequestMemRatio() }

// SetCacheInteractionRequestMemRatio safely sets the value for global configuration 'Cache.InteractionRequestMemRatio' field
func SetCacheInteractionRequestMemRatio(v float64) { global.SetCacheInteractionRequestMemRatio(v) }

// GetCacheListMemRatio safely fetches the Configuration value for state's 'Cache.ListMemRatio' field
func (st *ConfigState) GetCacheListMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.ListMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheListMemRatio safely sets the Configuration value for state's 'Cache.ListMemRatio' field
func (st *ConfigState) SetCacheListMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.ListMemRatio = v
	st.reloadToViper()
}

// GetCacheListMemRatio safely fetches the value for global configuration 'Cache.ListMemRatio' field
func GetCacheListMemRatio() float64 { return global.GetCacheListMemRatio() }

// SetCacheListMemRatio safely sets the value for global configuration 'Cache.ListMemRatio' field
func SetCacheListMemRatio(v float64) { global.SetCacheListMemRatio(v) }

// GetCacheListIDsMemRatio safely fetches the Configuration value for state's 'Cache.ListIDsMemRatio' field
func (st *ConfigState) GetCacheListIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.ListIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheListIDsMemRatio safely sets the Configuration value for state's 'Cache.ListIDsMemRatio' field
func (st *ConfigState) SetCacheListIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.ListIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheListIDsMemRatio safely fetches the value for global configuration 'Cache.ListIDsMemRatio' field
func GetCacheListIDsMemRatio() float64 { return global.GetCacheListIDsMemRatio() }

// SetCacheListIDsMemRatio safely sets the value for global configuration 'Cache.ListIDsMemRatio' field
func SetCacheListIDsMemRatio(v float64) { global.SetCacheListIDsMemRatio(v) }

// GetCacheListedIDsMemRatio safely fetches the Configuration value for state's 'Cache.ListedIDsMemRatio' field
func (st *ConfigState) GetCacheListedIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.ListedIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheListedIDsMemRatio safely sets the Configuration value for state's 'Cache.ListedIDsMemRatio' field
func (st *ConfigState) SetCacheListedIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.ListedIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheListedIDsMemRatio safely fetches the value for global configuration 'Cache.ListedIDsMemRatio' field
func GetCacheListedIDsMemRatio() float64 { return global.GetCacheListedIDsMemRatio() }

// SetCacheListedIDsMemRatio safely sets the value for global configuration 'Cache.ListedIDsMemRatio' field
func SetCacheListedIDsMemRatio(v float64) { global.SetCacheListedIDsMemRatio(v) }

// GetCacheMarkerMemRatio safely fetches the Configuration value for state's 'Cache.MarkerMemRatio' field
func (st *ConfigState) GetCacheMarkerMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.MarkerMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheMarkerMemRatio safely sets the Configuration value for state's 'Cache.MarkerMemRatio' field
func (st *ConfigState) SetCacheMarkerMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.MarkerMemRatio = v
	st.reloadToViper()
}

// GetCacheMarkerMemRatio safely fetches the value for global configuration 'Cache.MarkerMemRatio' field
func GetCacheMarkerMemRatio() float64 { return global.GetCacheMarkerMemRatio() }

// SetCacheMarkerMemRatio safely sets the value for global configuration 'Cache.MarkerMemRatio' field
func SetCacheMarkerMemRatio(v float64) { global.SetCacheMarkerMemRatio(v) }

// GetCacheMediaMemRatio safely fetches the Configuration value for state's 'Cache.MediaMemRatio' field
func (st *ConfigState) GetCacheMediaMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.MediaMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheMediaMemRatio safely sets the Configuration value for state's 'Cache.MediaMemRatio' field
func (st *ConfigState) SetCacheMediaMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.MediaMemRatio = v
	st.reloadToViper()
}

// GetCacheMediaMemRatio safely fetches the value for global configuration 'Cache.MediaMemRatio' field
func GetCacheMediaMemRatio() float64 { return global.GetCacheMediaMemRatio() }

// SetCacheMediaMemRatio safely sets the value for global configuration 'Cache.MediaMemRatio' field
func SetCacheMediaMemRatio(v float64) { global.SetCacheMediaMemRatio(v) }

// GetCacheMentionMemRatio safely fetches the Configuration value for state's 'Cache.MentionMemRatio' field
func (st *ConfigState) GetCacheMentionMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.MentionMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheMentionMemRatio safely sets the Configuration value for state's 'Cache.MentionMemRatio' field
func (st *ConfigState) SetCacheMentionMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.MentionMemRatio = v
	st.reloadToViper()
}

// GetCacheMentionMemRatio safely fetches the value for global configuration 'Cache.MentionMemRatio' field
func GetCacheMentionMemRatio() float64 { return global.GetCacheMentionMemRatio() }

// SetCacheMentionMemRatio safely sets the value for global configuration 'Cache.MentionMemRatio' field
func SetCacheMentionMemRatio(v float64) { global.SetCacheMentionMemRatio(v) }

// GetCacheMoveMemRatio safely fetches the Configuration value for state's 'Cache.MoveMemRatio' field
func (st *ConfigState) GetCacheMoveMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.MoveMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheMoveMemRatio safely sets the Configuration value for state's 'Cache.MoveMemRatio' field
func (st *ConfigState) SetCacheMoveMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.MoveMemRatio = v
	st.reloadToViper()
}

// GetCacheMoveMemRatio safely fetches the value for global configuration 'Cache.MoveMemRatio' field
func GetCacheMoveMemRatio() float64 { return global.GetCacheMoveMemRatio() }

// SetCacheMoveMemRatio safely sets the value for global configuration 'Cache.MoveMemRatio' field
func SetCacheMoveMemRatio(v float64) { global.SetCacheMoveMemRatio(v) }

// GetCacheNotificationMemRatio safely fetches the Configuration value for state's 'Cache.NotificationMemRatio' field
func (st *ConfigState) GetCacheNotificationMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.NotificationMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheNotificationMemRatio safely sets the Configuration value for state's 'Cache.NotificationMemRatio' field
func (st *ConfigState) SetCacheNotificationMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.NotificationMemRatio = v
	st.reloadToViper()
}

// GetCacheNotificationMemRatio safely fetches the value for global configuration 'Cache.NotificationMemRatio' field
func GetCacheNotificationMemRatio() float64 { return global.GetCacheNotificationMemRatio() }

// SetCacheNotificationMemRatio safely sets the value for global configuration 'Cache.NotificationMemRatio' field
func SetCacheNotificationMemRatio(v float64) { global.SetCacheNotificationMemRatio(v) }

// GetCachePollMemRatio safely fetches the Configuration value for state's 'Cache.PollMemRatio' field
func (st *ConfigState) GetCachePollMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.PollMemRatio
	st.mutex.RUnlock()
	return
}

// SetCachePollMemRatio safely sets the Configuration value for state's 'Cache.PollMemRatio' field
func (st *ConfigState) SetCachePollMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.PollMemRatio = v
	st.reloadToViper()
}

// GetCachePollMemRatio safely fetches the value for global configuration 'Cache.PollMemRatio' field
func GetCachePollMemRatio() float64 { return global.GetCachePollMemRatio() }

// SetCachePollMemRatio safely sets the value for global configuration 'Cache.PollMemRatio' field
func SetCachePollMemRatio(v float64) { global.SetCachePollMemRatio(v) }

// GetCachePollVoteMemRatio safely fetches the Configuration value for state's 'Cache.PollVoteMemRatio' field
func (st *ConfigState) GetCachePollVoteMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.PollVoteMemRatio
	st.mutex.RUnlock()
	return
}

// SetCachePollVoteMemRatio safely sets the Configuration value for state's 'Cache.PollVoteMemRatio' field
func (st *ConfigState) SetCachePollVoteMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.PollVoteMemRatio = v
	st.reloadToViper()
}

// GetCachePollVoteMemRatio safely fetches the value for global configuration 'Cache.PollVoteMemRatio' field
func GetCachePollVoteMemRatio() float64 { return global.GetCachePollVoteMemRatio() }

// SetCachePollVoteMemRatio safely sets the value for global configuration 'Cache.PollVoteMemRatio' field
func SetCachePollVoteMemRatio(v float64) { global.SetCachePollVoteMemRatio(v) }

// GetCachePollVoteIDsMemRatio safely fetches the Configuration value for state's 'Cache.PollVoteIDsMemRatio' field
func (st *ConfigState) GetCachePollVoteIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.PollVoteIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCachePollVoteIDsMemRatio safely sets the Configuration value for state's 'Cache.PollVoteIDsMemRatio' field
func (st *ConfigState) SetCachePollVoteIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.PollVoteIDsMemRatio = v
	st.reloadToViper()
}

// GetCachePollVoteIDsMemRatio safely fetches the value for global configuration 'Cache.PollVoteIDsMemRatio' field
func GetCachePollVoteIDsMemRatio() float64 { return global.GetCachePollVoteIDsMemRatio() }

// SetCachePollVoteIDsMemRatio safely sets the value for global configuration 'Cache.PollVoteIDsMemRatio' field
func SetCachePollVoteIDsMemRatio(v float64) { global.SetCachePollVoteIDsMemRatio(v) }

// GetCacheReportMemRatio safely fetches the Configuration value for state's 'Cache.ReportMemRatio' field
func (st *ConfigState) GetCacheReportMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.ReportMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheReportMemRatio safely sets the Configuration value for state's 'Cache.ReportMemRatio' field
func (st *ConfigState) SetCacheReportMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.ReportMemRatio = v
	st.reloadToViper()
}

// GetCacheReportMemRatio safely fetches the value for global configuration 'Cache.ReportMemRatio' field
func GetCacheReportMemRatio() float64 { return global.GetCacheReportMemRatio() }

// SetCacheReportMemRatio safely sets the value for global configuration 'Cache.ReportMemRatio' field
func SetCacheReportMemRatio(v float64) { global.SetCacheReportMemRatio(v) }

// GetCacheScheduledStatusMemRatio safely fetches the Configuration value for state's 'Cache.ScheduledStatusMemRatio' field
func (st *ConfigState) GetCacheScheduledStatusMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.ScheduledStatusMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheScheduledStatusMemRatio safely sets the Configuration value for state's 'Cache.ScheduledStatusMemRatio' field
func (st *ConfigState) SetCacheScheduledStatusMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.ScheduledStatusMemRatio = v
	st.reloadToViper()
}

// GetCacheScheduledStatusMemRatio safely fetches the value for global configuration 'Cache.ScheduledStatusMemRatio' field
func GetCacheScheduledStatusMemRatio() float64 { return global.GetCacheScheduledStatusMemRatio() }

// SetCacheScheduledStatusMemRatio safely sets the value for global configuration 'Cache.ScheduledStatusMemRatio' field
func SetCacheScheduledStatusMemRatio(v float64) { global.SetCacheScheduledStatusMemRatio(v) }

// GetCacheSinBinStatusMemRatio safely fetches the Configuration value for state's 'Cache.SinBinStatusMemRatio' field
func (st *ConfigState) GetCacheSinBinStatusMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.SinBinStatusMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheSinBinStatusMemRatio safely sets the Configuration value for state's 'Cache.SinBinStatusMemRatio' field
func (st *ConfigState) SetCacheSinBinStatusMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.SinBinStatusMemRatio = v
	st.reloadToViper()
}

// GetCacheSinBinStatusMemRatio safely fetches the value for global configuration 'Cache.SinBinStatusMemRatio' field
func GetCacheSinBinStatusMemRatio() float64 { return global.GetCacheSinBinStatusMemRatio() }

// SetCacheSinBinStatusMemRatio safely sets the value for global configuration 'Cache.SinBinStatusMemRatio' field
func SetCacheSinBinStatusMemRatio(v float64) { global.SetCacheSinBinStatusMemRatio(v) }

// GetCacheStatusMemRatio safely fetches the Configuration value for state's 'Cache.StatusMemRatio' field
func (st *ConfigState) GetCacheStatusMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.StatusMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheStatusMemRatio safely sets the Configuration value for state's 'Cache.StatusMemRatio' field
func (st *ConfigState) SetCacheStatusMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.StatusMemRatio = v
	st.reloadToViper()
}

// GetCacheStatusMemRatio safely fetches the value for global configuration 'Cache.StatusMemRatio' field
func GetCacheStatusMemRatio() float64 { return global.GetCacheStatusMemRatio() }

// SetCacheStatusMemRatio safely sets the value for global configuration 'Cache.StatusMemRatio' field
func SetCacheStatusMemRatio(v float64) { global.SetCacheStatusMemRatio(v) }

// GetCacheStatusBookmarkMemRatio safely fetches the Configuration value for state's 'Cache.StatusBookmarkMemRatio' field
func (st *ConfigState) GetCacheStatusBookmarkMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.StatusBookmarkMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheStatusBookmarkMemRatio safely sets the Configuration value for state's 'Cache.StatusBookmarkMemRatio' field
func (st *ConfigState) SetCacheStatusBookmarkMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.StatusBookmarkMemRatio = v
	st.reloadToViper()
}

// GetCacheStatusBookmarkMemRatio safely fetches the value for global configuration 'Cache.StatusBookmarkMemRatio' field
func GetCacheStatusBookmarkMemRatio() float64 { return global.GetCacheStatusBookmarkMemRatio() }

// SetCacheStatusBookmarkMemRatio safely sets the value for global configuration 'Cache.StatusBookmarkMemRatio' field
func SetCacheStatusBookmarkMemRatio(v float64) { global.SetCacheStatusBookmarkMemRatio(v) }

// GetCacheStatusBookmarkIDsMemRatio safely fetches the Configuration value for state's 'Cache.StatusBookmarkIDsMemRatio' field
func (st *ConfigState) GetCacheStatusBookmarkIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.StatusBookmarkIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheStatusBookmarkIDsMemRatio safely sets the Configuration value for state's 'Cache.StatusBookmarkIDsMemRatio' field
func (st *ConfigState) SetCacheStatusBookmarkIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.StatusBookmarkIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheStatusBookmarkIDsMemRatio safely fetches the value for global configuration 'Cache.StatusBookmarkIDsMemRatio' field
func GetCacheStatusBookmarkIDsMemRatio() float64 { return global.GetCacheStatusBookmarkIDsMemRatio() }

// SetCacheStatusBookmarkIDsMemRatio safely sets the value for global configuration 'Cache.StatusBookmarkIDsMemRatio' field
func SetCacheStatusBookmarkIDsMemRatio(v float64) { global.SetCacheStatusBookmarkIDsMemRatio(v) }

// GetCacheStatusEditMemRatio safely fetches the Configuration value for state's 'Cache.StatusEditMemRatio' field
func (st *ConfigState) GetCacheStatusEditMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.StatusEditMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheStatusEditMemRatio safely sets the Configuration value for state's 'Cache.StatusEditMemRatio' field
func (st *ConfigState) SetCacheStatusEditMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.StatusEditMemRatio = v
	st.reloadToViper()
}

// GetCacheStatusEditMemRatio safely fetches the value for global configuration 'Cache.StatusEditMemRatio' field
func GetCacheStatusEditMemRatio() float64 { return global.GetCacheStatusEditMemRatio() }

// SetCacheStatusEditMemRatio safely sets the value for global configuration 'Cache.StatusEditMemRatio' field
func SetCacheStatusEditMemRatio(v float64) { global.SetCacheStatusEditMemRatio(v) }

// GetCacheStatusFaveMemRatio safely fetches the Configuration value for state's 'Cache.StatusFaveMemRatio' field
func (st *ConfigState) GetCacheStatusFaveMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.StatusFaveMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheStatusFaveMemRatio safely sets the Configuration value for state's 'Cache.StatusFaveMemRatio' field
func (st *ConfigState) SetCacheStatusFaveMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.StatusFaveMemRatio = v
	st.reloadToViper()
}

// GetCacheStatusFaveMemRatio safely fetches the value for global configuration 'Cache.StatusFaveMemRatio' field
func GetCacheStatusFaveMemRatio() float64 { return global.GetCacheStatusFaveMemRatio() }

// SetCacheStatusFaveMemRatio safely sets the value for global configuration 'Cache.StatusFaveMemRatio' field
func SetCacheStatusFaveMemRatio(v float64) { global.SetCacheStatusFaveMemRatio(v) }

// GetCacheStatusFaveIDsMemRatio safely fetches the Configuration value for state's 'Cache.StatusFaveIDsMemRatio' field
func (st *ConfigState) GetCacheStatusFaveIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.StatusFaveIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheStatusFaveIDsMemRatio safely sets the Configuration value for state's 'Cache.StatusFaveIDsMemRatio' field
func (st *ConfigState) SetCacheStatusFaveIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.StatusFaveIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheStatusFaveIDsMemRatio safely fetches the value for global configuration 'Cache.StatusFaveIDsMemRatio' field
func GetCacheStatusFaveIDsMemRatio() float64 { return global.GetCacheStatusFaveIDsMemRatio() }

// SetCacheStatusFaveIDsMemRatio safely sets the value for global configuration 'Cache.StatusFaveIDsMemRatio' field
func SetCacheStatusFaveIDsMemRatio(v float64) { global.SetCacheStatusFaveIDsMemRatio(v) }

// GetCacheTagMemRatio safely fetches the Configuration value for state's 'Cache.TagMemRatio' field
func (st *ConfigState) GetCacheTagMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.TagMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheTagMemRatio safely sets the Configuration value for state's 'Cache.TagMemRatio' field
func (st *ConfigState) SetCacheTagMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.TagMemRatio = v
	st.reloadToViper()
}

// GetCacheTagMemRatio safely fetches the value for global configuration 'Cache.TagMemRatio' field
func GetCacheTagMemRatio() float64 { return global.GetCacheTagMemRatio() }

// SetCacheTagMemRatio safely sets the value for global configuration 'Cache.TagMemRatio' field
func SetCacheTagMemRatio(v float64) { global.SetCacheTagMemRatio(v) }

// GetCacheThreadMuteMemRatio safely fetches the Configuration value for state's 'Cache.ThreadMuteMemRatio' field
func (st *ConfigState) GetCacheThreadMuteMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.ThreadMuteMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheThreadMuteMemRatio safely sets the Configuration value for state's 'Cache.ThreadMuteMemRatio' field
func (st *ConfigState) SetCacheThreadMuteMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.ThreadMuteMemRatio = v
	st.reloadToViper()
}

// GetCacheThreadMuteMemRatio safely fetches the value for global configuration 'Cache.ThreadMuteMemRatio' field
func GetCacheThreadMuteMemRatio() float64 { return global.GetCacheThreadMuteMemRatio() }

// SetCacheThreadMuteMemRatio safely sets the value for global configuration 'Cache.ThreadMuteMemRatio' field
func SetCacheThreadMuteMemRatio(v float64) { global.SetCacheThreadMuteMemRatio(v) }

// GetCacheTokenMemRatio safely fetches the Configuration value for state's 'Cache.TokenMemRatio' field
func (st *ConfigState) GetCacheTokenMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.TokenMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheTokenMemRatio safely sets the Configuration value for state's 'Cache.TokenMemRatio' field
func (st *ConfigState) SetCacheTokenMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.TokenMemRatio = v
	st.reloadToViper()
}

// GetCacheTokenMemRatio safely fetches the value for global configuration 'Cache.TokenMemRatio' field
func GetCacheTokenMemRatio() float64 { return global.GetCacheTokenMemRatio() }

// SetCacheTokenMemRatio safely sets the value for global configuration 'Cache.TokenMemRatio' field
func SetCacheTokenMemRatio(v float64) { global.SetCacheTokenMemRatio(v) }

// GetCacheTombstoneMemRatio safely fetches the Configuration value for state's 'Cache.TombstoneMemRatio' field
func (st *ConfigState) GetCacheTombstoneMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.TombstoneMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheTombstoneMemRatio safely sets the Configuration value for state's 'Cache.TombstoneMemRatio' field
func (st *ConfigState) SetCacheTombstoneMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.TombstoneMemRatio = v
	st.reloadToViper()
}

// GetCacheTombstoneMemRatio safely fetches the value for global configuration 'Cache.TombstoneMemRatio' field
func GetCacheTombstoneMemRatio() float64 { return global.GetCacheTombstoneMemRatio() }

// SetCacheTombstoneMemRatio safely sets the value for global configuration 'Cache.TombstoneMemRatio' field
func SetCacheTombstoneMemRatio(v float64) { global.SetCacheTombstoneMemRatio(v) }

// GetCacheUserMemRatio safely fetches the Configuration value for state's 'Cache.UserMemRatio' field
func (st *ConfigState) GetCacheUserMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.UserMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheUserMemRatio safely sets the Configuration value for state's 'Cache.UserMemRatio' field
func (st *ConfigState) SetCacheUserMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.UserMemRatio = v
	st.reloadToViper()
}

// GetCacheUserMemRatio safely fetches the value for global configuration 'Cache.UserMemRatio' field
func GetCacheUserMemRatio() float64 { return global.GetCacheUserMemRatio() }

// SetCacheUserMemRatio safely sets the value for global configuration 'Cache.UserMemRatio' field
func SetCacheUserMemRatio(v float64) { global.SetCacheUserMemRatio(v) }

// GetCacheUserMuteMemRatio safely fetches the Configuration value for state's 'Cache.UserMuteMemRatio' field
func (st *ConfigState) GetCacheUserMuteMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.UserMuteMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheUserMuteMemRatio safely sets the Configuration value for state's 'Cache.UserMuteMemRatio' field
func (st *ConfigState) SetCacheUserMuteMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.UserMuteMemRatio = v
	st.reloadToViper()
}

// GetCacheUserMuteMemRatio safely fetches the value for global configuration 'Cache.UserMuteMemRatio' field
func GetCacheUserMuteMemRatio() float64 { return global.GetCacheUserMuteMemRatio() }

// SetCacheUserMuteMemRatio safely sets the value for global configuration 'Cache.UserMuteMemRatio' field
func SetCacheUserMuteMemRatio(v float64) { global.SetCacheUserMuteMemRatio(v) }

// GetCacheUserMuteIDsMemRatio safely fetches the Configuration value for state's 'Cache.UserMuteIDsMemRatio' field
func (st *ConfigState) GetCacheUserMuteIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.UserMuteIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheUserMuteIDsMemRatio safely sets the Configuration value for state's 'Cache.UserMuteIDsMemRatio' field
func (st *ConfigState) SetCacheUserMuteIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.UserMuteIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheUserMuteIDsMemRatio safely fetches the value for global configuration 'Cache.UserMuteIDsMemRatio' field
func GetCacheUserMuteIDsMemRatio() float64 { return global.GetCacheUserMuteIDsMemRatio() }

// SetCacheUserMuteIDsMemRatio safely sets the value for global configuration 'Cache.UserMuteIDsMemRatio' field
func SetCacheUserMuteIDsMemRatio(v float64) { global.SetCacheUserMuteIDsMemRatio(v) }

// GetCacheWebfingerMemRatio safely fetches the Configuration value for state's 'Cache.WebfingerMemRatio' field
func (st *ConfigState) GetCacheWebfingerMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.WebfingerMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheWebfingerMemRatio safely sets the Configuration value for state's 'Cache.WebfingerMemRatio' field
func (st *ConfigState) SetCacheWebfingerMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.WebfingerMemRatio = v
	st.reloadToViper()
}

// GetCacheWebfingerMemRatio safely fetches the value for global configuration 'Cache.WebfingerMemRatio' field
func GetCacheWebfingerMemRatio() float64 { return global.GetCacheWebfingerMemRatio() }

// SetCacheWebfingerMemRatio safely sets the value for global configuration 'Cache.WebfingerMemRatio' field
func SetCacheWebfingerMemRatio(v float64) { global.SetCacheWebfingerMemRatio(v) }

// GetCacheWebPushSubscriptionMemRatio safely fetches the Configuration value for state's 'Cache.WebPushSubscriptionMemRatio' field
func (st *ConfigState) GetCacheWebPushSubscriptionMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.WebPushSubscriptionMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheWebPushSubscriptionMemRatio safely sets the Configuration value for state's 'Cache.WebPushSubscriptionMemRatio' field
func (st *ConfigState) SetCacheWebPushSubscriptionMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.WebPushSubscriptionMemRatio = v
	st.reloadToViper()
}

// GetCacheWebPushSubscriptionMemRatio safely fetches the value for global configuration 'Cache.WebPushSubscriptionMemRatio' field
func GetCacheWebPushSubscriptionMemRatio() float64 {
	return global.GetCacheWebPushSubscriptionMemRatio()
}

// SetCacheWebPushSubscriptionMemRatio safely sets the value for global configuration 'Cache.WebPushSubscriptionMemRatio' field
func SetCacheWebPushSubscriptionMemRatio(v float64) { global.SetCacheWebPushSubscriptionMemRatio(v) }

// GetCacheWebPushSubscriptionIDsMemRatio safely fetches the Configuration value for state's 'Cache.WebPushSubscriptionIDsMemRatio' field
func (st *ConfigState) GetCacheWebPushSubscriptionIDsMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.WebPushSubscriptionIDsMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheWebPushSubscriptionIDsMemRatio safely sets the Configuration value for state's 'Cache.WebPushSubscriptionIDsMemRatio' field
func (st *ConfigState) SetCacheWebPushSubscriptionIDsMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.WebPushSubscriptionIDsMemRatio = v
	st.reloadToViper()
}

// GetCacheWebPushSubscriptionIDsMemRatio safely fetches the value for global configuration 'Cache.WebPushSubscriptionIDsMemRatio' field
func GetCacheWebPushSubscriptionIDsMemRatio() float64 {
	return global.GetCacheWebPushSubscriptionIDsMemRatio()
}

// SetCacheWebPushSubscriptionIDsMemRatio safely sets the value for global configuration 'Cache.WebPushSubscriptionIDsMemRatio' field
func SetCacheWebPushSubscriptionIDsMemRatio(v float64) {
	global.SetCacheWebPushSubscriptionIDsMemRatio(v)
}

// GetCacheMutesMemRatio safely fetches the Configuration value for state's 'Cache.MutesMemRatio' field
func (st *ConfigState) GetCacheMutesMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.MutesMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheMutesMemRatio safely sets the Configuration value for state's 'Cache.MutesMemRatio' field
func (st *ConfigState) SetCacheMutesMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.MutesMemRatio = v
	st.reloadToViper()
}

// GetCacheMutesMemRatio safely fetches the value for global configuration 'Cache.MutesMemRatio' field
func GetCacheMutesMemRatio() float64 { return global.GetCacheMutesMemRatio() }

// SetCacheMutesMemRatio safely sets the value for global configuration 'Cache.MutesMemRatio' field
func SetCacheMutesMemRatio(v float64) { global.SetCacheMutesMemRatio(v) }

// GetCacheStatusFilterMemRatio safely fetches the Configuration value for state's 'Cache.StatusFilterMemRatio' field
func (st *ConfigState) GetCacheStatusFilterMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.StatusFilterMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheStatusFilterMemRatio safely sets the Configuration value for state's 'Cache.StatusFilterMemRatio' field
func (st *ConfigState) SetCacheStatusFilterMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.StatusFilterMemRatio = v
	st.reloadToViper()
}

// GetCacheStatusFilterMemRatio safely fetches the value for global configuration 'Cache.StatusFilterMemRatio' field
func GetCacheStatusFilterMemRatio() float64 { return global.GetCacheStatusFilterMemRatio() }

// SetCacheStatusFilterMemRatio safely sets the value for global configuration 'Cache.StatusFilterMemRatio' field
func SetCacheStatusFilterMemRatio(v float64) { global.SetCacheStatusFilterMemRatio(v) }

// GetCacheVisibilityMemRatio safely fetches the Configuration value for state's 'Cache.VisibilityMemRatio' field
func (st *ConfigState) GetCacheVisibilityMemRatio() (v float64) {
	st.mutex.RLock()
	v = st.config.Cache.VisibilityMemRatio
	st.mutex.RUnlock()
	return
}

// SetCacheVisibilityMemRatio safely sets the Configuration value for state's 'Cache.VisibilityMemRatio' field
func (st *ConfigState) SetCacheVisibilityMemRatio(v float64) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.VisibilityMemRatio = v
	st.reloadToViper()
}

// GetCacheVisibilityMemRatio safely fetches the value for global configuration 'Cache.VisibilityMemRatio' field
func GetCacheVisibilityMemRatio() float64 { return global.GetCacheVisibilityMemRatio() }

// SetCacheVisibilityMemRatio safely sets the value for global configuration 'Cache.VisibilityMemRatio' field
func SetCacheVisibilityMemRatio(v float64) { global.SetCacheVisibilityMemRatio(v) }

// GetAdminAccountUsername safely fetches the Configuration value for state's 'AdminAccountUsername' field
func (st *ConfigState) GetAdminAccountUsername() (v string) {
	st.mutex.RLock()
	v = st.config.AdminAccountUsername
	st.mutex.RUnlock()
	return
}

// SetAdminAccountUsername safely sets the Configuration value for state's 'AdminAccountUsername' field
func (st *ConfigState) SetAdminAccountUsername(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdminAccountUsername = v
	st.reloadToViper()
}

// GetAdminAccountUsername safely fetches the value for global configuration 'AdminAccountUsername' field
func GetAdminAccountUsername() string { return global.GetAdminAccountUsername() }

// SetAdminAccountUsername safely sets the value for global configuration 'AdminAccountUsername' field
func SetAdminAccountUsername(v string) { global.SetAdminAccountUsername(v) }

// GetAdminAccountEmail safely fetches the Configuration value for state's 'AdminAccountEmail' field
func (st *ConfigState) GetAdminAccountEmail() (v string) {
	st.mutex.RLock()
	v = st.config.AdminAccountEmail
	st.mutex.RUnlock()
	return
}

// SetAdminAccountEmail safely sets the Configuration value for state's 'AdminAccountEmail' field
func (st *ConfigState) SetAdminAccountEmail(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdminAccountEmail = v
	st.reloadToViper()
}

// GetAdminAccountEmail safely fetches the value for global configuration 'AdminAccountEmail' field
func GetAdminAccountEmail() string { return global.GetAdminAccountEmail() }

// SetAdminAccountEmail safely sets the value for global configuration 'AdminAccountEmail' field
func SetAdminAccountEmail(v string) { global.SetAdminAccountEmail(v) }

// GetAdminAccountPassword safely fetches the Configuration value for state's 'AdminAccountPassword' field
func (st *ConfigState) GetAdminAccountPassword() (v string) {
	st.mutex.RLock()
	v = st.config.AdminAccountPassword
	st.mutex.RUnlock()
	return
}

// SetAdminAccountPassword safely sets the Configuration value for state's 'AdminAccountPassword' field
func (st *ConfigState) SetAdminAccountPassword(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdminAccountPassword = v
	st.reloadToViper()
}

// GetAdminAccountPassword safely fetches the value for global configuration 'AdminAccountPassword' field
func GetAdminAccountPassword() string { return global.GetAdminAccountPassword() }

// SetAdminAccountPassword safely sets the value for global configuration 'AdminAccountPassword' field
func SetAdminAccountPassword(v string) { global.SetAdminAccountPassword(v) }

// GetAdminTransPath safely fetches the Configuration value for state's 'AdminTransPath' field
func (st *ConfigState) GetAdminTransPath() (v string) {
	st.mutex.RLock()
	v = st.config.AdminTransPath
	st.mutex.RUnlock()
	return
}

// SetAdminTransPath safely sets the Configuration value for state's 'AdminTransPath' field
func (st *ConfigState) SetAdminTransPath(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdminTransPath = v
	st.reloadToViper()
}

// GetAdminTransPath safely fetches the value for global configuration 'AdminTransPath' field
func GetAdminTransPath() string { return global.GetAdminTransPath() }

// SetAdminTransPath safely sets the value for global configuration 'AdminTransPath' field
func SetAdminTransPath(v string) { global.SetAdminTransPath(v) }

// GetAdminMediaPruneDryRun safely fetches the Configuration value for state's 'AdminMediaPruneDryRun' field
func (st *ConfigState) GetAdminMediaPruneDryRun() (v bool) {
	st.mutex.RLock()
	v = st.config.AdminMediaPruneDryRun
	st.mutex.RUnlock()
	return
}

// SetAdminMediaPruneDryRun safely sets the Configuration value for state's 'AdminMediaPruneDryRun' field
func (st *ConfigState) SetAdminMediaPruneDryRun(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdminMediaPruneDryRun = v
	st.reloadToViper()
}

// GetAdminMediaPruneDryRun safely fetches the value for global configuration 'AdminMediaPruneDryRun' field
func GetAdminMediaPruneDryRun() bool { return global.GetAdminMediaPruneDryRun() }

// SetAdminMediaPruneDryRun safely sets the value for global configuration 'AdminMediaPruneDryRun' field
func SetAdminMediaPruneDryRun(v bool) { global.SetAdminMediaPruneDryRun(v) }

// GetAdminMediaListLocalOnly safely fetches the Configuration value for state's 'AdminMediaListLocalOnly' field
func (st *ConfigState) GetAdminMediaListLocalOnly() (v bool) {
	st.mutex.RLock()
	v = st.config.AdminMediaListLocalOnly
	st.mutex.RUnlock()
	return
}

// SetAdminMediaListLocalOnly safely sets the Configuration value for state's 'AdminMediaListLocalOnly' field
func (st *ConfigState) SetAdminMediaListLocalOnly(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdminMediaListLocalOnly = v
	st.reloadToViper()
}

// GetAdminMediaListLocalOnly safely fetches the value for global configuration 'AdminMediaListLocalOnly' field
func GetAdminMediaListLocalOnly() bool { return global.GetAdminMediaListLocalOnly() }

// SetAdminMediaListLocalOnly safely sets the value for global configuration 'AdminMediaListLocalOnly' field
func SetAdminMediaListLocalOnly(v bool) { global.SetAdminMediaListLocalOnly(v) }

// GetAdminMediaListRemoteOnly safely fetches the Configuration value for state's 'AdminMediaListRemoteOnly' field
func (st *ConfigState) GetAdminMediaListRemoteOnly() (v bool) {
	st.mutex.RLock()
	v = st.config.AdminMediaListRemoteOnly
	st.mutex.RUnlock()
	return
}

// SetAdminMediaListRemoteOnly safely sets the Configuration value for state's 'AdminMediaListRemoteOnly' field
func (st *ConfigState) SetAdminMediaListRemoteOnly(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdminMediaListRemoteOnly = v
	st.reloadToViper()
}

// GetAdminMediaListRemoteOnly safely fetches the value for global configuration 'AdminMediaListRemoteOnly' field
func GetAdminMediaListRemoteOnly() bool { return global.GetAdminMediaListRemoteOnly() }

// SetAdminMediaListRemoteOnly safely sets the value for global configuration 'AdminMediaListRemoteOnly' field
func SetAdminMediaListRemoteOnly(v bool) { global.SetAdminMediaListRemoteOnly(v) }

// GetTestrigSkipDBSetup safely fetches the Configuration value for state's 'TestrigSkipDBSetup' field
func (st *ConfigState) GetTestrigSkipDBSetup() (v bool) {
	st.mutex.RLock()
	v = st.config.TestrigSkipDBSetup
	st.mutex.RUnlock()
	return
}

// SetTestrigSkipDBSetup safely sets the Configuration value for state's 'TestrigSkipDBSetup' field
func (st *ConfigState) SetTestrigSkipDBSetup(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.TestrigSkipDBSetup = v
	st.reloadToViper()
}

// GetTestrigSkipDBSetup safely fetches the value for global configuration 'TestrigSkipDBSetup' field
func GetTestrigSkipDBSetup() bool { return global.GetTestrigSkipDBSetup() }

// SetTestrigSkipDBSetup safely sets the value for global configuration 'TestrigSkipDBSetup' field
func SetTestrigSkipDBSetup(v bool) { global.SetTestrigSkipDBSetup(v) }

// GetTestrigSkipDBTeardown safely fetches the Configuration value for state's 'TestrigSkipDBTeardown' field
func (st *ConfigState) GetTestrigSkipDBTeardown() (v bool) {
	st.mutex.RLock()
	v = st.config.TestrigSkipDBTeardown
	st.mutex.RUnlock()
	return
}

// SetTestrigSkipDBTeardown safely sets the Configuration value for state's 'TestrigSkipDBTeardown' field
func (st *ConfigState) SetTestrigSkipDBTeardown(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.TestrigSkipDBTeardown = v
	st.reloadToViper()
}

// GetTestrigSkipDBTeardown safely fetches the value for global configuration 'TestrigSkipDBTeardown' field
func GetTestrigSkipDBTeardown() bool { return global.GetTestrigSkipDBTeardown() }

// SetTestrigSkipDBTeardown safely sets the value for global configuration 'TestrigSkipDBTeardown' field
func SetTestrigSkipDBTeardown(v bool) { global.SetTestrigSkipDBTeardown(v) }

// GetTotalOfMemRatios safely fetches the combined value for all the state's mem ratio fields
func (st *ConfigState) GetTotalOfMemRatios() (total float64) {
	st.mutex.RLock()
	total += st.config.Cache.AccountMemRatio
	total += st.config.Cache.AccountNoteMemRatio
	total += st.config.Cache.AccountSettingsMemRatio
	total += st.config.Cache.AccountStatsMemRatio
	total += st.config.Cache.ApplicationMemRatio
	total += st.config.Cache.BlockMemRatio
	total += st.config.Cache.BlockIDsMemRatio
	total += st.config.Cache.BoostOfIDsMemRatio
	total += st.config.Cache.ClientMemRatio
	total += st.config.Cache.ConversationMemRatio
	total += st.config.Cache.ConversationLastStatusIDsMemRatio
	total += st.config.Cache.DomainPermissionDraftMemRation
	total += st.config.Cache.DomainPermissionSubscriptionMemRation
	total += st.config.Cache.EmojiMemRatio
	total += st.config.Cache.EmojiCategoryMemRatio
	total += st.config.Cache.FilterMemRatio
	total += st.config.Cache.FilterIDsMemRatio
	total += st.config.Cache.FilterKeywordMemRatio
	total += st.config.Cache.FilterStatusMemRatio
	total += st.config.Cache.FollowMemRatio
	total += st.config.Cache.FollowIDsMemRatio
	total += st.config.Cache.FollowRequestMemRatio
	total += st.config.Cache.FollowRequestIDsMemRatio
	total += st.config.Cache.FollowingTagIDsMemRatio
	total += st.config.Cache.InReplyToIDsMemRatio
	total += st.config.Cache.InstanceMemRatio
	total += st.config.Cache.InteractionRequestMemRatio
	total += st.config.Cache.ListMemRatio
	total += st.config.Cache.ListIDsMemRatio
	total += st.config.Cache.ListedIDsMemRatio
	total += st.config.Cache.MarkerMemRatio
	total += st.config.Cache.MediaMemRatio
	total += st.config.Cache.MentionMemRatio
	total += st.config.Cache.MoveMemRatio
	total += st.config.Cache.NotificationMemRatio
	total += st.config.Cache.PollMemRatio
	total += st.config.Cache.PollVoteMemRatio
	total += st.config.Cache.PollVoteIDsMemRatio
	total += st.config.Cache.ReportMemRatio
	total += st.config.Cache.ScheduledStatusMemRatio
	total += st.config.Cache.SinBinStatusMemRatio
	total += st.config.Cache.StatusMemRatio
	total += st.config.Cache.StatusBookmarkMemRatio
	total += st.config.Cache.StatusBookmarkIDsMemRatio
	total += st.config.Cache.StatusEditMemRatio
	total += st.config.Cache.StatusFaveMemRatio
	total += st.config.Cache.StatusFaveIDsMemRatio
	total += st.config.Cache.TagMemRatio
	total += st.config.Cache.ThreadMuteMemRatio
	total += st.config.Cache.TokenMemRatio
	total += st.config.Cache.TombstoneMemRatio
	total += st.config.Cache.UserMemRatio
	total += st.config.Cache.UserMuteMemRatio
	total += st.config.Cache.UserMuteIDsMemRatio
	total += st.config.Cache.WebfingerMemRatio
	total += st.config.Cache.WebPushSubscriptionMemRatio
	total += st.config.Cache.WebPushSubscriptionIDsMemRatio
	total += st.config.Cache.MutesMemRatio
	total += st.config.Cache.StatusFilterMemRatio
	total += st.config.Cache.VisibilityMemRatio
	st.mutex.RUnlock()
	return
}

// GetTotalOfMemRatios safely fetches the combined value for all the global state's mem ratio fields
func GetTotalOfMemRatios() (total float64) { return global.GetTotalOfMemRatios() }

func flattenConfigMap(cfgmap map[string]any) {
	nestedKeys := make(map[string]struct{})
	for _, key := range [][]string{
		{"advanced", "cookies-samesite"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["advanced-cookies-samesite"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"advanced", "sender-multiplier"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["advanced-sender-multiplier"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"advanced", "csp-extra-uris"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["advanced-csp-extra-uris"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"advanced", "header-filter-mode"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["advanced-header-filter-mode"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"advanced-rate-limit", "requests"},
		{"advanced", "rate-limit", "requests"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["advanced-rate-limit-requests"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"advanced-rate-limit", "exceptions"},
		{"advanced", "rate-limit", "exceptions"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["advanced-rate-limit-exceptions"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"advanced-throttling", "multiplier"},
		{"advanced", "throttling", "multiplier"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["advanced-throttling-multiplier"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"advanced-throttling", "retry-after"},
		{"advanced", "throttling", "retry-after"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["advanced-throttling-retry-after"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"advanced-scraper-deterrence", "enabled"},
		{"advanced", "scraper-deterrence", "enabled"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["advanced-scraper-deterrence-enabled"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"advanced-scraper-deterrence", "difficulty"},
		{"advanced", "scraper-deterrence", "difficulty"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["advanced-scraper-deterrence-difficulty"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"http-client", "allow-ips"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["http-client-allow-ips"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"http-client", "block-ips"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["http-client-block-ips"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"http-client", "timeout"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["http-client-timeout"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"http-client", "tls-insecure-skip-verify"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["http-client-tls-insecure-skip-verify"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"http-client", "insecure-outgoing"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["http-client-insecure-outgoing"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "description-min-chars"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-description-min-chars"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "description-max-chars"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-description-max-chars"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "remote-cache-days"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-remote-cache-days"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "emoji-local-max-size"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-emoji-local-max-size"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "emoji-remote-max-size"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-emoji-remote-max-size"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "image-size-hint"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-image-size-hint"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "video-size-hint"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-video-size-hint"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "local-max-size"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-local-max-size"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "remote-max-size"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-remote-max-size"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "cleanup-from"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-cleanup-from"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "cleanup-every"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-cleanup-every"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "ffmpeg-pool-size"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-ffmpeg-pool-size"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"media", "thumb-max-pixels"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["media-thumb-max-pixels"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "memory-target"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-memory-target"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "account-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-account-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "account-note-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-account-note-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "account-settings-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-account-settings-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "account-stats-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-account-stats-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "application-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-application-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "block-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-block-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "block-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-block-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "boost-of-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-boost-of-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "client-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-client-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "conversation-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-conversation-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "conversation-last-status-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-conversation-last-status-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "domain-permission-draft-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-domain-permission-draft-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "domain-permission-subscription-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-domain-permission-subscription-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "emoji-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-emoji-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "emoji-category-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-emoji-category-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "filter-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-filter-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "filter-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-filter-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "filter-keyword-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-filter-keyword-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "filter-status-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-filter-status-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "follow-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-follow-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "follow-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-follow-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "follow-request-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-follow-request-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "follow-request-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-follow-request-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "following-tag-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-following-tag-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "in-reply-to-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-in-reply-to-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "instance-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-instance-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "interaction-request-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-interaction-request-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "list-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-list-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "list-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-list-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "listed-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-listed-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "marker-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-marker-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "media-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-media-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "mention-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-mention-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "move-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-move-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "notification-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-notification-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "poll-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-poll-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "poll-vote-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-poll-vote-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "poll-vote-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-poll-vote-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "report-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-report-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "scheduled-status-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-scheduled-status-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "sin-bin-status-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-sin-bin-status-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "status-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-status-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "status-bookmark-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-status-bookmark-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "status-bookmark-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-status-bookmark-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "status-edit-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-status-edit-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "status-fave-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-status-fave-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "status-fave-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-status-fave-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "tag-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-tag-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "thread-mute-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-thread-mute-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "token-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-token-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "tombstone-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-tombstone-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "user-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-user-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "user-mute-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-user-mute-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "user-mute-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-user-mute-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "webfinger-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-webfinger-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "web-push-subscription-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-web-push-subscription-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "web-push-subscription-ids-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-web-push-subscription-ids-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "mutes-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-mutes-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "status-filter-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-status-filter-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for _, key := range [][]string{
		{"cache", "visibility-mem-ratio"},
	} {
		ival, ok := mapGet(cfgmap, key...)
		if ok {
			cfgmap["cache-visibility-mem-ratio"] = ival
			nestedKeys[key[0]] = struct{}{}
			break
		}
	}

	for key := range nestedKeys {
		delete(cfgmap, key)
	}
}
