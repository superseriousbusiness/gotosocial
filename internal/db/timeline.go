package db

import "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

type Timeline interface {
	// GetHomeTimelineForAccount returns a slice of statuses from accounts that are followed by the given account id.
	//
	// Statuses should be returned in descending order of when they were created (newest first).
	GetHomeTimelineForAccount(accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, DBError)

	// GetPublicTimelineForAccount fetches the account's PUBLIC timeline -- ie., posts and replies that are public.
	// It will use the given filters and try to return as many statuses as possible up to the limit.
	//
	// Statuses should be returned in descending order of when they were created (newest first).
	GetPublicTimelineForAccount(accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, DBError)

	// GetFavedTimelineForAccount fetches the account's FAVED timeline -- ie., posts and replies that the requesting account has faved.
	// It will use the given filters and try to return as many statuses as possible up to the limit.
	//
	// Note that unlike the other GetTimeline functions, the returned statuses will be arranged by their FAVE id, not the STATUS id.
	// In other words, they'll be returned in descending order of when they were faved by the requesting user, not when they were created.
	//
	// Also note the extra return values, which correspond to the nextMaxID and prevMinID for building Link headers.
	GetFavedTimelineForAccount(accountID string, maxID string, minID string, limit int) ([]*gtsmodel.Status, string, string, DBError)
}
