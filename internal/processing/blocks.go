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

package processing

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// BlocksGet ...
func (p *Processor) BlocksGet(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	maxID string,
	minID string,
	limit int,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	blocks, err := p.state.DB.GetAccountBlocks(ctx,
		requestingAccount.ID,
		&paging.Pager{
			MinID: minID,
			MaxID: maxID,
			Limit: limit,
		},
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for zero length.
	count := len(blocks)
	if len(blocks) == 0 {
		return util.EmptyPageableResponse(), nil
	}

	var (
		items = make([]interface{}, 0, count)

		// Set next + prev values before API converting
		// so the caller can still page even on error.
		nextMaxIDValue = blocks[count-1].ID
		prevMinIDValue = blocks[0].ID
	)

	for _, block := range blocks {
		if block.TargetAccount == nil {
			// All models should be populated at this point.
			log.Warnf(ctx, "block target account was nil: %v", err)
			continue
		}

		// Convert target account to frontend API model.
		account, err := p.tc.AccountToAPIAccountBlocked(ctx, block.TargetAccount)
		if err != nil {
			log.Errorf(ctx, "error converting account to public api account: %v", err)
			continue
		}

		// Append target to return items.
		items = append(items, account)
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "/api/v1/blocks",
		NextMaxIDKey:   "max_id",
		PrevMinIDKey:   "since_id",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}
