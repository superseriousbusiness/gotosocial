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

package admin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func (p *processor) EmojiCreate(ctx context.Context, account *gtsmodel.Account, user *gtsmodel.User, form *apimodel.EmojiCreateRequest) (*apimodel.Emoji, error) {
	if user.Admin {
		return nil, fmt.Errorf("user %s not an admin", user.ID)
	}

	// open the emoji and extract the bytes from it
	f, err := form.Image.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening emoji: %s", err)
	}
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		return nil, fmt.Errorf("error reading emoji: %s", err)
	}
	if size == 0 {
		return nil, errors.New("could not read provided emoji: size 0 bytes")
	}

	// allow the mediaHandler to work its magic of processing the emoji bytes, and putting them in whatever storage backend we're using
	emoji, err := p.mediaHandler.ProcessLocalEmoji(ctx, buf.Bytes(), form.Shortcode)
	if err != nil {
		return nil, fmt.Errorf("error reading emoji: %s", err)
	}

	emojiID, err := id.NewULID()
	if err != nil {
		return nil, err
	}
	emoji.ID = emojiID

	apiEmoji, err := p.tc.EmojiToAPIEmoji(ctx, emoji)
	if err != nil {
		return nil, fmt.Errorf("error converting emoji to apitype: %s", err)
	}

	if err := p.db.Put(ctx, emoji); err != nil {
		return nil, fmt.Errorf("database error while processing emoji: %s", err)
	}

	return &apiEmoji, nil
}
