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

package dereferencing

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (d *deref) DereferenceAnnounce(announce *gtsmodel.Status, requestingUsername string) error {
	if announce.GTSBoostedStatus == nil || announce.GTSBoostedStatus.URI == "" {
		// we can't do anything unfortunately
		return errors.New("DereferenceAnnounce: no URI to dereference")
	}

	boostedStatusURI, err := url.Parse(announce.GTSBoostedStatus.URI)
	if err != nil {
		return fmt.Errorf("DereferenceAnnounce: couldn't parse boosted status URI %s: %s", announce.GTSBoostedStatus.URI, err)
	}
	if blocked, err := d.blockedDomain(boostedStatusURI.Host); blocked || err != nil {
		return fmt.Errorf("DereferenceAnnounce: domain %s is blocked", boostedStatusURI.Host)
	}

	// dereference statuses in the thread of the boosted status
	if err := d.DereferenceThread(requestingUsername, boostedStatusURI); err != nil {
		return fmt.Errorf("DereferenceAnnounce: error dereferencing thread of boosted status: %s", err)
	}

	boostedStatus, _, _, err := d.GetRemoteStatus(requestingUsername, boostedStatusURI, false)
	if err != nil {
		return fmt.Errorf("DereferenceAnnounce: error dereferencing remote status with id %s: %s", announce.GTSBoostedStatus.URI, err)
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
	announce.VisibilityAdvanced = boostedStatus.VisibilityAdvanced
	announce.GTSBoostedStatus = boostedStatus
	return nil
}
