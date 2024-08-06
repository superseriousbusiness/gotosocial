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

package gtsmodel

// PendingInteraction represents one interaction
// (status, status fave) with approval set to pending.
//
// **This is not a database model and should not be
// stored in the database**, it's just for passing
// around internally, and doing fancy UNION selects.
type PendingInteraction struct {
	ID              string
	InteractionType InteractionType
	Like            *StatusFave // Only set if InteractionType = InteractionLike.
	Reply           *Status     // Only set if InteractionType = InteractionReply.
	Announce        *Status     // Only set if InteractionType = InteractionAnnounce.
}
