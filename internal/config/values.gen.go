// THIS IS A GENERATED FILE, DO NOT EDIT BY HAND
package config

// LogLevelFlag returns the flag name for the 'LogLevel' field
func LogLevelFlag() string {
	return "log-level"
}

// GetLogLevel safely fetches the value for global configuration 'LogLevel' field
func GetLogLevel() (v string) {
	mutex.Lock()
	v = global.LogLevel
	mutex.Unlock()
	return
}

// SetLogLevel safely sets the value for global configuration 'LogLevel' field
func SetLogLevel(v string) {
	mutex.Lock()
	global.LogLevel = v
	mutex.Unlock()
}

// LogDbQueriesFlag returns the flag name for the 'LogDbQueries' field
func LogDbQueriesFlag() string {
	return "log-db-queries"
}

// GetLogDbQueries safely fetches the value for global configuration 'LogDbQueries' field
func GetLogDbQueries() (v bool) {
	mutex.Lock()
	v = global.LogDbQueries
	mutex.Unlock()
	return
}

// SetLogDbQueries safely sets the value for global configuration 'LogDbQueries' field
func SetLogDbQueries(v bool) {
	mutex.Lock()
	global.LogDbQueries = v
	mutex.Unlock()
}

// ApplicationNameFlag returns the flag name for the 'ApplicationName' field
func ApplicationNameFlag() string {
	return "application-name"
}

// GetApplicationName safely fetches the value for global configuration 'ApplicationName' field
func GetApplicationName() (v string) {
	mutex.Lock()
	v = global.ApplicationName
	mutex.Unlock()
	return
}

// SetApplicationName safely sets the value for global configuration 'ApplicationName' field
func SetApplicationName(v string) {
	mutex.Lock()
	global.ApplicationName = v
	mutex.Unlock()
}

// ConfigPathFlag returns the flag name for the 'ConfigPath' field
func ConfigPathFlag() string {
	return "config-path"
}

// GetConfigPath safely fetches the value for global configuration 'ConfigPath' field
func GetConfigPath() (v string) {
	mutex.Lock()
	v = global.ConfigPath
	mutex.Unlock()
	return
}

// SetConfigPath safely sets the value for global configuration 'ConfigPath' field
func SetConfigPath(v string) {
	mutex.Lock()
	global.ConfigPath = v
	mutex.Unlock()
}

// HostFlag returns the flag name for the 'Host' field
func HostFlag() string {
	return "host"
}

// GetHost safely fetches the value for global configuration 'Host' field
func GetHost() (v string) {
	mutex.Lock()
	v = global.Host
	mutex.Unlock()
	return
}

// SetHost safely sets the value for global configuration 'Host' field
func SetHost(v string) {
	mutex.Lock()
	global.Host = v
	mutex.Unlock()
}

// AccountDomainFlag returns the flag name for the 'AccountDomain' field
func AccountDomainFlag() string {
	return "account-domain"
}

// GetAccountDomain safely fetches the value for global configuration 'AccountDomain' field
func GetAccountDomain() (v string) {
	mutex.Lock()
	v = global.AccountDomain
	mutex.Unlock()
	return
}

// SetAccountDomain safely sets the value for global configuration 'AccountDomain' field
func SetAccountDomain(v string) {
	mutex.Lock()
	global.AccountDomain = v
	mutex.Unlock()
}

// ProtocolFlag returns the flag name for the 'Protocol' field
func ProtocolFlag() string {
	return "protocol"
}

// GetProtocol safely fetches the value for global configuration 'Protocol' field
func GetProtocol() (v string) {
	mutex.Lock()
	v = global.Protocol
	mutex.Unlock()
	return
}

// SetProtocol safely sets the value for global configuration 'Protocol' field
func SetProtocol(v string) {
	mutex.Lock()
	global.Protocol = v
	mutex.Unlock()
}

// BindAddressFlag returns the flag name for the 'BindAddress' field
func BindAddressFlag() string {
	return "bind-address"
}

// GetBindAddress safely fetches the value for global configuration 'BindAddress' field
func GetBindAddress() (v string) {
	mutex.Lock()
	v = global.BindAddress
	mutex.Unlock()
	return
}

// SetBindAddress safely sets the value for global configuration 'BindAddress' field
func SetBindAddress(v string) {
	mutex.Lock()
	global.BindAddress = v
	mutex.Unlock()
}

// PortFlag returns the flag name for the 'Port' field
func PortFlag() string {
	return "port"
}

// GetPort safely fetches the value for global configuration 'Port' field
func GetPort() (v int) {
	mutex.Lock()
	v = global.Port
	mutex.Unlock()
	return
}

