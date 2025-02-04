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
	"net/netip"
	"time"

	"codeberg.org/gruf/go-bytesize"
	"github.com/superseriousbusiness/gotosocial/internal/language"
)

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

// LogLevelFlag returns the flag name for the 'LogLevel' field
func LogLevelFlag() string { return "log-level" }

// GetLogLevel safely fetches the value for global configuration 'LogLevel' field
func GetLogLevel() string { return global.GetLogLevel() }

// SetLogLevel safely sets the value for global configuration 'LogLevel' field
func SetLogLevel(v string) { global.SetLogLevel(v) }

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

// LogTimestampFormatFlag returns the flag name for the 'LogTimestampFormat' field
func LogTimestampFormatFlag() string { return "log-timestamp-format" }

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

// LogDbQueriesFlag returns the flag name for the 'LogDbQueries' field
func LogDbQueriesFlag() string { return "log-db-queries" }

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

// LogClientIPFlag returns the flag name for the 'LogClientIP' field
func LogClientIPFlag() string { return "log-client-ip" }

// GetLogClientIP safely fetches the value for global configuration 'LogClientIP' field
func GetLogClientIP() bool { return global.GetLogClientIP() }

// SetLogClientIP safely sets the value for global configuration 'LogClientIP' field
func SetLogClientIP(v bool) { global.SetLogClientIP(v) }

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

// ApplicationNameFlag returns the flag name for the 'ApplicationName' field
func ApplicationNameFlag() string { return "application-name" }

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

// LandingPageUserFlag returns the flag name for the 'LandingPageUser' field
func LandingPageUserFlag() string { return "landing-page-user" }

// GetLandingPageUser safely fetches the value for global configuration 'LandingPageUser' field
func GetLandingPageUser() string { return global.GetLandingPageUser() }

// SetLandingPageUser safely sets the value for global configuration 'LandingPageUser' field
func SetLandingPageUser(v string) { global.SetLandingPageUser(v) }

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

// ConfigPathFlag returns the flag name for the 'ConfigPath' field
func ConfigPathFlag() string { return "config-path" }

// GetConfigPath safely fetches the value for global configuration 'ConfigPath' field
func GetConfigPath() string { return global.GetConfigPath() }

// SetConfigPath safely sets the value for global configuration 'ConfigPath' field
func SetConfigPath(v string) { global.SetConfigPath(v) }

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

// HostFlag returns the flag name for the 'Host' field
func HostFlag() string { return "host" }

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

// AccountDomainFlag returns the flag name for the 'AccountDomain' field
func AccountDomainFlag() string { return "account-domain" }

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

// ProtocolFlag returns the flag name for the 'Protocol' field
func ProtocolFlag() string { return "protocol" }

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

// BindAddressFlag returns the flag name for the 'BindAddress' field
func BindAddressFlag() string { return "bind-address" }

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

// PortFlag returns the flag name for the 'Port' field
func PortFlag() string { return "port" }

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

// TrustedProxiesFlag returns the flag name for the 'TrustedProxies' field
func TrustedProxiesFlag() string { return "trusted-proxies" }

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

// SoftwareVersionFlag returns the flag name for the 'SoftwareVersion' field
func SoftwareVersionFlag() string { return "software-version" }

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

// DbTypeFlag returns the flag name for the 'DbType' field
func DbTypeFlag() string { return "db-type" }

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

// DbAddressFlag returns the flag name for the 'DbAddress' field
func DbAddressFlag() string { return "db-address" }

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

// DbPortFlag returns the flag name for the 'DbPort' field
func DbPortFlag() string { return "db-port" }

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

// DbUserFlag returns the flag name for the 'DbUser' field
func DbUserFlag() string { return "db-user" }

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

// DbPasswordFlag returns the flag name for the 'DbPassword' field
func DbPasswordFlag() string { return "db-password" }

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

// DbDatabaseFlag returns the flag name for the 'DbDatabase' field
func DbDatabaseFlag() string { return "db-database" }

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

// DbTLSModeFlag returns the flag name for the 'DbTLSMode' field
func DbTLSModeFlag() string { return "db-tls-mode" }

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

// DbTLSCACertFlag returns the flag name for the 'DbTLSCACert' field
func DbTLSCACertFlag() string { return "db-tls-ca-cert" }

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

// DbMaxOpenConnsMultiplierFlag returns the flag name for the 'DbMaxOpenConnsMultiplier' field
func DbMaxOpenConnsMultiplierFlag() string { return "db-max-open-conns-multiplier" }

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

// DbSqliteJournalModeFlag returns the flag name for the 'DbSqliteJournalMode' field
func DbSqliteJournalModeFlag() string { return "db-sqlite-journal-mode" }

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

// DbSqliteSynchronousFlag returns the flag name for the 'DbSqliteSynchronous' field
func DbSqliteSynchronousFlag() string { return "db-sqlite-synchronous" }

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

// DbSqliteCacheSizeFlag returns the flag name for the 'DbSqliteCacheSize' field
func DbSqliteCacheSizeFlag() string { return "db-sqlite-cache-size" }

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

// DbSqliteBusyTimeoutFlag returns the flag name for the 'DbSqliteBusyTimeout' field
func DbSqliteBusyTimeoutFlag() string { return "db-sqlite-busy-timeout" }

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

// DbPostgresConnectionStringFlag returns the flag name for the 'DbPostgresConnectionString' field
func DbPostgresConnectionStringFlag() string { return "db-postgres-connection-string" }

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

// WebTemplateBaseDirFlag returns the flag name for the 'WebTemplateBaseDir' field
func WebTemplateBaseDirFlag() string { return "web-template-base-dir" }

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

// WebAssetBaseDirFlag returns the flag name for the 'WebAssetBaseDir' field
func WebAssetBaseDirFlag() string { return "web-asset-base-dir" }

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

// InstanceFederationModeFlag returns the flag name for the 'InstanceFederationMode' field
func InstanceFederationModeFlag() string { return "instance-federation-mode" }

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

// InstanceFederationSpamFilterFlag returns the flag name for the 'InstanceFederationSpamFilter' field
func InstanceFederationSpamFilterFlag() string { return "instance-federation-spam-filter" }

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

// InstanceExposePeersFlag returns the flag name for the 'InstanceExposePeers' field
func InstanceExposePeersFlag() string { return "instance-expose-peers" }

// GetInstanceExposePeers safely fetches the value for global configuration 'InstanceExposePeers' field
func GetInstanceExposePeers() bool { return global.GetInstanceExposePeers() }

// SetInstanceExposePeers safely sets the value for global configuration 'InstanceExposePeers' field
func SetInstanceExposePeers(v bool) { global.SetInstanceExposePeers(v) }

