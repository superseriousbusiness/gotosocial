// THIS IS A GENERATED FILE, DO NOT EDIT BY HAND
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

// GetLogLevel safely fetches the Configuration value for state's 'LogLevel' field
func (st *ConfigState) GetLogLevel() (v string) {
	st.mutex.Lock()
	v = st.config.LogLevel
	st.mutex.Unlock()
	return
}

// SetLogLevel safely sets the Configuration value for state's 'LogLevel' field
func (st *ConfigState) SetLogLevel(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LogLevel = v
	st.reloadToViper()
}

// LogLevelFlag returns the flag name for the 'LogLevel' field
func LogLevelFlag() string { return "log-level" }

// GetLogLevel safely fetches the value for global configuration 'LogLevel' field
func GetLogLevel() string { return global.GetLogLevel() }

// SetLogLevel safely sets the value for global configuration 'LogLevel' field
func SetLogLevel(v string) { global.SetLogLevel(v) }

// GetLogDbQueries safely fetches the Configuration value for state's 'LogDbQueries' field
func (st *ConfigState) GetLogDbQueries() (v bool) {
	st.mutex.Lock()
	v = st.config.LogDbQueries
	st.mutex.Unlock()
	return
}

// SetLogDbQueries safely sets the Configuration value for state's 'LogDbQueries' field
func (st *ConfigState) SetLogDbQueries(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LogDbQueries = v
	st.reloadToViper()
}

// LogDbQueriesFlag returns the flag name for the 'LogDbQueries' field
func LogDbQueriesFlag() string { return "log-db-queries" }

// GetLogDbQueries safely fetches the value for global configuration 'LogDbQueries' field
func GetLogDbQueries() bool { return global.GetLogDbQueries() }

// SetLogDbQueries safely sets the value for global configuration 'LogDbQueries' field
func SetLogDbQueries(v bool) { global.SetLogDbQueries(v) }

// GetApplicationName safely fetches the Configuration value for state's 'ApplicationName' field
func (st *ConfigState) GetApplicationName() (v string) {
	st.mutex.Lock()
	v = st.config.ApplicationName
	st.mutex.Unlock()
	return
}

// SetApplicationName safely sets the Configuration value for state's 'ApplicationName' field
func (st *ConfigState) SetApplicationName(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.ApplicationName = v
	st.reloadToViper()
}

// ApplicationNameFlag returns the flag name for the 'ApplicationName' field
func ApplicationNameFlag() string { return "application-name" }

// GetApplicationName safely fetches the value for global configuration 'ApplicationName' field
func GetApplicationName() string { return global.GetApplicationName() }

// SetApplicationName safely sets the value for global configuration 'ApplicationName' field
func SetApplicationName(v string) { global.SetApplicationName(v) }

// GetConfigPath safely fetches the Configuration value for state's 'ConfigPath' field
func (st *ConfigState) GetConfigPath() (v string) {
	st.mutex.Lock()
	v = st.config.ConfigPath
	st.mutex.Unlock()
	return
}

// SetConfigPath safely sets the Configuration value for state's 'ConfigPath' field
func (st *ConfigState) SetConfigPath(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.ConfigPath = v
	st.reloadToViper()
}

// ConfigPathFlag returns the flag name for the 'ConfigPath' field
func ConfigPathFlag() string { return "config-path" }

// GetConfigPath safely fetches the value for global configuration 'ConfigPath' field
func GetConfigPath() string { return global.GetConfigPath() }

// SetConfigPath safely sets the value for global configuration 'ConfigPath' field
func SetConfigPath(v string) { global.SetConfigPath(v) }

// GetHost safely fetches the Configuration value for state's 'Host' field
func (st *ConfigState) GetHost() (v string) {
	st.mutex.Lock()
	v = st.config.Host
	st.mutex.Unlock()
	return
}

// SetHost safely sets the Configuration value for state's 'Host' field
func (st *ConfigState) SetHost(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Host = v
	st.reloadToViper()
}

// HostFlag returns the flag name for the 'Host' field
func HostFlag() string { return "host" }

// GetHost safely fetches the value for global configuration 'Host' field
func GetHost() string { return global.GetHost() }

// SetHost safely sets the value for global configuration 'Host' field
func SetHost(v string) { global.SetHost(v) }

// GetAccountDomain safely fetches the Configuration value for state's 'AccountDomain' field
func (st *ConfigState) GetAccountDomain() (v string) {
	st.mutex.Lock()
	v = st.config.AccountDomain
	st.mutex.Unlock()
	return
}

// SetAccountDomain safely sets the Configuration value for state's 'AccountDomain' field
func (st *ConfigState) SetAccountDomain(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountDomain = v
	st.reloadToViper()
}

// AccountDomainFlag returns the flag name for the 'AccountDomain' field
func AccountDomainFlag() string { return "account-domain" }

// GetAccountDomain safely fetches the value for global configuration 'AccountDomain' field
func GetAccountDomain() string { return global.GetAccountDomain() }

// SetAccountDomain safely sets the value for global configuration 'AccountDomain' field
func SetAccountDomain(v string) { global.SetAccountDomain(v) }

// GetProtocol safely fetches the Configuration value for state's 'Protocol' field
func (st *ConfigState) GetProtocol() (v string) {
	st.mutex.Lock()
	v = st.config.Protocol
	st.mutex.Unlock()
	return
}

// SetProtocol safely sets the Configuration value for state's 'Protocol' field
func (st *ConfigState) SetProtocol(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Protocol = v
	st.reloadToViper()
}

// ProtocolFlag returns the flag name for the 'Protocol' field
func ProtocolFlag() string { return "protocol" }

// GetProtocol safely fetches the value for global configuration 'Protocol' field
func GetProtocol() string { return global.GetProtocol() }

// SetProtocol safely sets the value for global configuration 'Protocol' field
func SetProtocol(v string) { global.SetProtocol(v) }

// GetBindAddress safely fetches the Configuration value for state's 'BindAddress' field
func (st *ConfigState) GetBindAddress() (v string) {
	st.mutex.Lock()
	v = st.config.BindAddress
	st.mutex.Unlock()
	return
}

// SetBindAddress safely sets the Configuration value for state's 'BindAddress' field
func (st *ConfigState) SetBindAddress(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.BindAddress = v
	st.reloadToViper()
}

// BindAddressFlag returns the flag name for the 'BindAddress' field
func BindAddressFlag() string { return "bind-address" }

// GetBindAddress safely fetches the value for global configuration 'BindAddress' field
func GetBindAddress() string { return global.GetBindAddress() }

// SetBindAddress safely sets the value for global configuration 'BindAddress' field
func SetBindAddress(v string) { global.SetBindAddress(v) }

// GetPort safely fetches the Configuration value for state's 'Port' field
func (st *ConfigState) GetPort() (v int) {
	st.mutex.Lock()
	v = st.config.Port
	st.mutex.Unlock()
	return
}

// SetPort safely sets the Configuration value for state's 'Port' field
func (st *ConfigState) SetPort(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Port = v
	st.reloadToViper()
}

// PortFlag returns the flag name for the 'Port' field
func PortFlag() string { return "port" }

// GetPort safely fetches the value for global configuration 'Port' field
func GetPort() int { return global.GetPort() }

// SetPort safely sets the value for global configuration 'Port' field
func SetPort(v int) { global.SetPort(v) }

// GetTrustedProxies safely fetches the Configuration value for state's 'TrustedProxies' field
func (st *ConfigState) GetTrustedProxies() (v []string) {
	st.mutex.Lock()
	v = st.config.TrustedProxies
	st.mutex.Unlock()
	return
}

