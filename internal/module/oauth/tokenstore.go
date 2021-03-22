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

package oauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotosocial/gotosocial/internal/db"
	"github.com/gotosocial/oauth2/v4"
	"github.com/gotosocial/oauth2/v4/models"
	"github.com/sirupsen/logrus"
)

// tokenStore is an implementation of oauth2.TokenStore, which uses our db interface as a storage backend.
type tokenStore struct {
	oauth2.TokenStore
	db  db.DB
	log *logrus.Logger
}

// newTokenStore returns a token store that satisfies the oauth2.TokenStore interface.
//
// In order to allow tokens to 'expire', it will also set off a goroutine that iterates through
// the tokens in the DB once per minute and deletes any that have expired.
func newTokenStore(ctx context.Context, db db.DB, log *logrus.Logger) oauth2.TokenStore {
	pts := &tokenStore{
		db:  db,
		log: log,
	}

	// set the token store to clean out expired tokens once per minute, or return if we're done
	go func(ctx context.Context, pts *tokenStore, log *logrus.Logger) {
	cleanloop:
		for {
			select {
			case <-ctx.Done():
				log.Info("breaking cleanloop")
				break cleanloop
			case <-time.After(1 * time.Minute):
				log.Debug("sweeping out old oauth entries broom broom")
				if err := pts.sweep(); err != nil {
					log.Errorf("error while sweeping oauth entries: %s", err)
				}
			}
		}
	}(ctx, pts, log)
	return pts
}

// sweep clears out old tokens that have expired; it should be run on a loop about once per minute or so.
func (pts *tokenStore) sweep() error {
	// select *all* tokens from the db
	// todo: if this becomes expensive (ie., there are fucking LOADS of tokens) then figure out a better way.
	tokens := new([]*oauthToken)
	if err := pts.db.GetAll(tokens); err != nil {
		return err
	}

	// iterate through and remove expired tokens
	now := time.Now()
	for _, pgt := range *tokens {
		// The zero value of a time.Time is 00:00 january 1 1970, which will always be before now. So:
		// we only want to check if a token expired before now if the expiry time is *not zero*;
		// ie., if it's been explicity set.
		if !pgt.CodeExpiresAt.IsZero() && pgt.CodeExpiresAt.Before(now) || !pgt.RefreshExpiresAt.IsZero() && pgt.RefreshExpiresAt.Before(now) || !pgt.AccessExpiresAt.IsZero() && pgt.AccessExpiresAt.Before(now) {
			if err := pts.db.DeleteByID(pgt.ID, &pgt); err != nil {
				return err
			}
		}
	}

	return nil
}

// Create creates and store the new token information.
// For the original implementation, see https://github.com/gotosocial/oauth2/blob/master/store/token.go#L34
func (pts *tokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	t, ok := info.(*models.Token)
	if !ok {
		return errors.New("info param was not a models.Token")
	}
	if err := pts.db.Put(oauthTokenToPGToken(t)); err != nil {
		return fmt.Errorf("error in tokenstore create: %s", err)
	}
	return nil
}

// RemoveByCode deletes a token from the DB based on the Code field
func (pts *tokenStore) RemoveByCode(ctx context.Context, code string) error {
	return pts.db.DeleteWhere("code", code, &oauthToken{})
}

// RemoveByAccess deletes a token from the DB based on the Access field
func (pts *tokenStore) RemoveByAccess(ctx context.Context, access string) error {
	return pts.db.DeleteWhere("access", access, &oauthToken{})
}

// RemoveByRefresh deletes a token from the DB based on the Refresh field
func (pts *tokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	return pts.db.DeleteWhere("refresh", refresh, &oauthToken{})
}

// GetByCode selects a token from the DB based on the Code field
func (pts *tokenStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	pgt := &oauthToken{
		Code: code,
	}
	if err := pts.db.GetWhere("code", code, pgt); err != nil {
		return nil, err
	}
	return pgTokenToOauthToken(pgt), nil
}

// GetByAccess selects a token from the DB based on the Access field
func (pts *tokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	pgt := &oauthToken{
		Access: access,
	}
	if err := pts.db.GetWhere("access", access, pgt); err != nil {
		return nil, err
	}
	return pgTokenToOauthToken(pgt), nil
}

