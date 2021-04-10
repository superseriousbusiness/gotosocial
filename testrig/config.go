package testrig

import "github.com/superseriousbusiness/gotosocial/internal/config"

// NewTestConfig returns a config initialized with test defaults
func NewTestConfig() *config.Config {
	return config.TestDefault()
}
