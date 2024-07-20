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

	"codeberg.org/gruf/go-ffmpreg/wasm"
)

// wasmInstancePool wraps a wasm.Instantiator{} and a
// channel of wasm.Instance{}s to provide a concurrency
// safe pool of WebAssembly module instances capable of
// compiling new instances on-the-fly, with a predetermined
// maximum number of concurrent instances at any one time.
type wasmInstancePool struct {
	inst wasm.Instantiator
	pool chan *wasm.Instance
}

func (p *wasmInstancePool) Init(ctx context.Context, sz int) error {
	// Initialize for first time
	// to preload module into the
	// wazero compilation cache.
	inst, err := p.inst.New(ctx)
	if err != nil {
		return err
	}

	// Clamp to 1.
	if sz <= 0 {
		sz = 1
	}

	// Allocate new pool instance channel.
	p.pool = make(chan *wasm.Instance, sz)

	// Store only one
	// open instance
	// at init time.
	p.pool <- inst

	// Fill reminaing with closed
	// instances for later opening.
	for i := 0; i < sz-1; i++ {
		p.pool <- new(wasm.Instance)
	}

	return nil
}

func (p *wasmInstancePool) Run(ctx context.Context, args wasm.Args) (uint32, error) {
	var inst *wasm.Instance

	select {
	// Context canceled.
	case <-ctx.Done():
		return 0, ctx.Err()

	// Acquire instance.
	case inst = <-p.pool:

		// Ensure instance is
		// ready for running.
		if inst.IsClosed() {
			var err error
			inst, err = p.inst.New(ctx)
			if err != nil {
				return 0, err
			}
		}
	}

	// Release instance to pool on end.
	defer func() { p.pool <- inst }()

	// Pass args to instance.
	return inst.Run(ctx, args)
}