// GetInstanceExposeSuspended safely fetches the Configuration value for state's 'InstanceExposeSuspended' field
func (st *ConfigState) GetInstanceExposeSuspended() (v bool) {
	st.mutex.RLock()
	v = st.config.InstanceExposeSuspended
	st.mutex.RUnlock()
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
	st.mutex.RLock()
	v = st.config.InstanceExposeSuspendedWeb
	st.mutex.RUnlock()
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

// InstanceExposePublicTimelineFlag returns the flag name for the 'InstanceExposePublicTimeline' field
func InstanceExposePublicTimelineFlag() string { return "instance-expose-public-timeline" }

// GetInstanceExposePublicTimeline safely fetches the value for global configuration 'InstanceExposePublicTimeline' field
func GetInstanceExposePublicTimeline() bool { return global.GetInstanceExposePublicTimeline() }

// SetInstanceExposePublicTimeline safely sets the value for global configuration 'InstanceExposePublicTimeline' field
func SetInstanceExposePublicTimeline(v bool) { global.SetInstanceExposePublicTimeline(v) }

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

// InstanceDeliverToSharedInboxesFlag returns the flag name for the 'InstanceDeliverToSharedInboxes' field
func InstanceDeliverToSharedInboxesFlag() string { return "instance-deliver-to-shared-inboxes" }

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

// InstanceInjectMastodonVersionFlag returns the flag name for the 'InstanceInjectMastodonVersion' field
func InstanceInjectMastodonVersionFlag() string { return "instance-inject-mastodon-version" }

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

// InstanceLanguagesFlag returns the flag name for the 'InstanceLanguages' field
func InstanceLanguagesFlag() string { return "instance-languages" }

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

// InstanceSubscriptionsProcessFromFlag returns the flag name for the 'InstanceSubscriptionsProcessFrom' field
func InstanceSubscriptionsProcessFromFlag() string { return "instance-subscriptions-process-from" }

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

// InstanceSubscriptionsProcessEveryFlag returns the flag name for the 'InstanceSubscriptionsProcessEvery' field
func InstanceSubscriptionsProcessEveryFlag() string { return "instance-subscriptions-process-every" }

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

// InstanceStatsModeFlag returns the flag name for the 'InstanceStatsMode' field
func InstanceStatsModeFlag() string { return "instance-stats-mode" }

// GetInstanceStatsMode safely fetches the value for global configuration 'InstanceStatsMode' field
func GetInstanceStatsMode() string { return global.GetInstanceStatsMode() }

// SetInstanceStatsMode safely sets the value for global configuration 'InstanceStatsMode' field
func SetInstanceStatsMode(v string) { global.SetInstanceStatsMode(v) }

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

// AccountsRegistrationOpenFlag returns the flag name for the 'AccountsRegistrationOpen' field
func AccountsRegistrationOpenFlag() string { return "accounts-registration-open" }

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

// AccountsReasonRequiredFlag returns the flag name for the 'AccountsReasonRequired' field
func AccountsReasonRequiredFlag() string { return "accounts-reason-required" }

// GetAccountsReasonRequired safely fetches the value for global configuration 'AccountsReasonRequired' field
func GetAccountsReasonRequired() bool { return global.GetAccountsReasonRequired() }

// SetAccountsReasonRequired safely sets the value for global configuration 'AccountsReasonRequired' field
func SetAccountsReasonRequired(v bool) { global.SetAccountsReasonRequired(v) }

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

// AccountsAllowCustomCSSFlag returns the flag name for the 'AccountsAllowCustomCSS' field
func AccountsAllowCustomCSSFlag() string { return "accounts-allow-custom-css" }

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

// AccountsCustomCSSLengthFlag returns the flag name for the 'AccountsCustomCSSLength' field
func AccountsCustomCSSLengthFlag() string { return "accounts-custom-css-length" }

// GetAccountsCustomCSSLength safely fetches the value for global configuration 'AccountsCustomCSSLength' field
func GetAccountsCustomCSSLength() int { return global.GetAccountsCustomCSSLength() }

// SetAccountsCustomCSSLength safely sets the value for global configuration 'AccountsCustomCSSLength' field
func SetAccountsCustomCSSLength(v int) { global.SetAccountsCustomCSSLength(v) }

// GetMediaDescriptionMinChars safely fetches the Configuration value for state's 'MediaDescriptionMinChars' field
func (st *ConfigState) GetMediaDescriptionMinChars() (v int) {
	st.mutex.RLock()
	v = st.config.MediaDescriptionMinChars
	st.mutex.RUnlock()
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
	st.mutex.RLock()
	v = st.config.MediaDescriptionMaxChars
	st.mutex.RUnlock()
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
	st.mutex.RLock()
	v = st.config.MediaRemoteCacheDays
	st.mutex.RUnlock()
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
	st.mutex.RLock()
	v = st.config.MediaEmojiLocalMaxSize
	st.mutex.RUnlock()
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
	st.mutex.RLock()
	v = st.config.MediaEmojiRemoteMaxSize
	st.mutex.RUnlock()
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

// GetMediaImageSizeHint safely fetches the Configuration value for state's 'MediaImageSizeHint' field
func (st *ConfigState) GetMediaImageSizeHint() (v bytesize.Size) {
	st.mutex.RLock()
	v = st.config.MediaImageSizeHint
	st.mutex.RUnlock()
	return
}

// SetMediaImageSizeHint safely sets the Configuration value for state's 'MediaImageSizeHint' field
func (st *ConfigState) SetMediaImageSizeHint(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaImageSizeHint = v
	st.reloadToViper()
}

// MediaImageSizeHintFlag returns the flag name for the 'MediaImageSizeHint' field
func MediaImageSizeHintFlag() string { return "media-image-size-hint" }

// GetMediaImageSizeHint safely fetches the value for global configuration 'MediaImageSizeHint' field
func GetMediaImageSizeHint() bytesize.Size { return global.GetMediaImageSizeHint() }

// SetMediaImageSizeHint safely sets the value for global configuration 'MediaImageSizeHint' field
func SetMediaImageSizeHint(v bytesize.Size) { global.SetMediaImageSizeHint(v) }

// GetMediaVideoSizeHint safely fetches the Configuration value for state's 'MediaVideoSizeHint' field
func (st *ConfigState) GetMediaVideoSizeHint() (v bytesize.Size) {
	st.mutex.RLock()
	v = st.config.MediaVideoSizeHint
	st.mutex.RUnlock()
	return
}

// SetMediaVideoSizeHint safely sets the Configuration value for state's 'MediaVideoSizeHint' field
func (st *ConfigState) SetMediaVideoSizeHint(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaVideoSizeHint = v
	st.reloadToViper()
}

// MediaVideoSizeHintFlag returns the flag name for the 'MediaVideoSizeHint' field
func MediaVideoSizeHintFlag() string { return "media-video-size-hint" }

// GetMediaVideoSizeHint safely fetches the value for global configuration 'MediaVideoSizeHint' field
func GetMediaVideoSizeHint() bytesize.Size { return global.GetMediaVideoSizeHint() }

// SetMediaVideoSizeHint safely sets the value for global configuration 'MediaVideoSizeHint' field
func SetMediaVideoSizeHint(v bytesize.Size) { global.SetMediaVideoSizeHint(v) }

// GetMediaLocalMaxSize safely fetches the Configuration value for state's 'MediaLocalMaxSize' field
func (st *ConfigState) GetMediaLocalMaxSize() (v bytesize.Size) {
	st.mutex.RLock()
	v = st.config.MediaLocalMaxSize
	st.mutex.RUnlock()
	return
}

// SetMediaLocalMaxSize safely sets the Configuration value for state's 'MediaLocalMaxSize' field
func (st *ConfigState) SetMediaLocalMaxSize(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaLocalMaxSize = v
	st.reloadToViper()
}

// MediaLocalMaxSizeFlag returns the flag name for the 'MediaLocalMaxSize' field
func MediaLocalMaxSizeFlag() string { return "media-local-max-size" }

// GetMediaLocalMaxSize safely fetches the value for global configuration 'MediaLocalMaxSize' field
func GetMediaLocalMaxSize() bytesize.Size { return global.GetMediaLocalMaxSize() }

// SetMediaLocalMaxSize safely sets the value for global configuration 'MediaLocalMaxSize' field
func SetMediaLocalMaxSize(v bytesize.Size) { global.SetMediaLocalMaxSize(v) }

// GetMediaRemoteMaxSize safely fetches the Configuration value for state's 'MediaRemoteMaxSize' field
func (st *ConfigState) GetMediaRemoteMaxSize() (v bytesize.Size) {
	st.mutex.RLock()
	v = st.config.MediaRemoteMaxSize
	st.mutex.RUnlock()
	return
}

// SetMediaRemoteMaxSize safely sets the Configuration value for state's 'MediaRemoteMaxSize' field
func (st *ConfigState) SetMediaRemoteMaxSize(v bytesize.Size) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaRemoteMaxSize = v
	st.reloadToViper()
}

// MediaRemoteMaxSizeFlag returns the flag name for the 'MediaRemoteMaxSize' field
func MediaRemoteMaxSizeFlag() string { return "media-remote-max-size" }

// GetMediaRemoteMaxSize safely fetches the value for global configuration 'MediaRemoteMaxSize' field
func GetMediaRemoteMaxSize() bytesize.Size { return global.GetMediaRemoteMaxSize() }

// SetMediaRemoteMaxSize safely sets the value for global configuration 'MediaRemoteMaxSize' field
func SetMediaRemoteMaxSize(v bytesize.Size) { global.SetMediaRemoteMaxSize(v) }

// GetMediaCleanupFrom safely fetches the Configuration value for state's 'MediaCleanupFrom' field
func (st *ConfigState) GetMediaCleanupFrom() (v string) {
	st.mutex.RLock()
	v = st.config.MediaCleanupFrom
	st.mutex.RUnlock()
	return
}

// SetMediaCleanupFrom safely sets the Configuration value for state's 'MediaCleanupFrom' field
func (st *ConfigState) SetMediaCleanupFrom(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaCleanupFrom = v
	st.reloadToViper()
}

// MediaCleanupFromFlag returns the flag name for the 'MediaCleanupFrom' field
func MediaCleanupFromFlag() string { return "media-cleanup-from" }

// GetMediaCleanupFrom safely fetches the value for global configuration 'MediaCleanupFrom' field
func GetMediaCleanupFrom() string { return global.GetMediaCleanupFrom() }

// SetMediaCleanupFrom safely sets the value for global configuration 'MediaCleanupFrom' field
func SetMediaCleanupFrom(v string) { global.SetMediaCleanupFrom(v) }

// GetMediaCleanupEvery safely fetches the Configuration value for state's 'MediaCleanupEvery' field
func (st *ConfigState) GetMediaCleanupEvery() (v time.Duration) {
	st.mutex.RLock()
	v = st.config.MediaCleanupEvery
	st.mutex.RUnlock()
	return
}

// SetMediaCleanupEvery safely sets the Configuration value for state's 'MediaCleanupEvery' field
func (st *ConfigState) SetMediaCleanupEvery(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaCleanupEvery = v
	st.reloadToViper()
}

// MediaCleanupEveryFlag returns the flag name for the 'MediaCleanupEvery' field
func MediaCleanupEveryFlag() string { return "media-cleanup-every" }

// GetMediaCleanupEvery safely fetches the value for global configuration 'MediaCleanupEvery' field
func GetMediaCleanupEvery() time.Duration { return global.GetMediaCleanupEvery() }

// SetMediaCleanupEvery safely sets the value for global configuration 'MediaCleanupEvery' field
func SetMediaCleanupEvery(v time.Duration) { global.SetMediaCleanupEvery(v) }

// GetMediaFfmpegPoolSize safely fetches the Configuration value for state's 'MediaFfmpegPoolSize' field
func (st *ConfigState) GetMediaFfmpegPoolSize() (v int) {
	st.mutex.RLock()
	v = st.config.MediaFfmpegPoolSize
	st.mutex.RUnlock()
	return
}

// SetMediaFfmpegPoolSize safely sets the Configuration value for state's 'MediaFfmpegPoolSize' field
func (st *ConfigState) SetMediaFfmpegPoolSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MediaFfmpegPoolSize = v
	st.reloadToViper()
}

