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
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) processFromClientAPI(clientMsg gtsmodel.FromClientAPI) error {
	switch clientMsg.APObjectType {
	case gtsmodel.ActivityStreamsNote:
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
	return fmt.Errorf("message type unprocessable: %+v", clientMsg)
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
