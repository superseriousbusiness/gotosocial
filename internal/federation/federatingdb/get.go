// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package federatingdb

import (
	"context"
	"errors"
	"net/url"

	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
)

var ErrNotImplemented = errors.New("not implemented")

// Get returns the database entry for the specified id.
//
// The library makes this call only after acquiring a lock first.
//
// Implementation notes: in GoToSocial this function should *only*
// be used for internal dereference calls. Everything coming from the
// outside goes via the handlers defined in internal/api/activitypub.
//
// Normally with go-fed this function would get used in lots of
// places for the side effect callback handlers, but since we override
// everything and handle side effects ourselves, the only two places
// this function actually ends up getting called are:
//
//   - vendor/code.superseriousbusiness.org/activity/pub/side_effect_actor.go
//     to get outbox actor inside the prepare function.
//   - internal/transport/controller.go to try to shortcut deref a local item.
//
// It may be useful in future to add more matching here so that more
// stuff can be shortcutted by the dereferencer, saving HTTP calls.
func (f *DB) Get(ctx context.Context, id *url.URL) (value vocab.Type, err error) {
	log.DebugKV(ctx, "id", id)

	// Ensure our host, for safety.
	if id.Host != config.GetHost() {
		return nil, gtserror.Newf("%s was not for our host", id)
	}

	if username, _ := uris.ParseUserPath(id); username != "" {
		acct, err := f.state.DB.GetAccountByUsernameDomain(
			gtscontext.SetBarebones(ctx),
			username,
			"",
		)
		if err != nil {
			return nil, err
		}
		return f.converter.AccountToAS(ctx, acct)

	} else if _, statusID, _ := uris.ParseStatusesPath(id); statusID != "" {
		status, err := f.state.DB.GetStatusByID(ctx, statusID)
		if err != nil {
			return nil, err
		}
		return f.converter.StatusToAS(ctx, status)

	} else if username, _ := uris.ParseFollowersPath(id); username != "" {
		acct, err := f.state.DB.GetAccountByUsernameDomain(
			gtscontext.SetBarebones(ctx),
			username,
			"",
		)
		if err != nil {
			return nil, err
		}

		acctURI, err := url.Parse(acct.URI)
		if err != nil {
			return nil, err
		}

		return f.Followers(ctx, acctURI)

	} else if username, _ := uris.ParseFollowingPath(id); username != "" {
		acct, err := f.state.DB.GetAccountByUsernameDomain(
			gtscontext.SetBarebones(ctx),
			username,
			"",
		)
		if err != nil {
			return nil, err
		}

		acctURI, err := url.Parse(acct.URI)
		if err != nil {
			return nil, err
		}

		return f.Following(ctx, acctURI)

	} else if uris.IsAcceptsPath(id) {
		return f.GetAccept(ctx, id)
	}

	// Nothing found, the caller
	// will have to deal with this.
	return nil, ErrNotImplemented
}
