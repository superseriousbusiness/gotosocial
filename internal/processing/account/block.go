// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package account

import (
	"context"
	"errors"
	"fmt"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// BlockCreate handles the creation of a block from requestingAccount to targetAccountID, either remote or local.
func (p *Processor) BlockCreate(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	targetAccount, existingBlock, errWithCode := p.getBlockTarget(ctx, requestingAccount, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if existingBlock != nil {
		// Block already exists, nothing to do.
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

	// Ensure each account unfollows the other.
	// We only care about processing unfollow side
	// effects from requesting account -> target
	// account, since requesting account is ours,
	// and target account might not be.
	msgs, err := p.unfollow(ctx, requestingAccount, targetAccount)
	if err != nil {
		err = fmt.Errorf("BlockCreate: error unfollowing: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Ensure unfollowed in other direction;
	// ignore/don't process returned messages.
	if _, err := p.unfollow(ctx, targetAccount, requestingAccount); err != nil {
		err = fmt.Errorf("BlockCreate: error unfollowing: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Process block side effects (federation etc).
	msgs = append(msgs, &messages.FromClientAPI{
		APObjectType:   ap.ActivityBlock,
		APActivityType: ap.ActivityCreate,
		GTSModel:       block,
		Origin:         requestingAccount,
		Target:         targetAccount,
	})

	// Batch queue accreted client api messages.
	p.state.Workers.Client.Queue.Push(msgs...)

	return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
}

// BlockRemove handles the removal of a block from requestingAccount to targetAccountID, either remote or local.
func (p *Processor) BlockRemove(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	targetAccount, existingBlock, errWithCode := p.getBlockTarget(ctx, requestingAccount, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if existingBlock == nil {
		// Already not blocked, nothing to do.
		return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
	}

	// We got a block, remove it from the db.
	if err := p.state.DB.DeleteBlockByID(ctx, existingBlock.ID); err != nil {
		err := fmt.Errorf("BlockRemove: error removing block from db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Populate account fields for convenience.
	existingBlock.Account = requestingAccount
	existingBlock.TargetAccount = targetAccount

	// Process block removal side effects (federation etc).
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityBlock,
		APActivityType: ap.ActivityUndo,
		GTSModel:       existingBlock,
		Origin:         requestingAccount,
		Target:         targetAccount,
	})

	return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
}

// BlocksGet ...
func (p *Processor) BlocksGet(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	blocks, err := p.state.DB.GetAccountBlocks(ctx,
		requestingAccount.ID,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for empty response.
	count := len(blocks)
	if len(blocks) == 0 {
		return util.EmptyPageableResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := blocks[count-1].ID
	hi := blocks[0].ID

	items := make([]interface{}, 0, count)

	for _, block := range blocks {
		// Convert target account to frontend API model. (target will never be nil)
		account, err := p.converter.AccountToAPIAccountBlocked(ctx, block.TargetAccount)
		if err != nil {
			log.Errorf(ctx, "error converting account to public api account: %v", err)
			continue
		}

		// Append target to return items.
		items = append(items, account)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/blocks",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}

func (p *Processor) getBlockTarget(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*gtsmodel.Account, *gtsmodel.Block, gtserror.WithCode) {
	// Account should not block or unblock itself.
	if requestingAccount.ID == targetAccountID {
		err := fmt.Errorf("getBlockTarget: account %s cannot block or unblock itself", requestingAccount.ID)
		return nil, nil, gtserror.NewErrorNotAcceptable(err, err.Error())
	}

	// Ensure target account retrievable.
	targetAccount, err := p.state.DB.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real db error.
			err = fmt.Errorf("getBlockTarget: db error looking for target account %s: %w", targetAccountID, err)
			return nil, nil, gtserror.NewErrorInternalError(err)
		}
		// Account not found.
		err = fmt.Errorf("getBlockTarget: target account %s not found in the db", targetAccountID)
		return nil, nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	// Check if currently blocked.
	block, err := p.state.DB.GetBlock(ctx, requestingAccount.ID, targetAccountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("getBlockTarget: db error checking existing block: %w", err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	return targetAccount, block, nil
}