// SetTrustedProxies safely sets the Configuration value for state's 'TrustedProxies' field
func (st *ConfigState) SetTrustedProxies(v []string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.TrustedProxies = v
	st.reloadToViper()
}

// TrustedProxiesFlag returns the flag name for the 'TrustedProxies' field
func TrustedProxiesFlag() string { return "trusted-proxies" }

// GetTrustedProxies safely fetches the value for global configuration 'TrustedProxies' field
func GetTrustedProxies() []string { return global.GetTrustedProxies() }

// SetTrustedProxies safely sets the value for global configuration 'TrustedProxies' field
func SetTrustedProxies(v []string) { global.SetTrustedProxies(v) }

// GetSoftwareVersion safely fetches the Configuration value for state's 'SoftwareVersion' field
func (st *ConfigState) GetSoftwareVersion() (v string) {
	st.mutex.Lock()
	v = st.config.SoftwareVersion
	st.mutex.Unlock()
	return
}

// SetSoftwareVersion safely sets the Configuration value for state's 'SoftwareVersion' field
func (st *ConfigState) SetSoftwareVersion(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SoftwareVersion = v
	st.reloadToViper()
}

// SoftwareVersionFlag returns the flag name for the 'SoftwareVersion' field
func SoftwareVersionFlag() string { return "software-version" }

// GetSoftwareVersion safely fetches the value for global configuration 'SoftwareVersion' field
func GetSoftwareVersion() string { return global.GetSoftwareVersion() }

// SetSoftwareVersion safely sets the value for global configuration 'SoftwareVersion' field
func SetSoftwareVersion(v string) { global.SetSoftwareVersion(v) }

// GetDbType safely fetches the Configuration value for state's 'DbType' field
func (st *ConfigState) GetDbType() (v string) {
	st.mutex.Lock()
	v = st.config.DbType
	st.mutex.Unlock()
	return
}

// SetDbType safely sets the Configuration value for state's 'DbType' field
func (st *ConfigState) SetDbType(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbType = v
	st.reloadToViper()
}

// DbTypeFlag returns the flag name for the 'DbType' field
func DbTypeFlag() string { return "db-type" }

// GetDbType safely fetches the value for global configuration 'DbType' field
func GetDbType() string { return global.GetDbType() }

// SetDbType safely sets the value for global configuration 'DbType' field
func SetDbType(v string) { global.SetDbType(v) }

// GetDbAddress safely fetches the Configuration value for state's 'DbAddress' field
func (st *ConfigState) GetDbAddress() (v string) {
	st.mutex.Lock()
	v = st.config.DbAddress
	st.mutex.Unlock()
	return
}

// SetDbAddress safely sets the Configuration value for state's 'DbAddress' field
func (st *ConfigState) SetDbAddress(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbAddress = v
	st.reloadToViper()
}

// DbAddressFlag returns the flag name for the 'DbAddress' field
func DbAddressFlag() string { return "db-address" }

// GetDbAddress safely fetches the value for global configuration 'DbAddress' field
func GetDbAddress() string { return global.GetDbAddress() }

// SetDbAddress safely sets the value for global configuration 'DbAddress' field
func SetDbAddress(v string) { global.SetDbAddress(v) }

// GetDbPort safely fetches the Configuration value for state's 'DbPort' field
func (st *ConfigState) GetDbPort() (v int) {
	st.mutex.Lock()
	v = st.config.DbPort
	st.mutex.Unlock()
	return
}

// SetDbPort safely sets the Configuration value for state's 'DbPort' field
func (st *ConfigState) SetDbPort(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbPort = v
	st.reloadToViper()
}

// DbPortFlag returns the flag name for the 'DbPort' field
func DbPortFlag() string { return "db-port" }

// GetDbPort safely fetches the value for global configuration 'DbPort' field
func GetDbPort() int { return global.GetDbPort() }

// SetDbPort safely sets the value for global configuration 'DbPort' field
func SetDbPort(v int) { global.SetDbPort(v) }

// GetDbUser safely fetches the Configuration value for state's 'DbUser' field
func (st *ConfigState) GetDbUser() (v string) {
	st.mutex.Lock()
	v = st.config.DbUser
	st.mutex.Unlock()
	return
}

// SetDbUser safely sets the Configuration value for state's 'DbUser' field
func (st *ConfigState) SetDbUser(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbUser = v
	st.reloadToViper()
}

// DbUserFlag returns the flag name for the 'DbUser' field
func DbUserFlag() string { return "db-user" }

// GetDbUser safely fetches the value for global configuration 'DbUser' field
func GetDbUser() string { return global.GetDbUser() }

// SetDbUser safely sets the value for global configuration 'DbUser' field
func SetDbUser(v string) { global.SetDbUser(v) }

// GetDbPassword safely fetches the Configuration value for state's 'DbPassword' field
func (st *ConfigState) GetDbPassword() (v string) {
	st.mutex.Lock()
	v = st.config.DbPassword
	st.mutex.Unlock()
	return
}

// SetDbPassword safely sets the Configuration value for state's 'DbPassword' field
func (st *ConfigState) SetDbPassword(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbPassword = v
	st.reloadToViper()
}

// DbPasswordFlag returns the flag name for the 'DbPassword' field
func DbPasswordFlag() string { return "db-password" }

// GetDbPassword safely fetches the value for global configuration 'DbPassword' field
func GetDbPassword() string { return global.GetDbPassword() }

// SetDbPassword safely sets the value for global configuration 'DbPassword' field
func SetDbPassword(v string) { global.SetDbPassword(v) }

// GetDbDatabase safely fetches the Configuration value for state's 'DbDatabase' field
func (st *ConfigState) GetDbDatabase() (v string) {
	st.mutex.Lock()
	v = st.config.DbDatabase
	st.mutex.Unlock()
	return
}

// SetDbDatabase safely sets the Configuration value for state's 'DbDatabase' field
func (st *ConfigState) SetDbDatabase(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbDatabase = v
	st.reloadToViper()
}

// DbDatabaseFlag returns the flag name for the 'DbDatabase' field
func DbDatabaseFlag() string { return "db-database" }

// GetDbDatabase safely fetches the value for global configuration 'DbDatabase' field
func GetDbDatabase() string { return global.GetDbDatabase() }

// SetDbDatabase safely sets the value for global configuration 'DbDatabase' field
func SetDbDatabase(v string) { global.SetDbDatabase(v) }

// GetDbTLSMode safely fetches the Configuration value for state's 'DbTLSMode' field
func (st *ConfigState) GetDbTLSMode() (v string) {
	st.mutex.Lock()
	v = st.config.DbTLSMode
	st.mutex.Unlock()
	return
}

// SetDbTLSMode safely sets the Configuration value for state's 'DbTLSMode' field
func (st *ConfigState) SetDbTLSMode(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbTLSMode = v
	st.reloadToViper()
}

// DbTLSModeFlag returns the flag name for the 'DbTLSMode' field
func DbTLSModeFlag() string { return "db-tls-mode" }

// GetDbTLSMode safely fetches the value for global configuration 'DbTLSMode' field
func GetDbTLSMode() string { return global.GetDbTLSMode() }

// SetDbTLSMode safely sets the value for global configuration 'DbTLSMode' field
func SetDbTLSMode(v string) { global.SetDbTLSMode(v) }

// GetDbTLSCACert safely fetches the Configuration value for state's 'DbTLSCACert' field
func (st *ConfigState) GetDbTLSCACert() (v string) {
	st.mutex.Lock()
	v = st.config.DbTLSCACert
	st.mutex.Unlock()
	return
}

