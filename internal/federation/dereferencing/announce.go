package dereferencing

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
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

	// check if we already have the boosted status in the database
	boostedStatus := &gtsmodel.Status{}
	err = d.db.GetWhere([]db.Where{{Key: "uri", Value: announce.GTSBoostedStatus.URI}}, boostedStatus)
	if err == nil {
		// nice, we already have it so we don't actually need to dereference it from remote
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

	// we don't have it so we need to dereference it
	statusable, err := d.DereferenceStatusable(requestingUsername, boostedStatusURI)
	if err != nil {
		return fmt.Errorf("dereferenceAnnounce: error dereferencing remote status with id %s: %s", announce.GTSBoostedStatus.URI, err)
	}

	// make sure we have the author account in the db
	attributedToProp := statusable.GetActivityStreamsAttributedTo()
	for iter := attributedToProp.Begin(); iter != attributedToProp.End(); iter = iter.Next() {
		accountURI := iter.GetIRI()
		if accountURI == nil {
			continue
		}

		if err := d.db.GetWhere([]db.Where{{Key: "uri", Value: accountURI.String()}}, &gtsmodel.Account{}); err == nil {
			// we already have it, fine
			continue
		}

		// we don't have the boosted status author account yet so dereference it
		accountable, err := d.DereferenceAccountable(requestingUsername, accountURI)
		if err != nil {
			return fmt.Errorf("dereferenceAnnounce: error dereferencing remote account with id %s: %s", accountURI.String(), err)
		}
		account, err := d.typeConverter.ASRepresentationToAccount(accountable, false)
		if err != nil {
			return fmt.Errorf("dereferenceAnnounce: error converting dereferenced account with id %s into account : %s", accountURI.String(), err)
		}

		accountID, err := id.NewRandomULID()
		if err != nil {
			return err
		}
		account.ID = accountID

		if err := d.db.Put(account); err != nil {
			return fmt.Errorf("dereferenceAnnounce: error putting dereferenced account with id %s into database : %s", accountURI.String(), err)
		}

		if err := d.PopulateAccountFields(account, requestingUsername, false); err != nil {
			return fmt.Errorf("dereferenceAnnounce: error dereferencing fields on account with id %s : %s", accountURI.String(), err)
		}
	}

	// now convert the statusable into something we can understand
	boostedStatus, err = d.typeConverter.ASStatusToStatus(statusable)
	if err != nil {
		return fmt.Errorf("dereferenceAnnounce: error converting dereferenced statusable with id %s into status : %s", announce.GTSBoostedStatus.URI, err)
	}

	boostedStatusID, err := id.NewULIDFromTime(boostedStatus.CreatedAt)
	if err != nil {
		return nil
	}
	boostedStatus.ID = boostedStatusID

	if err := d.db.Put(boostedStatus); err != nil {
		return fmt.Errorf("dereferenceAnnounce: error putting dereferenced status with id %s into the db: %s", announce.GTSBoostedStatus.URI, err)
	}

	// now dereference additional fields straight away (we're already async here so we have time)
	if err := d.PopulateStatusFields(boostedStatus, requestingUsername); err != nil {
		return fmt.Errorf("dereferenceAnnounce: error dereferencing status fields for status with id %s: %s", announce.GTSBoostedStatus.URI, err)
	}

	// update with the newly dereferenced fields
	if err := d.db.UpdateByID(boostedStatus.ID, boostedStatus); err != nil {
		return fmt.Errorf("dereferenceAnnounce: error updating dereferenced status in the db: %s", err)
	}

	// we have everything we need!
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
