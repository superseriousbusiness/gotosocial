package config

// GetLogLevel safely fetches the value for global configuration 'LogLevel' field
func GetLogLevel() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.LogLevel
	})
	return
}

// GetLogDbQueries safely fetches the value for global configuration 'LogDbQueries' field
func GetLogDbQueries() (v bool) {
	Config(func(cfg *Configuration) {
		v = cfg.LogDbQueries
	})
	return
}

// GetApplicationName safely fetches the value for global configuration 'ApplicationName' field
func GetApplicationName() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.ApplicationName
	})
	return
}

// GetConfigPath safely fetches the value for global configuration 'ConfigPath' field
func GetConfigPath() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.ConfigPath
	})
	return
}

// GetHost safely fetches the value for global configuration 'Host' field
func GetHost() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.Host
	})
	return
}

// GetAccountDomain safely fetches the value for global configuration 'AccountDomain' field
func GetAccountDomain() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.AccountDomain
	})
	return
}

// GetProtocol safely fetches the value for global configuration 'Protocol' field
func GetProtocol() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.Protocol
	})
	return
}

// GetBindAddress safely fetches the value for global configuration 'BindAddress' field
func GetBindAddress() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.BindAddress
	})
	return
}

// GetPort safely fetches the value for global configuration 'Port' field
func GetPort() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.Port
	})
	return
}

// GetTrustedProxies safely fetches the value for global configuration 'TrustedProxies' field
func GetTrustedProxies() (v []string) {
	Config(func(cfg *Configuration) {
		v = cfg.TrustedProxies
	})
	return
}

// GetSoftwareVersion safely fetches the value for global configuration 'SoftwareVersion' field
func GetSoftwareVersion() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.SoftwareVersion
	})
	return
}

// GetDbType safely fetches the value for global configuration 'DbType' field
func GetDbType() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.DbType
	})
	return
}

// GetDbAddress safely fetches the value for global configuration 'DbAddress' field
func GetDbAddress() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.DbAddress
	})
	return
}

// GetDbPort safely fetches the value for global configuration 'DbPort' field
func GetDbPort() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.DbPort
	})
	return
}

// GetDbUser safely fetches the value for global configuration 'DbUser' field
func GetDbUser() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.DbUser
	})
	return
}

// GetDbPassword safely fetches the value for global configuration 'DbPassword' field
func GetDbPassword() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.DbPassword
	})
	return
}

// GetDbDatabase safely fetches the value for global configuration 'DbDatabase' field
func GetDbDatabase() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.DbDatabase
	})
	return
}

// GetDbTLSMode safely fetches the value for global configuration 'DbTLSMode' field
func GetDbTLSMode() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.DbTLSMode
	})
	return
}

// GetDbTLSCACert safely fetches the value for global configuration 'DbTLSCACert' field
func GetDbTLSCACert() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.DbTLSCACert
	})
	return
}

// GetWebTemplateBaseDir safely fetches the value for global configuration 'WebTemplateBaseDir' field
func GetWebTemplateBaseDir() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.WebTemplateBaseDir
	})
	return
}

// GetWebAssetBaseDir safely fetches the value for global configuration 'WebAssetBaseDir' field
func GetWebAssetBaseDir() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.WebAssetBaseDir
	})
	return
}

// GetAccountsRegistrationOpen safely fetches the value for global configuration 'AccountsRegistrationOpen' field
func GetAccountsRegistrationOpen() (v bool) {
	Config(func(cfg *Configuration) {
		v = cfg.AccountsRegistrationOpen
	})
	return
}

// GetAccountsApprovalRequired safely fetches the value for global configuration 'AccountsApprovalRequired' field
func GetAccountsApprovalRequired() (v bool) {
	Config(func(cfg *Configuration) {
		v = cfg.AccountsApprovalRequired
	})
	return
}

// GetAccountsReasonRequired safely fetches the value for global configuration 'AccountsReasonRequired' field
func GetAccountsReasonRequired() (v bool) {
	Config(func(cfg *Configuration) {
		v = cfg.AccountsReasonRequired
	})
	return
}

// GetMediaImageMaxSize safely fetches the value for global configuration 'MediaImageMaxSize' field
func GetMediaImageMaxSize() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.MediaImageMaxSize
	})
	return
}

// GetMediaVideoMaxSize safely fetches the value for global configuration 'MediaVideoMaxSize' field
func GetMediaVideoMaxSize() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.MediaVideoMaxSize
	})
	return
}

// GetMediaDescriptionMinChars safely fetches the value for global configuration 'MediaDescriptionMinChars' field
func GetMediaDescriptionMinChars() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.MediaDescriptionMinChars
	})
	return
}

// GetMediaDescriptionMaxChars safely fetches the value for global configuration 'MediaDescriptionMaxChars' field
func GetMediaDescriptionMaxChars() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.MediaDescriptionMaxChars
	})
	return
}

// GetMediaRemoteCacheDays safely fetches the value for global configuration 'MediaRemoteCacheDays' field
func GetMediaRemoteCacheDays() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.MediaRemoteCacheDays
	})
	return
}

// GetStorageBackend safely fetches the value for global configuration 'StorageBackend' field
func GetStorageBackend() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.StorageBackend
	})
	return
}

// GetStorageLocalBasePath safely fetches the value for global configuration 'StorageLocalBasePath' field
func GetStorageLocalBasePath() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.StorageLocalBasePath
	})
	return
}