// SetDbTLSCACert safely sets the Configuration value for state's 'DbTLSCACert' field
func (st *ConfigState) SetDbTLSCACert(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbTLSCACert = v
	st.reloadToViper()
}

// DbTLSCACertFlag returns the flag name for the 'DbTLSCACert' field
func DbTLSCACertFlag() string { return "db-tls-ca-cert" }

// GetDbTLSCACert safely fetches the value for global configuration 'DbTLSCACert' field
func GetDbTLSCACert() string { return global.GetDbTLSCACert() }

// SetDbTLSCACert safely sets the value for global configuration 'DbTLSCACert' field
func SetDbTLSCACert(v string) { global.SetDbTLSCACert(v) }

// GetWebTemplateBaseDir safely fetches the Configuration value for state's 'WebTemplateBaseDir' field
func (st *ConfigState) GetWebTemplateBaseDir() (v string) {
	st.mutex.Lock()
	v = st.config.WebTemplateBaseDir
	st.mutex.Unlock()
	return
}

// SetWebTemplateBaseDir safely sets the Configuration value for state's 'WebTemplateBaseDir' field
func (st *ConfigState) SetWebTemplateBaseDir(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.WebTemplateBaseDir = v
	st.reloadToViper()
}

// WebTemplateBaseDirFlag returns the flag name for the 'WebTemplateBaseDir' field
func WebTemplateBaseDirFlag() string { return "web-template-base-dir" }

// GetWebTemplateBaseDir safely fetches the value for global configuration 'WebTemplateBaseDir' field
func GetWebTemplateBaseDir() string { return global.GetWebTemplateBaseDir() }

// SetWebTemplateBaseDir safely sets the value for global configuration 'WebTemplateBaseDir' field
func SetWebTemplateBaseDir(v string) { global.SetWebTemplateBaseDir(v) }

// GetWebAssetBaseDir safely fetches the Configuration value for state's 'WebAssetBaseDir' field
func (st *ConfigState) GetWebAssetBaseDir() (v string) {
	st.mutex.Lock()
	v = st.config.WebAssetBaseDir
	st.mutex.Unlock()
	return
}

// SetWebAssetBaseDir safely sets the Configuration value for state's 'WebAssetBaseDir' field
func (st *ConfigState) SetWebAssetBaseDir(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.WebAssetBaseDir = v
	st.reloadToViper()
}

// WebAssetBaseDirFlag returns the flag name for the 'WebAssetBaseDir' field
func WebAssetBaseDirFlag() string { return "web-asset-base-dir" }

// GetWebAssetBaseDir safely fetches the value for global configuration 'WebAssetBaseDir' field
func GetWebAssetBaseDir() string { return global.GetWebAssetBaseDir() }

// SetWebAssetBaseDir safely sets the value for global configuration 'WebAssetBaseDir' field
func SetWebAssetBaseDir(v string) { global.SetWebAssetBaseDir(v) }

// GetInstanceExposePeers safely fetches the Configuration value for state's 'InstanceExposePeers' field
func (st *ConfigState) GetInstanceExposePeers() (v bool) {
	st.mutex.Lock()
	v = st.config.InstanceExposePeers
	st.mutex.Unlock()
	return
}

// SetInstanceExposePeers safely sets the Configuration value for state's 'InstanceExposePeers' field
func (st *ConfigState) SetInstanceExposePeers(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceExposePeers = v
	st.reloadToViper()
}

// InstanceExposePeersFlag returns the flag name for the 'InstanceExposePeers' field
func InstanceExposePeersFlag() string { return "instance-expose-peers" }

// GetInstanceExposePeers safely fetches the value for global configuration 'InstanceExposePeers' field
func GetInstanceExposePeers() bool { return global.GetInstanceExposePeers() }

// SetInstanceExposePeers safely sets the value for global configuration 'InstanceExposePeers' field
func SetInstanceExposePeers(v bool) { global.SetInstanceExposePeers(v) }

// GetInstanceExposeSuspended safely fetches the Configuration value for state's 'InstanceExposeSuspended' field
func (st *ConfigState) GetInstanceExposeSuspended() (v bool) {
	st.mutex.Lock()
	v = st.config.InstanceExposeSuspended
	st.mutex.Unlock()
	return
}

// SetInstanceExposeSuspended safely sets the Configuration value for state's 'InstanceExposeSuspended' field
func (st *ConfigState) SetInstanceExposeSuspended(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceExposeSuspended = v
	st.reloadToViper()
}

// InstanceExposeSuspendedFlag returns the flag name for the 'InstanceExposeSuspended' field
func InstanceExposeSuspendedFlag() string { return "instance-expose-suspended" }

// GetInstanceExposeSuspended safely fetches the value for global configuration 'InstanceExposeSuspended' field
func GetInstanceExposeSuspended() bool { return global.GetInstanceExposeSuspended() }

// SetInstanceExposeSuspended safely sets the value for global configuration 'InstanceExposeSuspended' field
func SetInstanceExposeSuspended(v bool) { global.SetInstanceExposeSuspended(v) }

// GetAccountsRegistrationOpen safely fetches the Configuration value for state's 'AccountsRegistrationOpen' field
func (st *ConfigState) GetAccountsRegistrationOpen() (v bool) {
	st.mutex.Lock()
	v = st.config.AccountsRegistrationOpen
	st.mutex.Unlock()
	return
}

// SetAccountsRegistrationOpen safely sets the Configuration value for state's 'AccountsRegistrationOpen' field
func (st *ConfigState) SetAccountsRegistrationOpen(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountsRegistrationOpen = v
	st.reloadToViper()
}

// AccountsRegistrationOpenFlag returns the flag name for the 'AccountsRegistrationOpen' field
func AccountsRegistrationOpenFlag() string { return "accounts-registration-open" }

// GetAccountsRegistrationOpen safely fetches the value for global configuration 'AccountsRegistrationOpen' field
func GetAccountsRegistrationOpen() bool { return global.GetAccountsRegistrationOpen() }

// SetAccountsRegistrationOpen safely sets the value for global configuration 'AccountsRegistrationOpen' field
func SetAccountsRegistrationOpen(v bool) { global.SetAccountsRegistrationOpen(v) }

// GetAccountsApprovalRequired safely fetches the Configuration value for state's 'AccountsApprovalRequired' field
func (st *ConfigState) GetAccountsApprovalRequired() (v bool) {
	st.mutex.Lock()
	v = st.config.AccountsApprovalRequired
	st.mutex.Unlock()
	return
}

// SetAccountsApprovalRequired safely sets the Configuration value for state's 'AccountsApprovalRequired' field
func (st *ConfigState) SetAccountsApprovalRequired(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountsApprovalRequired = v
	st.reloadToViper()
}

// AccountsApprovalRequiredFlag returns the flag name for the 'AccountsApprovalRequired' field
func AccountsApprovalRequiredFlag() string { return "accounts-approval-required" }

// GetAccountsApprovalRequired safely fetches the value for global configuration 'AccountsApprovalRequired' field
func GetAccountsApprovalRequired() bool { return global.GetAccountsApprovalRequired() }

// SetAccountsApprovalRequired safely sets the value for global configuration 'AccountsApprovalRequired' field
func SetAccountsApprovalRequired(v bool) { global.SetAccountsApprovalRequired(v) }

// GetAccountsReasonRequired safely fetches the Configuration value for state's 'AccountsReasonRequired' field
func (st *ConfigState) GetAccountsReasonRequired() (v bool) {
	st.mutex.Lock()
	v = st.config.AccountsReasonRequired
	st.mutex.Unlock()
	return
}