// MediaFfmpegPoolSizeFlag returns the flag name for the 'MediaFfmpegPoolSize' field
func MediaFfmpegPoolSizeFlag() string { return "media-ffmpeg-pool-size" }

// GetMediaFfmpegPoolSize safely fetches the value for global configuration 'MediaFfmpegPoolSize' field
func GetMediaFfmpegPoolSize() int { return global.GetMediaFfmpegPoolSize() }

// SetMediaFfmpegPoolSize safely sets the value for global configuration 'MediaFfmpegPoolSize' field
func SetMediaFfmpegPoolSize(v int) { global.SetMediaFfmpegPoolSize(v) }

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

// StorageBackendFlag returns the flag name for the 'StorageBackend' field
func StorageBackendFlag() string { return "storage-backend" }

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

// StorageLocalBasePathFlag returns the flag name for the 'StorageLocalBasePath' field
func StorageLocalBasePathFlag() string { return "storage-local-base-path" }

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

// StorageS3EndpointFlag returns the flag name for the 'StorageS3Endpoint' field
func StorageS3EndpointFlag() string { return "storage-s3-endpoint" }

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

// StorageS3AccessKeyFlag returns the flag name for the 'StorageS3AccessKey' field
func StorageS3AccessKeyFlag() string { return "storage-s3-access-key" }

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

// StorageS3SecretKeyFlag returns the flag name for the 'StorageS3SecretKey' field
func StorageS3SecretKeyFlag() string { return "storage-s3-secret-key" }

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

// StorageS3UseSSLFlag returns the flag name for the 'StorageS3UseSSL' field
func StorageS3UseSSLFlag() string { return "storage-s3-use-ssl" }

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

// StorageS3BucketNameFlag returns the flag name for the 'StorageS3BucketName' field
func StorageS3BucketNameFlag() string { return "storage-s3-bucket" }

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

// StorageS3ProxyFlag returns the flag name for the 'StorageS3Proxy' field
func StorageS3ProxyFlag() string { return "storage-s3-proxy" }

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

// StorageS3RedirectURLFlag returns the flag name for the 'StorageS3RedirectURL' field
func StorageS3RedirectURLFlag() string { return "storage-s3-redirect-url" }

// GetStorageS3RedirectURL safely fetches the value for global configuration 'StorageS3RedirectURL' field
func GetStorageS3RedirectURL() string { return global.GetStorageS3RedirectURL() }

// SetStorageS3RedirectURL safely sets the value for global configuration 'StorageS3RedirectURL' field
func SetStorageS3RedirectURL(v string) { global.SetStorageS3RedirectURL(v) }

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

// StatusesMaxCharsFlag returns the flag name for the 'StatusesMaxChars' field
func StatusesMaxCharsFlag() string { return "statuses-max-chars" }

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

// StatusesPollMaxOptionsFlag returns the flag name for the 'StatusesPollMaxOptions' field
func StatusesPollMaxOptionsFlag() string { return "statuses-poll-max-options" }

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

// StatusesPollOptionMaxCharsFlag returns the flag name for the 'StatusesPollOptionMaxChars' field
func StatusesPollOptionMaxCharsFlag() string { return "statuses-poll-option-max-chars" }

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

