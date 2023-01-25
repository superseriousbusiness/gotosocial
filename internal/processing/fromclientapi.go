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
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (p *processor) ProcessFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	// Allocate new log fields slice
	fields := make([]kv.Field, 3, 4)
	fields[0] = kv.Field{"activityType", clientMsg.APActivityType}
	fields[1] = kv.Field{"objectType", clientMsg.APObjectType}
	fields[2] = kv.Field{"fromAccount", clientMsg.OriginAccount.Username}

	if clientMsg.GTSModel != nil &&
		log.Level() >= level.DEBUG {
		// Append converted model to log
		fields = append(fields, kv.Field{
			"model", clientMsg.GTSModel,
		})
	}

	// Log this federated message
	l := log.WithFields(fields...)
	l.Info("processing from client")

	switch clientMsg.APActivityType {
	case ap.ActivityCreate:
		// CREATE
		switch clientMsg.APObjectType {
		case ap.ObjectProfile, ap.ActorPerson:
			// CREATE ACCOUNT/PROFILE
			return p.processCreateAccountFromClientAPI(ctx, clientMsg)
		case ap.ObjectNote:
			// CREATE NOTE
			return p.processCreateStatusFromClientAPI(ctx, clientMsg)
		case ap.ActivityFollow:
			// CREATE FOLLOW REQUEST
			return p.processCreateFollowRequestFromClientAPI(ctx, clientMsg)
		case ap.ActivityLike:
			// CREATE LIKE/FAVE
			return p.processCreateFaveFromClientAPI(ctx, clientMsg)
		case ap.ActivityAnnounce:
			// CREATE BOOST/ANNOUNCE
			return p.processCreateAnnounceFromClientAPI(ctx, clientMsg)
		case ap.ActivityBlock:
			// CREATE BLOCK
			return p.processCreateBlockFromClientAPI(ctx, clientMsg)
		}
	case ap.ActivityUpdate:
		// UPDATE
		switch clientMsg.APObjectType {
		case ap.ObjectProfile, ap.ActorPerson:
			// UPDATE ACCOUNT/PROFILE
			return p.processUpdateAccountFromClientAPI(ctx, clientMsg)
		}
	case ap.ActivityAccept:
		// ACCEPT
		if clientMsg.APObjectType == ap.ActivityFollow {
			// ACCEPT FOLLOW
			return p.processAcceptFollowFromClientAPI(ctx, clientMsg)
		}
	case ap.ActivityReject:
		// REJECT
		if clientMsg.APObjectType == ap.ActivityFollow {
			// REJECT FOLLOW (request)
			return p.processRejectFollowFromClientAPI(ctx, clientMsg)
		}
	case ap.ActivityUndo:
		// UNDO
		switch clientMsg.APObjectType {
		case ap.ActivityFollow:
			// UNDO FOLLOW
			return p.processUndoFollowFromClientAPI(ctx, clientMsg)
		case ap.ActivityBlock:
			// UNDO BLOCK
			return p.processUndoBlockFromClientAPI(ctx, clientMsg)
		case ap.ActivityLike:
			// UNDO LIKE/FAVE
			return p.processUndoFaveFromClientAPI(ctx, clientMsg)
		case ap.ActivityAnnounce:
			// UNDO ANNOUNCE/BOOST
			return p.processUndoAnnounceFromClientAPI(ctx, clientMsg)
		}
	case ap.ActivityDelete:
		// DELETE
		switch clientMsg.APObjectType {
		case ap.ObjectNote:
			// DELETE STATUS/NOTE
			return p.processDeleteStatusFromClientAPI(ctx, clientMsg)
		case ap.ObjectProfile, ap.ActorPerson:
			// DELETE ACCOUNT/PROFILE
			return p.processDeleteAccountFromClientAPI(ctx, clientMsg)
		}
	case ap.ActivityFlag:
		// FLAG
		if clientMsg.APObjectType == ap.ObjectProfile {
			// FLAG/REPORT A PROFILE
			return p.processReportAccountFromClientAPI(ctx, clientMsg)
		}
	}
	return nil
}

