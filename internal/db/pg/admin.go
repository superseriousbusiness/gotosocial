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

package pg

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
	"net/mail"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"golang.org/x/crypto/bcrypt"
)

type adminDB struct {
	config *config.Config
	conn   *pg.DB
	log    *logrus.Logger
	cancel context.CancelFunc
}

func (a *adminDB) IsUsernameAvailable(username string) db.Error {
	// if no error we fail because it means we found something
	// if error but it's not pg.ErrNoRows then we fail
	// if err is pg.ErrNoRows we're good, we found nothing so continue
	if err := a.conn.Model(&gtsmodel.Account{}).Where("username = ?", username).Where("domain = ?", nil).Select(); err == nil {
		return fmt.Errorf("username %s already in use", username)
	} else if err != pg.ErrNoRows {
		return fmt.Errorf("db error: %s", err)
	}
	return nil
}

func (a *adminDB) IsEmailAvailable(email string) db.Error {
	// parse the domain from the email
	m, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("error parsing email address %s: %s", email, err)
	}
	domain := strings.Split(m.Address, "@")[1] // domain will always be the second part after @

	// check if the email domain is blocked
	if err := a.conn.Model(&gtsmodel.EmailDomainBlock{}).Where("domain = ?", domain).Select(); err == nil {
		// fail because we found something
		return fmt.Errorf("email domain %s is blocked", domain)
	} else if err != pg.ErrNoRows {
		// fail because we got an unexpected error
		return fmt.Errorf("db error: %s", err)
	}

	// check if this email is associated with a user already
	if err := a.conn.Model(&gtsmodel.User{}).Where("email = ?", email).WhereOr("unconfirmed_email = ?", email).Select(); err == nil {
		// fail because we found something
		return fmt.Errorf("email %s already in use", email)
	} else if err != pg.ErrNoRows {
		// fail because we got an unexpected error
		return fmt.Errorf("db error: %s", err)
	}
	return nil
}

func (a *adminDB) NewSignup(username string, reason string, requireApproval bool, email string, password string, signUpIP net.IP, locale string, appID string, emailVerified bool, admin bool) (*gtsmodel.User, db.Error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		a.log.Errorf("error creating new rsa key: %s", err)
		return nil, err
	}

	// if something went wrong while creating a user, we might already have an account, so check here first...
	acct := &gtsmodel.Account{}
	err = a.conn.Model(acct).Where("username = ?", username).Where("? IS NULL", pg.Ident("domain")).Select()
	if err != nil {
		// there's been an actual error
		if err != pg.ErrNoRows {
			return nil, fmt.Errorf("db error checking existence of account: %s", err)
		}

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
		if _, err = a.conn.Model(acct).Insert(); err != nil {
			return nil, err
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

	if _, err = a.conn.Model(u).Insert(); err != nil {
		return nil, err
	}

	return u, nil
}

func (a *adminDB) CreateInstanceAccount() db.Error {
	username := a.config.Host
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		a.log.Errorf("error creating new rsa key: %s", err)
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
	inserted, err := a.conn.Model(acct).Where("username = ?", username).SelectOrInsert()
	if err != nil {
		return err
	}
	if inserted {
		a.log.Infof("created instance account %s with id %s", username, acct.ID)
	} else {
		a.log.Infof("instance account %s already exists with id %s", username, acct.ID)
	}
	return nil
}

func (a *adminDB) CreateInstanceInstance() db.Error {
	iID, err := id.NewRandomULID()
	if err != nil {
		return err
	}

	i := &gtsmodel.Instance{
		ID:     iID,
		Domain: a.config.Host,
		Title:  a.config.Host,
		URI:    fmt.Sprintf("%s://%s", a.config.Protocol, a.config.Host),
	}
	inserted, err := a.conn.Model(i).Where("domain = ?", a.config.Host).SelectOrInsert()
	if err != nil {
		return err
	}
	if inserted {
		a.log.Infof("created instance instance %s with id %s", a.config.Host, i.ID)
	} else {
		a.log.Infof("instance instance %s already exists with id %s", a.config.Host, i.ID)
	}
	return nil
}
