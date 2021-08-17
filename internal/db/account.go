package db

import "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

type Account interface {
	// GetAccountByUserID is a shortcut for the common action of fetching an account corresponding to a user ID.
	// The given account pointer will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetAccountByUserID(userID string, account *gtsmodel.Account) DBError

	// GetAccountByID returns one account with the given ID, or an error if something goes wrong.
	GetAccountByID(id string) (*gtsmodel.Account, DBError)

	// GetAccountByURI returns one account with the given URI, or an error if something goes wrong.
	GetAccountByURI(uri string) (*gtsmodel.Account, DBError)

	// GetLocalAccountByUsername is a shortcut for the common action of fetching an account ON THIS INSTANCE
	// according to its username, which should be unique.
	// The given account pointer will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetLocalAccountByUsername(username string, account *gtsmodel.Account) DBError

	// GetAccountFollowRequests is a shortcut for the common action of fetching a list of follow requests targeting the given account ID.
	// The given slice 'followRequests' will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetAccountFollowRequests(accountID string, followRequests *[]gtsmodel.FollowRequest) DBError

	// GetAccountFollowing is a shortcut for the common action of fetching a list of accounts that accountID is following.
	// The given slice 'following' will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetAccountFollowing(accountID string, following *[]gtsmodel.Follow) DBError

	CountAccountFollowing(accountID string, localOnly bool) (int, DBError)

	// GetAccountFollowers is a shortcut for the common action of fetching a list of accounts that accountID is followed by.
	// The given slice 'followers' will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	//
	// If localOnly is set to true, then only followers from *this instance* will be returned.
	GetAccountFollowers(accountID string, followers *[]gtsmodel.Follow, localOnly bool) DBError

	CountAccountFollowers(accountID string, localOnly bool) (int, DBError)

	// GetAccountFaves is a shortcut for the common action of fetching a list of faves made by the given accountID.
	// The given slice 'faves' will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetAccountFaves(accountID string, faves *[]gtsmodel.StatusFave) DBError

	// GetAccountStatusesCount is a shortcut for the common action of counting statuses produced by accountID.
	CountAccountStatuses(accountID string) (int, DBError)

	// GetAccountStatuses is a shortcut for getting the most recent statuses. accountID is optional, if not provided
	// then all statuses will be returned. If limit is set to 0, the size of the returned slice will not be limited. This can
	// be very memory intensive so you probably shouldn't do this!
	// In case of no entries, a 'no entries' error will be returned
	GetAccountStatuses(accountID string, limit int, excludeReplies bool, maxID string, pinnedOnly bool, mediaOnly bool) ([]*gtsmodel.Status, DBError)

	GetAccountBlocks(accountID string, maxID string, sinceID string, limit int) ([]*gtsmodel.Account, string, string, DBError)

	// GetAccountLastStatus simply gets the most recent status by the given account.
	// The given slice 'status' pointer will be set to the result of the query, whatever it is.
	// In case of no entries, a 'no entries' error will be returned
	GetAccountLastStatus(accountID string, status *gtsmodel.Status) DBError

	// SetAccountHeaderOrAvatar sets the header or avatar for the given accountID to the given media attachment.
	SetAccountHeaderOrAvatar(mediaAttachment *gtsmodel.MediaAttachment, accountID string) DBError

	// GetInstanceAccount returns the instance account for the given domain.
	// If domain is empty, this instance account will be returned.
	GetInstanceAccount(domain string) (*gtsmodel.Account, DBError)
}
