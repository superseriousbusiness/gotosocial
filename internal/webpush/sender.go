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

package webpush

import (
	"context"
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

// Sender can send Web Push notifications.
type Sender interface {
	// Send queues up a notification for delivery to all of an account's Web Push subscriptions.
	Send(
		ctx context.Context,
		notification *gtsmodel.Notification,
		filters []*gtsmodel.Filter,
		mutes *usermute.CompiledUserMuteList,
	) error
}

// NewSender creates a new sender from an HTTP client, DB, and worker pool.
func NewSender(httpClient *httpclient.Client, state *state.State) Sender {
	return NewRealSender(
		&http.Client{
			Transport: &gtsHttpClientRoundTripper{
				httpClient: httpClient,
			},
			// Other fields are already set on the http.Client inside the httpclient.Client.
		},
		state,
	)
}
