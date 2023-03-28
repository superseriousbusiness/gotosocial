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

package status

import (
	"context"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *Processor) apiStatus(ctx context.Context, targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account) (*apimodel.Status, gtserror.WithCode) {
	apiStatus, err := p.tc.StatusToAPIStatus(ctx, targetStatus, requestingAccount)
	if err != nil {
		err = fmt.Errorf("error converting status %s to frontend representation: %w", targetStatus.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiStatus, nil
}

func (p *Processor) getVisibleStatus(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*gtsmodel.Status, gtserror.WithCode) {
	targetStatus, err := p.state.DB.GetStatusByID(ctx, targetStatusID)
	if err != nil {
		err = fmt.Errorf("getVisibleStatus: db error fetching status %s: %w", targetStatusID, err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	visible, err := p.filter.StatusVisible(ctx, requestingAccount, targetStatus)
	if err != nil {
		err = fmt.Errorf("getVisibleStatus: error seeing if status %s is visible: %w", targetStatus.ID, err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if !visible {
		err = fmt.Errorf("getVisibleStatus: status %s is not visible to requesting account", targetStatusID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	return targetStatus, nil
}
