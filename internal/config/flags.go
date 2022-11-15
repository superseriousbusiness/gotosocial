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
	"github.com/spf13/cobra"
)

// TODO: consolidate these methods into the Configuration{} or ConfigState{} structs.

// AddGlobalFlags will attach global configuration flags to given cobra command, loading defaults from global config.
func AddGlobalFlags(cmd *cobra.Command) {
	global.AddGlobalFlags(cmd)
}

// AddGlobalFlags will attach global configuration flags to given cobra command, loading defaults from State.
func (s *ConfigState) AddGlobalFlags(cmd *cobra.Command) {
	s.Config(func(cfg *Configuration) {
		// General
		cmd.PersistentFlags().String(ApplicationNameFlag(), cfg.ApplicationName, fieldtag("ApplicationName", "usage"))
		cmd.PersistentFlags().String(LandingPageUserFlag(), cfg.LandingPageUser, fieldtag("LandingPageUser", "usage"))
		cmd.PersistentFlags().String(HostFlag(), cfg.Host, fieldtag("Host", "usage"))
		cmd.PersistentFlags().String(AccountDomainFlag(), cfg.AccountDomain, fieldtag("AccountDomain", "usage"))
		cmd.PersistentFlags().String(ProtocolFlag(), cfg.Protocol, fieldtag("Protocol", "usage"))
		cmd.PersistentFlags().String(LogLevelFlag(), cfg.LogLevel, fieldtag("LogLevel", "usage"))
		cmd.PersistentFlags().Bool(LogDbQueriesFlag(), cfg.LogDbQueries, fieldtag("LogDbQueries", "usage"))
		cmd.PersistentFlags().String(ConfigPathFlag(), cfg.ConfigPath, fieldtag("ConfigPath", "usage"))

		// Database
		cmd.PersistentFlags().String(DbTypeFlag(), cfg.DbType, fieldtag("DbType", "usage"))
		cmd.PersistentFlags().String(DbAddressFlag(), cfg.DbAddress, fieldtag("DbAddress", "usage"))
		cmd.PersistentFlags().Int(DbPortFlag(), cfg.DbPort, fieldtag("DbPort", "usage"))
		cmd.PersistentFlags().String(DbUserFlag(), cfg.DbUser, fieldtag("DbUser", "usage"))
		cmd.PersistentFlags().String(DbPasswordFlag(), cfg.DbPassword, fieldtag("DbPassword", "usage"))
		cmd.PersistentFlags().String(DbDatabaseFlag(), cfg.DbDatabase, fieldtag("DbDatabase", "usage"))
		cmd.PersistentFlags().String(DbTLSModeFlag(), cfg.DbTLSMode, fieldtag("DbTLSMode", "usage"))
		cmd.PersistentFlags().String(DbTLSCACertFlag(), cfg.DbTLSCACert, fieldtag("DbTLSCACert", "usage"))
	})
}

// AddServerFlags will attach server configuration flags to given cobra command, loading defaults from global config.
func AddServerFlags(cmd *cobra.Command) {
	global.AddServerFlags(cmd)
}

