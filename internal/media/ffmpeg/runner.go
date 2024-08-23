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

package ffmpeg

import (
	"context"

	"github.com/tetratelabs/wazero"
)

// runner simply abstracts away the complexities
// of limiting the number of concurrent running
// instances of a particular WebAssembly module.
type runner struct{ pool chan struct{} }

// Init initializes the runner to
// only allow 'n' concurrently running
// instances. Special cases include 0
// which clamps to 1, and < 0 which
// disables the limit alltogether.
func (r *runner) Init(n int) {

	// Reset pool.
	r.pool = nil

	// Clamp to 1.
	if n <= 0 {
		n = 1
	}

	// Allocate new pool channel.
	r.pool = make(chan struct{}, n)
	for i := 0; i < n; i++ {
		r.pool <- struct{}{}
	}
}

// Run will attempt to pass the given compiled WebAssembly module with args to run(), waiting on
// the receiving runner until a free slot is available to run an instance, (if a limit is enabled).
func (r *runner) Run(ctx context.Context, cmod wazero.CompiledModule, args Args) (uint32, error) {
	select {
	// Context canceled.
	case <-ctx.Done():
		return 0, ctx.Err()

	// Slot acquired.
	case <-r.pool:
	}

	// Release slot back to pool on end.
	defer func() { r.pool <- struct{}{} }()

	// Pass to main module runner.
	return run(ctx, cmod, args)
}