// StatusesMediaMaxFilesFlag returns the flag name for the 'StatusesMediaMaxFiles' field
func StatusesMediaMaxFilesFlag() string { return "statuses-media-max-files" }

// GetStatusesMediaMaxFiles safely fetches the value for global configuration 'StatusesMediaMaxFiles' field
func GetStatusesMediaMaxFiles() int { return global.GetStatusesMediaMaxFiles() }

// SetStatusesMediaMaxFiles safely sets the value for global configuration 'StatusesMediaMaxFiles' field
func SetStatusesMediaMaxFiles(v int) { global.SetStatusesMediaMaxFiles(v) }

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

// LetsEncryptEnabledFlag returns the flag name for the 'LetsEncryptEnabled' field
func LetsEncryptEnabledFlag() string { return "letsencrypt-enabled" }

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

// LetsEncryptPortFlag returns the flag name for the 'LetsEncryptPort' field
func LetsEncryptPortFlag() string { return "letsencrypt-port" }

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

// LetsEncryptCertDirFlag returns the flag name for the 'LetsEncryptCertDir' field
func LetsEncryptCertDirFlag() string { return "letsencrypt-cert-dir" }

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

// LetsEncryptEmailAddressFlag returns the flag name for the 'LetsEncryptEmailAddress' field
func LetsEncryptEmailAddressFlag() string { return "letsencrypt-email-address" }

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

// TLSCertificateChainFlag returns the flag name for the 'TLSCertificateChain' field
func TLSCertificateChainFlag() string { return "tls-certificate-chain" }

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

// TLSCertificateKeyFlag returns the flag name for the 'TLSCertificateKey' field
func TLSCertificateKeyFlag() string { return "tls-certificate-key" }

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

// OIDCEnabledFlag returns the flag name for the 'OIDCEnabled' field
func OIDCEnabledFlag() string { return "oidc-enabled" }

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

// OIDCIdpNameFlag returns the flag name for the 'OIDCIdpName' field
func OIDCIdpNameFlag() string { return "oidc-idp-name" }

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

// OIDCSkipVerificationFlag returns the flag name for the 'OIDCSkipVerification' field
func OIDCSkipVerificationFlag() string { return "oidc-skip-verification" }

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

// OIDCIssuerFlag returns the flag name for the 'OIDCIssuer' field
func OIDCIssuerFlag() string { return "oidc-issuer" }

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

// OIDCClientIDFlag returns the flag name for the 'OIDCClientID' field
func OIDCClientIDFlag() string { return "oidc-client-id" }

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

// OIDCClientSecretFlag returns the flag name for the 'OIDCClientSecret' field
func OIDCClientSecretFlag() string { return "oidc-client-secret" }

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

// OIDCScopesFlag returns the flag name for the 'OIDCScopes' field
func OIDCScopesFlag() string { return "oidc-scopes" }

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

// OIDCLinkExistingFlag returns the flag name for the 'OIDCLinkExisting' field
func OIDCLinkExistingFlag() string { return "oidc-link-existing" }

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

// OIDCAllowedGroupsFlag returns the flag name for the 'OIDCAllowedGroups' field
func OIDCAllowedGroupsFlag() string { return "oidc-allowed-groups" }

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

// OIDCAdminGroupsFlag returns the flag name for the 'OIDCAdminGroups' field
func OIDCAdminGroupsFlag() string { return "oidc-admin-groups" }

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

// TracingEnabledFlag returns the flag name for the 'TracingEnabled' field
func TracingEnabledFlag() string { return "tracing-enabled" }

// GetTracingEnabled safely fetches the value for global configuration 'TracingEnabled' field
func GetTracingEnabled() bool { return global.GetTracingEnabled() }

// SetTracingEnabled safely sets the value for global configuration 'TracingEnabled' field
func SetTracingEnabled(v bool) { global.SetTracingEnabled(v) }

// GetTracingTransport safely fetches the Configuration value for state's 'TracingTransport' field
func (st *ConfigState) GetTracingTransport() (v string) {
	st.mutex.RLock()
	v = st.config.TracingTransport
	st.mutex.RUnlock()
	return
}

// SetTracingTransport safely sets the Configuration value for state's 'TracingTransport' field
func (st *ConfigState) SetTracingTransport(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.TracingTransport = v
	st.reloadToViper()
}

// TracingTransportFlag returns the flag name for the 'TracingTransport' field
func TracingTransportFlag() string { return "tracing-transport" }

// GetTracingTransport safely fetches the value for global configuration 'TracingTransport' field
func GetTracingTransport() string { return global.GetTracingTransport() }

// SetTracingTransport safely sets the value for global configuration 'TracingTransport' field
func SetTracingTransport(v string) { global.SetTracingTransport(v) }

// GetTracingEndpoint safely fetches the Configuration value for state's 'TracingEndpoint' field
func (st *ConfigState) GetTracingEndpoint() (v string) {
	st.mutex.RLock()
	v = st.config.TracingEndpoint
	st.mutex.RUnlock()
	return
}

// SetTracingEndpoint safely sets the Configuration value for state's 'TracingEndpoint' field
func (st *ConfigState) SetTracingEndpoint(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.TracingEndpoint = v
	st.reloadToViper()
}

// TracingEndpointFlag returns the flag name for the 'TracingEndpoint' field
func TracingEndpointFlag() string { return "tracing-endpoint" }

// GetTracingEndpoint safely fetches the value for global configuration 'TracingEndpoint' field
func GetTracingEndpoint() string { return global.GetTracingEndpoint() }

// SetTracingEndpoint safely sets the value for global configuration 'TracingEndpoint' field
func SetTracingEndpoint(v string) { global.SetTracingEndpoint(v) }

// GetTracingInsecureTransport safely fetches the Configuration value for state's 'TracingInsecureTransport' field
func (st *ConfigState) GetTracingInsecureTransport() (v bool) {
	st.mutex.RLock()
	v = st.config.TracingInsecureTransport
	st.mutex.RUnlock()
	return
}

// SetTracingInsecureTransport safely sets the Configuration value for state's 'TracingInsecureTransport' field
func (st *ConfigState) SetTracingInsecureTransport(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.TracingInsecureTransport = v
	st.reloadToViper()
}

// TracingInsecureTransportFlag returns the flag name for the 'TracingInsecureTransport' field
func TracingInsecureTransportFlag() string { return "tracing-insecure-transport" }

// GetTracingInsecureTransport safely fetches the value for global configuration 'TracingInsecureTransport' field
func GetTracingInsecureTransport() bool { return global.GetTracingInsecureTransport() }

// SetTracingInsecureTransport safely sets the value for global configuration 'TracingInsecureTransport' field
func SetTracingInsecureTransport(v bool) { global.SetTracingInsecureTransport(v) }

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

// MetricsEnabledFlag returns the flag name for the 'MetricsEnabled' field
func MetricsEnabledFlag() string { return "metrics-enabled" }

// GetMetricsEnabled safely fetches the value for global configuration 'MetricsEnabled' field
func GetMetricsEnabled() bool { return global.GetMetricsEnabled() }

// SetMetricsEnabled safely sets the value for global configuration 'MetricsEnabled' field
func SetMetricsEnabled(v bool) { global.SetMetricsEnabled(v) }

