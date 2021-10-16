/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Flags and usage strings for configuration.
const (
	UsernameFlag  = "username"
	UsernameUsage = "the username to create/delete/etc"

	EmailFlag  = "email"
	EmailUsage = "the email address of this account"

	PasswordFlag  = "password"
	PasswordUsage = "the password to set for this account"

	TransPathFlag  = "path"
	TransPathUsage = "the path of the file to import from/export to"
)

// Config pulls together all the configuration needed to run gotosocial
type Config struct {
	/*
		Parseable from .yaml configuration file.
		For long-running commands (server start etc).
	*/

	LogLevel          string             `yaml:"logLevel"`
	ApplicationName   string             `yaml:"applicationName"`
	Host              string             `yaml:"host"`
	AccountDomain     string             `yaml:"accountDomain"`
	Protocol          string             `yaml:"protocol"`
	Port              int                `yaml:"port"`
	TrustedProxies    []string           `yaml:"trustedProxies"`
	DBConfig          *DBConfig          `yaml:"db"`
	TemplateConfig    *TemplateConfig    `yaml:"template"`
	AccountsConfig    *AccountsConfig    `yaml:"accounts"`
	MediaConfig       *MediaConfig       `yaml:"media"`
	StorageConfig     *StorageConfig     `yaml:"storage"`
	StatusesConfig    *StatusesConfig    `yaml:"statuses"`
	LetsEncryptConfig *LetsEncryptConfig `yaml:"letsEncrypt"`
	OIDCConfig        *OIDCConfig        `yaml:"oidc"`
	SMTPConfig        *SMTPConfig        `yaml:"smtp"`

	/*
		Not parsed from .yaml configuration file.
	*/
	AccountCLIFlags map[string]string
	ExportCLIFlags  map[string]string
	SoftwareVersion string
}

// FromFile returns a new config from a file, or an error if something goes amiss.
func FromFile(path string) (*Config, error) {
	if path != "" {
		c, err := loadFromFile(path)
		if err != nil {
			return nil, fmt.Errorf("error creating config: %s", err)
		}
		return c, nil
	}
	return Empty(), nil
}

// Empty just returns a new empty config
func Empty() *Config {
	return &Config{
		DBConfig:          &DBConfig{},
		TemplateConfig:    &TemplateConfig{},
		AccountsConfig:    &AccountsConfig{},
		MediaConfig:       &MediaConfig{},
		StorageConfig:     &StorageConfig{},
		StatusesConfig:    &StatusesConfig{},
		LetsEncryptConfig: &LetsEncryptConfig{},
		OIDCConfig:        &OIDCConfig{},
		SMTPConfig:        &SMTPConfig{},
		AccountCLIFlags:   make(map[string]string),
		ExportCLIFlags:    make(map[string]string),
	}
}

// loadFromFile takes a path to a yaml file and attempts to load a Config object from it
func loadFromFile(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file at path %s: %s", path, err)
	}

	config := Empty()
	if err := yaml.Unmarshal(bytes, config); err != nil {
		return nil, fmt.Errorf("could not unmarshal file at path %s: %s", path, err)
	}

	return config, nil
}