func (p *processor) processCreateAccountFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	account, ok := clientMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return errors.New("account was not parseable as *gtsmodel.Account")
	}

	// return if the account isn't from this domain
	if account.Domain != "" {
		return nil
	}

	// get the user this account belongs to
	user, err := p.db.GetUserByAccountID(ctx, account.ID)
	if err != nil {
		return err
	}

	// email a confirmation to this user
	return p.userProcessor.SendConfirmEmail(ctx, user, account.Username)
}

func (p *processor) processCreateStatusFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	status, ok := clientMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return errors.New("note was not parseable as *gtsmodel.Status")
	}

	if err := p.timelineStatus(ctx, status); err != nil {
		return err
	}

	if err := p.notifyStatus(ctx, status); err != nil {
		return err
	}

	return p.federateStatus(ctx, status)
}

func (p *processor) processCreateFollowRequestFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	followRequest, ok := clientMsg.GTSModel.(*gtsmodel.FollowRequest)
	if !ok {
		return errors.New("followrequest was not parseable as *gtsmodel.FollowRequest")
	}

	if err := p.notifyFollowRequest(ctx, followRequest); err != nil {
		return err
	}

	return p.federateFollow(ctx, followRequest, clientMsg.OriginAccount, clientMsg.TargetAccount)
}

func (p *processor) processCreateFaveFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	fave, ok := clientMsg.GTSModel.(*gtsmodel.StatusFave)
	if !ok {
		return errors.New("fave was not parseable as *gtsmodel.StatusFave")
	}

	if err := p.notifyFave(ctx, fave); err != nil {
		return err
	}

	return p.federateFave(ctx, fave, clientMsg.OriginAccount, clientMsg.TargetAccount)
}

func (p *processor) processCreateAnnounceFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	boostWrapperStatus, ok := clientMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return errors.New("boost was not parseable as *gtsmodel.Status")
	}

	if err := p.timelineStatus(ctx, boostWrapperStatus); err != nil {
		return err
	}

	if err := p.notifyAnnounce(ctx, boostWrapperStatus); err != nil {
		return err
	}

	return p.federateAnnounce(ctx, boostWrapperStatus, clientMsg.OriginAccount, clientMsg.TargetAccount)
}

func (p *processor) processCreateBlockFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	block, ok := clientMsg.GTSModel.(*gtsmodel.Block)
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

	return p.federateBlock(ctx, block)
}

func (p *processor) processUpdateAccountFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	account, ok := clientMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return errors.New("account was not parseable as *gtsmodel.Account")
	}

	return p.federateAccountUpdate(ctx, account, clientMsg.OriginAccount)
}

func (p *processor) processAcceptFollowFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	follow, ok := clientMsg.GTSModel.(*gtsmodel.Follow)
	if !ok {
		return errors.New("accept was not parseable as *gtsmodel.Follow")
	}

	if err := p.notifyFollow(ctx, follow, clientMsg.TargetAccount); err != nil {
		return err
	}

	return p.federateAcceptFollowRequest(ctx, follow)
}

func (p *processor) processRejectFollowFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	followRequest, ok := clientMsg.GTSModel.(*gtsmodel.FollowRequest)
	if !ok {
		return errors.New("reject was not parseable as *gtsmodel.FollowRequest")
	}

	return p.federateRejectFollowRequest(ctx, followRequest)
}

func (p *processor) processUndoFollowFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	follow, ok := clientMsg.GTSModel.(*gtsmodel.Follow)
	if !ok {
		return errors.New("undo was not parseable as *gtsmodel.Follow")
	}
	return p.federateUnfollow(ctx, follow, clientMsg.OriginAccount, clientMsg.TargetAccount)
}

func (p *processor) processUndoBlockFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	block, ok := clientMsg.GTSModel.(*gtsmodel.Block)
	if !ok {
		return errors.New("undo was not parseable as *gtsmodel.Block")
	}
	return p.federateUnblock(ctx, block)
}

