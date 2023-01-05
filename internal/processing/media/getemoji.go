/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package media

import (
	"context"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (p *processor) GetCustomEmojis(ctx context.Context) ([]*apimodel.Emoji, gtserror.WithCode) {
	emojis, err := p.db.GetUseableEmojis(ctx)
	if err != nil {
		if err != db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("db error retrieving custom emojis: %s", err))
		}
	}

	apiEmojis := make([]*apimodel.Emoji, 0, len(emojis))
	for _, gtsEmoji := range emojis {
		apiEmoji, err := p.tc.EmojiToAPIEmoji(ctx, gtsEmoji)
		if err != nil {
			log.Errorf("error converting emoji with id %s: %s", gtsEmoji.ID, err)
			continue
		}
		apiEmojis = append(apiEmojis, &apiEmoji)
	}

	return apiEmojis, nil
}
