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

// noopSender drops anything sent to it.
// This should only be used in tests.
type noopSender struct{}

func NewNoopSender() Sender {
	return &noopSender{}
}

func (n *noopSender) Send(
	ctx context.Context,
	notification *gtsmodel.Notification,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) error {
	return nil
}
