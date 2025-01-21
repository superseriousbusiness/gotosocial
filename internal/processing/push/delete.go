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

package push

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// Delete deletes the Web Push subscription for the given access token, if there is one.
func (p *Processor) Delete(ctx context.Context, accessToken string) gtserror.WithCode {
	tokenID, errWithCode := p.getTokenID(ctx, accessToken)
	if errWithCode != nil {
		return errWithCode
	}

	if err := p.state.DB.DeleteWebPushSubscriptionByTokenID(ctx, tokenID); err != nil {
		err := gtserror.Newf("couldn't delete Web Push subscription for token ID %s: %w", tokenID, err)
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}
