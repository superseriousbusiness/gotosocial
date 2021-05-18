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

package message

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/pub"
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

			if err := p.notifyStatus(status); err != nil {
				return err
			}

			if status.VisibilityAdvanced.Federated {
				return p.federateStatus(status)
			}
			return nil
		}
	case gtsmodel.ActivityStreamsUpdate:
		// UPDATE
	case gtsmodel.ActivityStreamsAccept:
		// ACCEPT
		follow, ok := clientMsg.GTSModel.(*gtsmodel.Follow)
		if !ok {
			return errors.New("accept was not parseable as *gtsmodel.Follow")
		}
		return p.federateAcceptFollowRequest(follow)
	}
	return nil
}

func (p *processor) federateStatus(status *gtsmodel.Status) error {
	// // derive the sending account -- it might be attached to the status already
	// sendingAcct := &gtsmodel.Account{}
	// if status.GTSAccount != nil {
	// 	sendingAcct = status.GTSAccount
	// } else {
	// 	// it wasn't attached so get it from the db instead
	// 	if err := p.db.GetByID(status.AccountID, sendingAcct); err != nil {
	// 		return err
	// 	}
	// }

	// outboxURI, err := url.Parse(sendingAcct.OutboxURI)
	// if err != nil {
	// 	return err
	// }

	// // convert the status to AS format Note
	// note, err := p.tc.StatusToAS(status)
	// if err != nil {
	// 	return err
	// }

	// _, err = p.federator.FederatingActor().Send(context.Background(), outboxURI, note)
	return nil
}

func (p *processor) federateAcceptFollowRequest(follow *gtsmodel.Follow) error {

	followAccepter := &gtsmodel.Account{}
	if err := p.db.GetByID(follow.TargetAccountID, followAccepter); err != nil {
		return fmt.Errorf("error federating follow accept: %s", err)
	}
	followAccepterIRI, err := url.Parse(followAccepter.URI)
	if err != nil {
		return fmt.Errorf("error parsing URL: %s", err)
	}
	followAccepterOutboxIRI, err := url.Parse(followAccepter.OutboxURI)
	if err != nil {
		return fmt.Errorf("error parsing URL: %s", err)
	}
	me := streams.NewActivityStreamsActorProperty()
	me.AppendIRI(followAccepterIRI)

	followRequester := &gtsmodel.Account{}
	if err := p.db.GetByID(follow.AccountID, followRequester); err != nil {
		return fmt.Errorf("error federating follow accept: %s", err)
	}
	requesterIRI, err := url.Parse(followRequester.URI)
	if err != nil {
		return fmt.Errorf("error parsing URL: %s", err)
	}
	them := streams.NewActivityStreamsActorProperty()
	them.AppendIRI(requesterIRI)

	// prepare the follow
	ASFollow := streams.NewActivityStreamsFollow()
	// set the follow requester as the actor
	ASFollow.SetActivityStreamsActor(them)
	// set the ID from the follow
	ASFollowURI, err := url.Parse(follow.URI)
	if err != nil {
		return fmt.Errorf("error parsing URL: %s", err)
	}
	ASFollowIDProp := streams.NewJSONLDIdProperty()
	ASFollowIDProp.SetIRI(ASFollowURI)
	ASFollow.SetJSONLDId(ASFollowIDProp)

	// set the object as the accepter URI
	ASFollowObjectProp := streams.NewActivityStreamsObjectProperty()
	ASFollowObjectProp.AppendIRI(followAccepterIRI)

	// Prepare the response.
	ASAccept := streams.NewActivityStreamsAccept()
	// Set us as the 'actor'.
	ASAccept.SetActivityStreamsActor(me)

	// Set the Follow as the 'object' property.
	ASAcceptObject := streams.NewActivityStreamsObjectProperty()
	ASAcceptObject.AppendActivityStreamsFollow(ASFollow)
	ASAccept.SetActivityStreamsObject(ASAcceptObject)

	// Add all actors on the original Follow to the 'to' property.
	ASAcceptTo := streams.NewActivityStreamsToProperty()
	followActors := ASFollow.GetActivityStreamsActor()
	for iter := followActors.Begin(); iter != followActors.End(); iter = iter.Next() {
		id, err := pub.ToId(iter)
		if err != nil {
			return err
		}
		ASAcceptTo.AppendIRI(id)
	}
	ASAccept.SetActivityStreamsTo(ASAcceptTo)

	_, err = p.federator.FederatingActor().Send(context.Background(), followAccepterOutboxIRI, ASAccept)
	return err
}
