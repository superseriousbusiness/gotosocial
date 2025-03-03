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
	"time"

	"codeberg.org/gruf/go-mutexes"
	"codeberg.org/superseriousbusiness/oauth2/v4"
	"codeberg.org/superseriousbusiness/oauth2/v4/models"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

// tokenStore is an implementation of oauth2.TokenStore, which uses our db interface as a storage backend.
type tokenStore struct {
	oauth2.TokenStore
	state         *state.State
	lastUsedLocks mutexes.MutexMap
}

// newTokenStore returns a token store that satisfies the oauth2.TokenStore interface.
//
// In order to allow tokens to 'expire', it will also set off a goroutine that iterates through
// the tokens in the DB once per minute and deletes any that have expired.
func newTokenStore(ctx context.Context, state *state.State) oauth2.TokenStore {
	ts := &tokenStore{state: state}

	// Set the token store to clean out expired tokens
	// once per minute, or return if we're done.
	go func(ctx context.Context, ts *tokenStore) {
	cleanloop:
		for {
			select {
			case <-ctx.Done():
				log.Info(ctx, "breaking cleanloop")
				break cleanloop
			case <-time.After(1 * time.Minute):
				log.Trace(ctx, "sweeping out old oauth entries broom broom")
				if err := ts.sweep(ctx); err != nil {
					log.Errorf(ctx, "error while sweeping oauth entries: %s", err)
				}
			}
		}
	}(ctx, ts)
	return ts
}

// sweep clears out old tokens that have expired;
// it should be run on a loop about once per minute or so.
func (ts *tokenStore) sweep(ctx context.Context) error {
	// Select *all* tokens from the db
	//
	// TODO: if this becomes expensive
	// (ie., there are fucking LOADS of
	// tokens) then figure out a better way.
	tokens, err := ts.state.DB.GetAllTokens(ctx)
	if err != nil {
		return err
	}

	// Remove any expired tokens, bearing
	// in mind that zero time = no expiry.
	now := time.Now()
	for _, token := range tokens {
		var expired bool

		switch {
		case !token.CodeExpiresAt.IsZero() && token.CodeExpiresAt.Before(now):
			log.Tracef(ctx, "code token %s is expired", token.ID)
			expired = true

		case !token.RefreshExpiresAt.IsZero() && token.RefreshExpiresAt.Before(now):
			log.Tracef(ctx, "refresh token %s is expired", token.ID)
			expired = true

		case !token.AccessExpiresAt.IsZero() && token.AccessExpiresAt.Before(now):
			log.Tracef(ctx, "access token %s is expired", token.ID)
			expired = true
		}

		if !expired {
			// Token's
			// still good.
			continue
		}

		if err := ts.state.DB.DeleteTokenByID(ctx, token.ID); err != nil {
			err := gtserror.Newf("db error expiring token %s: %w", token.ID, err)
			return err
		}
	}

	return nil
}

// Create creates and store the new token information.
func (ts *tokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	t, ok := info.(*models.Token)
	if !ok {
		return errors.New("info param was not a models.Token")
	}

	dbt := TokenToDBToken(t)
	if dbt.ID == "" {
		dbt.ID = id.NewULID()
	}

	return ts.state.DB.PutToken(ctx, dbt)
}

// RemoveByCode deletes a token from the DB based on the Code field
func (ts *tokenStore) RemoveByCode(ctx context.Context, code string) error {
	return ts.state.DB.DeleteTokenByCode(ctx, code)
}

// RemoveByAccess deletes a token from the DB based on the Access field
func (ts *tokenStore) RemoveByAccess(ctx context.Context, access string) error {
	return ts.state.DB.DeleteTokenByAccess(ctx, access)
}

// RemoveByRefresh deletes a token from the DB based on the Refresh field
func (ts *tokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	return ts.state.DB.DeleteTokenByRefresh(ctx, refresh)
}

// GetByCode selects a token from
// the DB based on the Code field
func (ts *tokenStore) GetByCode(
	ctx context.Context,
	code string,
) (oauth2.TokenInfo, error) {
	return ts.getUpdateToken(
		ctx,
		ts.state.DB.GetTokenByCode,
		code,
	)
}

// GetByAccess selects a token from
// the DB based on the Access field.
func (ts *tokenStore) GetByAccess(
	ctx context.Context,
	access string,
) (oauth2.TokenInfo, error) {
	return ts.getUpdateToken(
		ctx,
		ts.state.DB.GetTokenByAccess,
		access,
	)
}

// GetByRefresh selects a token from
// the DB based on the Refresh field
func (ts *tokenStore) GetByRefresh(
	ctx context.Context,
	refresh string,
) (oauth2.TokenInfo, error) {
	return ts.getUpdateToken(
		ctx,
		ts.state.DB.GetTokenByRefresh,
		refresh,
	)
}

