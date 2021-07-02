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
	"errors"
	"fmt"
	"net"
	"net/mail"
	"strings"
	"time"

	"github.com/go-pg/pg/extra/pgdebug"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"golang.org/x/crypto/bcrypt"
)

// postgresService satisfies the DB interface
type postgresService struct {
	config *config.Config
	conn   *pg.DB
	log    *logrus.Logger
	cancel context.CancelFunc
}

// NewPostgresService returns a postgresService derived from the provided config, which implements the go-fed DB interface.
// Under the hood, it uses https://github.com/go-pg/pg to create and maintain a database connection.
func NewPostgresService(ctx context.Context, c *config.Config, log *logrus.Logger) (db.DB, error) {
	opts, err := derivePGOptions(c)
	if err != nil {
		return nil, fmt.Errorf("could not create postgres service: %s", err)
	}
	log.Debugf("using pg options: %+v", opts)

	// create a connection
	pgCtx, cancel := context.WithCancel(ctx)
	conn := pg.Connect(opts).WithContext(pgCtx)

	// this will break the logfmt format we normally log in,
	// since we can't choose where pg outputs to and it defaults to
	// stdout. So use this option with care!
	if log.GetLevel() >= logrus.TraceLevel {
		conn.AddQueryHook(pgdebug.DebugHook{
			// Print all queries.
			Verbose: true,
		})
	}

	// actually *begin* the connection so that we can tell if the db is there and listening
	if err := conn.Ping(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("db connection error: %s", err)
	}

	// print out discovered postgres version
	var version string
	if _, err = conn.QueryOneContext(ctx, pg.Scan(&version), "SELECT version()"); err != nil {
		cancel()
		return nil, fmt.Errorf("db connection error: %s", err)
	}
	log.Infof("connected to postgres version: %s", version)

	ps := &postgresService{
		config: c,
		conn:   conn,
		log:    log,
		cancel: cancel,
	}

	// we can confidently return this useable postgres service now
	return ps, nil
}

/*
	HANDY STUFF
*/

// derivePGOptions takes an application config and returns either a ready-to-use *pg.Options
// with sensible defaults, or an error if it's not satisfied by the provided config.
func derivePGOptions(c *config.Config) (*pg.Options, error) {
	if strings.ToUpper(c.DBConfig.Type) != db.DBTypePostgres {
		return nil, fmt.Errorf("expected db type of %s but got %s", db.DBTypePostgres, c.DBConfig.Type)
	}

	// validate port
	if c.DBConfig.Port == 0 {
		return nil, errors.New("no port set")
	}

	// validate address
	if c.DBConfig.Address == "" {
		return nil, errors.New("no address set")
	}

	// validate username
	if c.DBConfig.User == "" {
		return nil, errors.New("no user set")
	}

	// validate that there's a password
	if c.DBConfig.Password == "" {
		return nil, errors.New("no password set")
	}

	// validate database
	if c.DBConfig.Database == "" {
		return nil, errors.New("no database set")
	}

	// We can rely on the pg library we're using to set
	// sensible defaults for everything we don't set here.
	options := &pg.Options{
		Addr:            fmt.Sprintf("%s:%d", c.DBConfig.Address, c.DBConfig.Port),
		User:            c.DBConfig.User,
		Password:        c.DBConfig.Password,
		Database:        c.DBConfig.Database,
		ApplicationName: c.ApplicationName,
	}

	return options, nil
}

/*
	BASIC DB FUNCTIONALITY
*/

func (ps *postgresService) CreateTable(i interface{}) error {
	return ps.conn.Model(i).CreateTable(&orm.CreateTableOptions{
		IfNotExists: true,
	})
}

func (ps *postgresService) DropTable(i interface{}) error {
	return ps.conn.Model(i).DropTable(&orm.DropTableOptions{
		IfExists: true,
	})
}

func (ps *postgresService) Stop(ctx context.Context) error {
	ps.log.Info("closing db connection")
	if err := ps.conn.Close(); err != nil {
		// only cancel if there's a problem closing the db
		ps.cancel()
		return err
	}
	return nil
}

func (ps *postgresService) IsHealthy(ctx context.Context) error {
	return ps.conn.Ping(ctx)
}

