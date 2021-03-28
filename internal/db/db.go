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
	"fmt"
	"net"
	"strings"

	"github.com/go-fed/activity/pub"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/pkg/mastotypes"
)

const dbTypePostgres string = "POSTGRES"

// ErrNoEntries is to be returned from the DB interface when no entries are found for a given query.
type ErrNoEntries struct{}

func (e ErrNoEntries) Error() string {
	return "no entries"
}

// DB provides methods for interacting with an underlying database or other storage mechanism (for now, just postgres).
// Note that in all of the functions below, the passed interface should be a pointer or a slice, which will then be populated
// by whatever is returned from the database.
type DB interface {
	// Federation returns an interface that's compatible with go-fed, for performing federation storage/retrieval functions.
	// See: https://pkg.go.dev/github.com/go-fed/activity@v1.0.0/pub?utm_source=gopls#Database
	Federation() pub.Database

	/*
		BASIC DB FUNCTIONALITY
	*/

	// CreateTable creates a table for the given interface.
	// For implementations that don't use tables, this can just return nil.
	CreateTable(i interface{}) error

	// DropTable drops the table for the given interface.
	// For implementations that don't use tables, this can just return nil.
	DropTable(i interface{}) error

	// Stop should stop and close the database connection cleanly, returning an error if this is not possible.
	// If the database implementation doesn't need to be stopped, this can just return nil.
	Stop(ctx context.Context) error

	// IsHealthy should return nil if the database connection is healthy, or an error if not.
	IsHealthy(ctx context.Context) error

	// GetByID gets one entry by its id. In a database like postgres, this might be the 'id' field of the entry,
	// for other implementations (for example, in-memory) it might just be the key of a map.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	// In case of no entries, a 'no entries' error will be returned
	GetByID(id string, i interface{}) error

	// GetWhere gets one entry where key = value. This is similar to GetByID but allows the caller to specify the
	// name of the key to select from.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	// In case of no entries, a 'no entries' error will be returned
	GetWhere(key string, value interface{}, i interface{}) error

	// GetAll will try to get all entries of type i.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	// In case of no entries, a 'no entries' error will be returned
	GetAll(i interface{}) error

	// Put simply stores i. It is up to the implementation to figure out how to store it, and using what key.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	Put(i interface{}) error

	// UpdateByID updates i with id id.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	UpdateByID(id string, i interface{}) error

	// DeleteByID removes i with id id.
	// If i didn't exist anyway, then no error should be returned.
	DeleteByID(id string, i interface{}) error

	// DeleteWhere deletes i where key = value
	// If i didn't exist anyway, then no error should be returned.
	DeleteWhere(key string, value interface{}, i interface{}) error

	/*
		HANDY SHORTCUTS
	*/

	// GetAccountByUserID is a shortcut for the common action of fetching an account corresponding to a user ID.
	// The given account pointer will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetAccountByUserID(userID string, account *model.Account) error

	// GetFollowRequestsForAccountID is a shortcut for the common action of fetching a list of follow requests targeting the given account ID.
	// The given slice 'followRequests' will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetFollowRequestsForAccountID(accountID string, followRequests *[]model.FollowRequest) error

	// GetFollowingByAccountID is a shortcut for the common action of fetching a list of accounts that accountID is following.
	// The given slice 'following' will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetFollowingByAccountID(accountID string, following *[]model.Follow) error

	// GetFollowersByAccountID is a shortcut for the common action of fetching a list of accounts that accountID is followed by.
	// The given slice 'followers' will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetFollowersByAccountID(accountID string, followers *[]model.Follow) error

	// GetStatusesByAccountID is a shortcut for the common action of fetching a list of statuses produced by accountID.
	// The given slice 'statuses' will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetStatusesByAccountID(accountID string, statuses *[]model.Status) error

	// GetStatusesByTimeDescending is a shortcut for getting the most recent statuses. accountID is optional, if not provided
	// then all statuses will be returned. If limit is set to 0, the size of the returned slice will not be limited. This can
	// be very memory intensive so you probably shouldn't do this!
	// In case of no entries, a 'no entries' error will be returned
	GetStatusesByTimeDescending(accountID string, statuses *[]model.Status, limit int) error

	// GetLastStatusForAccountID simply gets the most recent status by the given account.
	// The given slice 'status' pointer will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetLastStatusForAccountID(accountID string, status *model.Status) error

	// IsUsernameAvailable checks whether a given username is available on our domain.
	// Returns an error if the username is already taken, or something went wrong in the db.
	IsUsernameAvailable(username string) error

	// IsEmailAvailable checks whether a given email address for a new account is available to be used on our domain.
	// Return an error if:
	// A) the email is already associated with an account
	// B) we block signups from this email domain
	// C) something went wrong in the db
	IsEmailAvailable(email string) error

	// NewSignup creates a new user in the database with the given parameters, with an *unconfirmed* email address.
	// By the time this function is called, it should be assumed that all the parameters have passed validation!
	NewSignup(username string, reason string, requireApproval bool, email string, password string, signUpIP net.IP, locale string, appID string) (*model.User, error)

	/*
		USEFUL CONVERSION FUNCTIONS
	*/

	// AccountToMastoSensitive takes a db model account as a param, and returns a populated mastotype account, or an error
	// if something goes wrong. The returned account should be ready to serialize on an API level, and may have sensitive fields,
	// so serve it only to an authorized user who should have permission to see it.
	AccountToMastoSensitive(account *model.Account) (*mastotypes.Account, error)
}

// New returns a new database service that satisfies the DB interface and, by extension,
// the go-fed database interface described here: https://github.com/go-fed/activity/blob/master/pub/database.go
func New(ctx context.Context, c *config.Config, log *logrus.Logger) (DB, error) {
	switch strings.ToUpper(c.DBConfig.Type) {
	case dbTypePostgres:
		return newPostgresService(ctx, c, log.WithField("service", "db"))
	default:
		return nil, fmt.Errorf("database type %s not supported", c.DBConfig.Type)
	}
}
