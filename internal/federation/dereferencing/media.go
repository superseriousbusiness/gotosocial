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

package dereferencing

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/media"
)

func (d *deref) GetRemoteMedia(ctx context.Context, requestingUsername string, accountID string, remoteURL string, ai *media.AdditionalMediaInfo) (*media.ProcessingMedia, error) {
	if accountID == "" {
		return nil, fmt.Errorf("GetRemoteMedia: account ID was empty")
	}

	t, err := d.transportController.NewTransportForUsername(ctx, requestingUsername)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteMedia: error creating transport: %s", err)
	}

	derefURI, err := url.Parse(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteMedia: error parsing url: %s", err)
	}

	dataFunc := func(innerCtx context.Context) (io.ReadCloser, int64, error) {
		return t.DereferenceMedia(innerCtx, derefURI)
	}

	processingMedia, err := d.mediaManager.ProcessMedia(ctx, dataFunc, nil, accountID, ai)
	if err != nil {
		return nil, fmt.Errorf("GetRemoteMedia: error processing attachment: %s", err)
	}

	return processingMedia, nil
}