// GetMetricsAuthEnabled safely fetches the Configuration value for state's 'MetricsAuthEnabled' field
func (st *ConfigState) GetMetricsAuthEnabled() (v bool) {
	st.mutex.RLock()
	v = st.config.MetricsAuthEnabled
	st.mutex.RUnlock()
	return
}

// SetMetricsAuthEnabled safely sets the Configuration value for state's 'MetricsAuthEnabled' field
func (st *ConfigState) SetMetricsAuthEnabled(v bool) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MetricsAuthEnabled = v
	st.reloadToViper()
}

// MetricsAuthEnabledFlag returns the flag name for the 'MetricsAuthEnabled' field
func MetricsAuthEnabledFlag() string { return "metrics-auth-enabled" }

// GetMetricsAuthEnabled safely fetches the value for global configuration 'MetricsAuthEnabled' field
func GetMetricsAuthEnabled() bool { return global.GetMetricsAuthEnabled() }

// SetMetricsAuthEnabled safely sets the value for global configuration 'MetricsAuthEnabled' field
func SetMetricsAuthEnabled(v bool) { global.SetMetricsAuthEnabled(v) }

// GetMetricsAuthUsername safely fetches the Configuration value for state's 'MetricsAuthUsername' field
func (st *ConfigState) GetMetricsAuthUsername() (v string) {
	st.mutex.RLock()
	v = st.config.MetricsAuthUsername
	st.mutex.RUnlock()
	return
}

// SetMetricsAuthUsername safely sets the Configuration value for state's 'MetricsAuthUsername' field
func (st *ConfigState) SetMetricsAuthUsername(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MetricsAuthUsername = v
	st.reloadToViper()
}

// MetricsAuthUsernameFlag returns the flag name for the 'MetricsAuthUsername' field
func MetricsAuthUsernameFlag() string { return "metrics-auth-username" }

// GetMetricsAuthUsername safely fetches the value for global configuration 'MetricsAuthUsername' field
func GetMetricsAuthUsername() string { return global.GetMetricsAuthUsername() }

// SetMetricsAuthUsername safely sets the value for global configuration 'MetricsAuthUsername' field
func SetMetricsAuthUsername(v string) { global.SetMetricsAuthUsername(v) }

// GetMetricsAuthPassword safely fetches the Configuration value for state's 'MetricsAuthPassword' field
func (st *ConfigState) GetMetricsAuthPassword() (v string) {
	st.mutex.RLock()
	v = st.config.MetricsAuthPassword
	st.mutex.RUnlock()
	return
}

// SetMetricsAuthPassword safely sets the Configuration value for state's 'MetricsAuthPassword' field
func (st *ConfigState) SetMetricsAuthPassword(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.MetricsAuthPassword = v
	st.reloadToViper()
}

// MetricsAuthPasswordFlag returns the flag name for the 'MetricsAuthPassword' field
func MetricsAuthPasswordFlag() string { return "metrics-auth-password" }

// GetMetricsAuthPassword safely fetches the value for global configuration 'MetricsAuthPassword' field
func GetMetricsAuthPassword() string { return global.GetMetricsAuthPassword() }

// SetMetricsAuthPassword safely sets the value for global configuration 'MetricsAuthPassword' field
func SetMetricsAuthPassword(v string) { global.SetMetricsAuthPassword(v) }

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

// SMTPHostFlag returns the flag name for the 'SMTPHost' field
func SMTPHostFlag() string { return "smtp-host" }

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

// SMTPPortFlag returns the flag name for the 'SMTPPort' field
func SMTPPortFlag() string { return "smtp-port" }

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

// SMTPUsernameFlag returns the flag name for the 'SMTPUsername' field
func SMTPUsernameFlag() string { return "smtp-username" }

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

// SMTPPasswordFlag returns the flag name for the 'SMTPPassword' field
func SMTPPasswordFlag() string { return "smtp-password" }

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

// SMTPFromFlag returns the flag name for the 'SMTPFrom' field
func SMTPFromFlag() string { return "smtp-from" }

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

// SMTPDiscloseRecipientsFlag returns the flag name for the 'SMTPDiscloseRecipients' field
func SMTPDiscloseRecipientsFlag() string { return "smtp-disclose-recipients" }

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

// SyslogEnabledFlag returns the flag name for the 'SyslogEnabled' field
func SyslogEnabledFlag() string { return "syslog-enabled" }

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

// SyslogProtocolFlag returns the flag name for the 'SyslogProtocol' field
func SyslogProtocolFlag() string { return "syslog-protocol" }

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

// SyslogAddressFlag returns the flag name for the 'SyslogAddress' field
func SyslogAddressFlag() string { return "syslog-address" }

// GetSyslogAddress safely fetches the value for global configuration 'SyslogAddress' field
func GetSyslogAddress() string { return global.GetSyslogAddress() }

// SetSyslogAddress safely sets the value for global configuration 'SyslogAddress' field
func SetSyslogAddress(v string) { global.SetSyslogAddress(v) }

// GetAdvancedCookiesSamesite safely fetches the Configuration value for state's 'AdvancedCookiesSamesite' field
func (st *ConfigState) GetAdvancedCookiesSamesite() (v string) {
	st.mutex.RLock()
	v = st.config.AdvancedCookiesSamesite
	st.mutex.RUnlock()
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
	st.mutex.RLock()
	v = st.config.AdvancedRateLimitRequests
	st.mutex.RUnlock()
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

// GetAdvancedRateLimitExceptions safely fetches the Configuration value for state's 'AdvancedRateLimitExceptions' field
func (st *ConfigState) GetAdvancedRateLimitExceptions() (v []string) {
	st.mutex.RLock()
	v = st.config.AdvancedRateLimitExceptions
	st.mutex.RUnlock()
	return
}

// SetAdvancedRateLimitExceptions safely sets the Configuration value for state's 'AdvancedRateLimitExceptions' field
func (st *ConfigState) SetAdvancedRateLimitExceptions(v []string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdvancedRateLimitExceptions = v
	st.reloadToViper()
}

// AdvancedRateLimitExceptionsFlag returns the flag name for the 'AdvancedRateLimitExceptions' field
func AdvancedRateLimitExceptionsFlag() string { return "advanced-rate-limit-exceptions" }

// GetAdvancedRateLimitExceptions safely fetches the value for global configuration 'AdvancedRateLimitExceptions' field
func GetAdvancedRateLimitExceptions() []string { return global.GetAdvancedRateLimitExceptions() }

// SetAdvancedRateLimitExceptions safely sets the value for global configuration 'AdvancedRateLimitExceptions' field
func SetAdvancedRateLimitExceptions(v []string) { global.SetAdvancedRateLimitExceptions(v) }

// GetAdvancedRateLimitExceptionsParsed safely fetches the Configuration value for state's 'AdvancedRateLimitExceptionsParsed' field
func (st *ConfigState) GetAdvancedRateLimitExceptionsParsed() (v []netip.Prefix) {
	st.mutex.RLock()
	v = st.config.AdvancedRateLimitExceptionsParsed
	st.mutex.RUnlock()
	return
}

// SetAdvancedRateLimitExceptionsParsed safely sets the Configuration value for state's 'AdvancedRateLimitExceptionsParsed' field
func (st *ConfigState) SetAdvancedRateLimitExceptionsParsed(v []netip.Prefix) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdvancedRateLimitExceptionsParsed = v
	st.reloadToViper()
}

// AdvancedRateLimitExceptionsParsedFlag returns the flag name for the 'AdvancedRateLimitExceptionsParsed' field
func AdvancedRateLimitExceptionsParsedFlag() string { return "advanced-rate-limit-exceptions-parsed" }

