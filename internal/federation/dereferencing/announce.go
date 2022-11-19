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
	"errors"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (d *deref) DereferenceAnnounce(ctx context.Context, announce *gtsmodel.Status, requestingUsername string) error {
	if announce.BoostOf == nil {
		// we can't do anything unfortunately
		return errors.New("DereferenceAnnounce: no URI to dereference")
	}

	// Parse the boosted status' URI
	boostedURI, err := url.Parse(announce.BoostOf.URI)
	if err != nil {
		return fmt.Errorf("DereferenceAnnounce: couldn't parse boosted status URI %s: %s", announce.BoostOf.URI, err)
	}

	// Check whether the originating status is from a blocked host
	if blocked, err := d.db.IsDomainBlocked(ctx, boostedURI.Host); blocked || err != nil {
		return fmt.Errorf("DereferenceAnnounce: domain %s is blocked", boostedURI.Host)
	}

	var boostedStatus *gtsmodel.Status

	if boostedURI.Host == config.GetHost() {
		// This is a local status, fetch from the database
		status, err := d.db.GetStatusByURI(ctx, boostedURI.String())
		if err != nil {
			return fmt.Errorf("DereferenceAnnounce: error fetching local status %q: %v", announce.BoostOf.URI, err)
		}

		// Set boosted status
		boostedStatus = status
	} else {
		// This is a boost of a remote status, we need to dereference it.
		status, statusable, err := d.GetStatus(ctx, requestingUsername, boostedURI, true, true)
		if err != nil {
			return fmt.Errorf("DereferenceAnnounce: error dereferencing remote status with id %s: %s", announce.BoostOf.URI, err)
		}

		// Dereference all statuses in the thread of the boosted status
		d.DereferenceThread(ctx, requestingUsername, boostedURI, status, statusable)

		// Set boosted status
		boostedStatus = status
	}

	announce.Content = boostedStatus.Content
	announce.ContentWarning = boostedStatus.ContentWarning
	announce.ActivityStreamsType = boostedStatus.ActivityStreamsType
	announce.Sensitive = boostedStatus.Sensitive
	announce.Language = boostedStatus.Language
	announce.Text = boostedStatus.Text
	announce.BoostOfID = boostedStatus.ID
	announce.BoostOfAccountID = boostedStatus.AccountID
	announce.Visibility = boostedStatus.Visibility
	announce.Federated = boostedStatus.Federated
	announce.Boostable = boostedStatus.Boostable
	announce.Replyable = boostedStatus.Replyable
	announce.Likeable = boostedStatus.Likeable
	announce.BoostOf = boostedStatus

	return nil
}