// SetAccountsReasonRequired safely sets the Configuration value for state's 'AccountsReasonRequired' field
func (st *ConfigState) SetAccountsReasonRequired(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountsReasonRequired = v
	st.reloadToViper()
}

// AccountsReasonRequiredFlag returns the flag name for the 'AccountsReasonRequired' field
func AccountsReasonRequiredFlag() string { return "accounts-reason-required" }

// GetAccountsReasonRequired safely fetches the value for global configuration 'AccountsReasonRequired' field
func GetAccountsReasonRequired() bool { return global.GetAccountsReasonRequired() }

// SetAccountsReasonRequired safely sets the value for global configuration 'AccountsReasonRequired' field
func SetAccountsReasonRequired(v bool) { global.SetAccountsReasonRequired(v) }

// GetMediaImageMaxSize safely fetches the Configuration value for state's 'MediaImageMaxSize' field
func (st *ConfigState) GetMediaImageMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.MediaImageMaxSize
	st.mutex.Unlock()
	return
}

// SetMediaImageMaxSize safely sets the Configuration value for state's 'MediaImageMaxSize' field
func (st *ConfigState) SetMediaImageMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaImageMaxSize = v
	st.reloadToViper()
}

// MediaImageMaxSizeFlag returns the flag name for the 'MediaImageMaxSize' field
func MediaImageMaxSizeFlag() string { return "media-image-max-size" }

// GetMediaImageMaxSize safely fetches the value for global configuration 'MediaImageMaxSize' field
func GetMediaImageMaxSize() int { return global.GetMediaImageMaxSize() }

// SetMediaImageMaxSize safely sets the value for global configuration 'MediaImageMaxSize' field
func SetMediaImageMaxSize(v int) { global.SetMediaImageMaxSize(v) }

// GetMediaVideoMaxSize safely fetches the Configuration value for state's 'MediaVideoMaxSize' field
func (st *ConfigState) GetMediaVideoMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.MediaVideoMaxSize
	st.mutex.Unlock()
	return
}

// SetMediaVideoMaxSize safely sets the Configuration value for state's 'MediaVideoMaxSize' field
func (st *ConfigState) SetMediaVideoMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaVideoMaxSize = v
	st.reloadToViper()
}

// MediaVideoMaxSizeFlag returns the flag name for the 'MediaVideoMaxSize' field
func MediaVideoMaxSizeFlag() string { return "media-video-max-size" }

// GetMediaVideoMaxSize safely fetches the value for global configuration 'MediaVideoMaxSize' field
func GetMediaVideoMaxSize() int { return global.GetMediaVideoMaxSize() }

// SetMediaVideoMaxSize safely sets the value for global configuration 'MediaVideoMaxSize' field
func SetMediaVideoMaxSize(v int) { global.SetMediaVideoMaxSize(v) }

// GetMediaDescriptionMinChars safely fetches the Configuration value for state's 'MediaDescriptionMinChars' field
func (st *ConfigState) GetMediaDescriptionMinChars() (v int) {
	st.mutex.Lock()
	v = st.config.MediaDescriptionMinChars
	st.mutex.Unlock()
	return
}

// SetMediaDescriptionMinChars safely sets the Configuration value for state's 'MediaDescriptionMinChars' field
func (st *ConfigState) SetMediaDescriptionMinChars(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaDescriptionMinChars = v
	st.reloadToViper()
}

// MediaDescriptionMinCharsFlag returns the flag name for the 'MediaDescriptionMinChars' field
func MediaDescriptionMinCharsFlag() string { return "media-description-min-chars" }

// GetMediaDescriptionMinChars safely fetches the value for global configuration 'MediaDescriptionMinChars' field
func GetMediaDescriptionMinChars() int { return global.GetMediaDescriptionMinChars() }

// SetMediaDescriptionMinChars safely sets the value for global configuration 'MediaDescriptionMinChars' field
func SetMediaDescriptionMinChars(v int) { global.SetMediaDescriptionMinChars(v) }

// GetMediaDescriptionMaxChars safely fetches the Configuration value for state's 'MediaDescriptionMaxChars' field
func (st *ConfigState) GetMediaDescriptionMaxChars() (v int) {
	st.mutex.Lock()
	v = st.config.MediaDescriptionMaxChars
	st.mutex.Unlock()
	return
}

// SetMediaDescriptionMaxChars safely sets the Configuration value for state's 'MediaDescriptionMaxChars' field
func (st *ConfigState) SetMediaDescriptionMaxChars(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaDescriptionMaxChars = v
	st.reloadToViper()
}

// MediaDescriptionMaxCharsFlag returns the flag name for the 'MediaDescriptionMaxChars' field
func MediaDescriptionMaxCharsFlag() string { return "media-description-max-chars" }

// GetMediaDescriptionMaxChars safely fetches the value for global configuration 'MediaDescriptionMaxChars' field
func GetMediaDescriptionMaxChars() int { return global.GetMediaDescriptionMaxChars() }

// SetMediaDescriptionMaxChars safely sets the value for global configuration 'MediaDescriptionMaxChars' field
func SetMediaDescriptionMaxChars(v int) { global.SetMediaDescriptionMaxChars(v) }

// GetMediaRemoteCacheDays safely fetches the Configuration value for state's 'MediaRemoteCacheDays' field
func (st *ConfigState) GetMediaRemoteCacheDays() (v int) {
	st.mutex.Lock()
	v = st.config.MediaRemoteCacheDays
	st.mutex.Unlock()
	return
}

// SetMediaRemoteCacheDays safely sets the Configuration value for state's 'MediaRemoteCacheDays' field
func (st *ConfigState) SetMediaRemoteCacheDays(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaRemoteCacheDays = v
	st.reloadToViper()
}

// MediaRemoteCacheDaysFlag returns the flag name for the 'MediaRemoteCacheDays' field
func MediaRemoteCacheDaysFlag() string { return "media-remote-cache-days" }

// GetMediaRemoteCacheDays safely fetches the value for global configuration 'MediaRemoteCacheDays' field
func GetMediaRemoteCacheDays() int { return global.GetMediaRemoteCacheDays() }

// SetMediaRemoteCacheDays safely sets the value for global configuration 'MediaRemoteCacheDays' field
func SetMediaRemoteCacheDays(v int) { global.SetMediaRemoteCacheDays(v) }

// GetStorageBackend safely fetches the Configuration value for state's 'StorageBackend' field
func (st *ConfigState) GetStorageBackend() (v string) {
	st.mutex.Lock()
	v = st.config.StorageBackend
	st.mutex.Unlock()
	return
}

// SetStorageBackend safely sets the Configuration value for state's 'StorageBackend' field
func (st *ConfigState) SetStorageBackend(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageBackend = v
	st.reloadToViper()
}

// StorageBackendFlag returns the flag name for the 'StorageBackend' field
func StorageBackendFlag() string { return "storage-backend" }

// GetStorageBackend safely fetches the value for global configuration 'StorageBackend' field
func GetStorageBackend() string { return global.GetStorageBackend() }

// SetStorageBackend safely sets the value for global configuration 'StorageBackend' field
func SetStorageBackend(v string) { global.SetStorageBackend(v) }

// GetStorageLocalBasePath safely fetches the Configuration value for state's 'StorageLocalBasePath' field
func (st *ConfigState) GetStorageLocalBasePath() (v string) {
	st.mutex.Lock()
	v = st.config.StorageLocalBasePath
	st.mutex.Unlock()
	return
}