// SetPort safely sets the value for global configuration 'Port' field
func SetPort(v int) {
	mutex.Lock()
	global.Port = v
	mutex.Unlock()
}

// TrustedProxiesFlag returns the flag name for the 'TrustedProxies' field
func TrustedProxiesFlag() string {
	return "trusted-proxies"
}

// GetTrustedProxies safely fetches the value for global configuration 'TrustedProxies' field
func GetTrustedProxies() (v []string) {
	mutex.Lock()
	v = global.TrustedProxies
	mutex.Unlock()
	return
}

// SetTrustedProxies safely sets the value for global configuration 'TrustedProxies' field
func SetTrustedProxies(v []string) {
	mutex.Lock()
	global.TrustedProxies = v
	mutex.Unlock()
}

// SoftwareVersionFlag returns the flag name for the 'SoftwareVersion' field
func SoftwareVersionFlag() string {
	return "software-version"
}

// GetSoftwareVersion safely fetches the value for global configuration 'SoftwareVersion' field
func GetSoftwareVersion() (v string) {
	mutex.Lock()
	v = global.SoftwareVersion
	mutex.Unlock()
	return
}

// SetSoftwareVersion safely sets the value for global configuration 'SoftwareVersion' field
func SetSoftwareVersion(v string) {
	mutex.Lock()
	global.SoftwareVersion = v
	mutex.Unlock()
}

// DbTypeFlag returns the flag name for the 'DbType' field
func DbTypeFlag() string {
	return "db-type"
}

// GetDbType safely fetches the value for global configuration 'DbType' field
func GetDbType() (v string) {
	mutex.Lock()
	v = global.DbType
	mutex.Unlock()
	return
}

// SetDbType safely sets the value for global configuration 'DbType' field
func SetDbType(v string) {
	mutex.Lock()
	global.DbType = v
	mutex.Unlock()
}

// DbAddressFlag returns the flag name for the 'DbAddress' field
func DbAddressFlag() string {
	return "db-address"
}

// GetDbAddress safely fetches the value for global configuration 'DbAddress' field
func GetDbAddress() (v string) {
	mutex.Lock()
	v = global.DbAddress
	mutex.Unlock()
	return
}

// SetDbAddress safely sets the value for global configuration 'DbAddress' field
func SetDbAddress(v string) {
	mutex.Lock()
	global.DbAddress = v
	mutex.Unlock()
}

// DbPortFlag returns the flag name for the 'DbPort' field
func DbPortFlag() string {
	return "db-port"
}

// GetDbPort safely fetches the value for global configuration 'DbPort' field
func GetDbPort() (v int) {
	mutex.Lock()
	v = global.DbPort
	mutex.Unlock()
	return
}

// SetDbPort safely sets the value for global configuration 'DbPort' field
func SetDbPort(v int) {
	mutex.Lock()
	global.DbPort = v
	mutex.Unlock()
}

// DbUserFlag returns the flag name for the 'DbUser' field
func DbUserFlag() string {
	return "db-user"
}

// GetDbUser safely fetches the value for global configuration 'DbUser' field
func GetDbUser() (v string) {
	mutex.Lock()
	v = global.DbUser
	mutex.Unlock()
	return
}

// SetDbUser safely sets the value for global configuration 'DbUser' field
func SetDbUser(v string) {
	mutex.Lock()
	global.DbUser = v
	mutex.Unlock()
}

// DbPasswordFlag returns the flag name for the 'DbPassword' field
func DbPasswordFlag() string {
	return "db-password"
}

// GetDbPassword safely fetches the value for global configuration 'DbPassword' field
func GetDbPassword() (v string) {
	mutex.Lock()
	v = global.DbPassword
	mutex.Unlock()
	return
}

// SetDbPassword safely sets the value for global configuration 'DbPassword' field
func SetDbPassword(v string) {
	mutex.Lock()
	global.DbPassword = v
	mutex.Unlock()
}

// DbDatabaseFlag returns the flag name for the 'DbDatabase' field
func DbDatabaseFlag() string {
	return "db-database"
}

// GetDbDatabase safely fetches the value for global configuration 'DbDatabase' field
func GetDbDatabase() (v string) {
	mutex.Lock()
	v = global.DbDatabase
	mutex.Unlock()
	return
}

// SetDbDatabase safely sets the value for global configuration 'DbDatabase' field
func SetDbDatabase(v string) {
	mutex.Lock()
	global.DbDatabase = v
	mutex.Unlock()
}

// DbTLSModeFlag returns the flag name for the 'DbTLSMode' field
func DbTLSModeFlag() string {
	return "db-tls-mode"
}

