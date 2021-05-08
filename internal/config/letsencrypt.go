package config

// LetsEncryptConfig wraps everything needed to manage letsencrypt certificates from within gotosocial.
type LetsEncryptConfig struct {
	// Should letsencrypt certificate fetching be enabled?
	Enabled bool
	// Where should certificates be stored?
	CertDir string
	// Email address to pass to letsencrypt for notifications about certificate expiry etc.
	EmailAddress string
}
