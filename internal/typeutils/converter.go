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
	crand "crypto/rand"
	"math/big"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/filter/interaction"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

type Converter struct {
	state          *state.State
	defaultAvatars []string
	randAvatars    sync.Map
	visFilter      *visibility.Filter
	intFilter      *interaction.Filter
	randStats      atomic.Pointer[apimodel.RandomStats]
}

func NewConverter(state *state.State) *Converter {
	return &Converter{
		state:          state,
		defaultAvatars: populateDefaultAvatars(),
		visFilter:      visibility.NewFilter(state),
		intFilter:      interaction.NewFilter(state),
	}
}

// RandomStats returns or generates
// and returns random instance stats.
func (c *Converter) RandomStats() apimodel.RandomStats {
	now := time.Now()
	stats := c.randStats.Load()
	if stats != nil && time.Since(stats.Generated) < time.Hour {
		// Random stats are still
		// fresh (less than 1hr old),
		// so return them as-is.
		return *stats
	}

	// Generate new random stats.
	newStats := genRandStats()
	newStats.Generated = now
	c.randStats.Store(&newStats)
	return newStats
}

func genRandStats() apimodel.RandomStats {
	const (
		statusesMax = 10000000
		usersMax    = 1000000
	)

	statusesB, err := crand.Int(crand.Reader, big.NewInt(statusesMax))
	if err != nil {
		// Only errs if something is buggered with the OS.
		log.Panicf(nil, "error randomly generating statuses count: %v", err)
	}

	totalUsersB, err := crand.Int(crand.Reader, big.NewInt(usersMax))
	if err != nil {
		// Only errs if something is buggered with the OS.
		log.Panicf(nil, "error randomly generating users count: %v", err)
	}

	// Monthly users should only ever
	// be <= 100% of total users.
	totalUsers := totalUsersB.Int64()
	activeRatio := rand.Float64() //nolint
	mau := int64(float64(totalUsers) * activeRatio)

	return apimodel.RandomStats{
		Statuses:           statusesB.Int64(),
		TotalUsers:         totalUsers,
		MonthlyActiveUsers: mau,
	}
}
