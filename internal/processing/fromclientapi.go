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
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) processFromClientAPI(clientMsg gtsmodel.FromClientAPI) error {
	switch clientMsg.APActivityType {
	case gtsmodel.ActivityStreamsCreate:
		// CREATE
		switch clientMsg.APObjectType {
		case gtsmodel.ActivityStreamsNote:
			// CREATE NOTE
			status, ok := clientMsg.GTSModel.(*gtsmodel.Status)
			if !ok {
				return errors.New("note was not parseable as *gtsmodel.Status")
			}

			if err := p.timelineStatus(status); err != nil {
				return err
			}

			if err := p.notifyStatus(status); err != nil {
				return err
			}

			if status.VisibilityAdvanced != nil && status.VisibilityAdvanced.Federated {
				return p.federateStatus(status)
			}
		case gtsmodel.ActivityStreamsFollow:
			// CREATE FOLLOW REQUEST
			followRequest, ok := clientMsg.GTSModel.(*gtsmodel.FollowRequest)
			if !ok {
				return errors.New("followrequest was not parseable as *gtsmodel.FollowRequest")
			}

			if err := p.notifyFollowRequest(followRequest, clientMsg.TargetAccount); err != nil {
				return err
			}

			return p.federateFollow(followRequest, clientMsg.OriginAccount, clientMsg.TargetAccount)
		case gtsmodel.ActivityStreamsLike:
			// CREATE LIKE/FAVE
			fave, ok := clientMsg.GTSModel.(*gtsmodel.StatusFave)
			if !ok {
				return errors.New("fave was not parseable as *gtsmodel.StatusFave")
			}

			if err := p.notifyFave(fave, clientMsg.TargetAccount); err != nil {
				return err
			}

			return p.federateFave(fave, clientMsg.OriginAccount, clientMsg.TargetAccount)

		case gtsmodel.ActivityStreamsAnnounce:
			// CREATE BOOST/ANNOUNCE
			boostWrapperStatus, ok := clientMsg.GTSModel.(*gtsmodel.Status)
			if !ok {
				return errors.New("boost was not parseable as *gtsmodel.Status")
			}

			if err := p.notifyAnnounce(boostWrapperStatus); err != nil {
				return err
			}

			return p.federateAnnounce(boostWrapperStatus, clientMsg.OriginAccount, clientMsg.TargetAccount)
		}
	case gtsmodel.ActivityStreamsUpdate:
		// UPDATE
		switch clientMsg.APObjectType {
		case gtsmodel.ActivityStreamsProfile, gtsmodel.ActivityStreamsPerson:
			// UPDATE ACCOUNT/PROFILE
			account, ok := clientMsg.GTSModel.(*gtsmodel.Account)
			if !ok {
				return errors.New("account was not parseable as *gtsmodel.Account")
			}

			return p.federateAccountUpdate(account, clientMsg.OriginAccount)
		}
	case gtsmodel.ActivityStreamsAccept:
		// ACCEPT
		switch clientMsg.APObjectType {
		case gtsmodel.ActivityStreamsFollow:
			// ACCEPT FOLLOW
			follow, ok := clientMsg.GTSModel.(*gtsmodel.Follow)
			if !ok {
				return errors.New("accept was not parseable as *gtsmodel.Follow")
			}

			if err := p.notifyFollow(follow, clientMsg.TargetAccount); err != nil {
				return err
			}

			return p.federateAcceptFollowRequest(follow, clientMsg.OriginAccount, clientMsg.TargetAccount)
		}
	case gtsmodel.ActivityStreamsUndo:
		// UNDO
		switch clientMsg.APObjectType {
		case gtsmodel.ActivityStreamsFollow:
			// UNDO FOLLOW
			follow, ok := clientMsg.GTSModel.(*gtsmodel.Follow)
			if !ok {
				return errors.New("undo was not parseable as *gtsmodel.Follow")
			}
			return p.federateUnfollow(follow, clientMsg.OriginAccount, clientMsg.TargetAccount)
		}
	}
	return nil
}

