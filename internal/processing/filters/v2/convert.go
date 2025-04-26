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

package v2

import (
	"context"
	"fmt"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// apiFilter is a shortcut to return the API v2 filter version of the given
// filter, or return an appropriate error if conversion fails.
func (p *Processor) apiFilter(ctx context.Context, filterKeyword *gtsmodel.Filter) (*apimodel.FilterV2, gtserror.WithCode) {
	apiFilter, err := p.converter.FilterToAPIFilterV2(ctx, filterKeyword)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting filter to API v2 filter: %w", err))
	}

	return apiFilter, nil
}
