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
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/media"
)

func (d *deref) GetRemoteMedia(ctx context.Context, requestingUsername string, accountID string, remoteURL string) (*media.Media, error) {
	if accountID == "" {
		return nil, fmt.Errorf("RefreshAttachment: minAttachment account ID was empty")
	}

	t, err := d.transportController.NewTransportForUsername(ctx, requestingUsername)
	if err != nil {
		return nil, fmt.Errorf("RefreshAttachment: error creating transport: %s", err)
	}

	derefURI, err := url.Parse(remoteURL)
	if err != nil {
		return nil, err
	}

	data, err := t.DereferenceMedia(ctx, derefURI)
	if err != nil {
		return nil, fmt.Errorf("RefreshAttachment: error dereferencing media: %s", err)
	}

	m, err := d.mediaManager.ProcessMedia(ctx, data, accountID, remoteURL)
	if err != nil {
		return nil, fmt.Errorf("RefreshAttachment: error processing attachment: %s", err)
	}

	return m, nil
}
