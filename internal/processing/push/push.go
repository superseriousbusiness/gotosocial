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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

type Processor struct {
	state     *state.State
	converter *typeutils.Converter
}

func New(state *state.State, converter *typeutils.Converter) Processor {
	return Processor{
		state:     state,
		converter: converter,
	}
}

// getTokenID returns the token ID for a given access token.
// Since all push API calls require authentication, this should always be available.
func (p *Processor) getTokenID(ctx context.Context, accessToken string) (string, gtserror.WithCode) {
	token, err := p.state.DB.GetTokenByAccess(ctx, accessToken)
	if err != nil {
		return "", gtserror.NewErrorInternalError(
			gtserror.Newf("couldn't find token ID for access token: %w", err),
		)
	}

	return token.ID, nil
}

// apiSubscription is a shortcut to return the API version of the given Web Push subscription,
// or return an appropriate error if conversion fails.
func (p *Processor) apiSubscription(ctx context.Context, subscription *gtsmodel.WebPushSubscription) (*apimodel.WebPushSubscription, gtserror.WithCode) {
	apiSubscription, err := p.converter.WebPushSubscriptionToAPIWebPushSubscription(ctx, subscription)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(
			gtserror.Newf("error converting Web Push subscription %s to API representation: %w", subscription.ID, err),
		)
	}

	return apiSubscription, nil
}
