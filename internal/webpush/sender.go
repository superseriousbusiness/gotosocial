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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/httpclient"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// Sender can send Web Push notifications.
type Sender interface {

	// Send queues up a notification for delivery to all of an account's Web Push subscriptions.
	Send(ctx context.Context, notif *gtsmodel.Notification, apiNotif *apimodel.Notification) error
}

// NewSender creates a new sender from an HTTP client, DB, and worker pool.
func NewSender(httpClient *httpclient.Client, state *state.State, converter *typeutils.Converter) Sender {
	return &realSender{
		httpClient: &http.Client{
			// Pass in our wrapped httpclient.Client{}
			// type as http.Transport{} in order to take
			// advantage of retries, SSF protection etc.
			Transport: httpClient,

			// Other http.Client{} fields are already
			// set in embedded httpclient.Client{}.
		},
		state:     state,
		converter: converter,
	}
}

// an internal function purely existing for the webpush test package to link to and use a custom http.Client{}.
func newSenderWith(client *http.Client, state *state.State, converter *typeutils.Converter) Sender { //nolint:unused
	return &realSender{
		httpClient: client,
		state:      state,
		converter:  converter,
	}
}
