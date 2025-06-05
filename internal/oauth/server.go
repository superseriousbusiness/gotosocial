// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/oauth2/v4"
	oautherr "code.superseriousbusiness.org/oauth2/v4/errors"
	"code.superseriousbusiness.org/oauth2/v4/manage"
	"code.superseriousbusiness.org/oauth2/v4/server"
	errorsv2 "codeberg.org/gruf/go-errors/v2"
)

const (
	// SessionAuthorizedToken is the key set in the gin context for the Token
	// of a User who has successfully passed Bearer token authorization.
	// The interface returned from grabbing this key should be parsed as oauth2.TokenInfo
	SessionAuthorizedToken = "authorized_token"
	// SessionAuthorizedUser is the key set in the gin context for the id of
	// a User who has successfully passed Bearer token authorization.
	// The interface returned from grabbing this key should be parsed as a *gtsmodel.User
	SessionAuthorizedUser = "authorized_user"
	// SessionAuthorizedAccount is the key set in the gin context for the Account
	// of a User who has successfully passed Bearer token authorization.
	// The interface returned from grabbing this key should be parsed as a *gtsmodel.Account
	SessionAuthorizedAccount = "authorized_account"
	// SessionAuthorizedApplication is the key set in the gin context for the Application
	// of a Client who has successfully passed Bearer token authorization.
	// The interface returned from grabbing this key should be parsed as a *gtsmodel.Application
	SessionAuthorizedApplication = "authorized_app"
	// OOBURI is the out-of-band oauth token uri
	OOBURI = "urn:ietf:wg:oauth:2.0:oob"
	// OOBTokenPath is the path to redirect out-of-band token requests to.
	OOBTokenPath = "/oauth/oob" // #nosec G101 else we get a hardcoded credentials warning
	// HelpfulAdvice is a handy hint to users;
	// particularly important during the login flow
	HelpfulAdvice      = "If you arrived at this error during a sign in/oauth flow, please try clearing your session cookies and signing in again; if problems persist, make sure you're using the correct credentials"
	HelpfulAdviceGrant = "If you arrived at this error during a sign in/oauth flow, your client is trying to use an unsupported OAuth grant type. Supported grant types are: authorization_code, client_credentials; please reach out to developer of your client"
)

// Server wraps some oauth2 server functions
// in an interface, exposing only what is needed.
type Server interface {
	HandleTokenRequest(r *http.Request) (map[string]interface{}, gtserror.WithCode)
	HandleAuthorizeRequest(w http.ResponseWriter, r *http.Request) gtserror.WithCode
	ValidationBearerToken(r *http.Request) (oauth2.TokenInfo, error)
	GenerateUserAccessToken(ctx context.Context, ti oauth2.TokenInfo, clientSecret string, userID string) (accessToken oauth2.TokenInfo, err error)
	LoadAccessToken(ctx context.Context, access string) (accessToken oauth2.TokenInfo, err error)
	RevokeAccessToken(ctx context.Context, clientID string, clientSecret string, access string) gtserror.WithCode
}

// s fulfils the Server interface
// using the underlying oauth2 server.
type s struct {
	server *server.Server
}

// New returns a new oauth server that implements the Server interface
func New(
	ctx context.Context,
	state *state.State,
	validateURIHandler manage.ValidateURIHandler,
	clientScopeHandler server.ClientScopeHandler,
	authorizeScopeHandler server.AuthorizeScopeHandler,
	internalErrorHandler server.InternalErrorHandler,
	responseErrorHandler server.ResponseErrorHandler,
	userAuthorizationHandler server.UserAuthorizationHandler,
) Server {
	ts := newTokenStore(ctx, state)
	cs := NewClientStore(state)

	// Set up OAuth2 manager.
	manager := manage.NewDefaultManager()
	manager.SetValidateURIHandler(validateURIHandler)
	manager.MapTokenStorage(ts)
	manager.MapClientStorage(cs)
	manager.SetAuthorizeCodeTokenCfg(
		&manage.Config{
			// Following the Mastodon API,
			// access tokens don't expire.
			AccessTokenExp: 0,
			// Don't use refresh tokens.
			IsGenerateRefresh: false,
		},
	)

	// Set up OAuth2 server.
	srv := server.NewServer(
		&server.Config{
			TokenType: "Bearer",
			// Must follow the spec.
			AllowGetAccessRequest: false,
			// Support only the non-implicit flow.
			AllowedResponseTypes: []oauth2.ResponseType{oauth2.Code},
			// Allow:
			// - Authorization Code (for first & third parties)
			// - Client Credentials (for applications)
			AllowedGrantTypes: []oauth2.GrantType{
				oauth2.AuthorizationCode,
				oauth2.ClientCredentials,
			},
			AllowedCodeChallengeMethods: []oauth2.CodeChallengeMethod{
				oauth2.CodeChallengeS256,
			},
			DefaultCodeChallengeMethod: oauth2.CodeChallengeS256,
		},
		manager,
	)
	srv.SetAuthorizeScopeHandler(authorizeScopeHandler)
	srv.SetClientScopeHandler(clientScopeHandler)
	srv.SetInternalErrorHandler(internalErrorHandler)
	srv.SetResponseErrorHandler(responseErrorHandler)
	srv.SetUserAuthorizationHandler(userAuthorizationHandler)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	return &s{srv}
}

