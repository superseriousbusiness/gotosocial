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

package db

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Instance contains functions for instance-level actions (counting instance users etc.).
type Instance interface {
	// CountInstanceUsers returns the number of known accounts registered with the given domain.
	CountInstanceUsers(ctx context.Context, domain string) (int, error)

	// CountInstanceStatuses returns the number of known statuses posted from the given domain.
	CountInstanceStatuses(ctx context.Context, domain string) (int, error)

	// CountInstanceDomains returns the number of known instances known that the given domain federates with.
	CountInstanceDomains(ctx context.Context, domain string) (int, error)

	// GetInstance returns the instance entry for the given domain, if it exists.
	GetInstance(ctx context.Context, domain string) (*gtsmodel.Instance, error)

	// GetInstanceByID returns the instance entry corresponding to the given id, if it exists.
	GetInstanceByID(ctx context.Context, id string) (*gtsmodel.Instance, error)

	// PopulateInstance populates the struct pointers on the given instance.
	PopulateInstance(ctx context.Context, instance *gtsmodel.Instance) error

	// PutInstance inserts the given instance into the database.
	PutInstance(ctx context.Context, instance *gtsmodel.Instance) error

	// UpdateInstance updates the given instance entry.
	UpdateInstance(ctx context.Context, instance *gtsmodel.Instance, columns ...string) error

	// GetInstanceAccounts returns a slice of accounts from the given instance, arranged by ID.
	GetInstanceAccounts(ctx context.Context, domain string, maxID string, limit int) ([]*gtsmodel.Account, error)

	// GetInstancePeers returns a slice of instances that the host instance knows about.
	GetInstancePeers(ctx context.Context, includeSuspended bool) ([]*gtsmodel.Instance, error)

	// GetInstanceModeratorAddresses returns a slice of email addresses belonging to active
	// (as in, not suspended) moderators + admins on this instance.
	GetInstanceModeratorAddresses(ctx context.Context) ([]string, error)

	// GetInstanceModerators returns a slice of accounts belonging to active
	// (as in, non suspended) moderators + admins on this instance.
	GetInstanceModerators(ctx context.Context) ([]*gtsmodel.Account, error)
}