// SetStorageLocalBasePath safely sets the Configuration value for state's 'StorageLocalBasePath' field
func (st *ConfigState) SetStorageLocalBasePath(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageLocalBasePath = v
	st.reloadToViper()
}

// StorageLocalBasePathFlag returns the flag name for the 'StorageLocalBasePath' field
func StorageLocalBasePathFlag() string { return "storage-local-base-path" }

// GetStorageLocalBasePath safely fetches the value for global configuration 'StorageLocalBasePath' field
func GetStorageLocalBasePath() string { return global.GetStorageLocalBasePath() }

// SetStorageLocalBasePath safely sets the value for global configuration 'StorageLocalBasePath' field
func SetStorageLocalBasePath(v string) { global.SetStorageLocalBasePath(v) }

// GetStatusesMaxChars safely fetches the Configuration value for state's 'StatusesMaxChars' field
func (st *ConfigState) GetStatusesMaxChars() (v int) {
	st.mutex.Lock()
	v = st.config.StatusesMaxChars
	st.mutex.Unlock()
	return
}

// SetStatusesMaxChars safely sets the Configuration value for state's 'StatusesMaxChars' field
func (st *ConfigState) SetStatusesMaxChars(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StatusesMaxChars = v
	st.reloadToViper()
}

// StatusesMaxCharsFlag returns the flag name for the 'StatusesMaxChars' field
func StatusesMaxCharsFlag() string { return "statuses-max-chars" }

// GetStatusesMaxChars safely fetches the value for global configuration 'StatusesMaxChars' field
func GetStatusesMaxChars() int { return global.GetStatusesMaxChars() }

// SetStatusesMaxChars safely sets the value for global configuration 'StatusesMaxChars' field
func SetStatusesMaxChars(v int) { global.SetStatusesMaxChars(v) }

// GetStatusesCWMaxChars safely fetches the Configuration value for state's 'StatusesCWMaxChars' field
func (st *ConfigState) GetStatusesCWMaxChars() (v int) {
	st.mutex.Lock()
	v = st.config.StatusesCWMaxChars
	st.mutex.Unlock()
	return
}

// SetStatusesCWMaxChars safely sets the Configuration value for state's 'StatusesCWMaxChars' field
func (st *ConfigState) SetStatusesCWMaxChars(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StatusesCWMaxChars = v
	st.reloadToViper()
}

// StatusesCWMaxCharsFlag returns the flag name for the 'StatusesCWMaxChars' field
func StatusesCWMaxCharsFlag() string { return "statuses-cw-max-chars" }

// GetStatusesCWMaxChars safely fetches the value for global configuration 'StatusesCWMaxChars' field
func GetStatusesCWMaxChars() int { return global.GetStatusesCWMaxChars() }

// SetStatusesCWMaxChars safely sets the value for global configuration 'StatusesCWMaxChars' field
func SetStatusesCWMaxChars(v int) { global.SetStatusesCWMaxChars(v) }

// GetStatusesPollMaxOptions safely fetches the Configuration value for state's 'StatusesPollMaxOptions' field
func (st *ConfigState) GetStatusesPollMaxOptions() (v int) {
	st.mutex.Lock()
	v = st.config.StatusesPollMaxOptions
	st.mutex.Unlock()
	return
}

// SetStatusesPollMaxOptions safely sets the Configuration value for state's 'StatusesPollMaxOptions' field
func (st *ConfigState) SetStatusesPollMaxOptions(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StatusesPollMaxOptions = v
	st.reloadToViper()
}

// StatusesPollMaxOptionsFlag returns the flag name for the 'StatusesPollMaxOptions' field
func StatusesPollMaxOptionsFlag() string { return "statuses-poll-max-options" }

// GetStatusesPollMaxOptions safely fetches the value for global configuration 'StatusesPollMaxOptions' field
func GetStatusesPollMaxOptions() int { return global.GetStatusesPollMaxOptions() }

// SetStatusesPollMaxOptions safely sets the value for global configuration 'StatusesPollMaxOptions' field
func SetStatusesPollMaxOptions(v int) { global.SetStatusesPollMaxOptions(v) }

// GetStatusesPollOptionMaxChars safely fetches the Configuration value for state's 'StatusesPollOptionMaxChars' field
func (st *ConfigState) GetStatusesPollOptionMaxChars() (v int) {
	st.mutex.Lock()
	v = st.config.StatusesPollOptionMaxChars
	st.mutex.Unlock()
	return
}

// SetStatusesPollOptionMaxChars safely sets the Configuration value for state's 'StatusesPollOptionMaxChars' field
func (st *ConfigState) SetStatusesPollOptionMaxChars(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StatusesPollOptionMaxChars = v
	st.reloadToViper()
}

// StatusesPollOptionMaxCharsFlag returns the flag name for the 'StatusesPollOptionMaxChars' field
func StatusesPollOptionMaxCharsFlag() string { return "statuses-poll-option-max-chars" }

// GetStatusesPollOptionMaxChars safely fetches the value for global configuration 'StatusesPollOptionMaxChars' field
func GetStatusesPollOptionMaxChars() int { return global.GetStatusesPollOptionMaxChars() }

// SetStatusesPollOptionMaxChars safely sets the value for global configuration 'StatusesPollOptionMaxChars' field
func SetStatusesPollOptionMaxChars(v int) { global.SetStatusesPollOptionMaxChars(v) }

// GetStatusesMediaMaxFiles safely fetches the Configuration value for state's 'StatusesMediaMaxFiles' field
func (st *ConfigState) GetStatusesMediaMaxFiles() (v int) {
	st.mutex.Lock()
	v = st.config.StatusesMediaMaxFiles
	st.mutex.Unlock()
	return
}

// SetStatusesMediaMaxFiles safely sets the Configuration value for state's 'StatusesMediaMaxFiles' field
func (st *ConfigState) SetStatusesMediaMaxFiles(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StatusesMediaMaxFiles = v
	st.reloadToViper()
}

// StatusesMediaMaxFilesFlag returns the flag name for the 'StatusesMediaMaxFiles' field
func StatusesMediaMaxFilesFlag() string { return "statuses-media-max-files" }

// GetStatusesMediaMaxFiles safely fetches the value for global configuration 'StatusesMediaMaxFiles' field
func GetStatusesMediaMaxFiles() int { return global.GetStatusesMediaMaxFiles() }

// SetStatusesMediaMaxFiles safely sets the value for global configuration 'StatusesMediaMaxFiles' field
func SetStatusesMediaMaxFiles(v int) { global.SetStatusesMediaMaxFiles(v) }

// GetLetsEncryptEnabled safely fetches the Configuration value for state's 'LetsEncryptEnabled' field
func (st *ConfigState) GetLetsEncryptEnabled() (v bool) {
	st.mutex.Lock()
	v = st.config.LetsEncryptEnabled
	st.mutex.Unlock()
	return
}

// SetLetsEncryptEnabled safely sets the Configuration value for state's 'LetsEncryptEnabled' field
func (st *ConfigState) SetLetsEncryptEnabled(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LetsEncryptEnabled = v
	st.reloadToViper()
}

// LetsEncryptEnabledFlag returns the flag name for the 'LetsEncryptEnabled' field
func LetsEncryptEnabledFlag() string { return "letsencrypt-enabled" }

// GetLetsEncryptEnabled safely fetches the value for global configuration 'LetsEncryptEnabled' field
func GetLetsEncryptEnabled() bool { return global.GetLetsEncryptEnabled() }

// SetLetsEncryptEnabled safely sets the value for global configuration 'LetsEncryptEnabled' field
func SetLetsEncryptEnabled(v bool) { global.SetLetsEncryptEnabled(v) }

