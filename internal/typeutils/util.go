package typeutils

import (
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (c *converter) interactionsWithStatusForAccount(s *gtsmodel.Status, requestingAccount *gtsmodel.Account) (*statusInteractions, error) {
	si := &statusInteractions{}

	if requestingAccount != nil {
		faved, err := c.db.IsStatusFavedBy(s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has faved status: %s", err)
		}
		si.Faved = faved

		reblogged, err := c.db.IsStatusRebloggedBy(s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has reblogged status: %s", err)
		}
		si.Reblogged = reblogged

		muted, err := c.db.IsStatusMutedBy(s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has muted status: %s", err)
		}
		si.Muted = muted

		bookmarked, err := c.db.IsStatusBookmarkedBy(s, requestingAccount.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking if requesting account has bookmarked status: %s", err)
		}
		si.Bookmarked = bookmarked
	}
	return si, nil
}

// StatusInteractions denotes interactions with a status on behalf of an account.
type statusInteractions struct {
	Faved      bool
	Muted      bool
	Bookmarked bool
	Reblogged  bool
}
