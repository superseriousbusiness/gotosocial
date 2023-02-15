package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
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

func (p *AdminProcessor) AdminDomainBlockCreate(ctx context.Context, account *gtsmodel.Account, domain string, obfuscate bool, publicComment string, privateComment string, subscriptionID string) (*apimodel.DomainBlock, gtserror.WithCode) {
	// domain blocks will always be lowercase
	domain = strings.ToLower(domain)

	// first check if we already have a block -- if err == nil we already had a block so we can skip a whole lot of work
	block, err := p.db.GetDomainBlock(ctx, domain)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// something went wrong in the DB
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error checking for existence of domain block %s: %s", domain, err))
		}

		// there's no block for this domain yet so create one
		newBlock := &gtsmodel.DomainBlock{
			ID:                 id.NewULID(),
			Domain:             domain,
			CreatedByAccountID: account.ID,
			PrivateComment:     text.SanitizePlaintext(privateComment),
			PublicComment:      text.SanitizePlaintext(publicComment),
			Obfuscate:          &obfuscate,
			SubscriptionID:     subscriptionID,
		}

		// Insert the new block into the database
		if err := p.db.CreateDomainBlock(ctx, newBlock); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error putting new domain block %s: %s", domain, err))
		}

		// Set the newly created block
		block = newBlock

		// Process the side effects of the domain block asynchronously since it might take a while
		go func() {
			p.initiateDomainBlockSideEffects(context.Background(), account, block)
		}()
	}

	// Convert our gts model domain block into an API model
	apiDomainBlock, err := p.tc.DomainBlockToAPIDomainBlock(ctx, block, false)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting domain block to frontend/api representation %s: %s", domain, err))
	}

	return apiDomainBlock, nil
}

// initiateDomainBlockSideEffects should be called asynchronously, to process the side effects of a domain block:
//
// 1. Strip most info away from the instance entry for the domain.
// 2. Delete the instance account for that instance if it exists.
// 3. Select all accounts from this instance and pass them through the delete functionality of the processor.
func (p *AdminProcessor) initiateDomainBlockSideEffects(ctx context.Context, account *gtsmodel.Account, block *gtsmodel.DomainBlock) {
	l := log.WithFields(kv.Fields{{"domain", block.Domain}}...)

	l.Debug("processing domain block side effects")

	// if we have an instance entry for this domain, update it with the new block ID and clear all fields
	instance := &gtsmodel.Instance{}
	if err := p.db.GetWhere(ctx, []db.Where{{Key: "domain", Value: block.Domain}}, instance); err == nil {
		updatingColumns := []string{
			"title",
			"updated_at",
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
		if err := p.db.UpdateByID(ctx, instance, instance.ID, updatingColumns...); err != nil {
			l.Errorf("domainBlockProcessSideEffects: db error updating instance: %s", err)
		}
		l.Debug("domainBlockProcessSideEffects: instance entry updated")
	}

	// if we have an instance account for this instance, delete it
	if instanceAccount, err := p.db.GetAccountByUsernameDomain(ctx, block.Domain, block.Domain); err == nil {
		if err := p.db.DeleteAccount(ctx, instanceAccount.ID); err != nil {
			l.Errorf("domainBlockProcessSideEffects: db error deleting instance account: %s", err)
		}
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
			p.clientWorker.Queue(messages.FromClientAPI{
				APObjectType:   ap.ActorPerson,
				APActivityType: ap.ActivityDelete,
				GTSModel:       block,
				OriginAccount:  account,
				TargetAccount:  a,
			})

			// if this is the last account in the slice, set the maxID appropriately for the next query
			if i == len(accounts)-1 {
				maxID = a.ID
			}
		}
	}
}

