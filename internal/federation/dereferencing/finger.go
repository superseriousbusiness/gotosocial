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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

func (d *deref) fingerRemoteAccount(ctx context.Context, username string, targetUsername string, targetHost string) (accountDomain string, accountURI *url.URL, err error) {
	t, err := d.transportController.NewTransportForUsername(ctx, username)
	if err != nil {
		err = fmt.Errorf("fingerRemoteAccount: error getting transport for %s: %s", username, err)
		return
	}

	b, err := t.Finger(ctx, targetUsername, targetHost)
	if err != nil {
		err = fmt.Errorf("fingerRemoteAccount: error fingering @%s@%s: %s", targetUsername, targetHost, err)
		return
	}

	resp := &apimodel.WellKnownResponse{}
	if err = json.Unmarshal(b, resp); err != nil {
		err = fmt.Errorf("fingerRemoteAccount: could not unmarshal server response as WebfingerAccountResponse while dereferencing @%s@%s: %s", targetUsername, targetHost, err)
		return
	}

	if len(resp.Links) == 0 {
		err = fmt.Errorf("fingerRemoteAccount: no links found in webfinger response %s", string(b))
		return
	}

	if resp.Subject == "" {
		err = fmt.Errorf("fingerRemoteAccount: no subject found in webfinger response %s", string(b))
		return
	}

	aaaaaaaaaaaaaaaaccountDomain

	// look through the links for the first one that matches "application/activity+json", this is what we need
	for _, l := range resp.Links {
		if strings.EqualFold(l.Type, "application/activity+json") && l.Href != "" && l.Rel == "self" {
			accountURI, err = url.Parse(l.Href)
			if err != nil {
				err = fmt.Errorf("fingerRemoteAccount: couldn't parse url %s: %s", l.Href, err)
				return
			}
			// found it!
			return
		}
	}

	err = errors.New("fingerRemoteAccount: no match found in webfinger response")
	return
}
