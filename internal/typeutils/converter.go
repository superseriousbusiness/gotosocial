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

package typeutils

import (
	"log"
	"sync"
	"time"

	"codeberg.org/gruf/go-cache/v3"
	"github.com/superseriousbusiness/gotosocial/internal/filter/interaction"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

type Converter struct {
	state          *state.State
	defaultAvatars []string
	randAvatars    sync.Map
	visFilter      *visibility.Filter
	intFilter      *interaction.Filter

	// TTL cache of statuses -> filterable text fields.
	// To ensure up-to-date fields, cache is keyed as:
	// [status.ID][status.UpdatedAt.Unix()]`
	statusesFilterableFields cache.TTLCache[string, []string]
}

func NewConverter(state *state.State) *Converter {
	statusHashesToFilterableText := cache.NewTTL[string, []string](0, 512, 0)
	statusHashesToFilterableText.SetTTL(time.Hour, true)
	if !statusHashesToFilterableText.Start(time.Minute) {
		log.Panic(nil, "failed to start statusHashesToFilterableText cache")
	}

	return &Converter{
		state:                    state,
		defaultAvatars:           populateDefaultAvatars(),
		visFilter:                visibility.NewFilter(state),
		intFilter:                interaction.NewFilter(state),
		statusesFilterableFields: statusHashesToFilterableText,
	}
}
