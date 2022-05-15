package config

import (
	"github.com/spf13/cobra"
)

// BindConfigPath binds the given command's .ConfigPath pflag to global viper instance.
func BindConfigPath(cmd *cobra.Command) (err error) {
	Config(func(cfg *Configuration) {
		name := ConfigPathFlag()
		flag := cmd.Flags().Lookup(name)

		// Bind the config path pflag to global viper
		if err = gviper.BindPFlag(name, flag); err != nil {
			return
		}

		// Manually set config path from flag
		cfg.ConfigPath = flag.Value.String()
	})
	return
}

// BindFlags binds given command's pflags to the global viper instance.
func BindFlags(cmd *cobra.Command) (err error) {
	Config(func(cfg *Configuration) {
		// Bind the command pflags to global viper
		if err = gviper.BindPFlags(cmd.Flags()); err != nil {
			return
		}

		// Unmarshal viper values into global
		if err := cfg.unmarshal(gviper.AllSettings()); err != nil {
			return
		}
	})
	return
}

// AddGlobalFlags will attach global configuration flags to given cobra command, loading defaults from global config.
func AddGlobalFlags(cmd *cobra.Command) {
	Config(func(cfg *Configuration) {
		// General
		cmd.PersistentFlags().String(ApplicationNameFlag(), cfg.ApplicationName, fieldtag("ApplicationName", "usage"))
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
	Config(func(cfg *Configuration) {
		// Router
		cmd.PersistentFlags().String(BindAddressFlag(), cfg.BindAddress, fieldtag("BindAddress", "usage"))
		cmd.PersistentFlags().Int(PortFlag(), cfg.Port, fieldtag("Port", "usage"))
		cmd.PersistentFlags().StringSlice(TrustedProxiesFlag(), cfg.TrustedProxies, fieldtag("TrustedProxies", "usage"))

		// Template
		cmd.Flags().String(WebTemplateBaseDirFlag(), cfg.WebTemplateBaseDir, fieldtag("WebTemplateBaseDir", "usage"))
		cmd.Flags().String(WebAssetBaseDirFlag(), cfg.WebAssetBaseDir, fieldtag("WebAssetBaseDir", "usage"))

		// Accounts
		cmd.Flags().Bool(AccountsRegistrationOpenFlag(), cfg.AccountsRegistrationOpen, fieldtag("AccountsRegistrationOpen", "usage"))
		cmd.Flags().Bool(AccountsApprovalRequiredFlag(), cfg.AccountsApprovalRequired, fieldtag("AccountsApprovalRequired", "usage"))
		cmd.Flags().Bool(AccountsReasonRequiredFlag(), cfg.AccountsReasonRequired, fieldtag("AccountsReasonRequired", "usage"))

		// Media
		cmd.Flags().Int(MediaImageMaxSizeFlag(), cfg.MediaImageMaxSize, fieldtag("MediaImageMaxSize", "usage"))
		cmd.Flags().Int(MediaVideoMaxSizeFlag(), cfg.MediaVideoMaxSize, fieldtag("MediaVideoMaxSize", "usage"))
		cmd.Flags().Int(MediaDescriptionMinCharsFlag(), cfg.MediaDescriptionMinChars, fieldtag("MediaDescriptionMinChars", "usage"))
		cmd.Flags().Int(MediaDescriptionMaxCharsFlag(), cfg.MediaDescriptionMaxChars, fieldtag("MediaDescriptionMaxChars", "usage"))
		cmd.Flags().Int(MediaRemoteCacheDaysFlag(), cfg.MediaRemoteCacheDays, fieldtag("MediaRemoteCacheDays", "usage"))

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
	AddAdminAccount(cmd)
	name := AdminAccountPasswordFlag()
	usage := fieldtag("AdminAccountPassword", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// AddAdminAccountCreate attaches flags pertaining to admin account creation.
func AddAdminAccountCreate(cmd *cobra.Command) {
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