func (ps *postgresService) CreateSchema(ctx context.Context) error {
	models := []interface{}{
		(*gtsmodel.Account)(nil),
		(*gtsmodel.Status)(nil),
		(*gtsmodel.User)(nil),
	}
	ps.log.Info("creating db schema")

	for _, model := range models {
		err := ps.conn.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}

	ps.log.Info("db schema created")
	return nil
}

func (ps *postgresService) GetByID(id string, i interface{}) error {
	if err := ps.conn.Model(i).Where("id = ?", id).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err

	}
	return nil
}

func (ps *postgresService) GetWhere(where []db.Where, i interface{}) error {
	if len(where) == 0 {
		return errors.New("no queries provided")
	}

	q := ps.conn.Model(i)
	for _, w := range where {

		if w.Value == nil {
			q = q.Where("? IS NULL", pg.Ident(w.Key))
		} else {
			if w.CaseInsensitive {
				q = q.Where("LOWER(?) = LOWER(?)", pg.Safe(w.Key), w.Value)
			} else {
				q = q.Where("? = ?", pg.Safe(w.Key), w.Value)
			}
		}
	}

	if err := q.Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetAll(i interface{}) error {
	if err := ps.conn.Model(i).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) Put(i interface{}) error {
	_, err := ps.conn.Model(i).Insert(i)
	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return db.ErrAlreadyExists{}
	}
	return err
}