// HandleTokenRequest wraps the oauth2 library's HandleTokenRequest function,
// providing some custom error handling (with more informative messages),
// and a slightly different token serialization format.
func (s *s) HandleTokenRequest(r *http.Request) (map[string]interface{}, gtserror.WithCode) {
	ctx := r.Context()

	gt, tgr, err := s.server.ValidationTokenRequest(r)
	if err != nil {
		help := fmt.Sprintf("could not validate token request: %s", err)
		adv := HelpfulAdvice
		if errors.Is(err, oautherr.ErrUnsupportedGrantType) {
			adv = HelpfulAdviceGrant
		}
		return nil, gtserror.NewErrorBadRequest(err, help, adv)
	}

	// Get access token + do our own nicer error handling.
	ti, err := s.server.GetAccessToken(ctx, gt, tgr)
	switch {
	case err == nil:
		// No problem.
		break

	case errors.Is(err, oautherr.ErrInvalidScope):
		help := fmt.Sprintf("requested scope %s was not covered by client scope", tgr.Scope)
		return nil, gtserror.NewErrorForbidden(err, help, HelpfulAdvice)

	case errors.Is(err, oautherr.ErrInvalidRedirectURI):
		help := fmt.Sprintf("requested redirect URI %s was not covered by client redirect URIs", tgr.RedirectURI)
		return nil, gtserror.NewErrorForbidden(err, help, HelpfulAdvice)

	default:
		help := fmt.Sprintf("could not get access token: %v", err)
		return nil, gtserror.NewErrorBadRequest(err, help, HelpfulAdvice)
	}

	// Wrangle data a bit.
	data := s.server.GetTokenData(ti)

	// Add created_at for Mastodon API compatibility.
	data["created_at"] = ti.GetAccessCreateAt().Unix()

	// If expires_in is 0 or less, omit it
	// from serialization so that clients don't
	// interpret the token as already expired.
	if expiresInI, ok := data["expires_in"]; ok {
		// This will panic if expiresIn is
		// not an int64, which is what we want.
		if expiresInI.(int64) <= 0 {
			delete(data, "expires_in")
		}
	}

	return data, nil
}

func (s *s) errorOrRedirect(err error, w http.ResponseWriter, req *server.AuthorizeRequest) gtserror.WithCode {
	if req == nil {
		return gtserror.NewErrorUnauthorized(err, HelpfulAdvice)
	}

	data, _, _ := s.server.GetErrorData(err)
	uri, err := s.server.GetRedirectURI(req, data)
	if err != nil {
		return gtserror.NewErrorInternalError(err, HelpfulAdvice)
	}

	w.Header().Set("Location", uri)
	w.WriteHeader(http.StatusFound)
	return nil
}

// HandleAuthorizeRequest wraps the oauth2 library's HandleAuthorizeRequest function
func (s *s) HandleAuthorizeRequest(w http.ResponseWriter, r *http.Request) gtserror.WithCode {
	ctx := r.Context()

	req, err := s.server.ValidationAuthorizeRequest(r)
	if err != nil {
		return s.errorOrRedirect(err, w, req)
	}

	// user authorization
	userID, err := s.server.UserAuthorizationHandler(w, r)
	if err != nil {
		return s.errorOrRedirect(err, w, req)
	}
	if userID == "" {
		help := "userID was empty"
		return gtserror.NewErrorUnauthorized(err, help, HelpfulAdvice)
	}
	req.UserID = userID

	// Specify the scope of authorization.
	if fn := s.server.AuthorizeScopeHandler; fn != nil {
		scope, err := fn(w, r)
		if err != nil {
			return s.errorOrRedirect(err, w, req)
		} else if scope != "" {
			req.Scope = scope
		}
	}

	// Specify the expiration time of access token.
	if fn := s.server.AccessTokenExpHandler; fn != nil {
		exp, err := fn(w, r)
		if err != nil {
			return s.errorOrRedirect(err, w, req)
		}
		req.AccessTokenExp = exp
	}

	ti, err := s.server.GetAuthorizeToken(ctx, req)
	if err != nil {
		return s.errorOrRedirect(err, w, req)
	}

	// If the redirect URI is empty, use the
	// first of the client's redirect URIs.
	if req.RedirectURI == "" {
		client, err := s.server.Manager.GetClient(ctx, req.ClientID)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			// Real error.
			err := gtserror.Newf("db error getting application with client id %s: %w", req.ClientID, err)
			return gtserror.NewErrorInternalError(err)
		}

		if util.IsNil(client) {
			// Application just not found.
			return gtserror.NewErrorUnauthorized(err, HelpfulAdvice)
		}

		// This will panic if client is not a
		// *gtsmodel.Application, which is what we want.
		req.RedirectURI = client.(*gtsmodel.Application).RedirectURIs[0]
	}

	uri, err := s.server.GetRedirectURI(req, s.server.GetAuthorizeData(req.ResponseType, ti))
	if err != nil {
		return gtserror.NewErrorUnauthorized(err, HelpfulAdvice)
	}

	if strings.Contains(uri, OOBURI) {
		w.Header().Set("Location", strings.ReplaceAll(uri, OOBURI, OOBTokenPath))
	} else {
		w.Header().Set("Location", uri)
	}

	w.WriteHeader(http.StatusFound)
	return nil
}

