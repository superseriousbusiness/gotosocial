package config

import "github.com/spf13/cobra"

// BindConfigPath binds the given command's .ConfigPath pflag to global viper instance.
func BindConfigPath(cmd *cobra.Command) error {
	mutex.Lock()
	defer mutex.Unlock()

	name := fieldtag("ConfigPath", "name")
	flag := cmd.PersistentFlags().Lookup(name)

	// Bind the config path pflag to global viper
	if err := gviper.BindPFlag(name, flag); err != nil {
		return err
	}

	// Manually set the config path variable
	global.ConfigPath = gviper.GetString(name)

	return nil
}

// BindFlags binds given command's pflags to the global viper instance.
func BindFlags(cmd *cobra.Command) (err error) {
	mutex.Lock()
	defer mutex.Unlock()

	// Bind the command pflags to global viper
	if err := gviper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	// Unmarshal viper values into global
	if err := gviper.Unmarshal(&global); err != nil {
		return err
	}

	return nil
}

// AddGlobalFlags will attach global configuration flags to given cobra command, loading defaults from global config.
func AddGlobalFlags(cmd *cobra.Command) {
	Config(func(cfg *Configuration) {
		// General
		cmd.PersistentFlags().String(fieldtag("ApplicationName", "name"), cfg.ApplicationName, fieldtag("ApplicationName", "usage"))
		cmd.PersistentFlags().String(fieldtag("Host", "name"), cfg.Host, fieldtag("Host", "usage"))
		cmd.PersistentFlags().String(fieldtag("AccountDomain", "name"), cfg.AccountDomain, fieldtag("AccountDomain", "usage"))
		cmd.PersistentFlags().String(fieldtag("Protocol", "name"), cfg.Protocol, fieldtag("Protocol", "usage"))
		cmd.PersistentFlags().String(fieldtag("LogLevel", "name"), cfg.LogLevel, fieldtag("LogLevel", "usage"))
		cmd.PersistentFlags().Bool(fieldtag("LogDbQueries", "name"), cfg.LogDbQueries, fieldtag("LogDbQueries", "usage"))
		cmd.PersistentFlags().String(fieldtag("ConfigPath", "name"), cfg.ConfigPath, fieldtag("ConfigPath", "usage"))

		// Database
		cmd.PersistentFlags().String(fieldtag("DbType", "name"), cfg.DbType, fieldtag("DbType", "usage"))
		cmd.PersistentFlags().String(fieldtag("DbAddress", "name"), cfg.DbAddress, fieldtag("DbAddress", "usage"))
		cmd.PersistentFlags().Int(fieldtag("DbPort", "name"), cfg.DbPort, fieldtag("DbPort", "usage"))
		cmd.PersistentFlags().String(fieldtag("DbUser", "name"), cfg.DbUser, fieldtag("DbUser", "usage"))
		cmd.PersistentFlags().String(fieldtag("DbPassword", "name"), cfg.DbPassword, fieldtag("DbPassword", "usage"))
		cmd.PersistentFlags().String(fieldtag("DbDatabase", "name"), cfg.DbDatabase, fieldtag("DbDatabase", "usage"))
		cmd.PersistentFlags().String(fieldtag("DbTLSMode", "name"), cfg.DbTLSMode, fieldtag("DbTLSMode", "usage"))
		cmd.PersistentFlags().String(fieldtag("DbTLSCACert", "name"), cfg.DbTLSCACert, fieldtag("DbTLSCACert", "usage"))
	})
}

