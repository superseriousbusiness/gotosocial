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

package gtsmodel

import (
	"path"
	"time"
)

// AdminActionCategory describes the category
// of entity that this admin action targets.
type AdminActionCategory string

const (
	AdminActionCategoryAccount AdminActionCategory = "account"
	AdminActionCategoryDomain  AdminActionCategory = "domain"
)

// AdminActionType describes a type of
// action taken on an entity by an admin.
type AdminActionType string

const (
	AdminActionDisable  AdminActionType = "disable"
	AdminActionReenable AdminActionType = "reenable"

	AdminActionSilence   AdminActionType = "silence"
	AdminActionUnsilence AdminActionType = "unsilence"

	AdminActionSuspend   AdminActionType = "suspend"
	AdminActionUnsuspend AdminActionType = "unsuspend"
)

// AdminAction models an action taken by an instance administrator towards an account, domain, etc.
type AdminAction struct {
	ID             string              `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // ID of this item in the database.
	CreatedAt      time.Time           `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // Creation time of this item.
	UpdatedAt      time.Time           `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // Last updated time of this item.
	CompletedAt    time.Time           `bun:"type:timestamptz,nullzero"`                                   // Completion time of this item.
	TargetCategory AdminActionCategory `bun:",nullzero,notnull"`                                           // Category of the entity targeted by this action.
	TargetID       string              `bun:",nullzero,notnull"`                                           // Identifier of the target. May be a ULID (in case of accounts), or a domain name (in case of domains).
	Target         interface{}         `bun:"-"`                                                           // Target of the action. Might be a domain string, might be an account.
	Type           AdminActionType     `bun:",nullzero,notnull"`                                           // Type of action that was taken.
	AccountID      string              `bun:"type:CHAR(26),notnull,nullzero"`                              // Who performed this admin action.
	Account        *Account            `bun:"rel:has-one"`                                                 // Account corresponding to accountID
	Text           string              `bun:",nullzero"`                                                   // Free text field for explaining why this action was taken, or adding a note about this action.
	SendEmail      *bool               `bun:",nullzero,notnull,default:false"`                             // Send an email to the target account's user to explain what happened (local accounts only).
	ReportIDs      []string            `bun:"reports,array"`                                               // IDs of any reports cited when creating this action.
	Reports        []*Report           `bun:"-"`                                                           // Reports corresponding to ReportIDs.
	Errors         []string            `bun:",array"`                                                      // String value of any error(s) encountered while processing. May be helpful for admins to debug.
}

// Key returns a key for the AdminAction which is
// unique only on its TargetCategory and TargetID
// fields. This key can be used to check if this
// AdminAction overlaps with another action performed
// on the same target, regardless of the Type of
// either this or the other action.
func (a *AdminAction) Key() string {
	return path.Join(
		string(a.TargetCategory),
		a.TargetID,
	)
}