// ValidationBearerToken wraps the oauth2 library's ValidationBearerToken function
func (s *s) ValidationBearerToken(r *http.Request) (oauth2.TokenInfo, error) {
	return s.server.ValidationBearerToken(r)
}

// GenerateUserAccessToken shortcuts the normal oauth flow to create an user-level
// bearer token *without* requiring that user to log in. This is useful when we
// need to create a token for new users who haven't validated their email or logged in yet.
//
// The ti parameter refers to an existing Application token that was used to make the upstream
// request. This token needs to be validated and exist in database in order to create a new token.
func (s *s) GenerateUserAccessToken(ctx context.Context, ti oauth2.TokenInfo, clientSecret string, userID string) (oauth2.TokenInfo, error) {
	authToken, err := s.server.Manager.GenerateAuthToken(ctx, oauth2.Code, &oauth2.TokenGenerateRequest{
		ClientID:     ti.GetClientID(),
		ClientSecret: clientSecret,
		UserID:       userID,
		RedirectURI:  ti.GetRedirectURI(),
		Scope:        ti.GetScope(),
	})
	if err != nil {
		return nil, fmt.Errorf("error generating auth token: %s", err)
	}
	if authToken == nil {
		return nil, errors.New("generated auth token was empty")
	}
	log.Tracef(ctx, "obtained auth token: %+v", authToken)

	accessToken, err := s.server.Manager.GenerateAccessToken(ctx, oauth2.AuthorizationCode, &oauth2.TokenGenerateRequest{
		ClientID:     authToken.GetClientID(),
		ClientSecret: clientSecret,
		RedirectURI:  authToken.GetRedirectURI(),
		Scope:        authToken.GetScope(),
		Code:         authToken.GetCode(),
	})
	if err != nil {
		return nil, fmt.Errorf("error generating user-level access token: %s", err)
	}
	if accessToken == nil {
		return nil, errors.New("generated user-level access token was empty")
	}
	log.Tracef(ctx, "obtained user-level access token: %+v", accessToken)
	return accessToken, nil
}

func (s *s) LoadAccessToken(ctx context.Context, access string) (accessToken oauth2.TokenInfo, err error) {
	return s.server.Manager.LoadAccessToken(ctx, access)
}

func (s *s) RevokeAccessToken(
	ctx context.Context,
	clientID string,
	clientSecret string,
	access string,
) gtserror.WithCode {
	token, err := s.server.Manager.LoadAccessToken(ctx, access)
	switch {
	case err == nil:
		// Got the token, can
		// proceed to invalidate.

	case errorsv2.IsV2(
		err,
		db.ErrNoEntries,
		oautherr.ErrExpiredAccessToken,
	):
		// Token already deleted, expired,
		// or doesn't exist, nothing to do.
		return nil

	default:
		// Real error.
		log.Errorf(ctx, "db error loading access token: %v", err)
		return gtserror.NewErrorInternalError(
			oautherr.ErrServerError,
			"db error loading access token, check logs",
		)
	}

	// Ensure token's client ID matches provided client ID.
	if token.GetClientID() != clientID {
		log.Debug(ctx, "client id of token does not match provided client_id")
		return gtserror.NewErrorForbidden(
			oautherr.ErrUnauthorizedClient,
			"You are not authorized to revoke this token",
		)
	}

	// Get client from the db using provided client ID.
	client, err := s.server.Manager.GetClient(ctx, clientID)
	if err != nil {
		log.Errorf(ctx, "db error loading client: %v", err)
		return gtserror.NewErrorInternalError(
			oautherr.ErrServerError,
			"db error loading client, check logs",
		)
	}

	// Ensure requester also knows the client secret,
	// which confirms that they indeed created the client.
	if client.GetSecret() != clientSecret {
		log.Debug(ctx, "secret of client does not match provided client_secret")
		return gtserror.NewErrorForbidden(
			oautherr.ErrUnauthorizedClient,
			"You are not authorized to revoke this token",
		)
	}

	// All good, invalidate the token.
	err = s.server.Manager.RemoveAccessToken(ctx, access)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		log.Errorf(ctx, "db error removing access token: %v", err)
		return gtserror.NewErrorInternalError(
			oautherr.ErrServerError,
			"db error removing access token, check logs",
		)
	}

	return nil
}