// GetStatusesMaxChars safely fetches the value for global configuration 'StatusesMaxChars' field
func GetStatusesMaxChars() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.StatusesMaxChars
	})
	return
}

// GetStatusesCWMaxChars safely fetches the value for global configuration 'StatusesCWMaxChars' field
func GetStatusesCWMaxChars() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.StatusesCWMaxChars
	})
	return
}

// GetStatusesPollMaxOptions safely fetches the value for global configuration 'StatusesPollMaxOptions' field
func GetStatusesPollMaxOptions() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.StatusesPollMaxOptions
	})
	return
}

// GetStatusesPollOptionMaxChars safely fetches the value for global configuration 'StatusesPollOptionMaxChars' field
func GetStatusesPollOptionMaxChars() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.StatusesPollOptionMaxChars
	})
	return
}

// GetStatusesMediaMaxFiles safely fetches the value for global configuration 'StatusesMediaMaxFiles' field
func GetStatusesMediaMaxFiles() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.StatusesMediaMaxFiles
	})
	return
}

// GetLetsEncryptEnabled safely fetches the value for global configuration 'LetsEncryptEnabled' field
func GetLetsEncryptEnabled() (v bool) {
	Config(func(cfg *Configuration) {
		v = cfg.LetsEncryptEnabled
	})
	return
}

// GetLetsEncryptPort safely fetches the value for global configuration 'LetsEncryptPort' field
func GetLetsEncryptPort() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.LetsEncryptPort
	})
	return
}

// GetLetsEncryptCertDir safely fetches the value for global configuration 'LetsEncryptCertDir' field
func GetLetsEncryptCertDir() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.LetsEncryptCertDir
	})
	return
}

// GetLetsEncryptEmailAddress safely fetches the value for global configuration 'LetsEncryptEmailAddress' field
func GetLetsEncryptEmailAddress() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.LetsEncryptEmailAddress
	})
	return
}

// GetOIDCEnabled safely fetches the value for global configuration 'OIDCEnabled' field
func GetOIDCEnabled() (v bool) {
	Config(func(cfg *Configuration) {
		v = cfg.OIDCEnabled
	})
	return
}

// GetOIDCIdpName safely fetches the value for global configuration 'OIDCIdpName' field
func GetOIDCIdpName() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.OIDCIdpName
	})
	return
}

// GetOIDCSkipVerification safely fetches the value for global configuration 'OIDCSkipVerification' field
func GetOIDCSkipVerification() (v bool) {
	Config(func(cfg *Configuration) {
		v = cfg.OIDCSkipVerification
	})
	return
}

// GetOIDCIssuer safely fetches the value for global configuration 'OIDCIssuer' field
func GetOIDCIssuer() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.OIDCIssuer
	})
	return
}

// GetOIDCClientID safely fetches the value for global configuration 'OIDCClientID' field
func GetOIDCClientID() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.OIDCClientID
	})
	return
}

// GetOIDCClientSecret safely fetches the value for global configuration 'OIDCClientSecret' field
func GetOIDCClientSecret() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.OIDCClientSecret
	})
	return
}

// GetOIDCScopes safely fetches the value for global configuration 'OIDCScopes' field
func GetOIDCScopes() (v []string) {
	Config(func(cfg *Configuration) {
		v = cfg.OIDCScopes
	})
	return
}

// GetSMTPHost safely fetches the value for global configuration 'SMTPHost' field
func GetSMTPHost() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.SMTPHost
	})
	return
}

// GetSMTPPort safely fetches the value for global configuration 'SMTPPort' field
func GetSMTPPort() (v int) {
	Config(func(cfg *Configuration) {
		v = cfg.SMTPPort
	})
	return
}

// GetSMTPUsername safely fetches the value for global configuration 'SMTPUsername' field
func GetSMTPUsername() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.SMTPUsername
	})
	return
}

// GetSMTPPassword safely fetches the value for global configuration 'SMTPPassword' field
func GetSMTPPassword() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.SMTPPassword
	})
	return
}

// GetSMTPFrom safely fetches the value for global configuration 'SMTPFrom' field
func GetSMTPFrom() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.SMTPFrom
	})
	return
}

// GetSyslogEnabled safely fetches the value for global configuration 'SyslogEnabled' field
func GetSyslogEnabled() (v bool) {
	Config(func(cfg *Configuration) {
		v = cfg.SyslogEnabled
	})
	return
}

// GetSyslogProtocol safely fetches the value for global configuration 'SyslogProtocol' field
func GetSyslogProtocol() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.SyslogProtocol
	})
	return
}

// GetSyslogAddress safely fetches the value for global configuration 'SyslogAddress' field
func GetSyslogAddress() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.SyslogAddress
	})
	return
}

// GetAdminAccountUsername safely fetches the value for global configuration 'AdminAccountUsername' field
func GetAdminAccountUsername() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.AdminAccountUsername
	})
	return
}

// GetAdminAccountEmail safely fetches the value for global configuration 'AdminAccountEmail' field
func GetAdminAccountEmail() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.AdminAccountEmail
	})
	return
}

// GetAdminAccountPassword safely fetches the value for global configuration 'AdminAccountPassword' field
func GetAdminAccountPassword() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.AdminAccountPassword
	})
	return
}

// GetAdminTransPath safely fetches the value for global configuration 'AdminTransPath' field
func GetAdminTransPath() (v string) {
	Config(func(cfg *Configuration) {
		v = cfg.AdminTransPath
	})
	return
}
