// THIS IS A GENERATED FILE, DO NOT EDIT BY HAND
/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"time"

	"codeberg.org/gruf/go-bytesize"
)

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

// GetLandingPageUser safely fetches the Configuration value for state's 'LandingPageUser' field
func (st *ConfigState) GetLandingPageUser() (v string) {
	st.mutex.Lock()
	v = st.config.LandingPageUser
	st.mutex.Unlock()
	return
}

// SetLandingPageUser safely sets the Configuration value for state's 'LandingPageUser' field
func (st *ConfigState) SetLandingPageUser(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.LandingPageUser = v
	st.reloadToViper()
}

// LandingPageUserFlag returns the flag name for the 'LandingPageUser' field
func LandingPageUserFlag() string { return "landing-page-user" }

// GetLandingPageUser safely fetches the value for global configuration 'LandingPageUser' field
func GetLandingPageUser() string { return global.GetLandingPageUser() }

// SetLandingPageUser safely sets the value for global configuration 'LandingPageUser' field
func SetLandingPageUser(v string) { global.SetLandingPageUser(v) }

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

// GetDbMaxOpenConnsMultiplier safely fetches the Configuration value for state's 'DbMaxOpenConnsMultiplier' field
func (st *ConfigState) GetDbMaxOpenConnsMultiplier() (v int) {
	st.mutex.Lock()
	v = st.config.DbMaxOpenConnsMultiplier
	st.mutex.Unlock()
	return
}

// SetDbMaxOpenConnsMultiplier safely sets the Configuration value for state's 'DbMaxOpenConnsMultiplier' field
func (st *ConfigState) SetDbMaxOpenConnsMultiplier(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbMaxOpenConnsMultiplier = v
	st.reloadToViper()
}

// DbMaxOpenConnsMultiplierFlag returns the flag name for the 'DbMaxOpenConnsMultiplier' field
func DbMaxOpenConnsMultiplierFlag() string { return "db-max-open-conns-multiplier" }

// GetDbMaxOpenConnsMultiplier safely fetches the value for global configuration 'DbMaxOpenConnsMultiplier' field
func GetDbMaxOpenConnsMultiplier() int { return global.GetDbMaxOpenConnsMultiplier() }

// SetDbMaxOpenConnsMultiplier safely sets the value for global configuration 'DbMaxOpenConnsMultiplier' field
func SetDbMaxOpenConnsMultiplier(v int) { global.SetDbMaxOpenConnsMultiplier(v) }

// GetDbSqliteJournalMode safely fetches the Configuration value for state's 'DbSqliteJournalMode' field
func (st *ConfigState) GetDbSqliteJournalMode() (v string) {
	st.mutex.Lock()
	v = st.config.DbSqliteJournalMode
	st.mutex.Unlock()
	return
}

// SetDbSqliteJournalMode safely sets the Configuration value for state's 'DbSqliteJournalMode' field
func (st *ConfigState) SetDbSqliteJournalMode(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbSqliteJournalMode = v
	st.reloadToViper()
}

// DbSqliteJournalModeFlag returns the flag name for the 'DbSqliteJournalMode' field
func DbSqliteJournalModeFlag() string { return "db-sqlite-journal-mode" }

// GetDbSqliteJournalMode safely fetches the value for global configuration 'DbSqliteJournalMode' field
func GetDbSqliteJournalMode() string { return global.GetDbSqliteJournalMode() }

// SetDbSqliteJournalMode safely sets the value for global configuration 'DbSqliteJournalMode' field
func SetDbSqliteJournalMode(v string) { global.SetDbSqliteJournalMode(v) }

// GetDbSqliteSynchronous safely fetches the Configuration value for state's 'DbSqliteSynchronous' field
func (st *ConfigState) GetDbSqliteSynchronous() (v string) {
	st.mutex.Lock()
	v = st.config.DbSqliteSynchronous
	st.mutex.Unlock()
	return
}

// SetDbSqliteSynchronous safely sets the Configuration value for state's 'DbSqliteSynchronous' field
func (st *ConfigState) SetDbSqliteSynchronous(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbSqliteSynchronous = v
	st.reloadToViper()
}

// DbSqliteSynchronousFlag returns the flag name for the 'DbSqliteSynchronous' field
func DbSqliteSynchronousFlag() string { return "db-sqlite-synchronous" }

// GetDbSqliteSynchronous safely fetches the value for global configuration 'DbSqliteSynchronous' field
func GetDbSqliteSynchronous() string { return global.GetDbSqliteSynchronous() }

// SetDbSqliteSynchronous safely sets the value for global configuration 'DbSqliteSynchronous' field
func SetDbSqliteSynchronous(v string) { global.SetDbSqliteSynchronous(v) }

// GetDbSqliteCacheSize safely fetches the Configuration value for state's 'DbSqliteCacheSize' field
func (st *ConfigState) GetDbSqliteCacheSize() (v bytesize.Size) {
	st.mutex.Lock()
	v = st.config.DbSqliteCacheSize
	st.mutex.Unlock()
	return
}

// SetDbSqliteCacheSize safely sets the Configuration value for state's 'DbSqliteCacheSize' field
func (st *ConfigState) SetDbSqliteCacheSize(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbSqliteCacheSize = v
	st.reloadToViper()
}

// DbSqliteCacheSizeFlag returns the flag name for the 'DbSqliteCacheSize' field
func DbSqliteCacheSizeFlag() string { return "db-sqlite-cache-size" }

// GetDbSqliteCacheSize safely fetches the value for global configuration 'DbSqliteCacheSize' field
func GetDbSqliteCacheSize() bytesize.Size { return global.GetDbSqliteCacheSize() }

// SetDbSqliteCacheSize safely sets the value for global configuration 'DbSqliteCacheSize' field
func SetDbSqliteCacheSize(v bytesize.Size) { global.SetDbSqliteCacheSize(v) }

// GetDbSqliteBusyTimeout safely fetches the Configuration value for state's 'DbSqliteBusyTimeout' field
func (st *ConfigState) GetDbSqliteBusyTimeout() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.DbSqliteBusyTimeout
	st.mutex.Unlock()
	return
}

// SetDbSqliteBusyTimeout safely sets the Configuration value for state's 'DbSqliteBusyTimeout' field
func (st *ConfigState) SetDbSqliteBusyTimeout(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.DbSqliteBusyTimeout = v
	st.reloadToViper()
}

// DbSqliteBusyTimeoutFlag returns the flag name for the 'DbSqliteBusyTimeout' field
func DbSqliteBusyTimeoutFlag() string { return "db-sqlite-busy-timeout" }

// GetDbSqliteBusyTimeout safely fetches the value for global configuration 'DbSqliteBusyTimeout' field
func GetDbSqliteBusyTimeout() time.Duration { return global.GetDbSqliteBusyTimeout() }

// SetDbSqliteBusyTimeout safely sets the value for global configuration 'DbSqliteBusyTimeout' field
func SetDbSqliteBusyTimeout(v time.Duration) { global.SetDbSqliteBusyTimeout(v) }

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

// GetInstanceExposeSuspendedWeb safely fetches the Configuration value for state's 'InstanceExposeSuspendedWeb' field
func (st *ConfigState) GetInstanceExposeSuspendedWeb() (v bool) {
	st.mutex.Lock()
	v = st.config.InstanceExposeSuspendedWeb
	st.mutex.Unlock()
	return
}

// SetInstanceExposeSuspendedWeb safely sets the Configuration value for state's 'InstanceExposeSuspendedWeb' field
func (st *ConfigState) SetInstanceExposeSuspendedWeb(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceExposeSuspendedWeb = v
	st.reloadToViper()
}

// InstanceExposeSuspendedWebFlag returns the flag name for the 'InstanceExposeSuspendedWeb' field
func InstanceExposeSuspendedWebFlag() string { return "instance-expose-suspended-web" }

// GetInstanceExposeSuspendedWeb safely fetches the value for global configuration 'InstanceExposeSuspendedWeb' field
func GetInstanceExposeSuspendedWeb() bool { return global.GetInstanceExposeSuspendedWeb() }

// SetInstanceExposeSuspendedWeb safely sets the value for global configuration 'InstanceExposeSuspendedWeb' field
func SetInstanceExposeSuspendedWeb(v bool) { global.SetInstanceExposeSuspendedWeb(v) }

// GetInstanceExposePublicTimeline safely fetches the Configuration value for state's 'InstanceExposePublicTimeline' field
func (st *ConfigState) GetInstanceExposePublicTimeline() (v bool) {
	st.mutex.Lock()
	v = st.config.InstanceExposePublicTimeline
	st.mutex.Unlock()
	return
}

// SetInstanceExposePublicTimeline safely sets the Configuration value for state's 'InstanceExposePublicTimeline' field
func (st *ConfigState) SetInstanceExposePublicTimeline(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceExposePublicTimeline = v
	st.reloadToViper()
}

// InstanceExposePublicTimelineFlag returns the flag name for the 'InstanceExposePublicTimeline' field
func InstanceExposePublicTimelineFlag() string { return "instance-expose-public-timeline" }

// GetInstanceExposePublicTimeline safely fetches the value for global configuration 'InstanceExposePublicTimeline' field
func GetInstanceExposePublicTimeline() bool { return global.GetInstanceExposePublicTimeline() }

