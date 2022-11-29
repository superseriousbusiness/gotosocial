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
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *federator) GetAccount(ctx context.Context, params dereferencing.GetAccountParams) (*gtsmodel.Account, error) {
	return f.dereferencer.GetAccount(ctx, params)
}

func (f *federator) GetStatus(ctx context.Context, username string, remoteStatusID *url.URL, refetch, includeParent bool) (*gtsmodel.Status, ap.Statusable, error) {
	return f.dereferencer.GetStatus(ctx, username, remoteStatusID, refetch, includeParent)
}

func (f *federator) EnrichRemoteStatus(ctx context.Context, username string, status *gtsmodel.Status, includeParent bool) (*gtsmodel.Status, error) {
	return f.dereferencer.EnrichRemoteStatus(ctx, username, status, includeParent)
}

func (f *federator) DereferenceRemoteThread(ctx context.Context, username string, statusIRI *url.URL, status *gtsmodel.Status, statusable ap.Statusable) {
	f.dereferencer.DereferenceThread(ctx, username, statusIRI, status, statusable)
}

func (f *federator) GetRemoteInstance(ctx context.Context, username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error) {
	return f.dereferencer.GetRemoteInstance(ctx, username, remoteInstanceURI)
}

func (f *federator) DereferenceAnnounce(ctx context.Context, announce *gtsmodel.Status, requestingUsername string) error {
	return f.dereferencer.DereferenceAnnounce(ctx, announce, requestingUsername)
}
