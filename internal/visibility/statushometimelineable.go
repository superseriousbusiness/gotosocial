package visibility

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *filter) StatusHometimelineable(targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account) (bool, error) {
	l := f.log.WithFields(logrus.Fields{
		"func":                "StatusHometimelineable",
		"statusID":            targetStatus.ID,
		"requestingAccountID": requestingAccount.ID,
	})

	// status owner should always be able to see their status in their timeline so we can return early if this is the case
	if targetStatus.AccountID == requestingAccount.ID {
		return true, nil
	}

	v, err := f.StatusVisible(targetStatus, requestingAccount)
	if err != nil {
		return false, fmt.Errorf("StatusHometimelineable: error checking visibility of status with id %s: %s", targetStatus.ID, err)
	}

	if !v {
		l.Debug("status is not hometimelineable because it's not visible to the requester")
		return false, nil
	}

	// we don't want to timeline a reply to a status whose owner isn't followed by the requesting account
	if targetStatus.InReplyToID != "" {
		// pin the reply to status on to this status if it hasn't been done already
		if targetStatus.GTSReplyToStatus == nil {
			rs := &gtsmodel.Status{}
			if err := f.db.GetByID(targetStatus.InReplyToID, rs); err != nil {
				return false, fmt.Errorf("StatusHometimelineable: error getting replied to status with id %s: %s", targetStatus.InReplyToID, err)
			}
			targetStatus.GTSReplyToStatus = rs
		}

		// pin the reply to account on to this status if it hasn't been done already
		if targetStatus.GTSReplyToAccount == nil {
			ra := &gtsmodel.Account{}
			if err := f.db.GetByID(targetStatus.InReplyToAccountID, ra); err != nil {
				return false, fmt.Errorf("StatusHometimelineable: error getting replied to account with id %s: %s", targetStatus.InReplyToAccountID, err)
			}
			targetStatus.GTSReplyToAccount = ra
		}

		// make sure the requesting account follows the replied-to account
		follows, err := f.db.Follows(requestingAccount, targetStatus.GTSReplyToAccount)
		if err != nil {
			return false, fmt.Errorf("StatusHometimelineable: error checking follow from account %s to account %s: %s", requestingAccount.ID, targetStatus.InReplyToAccountID, err)
		}

		if !follows {
			return false, nil
		}
	}

	return true, nil
}
