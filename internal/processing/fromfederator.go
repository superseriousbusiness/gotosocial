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
	"errors"
	"fmt"
	"net/url"

	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// ProcessFromFederator reads the APActivityType and APObjectType of an incoming message from the federator,
// and directs the message into the appropriate side effect handler function, or simply does nothing if there's
// no handler function defined for the combination of Activity and Object.
func (p *processor) ProcessFromFederator(ctx context.Context, federatorMsg messages.FromFederator) error {
	// Allocate new log fields slice
	fields := make([]kv.Field, 3, 5)
	fields[0] = kv.Field{"activityType", federatorMsg.APActivityType}
	fields[1] = kv.Field{"objectType", federatorMsg.APObjectType}
	fields[2] = kv.Field{"toAccount", federatorMsg.ReceivingAccount.Username}

	if federatorMsg.APIri != nil {
		// An IRI was supplied, append to log
		fields = append(fields, kv.Field{
			"iri", federatorMsg.APIri,
		})
	}

	if federatorMsg.GTSModel != nil &&
		log.Level() >= level.DEBUG {
		// Append converted model to log
		fields = append(fields, kv.Field{
			"model", federatorMsg.GTSModel,
		})
	}

	// Log this federated message
	l := log.WithContext(ctx).WithFields(fields...)
	l.Info("processing from federator")

	switch federatorMsg.APActivityType {
	case ap.ActivityCreate:
		// CREATE SOMETHING
		switch federatorMsg.APObjectType {
		case ap.ObjectNote:
			// CREATE A STATUS
			return p.processCreateStatusFromFederator(ctx, federatorMsg)
		case ap.ActivityLike:
			// CREATE A FAVE
			return p.processCreateFaveFromFederator(ctx, federatorMsg)
		case ap.ActivityFollow:
			// CREATE A FOLLOW REQUEST
			return p.processCreateFollowRequestFromFederator(ctx, federatorMsg)
		case ap.ActivityAnnounce:
			// CREATE AN ANNOUNCE
			return p.processCreateAnnounceFromFederator(ctx, federatorMsg)
		case ap.ActivityBlock:
			// CREATE A BLOCK
			return p.processCreateBlockFromFederator(ctx, federatorMsg)
		case ap.ActivityFlag:
			// CREATE A FLAG / REPORT
			return p.processCreateFlagFromFederator(ctx, federatorMsg)
		}
	case ap.ActivityUpdate:
		// UPDATE SOMETHING
		if federatorMsg.APObjectType == ap.ObjectProfile {
			// UPDATE AN ACCOUNT
			return p.processUpdateAccountFromFederator(ctx, federatorMsg)
		}
	case ap.ActivityDelete:
		// DELETE SOMETHING
		switch federatorMsg.APObjectType {
		case ap.ObjectNote:
			// DELETE A STATUS
			return p.processDeleteStatusFromFederator(ctx, federatorMsg)
		case ap.ObjectProfile:
			// DELETE A PROFILE/ACCOUNT
			return p.processDeleteAccountFromFederator(ctx, federatorMsg)
		}
	}

	// not a combination we can/need to process
	return nil
}