func (p *processor) processUndoFaveFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	fave, ok := clientMsg.GTSModel.(*gtsmodel.StatusFave)
	if !ok {
		return errors.New("undo was not parseable as *gtsmodel.StatusFave")
	}
	return p.federateUnfave(ctx, fave, clientMsg.OriginAccount, clientMsg.TargetAccount)
}

func (p *processor) processUndoAnnounceFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	boost, ok := clientMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return errors.New("undo was not parseable as *gtsmodel.Status")
	}

	if err := p.db.DeleteStatusByID(ctx, boost.ID); err != nil {
		return err
	}

	if err := p.deleteStatusFromTimelines(ctx, boost); err != nil {
		return err
	}

	return p.federateUnannounce(ctx, boost, clientMsg.OriginAccount, clientMsg.TargetAccount)
}

func (p *processor) processDeleteStatusFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	statusToDelete, ok := clientMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return errors.New("note was not parseable as *gtsmodel.Status")
	}

	if statusToDelete.Account == nil {
		statusToDelete.Account = clientMsg.OriginAccount
	}

	// don't delete attachments, just unattach them;
	// since this request comes from the client API
	// and the poster might want to use the attachments
	// again in a new post
	deleteAttachments := false
	if err := p.wipeStatus(ctx, statusToDelete, deleteAttachments); err != nil {
		return err
	}

	return p.federateStatusDelete(ctx, statusToDelete)
}

func (p *processor) processDeleteAccountFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	// the origin of the delete could be either a domain block, or an action by another (or this) account
	var origin string
	if domainBlock, ok := clientMsg.GTSModel.(*gtsmodel.DomainBlock); ok {
		// origin is a domain block
		origin = domainBlock.ID
	} else {
		// origin is whichever account caused this message
		origin = clientMsg.OriginAccount.ID
	}

	if err := p.federateAccountDelete(ctx, clientMsg.TargetAccount); err != nil {
		return err
	}

	return p.accountProcessor.Delete(ctx, clientMsg.TargetAccount, origin)
}

func (p *processor) processReportAccountFromClientAPI(ctx context.Context, clientMsg messages.FromClientAPI) error {
	report, ok := clientMsg.GTSModel.(*gtsmodel.Report)
	if !ok {
		return errors.New("report was not parseable as *gtsmodel.Report")
	}

	// TODO: in a separate PR, also email admin(s)

	if !*report.Forwarded {
		// nothing to do, don't federate the report
		return nil
	}

	return p.federateReport(ctx, report)
}

// TODO: move all the below functions into federation.Federator

func (p *processor) federateAccountDelete(ctx context.Context, account *gtsmodel.Account) error {
	// do nothing if this isn't our account
	if account.Domain != "" {
		return nil
	}

	outboxIRI, err := url.Parse(account.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateAccountDelete: error parsing outboxURI %s: %s", account.OutboxURI, err)
	}

	actorIRI, err := url.Parse(account.URI)
	if err != nil {
		return fmt.Errorf("federateAccountDelete: error parsing actorIRI %s: %s", account.URI, err)
	}

	followersIRI, err := url.Parse(account.FollowersURI)
	if err != nil {
		return fmt.Errorf("federateAccountDelete: error parsing followersIRI %s: %s", account.FollowersURI, err)
	}

	publicIRI, err := url.Parse(pub.PublicActivityPubIRI)
	if err != nil {
		return fmt.Errorf("federateAccountDelete: error parsing url %s: %s", pub.PublicActivityPubIRI, err)
	}

	// create a delete and set the appropriate actor on it
	delete := streams.NewActivityStreamsDelete()

	// set the actor for the delete; no matter who deleted it we should use the account owner for this
	deleteActor := streams.NewActivityStreamsActorProperty()
	deleteActor.AppendIRI(actorIRI)
	delete.SetActivityStreamsActor(deleteActor)

	// Set the account IRI as the 'object' property.
	deleteObject := streams.NewActivityStreamsObjectProperty()
	deleteObject.AppendIRI(actorIRI)
	delete.SetActivityStreamsObject(deleteObject)

	// send to followers...
	deleteTo := streams.NewActivityStreamsToProperty()
	deleteTo.AppendIRI(followersIRI)
	delete.SetActivityStreamsTo(deleteTo)

	// ... and CC to public
	deleteCC := streams.NewActivityStreamsCcProperty()
	deleteCC.AppendIRI(publicIRI)
	delete.SetActivityStreamsCc(deleteCC)

	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, delete)
	return err
}