// AdminDomainBlocksImport handles the import of a bunch of domain blocks at once, by calling the AdminDomainBlockCreate function for each domain in the provided file.
func (p *AdminProcessor) AdminDomainBlocksImport(ctx context.Context, account *gtsmodel.Account, domains *multipart.FileHeader) ([]*apimodel.DomainBlock, gtserror.WithCode) {
	f, err := domains.Open()
	if err != nil {
		return nil, gtserror.NewErrorBadRequest(fmt.Errorf("DomainBlocksImport: error opening attachment: %s", err))
	}
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		return nil, gtserror.NewErrorBadRequest(fmt.Errorf("DomainBlocksImport: error reading attachment: %s", err))
	}
	if size == 0 {
		return nil, gtserror.NewErrorBadRequest(errors.New("DomainBlocksImport: could not read provided attachment: size 0 bytes"))
	}

	d := []apimodel.DomainBlock{}
	if err := json.Unmarshal(buf.Bytes(), &d); err != nil {
		return nil, gtserror.NewErrorBadRequest(fmt.Errorf("DomainBlocksImport: could not read provided attachment: %s", err))
	}

	blocks := []*apimodel.DomainBlock{}
	for _, d := range d {
		block, err := p.AdminDomainBlockCreate(ctx, account, d.Domain.Domain, false, d.PublicComment, "", "")
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (p *AdminProcessor) AdminDomainBlocksGet(ctx context.Context, account *gtsmodel.Account, export bool) ([]*apimodel.DomainBlock, gtserror.WithCode) {
	domainBlocks := []*gtsmodel.DomainBlock{}

	if err := p.db.GetAll(ctx, &domainBlocks); err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// something has gone really wrong
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	apiDomainBlocks := []*apimodel.DomainBlock{}
	for _, b := range domainBlocks {
		apiDomainBlock, err := p.tc.DomainBlockToAPIDomainBlock(ctx, b, export)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		apiDomainBlocks = append(apiDomainBlocks, apiDomainBlock)
	}

	return apiDomainBlocks, nil
}

func (p *AdminProcessor) AdminDomainBlockGet(ctx context.Context, account *gtsmodel.Account, id string, export bool) (*apimodel.DomainBlock, gtserror.WithCode) {
	domainBlock := &gtsmodel.DomainBlock{}

	if err := p.db.GetByID(ctx, id, domainBlock); err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// something has gone really wrong
			return nil, gtserror.NewErrorInternalError(err)
		}
		// there are no entries for this ID
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("no entry for ID %s", id))
	}

	apiDomainBlock, err := p.tc.DomainBlockToAPIDomainBlock(ctx, domainBlock, export)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiDomainBlock, nil
}

func (p *AdminProcessor) AdminDomainBlockDelete(ctx context.Context, account *gtsmodel.Account, id string) (*apimodel.DomainBlock, gtserror.WithCode) {
	domainBlock := &gtsmodel.DomainBlock{}

	if err := p.db.GetByID(ctx, id, domainBlock); err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// something has gone really wrong
			return nil, gtserror.NewErrorInternalError(err)
		}
		// there are no entries for this ID
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("no entry for ID %s", id))
	}

	// prepare the domain block to return
	apiDomainBlock, err := p.tc.DomainBlockToAPIDomainBlock(ctx, domainBlock, false)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Delete the domain block
	if err := p.db.DeleteDomainBlock(ctx, domainBlock.Domain); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// remove the domain block reference from the instance, if we have an entry for it
	i := &gtsmodel.Instance{}
	if err := p.db.GetWhere(ctx, []db.Where{
		{Key: "domain", Value: domainBlock.Domain},
		{Key: "domain_block_id", Value: id},
	}, i); err == nil {
		updatingColumns := []string{"suspended_at", "domain_block_id", "updated_at"}
		i.SuspendedAt = time.Time{}
		i.DomainBlockID = ""
		i.UpdatedAt = time.Now()
		if err := p.db.UpdateByID(ctx, i, i.ID, updatingColumns...); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("couldn't update database entry for instance %s: %s", domainBlock.Domain, err))
		}
	}

	// unsuspend all accounts whose suspension origin was this domain block
	// 1. remove the 'suspended_at' entry from their accounts
	if err := p.db.UpdateWhere(ctx, []db.Where{
		{Key: "suspension_origin", Value: domainBlock.ID},
	}, "suspended_at", nil, &[]*gtsmodel.Account{}); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("database error removing suspended_at from accounts: %s", err))
	}

	// 2. remove the 'suspension_origin' entry from their accounts
	if err := p.db.UpdateWhere(ctx, []db.Where{
		{Key: "suspension_origin", Value: domainBlock.ID},
	}, "suspension_origin", nil, &[]*gtsmodel.Account{}); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("database error removing suspension_origin from accounts: %s", err))
	}

	return apiDomainBlock, nil
}
