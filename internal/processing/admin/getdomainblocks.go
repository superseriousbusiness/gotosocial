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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) DomainBlocksGet(ctx context.Context, account *gtsmodel.Account, export bool) ([]*apimodel.DomainBlock, gtserror.WithCode) {
	domainBlocks := []*gtsmodel.DomainBlock{}

	if err := p.db.GetAll(ctx, &domainBlocks); err != nil {
		if err != db.ErrNoEntries {
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
