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

// GetCacheAccountMaxSize safely fetches the Configuration value for state's 'Cache.AccountMaxSize' field
func (st *ConfigState) GetCacheAccountMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.AccountMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheAccountMaxSize safely sets the Configuration value for state's 'Cache.AccountMaxSize' field
func (st *ConfigState) SetCacheAccountMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.AccountMaxSize = v
	st.reloadToViper()
}

// CacheAccountMaxSizeFlag returns the flag name for the 'Cache.AccountMaxSize' field
func CacheAccountMaxSizeFlag() string { return "cache-account-max-size" }

// GetCacheAccountMaxSize safely fetches the value for global configuration 'Cache.AccountMaxSize' field
func GetCacheAccountMaxSize() int { return global.GetCacheAccountMaxSize() }

// SetCacheAccountMaxSize safely sets the value for global configuration 'Cache.AccountMaxSize' field
func SetCacheAccountMaxSize(v int) { global.SetCacheAccountMaxSize(v) }

// GetCacheAccountTTL safely fetches the Configuration value for state's 'Cache.AccountTTL' field
func (st *ConfigState) GetCacheAccountTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.AccountTTL
	st.mutex.Unlock()
	return
}

// SetCacheAccountTTL safely sets the Configuration value for state's 'Cache.AccountTTL' field
func (st *ConfigState) SetCacheAccountTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.AccountTTL = v
	st.reloadToViper()
}

// CacheAccountTTLFlag returns the flag name for the 'Cache.AccountTTL' field
func CacheAccountTTLFlag() string { return "cache-account-ttl" }

// GetCacheAccountTTL safely fetches the value for global configuration 'Cache.AccountTTL' field
func GetCacheAccountTTL() time.Duration { return global.GetCacheAccountTTL() }

// SetCacheAccountTTL safely sets the value for global configuration 'Cache.AccountTTL' field
func SetCacheAccountTTL(v time.Duration) { global.SetCacheAccountTTL(v) }

// GetCacheAccountSweepFreq safely fetches the Configuration value for state's 'Cache.AccountSweepFreq' field
func (st *ConfigState) GetCacheAccountSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.AccountSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheAccountSweepFreq safely sets the Configuration value for state's 'Cache.AccountSweepFreq' field
func (st *ConfigState) SetCacheAccountSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.AccountSweepFreq = v
	st.reloadToViper()
}

// CacheAccountSweepFreqFlag returns the flag name for the 'Cache.AccountSweepFreq' field
func CacheAccountSweepFreqFlag() string { return "cache-account-sweep-freq" }

// GetCacheAccountSweepFreq safely fetches the value for global configuration 'Cache.AccountSweepFreq' field
func GetCacheAccountSweepFreq() time.Duration { return global.GetCacheAccountSweepFreq() }

// SetCacheAccountSweepFreq safely sets the value for global configuration 'Cache.AccountSweepFreq' field
func SetCacheAccountSweepFreq(v time.Duration) { global.SetCacheAccountSweepFreq(v) }

// GetCacheBlockMaxSize safely fetches the Configuration value for state's 'Cache.BlockMaxSize' field
func (st *ConfigState) GetCacheBlockMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.BlockMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheBlockMaxSize safely sets the Configuration value for state's 'Cache.BlockMaxSize' field
func (st *ConfigState) SetCacheBlockMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.BlockMaxSize = v
	st.reloadToViper()
}

// CacheBlockMaxSizeFlag returns the flag name for the 'Cache.BlockMaxSize' field
func CacheBlockMaxSizeFlag() string { return "cache-block-max-size" }

// GetCacheBlockMaxSize safely fetches the value for global configuration 'Cache.BlockMaxSize' field
func GetCacheBlockMaxSize() int { return global.GetCacheBlockMaxSize() }

// SetCacheBlockMaxSize safely sets the value for global configuration 'Cache.BlockMaxSize' field
func SetCacheBlockMaxSize(v int) { global.SetCacheBlockMaxSize(v) }

