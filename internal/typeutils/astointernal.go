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

package typeutils

import (
	"errors"
	"fmt"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (c *converter) ASPersonToAccount(person vocab.ActivityStreamsPerson) (*gtsmodel.Account, error) {
	// first check if we actually already know this person
	uriProp := person.GetJSONLDId()
	if uriProp == nil || !uriProp.IsIRI() {
		return nil, errors.New("no id property found on person, or id was not an iri")
	}
	uri := uriProp.GetIRI()

	acct := &gtsmodel.Account{}
	if err := c.db.GetWhere("uri", uri.String(), acct); err == nil {
		// we already know this account so we can skip generating it
		return acct, nil
	} else {
		if _, ok := err.(db.ErrNoEntries); !ok {
			// we don't know the account and there's been a real error
			return nil, fmt.Errorf("error getting account with uri %s from the database: %s", uri.String(), err)
		}
	}

	// we don't know the account so we need to generate it from the person -- at least we already have the URI!
	acct = &gtsmodel.Account{}
	acct.URI = uri.String()

	// Username aka preferredUsername
	// We need this one so bail if it's not set.
	username, err := extractPreferredUsername(person)
	if err != nil {
		return nil, fmt.Errorf("couldn't extract username: %s", err)
	}
	acct.Username = username

	// Domain
	acct.Domain = uri.Host

	// avatar aka icon
	// if this one isn't extractable in a format we recognise we'll just skip it
	if avatarURL, err := extractIconURL(person); err == nil {
		acct.AvatarRemoteURL = avatarURL.String()
	}

	// header aka image
	// if this one isn't extractable in a format we recognise we'll just skip it
	if headerURL, err := extractImageURL(person); err == nil {
		acct.HeaderRemoteURL = headerURL.String()
	}

	// display name aka name
	// we default to the username, but take the more nuanced name property if it exists
	acct.DisplayName = username
	if displayName, err := extractName(person); err == nil {
		acct.DisplayName = displayName
	}

	// fields aka attachment array
    // TODO

    // note aka summary
	// TODO

    // bot
	// TODO: parse this from application vs. person type

    // locked aka manuallyApprovesFollowers
    // TODO

    // discoverable
	// TODO

    // url property
	// TODO

    // InboxURI
    // TODO

    // OutboxURI
	// TODO

	// FollowingURI
	// TODO

	// FollowersURI
	// TODO

    // FeaturedURI
	// TODO

	// FeaturedTagsURI
	// TODO

    // alsoKnownAs
	// TODO

    // publicKey
	// TODO

	return acct, nil
}
