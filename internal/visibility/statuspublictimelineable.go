package visibility

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *filter) StatusPublictimelineable(targetStatus *gtsmodel.Status, timelineOwnerAccount *gtsmodel.Account) (bool, error) {
	l := f.log.WithFields(logrus.Fields{
		"func":     "StatusPublictimelineable",
		"statusID": targetStatus.ID,
	})

	// Don't timeline a reply
	if targetStatus.InReplyToURI != "" || targetStatus.InReplyToID != "" || targetStatus.InReplyToAccountID != "" {
		return false, nil
	}

	// status owner should always be able to see their own status in their timeline so we can return early if this is the case
	if timelineOwnerAccount != nil && targetStatus.AccountID == timelineOwnerAccount.ID {
		return true, nil
	}

	v, err := f.StatusVisible(targetStatus, timelineOwnerAccount)
	if err != nil {
		return false, fmt.Errorf("StatusPublictimelineable: error checking visibility of status with id %s: %s", targetStatus.ID, err)
	}

	if !v {
		l.Debug("status is not publicTimelineable because it's not visible to the requester")
		return false, nil
	}

	return true, nil
}