func (p *processor) federateStatus(status *gtsmodel.Status) error {
	asStatus, err := p.tc.StatusToAS(status)
	if err != nil {
		return fmt.Errorf("federateStatus: error converting status to as format: %s", err)
	}

	outboxIRI, err := url.Parse(status.GTSAuthorAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateStatus: error parsing outboxURI %s: %s", status.GTSAuthorAccount.OutboxURI, err)
	}

	_, err = p.federator.FederatingActor().Send(context.Background(), outboxIRI, asStatus)
	return err
}

func (p *processor) federateFollow(followRequest *gtsmodel.FollowRequest, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) error {
	// if both accounts are local there's nothing to do here
	if originAccount.Domain == "" && targetAccount.Domain == "" {
		return nil
	}

	follow := p.tc.FollowRequestToFollow(followRequest)

	asFollow, err := p.tc.FollowToAS(follow, originAccount, targetAccount)
	if err != nil {
		return fmt.Errorf("federateFollow: error converting follow to as format: %s", err)
	}

	outboxIRI, err := url.Parse(originAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateFollow: error parsing outboxURI %s: %s", originAccount.OutboxURI, err)
	}

	_, err = p.federator.FederatingActor().Send(context.Background(), outboxIRI, asFollow)
	return err
}

func (p *processor) federateUnfollow(follow *gtsmodel.Follow, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) error {
	// if both accounts are local there's nothing to do here
	if originAccount.Domain == "" && targetAccount.Domain == "" {
		return nil
	}

	// recreate the follow
	asFollow, err := p.tc.FollowToAS(follow, originAccount, targetAccount)
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
	_, err = p.federator.FederatingActor().Send(context.Background(), outboxIRI, undo)
	return err
}

func (p *processor) federateAcceptFollowRequest(follow *gtsmodel.Follow, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) error {
	// if both accounts are local there's nothing to do here
	if originAccount.Domain == "" && targetAccount.Domain == "" {
		return nil
	}

	// recreate the AS follow
	asFollow, err := p.tc.FollowToAS(follow, originAccount, targetAccount)
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
	_, err = p.federator.FederatingActor().Send(context.Background(), outboxIRI, accept)
	return err
}

func (p *processor) federateFave(fave *gtsmodel.StatusFave, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) error {
	// if both accounts are local there's nothing to do here
	if originAccount.Domain == "" && targetAccount.Domain == "" {
		return nil
	}

	// create the AS fave
	asFave, err := p.tc.FaveToAS(fave)
	if err != nil {
		return fmt.Errorf("federateFave: error converting fave to as format: %s", err)
	}

	outboxIRI, err := url.Parse(originAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateFave: error parsing outboxURI %s: %s", originAccount.OutboxURI, err)
	}
	_, err = p.federator.FederatingActor().Send(context.Background(), outboxIRI, asFave)
	return err
}

func (p *processor) federateAnnounce(boostWrapperStatus *gtsmodel.Status, boostingAccount *gtsmodel.Account, boostedAccount *gtsmodel.Account) error {
	announce, err := p.tc.BoostToAS(boostWrapperStatus, boostingAccount, boostedAccount)
	if err != nil {
		return fmt.Errorf("federateAnnounce: error converting status to announce: %s", err)
	}

	outboxIRI, err := url.Parse(boostingAccount.OutboxURI)
	if err != nil {
		return fmt.Errorf("federateAnnounce: error parsing outboxURI %s: %s", boostingAccount.OutboxURI, err)
	}

	_, err = p.federator.FederatingActor().Send(context.Background(), outboxIRI, announce)
	return err
}

func (p *processor) federateAccountUpdate(updatedAccount *gtsmodel.Account, originAccount *gtsmodel.Account) error {
	person, err := p.tc.AccountToAS(updatedAccount)
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

	_, err = p.federator.FederatingActor().Send(context.Background(), outboxIRI, update)
	return err
}
