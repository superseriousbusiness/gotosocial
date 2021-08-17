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

package account

import (
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) FollowRemove(requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	// if there's a block between the accounts we shouldn't do anything
	blocked, err := p.db.Blocked(requestingAccount.ID, targetAccountID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	if blocked {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("AccountFollowRemove: block exists between accounts"))
	}

	// make sure the target account actually exists in our db
	targetAcct := &gtsmodel.Account{}
	if err := p.db.GetByID(targetAccountID, targetAcct); err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("AccountFollowRemove: account %s not found in the db: %s", targetAccountID, err))
		}
	}

	// check if a follow request exists, and remove it if it does (storing the URI for later)
	var frChanged bool
	var frURI string
	fr := &gtsmodel.FollowRequest{}
	if err := p.db.GetWhere([]db.Where{
		{Key: "account_id", Value: requestingAccount.ID},
		{Key: "target_account_id", Value: targetAccountID},
	}, fr); err == nil {
		frURI = fr.URI
		if err := p.db.DeleteByID(fr.ID, fr); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("AccountFollowRemove: error removing follow request from db: %s", err))
		}
		frChanged = true
	}

	// now do the same thing for any existing follow
	var fChanged bool
	var fURI string
	f := &gtsmodel.Follow{}
	if err := p.db.GetWhere([]db.Where{
		{Key: "account_id", Value: requestingAccount.ID},
		{Key: "target_account_id", Value: targetAccountID},
	}, f); err == nil {
		fURI = f.URI
		if err := p.db.DeleteByID(f.ID, f); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("AccountFollowRemove: error removing follow from db: %s", err))
		}
		fChanged = true
	}

	// follow request status changed so send the UNDO activity to the channel for async processing
	if frChanged {
		p.fromClientAPI <- gtsmodel.FromClientAPI{
			APObjectType:   gtsmodel.ActivityStreamsFollow,
			APActivityType: gtsmodel.ActivityStreamsUndo,
			GTSModel: &gtsmodel.Follow{
				AccountID:       requestingAccount.ID,
				TargetAccountID: targetAccountID,
				URI:             frURI,
			},
			OriginAccount: requestingAccount,
			TargetAccount: targetAcct,
		}
	}

	// follow status changed so send the UNDO activity to the channel for async processing
	if fChanged {
		p.fromClientAPI <- gtsmodel.FromClientAPI{
			APObjectType:   gtsmodel.ActivityStreamsFollow,
			APActivityType: gtsmodel.ActivityStreamsUndo,
			GTSModel: &gtsmodel.Follow{
				AccountID:       requestingAccount.ID,
				TargetAccountID: targetAccountID,
				URI:             fURI,
			},
			OriginAccount: requestingAccount,
			TargetAccount: targetAcct,
		}
	}

	// return whatever relationship results from all this
	return p.RelationshipGet(requestingAccount, targetAccountID)
}