// GetCacheBlockTTL safely fetches the Configuration value for state's 'Cache.BlockTTL' field
func (st *ConfigState) GetCacheBlockTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.BlockTTL
	st.mutex.Unlock()
	return
}

// SetCacheBlockTTL safely sets the Configuration value for state's 'Cache.BlockTTL' field
func (st *ConfigState) SetCacheBlockTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.BlockTTL = v
	st.reloadToViper()
}

// CacheBlockTTLFlag returns the flag name for the 'Cache.BlockTTL' field
func CacheBlockTTLFlag() string { return "cache-block-ttl" }

// GetCacheBlockTTL safely fetches the value for global configuration 'Cache.BlockTTL' field
func GetCacheBlockTTL() time.Duration { return global.GetCacheBlockTTL() }

// SetCacheBlockTTL safely sets the value for global configuration 'Cache.BlockTTL' field
func SetCacheBlockTTL(v time.Duration) { global.SetCacheBlockTTL(v) }

// GetCacheBlockSweepFreq safely fetches the Configuration value for state's 'Cache.BlockSweepFreq' field
func (st *ConfigState) GetCacheBlockSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.BlockSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheBlockSweepFreq safely sets the Configuration value for state's 'Cache.BlockSweepFreq' field
func (st *ConfigState) SetCacheBlockSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.BlockSweepFreq = v
	st.reloadToViper()
}

// CacheBlockSweepFreqFlag returns the flag name for the 'Cache.BlockSweepFreq' field
func CacheBlockSweepFreqFlag() string { return "cache-block-sweep-freq" }

// GetCacheBlockSweepFreq safely fetches the value for global configuration 'Cache.BlockSweepFreq' field
func GetCacheBlockSweepFreq() time.Duration { return global.GetCacheBlockSweepFreq() }

// SetCacheBlockSweepFreq safely sets the value for global configuration 'Cache.BlockSweepFreq' field
func SetCacheBlockSweepFreq(v time.Duration) { global.SetCacheBlockSweepFreq(v) }

// GetCacheDomainBlockMaxSize safely fetches the Configuration value for state's 'Cache.DomainBlockMaxSize' field
func (st *ConfigState) GetCacheDomainBlockMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.DomainBlockMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheDomainBlockMaxSize safely sets the Configuration value for state's 'Cache.DomainBlockMaxSize' field
func (st *ConfigState) SetCacheDomainBlockMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.DomainBlockMaxSize = v
	st.reloadToViper()
}

// CacheDomainBlockMaxSizeFlag returns the flag name for the 'Cache.DomainBlockMaxSize' field
func CacheDomainBlockMaxSizeFlag() string { return "cache-domain-block-max-size" }

// GetCacheDomainBlockMaxSize safely fetches the value for global configuration 'Cache.DomainBlockMaxSize' field
func GetCacheDomainBlockMaxSize() int { return global.GetCacheDomainBlockMaxSize() }

// SetCacheDomainBlockMaxSize safely sets the value for global configuration 'Cache.DomainBlockMaxSize' field
func SetCacheDomainBlockMaxSize(v int) { global.SetCacheDomainBlockMaxSize(v) }

// GetCacheDomainBlockTTL safely fetches the Configuration value for state's 'Cache.DomainBlockTTL' field
func (st *ConfigState) GetCacheDomainBlockTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.DomainBlockTTL
	st.mutex.Unlock()
	return
}

// SetCacheDomainBlockTTL safely sets the Configuration value for state's 'Cache.DomainBlockTTL' field
func (st *ConfigState) SetCacheDomainBlockTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.DomainBlockTTL = v
	st.reloadToViper()
}

// CacheDomainBlockTTLFlag returns the flag name for the 'Cache.DomainBlockTTL' field
func CacheDomainBlockTTLFlag() string { return "cache-domain-block-ttl" }

// GetCacheDomainBlockTTL safely fetches the value for global configuration 'Cache.DomainBlockTTL' field
func GetCacheDomainBlockTTL() time.Duration { return global.GetCacheDomainBlockTTL() }

