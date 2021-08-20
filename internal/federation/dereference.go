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

package federation

import (
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *federator) GetRemoteAccount(username string, remoteAccountID *url.URL, refresh bool) (*gtsmodel.Account, bool, error) {
	return f.dereferencer.GetRemoteAccount(username, remoteAccountID, refresh)
}

func (f *federator) EnrichRemoteAccount(username string, account *gtsmodel.Account) (*gtsmodel.Account, error) {
	return f.dereferencer.EnrichRemoteAccount(username, account)
}

func (f *federator) GetRemoteStatus(username string, remoteStatusID *url.URL, refresh bool) (*gtsmodel.Status, ap.Statusable, bool, error) {
	return f.dereferencer.GetRemoteStatus(username, remoteStatusID, refresh)
}

func (f *federator) EnrichRemoteStatus(username string, status *gtsmodel.Status) (*gtsmodel.Status, error) {
	return f.dereferencer.EnrichRemoteStatus(username, status)
}

func (f *federator) DereferenceRemoteThread(username string, statusIRI *url.URL) error {
	return f.dereferencer.DereferenceThread(username, statusIRI)
}

func (f *federator) GetRemoteInstance(username string, remoteInstanceURI *url.URL) (*gtsmodel.Instance, error) {
	return f.dereferencer.GetRemoteInstance(username, remoteInstanceURI)
}

func (f *federator) DereferenceAnnounce(announce *gtsmodel.Status, requestingUsername string) error {
	return f.dereferencer.DereferenceAnnounce(announce, requestingUsername)
}
