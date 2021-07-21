package oidc

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"golang.org/x/oauth2"
)

const (
	// CallbackPath is the API path for receiving callback tokens from external OIDC providers
	CallbackPath = "/auth/callback"
	profileScope = "profile"
	emailScope   = "email"
	groupsScope  = "groups"
)

type IDP interface {
}

type idp struct {
	oauth2Config    oauth2.Config
	provider        *oidc.Provider
	idTokenVerifier *oidc.IDTokenVerifier
}

func NewIDP(config *config.Config) (IDP, error) {

	// oidc isn't enabled so we don't need to do anything
	if !config.OIDCConfig.Enabled {
		return nil, nil
	}

	// validate config fields
	if config.OIDCConfig.IDPID == "" {
		return nil, fmt.Errorf("not set: IDPID")
	}

	if config.OIDCConfig.IDPName == "" {
		return nil, fmt.Errorf("not set: IDPName")
	}

	aaaaaaaaaaaaaaaaaaaaaaaaaaaa


	provider, err := oidc.NewProvider(context.Background(), config.OIDCConfig.Issuer)
	if err != nil {
		return nil, err
	}

	oauth2Config := oauth2.Config{
		// client_id and client_secret of the client.
		ClientID:     config.OIDCConfig.ClientID,
		ClientSecret: config.OIDCConfig.ClientSecret,

		// The redirectURL.
		RedirectURL: fmt.Sprintf("%s://%s%s", config.Protocol, config.Host, CallbackPath),

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		//
		// Other scopes, such as "groups" can be requested.
		Scopes: config.OIDCConfig.Scopes,
	}

	idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: config.OIDCConfig.ClientID})

	return &idp{
		oauth2Config:    oauth2Config,
		idTokenVerifier: idTokenVerifier,
	}, nil
}
