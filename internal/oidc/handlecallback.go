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

package oidc

import (
	"context"
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (i *idp) HandleCallback(ctx context.Context, code string) (*Claims, gtserror.WithCode) {
	if code == "" {
		err := errors.New("code was empty string")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	log.Debug("exchanging code for oauth2token")
	oauth2Token, err := i.oauth2Config.Exchange(ctx, code)
	if err != nil {
		err := fmt.Errorf("error exchanging code for oauth2token: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	log.Debug("extracting id_token")
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		err := errors.New("no id_token in oauth2token")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}
	log.Debugf("raw id token: %s", rawIDToken)

	// Parse and verify ID Token payload.
	log.Debug("verifying id_token")
	idTokenVerifier := i.provider.Verifier(i.oidcConf)
	idToken, err := idTokenVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		err = fmt.Errorf("could not verify id token: %s", err)
		return nil, gtserror.NewErrorUnauthorized(err, err.Error())
	}

	log.Debug("extracting claims from id_token")
	claims := &Claims{}
	if err := idToken.Claims(claims); err != nil {
		err := fmt.Errorf("could not parse claims from idToken: %s", err)
		return nil, gtserror.NewErrorInternalError(err, err.Error())
	}

	return claims, nil
}

func (i *idp) AuthCodeURL(state string) string {
	return i.oauth2Config.AuthCodeURL(state)
}
