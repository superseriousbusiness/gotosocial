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
	"code.superseriousbusiness.org/gotosocial/internal/cleaner"
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/filter/interaction"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/subscriptions"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/webpush"
)

// NewTestProcessor returns a Processor suitable for testing purposes.
// The passed in state will have its worker functions set appropriately,
// but the state will not be initialized.
func NewTestProcessor(
	state *state.State,
	federator *federation.Federator,
	emailSender email.Sender,
	webPushSender webpush.Sender,
	mediaManager *media.Manager,
) *processing.Processor {

	return processing.NewProcessor(
		cleaner.New(state),
		subscriptions.New(
			state,
			federator.TransportController(),
			typeutils.NewConverter(state),
		),
		typeutils.NewConverter(state),
		federator,
		NewTestOauthServer(state),
		mediaManager,
		state,
		emailSender,
		webPushSender,
		visibility.NewFilter(state),
		interaction.NewFilter(state),
	)
}
