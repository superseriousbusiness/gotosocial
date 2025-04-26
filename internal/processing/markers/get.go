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

package markers

import (
	"context"
	"errors"
	"fmt"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// Get returns an API model for the markers of the requested timelines.
// If a timeline marker hasn't been set yet, it's not included in the response.
func (p *Processor) Get(ctx context.Context, account *gtsmodel.Account, names []apimodel.MarkerName) (*apimodel.Marker, gtserror.WithCode) {
	markers := make([]*gtsmodel.Marker, 0, len(names))
	for _, name := range names {
		marker, err := p.state.DB.GetMarker(ctx, account.ID, typeutils.APIMarkerNameToMarkerName(name))
		if err != nil {
			if errors.Is(err, db.ErrNoEntries) {
				continue
			}
			// Real database error.
			return nil, gtserror.NewErrorInternalError(err)
		}
		markers = append(markers, marker)
	}

	apiMarker, err := p.converter.MarkersToAPIMarker(ctx, markers)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting marker to api: %w", err))
	}

	return apiMarker, nil
}
