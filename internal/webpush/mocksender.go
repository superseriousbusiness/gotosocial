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

	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// MockSender collects a map of notifications sent to each account ID.
// This should only be used in tests.
type MockSender struct {
	Sent map[string][]*gtsmodel.Notification
}

func NewMockSender() *MockSender {
	return &MockSender{
		Sent: map[string][]*gtsmodel.Notification{},
	}
}

func (m *MockSender) Send(
	ctx context.Context,
	notification *gtsmodel.Notification,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) error {
	m.Sent[notification.TargetAccountID] = append(m.Sent[notification.TargetAccountID], notification)
	return nil
}
