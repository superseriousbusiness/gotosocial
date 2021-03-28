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

package db

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-pg/pg/extra/pgdebug"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/pkg/mastotypes"
	"golang.org/x/crypto/bcrypt"
)

// postgresService satisfies the DB interface
type postgresService struct {
	config       *config.DBConfig
	conn         *pg.DB
	log          *logrus.Entry
	cancel       context.CancelFunc
	federationDB pub.Database
}

// newPostgresService returns a postgresService derived from the provided config, which implements the go-fed DB interface.
// Under the hood, it uses https://github.com/go-pg/pg to create and maintain a database connection.
func newPostgresService(ctx context.Context, c *config.Config, log *logrus.Entry) (DB, error) {
	opts, err := derivePGOptions(c)
	if err != nil {
		return nil, fmt.Errorf("could not create postgres service: %s", err)
	}
	log.Debugf("using pg options: %+v", opts)

	readyChan := make(chan interface{})
	opts.OnConnect = func(ctx context.Context, c *pg.Conn) error {
		close(readyChan)
		return nil
	}

	// create a connection
	pgCtx, cancel := context.WithCancel(ctx)
	conn := pg.Connect(opts).WithContext(pgCtx)

	// this will break the logfmt format we normally log in,
	// since we can't choose where pg outputs to and it defaults to
	// stdout. So use this option with care!
	if log.Logger.GetLevel() >= logrus.TraceLevel {
		conn.AddQueryHook(pgdebug.DebugHook{
			// Print all queries.
			Verbose: true,
		})
	}

	// actually *begin* the connection so that we can tell if the db is there
	// and listening, and also trigger the opts.OnConnect function passed in above
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

	// make sure the opts.OnConnect function has been triggered
	// and closed the ready channel
	select {
	case <-readyChan:
		log.Infof("postgres connection ready")
	case <-time.After(5 * time.Second):
		cancel()
		return nil, errors.New("db connection timeout")
	}

	// we can confidently return this useable postgres service now
	return &postgresService{
		config:       c.DBConfig,
		conn:         conn,
		log:          log,
		cancel:       cancel,
		federationDB: newPostgresFederation(conn),
	}, nil
}

/*
	HANDY STUFF
*/

// derivePGOptions takes an application config and returns either a ready-to-use *pg.Options
// with sensible defaults, or an error if it's not satisfied by the provided config.
func derivePGOptions(c *config.Config) (*pg.Options, error) {
	if strings.ToUpper(c.DBConfig.Type) != dbTypePostgres {
		return nil, fmt.Errorf("expected db type of %s but got %s", dbTypePostgres, c.DBConfig.Type)
	}

	// validate port
	if c.DBConfig.Port == 0 {
		return nil, errors.New("no port set")
	}

	// validate address
	if c.DBConfig.Address == "" {
		return nil, errors.New("no address set")
	}

	ipv4Regex := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	hostnameRegex := regexp.MustCompile(`^(?:[a-z0-9]+(?:-[a-z0-9]+)*\.)+[a-z]{2,}$`)
	if !hostnameRegex.MatchString(c.DBConfig.Address) && !ipv4Regex.MatchString(c.DBConfig.Address) && c.DBConfig.Address != "localhost" {
		return nil, fmt.Errorf("address %s was neither an ipv4 address nor a valid hostname", c.DBConfig.Address)
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
	FEDERATION FUNCTIONALITY
*/

func (ps *postgresService) Federation() pub.Database {
	return ps.federationDB
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
		(*model.Account)(nil),
		(*model.Status)(nil),
		(*model.User)(nil),
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
			return ErrNoEntries{}
		}
		return err

	}
	return nil
}

