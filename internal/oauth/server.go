/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package oauth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/oauth2/v4"
	"github.com/superseriousbusiness/oauth2/v4/errors"
	"github.com/superseriousbusiness/oauth2/v4/manage"
	"github.com/superseriousbusiness/oauth2/v4/server"
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
	OOBTokenPath = "/oob"
	// HelpfulAdvice is a handy hint to users;
	// particularly important during the login flow
	HelpfulAdvice = "If you arrived at this error during a login/oauth flow, please try clearing your session cookies and logging in again; if problems persist, make sure you're using the correct credentials"
)

// Server wraps some oauth2 server functions in an interface, exposing only what is needed
type Server interface {
	HandleTokenRequest(r *http.Request) (map[string]interface{}, gtserror.WithCode)
	HandleAuthorizeRequest(w http.ResponseWriter, r *http.Request) gtserror.WithCode
	ValidationBearerToken(r *http.Request) (oauth2.TokenInfo, error)
	GenerateUserAccessToken(ctx context.Context, ti oauth2.TokenInfo, clientSecret string, userID string) (accessToken oauth2.TokenInfo, err error)
	LoadAccessToken(ctx context.Context, access string) (accessToken oauth2.TokenInfo, err error)
}

// s fulfils the Server interface using the underlying oauth2 server
type s struct {
	server *server.Server
}

// New returns a new oauth server that implements the Server interface
func New(ctx context.Context, database db.Basic) Server {
	ts := newTokenStore(ctx, database)
	cs := NewClientStore(database)

	manager := manage.NewDefaultManager()
	manager.MapTokenStorage(ts)
	manager.MapClientStorage(cs)
	manager.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    0,     // access tokens don't expire -- they must be revoked
		IsGenerateRefresh: false, // don't use refresh tokens
	})
	sc := &server.Config{
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
		AllowedCodeChallengeMethods: []oauth2.CodeChallengeMethod{oauth2.CodeChallengePlain},
	}

	srv := server.NewServer(sc, manager)
	srv.SetInternalErrorHandler(func(err error) *errors.Response {
		log.Errorf("internal oauth error: %s", err)
		return nil
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Errorf("internal response error: %s", re.Error)
	})

	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (string, error) {
		userID := r.FormValue("userid")
		if userID == "" {
			return "", errors.New("userid was empty")
		}
		return userID, nil
	})
	srv.SetClientInfoHandler(server.ClientFormHandler)
	return &s{
		server: srv,
	}
}

// HandleTokenRequest wraps the oauth2 library's HandleTokenRequest function
func (s *s) HandleTokenRequest(r *http.Request) (map[string]interface{}, gtserror.WithCode) {
	ctx := r.Context()

	gt, tgr, err := s.server.ValidationTokenRequest(r)
	if err != nil {
		help := fmt.Sprintf("could not validate token request: %s", err)
		return nil, gtserror.NewErrorBadRequest(err, help, HelpfulAdvice)
	}

	ti, err := s.server.GetAccessToken(ctx, gt, tgr)
	if err != nil {
		help := fmt.Sprintf("could not get access token: %s", err)
		return nil, gtserror.NewErrorBadRequest(err, help, HelpfulAdvice)
	}

	data := s.server.GetTokenData(ti)

	if expiresInI, ok := data["expires_in"]; ok {
		switch expiresIn := expiresInI.(type) {
		case int64:
			// remove this key from the returned map
			// if the value is 0 or less, so that clients
			// don't interpret the token as already expired
			if expiresIn <= 0 {
				delete(data, "expires_in")
			}
		default:
			err := errors.New("expires_in was set on token response, but was not an int64")
			return nil, gtserror.NewErrorInternalError(err, HelpfulAdvice)
		}
	}

	// add this for mastodon api compatibility
	data["created_at"] = ti.GetAccessCreateAt().Unix()

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

	// specify the scope of authorization
	if fn := s.server.AuthorizeScopeHandler; fn != nil {
		scope, err := fn(w, r)
		if err != nil {
			return s.errorOrRedirect(err, w, req)
		} else if scope != "" {
			req.Scope = scope
		}
	}

	// specify the expiration time of access token
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

	// If the redirect URI is empty, the default domain provided by the client is used.
	if req.RedirectURI == "" {
		client, err := s.server.Manager.GetClient(ctx, req.ClientID)
		if err != nil {
			return gtserror.NewErrorUnauthorized(err, HelpfulAdvice)
		}
		req.RedirectURI = client.GetDomain()
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
	log.Tracef("obtained auth token: %+v", authToken)

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
	log.Tracef("obtained user-level access token: %+v", accessToken)
	return accessToken, nil
}

func (s *s) LoadAccessToken(ctx context.Context, access string) (accessToken oauth2.TokenInfo, err error) {
	return s.server.Manager.LoadAccessToken(ctx, access)
}
