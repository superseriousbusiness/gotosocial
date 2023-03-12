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

package trans

import "time"

// Follow represents an account follow as serialized in an export file.
type Follow struct {
	Type            Type       `json:"type" bun:"-"`
	ID              string     `json:"id" bun:",nullzero"`
	CreatedAt       *time.Time `json:"createdAt" bun:",nullzero"`
	URI             string     `json:"uri" bun:",nullzero"`
	AccountID       string     `json:"accountId" bun:",nullzero"`
	TargetAccountID string     `json:"targetAccountId" bun:",nullzero"`
}