// ParseCLIFlags sets flags on the config using the provided Flags object
func (c *Config) ParseCLIFlags(f KeyedFlags, version string) error {
	fn := GetFlagNames()

	// For all of these flags, we only want to set them on the config if:
	//
	// a) They haven't been set at all in the config file we already parsed,
	// 	  and so we take the default from the flags object.
	//
	// b) They may have been set in the config, but they've *also* been set explicitly
	//    as a command-line argument or an env variable, which takes priority.

	// general flags
	if c.LogLevel == "" || f.IsSet(fn.LogLevel) {
		c.LogLevel = f.String(fn.LogLevel)
	}

	if c.ApplicationName == "" || f.IsSet(fn.ApplicationName) {
		c.ApplicationName = f.String(fn.ApplicationName)
	}

	if c.Host == "" || f.IsSet(fn.Host) {
		c.Host = f.String(fn.Host)
	}
	if c.Host == "" {
		return errors.New("host was not set")
	}

	if c.AccountDomain == "" || f.IsSet(fn.AccountDomain) {
		c.AccountDomain = f.String(fn.AccountDomain)
	}
	if c.AccountDomain == "" {
		c.AccountDomain = c.Host // default to whatever the host is, if this is empty
	}

	if c.Protocol == "" || f.IsSet(fn.Protocol) {
		c.Protocol = f.String(fn.Protocol)
	}
	if c.Protocol == "" {
		return errors.New("protocol was not set")
	}

	if c.Port == 0 || f.IsSet(fn.Port) {
		c.Port = f.Int(fn.Port)
	}

	if len(c.TrustedProxies) == 0 || f.IsSet(fn.TrustedProxies) {
		c.TrustedProxies = f.StringSlice(fn.TrustedProxies)
	}

	// db flags
	if c.DBConfig.Type == "" || f.IsSet(fn.DbType) {
		c.DBConfig.Type = f.String(fn.DbType)
	}

	if c.DBConfig.Address == "" || f.IsSet(fn.DbAddress) {
		c.DBConfig.Address = f.String(fn.DbAddress)
	}

	if c.DBConfig.Port == 0 || f.IsSet(fn.DbPort) {
		c.DBConfig.Port = f.Int(fn.DbPort)
	}

	if c.DBConfig.User == "" || f.IsSet(fn.DbUser) {
		c.DBConfig.User = f.String(fn.DbUser)
	}

	if c.DBConfig.Password == "" || f.IsSet(fn.DbPassword) {
		c.DBConfig.Password = f.String(fn.DbPassword)
	}

	if c.DBConfig.Database == "" || f.IsSet(fn.DbDatabase) {
		c.DBConfig.Database = f.String(fn.DbDatabase)
	}

	if c.DBConfig.TLSMode == DBTLSModeUnset || f.IsSet(fn.DbTLSMode) {
		c.DBConfig.TLSMode = DBTLSMode(f.String(fn.DbTLSMode))
	}

	if c.DBConfig.TLSCACert == "" || f.IsSet(fn.DbTLSCACert) {
		c.DBConfig.TLSCACert = f.String(fn.DbTLSCACert)
	}

	// template flags
	if c.TemplateConfig.BaseDir == "" || f.IsSet(fn.TemplateBaseDir) {
		c.TemplateConfig.BaseDir = f.String(fn.TemplateBaseDir)
	}

	// template flags
	if c.TemplateConfig.AssetBaseDir == "" || f.IsSet(fn.AssetBaseDir) {
		c.TemplateConfig.AssetBaseDir = f.String(fn.AssetBaseDir)
	}

	// accounts flags
	if f.IsSet(fn.AccountsOpenRegistration) {
		c.AccountsConfig.OpenRegistration = f.Bool(fn.AccountsOpenRegistration)
	}

	if f.IsSet(fn.AccountsApprovalRequired) {
		c.AccountsConfig.RequireApproval = f.Bool(fn.AccountsApprovalRequired)
	}

	// media flags
	if c.MediaConfig.MaxImageSize == 0 || f.IsSet(fn.MediaMaxImageSize) {
		c.MediaConfig.MaxImageSize = f.Int(fn.MediaMaxImageSize)
	}

	if c.MediaConfig.MaxVideoSize == 0 || f.IsSet(fn.MediaMaxVideoSize) {
		c.MediaConfig.MaxVideoSize = f.Int(fn.MediaMaxVideoSize)
	}

	if c.MediaConfig.MinDescriptionChars == 0 || f.IsSet(fn.MediaMinDescriptionChars) {
		c.MediaConfig.MinDescriptionChars = f.Int(fn.MediaMinDescriptionChars)
	}

	if c.MediaConfig.MaxDescriptionChars == 0 || f.IsSet(fn.MediaMaxDescriptionChars) {
		c.MediaConfig.MaxDescriptionChars = f.Int(fn.MediaMaxDescriptionChars)
	}

	// storage flags
	if c.StorageConfig.Backend == "" || f.IsSet(fn.StorageBackend) {
		c.StorageConfig.Backend = f.String(fn.StorageBackend)
	}

	if c.StorageConfig.BasePath == "" || f.IsSet(fn.StorageBasePath) {
		c.StorageConfig.BasePath = f.String(fn.StorageBasePath)
	}

	if c.StorageConfig.ServeProtocol == "" || f.IsSet(fn.StorageServeProtocol) {
		c.StorageConfig.ServeProtocol = f.String(fn.StorageServeProtocol)
	}

	if c.StorageConfig.ServeHost == "" || f.IsSet(fn.StorageServeHost) {
		c.StorageConfig.ServeHost = f.String(fn.StorageServeHost)
	}

	if c.StorageConfig.ServeBasePath == "" || f.IsSet(fn.StorageServeBasePath) {
		c.StorageConfig.ServeBasePath = f.String(fn.StorageServeBasePath)
	}

	// statuses flags
	if c.StatusesConfig.MaxChars == 0 || f.IsSet(fn.StatusesMaxChars) {
		c.StatusesConfig.MaxChars = f.Int(fn.StatusesMaxChars)
	}
	if c.StatusesConfig.CWMaxChars == 0 || f.IsSet(fn.StatusesCWMaxChars) {
		c.StatusesConfig.CWMaxChars = f.Int(fn.StatusesCWMaxChars)
	}
	if c.StatusesConfig.PollMaxOptions == 0 || f.IsSet(fn.StatusesPollMaxOptions) {
		c.StatusesConfig.PollMaxOptions = f.Int(fn.StatusesPollMaxOptions)
	}
	if c.StatusesConfig.PollOptionMaxChars == 0 || f.IsSet(fn.StatusesPollOptionMaxChars) {
		c.StatusesConfig.PollOptionMaxChars = f.Int(fn.StatusesPollOptionMaxChars)
	}
	if c.StatusesConfig.MaxMediaFiles == 0 || f.IsSet(fn.StatusesMaxMediaFiles) {
		c.StatusesConfig.MaxMediaFiles = f.Int(fn.StatusesMaxMediaFiles)
	}

	// letsencrypt flags
	if f.IsSet(fn.LetsEncryptEnabled) {
		c.LetsEncryptConfig.Enabled = f.Bool(fn.LetsEncryptEnabled)
	}

	if c.LetsEncryptConfig.Port == 0 || f.IsSet(fn.LetsEncryptPort) {
		c.LetsEncryptConfig.Port = f.Int(fn.LetsEncryptPort)
	}

	if c.LetsEncryptConfig.CertDir == "" || f.IsSet(fn.LetsEncryptCertDir) {
		c.LetsEncryptConfig.CertDir = f.String(fn.LetsEncryptCertDir)
	}

	if c.LetsEncryptConfig.EmailAddress == "" || f.IsSet(fn.LetsEncryptEmailAddress) {
		c.LetsEncryptConfig.EmailAddress = f.String(fn.LetsEncryptEmailAddress)
	}

	// OIDC flags
	if f.IsSet(fn.OIDCEnabled) {
		c.OIDCConfig.Enabled = f.Bool(fn.OIDCEnabled)
	}

	if c.OIDCConfig.IDPName == "" || f.IsSet(fn.OIDCIdpName) {
		c.OIDCConfig.IDPName = f.String(fn.OIDCIdpName)
	}

	if f.IsSet(fn.OIDCSkipVerification) {
		c.OIDCConfig.SkipVerification = f.Bool(fn.OIDCSkipVerification)
	}

	if c.OIDCConfig.Issuer == "" || f.IsSet(fn.OIDCIssuer) {
		c.OIDCConfig.Issuer = f.String(fn.OIDCIssuer)
	}

	if c.OIDCConfig.ClientID == "" || f.IsSet(fn.OIDCClientID) {
		c.OIDCConfig.ClientID = f.String(fn.OIDCClientID)
	}

	if c.OIDCConfig.ClientSecret == "" || f.IsSet(fn.OIDCClientSecret) {
		c.OIDCConfig.ClientSecret = f.String(fn.OIDCClientSecret)
	}

	if len(c.OIDCConfig.Scopes) == 0 || f.IsSet(fn.OIDCScopes) {
		c.OIDCConfig.Scopes = f.StringSlice(fn.OIDCScopes)
	}

	// smtp flags
	if c.SMTPConfig.Host == "" || f.IsSet(fn.SMTPHost) {
		c.SMTPConfig.Host = f.String(fn.SMTPHost)
	}

	if c.SMTPConfig.Port == 0 || f.IsSet(fn.SMTPPort) {
		c.SMTPConfig.Port = f.Int(fn.SMTPPort)
	}

	if c.SMTPConfig.Username == "" || f.IsSet(fn.SMTPUsername) {
		c.SMTPConfig.Username = f.String(fn.SMTPUsername)
	}

	if c.SMTPConfig.Password == "" || f.IsSet(fn.SMTPPassword) {
		c.SMTPConfig.Password = f.String(fn.SMTPPassword)
	}

	if c.SMTPConfig.From == "" || f.IsSet(fn.SMTPFrom) {
		c.SMTPConfig.From = f.String(fn.SMTPFrom)
	}

	// command-specific flags

	// admin account CLI flags
	c.AccountCLIFlags[UsernameFlag] = f.String(UsernameFlag)
	c.AccountCLIFlags[EmailFlag] = f.String(EmailFlag)
	c.AccountCLIFlags[PasswordFlag] = f.String(PasswordFlag)

	// export CLI flags
	c.ExportCLIFlags[TransPathFlag] = f.String(TransPathFlag)

	c.SoftwareVersion = version
	return nil
}

