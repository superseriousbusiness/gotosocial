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

package dereferencing

import (
	"net/url"
	"sync"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// FreshnessWindow represents a duration in which a
// Status or Account is still considered to be "fresh"
// (ie., not in need of a refresh from remote), if its
// last FetchedAt value falls within the window.
//
// For example, if an Account was FetchedAt 09:00, and it
// is now 12:00, then it would be considered "fresh"
// according to DefaultAccountFreshness, but not according
// to Fresh, which would indicate that the Account requires
// refreshing from remote.
type FreshnessWindow time.Duration

var (
	// 6 hours.
	//
	// Default window for doing a
	// fresh dereference of an Account.
	DefaultAccountFreshness = util.Ptr(FreshnessWindow(6 * time.Hour))

	// 2 hours.
	//
	// Default window for doing a
	// fresh dereference of a Status.
	DefaultStatusFreshness = util.Ptr(FreshnessWindow(2 * time.Hour))

	// 5 minutes.
	//
	// Fresh is useful when you're wanting
	// a more up-to-date model of something
	// that exceeds default freshness windows.
	//
	// This is tuned to be quite fresh without
	// causing loads of dereferencing calls.
	Fresh = util.Ptr(FreshnessWindow(5 * time.Minute))
)

// Dereferencer wraps logic and functionality for doing dereferencing
// of remote accounts, statuses, etc, from federated instances.
type Dereferencer struct {
	state               *state.State
	converter           *typeutils.Converter
	transportController transport.Controller
	mediaManager        *media.Manager
	visibility          *visibility.Filter

	// all protected by State{}.FedLocks.
	derefAvatars map[string]*media.ProcessingMedia
	derefHeaders map[string]*media.ProcessingMedia
	derefEmojis  map[string]*media.ProcessingEmoji

	handshakes   map[string][]*url.URL
	handshakesMu sync.Mutex
}

// NewDereferencer returns a Dereferencer initialized with the given parameters.
func NewDereferencer(
	state *state.State,
	converter *typeutils.Converter,
	transportController transport.Controller,
	visFilter *visibility.Filter,
	mediaManager *media.Manager,
) Dereferencer {
	return Dereferencer{
		state:               state,
		converter:           converter,
		transportController: transportController,
		mediaManager:        mediaManager,
		visibility:          visFilter,
		derefAvatars:        make(map[string]*media.ProcessingMedia),
		derefHeaders:        make(map[string]*media.ProcessingMedia),
		derefEmojis:         make(map[string]*media.ProcessingEmoji),
		handshakes:          make(map[string][]*url.URL),
	}
}
