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

package cache

import "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

func copyStatus(s1 *gtsmodel.Status) *gtsmodel.Status {
	s2 := new(gtsmodel.Status)
	*s2 = *s1

	// Don't include ptr fields that
	// will be populated separately.
	// See internal/db/bundb/status.go.
	s2.Account = nil
	s2.InReplyTo = nil
	s2.InReplyToAccount = nil
	s2.BoostOf = nil
	s2.BoostOfAccount = nil
	s2.Poll = nil
	s2.Attachments = nil
	s2.Tags = nil
	s2.Mentions = nil
	s2.Emojis = nil
	s2.CreatedWithApplication = nil

	return s2
}