// package-internal function for getting a token
// and potentially updating its last_used value.
func (ts *tokenStore) getUpdateToken(
	ctx context.Context,
	getBy func(context.Context, string) (*gtsmodel.Token, error),
	key string,
) (oauth2.TokenInfo, error) {
	// Hold a lock to get the token based on
	// whatever func + key we've been given.
	unlock := ts.lastUsedLocks.Lock(key)

	token, err := getBy(ctx, key)
	if err != nil {
		// Unlock on error.
		unlock()
		return nil, err
	}

	// If token was last used more than
	// an hour ago, update this in the db.
	wasLastUsed := token.LastUsed
	if now := time.Now(); now.Sub(wasLastUsed) > 1*time.Hour {
		token.LastUsed = now
		if err := ts.state.DB.UpdateToken(ctx, token, "last_used"); err != nil {
			// Unlock on error.
			unlock()
			err := gtserror.Newf("error updating last_used on token: %w", err)
			return nil, err
		}
	}

	// We're done, unlock.
	unlock()
	return DBTokenToToken(token), nil
}

/*
	The following models are basically helpers for the token store implementation, they should only be used internally.
*/

// TokenToDBToken is a lil util function that takes a gotosocial token and gives back a token for inserting into a database.
func TokenToDBToken(tkn *models.Token) *gtsmodel.Token {
	now := time.Now()

	// For the following, we want to make sure we're not adding a time.Now() to an *empty* ExpiresIn, otherwise that's
	// going to cause all sorts of interesting problems. So check first to make sure that the ExpiresIn is not equal
	// to the zero value of a time.Duration, which is 0s. If it *is* empty/nil, just leave the ExpiresAt at nil as well.

	cea := time.Time{}
	if tkn.CodeExpiresIn != 0*time.Second {
		cea = now.Add(tkn.CodeExpiresIn)
	}

	aea := time.Time{}
	if tkn.AccessExpiresIn != 0*time.Second {
		aea = now.Add(tkn.AccessExpiresIn)
	}

	rea := time.Time{}
	if tkn.RefreshExpiresIn != 0*time.Second {
		rea = now.Add(tkn.RefreshExpiresIn)
	}

	return &gtsmodel.Token{
		ClientID:            tkn.ClientID,
		UserID:              tkn.UserID,
		RedirectURI:         tkn.RedirectURI,
		Scope:               tkn.Scope,
		Code:                tkn.Code,
		CodeChallenge:       tkn.CodeChallenge,
		CodeChallengeMethod: tkn.CodeChallengeMethod,
		CodeCreateAt:        tkn.CodeCreateAt,
		CodeExpiresAt:       cea,
		Access:              tkn.Access,
		AccessCreateAt:      tkn.AccessCreateAt,
		AccessExpiresAt:     aea,
		Refresh:             tkn.Refresh,
		RefreshCreateAt:     tkn.RefreshCreateAt,
		RefreshExpiresAt:    rea,
	}
}

// DBTokenToToken is a lil util function that takes a database token and gives back a gotosocial token
func DBTokenToToken(dbt *gtsmodel.Token) *models.Token {
	now := time.Now()

	var codeExpiresIn time.Duration
	if !dbt.CodeExpiresAt.IsZero() {
		codeExpiresIn = dbt.CodeExpiresAt.Sub(now)
	}

	var accessExpiresIn time.Duration
	if !dbt.AccessExpiresAt.IsZero() {
		accessExpiresIn = dbt.AccessExpiresAt.Sub(now)
	}

	var refreshExpiresIn time.Duration
	if !dbt.RefreshExpiresAt.IsZero() {
		refreshExpiresIn = dbt.RefreshExpiresAt.Sub(now)
	}

	return &models.Token{
		ClientID:            dbt.ClientID,
		UserID:              dbt.UserID,
		RedirectURI:         dbt.RedirectURI,
		Scope:               dbt.Scope,
		Code:                dbt.Code,
		CodeChallenge:       dbt.CodeChallenge,
		CodeChallengeMethod: dbt.CodeChallengeMethod,
		CodeCreateAt:        dbt.CodeCreateAt,
		CodeExpiresIn:       codeExpiresIn,
		Access:              dbt.Access,
		AccessCreateAt:      dbt.AccessCreateAt,
		AccessExpiresIn:     accessExpiresIn,
		Refresh:             dbt.Refresh,
		RefreshCreateAt:     dbt.RefreshCreateAt,
		RefreshExpiresIn:    refreshExpiresIn,
	}
}
