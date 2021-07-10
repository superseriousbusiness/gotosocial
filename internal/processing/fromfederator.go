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
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func (p *processor) processFromFederator(federatorMsg gtsmodel.FromFederator) error {
	l := p.log.WithFields(logrus.Fields{
		"func":         "processFromFederator",
		"federatorMsg": fmt.Sprintf("%+v", federatorMsg),
	})

	l.Trace("entering function PROCESS FROM FEDERATOR")

	switch federatorMsg.APActivityType {
	case gtsmodel.ActivityStreamsCreate:
		// CREATE
		switch federatorMsg.APObjectType {
		case gtsmodel.ActivityStreamsNote:
			// CREATE A STATUS
			incomingStatus, ok := federatorMsg.GTSModel.(*gtsmodel.Status)
			if !ok {
				return errors.New("note was not parseable as *gtsmodel.Status")
			}

			l.Trace("will now derefence incoming status")
			if err := p.federator.DereferenceStatusFields(incomingStatus, federatorMsg.ReceivingAccount.Username); err != nil {
				return fmt.Errorf("error dereferencing status from federator: %s", err)
			}
			if err := p.db.UpdateByID(incomingStatus.ID, incomingStatus); err != nil {
				return fmt.Errorf("error updating dereferenced status in the db: %s", err)
			}

			if err := p.timelineStatus(incomingStatus); err != nil {
				return err
			}

			if err := p.notifyStatus(incomingStatus); err != nil {
				return err
			}

		case gtsmodel.ActivityStreamsProfile:
			// CREATE AN ACCOUNT
			incomingAccount, ok := federatorMsg.GTSModel.(*gtsmodel.Account)
			if !ok {
				return errors.New("profile was not parseable as *gtsmodel.Account")
			}

			l.Trace("will now derefence incoming account")
			if err := p.federator.DereferenceAccountFields(incomingAccount, "", false); err != nil {
				return fmt.Errorf("error dereferencing account from federator: %s", err)
			}
			if err := p.db.UpdateByID(incomingAccount.ID, incomingAccount); err != nil {
				return fmt.Errorf("error updating dereferenced account in the db: %s", err)
			}
		case gtsmodel.ActivityStreamsLike:
			// CREATE A FAVE
			incomingFave, ok := federatorMsg.GTSModel.(*gtsmodel.StatusFave)
			if !ok {
				return errors.New("like was not parseable as *gtsmodel.StatusFave")
			}

			if err := p.notifyFave(incomingFave, federatorMsg.ReceivingAccount); err != nil {
				return err
			}
		case gtsmodel.ActivityStreamsFollow:
			// CREATE A FOLLOW REQUEST
			incomingFollowRequest, ok := federatorMsg.GTSModel.(*gtsmodel.FollowRequest)
			if !ok {
				return errors.New("incomingFollowRequest was not parseable as *gtsmodel.FollowRequest")
			}

			if err := p.notifyFollowRequest(incomingFollowRequest, federatorMsg.ReceivingAccount); err != nil {
				return err
			}
		case gtsmodel.ActivityStreamsAnnounce:
			// CREATE AN ANNOUNCE
			incomingAnnounce, ok := federatorMsg.GTSModel.(*gtsmodel.Status)
			if !ok {
				return errors.New("announce was not parseable as *gtsmodel.Status")
			}

			if err := p.federator.DereferenceAnnounce(incomingAnnounce, federatorMsg.ReceivingAccount.Username); err != nil {
				return fmt.Errorf("error dereferencing announce from federator: %s", err)
			}

			incomingAnnounceID, err := id.NewULIDFromTime(incomingAnnounce.CreatedAt)
			if err != nil {
				return err
			}
			incomingAnnounce.ID = incomingAnnounceID

			if err := p.db.Put(incomingAnnounce); err != nil {
				if _, ok := err.(db.ErrAlreadyExists); !ok {
					return fmt.Errorf("error adding dereferenced announce to the db: %s", err)
				}
			}

			if err := p.timelineStatus(incomingAnnounce); err != nil {
				return err
			}

			if err := p.notifyAnnounce(incomingAnnounce); err != nil {
				return err
			}
		case gtsmodel.ActivityStreamsBlock:
			// CREATE A BLOCK

			// TODO: remove any of the blocking account's statuses from the blocked account's timeline
			// TODO: same with notifications
			// TODO: same with bookmarks
		}
	case gtsmodel.ActivityStreamsUpdate:
		// UPDATE
		switch federatorMsg.APObjectType {
		case gtsmodel.ActivityStreamsProfile:
			// UPDATE AN ACCOUNT
			incomingAccount, ok := federatorMsg.GTSModel.(*gtsmodel.Account)
			if !ok {
				return errors.New("profile was not parseable as *gtsmodel.Account")
			}

			l.Trace("will now derefence incoming account")
			if err := p.federator.DereferenceAccountFields(incomingAccount, federatorMsg.ReceivingAccount.Username, true); err != nil {
				return fmt.Errorf("error dereferencing account from federator: %s", err)
			}
			if err := p.db.UpdateByID(incomingAccount.ID, incomingAccount); err != nil {
				return fmt.Errorf("error updating dereferenced account in the db: %s", err)
			}
		}
	case gtsmodel.ActivityStreamsDelete:
		// DELETE
		switch federatorMsg.APObjectType {
		case gtsmodel.ActivityStreamsNote:
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
			for _, a := range statusToDelete.Attachments {
				if err := p.mediaProcessor.Delete(a); err != nil {
					return err
				}
			}

			// delete all mentions for this status
			for _, m := range statusToDelete.Mentions {
				if err := p.db.DeleteByID(m, &gtsmodel.Mention{}); err != nil {
					return err
				}
			}

			// delete all notifications for this status
			if err := p.db.DeleteWhere([]db.Where{{Key: "status_id", Value: statusToDelete.ID}}, &[]*gtsmodel.Notification{}); err != nil {
				return err
			}

			// remove this status from any and all timelines
			return p.deleteStatusFromTimelines(statusToDelete)
		case gtsmodel.ActivityStreamsProfile:
			// DELETE A PROFILE/ACCOUNT
			// TODO: handle side effects of account deletion here: delete all objects, statuses, media etc associated with account
		}
	case gtsmodel.ActivityStreamsAccept:
		// ACCEPT
		switch federatorMsg.APObjectType {
		case gtsmodel.ActivityStreamsFollow:
			// ACCEPT A FOLLOW
			follow, ok := federatorMsg.GTSModel.(*gtsmodel.Follow)
			if !ok {
				return errors.New("follow was not parseable as *gtsmodel.Follow")
			}

			if err := p.notifyFollow(follow, federatorMsg.ReceivingAccount); err != nil {
				return err
			}
		}
	}

	return nil
}
