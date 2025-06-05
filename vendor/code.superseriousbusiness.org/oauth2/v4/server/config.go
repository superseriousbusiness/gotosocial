package server

import (
	"net/http"
	"time"

	"code.superseriousbusiness.org/oauth2/v4"
)

// Config configuration parameters
type Config struct {
	// token type
	TokenType string

	// to allow GET requests for the token
	AllowGetAccessRequest bool

	// allow the authorization type
	AllowedResponseTypes []oauth2.ResponseType

	// allow the grant type
	AllowedGrantTypes []oauth2.GrantType

	// Allowed values for "code_challenge_method".
	AllowedCodeChallengeMethods []oauth2.CodeChallengeMethod

	// Default to fall back to
	// if "code_challenge_method"
	// was not set in the request.
	DefaultCodeChallengeMethod oauth2.CodeChallengeMethod

	ForcePKCE bool
}

// NewConfig create to configuration instance
func NewConfig() *Config {
	return &Config{
		TokenType:            "Bearer",
		AllowedResponseTypes: []oauth2.ResponseType{oauth2.Code, oauth2.Token},
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
			oauth2.PasswordCredentials,
			oauth2.ClientCredentials,
			oauth2.Refreshing,
		},
		AllowedCodeChallengeMethods: []oauth2.CodeChallengeMethod{
			oauth2.CodeChallengePlain,
			oauth2.CodeChallengeS256,
		},
	}
}

// AuthorizeRequest authorization request
type AuthorizeRequest struct {
	ResponseType        oauth2.ResponseType
	ClientID            string
	Scope               string
	RedirectURI         string
	State               string
	UserID              string
	CodeChallenge       string
	CodeChallengeMethod oauth2.CodeChallengeMethod
	AccessTokenExp      time.Duration
	Request             *http.Request
}
