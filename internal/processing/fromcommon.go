/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package processing

import (
	"fmt"
	"strings"
	"sync"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) notifyStatus(status *gtsmodel.Status) error {
	// if there are no mentions in this status then just bail
	if len(status.Mentions) == 0 {
		return nil
	}

	if status.GTSMentions == nil {
		// there are mentions but they're not fully populated on the status yet so do this
		menchies := []*gtsmodel.Mention{}
		for _, m := range status.Mentions {
			gtsm := &gtsmodel.Mention{}
			if err := p.db.GetByID(m, gtsm); err != nil {
				return fmt.Errorf("notifyStatus: error getting mention with id %s from the db: %s", m, err)
			}
			menchies = append(menchies, gtsm)
		}
		status.GTSMentions = menchies
	}

	// now we have mentions as full gtsmodel.Mention structs on the status we can continue
	for _, m := range status.GTSMentions {
		// make sure this is a local account, otherwise we don't need to create a notification for it
		if m.GTSAccount == nil {
			a := &gtsmodel.Account{}
			if err := p.db.GetByID(m.TargetAccountID, a); err != nil {
				// we don't have the account or there's been an error
				return fmt.Errorf("notifyStatus: error getting account with id %s from the db: %s", m.TargetAccountID, err)
			}
			m.GTSAccount = a
		}
		if m.GTSAccount.Domain != "" {
			// not a local account so skip it
			continue
		}

		// make sure a notif doesn't already exist for this mention
		err := p.db.GetWhere([]db.Where{
			{Key: "notification_type", Value: gtsmodel.NotificationMention},
			{Key: "target_account_id", Value: m.TargetAccountID},
			{Key: "origin_account_id", Value: status.AccountID},
			{Key: "status_id", Value: status.ID},
		}, &gtsmodel.Notification{})
		if err == nil {
			// notification exists already so just continue
			continue
		}
		if _, ok := err.(db.ErrNoEntries); !ok {
			// there's a real error in the db
			return fmt.Errorf("notifyStatus: error checking existence of notification for mention with id %s : %s", m.ID, err)
		}

		// if we've reached this point we know the mention is for a local account, and the notification doesn't exist, so create it
		notif := &gtsmodel.Notification{
			NotificationType: gtsmodel.NotificationMention,
			TargetAccountID:  m.TargetAccountID,
			OriginAccountID:  status.AccountID,
			StatusID:         status.ID,
		}

		if err := p.db.Put(notif); err != nil {
			return fmt.Errorf("notifyStatus: error putting notification in database: %s", err)
		}
	}

	return nil
}

func (p *processor) notifyFollowRequest(followRequest *gtsmodel.FollowRequest, receivingAccount *gtsmodel.Account) error {
	// return if this isn't a local account
	if receivingAccount.Domain != "" {
		return nil
	}

	notif := &gtsmodel.Notification{
		NotificationType: gtsmodel.NotificationFollowRequest,
		TargetAccountID:  followRequest.TargetAccountID,
		OriginAccountID:  followRequest.AccountID,
	}

	if err := p.db.Put(notif); err != nil {
		return fmt.Errorf("notifyFollowRequest: error putting notification in database: %s", err)
	}

	return nil
}

func (p *processor) notifyFollow(follow *gtsmodel.Follow, receivingAccount *gtsmodel.Account) error {
	// return if this isn't a local account
	if receivingAccount.Domain != "" {
		return nil
	}

	// first remove the follow request notification
	if err := p.db.DeleteWhere([]db.Where{
		{Key: "notification_type", Value: gtsmodel.NotificationFollowRequest},
		{Key: "target_account_id", Value: follow.TargetAccountID},
		{Key: "origin_account_id", Value: follow.AccountID},
	}, &gtsmodel.Notification{}); err != nil {
		return fmt.Errorf("notifyFollow: error removing old follow request notification from database: %s", err)
	}

	// now create the new follow notification
	notif := &gtsmodel.Notification{
		NotificationType: gtsmodel.NotificationFollow,
		TargetAccountID:  follow.TargetAccountID,
		OriginAccountID:  follow.AccountID,
	}
	if err := p.db.Put(notif); err != nil {
		return fmt.Errorf("notifyFollow: error putting notification in database: %s", err)
	}

	return nil
}

func (p *processor) notifyFave(fave *gtsmodel.StatusFave, receivingAccount *gtsmodel.Account) error {
	// return if this isn't a local account
	if receivingAccount.Domain != "" {
		return nil
	}

	notif := &gtsmodel.Notification{
		NotificationType: gtsmodel.NotificationFave,
		TargetAccountID:  fave.TargetAccountID,
		OriginAccountID:  fave.AccountID,
		StatusID:         fave.StatusID,
	}

	if err := p.db.Put(notif); err != nil {
		return fmt.Errorf("notifyFave: error putting notification in database: %s", err)
	}

	return nil
}

