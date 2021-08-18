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
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func (p *processor) notifyStatus(status *gtsmodel.Status) error {
	// if there are no mentions in this status then just bail
	if len(status.MentionIDs) == 0 {
		return nil
	}

	if status.Mentions == nil {
		// there are mentions but they're not fully populated on the status yet so do this
		menchies, err := p.db.GetMentions(status.MentionIDs)
		if err != nil {
			return fmt.Errorf("notifyStatus: error getting mentions for status %s from the db: %s", status.ID, err)
		}
		status.Mentions = menchies
	}

	// now we have mentions as full gtsmodel.Mention structs on the status we can continue
	for _, m := range status.Mentions {
		// make sure this is a local account, otherwise we don't need to create a notification for it
		if m.TargetAccount == nil {
			a, err := p.db.GetAccountByID(m.TargetAccountID)
			if err != nil {
				// we don't have the account or there's been an error
				return fmt.Errorf("notifyStatus: error getting account with id %s from the db: %s", m.TargetAccountID, err)
			}
			m.TargetAccount = a
		}
		if m.TargetAccount.Domain != "" {
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
		if err != db.ErrNoEntries {
			// there's a real error in the db
			return fmt.Errorf("notifyStatus: error checking existence of notification for mention with id %s : %s", m.ID, err)
		}

		// if we've reached this point we know the mention is for a local account, and the notification doesn't exist, so create it
		notifID, err := id.NewULID()
		if err != nil {
			return err
		}

		notif := &gtsmodel.Notification{
			ID:               notifID,
			NotificationType: gtsmodel.NotificationMention,
			TargetAccountID:  m.TargetAccountID,
			TargetAccount:    m.TargetAccount,
			OriginAccountID:  status.AccountID,
			OriginAccount:    status.Account,
			StatusID:         status.ID,
			Status:           status,
		}

		if err := p.db.Put(notif); err != nil {
			return fmt.Errorf("notifyStatus: error putting notification in database: %s", err)
		}

		// now stream the notification to the user
		mastoNotif, err := p.tc.NotificationToMasto(notif)
		if err != nil {
			return fmt.Errorf("notifyStatus: error converting notification to masto representation: %s", err)
		}

		if err := p.streamingProcessor.StreamNotificationToAccount(mastoNotif, m.OriginAccount); err != nil {
			return fmt.Errorf("notifyStatus: error streaming notification to account: %s", err)
		}
	}

	return nil
}

func (p *processor) notifyFollowRequest(followRequest *gtsmodel.FollowRequest, receivingAccount *gtsmodel.Account) error {
	// return if this isn't a local account
	if receivingAccount.Domain != "" {
		return nil
	}

	notifID, err := id.NewULID()
	if err != nil {
		return err
	}

	notif := &gtsmodel.Notification{
		ID:               notifID,
		NotificationType: gtsmodel.NotificationFollowRequest,
		TargetAccountID:  followRequest.TargetAccountID,
		OriginAccountID:  followRequest.AccountID,
	}

	if err := p.db.Put(notif); err != nil {
		return fmt.Errorf("notifyFollowRequest: error putting notification in database: %s", err)
	}

	// now stream the notification to the user
	mastoNotif, err := p.tc.NotificationToMasto(notif)
	if err != nil {
		return fmt.Errorf("notifyStatus: error converting notification to masto representation: %s", err)
	}

	if err := p.streamingProcessor.StreamNotificationToAccount(mastoNotif, receivingAccount); err != nil {
		return fmt.Errorf("notifyStatus: error streaming notification to account: %s", err)
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
	notifID, err := id.NewULID()
	if err != nil {
		return err
	}

	notif := &gtsmodel.Notification{
		ID:               notifID,
		NotificationType: gtsmodel.NotificationFollow,
		TargetAccountID:  follow.TargetAccountID,
		OriginAccountID:  follow.AccountID,
	}
	if err := p.db.Put(notif); err != nil {
		return fmt.Errorf("notifyFollow: error putting notification in database: %s", err)
	}

	// now stream the notification to the user
	mastoNotif, err := p.tc.NotificationToMasto(notif)
	if err != nil {
		return fmt.Errorf("notifyStatus: error converting notification to masto representation: %s", err)
	}

	if err := p.streamingProcessor.StreamNotificationToAccount(mastoNotif, receivingAccount); err != nil {
		return fmt.Errorf("notifyStatus: error streaming notification to account: %s", err)
	}

	return nil
}

func (p *processor) notifyFave(fave *gtsmodel.StatusFave, receivingAccount *gtsmodel.Account) error {
	// return if this isn't a local account
	if receivingAccount.Domain != "" {
		return nil
	}

	notifID, err := id.NewULID()
	if err != nil {
		return err
	}

	notif := &gtsmodel.Notification{
		ID:               notifID,
		NotificationType: gtsmodel.NotificationFave,
		TargetAccountID:  fave.TargetAccountID,
		OriginAccountID:  fave.AccountID,
		StatusID:         fave.StatusID,
	}

	if err := p.db.Put(notif); err != nil {
		return fmt.Errorf("notifyFave: error putting notification in database: %s", err)
	}

	// now stream the notification to the user
	mastoNotif, err := p.tc.NotificationToMasto(notif)
	if err != nil {
		return fmt.Errorf("notifyStatus: error converting notification to masto representation: %s", err)
	}

	if err := p.streamingProcessor.StreamNotificationToAccount(mastoNotif, receivingAccount); err != nil {
		return fmt.Errorf("notifyStatus: error streaming notification to account: %s", err)
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
	notifID, err := id.NewULID()
	if err != nil {
		return err
	}

	notif := &gtsmodel.Notification{
		ID:               notifID,
		NotificationType: gtsmodel.NotificationReblog,
		TargetAccountID:  boostedAcct.ID,
		OriginAccountID:  status.AccountID,
		StatusID:         status.ID,
	}

	if err := p.db.Put(notif); err != nil {
		return fmt.Errorf("notifyAnnounce: error putting notification in database: %s", err)
	}

	// now stream the notification to the user
	mastoNotif, err := p.tc.NotificationToMasto(notif)
	if err != nil {
		return fmt.Errorf("notifyStatus: error converting notification to masto representation: %s", err)
	}

	if err := p.streamingProcessor.StreamNotificationToAccount(mastoNotif, boostedAcct); err != nil {
		return fmt.Errorf("notifyStatus: error streaming notification to account: %s", err)
	}

	return nil
}

func (p *processor) timelineStatus(status *gtsmodel.Status) error {
	// make sure the author account is pinned onto the status
	if status.Account == nil {
		a := &gtsmodel.Account{}
		if err := p.db.GetByID(status.AccountID, a); err != nil {
			return fmt.Errorf("timelineStatus: error getting author account with id %s: %s", status.AccountID, err)
		}
		status.Account = a
	}

	// get local followers of the account that posted the status
	followers := []gtsmodel.Follow{}
	if err := p.db.GetAccountFollowers(status.AccountID, &followers, true); err != nil {
		return fmt.Errorf("timelineStatus: error getting followers for account id %s: %s", status.AccountID, err)
	}

	// if the poster is local, add a fake entry for them to the followers list so they can see their own status in their timeline
	if status.Account.Domain == "" {
		followers = append(followers, gtsmodel.Follow{
			AccountID: status.AccountID,
		})
	}

	wg := sync.WaitGroup{}
	wg.Add(len(followers))
	errors := make(chan error, len(followers))

	for _, f := range followers {
		go p.timelineStatusForAccount(status, f.AccountID, errors, &wg)
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

func (p *processor) timelineStatusForAccount(status *gtsmodel.Status, accountID string, errors chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	// get the timeline owner account
	timelineAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(accountID, timelineAccount); err != nil {
		errors <- fmt.Errorf("timelineStatusForAccount: error getting account for timeline with id %s: %s", accountID, err)
		return
	}

	// make sure the status is timelineable
	timelineable, err := p.filter.StatusHometimelineable(status, timelineAccount)
	if err != nil {
		errors <- fmt.Errorf("timelineStatusForAccount: error getting timelineability for status for timeline with id %s: %s", accountID, err)
		return
	}

	if !timelineable {
		return
	}

	// stick the status in the timeline for the account and then immediately prepare it so they can see it right away
	inserted, err := p.timelineManager.IngestAndPrepare(status, timelineAccount.ID)
	if err != nil {
		errors <- fmt.Errorf("timelineStatusForAccount: error ingesting status %s: %s", status.ID, err)
		return
	}

	// the status was inserted to stream it to the user
	if inserted {
		mastoStatus, err := p.tc.StatusToMasto(status, timelineAccount)
		if err != nil {
			errors <- fmt.Errorf("timelineStatusForAccount: error converting status %s to frontend representation: %s", status.ID, err)
		} else {
			if err := p.streamingProcessor.StreamStatusToAccount(mastoStatus, timelineAccount); err != nil {
				errors <- fmt.Errorf("timelineStatusForAccount: error streaming status %s: %s", status.ID, err)
			}
		}
	}

	mastoStatus, err := p.tc.StatusToMasto(status, timelineAccount)
	if err != nil {
		errors <- fmt.Errorf("timelineStatusForAccount: error converting status %s to frontend representation: %s", status.ID, err)
	} else {
		if err := p.streamingProcessor.StreamStatusToAccount(mastoStatus, timelineAccount); err != nil {
			errors <- fmt.Errorf("timelineStatusForAccount: error streaming status %s: %s", status.ID, err)
		}
	}
}

func (p *processor) deleteStatusFromTimelines(status *gtsmodel.Status) error {
	if err := p.timelineManager.WipeStatusFromAllTimelines(status.ID); err != nil {
		return err
	}

	return p.streamingProcessor.StreamDelete(status.ID)
}