// GetLetsEncryptPort safely fetches the Configuration value for state's 'LetsEncryptPort' field
func (st *ConfigState) GetLetsEncryptPort() (v int) {
	st.mutex.Lock()
	v = st.config.LetsEncryptPort
	st.mutex.Unlock()
	return
}

// SetLetsEncryptPort safely sets the Configuration value for state's 'LetsEncryptPort' field
func (st *ConfigState) SetLetsEncryptPort(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LetsEncryptPort = v
	st.reloadToViper()
}

// LetsEncryptPortFlag returns the flag name for the 'LetsEncryptPort' field
func LetsEncryptPortFlag() string { return "letsencrypt-port" }

// GetLetsEncryptPort safely fetches the value for global configuration 'LetsEncryptPort' field
func GetLetsEncryptPort() int { return global.GetLetsEncryptPort() }

// SetLetsEncryptPort safely sets the value for global configuration 'LetsEncryptPort' field
func SetLetsEncryptPort(v int) { global.SetLetsEncryptPort(v) }

// GetLetsEncryptCertDir safely fetches the Configuration value for state's 'LetsEncryptCertDir' field
func (st *ConfigState) GetLetsEncryptCertDir() (v string) {
	st.mutex.Lock()
	v = st.config.LetsEncryptCertDir
	st.mutex.Unlock()
	return
}

// SetLetsEncryptCertDir safely sets the Configuration value for state's 'LetsEncryptCertDir' field
func (st *ConfigState) SetLetsEncryptCertDir(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LetsEncryptCertDir = v
	st.reloadToViper()
}

// LetsEncryptCertDirFlag returns the flag name for the 'LetsEncryptCertDir' field
func LetsEncryptCertDirFlag() string { return "letsencrypt-cert-dir" }

// GetLetsEncryptCertDir safely fetches the value for global configuration 'LetsEncryptCertDir' field
func GetLetsEncryptCertDir() string { return global.GetLetsEncryptCertDir() }

// SetLetsEncryptCertDir safely sets the value for global configuration 'LetsEncryptCertDir' field
func SetLetsEncryptCertDir(v string) { global.SetLetsEncryptCertDir(v) }

// GetLetsEncryptEmailAddress safely fetches the Configuration value for state's 'LetsEncryptEmailAddress' field
func (st *ConfigState) GetLetsEncryptEmailAddress() (v string) {
	st.mutex.Lock()
	v = st.config.LetsEncryptEmailAddress
	st.mutex.Unlock()
	return
}

// SetLetsEncryptEmailAddress safely sets the Configuration value for state's 'LetsEncryptEmailAddress' field
func (st *ConfigState) SetLetsEncryptEmailAddress(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LetsEncryptEmailAddress = v
	st.reloadToViper()
}

// LetsEncryptEmailAddressFlag returns the flag name for the 'LetsEncryptEmailAddress' field
func LetsEncryptEmailAddressFlag() string { return "letsencrypt-email-address" }

// GetLetsEncryptEmailAddress safely fetches the value for global configuration 'LetsEncryptEmailAddress' field
func GetLetsEncryptEmailAddress() string { return global.GetLetsEncryptEmailAddress() }

// SetLetsEncryptEmailAddress safely sets the value for global configuration 'LetsEncryptEmailAddress' field
func SetLetsEncryptEmailAddress(v string) { global.SetLetsEncryptEmailAddress(v) }

// GetOIDCEnabled safely fetches the Configuration value for state's 'OIDCEnabled' field
func (st *ConfigState) GetOIDCEnabled() (v bool) {
	st.mutex.Lock()
	v = st.config.OIDCEnabled
	st.mutex.Unlock()
	return
}

// SetOIDCEnabled safely sets the Configuration value for state's 'OIDCEnabled' field
func (st *ConfigState) SetOIDCEnabled(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCEnabled = v
	st.reloadToViper()
}

// OIDCEnabledFlag returns the flag name for the 'OIDCEnabled' field
func OIDCEnabledFlag() string { return "oidc-enabled" }

// GetOIDCEnabled safely fetches the value for global configuration 'OIDCEnabled' field
func GetOIDCEnabled() bool { return global.GetOIDCEnabled() }

// SetOIDCEnabled safely sets the value for global configuration 'OIDCEnabled' field
func SetOIDCEnabled(v bool) { global.SetOIDCEnabled(v) }

// GetOIDCIdpName safely fetches the Configuration value for state's 'OIDCIdpName' field
func (st *ConfigState) GetOIDCIdpName() (v string) {
	st.mutex.Lock()
	v = st.config.OIDCIdpName
	st.mutex.Unlock()
	return
}

// SetOIDCIdpName safely sets the Configuration value for state's 'OIDCIdpName' field
func (st *ConfigState) SetOIDCIdpName(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCIdpName = v
	st.reloadToViper()
}

// OIDCIdpNameFlag returns the flag name for the 'OIDCIdpName' field
func OIDCIdpNameFlag() string { return "oidc-idp-name" }

// GetOIDCIdpName safely fetches the value for global configuration 'OIDCIdpName' field
func GetOIDCIdpName() string { return global.GetOIDCIdpName() }

// SetOIDCIdpName safely sets the value for global configuration 'OIDCIdpName' field
func SetOIDCIdpName(v string) { global.SetOIDCIdpName(v) }

// GetOIDCSkipVerification safely fetches the Configuration value for state's 'OIDCSkipVerification' field
func (st *ConfigState) GetOIDCSkipVerification() (v bool) {
	st.mutex.Lock()
	v = st.config.OIDCSkipVerification
	st.mutex.Unlock()
	return
}

// SetOIDCSkipVerification safely sets the Configuration value for state's 'OIDCSkipVerification' field
func (st *ConfigState) SetOIDCSkipVerification(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCSkipVerification = v
	st.reloadToViper()
}

// OIDCSkipVerificationFlag returns the flag name for the 'OIDCSkipVerification' field
func OIDCSkipVerificationFlag() string { return "oidc-skip-verification" }

// GetOIDCSkipVerification safely fetches the value for global configuration 'OIDCSkipVerification' field
func GetOIDCSkipVerification() bool { return global.GetOIDCSkipVerification() }

// SetOIDCSkipVerification safely sets the value for global configuration 'OIDCSkipVerification' field
func SetOIDCSkipVerification(v bool) { global.SetOIDCSkipVerification(v) }

// GetOIDCIssuer safely fetches the Configuration value for state's 'OIDCIssuer' field
func (st *ConfigState) GetOIDCIssuer() (v string) {
	st.mutex.Lock()
	v = st.config.OIDCIssuer
	st.mutex.Unlock()
	return
}

// SetOIDCIssuer safely sets the Configuration value for state's 'OIDCIssuer' field
func (st *ConfigState) SetOIDCIssuer(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCIssuer = v
	st.reloadToViper()
}

// OIDCIssuerFlag returns the flag name for the 'OIDCIssuer' field
func OIDCIssuerFlag() string { return "oidc-issuer" }

// GetOIDCIssuer safely fetches the value for global configuration 'OIDCIssuer' field
func GetOIDCIssuer() string { return global.GetOIDCIssuer() }

// SetOIDCIssuer safely sets the value for global configuration 'OIDCIssuer' field
func SetOIDCIssuer(v string) { global.SetOIDCIssuer(v) }

// GetOIDCClientID safely fetches the Configuration value for state's 'OIDCClientID' field
func (st *ConfigState) GetOIDCClientID() (v string) {
	st.mutex.Lock()
	v = st.config.OIDCClientID
	st.mutex.Unlock()
	return
}