// SetInstanceExposePublicTimeline safely sets the value for global configuration 'InstanceExposePublicTimeline' field
func SetInstanceExposePublicTimeline(v bool) { global.SetInstanceExposePublicTimeline(v) }

// GetInstanceDeliverToSharedInboxes safely fetches the Configuration value for state's 'InstanceDeliverToSharedInboxes' field
func (st *ConfigState) GetInstanceDeliverToSharedInboxes() (v bool) {
	st.mutex.Lock()
	v = st.config.InstanceDeliverToSharedInboxes
	st.mutex.Unlock()
	return
}

// SetInstanceDeliverToSharedInboxes safely sets the Configuration value for state's 'InstanceDeliverToSharedInboxes' field
func (st *ConfigState) SetInstanceDeliverToSharedInboxes(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.InstanceDeliverToSharedInboxes = v
	st.reloadToViper()
}

// InstanceDeliverToSharedInboxesFlag returns the flag name for the 'InstanceDeliverToSharedInboxes' field
func InstanceDeliverToSharedInboxesFlag() string { return "instance-deliver-to-shared-inboxes" }

// GetInstanceDeliverToSharedInboxes safely fetches the value for global configuration 'InstanceDeliverToSharedInboxes' field
func GetInstanceDeliverToSharedInboxes() bool { return global.GetInstanceDeliverToSharedInboxes() }

// SetInstanceDeliverToSharedInboxes safely sets the value for global configuration 'InstanceDeliverToSharedInboxes' field
func SetInstanceDeliverToSharedInboxes(v bool) { global.SetInstanceDeliverToSharedInboxes(v) }

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

// GetAccountsAllowCustomCSS safely fetches the Configuration value for state's 'AccountsAllowCustomCSS' field
func (st *ConfigState) GetAccountsAllowCustomCSS() (v bool) {
	st.mutex.Lock()
	v = st.config.AccountsAllowCustomCSS
	st.mutex.Unlock()
	return
}

// SetAccountsAllowCustomCSS safely sets the Configuration value for state's 'AccountsAllowCustomCSS' field
func (st *ConfigState) SetAccountsAllowCustomCSS(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AccountsAllowCustomCSS = v
	st.reloadToViper()
}

// AccountsAllowCustomCSSFlag returns the flag name for the 'AccountsAllowCustomCSS' field
func AccountsAllowCustomCSSFlag() string { return "accounts-allow-custom-css" }

// GetAccountsAllowCustomCSS safely fetches the value for global configuration 'AccountsAllowCustomCSS' field
func GetAccountsAllowCustomCSS() bool { return global.GetAccountsAllowCustomCSS() }

// SetAccountsAllowCustomCSS safely sets the value for global configuration 'AccountsAllowCustomCSS' field
func SetAccountsAllowCustomCSS(v bool) { global.SetAccountsAllowCustomCSS(v) }

// GetMediaImageMaxSize safely fetches the Configuration value for state's 'MediaImageMaxSize' field
func (st *ConfigState) GetMediaImageMaxSize() (v bytesize.Size) {
	st.mutex.Lock()
	v = st.config.MediaImageMaxSize
	st.mutex.Unlock()
	return
}

// SetMediaImageMaxSize safely sets the Configuration value for state's 'MediaImageMaxSize' field
func (st *ConfigState) SetMediaImageMaxSize(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaImageMaxSize = v
	st.reloadToViper()
}

// MediaImageMaxSizeFlag returns the flag name for the 'MediaImageMaxSize' field
func MediaImageMaxSizeFlag() string { return "media-image-max-size" }

// GetMediaImageMaxSize safely fetches the value for global configuration 'MediaImageMaxSize' field
func GetMediaImageMaxSize() bytesize.Size { return global.GetMediaImageMaxSize() }

// SetMediaImageMaxSize safely sets the value for global configuration 'MediaImageMaxSize' field
func SetMediaImageMaxSize(v bytesize.Size) { global.SetMediaImageMaxSize(v) }

// GetMediaVideoMaxSize safely fetches the Configuration value for state's 'MediaVideoMaxSize' field
func (st *ConfigState) GetMediaVideoMaxSize() (v bytesize.Size) {
	st.mutex.Lock()
	v = st.config.MediaVideoMaxSize
	st.mutex.Unlock()
	return
}

// SetMediaVideoMaxSize safely sets the Configuration value for state's 'MediaVideoMaxSize' field
func (st *ConfigState) SetMediaVideoMaxSize(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaVideoMaxSize = v
	st.reloadToViper()
}

// MediaVideoMaxSizeFlag returns the flag name for the 'MediaVideoMaxSize' field
func MediaVideoMaxSizeFlag() string { return "media-video-max-size" }

// GetMediaVideoMaxSize safely fetches the value for global configuration 'MediaVideoMaxSize' field
func GetMediaVideoMaxSize() bytesize.Size { return global.GetMediaVideoMaxSize() }

// SetMediaVideoMaxSize safely sets the value for global configuration 'MediaVideoMaxSize' field
func SetMediaVideoMaxSize(v bytesize.Size) { global.SetMediaVideoMaxSize(v) }

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

// GetMediaEmojiLocalMaxSize safely fetches the Configuration value for state's 'MediaEmojiLocalMaxSize' field
func (st *ConfigState) GetMediaEmojiLocalMaxSize() (v bytesize.Size) {
	st.mutex.Lock()
	v = st.config.MediaEmojiLocalMaxSize
	st.mutex.Unlock()
	return
}

// SetMediaEmojiLocalMaxSize safely sets the Configuration value for state's 'MediaEmojiLocalMaxSize' field
func (st *ConfigState) SetMediaEmojiLocalMaxSize(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaEmojiLocalMaxSize = v
	st.reloadToViper()
}

// MediaEmojiLocalMaxSizeFlag returns the flag name for the 'MediaEmojiLocalMaxSize' field
func MediaEmojiLocalMaxSizeFlag() string { return "media-emoji-local-max-size" }

// GetMediaEmojiLocalMaxSize safely fetches the value for global configuration 'MediaEmojiLocalMaxSize' field
func GetMediaEmojiLocalMaxSize() bytesize.Size { return global.GetMediaEmojiLocalMaxSize() }

// SetMediaEmojiLocalMaxSize safely sets the value for global configuration 'MediaEmojiLocalMaxSize' field
func SetMediaEmojiLocalMaxSize(v bytesize.Size) { global.SetMediaEmojiLocalMaxSize(v) }

// GetMediaEmojiRemoteMaxSize safely fetches the Configuration value for state's 'MediaEmojiRemoteMaxSize' field
func (st *ConfigState) GetMediaEmojiRemoteMaxSize() (v bytesize.Size) {
	st.mutex.Lock()
	v = st.config.MediaEmojiRemoteMaxSize
	st.mutex.Unlock()
	return
}

// SetMediaEmojiRemoteMaxSize safely sets the Configuration value for state's 'MediaEmojiRemoteMaxSize' field
func (st *ConfigState) SetMediaEmojiRemoteMaxSize(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaEmojiRemoteMaxSize = v
	st.reloadToViper()
}

// MediaEmojiRemoteMaxSizeFlag returns the flag name for the 'MediaEmojiRemoteMaxSize' field
func MediaEmojiRemoteMaxSizeFlag() string { return "media-emoji-remote-max-size" }

// GetMediaEmojiRemoteMaxSize safely fetches the value for global configuration 'MediaEmojiRemoteMaxSize' field
func GetMediaEmojiRemoteMaxSize() bytesize.Size { return global.GetMediaEmojiRemoteMaxSize() }

// SetMediaEmojiRemoteMaxSize safely sets the value for global configuration 'MediaEmojiRemoteMaxSize' field
func SetMediaEmojiRemoteMaxSize(v bytesize.Size) { global.SetMediaEmojiRemoteMaxSize(v) }

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

// GetStorageS3Endpoint safely fetches the Configuration value for state's 'StorageS3Endpoint' field
func (st *ConfigState) GetStorageS3Endpoint() (v string) {
	st.mutex.Lock()
	v = st.config.StorageS3Endpoint
	st.mutex.Unlock()
	return
}

// SetStorageS3Endpoint safely sets the Configuration value for state's 'StorageS3Endpoint' field
func (st *ConfigState) SetStorageS3Endpoint(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3Endpoint = v
	st.reloadToViper()
}

// StorageS3EndpointFlag returns the flag name for the 'StorageS3Endpoint' field
func StorageS3EndpointFlag() string { return "storage-s3-endpoint" }

// GetStorageS3Endpoint safely fetches the value for global configuration 'StorageS3Endpoint' field
func GetStorageS3Endpoint() string { return global.GetStorageS3Endpoint() }

// SetStorageS3Endpoint safely sets the value for global configuration 'StorageS3Endpoint' field
func SetStorageS3Endpoint(v string) { global.SetStorageS3Endpoint(v) }

// GetStorageS3AccessKey safely fetches the Configuration value for state's 'StorageS3AccessKey' field
func (st *ConfigState) GetStorageS3AccessKey() (v string) {
	st.mutex.Lock()
	v = st.config.StorageS3AccessKey
	st.mutex.Unlock()
	return
}

