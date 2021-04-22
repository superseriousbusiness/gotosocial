/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package util

import (
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	mastotypes "github.com/superseriousbusiness/gotosocial/internal/mastotypes/mastomodel"
)

// ParseGTSVisFromMastoVis converts a mastodon visibility into its gts equivalent.
func ParseGTSVisFromMastoVis(m mastotypes.Visibility) gtsmodel.Visibility {
	switch m {
	case mastotypes.VisibilityPublic:
		return gtsmodel.VisibilityPublic
	case mastotypes.VisibilityUnlisted:
		return gtsmodel.VisibilityUnlocked
	case mastotypes.VisibilityPrivate:
		return gtsmodel.VisibilityFollowersOnly
	case mastotypes.VisibilityDirect:
		return gtsmodel.VisibilityDirect
	}
	return ""
}

// ParseMastoVisFromGTSVis converts a gts visibility into its mastodon equivalent
func ParseMastoVisFromGTSVis(m gtsmodel.Visibility) mastotypes.Visibility {
	switch m {
	case gtsmodel.VisibilityPublic:
		return mastotypes.VisibilityPublic
	case gtsmodel.VisibilityUnlocked:
		return mastotypes.VisibilityUnlisted
	case gtsmodel.VisibilityFollowersOnly, gtsmodel.VisibilityMutualsOnly:
		return mastotypes.VisibilityPrivate
	case gtsmodel.VisibilityDirect:
		return mastotypes.VisibilityDirect
	}
	return ""
}