// SetCacheDomainBlockTTL safely sets the value for global configuration 'Cache.DomainBlockTTL' field
func SetCacheDomainBlockTTL(v time.Duration) { global.SetCacheDomainBlockTTL(v) }

// GetCacheDomainBlockSweepFreq safely fetches the Configuration value for state's 'Cache.DomainBlockSweepFreq' field
func (st *ConfigState) GetCacheDomainBlockSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.DomainBlockSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheDomainBlockSweepFreq safely sets the Configuration value for state's 'Cache.DomainBlockSweepFreq' field
func (st *ConfigState) SetCacheDomainBlockSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.DomainBlockSweepFreq = v
	st.reloadToViper()
}

// CacheDomainBlockSweepFreqFlag returns the flag name for the 'Cache.DomainBlockSweepFreq' field
func CacheDomainBlockSweepFreqFlag() string { return "cache-domain-block-sweep-freq" }

// GetCacheDomainBlockSweepFreq safely fetches the value for global configuration 'Cache.DomainBlockSweepFreq' field
func GetCacheDomainBlockSweepFreq() time.Duration { return global.GetCacheDomainBlockSweepFreq() }

// SetCacheDomainBlockSweepFreq safely sets the value for global configuration 'Cache.DomainBlockSweepFreq' field
func SetCacheDomainBlockSweepFreq(v time.Duration) { global.SetCacheDomainBlockSweepFreq(v) }

// GetCacheEmojiMaxSize safely fetches the Configuration value for state's 'Cache.EmojiMaxSize' field
func (st *ConfigState) GetCacheEmojiMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.EmojiMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheEmojiMaxSize safely sets the Configuration value for state's 'Cache.EmojiMaxSize' field
func (st *ConfigState) SetCacheEmojiMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.EmojiMaxSize = v
	st.reloadToViper()
}

// CacheEmojiMaxSizeFlag returns the flag name for the 'Cache.EmojiMaxSize' field
func CacheEmojiMaxSizeFlag() string { return "cache-emoji-max-size" }

// GetCacheEmojiMaxSize safely fetches the value for global configuration 'Cache.EmojiMaxSize' field
func GetCacheEmojiMaxSize() int { return global.GetCacheEmojiMaxSize() }

// SetCacheEmojiMaxSize safely sets the value for global configuration 'Cache.EmojiMaxSize' field
func SetCacheEmojiMaxSize(v int) { global.SetCacheEmojiMaxSize(v) }

// GetCacheEmojiTTL safely fetches the Configuration value for state's 'Cache.EmojiTTL' field
func (st *ConfigState) GetCacheEmojiTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.EmojiTTL
	st.mutex.Unlock()
	return
}

// SetCacheEmojiTTL safely sets the Configuration value for state's 'Cache.EmojiTTL' field
func (st *ConfigState) SetCacheEmojiTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.EmojiTTL = v
	st.reloadToViper()
}

// CacheEmojiTTLFlag returns the flag name for the 'Cache.EmojiTTL' field
func CacheEmojiTTLFlag() string { return "cache-emoji-ttl" }

// GetCacheEmojiTTL safely fetches the value for global configuration 'Cache.EmojiTTL' field
func GetCacheEmojiTTL() time.Duration { return global.GetCacheEmojiTTL() }

// SetCacheEmojiTTL safely sets the value for global configuration 'Cache.EmojiTTL' field
func SetCacheEmojiTTL(v time.Duration) { global.SetCacheEmojiTTL(v) }

// GetCacheEmojiSweepFreq safely fetches the Configuration value for state's 'Cache.EmojiSweepFreq' field
func (st *ConfigState) GetCacheEmojiSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.EmojiSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheEmojiSweepFreq safely sets the Configuration value for state's 'Cache.EmojiSweepFreq' field
func (st *ConfigState) SetCacheEmojiSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.EmojiSweepFreq = v
	st.reloadToViper()
}