// SetStorageS3AccessKey safely sets the Configuration value for state's 'StorageS3AccessKey' field
func (st *ConfigState) SetStorageS3AccessKey(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3AccessKey = v
	st.reloadToViper()
}

// StorageS3AccessKeyFlag returns the flag name for the 'StorageS3AccessKey' field
func StorageS3AccessKeyFlag() string { return "storage-s3-access-key" }

// GetStorageS3AccessKey safely fetches the value for global configuration 'StorageS3AccessKey' field
func GetStorageS3AccessKey() string { return global.GetStorageS3AccessKey() }

// SetStorageS3AccessKey safely sets the value for global configuration 'StorageS3AccessKey' field
func SetStorageS3AccessKey(v string) { global.SetStorageS3AccessKey(v) }

// GetStorageS3SecretKey safely fetches the Configuration value for state's 'StorageS3SecretKey' field
func (st *ConfigState) GetStorageS3SecretKey() (v string) {
	st.mutex.Lock()
	v = st.config.StorageS3SecretKey
	st.mutex.Unlock()
	return
}

// SetStorageS3SecretKey safely sets the Configuration value for state's 'StorageS3SecretKey' field
func (st *ConfigState) SetStorageS3SecretKey(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3SecretKey = v
	st.reloadToViper()
}

// StorageS3SecretKeyFlag returns the flag name for the 'StorageS3SecretKey' field
func StorageS3SecretKeyFlag() string { return "storage-s3-secret-key" }

// GetStorageS3SecretKey safely fetches the value for global configuration 'StorageS3SecretKey' field
func GetStorageS3SecretKey() string { return global.GetStorageS3SecretKey() }

// SetStorageS3SecretKey safely sets the value for global configuration 'StorageS3SecretKey' field
func SetStorageS3SecretKey(v string) { global.SetStorageS3SecretKey(v) }

// GetStorageS3UseSSL safely fetches the Configuration value for state's 'StorageS3UseSSL' field
func (st *ConfigState) GetStorageS3UseSSL() (v bool) {
	st.mutex.Lock()
	v = st.config.StorageS3UseSSL
	st.mutex.Unlock()
	return
}

// SetStorageS3UseSSL safely sets the Configuration value for state's 'StorageS3UseSSL' field
func (st *ConfigState) SetStorageS3UseSSL(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3UseSSL = v
	st.reloadToViper()
}

// StorageS3UseSSLFlag returns the flag name for the 'StorageS3UseSSL' field
func StorageS3UseSSLFlag() string { return "storage-s3-use-ssl" }

// GetStorageS3UseSSL safely fetches the value for global configuration 'StorageS3UseSSL' field
func GetStorageS3UseSSL() bool { return global.GetStorageS3UseSSL() }

// SetStorageS3UseSSL safely sets the value for global configuration 'StorageS3UseSSL' field
func SetStorageS3UseSSL(v bool) { global.SetStorageS3UseSSL(v) }

// GetStorageS3BucketName safely fetches the Configuration value for state's 'StorageS3BucketName' field
func (st *ConfigState) GetStorageS3BucketName() (v string) {
	st.mutex.Lock()
	v = st.config.StorageS3BucketName
	st.mutex.Unlock()
	return
}

// SetStorageS3BucketName safely sets the Configuration value for state's 'StorageS3BucketName' field
func (st *ConfigState) SetStorageS3BucketName(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3BucketName = v
	st.reloadToViper()
}

// StorageS3BucketNameFlag returns the flag name for the 'StorageS3BucketName' field
func StorageS3BucketNameFlag() string { return "storage-s3-bucket" }

// GetStorageS3BucketName safely fetches the value for global configuration 'StorageS3BucketName' field
func GetStorageS3BucketName() string { return global.GetStorageS3BucketName() }

// SetStorageS3BucketName safely sets the value for global configuration 'StorageS3BucketName' field
func SetStorageS3BucketName(v string) { global.SetStorageS3BucketName(v) }

// GetStorageS3Proxy safely fetches the Configuration value for state's 'StorageS3Proxy' field
func (st *ConfigState) GetStorageS3Proxy() (v bool) {
	st.mutex.Lock()
	v = st.config.StorageS3Proxy
	st.mutex.Unlock()
	return
}

// SetStorageS3Proxy safely sets the Configuration value for state's 'StorageS3Proxy' field
func (st *ConfigState) SetStorageS3Proxy(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.StorageS3Proxy = v
	st.reloadToViper()
}

// StorageS3ProxyFlag returns the flag name for the 'StorageS3Proxy' field
func StorageS3ProxyFlag() string { return "storage-s3-proxy" }

// GetStorageS3Proxy safely fetches the value for global configuration 'StorageS3Proxy' field
func GetStorageS3Proxy() bool { return global.GetStorageS3Proxy() }

// SetStorageS3Proxy safely sets the value for global configuration 'StorageS3Proxy' field
func SetStorageS3Proxy(v bool) { global.SetStorageS3Proxy(v) }

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

// GetOIDCLinkExisting safely fetches the Configuration value for state's 'OIDCLinkExisting' field
func (st *ConfigState) GetOIDCLinkExisting() (v bool) {
	st.mutex.Lock()
	v = st.config.OIDCLinkExisting
	st.mutex.Unlock()
	return
}

// SetOIDCLinkExisting safely sets the Configuration value for state's 'OIDCLinkExisting' field
func (st *ConfigState) SetOIDCLinkExisting(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCLinkExisting = v
	st.reloadToViper()
}

// OIDCLinkExistingFlag returns the flag name for the 'OIDCLinkExisting' field
func OIDCLinkExistingFlag() string { return "oidc-link-existing" }

// GetOIDCLinkExisting safely fetches the value for global configuration 'OIDCLinkExisting' field
func GetOIDCLinkExisting() bool { return global.GetOIDCLinkExisting() }

// SetOIDCLinkExisting safely sets the value for global configuration 'OIDCLinkExisting' field
func SetOIDCLinkExisting(v bool) { global.SetOIDCLinkExisting(v) }

// GetOIDCAdminGroups safely fetches the Configuration value for state's 'OIDCAdminGroups' field
func (st *ConfigState) GetOIDCAdminGroups() (v []string) {
	st.mutex.Lock()
	v = st.config.OIDCAdminGroups
	st.mutex.Unlock()
	return
}

// SetOIDCAdminGroups safely sets the Configuration value for state's 'OIDCAdminGroups' field
func (st *ConfigState) SetOIDCAdminGroups(v []string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.OIDCAdminGroups = v
	st.reloadToViper()
}

// OIDCAdminGroupsFlag returns the flag name for the 'OIDCAdminGroups' field
func OIDCAdminGroupsFlag() string { return "oidc-admin-groups" }

// GetOIDCAdminGroups safely fetches the value for global configuration 'OIDCAdminGroups' field
func GetOIDCAdminGroups() []string { return global.GetOIDCAdminGroups() }

// SetOIDCAdminGroups safely sets the value for global configuration 'OIDCAdminGroups' field
func SetOIDCAdminGroups(v []string) { global.SetOIDCAdminGroups(v) }

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

// GetAdvancedRateLimitRequests safely fetches the Configuration value for state's 'AdvancedRateLimitRequests' field
func (st *ConfigState) GetAdvancedRateLimitRequests() (v int) {
	st.mutex.Lock()
	v = st.config.AdvancedRateLimitRequests
	st.mutex.Unlock()
	return
}

// SetAdvancedRateLimitRequests safely sets the Configuration value for state's 'AdvancedRateLimitRequests' field
func (st *ConfigState) SetAdvancedRateLimitRequests(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdvancedRateLimitRequests = v
	st.reloadToViper()
}

// AdvancedRateLimitRequestsFlag returns the flag name for the 'AdvancedRateLimitRequests' field
func AdvancedRateLimitRequestsFlag() string { return "advanced-rate-limit-requests" }

// GetAdvancedRateLimitRequests safely fetches the value for global configuration 'AdvancedRateLimitRequests' field
func GetAdvancedRateLimitRequests() int { return global.GetAdvancedRateLimitRequests() }

// SetAdvancedRateLimitRequests safely sets the value for global configuration 'AdvancedRateLimitRequests' field
func SetAdvancedRateLimitRequests(v int) { global.SetAdvancedRateLimitRequests(v) }

// GetAdvancedThrottlingMultiplier safely fetches the Configuration value for state's 'AdvancedThrottlingMultiplier' field
func (st *ConfigState) GetAdvancedThrottlingMultiplier() (v int) {
	st.mutex.Lock()
	v = st.config.AdvancedThrottlingMultiplier
	st.mutex.Unlock()
	return
}

// SetAdvancedThrottlingMultiplier safely sets the Configuration value for state's 'AdvancedThrottlingMultiplier' field
func (st *ConfigState) SetAdvancedThrottlingMultiplier(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdvancedThrottlingMultiplier = v
	st.reloadToViper()
}

// AdvancedThrottlingMultiplierFlag returns the flag name for the 'AdvancedThrottlingMultiplier' field
func AdvancedThrottlingMultiplierFlag() string { return "advanced-throttling-multiplier" }

