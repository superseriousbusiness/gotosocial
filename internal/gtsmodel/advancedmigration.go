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
	"encoding/json"
	"time"
)

// AdvancedMigration stores state for an "advanced migration", which is a migration
// that doesn't fit into the Bun migration framework.
type AdvancedMigration struct {
	ID        string    `bun:",pk,nullzero,notnull,unique"`                                 // id of this migration (preassigned, not a ULID)
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	StateJSON string    `bun:",nullzero"`                                                   // JSON dump of the migration state
	Finished  *bool     `bun:",nullzero,notnull,default:false"`                             // has this migration finished?
}

func AdvancedMigrationLoad[State any](a *AdvancedMigration) (State, error) {
	var state State
	err := json.Unmarshal([]byte(a.StateJSON), state)
	return state, err
}

func AdvancedMigrationStore[State any](a *AdvancedMigration, state State) error {
	bytes, err := json.Marshal(state)
	if err != nil {
		return err
	}
	a.StateJSON = string(bytes)
	return nil
}