func (ps *postgresService) GetWhere(key string, value interface{}, i interface{}) error {
	if err := ps.conn.Model(i).Where(fmt.Sprintf("%s = ?", key), value).Select(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetAll(i interface{}) error {
	if err := ps.conn.Model(i).Select(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) Put(i interface{}) error {
	_, err := ps.conn.Model(i).Insert(i)
	return err
}

func (ps *postgresService) UpdateByID(id string, i interface{}) error {
	if _, err := ps.conn.Model(i).OnConflict("(id) DO UPDATE").Insert(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) DeleteByID(id string, i interface{}) error {
	if _, err := ps.conn.Model(i).Where("id = ?", id).Delete(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) DeleteWhere(key string, value interface{}, i interface{}) error {
	if _, err := ps.conn.Model(i).Where(fmt.Sprintf("%s = ?", key), value).Delete(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	return nil
}

/*
	HANDY SHORTCUTS
*/

func (ps *postgresService) GetAccountByUserID(userID string, account *model.Account) error {
	user := &model.User{
		ID: userID,
	}
	if err := ps.conn.Model(user).Where("id = ?", userID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	if err := ps.conn.Model(account).Where("id = ?", user.AccountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetFollowRequestsForAccountID(accountID string, followRequests *[]model.FollowRequest) error {
	if err := ps.conn.Model(followRequests).Where("target_account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetFollowingByAccountID(accountID string, following *[]model.Follow) error {
	if err := ps.conn.Model(following).Where("account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetFollowersByAccountID(accountID string, followers *[]model.Follow) error {
	if err := ps.conn.Model(followers).Where("target_account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetStatusesByAccountID(accountID string, statuses *[]model.Status) error {
	if err := ps.conn.Model(statuses).Where("account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetStatusesByTimeDescending(accountID string, statuses *[]model.Status, limit int) error {
	q := ps.conn.Model(statuses).Order("created_at DESC")
	if limit != 0 {
		q = q.Limit(limit)
	}
	if accountID != "" {
		q = q.Where("account_id = ?", accountID)
	}
	if err := q.Select(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	return nil
}

func (ps *postgresService) GetLastStatusForAccountID(accountID string, status *model.Status) error {
	if err := ps.conn.Model(status).Order("created_at DESC").Limit(1).Where("account_id = ?", accountID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return ErrNoEntries{}
		}
		return err
	}
	return nil

}

func (ps *postgresService) IsUsernameAvailable(username string) error {
	// if no error we fail because it means we found something
	// if error but it's not pg.ErrNoRows then we fail
	// if err is pg.ErrNoRows we're good, we found nothing so continue
	if err := ps.conn.Model(&model.Account{}).Where("username = ?", username).Where("domain = ?", nil).Select(); err == nil {
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
	if err := ps.conn.Model(&model.EmailDomainBlock{}).Where("domain = ?", domain).Select(); err == nil {
		// fail because we found something
		return fmt.Errorf("email domain %s is blocked", domain)
	} else if err != pg.ErrNoRows {
		// fail because we got an unexpected error
		return fmt.Errorf("db error: %s", err)
	}

	// check if this email is associated with a user already
	if err := ps.conn.Model(&model.User{}).Where("email = ?", email).WhereOr("unconfirmed_email = ?", email).Select(); err == nil {
		// fail because we found something
		return fmt.Errorf("email %s already in use", email)
	} else if err != pg.ErrNoRows {
		// fail because we got an unexpected error
		return fmt.Errorf("db error: %s", err)
	}
	return nil
}

func (ps *postgresService) NewSignup(username string, reason string, requireApproval bool, email string, password string, signUpIP net.IP, locale string, appID string) (*model.User, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		ps.log.Errorf("error creating new rsa key: %s", err)
		return nil, err
	}

	a := &model.Account{
		Username:    username,
		DisplayName: username,
		Reason:      reason,
		PrivateKey:  key,
		PublicKey:   &key.PublicKey,
		ActorType:   "Person",
	}
	if _, err = ps.conn.Model(a).Insert(); err != nil {
		return nil, err
	}

	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %s", err)
	}
	u := &model.User{
		AccountID:              a.ID,
		EncryptedPassword:      string(pw),
		SignUpIP:               signUpIP,
		Locale:                 locale,
		UnconfirmedEmail:       email,
		CreatedByApplicationID: appID,
	}
	if _, err = ps.conn.Model(u).Insert(); err != nil {
		return nil, err
	}

	return u, nil
}

/*
	CONVERSION FUNCTIONS
*/

// AccountToMastoSensitive takes an internal account model and transforms it into an account ready to be served through the API.
// The resulting account fits the specifications for the path /api/v1/accounts/verify_credentials, as described here:
// https://docs.joinmastodon.org/methods/accounts/. Note that it's *sensitive* because it's only meant to be exposed to the user
// that the account actually belongs to.
func (ps *postgresService) AccountToMastoSensitive(a *model.Account) (*mastotypes.Account, error) {

	fields := []mastotypes.Field{}
	for _, f := range a.Fields {
		mField := mastotypes.Field{
			Name:  f.Name,
			Value: f.Value,
		}
		if !f.VerifiedAt.IsZero() {
			mField.VerifiedAt = f.VerifiedAt.Format(time.RFC3339)
		}
		fields = append(fields, mField)
	}

	// count followers
	followers := []model.Follow{}
	if err := ps.GetFollowersByAccountID(a.ID, &followers); err != nil {
		if _, ok := err.(ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting followers: %s", err)
		}
	}
	var followersCount int
	if followers != nil {
		followersCount = len(followers)
	}

	// count following
	following := []model.Follow{}
	if err := ps.GetFollowingByAccountID(a.ID, &following); err != nil {
		if _, ok := err.(ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting following: %s", err)
		}
	}
	var followingCount int
	if following != nil {
		followingCount = len(following)
	}

	// count statuses
	statuses := []model.Status{}
	if err := ps.GetStatusesByAccountID(a.ID, &statuses); err != nil {
		if _, ok := err.(ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting last statuses: %s", err)
		}
	}
	var statusesCount int
	if statuses != nil {
		statusesCount = len(statuses)
	}

	// check when the last status was
	lastStatus := &model.Status{}
	if err := ps.GetLastStatusForAccountID(a.ID, lastStatus); err != nil {
		if _, ok := err.(ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting last status: %s", err)
		}
	}
	var lastStatusAt string
	if lastStatus != nil {
		lastStatusAt = lastStatus.CreatedAt.Format(time.RFC3339)
	}

	fr := []model.FollowRequest{}
	if err := ps.GetFollowRequestsForAccountID(a.ID, &fr); err != nil {
		if _, ok := err.(ErrNoEntries); !ok {
			return nil, fmt.Errorf("error getting follow requests: %s", err)
		}
	}
	var frc int
	if fr != nil {
		frc = len(fr)
	}

	source := &mastotypes.Source{
		Privacy:             a.Privacy,
		Sensitive:           a.Sensitive,
		Language:            a.Language,
		Note:                a.Note,
		Fields:              fields,
		FollowRequestsCount: frc,
	}

	return &mastotypes.Account{
		ID:             a.ID,
		Username:       a.Username,
		Acct:           a.Username, // equivalent to username for local users only, which sensitive always is
		DisplayName:    a.DisplayName,
		Locked:         a.Locked,
		Bot:            a.Bot,
		CreatedAt:      a.CreatedAt.Format(time.RFC3339),
		Note:           a.Note,
		URL:            a.URL,
		Avatar:         a.AvatarRemoteURL.String(),
		AvatarStatic:   a.AvatarRemoteURL.String(),
		Header:         a.HeaderRemoteURL.String(),
		HeaderStatic:   a.HeaderRemoteURL.String(),
		FollowersCount: followersCount,
		FollowingCount: followingCount,
		StatusesCount:  statusesCount,
		LastStatusAt:   lastStatusAt,
		Source:         source,
		Emojis:         nil,
		Fields:         fields,
	}, nil
}
