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

func (p *Processor) createDomainAllow(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	domain string,
	obfuscate bool,
	publicComment string,
	privateComment string,
	subscriptionID string,
) (*apimodel.DomainPermission, string, gtserror.WithCode) {
	// TODO
	return nil, "", nil
}

func (p *Processor) domainAllowSideEffects(
	ctx context.Context,
	allow *gtsmodel.DomainAllow,
) gtserror.MultiError {
	// TODO
	return nil
}

func (p *Processor) deleteDomainAllow(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	domainAllowID string,
) (*apimodel.DomainPermission, string, gtserror.WithCode) {
	// TODO
	return nil, "", nil
}

func (p *Processor) domainUnallowSideEffects(
	ctx context.Context,
	allow *gtsmodel.DomainAllow,
) gtserror.MultiError {
	// TODO
	return nil
}
