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

package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

func (p *processor) DomainBlockCreate(ctx context.Context, account *gtsmodel.Account, domain string, obfuscate bool, publicComment string, privateComment string, subscriptionID string) (*apimodel.DomainBlock, gtserror.WithCode) {
	// first check if we already have a block -- if err == nil we already had a block so we can skip a whole lot of work
	domainBlock := &gtsmodel.DomainBlock{}
	err := p.db.GetWhere(ctx, []db.Where{{Key: "domain", Value: domain, CaseInsensitive: true}}, domainBlock)
	if err != nil {
		if err != db.ErrNoEntries {
			// something went wrong in the DB
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("DomainBlockCreate: db error checking for existence of domain block %s: %s", domain, err))
		}

		// there's no block for this domain yet so create one
		// note: we take a new ulid from timestamp here in case we need to sort blocks
		blockID, err := id.NewULID()
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("DomainBlockCreate: error creating id for new domain block %s: %s", domain, err))
		}

		domainBlock = &gtsmodel.DomainBlock{
			ID:                 blockID,
			Domain:             domain,
			CreatedByAccountID: account.ID,
			PrivateComment:     text.RemoveHTML(privateComment),
			PublicComment:      text.RemoveHTML(publicComment),
			Obfuscate:          obfuscate,
			SubscriptionID:     subscriptionID,
		}

		// put the new block in the database
		if err := p.db.Put(ctx, domainBlock); err != nil {
			if err != db.ErrNoEntries {
				// there's a real error creating the block
				return nil, gtserror.NewErrorInternalError(fmt.Errorf("DomainBlockCreate: db error putting new domain block %s: %s", domain, err))
			}
		}

		// process the side effects of the domain block asynchronously since it might take a while
		go p.initiateDomainBlockSideEffects(ctx, account, domainBlock) // TODO: add this to a queuing system so it can retry/resume
	}

	apiDomainBlock, err := p.tc.DomainBlockToAPIDomainBlock(ctx, domainBlock, false)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("DomainBlockCreate: error converting domain block to frontend/api representation %s: %s", domain, err))
	}

	return apiDomainBlock, nil
}

// initiateDomainBlockSideEffects should be called asynchronously, to process the side effects of a domain block:
//
// 1. Strip most info away from the instance entry for the domain.
// 2. Delete the instance account for that instance if it exists.
// 3. Select all accounts from this instance and pass them through the delete functionality of the processor.
func (p *processor) initiateDomainBlockSideEffects(ctx context.Context, account *gtsmodel.Account, block *gtsmodel.DomainBlock) {
	l := logrus.WithFields(logrus.Fields{
		"func":   "domainBlockProcessSideEffects",
		"domain": block.Domain,
	})

	l.Debug("processing domain block side effects")

	// if we have an instance entry for this domain, update it with the new block ID and clear all fields
	instance := &gtsmodel.Instance{}
	if err := p.db.GetWhere(ctx, []db.Where{{Key: "domain", Value: block.Domain, CaseInsensitive: true}}, instance); err == nil {
		instance.Title = ""
		instance.UpdatedAt = time.Now()
		instance.SuspendedAt = time.Now()
		instance.DomainBlockID = block.ID
		instance.ShortDescription = ""
		instance.Description = ""
		instance.Terms = ""
		instance.ContactEmail = ""
		instance.ContactAccountUsername = ""
		instance.ContactAccountID = ""
		instance.Version = ""
		if err := p.db.UpdateByPrimaryKey(ctx, instance); err != nil {
			l.Errorf("domainBlockProcessSideEffects: db error updating instance: %s", err)
		}
		l.Debug("domainBlockProcessSideEffects: instance entry updated")
	}

	// if we have an instance account for this instance, delete it
	if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "username", Value: block.Domain, CaseInsensitive: true}}, &gtsmodel.Account{}); err != nil {
		l.Errorf("domainBlockProcessSideEffects: db error removing instance account: %s", err)
	}

	// delete accounts through the normal account deletion system (which should also delete media + posts + remove posts from timelines)

	limit := 20      // just select 20 accounts at a time so we don't nuke our DB/mem with one huge query
	var maxID string // this is initially an empty string so we'll start at the top of accounts list (sorted by ID)

selectAccountsLoop:
	for {
		accounts, err := p.db.GetInstanceAccounts(ctx, block.Domain, maxID, limit)
		if err != nil {
			if err == db.ErrNoEntries {
				// no accounts left for this instance so we're done
				l.Infof("domainBlockProcessSideEffects: done iterating through accounts for domain %s", block.Domain)
				break selectAccountsLoop
			}
			// an actual error has occurred
			l.Errorf("domainBlockProcessSideEffects: db error selecting accounts for domain %s: %s", block.Domain, err)
			break selectAccountsLoop
		}

		for i, a := range accounts {
			l.Debugf("putting delete for account %s in the clientAPI channel", a.Username)

			// pass the account delete through the client api channel for processing
			p.fromClientAPI <- messages.FromClientAPI{
				APObjectType:   ap.ActorPerson,
				APActivityType: ap.ActivityDelete,
				GTSModel:       block,
				OriginAccount:  account,
				TargetAccount:  a,
			}

			// if this is the last account in the slice, set the maxID appropriately for the next query
			if i == len(accounts)-1 {
				maxID = a.ID
			}
		}
	}
}