// AddServerFlags will attach server configuration flags to given cobra command, loading defaults from State.
func (s *ConfigState) AddServerFlags(cmd *cobra.Command) {
	s.Config(func(cfg *Configuration) {
		// Router
		cmd.PersistentFlags().String(BindAddressFlag(), cfg.BindAddress, fieldtag("BindAddress", "usage"))
		cmd.PersistentFlags().Int(PortFlag(), cfg.Port, fieldtag("Port", "usage"))
		cmd.PersistentFlags().StringSlice(TrustedProxiesFlag(), cfg.TrustedProxies, fieldtag("TrustedProxies", "usage"))

		// Template
		cmd.Flags().String(WebTemplateBaseDirFlag(), cfg.WebTemplateBaseDir, fieldtag("WebTemplateBaseDir", "usage"))
		cmd.Flags().String(WebAssetBaseDirFlag(), cfg.WebAssetBaseDir, fieldtag("WebAssetBaseDir", "usage"))

		// Instance
		cmd.Flags().Bool(InstanceExposePeersFlag(), cfg.InstanceExposePeers, fieldtag("InstanceExposePeers", "usage"))
		cmd.Flags().Bool(InstanceExposeSuspendedFlag(), cfg.InstanceExposeSuspended, fieldtag("InstanceExposeSuspended", "usage"))
		cmd.Flags().Bool(InstanceDeliverToSharedInboxesFlag(), cfg.InstanceDeliverToSharedInboxes, fieldtag("InstanceDeliverToSharedInboxes", "usage"))

		// Accounts
		cmd.Flags().Bool(AccountsRegistrationOpenFlag(), cfg.AccountsRegistrationOpen, fieldtag("AccountsRegistrationOpen", "usage"))
		cmd.Flags().Bool(AccountsApprovalRequiredFlag(), cfg.AccountsApprovalRequired, fieldtag("AccountsApprovalRequired", "usage"))
		cmd.Flags().Bool(AccountsReasonRequiredFlag(), cfg.AccountsReasonRequired, fieldtag("AccountsReasonRequired", "usage"))
		cmd.Flags().Bool(AccountsAllowCustomCSSFlag(), cfg.AccountsAllowCustomCSS, fieldtag("AccountsAllowCustomCSS", "usage"))

		// Media
		cmd.Flags().Uint64(MediaImageMaxSizeFlag(), uint64(cfg.MediaImageMaxSize), fieldtag("MediaImageMaxSize", "usage"))
		cmd.Flags().Uint64(MediaVideoMaxSizeFlag(), uint64(cfg.MediaVideoMaxSize), fieldtag("MediaVideoMaxSize", "usage"))
		cmd.Flags().Int(MediaDescriptionMinCharsFlag(), cfg.MediaDescriptionMinChars, fieldtag("MediaDescriptionMinChars", "usage"))
		cmd.Flags().Int(MediaDescriptionMaxCharsFlag(), cfg.MediaDescriptionMaxChars, fieldtag("MediaDescriptionMaxChars", "usage"))
		cmd.Flags().Int(MediaRemoteCacheDaysFlag(), cfg.MediaRemoteCacheDays, fieldtag("MediaRemoteCacheDays", "usage"))
		cmd.Flags().Uint64(MediaEmojiLocalMaxSizeFlag(), uint64(cfg.MediaEmojiLocalMaxSize), fieldtag("MediaEmojiLocalMaxSize", "usage"))
		cmd.Flags().Uint64(MediaEmojiRemoteMaxSizeFlag(), uint64(cfg.MediaEmojiRemoteMaxSize), fieldtag("MediaEmojiRemoteMaxSize", "usage"))

		// Storage
		cmd.Flags().String(StorageBackendFlag(), cfg.StorageBackend, fieldtag("StorageBackend", "usage"))
		cmd.Flags().String(StorageLocalBasePathFlag(), cfg.StorageLocalBasePath, fieldtag("StorageLocalBasePath", "usage"))

		// Statuses
		cmd.Flags().Int(StatusesMaxCharsFlag(), cfg.StatusesMaxChars, fieldtag("StatusesMaxChars", "usage"))
		cmd.Flags().Int(StatusesCWMaxCharsFlag(), cfg.StatusesCWMaxChars, fieldtag("StatusesCWMaxChars", "usage"))
		cmd.Flags().Int(StatusesPollMaxOptionsFlag(), cfg.StatusesPollMaxOptions, fieldtag("StatusesPollMaxOptions", "usage"))
		cmd.Flags().Int(StatusesPollOptionMaxCharsFlag(), cfg.StatusesPollOptionMaxChars, fieldtag("StatusesPollOptionMaxChars", "usage"))
		cmd.Flags().Int(StatusesMediaMaxFilesFlag(), cfg.StatusesMediaMaxFiles, fieldtag("StatusesMediaMaxFiles", "usage"))

		// LetsEncrypt
		cmd.Flags().Bool(LetsEncryptEnabledFlag(), cfg.LetsEncryptEnabled, fieldtag("LetsEncryptEnabled", "usage"))
		cmd.Flags().Int(LetsEncryptPortFlag(), cfg.LetsEncryptPort, fieldtag("LetsEncryptPort", "usage"))
		cmd.Flags().String(LetsEncryptCertDirFlag(), cfg.LetsEncryptCertDir, fieldtag("LetsEncryptCertDir", "usage"))
		cmd.Flags().String(LetsEncryptEmailAddressFlag(), cfg.LetsEncryptEmailAddress, fieldtag("LetsEncryptEmailAddress", "usage"))

		// OIDC
		cmd.Flags().Bool(OIDCEnabledFlag(), cfg.OIDCEnabled, fieldtag("OIDCEnabled", "usage"))
		cmd.Flags().String(OIDCIdpNameFlag(), cfg.OIDCIdpName, fieldtag("OIDCIdpName", "usage"))
		cmd.Flags().Bool(OIDCSkipVerificationFlag(), cfg.OIDCSkipVerification, fieldtag("OIDCSkipVerification", "usage"))
		cmd.Flags().String(OIDCIssuerFlag(), cfg.OIDCIssuer, fieldtag("OIDCIssuer", "usage"))
		cmd.Flags().String(OIDCClientIDFlag(), cfg.OIDCClientID, fieldtag("OIDCClientID", "usage"))
		cmd.Flags().String(OIDCClientSecretFlag(), cfg.OIDCClientSecret, fieldtag("OIDCClientSecret", "usage"))
		cmd.Flags().StringSlice(OIDCScopesFlag(), cfg.OIDCScopes, fieldtag("OIDCScopes", "usage"))

		// SMTP
		cmd.Flags().String(SMTPHostFlag(), cfg.SMTPHost, fieldtag("SMTPHost", "usage"))
		cmd.Flags().Int(SMTPPortFlag(), cfg.SMTPPort, fieldtag("SMTPPort", "usage"))
		cmd.Flags().String(SMTPUsernameFlag(), cfg.SMTPUsername, fieldtag("SMTPUsername", "usage"))
		cmd.Flags().String(SMTPPasswordFlag(), cfg.SMTPPassword, fieldtag("SMTPPassword", "usage"))
		cmd.Flags().String(SMTPFromFlag(), cfg.SMTPFrom, fieldtag("SMTPFrom", "usage"))

		// Syslog
		cmd.Flags().Bool(SyslogEnabledFlag(), cfg.SyslogEnabled, fieldtag("SyslogEnabled", "usage"))
		cmd.Flags().String(SyslogProtocolFlag(), cfg.SyslogProtocol, fieldtag("SyslogProtocol", "usage"))
		cmd.Flags().String(SyslogAddressFlag(), cfg.SyslogAddress, fieldtag("SyslogAddress", "usage"))

		// Advanced flags
		cmd.Flags().String(AdvancedCookiesSamesiteFlag(), cfg.AdvancedCookiesSamesite, fieldtag("AdvancedCookiesSamesite", "usage"))
		cmd.Flags().Int(AdvancedRateLimitRequestsFlag(), cfg.AdvancedRateLimitRequests, fieldtag("AdvancedRateLimitRequests", "usage"))
	})
}

// AddAdminAccount attaches flags pertaining to admin account actions.
func AddAdminAccount(cmd *cobra.Command) {
	name := AdminAccountUsernameFlag()
	usage := fieldtag("AdminAccountUsername", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// AddAdminAccountPassword attaches flags pertaining to admin account password reset.
func AddAdminAccountPassword(cmd *cobra.Command) {
	name := AdminAccountPasswordFlag()
	usage := fieldtag("AdminAccountPassword", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// AddAdminAccountCreate attaches flags pertaining to admin account creation.
func AddAdminAccountCreate(cmd *cobra.Command) {
	// Requires both account and password
	AddAdminAccount(cmd)
	AddAdminAccountPassword(cmd)

	name := AdminAccountEmailFlag()
	usage := fieldtag("AdminAccountEmail", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// AddAdminTrans attaches flags pertaining to import/export commands.
func AddAdminTrans(cmd *cobra.Command) {
	name := AdminTransPathFlag()
	usage := fieldtag("AdminTransPath", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}