// AddServerFlags will attach server configuration flags to given cobra command, loading defaults from global config.
func AddServerFlags(cmd *cobra.Command) {
	Config(func(cfg *Configuration) {
		// Router
		cmd.PersistentFlags().String(fieldtag("BindAddress", "name"), cfg.BindAddress, fieldtag("BindAddress", "usage"))
		cmd.PersistentFlags().Int(fieldtag("Port", "name"), cfg.Port, fieldtag("Port", "usage"))
		cmd.PersistentFlags().StringSlice(fieldtag("TrustedProxies", "name"), cfg.TrustedProxies, fieldtag("TrustedProxies", "usage"))

		// Template
		cmd.Flags().String(fieldtag("WebTemplateBaseDir", "name"), cfg.WebTemplateBaseDir, fieldtag("WebTemplateBaseDir", "usage"))
		cmd.Flags().String(fieldtag("WebAssetBaseDir", "name"), cfg.WebAssetBaseDir, fieldtag("WebAssetBaseDir", "usage"))

		// Accounts
		cmd.Flags().Bool(fieldtag("AccountsRegistrationOpen", "name"), cfg.AccountsRegistrationOpen, fieldtag("AccountsRegistrationOpen", "usage"))
		cmd.Flags().Bool(fieldtag("AccountsApprovalRequired", "name"), cfg.AccountsApprovalRequired, fieldtag("AccountsApprovalRequired", "usage"))
		cmd.Flags().Bool(fieldtag("AccountsReasonRequired", "name"), cfg.AccountsReasonRequired, fieldtag("AccountsReasonRequired", "usage"))

		// Media
		cmd.Flags().Int(fieldtag("MediaImageMaxSize", "name"), cfg.MediaImageMaxSize, fieldtag("MediaImageMaxSize", "usage"))
		cmd.Flags().Int(fieldtag("MediaVideoMaxSize", "name"), cfg.MediaVideoMaxSize, fieldtag("MediaVideoMaxSize", "usage"))
		cmd.Flags().Int(fieldtag("MediaDescriptionMinChars", "name"), cfg.MediaDescriptionMinChars, fieldtag("MediaDescriptionMinChars", "usage"))
		cmd.Flags().Int(fieldtag("MediaDescriptionMaxChars", "name"), cfg.MediaDescriptionMaxChars, fieldtag("MediaDescriptionMaxChars", "usage"))
		cmd.Flags().Int(fieldtag("MediaRemoteCacheDays", "name"), cfg.MediaRemoteCacheDays, fieldtag("MediaRemoteCacheDays", "usage"))

		// Storage
		cmd.Flags().String(fieldtag("StorageBackend", "name"), cfg.StorageBackend, fieldtag("StorageBackend", "usage"))
		cmd.Flags().String(fieldtag("StorageLocalBasePath", "name"), cfg.StorageLocalBasePath, fieldtag("StorageLocalBasePath", "usage"))

		// Statuses
		cmd.Flags().Int(fieldtag("StatusesMaxChars", "name"), cfg.StatusesMaxChars, fieldtag("StatusesMaxChars", "usage"))
		cmd.Flags().Int(fieldtag("StatusesCWMaxChars", "name"), cfg.StatusesCWMaxChars, fieldtag("StatusesCWMaxChars", "usage"))
		cmd.Flags().Int(fieldtag("StatusesPollMaxOptions", "name"), cfg.StatusesPollMaxOptions, fieldtag("StatusesPollMaxOptions", "usage"))
		cmd.Flags().Int(fieldtag("StatusesPollOptionMaxChars", "name"), cfg.StatusesPollOptionMaxChars, fieldtag("StatusesPollOptionMaxChars", "usage"))
		cmd.Flags().Int(fieldtag("StatusesMediaMaxFiles", "name"), cfg.StatusesMediaMaxFiles, fieldtag("StatusesMediaMaxFiles", "usage"))

		// LetsEncrypt
		cmd.Flags().Bool(fieldtag("LetsEncryptEnabled", "name"), cfg.LetsEncryptEnabled, fieldtag("LetsEncryptEnabled", "usage"))
		cmd.Flags().Int(fieldtag("LetsEncryptPort", "name"), cfg.LetsEncryptPort, fieldtag("LetsEncryptPort", "usage"))
		cmd.Flags().String(fieldtag("LetsEncryptCertDir", "name"), cfg.LetsEncryptCertDir, fieldtag("LetsEncryptCertDir", "usage"))
		cmd.Flags().String(fieldtag("LetsEncryptEmailAddress", "name"), cfg.LetsEncryptEmailAddress, fieldtag("LetsEncryptEmailAddress", "usage"))

		// OIDC
		cmd.Flags().Bool(fieldtag("OIDCEnabled", "name"), cfg.OIDCEnabled, fieldtag("OIDCEnabled", "usage"))
		cmd.Flags().String(fieldtag("OIDCIdpName", "name"), cfg.OIDCIdpName, fieldtag("OIDCIdpName", "usage"))
		cmd.Flags().Bool(fieldtag("OIDCSkipVerification", "name"), cfg.OIDCSkipVerification, fieldtag("OIDCSkipVerification", "usage"))
		cmd.Flags().String(fieldtag("OIDCIssuer", "name"), cfg.OIDCIssuer, fieldtag("OIDCIssuer", "usage"))
		cmd.Flags().String(fieldtag("OIDCClientID", "name"), cfg.OIDCClientID, fieldtag("OIDCClientID", "usage"))
		cmd.Flags().String(fieldtag("OIDCClientSecret", "name"), cfg.OIDCClientSecret, fieldtag("OIDCClientSecret", "usage"))
		cmd.Flags().StringSlice(fieldtag("OIDCScopes", "name"), cfg.OIDCScopes, fieldtag("OIDCScopes", "usage"))

		// SMTP
		cmd.Flags().String(fieldtag("SMTPHost", "name"), cfg.SMTPHost, fieldtag("SMTPHost", "usage"))
		cmd.Flags().Int(fieldtag("SMTPPort", "name"), cfg.SMTPPort, fieldtag("SMTPPort", "usage"))
		cmd.Flags().String(fieldtag("SMTPUsername", "name"), cfg.SMTPUsername, fieldtag("SMTPUsername", "usage"))
		cmd.Flags().String(fieldtag("SMTPPassword", "name"), cfg.SMTPPassword, fieldtag("SMTPPassword", "usage"))
		cmd.Flags().String(fieldtag("SMTPFrom", "name"), cfg.SMTPFrom, fieldtag("SMTPFrom", "usage"))

		// Syslog
		cmd.Flags().Bool(fieldtag("SyslogEnabled", "name"), cfg.SyslogEnabled, fieldtag("SyslogEnabled", "usage"))
		cmd.Flags().String(fieldtag("SyslogProtocol", "name"), cfg.SyslogProtocol, fieldtag("SyslogProtocol", "usage"))
		cmd.Flags().String(fieldtag("SyslogAddress", "name"), cfg.SyslogAddress, fieldtag("SyslogAddress", "usage"))
	})
}

// AddAdminAccount attaches flags pertaining to admin account actions.
func AddAdminAccount(cmd *cobra.Command) {
	name := fieldtag("AdminAccountUsername", "name")
	usage := fieldtag("AdminAccountUsername", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// AddAdminAccountPassword attaches flags pertaining to admin account password reset.
func AddAdminAccountPassword(cmd *cobra.Command) {
	AddAdminAccount(cmd)
	name := fieldtag("AdminAccountPassword", "name")
	usage := fieldtag("AdminAccountPassword", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// AddAdminAccountCreate attaches flags pertaining to admin account creation.
func AddAdminAccountCreate(cmd *cobra.Command) {
	AddAdminAccountPassword(cmd)
	name := fieldtag("AdminAccountEmail", "name")
	usage := fieldtag("AdminAccountEmail", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// AddAdminTrans attaches flags pertaining to import/export commands.
func AddAdminTrans(cmd *cobra.Command) {
	name := fieldtag("AdminTransPath", "name")
	usage := fieldtag("AdminTransPath", "usage")
	cmd.Flags().String(name, "", usage) // REQUIRED
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}
