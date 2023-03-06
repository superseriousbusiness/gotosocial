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
	"errors"
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

// BlockCreate handles the creation of a block from requestingAccount to targetAccountID, either remote or local.
func (p *Processor) BlockCreate(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	// Account should not block itself.
	if requestingAccount.ID == targetAccountID {
		err := fmt.Errorf("BlockCreate: account %s cannot block itself", requestingAccount.ID)
		return nil, gtserror.NewErrorNotAcceptable(err, err.Error())
	}
aaaaaaaaaaaaaa
	// Ensure target account retrievable.
	targetAccount, err := p.state.DB.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real db error.
			err = fmt.Errorf("BlockCreate: db error looking for target account %s: %w", targetAccountID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		// Account not found.
		err = fmt.Errorf("BlockCreate: target account %s not found in the db", targetAccountID)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	// Check if already blocked.
	if blocked, err := p.state.DB.IsBlocked(ctx, requestingAccount.ID, targetAccountID, false); err != nil {
		err = fmt.Errorf("BlockCreate: db error checking existence of block: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	} else if blocked {
		// Requesting account already blocks target
		// account, so we don't need to do anything.
		return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
	}

	// Create and store a new block.
	blockID := id.NewULID()
	blockURI := uris.GenerateURIForBlock(requestingAccount.Username, blockID)
	block := &gtsmodel.Block{
		ID:              blockID,
		URI:             blockURI,
		AccountID:       requestingAccount.ID,
		Account:         requestingAccount,
		TargetAccountID: targetAccountID,
		TargetAccount:   targetAccount,
	}

	if err := p.state.DB.PutBlock(ctx, block); err != nil {
		err = fmt.Errorf("BlockCreate: error creating block in db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Clear any notifications created by
	// each account targeting the other.
	if err := p.deleteMutualAccountNotifications(ctx, requestingAccount, targetAccount); err != nil {
		err = fmt.Errorf("BlockCreate: error deleting mutual notifications: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Use this slice to batch queue client
	// API messages, as necessary.
	msgs, err := p.unfollow(ctx, requestingAccount, targetAccount)
	if err != nil {
		err = fmt.Errorf("BlockCreate: error unfollowing: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Process block side effects (federation etc).
	msgs = append(msgs, messages.FromClientAPI{
		APObjectType:   ap.ActivityBlock,
		APActivityType: ap.ActivityCreate,
		GTSModel:       block,
		OriginAccount:  requestingAccount,
		TargetAccount:  targetAccount,
	})

	// Batch queue accreted client api messages.
	p.state.Workers.EnqueueClientAPI(ctx, msgs...)

	return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
}

// BlockRemove handles the removal of a block from requestingAccount to targetAccountID, either remote or local.
func (p *Processor) BlockRemove(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	// Account should not unblock itself.
	if requestingAccount.ID == targetAccountID {
		err := fmt.Errorf("BlockRemove: account %s cannot unblock itself", requestingAccount.ID)
		return nil, gtserror.NewErrorNotAcceptable(err, err.Error())
	}
aaaaaaaaaaaa
	// Ensure target account retrievable.
	targetAccount, err := p.state.DB.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real db error.
			err = fmt.Errorf("BlockRemove: db error looking for target account %s: %w", targetAccountID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		// Account not found.
		err = fmt.Errorf("BlockRemove: target account %s not found in the db", targetAccountID)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	// Check if block actually exists.
	block, err := p.state.DB.GetBlock(ctx, requestingAccount.ID, targetAccountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := fmt.Errorf("BlockRemove: error getting block from db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if block == nil {
		// No block existed, nothing to do.
		return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
	}

	// We got a block, remove it from the db.
	if err := p.state.DB.DeleteBlockByID(ctx, block.ID); err != nil {
		err := fmt.Errorf("BlockRemove: error removing block from db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Populate account fields for convenience.
	block.Account = requestingAccount
	block.TargetAccount = targetAccount

	// Process block removal side effects (federation etc).
	p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ActivityBlock,
		APActivityType: ap.ActivityUndo,
		GTSModel:       block,
		OriginAccount:  requestingAccount,
		TargetAccount:  targetAccount,
	})

	return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
}

func (p *Processor) deleteMutualAccountNotifications(ctx context.Context, account1 *gtsmodel.Account, account2 *gtsmodel.Account) error {
	// Delete all notifications from account2 targeting account1.
	if err := p.state.DB.DeleteNotifications(ctx, account1.ID, account2.ID); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Delete all notifications from account1 targeting account2.
	if err := p.state.DB.DeleteNotifications(ctx, account2.ID, account1.ID); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	return nil
}