// GetAdvancedThrottlingMultiplier safely fetches the value for global configuration 'AdvancedThrottlingMultiplier' field
func GetAdvancedThrottlingMultiplier() int { return global.GetAdvancedThrottlingMultiplier() }

// SetAdvancedThrottlingMultiplier safely sets the value for global configuration 'AdvancedThrottlingMultiplier' field
func SetAdvancedThrottlingMultiplier(v int) { global.SetAdvancedThrottlingMultiplier(v) }

// GetAdvancedThrottlingRetryAfter safely fetches the Configuration value for state's 'AdvancedThrottlingRetryAfter' field
func (st *ConfigState) GetAdvancedThrottlingRetryAfter() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.AdvancedThrottlingRetryAfter
	st.mutex.Unlock()
	return
}

// SetAdvancedThrottlingRetryAfter safely sets the Configuration value for state's 'AdvancedThrottlingRetryAfter' field
func (st *ConfigState) SetAdvancedThrottlingRetryAfter(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdvancedThrottlingRetryAfter = v
	st.reloadToViper()
}

// AdvancedThrottlingRetryAfterFlag returns the flag name for the 'AdvancedThrottlingRetryAfter' field
func AdvancedThrottlingRetryAfterFlag() string { return "advanced-throttling-retry-after" }

// GetAdvancedThrottlingRetryAfter safely fetches the value for global configuration 'AdvancedThrottlingRetryAfter' field
func GetAdvancedThrottlingRetryAfter() time.Duration { return global.GetAdvancedThrottlingRetryAfter() }

// SetAdvancedThrottlingRetryAfter safely sets the value for global configuration 'AdvancedThrottlingRetryAfter' field
func SetAdvancedThrottlingRetryAfter(v time.Duration) { global.SetAdvancedThrottlingRetryAfter(v) }

// GetCacheGTSAccountMaxSize safely fetches the Configuration value for state's 'Cache.GTS.AccountMaxSize' field
func (st *ConfigState) GetCacheGTSAccountMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.AccountMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheGTSAccountMaxSize safely sets the Configuration value for state's 'Cache.GTS.AccountMaxSize' field
func (st *ConfigState) SetCacheGTSAccountMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.AccountMaxSize = v
	st.reloadToViper()
}

// CacheGTSAccountMaxSizeFlag returns the flag name for the 'Cache.GTS.AccountMaxSize' field
func CacheGTSAccountMaxSizeFlag() string { return "cache-gts-account-max-size" }

// GetCacheGTSAccountMaxSize safely fetches the value for global configuration 'Cache.GTS.AccountMaxSize' field
func GetCacheGTSAccountMaxSize() int { return global.GetCacheGTSAccountMaxSize() }

// SetCacheGTSAccountMaxSize safely sets the value for global configuration 'Cache.GTS.AccountMaxSize' field
func SetCacheGTSAccountMaxSize(v int) { global.SetCacheGTSAccountMaxSize(v) }

// GetCacheGTSAccountTTL safely fetches the Configuration value for state's 'Cache.GTS.AccountTTL' field
func (st *ConfigState) GetCacheGTSAccountTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.AccountTTL
	st.mutex.Unlock()
	return
}

// SetCacheGTSAccountTTL safely sets the Configuration value for state's 'Cache.GTS.AccountTTL' field
func (st *ConfigState) SetCacheGTSAccountTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.AccountTTL = v
	st.reloadToViper()
}

// CacheGTSAccountTTLFlag returns the flag name for the 'Cache.GTS.AccountTTL' field
func CacheGTSAccountTTLFlag() string { return "cache-gts-account-ttl" }

// GetCacheGTSAccountTTL safely fetches the value for global configuration 'Cache.GTS.AccountTTL' field
func GetCacheGTSAccountTTL() time.Duration { return global.GetCacheGTSAccountTTL() }

// SetCacheGTSAccountTTL safely sets the value for global configuration 'Cache.GTS.AccountTTL' field
func SetCacheGTSAccountTTL(v time.Duration) { global.SetCacheGTSAccountTTL(v) }

// GetCacheGTSAccountSweepFreq safely fetches the Configuration value for state's 'Cache.GTS.AccountSweepFreq' field
func (st *ConfigState) GetCacheGTSAccountSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.AccountSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheGTSAccountSweepFreq safely sets the Configuration value for state's 'Cache.GTS.AccountSweepFreq' field
func (st *ConfigState) SetCacheGTSAccountSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.AccountSweepFreq = v
	st.reloadToViper()
}

// CacheGTSAccountSweepFreqFlag returns the flag name for the 'Cache.GTS.AccountSweepFreq' field
func CacheGTSAccountSweepFreqFlag() string { return "cache-gts-account-sweep-freq" }

// GetCacheGTSAccountSweepFreq safely fetches the value for global configuration 'Cache.GTS.AccountSweepFreq' field
func GetCacheGTSAccountSweepFreq() time.Duration { return global.GetCacheGTSAccountSweepFreq() }

// SetCacheGTSAccountSweepFreq safely sets the value for global configuration 'Cache.GTS.AccountSweepFreq' field
func SetCacheGTSAccountSweepFreq(v time.Duration) { global.SetCacheGTSAccountSweepFreq(v) }

// GetCacheGTSBlockMaxSize safely fetches the Configuration value for state's 'Cache.GTS.BlockMaxSize' field
func (st *ConfigState) GetCacheGTSBlockMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.BlockMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheGTSBlockMaxSize safely sets the Configuration value for state's 'Cache.GTS.BlockMaxSize' field
func (st *ConfigState) SetCacheGTSBlockMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.BlockMaxSize = v
	st.reloadToViper()
}

// CacheGTSBlockMaxSizeFlag returns the flag name for the 'Cache.GTS.BlockMaxSize' field
func CacheGTSBlockMaxSizeFlag() string { return "cache-gts-block-max-size" }

// GetCacheGTSBlockMaxSize safely fetches the value for global configuration 'Cache.GTS.BlockMaxSize' field
func GetCacheGTSBlockMaxSize() int { return global.GetCacheGTSBlockMaxSize() }

// SetCacheGTSBlockMaxSize safely sets the value for global configuration 'Cache.GTS.BlockMaxSize' field
func SetCacheGTSBlockMaxSize(v int) { global.SetCacheGTSBlockMaxSize(v) }

// GetCacheGTSBlockTTL safely fetches the Configuration value for state's 'Cache.GTS.BlockTTL' field
func (st *ConfigState) GetCacheGTSBlockTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.BlockTTL
	st.mutex.Unlock()
	return
}

// SetCacheGTSBlockTTL safely sets the Configuration value for state's 'Cache.GTS.BlockTTL' field
func (st *ConfigState) SetCacheGTSBlockTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.BlockTTL = v
	st.reloadToViper()
}

// CacheGTSBlockTTLFlag returns the flag name for the 'Cache.GTS.BlockTTL' field
func CacheGTSBlockTTLFlag() string { return "cache-gts-block-ttl" }

// GetCacheGTSBlockTTL safely fetches the value for global configuration 'Cache.GTS.BlockTTL' field
func GetCacheGTSBlockTTL() time.Duration { return global.GetCacheGTSBlockTTL() }

// SetCacheGTSBlockTTL safely sets the value for global configuration 'Cache.GTS.BlockTTL' field
func SetCacheGTSBlockTTL(v time.Duration) { global.SetCacheGTSBlockTTL(v) }

// GetCacheGTSBlockSweepFreq safely fetches the Configuration value for state's 'Cache.GTS.BlockSweepFreq' field
func (st *ConfigState) GetCacheGTSBlockSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.BlockSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheGTSBlockSweepFreq safely sets the Configuration value for state's 'Cache.GTS.BlockSweepFreq' field
func (st *ConfigState) SetCacheGTSBlockSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.BlockSweepFreq = v
	st.reloadToViper()
}

// CacheGTSBlockSweepFreqFlag returns the flag name for the 'Cache.GTS.BlockSweepFreq' field
func CacheGTSBlockSweepFreqFlag() string { return "cache-gts-block-sweep-freq" }

// GetCacheGTSBlockSweepFreq safely fetches the value for global configuration 'Cache.GTS.BlockSweepFreq' field
func GetCacheGTSBlockSweepFreq() time.Duration { return global.GetCacheGTSBlockSweepFreq() }

// SetCacheGTSBlockSweepFreq safely sets the value for global configuration 'Cache.GTS.BlockSweepFreq' field
func SetCacheGTSBlockSweepFreq(v time.Duration) { global.SetCacheGTSBlockSweepFreq(v) }

// GetCacheGTSDomainBlockMaxSize safely fetches the Configuration value for state's 'Cache.GTS.DomainBlockMaxSize' field
func (st *ConfigState) GetCacheGTSDomainBlockMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.DomainBlockMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheGTSDomainBlockMaxSize safely sets the Configuration value for state's 'Cache.GTS.DomainBlockMaxSize' field
func (st *ConfigState) SetCacheGTSDomainBlockMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.DomainBlockMaxSize = v
	st.reloadToViper()
}

// CacheGTSDomainBlockMaxSizeFlag returns the flag name for the 'Cache.GTS.DomainBlockMaxSize' field
func CacheGTSDomainBlockMaxSizeFlag() string { return "cache-gts-domain-block-max-size" }

