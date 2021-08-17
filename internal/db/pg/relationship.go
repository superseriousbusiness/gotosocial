package pg

import (
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (ps *postgresService) Blocked(account1 string, account2 string) (bool, error) {
	// TODO: check domain blocks as well
	var blocked bool
	if err := ps.conn.Model(&gtsmodel.Block{}).
		Where("account_id = ?", account1).Where("target_account_id = ?", account2).
		WhereOr("target_account_id = ?", account1).Where("account_id = ?", account2).
		Select(); err != nil {
		if err == pg.ErrNoRows {
			blocked = false
			return blocked, nil
		}
		return blocked, err
	}
	blocked = true
	return blocked, nil
}

func (ps *postgresService) GetRelationship(requestingAccount string, targetAccount string) (*gtsmodel.Relationship, error) {
	r := &gtsmodel.Relationship{
		ID: targetAccount,
	}

	// check if the requesting account follows the target account
	follow := &gtsmodel.Follow{}
	if err := ps.conn.Model(follow).Where("account_id = ?", requestingAccount).Where("target_account_id = ?", targetAccount).Select(); err != nil {
		if err != pg.ErrNoRows {
			// a proper error
			return nil, fmt.Errorf("getrelationship: error checking follow existence: %s", err)
		}
		// no follow exists so these are all false
		r.Following = false
		r.ShowingReblogs = false
		r.Notifying = false
	} else {
		// follow exists so we can fill these fields out...
		r.Following = true
		r.ShowingReblogs = follow.ShowReblogs
		r.Notifying = follow.Notify
	}

	// check if the target account follows the requesting account
	followedBy, err := ps.conn.Model(&gtsmodel.Follow{}).Where("account_id = ?", targetAccount).Where("target_account_id = ?", requestingAccount).Exists()
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking followed_by existence: %s", err)
	}
	r.FollowedBy = followedBy

	// check if the requesting account blocks the target account
	blocking, err := ps.conn.Model(&gtsmodel.Block{}).Where("account_id = ?", requestingAccount).Where("target_account_id = ?", targetAccount).Exists()
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking blocking existence: %s", err)
	}
	r.Blocking = blocking

	// check if the target account blocks the requesting account
	blockedBy, err := ps.conn.Model(&gtsmodel.Block{}).Where("account_id = ?", targetAccount).Where("target_account_id = ?", requestingAccount).Exists()
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking blocked existence: %s", err)
	}
	r.BlockedBy = blockedBy

	// check if there's a pending following request from requesting account to target account
	requested, err := ps.conn.Model(&gtsmodel.FollowRequest{}).Where("account_id = ?", requestingAccount).Where("target_account_id = ?", targetAccount).Exists()
	if err != nil {
		return nil, fmt.Errorf("getrelationship: error checking blocked existence: %s", err)
	}
	r.Requested = requested

	return r, nil
}

func (ps *postgresService) Follows(sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, error) {
	if sourceAccount == nil || targetAccount == nil {
		return false, nil
	}

	return ps.conn.Model(&gtsmodel.Follow{}).Where("account_id = ?", sourceAccount.ID).Where("target_account_id = ?", targetAccount.ID).Exists()
}

func (ps *postgresService) FollowRequested(sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, error) {
	if sourceAccount == nil || targetAccount == nil {
		return false, nil
	}

	return ps.conn.Model(&gtsmodel.FollowRequest{}).Where("account_id = ?", sourceAccount.ID).Where("target_account_id = ?", targetAccount.ID).Exists()
}

func (ps *postgresService) Mutuals(account1 *gtsmodel.Account, account2 *gtsmodel.Account) (bool, error) {
	if account1 == nil || account2 == nil {
		return false, nil
	}

	// make sure account 1 follows account 2
	f1, err := ps.conn.Model(&gtsmodel.Follow{}).Where("account_id = ?", account1.ID).Where("target_account_id = ?", account2.ID).Exists()
	if err != nil {
		if err == pg.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	// make sure account 2 follows account 1
	f2, err := ps.conn.Model(&gtsmodel.Follow{}).Where("account_id = ?", account2.ID).Where("target_account_id = ?", account1.ID).Exists()
	if err != nil {
		if err == pg.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return f1 && f2, nil
}

func (ps *postgresService) AcceptFollowRequest(originAccountID string, targetAccountID string) (*gtsmodel.Follow, error) {
	// make sure the original follow request exists
	fr := &gtsmodel.FollowRequest{}
	if err := ps.conn.Model(fr).Where("account_id = ?", originAccountID).Where("target_account_id = ?", targetAccountID).Select(); err != nil {
		if err == pg.ErrMultiRows {
			return nil, db.ErrNoEntries{}
		}
		return nil, err
	}

	// create a new follow to 'replace' the request with
	follow := &gtsmodel.Follow{
		ID:              fr.ID,
		AccountID:       originAccountID,
		TargetAccountID: targetAccountID,
		URI:             fr.URI,
	}

	// if the follow already exists, just update the URI -- we don't need to do anything else
	if _, err := ps.conn.Model(follow).OnConflict("ON CONSTRAINT follows_account_id_target_account_id_key DO UPDATE set uri = ?", follow.URI).Insert(); err != nil {
		return nil, err
	}

	// now remove the follow request
	if _, err := ps.conn.Model(&gtsmodel.FollowRequest{}).Where("account_id = ?", originAccountID).Where("target_account_id = ?", targetAccountID).Delete(); err != nil {
		return nil, err
	}

	return follow, nil
}
