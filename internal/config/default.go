package config

// TestDefault returns a default config for testing
func TestDefault() *Config {
	defaults := GetTestDefaults()
	return &Config{
		LogLevel:        defaults.LogLevel,
		ApplicationName: defaults.ApplicationName,
		Host:            defaults.Host,
		Protocol:        defaults.Protocol,
		DBConfig: &DBConfig{
			Type:            defaults.DbType,
			Address:         defaults.DbAddress,
			Port:            defaults.DbPort,
			User:            defaults.DbUser,
			Password:        defaults.DbPassword,
			Database:        defaults.DbDatabase,
			ApplicationName: defaults.ApplicationName,
		},
		TemplateConfig: &TemplateConfig{
			BaseDir:      defaults.TemplateBaseDir,
			AssetBaseDir: defaults.AssetBaseDir,
		},
		AccountsConfig: &AccountsConfig{
			OpenRegistration: defaults.AccountsOpenRegistration,
			RequireApproval:  defaults.AccountsRequireApproval,
			ReasonRequired:   defaults.AccountsReasonRequired,
		},
		MediaConfig: &MediaConfig{
			MaxImageSize:        defaults.MediaMaxImageSize,
			MaxVideoSize:        defaults.MediaMaxVideoSize,
			MinDescriptionChars: defaults.MediaMinDescriptionChars,
			MaxDescriptionChars: defaults.MediaMaxDescriptionChars,
		},
		StorageConfig: &StorageConfig{
			Backend:       defaults.StorageBackend,
			BasePath:      defaults.StorageBasePath,
			ServeProtocol: defaults.StorageServeProtocol,
			ServeHost:     defaults.StorageServeHost,
			ServeBasePath: defaults.StorageServeBasePath,
		},
		StatusesConfig: &StatusesConfig{
			MaxChars:           defaults.StatusesMaxChars,
			CWMaxChars:         defaults.StatusesCWMaxChars,
			PollMaxOptions:     defaults.StatusesPollMaxOptions,
			PollOptionMaxChars: defaults.StatusesPollOptionMaxChars,
			MaxMediaFiles:      defaults.StatusesMaxMediaFiles,
		},
		LetsEncryptConfig: &LetsEncryptConfig{
			Enabled:      defaults.LetsEncryptEnabled,
			CertDir:      defaults.LetsEncryptCertDir,
			EmailAddress: defaults.LetsEncryptEmailAddress,
		},
	}
}

// Default returns a config with all default values set
func Default() *Config {
	defaults := GetDefaults()
	return &Config{
		LogLevel:        defaults.LogLevel,
		ApplicationName: defaults.ApplicationName,
		Host:            defaults.Host,
		Protocol:        defaults.Protocol,
		DBConfig: &DBConfig{
			Type:            defaults.DbType,
			Address:         defaults.DbAddress,
			Port:            defaults.DbPort,
			User:            defaults.DbUser,
			Password:        defaults.DbPassword,
			Database:        defaults.DbDatabase,
			ApplicationName: defaults.ApplicationName,
		},
		TemplateConfig: &TemplateConfig{
			BaseDir:      defaults.TemplateBaseDir,
			AssetBaseDir: defaults.AssetBaseDir,
		},
		AccountsConfig: &AccountsConfig{
			OpenRegistration: defaults.AccountsOpenRegistration,
			RequireApproval:  defaults.AccountsRequireApproval,
			ReasonRequired:   defaults.AccountsReasonRequired,
		},
		MediaConfig: &MediaConfig{
			MaxImageSize:        defaults.MediaMaxImageSize,
			MaxVideoSize:        defaults.MediaMaxVideoSize,
			MinDescriptionChars: defaults.MediaMinDescriptionChars,
			MaxDescriptionChars: defaults.MediaMaxDescriptionChars,
		},
		StorageConfig: &StorageConfig{
			Backend:       defaults.StorageBackend,
			BasePath:      defaults.StorageBasePath,
			ServeProtocol: defaults.StorageServeProtocol,
			ServeHost:     defaults.StorageServeHost,
			ServeBasePath: defaults.StorageServeBasePath,
		},
		StatusesConfig: &StatusesConfig{
			MaxChars:           defaults.StatusesMaxChars,
			CWMaxChars:         defaults.StatusesCWMaxChars,
			PollMaxOptions:     defaults.StatusesPollMaxOptions,
			PollOptionMaxChars: defaults.StatusesPollOptionMaxChars,
			MaxMediaFiles:      defaults.StatusesMaxMediaFiles,
		},
		LetsEncryptConfig: &LetsEncryptConfig{
			Enabled:      defaults.LetsEncryptEnabled,
			CertDir:      defaults.LetsEncryptCertDir,
			EmailAddress: defaults.LetsEncryptEmailAddress,
		},
	}
}