func (p *processor) federateStatus(ctx context.Context, status *gtsmodel.Status) error {
	// do nothing if the status shouldn't be federated
	if !*status.Federated {
		return nil
	}

	if status.Account == nil {
		statusAccount, err := p.db.GetAccountByID(ctx, status.AccountID)
		if err != nil {
			return fmt.Errorf("federateStatus: error fetching status author account: %s", err)
		}
		status.Account = statusAccount
	}

	// do nothing if this isn't our status
	if status.Account.Domain != "" {
		return nil
	}

	asStatus, err := p.tc.StatusToAS(ctx, status)
	if err != nil {
		return fmt.Errorf("federateStatus: error converting status to as format: %s", err)
	}

	create, err := p.tc.WrapNoteInCreate(asStatus, false)
	if err != nil {
		return fmt.Errorf("federateStatus: error wrapping status in create: %s", err)
	}

	outboxIRI, err := url.Parse(status.Account.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateStatus: error parsing outboxURI %s: %s", status.Account.OutboxURI, err)
	}

	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, create)
	return err
}

func (p *processor) federateStatusDelete(ctx context.Context, status *gtsmodel.Status) error {
	if status.Account == nil {
		statusAccount, err := p.db.GetAccountByID(ctx, status.AccountID)
		if err != nil {
			return fmt.Errorf("federateStatusDelete: error fetching status author account: %s", err)
		}
		status.Account = statusAccount
	}

	// do nothing if this isn't our status
	if status.Account.Domain != "" {
		return nil
	}

	asStatus, err := p.tc.StatusToAS(ctx, status)
	if err != nil {
		return fmt.Errorf("federateStatusDelete: error converting status to as format: %s", err)
	}

	outboxIRI, err := url.Parse(status.Account.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateStatusDelete: error parsing outboxURI %s: %s", status.Account.OutboxURI, err)
	}

	actorIRI, err := url.Parse(status.Account.URI)
	if err != nil {
		return fmt.Errorf("federateStatusDelete: error parsing actorIRI %s: %s", status.Account.URI, err)
	}

	// create a delete and set the appropriate actor on it
	delete := streams.NewActivityStreamsDelete()

	// set the actor for the delete
	deleteActor := streams.NewActivityStreamsActorProperty()
	deleteActor.AppendIRI(actorIRI)
	delete.SetActivityStreamsActor(deleteActor)

	// Set the status as the 'object' property.
	deleteObject := streams.NewActivityStreamsObjectProperty()
	deleteObject.AppendActivityStreamsNote(asStatus)
	delete.SetActivityStreamsObject(deleteObject)

	// set the to and cc as the original to/cc of the original status
	delete.SetActivityStreamsTo(asStatus.GetActivityStreamsTo())
	delete.SetActivityStreamsCc(asStatus.GetActivityStreamsCc())

	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, delete)
	return err
}

func (p *processor) federateFollow(ctx context.Context, followRequest *gtsmodel.FollowRequest, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) error {
	// if both accounts are local there's nothing to do here
	if originAccount.Domain == "" && targetAccount.Domain == "" {
		return nil
	}

	follow := p.tc.FollowRequestToFollow(ctx, followRequest)

	asFollow, err := p.tc.FollowToAS(ctx, follow, originAccount, targetAccount)
	if err != nil {
		return fmt.Errorf("federateFollow: error converting follow to as format: %s", err)
	}

	outboxIRI, err := url.Parse(originAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateFollow: error parsing outboxURI %s: %s", originAccount.OutboxURI, err)
	}

	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, asFollow)
	return err
}