// CacheEmojiSweepFreqFlag returns the flag name for the 'Cache.EmojiSweepFreq' field
func CacheEmojiSweepFreqFlag() string { return "cache-emoji-sweep-freq" }

// GetCacheEmojiSweepFreq safely fetches the value for global configuration 'Cache.EmojiSweepFreq' field
func GetCacheEmojiSweepFreq() time.Duration { return global.GetCacheEmojiSweepFreq() }

// SetCacheEmojiSweepFreq safely sets the value for global configuration 'Cache.EmojiSweepFreq' field
func SetCacheEmojiSweepFreq(v time.Duration) { global.SetCacheEmojiSweepFreq(v) }

// GetCacheEmojiCategoryMaxSize safely fetches the Configuration value for state's 'Cache.EmojiCategoryMaxSize' field
func (st *ConfigState) GetCacheEmojiCategoryMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.EmojiCategoryMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheEmojiCategoryMaxSize safely sets the Configuration value for state's 'Cache.EmojiCategoryMaxSize' field
func (st *ConfigState) SetCacheEmojiCategoryMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.EmojiCategoryMaxSize = v
	st.reloadToViper()
}

// CacheEmojiCategoryMaxSizeFlag returns the flag name for the 'Cache.EmojiCategoryMaxSize' field
func CacheEmojiCategoryMaxSizeFlag() string { return "cache-emoji-category-max-size" }

// GetCacheEmojiCategoryMaxSize safely fetches the value for global configuration 'Cache.EmojiCategoryMaxSize' field
func GetCacheEmojiCategoryMaxSize() int { return global.GetCacheEmojiCategoryMaxSize() }

// SetCacheEmojiCategoryMaxSize safely sets the value for global configuration 'Cache.EmojiCategoryMaxSize' field
func SetCacheEmojiCategoryMaxSize(v int) { global.SetCacheEmojiCategoryMaxSize(v) }

// GetCacheEmojiCategoryTTL safely fetches the Configuration value for state's 'Cache.EmojiCategoryTTL' field
func (st *ConfigState) GetCacheEmojiCategoryTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.EmojiCategoryTTL
	st.mutex.Unlock()
	return
}

// SetCacheEmojiCategoryTTL safely sets the Configuration value for state's 'Cache.EmojiCategoryTTL' field
func (st *ConfigState) SetCacheEmojiCategoryTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.EmojiCategoryTTL = v
	st.reloadToViper()
}

// CacheEmojiCategoryTTLFlag returns the flag name for the 'Cache.EmojiCategoryTTL' field
func CacheEmojiCategoryTTLFlag() string { return "cache-emoji-category-ttl" }

// GetCacheEmojiCategoryTTL safely fetches the value for global configuration 'Cache.EmojiCategoryTTL' field
func GetCacheEmojiCategoryTTL() time.Duration { return global.GetCacheEmojiCategoryTTL() }

// SetCacheEmojiCategoryTTL safely sets the value for global configuration 'Cache.EmojiCategoryTTL' field
func SetCacheEmojiCategoryTTL(v time.Duration) { global.SetCacheEmojiCategoryTTL(v) }

// GetCacheEmojiCategorySweepFreq safely fetches the Configuration value for state's 'Cache.EmojiCategorySweepFreq' field
func (st *ConfigState) GetCacheEmojiCategorySweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.EmojiCategorySweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheEmojiCategorySweepFreq safely sets the Configuration value for state's 'Cache.EmojiCategorySweepFreq' field
func (st *ConfigState) SetCacheEmojiCategorySweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.EmojiCategorySweepFreq = v
	st.reloadToViper()
}

// CacheEmojiCategorySweepFreqFlag returns the flag name for the 'Cache.EmojiCategorySweepFreq' field
func CacheEmojiCategorySweepFreqFlag() string { return "cache-emoji-category-sweep-freq" }

// GetCacheEmojiCategorySweepFreq safely fetches the value for global configuration 'Cache.EmojiCategorySweepFreq' field
func GetCacheEmojiCategorySweepFreq() time.Duration { return global.GetCacheEmojiCategorySweepFreq() }