// GetDefaults returns a populated Defaults struct with most of the values set to reasonable defaults.
// Note that if you use this function, you still need to set Host and, if desired, ConfigPath.
func GetDefaults() Defaults {
	return Defaults{
		LogLevel:        "info",
		ApplicationName: "gotosocial",
		ConfigPath:      "",
		Host:            "",
		Protocol:        "https",

		DbType:     "postgres",
		DbAddress:  "localhost",
		DbPort:     5432,
		DbUser:     "postgres",
		DbPassword: "postgres",
		DbDatabase: "postgres",

		TemplateBaseDir: "./web/template/",
		AssetBaseDir:    "./web/assets/",

		AccountsOpenRegistration: true,
		AccountsRequireApproval:  true,
		AccountsReasonRequired:   true,

		MediaMaxImageSize:        2097152,  //2mb
		MediaMaxVideoSize:        10485760, //10mb
		MediaMinDescriptionChars: 0,
		MediaMaxDescriptionChars: 500,

		StorageBackend:       "local",
		StorageBasePath:      "/gotosocial/storage",
		StorageServeProtocol: "https",
		StorageServeHost:     "localhost",
		StorageServeBasePath: "/fileserver",

		StatusesMaxChars:           5000,
		StatusesCWMaxChars:         100,
		StatusesPollMaxOptions:     6,
		StatusesPollOptionMaxChars: 50,
		StatusesMaxMediaFiles:      6,

		LetsEncryptEnabled:      true,
		LetsEncryptCertDir:      "/gotosocial/storage/certs",
		LetsEncryptEmailAddress: "",
	}
}

// GetTestDefaults returns a Defaults struct with values set that are suitable for local testing.
func GetTestDefaults() Defaults {
	return Defaults{
		LogLevel:        "trace",
		ApplicationName: "gotosocial",
		ConfigPath:      "",
		Host:            "localhost:8080",
		Protocol:        "http",

		DbType:     "postgres",
		DbAddress:  "localhost",
		DbPort:     5432,
		DbUser:     "postgres",
		DbPassword: "postgres",
		DbDatabase: "postgres",

		TemplateBaseDir: "./web/template/",
		AssetBaseDir:    "./web/assets/",

		AccountsOpenRegistration: true,
		AccountsRequireApproval:  true,
		AccountsReasonRequired:   true,

		MediaMaxImageSize:        1048576, //1mb
		MediaMaxVideoSize:        5242880, //5mb
		MediaMinDescriptionChars: 0,
		MediaMaxDescriptionChars: 500,

		StorageBackend:       "local",
		StorageBasePath:      "/gotosocial/storage",
		StorageServeProtocol: "http",
		StorageServeHost:     "localhost:8080",
		StorageServeBasePath: "/fileserver",

		StatusesMaxChars:           5000,
		StatusesCWMaxChars:         100,
		StatusesPollMaxOptions:     6,
		StatusesPollOptionMaxChars: 50,
		StatusesMaxMediaFiles:      6,

		LetsEncryptEnabled:      false,
		LetsEncryptCertDir:      "",
		LetsEncryptEmailAddress: "",
	}
}