// GetByRefresh selects a token from the DB based on the Refresh field
func (pts *tokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	pgt := &oauthToken{
		Refresh: refresh,
	}
	if err := pts.db.GetWhere("refresh", refresh, pgt); err != nil {
		return nil, err
	}
	return pgTokenToOauthToken(pgt), nil
}

/*
	The following models are basically helpers for the postgres token store implementation, they should only be used internally.
*/

// oauthToken is a translation of the gotosocial token with the ExpiresIn fields replaced with ExpiresAt.
//
// Explanation for this: gotosocial assumes an in-memory or file database of some kind, where a time-to-live parameter (TTL) can be defined,
// and tokens with expired TTLs are automatically removed. Since Postgres doesn't have that feature, it's easier to set an expiry time and
// then periodically sweep out tokens when that time has passed.
//
// Note that this struct does *not* satisfy the token interface shown here: https://github.com/gotosocial/oauth2/blob/master/model.go#L22
// and implemented here: https://github.com/gotosocial/oauth2/blob/master/models/token.go.
// As such, manual translation is always required between oauthToken and the gotosocial *model.Token. The helper functions oauthTokenToPGToken
// and pgTokenToOauthToken can be used for that.
type oauthToken struct {
	ID                  string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull"`
	ClientID            string
	UserID              string
	RedirectURI         string
	Scope               string
	Code                string `pg:"default:'',pk"`
	CodeChallenge       string
	CodeChallengeMethod string
	CodeCreateAt        time.Time `pg:"type:timestamp"`
	CodeExpiresAt       time.Time `pg:"type:timestamp"`
	Access              string    `pg:"default:'',pk"`
	AccessCreateAt      time.Time `pg:"type:timestamp"`
	AccessExpiresAt     time.Time `pg:"type:timestamp"`
	Refresh             string    `pg:"default:'',pk"`
	RefreshCreateAt     time.Time `pg:"type:timestamp"`
	RefreshExpiresAt    time.Time `pg:"type:timestamp"`
}

// oauthTokenToPGToken is a lil util function that takes a gotosocial token and gives back a token for inserting into postgres
func oauthTokenToPGToken(tkn *models.Token) *oauthToken {
	now := time.Now()

	// For the following, we want to make sure we're not adding a time.Now() to an *empty* ExpiresIn, otherwise that's
	// going to cause all sorts of interesting problems. So check first to make sure that the ExpiresIn is not equal
	// to the zero value of a time.Duration, which is 0s. If it *is* empty/nil, just leave the ExpiresAt at nil as well.

	var cea time.Time
	if tkn.CodeExpiresIn != 0*time.Second {
		cea = now.Add(tkn.CodeExpiresIn)
	}

	var aea time.Time
	if tkn.AccessExpiresIn != 0*time.Second {
		aea = now.Add(tkn.AccessExpiresIn)
	}

	var rea time.Time
	if tkn.RefreshExpiresIn != 0*time.Second {
		rea = now.Add(tkn.RefreshExpiresIn)
	}

	return &oauthToken{
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

// pgTokenToOauthToken is a lil util function that takes a postgres token and gives back a gotosocial token
func pgTokenToOauthToken(pgt *oauthToken) *models.Token {
	now := time.Now()

	return &models.Token{
		ClientID:            pgt.ClientID,
		UserID:              pgt.UserID,
		RedirectURI:         pgt.RedirectURI,
		Scope:               pgt.Scope,
		Code:                pgt.Code,
		CodeChallenge:       pgt.CodeChallenge,
		CodeChallengeMethod: pgt.CodeChallengeMethod,
		CodeCreateAt:        pgt.CodeCreateAt,
		CodeExpiresIn:       pgt.CodeExpiresAt.Sub(now),
		Access:              pgt.Access,
		AccessCreateAt:      pgt.AccessCreateAt,
		AccessExpiresIn:     pgt.AccessExpiresAt.Sub(now),
		Refresh:             pgt.Refresh,
		RefreshCreateAt:     pgt.RefreshCreateAt,
		RefreshExpiresIn:    pgt.RefreshExpiresAt.Sub(now),
	}
}
