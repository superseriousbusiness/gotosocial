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
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

func (d *deref) fingerRemoteAccount(ctx context.Context, t transport.Transport, targetUsername string, targetDomain string) (*url.URL, error) {
	b, err := t.Finger(ctx, targetUsername, targetDomain)
	if err != nil {
		return nil, fmt.Errorf("FingerRemoteAccount: error fingering @%s@%s: %s", targetUsername, targetDomain, err)
	}

	resp := &apimodel.WellKnownResponse{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, fmt.Errorf("FingerRemoteAccount: could not unmarshal server response as WebfingerAccountResponse while dereferencing @%s@%s: %s", targetUsername, targetDomain, err)
	}

	if len(resp.Links) == 0 {
		return nil, fmt.Errorf("FingerRemoteAccount: no links found in webfinger response %s", string(b))
	}

	// look through the links for the first one that matches "application/activity+json", this is what we need
	for _, l := range resp.Links {
		if strings.EqualFold(l.Type, "application/activity+json") {
			if l.Href == "" || l.Rel != "self" {
				continue
			}
			accountURI, err := url.Parse(l.Href)
			if err != nil {
				return nil, fmt.Errorf("FingerRemoteAccount: couldn't parse url %s: %s", l.Href, err)
			}
			// found it!
			return accountURI, nil
		}
	}

	return nil, errors.New("FingerRemoteAccount: no match found in webfinger response")
}