// GetDbTLSMode safely fetches the value for global configuration 'DbTLSMode' field
func GetDbTLSMode() (v string) {
	mutex.Lock()
	v = global.DbTLSMode
	mutex.Unlock()
	return
}

// SetDbTLSMode safely sets the value for global configuration 'DbTLSMode' field
func SetDbTLSMode(v string) {
	mutex.Lock()
	global.DbTLSMode = v
	mutex.Unlock()
}

// DbTLSCACertFlag returns the flag name for the 'DbTLSCACert' field
func DbTLSCACertFlag() string {
	return "db-tls-ca-cert"
}

// GetDbTLSCACert safely fetches the value for global configuration 'DbTLSCACert' field
func GetDbTLSCACert() (v string) {
	mutex.Lock()
	v = global.DbTLSCACert
	mutex.Unlock()
	return
}

// SetDbTLSCACert safely sets the value for global configuration 'DbTLSCACert' field
func SetDbTLSCACert(v string) {
	mutex.Lock()
	global.DbTLSCACert = v
	mutex.Unlock()
}

// WebTemplateBaseDirFlag returns the flag name for the 'WebTemplateBaseDir' field
func WebTemplateBaseDirFlag() string {
	return "web-template-base-dir"
}

// GetWebTemplateBaseDir safely fetches the value for global configuration 'WebTemplateBaseDir' field
func GetWebTemplateBaseDir() (v string) {
	mutex.Lock()
	v = global.WebTemplateBaseDir
	mutex.Unlock()
	return
}

// SetWebTemplateBaseDir safely sets the value for global configuration 'WebTemplateBaseDir' field
func SetWebTemplateBaseDir(v string) {
	mutex.Lock()
	global.WebTemplateBaseDir = v
	mutex.Unlock()
}

// WebAssetBaseDirFlag returns the flag name for the 'WebAssetBaseDir' field
func WebAssetBaseDirFlag() string {
	return "web-asset-base-dir"
}

// GetWebAssetBaseDir safely fetches the value for global configuration 'WebAssetBaseDir' field
func GetWebAssetBaseDir() (v string) {
	mutex.Lock()
	v = global.WebAssetBaseDir
	mutex.Unlock()
	return
}

// SetWebAssetBaseDir safely sets the value for global configuration 'WebAssetBaseDir' field
func SetWebAssetBaseDir(v string) {
	mutex.Lock()
	global.WebAssetBaseDir = v
	mutex.Unlock()
}

// AccountsRegistrationOpenFlag returns the flag name for the 'AccountsRegistrationOpen' field
func AccountsRegistrationOpenFlag() string {
	return "accounts-registration-open"
}

// GetAccountsRegistrationOpen safely fetches the value for global configuration 'AccountsRegistrationOpen' field
func GetAccountsRegistrationOpen() (v bool) {
	mutex.Lock()
	v = global.AccountsRegistrationOpen
	mutex.Unlock()
	return
}

// SetAccountsRegistrationOpen safely sets the value for global configuration 'AccountsRegistrationOpen' field
func SetAccountsRegistrationOpen(v bool) {
	mutex.Lock()
	global.AccountsRegistrationOpen = v
	mutex.Unlock()
}

// AccountsApprovalRequiredFlag returns the flag name for the 'AccountsApprovalRequired' field
func AccountsApprovalRequiredFlag() string {
	return "accounts-approval-required"
}

// GetAccountsApprovalRequired safely fetches the value for global configuration 'AccountsApprovalRequired' field
func GetAccountsApprovalRequired() (v bool) {
	mutex.Lock()
	v = global.AccountsApprovalRequired
	mutex.Unlock()
	return
}

// SetAccountsApprovalRequired safely sets the value for global configuration 'AccountsApprovalRequired' field
func SetAccountsApprovalRequired(v bool) {
	mutex.Lock()
	global.AccountsApprovalRequired = v
	mutex.Unlock()
}

// AccountsReasonRequiredFlag returns the flag name for the 'AccountsReasonRequired' field
func AccountsReasonRequiredFlag() string {
	return "accounts-reason-required"
}

// GetAccountsReasonRequired safely fetches the value for global configuration 'AccountsReasonRequired' field
func GetAccountsReasonRequired() (v bool) {
	mutex.Lock()
	v = global.AccountsReasonRequired
	mutex.Unlock()
	return
}

// SetAccountsReasonRequired safely sets the value for global configuration 'AccountsReasonRequired' field
func SetAccountsReasonRequired(v bool) {
	mutex.Lock()
	global.AccountsReasonRequired = v
	mutex.Unlock()
}

