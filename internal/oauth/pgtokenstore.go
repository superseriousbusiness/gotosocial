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
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/gotosocial/oauth2/v4"
	"github.com/gotosocial/oauth2/v4/models"
	"github.com/sirupsen/logrus"
)

// pgTokenStore is an implementation of oauth2.TokenStore, which uses Postgres as a storage backend.
type pgTokenStore struct {
	oauth2.TokenStore
	conn *pg.DB
	log  *logrus.Logger
}

// NewPGTokenStore returns a token store, using postgres, that satisfies the oauth2.TokenStore interface.
//
// In order to allow tokens to 'expire' (not really a thing in Postgres world), it will also set off a
// goroutine that iterates through the tokens in the DB once per minute and deletes any that have expired.
func NewPGTokenStore(ctx context.Context, conn *pg.DB, log *logrus.Logger) oauth2.TokenStore {
	pts := &pgTokenStore{
		conn: conn,
		log:  log,
	}

	// set the token store to clean out expired tokens once per minute, or return if we're done
	go func(ctx context.Context, pts *pgTokenStore, log *logrus.Logger) {
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
func (pts *pgTokenStore) sweep() error {
	// select *all* tokens from the db
	// todo: if this becomes expensive (ie., there are fucking LOADS of tokens) then figure out a better way.
	var tokens []oauthToken
	if err := pts.conn.Model(&tokens).Select(); err != nil {
		return err
	}

	// iterate through and remove expired tokens
	now := time.Now()
	for _, pgt := range tokens {
		// The zero value of a time.Time is 00:00 january 1 1970, which will always be before now. So:
		// we only want to check if a token expired before now if the expiry time is *not zero*;
		// ie., if it's been explicity set.
		if !pgt.CodeExpiresAt.IsZero() && pgt.CodeExpiresAt.Before(now) || !pgt.RefreshExpiresAt.IsZero() && pgt.RefreshExpiresAt.Before(now) || !pgt.AccessExpiresAt.IsZero() && pgt.AccessExpiresAt.Before(now) {
			if _, err := pts.conn.Model(&pgt).Delete(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Create creates and store the new token information.
// For the original implementation, see https://github.com/gotosocial/oauth2/blob/master/store/token.go#L34
func (pts *pgTokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	t, ok := info.(*models.Token)
	if !ok {
		return errors.New("info param was not a models.Token")
	}
	_, err := pts.conn.WithContext(ctx).Model(oauthTokenToPGToken(t)).Insert()
	return err
}

// RemoveByCode deletes a token from the DB based on the Code field
func (pts *pgTokenStore) RemoveByCode(ctx context.Context, code string) error {
	_, err := pts.conn.Model(&oauthToken{}).Where("code = ?", code).Delete()
	return err
}

// RemoveByAccess deletes a token from the DB based on the Access field
func (pts *pgTokenStore) RemoveByAccess(ctx context.Context, access string) error {
	_, err := pts.conn.Model(&oauthToken{}).Where("access = ?", access).Delete()
	return err
}

// RemoveByRefresh deletes a token from the DB based on the Refresh field
func (pts *pgTokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	_, err := pts.conn.Model(&oauthToken{}).Where("refresh = ?", refresh).Delete()
	return err
}

// GetByCode selects a token from the DB based on the Code field
func (pts *pgTokenStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	pgt := &oauthToken{}
	if err := pts.conn.Model(pgt).Where("code = ?", code).Select(); err != nil {
		return nil, err
	}
	return pgTokenToOauthToken(pgt), nil
}

// GetByAccess selects a token from the DB based on the Access field
func (pts *pgTokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	pgt := &oauthToken{}
	if err := pts.conn.Model(pgt).Where("access = ?", access).Select(); err != nil {
		return nil, err
	}
	return pgTokenToOauthToken(pgt), nil
}

// GetByRefresh selects a token from the DB based on the Refresh field
func (pts *pgTokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	pgt := &oauthToken{}
	if err := pts.conn.Model(pgt).Where("refresh = ?", refresh).Select(); err != nil {
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
	ClientID            string
	UserID              string
	RedirectURI         string
	Scope               string
	Code                string `pg:",pk"`
	CodeChallenge       string
	CodeChallengeMethod string
	CodeCreateAt        time.Time `pg:"type:timestamp"`
	CodeExpiresAt       time.Time `pg:"type:timestamp"`
	Access              string    `pg:",pk"`
	AccessCreateAt      time.Time `pg:"type:timestamp"`
	AccessExpiresAt     time.Time `pg:"type:timestamp"`
	Refresh             string    `pg:",pk"`
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