// SetCacheEmojiCategorySweepFreq safely sets the value for global configuration 'Cache.EmojiCategorySweepFreq' field
func SetCacheEmojiCategorySweepFreq(v time.Duration) { global.SetCacheEmojiCategorySweepFreq(v) }

// GetCacheMentionMaxSize safely fetches the Configuration value for state's 'Cache.MentionMaxSize' field
func (st *ConfigState) GetCacheMentionMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.MentionMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheMentionMaxSize safely sets the Configuration value for state's 'Cache.MentionMaxSize' field
func (st *ConfigState) SetCacheMentionMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.MentionMaxSize = v
	st.reloadToViper()
}

// CacheMentionMaxSizeFlag returns the flag name for the 'Cache.MentionMaxSize' field
func CacheMentionMaxSizeFlag() string { return "cache-mention-max-size" }

// GetCacheMentionMaxSize safely fetches the value for global configuration 'Cache.MentionMaxSize' field
func GetCacheMentionMaxSize() int { return global.GetCacheMentionMaxSize() }

// SetCacheMentionMaxSize safely sets the value for global configuration 'Cache.MentionMaxSize' field
func SetCacheMentionMaxSize(v int) { global.SetCacheMentionMaxSize(v) }

// GetCacheMentionTTL safely fetches the Configuration value for state's 'Cache.MentionTTL' field
func (st *ConfigState) GetCacheMentionTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.MentionTTL
	st.mutex.Unlock()
	return
}

// SetCacheMentionTTL safely sets the Configuration value for state's 'Cache.MentionTTL' field
func (st *ConfigState) SetCacheMentionTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.MentionTTL = v
	st.reloadToViper()
}

// CacheMentionTTLFlag returns the flag name for the 'Cache.MentionTTL' field
func CacheMentionTTLFlag() string { return "cache-mention-ttl" }

// GetCacheMentionTTL safely fetches the value for global configuration 'Cache.MentionTTL' field
func GetCacheMentionTTL() time.Duration { return global.GetCacheMentionTTL() }

// SetCacheMentionTTL safely sets the value for global configuration 'Cache.MentionTTL' field
func SetCacheMentionTTL(v time.Duration) { global.SetCacheMentionTTL(v) }

// GetCacheMentionSweepFreq safely fetches the Configuration value for state's 'Cache.MentionSweepFreq' field
func (st *ConfigState) GetCacheMentionSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.MentionSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheMentionSweepFreq safely sets the Configuration value for state's 'Cache.MentionSweepFreq' field
func (st *ConfigState) SetCacheMentionSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.MentionSweepFreq = v
	st.reloadToViper()
}

// CacheMentionSweepFreqFlag returns the flag name for the 'Cache.MentionSweepFreq' field
func CacheMentionSweepFreqFlag() string { return "cache-mention-sweep-freq" }

// GetCacheMentionSweepFreq safely fetches the value for global configuration 'Cache.MentionSweepFreq' field
func GetCacheMentionSweepFreq() time.Duration { return global.GetCacheMentionSweepFreq() }

// SetCacheMentionSweepFreq safely sets the value for global configuration 'Cache.MentionSweepFreq' field
func SetCacheMentionSweepFreq(v time.Duration) { global.SetCacheMentionSweepFreq(v) }

// GetCacheNotificationMaxSize safely fetches the Configuration value for state's 'Cache.NotificationMaxSize' field
func (st *ConfigState) GetCacheNotificationMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.NotificationMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheNotificationMaxSize safely sets the Configuration value for state's 'Cache.NotificationMaxSize' field
func (st *ConfigState) SetCacheNotificationMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.NotificationMaxSize = v
	st.reloadToViper()
}

// CacheNotificationMaxSizeFlag returns the flag name for the 'Cache.NotificationMaxSize' field
func CacheNotificationMaxSizeFlag() string { return "cache-notification-max-size" }

