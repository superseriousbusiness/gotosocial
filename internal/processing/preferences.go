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

package processing

import (
	"context"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

func (p *Processor) PreferencesGet(ctx context.Context, accountID string) (*apimodel.Preferences, gtserror.WithCode) {
	act, err := p.state.DB.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return &apimodel.Preferences{
		PostingDefaultVisibility: mastoPrefVisibility(act.Settings.Privacy),
		PostingDefaultSensitive:  *act.Settings.Sensitive,
		PostingDefaultLanguage:   act.Settings.Language,
		// The Reading* preferences don't appear to actually be settable by the
		// client, so forcing some sensible defaults here
		ReadingExpandMedia:    "default",
		ReadingExpandSpoilers: false,
		ReadingAutoPlayGifs:   false,
	}, nil
}

func mastoPrefVisibility(vis gtsmodel.Visibility) string {
	switch vis {
	case gtsmodel.VisibilityPublic, gtsmodel.VisibilityDirect:
		return vis.String()
	case gtsmodel.VisibilityUnlocked:
		return "unlisted"
	default:
		// this will catch gtsmodel.VisibilityMutualsOnly and other types Mastodon doesn't
		// have and map them to the most restrictive state
		return "private"
	}
}