func (p *processor) notifyAnnounce(status *gtsmodel.Status) error {
	if status.BoostOfID == "" {
		// not a boost, nothing to do
		return nil
	}

	boostedStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(status.BoostOfID, boostedStatus); err != nil {
		return fmt.Errorf("notifyAnnounce: error getting status with id %s: %s", status.BoostOfID, err)
	}

	boostedAcct := &gtsmodel.Account{}
	if err := p.db.GetByID(boostedStatus.AccountID, boostedAcct); err != nil {
		return fmt.Errorf("notifyAnnounce: error getting account with id %s: %s", boostedStatus.AccountID, err)
	}

	if boostedAcct.Domain != "" {
		// remote account, nothing to do
		return nil
	}

	if boostedStatus.AccountID == status.AccountID {
		// it's a self boost, nothing to do
		return nil
	}

	// make sure a notif doesn't already exist for this announce
	err := p.db.GetWhere([]db.Where{
		{Key: "notification_type", Value: gtsmodel.NotificationReblog},
		{Key: "target_account_id", Value: boostedAcct.ID},
		{Key: "origin_account_id", Value: status.AccountID},
		{Key: "status_id", Value: status.ID},
	}, &gtsmodel.Notification{})
	if err == nil {
		// notification exists already so just bail
		return nil
	}

	// now create the new reblog notification
	notif := &gtsmodel.Notification{
		NotificationType: gtsmodel.NotificationReblog,
		TargetAccountID:  boostedAcct.ID,
		OriginAccountID:  status.AccountID,
		StatusID:         status.ID,
	}

	if err := p.db.Put(notif); err != nil {
		return fmt.Errorf("notifyAnnounce: error putting notification in database: %s", err)
	}

	return nil
}

func (p *processor) timelineStatus(status *gtsmodel.Status) error {
	// make sure the author account is pinned onto the status
	if status.GTSAuthorAccount == nil {
		a := &gtsmodel.Account{}
		if err := p.db.GetByID(status.AccountID, a); err != nil {
			return fmt.Errorf("timelineStatus: error getting author account with id %s: %s", status.AccountID, err)
		}
		status.GTSAuthorAccount = a
	}

	// get all relevant accounts here once
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(status)
	if err != nil {
		return fmt.Errorf("timelineStatus: error getting relevant accounts from status: %s", err)
	}

	// get local followers of the account that posted the status
	followers := []gtsmodel.Follow{}
	if err := p.db.GetFollowersByAccountID(status.AccountID, &followers, true); err != nil {
		return fmt.Errorf("timelineStatus: error getting followers for account id %s: %s", status.AccountID, err)
	}

	// if the poster is local, add a fake entry for them to the followers list so they can see their own status in their timeline
	if status.GTSAuthorAccount.Domain == "" {
		followers = append(followers, gtsmodel.Follow{
			AccountID: status.AccountID,
		})
	}

	wg := sync.WaitGroup{}
	wg.Add(len(followers))
	errors := make(chan error, len(followers))

	for _, f := range followers {
		go p.timelineStatusForAccount(status, f.AccountID, relevantAccounts, errors, &wg)
	}

	// read any errors that come in from the async functions
	errs := []string{}
	go func() {
		for range errors {
			e := <-errors
			if e != nil {
				errs = append(errs, e.Error())
			}
		}
	}()

	// wait til all functions have returned and then close the error channel
	wg.Wait()
	close(errors)

	if len(errs) != 0 {
		// we have some errors
		return fmt.Errorf("timelineStatus: one or more errors timelining statuses: %s", strings.Join(errs, ";"))
	}

	// no errors, nice
	return nil
}

func (p *processor) timelineStatusForAccount(status *gtsmodel.Status, accountID string, relevantAccounts *gtsmodel.RelevantAccounts, errors chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	// get the targetAccount
	timelineAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(accountID, timelineAccount); err != nil {
		errors <- fmt.Errorf("timelineStatus: error getting account for timeline with id %s: %s", accountID, err)
		return
	}

	// make sure the status is visible
	visible, err := p.db.StatusVisible(status, status.GTSAuthorAccount, timelineAccount, relevantAccounts)
	if err != nil {
		errors <- fmt.Errorf("timelineStatus: error getting visibility for status for timeline with id %s: %s", accountID, err)
		return
	}

	if !visible {
		return
	}

	if err := p.timelineManager.IngestAndPrepare(status, timelineAccount.ID); err != nil {
		errors <- fmt.Errorf("initTimelineFor: error ingesting status %s: %s", status.ID, err)
	}
}

func (p *processor) fullyDeleteStatus(status *gtsmodel.Status, accountID string) error {
	return nil
}
