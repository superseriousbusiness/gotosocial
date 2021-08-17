package db

import "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

type Status interface {
	// GetStatusByID returns one status from the database, with all rel fields populated (if possible).
	GetStatusByID(id string) (*gtsmodel.Status, DBError)

	// GetStatusByURI returns one status from the database, with all rel fields populated (if possible).
	GetStatusByURI(uri string) (*gtsmodel.Status, DBError)

	// GetReplyCountForStatus returns the amount of replies recorded for a status, or an error if something goes wrong
	GetReplyCountForStatus(status *gtsmodel.Status) (int, DBError)

	// GetReblogCountForStatus returns the amount of reblogs/boosts recorded for a status, or an error if something goes wrong
	GetReblogCountForStatus(status *gtsmodel.Status) (int, DBError)

	// GetFaveCountForStatus returns the amount of faves/likes recorded for a status, or an error if something goes wrong
	GetFaveCountForStatus(status *gtsmodel.Status) (int, DBError)

	// StatusParents get the parent statuses of a given status.
	//
	// If onlyDirect is true, only the immediate parent will be returned.
	StatusParents(status *gtsmodel.Status, onlyDirect bool) ([]*gtsmodel.Status, DBError)

	// StatusChildren gets the child statuses of a given status.
	//
	// If onlyDirect is true, only the immediate children will be returned.
	StatusChildren(status *gtsmodel.Status, onlyDirect bool, minID string) ([]*gtsmodel.Status, DBError)

	// StatusFavedBy checks if a given status has been faved by a given account ID
	StatusFavedBy(status *gtsmodel.Status, accountID string) (bool, DBError)

	// StatusRebloggedBy checks if a given status has been reblogged/boosted by a given account ID
	StatusRebloggedBy(status *gtsmodel.Status, accountID string) (bool, DBError)

	// StatusMutedBy checks if a given status has been muted by a given account ID
	StatusMutedBy(status *gtsmodel.Status, accountID string) (bool, DBError)

	// StatusBookmarkedBy checks if a given status has been bookmarked by a given account ID
	StatusBookmarkedBy(status *gtsmodel.Status, accountID string) (bool, DBError)

	// WhoFavedStatus returns a slice of accounts who faved the given status.
	// This slice will be unfiltered, not taking account of blocks and whatnot, so filter it before serving it back to a user.
	WhoFavedStatus(status *gtsmodel.Status) ([]*gtsmodel.Account, DBError)

	// WhoBoostedStatus returns a slice of accounts who boosted the given status.
	// This slice will be unfiltered, not taking account of blocks and whatnot, so filter it before serving it back to a user.
	WhoBoostedStatus(status *gtsmodel.Status) ([]*gtsmodel.Account, DBError)
}
