package config

// Aliases contains aliases for all flags; this allows nested, differently-cased
// flags to be read from a configuration file via Viper.
//
// The key of the map is the command-line flag / viper key, and the value
// is the alias -- the field name used in the config file.
var Aliases = map[string]string{
	FlagNames.LogLevel:        "logLevel",
	FlagNames.ApplicationName: "applicationName",
	FlagNames.Host:            "host",
	FlagNames.AccountDomain:   "accountDomain",
	FlagNames.Protocol:        "protocol",
	FlagNames.BindAddress:     "bindAddress",
	FlagNames.Port:            "port",
	FlagNames.TrustedProxies:  "trustedProxies",
	FlagNames.SoftwareVersion: "softwareVersion",

	FlagNames.DbType:      "db.type",
	FlagNames.DbAddress:   "db.address",
	FlagNames.DbPort:      "db.port",
	FlagNames.DbUser:      "db.user",
	FlagNames.DbPassword:  "db.password",
	FlagNames.DbDatabase:  "db.database",
	FlagNames.DbTLSMode:   "db.tlsMode",
	FlagNames.DbTLSCACert: "db.tlsCACert",

	FlagNames.TemplateBaseDir: "template.baseDir",
	FlagNames.AssetBaseDir:    "template.assetBaseDir",

	FlagNames.AccountsOpenRegistration: "accounts.openRegistration",
	FlagNames.AccountsApprovalRequired: "accounts.requireApproval",
	FlagNames.AccountsReasonRequired:   "accounts.reasonRequired",

	FlagNames.MediaMaxImageSize:        "media.maxImageSize",
	FlagNames.MediaMaxVideoSize:        "media.maxVideoSize",
	FlagNames.MediaMinDescriptionChars: "media.minDescriptionChars",
	FlagNames.MediaMaxDescriptionChars: "media.maxDescriptionChars",

	FlagNames.StorageBackend:       "storage.backend",
	FlagNames.StorageBasePath:      "storage.basePath",
	FlagNames.StorageServeProtocol: "storage.serveProtocol",
	FlagNames.StorageServeHost:     "storage.serveHost",
	FlagNames.StorageServeBasePath: "storage.serveBasePath",

	FlagNames.StatusesMaxChars:           "statuses.maxChars",
	FlagNames.StatusesCWMaxChars:         "statuses.cwMaxChars",
	FlagNames.StatusesPollMaxOptions:     "statuses.pollMaxOptions",
	FlagNames.StatusesPollOptionMaxChars: "statuses.pollOptionMaxChars",
	FlagNames.StatusesMaxMediaFiles:      "statuses.maxMediaFiles",

	FlagNames.LetsEncryptEnabled:      "letsencrypt.enabled",
	FlagNames.LetsEncryptPort:         "letsencrypt.port",
	FlagNames.LetsEncryptCertDir:      "letsencrypt.certDir",
	FlagNames.LetsEncryptEmailAddress: "letsencrypt.emailAddress",

	FlagNames.OIDCEnabled:          "oidc.enabled",
	FlagNames.OIDCIdpName:          "oidc.idpName",
	FlagNames.OIDCSkipVerification: "oidc.skipVerification",
	FlagNames.OIDCIssuer:           "oidc.issuer",
	FlagNames.OIDCClientID:         "oidc.clientID",
	FlagNames.OIDCClientSecret:     "oidc.clientSecret",
	FlagNames.OIDCScopes:           "oidc.scopes",

	FlagNames.SMTPHost:     "smtp.host",
	FlagNames.SMTPPort:     "smtp.port",
	FlagNames.SMTPUsername: "smtp.username",
	FlagNames.SMTPPassword: "smtp.password",
	FlagNames.SMTPFrom:     "smtp.from",
}