// GetCacheNotificationMaxSize safely fetches the value for global configuration 'Cache.NotificationMaxSize' field
func GetCacheNotificationMaxSize() int { return global.GetCacheNotificationMaxSize() }

// SetCacheNotificationMaxSize safely sets the value for global configuration 'Cache.NotificationMaxSize' field
func SetCacheNotificationMaxSize(v int) { global.SetCacheNotificationMaxSize(v) }

// GetCacheNotificationTTL safely fetches the Configuration value for state's 'Cache.NotificationTTL' field
func (st *ConfigState) GetCacheNotificationTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.NotificationTTL
	st.mutex.Unlock()
	return
}

// SetCacheNotificationTTL safely sets the Configuration value for state's 'Cache.NotificationTTL' field
func (st *ConfigState) SetCacheNotificationTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.NotificationTTL = v
	st.reloadToViper()
}

// CacheNotificationTTLFlag returns the flag name for the 'Cache.NotificationTTL' field
func CacheNotificationTTLFlag() string { return "cache-notification-ttl" }

// GetCacheNotificationTTL safely fetches the value for global configuration 'Cache.NotificationTTL' field
func GetCacheNotificationTTL() time.Duration { return global.GetCacheNotificationTTL() }

// SetCacheNotificationTTL safely sets the value for global configuration 'Cache.NotificationTTL' field
func SetCacheNotificationTTL(v time.Duration) { global.SetCacheNotificationTTL(v) }

// GetCacheNotificationSweepFreq safely fetches the Configuration value for state's 'Cache.NotificationSweepFreq' field
func (st *ConfigState) GetCacheNotificationSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.NotificationSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheNotificationSweepFreq safely sets the Configuration value for state's 'Cache.NotificationSweepFreq' field
func (st *ConfigState) SetCacheNotificationSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.NotificationSweepFreq = v
	st.reloadToViper()
}

// CacheNotificationSweepFreqFlag returns the flag name for the 'Cache.NotificationSweepFreq' field
func CacheNotificationSweepFreqFlag() string { return "cache-notification-sweep-freq" }

// GetCacheNotificationSweepFreq safely fetches the value for global configuration 'Cache.NotificationSweepFreq' field
func GetCacheNotificationSweepFreq() time.Duration { return global.GetCacheNotificationSweepFreq() }

// SetCacheNotificationSweepFreq safely sets the value for global configuration 'Cache.NotificationSweepFreq' field
func SetCacheNotificationSweepFreq(v time.Duration) { global.SetCacheNotificationSweepFreq(v) }

// GetCacheStatusMaxSize safely fetches the Configuration value for state's 'Cache.StatusMaxSize' field
func (st *ConfigState) GetCacheStatusMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.StatusMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheStatusMaxSize safely sets the Configuration value for state's 'Cache.StatusMaxSize' field
func (st *ConfigState) SetCacheStatusMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.StatusMaxSize = v
	st.reloadToViper()
}

// CacheStatusMaxSizeFlag returns the flag name for the 'Cache.StatusMaxSize' field
func CacheStatusMaxSizeFlag() string { return "cache-status-max-size" }

// GetCacheStatusMaxSize safely fetches the value for global configuration 'Cache.StatusMaxSize' field
func GetCacheStatusMaxSize() int { return global.GetCacheStatusMaxSize() }

// SetCacheStatusMaxSize safely sets the value for global configuration 'Cache.StatusMaxSize' field
func SetCacheStatusMaxSize(v int) { global.SetCacheStatusMaxSize(v) }

// GetCacheStatusTTL safely fetches the Configuration value for state's 'Cache.StatusTTL' field
func (st *ConfigState) GetCacheStatusTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.StatusTTL
	st.mutex.Unlock()
	return
}

// SetCacheStatusTTL safely sets the Configuration value for state's 'Cache.StatusTTL' field
func (st *ConfigState) SetCacheStatusTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.StatusTTL = v
	st.reloadToViper()
}

// CacheStatusTTLFlag returns the flag name for the 'Cache.StatusTTL' field
func CacheStatusTTLFlag() string { return "cache-status-ttl" }

