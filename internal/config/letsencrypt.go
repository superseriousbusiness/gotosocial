package config

// LetsEncryptConfig wraps everything needed to manage letsencrypt certificates from within gotosocial.
type LetsEncryptConfig struct {
	// Should letsencrypt certificate fetching be enabled?
	Enabled bool `yaml:"enabled"`
	// What port should the server listen for letsencrypt challenges on?
	Port int `yaml:"port"`
	// Where should certificates be stored?
	CertDir string `yaml:"certDir"`
	// Email address to pass to letsencrypt for notifications about certificate expiry etc.
	EmailAddress string `yaml:"emailAddress"`
}
