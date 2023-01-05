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

package admin

import (
	"context"
	"fmt"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) DomainBlockDelete(ctx context.Context, account *gtsmodel.Account, id string) (*apimodel.DomainBlock, gtserror.WithCode) {
	domainBlock := &gtsmodel.DomainBlock{}

	if err := p.db.GetByID(ctx, id, domainBlock); err != nil {
		if err != db.ErrNoEntries {
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