// GetCacheStatusTTL safely fetches the value for global configuration 'Cache.StatusTTL' field
func GetCacheStatusTTL() time.Duration { return global.GetCacheStatusTTL() }

// SetCacheStatusTTL safely sets the value for global configuration 'Cache.StatusTTL' field
func SetCacheStatusTTL(v time.Duration) { global.SetCacheStatusTTL(v) }

// GetCacheStatusSweepFreq safely fetches the Configuration value for state's 'Cache.StatusSweepFreq' field
func (st *ConfigState) GetCacheStatusSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.StatusSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheStatusSweepFreq safely sets the Configuration value for state's 'Cache.StatusSweepFreq' field
func (st *ConfigState) SetCacheStatusSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.StatusSweepFreq = v
	st.reloadToViper()
}

// CacheStatusSweepFreqFlag returns the flag name for the 'Cache.StatusSweepFreq' field
func CacheStatusSweepFreqFlag() string { return "cache-status-sweep-freq" }

// GetCacheStatusSweepFreq safely fetches the value for global configuration 'Cache.StatusSweepFreq' field
func GetCacheStatusSweepFreq() time.Duration { return global.GetCacheStatusSweepFreq() }

// SetCacheStatusSweepFreq safely sets the value for global configuration 'Cache.StatusSweepFreq' field
func SetCacheStatusSweepFreq(v time.Duration) { global.SetCacheStatusSweepFreq(v) }

// GetCacheTombstoneMaxSize safely fetches the Configuration value for state's 'Cache.TombstoneMaxSize' field
func (st *ConfigState) GetCacheTombstoneMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.TombstoneMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheTombstoneMaxSize safely sets the Configuration value for state's 'Cache.TombstoneMaxSize' field
func (st *ConfigState) SetCacheTombstoneMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.TombstoneMaxSize = v
	st.reloadToViper()
}

// CacheTombstoneMaxSizeFlag returns the flag name for the 'Cache.TombstoneMaxSize' field
func CacheTombstoneMaxSizeFlag() string { return "cache-tombstone-max-size" }

// GetCacheTombstoneMaxSize safely fetches the value for global configuration 'Cache.TombstoneMaxSize' field
func GetCacheTombstoneMaxSize() int { return global.GetCacheTombstoneMaxSize() }

// SetCacheTombstoneMaxSize safely sets the value for global configuration 'Cache.TombstoneMaxSize' field
func SetCacheTombstoneMaxSize(v int) { global.SetCacheTombstoneMaxSize(v) }

// GetCacheTombstoneTTL safely fetches the Configuration value for state's 'Cache.TombstoneTTL' field
func (st *ConfigState) GetCacheTombstoneTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.TombstoneTTL
	st.mutex.Unlock()
	return
}

// SetCacheTombstoneTTL safely sets the Configuration value for state's 'Cache.TombstoneTTL' field
func (st *ConfigState) SetCacheTombstoneTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.TombstoneTTL = v
	st.reloadToViper()
}

// CacheTombstoneTTLFlag returns the flag name for the 'Cache.TombstoneTTL' field
func CacheTombstoneTTLFlag() string { return "cache-tombstone-ttl" }

// GetCacheTombstoneTTL safely fetches the value for global configuration 'Cache.TombstoneTTL' field
func GetCacheTombstoneTTL() time.Duration { return global.GetCacheTombstoneTTL() }

// SetCacheTombstoneTTL safely sets the value for global configuration 'Cache.TombstoneTTL' field
func SetCacheTombstoneTTL(v time.Duration) { global.SetCacheTombstoneTTL(v) }

// GetCacheTombstoneSweepFreq safely fetches the Configuration value for state's 'Cache.TombstoneSweepFreq' field
func (st *ConfigState) GetCacheTombstoneSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.TombstoneSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheTombstoneSweepFreq safely sets the Configuration value for state's 'Cache.TombstoneSweepFreq' field
func (st *ConfigState) SetCacheTombstoneSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.TombstoneSweepFreq = v
	st.reloadToViper()
}

