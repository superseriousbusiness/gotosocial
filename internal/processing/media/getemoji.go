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

package media

import (
	"context"
	"fmt"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

// GetCustomEmojis returns a list of all useable local custom emojis stored on this instance.
// 'useable' in this context means visible and picker, and not disabled.
func (p *Processor) GetCustomEmojis(ctx context.Context) ([]*apimodel.Emoji, gtserror.WithCode) {
	emojis, err := p.state.DB.GetUseableEmojis(ctx)
	if err != nil {
		if err != db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("db error retrieving custom emojis: %s", err))
		}
	}

	apiEmojis := make([]*apimodel.Emoji, 0, len(emojis))
	for _, gtsEmoji := range emojis {
		apiEmoji, err := p.converter.EmojiToAPIEmoji(ctx, gtsEmoji)
		if err != nil {
			log.Errorf(ctx, "error converting emoji with id %s: %s", gtsEmoji.ID, err)
			continue
		}
		apiEmojis = append(apiEmojis, &apiEmoji)
	}

	return apiEmojis, nil
}
