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
	"net"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

const (
	// DBTypePostgres represents an underlying POSTGRES database type.
	DBTypePostgres string = "POSTGRES"
)

// DB provides methods for interacting with an underlying database or other storage mechanism (for now, just postgres).
// Note that in all of the functions below, the passed interface should be a pointer or a slice, which will then be populated
// by whatever is returned from the database.
type DB interface {
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
	GetWhere(where []Where, i interface{}) error

	// // GetWhereMany gets one entry where key = value for *ALL* parameters passed as "where".
	// // That is, if you pass 2 'where' entries, with 1 being Key username and Value test, and the second
	// // being Key domain and Value example.org, only entries will be returned where BOTH conditions are true.
	// GetWhereMany(i interface{}, where ...model.Where) error

	// GetAll will try to get all entries of type i.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	// In case of no entries, a 'no entries' error will be returned
	GetAll(i interface{}) error

	// Put simply stores i. It is up to the implementation to figure out how to store it, and using what key.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	Put(i interface{}) error

	// Upsert stores or updates i based on the given conflict column, as in https://www.postgresqltutorial.com/postgresql-upsert/
	// It is up to the implementation to figure out how to store it, and using what key.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	Upsert(i interface{}, conflictColumn string) error

	// UpdateByID updates i with id id.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	UpdateByID(id string, i interface{}) error

	// UpdateOneByID updates interface i with database the given database id. It will update one field of key key and value value.
	UpdateOneByID(id string, key string, value interface{}, i interface{}) error

	// DeleteByID removes i with id id.
	// If i didn't exist anyway, then no error should be returned.
	DeleteByID(id string, i interface{}) error

	// DeleteWhere deletes i where key = value
	// If i didn't exist anyway, then no error should be returned.
	DeleteWhere(where []Where, i interface{}) error

	/*
		HANDY SHORTCUTS
	*/

	// AcceptFollowRequest moves a follow request in the database from the follow_requests table to the follows table.
	// In other words, it should create the follow, and delete the existing follow request.
	//
	// It will return the newly created follow for further processing.
	AcceptFollowRequest(originAccountID string, targetAccountID string) (*gtsmodel.Follow, error)

	// CreateInstanceAccount creates an account in the database with the same username as the instance host value.
	// Ie., if the instance is hosted at 'example.org' the instance user will have a username of 'example.org'.
	// This is needed for things like serving files that belong to the instance and not an individual user/account.
	CreateInstanceAccount() error

	// CreateInstanceInstance creates an instance in the database with the same domain as the instance host value.
	// Ie., if the instance is hosted at 'example.org' the instance will have a domain of 'example.org'.
	// This is needed for things like serving instance information through /api/v1/instance
	CreateInstanceInstance() error

	// GetAccountByUserID is a shortcut for the common action of fetching an account corresponding to a user ID.
	// The given account pointer will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetAccountByUserID(userID string, account *gtsmodel.Account) error

	// GetLocalAccountByUsername is a shortcut for the common action of fetching an account ON THIS INSTANCE
	// according to its username, which should be unique.
	// The given account pointer will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetLocalAccountByUsername(username string, account *gtsmodel.Account) error

	// GetFollowRequestsForAccountID is a shortcut for the common action of fetching a list of follow requests targeting the given account ID.
	// The given slice 'followRequests' will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetFollowRequestsForAccountID(accountID string, followRequests *[]gtsmodel.FollowRequest) error

	// GetFollowingByAccountID is a shortcut for the common action of fetching a list of accounts that accountID is following.
	// The given slice 'following' will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetFollowingByAccountID(accountID string, following *[]gtsmodel.Follow) error

	// GetFollowersByAccountID is a shortcut for the common action of fetching a list of accounts that accountID is followed by.
	// The given slice 'followers' will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	//
	// If localOnly is set to true, then only followers from *this instance* will be returned.
	GetFollowersByAccountID(accountID string, followers *[]gtsmodel.Follow, localOnly bool) error

	// GetFavesByAccountID is a shortcut for the common action of fetching a list of faves made by the given accountID.
	// The given slice 'faves' will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetFavesByAccountID(accountID string, faves *[]gtsmodel.StatusFave) error

	// CountStatusesByAccountID is a shortcut for the common action of counting statuses produced by accountID.
	CountStatusesByAccountID(accountID string) (int, error)

	// GetStatusesByTimeDescending is a shortcut for getting the most recent statuses. accountID is optional, if not provided
	// then all statuses will be returned. If limit is set to 0, the size of the returned slice will not be limited. This can
	// be very memory intensive so you probably shouldn't do this!
	// In case of no entries, a 'no entries' error will be returned
	GetStatusesByTimeDescending(accountID string, statuses *[]gtsmodel.Status, limit int, excludeReplies bool, maxID string, pinned bool, mediaOnly bool) error

	// GetLastStatusForAccountID simply gets the most recent status by the given account.
	// The given slice 'status' pointer will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetLastStatusForAccountID(accountID string, status *gtsmodel.Status) error

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
	NewSignup(username string, reason string, requireApproval bool, email string, password string, signUpIP net.IP, locale string, appID string) (*gtsmodel.User, error)

	// SetHeaderOrAvatarForAccountID sets the header or avatar for the given accountID to the given media attachment.
	SetHeaderOrAvatarForAccountID(mediaAttachment *gtsmodel.MediaAttachment, accountID string) error

	// GetHeaderAvatarForAccountID gets the current avatar for the given account ID.
	// The passed mediaAttachment pointer will be populated with the value of the avatar, if it exists.
	GetAvatarForAccountID(avatar *gtsmodel.MediaAttachment, accountID string) error

	// GetHeaderForAccountID gets the current header for the given account ID.
	// The passed mediaAttachment pointer will be populated with the value of the header, if it exists.
	GetHeaderForAccountID(header *gtsmodel.MediaAttachment, accountID string) error

	// Blocked checks whether a block exists in eiher direction between two accounts.
	// That is, it returns true if account1 blocks account2, OR if account2 blocks account1.
	Blocked(account1 string, account2 string) (bool, error)

	// GetRelationship retrieves the relationship of the targetAccount to the requestingAccount.
	GetRelationship(requestingAccount string, targetAccount string) (*gtsmodel.Relationship, error)

	// StatusVisible returns true if targetStatus is visible to requestingAccount, based on the
	// privacy settings of the status, and any blocks/mutes that might exist between the two accounts
	// or account domains.
	//
	// StatusVisible will also check through the given slice of 'otherRelevantAccounts', which should include:
	//
	// 1. Accounts mentioned in the targetStatus
	//
	// 2. Accounts replied to by the target status
	//
	// 3. Accounts boosted by the target status
	//
	// Will return an error if something goes wrong while pulling stuff out of the database.
	StatusVisible(targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account, relevantAccounts *gtsmodel.RelevantAccounts) (bool, error)

	// Follows returns true if sourceAccount follows target account, or an error if something goes wrong while finding out.
	Follows(sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, error)

	// FollowRequested returns true if sourceAccount has requested to follow target account, or an error if something goes wrong while finding out.
	FollowRequested(sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, error)

	// Mutuals returns true if account1 and account2 both follow each other, or an error if something goes wrong while finding out.
	Mutuals(account1 *gtsmodel.Account, account2 *gtsmodel.Account) (bool, error)

	// PullRelevantAccountsFromStatus returns all accounts mentioned in a status, replied to by a status, or boosted by a status
	PullRelevantAccountsFromStatus(status *gtsmodel.Status) (*gtsmodel.RelevantAccounts, error)

	// GetReplyCountForStatus returns the amount of replies recorded for a status, or an error if something goes wrong
	GetReplyCountForStatus(status *gtsmodel.Status) (int, error)

	// GetReblogCountForStatus returns the amount of reblogs/boosts recorded for a status, or an error if something goes wrong
	GetReblogCountForStatus(status *gtsmodel.Status) (int, error)

	// GetFaveCountForStatus returns the amount of faves/likes recorded for a status, or an error if something goes wrong
	GetFaveCountForStatus(status *gtsmodel.Status) (int, error)

	// StatusFavedBy checks if a given status has been faved by a given account ID
	StatusFavedBy(status *gtsmodel.Status, accountID string) (bool, error)

	// StatusRebloggedBy checks if a given status has been reblogged/boosted by a given account ID
	StatusRebloggedBy(status *gtsmodel.Status, accountID string) (bool, error)

	// StatusMutedBy checks if a given status has been muted by a given account ID
	StatusMutedBy(status *gtsmodel.Status, accountID string) (bool, error)

	// StatusBookmarkedBy checks if a given status has been bookmarked by a given account ID
	StatusBookmarkedBy(status *gtsmodel.Status, accountID string) (bool, error)

	// WhoFavedStatus returns a slice of accounts who faved the given status.
	// This slice will be unfiltered, not taking account of blocks and whatnot, so filter it before serving it back to a user.
	WhoFavedStatus(status *gtsmodel.Status) ([]*gtsmodel.Account, error)

	// WhoBoostedStatus returns a slice of accounts who boosted the given status.
	// This slice will be unfiltered, not taking account of blocks and whatnot, so filter it before serving it back to a user.
	WhoBoostedStatus(status *gtsmodel.Status) ([]*gtsmodel.Account, error)

	GetStatusesWhereFollowing(accountID string, limit int, offsetStatusID string) ([]*gtsmodel.Status, error)

	// GetPublicTimelineForAccount fetches the account's PUBLIC timline -- ie., posts and replies that are public.
	// It will use the given filters and try to return as many statuses as possible up to the limit.
	GetPublicTimelineForAccount(accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, error)

	// GetNotificationsForAccount returns a list of notifications that pertain to the given accountID.
	GetNotificationsForAccount(accountID string, limit int, maxID string, sinceID string) ([]*gtsmodel.Notification, error)

	/*
		USEFUL CONVERSION FUNCTIONS
	*/

	// MentionStringsToMentions takes a slice of deduplicated, lowercase account names in the form "@test@whatever.example.org" for a remote account,
	// or @test for a local account, which have been mentioned in a status.
	// It takes the id of the account that wrote the status, and the id of the status itself, and then
	// checks in the database for the mentioned accounts, and returns a slice of mentions generated based on the given parameters.
	//
	// Note: this func doesn't/shouldn't do any manipulation of the accounts in the DB, it's just for checking
	// if they exist in the db and conveniently returning them if they do.
	MentionStringsToMentions(targetAccounts []string, originAccountID string, statusID string) ([]*gtsmodel.Mention, error)

	// TagStringsToTags takes a slice of deduplicated, lowercase tags in the form "somehashtag", which have been
	// used in a status. It takes the id of the account that wrote the status, and the id of the status itself, and then
	// returns a slice of *model.Tag corresponding to the given tags. If the tag already exists in database, that tag
	// will be returned. Otherwise a pointer to a new tag struct will be created and returned.
	//
	// Note: this func doesn't/shouldn't do any manipulation of the tags in the DB, it's just for checking
	// if they exist in the db already, and conveniently returning them, or creating new tag structs.
	TagStringsToTags(tags []string, originAccountID string, statusID string) ([]*gtsmodel.Tag, error)

	// EmojiStringsToEmojis takes a slice of deduplicated, lowercase emojis in the form ":emojiname:", which have been
	// used in a status. It takes the id of the account that wrote the status, and the id of the status itself, and then
	// returns a slice of *model.Emoji corresponding to the given emojis.
	//
	// Note: this func doesn't/shouldn't do any manipulation of the emoji in the DB, it's just for checking
	// if they exist in the db and conveniently returning them if they do.
	EmojiStringsToEmojis(emojis []string, originAccountID string, statusID string) ([]*gtsmodel.Emoji, error)
}
