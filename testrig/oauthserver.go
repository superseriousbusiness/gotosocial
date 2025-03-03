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

package testrig

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/oauth/handlers"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

// NewTestOauthServer returns an oauth server with the given db
func NewTestOauthServer(state *state.State) oauth.Server {
	ctx := context.Background()
	return oauth.New(
		ctx,
		state,
		handlers.GetValidateURIHandler(ctx),
		handlers.GetClientScopeHandler(ctx, state),
		handlers.GetAuthorizeScopeHandler(),
		handlers.GetInternalErrorHandler(ctx),
		handlers.GetResponseErrorHandler(ctx),
		handlers.GetUserAuthorizationHandler(),
	)
}
