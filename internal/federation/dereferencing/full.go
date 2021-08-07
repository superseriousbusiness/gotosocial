package dereferencing

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

func (d *deref) FullyDereferenceStatusableAndAccount(username string, statusable typeutils.Statusable) error {
	l := d.log.WithFields(logrus.Fields{
		"func":     "FullyDereferenceStatusableAndAccount",
		"username": username,
	})

	idProp := statusable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return errors.New("FullyDereferenceStatusableAndAccount: couldn't extract iri from statusable")
	}
	statusURI := idProp.GetIRI()

	// make sure we don't already have this status in our db
	if err := d.db.GetWhere([]db.Where{{Key: "uri", Value: statusURI.String()}}, &gtsmodel.Status{}); err == nil {
		// we already have it
		l.Debugf("status with uri %s is already in the database", statusURI.String())
		return nil
	}

	// make sure we have the author account in the db
	attributedToProp := statusable.GetActivityStreamsAttributedTo()
	for iter := attributedToProp.Begin(); iter != attributedToProp.End(); iter = iter.Next() {
		if !iter.IsIRI() {
			continue
		}

		accountURI := iter.GetIRI()
		if err := d.db.GetWhere([]db.Where{{Key: "uri", Value: accountURI.String()}}, &gtsmodel.Account{}); err == nil {
			// we already have it, nice
			l.Debugf("account with uri %s is already in the database", accountURI.String())
			continue
		}

		// we don't have the status author account yet so dereference it
		accountable, err := d.DereferenceAccountable(username, accountURI)
		if err != nil {
			return fmt.Errorf("FullyDereferenceStatusAndAccount: error dereferencing remote account with id %s: %s", accountURI.String(), err)
		}
		account, err := d.typeConverter.ASRepresentationToAccount(accountable, false)
		if err != nil {
			return fmt.Errorf("FullyDereferenceStatusAndAccount: error converting dereferenced account with id %s into account : %s", accountURI.String(), err)
		}

		accountID, err := id.NewRandomULID()
		if err != nil {
			return err
		}
		account.ID = accountID

		if err := d.db.Put(account); err != nil {
			return fmt.Errorf("FullyDereferenceStatusAndAccount: error putting dereferenced account with id %s into database : %s", accountURI.String(), err)
		}

		if err := d.PopulateAccountFields(account, username, false); err != nil {
			return fmt.Errorf("FullyDereferenceStatusAndAccount: error dereferencing fields on account with id %s : %s", accountURI.String(), err)
		}
	}

	gtsStatus, err := d.typeConverter.ASStatusToStatus(statusable)
	if err != nil {
		return fmt.Errorf("FullyDereferenceStatusAndAccount: error converting statusable: %s", err)
	}

	id, err := id.NewULIDFromTime(gtsStatus.CreatedAt)
	if err != nil {
		return fmt.Errorf("FullyDereferenceStatusAndAccount: error generating id: %s", err)
	}
	gtsStatus.ID = id

	if err := d.db.Put(gtsStatus); err != nil {
		return fmt.Errorf("FullyDereferenceStatusAndAccount: error putting dereferenced status with id %s into the db: %s", gtsStatus.URI, err)
	}
	l.Debugf("put status %s in the db", gtsStatus.URI)

	// now dereference additional fields straight away (we're already async here so we have time)
	if err := d.PopulateStatusFields(gtsStatus, username); err != nil {
		return fmt.Errorf("FullyDereferenceStatusAndAccount: error dereferencing status fields for status with id %s: %s", gtsStatus.URI, err)
	}

	// update with the newly dereferenced fields
	if err := d.db.UpdateByID(gtsStatus.ID, gtsStatus); err != nil {
		return fmt.Errorf("FullyDereferenceStatusAndAccount: error updating dereferenced status in the db: %s", err)
	}
	l.Debugf("updated status %s in the db", gtsStatus.URI)

	return nil
}
