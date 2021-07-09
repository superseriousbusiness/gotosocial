package visibility

import (
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Filter packages up a bunch of logic for checking whether given statuses or accounts are visible to a requester.
type Filter interface {
	// StatusVisible returns true if targetStatus is visible to requestingAccount, based on the
	// privacy settings of the status, and any blocks/mutes that might exist between the two accounts
	// or account domains, and other relevant accounts mentioned in or replied to by the status.
	StatusVisible(targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account) (bool, error)

	// StatusHometimelineable returns true if targetStatus should be in the home timeline of the requesting account.
	//
	// This function will call StatusVisible internally, so it's not necessary to call it beforehand.
	StatusHometimelineable(targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account) (bool, error)

	// StatusPublictimelineable returns true if targetStatus should be in the public timeline of the requesting account.
	//
	// This function will call StatusVisible internally, so it's not necessary to call it beforehand.
	StatusPublictimelineable(targetStatus *gtsmodel.Status, timelineOwnerAccount *gtsmodel.Account) (bool, error)
}

type filter struct {
	db  db.DB
	log *logrus.Logger
}

// NewFilter returns a new Filter interface that will use the provided database and logger.
func NewFilter(db db.DB, log *logrus.Logger) Filter {
	return &filter{
		db:  db,
		log: log,
	}
}
