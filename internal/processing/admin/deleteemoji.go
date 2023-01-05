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

package admin

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (p *processor) EmojiDelete(ctx context.Context, id string) (*apimodel.AdminEmoji, gtserror.WithCode) {
	emoji, err := p.db.GetEmojiByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("EmojiDelete: no emoji with id %s found in the db", id)
			return nil, gtserror.NewErrorNotFound(err)
		}
		err := fmt.Errorf("EmojiDelete: db error: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if emoji.Domain != "" {
		err = fmt.Errorf("EmojiDelete: emoji with id %s was not a local emoji, will not delete", id)
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	adminEmoji, err := p.tc.EmojiToAdminAPIEmoji(ctx, emoji)
	if err != nil {
		err = fmt.Errorf("EmojiDelete: error converting emoji to admin api emoji: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.db.DeleteEmojiByID(ctx, id); err != nil {
		err := fmt.Errorf("EmojiDelete: db error: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return adminEmoji, nil
}