func (p *processor) federateUnfollow(ctx context.Context, follow *gtsmodel.Follow, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) error {
	// if both accounts are local there's nothing to do here
	if originAccount.Domain == "" && targetAccount.Domain == "" {
		return nil
	}

	// recreate the follow
	asFollow, err := p.tc.FollowToAS(ctx, follow, originAccount, targetAccount)
	if err != nil {
		return fmt.Errorf("federateUnfollow: error converting follow to as format: %s", err)
	}

	targetAccountURI, err := url.Parse(targetAccount.URI)
	if err != nil {
		return fmt.Errorf("error parsing uri %s: %s", targetAccount.URI, err)
	}

	// create an Undo and set the appropriate actor on it
	undo := streams.NewActivityStreamsUndo()
	undo.SetActivityStreamsActor(asFollow.GetActivityStreamsActor())

	// Set the recreated follow as the 'object' property.
	undoObject := streams.NewActivityStreamsObjectProperty()
	undoObject.AppendActivityStreamsFollow(asFollow)
	undo.SetActivityStreamsObject(undoObject)

	// Set the To of the undo as the target of the recreated follow
	undoTo := streams.NewActivityStreamsToProperty()
	undoTo.AppendIRI(targetAccountURI)
	undo.SetActivityStreamsTo(undoTo)

	outboxIRI, err := url.Parse(originAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateUnfollow: error parsing outboxURI %s: %s", originAccount.OutboxURI, err)
	}

	// send off the Undo
	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, undo)
	return err
}

func (p *processor) federateUnfave(ctx context.Context, fave *gtsmodel.StatusFave, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) error {
	// if both accounts are local there's nothing to do here
	if originAccount.Domain == "" && targetAccount.Domain == "" {
		return nil
	}

	// create the AS fave
	asFave, err := p.tc.FaveToAS(ctx, fave)
	if err != nil {
		return fmt.Errorf("federateFave: error converting fave to as format: %s", err)
	}

	targetAccountURI, err := url.Parse(targetAccount.URI)
	if err != nil {
		return fmt.Errorf("error parsing uri %s: %s", targetAccount.URI, err)
	}

	// create an Undo and set the appropriate actor on it
	undo := streams.NewActivityStreamsUndo()
	undo.SetActivityStreamsActor(asFave.GetActivityStreamsActor())

	// Set the fave as the 'object' property.
	undoObject := streams.NewActivityStreamsObjectProperty()
	undoObject.AppendActivityStreamsLike(asFave)
	undo.SetActivityStreamsObject(undoObject)

	// Set the To of the undo as the target of the fave
	undoTo := streams.NewActivityStreamsToProperty()
	undoTo.AppendIRI(targetAccountURI)
	undo.SetActivityStreamsTo(undoTo)

	outboxIRI, err := url.Parse(originAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateFave: error parsing outboxURI %s: %s", originAccount.OutboxURI, err)
	}
	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, undo)
	return err
}

func (p *processor) federateUnannounce(ctx context.Context, boost *gtsmodel.Status, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) error {
	if originAccount.Domain != "" {
		// nothing to do here
		return nil
	}

	asAnnounce, err := p.tc.BoostToAS(ctx, boost, originAccount, targetAccount)
	if err != nil {
		return fmt.Errorf("federateUnannounce: error converting status to announce: %s", err)
	}

	// create an Undo and set the appropriate actor on it
	undo := streams.NewActivityStreamsUndo()
	undo.SetActivityStreamsActor(asAnnounce.GetActivityStreamsActor())

	// Set the boost as the 'object' property.
	undoObject := streams.NewActivityStreamsObjectProperty()
	undoObject.AppendActivityStreamsAnnounce(asAnnounce)
	undo.SetActivityStreamsObject(undoObject)

	// set the to
	undo.SetActivityStreamsTo(asAnnounce.GetActivityStreamsTo())

	// set the cc
	undo.SetActivityStreamsCc(asAnnounce.GetActivityStreamsCc())

	outboxIRI, err := url.Parse(originAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateUnannounce: error parsing outboxURI %s: %s", originAccount.OutboxURI, err)
	}

	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, undo)
	return err
}