// SetOIDCClientID safely sets the Configuration value for state's 'OIDCClientID' field
func (st *ConfigState) SetOIDCClientID(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCClientID = v
	st.reloadToViper()
}

// OIDCClientIDFlag returns the flag name for the 'OIDCClientID' field
func OIDCClientIDFlag() string { return "oidc-client-id" }

// GetOIDCClientID safely fetches the value for global configuration 'OIDCClientID' field
func GetOIDCClientID() string { return global.GetOIDCClientID() }

// SetOIDCClientID safely sets the value for global configuration 'OIDCClientID' field
func SetOIDCClientID(v string) { global.SetOIDCClientID(v) }

// GetOIDCClientSecret safely fetches the Configuration value for state's 'OIDCClientSecret' field
func (st *ConfigState) GetOIDCClientSecret() (v string) {
	st.mutex.Lock()
	v = st.config.OIDCClientSecret
	st.mutex.Unlock()
	return
}

// SetOIDCClientSecret safely sets the Configuration value for state's 'OIDCClientSecret' field
func (st *ConfigState) SetOIDCClientSecret(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCClientSecret = v
	st.reloadToViper()
}

// OIDCClientSecretFlag returns the flag name for the 'OIDCClientSecret' field
func OIDCClientSecretFlag() string { return "oidc-client-secret" }

// GetOIDCClientSecret safely fetches the value for global configuration 'OIDCClientSecret' field
func GetOIDCClientSecret() string { return global.GetOIDCClientSecret() }

// SetOIDCClientSecret safely sets the value for global configuration 'OIDCClientSecret' field
func SetOIDCClientSecret(v string) { global.SetOIDCClientSecret(v) }

// GetOIDCScopes safely fetches the Configuration value for state's 'OIDCScopes' field
func (st *ConfigState) GetOIDCScopes() (v []string) {
	st.mutex.Lock()
	v = st.config.OIDCScopes
	st.mutex.Unlock()
	return
}

// SetOIDCScopes safely sets the Configuration value for state's 'OIDCScopes' field
func (st *ConfigState) SetOIDCScopes(v []string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCScopes = v
	st.reloadToViper()
}

// OIDCScopesFlag returns the flag name for the 'OIDCScopes' field
func OIDCScopesFlag() string { return "oidc-scopes" }

// GetOIDCScopes safely fetches the value for global configuration 'OIDCScopes' field
func GetOIDCScopes() []string { return global.GetOIDCScopes() }

// SetOIDCScopes safely sets the value for global configuration 'OIDCScopes' field
func SetOIDCScopes(v []string) { global.SetOIDCScopes(v) }

// GetSMTPHost safely fetches the Configuration value for state's 'SMTPHost' field
func (st *ConfigState) GetSMTPHost() (v string) {
	st.mutex.Lock()
	v = st.config.SMTPHost
	st.mutex.Unlock()
	return
}

// SetSMTPHost safely sets the Configuration value for state's 'SMTPHost' field
func (st *ConfigState) SetSMTPHost(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SMTPHost = v
	st.reloadToViper()
}

// SMTPHostFlag returns the flag name for the 'SMTPHost' field
func SMTPHostFlag() string { return "smtp-host" }

// GetSMTPHost safely fetches the value for global configuration 'SMTPHost' field
func GetSMTPHost() string { return global.GetSMTPHost() }

// SetSMTPHost safely sets the value for global configuration 'SMTPHost' field
func SetSMTPHost(v string) { global.SetSMTPHost(v) }

// GetSMTPPort safely fetches the Configuration value for state's 'SMTPPort' field
func (st *ConfigState) GetSMTPPort() (v int) {
	st.mutex.Lock()
	v = st.config.SMTPPort
	st.mutex.Unlock()
	return
}

// SetSMTPPort safely sets the Configuration value for state's 'SMTPPort' field
func (st *ConfigState) SetSMTPPort(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SMTPPort = v
	st.reloadToViper()
}

// SMTPPortFlag returns the flag name for the 'SMTPPort' field
func SMTPPortFlag() string { return "smtp-port" }

// GetSMTPPort safely fetches the value for global configuration 'SMTPPort' field
func GetSMTPPort() int { return global.GetSMTPPort() }

// SetSMTPPort safely sets the value for global configuration 'SMTPPort' field
func SetSMTPPort(v int) { global.SetSMTPPort(v) }

// GetSMTPUsername safely fetches the Configuration value for state's 'SMTPUsername' field
func (st *ConfigState) GetSMTPUsername() (v string) {
	st.mutex.Lock()
	v = st.config.SMTPUsername
	st.mutex.Unlock()
	return
}

// SetSMTPUsername safely sets the Configuration value for state's 'SMTPUsername' field
func (st *ConfigState) SetSMTPUsername(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SMTPUsername = v
	st.reloadToViper()
}

// SMTPUsernameFlag returns the flag name for the 'SMTPUsername' field
func SMTPUsernameFlag() string { return "smtp-username" }

// GetSMTPUsername safely fetches the value for global configuration 'SMTPUsername' field
func GetSMTPUsername() string { return global.GetSMTPUsername() }

// SetSMTPUsername safely sets the value for global configuration 'SMTPUsername' field
func SetSMTPUsername(v string) { global.SetSMTPUsername(v) }

// GetSMTPPassword safely fetches the Configuration value for state's 'SMTPPassword' field
func (st *ConfigState) GetSMTPPassword() (v string) {
	st.mutex.Lock()
	v = st.config.SMTPPassword
	st.mutex.Unlock()
	return
}

// SetSMTPPassword safely sets the Configuration value for state's 'SMTPPassword' field
func (st *ConfigState) SetSMTPPassword(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SMTPPassword = v
	st.reloadToViper()
}

// SMTPPasswordFlag returns the flag name for the 'SMTPPassword' field
func SMTPPasswordFlag() string { return "smtp-password" }

// GetSMTPPassword safely fetches the value for global configuration 'SMTPPassword' field
func GetSMTPPassword() string { return global.GetSMTPPassword() }

// SetSMTPPassword safely sets the value for global configuration 'SMTPPassword' field
func SetSMTPPassword(v string) { global.SetSMTPPassword(v) }

// GetSMTPFrom safely fetches the Configuration value for state's 'SMTPFrom' field
func (st *ConfigState) GetSMTPFrom() (v string) {
	st.mutex.Lock()
	v = st.config.SMTPFrom
	st.mutex.Unlock()
	return
}

// SetSMTPFrom safely sets the Configuration value for state's 'SMTPFrom' field
func (st *ConfigState) SetSMTPFrom(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SMTPFrom = v
	st.reloadToViper()
}

// SMTPFromFlag returns the flag name for the 'SMTPFrom' field
func SMTPFromFlag() string { return "smtp-from" }

// GetSMTPFrom safely fetches the value for global configuration 'SMTPFrom' field
func GetSMTPFrom() string { return global.GetSMTPFrom() }

// SetSMTPFrom safely sets the value for global configuration 'SMTPFrom' field
func SetSMTPFrom(v string) { global.SetSMTPFrom(v) }

// GetSyslogEnabled safely fetches the Configuration value for state's 'SyslogEnabled' field
func (st *ConfigState) GetSyslogEnabled() (v bool) {
	st.mutex.Lock()
	v = st.config.SyslogEnabled
	st.mutex.Unlock()
	return
}

// SetSyslogEnabled safely sets the Configuration value for state's 'SyslogEnabled' field
func (st *ConfigState) SetSyslogEnabled(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SyslogEnabled = v
	st.reloadToViper()
}