// MediaImageMaxSizeFlag returns the flag name for the 'MediaImageMaxSize' field
func MediaImageMaxSizeFlag() string {
	return "media-image-max-size"
}

// GetMediaImageMaxSize safely fetches the value for global configuration 'MediaImageMaxSize' field
func GetMediaImageMaxSize() (v int) {
	mutex.Lock()
	v = global.MediaImageMaxSize
	mutex.Unlock()
	return
}

// SetMediaImageMaxSize safely sets the value for global configuration 'MediaImageMaxSize' field
func SetMediaImageMaxSize(v int) {
	mutex.Lock()
	global.MediaImageMaxSize = v
	mutex.Unlock()
}

// MediaVideoMaxSizeFlag returns the flag name for the 'MediaVideoMaxSize' field
func MediaVideoMaxSizeFlag() string {
	return "media-video-max-size"
}

// GetMediaVideoMaxSize safely fetches the value for global configuration 'MediaVideoMaxSize' field
func GetMediaVideoMaxSize() (v int) {
	mutex.Lock()
	v = global.MediaVideoMaxSize
	mutex.Unlock()
	return
}

// SetMediaVideoMaxSize safely sets the value for global configuration 'MediaVideoMaxSize' field
func SetMediaVideoMaxSize(v int) {
	mutex.Lock()
	global.MediaVideoMaxSize = v
	mutex.Unlock()
}

// MediaDescriptionMinCharsFlag returns the flag name for the 'MediaDescriptionMinChars' field
func MediaDescriptionMinCharsFlag() string {
	return "media-description-min-chars"
}

// GetMediaDescriptionMinChars safely fetches the value for global configuration 'MediaDescriptionMinChars' field
func GetMediaDescriptionMinChars() (v int) {
	mutex.Lock()
	v = global.MediaDescriptionMinChars
	mutex.Unlock()
	return
}

// SetMediaDescriptionMinChars safely sets the value for global configuration 'MediaDescriptionMinChars' field
func SetMediaDescriptionMinChars(v int) {
	mutex.Lock()
	global.MediaDescriptionMinChars = v
	mutex.Unlock()
}

// MediaDescriptionMaxCharsFlag returns the flag name for the 'MediaDescriptionMaxChars' field
func MediaDescriptionMaxCharsFlag() string {
	return "media-description-max-chars"
}

// GetMediaDescriptionMaxChars safely fetches the value for global configuration 'MediaDescriptionMaxChars' field
func GetMediaDescriptionMaxChars() (v int) {
	mutex.Lock()
	v = global.MediaDescriptionMaxChars
	mutex.Unlock()
	return
}

// SetMediaDescriptionMaxChars safely sets the value for global configuration 'MediaDescriptionMaxChars' field
func SetMediaDescriptionMaxChars(v int) {
	mutex.Lock()
	global.MediaDescriptionMaxChars = v
	mutex.Unlock()
}

// MediaRemoteCacheDaysFlag returns the flag name for the 'MediaRemoteCacheDays' field
func MediaRemoteCacheDaysFlag() string {
	return "media-remote-cache-days"
}

// GetMediaRemoteCacheDays safely fetches the value for global configuration 'MediaRemoteCacheDays' field
func GetMediaRemoteCacheDays() (v int) {
	mutex.Lock()
	v = global.MediaRemoteCacheDays
	mutex.Unlock()
	return
}

// SetMediaRemoteCacheDays safely sets the value for global configuration 'MediaRemoteCacheDays' field
func SetMediaRemoteCacheDays(v int) {
	mutex.Lock()
	global.MediaRemoteCacheDays = v
	mutex.Unlock()
}

// StorageBackendFlag returns the flag name for the 'StorageBackend' field
func StorageBackendFlag() string {
	return "storage-backend"
}

// GetStorageBackend safely fetches the value for global configuration 'StorageBackend' field
func GetStorageBackend() (v string) {
	mutex.Lock()
	v = global.StorageBackend
	mutex.Unlock()
	return
}

// SetStorageBackend safely sets the value for global configuration 'StorageBackend' field
func SetStorageBackend(v string) {
	mutex.Lock()
	global.StorageBackend = v
	mutex.Unlock()
}

// StorageLocalBasePathFlag returns the flag name for the 'StorageLocalBasePath' field
func StorageLocalBasePathFlag() string {
	return "storage-local-base-path"
}

// GetStorageLocalBasePath safely fetches the value for global configuration 'StorageLocalBasePath' field
func GetStorageLocalBasePath() (v string) {
	mutex.Lock()
	v = global.StorageLocalBasePath
	mutex.Unlock()
	return
}