func (p *processor) federateAcceptFollowRequest(ctx context.Context, follow *gtsmodel.Follow) error {
	if follow.Account == nil {
		a, err := p.db.GetAccountByID(ctx, follow.AccountID)
		if err != nil {
			return err
		}
		follow.Account = a
	}
	originAccount := follow.Account

	if follow.TargetAccount == nil {
		a, err := p.db.GetAccountByID(ctx, follow.TargetAccountID)
		if err != nil {
			return err
		}
		follow.TargetAccount = a
	}
	targetAccount := follow.TargetAccount

	// if target account isn't from our domain we shouldn't do anything
	if targetAccount.Domain != "" {
		return nil
	}

	// if both accounts are local there's nothing to do here
	if originAccount.Domain == "" && targetAccount.Domain == "" {
		return nil
	}

	// recreate the AS follow
	asFollow, err := p.tc.FollowToAS(ctx, follow, originAccount, targetAccount)
	if err != nil {
		return fmt.Errorf("federateUnfollow: error converting follow to as format: %s", err)
	}

	acceptingAccountURI, err := url.Parse(targetAccount.URI)
	if err != nil {
		return fmt.Errorf("error parsing uri %s: %s", targetAccount.URI, err)
	}

	requestingAccountURI, err := url.Parse(originAccount.URI)
	if err != nil {
		return fmt.Errorf("error parsing uri %s: %s", targetAccount.URI, err)
	}

	// create an Accept
	accept := streams.NewActivityStreamsAccept()

	// set the accepting actor on it
	acceptActorProp := streams.NewActivityStreamsActorProperty()
	acceptActorProp.AppendIRI(acceptingAccountURI)
	accept.SetActivityStreamsActor(acceptActorProp)

	// Set the recreated follow as the 'object' property.
	acceptObject := streams.NewActivityStreamsObjectProperty()
	acceptObject.AppendActivityStreamsFollow(asFollow)
	accept.SetActivityStreamsObject(acceptObject)

	// Set the To of the accept as the originator of the follow
	acceptTo := streams.NewActivityStreamsToProperty()
	acceptTo.AppendIRI(requestingAccountURI)
	accept.SetActivityStreamsTo(acceptTo)

	outboxIRI, err := url.Parse(targetAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateAcceptFollowRequest: error parsing outboxURI %s: %s", originAccount.OutboxURI, err)
	}

	// send off the accept using the accepter's outbox
	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, accept)
	return err
}

func (p *processor) federateRejectFollowRequest(ctx context.Context, followRequest *gtsmodel.FollowRequest) error {
	if followRequest.Account == nil {
		a, err := p.db.GetAccountByID(ctx, followRequest.AccountID)
		if err != nil {
			return err
		}
		followRequest.Account = a
	}
	originAccount := followRequest.Account

	if followRequest.TargetAccount == nil {
		a, err := p.db.GetAccountByID(ctx, followRequest.TargetAccountID)
		if err != nil {
			return err
		}
		followRequest.TargetAccount = a
	}
	targetAccount := followRequest.TargetAccount

	// if target account isn't from our domain we shouldn't do anything
	if targetAccount.Domain != "" {
		return nil
	}

	// if both accounts are local there's nothing to do here
	if originAccount.Domain == "" && targetAccount.Domain == "" {
		return nil
	}

	// recreate the AS follow
	follow := p.tc.FollowRequestToFollow(ctx, followRequest)
	asFollow, err := p.tc.FollowToAS(ctx, follow, originAccount, targetAccount)
	if err != nil {
		return fmt.Errorf("federateUnfollow: error converting follow to as format: %s", err)
	}

	rejectingAccountURI, err := url.Parse(targetAccount.URI)
	if err != nil {
		return fmt.Errorf("error parsing uri %s: %s", targetAccount.URI, err)
	}

	requestingAccountURI, err := url.Parse(originAccount.URI)
	if err != nil {
		return fmt.Errorf("error parsing uri %s: %s", targetAccount.URI, err)
	}

	// create a Reject
	reject := streams.NewActivityStreamsReject()

	// set the rejecting actor on it
	acceptActorProp := streams.NewActivityStreamsActorProperty()
	acceptActorProp.AppendIRI(rejectingAccountURI)
	reject.SetActivityStreamsActor(acceptActorProp)

	// Set the recreated follow as the 'object' property.
	acceptObject := streams.NewActivityStreamsObjectProperty()
	acceptObject.AppendActivityStreamsFollow(asFollow)
	reject.SetActivityStreamsObject(acceptObject)

	// Set the To of the reject as the originator of the follow
	acceptTo := streams.NewActivityStreamsToProperty()
	acceptTo.AppendIRI(requestingAccountURI)
	reject.SetActivityStreamsTo(acceptTo)

	outboxIRI, err := url.Parse(targetAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateRejectFollowRequest: error parsing outboxURI %s: %s", originAccount.OutboxURI, err)
	}

	// send off the reject using the rejecting account's outbox
	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, reject)
	return err
}