// GetCacheGTSDomainBlockMaxSize safely fetches the value for global configuration 'Cache.GTS.DomainBlockMaxSize' field
func GetCacheGTSDomainBlockMaxSize() int { return global.GetCacheGTSDomainBlockMaxSize() }

// SetCacheGTSDomainBlockMaxSize safely sets the value for global configuration 'Cache.GTS.DomainBlockMaxSize' field
func SetCacheGTSDomainBlockMaxSize(v int) { global.SetCacheGTSDomainBlockMaxSize(v) }

// GetCacheGTSDomainBlockTTL safely fetches the Configuration value for state's 'Cache.GTS.DomainBlockTTL' field
func (st *ConfigState) GetCacheGTSDomainBlockTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.DomainBlockTTL
	st.mutex.Unlock()
	return
}

// SetCacheGTSDomainBlockTTL safely sets the Configuration value for state's 'Cache.GTS.DomainBlockTTL' field
func (st *ConfigState) SetCacheGTSDomainBlockTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.DomainBlockTTL = v
	st.reloadToViper()
}

// CacheGTSDomainBlockTTLFlag returns the flag name for the 'Cache.GTS.DomainBlockTTL' field
func CacheGTSDomainBlockTTLFlag() string { return "cache-gts-domain-block-ttl" }

// GetCacheGTSDomainBlockTTL safely fetches the value for global configuration 'Cache.GTS.DomainBlockTTL' field
func GetCacheGTSDomainBlockTTL() time.Duration { return global.GetCacheGTSDomainBlockTTL() }

// SetCacheGTSDomainBlockTTL safely sets the value for global configuration 'Cache.GTS.DomainBlockTTL' field
func SetCacheGTSDomainBlockTTL(v time.Duration) { global.SetCacheGTSDomainBlockTTL(v) }

// GetCacheGTSDomainBlockSweepFreq safely fetches the Configuration value for state's 'Cache.GTS.DomainBlockSweepFreq' field
func (st *ConfigState) GetCacheGTSDomainBlockSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.DomainBlockSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheGTSDomainBlockSweepFreq safely sets the Configuration value for state's 'Cache.GTS.DomainBlockSweepFreq' field
func (st *ConfigState) SetCacheGTSDomainBlockSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.DomainBlockSweepFreq = v
	st.reloadToViper()
}

// CacheGTSDomainBlockSweepFreqFlag returns the flag name for the 'Cache.GTS.DomainBlockSweepFreq' field
func CacheGTSDomainBlockSweepFreqFlag() string { return "cache-gts-domain-block-sweep-freq" }

// GetCacheGTSDomainBlockSweepFreq safely fetches the value for global configuration 'Cache.GTS.DomainBlockSweepFreq' field
func GetCacheGTSDomainBlockSweepFreq() time.Duration { return global.GetCacheGTSDomainBlockSweepFreq() }

// SetCacheGTSDomainBlockSweepFreq safely sets the value for global configuration 'Cache.GTS.DomainBlockSweepFreq' field
func SetCacheGTSDomainBlockSweepFreq(v time.Duration) { global.SetCacheGTSDomainBlockSweepFreq(v) }

// GetCacheGTSEmojiMaxSize safely fetches the Configuration value for state's 'Cache.GTS.EmojiMaxSize' field
func (st *ConfigState) GetCacheGTSEmojiMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.EmojiMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheGTSEmojiMaxSize safely sets the Configuration value for state's 'Cache.GTS.EmojiMaxSize' field
func (st *ConfigState) SetCacheGTSEmojiMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.EmojiMaxSize = v
	st.reloadToViper()
}

// CacheGTSEmojiMaxSizeFlag returns the flag name for the 'Cache.GTS.EmojiMaxSize' field
func CacheGTSEmojiMaxSizeFlag() string { return "cache-gts-emoji-max-size" }

// GetCacheGTSEmojiMaxSize safely fetches the value for global configuration 'Cache.GTS.EmojiMaxSize' field
func GetCacheGTSEmojiMaxSize() int { return global.GetCacheGTSEmojiMaxSize() }

// SetCacheGTSEmojiMaxSize safely sets the value for global configuration 'Cache.GTS.EmojiMaxSize' field
func SetCacheGTSEmojiMaxSize(v int) { global.SetCacheGTSEmojiMaxSize(v) }

// GetCacheGTSEmojiTTL safely fetches the Configuration value for state's 'Cache.GTS.EmojiTTL' field
func (st *ConfigState) GetCacheGTSEmojiTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.EmojiTTL
	st.mutex.Unlock()
	return
}

// SetCacheGTSEmojiTTL safely sets the Configuration value for state's 'Cache.GTS.EmojiTTL' field
func (st *ConfigState) SetCacheGTSEmojiTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.EmojiTTL = v
	st.reloadToViper()
}

// CacheGTSEmojiTTLFlag returns the flag name for the 'Cache.GTS.EmojiTTL' field
func CacheGTSEmojiTTLFlag() string { return "cache-gts-emoji-ttl" }

// GetCacheGTSEmojiTTL safely fetches the value for global configuration 'Cache.GTS.EmojiTTL' field
func GetCacheGTSEmojiTTL() time.Duration { return global.GetCacheGTSEmojiTTL() }

// SetCacheGTSEmojiTTL safely sets the value for global configuration 'Cache.GTS.EmojiTTL' field
func SetCacheGTSEmojiTTL(v time.Duration) { global.SetCacheGTSEmojiTTL(v) }

// GetCacheGTSEmojiSweepFreq safely fetches the Configuration value for state's 'Cache.GTS.EmojiSweepFreq' field
func (st *ConfigState) GetCacheGTSEmojiSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.EmojiSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheGTSEmojiSweepFreq safely sets the Configuration value for state's 'Cache.GTS.EmojiSweepFreq' field
func (st *ConfigState) SetCacheGTSEmojiSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.EmojiSweepFreq = v
	st.reloadToViper()
}

// CacheGTSEmojiSweepFreqFlag returns the flag name for the 'Cache.GTS.EmojiSweepFreq' field
func CacheGTSEmojiSweepFreqFlag() string { return "cache-gts-emoji-sweep-freq" }

// GetCacheGTSEmojiSweepFreq safely fetches the value for global configuration 'Cache.GTS.EmojiSweepFreq' field
func GetCacheGTSEmojiSweepFreq() time.Duration { return global.GetCacheGTSEmojiSweepFreq() }

// SetCacheGTSEmojiSweepFreq safely sets the value for global configuration 'Cache.GTS.EmojiSweepFreq' field
func SetCacheGTSEmojiSweepFreq(v time.Duration) { global.SetCacheGTSEmojiSweepFreq(v) }

// GetCacheGTSEmojiCategoryMaxSize safely fetches the Configuration value for state's 'Cache.GTS.EmojiCategoryMaxSize' field
func (st *ConfigState) GetCacheGTSEmojiCategoryMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.EmojiCategoryMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheGTSEmojiCategoryMaxSize safely sets the Configuration value for state's 'Cache.GTS.EmojiCategoryMaxSize' field
func (st *ConfigState) SetCacheGTSEmojiCategoryMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.EmojiCategoryMaxSize = v
	st.reloadToViper()
}

// CacheGTSEmojiCategoryMaxSizeFlag returns the flag name for the 'Cache.GTS.EmojiCategoryMaxSize' field
func CacheGTSEmojiCategoryMaxSizeFlag() string { return "cache-gts-emoji-category-max-size" }

// GetCacheGTSEmojiCategoryMaxSize safely fetches the value for global configuration 'Cache.GTS.EmojiCategoryMaxSize' field
func GetCacheGTSEmojiCategoryMaxSize() int { return global.GetCacheGTSEmojiCategoryMaxSize() }

// SetCacheGTSEmojiCategoryMaxSize safely sets the value for global configuration 'Cache.GTS.EmojiCategoryMaxSize' field
func SetCacheGTSEmojiCategoryMaxSize(v int) { global.SetCacheGTSEmojiCategoryMaxSize(v) }

// GetCacheGTSEmojiCategoryTTL safely fetches the Configuration value for state's 'Cache.GTS.EmojiCategoryTTL' field
func (st *ConfigState) GetCacheGTSEmojiCategoryTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.EmojiCategoryTTL
	st.mutex.Unlock()
	return
}

// SetCacheGTSEmojiCategoryTTL safely sets the Configuration value for state's 'Cache.GTS.EmojiCategoryTTL' field
func (st *ConfigState) SetCacheGTSEmojiCategoryTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.EmojiCategoryTTL = v
	st.reloadToViper()
}

// CacheGTSEmojiCategoryTTLFlag returns the flag name for the 'Cache.GTS.EmojiCategoryTTL' field
func CacheGTSEmojiCategoryTTLFlag() string { return "cache-gts-emoji-category-ttl" }

// GetCacheGTSEmojiCategoryTTL safely fetches the value for global configuration 'Cache.GTS.EmojiCategoryTTL' field
func GetCacheGTSEmojiCategoryTTL() time.Duration { return global.GetCacheGTSEmojiCategoryTTL() }

// SetCacheGTSEmojiCategoryTTL safely sets the value for global configuration 'Cache.GTS.EmojiCategoryTTL' field
func SetCacheGTSEmojiCategoryTTL(v time.Duration) { global.SetCacheGTSEmojiCategoryTTL(v) }

