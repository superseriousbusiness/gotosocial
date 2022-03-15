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

package transport

import (
	"context"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (t *transport) Dereference(ctx context.Context, iri *url.URL) ([]byte, error) {
	l := logrus.WithField("func", "Dereference")

	// if the request is to us, we can shortcut for certain URIs rather than going through
	// the normal request flow, thereby saving time and energy
	if iri.Host == viper.GetString(config.Keys.Host) {
		if uris.IsFollowersPath(iri) {
			// the request is for followers of one of our accounts, which we can shortcut
			return t.dereferenceFollowersShortcut(ctx, iri)
		}

		if uris.IsUserPath(iri) {
			// the request is for one of our accounts, which we can shortcut
			return t.dereferenceUserShortcut(ctx, iri)
		}
	}

	// the request is either for a remote host or for us but we don't have a shortcut, so continue as normal
	l.Debugf("performing GET to %s", iri.String())
	return t.sigTransport.Dereference(ctx, iri)
}
