/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package admin

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) EmojisGet(ctx context.Context, account *gtsmodel.Account, user *gtsmodel.User, domain string, includeDisabled bool, includeEnabled bool, shortcode string, maxShortcodeDomain string, minShortcodeDomain string, limit int) ([]*apimodel.AdminEmoji, gtserror.WithCode) {
	if !*user.Admin {
		return nil, gtserror.NewErrorUnauthorized(fmt.Errorf("user %s not an admin", user.ID), "user is not an admin")
	}

	emojis, err := p.db.GetEmojis(ctx, domain, includeDisabled, includeEnabled, shortcode, maxShortcodeDomain, minShortcodeDomain, limit)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := fmt.Errorf("EmojisGet: db error: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	adminEmojis := []*apimodel.AdminEmoji{}
	for _, emoji := range emojis {
		adminEmoji, err := p.tc.EmojiToAdminAPIEmoji(ctx, emoji)
		if err != nil {
			err := fmt.Errorf("EmojisGet: error converting emoji to admin model emoji: %s", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		adminEmojis = append(adminEmojis, adminEmoji)
	}

	return []*apimodel.AdminEmoji{}, nil
}
