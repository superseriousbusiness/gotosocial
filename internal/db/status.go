package db

import "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

type Status interface {
	// GetReplyCountForStatus returns the amount of replies recorded for a status, or an error if something goes wrong
	GetReplyCountForStatus(status *gtsmodel.Status) (int, error)

	// GetReblogCountForStatus returns the amount of reblogs/boosts recorded for a status, or an error if something goes wrong
	GetReblogCountForStatus(status *gtsmodel.Status) (int, error)

	// GetFaveCountForStatus returns the amount of faves/likes recorded for a status, or an error if something goes wrong
	GetFaveCountForStatus(status *gtsmodel.Status) (int, error)

	// StatusParents get the parent statuses of a given status.
	//
	// If onlyDirect is true, only the immediate parent will be returned.
	StatusParents(status *gtsmodel.Status, onlyDirect bool) ([]*gtsmodel.Status, error)

	// StatusChildren gets the child statuses of a given status.
	//
	// If onlyDirect is true, only the immediate children will be returned.
	StatusChildren(status *gtsmodel.Status, onlyDirect bool, minID string) ([]*gtsmodel.Status, error)

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
}