// SetStorageLocalBasePath safely sets the value for global configuration 'StorageLocalBasePath' field
func SetStorageLocalBasePath(v string) {
	mutex.Lock()
	global.StorageLocalBasePath = v
	mutex.Unlock()
}

// StatusesMaxCharsFlag returns the flag name for the 'StatusesMaxChars' field
func StatusesMaxCharsFlag() string {
	return "statuses-max-chars"
}

// GetStatusesMaxChars safely fetches the value for global configuration 'StatusesMaxChars' field
func GetStatusesMaxChars() (v int) {
	mutex.Lock()
	v = global.StatusesMaxChars
	mutex.Unlock()
	return
}

// SetStatusesMaxChars safely sets the value for global configuration 'StatusesMaxChars' field
func SetStatusesMaxChars(v int) {
	mutex.Lock()
	global.StatusesMaxChars = v
	mutex.Unlock()
}

// StatusesCWMaxCharsFlag returns the flag name for the 'StatusesCWMaxChars' field
func StatusesCWMaxCharsFlag() string {
	return "statuses-cw-max-chars"
}

// GetStatusesCWMaxChars safely fetches the value for global configuration 'StatusesCWMaxChars' field
func GetStatusesCWMaxChars() (v int) {
	mutex.Lock()
	v = global.StatusesCWMaxChars
	mutex.Unlock()
	return
}

// SetStatusesCWMaxChars safely sets the value for global configuration 'StatusesCWMaxChars' field
func SetStatusesCWMaxChars(v int) {
	mutex.Lock()
	global.StatusesCWMaxChars = v
	mutex.Unlock()
}

// StatusesPollMaxOptionsFlag returns the flag name for the 'StatusesPollMaxOptions' field
func StatusesPollMaxOptionsFlag() string {
	return "statuses-poll-max-options"
}

// GetStatusesPollMaxOptions safely fetches the value for global configuration 'StatusesPollMaxOptions' field
func GetStatusesPollMaxOptions() (v int) {
	mutex.Lock()
	v = global.StatusesPollMaxOptions
	mutex.Unlock()
	return
}

// SetStatusesPollMaxOptions safely sets the value for global configuration 'StatusesPollMaxOptions' field
func SetStatusesPollMaxOptions(v int) {
	mutex.Lock()
	global.StatusesPollMaxOptions = v
	mutex.Unlock()
}

// StatusesPollOptionMaxCharsFlag returns the flag name for the 'StatusesPollOptionMaxChars' field
func StatusesPollOptionMaxCharsFlag() string {
	return "statuses-poll-option-max-chars"
}

// GetStatusesPollOptionMaxChars safely fetches the value for global configuration 'StatusesPollOptionMaxChars' field
func GetStatusesPollOptionMaxChars() (v int) {
	mutex.Lock()
	v = global.StatusesPollOptionMaxChars
	mutex.Unlock()
	return
}

// SetStatusesPollOptionMaxChars safely sets the value for global configuration 'StatusesPollOptionMaxChars' field
func SetStatusesPollOptionMaxChars(v int) {
	mutex.Lock()
	global.StatusesPollOptionMaxChars = v
	mutex.Unlock()
}

// StatusesMediaMaxFilesFlag returns the flag name for the 'StatusesMediaMaxFiles' field
func StatusesMediaMaxFilesFlag() string {
	return "statuses-media-max-files"
}

// GetStatusesMediaMaxFiles safely fetches the value for global configuration 'StatusesMediaMaxFiles' field
func GetStatusesMediaMaxFiles() (v int) {
	mutex.Lock()
	v = global.StatusesMediaMaxFiles
	mutex.Unlock()
	return
}

// SetStatusesMediaMaxFiles safely sets the value for global configuration 'StatusesMediaMaxFiles' field
func SetStatusesMediaMaxFiles(v int) {
	mutex.Lock()
	global.StatusesMediaMaxFiles = v
	mutex.Unlock()
}

// LetsEncryptEnabledFlag returns the flag name for the 'LetsEncryptEnabled' field
func LetsEncryptEnabledFlag() string {
	return "letsencrypt-enabled"
}

// GetLetsEncryptEnabled safely fetches the value for global configuration 'LetsEncryptEnabled' field
func GetLetsEncryptEnabled() (v bool) {
	mutex.Lock()
	v = global.LetsEncryptEnabled
	mutex.Unlock()
	return
}

// SetLetsEncryptEnabled safely sets the value for global configuration 'LetsEncryptEnabled' field
func SetLetsEncryptEnabled(v bool) {
	mutex.Lock()
	global.LetsEncryptEnabled = v
	mutex.Unlock()
}