func (p *processor) federateFave(ctx context.Context, fave *gtsmodel.StatusFave, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) error {
	// if both accounts are local there's nothing to do here
	if originAccount.Domain == "" && targetAccount.Domain == "" {
		return nil
	}

	// create the AS fave
	asFave, err := p.tc.FaveToAS(ctx, fave)
	if err != nil {
		return fmt.Errorf("federateFave: error converting fave to as format: %s", err)
	}

	outboxIRI, err := url.Parse(originAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateFave: error parsing outboxURI %s: %s", originAccount.OutboxURI, err)
	}
	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, asFave)
	return err
}

func (p *processor) federateAnnounce(ctx context.Context, boostWrapperStatus *gtsmodel.Status, boostingAccount *gtsmodel.Account, boostedAccount *gtsmodel.Account) error {
	announce, err := p.tc.BoostToAS(ctx, boostWrapperStatus, boostingAccount, boostedAccount)
	if err != nil {
		return fmt.Errorf("federateAnnounce: error converting status to announce: %s", err)
	}

	outboxIRI, err := url.Parse(boostingAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateAnnounce: error parsing outboxURI %s: %s", boostingAccount.OutboxURI, err)
	}

	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, announce)
	return err
}

func (p *processor) federateAccountUpdate(ctx context.Context, updatedAccount *gtsmodel.Account, originAccount *gtsmodel.Account) error {
	person, err := p.tc.AccountToAS(ctx, updatedAccount)
	if err != nil {
		return fmt.Errorf("federateAccountUpdate: error converting account to person: %s", err)
	}

	update, err := p.tc.WrapPersonInUpdate(person, originAccount)
	if err != nil {
		return fmt.Errorf("federateAccountUpdate: error wrapping person in update: %s", err)
	}

	outboxIRI, err := url.Parse(originAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateAnnounce: error parsing outboxURI %s: %s", originAccount.OutboxURI, err)
	}

	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, update)
	return err
}

func (p *processor) federateBlock(ctx context.Context, block *gtsmodel.Block) error {
	if block.Account == nil {
		blockAccount, err := p.db.GetAccountByID(ctx, block.AccountID)
		if err != nil {
			return fmt.Errorf("federateBlock: error getting block account from database: %s", err)
		}
		block.Account = blockAccount
	}

	if block.TargetAccount == nil {
		blockTargetAccount, err := p.db.GetAccountByID(ctx, block.TargetAccountID)
		if err != nil {
			return fmt.Errorf("federateBlock: error getting block target account from database: %s", err)
		}
		block.TargetAccount = blockTargetAccount
	}

	// if both accounts are local there's nothing to do here
	if block.Account.Domain == "" && block.TargetAccount.Domain == "" {
		return nil
	}

	asBlock, err := p.tc.BlockToAS(ctx, block)
	if err != nil {
		return fmt.Errorf("federateBlock: error converting block to AS format: %s", err)
	}

	outboxIRI, err := url.Parse(block.Account.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateBlock: error parsing outboxURI %s: %s", block.Account.OutboxURI, err)
	}

	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, asBlock)
	return err
}

