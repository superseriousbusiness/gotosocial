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
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (d *deref) GetRemoteInstance(ctx context.Context, username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error) {
	if blocked, err := d.db.IsDomainBlocked(ctx, remoteInstanceURI.Host); blocked || err != nil {
		return nil, fmt.Errorf("GetRemoteInstance: domain %s is blocked", remoteInstanceURI.Host)
	}

	transport, err := d.transportController.NewTransportForUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("transport err: %s", err)
	}

	return transport.DereferenceInstance(ctx, remoteInstanceURI)
}