// LetsEncryptPortFlag returns the flag name for the 'LetsEncryptPort' field
func LetsEncryptPortFlag() string {
	return "letsencrypt-port"
}

// GetLetsEncryptPort safely fetches the value for global configuration 'LetsEncryptPort' field
func GetLetsEncryptPort() (v int) {
	mutex.Lock()
	v = global.LetsEncryptPort
	mutex.Unlock()
	return
}

// SetLetsEncryptPort safely sets the value for global configuration 'LetsEncryptPort' field
func SetLetsEncryptPort(v int) {
	mutex.Lock()
	global.LetsEncryptPort = v
	mutex.Unlock()
}

// LetsEncryptCertDirFlag returns the flag name for the 'LetsEncryptCertDir' field
func LetsEncryptCertDirFlag() string {
	return "letsencrypt-cert-dir"
}

// GetLetsEncryptCertDir safely fetches the value for global configuration 'LetsEncryptCertDir' field
func GetLetsEncryptCertDir() (v string) {
	mutex.Lock()
	v = global.LetsEncryptCertDir
	mutex.Unlock()
	return
}

// SetLetsEncryptCertDir safely sets the value for global configuration 'LetsEncryptCertDir' field
func SetLetsEncryptCertDir(v string) {
	mutex.Lock()
	global.LetsEncryptCertDir = v
	mutex.Unlock()
}

// LetsEncryptEmailAddressFlag returns the flag name for the 'LetsEncryptEmailAddress' field
func LetsEncryptEmailAddressFlag() string {
	return "letsencrypt-email-address"
}

// GetLetsEncryptEmailAddress safely fetches the value for global configuration 'LetsEncryptEmailAddress' field
func GetLetsEncryptEmailAddress() (v string) {
	mutex.Lock()
	v = global.LetsEncryptEmailAddress
	mutex.Unlock()
	return
}

// SetLetsEncryptEmailAddress safely sets the value for global configuration 'LetsEncryptEmailAddress' field
func SetLetsEncryptEmailAddress(v string) {
	mutex.Lock()
	global.LetsEncryptEmailAddress = v
	mutex.Unlock()
}

// OIDCEnabledFlag returns the flag name for the 'OIDCEnabled' field
func OIDCEnabledFlag() string {
	return "oidc-enabled"
}

// GetOIDCEnabled safely fetches the value for global configuration 'OIDCEnabled' field
func GetOIDCEnabled() (v bool) {
	mutex.Lock()
	v = global.OIDCEnabled
	mutex.Unlock()
	return
}

// SetOIDCEnabled safely sets the value for global configuration 'OIDCEnabled' field
func SetOIDCEnabled(v bool) {
	mutex.Lock()
	global.OIDCEnabled = v
	mutex.Unlock()
}

// OIDCIdpNameFlag returns the flag name for the 'OIDCIdpName' field
func OIDCIdpNameFlag() string {
	return "oidc-idp-name"
}

// GetOIDCIdpName safely fetches the value for global configuration 'OIDCIdpName' field
func GetOIDCIdpName() (v string) {
	mutex.Lock()
	v = global.OIDCIdpName
	mutex.Unlock()
	return
}

// SetOIDCIdpName safely sets the value for global configuration 'OIDCIdpName' field
func SetOIDCIdpName(v string) {
	mutex.Lock()
	global.OIDCIdpName = v
	mutex.Unlock()
}

// OIDCSkipVerificationFlag returns the flag name for the 'OIDCSkipVerification' field
func OIDCSkipVerificationFlag() string {
	return "oidc-skip-verification"
}

// GetOIDCSkipVerification safely fetches the value for global configuration 'OIDCSkipVerification' field
func GetOIDCSkipVerification() (v bool) {
	mutex.Lock()
	v = global.OIDCSkipVerification
	mutex.Unlock()
	return
}

// SetOIDCSkipVerification safely sets the value for global configuration 'OIDCSkipVerification' field
func SetOIDCSkipVerification(v bool) {
	mutex.Lock()
	global.OIDCSkipVerification = v
	mutex.Unlock()
}

// OIDCIssuerFlag returns the flag name for the 'OIDCIssuer' field
func OIDCIssuerFlag() string {
	return "oidc-issuer"
}

// GetOIDCIssuer safely fetches the value for global configuration 'OIDCIssuer' field
func GetOIDCIssuer() (v string) {
	mutex.Lock()
	v = global.OIDCIssuer
	mutex.Unlock()
	return
}

