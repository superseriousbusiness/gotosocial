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

package dereferencing

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/media"
)

func (d *deref) GetRemoteEmoji(ctx context.Context, requestingUsername string, remoteURL string, shortcode string, id string, emojiURI string, ai *media.AdditionalEmojiInfo) (*media.ProcessingEmoji, error) {
	t, err := d.transportController.NewTransportForUsername(ctx, requestingUsername)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteEmoji: error creating transport: %s", err)
	}

	derefURI, err := url.Parse(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteEmoji: error parsing url: %s", err)
	}

	dataFunc := func(innerCtx context.Context) (io.Reader, int64, error) {
		return t.DereferenceMedia(innerCtx, derefURI)
	}

	processingMedia, err := d.mediaManager.ProcessEmoji(ctx, dataFunc, nil, shortcode, id, emojiURI, ai)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteEmoji: error processing emoji: %s", err)
	}

	return processingMedia, nil
}
