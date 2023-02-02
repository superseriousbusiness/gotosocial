/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
)

func (p *processor) notifyStatus(ctx context.Context, status *gtsmodel.Status) error {
	// if there are no mentions in this status then just bail
	if len(status.MentionIDs) == 0 {
		return nil
	}

	if status.Mentions == nil {
		// there are mentions but they're not fully populated on the status yet so do this
		menchies, err := p.db.GetMentions(ctx, status.MentionIDs)
		if err != nil {
			return fmt.Errorf("notifyStatus: error getting mentions for status %s from the db: %s", status.ID, err)
		}
		status.Mentions = menchies
	}

	// now we have mentions as full gtsmodel.Mention structs on the status we can continue
	for _, m := range status.Mentions {
		// make sure this is a local account, otherwise we don't need to create a notification for it
		if m.TargetAccount == nil {
			a, err := p.db.GetAccountByID(ctx, m.TargetAccountID)
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
		if err := p.db.GetWhere(ctx, []db.Where{
			{Key: "notification_type", Value: gtsmodel.NotificationMention},
			{Key: "target_account_id", Value: m.TargetAccountID},
			{Key: "origin_account_id", Value: m.OriginAccountID},
			{Key: "status_id", Value: m.StatusID},
		}, &gtsmodel.Notification{}); err == nil {
			// notification exists already so just continue
			continue
		} else if err != db.ErrNoEntries {
			// there's a real error in the db
			return fmt.Errorf("notifyStatus: error checking existence of notification for mention with id %s : %s", m.ID, err)
		}

		// if we've reached this point we know the mention is for a local account, and the notification doesn't exist, so create it
		notif := &gtsmodel.Notification{
			ID:               id.NewULID(),
			NotificationType: gtsmodel.NotificationMention,
			TargetAccountID:  m.TargetAccountID,
			TargetAccount:    m.TargetAccount,
			OriginAccountID:  status.AccountID,
			OriginAccount:    status.Account,
			StatusID:         status.ID,
			Status:           status,
		}

		if err := p.db.Put(ctx, notif); err != nil {
			return fmt.Errorf("notifyStatus: error putting notification in database: %s", err)
		}

		// now stream the notification to the user
		apiNotif, err := p.tc.NotificationToAPINotification(ctx, notif)
		if err != nil {
			return fmt.Errorf("notifyStatus: error converting notification to api representation: %s", err)
		}

		if err := p.streamingProcessor.StreamNotificationToAccount(apiNotif, m.TargetAccount); err != nil {
			return fmt.Errorf("notifyStatus: error streaming notification to account: %s", err)
		}
	}

	return nil
}

func (p *processor) notifyFollowRequest(ctx context.Context, followRequest *gtsmodel.FollowRequest) error {
	// make sure we have the target account pinned on the follow request
	if followRequest.TargetAccount == nil {
		a, err := p.db.GetAccountByID(ctx, followRequest.TargetAccountID)
		if err != nil {
			return err
		}
		followRequest.TargetAccount = a
	}
	targetAccount := followRequest.TargetAccount

	// return if this isn't a local account
	if targetAccount.Domain != "" {
		// this isn't a local account so we've got nothing to do here
		return nil
	}

	notif := &gtsmodel.Notification{
		ID:               id.NewULID(),
		NotificationType: gtsmodel.NotificationFollowRequest,
		TargetAccountID:  followRequest.TargetAccountID,
		OriginAccountID:  followRequest.AccountID,
	}

	if err := p.db.Put(ctx, notif); err != nil {
		return fmt.Errorf("notifyFollowRequest: error putting notification in database: %s", err)
	}

	// now stream the notification to the user
	apiNotif, err := p.tc.NotificationToAPINotification(ctx, notif)
	if err != nil {
		return fmt.Errorf("notifyStatus: error converting notification to api representation: %s", err)
	}

	if err := p.streamingProcessor.StreamNotificationToAccount(apiNotif, targetAccount); err != nil {
		return fmt.Errorf("notifyStatus: error streaming notification to account: %s", err)
	}

	return nil
}

func (p *processor) notifyFollow(ctx context.Context, follow *gtsmodel.Follow, targetAccount *gtsmodel.Account) error {
	// return if this isn't a local account
	if targetAccount.Domain != "" {
		return nil
	}

	// first remove the follow request notification
	if err := p.db.DeleteWhere(ctx, []db.Where{
		{Key: "notification_type", Value: gtsmodel.NotificationFollowRequest},
		{Key: "target_account_id", Value: follow.TargetAccountID},
		{Key: "origin_account_id", Value: follow.AccountID},
	}, &gtsmodel.Notification{}); err != nil {
		return fmt.Errorf("notifyFollow: error removing old follow request notification from database: %s", err)
	}

	// now create the new follow notification
	notif := &gtsmodel.Notification{
		ID:               id.NewULID(),
		NotificationType: gtsmodel.NotificationFollow,
		TargetAccountID:  follow.TargetAccountID,
		TargetAccount:    follow.TargetAccount,
		OriginAccountID:  follow.AccountID,
		OriginAccount:    follow.Account,
	}
	if err := p.db.Put(ctx, notif); err != nil {
		return fmt.Errorf("notifyFollow: error putting notification in database: %s", err)
	}

	// now stream the notification to the user
	apiNotif, err := p.tc.NotificationToAPINotification(ctx, notif)
	if err != nil {
		return fmt.Errorf("notifyStatus: error converting notification to api representation: %s", err)
	}

	if err := p.streamingProcessor.StreamNotificationToAccount(apiNotif, targetAccount); err != nil {
		return fmt.Errorf("notifyStatus: error streaming notification to account: %s", err)
	}

	return nil
}

func (p *processor) notifyFave(ctx context.Context, fave *gtsmodel.StatusFave) error {
	// ignore self-faves
	if fave.TargetAccountID == fave.AccountID {
		return nil
	}

	if fave.TargetAccount == nil {
		a, err := p.db.GetAccountByID(ctx, fave.TargetAccountID)
		if err != nil {
			return err
		}
		fave.TargetAccount = a
	}
	targetAccount := fave.TargetAccount

	// just return if target isn't a local account
	if targetAccount.Domain != "" {
		return nil
	}

	notif := &gtsmodel.Notification{
		ID:               id.NewULID(),
		NotificationType: gtsmodel.NotificationFave,
		TargetAccountID:  fave.TargetAccountID,
		TargetAccount:    fave.TargetAccount,
		OriginAccountID:  fave.AccountID,
		OriginAccount:    fave.Account,
		StatusID:         fave.StatusID,
		Status:           fave.Status,
	}

	if err := p.db.Put(ctx, notif); err != nil {
		return fmt.Errorf("notifyFave: error putting notification in database: %s", err)
	}

	// now stream the notification to the user
	apiNotif, err := p.tc.NotificationToAPINotification(ctx, notif)
	if err != nil {
		return fmt.Errorf("notifyStatus: error converting notification to api representation: %s", err)
	}

	if err := p.streamingProcessor.StreamNotificationToAccount(apiNotif, targetAccount); err != nil {
		return fmt.Errorf("notifyStatus: error streaming notification to account: %s", err)
	}

	return nil
}

func (p *processor) notifyAnnounce(ctx context.Context, status *gtsmodel.Status) error {
	if status.BoostOfID == "" {
		// not a boost, nothing to do
		return nil
	}

	if status.BoostOf == nil {
		boostedStatus, err := p.db.GetStatusByID(ctx, status.BoostOfID)
		if err != nil {
			return fmt.Errorf("notifyAnnounce: error getting status with id %s: %s", status.BoostOfID, err)
		}
		status.BoostOf = boostedStatus
	}

	if status.BoostOfAccount == nil {
		boostedAcct, err := p.db.GetAccountByID(ctx, status.BoostOfAccountID)
		if err != nil {
			return fmt.Errorf("notifyAnnounce: error getting account with id %s: %s", status.BoostOfAccountID, err)
		}
		status.BoostOf.Account = boostedAcct
		status.BoostOfAccount = boostedAcct
	}

	if status.BoostOfAccount.Domain != "" {
		// remote account, nothing to do
		return nil
	}

	if status.BoostOfAccountID == status.AccountID {
		// it's a self boost, nothing to do
		return nil
	}

	// make sure a notif doesn't already exist for this announce
	err := p.db.GetWhere(ctx, []db.Where{
		{Key: "notification_type", Value: gtsmodel.NotificationReblog},
		{Key: "target_account_id", Value: status.BoostOfAccountID},
		{Key: "origin_account_id", Value: status.AccountID},
		{Key: "status_id", Value: status.ID},
	}, &gtsmodel.Notification{})
	if err == nil {
		// notification exists already so just bail
		return nil
	}

	// now create the new reblog notification
	notif := &gtsmodel.Notification{
		ID:               id.NewULID(),
		NotificationType: gtsmodel.NotificationReblog,
		TargetAccountID:  status.BoostOfAccountID,
		TargetAccount:    status.BoostOfAccount,
		OriginAccountID:  status.AccountID,
		OriginAccount:    status.Account,
		StatusID:         status.ID,
		Status:           status,
	}

	if err := p.db.Put(ctx, notif); err != nil {
		return fmt.Errorf("notifyAnnounce: error putting notification in database: %s", err)
	}

	// now stream the notification to the user
	apiNotif, err := p.tc.NotificationToAPINotification(ctx, notif)
	if err != nil {
		return fmt.Errorf("notifyStatus: error converting notification to api representation: %s", err)
	}

	if err := p.streamingProcessor.StreamNotificationToAccount(apiNotif, status.BoostOfAccount); err != nil {
		return fmt.Errorf("notifyStatus: error streaming notification to account: %s", err)
	}

	return nil
}

// timelineStatus processes the given new status and inserts it into
// the HOME timelines of accounts that follow the status author.
func (p *processor) timelineStatus(ctx context.Context, status *gtsmodel.Status) error {
	// make sure the author account is pinned onto the status
	if status.Account == nil {
		a, err := p.db.GetAccountByID(ctx, status.AccountID)
		if err != nil {
			return fmt.Errorf("timelineStatus: error getting author account with id %s: %s", status.AccountID, err)
		}
		status.Account = a
	}

	// get local followers of the account that posted the status
	follows, err := p.db.GetAccountFollowedBy(ctx, status.AccountID, true)
	if err != nil {
		return fmt.Errorf("timelineStatus: error getting followers for account id %s: %s", status.AccountID, err)
	}

	// if the poster is local, add a fake entry for them to the followers list so they can see their own status in their timeline
	if status.Account.Domain == "" {
		follows = append(follows, &gtsmodel.Follow{
			AccountID: status.AccountID,
			Account:   status.Account,
		})
	}

	wg := sync.WaitGroup{}
	wg.Add(len(follows))
	errors := make(chan error, len(follows))

	for _, f := range follows {
		go p.timelineStatusForAccount(ctx, status, f.AccountID, errors, &wg)
	}

	// read any errors that come in from the async functions
	errs := []string{}
	go func(errs []string) {
		for range errors {
			if e := <-errors; e != nil {
				errs = append(errs, e.Error())
			}
		}
	}(errs)

	// wait til all functions have returned and then close the error channel
	wg.Wait()
	close(errors)

	if len(errs) != 0 {
		// we have at least one error
		return fmt.Errorf("timelineStatus: one or more errors timelining statuses: %s", strings.Join(errs, ";"))
	}

	return nil
}

// timelineStatusForAccount puts the given status in the HOME timeline
// of the account with given accountID, if it's hometimelineable.
//
// If the status was inserted into the home timeline of the given account,
// it will also be streamed via websockets to the user.
func (p *processor) timelineStatusForAccount(ctx context.Context, status *gtsmodel.Status, accountID string, errors chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	// get the timeline owner account
	timelineAccount, err := p.db.GetAccountByID(ctx, accountID)
	if err != nil {
		errors <- fmt.Errorf("timelineStatusForAccount: error getting account for timeline with id %s: %s", accountID, err)
		return
	}

	// make sure the status is timelineable
	timelineable, err := p.filter.StatusHometimelineable(ctx, status, timelineAccount)
	if err != nil {
		errors <- fmt.Errorf("timelineStatusForAccount: error getting timelineability for status for timeline with id %s: %s", accountID, err)
		return
	}

	if !timelineable {
		return
	}

	// stick the status in the timeline for the account and then immediately prepare it so they can see it right away
	inserted, err := p.statusTimelines.IngestAndPrepare(ctx, status, timelineAccount.ID)
	if err != nil {
		errors <- fmt.Errorf("timelineStatusForAccount: error ingesting status %s: %s", status.ID, err)
		return
	}

	// the status was inserted so stream it to the user
	if inserted {
		apiStatus, err := p.tc.StatusToAPIStatus(ctx, status, timelineAccount)
		if err != nil {
			errors <- fmt.Errorf("timelineStatusForAccount: error converting status %s to frontend representation: %s", status.ID, err)
			return
		}

		if err := p.streamingProcessor.StreamUpdateToAccount(apiStatus, timelineAccount, stream.TimelineHome); err != nil {
			errors <- fmt.Errorf("timelineStatusForAccount: error streaming status %s: %s", status.ID, err)
		}
	}
}

// deleteStatusFromTimelines completely removes the given status from all timelines.
// It will also stream deletion of the status to all open streams.
func (p *processor) deleteStatusFromTimelines(ctx context.Context, status *gtsmodel.Status) error {
	if err := p.statusTimelines.WipeItemFromAllTimelines(ctx, status.ID); err != nil {
		return err
	}

	return p.streamingProcessor.StreamDelete(status.ID)
}

// wipeStatus contains common logic used to totally delete a status
// + all its attachments, notifications, boosts, and timeline entries.
func (p *processor) wipeStatus(ctx context.Context, statusToDelete *gtsmodel.Status, deleteAttachments bool) error {
	// either delete all attachments for this status, or simply
	// unattach all attachments for this status, so they'll be
	// cleaned later by a separate process; reason to unattach rather
	// than delete is that the poster might want to reattach them
	// to another status immediately (in case of delete + redraft)
	if deleteAttachments {
		for _, a := range statusToDelete.AttachmentIDs {
			if err := p.mediaProcessor.Delete(ctx, a); err != nil {
				return err
			}
		}
	} else {
		for _, a := range statusToDelete.AttachmentIDs {
			if _, err := p.mediaProcessor.Unattach(ctx, statusToDelete.Account, a); err != nil {
				return err
			}
		}
	}

	// delete all mention entries generated by this status
	for _, m := range statusToDelete.MentionIDs {
		if err := p.db.DeleteByID(ctx, m, &gtsmodel.Mention{}); err != nil {
			return err
		}
	}

	// delete all notification entries generated by this status
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "status_id", Value: statusToDelete.ID}}, &[]*gtsmodel.Notification{}); err != nil {
		return err
	}

	// delete all boosts for this status + remove them from timelines
	if boosts, err := p.db.GetStatusReblogs(ctx, statusToDelete); err == nil {
		for _, b := range boosts {
			if err := p.deleteStatusFromTimelines(ctx, b); err != nil {
				return err
			}
			if err := p.db.DeleteStatusByID(ctx, b.ID); err != nil {
				return err
			}
		}
	}

	// delete this status from any and all timelines
	if err := p.deleteStatusFromTimelines(ctx, statusToDelete); err != nil {
		return err
	}

	// delete the status itself
	if err := p.db.DeleteStatusByID(ctx, statusToDelete.ID); err != nil {
		return err
	}

	return nil
}