// GetCacheGTSEmojiCategorySweepFreq safely fetches the Configuration value for state's 'Cache.GTS.EmojiCategorySweepFreq' field
func (st *ConfigState) GetCacheGTSEmojiCategorySweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.EmojiCategorySweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheGTSEmojiCategorySweepFreq safely sets the Configuration value for state's 'Cache.GTS.EmojiCategorySweepFreq' field
func (st *ConfigState) SetCacheGTSEmojiCategorySweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.EmojiCategorySweepFreq = v
	st.reloadToViper()
}

// CacheGTSEmojiCategorySweepFreqFlag returns the flag name for the 'Cache.GTS.EmojiCategorySweepFreq' field
func CacheGTSEmojiCategorySweepFreqFlag() string { return "cache-gts-emoji-category-sweep-freq" }

// GetCacheGTSEmojiCategorySweepFreq safely fetches the value for global configuration 'Cache.GTS.EmojiCategorySweepFreq' field
func GetCacheGTSEmojiCategorySweepFreq() time.Duration {
	return global.GetCacheGTSEmojiCategorySweepFreq()
}

// SetCacheGTSEmojiCategorySweepFreq safely sets the value for global configuration 'Cache.GTS.EmojiCategorySweepFreq' field
func SetCacheGTSEmojiCategorySweepFreq(v time.Duration) { global.SetCacheGTSEmojiCategorySweepFreq(v) }

// GetCacheGTSMediaMaxSize safely fetches the Configuration value for state's 'Cache.GTS.MediaMaxSize' field
func (st *ConfigState) GetCacheGTSMediaMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.MediaMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheGTSMediaMaxSize safely sets the Configuration value for state's 'Cache.GTS.MediaMaxSize' field
func (st *ConfigState) SetCacheGTSMediaMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.MediaMaxSize = v
	st.reloadToViper()
}

// CacheGTSMediaMaxSizeFlag returns the flag name for the 'Cache.GTS.MediaMaxSize' field
func CacheGTSMediaMaxSizeFlag() string { return "cache-gts-media-max-size" }

// GetCacheGTSMediaMaxSize safely fetches the value for global configuration 'Cache.GTS.MediaMaxSize' field
func GetCacheGTSMediaMaxSize() int { return global.GetCacheGTSMediaMaxSize() }

// SetCacheGTSMediaMaxSize safely sets the value for global configuration 'Cache.GTS.MediaMaxSize' field
func SetCacheGTSMediaMaxSize(v int) { global.SetCacheGTSMediaMaxSize(v) }

// GetCacheGTSMediaTTL safely fetches the Configuration value for state's 'Cache.GTS.MediaTTL' field
func (st *ConfigState) GetCacheGTSMediaTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.MediaTTL
	st.mutex.Unlock()
	return
}

// SetCacheGTSMediaTTL safely sets the Configuration value for state's 'Cache.GTS.MediaTTL' field
func (st *ConfigState) SetCacheGTSMediaTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.MediaTTL = v
	st.reloadToViper()
}

// CacheGTSMediaTTLFlag returns the flag name for the 'Cache.GTS.MediaTTL' field
func CacheGTSMediaTTLFlag() string { return "cache-gts-media-ttl" }

// GetCacheGTSMediaTTL safely fetches the value for global configuration 'Cache.GTS.MediaTTL' field
func GetCacheGTSMediaTTL() time.Duration { return global.GetCacheGTSMediaTTL() }

// SetCacheGTSMediaTTL safely sets the value for global configuration 'Cache.GTS.MediaTTL' field
func SetCacheGTSMediaTTL(v time.Duration) { global.SetCacheGTSMediaTTL(v) }

// GetCacheGTSMediaSweepFreq safely fetches the Configuration value for state's 'Cache.GTS.MediaSweepFreq' field
func (st *ConfigState) GetCacheGTSMediaSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.MediaSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheGTSMediaSweepFreq safely sets the Configuration value for state's 'Cache.GTS.MediaSweepFreq' field
func (st *ConfigState) SetCacheGTSMediaSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.MediaSweepFreq = v
	st.reloadToViper()
}

// CacheGTSMediaSweepFreqFlag returns the flag name for the 'Cache.GTS.MediaSweepFreq' field
func CacheGTSMediaSweepFreqFlag() string { return "cache-gts-media-sweep-freq" }

// GetCacheGTSMediaSweepFreq safely fetches the value for global configuration 'Cache.GTS.MediaSweepFreq' field
func GetCacheGTSMediaSweepFreq() time.Duration { return global.GetCacheGTSMediaSweepFreq() }

// SetCacheGTSMediaSweepFreq safely sets the value for global configuration 'Cache.GTS.MediaSweepFreq' field
func SetCacheGTSMediaSweepFreq(v time.Duration) { global.SetCacheGTSMediaSweepFreq(v) }

// GetCacheGTSMentionMaxSize safely fetches the Configuration value for state's 'Cache.GTS.MentionMaxSize' field
func (st *ConfigState) GetCacheGTSMentionMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.MentionMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheGTSMentionMaxSize safely sets the Configuration value for state's 'Cache.GTS.MentionMaxSize' field
func (st *ConfigState) SetCacheGTSMentionMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.MentionMaxSize = v
	st.reloadToViper()
}

// CacheGTSMentionMaxSizeFlag returns the flag name for the 'Cache.GTS.MentionMaxSize' field
func CacheGTSMentionMaxSizeFlag() string { return "cache-gts-mention-max-size" }

// GetCacheGTSMentionMaxSize safely fetches the value for global configuration 'Cache.GTS.MentionMaxSize' field
func GetCacheGTSMentionMaxSize() int { return global.GetCacheGTSMentionMaxSize() }

// SetCacheGTSMentionMaxSize safely sets the value for global configuration 'Cache.GTS.MentionMaxSize' field
func SetCacheGTSMentionMaxSize(v int) { global.SetCacheGTSMentionMaxSize(v) }

// GetCacheGTSMentionTTL safely fetches the Configuration value for state's 'Cache.GTS.MentionTTL' field
func (st *ConfigState) GetCacheGTSMentionTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.MentionTTL
	st.mutex.Unlock()
	return
}

// SetCacheGTSMentionTTL safely sets the Configuration value for state's 'Cache.GTS.MentionTTL' field
func (st *ConfigState) SetCacheGTSMentionTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.MentionTTL = v
	st.reloadToViper()
}

// CacheGTSMentionTTLFlag returns the flag name for the 'Cache.GTS.MentionTTL' field
func CacheGTSMentionTTLFlag() string { return "cache-gts-mention-ttl" }

// GetCacheGTSMentionTTL safely fetches the value for global configuration 'Cache.GTS.MentionTTL' field
func GetCacheGTSMentionTTL() time.Duration { return global.GetCacheGTSMentionTTL() }

// SetCacheGTSMentionTTL safely sets the value for global configuration 'Cache.GTS.MentionTTL' field
func SetCacheGTSMentionTTL(v time.Duration) { global.SetCacheGTSMentionTTL(v) }

// GetCacheGTSMentionSweepFreq safely fetches the Configuration value for state's 'Cache.GTS.MentionSweepFreq' field
func (st *ConfigState) GetCacheGTSMentionSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.MentionSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheGTSMentionSweepFreq safely sets the Configuration value for state's 'Cache.GTS.MentionSweepFreq' field
func (st *ConfigState) SetCacheGTSMentionSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.MentionSweepFreq = v
	st.reloadToViper()
}

// CacheGTSMentionSweepFreqFlag returns the flag name for the 'Cache.GTS.MentionSweepFreq' field
func CacheGTSMentionSweepFreqFlag() string { return "cache-gts-mention-sweep-freq" }

// GetCacheGTSMentionSweepFreq safely fetches the value for global configuration 'Cache.GTS.MentionSweepFreq' field
func GetCacheGTSMentionSweepFreq() time.Duration { return global.GetCacheGTSMentionSweepFreq() }

// SetCacheGTSMentionSweepFreq safely sets the value for global configuration 'Cache.GTS.MentionSweepFreq' field
func SetCacheGTSMentionSweepFreq(v time.Duration) { global.SetCacheGTSMentionSweepFreq(v) }

// GetCacheGTSNotificationMaxSize safely fetches the Configuration value for state's 'Cache.GTS.NotificationMaxSize' field
func (st *ConfigState) GetCacheGTSNotificationMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.NotificationMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheGTSNotificationMaxSize safely sets the Configuration value for state's 'Cache.GTS.NotificationMaxSize' field
func (st *ConfigState) SetCacheGTSNotificationMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.NotificationMaxSize = v
	st.reloadToViper()
}

// CacheGTSNotificationMaxSizeFlag returns the flag name for the 'Cache.GTS.NotificationMaxSize' field
func CacheGTSNotificationMaxSizeFlag() string { return "cache-gts-notification-max-size" }

// GetCacheGTSNotificationMaxSize safely fetches the value for global configuration 'Cache.GTS.NotificationMaxSize' field
func GetCacheGTSNotificationMaxSize() int { return global.GetCacheGTSNotificationMaxSize() }

