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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// verifyAccountKeysOnUpdate verifies that account's public key hasn't changed on update from
// our existing stored representation, UNLESS the key has been explicitly expired (i.e. key rotation).
func verifyAccountKeysOnUpdate(existing, latest *gtsmodel.Account, now time.Time, federated bool) bool {
	if federated {
		// If this data was federated
		// to us then we implicitly trust
		// it on the grounds that it
		// passed any signature checks.
		return true
	}

	if existing.PublicKey == nil {
		// New account which has been
		// passed as a placeholder.
		// This is always permitted.
		return true
	}

	// Ensure that public keys have not changed.
	if existing.PublicKey.Equal(latest.PublicKey) &&
		existing.PublicKeyURI == latest.PublicKeyURI {
		return true
	}

	// The only time that an account key change is
	// permitted is when it is marked as expired.
	return !existing.PublicKeyExpiresAt.IsZero() &&
		existing.PublicKeyExpiresAt.Before(now)
}