// KeyedFlags is a wrapper for any type that can store keyed flags and give them back.
// HINT: This works with a urfave cli context struct ;)
type KeyedFlags interface {
	Bool(k string) bool
	String(k string) string
	StringSlice(k string) []string
	Int(k string) int
	IsSet(k string) bool
}

// Flags is used for storing the names of the various flags used for
// initializing and storing urfavecli flag variables.
type Flags struct {
	LogLevel        string
	ApplicationName string
	ConfigPath      string
	Host            string
	AccountDomain   string
	Protocol        string
	Port            string
	TrustedProxies  string

	DbType      string
	DbAddress   string
	DbPort      string
	DbUser      string
	DbPassword  string
	DbDatabase  string
	DbTLSMode   string
	DbTLSCACert string

	TemplateBaseDir string
	AssetBaseDir    string

	AccountsOpenRegistration string
	AccountsApprovalRequired string
	AccountsReasonRequired   string

	MediaMaxImageSize        string
	MediaMaxVideoSize        string
	MediaMinDescriptionChars string
	MediaMaxDescriptionChars string

	StorageBackend       string
	StorageBasePath      string
	StorageServeProtocol string
	StorageServeHost     string
	StorageServeBasePath string

	StatusesMaxChars           string
	StatusesCWMaxChars         string
	StatusesPollMaxOptions     string
	StatusesPollOptionMaxChars string
	StatusesMaxMediaFiles      string

	LetsEncryptEnabled      string
	LetsEncryptCertDir      string
	LetsEncryptEmailAddress string
	LetsEncryptPort         string

	OIDCEnabled          string
	OIDCIdpName          string
	OIDCSkipVerification string
	OIDCIssuer           string
	OIDCClientID         string
	OIDCClientSecret     string
	OIDCScopes           string

	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

// Defaults contains all the default values for a gotosocial config
type Defaults struct {
	LogLevel        string
	ApplicationName string
	ConfigPath      string
	Host            string
	AccountDomain   string
	Protocol        string
	Port            int
	TrustedProxies  []string
	SoftwareVersion string

	DbType      string
	DbAddress   string
	DbPort      int
	DbUser      string
	DbPassword  string
	DbDatabase  string
	DBTlsMode   string
	DBTlsCACert string

	TemplateBaseDir string
	AssetBaseDir    string

	AccountsOpenRegistration bool
	AccountsRequireApproval  bool
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

// GetFlagNames returns a struct containing the names of the various flags used for
// initializing and storing urfavecli flag variables.
func GetFlagNames() Flags {
	return Flags{
		LogLevel:        "log-level",
		ApplicationName: "application-name",
		ConfigPath:      "config-path",
		Host:            "host",
		AccountDomain:   "account-domain",
		Protocol:        "protocol",
		Port:            "port",
		TrustedProxies:  "trusted-proxies",

		DbType:      "db-type",
		DbAddress:   "db-address",
		DbPort:      "db-port",
		DbUser:      "db-user",
		DbPassword:  "db-password",
		DbDatabase:  "db-database",
		DbTLSMode:   "db-tls-mode",
		DbTLSCACert: "db-tls-ca-cert",

		TemplateBaseDir: "template-basedir",
		AssetBaseDir:    "asset-basedir",

		AccountsOpenRegistration: "accounts-open-registration",
		AccountsApprovalRequired: "accounts-approval-required",
		AccountsReasonRequired:   "accounts-reason-required",

		MediaMaxImageSize:        "media-max-image-size",
		MediaMaxVideoSize:        "media-max-video-size",
		MediaMinDescriptionChars: "media-min-description-chars",
		MediaMaxDescriptionChars: "media-max-description-chars",

		StorageBackend:       "storage-backend",
		StorageBasePath:      "storage-base-path",
		StorageServeProtocol: "storage-serve-protocol",
		StorageServeHost:     "storage-serve-host",
		StorageServeBasePath: "storage-serve-base-path",

		StatusesMaxChars:           "statuses-max-chars",
		StatusesCWMaxChars:         "statuses-cw-max-chars",
		StatusesPollMaxOptions:     "statuses-poll-max-options",
		StatusesPollOptionMaxChars: "statuses-poll-option-max-chars",
		StatusesMaxMediaFiles:      "statuses-max-media-files",

		LetsEncryptEnabled:      "letsencrypt-enabled",
		LetsEncryptPort:         "letsencrypt-port",
		LetsEncryptCertDir:      "letsencrypt-cert-dir",
		LetsEncryptEmailAddress: "letsencrypt-email",

		OIDCEnabled:          "oidc-enabled",
		OIDCIdpName:          "oidc-idp-name",
		OIDCSkipVerification: "oidc-skip-verification",
		OIDCIssuer:           "oidc-issuer",
		OIDCClientID:         "oidc-client-id",
		OIDCClientSecret:     "oidc-client-secret",
		OIDCScopes:           "oidc-scopes",

		SMTPHost:     "smtp-host",
		SMTPPort:     "smtp-port",
		SMTPUsername: "smtp-username",
		SMTPPassword: "smtp-password",
		SMTPFrom:     "smtp-from",
	}
}

// GetEnvNames returns a struct containing the names of the environment variable keys used for
// initializing and storing urfavecli flag variables.
func GetEnvNames() Flags {
	return Flags{
		LogLevel:        "GTS_LOG_LEVEL",
		ApplicationName: "GTS_APPLICATION_NAME",
		ConfigPath:      "GTS_CONFIG_PATH",
		Host:            "GTS_HOST",
		AccountDomain:   "GTS_ACCOUNT_DOMAIN",
		Protocol:        "GTS_PROTOCOL",
		Port:            "GTS_PORT",
		TrustedProxies:  "GTS_TRUSTED_PROXIES",

		DbType:      "GTS_DB_TYPE",
		DbAddress:   "GTS_DB_ADDRESS",
		DbPort:      "GTS_DB_PORT",
		DbUser:      "GTS_DB_USER",
		DbPassword:  "GTS_DB_PASSWORD",
		DbDatabase:  "GTS_DB_DATABASE",
		DbTLSMode:   "GTS_DB_TLS_MODE",
		DbTLSCACert: "GTS_DB_CA_CERT",

		TemplateBaseDir: "GTS_TEMPLATE_BASEDIR",
		AssetBaseDir:    "GTS_ASSET_BASEDIR",

		AccountsOpenRegistration: "GTS_ACCOUNTS_OPEN_REGISTRATION",
		AccountsApprovalRequired: "GTS_ACCOUNTS_APPROVAL_REQUIRED",
		AccountsReasonRequired:   "GTS_ACCOUNTS_REASON_REQUIRED",

		MediaMaxImageSize:        "GTS_MEDIA_MAX_IMAGE_SIZE",
		MediaMaxVideoSize:        "GTS_MEDIA_MAX_VIDEO_SIZE",
		MediaMinDescriptionChars: "GTS_MEDIA_MIN_DESCRIPTION_CHARS",
		MediaMaxDescriptionChars: "GTS_MEDIA_MAX_DESCRIPTION_CHARS",

		StorageBackend:       "GTS_STORAGE_BACKEND",
		StorageBasePath:      "GTS_STORAGE_BASE_PATH",
		StorageServeProtocol: "GTS_STORAGE_SERVE_PROTOCOL",
		StorageServeHost:     "GTS_STORAGE_SERVE_HOST",
		StorageServeBasePath: "GTS_STORAGE_SERVE_BASE_PATH",

		StatusesMaxChars:           "GTS_STATUSES_MAX_CHARS",
		StatusesCWMaxChars:         "GTS_STATUSES_CW_MAX_CHARS",
		StatusesPollMaxOptions:     "GTS_STATUSES_POLL_MAX_OPTIONS",
		StatusesPollOptionMaxChars: "GTS_STATUSES_POLL_OPTION_MAX_CHARS",
		StatusesMaxMediaFiles:      "GTS_STATUSES_MAX_MEDIA_FILES",

		LetsEncryptEnabled:      "GTS_LETSENCRYPT_ENABLED",
		LetsEncryptPort:         "GTS_LETSENCRYPT_PORT",
		LetsEncryptCertDir:      "GTS_LETSENCRYPT_CERT_DIR",
		LetsEncryptEmailAddress: "GTS_LETSENCRYPT_EMAIL",

		OIDCEnabled:          "GTS_OIDC_ENABLED",
		OIDCIdpName:          "GTS_OIDC_IDP_NAME",
		OIDCSkipVerification: "GTS_OIDC_SKIP_VERIFICATION",
		OIDCIssuer:           "GTS_OIDC_ISSUER",
		OIDCClientID:         "GTS_OIDC_CLIENT_ID",
		OIDCClientSecret:     "GTS_OIDC_CLIENT_SECRET",
		OIDCScopes:           "GTS_OIDC_SCOPES",

		SMTPHost:     "SMTP_HOST",
		SMTPPort:     "SMTP_PORT",
		SMTPUsername: "SMTP_USERNAME",
		SMTPPassword: "SMTP_PASSWORD",
		SMTPFrom:     "SMTP_FROM",
	}
}