// SetCacheGTSNotificationMaxSize safely sets the value for global configuration 'Cache.GTS.NotificationMaxSize' field
func SetCacheGTSNotificationMaxSize(v int) { global.SetCacheGTSNotificationMaxSize(v) }

// GetCacheGTSNotificationTTL safely fetches the Configuration value for state's 'Cache.GTS.NotificationTTL' field
func (st *ConfigState) GetCacheGTSNotificationTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.NotificationTTL
	st.mutex.Unlock()
	return
}

// SetCacheGTSNotificationTTL safely sets the Configuration value for state's 'Cache.GTS.NotificationTTL' field
func (st *ConfigState) SetCacheGTSNotificationTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.NotificationTTL = v
	st.reloadToViper()
}

// CacheGTSNotificationTTLFlag returns the flag name for the 'Cache.GTS.NotificationTTL' field
func CacheGTSNotificationTTLFlag() string { return "cache-gts-notification-ttl" }

// GetCacheGTSNotificationTTL safely fetches the value for global configuration 'Cache.GTS.NotificationTTL' field
func GetCacheGTSNotificationTTL() time.Duration { return global.GetCacheGTSNotificationTTL() }

// SetCacheGTSNotificationTTL safely sets the value for global configuration 'Cache.GTS.NotificationTTL' field
func SetCacheGTSNotificationTTL(v time.Duration) { global.SetCacheGTSNotificationTTL(v) }

// GetCacheGTSNotificationSweepFreq safely fetches the Configuration value for state's 'Cache.GTS.NotificationSweepFreq' field
func (st *ConfigState) GetCacheGTSNotificationSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.NotificationSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheGTSNotificationSweepFreq safely sets the Configuration value for state's 'Cache.GTS.NotificationSweepFreq' field
func (st *ConfigState) SetCacheGTSNotificationSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.NotificationSweepFreq = v
	st.reloadToViper()
}

// CacheGTSNotificationSweepFreqFlag returns the flag name for the 'Cache.GTS.NotificationSweepFreq' field
func CacheGTSNotificationSweepFreqFlag() string { return "cache-gts-notification-sweep-freq" }

// GetCacheGTSNotificationSweepFreq safely fetches the value for global configuration 'Cache.GTS.NotificationSweepFreq' field
func GetCacheGTSNotificationSweepFreq() time.Duration {
	return global.GetCacheGTSNotificationSweepFreq()
}

// SetCacheGTSNotificationSweepFreq safely sets the value for global configuration 'Cache.GTS.NotificationSweepFreq' field
func SetCacheGTSNotificationSweepFreq(v time.Duration) { global.SetCacheGTSNotificationSweepFreq(v) }

// GetCacheGTSReportMaxSize safely fetches the Configuration value for state's 'Cache.GTS.ReportMaxSize' field
func (st *ConfigState) GetCacheGTSReportMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.ReportMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheGTSReportMaxSize safely sets the Configuration value for state's 'Cache.GTS.ReportMaxSize' field
func (st *ConfigState) SetCacheGTSReportMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.ReportMaxSize = v
	st.reloadToViper()
}

// CacheGTSReportMaxSizeFlag returns the flag name for the 'Cache.GTS.ReportMaxSize' field
func CacheGTSReportMaxSizeFlag() string { return "cache-gts-report-max-size" }

// GetCacheGTSReportMaxSize safely fetches the value for global configuration 'Cache.GTS.ReportMaxSize' field
func GetCacheGTSReportMaxSize() int { return global.GetCacheGTSReportMaxSize() }

// SetCacheGTSReportMaxSize safely sets the value for global configuration 'Cache.GTS.ReportMaxSize' field
func SetCacheGTSReportMaxSize(v int) { global.SetCacheGTSReportMaxSize(v) }

// GetCacheGTSReportTTL safely fetches the Configuration value for state's 'Cache.GTS.ReportTTL' field
func (st *ConfigState) GetCacheGTSReportTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.ReportTTL
	st.mutex.Unlock()
	return
}

// SetCacheGTSReportTTL safely sets the Configuration value for state's 'Cache.GTS.ReportTTL' field
func (st *ConfigState) SetCacheGTSReportTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.ReportTTL = v
	st.reloadToViper()
}

// CacheGTSReportTTLFlag returns the flag name for the 'Cache.GTS.ReportTTL' field
func CacheGTSReportTTLFlag() string { return "cache-gts-report-ttl" }

// GetCacheGTSReportTTL safely fetches the value for global configuration 'Cache.GTS.ReportTTL' field
func GetCacheGTSReportTTL() time.Duration { return global.GetCacheGTSReportTTL() }

// SetCacheGTSReportTTL safely sets the value for global configuration 'Cache.GTS.ReportTTL' field
func SetCacheGTSReportTTL(v time.Duration) { global.SetCacheGTSReportTTL(v) }

// GetCacheGTSReportSweepFreq safely fetches the Configuration value for state's 'Cache.GTS.ReportSweepFreq' field
func (st *ConfigState) GetCacheGTSReportSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.ReportSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheGTSReportSweepFreq safely sets the Configuration value for state's 'Cache.GTS.ReportSweepFreq' field
func (st *ConfigState) SetCacheGTSReportSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.ReportSweepFreq = v
	st.reloadToViper()
}

// CacheGTSReportSweepFreqFlag returns the flag name for the 'Cache.GTS.ReportSweepFreq' field
func CacheGTSReportSweepFreqFlag() string { return "cache-gts-report-sweep-freq" }

// GetCacheGTSReportSweepFreq safely fetches the value for global configuration 'Cache.GTS.ReportSweepFreq' field
func GetCacheGTSReportSweepFreq() time.Duration { return global.GetCacheGTSReportSweepFreq() }

// SetCacheGTSReportSweepFreq safely sets the value for global configuration 'Cache.GTS.ReportSweepFreq' field
func SetCacheGTSReportSweepFreq(v time.Duration) { global.SetCacheGTSReportSweepFreq(v) }

// GetCacheGTSStatusMaxSize safely fetches the Configuration value for state's 'Cache.GTS.StatusMaxSize' field
func (st *ConfigState) GetCacheGTSStatusMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.StatusMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheGTSStatusMaxSize safely sets the Configuration value for state's 'Cache.GTS.StatusMaxSize' field
func (st *ConfigState) SetCacheGTSStatusMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.StatusMaxSize = v
	st.reloadToViper()
}

// CacheGTSStatusMaxSizeFlag returns the flag name for the 'Cache.GTS.StatusMaxSize' field
func CacheGTSStatusMaxSizeFlag() string { return "cache-gts-status-max-size" }

// GetCacheGTSStatusMaxSize safely fetches the value for global configuration 'Cache.GTS.StatusMaxSize' field
func GetCacheGTSStatusMaxSize() int { return global.GetCacheGTSStatusMaxSize() }

// SetCacheGTSStatusMaxSize safely sets the value for global configuration 'Cache.GTS.StatusMaxSize' field
func SetCacheGTSStatusMaxSize(v int) { global.SetCacheGTSStatusMaxSize(v) }

// GetCacheGTSStatusTTL safely fetches the Configuration value for state's 'Cache.GTS.StatusTTL' field
func (st *ConfigState) GetCacheGTSStatusTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.StatusTTL
	st.mutex.Unlock()
	return
}

// SetCacheGTSStatusTTL safely sets the Configuration value for state's 'Cache.GTS.StatusTTL' field
func (st *ConfigState) SetCacheGTSStatusTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.StatusTTL = v
	st.reloadToViper()
}

// CacheGTSStatusTTLFlag returns the flag name for the 'Cache.GTS.StatusTTL' field
func CacheGTSStatusTTLFlag() string { return "cache-gts-status-ttl" }

// GetCacheGTSStatusTTL safely fetches the value for global configuration 'Cache.GTS.StatusTTL' field
func GetCacheGTSStatusTTL() time.Duration { return global.GetCacheGTSStatusTTL() }

// SetCacheGTSStatusTTL safely sets the value for global configuration 'Cache.GTS.StatusTTL' field
func SetCacheGTSStatusTTL(v time.Duration) { global.SetCacheGTSStatusTTL(v) }

// GetCacheGTSStatusSweepFreq safely fetches the Configuration value for state's 'Cache.GTS.StatusSweepFreq' field
func (st *ConfigState) GetCacheGTSStatusSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.StatusSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheGTSStatusSweepFreq safely sets the Configuration value for state's 'Cache.GTS.StatusSweepFreq' field
func (st *ConfigState) SetCacheGTSStatusSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.StatusSweepFreq = v
	st.reloadToViper()
}

// CacheGTSStatusSweepFreqFlag returns the flag name for the 'Cache.GTS.StatusSweepFreq' field
func CacheGTSStatusSweepFreqFlag() string { return "cache-gts-status-sweep-freq" }

// GetCacheGTSStatusSweepFreq safely fetches the value for global configuration 'Cache.GTS.StatusSweepFreq' field
func GetCacheGTSStatusSweepFreq() time.Duration { return global.GetCacheGTSStatusSweepFreq() }

// SetCacheGTSStatusSweepFreq safely sets the value for global configuration 'Cache.GTS.StatusSweepFreq' field
func SetCacheGTSStatusSweepFreq(v time.Duration) { global.SetCacheGTSStatusSweepFreq(v) }

