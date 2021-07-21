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

package oidc

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sirupsen/logrus"
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
	HandleCallback(ctx context.Context, state string, code string) (*Claims, error)
}

type idp struct {
	oauth2Config oauth2.Config
	provider     *oidc.Provider
	oidcConf     *oidc.Config
	log          *logrus.Logger
}

func NewIDP(config *config.Config, log *logrus.Logger) (IDP, error) {

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
	if config.OIDCConfig.Issuer == "" {
		return nil, fmt.Errorf("not set: Issuer")
	}
	if config.OIDCConfig.ClientID == "" {
		return nil, fmt.Errorf("not set: ClientID")
	}
	if config.OIDCConfig.ClientSecret == "" {
		return nil, fmt.Errorf("not set: ClientSecret")
	}
	if len(config.OIDCConfig.Scopes) == 0 {
		return nil, fmt.Errorf("not set: Scopes")
	}

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

	// create a config for verifier creation
	oidcConf := &oidc.Config{
		ClientID: config.OIDCConfig.ClientID,
	}
	if config.OIDCConfig.SkipVerification {
		oidcConf.SkipClientIDCheck = true
		oidcConf.SkipExpiryCheck = true
		oidcConf.SkipIssuerCheck = true
	}

	return &idp{
		oauth2Config: oauth2Config,
		oidcConf:     oidcConf,
		provider:     provider,
		log:          log,
	}, nil
}
