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
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
)

// DomainKeysExpire iterates through all
// accounts belonging to the given domain,
// and expires the public key of each
// account found this way.
//
// The PublicKey for each account will be
// re-fetched next time a signed request
// from that account is received.
func (p *Processor) DomainKeysExpire(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	domain string,
) (string, gtserror.WithCode) {
	// Run admin action to process
	// side effects of key expiry.
	action := &gtsmodel.AdminAction{
		ID:             id.NewULID(),
		TargetCategory: gtsmodel.AdminActionCategoryDomain,
		TargetID:       domain,
		Type:           gtsmodel.AdminActionExpireKeys,
		AccountID:      adminAcct.ID,
	}

	if errWithCode := p.state.AdminActions.Run(
		ctx,
		action,
		p.state.AdminActions.DomainKeysExpireF(domain),
	); errWithCode != nil {
		return action.ID, errWithCode
	}

	return action.ID, nil
}
