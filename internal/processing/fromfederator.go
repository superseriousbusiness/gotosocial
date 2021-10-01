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
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (p *processor) ProcessFromFederator(ctx context.Context, federatorMsg messages.FromFederator) error {
	l := p.log.WithFields(logrus.Fields{
		"func":         "processFromFederator",
		"federatorMsg": fmt.Sprintf("%+v", federatorMsg),
	})

	l.Trace("entering function PROCESS FROM FEDERATOR")

	switch federatorMsg.APActivityType {
	case ap.ActivityCreate:
		// CREATE
		switch federatorMsg.APObjectType {
		case ap.ObjectNote:
			// CREATE A STATUS
			incomingStatus, ok := federatorMsg.GTSModel.(*gtsmodel.Status)
			if !ok {
				return errors.New("note was not parseable as *gtsmodel.Status")
			}

			status, err := p.federator.EnrichRemoteStatus(ctx, federatorMsg.ReceivingAccount.Username, incomingStatus, true)
			if err != nil {
				return err
			}

			if err := p.timelineStatus(ctx, status); err != nil {
				return err
			}

			if err := p.notifyStatus(ctx, status); err != nil {
				return err
			}
		case ap.ObjectProfile:
			// CREATE AN ACCOUNT
			// nothing to do here
		case ap.ActivityLike:
			// CREATE A FAVE
			incomingFave, ok := federatorMsg.GTSModel.(*gtsmodel.StatusFave)
			if !ok {
				return errors.New("like was not parseable as *gtsmodel.StatusFave")
			}

			if err := p.notifyFave(ctx, incomingFave); err != nil {
				return err
			}
		case ap.ActivityFollow:
			// CREATE A FOLLOW REQUEST
			followRequest, ok := federatorMsg.GTSModel.(*gtsmodel.FollowRequest)
			if !ok {
				return errors.New("incomingFollowRequest was not parseable as *gtsmodel.FollowRequest")
			}

			if followRequest.TargetAccount == nil {
				a, err := p.db.GetAccountByID(ctx, followRequest.TargetAccountID)
				if err != nil {
					return err
				}
				followRequest.TargetAccount = a
			}
			targetAccount := followRequest.TargetAccount

			if targetAccount.Locked {
				// if the account is locked just notify the follow request and nothing else
				return p.notifyFollowRequest(ctx, followRequest)
			}

			if followRequest.Account == nil {
				a, err := p.db.GetAccountByID(ctx, followRequest.AccountID)
				if err != nil {
					return err
				}
				followRequest.Account = a
			}
			originAccount := followRequest.Account

			// if the target account isn't locked, we should already accept the follow and notify about the new follower instead
			follow, err := p.db.AcceptFollowRequest(ctx, followRequest.AccountID, followRequest.TargetAccountID)
			if err != nil {
				return err
			}

			if err := p.federateAcceptFollowRequest(ctx, follow, originAccount, targetAccount); err != nil {
				return err
			}

			return p.notifyFollow(ctx, follow, targetAccount)
		case ap.ActivityAnnounce:
			// CREATE AN ANNOUNCE
			incomingAnnounce, ok := federatorMsg.GTSModel.(*gtsmodel.Status)
			if !ok {
				return errors.New("announce was not parseable as *gtsmodel.Status")
			}

			if err := p.federator.DereferenceAnnounce(ctx, incomingAnnounce, federatorMsg.ReceivingAccount.Username); err != nil {
				return fmt.Errorf("error dereferencing announce from federator: %s", err)
			}

			incomingAnnounceID, err := id.NewULIDFromTime(incomingAnnounce.CreatedAt)
			if err != nil {
				return err
			}
			incomingAnnounce.ID = incomingAnnounceID

			if err := p.db.PutStatus(ctx, incomingAnnounce); err != nil {
				return fmt.Errorf("error adding dereferenced announce to the db: %s", err)
			}

			if err := p.timelineStatus(ctx, incomingAnnounce); err != nil {
				return err
			}

			if err := p.notifyAnnounce(ctx, incomingAnnounce); err != nil {
				return err
			}
		case ap.ActivityBlock:
			// CREATE A BLOCK
			block, ok := federatorMsg.GTSModel.(*gtsmodel.Block)
			if !ok {
				return errors.New("block was not parseable as *gtsmodel.Block")
			}

			// remove any of the blocking account's statuses from the blocked account's timeline, and vice versa
			if err := p.timelineManager.WipeStatusesFromAccountID(ctx, block.AccountID, block.TargetAccountID); err != nil {
				return err
			}
			if err := p.timelineManager.WipeStatusesFromAccountID(ctx, block.TargetAccountID, block.AccountID); err != nil {
				return err
			}
			// TODO: same with notifications
			// TODO: same with bookmarks
		}
	case ap.ActivityUpdate:
		// UPDATE
		switch federatorMsg.APObjectType {
		case ap.ObjectProfile:
			// UPDATE AN ACCOUNT
			incomingAccount, ok := federatorMsg.GTSModel.(*gtsmodel.Account)
			if !ok {
				return errors.New("profile was not parseable as *gtsmodel.Account")
			}

			if _, err := p.federator.EnrichRemoteAccount(ctx, federatorMsg.ReceivingAccount.Username, incomingAccount); err != nil {
				return fmt.Errorf("error enriching updated account from federator: %s", err)
			}
		}
	case ap.ActivityDelete:
		// DELETE
		switch federatorMsg.APObjectType {
		case ap.ObjectNote:
			// DELETE A STATUS
			// TODO: handle side effects of status deletion here:
			// 1. delete all media associated with status
			// 2. delete boosts of status
			// 3. etc etc etc
			statusToDelete, ok := federatorMsg.GTSModel.(*gtsmodel.Status)
			if !ok {
				return errors.New("note was not parseable as *gtsmodel.Status")
			}

			// delete all attachments for this status
			for _, a := range statusToDelete.AttachmentIDs {
				if err := p.mediaProcessor.Delete(ctx, a); err != nil {
					return err
				}
			}

			// delete all mentions for this status
			for _, m := range statusToDelete.MentionIDs {
				if err := p.db.DeleteByID(ctx, m, &gtsmodel.Mention{}); err != nil {
					return err
				}
			}

			// delete all notifications for this status
			if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "status_id", Value: statusToDelete.ID}}, &[]*gtsmodel.Notification{}); err != nil {
				return err
			}

			// remove this status from any and all timelines
			return p.deleteStatusFromTimelines(ctx, statusToDelete)
		case ap.ObjectProfile:
			// DELETE A PROFILE/ACCOUNT
			// handle side effects of account deletion here: delete all objects, statuses, media etc associated with account
			account, ok := federatorMsg.GTSModel.(*gtsmodel.Account)
			if !ok {
				return errors.New("account delete was not parseable as *gtsmodel.Account")
			}

			return p.accountProcessor.Delete(ctx, account, account.ID)
		}
	case ap.ActivityAccept:
		// ACCEPT
		switch federatorMsg.APObjectType {
		case ap.ActivityFollow:
			// ACCEPT A FOLLOW
			// nothing to do here
		}
	}

	return nil
}
