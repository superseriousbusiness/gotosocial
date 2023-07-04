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

package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

// DomainBlockCreate creates an instance-level block against the given domain,
// and then processes side effects of that block (deleting accounts, media, etc).
//
// If a domain block already exists for the domain, side effects will be retried.
func (p *Processor) DomainBlockCreate(
	ctx context.Context,
	account *gtsmodel.Account,
	domain string,
	obfuscate bool,
	publicComment string,
	privateComment string,
	subscriptionID string,
) (*apimodel.DomainBlock, gtserror.WithCode) {
	// Check if a block already exists for this domain.
	domainBlock, err := p.state.DB.GetDomainBlock(ctx, domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Something went wrong in the DB.
		err = gtserror.Newf("db error getting domain block %s: %w", domain, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if domainBlock == nil {
		// No block exists yet, create it.
		domainBlock = &gtsmodel.DomainBlock{
			ID:                 id.NewULID(),
			Domain:             domain,
			CreatedByAccountID: account.ID,
			PrivateComment:     text.SanitizePlaintext(privateComment),
			PublicComment:      text.SanitizePlaintext(publicComment),
			Obfuscate:          &obfuscate,
			SubscriptionID:     subscriptionID,
		}

		// Insert the new block into the database.
		if err := p.state.DB.CreateDomainBlock(ctx, domainBlock); err != nil {
			err = gtserror.Newf("db error putting domain block %s: %s", domain, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	// Process the side effects of the domain block
	// asynchronously since it might take a while.
	p.state.Workers.ClientAPI.Enqueue(func(ctx context.Context) {
		p.domainBlockSideEffects(ctx, account, domainBlock)
	})

	return p.apiDomainBlock(ctx, domainBlock)
}

// DomainBlocksImport handles the import of multiple domain blocks,
// by calling the DomainBlockCreate function for each domain in the
// provided file. Will return a slice of processed domain blocks.
//
// In the case of total failure, a gtserror.WithCode will be returned
// so that the caller can respond appropriately. In the case of
// partial or total success, a MultiStatus model will be returned,
// which contains information about success/failure count, so that
// the caller can retry any failures as they wish.
func (p *Processor) DomainBlocksImport(
	ctx context.Context,
	account *gtsmodel.Account,
	domainsF *multipart.FileHeader,
) (*apimodel.MultiStatus, gtserror.WithCode) {
	// Open the provided file.
	file, err := domainsF.Open()
	if err != nil {
		err = gtserror.Newf("error opening attachment: %w", err)
		return nil, gtserror.NewErrorBadRequest(err)
	}
	defer file.Close()

	// Copy the file contents into a buffer.
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, file)
	if err != nil {
		err = gtserror.Newf("error reading attachment: %w", err)
		return nil, gtserror.NewErrorBadRequest(err)
	}

	// Ensure we actually read something.
	if size == 0 {
		err = gtserror.New("error reading attachment: size 0 bytes")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Parse bytes as slice of domain blocks.
	domainBlocks := make([]*apimodel.DomainBlock, 0)
	if err := json.Unmarshal(buf.Bytes(), &domainBlocks); err != nil {
		err = gtserror.Newf("error parsing attachment as domain blocks: %w", err)
		return nil, gtserror.NewErrorBadRequest(err)
	}

	count := len(domainBlocks)
	if count == 0 {
		err = gtserror.New("error importing domain blocks: 0 entries provided")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Try to process each domain block, differentiating
	// between successes and errors so that the caller can
	// try failed imports again if desired.
	multiStatusEntries := make([]apimodel.MultiStatusEntry, 0, count)

	for _, domainBlock := range domainBlocks {
		var (
			domain         = domainBlock.Domain.Domain
			obfuscate      = domainBlock.Obfuscate
			publicComment  = domainBlock.PublicComment
			privateComment = domainBlock.PrivateComment
			subscriptionID = "" // No sub ID for imports.
			errWithCode    gtserror.WithCode
		)

		domainBlock, errWithCode = p.DomainBlockCreate(
			ctx,
			account,
			domain,
			obfuscate,
			publicComment,
			privateComment,
			subscriptionID,
		)

		var entry *apimodel.MultiStatusEntry

		if errWithCode != nil {
			entry = &apimodel.MultiStatusEntry{
				// Use the failed domain entry as the resource value.
				Resource: domain,
				Message:  errWithCode.Safe(),
				Status:   errWithCode.Code(),
			}
		} else {
			entry = &apimodel.MultiStatusEntry{
				// Use successfully created API model domain block as the resource value.
				Resource: domainBlock,
				Message:  http.StatusText(http.StatusOK),
				Status:   http.StatusOK,
			}
		}

		multiStatusEntries = append(multiStatusEntries, *entry)
	}

	return apimodel.NewMultiStatus(multiStatusEntries), nil
}

// DomainBlocksGet returns all existing domain blocks. If export is
// true, the format will be suitable for writing out to an export.
func (p *Processor) DomainBlocksGet(ctx context.Context, account *gtsmodel.Account, export bool) ([]*apimodel.DomainBlock, gtserror.WithCode) {
	domainBlocks, err := p.state.DB.GetDomainBlocks(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting domain blocks: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiDomainBlocks := make([]*apimodel.DomainBlock, 0, len(domainBlocks))
	for _, domainBlock := range domainBlocks {
		apiDomainBlock, errWithCode := p.apiDomainBlock(ctx, domainBlock)
		if errWithCode != nil {
			return nil, errWithCode
		}

		apiDomainBlocks = append(apiDomainBlocks, apiDomainBlock)
	}

	return apiDomainBlocks, nil
}

// DomainBlockGet returns one domain block with the given id. If export
// is true, the format will be suitable for writing out to an export.
func (p *Processor) DomainBlockGet(ctx context.Context, id string, export bool) (*apimodel.DomainBlock, gtserror.WithCode) {
	domainBlock, err := p.state.DB.GetDomainBlockByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("no domain block exists with id %s", id)
			return nil, gtserror.NewErrorNotFound(err, err.Error())
		}

		// Something went wrong in the DB.
		err = gtserror.Newf("db error getting domain block %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiDomainBlock(ctx, domainBlock)
}

// DomainBlockDelete removes one domain block with the given ID,
// and processes side effects of removing the block asynchronously.
func (p *Processor) DomainBlockDelete(ctx context.Context, account *gtsmodel.Account, id string) (*apimodel.DomainBlock, gtserror.WithCode) {
	domainBlock, err := p.state.DB.GetDomainBlockByID(ctx, id)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real error.
			err = gtserror.Newf("db error getting domain block: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// There are just no entries for this ID.
		err = fmt.Errorf("no domain block entry exists with ID %s", id)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	// Prepare the domain block to return, *before* the deletion goes through.
	apiDomainBlock, errWithCode := p.apiDomainBlock(ctx, domainBlock)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Copy value of the domain block.
	domainBlockC := new(gtsmodel.DomainBlock)
	*domainBlockC = *domainBlock

	// Delete the original domain block.
	if err := p.state.DB.DeleteDomainBlock(ctx, domainBlock.Domain); err != nil {
		err = gtserror.Newf("db error deleting domain block: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Process the side effects of the domain unblock
	// asynchronously since it might take a while.
	p.state.Workers.ClientAPI.Enqueue(func(ctx context.Context) {
		p.domainUnblockSideEffects(ctx, domainBlockC) // Use the copy.
	})

	return apiDomainBlock, nil
}

// stubbifyInstance renders the given instance as a stub,
// removing most information from it and marking it as
// suspended.
//
// For caller's convenience, this function returns the db
// names of all columns that are updated by it.
func stubbifyInstance(instance *gtsmodel.Instance, domainBlockID string) []string {
	instance.Title = ""
	instance.SuspendedAt = time.Now()
	instance.DomainBlockID = domainBlockID
	instance.ShortDescription = ""
	instance.Description = ""
	instance.Terms = ""
	instance.ContactEmail = ""
	instance.ContactAccountUsername = ""
	instance.ContactAccountID = ""
	instance.Version = ""

	return []string{
		"title",
		"suspended_at",
		"domain_block_id",
		"short_description",
		"description",
		"terms",
		"contact_email",
		"contact_account_username",
		"contact_account_id",
		"version",
	}
}

// apiDomainBlock is a cheeky shortcut function for returning the API
// version of the given domainBlock, or an appropriate error if
// something goes wrong.
func (p *Processor) apiDomainBlock(ctx context.Context, domainBlock *gtsmodel.DomainBlock) (*apimodel.DomainBlock, gtserror.WithCode) {
	apiDomainBlock, err := p.tc.DomainBlockToAPIDomainBlock(ctx, domainBlock, false)
	if err != nil {
		err = gtserror.Newf("error converting domain block for %s to api model : %w", domainBlock.Domain, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiDomainBlock, nil
}

// rangeAccounts iterates through all accounts originating from the
// given domain, and calls the provided range function on each account.
// If an error is returned from the range function, the loop will stop
// and return the error.
func (p *Processor) rangeAccounts(
	ctx context.Context,
	domain string,
	rangeF func(*gtsmodel.Account) error,
) error {
	var (
		limit = 50   // Limit selection to avoid spiking mem/cpu.
		maxID string // Start with empty string to select from top.
	)

	for {
		// Get (next) page of accounts.
		accounts, err := p.state.DB.GetInstanceAccounts(ctx, domain, maxID, limit)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			// Real db error.
			return gtserror.Newf("db error getting instance accounts: %w", err)
		}

		if len(accounts) == 0 {
			// No accounts left, we're done.
			return nil
		}

		// Set next max ID for paging down.
		maxID = accounts[len(accounts)-1].ID

		// Call provided range function.
		for _, account := range accounts {
			if err := rangeF(account); err != nil {
				return err
			}
		}
	}
}

// domainBlockSideEffects processes the side effects of a domain block:
//
//  1. Strip most info away from the instance entry for the domain.
//  2. Pass each account from the domain to the processor for deletion.
//
// It should be called asynchronously, since it can take a while when
// there are many accounts present on the given domain.
func (p *Processor) domainBlockSideEffects(ctx context.Context, account *gtsmodel.Account, block *gtsmodel.DomainBlock) {
	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"domain", block.Domain},
		}...)
	l.Debug("processing domain block side effects")

	// If we have an instance entry for this domain,
	// update it with the new block ID and clear all fields
	instance, err := p.state.DB.GetInstance(ctx, block.Domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		l.Errorf("db error getting instance %s: %q", block.Domain, err)
	}

	if instance != nil {
		// We had an entry for this domain.
		columns := stubbifyInstance(instance, block.ID)
		if err := p.state.DB.UpdateInstance(ctx, instance, columns...); err != nil {
			l.Errorf("db error updating instance: %s", err)
		} else {
			l.Debug("instance entry updated")
		}
	}

	// For each account that belongs to this domain, create
	// an account delete message to process via the client API
	// worker pool, to remove that account's posts, media, etc.
	msgs := []messages.FromClientAPI{}
	if err := p.rangeAccounts(ctx, block.Domain, func(account *gtsmodel.Account) error {
		msgs = append(msgs, messages.FromClientAPI{
			APObjectType:   ap.ActorPerson,
			APActivityType: ap.ActivityDelete,
			GTSModel:       block,
			OriginAccount:  account,
			TargetAccount:  account,
		})

		return nil
	}); err != nil {
		l.Errorf("error while ranging through accounts: %q", err)
	}

	// Batch process all accreted messages.
	p.state.Workers.EnqueueClientAPI(ctx, msgs...)
}

// domainUnblockSideEffects processes the side effects of undoing a
// domain block:
//
//  1. Mark instance entry as no longer suspended.
//  2. Mark each account from the domain as no longer suspended, if the
//     suspension origin corresponds to the ID of the provided domain block.
//
// It should be called asynchronously, since it can take a while when
// there are many accounts present on the given domain.
func (p *Processor) domainUnblockSideEffects(ctx context.Context, block *gtsmodel.DomainBlock) {
	l := log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"domain", block.Domain},
		}...)
	l.Debug("processing domain unblock side effects")

	// Update instance entry for this domain, if we have it.
	instance, err := p.state.DB.GetInstance(ctx, block.Domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		l.Errorf("db error getting instance %s: %q", block.Domain, err)
	}

	if instance != nil {
		// We had an entry, update it to signal
		// that it's no longer suspended.
		instance.SuspendedAt = time.Time{}
		instance.DomainBlockID = ""
		if err := p.state.DB.UpdateInstance(
			ctx,
			instance,
			"suspended_at",
			"domain_block_id",
		); err != nil {
			l.Errorf("db error updating instance: %s", err)
		} else {
			l.Debug("instance entry updated")
		}
	}

	// Unsuspend all accounts whose suspension origin was this domain block.
	if err := p.rangeAccounts(ctx, block.Domain, func(account *gtsmodel.Account) error {
		if account.SuspensionOrigin == "" || account.SuspendedAt.IsZero() {
			// Account wasn't suspended, nothing to do.
			return nil
		}

		if account.SuspensionOrigin != block.ID {
			// Account was suspended, but not by
			// this domain block, leave it alone.
			return nil
		}

		// Account was suspended by this domain
		// block, mark it as unsuspended.
		account.SuspendedAt = time.Time{}
		account.SuspensionOrigin = ""

		if err := p.state.DB.UpdateAccount(
			ctx,
			account,
			"suspended_at",
			"suspension_origin",
		); err != nil {
			return gtserror.Newf("db error updating account %s: %w", account.Username, err)
		}

		return nil
	}); err != nil {
		l.Errorf("error while ranging through accounts: %q", err)
	}
}
