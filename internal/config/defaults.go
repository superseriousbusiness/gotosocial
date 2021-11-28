package config

import "github.com/coreos/go-oidc/v3/oidc"

// Values contains all the default values for a gotosocial config
type Values struct {
	LogLevel        string
	ApplicationName string
	ConfigPath      string
	Host            string
	AccountDomain   string
	Protocol        string
	BindAddress     string
	Port            int
	TrustedProxies  []string
	SoftwareVersion string

	DbType      string
	DbAddress   string
	DbPort      int
	DbUser      string
	DbPassword  string
	DbDatabase  string
	DbTLSMode   string
	DbTLSCACert string

	TemplateBaseDir string
	AssetBaseDir    string

	AccountsOpenRegistration bool
	AccountsApprovalRequired bool
	AccountsReasonRequired   bool

	MediaMaxImageSize        int
	MediaMaxVideoSize        int
	MediaMinDescriptionChars int
	MediaMaxDescriptionChars int

	StorageBackend       string
	StorageBasePath      string
	StorageServeProtocol string
	StorageServeHost     string
	StorageServeBasePath string

	StatusesMaxChars           int
	StatusesCWMaxChars         int
	StatusesPollMaxOptions     int
	StatusesPollOptionMaxChars int
	StatusesMaxMediaFiles      int

	LetsEncryptEnabled      bool
	LetsEncryptCertDir      string
	LetsEncryptEmailAddress string
	LetsEncryptPort         int

	OIDCEnabled          bool
	OIDCIdpName          string
	OIDCSkipVerification bool
	OIDCIssuer           string
	OIDCClientID         string
	OIDCClientSecret     string
	OIDCScopes           []string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

// Defaults returns a populated Defaults struct with most of the values set to reasonable defaults.
// Note that if you use this, you still need to set Host and, if desired, ConfigPath.
var Defaults = Values{
	LogLevel:        "info",
	ApplicationName: "gotosocial",
	ConfigPath:      "",
	Host:            "",
	AccountDomain:   "",
	Protocol:        "https",
	BindAddress:     "0.0.0.0",
	Port:            8080,
	TrustedProxies:  []string{"127.0.0.1/32"}, // localhost

	DbType:      "postgres",
	DbAddress:   "localhost",
	DbPort:      5432,
	DbUser:      "postgres",
	DbPassword:  "postgres",
	DbDatabase:  "postgres",
	DbTLSMode:   "disable",
	DbTLSCACert: "",

	TemplateBaseDir: "./web/template/",
	AssetBaseDir:    "./web/assets/",

	AccountsOpenRegistration: true,
	AccountsApprovalRequired: true,
	AccountsReasonRequired:   true,

	MediaMaxImageSize:        2097152,  // 2mb
	MediaMaxVideoSize:        10485760, // 10mb
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
	LetsEncryptPort:         80,
	LetsEncryptCertDir:      "/gotosocial/storage/certs",
	LetsEncryptEmailAddress: "",

	OIDCEnabled:          false,
	OIDCIdpName:          "",
	OIDCSkipVerification: false,
	OIDCIssuer:           "",
	OIDCClientID:         "",
	OIDCClientSecret:     "",
	OIDCScopes:           []string{oidc.ScopeOpenID, "profile", "email", "groups"},

	SMTPHost:     "",
	SMTPPort:     0,
	SMTPUsername: "",
	SMTPPassword: "",
	SMTPFrom:     "GoToSocial",
}

// TestDefaults returns a Defaults struct with values set that are suitable for local testing.
var TestDefaults = Values{
	LogLevel:        "trace",
	ApplicationName: "gotosocial",
	ConfigPath:      "",
	Host:            "localhost:8080",
	AccountDomain:   "localhost:8080",
	Protocol:        "http",
	BindAddress:     "127.0.0.1",
	Port:            8080,
	TrustedProxies:  []string{"127.0.0.1/32"},

	DbType:     "sqlite",
	DbAddress:  ":memory:",
	DbPort:     5432,
	DbUser:     "postgres",
	DbPassword: "postgres",
	DbDatabase: "postgres",

	TemplateBaseDir: "./web/template/",
	AssetBaseDir:    "./web/assets/",

	AccountsOpenRegistration: true,
	AccountsApprovalRequired: true,
	AccountsReasonRequired:   true,

	MediaMaxImageSize:        1048576, // 1mb
	MediaMaxVideoSize:        5242880, // 5mb
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
	LetsEncryptPort:         0,
	LetsEncryptCertDir:      "",
	LetsEncryptEmailAddress: "",

	OIDCEnabled:          false,
	OIDCIdpName:          "",
	OIDCSkipVerification: false,
	OIDCIssuer:           "",
	OIDCClientID:         "",
	OIDCClientSecret:     "",
	OIDCScopes:           []string{oidc.ScopeOpenID, "profile", "email", "groups"},

	SMTPHost:     "",
	SMTPPort:     0,
	SMTPUsername: "",
	SMTPPassword: "",
	SMTPFrom:     "GoToSocial",
}