func (ps *postgresService) Upsert(i interface{}, conflictColumn string) error {
	if _, err := ps.conn.Model(i).OnConflict(fmt.Sprintf("(%s) DO UPDATE", conflictColumn)).Insert(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) UpdateByID(id string, i interface{}) error {
	if _, err := ps.conn.Model(i).Where("id = ?", id).OnConflict("(id) DO UPDATE").Insert(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) UpdateOneByID(id string, key string, value interface{}, i interface{}) error {
	_, err := ps.conn.Model(i).Set("? = ?", pg.Safe(key), value).Where("id = ?", id).Update()
	return err
}

func (ps *postgresService) DeleteByID(id string, i interface{}) error {
	if _, err := ps.conn.Model(i).Where("id = ?", id).Delete(); err != nil {
		// if there are no rows *anyway* then that's fine
		// just return err if there's an actual error
		if err != pg.ErrNoRows {
			return err
		}
	}
	return nil
}

func (ps *postgresService) DeleteWhere(where []db.Where, i interface{}) error {
	if len(where) == 0 {
		return errors.New("no queries provided")
	}

	q := ps.conn.Model(i)
	for _, w := range where {
		q = q.Where("? = ?", pg.Safe(w.Key), w.Value)
	}

	if _, err := q.Delete(); err != nil {
		// if there are no rows *anyway* then that's fine
		// just return err if there's an actual error
		if err != pg.ErrNoRows {
			return err
		}
	}
	return nil
}

/*
	HANDY SHORTCUTS
*/

func (ps *postgresService) AcceptFollowRequest(originAccountID string, targetAccountID string) (*gtsmodel.Follow, error) {
	// make sure the original follow request exists
	fr := &gtsmodel.FollowRequest{}
	if err := ps.conn.Model(fr).Where("account_id = ?", originAccountID).Where("target_account_id = ?", targetAccountID).Select(); err != nil {
		if err == pg.ErrMultiRows {
			return nil, db.ErrNoEntries{}
		}
		return nil, err
	}

	// create a new follow to 'replace' the request with
	follow := &gtsmodel.Follow{
		ID:              fr.ID,
		AccountID:       originAccountID,
		TargetAccountID: targetAccountID,
		URI:             fr.URI,
	}

	// if the follow already exists, just update the URI -- we don't need to do anything else
	if _, err := ps.conn.Model(follow).OnConflict("ON CONSTRAINT follows_account_id_target_account_id_key DO UPDATE set uri = ?", follow.URI).Insert(); err != nil {
		return nil, err
	}

	// now remove the follow request
	if _, err := ps.conn.Model(&gtsmodel.FollowRequest{}).Where("account_id = ?", originAccountID).Where("target_account_id = ?", targetAccountID).Delete(); err != nil {
		return nil, err
	}

	return follow, nil
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

func (ps *postgresService) GetAccountByUserID(userID string, account *gtsmodel.Account) error {
	user := &gtsmodel.User{
		ID: userID,
	}
	if err := ps.conn.Model(user).Where("id = ?", userID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	if err := ps.conn.Model(account).Where("id = ?", user.AccountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetLocalAccountByUsername(username string, account *gtsmodel.Account) error {
	if err := ps.conn.Model(account).Where("username = ?", username).Where("? IS NULL", pg.Ident("domain")).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetFollowRequestsForAccountID(accountID string, followRequests *[]gtsmodel.FollowRequest) error {
	if err := ps.conn.Model(followRequests).Where("target_account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetFollowingByAccountID(accountID string, following *[]gtsmodel.Follow) error {
	if err := ps.conn.Model(following).Where("account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetFollowersByAccountID(accountID string, followers *[]gtsmodel.Follow, localOnly bool) error {

	q := ps.conn.Model(followers)

	if localOnly {
		// for local accounts let's get where domain is null OR where domain is an empty string, just to be safe
		whereGroup := func(q *pg.Query) (*pg.Query, error) {
			q = q.
				WhereOr("? IS NULL", pg.Ident("a.domain")).
				WhereOr("a.domain = ?", "")
			return q, nil
		}

		q = q.ColumnExpr("follow.*").
			Join("JOIN accounts AS a ON follow.account_id = TEXT(a.id)").
			Where("follow.target_account_id = ?", accountID).
			WhereGroup(whereGroup)
	} else {
		q = q.Where("target_account_id = ?", accountID)
	}

	if err := q.Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetFavesByAccountID(accountID string, faves *[]gtsmodel.StatusFave) error {
	if err := ps.conn.Model(faves).Where("account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func (ps *postgresService) CountStatusesByAccountID(accountID string) (int, error) {
	count, err := ps.conn.Model(&gtsmodel.Status{}).Where("account_id = ?", accountID).Count()
	if err != nil {
		if err == pg.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

func (ps *postgresService) GetStatusesForAccount(accountID string, limit int, excludeReplies bool, maxID string, pinnedOnly bool, mediaOnly bool) ([]*gtsmodel.Status, error) {
	statuses := []*gtsmodel.Status{}

	q := ps.conn.Model(&statuses).Order("id DESC")
	if accountID != "" {
		q = q.Where("account_id = ?", accountID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if excludeReplies {
		q = q.Where("? IS NULL", pg.Ident("in_reply_to_id"))
	}

	if pinnedOnly {
		q = q.Where("pinned = ?", true)
	}

	if mediaOnly {
		q = q.WhereGroup(func(q *pg.Query) (*pg.Query, error) {
			return q.Where("? IS NOT NULL", pg.Ident("attachments")).Where("attachments != '{}'"), nil
		})
	}

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if err := q.Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, db.ErrNoEntries{}
		}
		return nil, err
	}
	return statuses, nil
}

func (ps *postgresService) GetLastStatusForAccountID(accountID string, status *gtsmodel.Status) error {
	if err := ps.conn.Model(status).Order("created_at DESC").Limit(1).Where("account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil

}

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

func (ps *postgresService) NewSignup(username string, reason string, requireApproval bool, email string, password string, signUpIP net.IP, locale string, appID string) (*gtsmodel.User, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		ps.log.Errorf("error creating new rsa key: %s", err)
		return nil, err
	}

	newAccountURIs := util.GenerateURIsForAccount(username, ps.config.Protocol, ps.config.Host)
	newAccountID, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	a := &gtsmodel.Account{
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
		AccountID:              newAccountID,
		EncryptedPassword:      string(pw),
		SignUpIP:               signUpIP,
		Locale:                 locale,
		UnconfirmedEmail:       email,
		CreatedByApplicationID: appID,
		Approved:               !requireApproval, // if we don't require moderator approval, just pre-approve the user
	}
	if _, err = ps.conn.Model(u).Insert(); err != nil {
		return nil, err
	}

	return u, nil
}

func (ps *postgresService) SetHeaderOrAvatarForAccountID(mediaAttachment *gtsmodel.MediaAttachment, accountID string) error {
	if mediaAttachment.Avatar && mediaAttachment.Header {
		return errors.New("one media attachment cannot be both header and avatar")
	}

	var headerOrAVI string
	if mediaAttachment.Avatar {
		headerOrAVI = "avatar"
	} else if mediaAttachment.Header {
		headerOrAVI = "header"
	} else {
		return errors.New("given media attachment was neither a header nor an avatar")
	}

	// TODO: there are probably more side effects here that need to be handled
	if _, err := ps.conn.Model(mediaAttachment).OnConflict("(id) DO UPDATE").Insert(); err != nil {
		return err
	}

	if _, err := ps.conn.Model(&gtsmodel.Account{}).Set(fmt.Sprintf("%s_media_attachment_id = ?", headerOrAVI), mediaAttachment.ID).Where("id = ?", accountID).Update(); err != nil {
		return err
	}
	return nil
}

func (ps *postgresService) GetHeaderForAccountID(header *gtsmodel.MediaAttachment, accountID string) error {
	acct := &gtsmodel.Account{}
	if err := ps.conn.Model(acct).Where("id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}

	if acct.HeaderMediaAttachmentID == "" {
		return db.ErrNoEntries{}
	}

	if err := ps.conn.Model(header).Where("id = ?", acct.HeaderMediaAttachmentID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetAvatarForAccountID(avatar *gtsmodel.MediaAttachment, accountID string) error {
	acct := &gtsmodel.Account{}
	if err := ps.conn.Model(acct).Where("id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}

	if acct.AvatarMediaAttachmentID == "" {
		return db.ErrNoEntries{}
	}

	if err := ps.conn.Model(avatar).Where("id = ?", acct.AvatarMediaAttachmentID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return db.ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) Blocked(account1 string, account2 string) (bool, error) {
	// TODO: check domain blocks as well
	var blocked bool
	if err := ps.conn.Model(&gtsmodel.Block{}).
		Where("account_id = ?", account1).Where("target_account_id = ?", account2).
		WhereOr("target_account_id = ?", account1).Where("account_id = ?", account2).
		Select(); err != nil {
		if err == pg.ErrNoRows {
			blocked = false
			return blocked, nil
		}
		return blocked, err
	}
	blocked = true
	return blocked, nil
}

func (ps *postgresService) GetRelationship(requestingAccount string, targetAccount string) (*gtsmodel.Relationship, error) {
	r := &gtsmodel.Relationship{
		ID: targetAccount,
	}

	// check if the requesting account follows the target account
	follow := &gtsmodel.Follow{}
	if err := ps.conn.Model(follow).Where("account_id = ?", requestingAccount).Where("target_account_id = ?", targetAccount).Select(); err != nil {
		if err != pg.ErrNoRows {
			// a proper error
			return nil, fmt.Errorf("getrelationship: error checking follow existence: %s", err)
		}
		// no follow exists so these are all false
		r.Following = false
		r.ShowingReblogs = false
		r.Notifying = false
	} else {
		// follow exists so we can fill these fields out...
		r.Following = true
		r.ShowingReblogs = follow.ShowReblogs
		r.Notifying = follow.Notify
	}

	// check if the target account follows the requesting account
	followedBy, err := ps.conn.Model(&gtsmodel.Follow{}).Where("account_id = ?", targetAccount).Where("target_account_id = ?", requestingAccount).Exists()
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking followed_by existence: %s", err)
	}
	r.FollowedBy = followedBy

	// check if the requesting account blocks the target account
	blocking, err := ps.conn.Model(&gtsmodel.Block{}).Where("account_id = ?", requestingAccount).Where("target_account_id = ?", targetAccount).Exists()
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking blocking existence: %s", err)
	}
	r.Blocking = blocking

	// check if the target account blocks the requesting account
	blockedBy, err := ps.conn.Model(&gtsmodel.Block{}).Where("account_id = ?", targetAccount).Where("target_account_id = ?", requestingAccount).Exists()
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking blocked existence: %s", err)
	}
	r.BlockedBy = blockedBy

	// check if there's a pending following request from requesting account to target account
	requested, err := ps.conn.Model(&gtsmodel.FollowRequest{}).Where("account_id = ?", requestingAccount).Where("target_account_id = ?", targetAccount).Exists()
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking blocked existence: %s", err)
	}
	r.Requested = requested

	return r, nil
}

func (ps *postgresService) Follows(sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, error) {
	if sourceAccount == nil || targetAccount == nil {
		return false, nil
	}

	return ps.conn.Model(&gtsmodel.Follow{}).Where("account_id = ?", sourceAccount.ID).Where("target_account_id = ?", targetAccount.ID).Exists()
}

func (ps *postgresService) FollowRequested(sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, error) {
	if sourceAccount == nil || targetAccount == nil {
		return false, nil
	}

	return ps.conn.Model(&gtsmodel.FollowRequest{}).Where("account_id = ?", sourceAccount.ID).Where("target_account_id = ?", targetAccount.ID).Exists()
}

func (ps *postgresService) Mutuals(account1 *gtsmodel.Account, account2 *gtsmodel.Account) (bool, error) {
	if account1 == nil || account2 == nil {
		return false, nil
	}

	// make sure account 1 follows account 2
	f1, err := ps.conn.Model(&gtsmodel.Follow{}).Where("account_id = ?", account1.ID).Where("target_account_id = ?", account2.ID).Exists()
	if err != nil {
		if err == pg.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	// make sure account 2 follows account 1
	f2, err := ps.conn.Model(&gtsmodel.Follow{}).Where("account_id = ?", account2.ID).Where("target_account_id = ?", account1.ID).Exists()
	if err != nil {
		if err == pg.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return f1 && f2, nil
}

func (ps *postgresService) GetReplyCountForStatus(status *gtsmodel.Status) (int, error) {
	return ps.conn.Model(&gtsmodel.Status{}).Where("in_reply_to_id = ?", status.ID).Count()
}

func (ps *postgresService) GetReblogCountForStatus(status *gtsmodel.Status) (int, error) {
	return ps.conn.Model(&gtsmodel.Status{}).Where("boost_of_id = ?", status.ID).Count()
}

func (ps *postgresService) GetFaveCountForStatus(status *gtsmodel.Status) (int, error) {
	return ps.conn.Model(&gtsmodel.StatusFave{}).Where("status_id = ?", status.ID).Count()
}

func (ps *postgresService) StatusFavedBy(status *gtsmodel.Status, accountID string) (bool, error) {
	return ps.conn.Model(&gtsmodel.StatusFave{}).Where("status_id = ?", status.ID).Where("account_id = ?", accountID).Exists()
}

func (ps *postgresService) StatusRebloggedBy(status *gtsmodel.Status, accountID string) (bool, error) {
	return ps.conn.Model(&gtsmodel.Status{}).Where("boost_of_id = ?", status.ID).Where("account_id = ?", accountID).Exists()
}

func (ps *postgresService) StatusMutedBy(status *gtsmodel.Status, accountID string) (bool, error) {
	return ps.conn.Model(&gtsmodel.StatusMute{}).Where("status_id = ?", status.ID).Where("account_id = ?", accountID).Exists()
}

func (ps *postgresService) StatusBookmarkedBy(status *gtsmodel.Status, accountID string) (bool, error) {
	return ps.conn.Model(&gtsmodel.StatusBookmark{}).Where("status_id = ?", status.ID).Where("account_id = ?", accountID).Exists()
}

func (ps *postgresService) WhoFavedStatus(status *gtsmodel.Status) ([]*gtsmodel.Account, error) {
	accounts := []*gtsmodel.Account{}

	faves := []*gtsmodel.StatusFave{}
	if err := ps.conn.Model(&faves).Where("status_id = ?", status.ID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return accounts, nil // no rows just means nobody has faved this status, so that's fine
		}
		return nil, err // an actual error has occurred
	}

	for _, f := range faves {
		acc := &gtsmodel.Account{}
		if err := ps.conn.Model(acc).Where("id = ?", f.AccountID).Select(); err != nil {
			if err == pg.ErrNoRows {
				continue // the account doesn't exist for some reason??? but this isn't the place to worry about that so just skip it
			}
			return nil, err // an actual error has occurred
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (ps *postgresService) WhoBoostedStatus(status *gtsmodel.Status) ([]*gtsmodel.Account, error) {
	accounts := []*gtsmodel.Account{}

	boosts := []*gtsmodel.Status{}
	if err := ps.conn.Model(&boosts).Where("boost_of_id = ?", status.ID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return accounts, nil // no rows just means nobody has boosted this status, so that's fine
		}
		return nil, err // an actual error has occurred
	}

	for _, f := range boosts {
		acc := &gtsmodel.Account{}
		if err := ps.conn.Model(acc).Where("id = ?", f.AccountID).Select(); err != nil {
			if err == pg.ErrNoRows {
				continue // the account doesn't exist for some reason??? but this isn't the place to worry about that so just skip it
			}
			return nil, err // an actual error has occurred
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (ps *postgresService) GetStatusesWhereFollowing(accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, error) {
	statuses := []*gtsmodel.Status{}

	q := ps.conn.Model(&statuses)

	q = q.ColumnExpr("status.*").
		Join("JOIN follows AS f ON f.target_account_id = status.account_id").
		Where("f.account_id = ?", accountID).
		Order("status.id DESC")

	if maxID != "" {
		q = q.Where("status.id < ?", maxID)
	}

	if sinceID != "" {
		q = q.Where("status.id > ?", sinceID)
	}

	if minID != "" {
		q = q.Where("status.id > ?", minID)
	}

	if local {
		q = q.Where("status.local = ?", local)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	err := q.Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, db.ErrNoEntries{}
		}
		return nil, err
	}

	if len(statuses) == 0 {
		return nil, db.ErrNoEntries{}
	}

	return statuses, nil
}

func (ps *postgresService) GetPublicTimelineForAccount(accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, error) {
	statuses := []*gtsmodel.Status{}

	q := ps.conn.Model(&statuses).
		Where("visibility = ?", gtsmodel.VisibilityPublic).
		Where("? IS NULL", pg.Ident("in_reply_to_id")).
		Where("? IS NULL", pg.Ident("in_reply_to_uri")).
		Where("? IS NULL", pg.Ident("boost_of_id")).
		Order("status.id DESC")

	if maxID != "" {
		q = q.Where("status.id < ?", maxID)
	}

	if sinceID != "" {
		q = q.Where("status.id > ?", sinceID)
	}

	if minID != "" {
		q = q.Where("status.id > ?", minID)
	}

	if local {
		q = q.Where("status.local = ?", local)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	err := q.Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, db.ErrNoEntries{}
		}
		return nil, err
	}

	return statuses, nil
}

func (ps *postgresService) GetNotificationsForAccount(accountID string, limit int, maxID string, sinceID string) ([]*gtsmodel.Notification, error) {
	notifications := []*gtsmodel.Notification{}

	q := ps.conn.Model(&notifications).Where("target_account_id = ?", accountID)

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if sinceID != "" {
		q = q.Where("id > ?", sinceID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	q = q.Order("created_at DESC")

	if err := q.Select(); err != nil {
		if err != pg.ErrNoRows {
			return nil, err
		}

	}
	return notifications, nil
}

/*
	CONVERSION FUNCTIONS
*/

// TODO: move these to the type converter, it's bananas that they're here and not there

func (ps *postgresService) MentionStringsToMentions(targetAccounts []string, originAccountID string, statusID string) ([]*gtsmodel.Mention, error) {
	ogAccount := &gtsmodel.Account{}
	if err := ps.conn.Model(ogAccount).Where("id = ?", originAccountID).Select(); err != nil {
		return nil, err
	}

	menchies := []*gtsmodel.Mention{}
	for _, a := range targetAccounts {
		// A mentioned account looks like "@test@example.org" or just "@test" for a local account
		// -- we can guarantee this from the regex that targetAccounts should have been derived from.
		// But we still need to do a bit of fiddling to get what we need here -- the username and domain (if given).

		// 1.  trim off the first @
		t := strings.TrimPrefix(a, "@")

		// 2. split the username and domain
		s := strings.Split(t, "@")

		// 3. if it's length 1 it's a local account, length 2 means remote, anything else means something is wrong
		var local bool
		switch len(s) {
		case 1:
			local = true
		case 2:
			local = false
		default:
			return nil, fmt.Errorf("mentioned account format '%s' was not valid", a)
		}

		var username, domain string
		username = s[0]
		if !local {
			domain = s[1]
		}

		// 4. check we now have a proper username and domain
		if username == "" || (!local && domain == "") {
			return nil, fmt.Errorf("username or domain for '%s' was nil", a)
		}

		// okay we're good now, we can start pulling accounts out of the database
		mentionedAccount := &gtsmodel.Account{}
		var err error

		// match username + account, case insensitive
		if local {
			// local user -- should have a null domain
			err = ps.conn.Model(mentionedAccount).Where("LOWER(?) = LOWER(?)", pg.Ident("username"), username).Where("? IS NULL", pg.Ident("domain")).Select()
		} else {
			// remote user -- should have domain defined
			err = ps.conn.Model(mentionedAccount).Where("LOWER(?) = LOWER(?)", pg.Ident("username"), username).Where("LOWER(?) = LOWER(?)", pg.Ident("domain"), domain).Select()
		}

		if err != nil {
			if err == pg.ErrNoRows {
				// no result found for this username/domain so just don't include it as a mencho and carry on about our business
				ps.log.Debugf("no account found with username '%s' and domain '%s', skipping it", username, domain)
				continue
			}
			// a serious error has happened so bail
			return nil, fmt.Errorf("error getting account with username '%s' and domain '%s': %s", username, domain, err)
		}

		// id, createdAt and updatedAt will be populated by the db, so we have everything we need!
		menchies = append(menchies, &gtsmodel.Mention{
			StatusID:            statusID,
			OriginAccountID:     ogAccount.ID,
			OriginAccountURI:    ogAccount.URI,
			TargetAccountID:     mentionedAccount.ID,
			NameString:          a,
			MentionedAccountURI: mentionedAccount.URI,
			MentionedAccountURL: mentionedAccount.URL,
			GTSAccount:          mentionedAccount,
		})
	}
	return menchies, nil
}

func (ps *postgresService) TagStringsToTags(tags []string, originAccountID string, statusID string) ([]*gtsmodel.Tag, error) {
	newTags := []*gtsmodel.Tag{}
	for _, t := range tags {
		tag := &gtsmodel.Tag{}
		// we can use selectorinsert here to create the new tag if it doesn't exist already
		// inserted will be true if this is a new tag we just created
		if err := ps.conn.Model(tag).Where("LOWER(?) = LOWER(?)", pg.Ident("name"), t).Select(); err != nil {
			if err == pg.ErrNoRows {
				// tag doesn't exist yet so populate it
				newID, err := id.NewRandomULID()
				if err != nil {
					return nil, err
				}
				tag.ID = newID
				tag.URL = fmt.Sprintf("%s://%s/tags/%s", ps.config.Protocol, ps.config.Host, t)
				tag.Name = t
				tag.FirstSeenFromAccountID = originAccountID
				tag.CreatedAt = time.Now()
				tag.UpdatedAt = time.Now()
				tag.Useable = true
				tag.Listable = true
			} else {
				return nil, fmt.Errorf("error getting tag with name %s: %s", t, err)
			}
		}

		// bail already if the tag isn't useable
		if !tag.Useable {
			continue
		}
		tag.LastStatusAt = time.Now()
		newTags = append(newTags, tag)
	}
	return newTags, nil
}

func (ps *postgresService) EmojiStringsToEmojis(emojis []string, originAccountID string, statusID string) ([]*gtsmodel.Emoji, error) {
	newEmojis := []*gtsmodel.Emoji{}
	for _, e := range emojis {
		emoji := &gtsmodel.Emoji{}
		err := ps.conn.Model(emoji).Where("shortcode = ?", e).Where("visible_in_picker = true").Where("disabled = false").Select()
		if err != nil {
			if err == pg.ErrNoRows {
				// no result found for this username/domain so just don't include it as an emoji and carry on about our business
				ps.log.Debugf("no emoji found with shortcode %s, skipping it", e)
				continue
			}
			// a serious error has happened so bail
			return nil, fmt.Errorf("error getting emoji with shortcode %s: %s", e, err)
		}
		newEmojis = append(newEmojis, emoji)
	}
	return newEmojis, nil
}