// processCreateStatusFromFederator handles Activity Create and Object Note
func (p *processor) processCreateStatusFromFederator(ctx context.Context, federatorMsg messages.FromFederator) error {
	// check for either an IRI that we still need to dereference, OR an already dereferenced
	// and converted status pinned to the message.
	var status *gtsmodel.Status

	if federatorMsg.GTSModel != nil {
		// there's a gts model already pinned to the message, it should be a status
		var ok bool
		if status, ok = federatorMsg.GTSModel.(*gtsmodel.Status); !ok {
			return errors.New("ProcessFromFederator: note was not parseable as *gtsmodel.Status")
		}

		var err error
		status, err = p.federator.EnrichRemoteStatus(ctx, federatorMsg.ReceivingAccount.Username, status, true)
		if err != nil {
			return err
		}
	} else {
		// no model pinned, we need to dereference based on the IRI
		if federatorMsg.APIri == nil {
			return errors.New("ProcessFromFederator: status was not pinned to federatorMsg, and neither was an IRI for us to dereference")
		}
		var err error
		status, _, err = p.federator.GetStatus(ctx, federatorMsg.ReceivingAccount.Username, federatorMsg.APIri, false, false)
		if err != nil {
			return err
		}
	}

	// make sure the account is pinned
	if status.Account == nil {
		a, err := p.db.GetAccountByID(ctx, status.AccountID)
		if err != nil {
			return err
		}
		status.Account = a
	}

	// do a BLOCKING get of the remote account to make sure the avi and header are cached
	if status.Account.Domain != "" {
		remoteAccountID, err := url.Parse(status.Account.URI)
		if err != nil {
			return err
		}

		a, err := p.federator.GetAccountByURI(ctx,
			federatorMsg.ReceivingAccount.Username,
			remoteAccountID,
			true,
		)
		if err != nil {
			return err
		}

		status.Account = a
	}

	if err := p.timelineStatus(ctx, status); err != nil {
		return err
	}

	if err := p.notifyStatus(ctx, status); err != nil {
		return err
	}

	return nil
}

// processCreateFaveFromFederator handles Activity Create and Object Like
func (p *processor) processCreateFaveFromFederator(ctx context.Context, federatorMsg messages.FromFederator) error {
	incomingFave, ok := federatorMsg.GTSModel.(*gtsmodel.StatusFave)
	if !ok {
		return errors.New("like was not parseable as *gtsmodel.StatusFave")
	}

	// make sure the account is pinned
	if incomingFave.Account == nil {
		a, err := p.db.GetAccountByID(ctx, incomingFave.AccountID)
		if err != nil {
			return err
		}
		incomingFave.Account = a
	}

	// do a BLOCKING get of the remote account to make sure the avi and header are cached
	if incomingFave.Account.Domain != "" {
		remoteAccountID, err := url.Parse(incomingFave.Account.URI)
		if err != nil {
			return err
		}

		a, err := p.federator.GetAccountByURI(ctx,
			federatorMsg.ReceivingAccount.Username,
			remoteAccountID,
			true,
		)
		if err != nil {
			return err
		}

		incomingFave.Account = a
	}

	if err := p.notifyFave(ctx, incomingFave); err != nil {
		return err
	}

	return nil
}

// processCreateFollowRequestFromFederator handles Activity Create and Object Follow
func (p *processor) processCreateFollowRequestFromFederator(ctx context.Context, federatorMsg messages.FromFederator) error {
	followRequest, ok := federatorMsg.GTSModel.(*gtsmodel.FollowRequest)
	if !ok {
		return errors.New("incomingFollowRequest was not parseable as *gtsmodel.FollowRequest")
	}

	// make sure the account is pinned
	if followRequest.Account == nil {
		a, err := p.db.GetAccountByID(ctx, followRequest.AccountID)
		if err != nil {
			return err
		}
		followRequest.Account = a
	}

	// do a BLOCKING get of the remote account to make sure the avi and header are cached
	if followRequest.Account.Domain != "" {
		remoteAccountID, err := url.Parse(followRequest.Account.URI)
		if err != nil {
			return err
		}

		a, err := p.federator.GetAccountByURI(ctx,
			federatorMsg.ReceivingAccount.Username,
			remoteAccountID,
			true,
		)
		if err != nil {
			return err
		}

		followRequest.Account = a
	}

	if followRequest.TargetAccount == nil {
		a, err := p.db.GetAccountByID(ctx, followRequest.TargetAccountID)
		if err != nil {
			return err
		}
		followRequest.TargetAccount = a
	}

	if *followRequest.TargetAccount.Locked {
		// if the account is locked just notify the follow request and nothing else
		return p.notifyFollowRequest(ctx, followRequest)
	}

	// if the target account isn't locked, we should already accept the follow and notify about the new follower instead
	follow, err := p.db.AcceptFollowRequest(ctx, followRequest.AccountID, followRequest.TargetAccountID)
	if err != nil {
		return err
	}

	if err := p.federateAcceptFollowRequest(ctx, follow); err != nil {
		return err
	}

	return p.notifyFollow(ctx, follow, followRequest.TargetAccount)
}