func (p *processor) federateUnblock(ctx context.Context, block *gtsmodel.Block) error {
	if block.Account == nil {
		blockAccount, err := p.db.GetAccountByID(ctx, block.AccountID)
		if err != nil {
			return fmt.Errorf("federateUnblock: error getting block account from database: %s", err)
		}
		block.Account = blockAccount
	}

	if block.TargetAccount == nil {
		blockTargetAccount, err := p.db.GetAccountByID(ctx, block.TargetAccountID)
		if err != nil {
			return fmt.Errorf("federateUnblock: error getting block target account from database: %s", err)
		}
		block.TargetAccount = blockTargetAccount
	}

	// if both accounts are local there's nothing to do here
	if block.Account.Domain == "" && block.TargetAccount.Domain == "" {
		return nil
	}

	asBlock, err := p.tc.BlockToAS(ctx, block)
	if err != nil {
		return fmt.Errorf("federateUnblock: error converting block to AS format: %s", err)
	}

	targetAccountURI, err := url.Parse(block.TargetAccount.URI)
	if err != nil {
		return fmt.Errorf("federateUnblock: error parsing uri %s: %s", block.TargetAccount.URI, err)
	}

	// create an Undo and set the appropriate actor on it
	undo := streams.NewActivityStreamsUndo()
	undo.SetActivityStreamsActor(asBlock.GetActivityStreamsActor())

	// Set the block as the 'object' property.
	undoObject := streams.NewActivityStreamsObjectProperty()
	undoObject.AppendActivityStreamsBlock(asBlock)
	undo.SetActivityStreamsObject(undoObject)

	// Set the To of the undo as the target of the block
	undoTo := streams.NewActivityStreamsToProperty()
	undoTo.AppendIRI(targetAccountURI)
	undo.SetActivityStreamsTo(undoTo)

	outboxIRI, err := url.Parse(block.Account.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateUnblock: error parsing outboxURI %s: %s", block.Account.OutboxURI, err)
	}
	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, undo)
	return err
}

func (p *processor) federateReport(ctx context.Context, report *gtsmodel.Report) error {
	if report.TargetAccount == nil {
		reportTargetAccount, err := p.db.GetAccountByID(ctx, report.TargetAccountID)
		if err != nil {
			return fmt.Errorf("federateReport: error getting report target account from database: %w", err)
		}
		report.TargetAccount = reportTargetAccount
	}

	if len(report.StatusIDs) > 0 && len(report.Statuses) == 0 {
		statuses, err := p.db.GetStatuses(ctx, report.StatusIDs)
		if err != nil {
			return fmt.Errorf("federateReport: error getting report statuses from database: %w", err)
		}
		report.Statuses = statuses
	}

	flag, err := p.tc.ReportToASFlag(ctx, report)
	if err != nil {
		return fmt.Errorf("federateReport: error converting report to AS flag: %w", err)
	}

	// add bto so that our federating actor knows where to
	// send the Flag; it'll still use a shared inbox if possible
	reportTargetURI, err := url.Parse(report.TargetAccount.URI)
	if err != nil {
		return fmt.Errorf("federateReport: error parsing outboxURI %s: %w", report.TargetAccount.URI, err)
	}
	bTo := streams.NewActivityStreamsBtoProperty()
	bTo.AppendIRI(reportTargetURI)
	flag.SetActivityStreamsBto(bTo)

	// deliver the flag using the outbox of the
	// instance account to anonymize the report
	instanceAccount, err := p.db.GetInstanceAccount(ctx, "")
	if err != nil {
		return fmt.Errorf("federateReport: error getting instance account: %w", err)
	}

	outboxIRI, err := url.Parse(instanceAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateReport: error parsing outboxURI %s: %w", instanceAccount.OutboxURI, err)
	}

	_, err = p.federator.FederatingActor().Send(ctx, outboxIRI, flag)
	return err
}
