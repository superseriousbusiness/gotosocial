package dereferencing

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (d *deref) DereferenceAnnounce(announce *gtsmodel.Status, requestingUsername string) error {
	if announce.GTSBoostedStatus == nil || announce.GTSBoostedStatus.URI == "" {
		// we can't do anything unfortunately
		return errors.New("DereferenceAnnounce: no URI to dereference")
	}

	boostedStatusURI, err := url.Parse(announce.GTSBoostedStatus.URI)
	if err != nil {
		return fmt.Errorf("DereferenceAnnounce: couldn't parse boosted status URI %s: %s", announce.GTSBoostedStatus.URI, err)
	}
	if blocked, err := d.blockedDomain(boostedStatusURI.Host); blocked || err != nil {
		return fmt.Errorf("DereferenceAnnounce: domain %s is blocked", boostedStatusURI.Host)
	}

	// dereference statuses in the thread of the boosted status
	if err := d.DereferenceThread(requestingUsername, boostedStatusURI); err != nil {
		return fmt.Errorf("DereferenceAnnounce: error dereferencing thread of boosted status: %s", err)
	}

	boostedStatus, _, _, err := d.GetRemoteStatus(requestingUsername, boostedStatusURI, false)
	if err != nil {
		return fmt.Errorf("DereferenceAnnounce: error dereferencing remote status with id %s: %s", announce.GTSBoostedStatus.URI, err)
	}

	announce.Content = boostedStatus.Content
	announce.ContentWarning = boostedStatus.ContentWarning
	announce.ActivityStreamsType = boostedStatus.ActivityStreamsType
	announce.Sensitive = boostedStatus.Sensitive
	announce.Language = boostedStatus.Language
	announce.Text = boostedStatus.Text
	announce.BoostOfID = boostedStatus.ID
	announce.BoostOfAccountID = boostedStatus.AccountID
	announce.Visibility = boostedStatus.Visibility
	announce.VisibilityAdvanced = boostedStatus.VisibilityAdvanced
	announce.GTSBoostedStatus = boostedStatus
	return nil
}
