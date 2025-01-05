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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// apiDomainPerm is a cheeky shortcut for returning
// the API version of the given domain permission,
// or an appropriate error if something goes wrong.
func (p *Processor) apiDomainPerm(
	ctx context.Context,
	domainPermission gtsmodel.DomainPermission,
	export bool,
) (*apimodel.DomainPermission, gtserror.WithCode) {
	apiDomainPerm, err := p.converter.DomainPermToAPIDomainPerm(ctx, domainPermission, export)
	if err != nil {
		err := gtserror.NewfAt(3, "error converting domain permission to api model: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiDomainPerm, nil
}

// apiDomainPermSub is a cheeky shortcut for returning the
// API version of the given domain permission subscription,
// or an appropriate error if something goes wrong.
func (p *Processor) apiDomainPermSub(
	ctx context.Context,
	domainPermSub *gtsmodel.DomainPermissionSubscription,
) (*apimodel.DomainPermissionSubscription, gtserror.WithCode) {
	apiDomainPermSub, err := p.converter.DomainPermSubToAPIDomainPermSub(ctx, domainPermSub)
	if err != nil {
		err := gtserror.NewfAt(3, "error converting domain permission subscription to api model: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiDomainPermSub, nil
}
