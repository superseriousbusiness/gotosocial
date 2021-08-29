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

package bundb

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"net"
	"net/mail"
	"strings"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"golang.org/x/crypto/bcrypt"
)

type adminDB struct {
	config *config.Config
	conn   *DBConn
}

func (a *adminDB) IsUsernameAvailable(ctx context.Context, username string) (bool, db.Error) {
	q := a.conn.
		NewSelect().
		Model(&gtsmodel.Account{}).
		Where("username = ?", username).
		Where("domain = ?", nil)

	return a.conn.NotExists(ctx, q)
}

func (a *adminDB) IsEmailAvailable(ctx context.Context, email string) (bool, db.Error) {
	// parse the domain from the email
	m, err := mail.ParseAddress(email)
	if err != nil {
		return false, fmt.Errorf("error parsing email address %s: %s", email, err)
	}
	domain := strings.Split(m.Address, "@")[1] // domain will always be the second part after @

	// check if the email domain is blocked
	if err := a.conn.
		NewSelect().
		Model(&gtsmodel.EmailDomainBlock{}).
		Where("domain = ?", domain).
		Scan(ctx); err == nil {
		// fail because we found something
		return false, fmt.Errorf("email domain %s is blocked", domain)
	} else if err != sql.ErrNoRows {
		return false, a.conn.ProcessError(err)
	}

	// check if this email is associated with a user already
	q := a.conn.
		NewSelect().
		Model(&gtsmodel.User{}).
		Where("email = ?", email).
		WhereOr("unconfirmed_email = ?", email)

	return a.conn.NotExists(ctx, q)
}

func (a *adminDB) NewSignup(ctx context.Context, username string, reason string, requireApproval bool, email string, password string, signUpIP net.IP, locale string, appID string, emailVerified bool, admin bool) (*gtsmodel.User, db.Error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		a.conn.log.Errorf("error creating new rsa key: %s", err)
		return nil, err
	}

	// if something went wrong while creating a user, we might already have an account, so check here first...
	acct := &gtsmodel.Account{}
	err = a.conn.NewSelect().
		Model(acct).
		Where("username = ?", username).
		WhereGroup(" AND ", whereEmptyOrNull("domain")).
		Scan(ctx)
	if err != nil {
		// we just don't have an account yet create one
		newAccountURIs := util.GenerateURIsForAccount(username, a.config.Protocol, a.config.Host)
		newAccountID, err := id.NewRandomULID()
		if err != nil {
			return nil, err
		}

		acct = &gtsmodel.Account{
			ID:                    newAccountID,
			Username:              username,
			DisplayName:           username,
			Reason:                reason,
			URL:                   newAccountURIs.UserURL,
			PrivateKey:            key,
			PublicKey:             &key.PublicKey,
			PublicKeyURI:          newAccountURIs.PublicKeyURI,
			ActorType:             gtsmodel.ActivityStreamsPerson,
			URI:                   newAccountURIs.UserURI,
			InboxURI:              newAccountURIs.InboxURI,
			OutboxURI:             newAccountURIs.OutboxURI,
			FollowersURI:          newAccountURIs.FollowersURI,
			FollowingURI:          newAccountURIs.FollowingURI,
			FeaturedCollectionURI: newAccountURIs.CollectionURI,
		}
		if _, err = a.conn.
			NewInsert().
			Model(acct).
			Exec(ctx); err != nil {
			return nil, a.conn.ProcessError(err)
		}
	}

	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %s", err)
	}

	newUserID, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	u := &gtsmodel.User{
		ID:                     newUserID,
		AccountID:              acct.ID,
		EncryptedPassword:      string(pw),
		SignUpIP:               signUpIP.To4(),
		Locale:                 locale,
		UnconfirmedEmail:       email,
		CreatedByApplicationID: appID,
		Approved:               !requireApproval, // if we don't require moderator approval, just pre-approve the user
	}

	if emailVerified {
		u.ConfirmedAt = time.Now()
		u.Email = email
	}

	if admin {
		u.Admin = true
		u.Moderator = true
	}

	if _, err = a.conn.
		NewInsert().
		Model(u).
		Exec(ctx); err != nil {
		return nil, a.conn.ProcessError(err)
	}

	return u, nil
}

func (a *adminDB) CreateInstanceAccount(ctx context.Context) db.Error {
	username := a.config.Host

	// check if instance account already exists
	existsQ := a.conn.
		NewSelect().
		Model(&gtsmodel.Account{}).
		Where("username = ?", username).
		WhereGroup(" AND ", whereEmptyOrNull("domain"))
	count, err := existsQ.Count(ctx)
	if err != nil && count == 1 {
		a.conn.log.Infof("instance account %s already exists", username)
		return nil
	} else if err != sql.ErrNoRows {
		return a.conn.ProcessError(err)
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		a.conn.log.Errorf("error creating new rsa key: %s", err)
		return err
	}

	aID, err := id.NewRandomULID()
	if err != nil {
		return err
	}

	newAccountURIs := util.GenerateURIsForAccount(username, a.config.Protocol, a.config.Host)
	acct := &gtsmodel.Account{
		ID:                    aID,
		Username:              a.config.Host,
		DisplayName:           username,
		URL:                   newAccountURIs.UserURL,
		PrivateKey:            key,
		PublicKey:             &key.PublicKey,
		PublicKeyURI:          newAccountURIs.PublicKeyURI,
		ActorType:             gtsmodel.ActivityStreamsPerson,
		URI:                   newAccountURIs.UserURI,
		InboxURI:              newAccountURIs.InboxURI,
		OutboxURI:             newAccountURIs.OutboxURI,
		FollowersURI:          newAccountURIs.FollowersURI,
		FollowingURI:          newAccountURIs.FollowingURI,
		FeaturedCollectionURI: newAccountURIs.CollectionURI,
	}

	insertQ := a.conn.
		NewInsert().
		Model(acct)

	if _, err := insertQ.Exec(ctx); err != nil {
		return a.conn.ProcessError(err)
	}

	a.conn.log.Infof("instance account %s CREATED with id %s", username, acct.ID)
	return nil
}

func (a *adminDB) CreateInstanceInstance(ctx context.Context) db.Error {
	domain := a.config.Host

	// check if instance entry already exists
	q := a.conn.
		NewSelect().
		Model(&gtsmodel.Instance{}).
		Where("domain = ?", domain)

	exists, err := a.conn.Exists(ctx, q)
	if err != nil {
		return err
	}
	if exists {
		a.conn.log.Infof("instance entry already exists")
		return nil
	}

	iID, err := id.NewRandomULID()
	if err != nil {
		return err
	}

	i := &gtsmodel.Instance{
		ID:     iID,
		Domain: domain,
		Title:  domain,
		URI:    fmt.Sprintf("%s://%s", a.config.Protocol, a.config.Host),
	}

	insertQ := a.conn.
		NewInsert().
		Model(i)

	_, err = insertQ.Exec(ctx)
	if err != nil {
		return a.conn.ProcessError(err)
	}

	a.conn.log.Infof("created instance instance %s with id %s", domain, i.ID)
	return nil
}