// SetOIDCIssuer safely sets the value for global configuration 'OIDCIssuer' field
func SetOIDCIssuer(v string) {
	mutex.Lock()
	global.OIDCIssuer = v
	mutex.Unlock()
}

// OIDCClientIDFlag returns the flag name for the 'OIDCClientID' field
func OIDCClientIDFlag() string {
	return "oidc-client-id"
}

// GetOIDCClientID safely fetches the value for global configuration 'OIDCClientID' field
func GetOIDCClientID() (v string) {
	mutex.Lock()
	v = global.OIDCClientID
	mutex.Unlock()
	return
}

// SetOIDCClientID safely sets the value for global configuration 'OIDCClientID' field
func SetOIDCClientID(v string) {
	mutex.Lock()
	global.OIDCClientID = v
	mutex.Unlock()
}

// OIDCClientSecretFlag returns the flag name for the 'OIDCClientSecret' field
func OIDCClientSecretFlag() string {
	return "oidc-client-secret"
}

// GetOIDCClientSecret safely fetches the value for global configuration 'OIDCClientSecret' field
func GetOIDCClientSecret() (v string) {
	mutex.Lock()
	v = global.OIDCClientSecret
	mutex.Unlock()
	return
}

// SetOIDCClientSecret safely sets the value for global configuration 'OIDCClientSecret' field
func SetOIDCClientSecret(v string) {
	mutex.Lock()
	global.OIDCClientSecret = v
	mutex.Unlock()
}

// OIDCScopesFlag returns the flag name for the 'OIDCScopes' field
func OIDCScopesFlag() string {
	return "oidc-scopes"
}

// GetOIDCScopes safely fetches the value for global configuration 'OIDCScopes' field
func GetOIDCScopes() (v []string) {
	mutex.Lock()
	v = global.OIDCScopes
	mutex.Unlock()
	return
}

// SetOIDCScopes safely sets the value for global configuration 'OIDCScopes' field
func SetOIDCScopes(v []string) {
	mutex.Lock()
	global.OIDCScopes = v
	mutex.Unlock()
}

// SMTPHostFlag returns the flag name for the 'SMTPHost' field
func SMTPHostFlag() string {
	return "smtp-host"
}

// GetSMTPHost safely fetches the value for global configuration 'SMTPHost' field
func GetSMTPHost() (v string) {
	mutex.Lock()
	v = global.SMTPHost
	mutex.Unlock()
	return
}

// SetSMTPHost safely sets the value for global configuration 'SMTPHost' field
func SetSMTPHost(v string) {
	mutex.Lock()
	global.SMTPHost = v
	mutex.Unlock()
}

// SMTPPortFlag returns the flag name for the 'SMTPPort' field
func SMTPPortFlag() string {
	return "smtp-port"
}

// GetSMTPPort safely fetches the value for global configuration 'SMTPPort' field
func GetSMTPPort() (v int) {
	mutex.Lock()
	v = global.SMTPPort
	mutex.Unlock()
	return
}

// SetSMTPPort safely sets the value for global configuration 'SMTPPort' field
func SetSMTPPort(v int) {
	mutex.Lock()
	global.SMTPPort = v
	mutex.Unlock()
}

// SMTPUsernameFlag returns the flag name for the 'SMTPUsername' field
func SMTPUsernameFlag() string {
	return "smtp-username"
}

// GetSMTPUsername safely fetches the value for global configuration 'SMTPUsername' field
func GetSMTPUsername() (v string) {
	mutex.Lock()
	v = global.SMTPUsername
	mutex.Unlock()
	return
}

// SetSMTPUsername safely sets the value for global configuration 'SMTPUsername' field
func SetSMTPUsername(v string) {
	mutex.Lock()
	global.SMTPUsername = v
	mutex.Unlock()
}

// SMTPPasswordFlag returns the flag name for the 'SMTPPassword' field
func SMTPPasswordFlag() string {
	return "smtp-password"
}

// GetSMTPPassword safely fetches the value for global configuration 'SMTPPassword' field
func GetSMTPPassword() (v string) {
	mutex.Lock()
	v = global.SMTPPassword
	mutex.Unlock()
	return
}

// SetSMTPPassword safely sets the value for global configuration 'SMTPPassword' field
func SetSMTPPassword(v string) {
	mutex.Lock()
	global.SMTPPassword = v
	mutex.Unlock()
}

// SMTPFromFlag returns the flag name for the 'SMTPFrom' field
func SMTPFromFlag() string {
	return "smtp-from"
}

// GetSMTPFrom safely fetches the value for global configuration 'SMTPFrom' field
func GetSMTPFrom() (v string) {
	mutex.Lock()
	v = global.SMTPFrom
	mutex.Unlock()
	return
}

