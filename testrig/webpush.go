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

	"code.superseriousbusiness.org/gotosocial/internal/filter/usermute"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/webpush"
)

// WebPushMockSender collects a map of notifications sent to each account ID.
type WebPushMockSender struct {
	Sent map[string][]*gtsmodel.Notification
}

// NewWebPushMockSender creates a mock sender that can record sent Web Push notifications for test expectations.
func NewWebPushMockSender() *WebPushMockSender {
	return &WebPushMockSender{
		Sent: map[string][]*gtsmodel.Notification{},
	}
}

func (m *WebPushMockSender) Send(
	ctx context.Context,
	notification *gtsmodel.Notification,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) error {
	m.Sent[notification.TargetAccountID] = append(m.Sent[notification.TargetAccountID], notification)
	return nil
}

// noopSender drops anything sent to it.
type noopWebPushSender struct{}

// NewNoopWebPushSender creates a no-op sender that does nothing.
func NewNoopWebPushSender() webpush.Sender {
	return &noopWebPushSender{}
}

func (n *noopWebPushSender) Send(
	ctx context.Context,
	notification *gtsmodel.Notification,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) error {
	return nil
}