// processCreateAnnounceFromFederator handles Activity Create and Object Announce
func (p *processor) processCreateAnnounceFromFederator(ctx context.Context, federatorMsg messages.FromFederator) error {
	incomingAnnounce, ok := federatorMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return errors.New("announce was not parseable as *gtsmodel.Status")
	}

	// make sure the account is pinned
	if incomingAnnounce.Account == nil {
		a, err := p.db.GetAccountByID(ctx, incomingAnnounce.AccountID)
		if err != nil {
			return err
		}
		incomingAnnounce.Account = a
	}

	// do a BLOCKING get of the remote account to make sure the avi and header are cached
	if incomingAnnounce.Account.Domain != "" {
		remoteAccountID, err := url.Parse(incomingAnnounce.Account.URI)
		if err != nil {
			return err
		}

		a, err := p.federator.GetAccountByURI(ctx,
			federatorMsg.ReceivingAccount.Username,
			remoteAccountID,
			true,
		)
		if err != nil {
			return err
		}

		incomingAnnounce.Account = a
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

	return nil
}

// processCreateBlockFromFederator handles Activity Create and Object Block
func (p *processor) processCreateBlockFromFederator(ctx context.Context, federatorMsg messages.FromFederator) error {
	block, ok := federatorMsg.GTSModel.(*gtsmodel.Block)
	if !ok {
		return errors.New("block was not parseable as *gtsmodel.Block")
	}

	// remove any of the blocking account's statuses from the blocked account's timeline, and vice versa
	if err := p.statusTimelines.WipeItemsFromAccountID(ctx, block.AccountID, block.TargetAccountID); err != nil {
		return err
	}
	if err := p.statusTimelines.WipeItemsFromAccountID(ctx, block.TargetAccountID, block.AccountID); err != nil {
		return err
	}
	// TODO: same with notifications
	// TODO: same with bookmarks

	return nil
}

func (p *processor) processCreateFlagFromFederator(ctx context.Context, federatorMsg messages.FromFederator) error {
	// TODO: handle side effects of flag creation:
	// - send email to admins
	// - notify admins
	return nil
}

// processUpdateAccountFromFederator handles Activity Update and Object Profile
func (p *processor) processUpdateAccountFromFederator(ctx context.Context, federatorMsg messages.FromFederator) error {
	incomingAccount, ok := federatorMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return errors.New("profile was not parseable as *gtsmodel.Account")
	}

	incomingAccountURL, err := url.Parse(incomingAccount.URI)
	if err != nil {
		return err
	}

	// further database updates occur inside getremoteaccount
	if _, err := p.federator.GetAccountByURI(ctx,
		federatorMsg.ReceivingAccount.Username,
		incomingAccountURL,
		true,
	); err != nil {
		return fmt.Errorf("error enriching updated account from federator: %s", err)
	}

	return nil
}

// processDeleteStatusFromFederator handles Activity Delete and Object Note
func (p *processor) processDeleteStatusFromFederator(ctx context.Context, federatorMsg messages.FromFederator) error {
	statusToDelete, ok := federatorMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return errors.New("note was not parseable as *gtsmodel.Status")
	}

	// delete attachments from this status since this request
	// comes from the federating API, and there's no way the
	// poster can do a delete + redraft for it on our instance
	deleteAttachments := true
	return p.wipeStatus(ctx, statusToDelete, deleteAttachments)
}

// processDeleteAccountFromFederator handles Activity Delete and Object Profile
func (p *processor) processDeleteAccountFromFederator(ctx context.Context, federatorMsg messages.FromFederator) error {
	account, ok := federatorMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return errors.New("account delete was not parseable as *gtsmodel.Account")
	}

	return p.accountProcessor.Delete(ctx, account, account.ID)
}