// SetSMTPFrom safely sets the value for global configuration 'SMTPFrom' field
func SetSMTPFrom(v string) {
	mutex.Lock()
	global.SMTPFrom = v
	mutex.Unlock()
}

// SyslogEnabledFlag returns the flag name for the 'SyslogEnabled' field
func SyslogEnabledFlag() string {
	return "syslog-enabled"
}

// GetSyslogEnabled safely fetches the value for global configuration 'SyslogEnabled' field
func GetSyslogEnabled() (v bool) {
	mutex.Lock()
	v = global.SyslogEnabled
	mutex.Unlock()
	return
}

// SetSyslogEnabled safely sets the value for global configuration 'SyslogEnabled' field
func SetSyslogEnabled(v bool) {
	mutex.Lock()
	global.SyslogEnabled = v
	mutex.Unlock()
}

// SyslogProtocolFlag returns the flag name for the 'SyslogProtocol' field
func SyslogProtocolFlag() string {
	return "syslog-protocol"
}

// GetSyslogProtocol safely fetches the value for global configuration 'SyslogProtocol' field
func GetSyslogProtocol() (v string) {
	mutex.Lock()
	v = global.SyslogProtocol
	mutex.Unlock()
	return
}

// SetSyslogProtocol safely sets the value for global configuration 'SyslogProtocol' field
func SetSyslogProtocol(v string) {
	mutex.Lock()
	global.SyslogProtocol = v
	mutex.Unlock()
}

// SyslogAddressFlag returns the flag name for the 'SyslogAddress' field
func SyslogAddressFlag() string {
	return "syslog-address"
}

// GetSyslogAddress safely fetches the value for global configuration 'SyslogAddress' field
func GetSyslogAddress() (v string) {
	mutex.Lock()
	v = global.SyslogAddress
	mutex.Unlock()
	return
}

// SetSyslogAddress safely sets the value for global configuration 'SyslogAddress' field
func SetSyslogAddress(v string) {
	mutex.Lock()
	global.SyslogAddress = v
	mutex.Unlock()
}

// AdminAccountUsernameFlag returns the flag name for the 'AdminAccountUsername' field
func AdminAccountUsernameFlag() string {
	return "username"
}

// GetAdminAccountUsername safely fetches the value for global configuration 'AdminAccountUsername' field
func GetAdminAccountUsername() (v string) {
	mutex.Lock()
	v = global.AdminAccountUsername
	mutex.Unlock()
	return
}

// SetAdminAccountUsername safely sets the value for global configuration 'AdminAccountUsername' field
func SetAdminAccountUsername(v string) {
	mutex.Lock()
	global.AdminAccountUsername = v
	mutex.Unlock()
}

// AdminAccountEmailFlag returns the flag name for the 'AdminAccountEmail' field
func AdminAccountEmailFlag() string {
	return "email"
}

// GetAdminAccountEmail safely fetches the value for global configuration 'AdminAccountEmail' field
func GetAdminAccountEmail() (v string) {
	mutex.Lock()
	v = global.AdminAccountEmail
	mutex.Unlock()
	return
}

// SetAdminAccountEmail safely sets the value for global configuration 'AdminAccountEmail' field
func SetAdminAccountEmail(v string) {
	mutex.Lock()
	global.AdminAccountEmail = v
	mutex.Unlock()
}

// AdminAccountPasswordFlag returns the flag name for the 'AdminAccountPassword' field
func AdminAccountPasswordFlag() string {
	return "password"
}

// GetAdminAccountPassword safely fetches the value for global configuration 'AdminAccountPassword' field
func GetAdminAccountPassword() (v string) {
	mutex.Lock()
	v = global.AdminAccountPassword
	mutex.Unlock()
	return
}

// SetAdminAccountPassword safely sets the value for global configuration 'AdminAccountPassword' field
func SetAdminAccountPassword(v string) {
	mutex.Lock()
	global.AdminAccountPassword = v
	mutex.Unlock()
}

// AdminTransPathFlag returns the flag name for the 'AdminTransPath' field
func AdminTransPathFlag() string {
	return "path"
}

// GetAdminTransPath safely fetches the value for global configuration 'AdminTransPath' field
func GetAdminTransPath() (v string) {
	mutex.Lock()
	v = global.AdminTransPath
	mutex.Unlock()
	return
}

// SetAdminTransPath safely sets the value for global configuration 'AdminTransPath' field
func SetAdminTransPath(v string) {
	mutex.Lock()
	global.AdminTransPath = v
	mutex.Unlock()
}