// GetAdvancedRateLimitExceptionsParsed safely fetches the value for global configuration 'AdvancedRateLimitExceptionsParsed' field
func GetAdvancedRateLimitExceptionsParsed() []netip.Prefix {
	return global.GetAdvancedRateLimitExceptionsParsed()
}

// SetAdvancedRateLimitExceptionsParsed safely sets the value for global configuration 'AdvancedRateLimitExceptionsParsed' field
func SetAdvancedRateLimitExceptionsParsed(v []netip.Prefix) {
	global.SetAdvancedRateLimitExceptionsParsed(v)
}

// GetAdvancedThrottlingMultiplier safely fetches the Configuration value for state's 'AdvancedThrottlingMultiplier' field
func (st *ConfigState) GetAdvancedThrottlingMultiplier() (v int) {
	st.mutex.RLock()
	v = st.config.AdvancedThrottlingMultiplier
	st.mutex.RUnlock()
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
	st.mutex.RLock()
	v = st.config.AdvancedThrottlingRetryAfter
	st.mutex.RUnlock()
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

// GetAdvancedSenderMultiplier safely fetches the Configuration value for state's 'AdvancedSenderMultiplier' field
func (st *ConfigState) GetAdvancedSenderMultiplier() (v int) {
	st.mutex.RLock()
	v = st.config.AdvancedSenderMultiplier
	st.mutex.RUnlock()
	return
}

// SetAdvancedSenderMultiplier safely sets the Configuration value for state's 'AdvancedSenderMultiplier' field
func (st *ConfigState) SetAdvancedSenderMultiplier(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdvancedSenderMultiplier = v
	st.reloadToViper()
}

// AdvancedSenderMultiplierFlag returns the flag name for the 'AdvancedSenderMultiplier' field
func AdvancedSenderMultiplierFlag() string { return "advanced-sender-multiplier" }

// GetAdvancedSenderMultiplier safely fetches the value for global configuration 'AdvancedSenderMultiplier' field
func GetAdvancedSenderMultiplier() int { return global.GetAdvancedSenderMultiplier() }

// SetAdvancedSenderMultiplier safely sets the value for global configuration 'AdvancedSenderMultiplier' field
func SetAdvancedSenderMultiplier(v int) { global.SetAdvancedSenderMultiplier(v) }

// GetAdvancedCSPExtraURIs safely fetches the Configuration value for state's 'AdvancedCSPExtraURIs' field
func (st *ConfigState) GetAdvancedCSPExtraURIs() (v []string) {
	st.mutex.RLock()
	v = st.config.AdvancedCSPExtraURIs
	st.mutex.RUnlock()
	return
}

// SetAdvancedCSPExtraURIs safely sets the Configuration value for state's 'AdvancedCSPExtraURIs' field
func (st *ConfigState) SetAdvancedCSPExtraURIs(v []string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdvancedCSPExtraURIs = v
	st.reloadToViper()
}

// AdvancedCSPExtraURIsFlag returns the flag name for the 'AdvancedCSPExtraURIs' field
func AdvancedCSPExtraURIsFlag() string { return "advanced-csp-extra-uris" }

// GetAdvancedCSPExtraURIs safely fetches the value for global configuration 'AdvancedCSPExtraURIs' field
func GetAdvancedCSPExtraURIs() []string { return global.GetAdvancedCSPExtraURIs() }

// SetAdvancedCSPExtraURIs safely sets the value for global configuration 'AdvancedCSPExtraURIs' field
func SetAdvancedCSPExtraURIs(v []string) { global.SetAdvancedCSPExtraURIs(v) }

// GetAdvancedHeaderFilterMode safely fetches the Configuration value for state's 'AdvancedHeaderFilterMode' field
func (st *ConfigState) GetAdvancedHeaderFilterMode() (v string) {
	st.mutex.RLock()
	v = st.config.AdvancedHeaderFilterMode
	st.mutex.RUnlock()
	return
}

// SetAdvancedHeaderFilterMode safely sets the Configuration value for state's 'AdvancedHeaderFilterMode' field
func (st *ConfigState) SetAdvancedHeaderFilterMode(v string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.AdvancedHeaderFilterMode = v
	st.reloadToViper()
}

// AdvancedHeaderFilterModeFlag returns the flag name for the 'AdvancedHeaderFilterMode' field
func AdvancedHeaderFilterModeFlag() string { return "advanced-header-filter-mode" }

// GetAdvancedHeaderFilterMode safely fetches the value for global configuration 'AdvancedHeaderFilterMode' field
func GetAdvancedHeaderFilterMode() string { return global.GetAdvancedHeaderFilterMode() }

// SetAdvancedHeaderFilterMode safely sets the value for global configuration 'AdvancedHeaderFilterMode' field
func SetAdvancedHeaderFilterMode(v string) { global.SetAdvancedHeaderFilterMode(v) }

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

// HTTPClientAllowIPsFlag returns the flag name for the 'HTTPClient.AllowIPs' field
func HTTPClientAllowIPsFlag() string { return "httpclient-allow-ips" }

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

// HTTPClientBlockIPsFlag returns the flag name for the 'HTTPClient.BlockIPs' field
func HTTPClientBlockIPsFlag() string { return "httpclient-block-ips" }

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

// HTTPClientTimeoutFlag returns the flag name for the 'HTTPClient.Timeout' field
func HTTPClientTimeoutFlag() string { return "httpclient-timeout" }

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

// HTTPClientTLSInsecureSkipVerifyFlag returns the flag name for the 'HTTPClient.TLSInsecureSkipVerify' field
func HTTPClientTLSInsecureSkipVerifyFlag() string { return "httpclient-tls-insecure-skip-verify" }

// GetHTTPClientTLSInsecureSkipVerify safely fetches the value for global configuration 'HTTPClient.TLSInsecureSkipVerify' field
func GetHTTPClientTLSInsecureSkipVerify() bool { return global.GetHTTPClientTLSInsecureSkipVerify() }

// SetHTTPClientTLSInsecureSkipVerify safely sets the value for global configuration 'HTTPClient.TLSInsecureSkipVerify' field
func SetHTTPClientTLSInsecureSkipVerify(v bool) { global.SetHTTPClientTLSInsecureSkipVerify(v) }

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

// CacheMemoryTargetFlag returns the flag name for the 'Cache.MemoryTarget' field
func CacheMemoryTargetFlag() string { return "cache-memory-target" }

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

// CacheAccountMemRatioFlag returns the flag name for the 'Cache.AccountMemRatio' field
func CacheAccountMemRatioFlag() string { return "cache-account-mem-ratio" }

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

// CacheAccountNoteMemRatioFlag returns the flag name for the 'Cache.AccountNoteMemRatio' field
func CacheAccountNoteMemRatioFlag() string { return "cache-account-note-mem-ratio" }

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

// CacheAccountSettingsMemRatioFlag returns the flag name for the 'Cache.AccountSettingsMemRatio' field
func CacheAccountSettingsMemRatioFlag() string { return "cache-account-settings-mem-ratio" }

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

// CacheAccountStatsMemRatioFlag returns the flag name for the 'Cache.AccountStatsMemRatio' field
func CacheAccountStatsMemRatioFlag() string { return "cache-account-stats-mem-ratio" }

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

// CacheApplicationMemRatioFlag returns the flag name for the 'Cache.ApplicationMemRatio' field
func CacheApplicationMemRatioFlag() string { return "cache-application-mem-ratio" }

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

// CacheBlockMemRatioFlag returns the flag name for the 'Cache.BlockMemRatio' field
func CacheBlockMemRatioFlag() string { return "cache-block-mem-ratio" }

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

// CacheBlockIDsMemRatioFlag returns the flag name for the 'Cache.BlockIDsMemRatio' field
func CacheBlockIDsMemRatioFlag() string { return "cache-block-ids-mem-ratio" }

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

// CacheBoostOfIDsMemRatioFlag returns the flag name for the 'Cache.BoostOfIDsMemRatio' field
func CacheBoostOfIDsMemRatioFlag() string { return "cache-boost-of-ids-mem-ratio" }

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

// CacheClientMemRatioFlag returns the flag name for the 'Cache.ClientMemRatio' field
func CacheClientMemRatioFlag() string { return "cache-client-mem-ratio" }

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

// CacheConversationMemRatioFlag returns the flag name for the 'Cache.ConversationMemRatio' field
func CacheConversationMemRatioFlag() string { return "cache-conversation-mem-ratio" }

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

// CacheConversationLastStatusIDsMemRatioFlag returns the flag name for the 'Cache.ConversationLastStatusIDsMemRatio' field
func CacheConversationLastStatusIDsMemRatioFlag() string {
	return "cache-conversation-last-status-ids-mem-ratio"
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

// CacheDomainPermissionDraftMemRationFlag returns the flag name for the 'Cache.DomainPermissionDraftMemRation' field
func CacheDomainPermissionDraftMemRationFlag() string {
	return "cache-domain-permission-draft-mem-ratio"
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

// CacheDomainPermissionSubscriptionMemRationFlag returns the flag name for the 'Cache.DomainPermissionSubscriptionMemRation' field
func CacheDomainPermissionSubscriptionMemRationFlag() string {
	return "cache-domain-permission-subscription-mem-ratio"
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

// CacheEmojiMemRatioFlag returns the flag name for the 'Cache.EmojiMemRatio' field
func CacheEmojiMemRatioFlag() string { return "cache-emoji-mem-ratio" }

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

// CacheEmojiCategoryMemRatioFlag returns the flag name for the 'Cache.EmojiCategoryMemRatio' field
func CacheEmojiCategoryMemRatioFlag() string { return "cache-emoji-category-mem-ratio" }

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

// CacheFilterMemRatioFlag returns the flag name for the 'Cache.FilterMemRatio' field
func CacheFilterMemRatioFlag() string { return "cache-filter-mem-ratio" }

// GetCacheFilterMemRatio safely fetches the value for global configuration 'Cache.FilterMemRatio' field
func GetCacheFilterMemRatio() float64 { return global.GetCacheFilterMemRatio() }

// SetCacheFilterMemRatio safely sets the value for global configuration 'Cache.FilterMemRatio' field
func SetCacheFilterMemRatio(v float64) { global.SetCacheFilterMemRatio(v) }

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

// CacheFilterKeywordMemRatioFlag returns the flag name for the 'Cache.FilterKeywordMemRatio' field
func CacheFilterKeywordMemRatioFlag() string { return "cache-filter-keyword-mem-ratio" }

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

// CacheFilterStatusMemRatioFlag returns the flag name for the 'Cache.FilterStatusMemRatio' field
func CacheFilterStatusMemRatioFlag() string { return "cache-filter-status-mem-ratio" }

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

// CacheFollowMemRatioFlag returns the flag name for the 'Cache.FollowMemRatio' field
func CacheFollowMemRatioFlag() string { return "cache-follow-mem-ratio" }

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

// CacheFollowIDsMemRatioFlag returns the flag name for the 'Cache.FollowIDsMemRatio' field
func CacheFollowIDsMemRatioFlag() string { return "cache-follow-ids-mem-ratio" }

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

// CacheFollowRequestMemRatioFlag returns the flag name for the 'Cache.FollowRequestMemRatio' field
func CacheFollowRequestMemRatioFlag() string { return "cache-follow-request-mem-ratio" }

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

// CacheFollowRequestIDsMemRatioFlag returns the flag name for the 'Cache.FollowRequestIDsMemRatio' field
func CacheFollowRequestIDsMemRatioFlag() string { return "cache-follow-request-ids-mem-ratio" }

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

// CacheFollowingTagIDsMemRatioFlag returns the flag name for the 'Cache.FollowingTagIDsMemRatio' field
func CacheFollowingTagIDsMemRatioFlag() string { return "cache-following-tag-ids-mem-ratio" }

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

// CacheInReplyToIDsMemRatioFlag returns the flag name for the 'Cache.InReplyToIDsMemRatio' field
func CacheInReplyToIDsMemRatioFlag() string { return "cache-in-reply-to-ids-mem-ratio" }

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

// CacheInstanceMemRatioFlag returns the flag name for the 'Cache.InstanceMemRatio' field
func CacheInstanceMemRatioFlag() string { return "cache-instance-mem-ratio" }

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

// CacheInteractionRequestMemRatioFlag returns the flag name for the 'Cache.InteractionRequestMemRatio' field
func CacheInteractionRequestMemRatioFlag() string { return "cache-interaction-request-mem-ratio" }

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

// CacheListMemRatioFlag returns the flag name for the 'Cache.ListMemRatio' field
func CacheListMemRatioFlag() string { return "cache-list-mem-ratio" }

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

// CacheListIDsMemRatioFlag returns the flag name for the 'Cache.ListIDsMemRatio' field
func CacheListIDsMemRatioFlag() string { return "cache-list-ids-mem-ratio" }

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

// CacheListedIDsMemRatioFlag returns the flag name for the 'Cache.ListedIDsMemRatio' field
func CacheListedIDsMemRatioFlag() string { return "cache-listed-ids-mem-ratio" }

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

// CacheMarkerMemRatioFlag returns the flag name for the 'Cache.MarkerMemRatio' field
func CacheMarkerMemRatioFlag() string { return "cache-marker-mem-ratio" }

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

// CacheMediaMemRatioFlag returns the flag name for the 'Cache.MediaMemRatio' field
func CacheMediaMemRatioFlag() string { return "cache-media-mem-ratio" }

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

// CacheMentionMemRatioFlag returns the flag name for the 'Cache.MentionMemRatio' field
func CacheMentionMemRatioFlag() string { return "cache-mention-mem-ratio" }

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

// CacheMoveMemRatioFlag returns the flag name for the 'Cache.MoveMemRatio' field
func CacheMoveMemRatioFlag() string { return "cache-move-mem-ratio" }

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

// CacheNotificationMemRatioFlag returns the flag name for the 'Cache.NotificationMemRatio' field
func CacheNotificationMemRatioFlag() string { return "cache-notification-mem-ratio" }

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

// CachePollMemRatioFlag returns the flag name for the 'Cache.PollMemRatio' field
func CachePollMemRatioFlag() string { return "cache-poll-mem-ratio" }

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

// CachePollVoteMemRatioFlag returns the flag name for the 'Cache.PollVoteMemRatio' field
func CachePollVoteMemRatioFlag() string { return "cache-poll-vote-mem-ratio" }

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

// CachePollVoteIDsMemRatioFlag returns the flag name for the 'Cache.PollVoteIDsMemRatio' field
func CachePollVoteIDsMemRatioFlag() string { return "cache-poll-vote-ids-mem-ratio" }

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

// CacheReportMemRatioFlag returns the flag name for the 'Cache.ReportMemRatio' field
func CacheReportMemRatioFlag() string { return "cache-report-mem-ratio" }

// GetCacheReportMemRatio safely fetches the value for global configuration 'Cache.ReportMemRatio' field
func GetCacheReportMemRatio() float64 { return global.GetCacheReportMemRatio() }

// SetCacheReportMemRatio safely sets the value for global configuration 'Cache.ReportMemRatio' field
func SetCacheReportMemRatio(v float64) { global.SetCacheReportMemRatio(v) }

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

// CacheSinBinStatusMemRatioFlag returns the flag name for the 'Cache.SinBinStatusMemRatio' field
func CacheSinBinStatusMemRatioFlag() string { return "cache-sin-bin-status-mem-ratio" }

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

// CacheStatusMemRatioFlag returns the flag name for the 'Cache.StatusMemRatio' field
func CacheStatusMemRatioFlag() string { return "cache-status-mem-ratio" }

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

// CacheStatusBookmarkMemRatioFlag returns the flag name for the 'Cache.StatusBookmarkMemRatio' field
func CacheStatusBookmarkMemRatioFlag() string { return "cache-status-bookmark-mem-ratio" }

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

// CacheStatusBookmarkIDsMemRatioFlag returns the flag name for the 'Cache.StatusBookmarkIDsMemRatio' field
func CacheStatusBookmarkIDsMemRatioFlag() string { return "cache-status-bookmark-ids-mem-ratio" }

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

// CacheStatusEditMemRatioFlag returns the flag name for the 'Cache.StatusEditMemRatio' field
func CacheStatusEditMemRatioFlag() string { return "cache-status-edit-mem-ratio" }

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

// CacheStatusFaveMemRatioFlag returns the flag name for the 'Cache.StatusFaveMemRatio' field
func CacheStatusFaveMemRatioFlag() string { return "cache-status-fave-mem-ratio" }

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

// CacheStatusFaveIDsMemRatioFlag returns the flag name for the 'Cache.StatusFaveIDsMemRatio' field
func CacheStatusFaveIDsMemRatioFlag() string { return "cache-status-fave-ids-mem-ratio" }

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

// CacheTagMemRatioFlag returns the flag name for the 'Cache.TagMemRatio' field
func CacheTagMemRatioFlag() string { return "cache-tag-mem-ratio" }

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

// CacheThreadMuteMemRatioFlag returns the flag name for the 'Cache.ThreadMuteMemRatio' field
func CacheThreadMuteMemRatioFlag() string { return "cache-thread-mute-mem-ratio" }

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

// CacheTokenMemRatioFlag returns the flag name for the 'Cache.TokenMemRatio' field
func CacheTokenMemRatioFlag() string { return "cache-token-mem-ratio" }

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

// CacheTombstoneMemRatioFlag returns the flag name for the 'Cache.TombstoneMemRatio' field
func CacheTombstoneMemRatioFlag() string { return "cache-tombstone-mem-ratio" }

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

// CacheUserMemRatioFlag returns the flag name for the 'Cache.UserMemRatio' field
func CacheUserMemRatioFlag() string { return "cache-user-mem-ratio" }

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

// CacheUserMuteMemRatioFlag returns the flag name for the 'Cache.UserMuteMemRatio' field
func CacheUserMuteMemRatioFlag() string { return "cache-user-mute-mem-ratio" }

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

// CacheUserMuteIDsMemRatioFlag returns the flag name for the 'Cache.UserMuteIDsMemRatio' field
func CacheUserMuteIDsMemRatioFlag() string { return "cache-user-mute-ids-mem-ratio" }

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

// CacheWebfingerMemRatioFlag returns the flag name for the 'Cache.WebfingerMemRatio' field
func CacheWebfingerMemRatioFlag() string { return "cache-webfinger-mem-ratio" }

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

// CacheWebPushSubscriptionMemRatioFlag returns the flag name for the 'Cache.WebPushSubscriptionMemRatio' field
func CacheWebPushSubscriptionMemRatioFlag() string { return "cache-web-push-subscription-mem-ratio" }

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

// CacheWebPushSubscriptionIDsMemRatioFlag returns the flag name for the 'Cache.WebPushSubscriptionIDsMemRatio' field
func CacheWebPushSubscriptionIDsMemRatioFlag() string {
	return "cache-web-push-subscription-ids-mem-ratio"
}

// GetCacheWebPushSubscriptionIDsMemRatio safely fetches the value for global configuration 'Cache.WebPushSubscriptionIDsMemRatio' field
func GetCacheWebPushSubscriptionIDsMemRatio() float64 {
	return global.GetCacheWebPushSubscriptionIDsMemRatio()
}

// SetCacheWebPushSubscriptionIDsMemRatio safely sets the value for global configuration 'Cache.WebPushSubscriptionIDsMemRatio' field
func SetCacheWebPushSubscriptionIDsMemRatio(v float64) {
	global.SetCacheWebPushSubscriptionIDsMemRatio(v)
}

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

// CacheVisibilityMemRatioFlag returns the flag name for the 'Cache.VisibilityMemRatio' field
func CacheVisibilityMemRatioFlag() string { return "cache-visibility-mem-ratio" }

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

// AdminAccountUsernameFlag returns the flag name for the 'AdminAccountUsername' field
func AdminAccountUsernameFlag() string { return "username" }

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

// AdminAccountEmailFlag returns the flag name for the 'AdminAccountEmail' field
func AdminAccountEmailFlag() string { return "email" }

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

// AdminAccountPasswordFlag returns the flag name for the 'AdminAccountPassword' field
func AdminAccountPasswordFlag() string { return "password" }

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

// AdminTransPathFlag returns the flag name for the 'AdminTransPath' field
func AdminTransPathFlag() string { return "path" }

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

// AdminMediaPruneDryRunFlag returns the flag name for the 'AdminMediaPruneDryRun' field
func AdminMediaPruneDryRunFlag() string { return "dry-run" }

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

// AdminMediaListLocalOnlyFlag returns the flag name for the 'AdminMediaListLocalOnly' field
func AdminMediaListLocalOnlyFlag() string { return "local-only" }

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

// AdminMediaListRemoteOnlyFlag returns the flag name for the 'AdminMediaListRemoteOnly' field
func AdminMediaListRemoteOnlyFlag() string { return "remote-only" }

// GetAdminMediaListRemoteOnly safely fetches the value for global configuration 'AdminMediaListRemoteOnly' field
func GetAdminMediaListRemoteOnly() bool { return global.GetAdminMediaListRemoteOnly() }

// SetAdminMediaListRemoteOnly safely sets the value for global configuration 'AdminMediaListRemoteOnly' field
func SetAdminMediaListRemoteOnly(v bool) { global.SetAdminMediaListRemoteOnly(v) }

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

// RequestIDHeaderFlag returns the flag name for the 'RequestIDHeader' field
func RequestIDHeaderFlag() string { return "request-id-header" }

// GetRequestIDHeader safely fetches the value for global configuration 'RequestIDHeader' field
func GetRequestIDHeader() string { return global.GetRequestIDHeader() }

// SetRequestIDHeader safely sets the value for global configuration 'RequestIDHeader' field
func SetRequestIDHeader(v string) { global.SetRequestIDHeader(v) }
