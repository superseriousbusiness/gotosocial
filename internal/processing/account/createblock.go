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

package account

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (p *processor) BlockCreate(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	// make sure the target account actually exists in our db
	targetAccount, err := p.db.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("BlockCreate: error getting account %s from the db: %s", targetAccountID, err))
	}

	// if requestingAccount already blocks target account, we don't need to do anything
	if blocked, err := p.db.IsBlocked(ctx, requestingAccount.ID, targetAccountID, false); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("BlockCreate: error checking existence of block: %s", err))
	} else if blocked {
		return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
	}

	// don't block yourself, silly
	if requestingAccount.ID == targetAccountID {
		return nil, gtserror.NewErrorNotAcceptable(fmt.Errorf("BlockCreate: account %s cannot block itself", requestingAccount.ID))
	}

	// make the block
	block := &gtsmodel.Block{}
	newBlockID, err := id.NewULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	block.ID = newBlockID
	block.AccountID = requestingAccount.ID
	block.Account = requestingAccount
	block.TargetAccountID = targetAccountID
	block.TargetAccount = targetAccount
	block.URI = uris.GenerateURIForBlock(requestingAccount.Username, newBlockID)

	// whack it in the database
	if err := p.db.PutBlock(ctx, block); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("BlockCreate: error creating block in db: %s", err))
	}

	// clear any follows or follow requests from the blocked account to the target account -- this is a simple delete
	if err := p.db.DeleteWhere(ctx, []db.Where{
		{Key: "account_id", Value: targetAccountID},
		{Key: "target_account_id", Value: requestingAccount.ID},
	}, &gtsmodel.Follow{}); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("BlockCreate: error removing follow in db: %s", err))
	}
	if err := p.db.DeleteWhere(ctx, []db.Where{
		{Key: "account_id", Value: targetAccountID},
		{Key: "target_account_id", Value: requestingAccount.ID},
	}, &gtsmodel.FollowRequest{}); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("BlockCreate: error removing follow in db: %s", err))
	}

	// clear any follows or follow requests from the requesting account to the target account --
	// this might require federation so we need to pass some messages around

	// check if a follow request exists from the requesting account to the target account, and remove it if it does (storing the URI for later)
	var frChanged bool
	var frURI string
	fr := &gtsmodel.FollowRequest{}
	if err := p.db.GetWhere(ctx, []db.Where{
		{Key: "account_id", Value: requestingAccount.ID},
		{Key: "target_account_id", Value: targetAccountID},
	}, fr); err == nil {
		frURI = fr.URI
		if err := p.db.DeleteByID(ctx, fr.ID, fr); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("BlockCreate: error removing follow request from db: %s", err))
		}
		frChanged = true
	}

	// now do the same thing for any existing follow
	var fChanged bool
	var fURI string
	f := &gtsmodel.Follow{}
	if err := p.db.GetWhere(ctx, []db.Where{
		{Key: "account_id", Value: requestingAccount.ID},
		{Key: "target_account_id", Value: targetAccountID},
	}, f); err == nil {
		fURI = f.URI
		if err := p.db.DeleteByID(ctx, f.ID, f); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("BlockCreate: error removing follow from db: %s", err))
		}
		fChanged = true
	}

	// follow request status changed so send the UNDO activity to the channel for async processing
	if frChanged {
		p.clientWorker.Queue(messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityUndo,
			GTSModel: &gtsmodel.Follow{
				AccountID:       requestingAccount.ID,
				TargetAccountID: targetAccountID,
				URI:             frURI,
			},
			OriginAccount: requestingAccount,
			TargetAccount: targetAccount,
		})
	}

	// follow status changed so send the UNDO activity to the channel for async processing
	if fChanged {
		p.clientWorker.Queue(messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityUndo,
			GTSModel: &gtsmodel.Follow{
				AccountID:       requestingAccount.ID,
				TargetAccountID: targetAccountID,
				URI:             fURI,
			},
			OriginAccount: requestingAccount,
			TargetAccount: targetAccount,
		})
	}

	// handle the rest of the block process asynchronously
	p.clientWorker.Queue(messages.FromClientAPI{
		APObjectType:   ap.ActivityBlock,
		APActivityType: ap.ActivityCreate,
		GTSModel:       block,
		OriginAccount:  requestingAccount,
		TargetAccount:  targetAccount,
	})

	return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
}