// GetCacheGTSTombstoneMaxSize safely fetches the Configuration value for state's 'Cache.GTS.TombstoneMaxSize' field
func (st *ConfigState) GetCacheGTSTombstoneMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.TombstoneMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheGTSTombstoneMaxSize safely sets the Configuration value for state's 'Cache.GTS.TombstoneMaxSize' field
func (st *ConfigState) SetCacheGTSTombstoneMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.TombstoneMaxSize = v
	st.reloadToViper()
}

// CacheGTSTombstoneMaxSizeFlag returns the flag name for the 'Cache.GTS.TombstoneMaxSize' field
func CacheGTSTombstoneMaxSizeFlag() string { return "cache-gts-tombstone-max-size" }

// GetCacheGTSTombstoneMaxSize safely fetches the value for global configuration 'Cache.GTS.TombstoneMaxSize' field
func GetCacheGTSTombstoneMaxSize() int { return global.GetCacheGTSTombstoneMaxSize() }

// SetCacheGTSTombstoneMaxSize safely sets the value for global configuration 'Cache.GTS.TombstoneMaxSize' field
func SetCacheGTSTombstoneMaxSize(v int) { global.SetCacheGTSTombstoneMaxSize(v) }

// GetCacheGTSTombstoneTTL safely fetches the Configuration value for state's 'Cache.GTS.TombstoneTTL' field
func (st *ConfigState) GetCacheGTSTombstoneTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.TombstoneTTL
	st.mutex.Unlock()
	return
}

// SetCacheGTSTombstoneTTL safely sets the Configuration value for state's 'Cache.GTS.TombstoneTTL' field
func (st *ConfigState) SetCacheGTSTombstoneTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.TombstoneTTL = v
	st.reloadToViper()
}

// CacheGTSTombstoneTTLFlag returns the flag name for the 'Cache.GTS.TombstoneTTL' field
func CacheGTSTombstoneTTLFlag() string { return "cache-gts-tombstone-ttl" }

// GetCacheGTSTombstoneTTL safely fetches the value for global configuration 'Cache.GTS.TombstoneTTL' field
func GetCacheGTSTombstoneTTL() time.Duration { return global.GetCacheGTSTombstoneTTL() }

// SetCacheGTSTombstoneTTL safely sets the value for global configuration 'Cache.GTS.TombstoneTTL' field
func SetCacheGTSTombstoneTTL(v time.Duration) { global.SetCacheGTSTombstoneTTL(v) }

// GetCacheGTSTombstoneSweepFreq safely fetches the Configuration value for state's 'Cache.GTS.TombstoneSweepFreq' field
func (st *ConfigState) GetCacheGTSTombstoneSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.TombstoneSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheGTSTombstoneSweepFreq safely sets the Configuration value for state's 'Cache.GTS.TombstoneSweepFreq' field
func (st *ConfigState) SetCacheGTSTombstoneSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.TombstoneSweepFreq = v
	st.reloadToViper()
}

// CacheGTSTombstoneSweepFreqFlag returns the flag name for the 'Cache.GTS.TombstoneSweepFreq' field
func CacheGTSTombstoneSweepFreqFlag() string { return "cache-gts-tombstone-sweep-freq" }

// GetCacheGTSTombstoneSweepFreq safely fetches the value for global configuration 'Cache.GTS.TombstoneSweepFreq' field
func GetCacheGTSTombstoneSweepFreq() time.Duration { return global.GetCacheGTSTombstoneSweepFreq() }

// SetCacheGTSTombstoneSweepFreq safely sets the value for global configuration 'Cache.GTS.TombstoneSweepFreq' field
func SetCacheGTSTombstoneSweepFreq(v time.Duration) { global.SetCacheGTSTombstoneSweepFreq(v) }

// GetCacheGTSUserMaxSize safely fetches the Configuration value for state's 'Cache.GTS.UserMaxSize' field
func (st *ConfigState) GetCacheGTSUserMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.UserMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheGTSUserMaxSize safely sets the Configuration value for state's 'Cache.GTS.UserMaxSize' field
func (st *ConfigState) SetCacheGTSUserMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.UserMaxSize = v
	st.reloadToViper()
}

// CacheGTSUserMaxSizeFlag returns the flag name for the 'Cache.GTS.UserMaxSize' field
func CacheGTSUserMaxSizeFlag() string { return "cache-gts-user-max-size" }

// GetCacheGTSUserMaxSize safely fetches the value for global configuration 'Cache.GTS.UserMaxSize' field
func GetCacheGTSUserMaxSize() int { return global.GetCacheGTSUserMaxSize() }

// SetCacheGTSUserMaxSize safely sets the value for global configuration 'Cache.GTS.UserMaxSize' field
func SetCacheGTSUserMaxSize(v int) { global.SetCacheGTSUserMaxSize(v) }

// GetCacheGTSUserTTL safely fetches the Configuration value for state's 'Cache.GTS.UserTTL' field
func (st *ConfigState) GetCacheGTSUserTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.UserTTL
	st.mutex.Unlock()
	return
}

// SetCacheGTSUserTTL safely sets the Configuration value for state's 'Cache.GTS.UserTTL' field
func (st *ConfigState) SetCacheGTSUserTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.UserTTL = v
	st.reloadToViper()
}

// CacheGTSUserTTLFlag returns the flag name for the 'Cache.GTS.UserTTL' field
func CacheGTSUserTTLFlag() string { return "cache-gts-user-ttl" }

// GetCacheGTSUserTTL safely fetches the value for global configuration 'Cache.GTS.UserTTL' field
func GetCacheGTSUserTTL() time.Duration { return global.GetCacheGTSUserTTL() }

// SetCacheGTSUserTTL safely sets the value for global configuration 'Cache.GTS.UserTTL' field
func SetCacheGTSUserTTL(v time.Duration) { global.SetCacheGTSUserTTL(v) }

// GetCacheGTSUserSweepFreq safely fetches the Configuration value for state's 'Cache.GTS.UserSweepFreq' field
func (st *ConfigState) GetCacheGTSUserSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.GTS.UserSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheGTSUserSweepFreq safely sets the Configuration value for state's 'Cache.GTS.UserSweepFreq' field
func (st *ConfigState) SetCacheGTSUserSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.GTS.UserSweepFreq = v
	st.reloadToViper()
}

// CacheGTSUserSweepFreqFlag returns the flag name for the 'Cache.GTS.UserSweepFreq' field
func CacheGTSUserSweepFreqFlag() string { return "cache-gts-user-sweep-freq" }

// GetCacheGTSUserSweepFreq safely fetches the value for global configuration 'Cache.GTS.UserSweepFreq' field
func GetCacheGTSUserSweepFreq() time.Duration { return global.GetCacheGTSUserSweepFreq() }

// SetCacheGTSUserSweepFreq safely sets the value for global configuration 'Cache.GTS.UserSweepFreq' field
func SetCacheGTSUserSweepFreq(v time.Duration) { global.SetCacheGTSUserSweepFreq(v) }

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

// GetAdminMediaPruneDryRun safely fetches the Configuration value for state's 'AdminMediaPruneDryRun' field
func (st *ConfigState) GetAdminMediaPruneDryRun() (v bool) {
	st.mutex.Lock()
	v = st.config.AdminMediaPruneDryRun
	st.mutex.Unlock()
	return
}

// SetAdminMediaPruneDryRun safely sets the Configuration value for state's 'AdminMediaPruneDryRun' field
func (st *ConfigState) SetAdminMediaPruneDryRun(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdminMediaPruneDryRun = v
	st.reloadToViper()
}

// AdminMediaPruneDryRunFlag returns the flag name for the 'AdminMediaPruneDryRun' field
func AdminMediaPruneDryRunFlag() string { return "dry-run" }

// GetAdminMediaPruneDryRun safely fetches the value for global configuration 'AdminMediaPruneDryRun' field
func GetAdminMediaPruneDryRun() bool { return global.GetAdminMediaPruneDryRun() }

// SetAdminMediaPruneDryRun safely sets the value for global configuration 'AdminMediaPruneDryRun' field
func SetAdminMediaPruneDryRun(v bool) { global.SetAdminMediaPruneDryRun(v) }

// GetRequestIDHeader safely fetches the Configuration value for state's 'RequestIDHeader' field
func (st *ConfigState) GetRequestIDHeader() (v string) {
	st.mutex.Lock()
	v = st.config.RequestIDHeader
	st.mutex.Unlock()
	return
}

// SetRequestIDHeader safely sets the Configuration value for state's 'RequestIDHeader' field
func (st *ConfigState) SetRequestIDHeader(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.RequestIDHeader = v
	st.reloadToViper()
}

// RequestIDHeaderFlag returns the flag name for the 'RequestIDHeader' field
func RequestIDHeaderFlag() string { return "request-id-header" }

// GetRequestIDHeader safely fetches the value for global configuration 'RequestIDHeader' field
func GetRequestIDHeader() string { return global.GetRequestIDHeader() }

// SetRequestIDHeader safely sets the value for global configuration 'RequestIDHeader' field
func SetRequestIDHeader(v string) { global.SetRequestIDHeader(v) }