// CacheTombstoneSweepFreqFlag returns the flag name for the 'Cache.TombstoneSweepFreq' field
func CacheTombstoneSweepFreqFlag() string { return "cache-tombstone-sweep-freq" }

// GetCacheTombstoneSweepFreq safely fetches the value for global configuration 'Cache.TombstoneSweepFreq' field
func GetCacheTombstoneSweepFreq() time.Duration { return global.GetCacheTombstoneSweepFreq() }

// SetCacheTombstoneSweepFreq safely sets the value for global configuration 'Cache.TombstoneSweepFreq' field
func SetCacheTombstoneSweepFreq(v time.Duration) { global.SetCacheTombstoneSweepFreq(v) }

// GetCacheUserMaxSize safely fetches the Configuration value for state's 'Cache.UserMaxSize' field
func (st *ConfigState) GetCacheUserMaxSize() (v int) {
	st.mutex.Lock()
	v = st.config.Cache.UserMaxSize
	st.mutex.Unlock()
	return
}

// SetCacheUserMaxSize safely sets the Configuration value for state's 'Cache.UserMaxSize' field
func (st *ConfigState) SetCacheUserMaxSize(v int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.UserMaxSize = v
	st.reloadToViper()
}

// CacheUserMaxSizeFlag returns the flag name for the 'Cache.UserMaxSize' field
func CacheUserMaxSizeFlag() string { return "cache-user-max-size" }

// GetCacheUserMaxSize safely fetches the value for global configuration 'Cache.UserMaxSize' field
func GetCacheUserMaxSize() int { return global.GetCacheUserMaxSize() }

// SetCacheUserMaxSize safely sets the value for global configuration 'Cache.UserMaxSize' field
func SetCacheUserMaxSize(v int) { global.SetCacheUserMaxSize(v) }

// GetCacheUserTTL safely fetches the Configuration value for state's 'Cache.UserTTL' field
func (st *ConfigState) GetCacheUserTTL() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.UserTTL
	st.mutex.Unlock()
	return
}

// SetCacheUserTTL safely sets the Configuration value for state's 'Cache.UserTTL' field
func (st *ConfigState) SetCacheUserTTL(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.UserTTL = v
	st.reloadToViper()
}

// CacheUserTTLFlag returns the flag name for the 'Cache.UserTTL' field
func CacheUserTTLFlag() string { return "cache-user-ttl" }

// GetCacheUserTTL safely fetches the value for global configuration 'Cache.UserTTL' field
func GetCacheUserTTL() time.Duration { return global.GetCacheUserTTL() }

// SetCacheUserTTL safely sets the value for global configuration 'Cache.UserTTL' field
func SetCacheUserTTL(v time.Duration) { global.SetCacheUserTTL(v) }

// GetCacheUserSweepFreq safely fetches the Configuration value for state's 'Cache.UserSweepFreq' field
func (st *ConfigState) GetCacheUserSweepFreq() (v time.Duration) {
	st.mutex.Lock()
	v = st.config.Cache.UserSweepFreq
	st.mutex.Unlock()
	return
}

// SetCacheUserSweepFreq safely sets the Configuration value for state's 'Cache.UserSweepFreq' field
func (st *ConfigState) SetCacheUserSweepFreq(v time.Duration) {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	st.config.Cache.UserSweepFreq = v
	st.reloadToViper()
}

// CacheUserSweepFreqFlag returns the flag name for the 'Cache.UserSweepFreq' field
func CacheUserSweepFreqFlag() string { return "cache-user-sweep-freq" }

// GetCacheUserSweepFreq safely fetches the value for global configuration 'Cache.UserSweepFreq' field
func GetCacheUserSweepFreq() time.Duration { return global.GetCacheUserSweepFreq() }

// SetCacheUserSweepFreq safely sets the value for global configuration 'Cache.UserSweepFreq' field
func SetCacheUserSweepFreq(v time.Duration) { global.SetCacheUserSweepFreq(v) }

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
