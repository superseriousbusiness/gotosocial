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

package federation

import (
	"context"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *federator) GetRemoteAccount(ctx context.Context, username string, remoteAccountID *url.URL, blocking bool, refresh bool) (*gtsmodel.Account, error) {
	return f.dereferencer.GetRemoteAccount(ctx, username, remoteAccountID, blocking, refresh)
}

func (f *federator) GetRemoteStatus(ctx context.Context, username string, remoteStatusID *url.URL, refresh, includeParent bool) (*gtsmodel.Status, ap.Statusable, bool, error) {
	return f.dereferencer.GetRemoteStatus(ctx, username, remoteStatusID, refresh, includeParent)
}

func (f *federator) EnrichRemoteStatus(ctx context.Context, username string, status *gtsmodel.Status, includeParent bool) (*gtsmodel.Status, error) {
	return f.dereferencer.EnrichRemoteStatus(ctx, username, status, includeParent)
}

func (f *federator) DereferenceRemoteThread(ctx context.Context, username string, statusIRI *url.URL) error {
	return f.dereferencer.DereferenceThread(ctx, username, statusIRI)
}

func (f *federator) GetRemoteInstance(ctx context.Context, username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error) {
	return f.dereferencer.GetRemoteInstance(ctx, username, remoteInstanceURI)
}

func (f *federator) DereferenceAnnounce(ctx context.Context, announce *gtsmodel.Status, requestingUsername string) error {
	return f.dereferencer.DereferenceAnnounce(ctx, announce, requestingUsername)
}
