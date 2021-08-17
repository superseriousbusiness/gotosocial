package pg

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
	"net/mail"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"golang.org/x/crypto/bcrypt"
)

func (ps *postgresService) IsUsernameAvailable(username string) error {
	// if no error we fail because it means we found something
	// if error but it's not pg.ErrNoRows then we fail
	// if err is pg.ErrNoRows we're good, we found nothing so continue
	if err := ps.conn.Model(&gtsmodel.Account{}).Where("username = ?", username).Where("domain = ?", nil).Select(); err == nil {
		return fmt.Errorf("username %s already in use", username)
	} else if err != pg.ErrNoRows {
		return fmt.Errorf("db error: %s", err)
	}
	return nil
}

func (ps *postgresService) IsEmailAvailable(email string) error {
	// parse the domain from the email
	m, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("error parsing email address %s: %s", email, err)
	}
	domain := strings.Split(m.Address, "@")[1] // domain will always be the second part after @

	// check if the email domain is blocked
	if err := ps.conn.Model(&gtsmodel.EmailDomainBlock{}).Where("domain = ?", domain).Select(); err == nil {
		// fail because we found something
		return fmt.Errorf("email domain %s is blocked", domain)
	} else if err != pg.ErrNoRows {
		// fail because we got an unexpected error
		return fmt.Errorf("db error: %s", err)
	}

	// check if this email is associated with a user already
	if err := ps.conn.Model(&gtsmodel.User{}).Where("email = ?", email).WhereOr("unconfirmed_email = ?", email).Select(); err == nil {
		// fail because we found something
		return fmt.Errorf("email %s already in use", email)
	} else if err != pg.ErrNoRows {
		// fail because we got an unexpected error
		return fmt.Errorf("db error: %s", err)
	}
	return nil
}

func (ps *postgresService) NewSignup(username string, reason string, requireApproval bool, email string, password string, signUpIP net.IP, locale string, appID string, emailVerified bool, admin bool) (*gtsmodel.User, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		ps.log.Errorf("error creating new rsa key: %s", err)
		return nil, err
	}

	// if something went wrong while creating a user, we might already have an account, so check here first...
	a := &gtsmodel.Account{}
	err = ps.conn.Model(a).Where("username = ?", username).Where("? IS NULL", pg.Ident("domain")).Select()
	if err != nil {
		// there's been an actual error
		if err != pg.ErrNoRows {
			return nil, fmt.Errorf("db error checking existence of account: %s", err)
		}

		// we just don't have an account yet create one
		newAccountURIs := util.GenerateURIsForAccount(username, ps.config.Protocol, ps.config.Host)
		newAccountID, err := id.NewRandomULID()
		if err != nil {
			return nil, err
		}

		a = &gtsmodel.Account{
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
		if _, err = ps.conn.Model(a).Insert(); err != nil {
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
		AccountID:              a.ID,
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

	if _, err = ps.conn.Model(u).Insert(); err != nil {
		return nil, err
	}

	return u, nil
}

func (ps *postgresService) CreateInstanceAccount() error {
	username := ps.config.Host
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		ps.log.Errorf("error creating new rsa key: %s", err)
		return err
	}

	aID, err := id.NewRandomULID()
	if err != nil {
		return err
	}

	newAccountURIs := util.GenerateURIsForAccount(username, ps.config.Protocol, ps.config.Host)
	a := &gtsmodel.Account{
		ID:                    aID,
		Username:              ps.config.Host,
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
	inserted, err := ps.conn.Model(a).Where("username = ?", username).SelectOrInsert()
	if err != nil {
		return err
	}
	if inserted {
		ps.log.Infof("created instance account %s with id %s", username, a.ID)
	} else {
		ps.log.Infof("instance account %s already exists with id %s", username, a.ID)
	}
	return nil
}

func (ps *postgresService) CreateInstanceInstance() error {
	iID, err := id.NewRandomULID()
	if err != nil {
		return err
	}

	i := &gtsmodel.Instance{
		ID:     iID,
		Domain: ps.config.Host,
		Title:  ps.config.Host,
		URI:    fmt.Sprintf("%s://%s", ps.config.Protocol, ps.config.Host),
	}
	inserted, err := ps.conn.Model(i).Where("domain = ?", ps.config.Host).SelectOrInsert()
	if err != nil {
		return err
	}
	if inserted {
		ps.log.Infof("created instance instance %s with id %s", ps.config.Host, i.ID)
	} else {
		ps.log.Infof("instance instance %s already exists with id %s", ps.config.Host, i.ID)
	}
	return nil
}