// SyslogEnabledFlag returns the flag name for the 'SyslogEnabled' field
func SyslogEnabledFlag() string { return "syslog-enabled" }

// GetSyslogEnabled safely fetches the value for global configuration 'SyslogEnabled' field
func GetSyslogEnabled() bool { return global.GetSyslogEnabled() }

// SetSyslogEnabled safely sets the value for global configuration 'SyslogEnabled' field
func SetSyslogEnabled(v bool) { global.SetSyslogEnabled(v) }

// GetSyslogProtocol safely fetches the Configuration value for state's 'SyslogProtocol' field
func (st *ConfigState) GetSyslogProtocol() (v string) {
	st.mutex.Lock()
	v = st.config.SyslogProtocol
	st.mutex.Unlock()
	return
}

// SetSyslogProtocol safely sets the Configuration value for state's 'SyslogProtocol' field
func (st *ConfigState) SetSyslogProtocol(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SyslogProtocol = v
	st.reloadToViper()
}

// SyslogProtocolFlag returns the flag name for the 'SyslogProtocol' field
func SyslogProtocolFlag() string { return "syslog-protocol" }

// GetSyslogProtocol safely fetches the value for global configuration 'SyslogProtocol' field
func GetSyslogProtocol() string { return global.GetSyslogProtocol() }

// SetSyslogProtocol safely sets the value for global configuration 'SyslogProtocol' field
func SetSyslogProtocol(v string) { global.SetSyslogProtocol(v) }

// GetSyslogAddress safely fetches the Configuration value for state's 'SyslogAddress' field
func (st *ConfigState) GetSyslogAddress() (v string) {
	st.mutex.Lock()
	v = st.config.SyslogAddress
	st.mutex.Unlock()
	return
}

// SetSyslogAddress safely sets the Configuration value for state's 'SyslogAddress' field
func (st *ConfigState) SetSyslogAddress(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.SyslogAddress = v
	st.reloadToViper()
}

// SyslogAddressFlag returns the flag name for the 'SyslogAddress' field
func SyslogAddressFlag() string { return "syslog-address" }

// GetSyslogAddress safely fetches the value for global configuration 'SyslogAddress' field
func GetSyslogAddress() string { return global.GetSyslogAddress() }

// SetSyslogAddress safely sets the value for global configuration 'SyslogAddress' field
func SetSyslogAddress(v string) { global.SetSyslogAddress(v) }

// GetAdminAccountUsername safely fetches the Configuration value for state's 'AdminAccountUsername' field
func (st *ConfigState) GetAdminAccountUsername() (v string) {
	st.mutex.Lock()
	v = st.config.AdminAccountUsername
	st.mutex.Unlock()
	return
}

// SetAdminAccountUsername safely sets the Configuration value for state's 'AdminAccountUsername' field
func (st *ConfigState) SetAdminAccountUsername(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdminAccountUsername = v
	st.reloadToViper()
}

// AdminAccountUsernameFlag returns the flag name for the 'AdminAccountUsername' field
func AdminAccountUsernameFlag() string { return "username" }

// GetAdminAccountUsername safely fetches the value for global configuration 'AdminAccountUsername' field
func GetAdminAccountUsername() string { return global.GetAdminAccountUsername() }

// SetAdminAccountUsername safely sets the value for global configuration 'AdminAccountUsername' field
func SetAdminAccountUsername(v string) { global.SetAdminAccountUsername(v) }

// GetAdminAccountEmail safely fetches the Configuration value for state's 'AdminAccountEmail' field
func (st *ConfigState) GetAdminAccountEmail() (v string) {
	st.mutex.Lock()
	v = st.config.AdminAccountEmail
	st.mutex.Unlock()
	return
}

// SetAdminAccountEmail safely sets the Configuration value for state's 'AdminAccountEmail' field
func (st *ConfigState) SetAdminAccountEmail(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdminAccountEmail = v
	st.reloadToViper()
}

// AdminAccountEmailFlag returns the flag name for the 'AdminAccountEmail' field
func AdminAccountEmailFlag() string { return "email" }

// GetAdminAccountEmail safely fetches the value for global configuration 'AdminAccountEmail' field
func GetAdminAccountEmail() string { return global.GetAdminAccountEmail() }

// SetAdminAccountEmail safely sets the value for global configuration 'AdminAccountEmail' field
func SetAdminAccountEmail(v string) { global.SetAdminAccountEmail(v) }

// GetAdminAccountPassword safely fetches the Configuration value for state's 'AdminAccountPassword' field
func (st *ConfigState) GetAdminAccountPassword() (v string) {
	st.mutex.Lock()
	v = st.config.AdminAccountPassword
	st.mutex.Unlock()
	return
}

// SetAdminAccountPassword safely sets the Configuration value for state's 'AdminAccountPassword' field
func (st *ConfigState) SetAdminAccountPassword(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdminAccountPassword = v
	st.reloadToViper()
}

// AdminAccountPasswordFlag returns the flag name for the 'AdminAccountPassword' field
func AdminAccountPasswordFlag() string { return "password" }

// GetAdminAccountPassword safely fetches the value for global configuration 'AdminAccountPassword' field
func GetAdminAccountPassword() string { return global.GetAdminAccountPassword() }

// SetAdminAccountPassword safely sets the value for global configuration 'AdminAccountPassword' field
func SetAdminAccountPassword(v string) { global.SetAdminAccountPassword(v) }

// GetAdminTransPath safely fetches the Configuration value for state's 'AdminTransPath' field
func (st *ConfigState) GetAdminTransPath() (v string) {
	st.mutex.Lock()
	v = st.config.AdminTransPath
	st.mutex.Unlock()
	return
}

// SetAdminTransPath safely sets the Configuration value for state's 'AdminTransPath' field
func (st *ConfigState) SetAdminTransPath(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdminTransPath = v
	st.reloadToViper()
}

// AdminTransPathFlag returns the flag name for the 'AdminTransPath' field
func AdminTransPathFlag() string { return "path" }

// GetAdminTransPath safely fetches the value for global configuration 'AdminTransPath' field
func GetAdminTransPath() string { return global.GetAdminTransPath() }

// SetAdminTransPath safely sets the value for global configuration 'AdminTransPath' field
func SetAdminTransPath(v string) { global.SetAdminTransPath(v) }

// GetAdvancedCookiesSamesite safely fetches the Configuration value for state's 'AdvancedCookiesSamesite' field
func (st *ConfigState) GetAdvancedCookiesSamesite() (v string) {
	st.mutex.Lock()
	v = st.config.AdvancedCookiesSamesite
	st.mutex.Unlock()
	return
}

// SetAdvancedCookiesSamesite safely sets the Configuration value for state's 'AdvancedCookiesSamesite' field
func (st *ConfigState) SetAdvancedCookiesSamesite(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdvancedCookiesSamesite = v
	st.reloadToViper()
}

// AdvancedCookiesSamesiteFlag returns the flag name for the 'AdvancedCookiesSamesite' field
func AdvancedCookiesSamesiteFlag() string { return "advanced-cookies-samesite" }

// GetAdvancedCookiesSamesite safely fetches the value for global configuration 'AdvancedCookiesSamesite' field
func GetAdvancedCookiesSamesite() string { return global.GetAdvancedCookiesSamesite() }

// SetAdvancedCookiesSamesite safely sets the value for global configuration 'AdvancedCookiesSamesite' field
func SetAdvancedCookiesSamesite(v string) { global.SetAdvancedCookiesSamesite(v) }
